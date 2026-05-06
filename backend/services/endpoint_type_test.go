package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEndpointTypeConstants_Complete(t *testing.T) {
	expected := map[string]string{
		"EndpointTypeChatCompletions":     EndpointTypeChatCompletions,
		"EndpointTypeEmbeddings":          EndpointTypeEmbeddings,
		"EndpointTypeImagesGenerations":   EndpointTypeImagesGenerations,
		"EndpointTypeImagesVariations":    EndpointTypeImagesVariations,
		"EndpointTypeImagesEdits":         EndpointTypeImagesEdits,
		"EndpointTypeAudioTranscriptions": EndpointTypeAudioTranscriptions,
		"EndpointTypeAudioTranslations":   EndpointTypeAudioTranslations,
		"EndpointTypeAudioSpeech":         EndpointTypeAudioSpeech,
		"EndpointTypeModerations":         EndpointTypeModerations,
		"EndpointTypeResponses":           EndpointTypeResponses,
	}

	values := map[string]string{
		"EndpointTypeChatCompletions":     "chat_completions",
		"EndpointTypeEmbeddings":          "embeddings",
		"EndpointTypeImagesGenerations":   "images_generations",
		"EndpointTypeImagesVariations":    "images_variations",
		"EndpointTypeImagesEdits":         "images_edits",
		"EndpointTypeAudioTranscriptions": "audio_transcriptions",
		"EndpointTypeAudioTranslations":   "audio_translations",
		"EndpointTypeAudioSpeech":         "audio_speech",
		"EndpointTypeModerations":         "moderations",
		"EndpointTypeResponses":           "responses",
	}

	for name, constant := range expected {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, values[name], constant)
		})
	}
}

func TestEndpointPathSuffixes(t *testing.T) {
	expected := map[string]string{
		EndpointTypeChatCompletions:     "/v1/chat/completions",
		EndpointTypeEmbeddings:          "/v1/embeddings",
		EndpointTypeImagesGenerations:   "/v1/images/generations",
		EndpointTypeImagesVariations:    "/v1/images/variations",
		EndpointTypeImagesEdits:         "/v1/images/edits",
		EndpointTypeAudioTranscriptions: "/v1/audio/transcriptions",
		EndpointTypeAudioTranslations:   "/v1/audio/translations",
		EndpointTypeAudioSpeech:         "/v1/audio/speech",
		EndpointTypeModerations:         "/v1/moderations",
		EndpointTypeResponses:           "/v1/responses",
	}

	for endpointType, expectedSuffix := range expected {
		t.Run(endpointType, func(t *testing.T) {
			suffix, ok := endpointPathSuffixes[endpointType]
			assert.True(t, ok, "endpoint type %s should have a path suffix", endpointType)
			assert.Equal(t, expectedSuffix, suffix)
		})
	}
}

func TestResolveEndpointByType_AllEndpointTypes(t *testing.T) {
	cfg := &ExecutionProviderConfig{
		APIBaseURL:     "https://api.openai.com",
		GatewayMode:    GatewayModeDirect,
		ProviderRegion: regionOverseas,
	}

	tests := []struct {
		name         string
		endpointType string
		expected     string
	}{
		{
			name:         "chat_completions",
			endpointType: EndpointTypeChatCompletions,
			expected:     "https://api.openai.com/v1/chat/completions",
		},
		{
			name:         "embeddings",
			endpointType: EndpointTypeEmbeddings,
			expected:     "https://api.openai.com/v1/embeddings",
		},
		{
			name:         "images_generations",
			endpointType: EndpointTypeImagesGenerations,
			expected:     "https://api.openai.com/v1/images/generations",
		},
		{
			name:         "images_variations",
			endpointType: EndpointTypeImagesVariations,
			expected:     "https://api.openai.com/v1/images/variations",
		},
		{
			name:         "images_edits",
			endpointType: EndpointTypeImagesEdits,
			expected:     "https://api.openai.com/v1/images/edits",
		},
		{
			name:         "audio_transcriptions",
			endpointType: EndpointTypeAudioTranscriptions,
			expected:     "https://api.openai.com/v1/audio/transcriptions",
		},
		{
			name:         "audio_translations",
			endpointType: EndpointTypeAudioTranslations,
			expected:     "https://api.openai.com/v1/audio/translations",
		},
		{
			name:         "audio_speech",
			endpointType: EndpointTypeAudioSpeech,
			expected:     "https://api.openai.com/v1/audio/speech",
		},
		{
			name:         "moderations",
			endpointType: EndpointTypeModerations,
			expected:     "https://api.openai.com/v1/moderations",
		},
		{
			name:         "responses",
			endpointType: EndpointTypeResponses,
			expected:     "https://api.openai.com/v1/responses",
		},
		{
			name:         "unknown defaults to chat_completions",
			endpointType: "unknown_type",
			expected:     "https://api.openai.com/v1/chat/completions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolveEndpointByType(cfg, tt.endpointType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResolveEndpointByType_WithEndpoints(t *testing.T) {
	cfg := &ExecutionProviderConfig{
		APIBaseURL:     "https://api.openai.com",
		GatewayMode:    GatewayModeDirect,
		ProviderRegion: regionOverseas,
		Endpoints: map[string]interface{}{
			GatewayModeDirect: map[string]interface{}{
				regionOverseas: "https://api.openai.com/v1",
			},
		},
	}

	result := ResolveEndpointByType(cfg, EndpointTypeEmbeddings)
	assert.Equal(t, "https://api.openai.com/v1/embeddings", result)
}

func TestResolveEndpointByType_EmptyConfig(t *testing.T) {
	result := ResolveEndpointByType(nil, EndpointTypeChatCompletions)
	assert.Equal(t, "", result)
}

func TestRoutingRequest_EndpointType(t *testing.T) {
	req := &RoutingRequest{
		RequestID:    "test-123",
		EndpointType: EndpointTypeEmbeddings,
	}
	assert.Equal(t, EndpointTypeEmbeddings, req.EndpointType)
}
