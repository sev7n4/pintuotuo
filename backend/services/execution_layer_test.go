package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExecutionLayer(t *testing.T) {
	engine := NewExecutionEngine()
	layer := NewExecutionLayer(nil, engine)

	assert.NotNil(t, layer)
	assert.Equal(t, engine, layer.engine)
}

func TestNewExecutionLayer_NilEngine(t *testing.T) {
	layer := NewExecutionLayer(nil, nil)
	assert.NotNil(t, layer)
	assert.NotNil(t, layer.engine)
}

func TestExecutionLayer_Execute_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"id":      "test-id",
			"model":   "gpt-4",
			"choices": []interface{}{},
			"usage": map[string]int{
				"prompt_tokens":     10,
				"completion_tokens": 20,
				"total_tokens":      30,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	engine := NewExecutionEngine()
	layer := NewExecutionLayer(nil, engine)

	decision := &RoutingDecision{
		SelectedModel: "gpt-4",
	}

	input := &ExecutionLayerInput{
		RoutingDecision: decision,
		ProviderConfig: &ExecutionProviderConfig{
			Code:       "openai",
			Name:       "OpenAI",
			APIBaseURL: server.URL,
			APIFormat:  "openai",
		},
		DecryptedAPIKey: "test-api-key",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
	}

	output, err := layer.Execute(context.Background(), input)

	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.True(t, output.Result.Success)
	assert.Equal(t, 200, output.Result.StatusCode)
	assert.NotNil(t, output.Result.Usage)
}

func TestExecutionLayer_Execute_MissingProviderConfig(t *testing.T) {
	layer := NewExecutionLayer(nil, nil)

	input := &ExecutionLayerInput{
		DecryptedAPIKey: "test-api-key",
	}

	output, err := layer.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "provider config is required")
}

func TestExecutionLayer_Execute_MissingAPIKey(t *testing.T) {
	layer := NewExecutionLayer(nil, nil)

	input := &ExecutionLayerInput{
		ProviderConfig: &ExecutionProviderConfig{
			Code:       "openai",
			APIBaseURL: "http://example.com",
		},
	}

	output, err := layer.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "decrypted API key is required")
}

