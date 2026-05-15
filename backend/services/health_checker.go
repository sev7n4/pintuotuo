package services

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pintuotuo/backend/billing"
	"github.com/pintuotuo/backend/cache"
	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/logger"
	"github.com/pintuotuo/backend/models"
	"github.com/pintuotuo/backend/utils"
)

type HealthCheckLevel string

const (
	HealthCheckLevelHigh   HealthCheckLevel = "high"
	HealthCheckLevelMedium HealthCheckLevel = "medium"
	HealthCheckLevelLow    HealthCheckLevel = "low"
	HealthCheckLevelDaily  HealthCheckLevel = "daily"
)

const (
	HealthStatusHealthy   = "healthy"
	HealthStatusDegraded  = "degraded"
	HealthStatusUnhealthy = "unhealthy"
	HealthStatusUnknown   = "unknown"
)

const fallbackProviderCode = "__default__"

// normalizeOpenAICompatBase trims the resolved provider/merchant base URL for path joining.
func normalizeOpenAICompatBase(endpoint string) string {
	return strings.TrimRight(strings.TrimSpace(endpoint), "/")
}

// OpenAICompatModelsProbeURL returns the GET URL for OpenAI-compatible model listing.
// When the base is already a versioned OpenAI-compat root (…/v1, …/v4, 智谱 paas/v4, 阿里 compatible-mode/v1, etc.),
// append "/models" only — matching api_key_validator (base + "/models") and avoiding paths like "…/v4/v1/models".
func OpenAICompatModelsProbeURL(endpoint string) string {
	b := normalizeOpenAICompatBase(endpoint)
	if hasOpenAICompatVersionedRootSuffix(b) {
		return b + "/models"
	}
	return b + "/v1/models"
}

// OpenAICompatChatCompletionsURL returns the POST URL for OpenAI-compatible chat completions.
func OpenAICompatChatCompletionsURL(endpoint string) string {
	b := normalizeOpenAICompatBase(endpoint)
	if hasOpenAICompatVersionedRootSuffix(b) {
		return b + "/chat/completions"
	}
	return b + "/v1/chat/completions"
}

// hasOpenAICompatVersionedRootSuffix reports whether the path already ends with a typical
// OpenAI-style API version segment. Order matters: check longer tokens before "/v1" to avoid
// false positives on hostnames like *ev1* (we only match a slash before the version).
func hasOpenAICompatVersionedRootSuffix(base string) bool {
	b := strings.ToLower(base)
	// v4: 智谱 `…/api/paas/v4`; v1: 多数厂商、DashScope compatible-mode/v1、MiniMax 国际版 等
	for _, suf := range []string{"/v4", "/v3", "/v2", "/v1"} {
		if strings.HasSuffix(b, suf) {
			return true
		}
	}
	return false
}

var healthCheckIntervalMap = map[HealthCheckLevel]int{
	HealthCheckLevelHigh:   60,
	HealthCheckLevelMedium: 300,
	HealthCheckLevelLow:    1800,
	HealthCheckLevelDaily:  86400,
}

type ProviderHealth struct {
	APIKeyID         int                    `json:"api_key_id"`
	Provider         string                 `json:"provider"`
	Status           string                 `json:"status"`
	LatencyMs        int                    `json:"latency_ms"`
	ErrorMessage     string                 `json:"error_message,omitempty"`
	ModelsAvailable  []string               `json:"models_available,omitempty"`
	PricingInfo      map[string]interface{} `json:"pricing_info,omitempty"`
	LastCheckedAt    time.Time              `json:"last_checked_at"`
	FailureCount     int                    `json:"failure_count"`
	ConsecutiveCount int                    `json:"consecutive_count"`
}

type HealthCheckResult struct {
	Success           bool
	Status            string
	LatencyMs         int
	ErrorMessage      string
	ErrorCategory     string
	ProviderErrorCode string
	ProviderRequestID string
	EndpointUsed      string
	StatusCode        int
	RawErrorExcerpt   string
	ModelsFound       []string
	PricingInfo       map[string]interface{}
	CheckType         string
}

type HealthChecker struct {
	httpClient *http.Client
	db         *sql.DB
}

func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		httpClient: &http.Client{
			Timeout:   10 * time.Second,
			Transport: http.DefaultTransport.(*http.Transport).Clone(),
		},
		db: config.GetDB(),
	}
}

