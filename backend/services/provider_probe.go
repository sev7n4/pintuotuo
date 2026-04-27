package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
func ProbeProviderModels(ctx context.Context, client *http.Client, modelsURL string, apiKey string) (*ProbeModelsResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, modelsURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	req.Header.Set("Content-Type", "application/json")

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
		if id != "" {
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
		Timeout: time.Duration(timeoutMs) * time.Millisecond,
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
