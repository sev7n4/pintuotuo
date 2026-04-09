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
	stopChan   chan struct{}
	reloadCh   chan struct{}
	running    bool
	mutex      sync.Mutex
	checkLevel HealthCheckLevel // 仅用于 GetStats / SetCheckLevel 展示，全局 tick 来自 platform_settings
}

var (
	scheduler     *HealthScheduler
	schedulerOnce sync.Once
)

func GetHealthScheduler() *HealthScheduler {
	schedulerOnce.Do(func() {
		scheduler = &HealthScheduler{
			checker:    NewHealthChecker(),
			stopChan:   make(chan struct{}, 1),
			reloadCh:   make(chan struct{}, 1),
			checkLevel: HealthCheckLevelMedium,
		}
	})
	return scheduler
}

// SignalReload 配置热更新时唤醒调度循环（非阻塞）。
func (s *HealthScheduler) SignalReload() {
	if s == nil {
		return
	}
	select {
	case s.reloadCh <- struct{}{}:
	default:
	}
}

func (s *HealthScheduler) Start() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.running {
		return
	}

	if err := ReloadPlatformSettingsCache(context.Background()); err != nil {
		log.Printf("Health scheduler: load platform settings: %v (using cache/defaults)", err)
	}

	s.running = true
	go s.run()

	log.Printf("Health scheduler loop started")
}

func (s *HealthScheduler) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.running {
		return
	}

	s.running = false
	s.stopChan <- struct{}{}

	log.Println("Health scheduler stopped")
}

func (s *HealthScheduler) run() {
	for {
		select {
		case <-s.stopChan:
			return
		default:
		}

		cfg := GetHealthSchedulerPlatformConfig()
		if !cfg.Enabled {
			select {
			case <-s.stopChan:
				return
			case <-s.reloadCh:
				continue
			case <-time.After(5 * time.Second):
				_ = ReloadPlatformSettingsCache(context.Background())
				continue
			}
		}

		timer := time.NewTimer(time.Duration(cfg.IntervalSeconds) * time.Second)
		select {
		case <-s.stopChan:
			if !timer.Stop() {
				<-timer.C
			}
			return
		case <-s.reloadCh:
			if !timer.Stop() {
				<-timer.C
			}
			_ = ReloadPlatformSettingsCache(context.Background())
			continue
		case <-timer.C:
			s.performChecks()
		}
	}
}

func (s *HealthScheduler) performChecks() {
	ctx := context.Background()
	db := config.GetDB()
	if db == nil {
		return
	}

	cfg := GetHealthSchedulerPlatformConfig()
	if !cfg.Enabled {
		return
	}
	if cfg.Batch < 1 {
		return
	}

	rows, err := db.Query(`
		SELECT id, merchant_id, provider, api_key_encrypted, endpoint_url, 
		       health_check_level, health_status, last_health_check_at
		FROM merchant_api_keys 
		WHERE status = 'active'
		ORDER BY last_health_check_at ASC NULLS FIRST
		LIMIT $1`, cfg.Batch)
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
	log.Printf("Health scheduler check_level (per-key default display) set to: %s", level)
}

func (s *HealthScheduler) TriggerImmediateCheck(apiKeyID int) error {
	ctx := context.Background()
	return s.checker.TriggerActiveCheck(ctx, apiKeyID)
}

func (s *HealthScheduler) GetStats() map[string]interface{} {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	cfg := GetHealthSchedulerPlatformConfig()
	refInterval := s.checker.GetHealthCheckInterval(string(s.checkLevel))

	return map[string]interface{}{
		"running":                           s.running,
		"check_level":                       string(s.checkLevel),
		"interval":                          refInterval,
		"health_scheduler_enabled":          cfg.Enabled,
		"health_scheduler_interval_seconds": cfg.IntervalSeconds,
		"health_scheduler_batch":            cfg.Batch,
	}
}
