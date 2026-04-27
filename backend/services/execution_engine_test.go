package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExecutionEngine(t *testing.T) {
	engine := NewExecutionEngine()
	assert.NotNil(t, engine)
	assert.NotNil(t, engine.httpClient)
	assert.Equal(t, 3, engine.retryPolicy.MaxRetries)
	assert.Equal(t, 100*time.Millisecond, engine.retryPolicy.InitialDelay)
}

func TestNewExecutionEngineWithOptions(t *testing.T) {
	customClient := &http.Client{Timeout: 30 * time.Second}
	customPolicy := &RetryPolicy{
		MaxRetries:    5,
		InitialDelay:  200 * time.Millisecond,
		MaxDelay:      10 * time.Second,
		BackoffFactor: 2.0,
	}

	engine := NewExecutionEngine(
		WithHTTPClient(customClient),
		WithExecutionRetryPolicy(customPolicy),
	)

	assert.Equal(t, customClient, engine.httpClient)
	assert.Equal(t, 5, engine.retryPolicy.MaxRetries)
	assert.Equal(t, 200*time.Millisecond, engine.retryPolicy.InitialDelay)
}

func TestExecutionEngine_Execute_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

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
	input := &ExecutionInput{
		Provider:      "openai",
		Model:         "gpt-4",
		APIKey:        "test-api-key",
		EndpointURL:   server.URL,
		RequestFormat: "openai",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
		Stream: false,
	}

	result, err := engine.Execute(context.Background(), input)

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, 200, result.StatusCode)
	assert.NotNil(t, result.Usage)
	assert.Equal(t, 10, result.Usage.PromptTokens)
	assert.Equal(t, 20, result.Usage.CompletionTokens)
	assert.Equal(t, 30, result.Usage.TotalTokens)
}

func TestExecutionEngine_Execute_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer server.Close()

	engine := NewExecutionEngine()
	input := &ExecutionInput{
		Provider:      "openai",
		Model:         "gpt-4",
		APIKey:        "test-api-key",
		EndpointURL:   server.URL,
		RequestFormat: "openai",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
	}

	result, err := engine.Execute(context.Background(), input)

	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, 500, result.StatusCode)
	assert.Contains(t, result.ErrorMessage, "internal server error")
}

func TestExecutionEngine_Execute_Anthropic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-api-key", r.Header.Get("x-api-key"))
		assert.Equal(t, "2023-06-01", r.Header.Get("anthropic-version"))

		response := map[string]interface{}{
			"id":    "test-id",
			"model": "claude-3",
			"content": []interface{}{
				map[string]string{"type": "text", "text": "Hello!"},
			},
			"usage": map[string]int{
				"input_tokens":  10,
				"output_tokens": 20,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	engine := NewExecutionEngine()
	input := &ExecutionInput{
		Provider:      "anthropic",
		Model:         "claude-3",
		APIKey:        "test-api-key",
		EndpointURL:   server.URL,
		RequestFormat: "anthropic",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
	}

	result, err := engine.Execute(context.Background(), input)

	require.NoError(t, err)
	assert.True(t, result.Success)
}

func TestExecutionEngine_ExecuteWithRetry_Success(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	engine := NewExecutionEngine(WithExecutionRetryPolicy(&RetryPolicy{
		MaxRetries:    3,
		InitialDelay:  10 * time.Millisecond,
		MaxDelay:      100 * time.Millisecond,
		BackoffFactor: 1.0,
	}))

	input := &ExecutionInput{
		Provider:      "openai",
		Model:         "gpt-4",
		APIKey:        "test-api-key",
		EndpointURL:   server.URL,
		RequestFormat: "openai",
		Messages:      []Message{{Role: "user", Content: "Hello"}},
	}

	result, err := engine.ExecuteWithRetry(context.Background(), input)

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, 3, attempts)
}

func TestExecutionEngine_ExecuteWithRetry_MaxRetriesExceeded(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	engine := NewExecutionEngine(WithExecutionRetryPolicy(&RetryPolicy{
		MaxRetries:    2,
		InitialDelay:  10 * time.Millisecond,
		MaxDelay:      100 * time.Millisecond,
		BackoffFactor: 1.0,
	}))

	input := &ExecutionInput{
		Provider:      "openai",
		Model:         "gpt-4",
		APIKey:        "test-api-key",
		EndpointURL:   server.URL,
		RequestFormat: "openai",
		Messages:      []Message{{Role: "user", Content: "Hello"}},
	}

	result, err := engine.ExecuteWithRetry(context.Background(), input)

	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, 502, result.StatusCode)
	assert.Equal(t, 3, attempts)
}

func TestExecutionEngine_buildHTTPRequest(t *testing.T) {
	engine := NewExecutionEngine()

	tests := []struct {
		name          string
		input         *ExecutionInput
		expectHeader  string
		expectHeaderV string
	}{
		{
			name: "openai format",
			input: &ExecutionInput{
				Provider:      "openai",
				Model:         "gpt-4",
				APIKey:        "test-key",
				EndpointURL:   "http://api.example.com/v1",
				RequestFormat: "openai",
				Messages:      []Message{{Role: "user", Content: "Hello"}},
			},
			expectHeader:  "Authorization",
			expectHeaderV: "Bearer test-key",
		},
		{
			name: "anthropic format",
			input: &ExecutionInput{
				Provider:      "anthropic",
				Model:         "claude-3",
				APIKey:        "test-key",
				EndpointURL:   "http://api.example.com/v1",
				RequestFormat: "anthropic",
				Messages:      []Message{{Role: "user", Content: "Hello"}},
			},
			expectHeader:  "x-api-key",
			expectHeaderV: "test-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := engine.buildHTTPRequest(context.Background(), tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expectHeaderV, req.Header.Get(tt.expectHeader))
		})
	}
}