func (s *HealthChecker) routeAwareHTTPClient(routeMode string) *http.Client {
	if routeMode == GatewayModeProxy {
		httpsProxy := os.Getenv("HTTPS_PROXY")
		if httpsProxy == "" {
			httpsProxy = os.Getenv("https_proxy")
		}
		if httpsProxy != "" {
			proxyURL, err := url.Parse(httpsProxy)
			if err == nil {
				transport := http.DefaultTransport.(*http.Transport).Clone()
				transport.Proxy = http.ProxyURL(proxyURL)
				return &http.Client{
					Timeout:   10 * time.Second,
					Transport: transport,
				}
			}
		}
	}
	return s.httpClient
}

func (s *HealthChecker) GetHealthCheckInterval(level string) int {
	levelEnum := HealthCheckLevel(level)
	if interval, ok := healthCheckIntervalMap[levelEnum]; ok {
		return interval
	}
	return 300
}

func (s *HealthChecker) LightweightPing(ctx context.Context, apiKey *models.MerchantAPIKey) (*HealthCheckResult, error) {
	baseURL, authToken, resolvedMode, resolveErr := s.ResolveProviderConnectivityBase(ctx, apiKey)
	if resolveErr != nil {
		return &HealthCheckResult{
			Success:      false,
			Status:       HealthStatusUnhealthy,
			ErrorMessage: resolveErr.Error(),
			CheckType:    "lightweight",
		}, nil
	}

	apiFmt := s.providerAPIFormat(apiKey.Provider)
	client := newProxyAwareHTTPClient(10*time.Second, resolvedMode)
	probe, err := ProbeProviderConnectivity(ctx, client, baseURL, authToken, apiKey.Provider, apiFmt)
	if err != nil {
		return &HealthCheckResult{
			Success:      false,
			Status:       HealthStatusUnhealthy,
			LatencyMs:    probeLatency(probe),
			ErrorMessage: fmt.Sprintf("connection failed: %v", err),
			CheckType:    "lightweight",
		}, nil
	}
	if probe != nil && probe.Success {
		return &HealthCheckResult{
			Success:   true,
			Status:    HealthStatusHealthy,
			LatencyMs: probe.LatencyMs,
			CheckType: "lightweight",
		}, nil
	}

	errMsg := "connectivity probe failed"
	if probe != nil && strings.TrimSpace(probe.ErrorMsg) != "" {
		errMsg = probe.ErrorMsg
	}
	status := HealthStatusDegraded
	if probe != nil && probe.StatusCode == http.StatusUnauthorized {
		status = HealthStatusUnhealthy
	}
	return &HealthCheckResult{
		Success:      false,
		Status:       status,
		LatencyMs:    probeLatency(probe),
		ErrorMessage: errMsg,
		CheckType:    "lightweight",
	}, nil
}

// ResolveProviderConnectivityBase returns the upstream base URL and auth token for connectivity probes.
func (s *HealthChecker) ResolveProviderConnectivityBase(ctx context.Context, apiKey *models.MerchantAPIKey) (baseURL string, authToken string, resolvedRouteMode string, err error) {
	trafficEp, resolveErr := s.resolveEndpoint(ctx, apiKey)
	if resolveErr != nil {
		return "", "", "", resolveErr
	}
	resolvedRouteMode = s.resolvedRouteMode(ctx, apiKey)
	modelsBase := trafficEp
	if resolvedRouteMode == GatewayModeLitellm {
		if upstream, upErr := s.resolveDirectEndpoint(ctx, apiKey); upErr == nil && strings.TrimSpace(upstream) != "" {
			modelsBase = upstream
		}
	}
	authToken = s.getDecryptedAPIKey(apiKey)
	if strings.TrimSpace(authToken) == "" && strings.TrimSpace(apiKey.APIKeyEncrypted) != "" {
		return "", "", resolvedRouteMode, fmt.Errorf("failed to decrypt API key")
	}
	return normalizeOpenAICompatBase(modelsBase), authToken, resolvedRouteMode, nil
}

// ResolveOpenAICompatModelsListProbe returns the GET URL and bearer token for OpenAI-compatible model listing
// for this merchant key. Single source of truth for FullVerification, Admin probe-models, and
// APIKeyValidator light/deep connection test — same path across direct / auto / proxy / litellm.
func (s *HealthChecker) ResolveOpenAICompatModelsListProbe(ctx context.Context, apiKey *models.MerchantAPIKey) (modelsListURL string, authToken string, resolvedRouteMode string, err error) {
	connectivityBase, token, routeMode, resolveErr := s.ResolveProviderConnectivityBase(ctx, apiKey)
	if resolveErr != nil {
		return "", "", "", resolveErr
	}
	return OpenAICompatModelsProbeURL(connectivityBase), token, routeMode, nil
}

