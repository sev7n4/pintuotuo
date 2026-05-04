package models

import (
	"testing"
)

func TestMerchantAPIKey_GetEndpointForMode(t *testing.T) {
	tests := []struct {
		name        string
		apiKey      MerchantAPIKey
		mode        string
		region      string
		wantEndpoint string
	}{
		{
			name: "Direct mode with endpoint_url column takes priority over route_config",
			apiKey: MerchantAPIKey{
				EndpointURL: "https://default.example.com",
				RouteConfig: map[string]interface{}{
					"endpoint_url": "https://custom.example.com",
				},
			},
			mode:        "direct",
			region:      "overseas",
			wantEndpoint: "https://default.example.com",
		},
		{
			name: "Direct mode with endpoints.direct.overseas falls back when EndpointURL empty",
			apiKey: MerchantAPIKey{
				RouteConfig: map[string]interface{}{
					"endpoints": map[string]interface{}{
						"direct": map[string]interface{}{
							"overseas": "https://api.overseas.example.com",
							"domestic": "https://api.domestic.example.com",
						},
					},
				},
			},
			mode:        "direct",
			region:      "overseas",
			wantEndpoint: "https://api.overseas.example.com",
		},
		{
			name: "Direct mode EndpointURL overrides endpoints.direct",
			apiKey: MerchantAPIKey{
				EndpointURL: "https://user-configured.example.com",
				RouteConfig: map[string]interface{}{
					"endpoints": map[string]interface{}{
						"direct": map[string]interface{}{
							"overseas": "https://stale-overseas.example.com",
							"domestic": "https://stale-domestic.example.com",
						},
					},
				},
			},
			mode:        "direct",
			region:      "overseas",
			wantEndpoint: "https://user-configured.example.com",
		},
		{
			name: "Direct mode with endpoints.direct.domestic falls back when EndpointURL empty",
			apiKey: MerchantAPIKey{
				RouteConfig: map[string]interface{}{
					"endpoints": map[string]interface{}{
						"direct": map[string]interface{}{
							"overseas": "https://api.overseas.example.com",
							"domestic": "https://api.domestic.example.com",
						},
					},
				},
			},
			mode:        "direct",
			region:      "domestic",
			wantEndpoint: "https://api.domestic.example.com",
		},
		{
			name: "LiteLLM mode with endpoints.litellm.domestic",
			apiKey: MerchantAPIKey{
				RouteConfig: map[string]interface{}{
					"endpoints": map[string]interface{}{
						"litellm": map[string]interface{}{
							"domestic": "http://litellm.local:4000/v1",
							"overseas": "http://litellm-overseas.local:4000/v1",
						},
					},
				},
			},
			mode:        "litellm",
			region:      "domestic",
			wantEndpoint: "http://litellm.local:4000/v1",
		},
		{
			name: "LiteLLM mode with base_url",
			apiKey: MerchantAPIKey{
				RouteConfig: map[string]interface{}{
					"base_url": "http://litellm.local:4000/v1",
				},
			},
			mode:        "litellm",
			region:      "domestic",
			wantEndpoint: "http://litellm.local:4000/v1",
		},
		{
			name: "Proxy mode with endpoints.proxy.gaap",
			apiKey: MerchantAPIKey{
				RouteConfig: map[string]interface{}{
					"endpoints": map[string]interface{}{
						"proxy": map[string]interface{}{
							"gaap": "https://proxy.example.com",
						},
					},
				},
			},
			mode:        "proxy",
			region:      "",
			wantEndpoint: "https://proxy.example.com",
		},
		{
			name: "Proxy mode with proxy_url",
			apiKey: MerchantAPIKey{
				RouteConfig: map[string]interface{}{
					"proxy_url": "https://proxy.example.com",
				},
			},
			mode:        "proxy",
			region:      "",
			wantEndpoint: "https://proxy.example.com",
		},
		{
			name: "Unknown mode falls back to EndpointURL",
			apiKey: MerchantAPIKey{
				EndpointURL: "https://default.example.com",
				RouteConfig: map[string]interface{}{},
			},
			mode:        "unknown",
			region:      "",
			wantEndpoint: "https://default.example.com",
		},
		{
			name: "Empty region defaults to overseas",
			apiKey: MerchantAPIKey{
				RouteConfig: map[string]interface{}{
					"endpoints": map[string]interface{}{
						"direct": map[string]interface{}{
							"overseas": "https://api.overseas.example.com",
						},
					},
				},
			},
			mode:        "direct",
			region:      "",
			wantEndpoint: "https://api.overseas.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoint := tt.apiKey.GetEndpointForMode(tt.mode, tt.region)
			if endpoint != tt.wantEndpoint {
				t.Errorf("GetEndpointForMode(%s, %s) = %s, want %s", tt.mode, tt.region, endpoint, tt.wantEndpoint)
			}
		})
	}
}