func TestExecutionEngine_buildRequestBody(t *testing.T) {
	engine := NewExecutionEngine()

	input := &ExecutionInput{
		Provider:    "openai",
		Model:       "gpt-4",
		EndpointURL: "http://api.example.com/v1/chat/completions",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
		Stream: false,
		Options: json.RawMessage(`{
			"temperature": 0.7,
			"max_tokens": 100
		}`),
	}

	req, err := engine.buildHTTPRequest(context.Background(), input)
	require.NoError(t, err)

	var body map[string]interface{}
	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&body)
	require.NoError(t, err)

	assert.Equal(t, "gpt-4", body["model"])
	assert.Equal(t, false, body["stream"])
	assert.Equal(t, 0.7, body["temperature"])
	assert.Equal(t, float64(100), body["max_tokens"])
}

func TestExecutionEngine_shouldRetry(t *testing.T) {
	engine := NewExecutionEngine()

	tests := []struct {
		name       string
		statusCode int
		err        error
		expected   bool
	}{
		{"success", 200, nil, false},
		{"client error", 400, nil, false},
		{"rate limit", 429, nil, true},
		{"server error", 500, nil, true},
		{"bad gateway", 502, nil, true},
		{"service unavailable", 503, nil, true},
		{"gateway timeout", 504, nil, true},
		{"with error", 0, assert.AnError, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.shouldRetry(tt.statusCode, tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExecutionEngine_Execute_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	engine := NewExecutionEngine()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	input := &ExecutionInput{
		Provider:      "openai",
		Model:         "gpt-4",
		APIKey:        "test-api-key",
		EndpointURL:   server.URL,
		RequestFormat: "openai",
		Messages:      []Message{{Role: "user", Content: "Hello"}},
	}

	result, err := engine.Execute(ctx, input)

	assert.Error(t, err)
	assert.False(t, result.Success)
}

func TestExecutionInput_NewFields(t *testing.T) {
	input := &ExecutionInput{
		Provider:        "openai",
		Model:           "gpt-4",
		APIKey:          "gateway-key",
		EndpointURL:     "http://litellm:4000/v1/chat/completions",
		RequestFormat:   "openai",
		OriginalAPIKey:  "original-provider-key",
		GatewayMode:     "litellm",
		ProviderBaseURL: "https://api.openai.com/v1",
		FallbackURL:     "http://proxy:8080/v1/chat/completions",
		Messages:        []Message{{Role: "user", Content: "Hello"}},
	}

	assert.Equal(t, "openai", input.Provider)
	assert.Equal(t, "gateway-key", input.APIKey)
	assert.Equal(t, "original-provider-key", input.OriginalAPIKey)
	assert.Equal(t, "litellm", input.GatewayMode)
	assert.Equal(t, "https://api.openai.com/v1", input.ProviderBaseURL)
	assert.Equal(t, "http://proxy:8080/v1/chat/completions", input.FallbackURL)
}

func TestExecutionInput_NewFields_Default(t *testing.T) {
	input := &ExecutionInput{
		Provider:    "openai",
		APIKey:      "test-key",
		EndpointURL: "https://api.openai.com/v1/chat/completions",
	}

	assert.Equal(t, "", input.OriginalAPIKey)
	assert.Equal(t, "", input.GatewayMode)
	assert.Equal(t, "", input.ProviderBaseURL)
	assert.Equal(t, "", input.FallbackURL)
}

func TestExecutionEngine_ExecuteStream_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "text/event-stream", r.Header.Get("Accept"))

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("data: {\"content\": \"Hello\"}\n\n"))
	}))
	defer server.Close()

	engine := NewExecutionEngine()
	input := &ExecutionInput{
		Provider:      "openai",
		Model:         "gpt-4",
		APIKey:        "test-api-key",
		EndpointURL:   server.URL,
		RequestFormat: "openai",
		Stream:        true,
	}

	result, err := engine.ExecuteStream(context.Background(), input)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 200, result.StatusCode)
	assert.NotNil(t, result.Body)
	defer result.Body.Close()
}

func TestExecutionEngine_ExecuteStream_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer server.Close()

	engine := NewExecutionEngine()
	input := &ExecutionInput{
		Provider:      "openai",
		Model:         "gpt-4",
		APIKey:        "test-api-key",
		EndpointURL:   server.URL,
		RequestFormat: "openai",
		Stream:        true,
	}

	result, err := engine.ExecuteStream(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "500")
}

func TestExecutionEngine_ExecuteStreamWithRetry_Success(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("data: {\"content\": \"Hello\"}\n\n"))
	}))
	defer server.Close()

	engine := NewExecutionEngine(WithExecutionRetryPolicy(&RetryPolicy{
		MaxRetries:    3,
		InitialDelay:  10 * time.Millisecond,
		MaxDelay:      100 * time.Millisecond,
		BackoffFactor: 1.0,
	}))

	input := &ExecutionInput{
		Provider:      "openai",
		Model:         "gpt-4",
		APIKey:        "test-api-key",
		EndpointURL:   server.URL,
		RequestFormat: "openai",
		Stream:        true,
	}

	result, err := engine.ExecuteStreamWithRetry(context.Background(), input)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 200, result.StatusCode)
	assert.Equal(t, 2, attempts)
	result.Body.Close()
}