func (s *HealthChecker) providerAPIFormat(provider string) string {
	if s.db == nil {
		s.db = config.GetDB()
	}
	if s.db == nil {
		return modelProviderOpenAI
	}
	var apiFormat string
	err := s.db.QueryRow(
		`SELECT COALESCE(NULLIF(TRIM(api_format), ''), $1) FROM model_providers WHERE lower(trim(code)) = lower(trim($2))`,
		modelProviderOpenAI, provider,
	).Scan(&apiFormat)
	if err != nil {
		return modelProviderOpenAI
	}
	return strings.ToLower(strings.TrimSpace(apiFormat))
}

func (s *HealthChecker) FullVerification(ctx context.Context, apiKey *models.MerchantAPIKey) (*HealthCheckResult, error) {
	baseURL, authToken, resolvedMode, resolveErr := s.ResolveProviderConnectivityBase(ctx, apiKey)
	if resolveErr != nil {
		return &HealthCheckResult{
			Success:      false,
			Status:       HealthStatusUnhealthy,
			ErrorMessage: resolveErr.Error(),
			CheckType:    "full",
		}, nil
	}

	apiFmt := s.providerAPIFormat(apiKey.Provider)
	endpointUsed := OpenAICompatModelsProbeURL(baseURL)
	if ProviderUsesAnthropicHTTP(apiKey.Provider, apiFmt) {
		endpointUsed = AnthropicMessagesProbeURL(baseURL)
	}

	client := newProxyAwareHTTPClient(10*time.Second, resolvedMode)
	probe, err := ProbeProviderConnectivity(ctx, client, baseURL, authToken, apiKey.Provider, apiFmt)
	if err != nil {
		return &HealthCheckResult{
			Success:      false,
			Status:       HealthStatusUnhealthy,
			LatencyMs:    probeLatency(probe),
			ErrorMessage: fmt.Sprintf("connection failed: %v", err),
			CheckType:    "full",
		}, nil
	}
	if probe != nil && probe.Success {
		return &HealthCheckResult{
			Success:      true,
			Status:       HealthStatusHealthy,
			LatencyMs:    probe.LatencyMs,
			ModelsFound:  probe.Models,
			PricingInfo:  s.extractPricingInfo(apiKey.Provider),
			EndpointUsed: endpointUsed,
			CheckType:    "full",
		}, nil
	}

	errMsg := "verification failed"
	if probe != nil && strings.TrimSpace(probe.ErrorMsg) != "" {
		errMsg = probe.ErrorMsg
	}
	return &HealthCheckResult{
		Success:           false,
		Status:            HealthStatusUnhealthy,
		LatencyMs:         probeLatency(probe),
		ErrorMessage:      errMsg,
		ErrorCategory:     safeProbeValue(probe, func(p *ProbeModelsResult) string { return p.ErrorCategory }),
		ProviderErrorCode: safeProbeValue(probe, func(p *ProbeModelsResult) string { return p.ErrorCode }),
		ProviderRequestID: safeProbeValue(probe, func(p *ProbeModelsResult) string { return p.ProviderRequestID }),
		StatusCode:        safeProbeInt(probe, func(p *ProbeModelsResult) int { return p.StatusCode }),
		RawErrorExcerpt:   safeProbeValue(probe, func(p *ProbeModelsResult) string { return p.RawErrorExcerpt }),
		EndpointUsed:      endpointUsed,
		CheckType:         "full",
	}, nil
}

func (s *HealthChecker) RecordRequestResult(ctx context.Context, apiKeyID int, isSuccess bool, latencyMs int) error {
	db := config.GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	if isSuccess {
		// 被动健康：真实请求成功时，将曾失败或尚未探测的状态收敛为 healthy，供 SmartRouter 等使用。
		_, err := db.ExecContext(ctx, `
			UPDATE merchant_api_keys 
			SET consecutive_failures = 0, 
			    health_status = CASE
			      WHEN health_status IN ('unhealthy', 'degraded', 'unknown') OR health_status IS NULL THEN 'healthy'
			      ELSE health_status
			    END,
			    last_health_check_at = CURRENT_TIMESTAMP
			WHERE id = $1`,
			apiKeyID,
		)
		return err
	}

	_, err := db.ExecContext(ctx, `
		UPDATE merchant_api_keys 
		SET consecutive_failures = consecutive_failures + 1,
		    last_health_check_at = CURRENT_TIMESTAMP
		WHERE id = $1`,
		apiKeyID,
	)
	if err != nil {
		return err
	}

	var consecutiveFailures int
	err = db.QueryRowContext(ctx, `SELECT consecutive_failures FROM merchant_api_keys WHERE id = $1`, apiKeyID).Scan(&consecutiveFailures)
	if err != nil {
		return err
	}

	if consecutiveFailures >= 5 {
		_, err = db.ExecContext(ctx, `
			UPDATE merchant_api_keys SET health_status = 'unhealthy' WHERE id = $1`,
			apiKeyID,
		)
		return err
	}

	if consecutiveFailures >= 2 {
		_, err = db.ExecContext(ctx, `
			UPDATE merchant_api_keys SET health_status = 'degraded' WHERE id = $1`,
			apiKeyID,
		)
	}

	return err
}

