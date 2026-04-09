package services

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/logger"
	"github.com/pintuotuo/backend/metrics"
)

type PricingData struct {
	Provider    string    `json:"provider"`
	Model       string    `json:"model"`
	InputPrice  float64   `json:"input_price"`
	OutputPrice float64   `json:"output_price"`
	Currency    string    `json:"currency"`
	EffectiveAt time.Time `json:"effective_at"`
}

type PricingHistory struct {
	ID             int       `json:"id"`
	EntityType     string    `json:"entity_type"`
	EntityID       int       `json:"entity_id"`
	OldInputPrice  float64   `json:"old_input_price"`
	OldOutputPrice float64   `json:"old_output_price"`
	NewInputPrice  float64   `json:"new_input_price"`
	NewOutputPrice float64   `json:"new_output_price"`
	ChangeReason   string    `json:"change_reason"`
	ChangedBy      int       `json:"changed_by"`
	ChangedAt      time.Time `json:"changed_at"`
	EffectiveAt    time.Time `json:"effective_at"`
}

type PricingSchedule struct {
	ID             int       `json:"id"`
	EntityType     string    `json:"entity_type"`
	EntityID       int       `json:"entity_id"`
	NewInputPrice  float64   `json:"new_input_price"`
	NewOutputPrice float64   `json:"new_output_price"`
	ScheduledAt    time.Time `json:"scheduled_at"`
	Status         string    `json:"status"`
	ChangeReason   string    `json:"change_reason"`
	CreatedBy      int       `json:"created_by"`
}

type PricingService struct {
	db           *sql.DB
	pricingCache map[string]PricingData
	cacheMutex   sync.RWMutex
	lastLoadTime time.Time
	cacheTTL     time.Duration
}

var (
	pricingService     *PricingService
	pricingServiceOnce sync.Once
)

func GetPricingService() *PricingService {
	pricingServiceOnce.Do(func() {
		pricingService = &PricingService{
			db:           config.GetDB(),
			pricingCache: make(map[string]PricingData),
			cacheTTL:     5 * time.Minute,
		}
		pricingService.loadPricing()
	})
	return pricingService
}

func (s *PricingService) loadPricing() {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	startTime := time.Now()
	source := "database"

	if s.db == nil {
		s.db = config.GetDB()
	}

	if s.db != nil {
		query := `
			SELECT 
				model_provider as provider,
				model_name as model,
				provider_input_rate as input_price,
				provider_output_rate as output_price,
				'CNY' as currency,
				COALESCE(updated_at, created_at) as effective_at
			FROM spus
			WHERE status = 'active'
			  AND provider_input_rate IS NOT NULL
			  AND provider_output_rate IS NOT NULL
		`

		rows, err := s.db.Query(query)
		if err != nil {
			logger.LogError(context.Background(), "pricing_service", "Failed to query pricing from database", err, map[string]interface{}{
				"query": query,
			})
		} else {
			defer rows.Close()
			for rows.Next() {
				var p PricingData
				err := rows.Scan(&p.Provider, &p.Model, &p.InputPrice, &p.OutputPrice, &p.Currency, &p.EffectiveAt)
				if err == nil {
					key := fmt.Sprintf("%s:%s", p.Provider, p.Model)
					s.pricingCache[key] = p
				}
			}
		}
	}

	if len(s.pricingCache) == 0 {
		s.loadDefaultPricing()
		source = "default"
	}

	s.lastLoadTime = time.Now()
	cacheSize := len(s.pricingCache)

	metrics.RecordPricingCacheReload(source)
	metrics.SetPricingCacheSize(cacheSize)

	logger.LogInfo(context.Background(), "pricing_service", "Pricing cache reloaded", map[string]interface{}{
		"source":       source,
		"items_loaded": cacheSize,
		"duration_ms":  time.Since(startTime).Milliseconds(),
	})
}

func (s *PricingService) loadDefaultPricing() {
	defaultPricing := []PricingData{
		{Provider: "openai", Model: "gpt-4-turbo-preview", InputPrice: 0.01, OutputPrice: 0.03, Currency: "CNY"},
		{Provider: "openai", Model: "gpt-4", InputPrice: 0.03, OutputPrice: 0.06, Currency: "CNY"},
		{Provider: "openai", Model: "gpt-4o", InputPrice: 0.005, OutputPrice: 0.015, Currency: "CNY"},
		{Provider: "openai", Model: "gpt-4o-mini", InputPrice: 0.00015, OutputPrice: 0.0006, Currency: "CNY"},
		{Provider: "openai", Model: "gpt-3.5-turbo", InputPrice: 0.0005, OutputPrice: 0.0015, Currency: "CNY"},
		{Provider: "anthropic", Model: "claude-3-opus-20240229", InputPrice: 0.015, OutputPrice: 0.075, Currency: "CNY"},
		{Provider: "anthropic", Model: "claude-3-sonnet-20240229", InputPrice: 0.003, OutputPrice: 0.015, Currency: "CNY"},
		{Provider: "anthropic", Model: "claude-3-haiku-20240307", InputPrice: 0.00025, OutputPrice: 0.00125, Currency: "CNY"},
		{Provider: "anthropic", Model: "claude-3-5-sonnet-20241022", InputPrice: 0.003, OutputPrice: 0.015, Currency: "CNY"},
		{Provider: "google", Model: "gemini-pro", InputPrice: 0.0005, OutputPrice: 0.0015, Currency: "CNY"},
		{Provider: "google", Model: "gemini-1.5-pro", InputPrice: 0.0035, OutputPrice: 0.0105, Currency: "CNY"},
		{Provider: "google", Model: "gemini-1.5-flash", InputPrice: 0.000075, OutputPrice: 0.0003, Currency: "CNY"},
	}

	for _, p := range defaultPricing {
		key := fmt.Sprintf("%s:%s", p.Provider, p.Model)
		s.pricingCache[key] = p
	}
}

