package services

import (
	"context"
	"testing"
)

func TestResolveEndpointByRouteMode(t *testing.T) {
	validator := &APIKeyValidator{}
	ctx := context.Background()

	tests := []struct {
		name        string
		provider    string
		routeMode   string
		routeConfig map[string]interface{}
		region      string
		wantErr     bool
	}{
		{
			name:      "Direct mode with endpoint_url",
			provider:  "openai",
			routeMode: "direct",
			routeConfig: map[string]interface{}{
				"endpoint_url": "https://custom.openai.com",
			},
			region:  "overseas",
			wantErr: false,
		},
		{
			name:      "Direct mode with endpoints.direct.overseas",
			provider:  "openai",
			routeMode: "direct",
			routeConfig: map[string]interface{}{
				"endpoints": map[string]interface{}{
					"direct": map[string]interface{}{
						"overseas": "https://api.openai.com",
						"domestic": "https://api.openai-proxy.com",
					},
				},
			},
			region:  "overseas",
			wantErr: false,
		},
		{
			name:      "LiteLLM mode with endpoints.litellm.domestic",
			provider:  "openai",
			routeMode: "litellm",
			routeConfig: map[string]interface{}{
				"endpoints": map[string]interface{}{
					"litellm": map[string]interface{}{
						"domestic": "http://litellm.local:4000/v1",
						"overseas": "http://litellm-overseas.local:4000/v1",
					},
				},
			},
			region:  "domestic",
			wantErr: false,
		},
		{
			name:      "Proxy mode with endpoints.proxy.gaap",
			provider:  "openai",
			routeMode: "proxy",
			routeConfig: map[string]interface{}{
				"endpoints": map[string]interface{}{
					"proxy": map[string]interface{}{
						"gaap": "https://proxy.gaap-real.com",
					},
				},
			},
			region:  "",
			wantErr: false,
		},
		{
			name:        "Auto mode with no config",
			provider:    "openai",
			routeMode:   "auto",
			routeConfig: map[string]interface{}{},
			region:      "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoint, err := validator.resolveEndpointByRouteMode(ctx, tt.provider, tt.routeMode, tt.routeConfig, tt.region)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveEndpointByRouteMode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && endpoint == "" {
				t.Errorf("resolveEndpointByRouteMode() endpoint is empty")
			}
		})
	}
}

func TestResolveDirectEndpoint(t *testing.T) {
	validator := &APIKeyValidator{}
	ctx := context.Background()

	tests := []struct {
		name        string
		provider    string
		routeConfig map[string]interface{}
		region      string
		wantErr     bool
	}{
		{
			name:     "Priority 1: endpoint_url",
			provider: "openai",
			routeConfig: map[string]interface{}{
				"endpoint_url": "https://custom.openai.com",
				"endpoints": map[string]interface{}{
					"direct": map[string]interface{}{
						"overseas": "https://api.openai.com",
					},
				},
			},
			region:  "overseas",
			wantErr: false,
		},
		{
			name:     "Priority 2: endpoints.direct.{region}",
			provider: "openai",
			routeConfig: map[string]interface{}{
				"endpoints": map[string]interface{}{
					"direct": map[string]interface{}{
						"overseas": "https://api.openai.com",
						"domestic": "https://api.openai-proxy.com",
					},
				},
			},
			region:  "domestic",
			wantErr: false,
		},
		{
			name:        "Default region: overseas",
			provider:    "openai",
			routeConfig: map[string]interface{}{},
			region:      "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoint, err := validator.resolveDirectEndpoint(ctx, tt.provider, tt.routeConfig, tt.region)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveDirectEndpoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && endpoint == "" {
				t.Errorf("resolveDirectEndpoint() endpoint is empty")
			}
		})
	}
}

func TestResolveLitellmEndpoint(t *testing.T) {
	validator := &APIKeyValidator{}
	ctx := context.Background()

	tests := []struct {
		name        string
		routeConfig map[string]interface{}
		region      string
		wantErr     bool
	}{
		{
			name: "Priority 1: endpoints.litellm.{region}",
			routeConfig: map[string]interface{}{
				"endpoints": map[string]interface{}{
					"litellm": map[string]interface{}{
						"domestic": "http://litellm.local:4000/v1",
						"overseas": "http://litellm-overseas.local:4000/v1",
					},
				},
			},
			region:  "domestic",
			wantErr: false,
		},
		{
			name: "Priority 2: base_url",
			routeConfig: map[string]interface{}{
				"base_url": "http://litellm.local:4000/v1",
			},
			region:  "",
			wantErr: false,
		},
		{
			name:        "No config",
			routeConfig: map[string]interface{}{},
			region:      "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoint, err := validator.resolveLitellmEndpoint(ctx, tt.routeConfig, tt.region)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveLitellmEndpoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && endpoint == "" {
				t.Errorf("resolveLitellmEndpoint() endpoint is empty")
			}
		})
	}
}

