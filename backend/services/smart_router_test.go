package services

import (
	"testing"
)

func TestRoutingStrategyConstants(t *testing.T) {
	tests := []struct {
		name     string
		strategy RoutingStrategy
		expected string
	}{
		{"price strategy", RoutingStrategyPrice, "price_first"},
		{"latency strategy", RoutingStrategyLatency, "latency_first"},
		{"balanced strategy", RoutingStrategyBalanced, "balanced"},
		{"cost strategy", RoutingStrategyCost, "cost_first"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.strategy) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, tt.strategy)
			}
		})
	}
}

func TestRoutingCandidate(t *testing.T) {
	candidate := RoutingCandidate{
		APIKeyID:     1,
		Provider:     "openai",
		Model:        "gpt-4",
		Score:        0.85,
		HealthStatus: "healthy",
		Verified:     true,
	}

	if candidate.APIKeyID != 1 {
		t.Errorf("Expected APIKeyID=1, got %d", candidate.APIKeyID)
	}
	if candidate.Provider != "openai" {
		t.Errorf("Expected Provider=openai, got %s", candidate.Provider)
	}
	if !candidate.Verified {
		t.Error("Expected Verified=true")
	}
}

func TestGetSmartRouter(t *testing.T) {
	router1 := GetSmartRouter()
	router2 := GetSmartRouter()

	if router1 != router2 {
		t.Error("Expected singleton instance")
	}
}

func TestFilterUnhealthy(t *testing.T) {
	router := GetSmartRouter()

	candidates := []RoutingCandidate{
		{APIKeyID: 1, Provider: "openai", HealthStatus: "healthy", Verified: true},
		{APIKeyID: 2, Provider: "anthropic", HealthStatus: "unhealthy", Verified: true},
		{APIKeyID: 3, Provider: "google", HealthStatus: "degraded", Verified: true},
	}

	filtered := router.FilterUnhealthy(candidates)

	if len(filtered) != 2 {
		t.Errorf("Expected 2 healthy candidates, got %d", len(filtered))
	}

	for _, c := range filtered {
		if c.HealthStatus == "unhealthy" {
			t.Error("Unhealthy candidate should be filtered out")
		}
	}
}

func TestFilterUnverified(t *testing.T) {
	router := GetSmartRouter()

	candidates := []RoutingCandidate{
		{APIKeyID: 1, Provider: "openai", Verified: true},
		{APIKeyID: 2, Provider: "anthropic", Verified: false},
		{APIKeyID: 3, Provider: "google", Verified: true},
	}

	filtered := router.FilterUnverified(candidates)

	if len(filtered) != 2 {
		t.Errorf("Expected 2 verified candidates, got %d", len(filtered))
	}

	for _, c := range filtered {
		if !c.Verified {
			t.Error("Unverified candidate should be filtered out")
		}
	}
}

func TestCalculateScores(t *testing.T) {
	router := GetSmartRouter()

	candidates := []RoutingCandidate{
		{
			APIKeyID:     1,
			Provider:     "openai",
			InputPrice:   0.03,
			OutputPrice:  0.06,
			AvgLatencyMs: 100,
			SuccessRate:  99.0,
		},
		{
			APIKeyID:     2,
			Provider:     "anthropic",
			InputPrice:   0.015,
			OutputPrice:  0.075,
			AvgLatencyMs: 150,
			SuccessRate:  98.0,
		},
	}

	router.CalculateScores(candidates, RoutingStrategyBalanced)

	for _, c := range candidates {
		if c.PriceScore < 0 || c.PriceScore > 1 {
			t.Errorf("PriceScore should be between 0 and 1, got %f", c.PriceScore)
		}
		if c.LatencyScore < 0 || c.LatencyScore > 1 {
			t.Errorf("LatencyScore should be between 0 and 1, got %f", c.LatencyScore)
		}
		if c.SuccessScore < 0 || c.SuccessScore > 1 {
			t.Errorf("SuccessScore should be between 0 and 1, got %f", c.SuccessScore)
		}
		if c.Score < 0 || c.Score > 1 {
			t.Errorf("Overall Score should be between 0 and 1, got %f", c.Score)
		}
	}
}

