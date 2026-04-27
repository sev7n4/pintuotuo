package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type ExecutionEngine struct {
	httpClient  *http.Client
	retryPolicy *RetryPolicy
}

type ExecutionInput struct {
	Provider        string            `json:"provider"`
	Model           string            `json:"model"`
	APIKey          string            `json:"api_key"`
	EndpointURL     string            `json:"endpoint_url"`
	RequestFormat   string            `json:"request_format"`
	RequestBody     []byte            `json:"request_body"`
	Headers         map[string]string `json:"headers"`
	Messages        []Message         `json:"messages"`
	Stream          bool              `json:"stream"`
	Options         json.RawMessage   `json:"options"`
	OriginalAPIKey  string            `json:"original_api_key"`
	GatewayMode     string            `json:"gateway_mode"`
	ProviderBaseURL string            `json:"provider_base_url"`
	FallbackURL     string            `json:"fallback_url"`
}

type ExecutionResult struct {
	Success      bool              `json:"success"`
	StatusCode   int               `json:"status_code"`
	LatencyMs    int               `json:"latency_ms"`
	ResponseBody []byte            `json:"response_body"`
	ErrorMessage string            `json:"error_message"`
	Usage        *TokenUsage       `json:"usage"`
	Provider     string            `json:"provider"`
	ActualModel  string            `json:"actual_model"`
	Headers      map[string]string `json:"headers"`
}

type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ExecutionEngineOption func(*ExecutionEngine)

func WithHTTPClient(client *http.Client) ExecutionEngineOption {
	return func(e *ExecutionEngine) {
		e.httpClient = client
	}
}

func WithExecutionRetryPolicy(policy *RetryPolicy) ExecutionEngineOption {
	return func(e *ExecutionEngine) {
		e.retryPolicy = policy
	}
}

func NewExecutionEngine(opts ...ExecutionEngineOption) *ExecutionEngine {
	engine := &ExecutionEngine{
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 20,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		retryPolicy: DefaultRetryPolicy,
	}
	for _, opt := range opts {
		opt(engine)
	}
	return engine
}

func (e *ExecutionEngine) Execute(ctx context.Context, input *ExecutionInput) (*ExecutionResult, error) {
	startTime := time.Now()

	req, err := e.buildHTTPRequest(ctx, input)
	if err != nil {
		return &ExecutionResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to build request: %v", err),
		}, err
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return &ExecutionResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("request failed: %v", err),
			Provider:     input.Provider,
			ActualModel:  input.Model,
		}, err
	}
	defer resp.Body.Close()

	result, err := e.parseResponse(resp)
	if err != nil {
		return result, err
	}

	result.LatencyMs = int(time.Since(startTime).Milliseconds())
	result.Provider = input.Provider
	result.ActualModel = input.Model

	return result, nil
}

func (e *ExecutionEngine) ExecuteWithRetry(ctx context.Context, input *ExecutionInput) (*ExecutionResult, error) {
	var lastResult *ExecutionResult
	var lastErr error

	for attempt := 0; attempt <= e.retryPolicy.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := e.retryPolicy.DelayForAttempt(attempt - 1)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		result, err := e.Execute(ctx, input)
		lastResult = result
		lastErr = err

		if err == nil && result.Success {
			return result, nil
		}

		if !e.shouldRetry(result.StatusCode, err) {
			return result, err
		}
	}

	return lastResult, lastErr
}

func (e *ExecutionEngine) shouldRetry(statusCode int, err error) bool {
	if err != nil {
		return true
	}

	retryableStatusCodes := []int{429, 500, 502, 503, 504}
	for _, code := range retryableStatusCodes {
		if statusCode == code {
			return true
		}
	}
	return false
}

func (e *ExecutionEngine) buildHTTPRequest(ctx context.Context, input *ExecutionInput) (*http.Request, error) {
	var requestBody []byte
	var err error

	if len(input.RequestBody) > 0 {
		requestBody = input.RequestBody
	} else {
		rb := map[string]interface{}{
			"model":    input.Model,
			"messages": input.Messages,
			"stream":   input.Stream,
		}
		if len(input.Options) > 0 {
			var options map[string]interface{}
			if unmarshalErr := json.Unmarshal(input.Options, &options); unmarshalErr == nil {
				for k, v := range options {
					rb[k] = v
				}
			}
		}
		requestBody, err = json.Marshal(rb)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	endpoint := input.EndpointURL
	if endpoint == "" {
		return nil, fmt.Errorf("endpoint URL is required")
	}

	if !strings.HasSuffix(endpoint, "/chat/completions") && !strings.HasSuffix(endpoint, "/messages") {
		if input.RequestFormat == modelProviderAnthropic {
			endpoint = strings.TrimRight(endpoint, "/") + "/messages"
		} else {
			endpoint = strings.TrimRight(endpoint, "/") + "/chat/completions"
		}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(requestBody)), nil
	}

	req.Header.Set("Content-Type", "application/json")

	switch input.RequestFormat {
	case modelProviderAnthropic:
		req.Header.Set("x-api-key", input.APIKey)
		req.Header.Set("anthropic-version", "2023-06-01")
	default:
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", input.APIKey))
	}

	for key, value := range input.Headers {
		req.Header.Set(key, value)
	}

	return req, nil
}

func (e *ExecutionEngine) parseResponse(resp *http.Response) (*ExecutionResult, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &ExecutionResult{
			Success:      false,
			StatusCode:   resp.StatusCode,
			ErrorMessage: fmt.Sprintf("failed to read response body: %v", err),
		}, err
	}

	result := &ExecutionResult{
		StatusCode:   resp.StatusCode,
		ResponseBody: body,
		Headers:      make(map[string]string),
	}

	for key, values := range resp.Header {
		if len(values) > 0 {
			result.Headers[key] = values[0]
		}
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.Success = true

		var response struct {
			Usage *struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			} `json:"usage"`
		}
		if json.Unmarshal(body, &response) == nil && response.Usage != nil {
			result.Usage = &TokenUsage{
				PromptTokens:     response.Usage.PromptTokens,
				CompletionTokens: response.Usage.CompletionTokens,
				TotalTokens:      response.Usage.TotalTokens,
			}
		}
	} else {
		result.Success = false
		result.ErrorMessage = string(body)
	}

	return result, nil
}

type StreamExecutionInput struct {
	ExecutionInput
	Writer  io.Writer
	Flusher http.Flusher
}

func (e *ExecutionEngine) ExecuteStream(ctx context.Context, input *StreamExecutionInput) error {
	req, err := e.buildHTTPRequest(ctx, &input.ExecutionInput)
	if err != nil {
		return fmt.Errorf("failed to build stream request: %w", err)
	}

	req.Header.Set("Accept", "text/event-stream")

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("stream request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("stream request returned status %d: %s", resp.StatusCode, string(body))
	}

	reader := resp.Body
	buf := make([]byte, 4096)

	for {
		n, err := reader.Read(buf)
		if n > 0 {
			if _, writeErr := input.Writer.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
			if input.Flusher != nil {
				input.Flusher.Flush()
			}
		}

		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}
