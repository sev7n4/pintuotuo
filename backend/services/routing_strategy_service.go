package services

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type RoutingStrategyConfig struct {
	ID                      int       `json:"id"`
	Name                    string    `json:"name"`
	Code                    string    `json:"code"`
	Description             *string   `json:"description"`
	PriceWeight             float64   `json:"price_weight"`
	LatencyWeight           float64   `json:"latency_weight"`
	ReliabilityWeight       float64   `json:"reliability_weight"`
	MaxRetryCount           int       `json:"max_retry_count"`
	RetryBackoffBase        int       `json:"retry_backoff_base"`
	CircuitBreakerThreshold int       `json:"circuit_breaker_threshold"`
	CircuitBreakerTimeout   int       `json:"circuit_breaker_timeout"`
	IsDefault               bool      `json:"is_default"`
	Status                  string    `json:"status"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

type RoutingStrategyService struct {
	db *sql.DB
}

func NewRoutingStrategyService(db *sql.DB) *RoutingStrategyService {
	return &RoutingStrategyService{db: db}
}

func (s *RoutingStrategyService) GetStrategies(page, pageSize int) ([]RoutingStrategyConfig, int, error) {
	offset := (page - 1) * pageSize

	query := `
		SELECT id, name, code, description, price_weight, latency_weight, reliability_weight,
		       max_retry_count, retry_backoff_base, circuit_breaker_threshold, circuit_breaker_timeout,
		       is_default, status, created_at, updated_at
		FROM routing_strategies
		WHERE status != 'deleted'
		ORDER BY is_default DESC, created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := s.db.Query(query, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query strategies: %w", err)
	}
	defer rows.Close()

	var strategies []RoutingStrategyConfig
	for rows.Next() {
		var strategy RoutingStrategyConfig
		if err = rows.Scan(
			&strategy.ID, &strategy.Name, &strategy.Code, &strategy.Description,
			&strategy.PriceWeight, &strategy.LatencyWeight, &strategy.ReliabilityWeight,
			&strategy.MaxRetryCount, &strategy.RetryBackoffBase,
			&strategy.CircuitBreakerThreshold, &strategy.CircuitBreakerTimeout,
			&strategy.IsDefault, &strategy.Status, &strategy.CreatedAt, &strategy.UpdatedAt,
		); err != nil {
			continue
		}
		strategies = append(strategies, strategy)
	}

	var total int
	err = s.db.QueryRow("SELECT COUNT(*) FROM routing_strategies WHERE status != 'deleted'").Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count strategies: %w", err)
	}

	return strategies, total, nil
}

func (s *RoutingStrategyService) GetStrategyByID(id int) (*RoutingStrategyConfig, error) {
	query := `
		SELECT id, name, code, description, price_weight, latency_weight, reliability_weight,
		       max_retry_count, retry_backoff_base, circuit_breaker_threshold, circuit_breaker_timeout,
		       is_default, status, created_at, updated_at
		FROM routing_strategies
		WHERE id = $1 AND status != 'deleted'
	`

	var strategy RoutingStrategyConfig
	err := s.db.QueryRow(query, id).Scan(
		&strategy.ID, &strategy.Name, &strategy.Code, &strategy.Description,
		&strategy.PriceWeight, &strategy.LatencyWeight, &strategy.ReliabilityWeight,
		&strategy.MaxRetryCount, &strategy.RetryBackoffBase,
		&strategy.CircuitBreakerThreshold, &strategy.CircuitBreakerTimeout,
		&strategy.IsDefault, &strategy.Status, &strategy.CreatedAt, &strategy.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("strategy not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query strategy: %w", err)
	}

	return &strategy, nil
}

func (s *RoutingStrategyService) CreateStrategy(strategy *RoutingStrategyConfig) error {
	query := `
		INSERT INTO routing_strategies (name, code, description, price_weight, latency_weight, reliability_weight,
		                                max_retry_count, retry_backoff_base, circuit_breaker_threshold, circuit_breaker_timeout,
		                                is_default, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at
	`

	err := s.db.QueryRow(
		query,
		strategy.Name, strategy.Code, strategy.Description,
		strategy.PriceWeight, strategy.LatencyWeight, strategy.ReliabilityWeight,
		strategy.MaxRetryCount, strategy.RetryBackoffBase,
		strategy.CircuitBreakerThreshold, strategy.CircuitBreakerTimeout,
		strategy.IsDefault, strategy.Status,
	).Scan(&strategy.ID, &strategy.CreatedAt, &strategy.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create strategy: %w", err)
	}

	return nil
}

func (s *RoutingStrategyService) UpdateStrategy(id int, strategy *RoutingStrategyConfig) error {
	query := `
		UPDATE routing_strategies
		SET name = $1, description = $2, price_weight = $3, latency_weight = $4, reliability_weight = $5,
		    max_retry_count = $6, retry_backoff_base = $7, circuit_breaker_threshold = $8, circuit_breaker_timeout = $9,
		    is_default = $10, status = $11, updated_at = CURRENT_TIMESTAMP
		WHERE id = $12 AND status != 'deleted'
		RETURNING updated_at
	`

	err := s.db.QueryRow(
		query,
		strategy.Name, strategy.Description,
		strategy.PriceWeight, strategy.LatencyWeight, strategy.ReliabilityWeight,
		strategy.MaxRetryCount, strategy.RetryBackoffBase,
		strategy.CircuitBreakerThreshold, strategy.CircuitBreakerTimeout,
		strategy.IsDefault, strategy.Status, id,
	).Scan(&strategy.UpdatedAt)

	if err == sql.ErrNoRows {
		return errors.New("strategy not found")
	}
	if err != nil {
		return fmt.Errorf("failed to update strategy: %w", err)
	}

	strategy.ID = id
	return nil
}

func (s *RoutingStrategyService) DeleteStrategy(id int) error {
	query := `UPDATE routing_strategies SET status = 'deleted', updated_at = CURRENT_TIMESTAMP WHERE id = $1 AND status != 'deleted'`

	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete strategy: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("strategy not found")
	}

	return nil
}