func TestGetStrategyWeights(t *testing.T) {
	router := GetSmartRouter()

	tests := []struct {
		name     string
		strategy RoutingStrategy
		expected StrategyWeights
	}{
		{
			name:     "price strategy",
			strategy: RoutingStrategyPrice,
			expected: StrategyWeights{Price: 0.6, Latency: 0.2, Success: 0.2},
		},
		{
			name:     "latency strategy",
			strategy: RoutingStrategyLatency,
			expected: StrategyWeights{Price: 0.2, Latency: 0.6, Success: 0.2},
		},
		{
			name:     "cost strategy",
			strategy: RoutingStrategyCost,
			expected: StrategyWeights{Price: 0.7, Latency: 0.1, Success: 0.2},
		},
		{
			name:     "balanced strategy",
			strategy: RoutingStrategyBalanced,
			expected: StrategyWeights{Price: 0.33, Latency: 0.34, Success: 0.33},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			weights := router.getStrategyWeights(tt.strategy)

			if weights.Price != tt.expected.Price {
				t.Errorf("Price weight: expected %f, got %f", tt.expected.Price, weights.Price)
			}
			if weights.Latency != tt.expected.Latency {
				t.Errorf("Latency weight: expected %f, got %f", tt.expected.Latency, weights.Latency)
			}
			if weights.Success != tt.expected.Success {
				t.Errorf("Success weight: expected %f, got %f", tt.expected.Success, weights.Success)
			}
		})
	}
}

func TestGetPriceRange(t *testing.T) {
	router := GetSmartRouter()

	candidates := []RoutingCandidate{
		{InputPrice: 0.01, OutputPrice: 0.02},
		{InputPrice: 0.03, OutputPrice: 0.06},
		{InputPrice: 0.02, OutputPrice: 0.04},
	}

	min, max := router.getPriceRange(candidates)

	expectedMin := 0.03
	expectedMax := 0.09

	if min != expectedMin {
		t.Errorf("Min price: expected %f, got %f", expectedMin, min)
	}
	if max != expectedMax {
		t.Errorf("Max price: expected %f, got %f", expectedMax, max)
	}
}

func TestGetLatencyRange(t *testing.T) {
	router := GetSmartRouter()

	candidates := []RoutingCandidate{
		{AvgLatencyMs: 100},
		{AvgLatencyMs: 200},
		{AvgLatencyMs: 150},
	}

	min, max := router.getLatencyRange(candidates)

	if min != 100 {
		t.Errorf("Min latency: expected 100, got %d", min)
	}
	if max != 200 {
		t.Errorf("Max latency: expected 200, got %d", max)
	}
}

func TestRecordRequestResult(t *testing.T) {
	router := GetSmartRouter()

	router.RecordRequestResult(999, true)
	router.RecordRequestResult(999, false)

	if router.IsCircuitBreakerOpen(999) {
		t.Error("Circuit breaker should not be open after 1 failure")
	}
}

func TestFilterByRouteDecision(t *testing.T) {
	router := GetSmartRouter()

	candidates := []RoutingCandidate{
		{APIKeyID: 1, Provider: "openai", HealthStatus: "healthy", Verified: true},
		{APIKeyID: 2, Provider: "anthropic", HealthStatus: "healthy", Verified: true},
		{APIKeyID: 3, Provider: "google", HealthStatus: "healthy", Verified: true},
	}

	tests := []struct {
		name     string
		decision *RouteDecision
		expected int
	}{
		{
			name:     "nil decision returns all candidates",
			decision: nil,
			expected: 3,
		},
		{
			name: "direct mode decision",
			decision: &RouteDecision{
				Mode:     "direct",
				Endpoint: "https://api.openai.com/v1",
				Reason:   "auto: direct connection",
			},
			expected: 3,
		},
		{
			name: "litellm mode decision",
			decision: &RouteDecision{
				Mode:     "litellm",
				Endpoint: "http://litellm-overseas:4000/v1",
				Reason:   "auto: domestic user accessing overseas provider",
			},
			expected: 3,
		},
		{
			name: "proxy mode decision with fallback",
			decision: &RouteDecision{
				Mode:             "litellm",
				Endpoint:         "http://litellm-overseas:4000/v1",
				FallbackMode:     "proxy",
				FallbackEndpoint: "https://gaap.example.com",
				Reason:           "configured route",
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := router.FilterByRouteDecision(candidates, tt.decision)
			if len(filtered) != tt.expected {
				t.Errorf("Expected %d candidates, got %d", tt.expected, len(filtered))
			}
		})
	}
}

func TestMatchesRouteDecision(t *testing.T) {
	router := GetSmartRouter()

	candidate := RoutingCandidate{
		APIKeyID:     1,
		Provider:     "openai",
		HealthStatus: "healthy",
		Verified:     true,
	}

	tests := []struct {
		name     string
		decision *RouteDecision
		expected bool
	}{
		{
			name:     "nil decision matches",
			decision: nil,
			expected: true,
		},
		{
			name: "direct mode matches",
			decision: &RouteDecision{
				Mode:     "direct",
				Endpoint: "https://api.openai.com/v1",
			},
			expected: true,
		},
		{
			name: "litellm mode matches",
			decision: &RouteDecision{
				Mode:     "litellm",
				Endpoint: "http://litellm-overseas:4000/v1",
			},
			expected: true,
		},
		{
			name: "proxy mode matches",
			decision: &RouteDecision{
				Mode:     "proxy",
				Endpoint: "https://gaap.example.com",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := router.matchesRouteDecision(candidate, tt.decision)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
