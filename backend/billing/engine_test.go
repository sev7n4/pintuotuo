package billing

import (
	"testing"
)

func TestBillingEngine_GetPricing(t *testing.T) {
	engine := GetBillingEngine()

	tests := []struct {
		name      string
		provider  string
		model     string
		wantFound bool
	}{
		{
			name:      "OpenAI GPT-4 Turbo",
			provider:  "openai",
			model:     "gpt-4-turbo-preview",
			wantFound: true,
		},
		{
			name:      "OpenAI GPT-3.5 Turbo",
			provider:  "openai",
			model:     "gpt-3.5-turbo",
			wantFound: true,
		},
		{
			name:      "Anthropic Claude 3 Opus",
			provider:  "anthropic",
			model:     "claude-3-opus-20240229",
			wantFound: true,
		},
		{
			name:      "Google Gemini Pro",
			provider:  "google",
			model:     "gemini-pro",
			wantFound: true,
		},
		{
			name:      "Unknown Model",
			provider:  "unknown",
			model:     "unknown-model",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pricing, found := engine.GetPricing(tt.provider, tt.model)
			if found != tt.wantFound {
				t.Errorf("GetPricing() found = %v, want %v", found, tt.wantFound)
			}
			if found && pricing.Provider != tt.provider {
				t.Errorf("GetPricing() provider = %v, want %v", pricing.Provider, tt.provider)
			}
		})
	}
}

func TestBillingEngine_CalculateCost(t *testing.T) {
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
			name:         "GPT-4 Turbo 1K tokens",
			provider:     "openai",
			model:        "gpt-4-turbo-preview",
			inputTokens:  1000,
			outputTokens: 500,
			wantMinCost:  0.01,
			wantMaxCost:  0.03,
		},
		{
			name:         "GPT-3.5 Turbo 10K tokens",
			provider:     "openai",
			model:        "gpt-3.5-turbo",
			inputTokens:  10000,
			outputTokens: 5000,
			wantMinCost:  0.005,
			wantMaxCost:  0.02,
		},
		{
			name:         "Claude 3 Opus 1M tokens",
			provider:     "anthropic",
			model:        "claude-3-opus-20240229",
			inputTokens:  1000000,
			outputTokens: 500000,
			wantMinCost:  15.0,
			wantMaxCost:  60.0,
		},
		{
			name:         "Unknown model uses default pricing",
			provider:     "unknown",
			model:        "unknown-model",
			inputTokens:  1000,
			outputTokens: 1000,
			wantMinCost:  0.001,
			wantMaxCost:  0.01,
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

func TestBillingEngine_CalculateCost_EdgeCases(t *testing.T) {
	engine := GetBillingEngine()

	t.Run("Zero tokens", func(t *testing.T) {
		cost := engine.CalculateCost("openai", "gpt-4-turbo-preview", 0, 0)
		if cost != 0 {
			t.Errorf("CalculateCost() with zero tokens = %v, want 0", cost)
		}
	})

	t.Run("Large token count", func(t *testing.T) {
		cost := engine.CalculateCost("openai", "gpt-4-turbo-preview", 10000000, 5000000)
		if cost <= 0 {
			t.Errorf("CalculateCost() with large tokens should be positive, got %v", cost)
		}
	})
}

func TestPricingTier_Values(t *testing.T) {
	engine := GetBillingEngine()

	t.Run("GPT-4 Turbo pricing is correct", func(t *testing.T) {
		pricing, found := engine.GetPricing("openai", "gpt-4-turbo-preview")
		if !found {
			t.Fatal("GPT-4 Turbo pricing not found")
		}
		if pricing.InputPrice != 10 {
			t.Errorf("GPT-4 Turbo input price = %v, want 10", pricing.InputPrice)
		}
		if pricing.OutputPrice != 30 {
			t.Errorf("GPT-4 Turbo output price = %v, want 30", pricing.OutputPrice)
		}
	})

	t.Run("Claude 3 Haiku is cheapest", func(t *testing.T) {
		haiku, found := engine.GetPricing("anthropic", "claude-3-haiku-20240307")
		if !found {
			t.Fatal("Claude 3 Haiku pricing not found")
		}
		opus, _ := engine.GetPricing("anthropic", "claude-3-opus-20240229")
		if haiku.InputPrice >= opus.InputPrice {
			t.Error("Claude 3 Haiku should be cheaper than Opus")
		}
	})
}
