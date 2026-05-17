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

// FilterBYOKModelsForProvider keeps model ids that belong to the BYOK provider and drops gateway/noise entries.
func FilterBYOKModelsForProvider(provider string, models []string) []string {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if len(models) == 0 {
		return nil
	}
	out := make([]string, 0, len(models))
	seen := make(map[string]struct{}, len(models))
	for _, m := range models {
		m = strings.TrimSpace(m)
		if m == "" || isGatewayCatalogNoise(m) {
			continue
		}
		if !modelMatchesProviderProbe(provider, strings.ToLower(m)) {
			continue
		}
		if _, ok := seen[m]; ok {
			continue
		}
		seen[m] = struct{}{}
		out = append(out, m)
	}
	sort.Strings(out)
	return out
}

// ProbeLitellmBYOKChatCompletion sends a minimal chat via LiteLLM with user_config (BYOK path through gateway).
func ProbeLitellmBYOKChatCompletion(
	ctx context.Context,
	client *http.Client,
	chatEndpoint, provider, catalogModel, decryptedBYOK, litellmMasterKey string,
) (ok bool, statusCode int, errCode, errMsg string) {
	if client == nil {
		client = newProxyAwareHTTPClient(30*time.Second, GatewayModeLitellm)
	}
	upstreamBaseURL := ResolveLitellmUpstreamBaseURL(provider)
	if upstreamBaseURL == "" {
		return false, 0, "QUOTA_PROBE_UPSTREAM_URL_MISSING", "upstream base URL not configured"
	}
	catalogModel = strings.TrimSpace(catalogModel)
	if catalogModel == "" {
		catalogModel = defaultCatalogProbeModel(provider)
	}
	userConfig := BuildLitellmUserConfig(provider, catalogModel, decryptedBYOK, upstreamBaseURL)
	body := map[string]interface{}{
		"model":       catalogModel,
		"messages":    []map[string]string{{"role": "user", "content": "ping"}},
		"max_tokens":  1,
		"user_config": userConfig,
	}
	jsonBody, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, chatEndpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return false, 0, ErrCodeQuotaProbeRequestBuildFailed, err.Error()
	}
	req.Header.Set("Authorization", "Bearer "+litellmMasterKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return false, 0, ErrCodeQuotaProbeNetworkError, err.Error()
	}
	defer resp.Body.Close()
	statusCode = resp.StatusCode
	if statusCode >= 200 && statusCode < 300 {
		return true, statusCode, "", ""
	}
	rawBody, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
	errCode, errMsg = ExtractProviderError(rawBody)
	if errCode == "" {
		errCode = formatHTTPStatusCode(statusCode)
	}
	if errMsg == "" {
		errMsg = strings.TrimSpace(string(rawBody))
	}
	if statusCode == http.StatusPaymentRequired || statusCode == http.StatusTooManyRequests {
		errCode = errorCategoryQuotaInsufficient
	}
	return false, statusCode, errCode, errMsg
}

