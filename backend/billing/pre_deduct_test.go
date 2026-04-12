package billing

import (
	"testing"
)

func TestBillingEngine_EstimateTokenUsage(t *testing.T) {
	engine := GetBillingEngine()

	tests := []struct {
		name        string
		inputTokens int
		config      *PreDeductConfig
		wantMin     int64
		wantMax     int64
	}{
		{
			name:        "Default config with 1000 input tokens",
			inputTokens: 1000,
			config:      nil,
			wantMin:     2000,
			wantMax:     2000,
		},
		{
			name:        "Custom multiplier 3x with 500 input tokens",
			inputTokens: 500,
			config:      &PreDeductConfig{Multiplier: 3, MaxMultiplier: 10},
			wantMin:     1500,
			wantMax:     1500,
		},
		{
			name:        "High multiplier capped by max",
			inputTokens: 1000,
			config:      &PreDeductConfig{Multiplier: 15, MaxMultiplier: 5},
			wantMin:     5000,
			wantMax:     5000,
		},
		{
			name:        "Zero multiplier uses default",
			inputTokens: 1000,
			config:      &PreDeductConfig{Multiplier: 0, MaxMultiplier: 10},
			wantMin:     2000,
			wantMax:     2000,
		},
		{
			name:        "Negative multiplier uses default",
			inputTokens: 1000,
			config:      &PreDeductConfig{Multiplier: -1, MaxMultiplier: 10},
			wantMin:     2000,
			wantMax:     2000,
		},
		{
			name:        "Large input tokens",
			inputTokens: 100000,
			config:      &PreDeductConfig{Multiplier: 2, MaxMultiplier: 10},
			wantMin:     200000,
			wantMax:     200000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := engine.EstimateTokenUsage(tt.inputTokens, tt.config)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("EstimateTokenUsage() = %v, want between %v and %v", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestBillingEngine_EstimateTokenUsage_EdgeCases(t *testing.T) {
	engine := GetBillingEngine()

	t.Run("Zero input tokens", func(t *testing.T) {
		got := engine.EstimateTokenUsage(0, nil)
		if got != 0 {
			t.Errorf("EstimateTokenUsage() with zero input = %v, want 0", got)
		}
	})

	t.Run("Estimate respects max multiplier cap", func(t *testing.T) {
		config := &PreDeductConfig{Multiplier: 100, MaxMultiplier: 3}
		got := engine.EstimateTokenUsage(1000, config)
		want := int64(3000)
		if got != want {
			t.Errorf("EstimateTokenUsage() with max cap = %v, want %v", got, want)
		}
	})
}

func TestPreDeductConfig_DefaultValues(t *testing.T) {
	t.Run("Default multiplier is 2", func(t *testing.T) {
		config := &PreDeductConfig{}
		if config.Multiplier != 0 {
			t.Errorf("Default Multiplier should be 0 (unset), got %v", config.Multiplier)
		}
	})
}

func TestBillingEngine_GetPreDeductConfig_Default(t *testing.T) {
	engine := GetBillingEngine()

	t.Run("Returns default config when no overrides", func(t *testing.T) {
		config := engine.GetPreDeductConfig(0, 0, "")
		if config == nil {
			t.Fatal("GetPreDeductConfig() returned nil")
		}
		if config.Multiplier != 2 {
			t.Errorf("Default Multiplier = %v, want 2", config.Multiplier)
		}
		if config.MaxMultiplier != 10 {
			t.Errorf("Default MaxMultiplier = %v, want 10", config.MaxMultiplier)
		}
	})
}

func TestBillingEngine_CalculateCostForSettlement(t *testing.T) {
	engine := GetBillingEngine()

	tests := []struct {
		name         string
		provider     string
		model        string
		inputTokens  int
		outputTokens int
		wantMinCost  float64
		wantMaxCost  float64
	}{
		{
			name:         "GPT-4 Turbo settlement cost",
			provider:     "openai",
			model:        "gpt-4-turbo-preview",
			inputTokens:  1000,
			outputTokens: 500,
			wantMinCost:  0.01,
			wantMaxCost:  0.03,
		},
		{
			name:         "Claude 3 Opus settlement cost",
			provider:     "anthropic",
			model:        "claude-3-opus-20240229",
			inputTokens:  10000,
			outputTokens: 5000,
			wantMinCost:  0.1,
			wantMaxCost:  1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost := engine.CalculateCost(tt.provider, tt.model, tt.inputTokens, tt.outputTokens)
			if cost < tt.wantMinCost || cost > tt.wantMaxCost {
				t.Errorf("CalculateCost() = %v, want between %v and %v", cost, tt.wantMinCost, tt.wantMaxCost)
			}
		})
	}
}
