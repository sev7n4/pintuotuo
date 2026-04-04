package services

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/models"
)

type HealthScheduler struct {
	checker    *HealthChecker
	ticker     *time.Ticker
	stopChan   chan struct{}
	running    bool
	mutex      sync.Mutex
	checkLevel HealthCheckLevel
}

var (
	scheduler     *HealthScheduler
	schedulerOnce sync.Once
)

func GetHealthScheduler() *HealthScheduler {
	schedulerOnce.Do(func() {
		scheduler = &HealthScheduler{
			checker:    NewHealthChecker(),
			stopChan:   make(chan struct{}),
			checkLevel: HealthCheckLevelMedium,
		}
	})
	return scheduler
}

func (s *HealthScheduler) Start() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.running {
		return
	}

	s.running = true

	interval := s.checker.GetHealthCheckInterval(string(s.checkLevel))
	s.ticker = time.NewTicker(time.Duration(interval) * time.Second)

	go s.run()

	log.Printf("Health scheduler started with interval: %d seconds", interval)
}

func (s *HealthScheduler) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.running {
		return
	}

	s.running = false
	s.ticker.Stop()
	s.stopChan <- struct{}{}

	log.Println("Health scheduler stopped")
}

func (s *HealthScheduler) run() {
	for {
		select {
		case <-s.ticker.C:
			s.performChecks()
		case <-s.stopChan:
			return
		}
	}
}

func (s *HealthScheduler) performChecks() {
	ctx := context.Background()
	db := config.GetDB()
	if db == nil {
		return
	}

	rows, err := db.Query(`
		SELECT id, merchant_id, provider, api_key_encrypted, endpoint_url, 
		       health_check_level, health_status, last_health_check_at
		FROM merchant_api_keys 
		WHERE status = 'active'
		ORDER BY last_health_check_at ASC NULLS FIRST
		LIMIT 10`)
	if err != nil {
		log.Printf("Failed to query api keys for health check: %v", err)
		return
	}
	defer rows.Close()

	var apiKeys []models.MerchantAPIKey
	for rows.Next() {
		var key models.MerchantAPIKey
		var lastCheckAt *time.Time
		err := rows.Scan(
			&key.ID, &key.MerchantID, &key.Provider, &key.APIKeyEncrypted,
			&key.EndpointURL, &key.HealthCheckLevel, &key.HealthStatus, &lastCheckAt,
		)
		if err != nil {
			continue
		}
		if lastCheckAt != nil {
			key.LastHealthCheckAt = lastCheckAt
		}
		apiKeys = append(apiKeys, key)
	}

	for _, apiKey := range apiKeys {
		if !s.checker.ShouldPerformCheck(&apiKey) {
			continue
		}

		go s.checkSingleProvider(ctx, &apiKey)
	}
}

func (s *HealthScheduler) checkSingleProvider(ctx context.Context, apiKey *models.MerchantAPIKey) {
	level := HealthCheckLevel(apiKey.HealthCheckLevel)

	var result *HealthCheckResult
	var err error

	switch level {
	case HealthCheckLevelHigh:
		result, err = s.checker.LightweightPing(ctx, apiKey)
	default:
		result, err = s.checker.FullVerification(ctx, apiKey)
	}

	if err != nil {
		log.Printf("Health check failed for api_key_id=%d: %v", apiKey.ID, err)
		return
	}

	if result != nil {
		if err := s.checker.SaveHealthCheckResult(ctx, apiKey.ID, result); err != nil {
			log.Printf("Failed to save health check result for api_key_id=%d: %v", apiKey.ID, err)
		}
	}
}

func (s *HealthScheduler) SetCheckLevel(level HealthCheckLevel) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.checkLevel = level

	if s.running && s.ticker != nil {
		interval := s.checker.GetHealthCheckInterval(string(level))
		s.ticker.Reset(time.Duration(interval) * time.Second)
		log.Printf("Health scheduler interval updated to: %d seconds", interval)
	}
}

func (s *HealthScheduler) TriggerImmediateCheck(apiKeyID int) error {
	ctx := context.Background()
	return s.checker.TriggerActiveCheck(ctx, apiKeyID)
}

func (s *HealthScheduler) GetStats() map[string]interface{} {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return map[string]interface{}{
		"running":     s.running,
		"check_level": string(s.checkLevel),
		"interval":    s.checker.GetHealthCheckInterval(string(s.checkLevel)),
	}
}