func (s *PricingService) GetPricing(provider, model string) (PricingData, bool) {
	s.cacheMutex.RLock()
	if time.Since(s.lastLoadTime) > s.cacheTTL {
		s.cacheMutex.RUnlock()
		s.loadPricing()
		s.cacheMutex.RLock()
	}
	defer s.cacheMutex.RUnlock()

	key := fmt.Sprintf("%s:%s", provider, model)
	pricing, ok := s.pricingCache[key]

	if ok {
		metrics.RecordPricingCacheHit(provider, model)
		logger.LogDebug(context.Background(), "pricing_service", "Pricing cache hit", map[string]interface{}{
			"provider": provider,
			"model":    model,
		})
	} else {
		metrics.RecordPricingCacheMiss(provider, model)
		logger.LogWarn(context.Background(), "pricing_service", "Pricing cache miss", map[string]interface{}{
			"provider": provider,
			"model":    model,
		})
	}

	return pricing, ok
}

func (s *PricingService) CalculateCost(provider, model string, inputTokens, outputTokens int) float64 {
	startTime := time.Now()

	pricing, ok := s.GetPricing(provider, model)
	if !ok {
		pricing = PricingData{
			InputPrice:  0.001,
			OutputPrice: 0.002,
		}
	}

	// 统一计费口径：价格字段为 元/1K tokens，按 token 精算。
	inputCost := float64(inputTokens) * pricing.InputPrice / 1_000
	outputCost := float64(outputTokens) * pricing.OutputPrice / 1_000
	totalCost := inputCost + outputCost

	duration := time.Since(startTime).Seconds()
	metrics.RecordPricingCalculation(provider, model, duration)

	logger.LogDebug(context.Background(), "pricing_service", "Pricing calculation completed", map[string]interface{}{
		"provider":      provider,
		"model":         model,
		"input_tokens":  inputTokens,
		"output_tokens": outputTokens,
		"total_cost":    totalCost,
		"duration_us":   duration * 1_000_000,
	})

	return totalCost
}

func (s *PricingService) RecordPricingHistory(entityType string, entityID int, oldInputPrice, oldOutputPrice, newInputPrice, newOutputPrice float64, changeReason string, changedBy int) error {
	if s.db == nil {
		s.db = config.GetDB()
	}
	if s.db == nil {
		return fmt.Errorf("database not available")
	}

	query := `
		INSERT INTO pricing_history (
			entity_type, entity_id,
			old_input_price, old_output_price,
			new_input_price, new_output_price,
			change_reason, changed_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := s.db.Exec(query,
		entityType, entityID,
		oldInputPrice, oldOutputPrice,
		newInputPrice, newOutputPrice,
		changeReason, changedBy,
	)

	if err != nil {
		logger.LogError(context.Background(), "pricing_service", "Failed to record pricing history", err, map[string]interface{}{
			"entity_type": entityType,
			"entity_id":   entityID,
		})
		return err
	}

	logger.LogInfo(context.Background(), "pricing_service", "Pricing history recorded", map[string]interface{}{
		"entity_type":     entityType,
		"entity_id":       entityID,
		"old_input_price": oldInputPrice,
		"new_input_price": newInputPrice,
		"changed_by":      changedBy,
	})

	return nil
}

func (s *PricingService) GetPricingHistory(entityType string, entityID int, limit int) ([]PricingHistory, error) {
	if s.db == nil {
		s.db = config.GetDB()
	}
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT id, entity_type, entity_id,
			   old_input_price, old_output_price,
			   new_input_price, new_output_price,
			   change_reason, changed_by, changed_at, effective_at
		FROM pricing_history
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY changed_at DESC
		LIMIT $3
	`

	rows, err := s.db.Query(query, entityType, entityID, limit)
	if err != nil {
		logger.LogError(context.Background(), "pricing_service", "Failed to query pricing history", err, map[string]interface{}{
			"entity_type": entityType,
			"entity_id":   entityID,
		})
		return nil, err
	}
	defer rows.Close()

	var history []PricingHistory
	for rows.Next() {
		var h PricingHistory
		err := rows.Scan(
			&h.ID, &h.EntityType, &h.EntityID,
			&h.OldInputPrice, &h.OldOutputPrice,
			&h.NewInputPrice, &h.NewOutputPrice,
			&h.ChangeReason, &h.ChangedBy, &h.ChangedAt, &h.EffectiveAt,
		)
		if err != nil {
			continue
		}
		history = append(history, h)
	}

	return history, nil
}

