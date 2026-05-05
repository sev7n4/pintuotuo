package handlers

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pintuotuo/backend/models"
	"github.com/pintuotuo/backend/services"
)

func TestResolveRouteMode(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   *models.MerchantAPIKey
		expected string
	}{
		{
			name:     "nil api key returns direct",
			apiKey:   nil,
			expected: routeModeDirect,
		},
		{
			name: "empty route mode returns direct",
			apiKey: &models.MerchantAPIKey{
				RouteMode: "",
			},
			expected: routeModeDirect,
		},
		{
			name: "auto route mode returns direct",
			apiKey: &models.MerchantAPIKey{
				RouteMode: "auto",
			},
			expected: routeModeDirect,
		},
		{
			name: "direct route mode returns direct",
			apiKey: &models.MerchantAPIKey{
				RouteMode: "direct",
			},
			expected: routeModeDirect,
		},
		{
			name: "litellm route mode returns litellm",
			apiKey: &models.MerchantAPIKey{
				RouteMode: "litellm",
			},
			expected: routeModeLitellm,
		},
		{
			name: "proxy route mode returns proxy",
			apiKey: &models.MerchantAPIKey{
				RouteMode: "proxy",
			},
			expected: routeModeProxy,
		},
		{
			name: "uppercase route mode is normalized",
			apiKey: &models.MerchantAPIKey{
				RouteMode: "LITELLM",
			},
			expected: routeModeLitellm,
		},
		{
			name: "whitespace route mode is trimmed",
			apiKey: &models.MerchantAPIKey{
				RouteMode: "  direct  ",
			},
			expected: routeModeDirect,
		},
		{
			name: "unknown route mode returns direct",
			apiKey: &models.MerchantAPIKey{
				RouteMode: "unknown",
			},
			expected: routeModeDirect,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolveRouteModeWithProvider(tt.apiKey, "")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResolveRouteModeWithProvider(t *testing.T) {
	tests := []struct {
		name           string
		apiKey         *models.MerchantAPIKey
		providerRegion string
		envHTTPSProxy  string
		envLitellmURL  string
		expected       string
	}{
		{
			name:           "nil api key returns direct",
			apiKey:         nil,
			providerRegion: "overseas",
			envHTTPSProxy:  "http://proxy:7890",
			expected:       routeModeDirect,
		},
		{
			name: "explicit direct not affected by proxy",
			apiKey: &models.MerchantAPIKey{
				RouteMode: "direct",
			},
			providerRegion: "overseas",
			envHTTPSProxy:  "http://proxy:7890",
			expected:       routeModeDirect,
		},
		{
			name: "explicit litellm not affected by proxy",
			apiKey: &models.MerchantAPIKey{
				RouteMode: "litellm",
			},
			providerRegion: "overseas",
			envHTTPSProxy:  "http://proxy:7890",
			expected:       routeModeLitellm,
		},
		{
			name: "explicit proxy not affected",
			apiKey: &models.MerchantAPIKey{
				RouteMode: "proxy",
			},
			providerRegion: "domestic",
			envHTTPSProxy:  "",
			expected:       routeModeProxy,
		},
		{
			name: "auto + overseas + HTTPS_PROXY returns proxy",
			apiKey: &models.MerchantAPIKey{
				RouteMode: "auto",
			},
			providerRegion: "overseas",
			envHTTPSProxy:  "http://host.docker.internal:7890",
			expected:       routeModeProxy,
		},
		{
			name: "auto + overseas + https_proxy lowercase returns proxy",
			apiKey: &models.MerchantAPIKey{
				RouteMode: "auto",
			},
			providerRegion: "overseas",
			envHTTPSProxy:  "",
			envLitellmURL:  "",
			expected:       routeModeProxy,
		},
		{
			name: "auto + overseas + no HTTPS_PROXY + LiteLLM URL returns litellm",
			apiKey: &models.MerchantAPIKey{
				RouteMode: "auto",
			},
			providerRegion: "overseas",
			envHTTPSProxy:  "",
			envLitellmURL:  "http://litellm:4000",
			expected:       routeModeLitellm,
		},
		{
			name: "auto + overseas + no proxy no litellm returns direct",
			apiKey: &models.MerchantAPIKey{
				RouteMode: "auto",
			},
			providerRegion: "overseas",
			envHTTPSProxy:  "",
			envLitellmURL:  "",
			expected:       routeModeDirect,
		},
		{
			name: "auto + domestic returns direct regardless of HTTPS_PROXY",
			apiKey: &models.MerchantAPIKey{
				RouteMode: "auto",
			},
			providerRegion: "domestic",
			envHTTPSProxy:  "http://host.docker.internal:7890",
			expected:       routeModeDirect,
		},
		{
			name: "empty route mode + overseas + HTTPS_PROXY returns proxy",
			apiKey: &models.MerchantAPIKey{
				RouteMode: "",
			},
			providerRegion: "overseas",
			envHTTPSProxy:  "http://host.docker.internal:7890",
			expected:       routeModeProxy,
		},
		{
			name: "auto + empty region returns direct",
			apiKey: &models.MerchantAPIKey{
				RouteMode: "auto",
			},
			providerRegion: "",
			envHTTPSProxy:  "http://host.docker.internal:7890",
			expected:       routeModeDirect,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origHTTPS := os.Getenv("HTTPS_PROXY")
			origHTTPSLower := os.Getenv("https_proxy")
			origLitellm := os.Getenv("LLM_GATEWAY_LITELLM_URL")
			defer func() {
				os.Setenv("HTTPS_PROXY", origHTTPS)
				os.Setenv("https_proxy", origHTTPSLower)
				os.Setenv("LLM_GATEWAY_LITELLM_URL", origLitellm)
			}()

			os.Setenv("HTTPS_PROXY", tt.envHTTPSProxy)
			os.Setenv("https_proxy", "")
			os.Setenv("LLM_GATEWAY_LITELLM_URL", tt.envLitellmURL)

			if tt.name == "auto + overseas + https_proxy lowercase returns proxy" {
				os.Setenv("HTTPS_PROXY", "")
				os.Setenv("https_proxy", "http://host.docker.internal:7890")
			}

			result := resolveRouteModeWithProvider(tt.apiKey, tt.providerRegion)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestApplyAPIKeyRouteConfig(t *testing.T) {
	tests := []struct {
		name            string
		cfg             providerRuntimeConfig
		apiKey          *models.MerchantAPIKey
		expectedBaseURL string
	}{
		{
			name: "direct mode with custom endpoint",
			cfg: providerRuntimeConfig{
				Code:       "openai",
				APIBaseURL: "https://api.openai.com/v1",
			},
			apiKey: &models.MerchantAPIKey{
				RouteMode:   "direct",
				EndpointURL: "https://custom.api.com/v1",
			},
			expectedBaseURL: "https://custom.api.com/v1",
		},
		{
			name: "litellm mode without custom endpoint uses env",
			cfg: providerRuntimeConfig{
				Code:       "openai",
				APIBaseURL: "https://api.openai.com/v1",
			},
			apiKey: &models.MerchantAPIKey{
				RouteMode: "litellm",
			},
			expectedBaseURL: "http://localhost:4000/v1",
		},
		{
			name: "auto mode defaults to direct",
			cfg: providerRuntimeConfig{
				Code:       "openai",
				APIBaseURL: "https://api.openai.com/v1",
			},
			apiKey: &models.MerchantAPIKey{
				RouteMode: "auto",
			},
			expectedBaseURL: "https://api.openai.com/v1",
		},
		{
			name: "nil api key uses provider base",
			cfg: providerRuntimeConfig{
				Code:       "openai",
				APIBaseURL: "https://api.openai.com/v1",
			},
			apiKey:          nil,
			expectedBaseURL: "https://api.openai.com/v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "litellm mode without custom endpoint uses env" {
				t.Skip("requires LLM_GATEWAY_LITELLM_URL env var")
			}
			result := applyAPIKeyRouteConfig(tt.cfg, tt.apiKey)
			assert.Equal(t, tt.expectedBaseURL, result.APIBaseURL)
		})
	}
}

func TestResolveAuthTokenFromRouteMode(t *testing.T) {
	tests := []struct {
		name          string
		routeMode     string
		fallbackToken string
		expected      string
	}{
		{
			name:          "non-litellm mode uses fallback token",
			routeMode:     "direct",
			fallbackToken: "sk-test",
			expected:      "sk-test",
		},
		{
			name:          "empty route mode uses fallback token",
			routeMode:     "",
			fallbackToken: "sk-test",
			expected:      "sk-test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolveAuthTokenFromRouteMode(tt.routeMode, tt.fallbackToken)
			assert.Equal(t, tt.expected, result)
		})
	}
}

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
