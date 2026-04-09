package services

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
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
			COALESCE(
				mak.cost_input_rate,
				MAX(ms.cost_input_rate),
				MAX(sp.provider_input_rate),
				0
			) as effective_input_rate,
			COALESCE(
				mak.cost_output_rate,
				MAX(ms.cost_output_rate),
				MAX(sp.provider_output_rate),
				0
			) as effective_output_rate,
			COALESCE(AVG(h.latency_ms), 0) as avg_latency,
			COALESCE(
				COUNT(CASE WHEN h.status = 'healthy' THEN 1 END)::float / 
				NULLIF(COUNT(*), 0) * 100, 
				100
			) as success_rate
		FROM merchant_api_keys mak
		LEFT JOIN api_key_health_history h ON mak.id = h.api_key_id
		LEFT JOIN merchant_skus ms ON ms.api_key_id = mak.id AND ms.status = 'active'
		LEFT JOIN skus s ON s.id = ms.sku_id
		LEFT JOIN spus sp ON sp.id = s.spu_id
		WHERE mak.status = 'active'
		AND mak.health_status IN ('healthy', 'degraded')
		AND mak.verified_at IS NOT NULL
		GROUP BY mak.id, mak.provider, mak.health_status, mak.verified_at, 
		         mak.cost_input_rate, mak.cost_output_rate
		ORDER BY mak.last_health_check_at DESC
	`)
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

func (r *SmartRouter) ConfigureCircuitBreaker(apiKeyID int, threshold int, timeout time.Duration) {
	r.cbMutex.Lock()
	defer r.cbMutex.Unlock()
	if threshold <= 0 {
		threshold = 5
	}
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	cb, exists := r.circuitBreaker[apiKeyID]
	if !exists {
		r.circuitBreaker[apiKeyID] = NewCircuitBreaker(threshold, timeout)
		return
	}

	cb.mu.Lock()
	cb.threshold = threshold
	cb.timeout = timeout
	cb.mu.Unlock()
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

type StrategyConfig struct {
	ID                      int
	Name                    string
	Code                    string
	PriceWeight             float64
	LatencyWeight           float64
	ReliabilityWeight       float64
	MaxRetryCount           int
	RetryBackoffBase        int
	CircuitBreakerThreshold int
	CircuitBreakerTimeout   int
	IsDefault               bool
}

// strategyCacheAtomic holds map[string]StrategyConfig (hot-reloaded via ReloadRoutingStrategies).
var strategyCacheAtomic atomic.Value

func init() {
	strategyCacheAtomic.Store(make(map[string]StrategyConfig))
}

func snapshotStrategyCache() map[string]StrategyConfig {
	v := strategyCacheAtomic.Load()
	if v == nil {
		return make(map[string]StrategyConfig)
	}
	m, ok := v.(map[string]StrategyConfig)
	if !ok {
		return make(map[string]StrategyConfig)
	}
	return m
}

// ReloadRoutingStrategies rebuilds the in-memory routing strategy cache from the database (or built-in defaults).
func (r *SmartRouter) ReloadRoutingStrategies() {
	m := r.loadStrategyCacheMapFromDB()
	if len(m) == 0 {
		m = defaultStrategiesMap()
	}
	strategyCacheAtomic.Store(m)
}

func (r *SmartRouter) GetStrategyConfig(strategyCode string) (StrategyConfig, bool) {
	m := snapshotStrategyCache()
	if len(m) == 0 {
		r.ReloadRoutingStrategies()
		m = snapshotStrategyCache()
	}
	config, found := m[strategyCode]
	return config, found
}

func (r *SmartRouter) GetDefaultStrategyCode() string {
	m := snapshotStrategyCache()
	if len(m) == 0 {
		r.ReloadRoutingStrategies()
		m = snapshotStrategyCache()
	}
	type pair struct {
		code string
		id   int
	}
	var defaults []pair
	for code, cfg := range m {
		if cfg.IsDefault {
			defaults = append(defaults, pair{code: code, id: cfg.ID})
		}
	}
	sort.Slice(defaults, func(i, j int) bool {
		if defaults[i].id != defaults[j].id {
			return defaults[i].id < defaults[j].id
		}
		return defaults[i].code < defaults[j].code
	})
	if len(defaults) > 0 {
		return defaults[0].code
	}
	return ""
}

func defaultStrategiesMap() map[string]StrategyConfig {
	defaultStrategies := []StrategyConfig{
		{
			ID:                      1,
			Name:                    "价格优先",
			Code:                    "price_first",
			PriceWeight:             0.6,
			LatencyWeight:           0.2,
			ReliabilityWeight:       0.2,
			MaxRetryCount:           3,
			RetryBackoffBase:        1000,
			CircuitBreakerThreshold: 5,
			CircuitBreakerTimeout:   60,
			IsDefault:               false,
		},
		{
			ID:                      2,
			Name:                    "延迟优先",
			Code:                    "latency_first",
			PriceWeight:             0.2,
			LatencyWeight:           0.6,
			ReliabilityWeight:       0.2,
			MaxRetryCount:           3,
			RetryBackoffBase:        1000,
			CircuitBreakerThreshold: 5,
			CircuitBreakerTimeout:   60,
			IsDefault:               false,
		},
		{
			ID:                      3,
			Name:                    "均衡策略",
			Code:                    "balanced",
			PriceWeight:             0.33,
			LatencyWeight:           0.34,
			ReliabilityWeight:       0.33,
			MaxRetryCount:           3,
			RetryBackoffBase:        1000,
			CircuitBreakerThreshold: 5,
			CircuitBreakerTimeout:   60,
			IsDefault:               true,
		},
		{
			ID:                      4,
			Name:                    "可靠性优先",
			Code:                    "reliability_first",
			PriceWeight:             0.2,
			LatencyWeight:           0.2,
			ReliabilityWeight:       0.6,
			MaxRetryCount:           3,
			RetryBackoffBase:        1000,
			CircuitBreakerThreshold: 5,
			CircuitBreakerTimeout:   60,
			IsDefault:               false,
		},
	}
	out := make(map[string]StrategyConfig, len(defaultStrategies))
	for _, s := range defaultStrategies {
		out[s.Code] = s
	}
	return out
}

func (r *SmartRouter) loadStrategyCacheMapFromDB() map[string]StrategyConfig {
	if r.db == nil {
		r.db = config.GetDB()
	}
	if r.db == nil {
		return nil
	}

	rows, err := r.db.Query(`
		SELECT 
			id, name, code,
			price_weight, latency_weight, reliability_weight,
			max_retry_count, retry_backoff_base,
			circuit_breaker_threshold, circuit_breaker_timeout,
			is_default
		FROM routing_strategies
		WHERE status = 'active'
		ORDER BY id ASC
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	m := make(map[string]StrategyConfig)
	for rows.Next() {
		var config StrategyConfig
		err := rows.Scan(
			&config.ID, &config.Name, &config.Code,
			&config.PriceWeight, &config.LatencyWeight, &config.ReliabilityWeight,
			&config.MaxRetryCount, &config.RetryBackoffBase,
			&config.CircuitBreakerThreshold, &config.CircuitBreakerTimeout,
			&config.IsDefault,
		)
		if err == nil {
			m[config.Code] = config
		}
	}
	return m
}
