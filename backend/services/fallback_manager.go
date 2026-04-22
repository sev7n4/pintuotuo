package services

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/logger"
)

type FallbackManager struct {
	db              *sql.DB
	fallbackCache   map[string]*FallbackState
	cacheMutex      sync.RWMutex
	fallbackHistory []FallbackEvent
	historyMutex    sync.Mutex
	maxHistorySize  int
}

type FallbackState struct {
	ProviderCode     string
	CurrentMode      string
	OriginalMode     string
	FallbackMode     string
	FallbackEndpoint string
	FallbackAt       time.Time
	FailureCount     int
	Reason           string
}

type FallbackEvent struct {
	Timestamp       time.Time `json:"timestamp"`
	ProviderCode    string    `json:"provider_code"`
	OriginalMode    string    `json:"original_mode"`
	FallbackMode    string    `json:"fallback_mode"`
	Reason          string    `json:"reason"`
	Success         bool      `json:"success"`
	RecoveryAttempt bool      `json:"recovery_attempt"`
}

var (
	fallbackManager     *FallbackManager
	fallbackManagerOnce sync.Once
)

func GetFallbackManager() *FallbackManager {
	fallbackManagerOnce.Do(func() {
		fallbackManager = &FallbackManager{
			db:              config.GetDB(),
			fallbackCache:   make(map[string]*FallbackState),
			fallbackHistory: make([]FallbackEvent, 0, 100),
			maxHistorySize:  100,
		}
	})
	return fallbackManager
}

func (m *FallbackManager) ShouldFallback(ctx context.Context, providerCode string, currentMode string, failureCount int) (*FallbackState, bool) {
	m.cacheMutex.RLock()
	state, exists := m.fallbackCache[providerCode]
	m.cacheMutex.RUnlock()

	if exists && time.Since(state.FallbackAt) < 5*time.Minute {
		return state, true
	}

	if failureCount >= 3 {
		return nil, false
	}

	return nil, false
}

func (m *FallbackManager) TriggerFallback(ctx context.Context, providerCode string, originalMode string, fallbackMode string, fallbackEndpoint string, reason string) error {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	state := &FallbackState{
		ProviderCode:     providerCode,
		CurrentMode:      fallbackMode,
		OriginalMode:     originalMode,
		FallbackMode:     fallbackMode,
		FallbackEndpoint: fallbackEndpoint,
		FallbackAt:       time.Now(),
		FailureCount:     1,
		Reason:           reason,
	}

	m.fallbackCache[providerCode] = state

	m.recordFallbackEvent(FallbackEvent{
		Timestamp:       time.Now(),
		ProviderCode:    providerCode,
		OriginalMode:    originalMode,
		FallbackMode:    fallbackMode,
		Reason:          reason,
		Success:         true,
		RecoveryAttempt: false,
	})

	logger.LogInfo(ctx, "fallback_manager", "Fallback triggered",
		map[string]interface{}{
			"provider":      providerCode,
			"original_mode": originalMode,
			"fallback_mode": fallbackMode,
			"reason":        reason,
		},
	)

	return nil
}

func (m *FallbackManager) GetFallbackState(providerCode string) *FallbackState {
	m.cacheMutex.RLock()
	defer m.cacheMutex.RUnlock()

	state, exists := m.fallbackCache[providerCode]
	if !exists {
		return nil
	}

	if time.Since(state.FallbackAt) > 10*time.Minute {
		return nil
	}

	return state
}

func (m *FallbackManager) ClearFallback(providerCode string, recovered bool) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	state, exists := m.fallbackCache[providerCode]
	if !exists {
		return
	}

	m.recordFallbackEvent(FallbackEvent{
		Timestamp:       time.Now(),
		ProviderCode:    providerCode,
		OriginalMode:    state.OriginalMode,
		FallbackMode:    state.FallbackMode,
		Reason:          "recovery",
		Success:         recovered,
		RecoveryAttempt: true,
	})

	delete(m.fallbackCache, providerCode)

	logger.LogInfo(context.Background(), "fallback_manager", "Fallback cleared",
		map[string]interface{}{
			"provider":  providerCode,
			"recovered": recovered,
		},
	)
}

func (m *FallbackManager) recordFallbackEvent(event FallbackEvent) {
	m.historyMutex.Lock()
	defer m.historyMutex.Unlock()

	if len(m.fallbackHistory) >= m.maxHistorySize {
		m.fallbackHistory = m.fallbackHistory[1:]
	}

	m.fallbackHistory = append(m.fallbackHistory, event)
}

func (m *FallbackManager) GetFallbackHistory(limit int) []FallbackEvent {
	m.historyMutex.Lock()
	defer m.historyMutex.Unlock()

	if limit <= 0 || limit > len(m.fallbackHistory) {
		limit = len(m.fallbackHistory)
	}

	start := len(m.fallbackHistory) - limit
	if start < 0 {
		start = 0
	}

	result := make([]FallbackEvent, limit)
	copy(result, m.fallbackHistory[start:])

	return result
}

func (m *FallbackManager) GetFallbackStats() map[string]interface{} {
	m.cacheMutex.RLock()
	activeFallbacks := len(m.fallbackCache)
	m.cacheMutex.RUnlock()

	m.historyMutex.Lock()
	totalEvents := len(m.fallbackHistory)
	m.historyMutex.Unlock()

	return map[string]interface{}{
		"active_fallbacks": activeFallbacks,
		"total_events":     totalEvents,
	}
}

func (m *FallbackManager) ExecuteWithFallback(
	ctx context.Context,
	providerCode string,
	decision *RouteDecision,
	primaryFunc func() error,
) error {
	err := primaryFunc()
	if err == nil {
		if state := m.GetFallbackState(providerCode); state != nil {
			m.ClearFallback(providerCode, true)
		}
		return nil
	}

	if decision.FallbackMode != "" && decision.FallbackEndpoint != "" {
		state, shouldFallback := m.ShouldFallback(ctx, providerCode, decision.Mode, 1)
		if shouldFallback || state == nil {
			fallbackErr := m.TriggerFallback(
				ctx,
				providerCode,
				decision.Mode,
				decision.FallbackMode,
				decision.FallbackEndpoint,
				fmt.Sprintf("primary endpoint failed: %v", err),
			)
			if fallbackErr != nil {
				logger.LogError(ctx, "fallback_manager", "Failed to trigger fallback",
					fallbackErr,
					map[string]interface{}{
						"provider": providerCode,
					},
				)
			}
		}
	}

	return err
}
