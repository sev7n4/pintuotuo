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
		name           string
		routeMode      string
		providerRegion string
		envHTTPSProxy  string
		envLitellmURL  string
		expected       string
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

func TestEndpointTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{
			name:     "EndpointTypeChatCompletions",
			constant: EndpointTypeChatCompletions,
			expected: "chat_completions",
		},
		{
			name:     "EndpointTypeEmbeddings",
			constant: EndpointTypeEmbeddings,
			expected: "embeddings",
		},
		{
			name:     "EndpointTypeImagesGenerations",
			constant: EndpointTypeImagesGenerations,
			expected: "images_generations",
		},
		{
			name:     "EndpointTypeImagesVariations",
			constant: EndpointTypeImagesVariations,
			expected: "images_variations",
		},
		{
			name:     "EndpointTypeImagesEdits",
			constant: EndpointTypeImagesEdits,
			expected: "images_edits",
		},
		{
			name:     "EndpointTypeAudioSpeech",
			constant: EndpointTypeAudioSpeech,
			expected: "audio_speech",
		},
		{
			name:     "EndpointTypeAudioTranscriptions",
			constant: EndpointTypeAudioTranscriptions,
			expected: "audio_transcriptions",
		},
		{
			name:     "EndpointTypeAudioTranslations",
			constant: EndpointTypeAudioTranslations,
			expected: "audio_translations",
		},
		{
			name:     "EndpointTypeModerations",
			constant: EndpointTypeModerations,
			expected: "moderations",
		},
		{
			name:     "EndpointTypeResponses",
			constant: EndpointTypeResponses,
			expected: "responses",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.constant)
		})
	}
}

