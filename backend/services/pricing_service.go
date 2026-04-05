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
				'USD' as currency,
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
		{Provider: "openai", Model: "gpt-4-turbo-preview", InputPrice: 10, OutputPrice: 30, Currency: "USD"},
		{Provider: "openai", Model: "gpt-4", InputPrice: 30, OutputPrice: 60, Currency: "USD"},
		{Provider: "openai", Model: "gpt-4o", InputPrice: 5, OutputPrice: 15, Currency: "USD"},
		{Provider: "openai", Model: "gpt-4o-mini", InputPrice: 0.15, OutputPrice: 0.6, Currency: "USD"},
		{Provider: "openai", Model: "gpt-3.5-turbo", InputPrice: 0.5, OutputPrice: 1.5, Currency: "USD"},
		{Provider: "anthropic", Model: "claude-3-opus-20240229", InputPrice: 15, OutputPrice: 75, Currency: "USD"},
		{Provider: "anthropic", Model: "claude-3-sonnet-20240229", InputPrice: 3, OutputPrice: 15, Currency: "USD"},
		{Provider: "anthropic", Model: "claude-3-haiku-20240307", InputPrice: 0.25, OutputPrice: 1.25, Currency: "USD"},
		{Provider: "anthropic", Model: "claude-3-5-sonnet-20241022", InputPrice: 3, OutputPrice: 15, Currency: "USD"},
		{Provider: "google", Model: "gemini-pro", InputPrice: 0.5, OutputPrice: 1.5, Currency: "USD"},
		{Provider: "google", Model: "gemini-1.5-pro", InputPrice: 3.5, OutputPrice: 10.5, Currency: "USD"},
		{Provider: "google", Model: "gemini-1.5-flash", InputPrice: 0.075, OutputPrice: 0.3, Currency: "USD"},
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
			InputPrice:  1,
			OutputPrice: 2,
		}
	}

	inputCost := float64(inputTokens) * pricing.InputPrice / 1_000_000
	outputCost := float64(outputTokens) * pricing.OutputPrice / 1_000_000
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
