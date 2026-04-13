package services

import (
	"testing"
	"time"
)

func TestGetPricingService(t *testing.T) {
	service1 := GetPricingService()
	service2 := GetPricingService()

	if service1 == nil {
		t.Fatal("GetPricingService() returned nil")
	}

	if service1 != service2 {
		t.Fatal("GetPricingService() should return singleton instance")
	}
}

func TestGetPricing(t *testing.T) {
	service := GetPricingService()

	tests := []struct {
		name      string
		provider  string
		model     string
		wantFound bool
	}{
		{
			name:      "OpenAI GPT-4",
			provider:  "openai",
			model:     "gpt-4-turbo-preview",
			wantFound: true,
		},
		{
			name:      "Anthropic Claude 3",
			provider:  "anthropic",
			model:     "claude-3-opus-20240229",
			wantFound: true,
		},
		{
			name:      "Unknown model",
			provider:  "unknown",
			model:     "unknown-model",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pricing, found := service.GetPricing(tt.provider, tt.model)
			if found != tt.wantFound {
				t.Errorf("GetPricing() found = %v, want %v", found, tt.wantFound)
			}
			if found && pricing.InputPrice <= 0 {
				t.Errorf("GetPricing() InputPrice should be positive, got %v", pricing.InputPrice)
			}
			if found && pricing.OutputPrice <= 0 {
				t.Errorf("GetPricing() OutputPrice should be positive, got %v", pricing.OutputPrice)
			}
		})
	}
}

func TestCalculateCost(t *testing.T) {
	service := GetPricingService()

	tests := []struct {
		name         string
		provider     string
		model        string
		inputTokens  int
		outputTokens int
		wantMinCost  float64
	}{
		{
			name:         "GPT-4 Turbo pricing",
			provider:     "openai",
			model:        "gpt-4-turbo-preview",
			inputTokens:  1000,
			outputTokens: 1000,
			wantMinCost:  0.00001,
		},
		{
			name:         "Unknown model uses default pricing",
			provider:     "unknown",
			model:        "unknown-model",
			inputTokens:  1000,
			outputTokens: 1000,
			wantMinCost:  0.000001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost := service.CalculateCost(tt.provider, tt.model, tt.inputTokens, tt.outputTokens)
			if cost < tt.wantMinCost {
				t.Errorf("CalculateCost() = %v, want at least %v", cost, tt.wantMinCost)
			}
		})
	}
}

func TestCacheExpiration(t *testing.T) {
	service := GetPricingService()

	service.cacheMutex.Lock()
	service.lastLoadTime = time.Now().Add(-10 * time.Minute)
	service.cacheMutex.Unlock()

	pricing, found := service.GetPricing("openai", "gpt-4-turbo-preview")
	if !found {
		t.Error("GetPricing() should reload cache when expired")
	}
	if pricing.InputPrice <= 0 {
		t.Errorf("GetPricing() should return valid pricing after reload, got InputPrice=%v", pricing.InputPrice)
	}
}

func TestGetStrategyConfig(t *testing.T) {
	router := GetSmartRouter()

	tests := []struct {
		name         string
		strategyCode string
		wantFound    bool
		wantPriceW   float64
		wantLatencyW float64
		wantReliabW  float64
	}{
		{
			name:         "Price first strategy",
			strategyCode: "price_first",
			wantFound:    true,
			wantPriceW:   0.6,
			wantLatencyW: 0.2,
			wantReliabW:  0.2,
		},
		{
			name:         "Balanced strategy",
			strategyCode: "balanced",
			wantFound:    true,
			wantPriceW:   0.33,
			wantLatencyW: 0.34,
			wantReliabW:  0.33,
		},
		{
			name:         "Unknown strategy",
			strategyCode: "unknown",
			wantFound:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, found := router.GetStrategyConfig(tt.strategyCode)
			if found != tt.wantFound {
				t.Errorf("GetStrategyConfig() found = %v, want %v", found, tt.wantFound)
			}
			if found {
				if config.PriceWeight != tt.wantPriceW {
					t.Errorf("GetStrategyConfig() PriceWeight = %v, want %v", config.PriceWeight, tt.wantPriceW)
				}
				if config.LatencyWeight != tt.wantLatencyW {
					t.Errorf("GetStrategyConfig() LatencyWeight = %v, want %v", config.LatencyWeight, tt.wantLatencyW)
				}
				if config.ReliabilityWeight != tt.wantReliabW {
					t.Errorf("GetStrategyConfig() ReliabilityWeight = %v, want %v", config.ReliabilityWeight, tt.wantReliabW)
				}
			}
		})
	}
}

func TestPricingHistoryStruct(t *testing.T) {
	history := PricingHistory{
		ID:             1,
		EntityType:     "spu",
		EntityID:       100,
		OldInputPrice:  10.0,
		OldOutputPrice: 30.0,
		NewInputPrice:  12.0,
		NewOutputPrice: 35.0,
		ChangeReason:   "Market adjustment",
		ChangedBy:      1,
		ChangedAt:      time.Now(),
		EffectiveAt:    time.Now(),
	}

	if history.EntityType != "spu" {
		t.Errorf("PricingHistory EntityType = %v, want spu", history.EntityType)
	}
	if history.OldInputPrice >= history.NewInputPrice {
		t.Errorf("PricingHistory OldInputPrice should be less than NewInputPrice")
	}
}

func TestPricingScheduleStruct(t *testing.T) {
	schedule := PricingSchedule{
		ID:             1,
		EntityType:     "spu",
		EntityID:       100,
		NewInputPrice:  15.0,
		NewOutputPrice: 40.0,
		ScheduledAt:    time.Now().Add(24 * time.Hour),
		Status:         "pending",
		ChangeReason:   "Scheduled price increase",
		CreatedBy:      1,
	}

	if schedule.Status != "pending" {
		t.Errorf("PricingSchedule Status = %v, want pending", schedule.Status)
	}
	if schedule.ScheduledAt.Before(time.Now()) {
		t.Errorf("PricingSchedule ScheduledAt should be in the future")
	}
}
