package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pintuotuo/backend/models"
)

type IRouteAwareness interface {
	GetRealtimeStatus(ctx context.Context, apiKeyID int) (*models.APIKeyRealtimeStatus, error)
	UpdateStatus(ctx context.Context, apiKeyID int, status *models.APIKeyRealtimeStatus) error
	GetBatchStatus(ctx context.Context, apiKeyIDs []int) ([]*models.APIKeyRealtimeStatus, error)
	UpdateLatency(ctx context.Context, apiKeyID int, latencyMs int) error
	UpdateErrorRate(ctx context.Context, apiKeyID int, errorRate float64) error
	UpdateConnectionPool(ctx context.Context, apiKeyID int, active int) error
	UpdateRateLimit(ctx context.Context, apiKeyID int, remaining int, resetAt *time.Time) error
}

type RouteAwarenessService struct {
	db *sql.DB
}

func NewRouteAwarenessService(db *sql.DB) *RouteAwarenessService {
	return &RouteAwarenessService{
		db: db,
	}
}

func (s *RouteAwarenessService) GetRealtimeStatus(ctx context.Context, apiKeyID int) (*models.APIKeyRealtimeStatus, error) {
	query := `
		SELECT api_key_id, latency_p50, latency_p95, latency_p99,
		       error_rate, success_rate, connection_pool_size, connection_pool_active,
		       rate_limit_remaining, rate_limit_reset_at, load_balance_weight,
		       last_request_at, updated_at
		FROM api_key_realtime_status
		WHERE api_key_id = $1
	`

	var status models.APIKeyRealtimeStatus
	err := s.db.QueryRowContext(ctx, query, apiKeyID).Scan(
		&status.APIKeyID,
		&status.LatencyP50,
		&status.LatencyP95,
		&status.LatencyP99,
		&status.ErrorRate,
		&status.SuccessRate,
		&status.ConnectionPoolSize,
		&status.ConnectionPoolActive,
		&status.RateLimitRemaining,
		&status.RateLimitResetAt,
		&status.LoadBalanceWeight,
		&status.LastRequestAt,
		&status.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return s.createDefaultStatus(ctx, apiKeyID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get realtime status: %w", err)
	}

	return &status, nil
}

func (s *RouteAwarenessService) createDefaultStatus(ctx context.Context, apiKeyID int) (*models.APIKeyRealtimeStatus, error) {
	status := &models.APIKeyRealtimeStatus{
		APIKeyID:             apiKeyID,
		LatencyP50:           0,
		LatencyP95:           0,
		LatencyP99:           0,
		ErrorRate:            0.0,
		SuccessRate:          1.0,
		ConnectionPoolSize:   10,
		ConnectionPoolActive: 0,
		RateLimitRemaining:   0,
		LoadBalanceWeight:    1.0,
		UpdatedAt:            time.Now(),
	}

	query := `
		INSERT INTO api_key_realtime_status (
			api_key_id, latency_p50, latency_p95, latency_p99,
			error_rate, success_rate, connection_pool_size, connection_pool_active,
			rate_limit_remaining, load_balance_weight, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (api_key_id) DO NOTHING
	`

	_, err := s.db.ExecContext(ctx, query,
		status.APIKeyID,
		status.LatencyP50,
		status.LatencyP95,
		status.LatencyP99,
		status.ErrorRate,
		status.SuccessRate,
		status.ConnectionPoolSize,
		status.ConnectionPoolActive,
		status.RateLimitRemaining,
		status.LoadBalanceWeight,
		status.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create default status: %w", err)
	}

	return status, nil
}

func (s *RouteAwarenessService) UpdateStatus(ctx context.Context, apiKeyID int, status *models.APIKeyRealtimeStatus) error {
	query := `
		INSERT INTO api_key_realtime_status (
			api_key_id, latency_p50, latency_p95, latency_p99,
			error_rate, success_rate, connection_pool_size, connection_pool_active,
			rate_limit_remaining, rate_limit_reset_at, load_balance_weight,
			last_request_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (api_key_id) DO UPDATE SET
			latency_p50 = EXCLUDED.latency_p50,
			latency_p95 = EXCLUDED.latency_p95,
			latency_p99 = EXCLUDED.latency_p99,
			error_rate = EXCLUDED.error_rate,
			success_rate = EXCLUDED.success_rate,
			connection_pool_size = EXCLUDED.connection_pool_size,
			connection_pool_active = EXCLUDED.connection_pool_active,
			rate_limit_remaining = EXCLUDED.rate_limit_remaining,
			rate_limit_reset_at = EXCLUDED.rate_limit_reset_at,
			load_balance_weight = EXCLUDED.load_balance_weight,
			last_request_at = EXCLUDED.last_request_at,
			updated_at = EXCLUDED.updated_at
	`

	status.UpdatedAt = time.Now()

	_, err := s.db.ExecContext(ctx, query,
		apiKeyID,
		status.LatencyP50,
		status.LatencyP95,
		status.LatencyP99,
		status.ErrorRate,
		status.SuccessRate,
		status.ConnectionPoolSize,
		status.ConnectionPoolActive,
		status.RateLimitRemaining,
		status.RateLimitResetAt,
		status.LoadBalanceWeight,
		status.LastRequestAt,
		status.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	return nil
}

func (s *RouteAwarenessService) GetBatchStatus(ctx context.Context, apiKeyIDs []int) ([]*models.APIKeyRealtimeStatus, error) {
	if len(apiKeyIDs) == 0 {
		return []*models.APIKeyRealtimeStatus{}, nil
	}

	query := `
		SELECT api_key_id, latency_p50, latency_p95, latency_p99,
		       error_rate, success_rate, connection_pool_size, connection_pool_active,
		       rate_limit_remaining, rate_limit_reset_at, load_balance_weight,
		       last_request_at, updated_at
		FROM api_key_realtime_status
		WHERE api_key_id = ANY($1)
	`

	idsJSON, _ := json.Marshal(apiKeyIDs)
	rows, err := s.db.QueryContext(ctx, query, idsJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to get batch status: %w", err)
	}
	defer rows.Close()

	var statuses []*models.APIKeyRealtimeStatus
	for rows.Next() {
		var status models.APIKeyRealtimeStatus
		err := rows.Scan(
			&status.APIKeyID,
			&status.LatencyP50,
			&status.LatencyP95,
			&status.LatencyP99,
			&status.ErrorRate,
			&status.SuccessRate,
			&status.ConnectionPoolSize,
			&status.ConnectionPoolActive,
			&status.RateLimitRemaining,
			&status.RateLimitResetAt,
			&status.LoadBalanceWeight,
			&status.LastRequestAt,
			&status.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan status: %w", err)
		}
		statuses = append(statuses, &status)
	}

	return statuses, nil
}

func (s *RouteAwarenessService) UpdateLatency(ctx context.Context, apiKeyID int, latencyMs int) error {
	query := `
		INSERT INTO api_key_realtime_status (api_key_id, latency_p50, latency_p95, latency_p99, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (api_key_id) DO UPDATE SET
			latency_p50 = EXCLUDED.latency_p50,
			latency_p95 = EXCLUDED.latency_p95,
			latency_p99 = EXCLUDED.latency_p99,
			updated_at = EXCLUDED.updated_at
	`

	now := time.Now()
	_, err := s.db.ExecContext(ctx, query, apiKeyID, latencyMs, latencyMs, latencyMs, now)
	if err != nil {
		return fmt.Errorf("failed to update latency: %w", err)
	}

	return nil
}

func (s *RouteAwarenessService) UpdateErrorRate(ctx context.Context, apiKeyID int, errorRate float64) error {
	query := `
		INSERT INTO api_key_realtime_status (api_key_id, error_rate, success_rate, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (api_key_id) DO UPDATE SET
			error_rate = EXCLUDED.error_rate,
			success_rate = EXCLUDED.success_rate,
			updated_at = EXCLUDED.updated_at
	`

	now := time.Now()
	successRate := 1.0 - errorRate
	_, err := s.db.ExecContext(ctx, query, apiKeyID, errorRate, successRate, now)
	if err != nil {
		return fmt.Errorf("failed to update error rate: %w", err)
	}

	return nil
}

func (s *RouteAwarenessService) UpdateConnectionPool(ctx context.Context, apiKeyID int, active int) error {
	query := `
		INSERT INTO api_key_realtime_status (api_key_id, connection_pool_active, updated_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (api_key_id) DO UPDATE SET
			connection_pool_active = EXCLUDED.connection_pool_active,
			updated_at = EXCLUDED.updated_at
	`

	now := time.Now()
	_, err := s.db.ExecContext(ctx, query, apiKeyID, active, now)
	if err != nil {
		return fmt.Errorf("failed to update connection pool: %w", err)
	}

	return nil
}

func (s *RouteAwarenessService) UpdateRateLimit(ctx context.Context, apiKeyID int, remaining int, resetAt *time.Time) error {
	query := `
		INSERT INTO api_key_realtime_status (api_key_id, rate_limit_remaining, rate_limit_reset_at, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (api_key_id) DO UPDATE SET
			rate_limit_remaining = EXCLUDED.rate_limit_remaining,
			rate_limit_reset_at = EXCLUDED.rate_limit_reset_at,
			updated_at = EXCLUDED.updated_at
	`

	now := time.Now()
	_, err := s.db.ExecContext(ctx, query, apiKeyID, remaining, resetAt, now)
	if err != nil {
		return fmt.Errorf("failed to update rate limit: %w", err)
	}

	return nil
}
