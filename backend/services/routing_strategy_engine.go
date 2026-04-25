package services

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pintuotuo/backend/config"
)

type IStrategyEngine interface {
	DefineGoal(ctx context.Context, reqCtx *RequestContext) (*StrategyOutput, error)
	GetStrategyWeights(strategy StrategyGoal) (*StrategyWeightsV2, error)
	ValidateConstraints(goal *StrategyOutput, candidate *RoutingCandidateV2) bool
	ApplyCostBudget(goal *StrategyOutput, budget float64) error
	ApplyComplianceConstraints(goal *StrategyOutput, regions []string) error
}

type RoutingCandidateV2 struct {
	APIKeyID          int     `json:"api_key_id"`
	Provider          string  `json:"provider"`
	Model             string  `json:"model"`
	Region            string  `json:"region"`
	SecurityLevel     string  `json:"security_level"`
	Latency           int     `json:"latency"`
	ErrorRate         float64 `json:"error_rate"`
	SuccessRate       float64 `json:"success_rate"`
	CostPerToken      float64 `json:"cost_per_token"`
	LoadBalanceWeight float64 `json:"load_balance_weight"`
}

type RoutingStrategyEngine struct {
	db              *sql.DB
	defaultStrategy StrategyGoal
}

func NewRoutingStrategyEngine() *RoutingStrategyEngine {
	return &RoutingStrategyEngine{
		db:              config.GetDB(),
		defaultStrategy: GoalBalanced,
	}
}

func (e *RoutingStrategyEngine) DefineGoal(ctx context.Context, reqCtx *RequestContext) (*StrategyOutput, error) {
	goal := e.determineStrategy(reqCtx)

	weights, err := e.GetStrategyWeights(goal)
	if err != nil {
		return nil, fmt.Errorf("failed to get strategy weights: %w", err)
	}

	constraints := StrategyConstraints{
		MinSuccessRate: 0.95,
	}

	output := &StrategyOutput{
		Goal:        goal,
		Weights:     *weights,
		Constraints: constraints,
		Priority:    e.calculatePriority(reqCtx),
		Reason:      e.getStrategyReason(goal, reqCtx),
	}

	if reqCtx.CostBudget != nil {
		if err := e.ApplyCostBudget(output, *reqCtx.CostBudget); err != nil {
			return nil, fmt.Errorf("failed to apply cost budget: %w", err)
		}
	}

	if len(reqCtx.ComplianceReqs) > 0 {
		if err := e.ApplyComplianceConstraints(output, reqCtx.ComplianceReqs); err != nil {
			return nil, fmt.Errorf("failed to apply compliance constraints: %w", err)
		}
	}

	return output, nil
}

func (e *RoutingStrategyEngine) determineStrategy(reqCtx *RequestContext) StrategyGoal {
	if reqCtx.UserPreferences != nil {
		if strategy, ok := reqCtx.UserPreferences["strategy"].(string); ok {
			switch StrategyGoal(strategy) {
			case GoalPerformanceFirst, GoalPriceFirst, GoalReliabilityFirst,
				GoalBalanced, GoalSecurityFirst:
				return StrategyGoal(strategy)
			}
		}
	}

	if reqCtx.RequestAnalysis != nil {
		analysis := reqCtx.RequestAnalysis

		if analysis.Complexity == ComplexityComplex {
			return GoalReliabilityFirst
		}

		if analysis.EstimatedTokens > 8000 {
			return GoalPriceFirst
		}

		if analysis.Stream {
			return GoalPerformanceFirst
		}
	}

	if reqCtx.CostBudget != nil && *reqCtx.CostBudget < 0.01 {
		return GoalPriceFirst
	}

	if len(reqCtx.ComplianceReqs) > 0 {
		return GoalSecurityFirst
	}

	return e.getDefaultStrategyFromDB()
}

func (e *RoutingStrategyEngine) GetStrategyWeights(strategy StrategyGoal) (*StrategyWeightsV2, error) {
	if e.db != nil {
		weights, err := e.getStrategyWeightsFromDB(strategy)
		if err == nil {
			return weights, nil
		}
	}

	return e.getDefaultStrategyWeights(strategy)
}