func TestExecutionLayer_RecordExecutionResult(t *testing.T) {
	layer := NewExecutionLayer(nil, nil)

	decision := &RoutingDecision{}
	result := &ExecutionResult{
		Success:     true,
		StatusCode:  200,
		LatencyMs:   150,
		Provider:    "openai",
		ActualModel: "gpt-4",
		Usage: &TokenUsage{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
	}

	layer.recordExecutionResult(decision, result)

	assert.True(t, decision.ExecutionSuccess)
	assert.Equal(t, 200, decision.ExecutionStatusCode)
	assert.Equal(t, 150, decision.ExecutionLatencyMs)
	assert.NotNil(t, decision.ExecutionLayerResult)

	var resultData map[string]interface{}
	err := json.Unmarshal(decision.ExecutionLayerResult, &resultData)
	require.NoError(t, err)
	assert.True(t, resultData["success"].(bool))
	assert.Equal(t, float64(200), resultData["status_code"])
}

func TestExecutionLayer_RecordExecutionResult_NilDecision(t *testing.T) {
	layer := NewExecutionLayer(nil, nil)

	result := &ExecutionResult{
		Success:    true,
		StatusCode: 200,
	}

	assert.NotPanics(t, func() {
		layer.recordExecutionResult(nil, result)
	})
}

func TestExecutionLayer_RecordExecutionResult_NilResult(t *testing.T) {
	layer := NewExecutionLayer(nil, nil)

	decision := &RoutingDecision{}

	assert.NotPanics(t, func() {
		layer.recordExecutionResult(decision, nil)
	})
}

func TestExecutionLayer_Execute_AnthropicFormat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-api-key", r.Header.Get("x-api-key"))
		assert.Equal(t, "2023-06-01", r.Header.Get("anthropic-version"))

		response := map[string]interface{}{
			"id":    "test-id",
			"model": "claude-3",
			"content": []interface{}{
				map[string]string{"type": "text", "text": "Hello!"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	engine := NewExecutionEngine()
	layer := NewExecutionLayer(nil, engine)

	input := &ExecutionLayerInput{
		ProviderConfig: &ExecutionProviderConfig{
			Code:       "anthropic",
			Name:       "Anthropic",
			APIBaseURL: server.URL,
			APIFormat:  "anthropic",
		},
		DecryptedAPIKey: "test-api-key",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
	}

	output, err := layer.Execute(context.Background(), input)

	require.NoError(t, err)
	assert.True(t, output.Result.Success)
}

func TestExecutionLayer_Execute_WithRetry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	engine := NewExecutionEngine()
	layer := NewExecutionLayer(nil, engine)

	input := &ExecutionLayerInput{
		ProviderConfig: &ExecutionProviderConfig{
			Code:       "openai",
			APIBaseURL: server.URL,
			APIFormat:  "openai",
		},
		DecryptedAPIKey: "test-key",
	}

	output, err := layer.Execute(context.Background(), input)

	require.NoError(t, err)
	assert.True(t, output.Result.Success)
	assert.Equal(t, 2, attempts)
}

func TestExecutionLayer_GetProviderConfig(t *testing.T) {
	t.Run("nil database", func(t *testing.T) {
		layer := NewExecutionLayer(nil, nil)
		_, err := layer.GetProviderConfig(context.Background(), "openai")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database connection is required")
	})
}

func TestExecutionLayer_UpdateRoutingDecisionLog(t *testing.T) {
	t.Run("nil database", func(t *testing.T) {
		layer := NewExecutionLayer(nil, nil)
		err := layer.UpdateRoutingDecisionLog(context.Background(), 1, &ExecutionResult{})
		assert.NoError(t, err)
	})
}

func TestExecutionLayer_Execute_WithMessages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		assert.Contains(t, body, "messages")
		assert.Contains(t, body, "model")

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	engine := NewExecutionEngine()
	layer := NewExecutionLayer(nil, engine)

	input := &ExecutionLayerInput{
		ProviderConfig: &ExecutionProviderConfig{
			Code:       "openai",
			APIBaseURL: server.URL,
			APIFormat:  "openai",
		},
		DecryptedAPIKey: "test-key",
		Messages: []Message{
			{Role: "system", Content: "You are helpful"},
			{Role: "user", Content: "Hello"},
		},
		Stream: false,
		Options: json.RawMessage(`{
			"temperature": 0.7,
			"max_tokens": 100
		}`),
	}

	output, err := layer.Execute(context.Background(), input)

	require.NoError(t, err)
	assert.True(t, output.Result.Success)
}

func TestExecutionProviderConfig_NewFields(t *testing.T) {
	cfg := &ExecutionProviderConfig{
		Code:           "openai",
		Name:           "OpenAI",
		APIBaseURL:     "https://api.openai.com/v1",
		APIFormat:      "openai",
		GatewayMode:    "litellm",
		ProviderRegion: "overseas",
		RouteStrategy: map[string]interface{}{
			"domestic_users": map[string]interface{}{"mode": "litellm"},
			"overseas_users": map[string]interface{}{"mode": "direct"},
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

	assert.Equal(t, "openai", cfg.Code)
	assert.Equal(t, "litellm", cfg.GatewayMode)
	assert.Equal(t, "overseas", cfg.ProviderRegion)
	assert.NotNil(t, cfg.RouteStrategy)
	assert.NotNil(t, cfg.Endpoints)

	domesticUsers, ok := cfg.RouteStrategy["domestic_users"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "litellm", domesticUsers["mode"])

	overseasUsers, ok := cfg.RouteStrategy["overseas_users"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "direct", overseasUsers["mode"])
}

func TestExecutionProviderConfig_ProviderRegion_Default(t *testing.T) {
	cfg := &ExecutionProviderConfig{
		Code:       "openai",
		APIBaseURL: "https://api.openai.com/v1",
	}

	assert.Equal(t, "", cfg.ProviderRegion)
}

func TestExecutionProviderConfig_RouteStrategy_Nil(t *testing.T) {
	cfg := &ExecutionProviderConfig{
		Code:       "openai",
		APIBaseURL: "https://api.openai.com/v1",
	}

	assert.Nil(t, cfg.RouteStrategy)
}

func TestExecutionProviderConfig_Endpoints_Nil(t *testing.T) {
	cfg := &ExecutionProviderConfig{
		Code:       "openai",
		APIBaseURL: "https://api.openai.com/v1",
	}

	assert.Nil(t, cfg.Endpoints)
}

func TestExecutionLayer_ResolveEndpoint_DirectMode(t *testing.T) {
	layer := NewExecutionLayer(nil, nil)

	cfg := &ExecutionProviderConfig{
		Code:           "openai",
		APIBaseURL:     "https://api.openai.com/v1",
		GatewayMode:    "direct",
		ProviderRegion: "overseas",
	}

	endpoint := layer.resolveEndpoint(cfg)
	assert.Equal(t, "https://api.openai.com/v1", endpoint)
}

func TestExecutionLayer_ResolveEndpoint_LitellmMode(t *testing.T) {
	layer := NewExecutionLayer(nil, nil)

	cfg := &ExecutionProviderConfig{
		Code:           "openai",
		APIBaseURL:     "https://api.openai.com/v1",
		GatewayMode:    "litellm",
		ProviderRegion: "domestic",
		Endpoints: map[string]interface{}{
			"litellm": map[string]interface{}{
				"domestic": "http://litellm-domestic:4000/v1",
				"overseas": "http://litellm-overseas:4000/v1",
			},
		},
	}

	endpoint := layer.resolveEndpoint(cfg)
	assert.Equal(t, "http://litellm-domestic:4000/v1", endpoint)
}

func TestExecutionLayer_ResolveEndpoint_ProxyMode(t *testing.T) {
	layer := NewExecutionLayer(nil, nil)

	cfg := &ExecutionProviderConfig{
		Code:           "openai",
		APIBaseURL:     "https://api.openai.com/v1",
		GatewayMode:    "proxy",
		ProviderRegion: "domestic",
		Endpoints: map[string]interface{}{
			"proxy": map[string]interface{}{
				"gaap": "https://openai-gaap.example.com",
			},
		},
	}

	endpoint := layer.resolveEndpoint(cfg)
	assert.Equal(t, "https://openai-gaap.example.com", endpoint)
}

func TestExecutionLayer_ResolveEndpoint_FallbackToAPIBaseURL(t *testing.T) {
	layer := NewExecutionLayer(nil, nil)

	cfg := &ExecutionProviderConfig{
		Code:           "openai",
		APIBaseURL:     "https://api.openai.com/v1",
		GatewayMode:    "direct",
		ProviderRegion: "overseas",
		Endpoints:      nil,
	}

	endpoint := layer.resolveEndpoint(cfg)
	assert.Equal(t, "https://api.openai.com/v1", endpoint)
}

func TestExecutionLayer_ResolveAuthToken_DirectMode(t *testing.T) {
	layer := NewExecutionLayer(nil, nil)

	cfg := &ExecutionProviderConfig{
		Code:        "openai",
		GatewayMode: "direct",
	}

	token := layer.resolveAuthToken(cfg, "sk-original-key")
	assert.Equal(t, "sk-original-key", token)
}

func TestExecutionLayer_ResolveAuthToken_LitellmMode(t *testing.T) {
	layer := NewExecutionLayer(nil, nil)

	cfg := &ExecutionProviderConfig{
		Code:        "openai",
		GatewayMode: "litellm",
	}

	os.Setenv("LITELLM_MASTER_KEY", "sk-litellm-master-key")
	defer os.Unsetenv("LITELLM_MASTER_KEY")

	token := layer.resolveAuthToken(cfg, "sk-original-key")
	assert.Equal(t, "sk-litellm-master-key", token)
}

func TestExecutionLayer_ResolveAuthToken_ProxyMode(t *testing.T) {
	layer := NewExecutionLayer(nil, nil)

	cfg := &ExecutionProviderConfig{
		Code:        "openai",
		GatewayMode: "proxy",
	}

	os.Setenv("LLM_GATEWAY_PROXY_TOKEN", "sk-proxy-token")
	defer os.Unsetenv("LLM_GATEWAY_PROXY_TOKEN")

	token := layer.resolveAuthToken(cfg, "sk-original-key")
	assert.Equal(t, "sk-proxy-token", token)
}

func TestExecutionLayer_DetermineGatewayMode_FromConfig(t *testing.T) {
	layer := NewExecutionLayer(nil, nil)

	cfg := &ExecutionProviderConfig{
		Code:        "openai",
		GatewayMode: "litellm",
		RouteStrategy: map[string]interface{}{
			"domestic_users": map[string]interface{}{"mode": "litellm"},
		},
	}

	mode := layer.determineGatewayMode(cfg)
	assert.Equal(t, "litellm", mode)
}

func TestExecutionLayer_DetermineGatewayMode_EmptyConfig(t *testing.T) {
	layer := NewExecutionLayer(nil, nil)

	cfg := &ExecutionProviderConfig{
		Code:        "openai",
		GatewayMode: "",
	}

	mode := layer.determineGatewayMode(cfg)
	assert.Equal(t, "direct", mode)
}

func TestExecutionLayer_Execute_WithGatewayMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer sk-litellm-master", r.Header.Get("Authorization"))

		response := map[string]interface{}{
			"id":      "test-id",
			"model":   "gpt-4",
			"choices": []interface{}{},
			"usage": map[string]int{
				"prompt_tokens":     10,
				"completion_tokens": 20,
				"total_tokens":      30,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	os.Setenv("LITELLM_MASTER_KEY", "sk-litellm-master")
	defer os.Unsetenv("LITELLM_MASTER_KEY")

	engine := NewExecutionEngine()
	layer := NewExecutionLayer(nil, engine)

	input := &ExecutionLayerInput{
		ProviderConfig: &ExecutionProviderConfig{
			Code:           "openai",
			Name:           "OpenAI",
			APIBaseURL:     "https://api.openai.com/v1",
			APIFormat:      "openai",
			GatewayMode:    GatewayModeLitellm,
			ProviderRegion: "domestic",
			Endpoints: map[string]interface{}{
				GatewayModeLitellm: map[string]interface{}{
					"domestic": server.URL,
				},
			},
		},
		DecryptedAPIKey: "sk-original-key",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
	}

	output, err := layer.Execute(context.Background(), input)

	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.True(t, output.Result.Success)
	assert.Equal(t, 200, output.Result.StatusCode)
}

func TestExecutionLayer_Execute_DirectMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer sk-original-key", r.Header.Get("Authorization"))

		response := map[string]interface{}{
			"id":      "test-id",
			"model":   "gpt-4",
			"choices": []interface{}{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	engine := NewExecutionEngine()
	layer := NewExecutionLayer(nil, engine)

	input := &ExecutionLayerInput{
		ProviderConfig: &ExecutionProviderConfig{
			Code:           "openai",
			Name:           "OpenAI",
			APIBaseURL:     server.URL,
			APIFormat:      "openai",
			GatewayMode:    GatewayModeDirect,
			ProviderRegion: "overseas",
		},
		DecryptedAPIKey: "sk-original-key",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
	}

	output, err := layer.Execute(context.Background(), input)

	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.True(t, output.Result.Success)
}

func TestExecutionLayer_Execute_ProxyMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer sk-proxy-token", r.Header.Get("Authorization"))

		response := map[string]interface{}{
			"id":      "test-id",
			"model":   "gpt-4",
			"choices": []interface{}{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	os.Setenv("LLM_GATEWAY_PROXY_TOKEN", "sk-proxy-token")
	defer os.Unsetenv("LLM_GATEWAY_PROXY_TOKEN")

	engine := NewExecutionEngine()
	layer := NewExecutionLayer(nil, engine)

	input := &ExecutionLayerInput{
		ProviderConfig: &ExecutionProviderConfig{
			Code:        "openai",
			Name:        "OpenAI",
			APIBaseURL:  "https://api.openai.com/v1",
			APIFormat:   "openai",
			GatewayMode: GatewayModeProxy,
			Endpoints: map[string]interface{}{
				GatewayModeProxy: map[string]interface{}{
					"gaap": server.URL,
				},
			},
		},
		DecryptedAPIKey: "sk-original-key",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
	}

	output, err := layer.Execute(context.Background(), input)

	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.True(t, output.Result.Success)
}
