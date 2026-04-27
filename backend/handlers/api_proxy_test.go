package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pintuotuo/backend/services"
)

func TestApplyGatewayOverrideWithUnifiedRouter(t *testing.T) {
	router := services.NewUnifiedRouter(nil)

	tests := []struct {
		name           string
		provider       string
		merchantRegion string
		merchantType   string
		providerRegion string
		expectedMode   string
	}{
		{
			name:           "domestic user accessing overseas provider",
			provider:       "openai",
			merchantRegion: "domestic",
			merchantType:   "standard",
			providerRegion: "overseas",
			expectedMode:   "litellm",
		},
		{
			name:           "overseas user accessing overseas provider",
			provider:       "openai",
			merchantRegion: "overseas",
			merchantType:   "standard",
			providerRegion: "overseas",
			expectedMode:   "direct",
		},
		{
			name:           "domestic user accessing domestic provider",
			provider:       "deepseek",
			merchantRegion: "domestic",
			merchantType:   "standard",
			providerRegion: "domestic",
			expectedMode:   "direct",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			providerConfig := &services.ProviderConfig{
				Code:           tt.provider,
				ProviderRegion: tt.providerRegion,
				RouteStrategy: map[string]interface{}{
					"default_mode": "auto",
				},
				Endpoints: map[string]interface{}{
					"direct": map[string]interface{}{
						"domestic": "https://domestic.example.com",
						"overseas": "https://overseas.example.com",
					},
					"litellm": map[string]interface{}{
						"domestic": "http://litellm:4000/v1",
						"overseas": "http://litellm:4000/v1",
					},
				},
			}

			merchantConfig := &services.MerchantConfig{
				ID:     1,
				Type:   tt.merchantType,
				Region: tt.merchantRegion,
			}

			decision, err := router.DecideRoute(nil, providerConfig, merchantConfig)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if decision.Mode != tt.expectedMode {
				t.Errorf("expected mode to be '%s', got '%s'", tt.expectedMode, decision.Mode)
			}
		})
	}
}

func TestRouteDecisionIntegration(t *testing.T) {
	router := services.NewUnifiedRouter(nil)

	providerConfig := &services.ProviderConfig{
		Code:           "openai",
		ProviderRegion: "overseas",
		RouteStrategy: map[string]interface{}{
			"default_mode": "auto",
			"domestic_users": map[string]interface{}{
				"mode":          "litellm",
				"fallback_mode": "proxy",
			},
		},
		Endpoints: map[string]interface{}{
			"litellm": map[string]interface{}{
				"domestic": "http://litellm-overseas:4000/v1",
			},
			"proxy": map[string]interface{}{
				"domestic": "https://gaap.example.com",
				"gaap":     "https://gaap.example.com",
			},
		},
	}

	merchantConfig := &services.MerchantConfig{
		ID:     1,
		Type:   "standard",
		Region: "domestic",
	}

	decision, err := router.DecideRoute(nil, providerConfig, merchantConfig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if decision.Mode != "litellm" {
		t.Errorf("expected mode to be 'litellm', got '%s'", decision.Mode)
	}

	if decision.Endpoint == "" {
		t.Error("expected endpoint to be set")
	}

	if decision.FallbackMode != "proxy" {
		t.Errorf("expected fallback mode to be 'proxy', got '%s'", decision.FallbackMode)
	}

	if decision.FallbackEndpoint == "" {
		t.Error("expected fallback endpoint to be set")
	}
}

func TestProviderRuntimeConfig_NewFields(t *testing.T) {
	cfg := providerRuntimeConfig{
		Code:           "openai",
		Name:           "OpenAI",
		APIBaseURL:     "https://api.openai.com/v1",
		APIFormat:      "openai",
		ProviderRegion: "overseas",
		RouteStrategy: map[string]interface{}{
			"domestic_users": map[string]interface{}{"mode": "litellm"},
			"overseas_users": map[string]interface{}{"mode": "direct"},
		},
		Endpoints: map[string]interface{}{
			"direct": map[string]interface{}{
				"overseas": "https://api.openai.com/v1",
			},
			"litellm": map[string]interface{}{
				"domestic": "http://litellm:4000/v1",
			},
		},
	}

	assert.Equal(t, "openai", cfg.Code)
	assert.Equal(t, "overseas", cfg.ProviderRegion)
	assert.NotNil(t, cfg.RouteStrategy)
	assert.NotNil(t, cfg.Endpoints)

	domesticUsers, ok := cfg.RouteStrategy["domestic_users"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "litellm", domesticUsers["mode"])
}

func TestProviderRuntimeConfig_EmptyFields(t *testing.T) {
	cfg := providerRuntimeConfig{
		Code:       "openai",
		APIBaseURL: "https://api.openai.com/v1",
	}

	assert.Equal(t, "", cfg.ProviderRegion)
	assert.Nil(t, cfg.RouteStrategy)
	assert.Nil(t, cfg.Endpoints)
}


