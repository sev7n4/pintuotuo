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

	"github.com/pintuotuo/backend/config"
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

// litellmBYOKPathReachable is true when the request reached the upstream via LiteLLM+BYOK (auth/quota errors count).
func litellmBYOKPathReachable(statusCode int) bool {
	if statusCode >= 200 && statusCode < 300 {
		return true
	}
	switch statusCode {
	case http.StatusUnauthorized, http.StatusForbidden, http.StatusPaymentRequired, http.StatusTooManyRequests:
		return true
	default:
		return false
	}
}

// probeLitellmBYOKChat probes chat via gateway user_config. billingOK is true only on 2xx; pathOK includes auth/quota responses.
func probeLitellmBYOKChat(
	ctx context.Context,
	client *http.Client,
	chatEndpoint, provider, catalogModel, decryptedBYOK, litellmMasterKey string,
) (pathOK, billingOK bool, statusCode int, errCode, errMsg string) {
	if client == nil {
		client = newProxyAwareHTTPClient(30*time.Second, GatewayModeLitellm)
	}
	upstreamBaseURL := ResolveProbeUpstreamBaseURL(provider)
	if upstreamBaseURL == "" {
		return false, false, 0, "QUOTA_PROBE_UPSTREAM_URL_MISSING", "upstream base URL not configured"
	}
	catalogModel = strings.TrimSpace(catalogModel)
	if catalogModel == "" {
		catalogModel = defaultCatalogProbeModel(context.Background(), provider)
	}
	userConfig := BuildLitellmUserConfig(provider, catalogModel, decryptedBYOK, upstreamBaseURL)
	gatewayModel := LitellmGatewayRequestModel(provider, catalogModel)
	if gatewayModel == "" {
		gatewayModel = catalogModel
	}
	body := map[string]interface{}{
		"model":       gatewayModel,
		"messages":    []map[string]string{{"role": "user", "content": "ping"}},
		"max_tokens":  1,
		"user_config": userConfig,
	}
	jsonBody, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, chatEndpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return false, false, 0, ErrCodeQuotaProbeRequestBuildFailed, err.Error()
	}
	req.Header.Set("Authorization", "Bearer "+litellmMasterKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return false, false, 0, ErrCodeQuotaProbeNetworkError, err.Error()
	}
	defer resp.Body.Close()
	statusCode = resp.StatusCode
	pathOK = litellmBYOKPathReachable(statusCode)
	billingOK = statusCode >= 200 && statusCode < 300
	if billingOK {
		return pathOK, billingOK, statusCode, "", ""
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
	return pathOK, billingOK, statusCode, errCode, errMsg
}

// ProbeLitellmBYOKChatCompletion sends a minimal chat via LiteLLM with user_config (2xx only).
func ProbeLitellmBYOKChatCompletion(
	ctx context.Context,
	client *http.Client,
	chatEndpoint, provider, catalogModel, decryptedBYOK, litellmMasterKey string,
) (ok bool, statusCode int, errCode, errMsg string) {
	pathOK, billingOK, statusCode, errCode, errMsg := probeLitellmBYOKChat(ctx, client, chatEndpoint, provider, catalogModel, decryptedBYOK, litellmMasterKey)
	if !pathOK {
		return false, statusCode, errCode, errMsg
	}
	return billingOK, statusCode, errCode, errMsg
}

func probeLitellmBYOKChatWithCandidates(
	ctx context.Context,
	client *http.Client,
	chatEndpoint, provider, decryptedBYOK, litellmMasterKey string,
	candidates []string,
) (pathOK, billingOK bool, statusCode int, errCode, errMsg, modelUsed string) {
	for _, model := range candidates {
		pathOK, billingOK, statusCode, errCode, errMsg = probeLitellmBYOKChat(ctx, client, chatEndpoint, provider, model, decryptedBYOK, litellmMasterKey)
		if pathOK {
			return pathOK, billingOK, statusCode, errCode, errMsg, model
		}
		if statusCode == http.StatusNotFound {
			continue
		}
	}
	return false, false, statusCode, errCode, errMsg, ""
}