func (s *PricingService) SchedulePricingChange(entityType string, entityID int, newInputPrice, newOutputPrice float64, scheduledAt time.Time, changeReason string, createdBy int) (int, error) {
	if s.db == nil {
		s.db = config.GetDB()
	}
	if s.db == nil {
		return 0, fmt.Errorf("database not available")
	}

	query := `
		INSERT INTO pricing_schedules (
			entity_type, entity_id,
			new_input_price, new_output_price,
			scheduled_at, change_reason, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	var scheduleID int
	err := s.db.QueryRow(query,
		entityType, entityID,
		newInputPrice, newOutputPrice,
		scheduledAt, changeReason, createdBy,
	).Scan(&scheduleID)

	if err != nil {
		logger.LogError(context.Background(), "pricing_service", "Failed to schedule pricing change", err, map[string]interface{}{
			"entity_type":  entityType,
			"entity_id":    entityID,
			"scheduled_at": scheduledAt,
		})
		return 0, err
	}

	logger.LogInfo(context.Background(), "pricing_service", "Pricing change scheduled", map[string]interface{}{
		"schedule_id":  scheduleID,
		"entity_type":  entityType,
		"entity_id":    entityID,
		"scheduled_at": scheduledAt,
		"created_by":   createdBy,
	})

	return scheduleID, nil
}

func (s *PricingService) ProcessScheduledPricing() (int, error) {
	if s.db == nil {
		s.db = config.GetDB()
	}
	if s.db == nil {
		return 0, fmt.Errorf("database not available")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	query := `
		SELECT id, entity_type, entity_id, new_input_price, new_output_price, change_reason, created_by
		FROM pricing_schedules
		WHERE status = 'pending' AND scheduled_at <= NOW()
		FOR UPDATE SKIP LOCKED
	`

	rows, err := tx.Query(query)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	processedCount := 0
	for rows.Next() {
		var schedule PricingSchedule
		if scanErr := rows.Scan(
			&schedule.ID, &schedule.EntityType, &schedule.EntityID,
			&schedule.NewInputPrice, &schedule.NewOutputPrice,
			&schedule.ChangeReason, &schedule.CreatedBy,
		); scanErr != nil {
			continue
		}

		var oldInputPrice, oldOutputPrice float64

		switch schedule.EntityType {
		case "spu":
			err = tx.QueryRow(`
				SELECT COALESCE(provider_input_rate, 0), COALESCE(provider_output_rate, 0)
				FROM spus WHERE id = $1
			`, schedule.EntityID).Scan(&oldInputPrice, &oldOutputPrice)

			if err != nil {
				continue
			}

			_, err = tx.Exec(`
				UPDATE spus 
				SET provider_input_rate = $1, provider_output_rate = $2, updated_at = NOW()
				WHERE id = $3
			`, schedule.NewInputPrice, schedule.NewOutputPrice, schedule.EntityID)

			if err != nil {
				continue
			}
		}

		_, err = tx.Exec(`
			INSERT INTO pricing_history (
				entity_type, entity_id,
				old_input_price, old_output_price,
				new_input_price, new_output_price,
				change_reason, changed_by
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, schedule.EntityType, schedule.EntityID,
			oldInputPrice, oldOutputPrice,
			schedule.NewInputPrice, schedule.NewOutputPrice,
			schedule.ChangeReason, schedule.CreatedBy)

		if err != nil {
			continue
		}

		_, err = tx.Exec(`
			UPDATE pricing_schedules 
			SET status = 'executed', executed_at = NOW(), executed_by = $1
			WHERE id = $2
		`, schedule.CreatedBy, schedule.ID)

		if err != nil {
			continue
		}

		processedCount++

		logger.LogInfo(context.Background(), "pricing_service", "Scheduled pricing executed", map[string]interface{}{
			"schedule_id":     schedule.ID,
			"entity_type":     schedule.EntityType,
			"entity_id":       schedule.EntityID,
			"old_input_price": oldInputPrice,
			"new_input_price": schedule.NewInputPrice,
		})
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	if processedCount > 0 {
		s.loadPricing()
	}

	return processedCount, nil
}

func (s *PricingService) GetPendingSchedules() ([]PricingSchedule, error) {
	if s.db == nil {
		s.db = config.GetDB()
	}
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	query := `
		SELECT id, entity_type, entity_id, new_input_price, new_output_price,
			   scheduled_at, status, change_reason, created_by
		FROM pricing_schedules
		WHERE status = 'pending'
		ORDER BY scheduled_at ASC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []PricingSchedule
	for rows.Next() {
		var ps PricingSchedule
		err := rows.Scan(
			&ps.ID, &ps.EntityType, &ps.EntityID,
			&ps.NewInputPrice, &ps.NewOutputPrice,
			&ps.ScheduledAt, &ps.Status, &ps.ChangeReason, &ps.CreatedBy,
		)
		if err != nil {
			continue
		}
		schedules = append(schedules, ps)
	}

	return schedules, nil
}
