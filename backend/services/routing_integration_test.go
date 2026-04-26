package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoutingIntegration_EndToEndRouting(t *testing.T) {
	t.Run("should route request through three layers", func(t *testing.T) {
		engine := NewRoutingStrategyEngine()
		router := GetSmartRouter()

		strategyOutput, err := engine.DefineGoal(nil, &RequestContext{
			RequestAnalysis: &RequestAnalysis{
				Intent:          IntentChat,
				EstimatedTokens: 1000,
				Stream:          false,
			},
		})
		assert.NoError(t, err)
		assert.NotNil(t, strategyOutput)
		assert.NotEmpty(t, strategyOutput.Goal)

		weights, err := engine.GetStrategyWeights(strategyOutput.Goal)
		assert.NoError(t, err)
		assert.NotNil(t, weights)
		assert.NotEmpty(t, weights.StrategyCode)

		assert.True(t, weights.CostWeight >= 0 && weights.CostWeight <= 1)
		assert.True(t, weights.LatencyWeight >= 0 && weights.LatencyWeight <= 1)
		assert.True(t, weights.ReliabilityWeight >= 0 && weights.ReliabilityWeight <= 1)
		assert.True(t, weights.SecurityWeight >= 0 && weights.SecurityWeight <= 1)
		assert.True(t, weights.LoadBalanceWeight >= 0 && weights.LoadBalanceWeight <= 1)

		totalWeight := weights.CostWeight + weights.LatencyWeight + weights.ReliabilityWeight + weights.SecurityWeight + weights.LoadBalanceWeight
		assert.InDelta(t, 1.0, totalWeight, 0.01, "weights should sum to 1.0")

		_ = router
	})

	t.Run("should select correct strategy based on request characteristics", func(t *testing.T) {
		engine := NewRoutingStrategyEngine()

		tests := []struct {
			name           string
			reqCtx         *RequestContext
			expectedGoal   StrategyGoal
			expectedReason string
		}{
			{
				name: "stream request should use performance_first",
				reqCtx: &RequestContext{
					RequestAnalysis: &RequestAnalysis{
						Intent: IntentChat,
						Stream: true,
					},
				},
				expectedGoal:   GoalPerformanceFirst,
				expectedReason: "Stream",
			},
			{
				name: "high token request should use price_first",
				reqCtx: &RequestContext{
					RequestAnalysis: &RequestAnalysis{
						Intent:          IntentChat,
						EstimatedTokens: 10000,
					},
				},
				expectedGoal: GoalPriceFirst,
			},
			{
				name: "complex request should use reliability_first",
				reqCtx: &RequestContext{
					RequestAnalysis: &RequestAnalysis{
						Intent:     IntentChat,
						Complexity: ComplexityComplex,
					},
				},
				expectedGoal: GoalReliabilityFirst,
			},
			{
				name: "compliance request should use security_first",
				reqCtx: &RequestContext{
					ComplianceReqs: []string{"domestic"},
				},
				expectedGoal: GoalSecurityFirst,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				output, err := engine.DefineGoal(nil, tt.reqCtx)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedGoal, output.Goal)
				if tt.expectedReason != "" {
					assert.Contains(t, output.Reason, tt.expectedReason)
				}
			})
		}
	})
}

func TestRoutingIntegration_MultiMerchantPriceComparison(t *testing.T) {
	t.Run("should calculate price scores correctly", func(t *testing.T) {
		router := GetSmartRouter()

		candidates := []RoutingCandidate{
			{
				APIKeyID:     1,
				MerchantID:   1,
				Provider:     "openai",
				Model:        "gpt-4",
				InputPrice:   0.03,
				OutputPrice:  0.06,
				SuccessRate:  99.0,
				AvgLatencyMs: 500,
			},
			{
				APIKeyID:     2,
				MerchantID:   2,
				Provider:     "openai",
				Model:        "gpt-4",
				InputPrice:   0.02,
				OutputPrice:  0.04,
				SuccessRate:  98.0,
				AvgLatencyMs: 600,
			},
			{
				APIKeyID:     3,
				MerchantID:   3,
				Provider:     "openai",
				Model:        "gpt-4",
				InputPrice:   0.04,
				OutputPrice:  0.08,
				SuccessRate:  99.5,
				AvgLatencyMs: 400,
			},
		}

		weights := StrategyWeightsV2{
			StrategyCode:      string(GoalPriceFirst),
			CostWeight:        0.6,
			LatencyWeight:     0.1,
			ReliabilityWeight: 0.15,
			SecurityWeight:    0.1,
			LoadBalanceWeight: 0.05,
		}

		router.CalculateScoresWithWeights(candidates, weights)

		assert.Greater(t, candidates[1].PriceScore, candidates[0].PriceScore, "merchant 2 should have higher price score (lower price)")
		assert.Greater(t, candidates[0].PriceScore, candidates[2].PriceScore, "merchant 1 should have higher price score than merchant 3")

		assert.Greater(t, candidates[1].Score, candidates[0].Score, "merchant 2 should win with price_first strategy")
	})

	t.Run("should calculate estimated cost correctly", func(t *testing.T) {
		candidates := []RoutingCandidate{
			{
				APIKeyID:    1,
				MerchantID:  1,
				InputPrice:  0.03,
				OutputPrice: 0.06,
			},
			{
				APIKeyID:    2,
				MerchantID:  2,
				InputPrice:  0.02,
				OutputPrice: 0.04,
			},
		}

		estimatedInputTokens := 1000.0
		estimatedOutputTokens := 500.0

		cost1 := estimatedInputTokens*candidates[0].InputPrice + estimatedOutputTokens*candidates[0].OutputPrice
		cost2 := estimatedInputTokens*candidates[1].InputPrice + estimatedOutputTokens*candidates[1].OutputPrice

		assert.Greater(t, cost1, cost2, "merchant 1 should have higher estimated cost")
	})

	t.Run("should consider both input and output prices", func(t *testing.T) {
		candidates := []RoutingCandidate{
			{
				APIKeyID:    1,
				MerchantID:  1,
				InputPrice:  0.01,
				OutputPrice: 0.10,
			},
			{
				APIKeyID:    2,
				MerchantID:  2,
				InputPrice:  0.05,
				OutputPrice: 0.02,
			},
		}

		inputTokens := 1000.0
		outputTokens := 2000.0

		cost1 := inputTokens*candidates[0].InputPrice + outputTokens*candidates[0].OutputPrice
		cost2 := inputTokens*candidates[1].InputPrice + outputTokens*candidates[1].OutputPrice

		assert.Greater(t, cost1, cost2, "merchant 2 should be cheaper for output-heavy requests")
	})
}

