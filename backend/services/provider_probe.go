package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

// ProbeModelsResult is a shared probe output for validator and health checker.
type ProbeModelsResult struct {
	Success           bool
	StatusCode        int
	LatencyMs         int
	Models            []string
	ErrorCode         string
	ErrorMsg          string
	ErrorCategory     string
	ProviderRequestID string
	RawErrorExcerpt   string
}

// ProbeProviderModels performs one GET <modelsURL> probe and parses OpenAI-style model list payload.
func SetProviderAuthHeaders(req *http.Request, provider, apiKey string) {
	if ProviderUsesAnthropicHTTP(provider, "") {
		req.Header.Set("x-api-key", apiKey)
		req.Header.Set("anthropic-version", "2023-06-01")
	} else {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	}
	req.Header.Set("Content-Type", "application/json")
}

// AnthropicMessagesProbeURL returns POST URL for Anthropic Messages API given a provider base URL.
func AnthropicMessagesProbeURL(endpoint string) string {
	b := normalizeOpenAICompatBase(endpoint)
	if hasOpenAICompatVersionedRootSuffix(b) {
		return b + "/messages"
	}
	return b + "/v1/messages"
}

// ProbeProviderConnectivity probes upstream reachability: OpenAI GET /models or Anthropic POST /messages.
func ProbeProviderConnectivity(ctx context.Context, client *http.Client, baseURL, apiKey, provider, apiFormat string) (*ProbeModelsResult, error) {
	baseURL = normalizeOpenAICompatBase(baseURL)
	if baseURL == "" {
		return &ProbeModelsResult{
			Success:   false,
			ErrorMsg:  "endpoint is empty",
			ErrorCode: "PROBE_CONFIG_ERROR",
		}, nil
	}
	if ProviderUsesAnthropicHTTP(provider, apiFormat) {
		model := selectProbeModel(provider, nil, "")
		return ProbeProviderAnthropicMessages(ctx, client, baseURL, apiKey, provider, model)
	}
	return ProbeProviderModels(ctx, client, OpenAICompatModelsProbeURL(baseURL), apiKey, provider)
}

