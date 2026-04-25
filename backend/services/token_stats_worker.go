package services

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/pintuotuo/backend/logger"
)

type TokenStatsWorker struct {
	db       *sql.DB
	stopChan chan struct{}
	running  bool
	mutex    sync.Mutex
}

func NewTokenStatsWorker(db *sql.DB) *TokenStatsWorker {
	return &TokenStatsWorker{
		db:       db,
		stopChan: make(chan struct{}, 1),
		running:  false,
	}
}

func (w *TokenStatsWorker) UpdateStatistics(ctx context.Context) error {
	logger.LogInfo(ctx, "token_stats_worker", "Starting token statistics update", nil)

	query := `
		INSERT INTO model_token_statistics (
			model_name,
			avg_input_tokens,
			avg_output_tokens,
			p50_input_tokens,
			p50_output_tokens,
			p90_input_tokens,
			p90_output_tokens,
			input_output_ratio,
			total_requests,
			sample_start_date,
			sample_end_date,
			created_at,
			updated_at
		)
		SELECT 
			model,
			COALESCE(AVG(input_tokens), 0),
			COALESCE(AVG(output_tokens), 0),
			COALESCE(PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY input_tokens), 0)::INT,
			COALESCE(PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY output_tokens), 0)::INT,
			COALESCE(PERCENTILE_CONT(0.9) WITHIN GROUP (ORDER BY input_tokens), 0)::INT,
			COALESCE(PERCENTILE_CONT(0.9) WITHIN GROUP (ORDER BY output_tokens), 0)::INT,
			CASE 
				WHEN AVG(output_tokens) > 0 THEN AVG(input_tokens::FLOAT / NULLIF(output_tokens, 0))
				ELSE 1.0 
			END,
			COUNT(*),
			CURRENT_DATE - INTERVAL '7 days',
			CURRENT_DATE,
			NOW(),
			NOW()
		FROM api_usage_logs
		WHERE created_at > NOW() - INTERVAL '7 days'
		GROUP BY model
		ON CONFLICT (model_name) DO UPDATE SET
			avg_input_tokens = EXCLUDED.avg_input_tokens,
			avg_output_tokens = EXCLUDED.avg_output_tokens,
			p50_input_tokens = EXCLUDED.p50_input_tokens,
			p50_output_tokens = EXCLUDED.p50_output_tokens,
			p90_input_tokens = EXCLUDED.p90_input_tokens,
			p90_output_tokens = EXCLUDED.p90_output_tokens,
			input_output_ratio = EXCLUDED.input_output_ratio,
			total_requests = EXCLUDED.total_requests,
			sample_start_date = EXCLUDED.sample_start_date,
			sample_end_date = EXCLUDED.sample_end_date,
			updated_at = NOW()
	`

	result, err := w.db.ExecContext(ctx, query)
	if err != nil {
		logger.LogError(ctx, "token_stats_worker", "Failed to update token statistics", err, nil)
		return fmt.Errorf("failed to update token statistics: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	logger.LogInfo(ctx, "token_stats_worker", "Token statistics update completed", map[string]interface{}{
		"rows_affected": rowsAffected,
	})

	return nil
}

func (w *TokenStatsWorker) Run() {
	ctx := context.Background()

	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	durationUntilMidnight := next.Sub(now)

	logger.LogInfo(ctx, "token_stats_worker", "Scheduling first run at midnight", map[string]interface{}{
		"duration_until_midnight": durationUntilMidnight.String(),
		"next_run":                next.Format("2006-01-02 15:04:05"),
	})

	select {
	case <-w.stopChan:
		return
	case <-time.After(durationUntilMidnight):
	}

	if err := w.UpdateStatistics(ctx); err != nil {
		logger.LogError(ctx, "token_stats_worker", "Initial statistics update failed", err, nil)
	}

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopChan:
			logger.LogInfo(ctx, "token_stats_worker", "Token stats worker stopped", nil)
			return
		case <-ticker.C:
			if err := w.UpdateStatistics(ctx); err != nil {
				logger.LogError(ctx, "token_stats_worker", "Scheduled statistics update failed", err, nil)
			}
		}
	}
}

func (w *TokenStatsWorker) Start() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.running {
		return
	}

	w.running = true
	go w.Run()

	logger.LogInfo(context.Background(), "token_stats_worker", "Token stats worker started", nil)
}

func (w *TokenStatsWorker) Stop() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if !w.running {
		return
	}

	w.running = false
	w.stopChan <- struct{}{}

	logger.LogInfo(context.Background(), "token_stats_worker", "Token stats worker stop signal sent", nil)
}

func (w *TokenStatsWorker) IsRunning() bool {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.running
}

func (w *TokenStatsWorker) TriggerImmediateUpdate() error {
	ctx := context.Background()
	return w.UpdateStatistics(ctx)
}