func (s *HealthChecker) CalculateFailureRate(ctx context.Context, apiKeyID int, windowMinutes int) (float64, error) {
	db := config.GetDB()
	if db == nil {
		return 0, fmt.Errorf("database not initialized")
	}

	var totalRequests, failedRequests int
	err := db.QueryRowContext(ctx, `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status != 'healthy') as failed
		FROM api_key_health_history 
		WHERE api_key_id = $1 
		  AND created_at > NOW() - INTERVAL '1 minute' * $2`,
		apiKeyID, windowMinutes,
	).Scan(&totalRequests, &failedRequests)

	if err != nil {
		return 0, err
	}

	if totalRequests == 0 {
		return 0, nil
	}

	return float64(failedRequests) / float64(totalRequests) * 100, nil
}

func (s *HealthChecker) MarkAsDegraded(ctx context.Context, apiKeyID int) error {
	db := config.GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	_, err := db.ExecContext(ctx, `
		UPDATE merchant_api_keys 
		SET health_status = 'degraded', 
		    last_health_check_at = CURRENT_TIMESTAMP
		WHERE id = $1`,
		apiKeyID,
	)
	return err
}

func (s *HealthChecker) GetProviderHealth(ctx context.Context, apiKeyID int) (*ProviderHealth, error) {
	db := config.GetDB()
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var key models.MerchantAPIKey
	err := db.QueryRowContext(ctx, `
		SELECT id, merchant_id, provider, health_status, health_check_level, 
		       last_health_check_at, consecutive_failures, endpoint_url
		FROM merchant_api_keys WHERE id = $1`,
		apiKeyID,
	).Scan(&key.ID, &key.MerchantID, &key.Provider, &key.HealthStatus, &key.HealthCheckLevel,
		&key.LastHealthCheckAt, &key.ConsecutiveFailures, &key.EndpointURL)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("api key not found: %d", apiKeyID)
	}
	if err != nil {
		return nil, err
	}

	var lastCheckAt time.Time
	if key.LastHealthCheckAt != nil {
		lastCheckAt = *key.LastHealthCheckAt
	}

	return &ProviderHealth{
		APIKeyID:         key.ID,
		Provider:         key.Provider,
		Status:           key.HealthStatus,
		LastCheckedAt:    lastCheckAt,
		ConsecutiveCount: key.ConsecutiveFailures,
	}, nil
}

