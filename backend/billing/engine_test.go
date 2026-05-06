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
		// 元/1K tokens（与 SPU/定价缓存口径一致）
		if pricing.InputPrice != 0.01 {
			t.Errorf("GPT-4 Turbo input price = %v, want 0.01", pricing.InputPrice)
		}
		if pricing.OutputPrice != 0.03 {
			t.Errorf("GPT-4 Turbo output price = %v, want 0.03", pricing.OutputPrice)
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

func TestBillingUnitConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant BillingUnit
		expected string
	}{
		{
			name:     "BillingUnitToken",
			constant: BillingUnitToken,
			expected: "token",
		},
		{
			name:     "BillingUnitImage",
			constant: BillingUnitImage,
			expected: "image",
		},
		{
			name:     "BillingUnitSecond",
			constant: BillingUnitSecond,
			expected: "second",
		},
		{
			name:     "BillingUnitCharacter",
			constant: BillingUnitCharacter,
			expected: "character",
		},
		{
			name:     "BillingUnitRequest",
			constant: BillingUnitRequest,
			expected: "request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.constant) != tt.expected {
				t.Errorf("BillingUnit constant = %v, want %v", tt.constant, tt.expected)
			}
		})
	}
}

func TestBillingRequest_Structure(t *testing.T) {
	req := BillingRequest{
		UserID:       123,
		EndpointType: "chat_completions",
		ProviderCode: "openai",
		UnitType:     BillingUnitToken,
		Quantity:     1000,
		RequestID:    "req-123",
		Reason:       "API usage",
	}

	if req.UserID != 123 {
		t.Errorf("UserID = %v, want 123", req.UserID)
	}
	if req.EndpointType != "chat_completions" {
		t.Errorf("EndpointType = %v, want chat_completions", req.EndpointType)
	}
	if req.ProviderCode != "openai" {
		t.Errorf("ProviderCode = %v, want openai", req.ProviderCode)
	}
	if req.UnitType != BillingUnitToken {
		t.Errorf("UnitType = %v, want token", req.UnitType)
	}
	if req.Quantity != 1000 {
		t.Errorf("Quantity = %v, want 1000", req.Quantity)
	}
	if req.RequestID != "req-123" {
		t.Errorf("RequestID = %v, want req-123", req.RequestID)
	}
	if req.Reason != "API usage" {
		t.Errorf("Reason = %v, want API usage", req.Reason)
	}
}

func TestBillingEngine_GetUnitPrice(t *testing.T) {
	engine := GetBillingEngine()

	tests := []struct {
		name         string
		endpointType string
		providerCode string
		unitType     BillingUnit
		wantMinPrice float64
		wantMaxPrice float64
	}{
		{
			name:         "chat_completions token price",
			endpointType: "chat_completions",
			providerCode: "openai",
			unitType:     BillingUnitToken,
			wantMinPrice: 0.0001,
			wantMaxPrice: 0.01,
		},
		{
			name:         "embeddings token price",
			endpointType: "embeddings",
			providerCode: "openai",
			unitType:     BillingUnitToken,
			wantMinPrice: 0.00001,
			wantMaxPrice: 0.001,
		},
		{
			name:         "images_generations image price",
			endpointType: "images_generations",
			providerCode: "openai",
			unitType:     BillingUnitImage,
			wantMinPrice: 1.0,
			wantMaxPrice: 50.0,
		},
		{
			name:         "images_variations image price",
			endpointType: "images_variations",
			providerCode: "openai",
			unitType:     BillingUnitImage,
			wantMinPrice: 1.0,
			wantMaxPrice: 50.0,
		},
		{
			name:         "images_edits image price",
			endpointType: "images_edits",
			providerCode: "openai",
			unitType:     BillingUnitImage,
			wantMinPrice: 1.0,
			wantMaxPrice: 50.0,
		},
		{
			name:         "audio_speech character price",
			endpointType: "audio_speech",
			providerCode: "openai",
			unitType:     BillingUnitCharacter,
			wantMinPrice: 0.00001,
			wantMaxPrice: 0.001,
		},
		{
			name:         "audio_transcriptions second price",
			endpointType: "audio_transcriptions",
			providerCode: "openai",
			unitType:     BillingUnitSecond,
			wantMinPrice: 0.001,
			wantMaxPrice: 0.01,
		},
		{
			name:         "audio_translations second price",
			endpointType: "audio_translations",
			providerCode: "openai",
			unitType:     BillingUnitSecond,
			wantMinPrice: 0.001,
			wantMaxPrice: 0.01,
		},
		{
			name:         "moderations token price",
			endpointType: "moderations",
			providerCode: "openai",
			unitType:     BillingUnitToken,
			wantMinPrice: 0.0,
			wantMaxPrice: 0.0,
		},
		{
			name:         "responses token price",
			endpointType: "responses",
			providerCode: "openai",
			unitType:     BillingUnitToken,
			wantMinPrice: 0.0001,
			wantMaxPrice: 0.01,
		},
		{
			name:         "responses request price",
			endpointType: "responses",
			providerCode: "openai",
			unitType:     BillingUnitRequest,
			wantMinPrice: 0.001,
			wantMaxPrice: 1.0,
		},
		{
			name:         "responses image price",
			endpointType: "responses",
			providerCode: "openai",
			unitType:     BillingUnitImage,
			wantMinPrice: 1.0,
			wantMaxPrice: 50.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			price := engine.getUnitPrice(tt.endpointType, tt.providerCode, tt.unitType)
			if price < tt.wantMinPrice || price > tt.wantMaxPrice {
				t.Errorf("getUnitPrice() = %v, want between %v and %v", price, tt.wantMinPrice, tt.wantMaxPrice)
			}
		})
	}
}

