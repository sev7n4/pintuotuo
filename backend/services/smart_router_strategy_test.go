package services

import (
	"testing"
)

func TestGetStrategyWeights_AllStrategies(t *testing.T) {
	router := GetSmartRouter()

	tests := []struct {
		name           string
		strategy       RoutingStrategy
		expectPrice    float64
		expectLatency  float64
		expectSuccess  float64
	}{
		{
			name:          "price_first strategy",
			strategy:      RoutingStrategyPrice,
			expectPrice:   0.6,
			expectLatency: 0.2,
			expectSuccess: 0.2,
		},
		{
			name:          "latency_first strategy",
			strategy:      RoutingStrategyLatency,
			expectPrice:   0.2,
			expectLatency: 0.6,
			expectSuccess: 0.2,
		},
		{
			name:          "balanced strategy",
			strategy:      RoutingStrategyBalanced,
			expectPrice:   0.33,
			expectLatency: 0.34,
			expectSuccess: 0.33,
		},
		{
			name:          "cost_first strategy",
			strategy:      RoutingStrategyCost,
			expectPrice:   0.7,
			expectLatency: 0.1,
			expectSuccess: 0.2,
		},
		{
			name:          "reliability_first strategy",
			strategy:      RoutingStrategyReliability,
			expectPrice:   0.2,
			expectLatency: 0.2,
			expectSuccess: 0.6,
		},
		{
			name:          "performance_first strategy",
			strategy:      RoutingStrategyPerformance,
			expectPrice:   0.1,
			expectLatency: 0.5,
			expectSuccess: 0.2,
		},
		{
			name:          "security_first strategy",
			strategy:      RoutingStrategySecurity,
			expectPrice:   0.1,
			expectLatency: 0.1,
			expectSuccess: 0.2,
		},
		{
			name:          "auto strategy",
			strategy:      RoutingStrategyAuto,
			expectPrice:   0.2,
			expectLatency: 0.3,
			expectSuccess: 0.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			weights := router.getStrategyWeights(tt.strategy)
			if weights.Price != tt.expectPrice {
				t.Errorf("Price weight = %v, want %v", weights.Price, tt.expectPrice)
			}
			if weights.Latency != tt.expectLatency {
				t.Errorf("Latency weight = %v, want %v", weights.Latency, tt.expectLatency)
			}
			if weights.Success != tt.expectSuccess {
				t.Errorf("Success weight = %v, want %v", weights.Success, tt.expectSuccess)
			}
		})
	}
}

func TestGetStrategyWeightsFromDB(t *testing.T) {
	router := GetSmartRouter()

	weights, err := router.getStrategyWeightsFromDB("price_first")
	if err != nil {
		t.Logf("getStrategyWeightsFromDB returned error (expected if DB not available): %v", err)
		return
	}

	if weights.Price <= 0 || weights.Latency <= 0 || weights.Success <= 0 {
		t.Errorf("Invalid weights from DB: Price=%v, Latency=%v, Success=%v", weights.Price, weights.Latency, weights.Success)
	}
}
