package services

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
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

func afterRoutingStrategiesChange() {
	GetSmartRouter().ReloadRoutingStrategies()
}

// mergeRoutingStrategyUpdate overlays patch onto existing for partial JSON bodies (e.g. name-only tests).
func mergeRoutingStrategyUpdate(existing *RoutingStrategyConfig, patch *RoutingStrategyConfig) RoutingStrategyConfig {
	out := *existing
	if patch.Name != "" {
		out.Name = patch.Name
	}
	if patch.Code != "" {
		out.Code = patch.Code
	}
	if patch.Description != nil {
		out.Description = patch.Description
	}
	wSum := patch.PriceWeight + patch.LatencyWeight + patch.ReliabilityWeight
	if wSum > 1e-9 {
		out.PriceWeight = patch.PriceWeight
		out.LatencyWeight = patch.LatencyWeight
		out.ReliabilityWeight = patch.ReliabilityWeight
		out.MaxRetryCount = patch.MaxRetryCount
		out.RetryBackoffBase = patch.RetryBackoffBase
		out.CircuitBreakerThreshold = patch.CircuitBreakerThreshold
		out.CircuitBreakerTimeout = patch.CircuitBreakerTimeout
	} else {
		if patch.MaxRetryCount != 0 || patch.RetryBackoffBase != 0 {
			out.MaxRetryCount = patch.MaxRetryCount
			out.RetryBackoffBase = patch.RetryBackoffBase
		}
		if patch.CircuitBreakerThreshold != 0 || patch.CircuitBreakerTimeout != 0 {
			out.CircuitBreakerThreshold = patch.CircuitBreakerThreshold
			out.CircuitBreakerTimeout = patch.CircuitBreakerTimeout
		}
	}
	out.IsDefault = patch.IsDefault
	if patch.Status != "" {
		out.Status = strings.TrimSpace(patch.Status)
	}
	return out
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

func getStrategyByIDTx(tx *sql.Tx, id int) (*RoutingStrategyConfig, error) {
	query := `
		SELECT id, name, code, description, price_weight, latency_weight, reliability_weight,
		       max_retry_count, retry_backoff_base, circuit_breaker_threshold, circuit_breaker_timeout,
		       is_default, status, created_at, updated_at
		FROM routing_strategies
		WHERE id = $1 AND status != 'deleted'
	`
	var strategy RoutingStrategyConfig
	err := tx.QueryRow(query, id).Scan(
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
		return nil, err
	}
	return &strategy, nil
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

func applyRoutingStrategyBusinessRules(merged *RoutingStrategyConfig) {
	if merged.IsDefault {
		merged.Status = "active"
		return
	}
	if merged.Status == "inactive" {
		merged.IsDefault = false
	}
}

func (s *RoutingStrategyService) CreateStrategy(strategy *RoutingStrategyConfig) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	if strategy.Status == "" {
		strategy.Status = "active"
	}
	strategy.Status = strings.TrimSpace(strategy.Status)
	if strategy.IsDefault {
		if _, err = tx.Exec(`UPDATE routing_strategies SET is_default = false WHERE status != 'deleted'`); err != nil {
			return fmt.Errorf("failed to clear default flags: %w", err)
		}
		strategy.Status = "active"
	}
	if strategy.Status == "inactive" {
		strategy.IsDefault = false
	}

	query := `
		INSERT INTO routing_strategies (name, code, description, price_weight, latency_weight, reliability_weight,
		                                max_retry_count, retry_backoff_base, circuit_breaker_threshold, circuit_breaker_timeout,
		                                is_default, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at
	`

	err = tx.QueryRow(
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
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	afterRoutingStrategiesChange()
	return nil
}

func (s *RoutingStrategyService) UpdateStrategy(id int, patch *RoutingStrategyConfig) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	existing, err := getStrategyByIDTx(tx, id)
	if err != nil {
		return err
	}

	merged := mergeRoutingStrategyUpdate(existing, patch)
	applyRoutingStrategyBusinessRules(&merged)

	if merged.IsDefault {
		if _, err = tx.Exec(`UPDATE routing_strategies SET is_default = false WHERE status != 'deleted' AND id != $1`, id); err != nil {
			return fmt.Errorf("failed to clear other defaults: %w", err)
		}
	}

	query := `
		UPDATE routing_strategies
		SET name = $1, description = $2, price_weight = $3, latency_weight = $4, reliability_weight = $5,
		    max_retry_count = $6, retry_backoff_base = $7, circuit_breaker_threshold = $8, circuit_breaker_timeout = $9,
		    is_default = $10, status = $11, updated_at = CURRENT_TIMESTAMP
		WHERE id = $12 AND status != 'deleted'
		RETURNING updated_at
	`

	err = tx.QueryRow(
		query,
		merged.Name, merged.Description,
		merged.PriceWeight, merged.LatencyWeight, merged.ReliabilityWeight,
		merged.MaxRetryCount, merged.RetryBackoffBase,
		merged.CircuitBreakerThreshold, merged.CircuitBreakerTimeout,
		merged.IsDefault, merged.Status, id,
	).Scan(&merged.UpdatedAt)

	if err == sql.ErrNoRows {
		return errors.New("strategy not found")
	}
	if err != nil {
		return fmt.Errorf("failed to update strategy: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	afterRoutingStrategiesChange()
	return nil
}

func (s *RoutingStrategyService) DeleteStrategy(id int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	query := `UPDATE routing_strategies SET status = 'deleted', updated_at = CURRENT_TIMESTAMP WHERE id = $1 AND status != 'deleted'`

	result, err := tx.Exec(query, id)
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

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	afterRoutingStrategiesChange()
	return nil
}