// ProbeLitellmBYOKModels lists models for a BYOK key under route_mode=litellm.
// Order: BYOK GET upstream (when reachable) → gateway chat path + platform catalog list → fail.
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
	upstream := ResolveProbeUpstreamBaseURL(provider)
	chatURL := OpenAICompatChatCompletionsURL(gatewayV1Base)
	candidates := CatalogProbeCandidates(ctx, provider)

	if ProviderUsesAnthropicHTTP(provider, apiFormat) {
		return probeLitellmAnthropicBYOKModels(ctx, client, chatURL, masterKey, provider, decryptedBYOK, openAISiblingBase, upstream, candidates)
	}

	if upstream != "" {
		probe, err := ProbeProviderModels(ctx, client, OpenAICompatModelsProbeURL(upstream), decryptedBYOK, provider)
		if probe != nil {
			probe.Models = FilterBYOKModelsForProvider(provider, probe.Models)
			if probe.Success && len(probe.Models) > 0 {
				return probe, err
			}
		}
	}

	pathOK, _, statusCode, code, msg, _ := probeLitellmBYOKChatWithCandidates(ctx, client, chatURL, provider, decryptedBYOK, masterKey, candidates)
	if pathOK {
		models := FilterBYOKModelsForProvider(provider, ProbeCatalogModelsForProvider(ctx, provider))
		return &ProbeModelsResult{Success: true, StatusCode: statusCode, Models: models}, nil
	}
	return &ProbeModelsResult{Success: false, StatusCode: statusCode, ErrorCode: code, ErrorMsg: msg}, nil
}

func probeLitellmAnthropicBYOKModels(
	ctx context.Context,
	client *http.Client,
	chatURL, masterKey, provider, decryptedBYOK, openAISiblingBase, upstream string,
	candidates []string,
) (*ProbeModelsResult, error) {
	pathOK, _, statusCode, code, msg, _ := probeLitellmBYOKChatWithCandidates(ctx, client, chatURL, provider, decryptedBYOK, masterKey, candidates)
	if !pathOK {
		return &ProbeModelsResult{Success: false, StatusCode: statusCode, ErrorCode: code, ErrorMsg: msg}, nil
	}

	models := FilterBYOKModelsForProvider(provider, ProbeCatalogModelsForProvider(ctx, provider))
	siblingBase := normalizeOpenAICompatBase(openAISiblingBase)
	if siblingBase != "" && upstream != "" {
		if modelsProbe, err := ProbeProviderModels(ctx, client, OpenAICompatModelsProbeURL(siblingBase), decryptedBYOK, provider); err == nil && modelsProbe != nil && modelsProbe.Success {
			if merged := FilterBYOKModelsForProvider(provider, modelsProbe.Models); len(merged) > 0 {
				models = merged
			}
		}
	}
	if len(models) == 0 {
		return &ProbeModelsResult{Success: false, StatusCode: statusCode, ErrorCode: "PROBE_MODEL_LIST_EMPTY", ErrorMsg: "no catalog models for provider"}, nil
	}
	return &ProbeModelsResult{Success: true, StatusCode: statusCode, Models: models}, nil
}

// ResolveProbeUpstreamBaseURL returns vendor api_base for BYOK probes (litellm cache, then model_providers).
func ResolveProbeUpstreamBaseURL(provider string) string {
	if u := ResolveLitellmUpstreamBaseURL(provider); u != "" {
		return u
	}
	return providerAPIBaseURLFromDB(provider)
}

func providerAPIBaseURLFromDB(provider string) string {
	db := config.GetDB()
	if db == nil {
		return ""
	}
	for _, pcode := range catalogProviderCodes(provider) {
		var apiBase string
		err := db.QueryRow(
			`SELECT COALESCE(NULLIF(TRIM(api_base_url), ''), '')
			 FROM model_providers WHERE lower(trim(code)) = lower(trim($1)) AND status = 'active'`,
			pcode,
		).Scan(&apiBase)
		if err == nil {
			apiBase = strings.TrimRight(strings.TrimSpace(apiBase), "/")
			if apiBase != "" {
				return apiBase
			}
		}
	}
	return ""
}

func formatHTTPStatusCode(code int) string {
	if code <= 0 {
		return "HTTP_000"
	}
	return fmt.Sprintf("HTTP_%d", code)
}