func (s *HealthChecker) GetAllProviderHealth(ctx context.Context) ([]ProviderHealth, error) {
	db := config.GetDB()
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	rows, err := db.QueryContext(ctx, `
		SELECT id, merchant_id, provider, health_status, health_check_level, 
		       last_health_check_at, consecutive_failures, endpoint_url
		FROM merchant_api_keys 
		WHERE status = 'active'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ProviderHealth
	for rows.Next() {
		var key models.MerchantAPIKey
		err := rows.Scan(&key.ID, &key.MerchantID, &key.Provider, &key.HealthStatus, &key.HealthCheckLevel,
			&key.LastHealthCheckAt, &key.ConsecutiveFailures, &key.EndpointURL)
		if err != nil {
			continue
		}

		var lastCheckAt time.Time
		if key.LastHealthCheckAt != nil {
			lastCheckAt = *key.LastHealthCheckAt
		}

		results = append(results, ProviderHealth{
			APIKeyID:         key.ID,
			Provider:         key.Provider,
			Status:           key.HealthStatus,
			LastCheckedAt:    lastCheckAt,
			ConsecutiveCount: key.ConsecutiveFailures,
		})
	}

	return results, nil
}

func (s *HealthChecker) SaveHealthCheckResult(ctx context.Context, apiKeyID int, result *HealthCheckResult) error {
	db := config.GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	modelsJSON, _ := json.Marshal(result.ModelsFound)
	pricingJSON, _ := json.Marshal(result.PricingInfo)

	_, err := db.ExecContext(ctx, `
		INSERT INTO api_key_health_history 
		(api_key_id, check_type, status, latency_ms, error_message, models_available, pricing_info,
		 status_code, provider_error_code, provider_request_id, endpoint_used, error_category, raw_error_excerpt)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NULLIF(TRIM($9), ''), NULLIF(TRIM($10), ''), NULLIF(TRIM($11), ''), NULLIF(TRIM($12), ''), NULLIF(TRIM($13), ''))`,
		apiKeyID, result.CheckType, result.Status, result.LatencyMs, result.ErrorMessage,
		modelsJSON, pricingJSON, result.StatusCode, result.ProviderErrorCode, result.ProviderRequestID, result.EndpointUsed, result.ErrorCategory, result.RawErrorExcerpt,
	)
	if err != nil {
		return err
	}

	statusToUpdate := result.Status
	if result.Status == HealthStatusHealthy && result.CheckType == "passive" {
		statusToUpdate = HealthStatusHealthy
	}

	var merchantID int
	err = db.QueryRowContext(ctx, `
		UPDATE merchant_api_keys 
		SET health_status = $1,
		    last_health_check_at = CURRENT_TIMESTAMP,
		    consecutive_failures = CASE WHEN $2 = $4 THEN 0 ELSE consecutive_failures END
		WHERE id = $3
		RETURNING merchant_id`,
		statusToUpdate, result.Status, apiKeyID, HealthStatusHealthy,
	).Scan(&merchantID)
	if err != nil {
		return err
	}

	cache.Delete(context.Background(), cache.MerchantAPIKeysKey(merchantID))

	return nil
}

func (s *HealthChecker) resolveEndpoint(ctx context.Context, apiKey *models.MerchantAPIKey) (string, error) {
	return s.resolveEndpointWithRouteMode(ctx, apiKey)
}

func (s *HealthChecker) resolvedRouteMode(ctx context.Context, apiKey *models.MerchantAPIKey) string {
	routeMode := apiKey.RouteMode
	if routeMode == RouteModeAuto || routeMode == string(GoalAuto) || routeMode == "" {
		providerRegion := s.getProviderRegion(apiKey.Provider)
		routeMode = resolveAutoRouteMode(providerRegion)
	}
	return routeMode
}

func (s *HealthChecker) resolveEndpointWithRouteMode(ctx context.Context, apiKey *models.MerchantAPIKey) (string, error) {
	routeMode := apiKey.RouteMode
	if routeMode == RouteModeAuto || routeMode == string(GoalAuto) || routeMode == "" {
		providerRegion := s.getProviderRegion(apiKey.Provider)
		resolved := resolveAutoRouteMode(providerRegion)
		logger.LogInfo(ctx, "health_checker", "Resolved auto route mode", map[string]interface{}{
			"api_key_id":      apiKey.ID,
			"original_mode":   routeMode,
			"resolved_mode":   resolved,
			"provider_region": providerRegion,
		})
		routeMode = resolved
	}

	switch routeMode {
	case GatewayModeDirect:
		return s.resolveDirectEndpoint(ctx, apiKey)
	case GatewayModeLitellm:
		return s.resolveLitellmEndpoint(ctx, apiKey)
	case GatewayModeProxy:
		return s.resolveProxyEndpoint(ctx, apiKey)
	default:
		return s.resolveDirectEndpoint(ctx, apiKey)
	}
}

func (s *HealthChecker) resolveDirectEndpoint(ctx context.Context, apiKey *models.MerchantAPIKey) (string, error) {
	if apiKey.RouteConfig != nil {
		if endpoint, ok := apiKey.RouteConfig["endpoint_url"].(string); ok && endpoint != "" {
			return endpoint, nil
		}

		if endpoints, ok := apiKey.RouteConfig["endpoints"].(map[string]interface{}); ok {
			if directEndpoints, ok := endpoints[GatewayModeDirect].(map[string]interface{}); ok {
				region := apiKey.Region
				if region == "" {
					region = regionOverseas
				}
				if url, ok := directEndpoints[region].(string); ok && url != "" {
					return url, nil
				}
			}
		}
	}

	if ep := strings.TrimSpace(apiKey.EndpointURL); ep != "" {
		return ep, nil
	}
	if ep, ok := s.getProviderBaseURL(ctx, apiKey.Provider); ok {
		return ep, nil
	}
	if ep, ok := s.getProviderBaseURL(ctx, fallbackProviderCode); ok {
		return ep, nil
	}
	return "", fmt.Errorf("provider endpoint not configured for code=%s and fallback=%s", strings.TrimSpace(apiKey.Provider), fallbackProviderCode)
}

func (s *HealthChecker) resolveLitellmEndpoint(ctx context.Context, apiKey *models.MerchantAPIKey) (string, error) {
	if apiKey.RouteConfig != nil {
		if endpoints, ok := apiKey.RouteConfig["endpoints"].(map[string]interface{}); ok {
			if litellmEndpoints, ok := endpoints[GatewayModeLitellm].(map[string]interface{}); ok {
				region := apiKey.Region
				if region == "" {
					region = regionDomestic
				}
				if url, ok := litellmEndpoints[region].(string); ok && url != "" {
					return NormalizeLegacyLitellmGatewayBaseURL(url), nil
				}
			}
		}

		if baseURL, ok := apiKey.RouteConfig["base_url"].(string); ok && baseURL != "" {
			return NormalizeLegacyLitellmGatewayBaseURL(baseURL), nil
		}
	}

	litellmURL := os.Getenv("LLM_GATEWAY_LITELLM_URL")
	if litellmURL != "" {
		return litellmURL + "/v1", nil
	}

	return "", fmt.Errorf("LiteLLM endpoint not configured")
}

func (s *HealthChecker) resolveProxyEndpoint(ctx context.Context, apiKey *models.MerchantAPIKey) (string, error) {
	if baseURL, ok := s.getProviderBaseURL(ctx, apiKey.Provider); ok && baseURL != "" {
		logger.LogInfo(ctx, "health_checker", "Proxy mode using api_base_url via HTTPS_PROXY", map[string]interface{}{
			"provider":     apiKey.Provider,
			"api_base_url": baseURL,
		})
		return baseURL, nil
	}

	if apiKey.RouteConfig != nil {
		if endpoints, ok := apiKey.RouteConfig["endpoints"].(map[string]interface{}); ok {
			if proxyEndpoints, ok := endpoints[GatewayModeProxy].(map[string]interface{}); ok {
				if gaapURL, ok := proxyEndpoints["gaap"].(string); ok && gaapURL != "" && !strings.Contains(gaapURL, "example.com") {
					return gaapURL, nil
				}
				for _, v := range proxyEndpoints {
					if url, ok := v.(string); ok && url != "" && !strings.Contains(url, "example.com") {
						return url, nil
					}
				}
			}
		}

		if proxyURL, ok := apiKey.RouteConfig["proxy_url"].(string); ok && proxyURL != "" {
			return proxyURL, nil
		}
	}

	return "", fmt.Errorf("Proxy endpoint not configured")
}

func (s *HealthChecker) getProviderBaseURL(ctx context.Context, provider string) (string, bool) {
	db := config.GetDB()
	if db == nil {
		return "", false
	}
	var baseURL string
	err := db.QueryRowContext(ctx, `
		SELECT COALESCE(NULLIF(TRIM(api_base_url), ''), '')
		FROM model_providers
		WHERE code = $1 AND status = 'active'
		ORDER BY updated_at DESC, id DESC
		LIMIT 1
	`, strings.TrimSpace(provider)).Scan(&baseURL)
	if err != nil {
		return "", false
	}
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return "", false
	}
	return baseURL, true
}

func (s *HealthChecker) getProviderRegion(provider string) string {
	db := config.GetDB()
	if db == nil {
		return regionDomestic
	}
	var providerRegion string
	err := db.QueryRow(
		"SELECT COALESCE(provider_region, 'domestic') FROM model_providers WHERE code = $1",
		provider,
	).Scan(&providerRegion)
	if err != nil {
		return regionDomestic
	}
	return providerRegion
}

func (s *HealthChecker) getDecryptedAPIKey(apiKey *models.MerchantAPIKey) string {
	if apiKey.APIKeyEncrypted == "" {
		return ""
	}

	decrypted, err := utils.Decrypt(apiKey.APIKeyEncrypted)
	if err != nil {
		return ""
	}

	return decrypted
}

func (s *HealthChecker) extractPricingInfo(provider string) map[string]interface{} {
	engine := billing.GetBillingEngine()

	pricingData := map[string]interface{}{
		"provider": provider,
		"updated":  time.Now().Format(time.RFC3339),
		"models":   []map[string]interface{}{},
	}

	models := s.getProviderModels(provider)
	modelPricings := make([]map[string]interface{}, 0)

	for _, model := range models {
		if pricing, ok := engine.GetPricing(provider, model); ok {
			modelPricings = append(modelPricings, map[string]interface{}{
				"model":        model,
				"input_price":  pricing.InputPrice,
				"output_price": pricing.OutputPrice,
				"currency":     pricing.Currency,
			})
		}
	}

	pricingData["models"] = modelPricings
	return pricingData
}

func (s *HealthChecker) getProviderModels(provider string) []string {
	modelsMap := map[string][]string{
		"openai": {
			"gpt-4-turbo-preview", "gpt-4", "gpt-4o", "gpt-4o-mini", "gpt-3.5-turbo",
		},
		"anthropic": {
			"claude-3-opus-20240229", "claude-3-sonnet-20240229",
			"claude-3-haiku-20240307", "claude-3-5-sonnet-20241022",
		},
		"google": {
			"gemini-pro", "gemini-1.5-pro", "gemini-1.5-flash",
		},
	}

	if models, ok := modelsMap[provider]; ok {
		return models
	}
	return []string{}
}

func (s *HealthChecker) ShouldPerformCheck(apiKey *models.MerchantAPIKey) bool {
	if apiKey.LastHealthCheckAt == nil {
		return true
	}

	interval := s.GetHealthCheckInterval(apiKey.HealthCheckLevel)
	elapsed := time.Since(*apiKey.LastHealthCheckAt)

	return elapsed.Seconds() >= float64(interval)
}

func (s *HealthChecker) TriggerActiveCheck(ctx context.Context, apiKeyID int) error {
	db := config.GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	var key models.MerchantAPIKey
	var routeConfigBytes []byte
	err := db.QueryRowContext(ctx, `
		SELECT id, merchant_id, provider, api_key_encrypted, COALESCE(endpoint_url, ''), health_check_level, 
		       COALESCE(route_mode, ''), route_config, COALESCE(region, '')
		FROM merchant_api_keys WHERE id = $1`,
		apiKeyID,
	).Scan(&key.ID, &key.MerchantID, &key.Provider, &key.APIKeyEncrypted, &key.EndpointURL, &key.HealthCheckLevel,
		&key.RouteMode, &routeConfigBytes, &key.Region)

	if err != nil {
		return err
	}

	if len(routeConfigBytes) > 0 {
		_ = json.Unmarshal(routeConfigBytes, &key.RouteConfig)
	}

	result, err := s.runByLevel(ctx, &key)
	if err != nil {
		return err
	}

	return s.SaveHealthCheckResult(ctx, apiKeyID, result)
}

func PerformHealthCheckAsync(apiKeyID int) {
	ctx := context.Background()
	checker := NewHealthChecker()

	var key models.MerchantAPIKey
	var routeConfigBytes []byte
	db := config.GetDB()
	if db == nil {
		return
	}

	err := db.QueryRowContext(ctx, `
		SELECT id, merchant_id, provider, api_key_encrypted, COALESCE(endpoint_url, ''), health_check_level, health_status,
		       COALESCE(route_mode, ''), route_config, COALESCE(region, '')
		FROM merchant_api_keys WHERE id = $1`,
		apiKeyID,
	).Scan(&key.ID, &key.MerchantID, &key.Provider, &key.APIKeyEncrypted, &key.EndpointURL, &key.HealthCheckLevel, &key.HealthStatus,
		&key.RouteMode, &routeConfigBytes, &key.Region)

	if err != nil {
		return
	}

	if len(routeConfigBytes) > 0 {
		_ = json.Unmarshal(routeConfigBytes, &key.RouteConfig)
	}

	if !checker.ShouldPerformCheck(&key) {
		return
	}

	result, _ := checker.runByLevel(ctx, &key)

	if result != nil {
		checker.SaveHealthCheckResult(ctx, apiKeyID, result)
	}
}

func (s *HealthChecker) runByLevel(ctx context.Context, apiKey *models.MerchantAPIKey) (*HealthCheckResult, error) {
	level := HealthCheckLevel(strings.ToLower(strings.TrimSpace(apiKey.HealthCheckLevel)))
	if level == HealthCheckLevelHigh {
		return s.LightweightPing(ctx, apiKey)
	}
	return s.FullVerification(ctx, apiKey)
}

func probeLatency(probe *ProbeModelsResult) int {
	if probe == nil {
		return 0
	}
	return probe.LatencyMs
}

func safeProbeValue(probe *ProbeModelsResult, getter func(*ProbeModelsResult) string) string {
	if probe == nil {
		return ""
	}
	return getter(probe)
}

func safeProbeInt(probe *ProbeModelsResult, getter func(*ProbeModelsResult) int) int {
	if probe == nil {
		return 0
	}
	return getter(probe)
}

func IsHealthy(status string) bool {
	return status == HealthStatusHealthy
}

func IsDegraded(status string) bool {
	return status == HealthStatusDegraded
}

func IsUnhealthy(status string) bool {
	return status == HealthStatusUnhealthy
}

type TestChatRequest struct {
	Model    string `json:"model"`
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
	MaxTokens int `json:"max_tokens"`
}

func (s *HealthChecker) TestChatCompletion(ctx context.Context, apiKey *models.MerchantAPIKey, model string) (*HealthCheckResult, error) {
	endpoint, resolveErr := s.resolveEndpoint(ctx, apiKey)
	if resolveErr != nil {
		return &HealthCheckResult{
			Success:      false,
			Status:       HealthStatusUnhealthy,
			ErrorMessage: resolveErr.Error(),
			CheckType:    "chat_test",
		}, nil
	}

	chatEndpoint := OpenAICompatChatCompletionsURL(endpoint)

	testReq := TestChatRequest{
		Model: model,
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{Role: "user", Content: "Hello"},
		},
		MaxTokens: 5,
	}

	body, _ := json.Marshal(testReq)
	req, err := http.NewRequestWithContext(ctx, "POST", chatEndpoint, bytes.NewReader(body))
	if err != nil {
		return &HealthCheckResult{
			Success:      false,
			Status:       HealthStatusUnhealthy,
			ErrorMessage: fmt.Sprintf("failed to create request: %v", err),
			CheckType:    "chat_test",
		}, nil
	}

	authToken := s.getDecryptedAPIKey(apiKey)
	resolvedMode := s.resolvedRouteMode(ctx, apiKey)
	if resolvedMode == GatewayModeLitellm {
		if masterKey := os.Getenv("LITELLM_MASTER_KEY"); masterKey != "" {
			authToken = strings.TrimSpace(masterKey)
		}
	}
	SetProviderAuthHeaders(req, apiKey.Provider, authToken)

	client := s.routeAwareHTTPClient(resolvedMode)
	start := time.Now()
	resp, err := client.Do(req)
	latencyMs := int(time.Since(start).Milliseconds())

	if err != nil {
		errInfo := MapProviderError(0, "", fmt.Sprintf("connection failed: %v", err), nil, err, "")
		return &HealthCheckResult{
			Success:       false,
			Status:        HealthStatusUnhealthy,
			LatencyMs:     latencyMs,
			ErrorMessage:  errInfo.ProviderMessage,
			ErrorCategory: errInfo.Category,
			CheckType:     "chat_test",
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return &HealthCheckResult{
			Success:   true,
			Status:    HealthStatusHealthy,
			LatencyMs: latencyMs,
			CheckType: "chat_test",
		}, nil
	}

	respBody, _ := io.ReadAll(resp.Body)
	code, msg := ExtractProviderError(respBody)
	if msg == "" {
		msg = strings.TrimSpace(string(respBody))
	}
	errInfo := MapProviderError(resp.StatusCode, code, msg, resp.Header, nil, string(respBody))
	return &HealthCheckResult{
		Success:           false,
		Status:            HealthStatusDegraded,
		LatencyMs:         latencyMs,
		ErrorMessage:      firstNonEmpty(errInfo.ProviderMessage, fmt.Sprintf("status: %d", resp.StatusCode)),
		ErrorCategory:     errInfo.Category,
		ProviderErrorCode: firstNonEmpty(code, errInfo.ProviderCode),
		ProviderRequestID: errInfo.ProviderRequestID,
		StatusCode:        resp.StatusCode,
		RawErrorExcerpt:   errInfo.RawErrorExcerpt,
		EndpointUsed:      chatEndpoint,
		CheckType:         "chat_test",
	}, nil
}

// ResolveMerchantAPIKeyUpstreamBase returns the HTTP base URL used to reach the upstream or gateway for this BYOK row,
// and the effective route mode (direct / litellm / proxy), matching HealthChecker probes.
func ResolveMerchantAPIKeyUpstreamBase(ctx context.Context, apiKey *models.MerchantAPIKey) (baseURL string, routeMode string, err error) {
	if apiKey == nil {
		return "", "", fmt.Errorf("nil apiKey")
	}
	s := NewHealthChecker()
	baseURL, err = s.resolveEndpoint(ctx, apiKey)
	if err != nil {
		return "", "", err
	}
	routeMode = s.resolvedRouteMode(ctx, apiKey)
	return baseURL, routeMode, nil
}
