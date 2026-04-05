package services

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/pintuotuo/backend/config"
)

type RoutingStrategy string

const (
	RoutingStrategyPrice    RoutingStrategy = "price_first"
	RoutingStrategyLatency  RoutingStrategy = "latency_first"
	RoutingStrategyBalanced RoutingStrategy = "balanced"
	RoutingStrategyCost     RoutingStrategy = "cost_first"
)

type RoutingCandidate struct {
	APIKeyID     int
	Provider     string
	Model        string
	Score        float64
	PriceScore   float64
	LatencyScore float64
	SuccessScore float64
	HealthStatus string
	Verified     bool
	InputPrice   float64
	OutputPrice  float64
	AvgLatencyMs int
	SuccessRate  float64
}

type SmartRouter struct {
	db             *sql.DB
	circuitBreaker map[int]*CircuitBreaker
	cbMutex        sync.RWMutex
}

var (
	router     *SmartRouter
	routerOnce sync.Once
)

func GetSmartRouter() *SmartRouter {
	routerOnce.Do(func() {
		router = &SmartRouter{
			db:             config.GetDB(),
			circuitBreaker: make(map[int]*CircuitBreaker),
		}
	})
	return router
}

func (r *SmartRouter) SelectProvider(ctx context.Context, model string, strategy RoutingStrategy) (*RoutingCandidate, error) {
	candidates, err := r.GetCandidates(ctx, model)
	if err != nil {
		return nil, fmt.Errorf("failed to get candidates: %w", err)
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no available providers for model: %s", model)
	}

	healthyCandidates := r.FilterUnhealthy(candidates)
	if len(healthyCandidates) == 0 {
		return nil, fmt.Errorf("no healthy providers for model: %s", model)
	}

	verifiedCandidates := r.FilterUnverified(healthyCandidates)
	if len(verifiedCandidates) == 0 {
		return nil, fmt.Errorf("no verified providers for model: %s", model)
	}

	r.CalculateScores(verifiedCandidates, strategy)

	sort.Slice(verifiedCandidates, func(i, j int) bool {
		return verifiedCandidates[i].Score > verifiedCandidates[j].Score
	})

	return &verifiedCandidates[0], nil
}