func TestMerchantAPIKey_HasRouteConfig(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   MerchantAPIKey
		expected bool
	}{
		{
			name: "Has route config",
			apiKey: MerchantAPIKey{
				RouteConfig: map[string]interface{}{
					"endpoint_url": "https://example.com",
				},
			},
			expected: true,
		},
		{
			name: "Empty route config",
			apiKey: MerchantAPIKey{
				RouteConfig: map[string]interface{}{},
			},
			expected: false,
		},
		{
			name: "Nil route config",
			apiKey: MerchantAPIKey{
				RouteConfig: nil,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.apiKey.HasRouteConfig()
			if result != tt.expected {
				t.Errorf("HasRouteConfig() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMerchantAPIKey_GetEndpoints(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   MerchantAPIKey
		expected map[string]interface{}
	}{
		{
			name: "Has endpoints",
			apiKey: MerchantAPIKey{
				RouteConfig: map[string]interface{}{
					"endpoints": map[string]interface{}{
						"direct": map[string]interface{}{
							"overseas": "https://api.example.com",
						},
					},
				},
			},
			expected: map[string]interface{}{
				"direct": map[string]interface{}{
					"overseas": "https://api.example.com",
				},
			},
		},
		{
			name: "No endpoints",
			apiKey: MerchantAPIKey{
				RouteConfig: map[string]interface{}{
					"endpoint_url": "https://example.com",
				},
			},
			expected: nil,
		},
		{
			name: "Nil route config",
			apiKey: MerchantAPIKey{
				RouteConfig: nil,
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.apiKey.GetEndpoints()
			if tt.expected == nil {
				if result != nil {
					t.Errorf("GetEndpoints() = %v, want nil", result)
				}
			} else {
				if result == nil {
					t.Errorf("GetEndpoints() = nil, want %v", tt.expected)
				}
			}
		})
	}
}

func TestMerchantAPIKey_GetEndpointByType(t *testing.T) {
	tests := []struct {
		name         string
		apiKey       MerchantAPIKey
		endpointType string
		expected     string
	}{
		{
			name: "Get direct endpoint",
			apiKey: MerchantAPIKey{
				Region: "overseas",
				RouteConfig: map[string]interface{}{
					"endpoints": map[string]interface{}{
						"direct": map[string]interface{}{
							"overseas": "https://api.overseas.example.com",
							"domestic": "https://api.domestic.example.com",
						},
					},
				},
			},
			endpointType: "direct",
			expected:     "https://api.overseas.example.com",
		},
		{
			name: "Get litellm endpoint",
			apiKey: MerchantAPIKey{
				Region: "domestic",
				RouteConfig: map[string]interface{}{
					"endpoints": map[string]interface{}{
						"litellm": map[string]interface{}{
							"domestic": "http://litellm.local:4000/v1",
						},
					},
				},
			},
			endpointType: "litellm",
			expected:     "http://litellm.local:4000/v1",
		},
		{
			name: "Endpoint type not found",
			apiKey: MerchantAPIKey{
				Region: "overseas",
				RouteConfig: map[string]interface{}{
					"endpoints": map[string]interface{}{
						"direct": map[string]interface{}{
							"overseas": "https://api.example.com",
						},
					},
				},
			},
			endpointType: "proxy",
			expected:     "",
		},
		{
			name: "Region not found",
			apiKey: MerchantAPIKey{
				Region: "asia",
				RouteConfig: map[string]interface{}{
					"endpoints": map[string]interface{}{
						"direct": map[string]interface{}{
							"overseas": "https://api.example.com",
						},
					},
				},
			},
			endpointType: "direct",
			expected:     "",
		},
		{
			name: "Empty region defaults to overseas",
			apiKey: MerchantAPIKey{
				Region: "",
				RouteConfig: map[string]interface{}{
					"endpoints": map[string]interface{}{
						"direct": map[string]interface{}{
							"overseas": "https://api.example.com",
						},
					},
				},
			},
			endpointType: "direct",
			expected:     "https://api.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.apiKey.GetEndpointByType(tt.endpointType)
			if result != tt.expected {
				t.Errorf("GetEndpointByType(%s) = %s, want %s", tt.endpointType, result, tt.expected)
			}
		})
	}
}
