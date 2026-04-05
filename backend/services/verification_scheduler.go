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

type VerificationScheduler struct {
	db        *sql.DB
	validator *APIKeyValidator
	interval  time.Duration
	stopChan  chan struct{}
	running   bool
	mu        sync.Mutex
}

var (
	verificationScheduler     *VerificationScheduler
	verificationSchedulerOnce sync.Once
)

func GetVerificationScheduler() *VerificationScheduler {
	verificationSchedulerOnce.Do(func() {
		verificationScheduler = &VerificationScheduler{
			db:        config.GetDB(),
			validator: GetAPIKeyValidator(),
			interval:  VerificationInterval,
			stopChan:  make(chan struct{}),
			running:   false,
		}
	})
	return verificationScheduler
}

func (s *VerificationScheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("scheduler is already running")
	}

	s.running = true
	go s.run()

	logger.LogInfo(context.Background(), "verification_scheduler", "Scheduler started", map[string]interface{}{
		"interval": s.interval.String(),
	})

	return nil
}

func (s *VerificationScheduler) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return fmt.Errorf("scheduler is not running")
	}

	close(s.stopChan)
	s.running = false

	logger.LogInfo(context.Background(), "verification_scheduler", "Scheduler stopped", nil)

	return nil
}

func (s *VerificationScheduler) run() {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	s.performPeriodicVerification()

	for {
		select {
		case <-ticker.C:
			s.performPeriodicVerification()
		case <-s.stopChan:
			return
		}
	}
}

func (s *VerificationScheduler) performPeriodicVerification() {
	ctx := context.Background()
	startTime := time.Now()

	logger.LogInfo(ctx, "verification_scheduler", "Starting periodic verification", nil)

	apiKeys, err := s.getAPIKeysForVerification()
	if err != nil {
		logger.LogError(ctx, "verification_scheduler", "Failed to get API keys for verification", err, nil)
		return
	}

	if len(apiKeys) == 0 {
		logger.LogInfo(ctx, "verification_scheduler", "No API keys need verification", nil)
		return
	}

	successCount := 0
	failedCount := 0

	for _, key := range apiKeys {
		err := s.validator.ValidateAsync(key.ID, key.Provider, key.APIKeyEncrypted, "periodic")
		if err != nil {
			logger.LogError(ctx, "verification_scheduler", "Failed to trigger verification", err, map[string]interface{}{
				"api_key_id": key.ID,
				"provider":   key.Provider,
			})
			failedCount++
		} else {
			successCount++
		}
	}

	duration := time.Since(startTime)
	logger.LogInfo(ctx, "verification_scheduler", "Periodic verification completed", map[string]interface{}{
		"total_keys":    len(apiKeys),
		"success_count": successCount,
		"failed_count":  failedCount,
		"duration_ms":   duration.Milliseconds(),
	})
}

type APIKeyForVerification struct {
	ID              int
	Provider        string
	APIKeyEncrypted string
}

func (s *VerificationScheduler) getAPIKeysForVerification() ([]APIKeyForVerification, error) {
	if s.db == nil {
		s.db = config.GetDB()
	}
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	query := `
		SELECT id, provider, api_key_encrypted
		FROM merchant_api_keys
		WHERE status = 'active'
		  AND (
		    verification_result IS NULL
		    OR verification_result = 'pending'
		    OR verification_result = 'failed'
		    OR verified_at IS NULL
		    OR verified_at < NOW() - INTERVAL '24 hours'
		  )
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apiKeys []APIKeyForVerification
	for rows.Next() {
		var key APIKeyForVerification
		if err := rows.Scan(&key.ID, &key.Provider, &key.APIKeyEncrypted); err != nil {
			continue
		}
		apiKeys = append(apiKeys, key)
	}

	return apiKeys, nil
}

func (s *VerificationScheduler) TriggerVerification(apiKeyID int) error {
	if s.db == nil {
		s.db = config.GetDB()
	}
	if s.db == nil {
		return fmt.Errorf("database not available")
	}

	var key APIKeyForVerification
	err := s.db.QueryRow(
		"SELECT id, provider, api_key_encrypted FROM merchant_api_keys WHERE id = $1 AND status = 'active'",
		apiKeyID,
	).Scan(&key.ID, &key.Provider, &key.APIKeyEncrypted)

	if err != nil {
		return err
	}

	return s.validator.ValidateAsync(key.ID, key.Provider, key.APIKeyEncrypted, "manual")
}

func (s *VerificationScheduler) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}