func (r *SmartRouter) GetCandidates(ctx context.Context, model string) ([]RoutingCandidate, error) {
	if r.db == nil {
		r.db = config.GetDB()
	}
	if r.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT 
			mak.id, 
			mak.provider, 
			mak.health_status,
			mak.verified_at IS NOT NULL as verified,
			mak.cost_input_rate,
			mak.cost_output_rate,
			COALESCE(AVG(h.latency_ms), 0) as avg_latency,
			COALESCE(
				COUNT(CASE WHEN h.status = 'healthy' THEN 1 END)::float / 
				NULLIF(COUNT(*), 0) * 100, 
				100
			) as success_rate
		FROM merchant_api_keys mak
		LEFT JOIN api_key_health_history h ON mak.id = h.api_key_id
		WHERE mak.status = 'active'
		AND mak.health_status IN ('healthy', 'degraded')
		AND mak.verified_at IS NOT NULL
		GROUP BY mak.id, mak.provider, mak.health_status, mak.verified_at, 
		         mak.cost_input_rate, mak.cost_output_rate
		ORDER BY mak.last_health_check_at DESC
	`, model)
	if err != nil {
		return nil, fmt.Errorf("failed to query candidates: %w", err)
	}
	defer rows.Close()

	var candidates []RoutingCandidate
	for rows.Next() {
		var c RoutingCandidate
		var healthStatus string
		var verified bool
		var avgLatency float64
		var successRate float64

		err := rows.Scan(
			&c.APIKeyID,
			&c.Provider,
			&healthStatus,
			&verified,
			&c.InputPrice,
			&c.OutputPrice,
			&avgLatency,
			&successRate,
		)
		if err != nil {
			continue
		}

		c.Model = model
		c.HealthStatus = healthStatus
		c.Verified = verified
		c.AvgLatencyMs = int(avgLatency)
		c.SuccessRate = successRate

		candidates = append(candidates, c)
	}

	return candidates, nil
}

func (r *SmartRouter) FilterUnhealthy(candidates []RoutingCandidate) []RoutingCandidate {
	var healthy []RoutingCandidate
	for _, c := range candidates {
		if c.HealthStatus == "healthy" || c.HealthStatus == "degraded" {
			if !r.IsCircuitBreakerOpen(c.APIKeyID) {
				healthy = append(healthy, c)
			}
		}
	}
	return healthy
}

func (r *SmartRouter) FilterUnverified(candidates []RoutingCandidate) []RoutingCandidate {
	var verified []RoutingCandidate
	for _, c := range candidates {
		if c.Verified {
			verified = append(verified, c)
		}
	}
	return verified
}

func (r *SmartRouter) CalculateScores(candidates []RoutingCandidate, strategy RoutingStrategy) {
	if len(candidates) == 0 {
		return
	}

	minPrice, maxPrice := r.getPriceRange(candidates)
	minLatency, maxLatency := r.getLatencyRange(candidates)

	for i := range candidates {
		candidates[i].PriceScore = r.calculatePriceScore(candidates[i], minPrice, maxPrice)
		candidates[i].LatencyScore = r.calculateLatencyScore(candidates[i], minLatency, maxLatency)
		candidates[i].SuccessScore = candidates[i].SuccessRate / 100.0

		candidates[i].Score = r.calculateWeightedScore(
			candidates[i].PriceScore,
			candidates[i].LatencyScore,
			candidates[i].SuccessScore,
			strategy,
		)
	}
}

func (r *SmartRouter) calculatePriceScore(c RoutingCandidate, minPrice, maxPrice float64) float64 {
	if maxPrice == minPrice {
		return 1.0
	}
	price := c.InputPrice + c.OutputPrice
	return 1.0 - (price-minPrice)/(maxPrice-minPrice)
}

func (r *SmartRouter) calculateLatencyScore(c RoutingCandidate, minLatency, maxLatency int) float64 {
	if maxLatency == minLatency {
		return 1.0
	}
	return 1.0 - float64(c.AvgLatencyMs-minLatency)/float64(maxLatency-minLatency)
}

func (r *SmartRouter) calculateWeightedScore(priceScore, latencyScore, successScore float64, strategy RoutingStrategy) float64 {
	weights := r.getStrategyWeights(strategy)
	return priceScore*weights.Price +
		latencyScore*weights.Latency +
		successScore*weights.Success
}

type StrategyWeights struct {
	Price   float64
	Latency float64
	Success float64
}

func (r *SmartRouter) getStrategyWeights(strategy RoutingStrategy) StrategyWeights {
	switch strategy {
	case RoutingStrategyPrice:
		return StrategyWeights{Price: 0.6, Latency: 0.2, Success: 0.2}
	case RoutingStrategyLatency:
		return StrategyWeights{Price: 0.2, Latency: 0.6, Success: 0.2}
	case RoutingStrategyCost:
		return StrategyWeights{Price: 0.7, Latency: 0.1, Success: 0.2}
	default:
		return StrategyWeights{Price: 0.33, Latency: 0.34, Success: 0.33}
	}
}

func (r *SmartRouter) getPriceRange(candidates []RoutingCandidate) (min, max float64) {
	if len(candidates) == 0 {
		return 0, 0
	}
	min = candidates[0].InputPrice + candidates[0].OutputPrice
	max = min
	for _, c := range candidates {
		price := c.InputPrice + c.OutputPrice
		if price < min {
			min = price
		}
		if price > max {
			max = price
		}
	}
	return min, max
}

func (r *SmartRouter) getLatencyRange(candidates []RoutingCandidate) (min, max int) {
	if len(candidates) == 0 {
		return 0, 0
	}
	min = candidates[0].AvgLatencyMs
	max = min
	for _, c := range candidates {
		if c.AvgLatencyMs < min {
			min = c.AvgLatencyMs
		}
		if c.AvgLatencyMs > max {
			max = c.AvgLatencyMs
		}
	}
	return min, max
}

func (r *SmartRouter) IsCircuitBreakerOpen(apiKeyID int) bool {
	r.cbMutex.RLock()
	cb, exists := r.circuitBreaker[apiKeyID]
	r.cbMutex.RUnlock()

	if !exists {
		return false
	}

	return !cb.AllowRequest()
}

func (r *SmartRouter) RecordRequestResult(apiKeyID int, success bool) {
	r.cbMutex.Lock()
	defer r.cbMutex.Unlock()

	if _, exists := r.circuitBreaker[apiKeyID]; !exists {
		r.circuitBreaker[apiKeyID] = NewCircuitBreaker(5, 60*time.Second)
	}

	if success {
		r.circuitBreaker[apiKeyID].RecordSuccess()
	} else {
		r.circuitBreaker[apiKeyID].RecordFailure()
	}
}

func (r *SmartRouter) GetRoutingStats(ctx context.Context) (map[string]interface{}, error) {
	if r.db == nil {
		r.db = config.GetDB()
	}
	if r.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	var totalProviders, healthyProviders, degradedProviders int
	err := r.db.QueryRowContext(ctx, `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE health_status = 'healthy') as healthy,
			COUNT(*) FILTER (WHERE health_status = 'degraded') as degraded
		FROM merchant_api_keys 
		WHERE status = 'active' AND verified_at IS NOT NULL
	`).Scan(&totalProviders, &healthyProviders, &degradedProviders)
	if err != nil {
		return nil, fmt.Errorf("failed to get routing stats: %w", err)
	}

	r.cbMutex.RLock()
	openCircuitBreakers := 0
	for _, cb := range r.circuitBreaker {
		if !cb.AllowRequest() {
			openCircuitBreakers++
		}
	}
	r.cbMutex.RUnlock()

	return map[string]interface{}{
		"total_providers":       totalProviders,
		"healthy_providers":     healthyProviders,
		"degraded_providers":    degradedProviders,
		"open_circuit_breakers": openCircuitBreakers,
		"circuit_breaker_count": len(r.circuitBreaker),
	}, nil
}