func (e *RoutingStrategyEngine) getStrategyWeightsFromDB(strategy StrategyGoal) (*StrategyWeightsV2, error) {
	var priceWeight, latencyWeight, reliabilityWeight, securityWeight, loadBalanceWeight float64
	err := e.db.QueryRow(`
		SELECT price_weight, latency_weight, reliability_weight,
		       COALESCE(security_weight, 0.1), COALESCE(load_balance_weight, 0.1)
		FROM routing_strategies 
		WHERE code = $1 AND status = 'active'`,
		string(strategy),
	).Scan(&priceWeight, &latencyWeight, &reliabilityWeight, &securityWeight, &loadBalanceWeight)
	if err != nil {
		return nil, fmt.Errorf("failed to get strategy weights from DB: %w", err)
	}

	return &StrategyWeightsV2{
		CostWeight:        priceWeight,
		LatencyWeight:     latencyWeight,
		ReliabilityWeight: reliabilityWeight,
		SecurityWeight:    securityWeight,
		LoadBalanceWeight: loadBalanceWeight,
	}, nil
}

func (e *RoutingStrategyEngine) getDefaultStrategyFromDB() StrategyGoal {
	if e.db == nil {
		return GoalBalanced
	}

	var code string
	err := e.db.QueryRow(`
		SELECT code 
		FROM routing_strategies 
		WHERE is_default = true AND status = 'active'
		LIMIT 1`,
	).Scan(&code)
	if err != nil {
		return GoalBalanced
	}

	strategy := StrategyGoal(code)
	switch strategy {
	case GoalPerformanceFirst, GoalPriceFirst, GoalReliabilityFirst,
		GoalBalanced, GoalSecurityFirst, GoalAuto:
		return strategy
	default:
		return GoalBalanced
	}
}

func (e *RoutingStrategyEngine) getDefaultStrategyWeights(strategy StrategyGoal) (*StrategyWeightsV2, error) {
	switch strategy {
	case GoalPerformanceFirst:
		return &StrategyWeightsV2{
			LatencyWeight:     0.5,
			CostWeight:        0.1,
			ReliabilityWeight: 0.2,
			SecurityWeight:    0.1,
			LoadBalanceWeight: 0.1,
		}, nil

	case GoalPriceFirst:
		return &StrategyWeightsV2{
			LatencyWeight:     0.1,
			CostWeight:        0.6,
			ReliabilityWeight: 0.15,
			SecurityWeight:    0.05,
			LoadBalanceWeight: 0.1,
		}, nil

	case GoalReliabilityFirst:
		return &StrategyWeightsV2{
			LatencyWeight:     0.2,
			CostWeight:        0.1,
			ReliabilityWeight: 0.5,
			SecurityWeight:    0.1,
			LoadBalanceWeight: 0.1,
		}, nil

	case GoalSecurityFirst:
		return &StrategyWeightsV2{
			LatencyWeight:     0.1,
			CostWeight:        0.1,
			ReliabilityWeight: 0.2,
			SecurityWeight:    0.5,
			LoadBalanceWeight: 0.1,
		}, nil

	case GoalBalanced:
		return &StrategyWeightsV2{
			LatencyWeight:     0.25,
			CostWeight:        0.25,
			ReliabilityWeight: 0.25,
			SecurityWeight:    0.15,
			LoadBalanceWeight: 0.1,
		}, nil

	case GoalAuto:
		return &StrategyWeightsV2{
			LatencyWeight:     0.3,
			CostWeight:        0.2,
			ReliabilityWeight: 0.3,
			SecurityWeight:    0.1,
			LoadBalanceWeight: 0.1,
		}, nil

	default:
		return nil, fmt.Errorf("unknown strategy: %s", strategy)
	}
}

func (e *RoutingStrategyEngine) DetermineAutoStrategyWeights(req *RoutingRequest) *StrategyWeightsV2 {
	weights := &StrategyWeightsV2{
		CostWeight:        0.25,
		LatencyWeight:     0.25,
		ReliabilityWeight: 0.25,
		SecurityWeight:    0.15,
		LoadBalanceWeight: 0.10,
	}

	if req == nil {
		return weights
	}

	if req.Stream {
		weights.LatencyWeight = 0.40
		weights.CostWeight = 0.15
		weights.ReliabilityWeight = 0.20
		weights.SecurityWeight = 0.15
		weights.LoadBalanceWeight = 0.10
		return weights
	}

	if req.MaxTokens > 4000 {
		weights.CostWeight = 0.45
		weights.LatencyWeight = 0.15
		weights.ReliabilityWeight = 0.20
		weights.SecurityWeight = 0.10
		weights.LoadBalanceWeight = 0.10
		return weights
	}

	if len(req.ComplianceReqs) > 0 {
		weights.SecurityWeight = 0.40
		weights.CostWeight = 0.15
		weights.LatencyWeight = 0.15
		weights.ReliabilityWeight = 0.20
		weights.LoadBalanceWeight = 0.10
		return weights
	}

	if req.Priority == "high" {
		weights.ReliabilityWeight = 0.40
		weights.CostWeight = 0.15
		weights.LatencyWeight = 0.20
		weights.SecurityWeight = 0.15
		weights.LoadBalanceWeight = 0.10
		return weights
	}

	return weights
}