// ProbeProviderAnthropicMessages sends a minimal Messages request (same auth path as proxy anthropic upstream).
func ProbeProviderAnthropicMessages(ctx context.Context, client *http.Client, baseURL, apiKey, provider, model string) (*ProbeModelsResult, error) {
	if model == "" {
		model = selectProbeModel(provider, nil, "")
	}
	if model == "" {
		return &ProbeModelsResult{
			Success:   false,
			ErrorCode: "PROBE_MODEL_MISSING",
			ErrorMsg:  "no model available for anthropic messages probe",
		}, nil
	}

	body := map[string]interface{}{
		"model":      model,
		"max_tokens": 1,
		"messages": []map[string]string{
			{"role": "user", "content": "ping"},
		},
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	messagesURL := AnthropicMessagesProbeURL(baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, messagesURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	SetProviderAuthHeaders(req, provider, apiKey)

	start := time.Now()
	resp, err := client.Do(req)
	latencyMs := int(time.Since(start).Milliseconds())
	if err != nil {
		errInfo := MapProviderError(0, "", fmt.Sprintf("connection failed: %v", err), nil, err, "")
		return &ProbeModelsResult{
			Success:       false,
			LatencyMs:     latencyMs,
			ErrorMsg:      errInfo.ProviderMessage,
			ErrorCategory: errInfo.Category,
		}, err
	}
	defer resp.Body.Close()

	rawBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	result := &ProbeModelsResult{
		StatusCode: resp.StatusCode,
		LatencyMs:  latencyMs,
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		code, msg := ExtractProviderError(rawBody)
		if code == "" {
			code = fmt.Sprintf("HTTP_%d", resp.StatusCode)
		}
		if msg == "" {
			msg = strings.TrimSpace(string(rawBody))
		}
		errInfo := MapProviderError(resp.StatusCode, code, msg, resp.Header, nil, string(rawBody))
		result.ErrorCode = code
		result.ErrorMsg = msg
		result.ErrorCategory = errInfo.Category
		result.ProviderRequestID = errInfo.ProviderRequestID
		result.RawErrorExcerpt = errInfo.RawErrorExcerpt
		return result, nil
	}

	result.Success = true
	models := GetPredefinedModels(provider)
	if len(models) == 0 {
		models = []string{model}
	}
	sort.Strings(models)
	result.Models = models
	return result, nil
}

// ProbeAnthropicMessagesQuota checks billable access via a minimal POST /messages (deep verification).
func ProbeAnthropicMessagesQuota(ctx context.Context, client *http.Client, baseURL, apiKey, provider, model string) (bool, string, string) {
	probe, err := ProbeProviderAnthropicMessages(ctx, client, baseURL, apiKey, provider, model)
	if err != nil {
		return false, ErrCodeQuotaProbeNetworkError, err.Error()
	}
	if probe != nil && probe.Success {
		return true, "", ""
	}
	code := "ANTHROPIC_PROBE_FAILED"
	msg := "anthropic messages probe failed"
	if probe != nil {
		if strings.TrimSpace(probe.ErrorCode) != "" {
			code = probe.ErrorCode
		}
		if strings.TrimSpace(probe.ErrorMsg) != "" {
			msg = probe.ErrorMsg
		}
	}
	return false, code, msg
}

func ProbeProviderModels(ctx context.Context, client *http.Client, modelsURL string, apiKey string, provider string) (*ProbeModelsResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, modelsURL, nil)
	if err != nil {
		return nil, err
	}
	SetProviderAuthHeaders(req, provider, apiKey)

	start := time.Now()
	resp, err := client.Do(req)
	latencyMs := int(time.Since(start).Milliseconds())
	if err != nil {
		errInfo := MapProviderError(0, "", fmt.Sprintf("connection failed: %v", err), nil, err, "")
		return &ProbeModelsResult{
			Success:       false,
			LatencyMs:     latencyMs,
			ErrorMsg:      errInfo.ProviderMessage,
			ErrorCategory: errInfo.Category,
		}, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	result := &ProbeModelsResult{
		StatusCode: resp.StatusCode,
		LatencyMs:  latencyMs,
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		code, msg := ExtractProviderError(body)
		if code == "" {
			code = fmt.Sprintf("HTTP_%d", resp.StatusCode)
		}
		if msg == "" {
			msg = strings.TrimSpace(string(body))
		}
		errInfo := MapProviderError(resp.StatusCode, code, msg, resp.Header, nil, string(body))
		result.ErrorCode = code
		result.ErrorMsg = msg
		result.ErrorCategory = errInfo.Category
		result.ProviderRequestID = errInfo.ProviderRequestID
		result.RawErrorExcerpt = errInfo.RawErrorExcerpt
		return result, nil
	}

	models, parseErr := parseOpenAIModels(body)
	if parseErr != nil {
		result.ErrorMsg = fmt.Sprintf("failed to parse models: %v", parseErr)
		return result, nil
	}
	result.Success = true
	sort.Strings(models)
	result.Models = models
	return result, nil
}

func parseOpenAIModels(body []byte) ([]string, error) {
	var modelsResp struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &modelsResp); err != nil {
		return nil, err
	}
	models := make([]string, 0, len(modelsResp.Data))
	for _, m := range modelsResp.Data {
		id := strings.TrimSpace(m.ID)
		if id != "" && !strings.Contains(id, "*") {
			models = append(models, id)
		}
	}
	return models, nil
}

type ProbeURLResult struct {
	Success    bool
	StatusCode int
	LatencyMs  int
	ErrorCode  string
	ErrorMsg   string
}

func ProbeEndpointURL(ctx context.Context, url string, apiKey string, timeoutMs int) *ProbeURLResult {
	client := &http.Client{
		Timeout:   time.Duration(timeoutMs) * time.Millisecond,
		Transport: http.DefaultTransport.(*http.Transport).Clone(),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return &ProbeURLResult{
			Success:  false,
			ErrorMsg: fmt.Sprintf("failed to create request: %v", err),
		}
	}

	if apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	start := time.Now()
	resp, err := client.Do(req)
	latencyMs := int(time.Since(start).Milliseconds())

	if err != nil {
		errMsg := err.Error()
		var errCode string
		if strings.Contains(errMsg, "timeout") {
			errCode = "TIMEOUT"
			errMsg = fmt.Sprintf("connection timeout after %dms", timeoutMs)
		} else if strings.Contains(errMsg, "connection refused") {
			errCode = "CONNECTION_REFUSED"
		} else if strings.Contains(errMsg, "no such host") || strings.Contains(errMsg, "lookup") {
			errCode = "DNS_ERROR"
		} else {
			errCode = "CONNECTION_ERROR"
		}
		return &ProbeURLResult{
			Success:   false,
			LatencyMs: latencyMs,
			ErrorCode: errCode,
			ErrorMsg:  errMsg,
		}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	_ = body

	result := &ProbeURLResult{
		StatusCode: resp.StatusCode,
		LatencyMs:  latencyMs,
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.Success = true
	} else {
		result.Success = false
		result.ErrorCode = fmt.Sprintf("HTTP_%d", resp.StatusCode)
		result.ErrorMsg = fmt.Sprintf("endpoint returned status %d", resp.StatusCode)
	}

	return result
}

var predefinedModels = map[string][]string{
	"openai":            {"gpt-4o", "gpt-4o-mini", "gpt-4-turbo", "gpt-3.5-turbo"},
	"anthropic":         {"claude-3-5-sonnet-20241022", "claude-3-opus-20240229", "claude-3-haiku-20240307"},
	"stepfun":           {"stepfun-step-1-8k", "stepfun-step-1-32k", "stepfun-step-2-16k"},
	"deepseek":          {"deepseek-chat", "deepseek-coder"},
	"moonshot":          {"moonshot-v1-8k", "moonshot-v1-32k", "moonshot-v1-128k"},
	"zhipu":             {"glm-4", "glm-4-flash", "glm-3-turbo"},
	"alibaba":           {"qwen-turbo", "qwen-plus", "qwen-max"},
	"alibaba_anthropic": {"qwen-plus", "qwen-turbo", "glm-4"},
	"qwen":              {"qwen-turbo", "qwen-plus", "qwen-max"},
	"google":            {"gemini-2.0-flash", "gemini-1.5-pro", "gemini-1.5-flash"},
}

func GetPredefinedModels(provider string) []string {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if models, ok := predefinedModels[provider]; ok {
		return models
	}
	if strings.HasSuffix(provider, AnthropicSiblingProviderSuffix) {
		primary := strings.TrimSuffix(provider, AnthropicSiblingProviderSuffix)
		if models, ok := predefinedModels[primary]; ok {
			return models
		}
	}
	return []string{}
}