func TestResolveEndpointByType(t *testing.T) {
	tests := []struct {
		name         string
		cfg          *ExecutionProviderConfig
		endpointType string
		expected     string
	}{
		{
			name: "chat_completions endpoint from direct mode",
			cfg: &ExecutionProviderConfig{
				APIBaseURL:     "https://api.openai.com",
				GatewayMode:    GatewayModeDirect,
				ProviderRegion: regionOverseas,
				Endpoints: map[string]interface{}{
					GatewayModeDirect: map[string]interface{}{
						regionOverseas: "https://api.openai.com/v1",
					},
				},
			},
			endpointType: EndpointTypeChatCompletions,
			expected:     "https://api.openai.com/v1/chat/completions",
		},
		{
			name: "embeddings endpoint from direct mode",
			cfg: &ExecutionProviderConfig{
				APIBaseURL:     "https://api.openai.com",
				GatewayMode:    GatewayModeDirect,
				ProviderRegion: regionOverseas,
				Endpoints: map[string]interface{}{
					GatewayModeDirect: map[string]interface{}{
						regionOverseas: "https://api.openai.com/v1",
					},
				},
			},
			endpointType: EndpointTypeEmbeddings,
			expected:     "https://api.openai.com/v1/embeddings",
		},
		{
			name: "images_generations endpoint from direct mode",
			cfg: &ExecutionProviderConfig{
				APIBaseURL:     "https://api.openai.com",
				GatewayMode:    GatewayModeDirect,
				ProviderRegion: regionOverseas,
				Endpoints: map[string]interface{}{
					GatewayModeDirect: map[string]interface{}{
						regionOverseas: "https://api.openai.com/v1",
					},
				},
			},
			endpointType: EndpointTypeImagesGenerations,
			expected:     "https://api.openai.com/v1/images/generations",
		},
		{
			name: "images_variations endpoint from direct mode",
			cfg: &ExecutionProviderConfig{
				APIBaseURL:     "https://api.openai.com",
				GatewayMode:    GatewayModeDirect,
				ProviderRegion: regionOverseas,
				Endpoints: map[string]interface{}{
					GatewayModeDirect: map[string]interface{}{
						regionOverseas: "https://api.openai.com/v1",
					},
				},
			},
			endpointType: EndpointTypeImagesVariations,
			expected:     "https://api.openai.com/v1/images/variations",
		},
		{
			name: "images_edits endpoint from direct mode",
			cfg: &ExecutionProviderConfig{
				APIBaseURL:     "https://api.openai.com",
				GatewayMode:    GatewayModeDirect,
				ProviderRegion: regionOverseas,
				Endpoints: map[string]interface{}{
					GatewayModeDirect: map[string]interface{}{
						regionOverseas: "https://api.openai.com/v1",
					},
				},
			},
			endpointType: EndpointTypeImagesEdits,
			expected:     "https://api.openai.com/v1/images/edits",
		},
		{
			name: "audio_speech endpoint from direct mode",
			cfg: &ExecutionProviderConfig{
				APIBaseURL:     "https://api.openai.com",
				GatewayMode:    GatewayModeDirect,
				ProviderRegion: regionOverseas,
				Endpoints: map[string]interface{}{
					GatewayModeDirect: map[string]interface{}{
						regionOverseas: "https://api.openai.com/v1",
					},
				},
			},
			endpointType: EndpointTypeAudioSpeech,
			expected:     "https://api.openai.com/v1/audio/speech",
		},
		{
			name: "audio_transcriptions endpoint from direct mode",
			cfg: &ExecutionProviderConfig{
				APIBaseURL:     "https://api.openai.com",
				GatewayMode:    GatewayModeDirect,
				ProviderRegion: regionOverseas,
				Endpoints: map[string]interface{}{
					GatewayModeDirect: map[string]interface{}{
						regionOverseas: "https://api.openai.com/v1",
					},
				},
			},
			endpointType: EndpointTypeAudioTranscriptions,
			expected:     "https://api.openai.com/v1/audio/transcriptions",
		},
		{
			name: "audio_translations endpoint from direct mode",
			cfg: &ExecutionProviderConfig{
				APIBaseURL:     "https://api.openai.com",
				GatewayMode:    GatewayModeDirect,
				ProviderRegion: regionOverseas,
				Endpoints: map[string]interface{}{
					GatewayModeDirect: map[string]interface{}{
						regionOverseas: "https://api.openai.com/v1",
					},
				},
			},
			endpointType: EndpointTypeAudioTranslations,
			expected:     "https://api.openai.com/v1/audio/translations",
		},
		{
			name: "moderations endpoint from direct mode",
			cfg: &ExecutionProviderConfig{
				APIBaseURL:     "https://api.openai.com",
				GatewayMode:    GatewayModeDirect,
				ProviderRegion: regionOverseas,
				Endpoints: map[string]interface{}{
					GatewayModeDirect: map[string]interface{}{
						regionOverseas: "https://api.openai.com/v1",
					},
				},
			},
			endpointType: EndpointTypeModerations,
			expected:     "https://api.openai.com/v1/moderations",
		},
		{
			name: "fallback to APIBaseURL when no endpoints configured",
			cfg: &ExecutionProviderConfig{
				APIBaseURL:     "https://api.openai.com/v1",
				GatewayMode:    GatewayModeDirect,
				ProviderRegion: regionOverseas,
				Endpoints:      nil,
			},
			endpointType: EndpointTypeChatCompletions,
			expected:     "https://api.openai.com/v1/chat/completions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolveEndpointByType(tt.cfg, tt.endpointType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizeLegacyLitellmGatewayBaseURL(t *testing.T) {
	t.Run("no env leaves legacy host unchanged", func(t *testing.T) {
		t.Setenv("LLM_GATEWAY_LITELLM_URL", "")
		assert.Equal(t, "http://litellm-domestic:4000/v1", NormalizeLegacyLitellmGatewayBaseURL("http://litellm-domestic:4000/v1"))
	})
	t.Run("env rewrites litellm-domestic", func(t *testing.T) {
		t.Setenv("LLM_GATEWAY_LITELLM_URL", "http://litellm:4000")
		assert.Equal(t, "http://litellm:4000/v1", NormalizeLegacyLitellmGatewayBaseURL("http://litellm-domestic:4000/v1"))
	})
	t.Run("env rewrites litellm-overseas", func(t *testing.T) {
		t.Setenv("LLM_GATEWAY_LITELLM_URL", "http://litellm:4000/")
		assert.Equal(t, "http://litellm:4000/v1", NormalizeLegacyLitellmGatewayBaseURL("http://litellm-overseas:4000/v1"))
	})
	t.Run("custom upstream unchanged", func(t *testing.T) {
		t.Setenv("LLM_GATEWAY_LITELLM_URL", "http://litellm:4000")
		assert.Equal(t, "https://gw.example.com/v1", NormalizeLegacyLitellmGatewayBaseURL("https://gw.example.com/v1"))
	})
}

func TestResolveEndpointByType_LitellmLegacyHostUsesEnv(t *testing.T) {
	t.Setenv("LLM_GATEWAY_LITELLM_URL", "http://litellm:4000")
	cfg := &ExecutionProviderConfig{
		GatewayMode:    GatewayModeLitellm,
		ProviderRegion: regionDomestic,
		Endpoints: map[string]interface{}{
			GatewayModeLitellm: map[string]interface{}{
				"domestic": "http://litellm-domestic:4000/v1",
			},
		},
	}
	got := ResolveEndpointByType(cfg, EndpointTypeEmbeddings)
	assert.Equal(t, "http://litellm:4000/v1/embeddings", got)
}
