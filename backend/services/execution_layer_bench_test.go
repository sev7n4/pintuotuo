package services

import (
	"context"
	"testing"
	"time"
)

func BenchmarkResolveRouteDecision(b *testing.B) {
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
			"overseas_users": map[string]interface{}{
				"mode": "direct",
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
		Type:   "standard",
		Region: "domestic",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = router.DecideRoute(context.Background(), providerConfig, merchantConfig)
	}
}

func BenchmarkResolveRouteDecision_Parallel(b *testing.B) {
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

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		merchantConfig := &MerchantConfig{
			ID:     1,
			Type:   "standard",
			Region: "domestic",
		}
		for pb.Next() {
			_, _ = router.DecideRoute(context.Background(), providerConfig, merchantConfig)
		}
	})
}

func BenchmarkRouteCache_Get(b *testing.B) {
	cache := NewRouteCache(5 * time.Minute)

	cfg := providerRuntimeConfig{
		Code:           "openai",
		Name:           "OpenAI",
		APIBaseURL:      "https://api.openai.com/v1",
		APIFormat:       "openai",
		ProviderRegion:  "overseas",
	}

	cache.Set("openai", cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cache.Get("openai")
	}
}

func BenchmarkRouteCache_Get_Parallel(b *testing.B) {
	cache := NewRouteCache(5 * time.Minute)

	providers := []string{"openai", "anthropic", "deepseek", "moonshot", "minimax",
		"google", "zhipu", "stepfun", "dashscope", "xai"}

	for _, p := range providers {
		cfg := providerRuntimeConfig{
			Code:           p,
			Name:           p,
			APIBaseURL:      "https://api.example.com/v1",
			APIFormat:       "openai",
			ProviderRegion:  "overseas",
		}
		cache.Set(p, cfg)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_, _ = cache.Get(providers[i%len(providers)])
			i++
		}
	})
}

func BenchmarkRouteCache_Set(b *testing.B) {
	cache := NewRouteCache(5 * time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cfg := providerRuntimeConfig{
			Code:           "openai",
			Name:           "OpenAI",
			APIBaseURL:      "https://api.openai.com/v1",
			APIFormat:       "openai",
			ProviderRegion:  "overseas",
		}
		cache.Set("openai", cfg)
	}
}

func BenchmarkRouteCache_Set_Parallel(b *testing.B) {
	cache := NewRouteCache(5 * time.Minute)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			cfg := providerRuntimeConfig{
				Code:           "openai",
				Name:           "OpenAI",
				APIBaseURL:      "https://api.openai.com/v1",
				APIFormat:       "openai",
				ProviderRegion:  "overseas",
			}
			cache.Set("openai", cfg)
			i++
		}
	})
}

func BenchmarkTokenEstimation(b *testing.B) {
	service := NewTokenEstimationService()

	req := &RoutingRequest{
		Model: "gpt-4",
		RequestBody: map[string]interface{}{
			"messages": []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "Hello, how are you? This is a longer message to test token estimation performance.",
				},
			},
			"max_tokens": 1000,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.EstimateTokens(req)
	}
}

func BenchmarkRoutingStrategyEngine_DefineGoal(b *testing.B) {
	engine := NewRoutingStrategyEngine()

	reqCtx := &RequestContext{
		RequestAnalysis: &RequestAnalysis{
			Intent:          IntentChat,
			EstimatedTokens: 1000,
			Stream:          false,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.DefineGoal(nil, reqCtx)
	}
}

func BenchmarkRoutingStrategyEngine_GetStrategyWeights(b *testing.B) {
	engine := NewRoutingStrategyEngine()

	strategies := []StrategyGoal{
		GoalPriceFirst,
		GoalPerformanceFirst,
		GoalReliabilityFirst,
		GoalSecurityFirst,
		GoalBalanced,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.GetStrategyWeights(strategies[i%len(strategies)])
	}
}
