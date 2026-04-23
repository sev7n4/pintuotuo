package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestAnalyzer_Analyze(t *testing.T) {
	analyzer := NewRequestAnalyzer()
	ctx := context.Background()

	t.Run("should analyze chat request correctly", func(t *testing.T) {
		body := []byte(`{
			"model": "gpt-4",
			"messages": [
				{"role": "user", "content": "Hello, how are you?"}
			],
			"stream": true,
			"temperature": 0.7
		}`)

		analysis, err := analyzer.Analyze(ctx, nil, body)
		assert.NoError(t, err)
		assert.NotNil(t, analysis)
		assert.Equal(t, IntentChat, analysis.Intent)
		assert.Equal(t, "gpt-4", analysis.Model)
		assert.True(t, analysis.Stream)
		assert.NotNil(t, analysis.Temperature)
		assert.Equal(t, 0.7, *analysis.Temperature)
		assert.Greater(t, analysis.EstimatedTokens, 0)
	})

	t.Run("should analyze completion request correctly", func(t *testing.T) {
		body := []byte(`{
			"model": "text-davinci-003",
			"prompt": "Once upon a time",
			"max_tokens": 100
		}`)

		analysis, err := analyzer.Analyze(ctx, nil, body)
		assert.NoError(t, err)
		assert.NotNil(t, analysis)
		assert.Equal(t, IntentCompletion, analysis.Intent)
		assert.Equal(t, "text-davinci-003", analysis.Model)
		assert.NotNil(t, analysis.MaxTokens)
		assert.Equal(t, 100, *analysis.MaxTokens)
	})

	t.Run("should analyze embedding request correctly", func(t *testing.T) {
		body := []byte(`{
			"model": "text-embedding-ada-002",
			"input": "This is a test"
		}`)

		analysis, err := analyzer.Analyze(ctx, nil, body)
		assert.NoError(t, err)
		assert.NotNil(t, analysis)
		assert.Equal(t, IntentEmbedding, analysis.Intent)
	})

	t.Run("should handle unknown intent", func(t *testing.T) {
		body := []byte(`{
			"unknown_field": "value"
		}`)

		analysis, err := analyzer.Analyze(ctx, nil, body)
		assert.NoError(t, err)
		assert.NotNil(t, analysis)
		assert.Equal(t, IntentUnknown, analysis.Intent)
	})
}

func TestRequestAnalyzer_EstimateTokens(t *testing.T) {
	analyzer := NewRequestAnalyzer()

	t.Run("should estimate tokens for English text", func(t *testing.T) {
		text := "Hello, how are you today? This is a test message."
		tokens := analyzer.EstimateTokens(text)
		assert.Greater(t, tokens, 0)
		assert.Less(t, tokens, 50)
	})

	t.Run("should estimate tokens for Chinese text", func(t *testing.T) {
		text := "你好，今天天气怎么样？这是一个测试消息。"
		tokens := analyzer.EstimateTokens(text)
		assert.Greater(t, tokens, 0)
	})

	t.Run("should handle empty text", func(t *testing.T) {
		tokens := analyzer.EstimateTokens("")
		assert.Equal(t, 0, tokens)
	})
}

func TestRequestAnalyzer_DetectIntent(t *testing.T) {
	analyzer := NewRequestAnalyzer()

	tests := []struct {
		name     string
		body     string
		expected RequestIntent
	}{
		{
			name:     "chat with messages",
			body:     `{"messages": [{"role": "user", "content": "test"}]}`,
			expected: IntentChat,
		},
		{
			name:     "completion with prompt",
			body:     `{"prompt": "Once upon a time"}`,
			expected: IntentCompletion,
		},
		{
			name:     "embedding with model",
			body:     `{"model": "text-embedding-ada-002", "input": "test"}`,
			expected: IntentEmbedding,
		},
		{
			name:     "unknown intent",
			body:     `{"unknown": "field"}`,
			expected: IntentUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			intent := analyzer.DetectIntent([]byte(tt.body))
			assert.Equal(t, tt.expected, intent)
		})
	}
}

