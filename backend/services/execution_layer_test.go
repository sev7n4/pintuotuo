package services

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetermineGatewayMode(t *testing.T) {
	tests := []struct {
		name          string
		cfg           *ExecutionProviderConfig
		envHTTPSProxy string
		envLitellmURL string
		expected      string
	}{
		{
			name:          "nil config returns direct",
			cfg:           nil,
			envHTTPSProxy: "",
			envLitellmURL: "",
			expected:      GatewayModeDirect,
		},
		{
			name: "explicit direct mode not affected",
			cfg: &ExecutionProviderConfig{
				BYOKRouteMode:  GatewayModeDirect,
				ProviderRegion: regionOverseas,
			},
			envHTTPSProxy: "http://proxy:7890",
			envLitellmURL: "",
			expected:      GatewayModeDirect,
		},
		{
			name: "explicit litellm mode not affected",
			cfg: &ExecutionProviderConfig{
				BYOKRouteMode:  GatewayModeLitellm,
				ProviderRegion: regionOverseas,
			},
			envHTTPSProxy: "http://proxy:7890",
			envLitellmURL: "",
			expected:      GatewayModeLitellm,
		},
		{
			name: "explicit proxy mode not affected",
			cfg: &ExecutionProviderConfig{
				BYOKRouteMode:  GatewayModeProxy,
				ProviderRegion: regionOverseas,
			},
			envHTTPSProxy: "",
			envLitellmURL: "",
			expected:      GatewayModeProxy,
		},
		{
			name: "auto + overseas + HTTPS_PROXY returns proxy",
			cfg: &ExecutionProviderConfig{
				BYOKRouteMode:  RouteModeAuto,
				ProviderRegion: regionOverseas,
			},
			envHTTPSProxy: "http://host.docker.internal:7890",
			envLitellmURL: "",
			expected:      GatewayModeProxy,
		},
		{
			name: "auto + overseas + https_proxy lowercase returns proxy",
			cfg: &ExecutionProviderConfig{
				BYOKRouteMode:  RouteModeAuto,
				ProviderRegion: regionOverseas,
			},
			envHTTPSProxy: "",
			envLitellmURL: "",
			expected:      GatewayModeProxy,
		},
		{
			name: "auto + overseas + no HTTPS_PROXY + LiteLLM URL returns litellm",
			cfg: &ExecutionProviderConfig{
				BYOKRouteMode:  RouteModeAuto,
				ProviderRegion: regionOverseas,
			},
			envHTTPSProxy: "",
			envLitellmURL: "http://litellm:4000",
			expected:      GatewayModeLitellm,
		},
		{
			name: "auto + overseas + no proxy no litellm returns direct",
			cfg: &ExecutionProviderConfig{
				BYOKRouteMode:  RouteModeAuto,
				ProviderRegion: regionOverseas,
			},
			envHTTPSProxy: "",
			envLitellmURL: "",
			expected:      GatewayModeDirect,
		},
		{
			name: "auto + domestic returns direct regardless of HTTPS_PROXY",
			cfg: &ExecutionProviderConfig{
				BYOKRouteMode:  RouteModeAuto,
				ProviderRegion: regionDomestic,
			},
			envHTTPSProxy: "http://host.docker.internal:7890",
			envLitellmURL: "",
			expected:      GatewayModeDirect,
		},
		{
			name: "auto + empty region + HTTPS_PROXY returns direct",
			cfg: &ExecutionProviderConfig{
				BYOKRouteMode:  RouteModeAuto,
				ProviderRegion: "",
			},
			envHTTPSProxy: "http://host.docker.internal:7890",
			envLitellmURL: "",
			expected:      GatewayModeDirect,
		},
		{
			name: "empty route mode treated as auto + overseas + HTTPS_PROXY returns proxy",
			cfg: &ExecutionProviderConfig{
				BYOKRouteMode:  "",
				ProviderRegion: regionOverseas,
			},
			envHTTPSProxy: "http://host.docker.internal:7890",
			envLitellmURL: "",
			expected:      GatewayModeProxy,
		},
		{
			name: "HTTPS_PROXY takes priority over LiteLLM URL",
			cfg: &ExecutionProviderConfig{
				BYOKRouteMode:  RouteModeAuto,
				ProviderRegion: regionOverseas,
			},
			envHTTPSProxy: "http://host.docker.internal:7890",
			envLitellmURL: "http://litellm:4000",
			expected:      GatewayModeProxy,
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

			layer := &ExecutionLayer{}
			result := layer.determineGatewayMode(tt.cfg)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResolveRouteModeWithProvider(t *testing.T) {
	tests := []struct {
		name          string
		routeMode     string
		providerRegion string
		envHTTPSProxy string
		envLitellmURL string
		expected      string
	}{
		{
			name:           "explicit direct not affected",
			routeMode:      GatewayModeDirect,
			providerRegion: regionOverseas,
			envHTTPSProxy:  "http://proxy:7890",
			expected:       GatewayModeDirect,
		},
		{
			name:           "explicit litellm not affected",
			routeMode:      GatewayModeLitellm,
			providerRegion: regionOverseas,
			envHTTPSProxy:  "http://proxy:7890",
			expected:       GatewayModeLitellm,
		},
		{
			name:           "explicit proxy not affected",
			routeMode:      GatewayModeProxy,
			providerRegion: regionDomestic,
			envHTTPSProxy:  "",
			expected:       GatewayModeProxy,
		},
		{
			name:           "auto + overseas + HTTPS_PROXY returns proxy",
			routeMode:      RouteModeAuto,
			providerRegion: regionOverseas,
			envHTTPSProxy:  "http://host.docker.internal:7890",
			expected:       GatewayModeProxy,
		},
		{
			name:           "auto + overseas + no HTTPS_PROXY + LiteLLM URL returns litellm",
			routeMode:      RouteModeAuto,
			providerRegion: regionOverseas,
			envHTTPSProxy:  "",
			envLitellmURL:  "http://litellm:4000",
			expected:       GatewayModeLitellm,
		},
		{
			name:           "auto + overseas + no proxy no litellm returns direct",
			routeMode:      RouteModeAuto,
			providerRegion: regionOverseas,
			envHTTPSProxy:  "",
			envLitellmURL:  "",
			expected:       GatewayModeDirect,
		},
		{
			name:           "auto + domestic returns direct regardless of HTTPS_PROXY",
			routeMode:      RouteModeAuto,
			providerRegion: regionDomestic,
			envHTTPSProxy:  "http://host.docker.internal:7890",
			expected:       GatewayModeDirect,
		},
		{
			name:           "empty route mode treated as auto + overseas + HTTPS_PROXY returns proxy",
			routeMode:      "",
			providerRegion: regionOverseas,
			envHTTPSProxy:  "http://host.docker.internal:7890",
			expected:       GatewayModeProxy,
		},
		{
			name:           "unknown route mode returns direct",
			routeMode:      "unknown",
			providerRegion: regionOverseas,
			envHTTPSProxy:  "http://proxy:7890",
			expected:       GatewayModeDirect,
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

			result := ResolveRouteModeWithProvider(tt.routeMode, tt.providerRegion)
			assert.Equal(t, tt.expected, result)
		})
	}
}
