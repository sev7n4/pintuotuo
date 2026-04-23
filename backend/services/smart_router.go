package services

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/lib/pq"
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
	APIKeyID      int
	Provider      string
	Model         string
	Score         float64
	PriceScore    float64
	LatencyScore  float64
	SuccessScore  float64
	HealthStatus  string
	Verified      bool
	InputPrice    float64
	OutputPrice   float64
	AvgLatencyMs  int
	SuccessRate   float64
	Region        string
	SecurityLevel string
}

type SmartRouter struct {
	db             *sql.DB
	circuitBreaker map[int]*CircuitBreaker
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

func (r *SmartRouter) SelectProvider(ctx context.Context, model string, provider string, strategy RoutingStrategy) (*RoutingCandidate, error) {
	return r.SelectProviderWithKeyAllowlist(ctx, model, provider, strategy, nil)
}

func (r *SmartRouter) SelectProviderWithKeyAllowlist(ctx context.Context, model string, provider string, strategy RoutingStrategy, allowedKeyIDs []int) (*RoutingCandidate, error) {
	candidates, err := r.GetCandidatesWithKeyAllowlist(ctx, model, provider, allowedKeyIDs)
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

func (r *SmartRouter) SelectProviderWithStrategyOutput(ctx context.Context, model string, provider string, strategyOutput *StrategyOutput, allowedKeyIDs []int) (*RoutingCandidate, error) {
	candidates, err := r.GetCandidatesWithKeyAllowlist(ctx, model, provider, allowedKeyIDs)
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

	filteredCandidates := r.FilterByConstraints(verifiedCandidates, strategyOutput.Constraints)
	if len(filteredCandidates) == 0 {
		return nil, fmt.Errorf("no candidates satisfy constraints for model: %s", model)
	}

	r.CalculateScoresWithWeights(filteredCandidates, strategyOutput.Weights)

	sort.Slice(filteredCandidates, func(i, j int) bool {
		return filteredCandidates[i].Score > filteredCandidates[j].Score
	})

	return &filteredCandidates[0], nil
}

func (r *SmartRouter) FilterByRouteDecision(candidates []RoutingCandidate, decision *RouteDecision) []RoutingCandidate {
	// 路由决策不影响候选者过滤，返回所有候选者
	return candidates
}

func (r *SmartRouter) MatchesRouteDecision(candidate RoutingCandidate, decision *RouteDecision) bool {
	// 路由决策不影响候选者匹配，所有候选者都匹配
	return true
}

func (r *SmartRouter) FilterByConstraints(candidates []RoutingCandidate, constraints StrategyConstraints) []RoutingCandidate {
	if constraints.MinSuccessRate == 0 && len(constraints.RequiredRegions) == 0 &&
		len(constraints.ExcludedProviders) == 0 && constraints.MinSecurityLevel == "" &&
		constraints.MaxLatencyMs == 0 && constraints.MaxCostPerToken == 0 {
		return candidates
	}

	var filtered []RoutingCandidate
	for _, c := range candidates {
		if constraints.MinSuccessRate > 0 && c.SuccessRate < constraints.MinSuccessRate {
			continue
		}

		if constraints.MaxLatencyMs > 0 && c.AvgLatencyMs > constraints.MaxLatencyMs {
			continue
		}

		if constraints.MaxCostPerToken > 0 {
			totalPrice := c.InputPrice + c.OutputPrice
			if totalPrice > constraints.MaxCostPerToken {
				continue
			}
		}

		if len(constraints.RequiredRegions) > 0 {
			regionMatch := false
			for _, region := range constraints.RequiredRegions {
				if c.Region == region {
					regionMatch = true
					break
				}
			}
			if !regionMatch {
				continue
			}
		}

		excluded := false
		for _, provider := range constraints.ExcludedProviders {
			if c.Provider == provider {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}

		if constraints.MinSecurityLevel != "" {
			securityLevels := map[string]int{
				"standard": 1,
				"high":     2,
			}
			candidateLevel := securityLevels[c.SecurityLevel]
			requiredLevel := securityLevels[constraints.MinSecurityLevel]
			if candidateLevel < requiredLevel {
				continue
			}
		}

		filtered = append(filtered, c)
	}

	return filtered
}

func (r *SmartRouter) GetCandidatesWithKeyAllowlist(ctx context.Context, model string, provider string, allowedKeyIDs []int) ([]RoutingCandidate, error) {
	query := `
		SELECT 
			mak.id, mak.provider, mak.model,
			CASE 
				WHEN mak.input_price IS NOT NULL AND mak.output_price IS NOT NULL 
				THEN mak.input_price + mak.output_price
				ELSE 0 
			END as price,
			COALESCE(mak.avg_latency_ms, 0) as latency,
			COALESCE(mak.success_rate, 1.0) as success_rate,
			COALESCE(mak.region, 'domestic') as region,
			COALESCE(mak.security_level, 'standard') as security_level
		FROM merchant_api_keys mak
		WHERE mak.status = 'active'
			AND mak.model = $1
	`

	args := []interface{}{model}
	argPos := 2

	if provider != "" {
		query += fmt.Sprintf(" AND mak.provider = $%d", argPos)
		args = append(args, provider)
		argPos++
	}

	if allowedKeyIDs != nil {
		if len(allowedKeyIDs) == 0 {
			return []RoutingCandidate{}, nil
		}
		query += fmt.Sprintf(" AND mak.id = ANY($%d)", argPos)
		args = append(args, pq.Array(allowedKeyIDs))
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query candidates: %w", err)
	}
	defer rows.Close()

	var candidates []RoutingCandidate
	for rows.Next() {
		var c RoutingCandidate
		var totalPrice float64
		err := rows.Scan(
			&c.APIKeyID,
			&c.Provider,
			&c.Model,
			&totalPrice,
			&c.AvgLatencyMs,
			&c.SuccessRate,
			&c.Region,
			&c.SecurityLevel,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan candidate: %w", err)
		}

		c.InputPrice = totalPrice / 2
		c.OutputPrice = totalPrice / 2
		c.HealthStatus = "healthy"
		c.Verified = true

		candidates = append(candidates, c)
	}

	return candidates, nil
}

func (r *SmartRouter) FilterUnhealthy(candidates []RoutingCandidate) []RoutingCandidate {
	var healthy []RoutingCandidate
	for _, c := range candidates {
		if c.HealthStatus == "healthy" || c.HealthStatus == "degraded" {
			healthy = append(healthy, c)
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
		priceScore := r.calculatePriceScore(candidates[i], minPrice, maxPrice)
		latencyScore := r.calculateLatencyScore(candidates[i], minLatency, maxLatency)
		successScore := candidates[i].SuccessRate / 100.0

		candidates[i].PriceScore = priceScore
		candidates[i].LatencyScore = latencyScore
		candidates[i].SuccessScore = successScore

		candidates[i].Score = r.calculateWeightedScore(priceScore, latencyScore, successScore, strategy)
	}
}

func (r *SmartRouter) CalculateScoresWithWeights(candidates []RoutingCandidate, weights StrategyWeightsV2) {
	if len(candidates) == 0 {
		return
	}

	minPrice, maxPrice := r.getPriceRange(candidates)
	minLatency, maxLatency := r.getLatencyRange(candidates)

	for i := range candidates {
		priceScore := r.calculatePriceScore(candidates[i], minPrice, maxPrice)
		latencyScore := r.calculateLatencyScore(candidates[i], minLatency, maxLatency)
		successScore := candidates[i].SuccessRate / 100.0

		candidates[i].PriceScore = priceScore
		candidates[i].LatencyScore = latencyScore
		candidates[i].SuccessScore = successScore

		candidates[i].Score = priceScore*weights.CostWeight +
			latencyScore*weights.LatencyWeight +
			successScore*weights.ReliabilityWeight
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

func (r *SmartRouter) ReloadRoutingStrategies() {
}

func (r *SmartRouter) RecordRequestResult(apiKeyID int, success bool) {
}

func (r *SmartRouter) GetDefaultStrategyCode() string {
	return string(RoutingStrategyBalanced)
}

func (r *SmartRouter) ConfigureCircuitBreaker(apiKeyID int, threshold int, timeout time.Duration) error {
	return nil
}

func (r *SmartRouter) IsCircuitBreakerOpen(apiKeyID int) bool {
	return false
}

func (r *SmartRouter) GetStrategyConfig(strategyCode string) (*StrategyConfig, bool) {
	strategy := RoutingStrategy(strategyCode)
	switch strategy {
	case RoutingStrategyPrice, RoutingStrategyLatency, RoutingStrategyBalanced, RoutingStrategyCost:
		weights := r.getStrategyWeights(strategy)
		return &StrategyConfig{
			Strategy:               strategyCode,
			MaxRetryCount:          3,
			RetryBackoffBase:       100,
			CircuitBreakerThreshold: 5,
			CircuitBreakerTimeout:  60,
			PriceWeight:            weights.Price,
			LatencyWeight:          weights.Latency,
			SuccessWeight:          weights.Success,
			ReliabilityWeight:      weights.Success,
		}, true
	default:
		return nil, false
	}
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