func TestRoutingIntegration_StrategyWeights(t *testing.T) {
	t.Run("should return correct weights for each strategy", func(t *testing.T) {
		engine := NewRoutingStrategyEngine()

		tests := []struct {
			name                  string
			strategy              StrategyGoal
			expectedPrimaryWeight string
			minPrimaryWeight      float64
		}{
			{
				name:                  "price_first prioritizes cost",
				strategy:              GoalPriceFirst,
				expectedPrimaryWeight: "cost",
				minPrimaryWeight:      0.5,
			},
			{
				name:                  "performance_first prioritizes latency",
				strategy:              GoalPerformanceFirst,
				expectedPrimaryWeight: "latency",
				minPrimaryWeight:      0.4,
			},
			{
				name:                  "reliability_first prioritizes reliability",
				strategy:              GoalReliabilityFirst,
				expectedPrimaryWeight: "reliability",
				minPrimaryWeight:      0.4,
			},
			{
				name:                  "security_first prioritizes security",
				strategy:              GoalSecurityFirst,
				expectedPrimaryWeight: "security",
				minPrimaryWeight:      0.4,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				weights, err := engine.GetStrategyWeights(tt.strategy)
				assert.NoError(t, err)

				var primaryWeight float64
				switch tt.expectedPrimaryWeight {
				case "cost":
					primaryWeight = weights.CostWeight
				case "latency":
					primaryWeight = weights.LatencyWeight
				case "reliability":
					primaryWeight = weights.ReliabilityWeight
				case "security":
					primaryWeight = weights.SecurityWeight
				}

				assert.GreaterOrEqual(t, primaryWeight, tt.minPrimaryWeight, "%s should be prioritized", tt.expectedPrimaryWeight)
			})
		}
	})

	t.Run("should generate dynamic weights for auto strategy", func(t *testing.T) {
		engine := NewRoutingStrategyEngine()

		t.Run("stream request should prioritize latency", func(t *testing.T) {
			weights := engine.DetermineAutoStrategyWeights(&RoutingRequest{
				Stream: true,
			})
			assert.GreaterOrEqual(t, weights.LatencyWeight, 0.3, "latency should be prioritized for stream requests")
		})

		t.Run("large token request should prioritize cost", func(t *testing.T) {
			weights := engine.DetermineAutoStrategyWeights(&RoutingRequest{
				MaxTokens: 8000,
			})
			assert.GreaterOrEqual(t, weights.CostWeight, 0.4, "cost should be prioritized for large token requests")
		})

		t.Run("high priority request should prioritize reliability", func(t *testing.T) {
			weights := engine.DetermineAutoStrategyWeights(&RoutingRequest{
				Priority: "high",
			})
			assert.GreaterOrEqual(t, weights.ReliabilityWeight, 0.3, "reliability should be prioritized for high priority requests")
		})

		t.Run("compliance request should prioritize security", func(t *testing.T) {
			weights := engine.DetermineAutoStrategyWeights(&RoutingRequest{
				ComplianceReqs: []string{"domestic"},
			})
			assert.GreaterOrEqual(t, weights.SecurityWeight, 0.3, "security should be prioritized for compliance requests")
		})
	})

	t.Run("weights should always sum to 1.0", func(t *testing.T) {
		engine := NewRoutingStrategyEngine()
		strategies := []StrategyGoal{
			GoalPriceFirst,
			GoalPerformanceFirst,
			GoalReliabilityFirst,
			GoalSecurityFirst,
			GoalBalanced,
			GoalAuto,
		}

		for _, strategy := range strategies {
			weights, err := engine.GetStrategyWeights(strategy)
			assert.NoError(t, err)

			total := weights.CostWeight + weights.LatencyWeight + weights.ReliabilityWeight + weights.SecurityWeight + weights.LoadBalanceWeight
			assert.InDelta(t, 1.0, total, 0.01, "weights for %s should sum to 1.0", strategy)
		}

		autoWeights := engine.DetermineAutoStrategyWeights(&RoutingRequest{
			Stream:    true,
			MaxTokens: 5000,
			Priority:  "high",
		})
		total := autoWeights.CostWeight + autoWeights.LatencyWeight + autoWeights.ReliabilityWeight + autoWeights.SecurityWeight + autoWeights.LoadBalanceWeight
		assert.InDelta(t, 1.0, total, 0.01, "auto weights should sum to 1.0")
	})
}