// ProbeLitellmBYOKModels lists models for a BYOK key via LiteLLM user_config, falling back to upstream GET with BYOK.
func ProbeLitellmBYOKModels(
	ctx context.Context,
	client *http.Client,
	gatewayV1Base, masterKey, provider, decryptedBYOK, apiFormat string,
	openAISiblingBase string,
) (*ProbeModelsResult, error) {
	if client == nil {
		client = newProxyAwareHTTPClient(15*time.Second, GatewayModeLitellm)
	}
	gatewayV1Base = normalizeOpenAICompatBase(gatewayV1Base)
	upstream := ResolveLitellmUpstreamBaseURL(provider)

	if ProviderUsesAnthropicHTTP(provider, apiFormat) {
		chatURL := OpenAICompatChatCompletionsURL(gatewayV1Base)
		ok, statusCode, code, msg := ProbeLitellmBYOKChatCompletion(ctx, client, chatURL, provider, "", decryptedBYOK, masterKey)
		models := anthropicProbeModelList(ctx, client, decryptedBYOK, provider, openAISiblingBase, defaultCatalogProbeModel(provider))
		models = FilterBYOKModelsForProvider(provider, models)
		result := &ProbeModelsResult{
			Success:    ok,
			StatusCode: statusCode,
			ErrorCode:  code,
			ErrorMsg:   msg,
			Models:     models,
		}
		if !ok && msg != "" {
			return result, nil
		}
		if len(models) > 0 {
			result.Success = true
		}
		return result, nil
	}

	catalogModel := defaultCatalogProbeModel(provider)
	userConfig := BuildLitellmUserConfig(provider, catalogModel, decryptedBYOK, upstream)

	// P3: LiteLLM clientside_auth — list models through gateway with user_config (same as domestic).
	if probe, err := probeLitellmGatewayModelsWithUserConfig(ctx, client, gatewayV1Base, masterKey, userConfig, provider); err == nil && probe != nil && probe.Success && len(probe.Models) > 0 {
		probe.Models = FilterBYOKModelsForProvider(provider, probe.Models)
		if len(probe.Models) > 0 {
			return probe, nil
		}
	}

	// Fallback: BYOK GET upstream /v1/models (same api_base as user_config).
	if upstream != "" {
		probe, err := ProbeProviderModels(ctx, client, OpenAICompatModelsProbeURL(upstream), decryptedBYOK, provider)
		if probe != nil {
			probe.Models = FilterBYOKModelsForProvider(provider, probe.Models)
			if probe.Success && len(probe.Models) > 0 {
				return probe, err
			}
		}
	}

	// Last resort: gateway BYOK chat path works → predefined catalog ids for this provider.
	chatURL := OpenAICompatChatCompletionsURL(gatewayV1Base)
	ok, statusCode, code, msg := ProbeLitellmBYOKChatCompletion(ctx, client, chatURL, provider, catalogModel, decryptedBYOK, masterKey)
	if ok {
		models := FilterBYOKModelsForProvider(provider, GetPredefinedModels(provider))
		return &ProbeModelsResult{Success: true, StatusCode: statusCode, Models: models}, nil
	}
	return &ProbeModelsResult{Success: false, StatusCode: statusCode, ErrorCode: code, ErrorMsg: msg}, nil
}

func probeLitellmGatewayModelsWithUserConfig(
	ctx context.Context,
	client *http.Client,
	gatewayV1Base, masterKey string,
	userConfig map[string]interface{},
	provider string,
) (*ProbeModelsResult, error) {
	modelsURL := OpenAICompatModelsProbeURL(gatewayV1Base)
	body, _ := json.Marshal(map[string]interface{}{"user_config": userConfig})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, modelsURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+masterKey)
	req.Header.Set("Content-Type", "application/json")
	return executeProbeModelsRequest(ctx, client, req, provider)
}

func executeProbeModelsRequest(ctx context.Context, client *http.Client, req *http.Request, provider string) (*ProbeModelsResult, error) {
	start := time.Now()
	resp, err := client.Do(req)
	latencyMs := int(time.Since(start).Milliseconds())
	if err != nil {
		errInfo := MapProviderError(0, "", "connection failed: "+err.Error(), nil, err, "")
		return &ProbeModelsResult{
			Success:       false,
			LatencyMs:     latencyMs,
			ErrorMsg:      errInfo.ProviderMessage,
			ErrorCategory: errInfo.Category,
		}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	result := &ProbeModelsResult{StatusCode: resp.StatusCode, LatencyMs: latencyMs}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		code, msg := ExtractProviderError(body)
		if code == "" {
			code = formatHTTPStatusCode(resp.StatusCode)
		}
		if msg == "" {
			msg = strings.TrimSpace(string(body))
		}
		errInfo := MapProviderError(resp.StatusCode, code, msg, resp.Header, nil, string(body))
		result.ErrorCode = code
		result.ErrorMsg = msg
		result.ErrorCategory = errInfo.Category
		return result, nil
	}
	models, parseErr := parseOpenAIModels(body)
	if parseErr != nil {
		result.ErrorMsg = "failed to parse models: " + parseErr.Error()
		return result, nil
	}
	result.Success = true
	result.Models = models
	return result, nil
}

func formatHTTPStatusCode(code int) string {
	if code <= 0 {
		return "HTTP_000"
	}
	return fmt.Sprintf("HTTP_%d", code)
}