func TestBillingEngine_PreDeductBalanceV2(t *testing.T) {
	engine := GetBillingEngine()

	t.Run("calculate amount from tokens", func(t *testing.T) {
		req := BillingRequest{
			UserID:       1,
			EndpointType: "chat_completions",
			ProviderCode: "openai",
			UnitType:     BillingUnitToken,
			Quantity:     1000,
			RequestID:    "test-req-1",
			Reason:       "test",
		}

		amount := engine.calculateAmountFromRequest(req)
		if amount <= 0 {
			t.Errorf("calculateAmountFromRequest() = %v, want > 0", amount)
		}
	})

	t.Run("calculate amount from images", func(t *testing.T) {
		req := BillingRequest{
			UserID:       1,
			EndpointType: "images_generations",
			ProviderCode: "openai",
			UnitType:     BillingUnitImage,
			Quantity:     1,
			RequestID:    "test-req-2",
			Reason:       "test",
		}

		amount := engine.calculateAmountFromRequest(req)
		if amount <= 0 {
			t.Errorf("calculateAmountFromRequest() = %v, want > 0", amount)
		}
	})

	t.Run("calculate amount from seconds", func(t *testing.T) {
		req := BillingRequest{
			UserID:       1,
			EndpointType: "audio_transcriptions",
			ProviderCode: "openai",
			UnitType:     BillingUnitSecond,
			Quantity:     60,
			RequestID:    "test-req-3",
			Reason:       "test",
		}

		amount := engine.calculateAmountFromRequest(req)
		if amount <= 0 {
			t.Errorf("calculateAmountFromRequest() = %v, want > 0", amount)
		}
	})

	t.Run("calculate amount from characters", func(t *testing.T) {
		req := BillingRequest{
			UserID:       1,
			EndpointType: "audio_speech",
			ProviderCode: "openai",
			UnitType:     BillingUnitCharacter,
			Quantity:     1000,
			RequestID:    "test-req-4",
			Reason:       "test",
		}

		amount := engine.calculateAmountFromRequest(req)
		if amount <= 0 {
			t.Errorf("calculateAmountFromRequest() = %v, want > 0", amount)
		}
	})

	t.Run("calculate amount from requests", func(t *testing.T) {
		req := BillingRequest{
			UserID:       1,
			EndpointType: "moderations",
			ProviderCode: "openai",
			UnitType:     BillingUnitToken,
			Quantity:     10,
			RequestID:    "test-req-5",
			Reason:       "test",
		}

		amount := engine.calculateAmountFromRequest(req)
		if amount != 0 {
			t.Errorf("calculateAmountFromRequest() = %v, want 0 for moderations", amount)
		}
	})
}