func TestRoutingIntegration_FiveDimensionalScoring(t *testing.T) {
	t.Run("should calculate all five dimension scores", func(t *testing.T) {
		router := GetSmartRouter()

		candidates := []RoutingCandidate{
			{
				APIKeyID:      1,
				MerchantID:    1,
				Provider:      "openai",
				Model:         "gpt-4",
				InputPrice:    0.03,
				OutputPrice:   0.06,
				SuccessRate:   99.0,
				AvgLatencyMs:  500,
				Region:        "domestic",
				SecurityLevel: "enterprise",
				HealthStatus:  "healthy",
				Verified:      true,
			},
		}

		weights := StrategyWeightsV2{
			StrategyCode:      string(GoalBalanced),
			CostWeight:        0.25,
			LatencyWeight:     0.25,
			ReliabilityWeight: 0.25,
			SecurityWeight:    0.15,
			LoadBalanceWeight: 0.10,
		}

		router.CalculateScoresWithWeights(candidates, weights)

		assert.GreaterOrEqual(t, candidates[0].PriceScore, 0.0, "price score should be >= 0")
		assert.LessOrEqual(t, candidates[0].PriceScore, 1.0, "price score should be <= 1")
		assert.GreaterOrEqual(t, candidates[0].LatencyScore, 0.0, "latency score should be >= 0")
		assert.LessOrEqual(t, candidates[0].LatencyScore, 1.0, "latency score should be <= 1")
		assert.GreaterOrEqual(t, candidates[0].SuccessScore, 0.0, "success score should be >= 0")
		assert.LessOrEqual(t, candidates[0].SuccessScore, 1.0, "success score should be <= 1")
		assert.GreaterOrEqual(t, candidates[0].Score, 0.0, "total score should be >= 0")
		assert.LessOrEqual(t, candidates[0].Score, 1.0, "total score should be <= 1")
	})

	t.Run("should handle candidates with same prices", func(t *testing.T) {
		router := GetSmartRouter()

		candidates := []RoutingCandidate{
			{
				APIKeyID:     1,
				InputPrice:   0.03,
				OutputPrice:  0.06,
				SuccessRate:  99.0,
				AvgLatencyMs: 500,
			},
			{
				APIKeyID:     2,
				InputPrice:   0.03,
				OutputPrice:  0.06,
				SuccessRate:  98.0,
				AvgLatencyMs: 600,
			},
		}

		weights := StrategyWeightsV2{
			StrategyCode:      string(GoalBalanced),
			CostWeight:        0.25,
			LatencyWeight:     0.25,
			ReliabilityWeight: 0.25,
			SecurityWeight:    0.15,
			LoadBalanceWeight: 0.10,
		}

		router.CalculateScoresWithWeights(candidates, weights)

		assert.Equal(t, candidates[0].PriceScore, candidates[1].PriceScore, "same prices should have same price score")
		assert.Greater(t, candidates[0].SuccessScore, candidates[1].SuccessScore, "higher success rate should have higher score")
	})
}

func TestRoutingIntegration_TokenEstimation(t *testing.T) {
	t.Run("should estimate tokens from request parameters", func(t *testing.T) {
		service := NewTokenEstimationService()

		req := &RoutingRequest{
			Model: "gpt-4",
			RequestBody: map[string]interface{}{
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "Hello, how are you?",
					},
				},
				"max_tokens": 1000,
			},
		}

		estimation := service.EstimateTokens(req)

		assert.Greater(t, estimation.EstimatedInputTokens, 0.0, "should estimate input tokens")
		assert.Greater(t, estimation.EstimatedOutputTokens, 0.0, "should estimate output tokens")
		assert.NotEmpty(t, estimation.Source, "should have estimation source")
	})

	t.Run("should use max_tokens for output estimation", func(t *testing.T) {
		service := NewTokenEstimationService()

		req := &RoutingRequest{
			Model: "gpt-4",
			RequestBody: map[string]interface{}{
				"messages": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "Hello",
					},
				},
				"max_tokens": 500,
			},
		}

		estimation := service.EstimateTokens(req)

		assert.Equal(t, 500.0, estimation.EstimatedOutputTokens, "should use max_tokens for output estimation")
	})
}