func (e *RoutingStrategyEngine) ValidateConstraints(goal *StrategyOutput, candidate *RoutingCandidateV2) bool {
	if goal.Constraints.MinSuccessRate > 0 && candidate.SuccessRate < goal.Constraints.MinSuccessRate {
		return false
	}

	if goal.Constraints.MaxLatencyMs > 0 && candidate.Latency > goal.Constraints.MaxLatencyMs {
		return false
	}

	if goal.Constraints.MaxCostPerToken > 0 && candidate.CostPerToken > goal.Constraints.MaxCostPerToken {
		return false
	}

	if len(goal.Constraints.RequiredRegions) > 0 {
		regionMatch := false
		for _, region := range goal.Constraints.RequiredRegions {
			if candidate.Region == region {
				regionMatch = true
				break
			}
		}
		if !regionMatch {
			return false
		}
	}

	for _, excluded := range goal.Constraints.ExcludedProviders {
		if candidate.Provider == excluded {
			return false
		}
	}

	if goal.Constraints.MinSecurityLevel != "" {
		securityLevels := map[string]int{
			"standard": 1,
			"high":     2,
		}
		candidateLevel := securityLevels[candidate.SecurityLevel]
		requiredLevel := securityLevels[goal.Constraints.MinSecurityLevel]
		if candidateLevel < requiredLevel {
			return false
		}
	}

	return true
}

func (e *RoutingStrategyEngine) ApplyCostBudget(goal *StrategyOutput, budget float64) error {
	goal.Constraints.MaxCostPerToken = budget / 1000.0
	return nil
}

func (e *RoutingStrategyEngine) ApplyComplianceConstraints(goal *StrategyOutput, regions []string) error {
	goal.Constraints.RequiredRegions = regions
	goal.Constraints.MinSecurityLevel = "high"
	return nil
}

func (e *RoutingStrategyEngine) calculatePriority(reqCtx *RequestContext) int {
	priority := 5

	if reqCtx.RequestAnalysis != nil {
		if reqCtx.RequestAnalysis.Complexity == ComplexityComplex {
			priority += 2
		} else if reqCtx.RequestAnalysis.Complexity == ComplexitySimple {
			priority -= 1
		}
	}

	if reqCtx.CostBudget != nil && *reqCtx.CostBudget > 0.1 {
		priority += 1
	}

	if len(reqCtx.ComplianceReqs) > 0 {
		priority += 1
	}

	if priority > 10 {
		priority = 10
	}
	if priority < 1 {
		priority = 1
	}

	return priority
}

func (e *RoutingStrategyEngine) getStrategyReason(strategy StrategyGoal, reqCtx *RequestContext) string {
	switch strategy {
	case GoalPerformanceFirst:
		if reqCtx.RequestAnalysis != nil && reqCtx.RequestAnalysis.Stream {
			return "Stream request requires low latency"
		}
		return "Performance priority based on user preference"

	case GoalPriceFirst:
		if reqCtx.CostBudget != nil {
			return fmt.Sprintf("Cost budget constraint: $%.4f", *reqCtx.CostBudget)
		}
		if reqCtx.RequestAnalysis != nil && reqCtx.RequestAnalysis.EstimatedTokens > 8000 {
			return "High token count requires cost optimization"
		}
		return "Price priority based on user preference"

	case GoalReliabilityFirst:
		if reqCtx.RequestAnalysis != nil && reqCtx.RequestAnalysis.Complexity == ComplexityComplex {
			return "Complex request requires high reliability"
		}
		return "Reliability priority based on user preference"

	case GoalSecurityFirst:
		if len(reqCtx.ComplianceReqs) > 0 {
			return fmt.Sprintf("Compliance requirements: %v", reqCtx.ComplianceReqs)
		}
		return "Security priority based on user preference"

	case GoalBalanced:
		return "Balanced strategy for general use"

	case GoalAuto:
		return "Auto strategy with adaptive weights"

	default:
		return "Unknown strategy"
	}
}
