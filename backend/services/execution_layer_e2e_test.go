package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2E_ConfigDrivenRouting(t *testing.T) {
	t.Run("domestic user should route to LiteLLM", func(t *testing.T) {
		router := NewUnifiedRouter(nil)

		providerConfig := &ProviderConfig{
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
				"direct": map[string]interface{}{
					"overseas": "https://api.openai.com/v1",
				},
				"litellm": map[string]interface{}{
					"domestic": "http://litellm:4000/v1",
				},
			},
		}

		merchantConfig := &MerchantConfig{
			ID:     1,
			Type:   "standard",
			Region: "domestic",
		}

		decision, err := router.DecideRoute(context.Background(), providerConfig, merchantConfig)
		require.NoError(t, err)
		assert.Equal(t, "litellm", decision.Mode, "domestic user should route to LiteLLM")
		assert.NotEmpty(t, decision.Endpoint, "endpoint should be set")
	})

	t.Run("overseas user should route to Direct", func(t *testing.T) {
		router := NewUnifiedRouter(nil)

		providerConfig := &ProviderConfig{
			Code:           "openai",
			ProviderRegion: "overseas",
			RouteStrategy: map[string]interface{}{
				"default_mode": "auto",
				"overseas_users": map[string]interface{}{
					"mode": "direct",
				},
			},
			Endpoints: map[string]interface{}{
				"direct": map[string]interface{}{
					"overseas": "https://api.openai.com/v1",
				},
			},
		}

		merchantConfig := &MerchantConfig{
			ID:     1,
			Type:   "standard",
			Region: "overseas",
		}

		decision, err := router.DecideRoute(context.Background(), providerConfig, merchantConfig)
		require.NoError(t, err)
		assert.Equal(t, "direct", decision.Mode, "overseas user should route to Direct")
		assert.NotEmpty(t, decision.Endpoint, "endpoint should be set")
	})

	t.Run("enterprise user should route to LiteLLM with Fallback", func(t *testing.T) {
		router := NewUnifiedRouter(nil)

		providerConfig := &ProviderConfig{
			Code:           "openai",
			ProviderRegion: "overseas",
			RouteStrategy: map[string]interface{}{
				"default_mode": "auto",
				"enterprise_users": map[string]interface{}{
					"mode":          "litellm",
					"fallback_mode": "proxy",
				},
			},
			Endpoints: map[string]interface{}{
				"direct": map[string]interface{}{
					"overseas": "https://api.openai.com/v1",
				},
				"litellm": map[string]interface{}{
					"domestic": "http://litellm:4000/v1",
				},
				"proxy": map[string]interface{}{
					"domestic": "https://gaap.example.com",
				},
			},
		}

		merchantConfig := &MerchantConfig{
			ID:     1,
			Type:   "enterprise",
			Region: "domestic",
		}

		decision, err := router.DecideRoute(context.Background(), providerConfig, merchantConfig)
		require.NoError(t, err)
		assert.Equal(t, "litellm", decision.Mode, "enterprise user should route to LiteLLM")
		assert.Equal(t, "proxy", decision.FallbackMode, "enterprise user should have proxy fallback")
		assert.NotEmpty(t, decision.FallbackEndpoint, "fallback endpoint should be set")
	})

	t.Run("missing config should fallback to env-driven routing", func(t *testing.T) {
		router := NewUnifiedRouter(nil)

		providerConfig := &ProviderConfig{
			Code:           "openai",
			ProviderRegion: "overseas",
			RouteStrategy:  nil,
			Endpoints:      nil,
		}

		merchantConfig := &MerchantConfig{
			ID:     1,
			Type:   "standard",
			Region: "domestic",
		}

		decision, err := router.DecideRoute(context.Background(), providerConfig, merchantConfig)
		require.NoError(t, err)
		assert.NotEmpty(t, decision.Mode, "should have a mode even with missing config")
	})
}

func TestE2E_DegradationMechanism(t *testing.T) {
	t.Run("LiteLLM failure should fallback to Proxy", func(t *testing.T) {
		router := NewUnifiedRouter(nil)

		providerConfig := &ProviderConfig{
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
					"domestic": "http://litellm:4000/v1",
				},
				"proxy": map[string]interface{}{
					"domestic": "https://gaap.example.com",
				},
			},
		}

		merchantConfig := &MerchantConfig{
			ID:     1,
			Type:   "standard",
			Region: "domestic",
		}

		decision, err := router.DecideRoute(context.Background(), providerConfig, merchantConfig)
		require.NoError(t, err)
		assert.Equal(t, "litellm", decision.Mode)
		assert.Equal(t, "proxy", decision.FallbackMode, "should have proxy as fallback")
		assert.NotEmpty(t, decision.FallbackEndpoint, "fallback endpoint should be set")
	})

	t.Run("config error should fallback to env variable", func(t *testing.T) {
		router := NewUnifiedRouter(nil)

		providerConfig := &ProviderConfig{
			Code:           "openai",
			ProviderRegion: "overseas",
			RouteStrategy: map[string]interface{}{
				"invalid_key": "invalid_value",
			},
			Endpoints: map[string]interface{}{},
		}

		merchantConfig := &MerchantConfig{
			ID:     1,
			Type:   "standard",
			Region: "domestic",
		}

		decision, err := router.DecideRoute(context.Background(), providerConfig, merchantConfig)
		require.NoError(t, err)
		assert.NotEmpty(t, decision.Mode, "should have a mode even with invalid config")
	})

	t.Run("missing env should fallback to Direct", func(t *testing.T) {
		router := NewUnifiedRouter(nil)

		providerConfig := &ProviderConfig{
			Code:           "openai",
			ProviderRegion: "overseas",
			RouteStrategy:  nil,
			Endpoints: map[string]interface{}{
				"direct": map[string]interface{}{
					"overseas": "https://api.openai.com/v1",
				},
			},
		}

		merchantConfig := &MerchantConfig{
			ID:     1,
			Type:   "standard",
			Region: "overseas",
		}

		decision, err := router.DecideRoute(context.Background(), providerConfig, merchantConfig)
		require.NoError(t, err)
		assert.Equal(t, "direct", decision.Mode, "overseas user with no config should use direct")
	})
}

func TestE2E_RouteCacheIntegration(t *testing.T) {
	t.Run("RouteCache should cache and return decisions", func(t *testing.T) {
		cache := NewRouteCache(5 * time.Minute)

		cfg := providerRuntimeConfig{
			Code:           "openai",
			Name:           "OpenAI",
			APIBaseURL:      "https://api.openai.com/v1",
			APIFormat:       "openai",
			ProviderRegion:  "overseas",
		}

		cache.Set("openai", cfg)

		cached, found := cache.Get("openai")
		assert.True(t, found, "should find cached decision")
		assert.Equal(t, "openai", cached.Code, "cached code should match")
	})

	t.Run("RouteCache should respect TTL", func(t *testing.T) {
		cache := NewRouteCache(0)

		cfg := providerRuntimeConfig{
			Code:       "openai",
			APIBaseURL: "https://api.openai.com/v1",
		}

		cache.Set("openai", cfg)

		_, found := cache.Get("openai")
		assert.False(t, found, "should not find expired cached decision")
	})
}