func TestRequestAnalyzer_AssessComplexity(t *testing.T) {
	analyzer := NewRequestAnalyzer()

	t.Run("should assess simple complexity", func(t *testing.T) {
		analysis := &RequestAnalysis{
			Intent:          IntentChat,
			EstimatedTokens: 500,
			Stream:          false,
		}
		complexity := analyzer.AssessComplexity(analysis)
		assert.Equal(t, ComplexitySimple, complexity)
	})

	t.Run("should assess medium complexity", func(t *testing.T) {
		analysis := &RequestAnalysis{
			Intent:          IntentChat,
			EstimatedTokens: 1500,
			Stream:          true,
		}
		complexity := analyzer.AssessComplexity(analysis)
		assert.Equal(t, ComplexityMedium, complexity)
	})

	t.Run("should assess complex complexity", func(t *testing.T) {
		maxTokens := 4000
		analysis := &RequestAnalysis{
			Intent:          IntentChat,
			EstimatedTokens: 5000,
			Stream:          true,
			MaxTokens:       &maxTokens,
		}
		complexity := analyzer.AssessComplexity(analysis)
		assert.Equal(t, ComplexityComplex, complexity)
	})
}

func TestRoutingStrategyEngine_DefineGoal(t *testing.T) {
	engine := NewRoutingStrategyEngine()
	ctx := context.Background()

	t.Run("should define performance_first strategy for stream request", func(t *testing.T) {
		reqCtx := &RequestContext{
			RequestAnalysis: &RequestAnalysis{
				Intent:   IntentChat,
				Stream:   true,
				EstimatedTokens: 1000,
			},
		}

		output, err := engine.DefineGoal(ctx, reqCtx)
		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, GoalPerformanceFirst, output.Goal)
		assert.Contains(t, output.Reason, "Stream")
	})

	t.Run("should define price_first strategy for high token count", func(t *testing.T) {
		reqCtx := &RequestContext{
			RequestAnalysis: &RequestAnalysis{
				Intent:          IntentChat,
				EstimatedTokens: 10000,
			},
		}

		output, err := engine.DefineGoal(ctx, reqCtx)
		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, GoalPriceFirst, output.Goal)
	})

	t.Run("should define reliability_first strategy for complex request", func(t *testing.T) {
		reqCtx := &RequestContext{
			RequestAnalysis: &RequestAnalysis{
				Intent:     IntentChat,
				Complexity: ComplexityComplex,
			},
		}

		output, err := engine.DefineGoal(ctx, reqCtx)
		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, GoalReliabilityFirst, output.Goal)
	})

	t.Run("should define security_first strategy for compliance requirements", func(t *testing.T) {
		reqCtx := &RequestContext{
			ComplianceReqs: []string{"domestic"},
		}

		output, err := engine.DefineGoal(ctx, reqCtx)
		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, GoalSecurityFirst, output.Goal)
	})

	t.Run("should apply cost budget constraint", func(t *testing.T) {
		budget := 0.05
		reqCtx := &RequestContext{
			CostBudget: &budget,
		}

		output, err := engine.DefineGoal(ctx, reqCtx)
		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Greater(t, output.Constraints.MaxCostPerToken, 0.0)
	})

	t.Run("should apply compliance constraints", func(t *testing.T) {
		reqCtx := &RequestContext{
			ComplianceReqs: []string{"domestic", "overseas"},
		}

		output, err := engine.DefineGoal(ctx, reqCtx)
		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, []string{"domestic", "overseas"}, output.Constraints.RequiredRegions)
		assert.Equal(t, "high", output.Constraints.MinSecurityLevel)
	})
}

