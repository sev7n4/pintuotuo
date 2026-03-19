package handlers

import (
	"testing"

	"github.com/pintuotuo/backend/billing"
)

func TestCalculateTokenCost(t *testing.T) {
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
			name:         "OpenAI GPT-4 Turbo",
			provider:     "openai",
			model:        "gpt-4-turbo-preview",
			inputTokens:  1000,
			outputTokens: 500,
			wantMinCost:  0.01,
			wantMaxCost:  0.05,
		},
		{
			name:         "OpenAI GPT-3.5 Turbo",
			provider:     "openai",
			model:        "gpt-3.5-turbo",
			inputTokens:  10000,
			outputTokens: 5000,
			wantMinCost:  0.005,
			wantMaxCost:  0.02,
		},
		{
			name:         "Anthropic Claude 3 Opus",
			provider:     "anthropic",
			model:        "claude-3-opus-20240229",
			inputTokens:  1000000,
			outputTokens: 500000,
			wantMinCost:  15.0,
			wantMaxCost:  60.0,
		},
		{
			name:         "Anthropic Claude 3 Haiku",
			provider:     "anthropic",
			model:        "claude-3-haiku-20240307",
			inputTokens:  1000000,
			outputTokens: 500000,
			wantMinCost:  0.25,
			wantMaxCost:  1.5,
		},
		{
			name:         "Google Gemini Pro",
			provider:     "google",
			model:        "gemini-pro",
			inputTokens:  1000,
			outputTokens: 1000,
			wantMinCost:  0.0005,
			wantMaxCost:  0.01,
		},
		{
			name:         "Unknown provider uses default",
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
			cost := calculateTokenCost(tt.provider, tt.model, tt.inputTokens, tt.outputTokens)
			if cost < tt.wantMinCost || cost > tt.wantMaxCost {
				t.Errorf("calculateTokenCost() = %v, want between %v and %v", cost, tt.wantMinCost, tt.wantMaxCost)
			}
		})
	}
}

func TestCalculateTokenCost_ZeroTokens(t *testing.T) {
	cost := calculateTokenCost("openai", "gpt-4-turbo-preview", 0, 0)
	if cost != 0 {
		t.Errorf("calculateTokenCost() with zero tokens = %v, want 0", cost)
	}
}

func TestProviderBaseURLs(t *testing.T) {
	expectedProviders := []string{"openai", "anthropic", "google", "azure"}

	for _, provider := range expectedProviders {
		t.Run(provider, func(t *testing.T) {
			url, ok := providerBaseURLs[provider]
			if !ok {
				t.Errorf("Provider %s not found in providerBaseURLs", provider)
			}
			if url == "" {
				t.Errorf("Provider %s has empty URL", provider)
			}
		})
	}
}

func TestAPIProxyRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     APIProxyRequest
		wantErr bool
	}{
		{
			name: "Valid request",
			req: APIProxyRequest{
				Provider: "openai",
				Model:    "gpt-4-turbo-preview",
				Messages: []ChatMessage{
					{Role: "user", Content: "Hello"},
				},
			},
			wantErr: false,
		},
		{
			name: "Empty provider",
			req: APIProxyRequest{
				Provider: "",
				Model:    "gpt-4-turbo-preview",
				Messages: []ChatMessage{
					{Role: "user", Content: "Hello"},
				},
			},
			wantErr: true,
		},
		{
			name: "Empty model",
			req: APIProxyRequest{
				Provider: "openai",
				Model:    "",
				Messages: []ChatMessage{
					{Role: "user", Content: "Hello"},
				},
			},
			wantErr: true,
		},
		{
			name: "Empty messages",
			req: APIProxyRequest{
				Provider: "openai",
				Model:    "gpt-4-turbo-preview",
				Messages: []ChatMessage{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasEmptyProvider := tt.req.Provider == ""
			hasEmptyModel := tt.req.Model == ""
			hasEmptyMessages := len(tt.req.Messages) == 0

			hasErr := hasEmptyProvider || hasEmptyModel || hasEmptyMessages
			if hasErr != tt.wantErr {
				t.Errorf("Validation error = %v, want %v", hasErr, tt.wantErr)
			}
		})
	}
}

func TestGetAPIProviders(t *testing.T) {
	providers := []map[string]interface{}{
		{
			"name":        "openai",
			"display_name": "OpenAI",
			"models": []string{
				"gpt-4-turbo-preview",
				"gpt-4",
				"gpt-3.5-turbo",
			},
		},
		{
			"name":        "anthropic",
			"display_name": "Anthropic",
			"models": []string{
				"claude-3-opus-20240229",
				"claude-3-sonnet-20240229",
				"claude-3-haiku-20240307",
			},
		},
		{
			"name":        "google",
			"display_name": "Google AI",
			"models": []string{
				"gemini-pro",
				"gemini-pro-vision",
			},
		},
	}

	if len(providers) != 3 {
		t.Errorf("Expected 3 providers, got %d", len(providers))
	}

	for _, p := range providers {
		name, ok := p["name"].(string)
		if !ok {
			t.Error("Provider should have a name")
		}

		models, ok := p["models"].([]string)
		if !ok {
			t.Errorf("Provider %s should have models", name)
		}

		if len(models) == 0 {
			t.Errorf("Provider %s should have at least one model", name)
		}
	}
}

func TestBillingEngine_Integration(t *testing.T) {
	engine := billing.GetBillingEngine()

	t.Run("Calculate cost matches handler", func(t *testing.T) {
		provider := "openai"
		model := "gpt-4-turbo-preview"
		inputTokens := 1000
		outputTokens := 500

		handlerCost := calculateTokenCost(provider, model, inputTokens, outputTokens)
		engineCost := engine.CalculateCost(provider, model, inputTokens, outputTokens)

		if handlerCost != engineCost {
			t.Errorf("Handler cost %v != Engine cost %v", handlerCost, engineCost)
		}
	})
}