func TestResolveProxyEndpoint(t *testing.T) {
	validator := &APIKeyValidator{}
	ctx := context.Background()

	tests := []struct {
		name        string
		provider    string
		routeConfig map[string]interface{}
		wantErr     bool
	}{
		{
			name:     "Priority 1: endpoints.proxy.gaap",
			provider: "openai",
			routeConfig: map[string]interface{}{
				"endpoints": map[string]interface{}{
					"proxy": map[string]interface{}{
						"gaap": "https://proxy.gaap-real.com",
					},
				},
			},
			wantErr: false,
		},
		{
			name:     "Priority 2: proxy_url",
			provider: "openai",
			routeConfig: map[string]interface{}{
				"proxy_url": "https://proxy.real-proxy.com",
			},
			wantErr: false,
		},
		{
			name:        "No config, no DB returns error",
			provider:    "openai",
			routeConfig: map[string]interface{}{},
			wantErr:     true,
		},
		{
			name:     "Skip example.com proxy endpoints",
			provider: "openai",
			routeConfig: map[string]interface{}{
				"endpoints": map[string]interface{}{
					"proxy": map[string]interface{}{
						"gaap":     "https://google-gaap.example.com",
						"nginx_hk": "https://google-proxy-hk.example.com",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoint, err := validator.resolveProxyEndpoint(ctx, tt.provider, tt.routeConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveProxyEndpoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && endpoint == "" {
				t.Errorf("resolveProxyEndpoint() endpoint is empty")
			}
		})
	}
}

func TestResolveAuthToken(t *testing.T) {
	validator := &APIKeyValidator{}

	tests := []struct {
		name           string
		routeMode      string
		originalAPIKey string
		wantToken      string
	}{
		{
			name:           "LiteLLM mode without master key",
			routeMode:      "litellm",
			originalAPIKey: "sk-test-key",
			wantToken:      "sk-test-key",
		},
		{
			name:           "Direct mode",
			routeMode:      "direct",
			originalAPIKey: "sk-test-key",
			wantToken:      "sk-test-key",
		},
		{
			name:           "Proxy mode",
			routeMode:      "proxy",
			originalAPIKey: "sk-test-key",
			wantToken:      "sk-test-key",
		},
		{
			name:           "Auto mode",
			routeMode:      "auto",
			originalAPIKey: "sk-test-key",
			wantToken:      "sk-test-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := validator.resolveAuthToken(tt.routeMode, tt.originalAPIKey)
			if token != tt.wantToken {
				t.Errorf("resolveAuthToken() token = %v, want %v", token, tt.wantToken)
			}
		})
	}
}

func TestRegionHandling(t *testing.T) {
	validator := &APIKeyValidator{}
	ctx := context.Background()

	tests := []struct {
		name        string
		routeMode   string
		routeConfig map[string]interface{}
		region      string
		wantErr     bool
	}{
		{
			name:      "Direct mode with domestic region",
			routeMode: "direct",
			routeConfig: map[string]interface{}{
				"endpoints": map[string]interface{}{
					"direct": map[string]interface{}{
						"domestic": "https://api.domestic.com",
						"overseas": "https://api.overseas.com",
					},
				},
			},
			region:  "domestic",
			wantErr: false,
		},
		{
			name:      "Direct mode with overseas region",
			routeMode: "direct",
			routeConfig: map[string]interface{}{
				"endpoints": map[string]interface{}{
					"direct": map[string]interface{}{
						"domestic": "https://api.domestic.com",
						"overseas": "https://api.overseas.com",
					},
				},
			},
			region:  "overseas",
			wantErr: false,
		},
		{
			name:      "LiteLLM mode with domestic region",
			routeMode: "litellm",
			routeConfig: map[string]interface{}{
				"endpoints": map[string]interface{}{
					"litellm": map[string]interface{}{
						"domestic": "http://litellm.local:4000/v1",
						"overseas": "http://litellm-overseas.local:4000/v1",
					},
				},
			},
			region:  "domestic",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoint, err := validator.resolveEndpointByRouteMode(ctx, "openai", tt.routeMode, tt.routeConfig, tt.region)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveEndpointByRouteMode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && endpoint == "" {
				t.Errorf("resolveEndpointByRouteMode() endpoint is empty")
			}
		})
	}
}
