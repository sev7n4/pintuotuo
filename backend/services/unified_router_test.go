package services

import (
	"context"
	"testing"
)

func TestNewUnifiedRouter(t *testing.T) {
	router := NewUnifiedRouter(nil)
	if router == nil {
		t.Error("expected router to be created, got nil")
	}
}

func TestRouteDecision_DomesticUserOverseasProvider(t *testing.T) {
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
				"domestic": "http://litellm-overseas:4000/v1",
			},
			"proxy": map[string]interface{}{
				"gaap": "https://gaap.example.com",
			},
		},
	}
	
	merchantConfig := &MerchantConfig{
		ID:     1,
		Type:   "standard",
		Region: "domestic",
	}
	
	decision, err := router.DecideRoute(context.Background(), providerConfig, merchantConfig)
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
}

func TestRouteDecision_OverseasUserOverseasProvider(t *testing.T) {
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
		ID:     2,
		Type:   "standard",
		Region: "overseas",
	}
	
	decision, err := router.DecideRoute(context.Background(), providerConfig, merchantConfig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if decision.Mode != "direct" {
		t.Errorf("expected mode to be 'direct', got '%s'", decision.Mode)
	}
	
	if decision.Endpoint != "https://api.openai.com/v1" {
		t.Errorf("expected endpoint to be 'https://api.openai.com/v1', got '%s'", decision.Endpoint)
	}
}

func TestRouteDecision_EnterpriseUser(t *testing.T) {
	router := NewUnifiedRouter(nil)
	
	providerConfig := &ProviderConfig{
		Code:           "deepseek",
		ProviderRegion: "domestic",
		RouteStrategy: map[string]interface{}{
			"default_mode": "auto",
			"enterprise_users": map[string]interface{}{
				"mode":          "litellm",
				"fallback_mode": "direct",
			},
		},
		Endpoints: map[string]interface{}{
			"litellm": map[string]interface{}{
				"domestic": "http://litellm-domestic:4000/v1",
			},
			"direct": map[string]interface{}{
				"domestic": "https://api.deepseek.com/v1",
			},
		},
	}
	
	merchantConfig := &MerchantConfig{
		ID:     3,
		Type:   "enterprise",
		Region: "domestic",
	}
	
	decision, err := router.DecideRoute(context.Background(), providerConfig, merchantConfig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if decision.Mode != "litellm" {
		t.Errorf("expected mode to be 'litellm', got '%s'", decision.Mode)
	}
	
	if decision.FallbackMode != "direct" {
		t.Errorf("expected fallback mode to be 'direct', got '%s'", decision.FallbackMode)
	}
}

func TestRouteDecision_AutoMode(t *testing.T) {
	router := NewUnifiedRouter(nil)
	
	tests := []struct {
		name            string
		providerRegion  string
		merchantRegion  string
		expectedMode    string
	}{
		{
			name:           "domestic user accessing overseas provider",
			providerRegion: "overseas",
			merchantRegion: "domestic",
			expectedMode:   "litellm",
		},
		{
			name:           "overseas user accessing overseas provider",
			providerRegion: "overseas",
			merchantRegion: "overseas",
			expectedMode:   "direct",
		},
		{
			name:           "domestic user accessing domestic provider",
			providerRegion: "domestic",
			merchantRegion: "domestic",
			expectedMode:   "direct",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			providerConfig := &ProviderConfig{
				Code:           "test",
				ProviderRegion: tt.providerRegion,
				RouteStrategy: map[string]interface{}{
					"default_mode": "auto",
				},
				Endpoints: map[string]interface{}{
					"direct": map[string]interface{}{
						"domestic":  "https://domestic.example.com",
						"overseas": "https://overseas.example.com",
					},
					"litellm": map[string]interface{}{
						"domestic":  "http://litellm:4000/v1",
						"overseas": "http://litellm:4000/v1",
					},
				},
			}
			
			merchantConfig := &MerchantConfig{
				ID:     1,
				Type:   "standard",
				Region: tt.merchantRegion,
			}
			
			decision, err := router.DecideRoute(context.Background(), providerConfig, merchantConfig)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			
			if decision.Mode != tt.expectedMode {
				t.Errorf("expected mode to be '%s', got '%s'", tt.expectedMode, decision.Mode)
			}
		})
	}
}

func TestSelectEndpoint(t *testing.T) {
	router := NewUnifiedRouter(nil)
	
	endpoints := map[string]interface{}{
		"direct": map[string]interface{}{
			"domestic":  "https://domestic.example.com",
			"overseas": "https://overseas.example.com",
		},
		"litellm": map[string]interface{}{
			"domestic": "http://litellm:4000/v1",
		},
	}
	
	tests := []struct {
		name          string
		mode          string
		region        string
		expectedURL   string
	}{
		{
			name:        "direct domestic",
			mode:        "direct",
			region:      "domestic",
			expectedURL: "https://domestic.example.com",
		},
		{
			name:        "direct overseas",
			mode:        "direct",
			region:      "overseas",
			expectedURL: "https://overseas.example.com",
		},
		{
			name:        "litellm domestic",
			mode:        "litellm",
			region:      "domestic",
			expectedURL: "http://litellm:4000/v1",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoint := router.SelectEndpoint(tt.mode, tt.region, endpoints)
			if endpoint != tt.expectedURL {
				t.Errorf("expected endpoint to be '%s', got '%s'", tt.expectedURL, endpoint)
			}
		})
	}
}