func TestRoutingStrategyEngine_GetStrategyWeights(t *testing.T) {
	engine := NewRoutingStrategyEngine()

	tests := []struct {
		name           string
		strategy       StrategyGoal
		expectedWeight float64
		weightField    string
	}{
		{
			name:           "performance_first should prioritize latency",
			strategy:       GoalPerformanceFirst,
			expectedWeight: 0.5,
			weightField:    "latency",
		},
		{
			name:           "price_first should prioritize cost",
			strategy:       GoalPriceFirst,
			expectedWeight: 0.6,
			weightField:    "cost",
		},
		{
			name:           "reliability_first should prioritize reliability",
			strategy:       GoalReliabilityFirst,
			expectedWeight: 0.5,
			weightField:    "reliability",
		},
		{
			name:           "security_first should prioritize security",
			strategy:       GoalSecurityFirst,
			expectedWeight: 0.5,
			weightField:    "security",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			weights, err := engine.GetStrategyWeights(tt.strategy)
			assert.NoError(t, err)
			assert.NotNil(t, weights)

			switch tt.weightField {
			case "latency":
				assert.Equal(t, tt.expectedWeight, weights.LatencyWeight)
			case "cost":
				assert.Equal(t, tt.expectedWeight, weights.CostWeight)
			case "reliability":
				assert.Equal(t, tt.expectedWeight, weights.ReliabilityWeight)
			case "security":
				assert.Equal(t, tt.expectedWeight, weights.SecurityWeight)
			}
		})
	}
}

func TestRoutingStrategyEngine_ValidateConstraints(t *testing.T) {
	engine := NewRoutingStrategyEngine()

	t.Run("should validate success rate constraint", func(t *testing.T) {
		goal := &StrategyOutput{
			Constraints: StrategyConstraints{
				MinSuccessRate: 0.95,
			},
		}

		candidate := &RoutingCandidateV2{
			SuccessRate: 0.98,
		}

		valid := engine.ValidateConstraints(goal, candidate)
		assert.True(t, valid)
	})

	t.Run("should reject candidate with low success rate", func(t *testing.T) {
		goal := &StrategyOutput{
			Constraints: StrategyConstraints{
				MinSuccessRate: 0.95,
			},
		}

		candidate := &RoutingCandidateV2{
			SuccessRate: 0.90,
		}

		valid := engine.ValidateConstraints(goal, candidate)
		assert.False(t, valid)
	})

	t.Run("should validate region constraint", func(t *testing.T) {
		goal := &StrategyOutput{
			Constraints: StrategyConstraints{
				RequiredRegions: []string{"domestic"},
			},
		}

		candidate := &RoutingCandidateV2{
			Region: "domestic",
		}

		valid := engine.ValidateConstraints(goal, candidate)
		assert.True(t, valid)
	})

	t.Run("should reject candidate from excluded region", func(t *testing.T) {
		goal := &StrategyOutput{
			Constraints: StrategyConstraints{
				RequiredRegions: []string{"domestic"},
			},
		}

		candidate := &RoutingCandidateV2{
			Region: "overseas",
		}

		valid := engine.ValidateConstraints(goal, candidate)
		assert.False(t, valid)
	})

	t.Run("should validate security level constraint", func(t *testing.T) {
		goal := &StrategyOutput{
			Constraints: StrategyConstraints{
				MinSecurityLevel: "high",
			},
		}

		candidate := &RoutingCandidateV2{
			SecurityLevel: "high",
		}

		valid := engine.ValidateConstraints(goal, candidate)
		assert.True(t, valid)
	})

	t.Run("should reject candidate with low security level", func(t *testing.T) {
		goal := &StrategyOutput{
			Constraints: StrategyConstraints{
				MinSecurityLevel: "high",
			},
		}

		candidate := &RoutingCandidateV2{
			SecurityLevel: "standard",
		}

		valid := engine.ValidateConstraints(goal, candidate)
		assert.False(t, valid)
	})
}
