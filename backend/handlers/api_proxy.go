package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pintuotuo/backend/billing"
	"github.com/pintuotuo/backend/cache"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/logger"
	"github.com/pintuotuo/backend/metrics"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/models"
	"github.com/pintuotuo/backend/services"
	"github.com/pintuotuo/backend/utils"
)

const (
	providerAnthropic = "anthropic"
	apiFormatOpenAI   = "openai"
)

// 路由策略来源（trace / 落库 effective_policy_source，与 JSON 对外字段一致）
const (
	policySourceEnv     = "env"
	policySourceDB      = "db"
	policySourceDefault = "default"
)

type providerRuntimeConfig struct {
	Code       string
	Name       string
	APIBaseURL string
	APIFormat  string
}

func legacyProviderList() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":         "openai",
			"display_name": "OpenAI",
			"api_format":   "openai",
		},
		{
			"name":         "anthropic",
			"display_name": "Anthropic",
			"api_format":   "anthropic",
		},
		{
			"name":         "google",
			"display_name": "Google AI",
			"api_format":   "openai",
		},
	}
}

type APIProxyRequest struct {
	Provider      string          `json:"provider" binding:"required"`
	Model         string          `json:"model" binding:"required"`
	Messages      []ChatMessage   `json:"messages" binding:"required"`
	Stream        bool            `json:"stream"`
	Options       json.RawMessage `json:"options,omitempty"`
	APIKeyID      *int            `json:"api_key_id,omitempty"`
	MerchantSKUID *int            `json:"merchant_sku_id,omitempty"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func estimateInputTokens(messages []ChatMessage) int {
	totalChars := 0
	for _, msg := range messages {
		totalChars += len(msg.Role) + len(msg.Content)
	}
	return totalChars / 4
}

type APIProxyResponse struct {
	ID      string      `json:"id"`
	Object  string      `json:"object"`
	Created int64       `json:"created"`
	Model   string      `json:"model"`
	Choices []APIChoice `json:"choices"`
	Usage   APIUsage    `json:"usage"`
	Error   *APIError   `json:"error,omitempty"`
}

type APIChoice struct {
	Index        int          `json:"index"`
	Message      *ChatMessage `json:"message"`
	Delta        *ChatMessage `json:"delta,omitempty"`
	FinishReason string       `json:"finish_reason"`
}

type APIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type APIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

func ProxyAPIRequest(c *gin.Context) {
	startTime := time.Now()
	requestID := uuid.New().String()

	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}
	userIDInt, ok := userID.(int)
	if !ok {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	var req APIProxyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	proxyAPIRequestCore(c, userIDInt, requestID, startTime, req, c.Request.URL.Path)
}

func proxyAPIRequestCore(c *gin.Context, userIDInt int, requestID string, startTime time.Time, req APIProxyRequest, requestPath string) {
	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	c.Header("X-Request-ID", requestID)
	c.Header("X-Trace-ID", requestID)
	traceSpan := services.StartLLMTrace(requestID, userIDInt)
	defer traceSpan.Finish(c.Request.Context())

	var strictPricingVID *int
	if services.EntitlementEnforcementStrict() {
		vid, _, ok, entErr := services.ResolveChosenPricingVersion(db, userIDInt, req.Provider, req.Model)
		if entErr != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		if !ok {
			middleware.RespondWithError(c, apperrors.ErrEntitlementDenied)
			return
		}
		strictPricingVID = &vid
	}

	var tokenBalance float64
	err := db.QueryRow("SELECT balance FROM tokens WHERE user_id = $1", userIDInt).Scan(&tokenBalance)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrTokenNotFound)
		return
	}

	if tokenBalance <= 0 {
		middleware.RespondWithError(c, apperrors.ErrInsufficientBalance)
		return
	}

	inputTokensEstimate := estimateInputTokens(req.Messages)
	billingEngine := billing.GetBillingEngine()
	preDeductConfig := billingEngine.GetPreDeductConfig(0, 0, req.Provider)
	estimatedUsage := billingEngine.EstimateTokenUsage(inputTokensEstimate, preDeductConfig)

	if tokenBalance < float64(estimatedUsage) {
		logger.LogWarn(c.Request.Context(), "api_proxy", "Insufficient balance for pre-deduction", map[string]interface{}{
			"user_id":         userIDInt,
			"balance":         tokenBalance,
			"estimated_usage": estimatedUsage,
			"input_estimate":  inputTokensEstimate,
			"multiplier":      preDeductConfig.Multiplier,
			"request_id":      requestID,
		})
		middleware.RespondWithError(c, apperrors.ErrInsufficientBalance)
		return
	}

	preDeductErr := billingEngine.PreDeductBalance(userIDInt, estimatedUsage, "API call pre-deduct", requestID)
	if preDeductErr != nil {
		logger.LogError(c.Request.Context(), "api_proxy", "Pre-deduction failed", preDeductErr, map[string]interface{}{
			"user_id":         userIDInt,
			"estimated_usage": estimatedUsage,
			"request_id":      requestID,
		})
		middleware.RespondWithError(c, apperrors.ErrInsufficientBalance)
		return
	}

	logger.LogInfo(c.Request.Context(), "api_proxy", "Pre-deduction successful", map[string]interface{}{
		"user_id":         userIDInt,
		"estimated_usage": estimatedUsage,
		"request_id":      requestID,
	})

	merchantID, merchantErr := resolveMerchantIDByUser(db, userIDInt)
	if merchantErr != nil {
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	selectedStrategy := "legacy_fallback"
	effectivePolicySource := ""
	var smartCandidatesJSON []byte
	decisionStart := time.Now()
	if req.APIKeyID == nil && req.MerchantSKUID == nil && shouldUseSmartRouting(userIDInt, requestID) {
		strategyCode, policySrc := routingStrategyWithSource()
		if smartReq := trySelectAPIKeyWithSmartRouter(req, strategyCode); smartReq.APIKeyID != nil {
			req.APIKeyID = smartReq.APIKeyID
			selectedStrategy = strategyCode
			smartCandidatesJSON = smartReq.CandidatesJSON
			effectivePolicySource = policySrc
		}
	}

	var apiKey models.MerchantAPIKey
	err = selectAPIKeyForRequest(db, userIDInt, merchantID, req, &apiKey)
	if err != nil {
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		if errors.Is(err, sql.ErrNoRows) {
			logger.LogWarn(c.Request.Context(), "api_proxy", "API key authorization miss", map[string]interface{}{
				"request_id":        requestID,
				"request_path":      requestPath,
				"user_id":           userIDInt,
				"merchant_id":       merchantID,
				"provider":          req.Provider,
				"model":             req.Model,
				"api_key_id":        req.APIKeyID,
				"merchant_sku_id":   req.MerchantSKUID,
				"selected_strategy": selectedStrategy,
			})
			middleware.RespondWithError(c, apperrors.NewAppError(
				"API_KEY_NOT_AUTHORIZED",
				"No authorized API key available for this provider",
				http.StatusForbidden,
				nil,
			))
			return
		}
		middleware.RespondWithError(c, apperrors.NewAppError(
			"API_KEY_NOT_FOUND",
			"No available API key for this provider",
			http.StatusServiceUnavailable,
			err,
		))
		return
	}

	decryptedKey, err := utils.Decrypt(apiKey.APIKeyEncrypted)
	if err != nil {
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		middleware.RespondWithError(c, apperrors.NewAppError(
			"DECRYPTION_FAILED",
			"Failed to decrypt API key",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	providerCfg, err := getProviderRuntimeConfig(db, req.Provider)
	if err == sql.ErrNoRows {
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		middleware.RespondWithError(c, apperrors.NewAppError(
			"UNSUPPORTED_PROVIDER",
			fmt.Sprintf("Provider %s is not supported", req.Provider),
			http.StatusBadRequest,
			nil,
		))
		return
	}
	if err != nil {
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	providerCfg = applyGatewayOverride(providerCfg)
	traceSpan.SetRoute(req.Provider, req.Model)

	baseURL := strings.TrimRight(providerCfg.APIBaseURL, "/")
	if baseURL == "" {
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		middleware.RespondWithError(c, apperrors.NewAppError(
			"UNSUPPORTED_PROVIDER",
			fmt.Sprintf("Provider %s is missing api_base_url", req.Provider),
			http.StatusBadRequest,
			nil,
		))
		return
	}

	endpoint := fmt.Sprintf("%s/chat/completions", baseURL)
	if providerCfg.APIFormat == providerAnthropic {
		endpoint = fmt.Sprintf("%s/messages", baseURL)
	}

	requestBody := map[string]interface{}{
		"model":    req.Model,
		"messages": req.Messages,
		"stream":   req.Stream,
	}

	if req.Options != nil {
		var options map[string]interface{}
		if unmarshalErr := json.Unmarshal(req.Options, &options); unmarshalErr == nil {
			for k, v := range options {
				requestBody[k] = v
			}
		}
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		middleware.RespondWithError(c, apperrors.NewAppError(
			"REQUEST_BUILD_FAILED",
			"Failed to build request body",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	httpReq, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		middleware.RespondWithError(c, apperrors.NewAppError(
			"REQUEST_CREATE_FAILED",
			"Failed to create request",
			http.StatusInternalServerError,
			err,
		))
		return
	}
	httpReq.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(jsonBody)), nil
	}

	httpReq.Header.Set("Content-Type", "application/json")
	switch providerCfg.APIFormat {
	case providerAnthropic:
		httpReq.Header.Set("x-api-key", decryptedKey)
		httpReq.Header.Set("anthropic-version", "2023-06-01")
	default:
		authToken := resolveGatewayAuthToken(providerCfg, decryptedKey)
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
	}

	client := &http.Client{Timeout: 60 * time.Second}
	strategySnapshot := buildStrategyRuntimeSnapshot(selectedStrategy)
	applyCircuitBreakerConfig(apiKey.ID, strategySnapshot)
	retryPolicy := buildRetryPolicy(strategySnapshot)
	resp, retryCount, err := executeProviderRequestWithRetry(client, httpReq, retryPolicy)
	if err != nil {
		traceSpan.SetStatusCode(http.StatusBadGateway)
		traceSpan.SetErrorCode("API_REQUEST_FAILED")
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		services.GetSmartRouter().RecordRequestResult(apiKey.ID, false)
		recordHealthCheckerProxyOutcome(c, apiKey.ID, false, startTime)
		middleware.RespondWithError(c, apperrors.NewAppError(
			"API_REQUEST_FAILED",
			"Failed to send request to provider",
			http.StatusBadGateway,
			err,
		))
		return
	}
	defer resp.Body.Close()
	proxyTransportOK := resp.StatusCode < http.StatusInternalServerError && resp.StatusCode != http.StatusTooManyRequests
	services.GetSmartRouter().RecordRequestResult(apiKey.ID, proxyTransportOK)
	recordHealthCheckerProxyOutcome(c, apiKey.ID, proxyTransportOK, startTime)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		recordHealthCheckerProxyOutcome(c, apiKey.ID, false, startTime)
		middleware.RespondWithError(c, apperrors.NewAppError(
			"RESPONSE_READ_FAILED",
			"Failed to read response",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	var apiResp APIProxyResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		traceSpan.SetStatusCode(resp.StatusCode)
		traceSpan.SetErrorCode("UNMARSHAL_PROXY_RESPONSE_FAILED")
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		c.Data(resp.StatusCode, "application/json", body)
		return
	}

	latency := int(time.Since(startTime).Milliseconds())

	var inputTokens, outputTokens int
	var tokenUsage int64
	var cost float64

	if apiResp.Usage.TotalTokens > 0 {
		inputTokens = apiResp.Usage.PromptTokens
		outputTokens = apiResp.Usage.CompletionTokens
		tokenUsage = billingEngine.CalculateTokenUsage(inputTokens, outputTokens)
		var cerr error
		cost, cerr = calculateTokenCost(db, userIDInt, req.Provider, req.Model, inputTokens, outputTokens, strictPricingVID)
		if cerr != nil {
			billingEngine.CancelPreDeduct(userIDInt, requestID)
			logger.LogError(context.Background(), "api_proxy", "Token cost resolution failed", cerr, map[string]interface{}{
				"user_id": userIDInt, "provider": req.Provider, "model": req.Model, "request_id": requestID,
			})
			middleware.RespondWithError(c, apperrors.ErrPricingSnapshotMiss)
			return
		}
	}

	if tokenUsage > 0 {
		settleErr := billingEngine.SettlePreDeduct(userIDInt, requestID, tokenUsage)
		if settleErr != nil {
			logger.LogError(context.Background(), "api_proxy", "Settle pre-deduct failed", settleErr, map[string]interface{}{
				"user_id":     userIDInt,
				"token_usage": tokenUsage,
				"request_id":  requestID,
			})
		}
	} else {
		billingEngine.CancelPreDeduct(userIDInt, requestID)
	}

	if cost > 0 {
		logMerchantSKUID, logProcurementCNY := resolveMerchantSKUProcurementForLog(db, req, apiKey.ID, merchantID, inputTokens, outputTokens)
		tx, err := db.Begin()
		if err == nil {
			_, updateErr := tx.Exec(
				"UPDATE merchant_api_keys SET quota_used = quota_used + $1, last_used_at = $2 WHERE id = $3",
				cost, time.Now(), apiKey.ID,
			)
			err = updateErr
			if err == nil {
				_, err = tx.Exec(
					"INSERT INTO api_usage_logs (user_id, key_id, request_id, provider, model, method, path, status_code, latency_ms, input_tokens, output_tokens, cost, token_usage, merchant_sku_id, procurement_cost_cny) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)",
					userIDInt, apiKey.ID, requestID, req.Provider, req.Model, "POST", requestPath, resp.StatusCode, latency, inputTokens, outputTokens, cost, tokenUsage, nullInt64Arg(logMerchantSKUID), nullFloat64Arg(logProcurementCNY),
				)
			}

			if err != nil {
				tx.Rollback()
				logger.LogError(context.Background(), "api_proxy", "Transaction rollback", err, map[string]interface{}{
					"user_id":     userIDInt,
					"provider":    req.Provider,
					"model":       req.Model,
					"cost":        cost,
					"token_usage": tokenUsage,
					"request_id":  requestID,
				})
			} else {
				tx.Commit()
				logger.LogInfo(context.Background(), "api_proxy", "API request completed", map[string]interface{}{
					"user_id":       userIDInt,
					"provider":      req.Provider,
					"model":         req.Model,
					"input_tokens":  inputTokens,
					"output_tokens": outputTokens,
					"token_usage":   tokenUsage,
					"cost":          cost,
					"latency_ms":    latency,
					"status_code":   resp.StatusCode,
					"request_id":    requestID,
				})

				metrics.RecordOrderCreation("completed", int64(cost*100), "USD")
			}
		}

		ctx := context.Background()
		cache.Delete(ctx, cache.TokenBalanceKey(userIDInt))
		cache.Delete(ctx, cache.ComputePointBalanceKey(userIDInt))
	}

	decisionPayload := buildRoutingDecisionPayload(smartCandidatesJSON, strategySnapshot, effectivePolicySource)
	_ = insertRoutingDecision(db, requestID, userIDInt, req, selectedStrategy, decisionPayload, apiKey.ID, int(time.Since(decisionStart).Milliseconds()), retryCount)

	traceSpan.SetStatusCode(resp.StatusCode)
	c.Data(resp.StatusCode, "application/json", body)
}

func applyGatewayOverride(cfg providerRuntimeConfig) providerRuntimeConfig {
	active := strings.TrimSpace(strings.ToLower(os.Getenv("LLM_GATEWAY_ACTIVE")))
	if cfg.APIFormat != apiFormatOpenAI || active == "" || active == "none" {
		return cfg
	}
	switch active {
	case "litellm":
		if base := strings.TrimSpace(os.Getenv("LLM_GATEWAY_LITELLM_URL")); base != "" {
			cfg.APIBaseURL = strings.TrimRight(base, "/") + "/v1"
		}
	}
	return cfg
}

func resolveGatewayAuthToken(cfg providerRuntimeConfig, fallbackToken string) string {
	active := strings.TrimSpace(strings.ToLower(os.Getenv("LLM_GATEWAY_ACTIVE")))
	if cfg.APIFormat != apiFormatOpenAI || active == "" || active == "none" {
		return fallbackToken
	}
	switch active {
	case "litellm":
		if token := strings.TrimSpace(os.Getenv("LITELLM_MASTER_KEY")); token != "" {
			return token
		}
	}
	return fallbackToken
}

func calculateTokenCost(db *sql.DB, userID int, provider, model string, inputTokens, outputTokens int, strictPricingVID *int) (float64, error) {
	if strictPricingVID != nil {
		cost, ok := services.CalculateCostFromPricingVersion(db, *strictPricingVID, provider, model, inputTokens, outputTokens)
		if !ok {
			return 0, fmt.Errorf("strict pricing snapshot miss for version %d", *strictPricingVID)
		}
		logger.LogDebug(context.Background(), "api_proxy", "Token cost from entitlement pricing_version", map[string]interface{}{
			"pricing_version_id": *strictPricingVID,
			"pricing_source":     "entitlement_strict",
			"provider":           provider,
			"model":              model,
			"input_tokens":       inputTokens,
			"output_tokens":      outputTokens,
			"cost":               cost,
		})
		return cost, nil
	}

	vid := services.LatestUserPricingVersionID(db, userID)
	if vid.Valid {
		if cost, ok := services.CalculateCostFromPricingVersion(db, int(vid.Int64), provider, model, inputTokens, outputTokens); ok {
			logger.LogDebug(context.Background(), "api_proxy", "Token cost from pricing_version snapshot", map[string]interface{}{
				"pricing_version_id": vid.Int64,
				"pricing_source":     "pricing_version_spu_rates",
				"provider":           provider,
				"model":              model,
				"input_tokens":       inputTokens,
				"output_tokens":      outputTokens,
				"cost":               cost,
			})
			return cost, nil
		}
		logger.LogDebug(context.Background(), "api_proxy", "pricing_version snapshot miss, fallback live SPU", map[string]interface{}{
			"pricing_version_id": vid.Int64,
			"provider":           provider,
			"model":              model,
		})
	}

	pricingService := services.GetPricingService()
	cost := pricingService.CalculateCost(provider, model, inputTokens, outputTokens)

	logger.LogDebug(context.Background(), "api_proxy", "Token cost calculated (live SPU)", map[string]interface{}{
		"provider":       provider,
		"model":          model,
		"input_tokens":   inputTokens,
		"output_tokens":  outputTokens,
		"cost":           cost,
		"pricing_source": "live_spu",
	})

	return cost, nil
}

func getProviderRuntimeConfig(db *sql.DB, providerCode string) (providerRuntimeConfig, error) {
	var cfg providerRuntimeConfig
	err := db.QueryRow(
		`SELECT code, name, COALESCE(api_base_url, ''), api_format
		 FROM model_providers
		 WHERE code = $1 AND status = 'active'
		 LIMIT 1`,
		providerCode,
	).Scan(&cfg.Code, &cfg.Name, &cfg.APIBaseURL, &cfg.APIFormat)
	return cfg, err
}

func scanMerchantAPIKeyQuotaRow(row *sql.Row, apiKey *models.MerchantAPIKey) error {
	var qLim sql.NullFloat64
	if err := row.Scan(&apiKey.ID, &apiKey.MerchantID, &apiKey.Provider, &apiKey.APIKeyEncrypted, &apiKey.APISecretEncrypted, &qLim, &apiKey.QuotaUsed, &apiKey.Status); err != nil {
		return err
	}
	apiKey.QuotaLimit = utils.NullFloat64Ptr(qLim)
	return nil
}

func selectAPIKeyForRequest(db *sql.DB, userID, merchantID int, req APIProxyRequest, apiKey *models.MerchantAPIKey) error {
	if req.APIKeyID != nil && *req.APIKeyID > 0 {
		query := `SELECT id, merchant_id, provider, api_key_encrypted, api_secret_encrypted, quota_limit, quota_used, status
			 FROM merchant_api_keys
			 WHERE id = $1 AND provider = $2 AND status = 'active'
			   AND (verified_at IS NOT NULL OR verification_result = 'verified')
			   AND (quota_limit IS NULL OR quota_used < quota_limit)`
		args := []interface{}{*req.APIKeyID, req.Provider}
		if merchantID <= 0 {
			return scanMerchantAPIKeyQuotaRow(
				db.QueryRow(query+" LIMIT 1", args...),
				apiKey,
			)
		}
		query += " AND merchant_id = $3 LIMIT 1"
		args = append(args, merchantID)
		err := scanMerchantAPIKeyQuotaRow(
			db.QueryRow(
				query,
				args...,
			),
			apiKey,
		)
		if err == nil {
			return nil
		}
		if err != sql.ErrNoRows {
			return err
		}
	}

	if req.MerchantSKUID != nil && *req.MerchantSKUID > 0 {
		if merchantID <= 0 {
			return sql.ErrNoRows
		}
		err := scanMerchantAPIKeyQuotaRow(
			db.QueryRow(
				`SELECT mak.id, mak.merchant_id, mak.provider, mak.api_key_encrypted, mak.api_secret_encrypted, mak.quota_limit, mak.quota_used, mak.status
				 FROM merchant_skus ms
				 JOIN merchant_api_keys mak ON mak.id = ms.api_key_id
				 JOIN merchants m ON m.id = ms.merchant_id
				 WHERE ms.id = $1 AND ms.status = 'active'
				   AND ms.merchant_id = $2
				   AND m.user_id = $3
				   AND mak.provider = $4 AND mak.status = 'active'
				   AND (mak.verified_at IS NOT NULL OR mak.verification_result = 'verified')`,
				*req.MerchantSKUID, merchantID, userID, req.Provider,
			),
			apiKey,
		)
		if err == nil {
			return nil
		}
		if err != sql.ErrNoRows {
			return err
		}
	}

	if merchantID > 0 {
		return scanMerchantAPIKeyQuotaRow(
			db.QueryRow(
				`SELECT id, merchant_id, provider, api_key_encrypted, api_secret_encrypted, quota_limit, quota_used, status
				 FROM merchant_api_keys
				 WHERE provider = $1 AND status = 'active'
				   AND merchant_id = $2
				   AND (verified_at IS NOT NULL OR verification_result = 'verified')
				   AND (quota_limit IS NULL OR quota_used < quota_limit)
				 ORDER BY COALESCE((quota_limit - quota_used)::double precision, 1e30::double precision) DESC
				 LIMIT 1`,
				req.Provider, merchantID,
			),
			apiKey,
		)
	}

	return scanMerchantAPIKeyQuotaRow(
		db.QueryRow(
			`SELECT id, merchant_id, provider, api_key_encrypted, api_secret_encrypted, quota_limit, quota_used, status
			 FROM merchant_api_keys
			 WHERE provider = $1 AND status = 'active'
			   AND (verified_at IS NOT NULL OR verification_result = 'verified')
			   AND (quota_limit IS NULL OR quota_used < quota_limit)
			 ORDER BY COALESCE((quota_limit - quota_used)::double precision, 1e30::double precision) DESC
			 LIMIT 1`,
			req.Provider,
		),
		apiKey,
	)
}

// resolveMerchantSKUProcurementForLog 按 merchant_skus 成本单价计算采购成本；无绑定在售 SKU 时返回空。
func resolveMerchantSKUProcurementForLog(db *sql.DB, req APIProxyRequest, apiKeyID int, merchantID int, inputTokens, outputTokens int) (sql.NullInt64, sql.NullFloat64) {
	var msID int
	var inRate, outRate float64
	var err error
	if req.MerchantSKUID != nil && *req.MerchantSKUID > 0 && merchantID > 0 {
		err = db.QueryRow(
			`SELECT ms.id, ms.cost_input_rate, ms.cost_output_rate
			 FROM merchant_skus ms
			 WHERE ms.id = $1 AND ms.api_key_id = $2 AND ms.merchant_id = $3 AND ms.status = 'active'`,
			*req.MerchantSKUID, apiKeyID, merchantID,
		).Scan(&msID, &inRate, &outRate)
	} else {
		err = db.QueryRow(
			`SELECT ms.id, ms.cost_input_rate, ms.cost_output_rate
			 FROM merchant_skus ms
			 WHERE ms.api_key_id = $1 AND ms.status = 'active'
			 LIMIT 1`,
			apiKeyID,
		).Scan(&msID, &inRate, &outRate)
	}
	if err != nil {
		if err == sql.ErrNoRows {
			return sql.NullInt64{}, sql.NullFloat64{}
		}
		logger.LogWarn(context.Background(), "api_proxy", "resolve procurement merchant_sku failed", map[string]interface{}{
			"error": err.Error(), "api_key_id": apiKeyID, "merchant_id": merchantID,
		})
		return sql.NullInt64{}, sql.NullFloat64{}
	}
	proc := services.ProcurementCostCNY(inRate, outRate, inputTokens, outputTokens)
	return sql.NullInt64{Int64: int64(msID), Valid: true}, sql.NullFloat64{Float64: proc, Valid: true}
}

func nullInt64Arg(n sql.NullInt64) interface{} {
	if n.Valid {
		return n.Int64
	}
	return nil
}

func nullFloat64Arg(n sql.NullFloat64) interface{} {
	if n.Valid {
		return n.Float64
	}
	return nil
}

func resolveMerchantIDByUser(db *sql.DB, userID int) (int, error) {
	var merchantID int
	err := db.QueryRow("SELECT id FROM merchants WHERE user_id = $1 AND status = 'active' LIMIT 1", userID).Scan(&merchantID)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return merchantID, nil
}

type smartRoutingPick struct {
	APIKeyID       *int
	CandidatesJSON []byte
}

type strategyRuntimeSnapshot struct {
	StrategyCode      string `json:"strategy_code"`
	MaxRetries        int    `json:"max_retries"`
	InitialDelayMs    int    `json:"initial_delay_ms"`
	CircuitThreshold  int    `json:"circuit_breaker_threshold"`
	CircuitTimeoutSec int    `json:"circuit_breaker_timeout_sec"`
}

type routingDecisionPayload struct {
	Candidates            json.RawMessage         `json:"candidates"`
	StrategyRuntime       strategyRuntimeSnapshot `json:"strategy_runtime"`
	EffectivePolicySource string                  `json:"effective_policy_source,omitempty"`
}

// traceTopCandidate 用于 trace 列表页展示的候选摘要（得分最高的一条）
type traceTopCandidate struct {
	APIKeyID int     `json:"api_key_id"`
	Provider string  `json:"provider"`
	Model    string  `json:"model,omitempty"`
	Score    float64 `json:"score"`
}

func trySelectAPIKeyWithSmartRouter(req APIProxyRequest, strategyCode string) smartRoutingPick {
	if strings.TrimSpace(req.Provider) == "" {
		return smartRoutingPick{}
	}
	router := services.GetSmartRouter()
	choice, err := router.SelectProvider(context.Background(), req.Model, req.Provider, services.RoutingStrategy(strategyCode))
	if err != nil || choice == nil {
		return smartRoutingPick{}
	}

	candidates, cErr := router.GetCandidates(context.Background(), req.Model, req.Provider)
	if cErr == nil {
		router.CalculateScores(candidates, services.RoutingStrategy(strategyCode))
	}

	apiKeyID := choice.APIKeyID
	var candidatesJSON []byte
	if cErr == nil {
		candidatesJSON, _ = json.Marshal(candidates)
	}

	return smartRoutingPick{
		APIKeyID:       &apiKeyID,
		CandidatesJSON: candidatesJSON,
	}
}

// routingStrategyWithSource 返回策略代码及其来源：env（环境变量）、db（库里的默认策略）、default（内置 balanced）
func routingStrategyWithSource() (code string, source string) {
	code = strings.TrimSpace(os.Getenv("SMART_ROUTING_STRATEGY"))
	if code != "" {
		return code, policySourceEnv
	}
	code = strings.TrimSpace(services.GetSmartRouter().GetDefaultStrategyCode())
	if code != "" {
		return code, policySourceDB
	}
	return string(services.RoutingStrategyBalanced), policySourceDefault
}

func shouldUseSmartRouting(userID int, requestID string) bool {
	enabled := strings.TrimSpace(strings.ToLower(os.Getenv("SMART_ROUTING_ENABLE")))
	if enabled == "false" || enabled == "0" || enabled == "off" {
		return false
	}
	percent := 100
	if raw := strings.TrimSpace(os.Getenv("SMART_ROUTING_GRAY_PERCENT")); raw != "" {
		if p, err := strconv.Atoi(raw); err == nil {
			if p < 0 {
				p = 0
			}
			if p > 100 {
				p = 100
			}
			percent = p
		}
	}
	if percent == 0 {
		return false
	}
	if percent == 100 {
		return true
	}
	seed := userID*31 + len(requestID)*17
	for _, ch := range requestID {
		seed += int(ch)
	}
	slot := seed % 100
	if slot < 0 {
		slot = -slot
	}
	return slot < percent
}

// recordHealthCheckerProxyOutcome 将上游请求成败写入 merchant_api_keys（被动健康），与 SmartRouter 熔断分离。
func recordHealthCheckerProxyOutcome(c *gin.Context, apiKeyID int, success bool, startTime time.Time) {
	if apiKeyID <= 0 {
		return
	}
	latencyMs := int(time.Since(startTime).Milliseconds())
	if err := services.NewHealthChecker().RecordRequestResult(c.Request.Context(), apiKeyID, success, latencyMs); err != nil {
		logger.LogError(c.Request.Context(), "api_proxy", "RecordRequestResult failed", err, map[string]interface{}{
			"api_key_id": apiKeyID,
		})
	}
}

func insertRoutingDecision(db *sql.DB, requestID string, userID int, req APIProxyRequest, strategy string, candidatesJSON []byte, selectedAPIKeyID int, latencyMs int, retryCount int) error {
	if db == nil {
		return nil
	}
	wasRetry := retryCount > 0
	_, err := db.Exec(
		`INSERT INTO routing_decisions
		(request_id, user_id, model_requested, strategy_used, candidates, selected_provider, selected_api_key_id, decision_latency_ms, was_retry, retry_count)
		VALUES ($1, $2, $3, $4, $5, NULL, $6, $7, $8, $9)`,
		requestID, userID, req.Model, strategy, candidatesJSON, selectedAPIKeyID, latencyMs, wasRetry, retryCount,
	)
	return err
}

func executeProviderRequestWithRetry(client *http.Client, baseReq *http.Request, policy *services.RetryPolicy) (*http.Response, int, error) {
	if policy == nil {
		policy = services.DefaultRetryPolicy
	}
	var (
		resp       *http.Response
		err        error
		retryCount int
	)
	ctx := baseReq.Context()
	for i := 0; i <= policy.MaxRetries; i++ {
		req := baseReq.Clone(ctx)
		if baseReq.GetBody != nil {
			body, bodyErr := baseReq.GetBody()
			if bodyErr == nil {
				req.Body = body
			}
		}

		resp, err = client.Do(req) // #nosec G704 -- upstream URL from admin-configured model_providers.api_base_url, not user-supplied host
		if err == nil && resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode < http.StatusInternalServerError {
			return resp, retryCount, nil
		}

		if resp != nil && (resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= http.StatusInternalServerError) {
			resp.Body.Close()
			err = fmt.Errorf("upstream status %d", resp.StatusCode)
		}

		shouldRetry, delay := policy.ShouldRetry(err, i)
		if !shouldRetry {
			return nil, retryCount, err
		}
		retryCount++
		time.Sleep(delay)
	}
	return nil, retryCount, err
}

func buildRetryPolicy(snapshot strategyRuntimeSnapshot) *services.RetryPolicy {
	policy := *services.DefaultRetryPolicy
	if snapshot.MaxRetries > 0 {
		policy.MaxRetries = snapshot.MaxRetries
	}
	if snapshot.InitialDelayMs > 0 {
		policy.InitialDelay = time.Duration(snapshot.InitialDelayMs) * time.Millisecond
	}
	return &policy
}

func applyCircuitBreakerConfig(apiKeyID int, snapshot strategyRuntimeSnapshot) {
	if apiKeyID <= 0 {
		return
	}
	threshold := snapshot.CircuitThreshold
	if threshold <= 0 {
		threshold = 5
	}
	timeoutSeconds := snapshot.CircuitTimeoutSec
	if timeoutSeconds <= 0 {
		timeoutSeconds = 60
	}
	services.GetSmartRouter().ConfigureCircuitBreaker(apiKeyID, threshold, time.Duration(timeoutSeconds)*time.Second)
}

func buildStrategyRuntimeSnapshot(strategyCode string) strategyRuntimeSnapshot {
	snapshot := strategyRuntimeSnapshot{
		StrategyCode:      strategyCode,
		MaxRetries:        services.DefaultRetryPolicy.MaxRetries,
		InitialDelayMs:    int(services.DefaultRetryPolicy.InitialDelay / time.Millisecond),
		CircuitThreshold:  5,
		CircuitTimeoutSec: 60,
	}
	if strategyCode == "" {
		return snapshot
	}
	cfg, ok := services.GetSmartRouter().GetStrategyConfig(strategyCode)
	if !ok {
		return snapshot
	}
	if cfg.MaxRetryCount > 0 {
		snapshot.MaxRetries = cfg.MaxRetryCount
	}
	if cfg.RetryBackoffBase > 0 {
		snapshot.InitialDelayMs = cfg.RetryBackoffBase
	}
	if cfg.CircuitBreakerThreshold > 0 {
		snapshot.CircuitThreshold = cfg.CircuitBreakerThreshold
	}
	if cfg.CircuitBreakerTimeout > 0 {
		snapshot.CircuitTimeoutSec = cfg.CircuitBreakerTimeout
	}
	return snapshot
}

func buildRoutingDecisionPayload(candidatesJSON []byte, snapshot strategyRuntimeSnapshot, effectivePolicySource string) []byte {
	var candidates any = []any{}
	if len(candidatesJSON) > 0 {
		var parsed []map[string]any
		if err := json.Unmarshal(candidatesJSON, &parsed); err == nil {
			candidates = parsed
		}
	}
	body := map[string]any{
		"candidates":       candidates,
		"strategy_runtime": snapshot,
	}
	if strings.TrimSpace(effectivePolicySource) != "" {
		body["effective_policy_source"] = strings.TrimSpace(effectivePolicySource)
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return candidatesJSON
	}
	return payload
}

func candidateScoreFromMap(m map[string]any) float64 {
	if m == nil {
		return 0
	}
	for _, k := range []string{"Score", "score"} {
		if v, ok := m[k]; ok {
			switch x := v.(type) {
			case float64:
				return x
			case json.Number:
				f, _ := x.Float64()
				return f
			}
		}
	}
	return 0
}

func intFromMap(m map[string]any, keys ...string) int {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			switch x := v.(type) {
			case float64:
				return int(x)
			case int:
				return x
			case json.Number:
				i64, _ := x.Int64()
				return int(i64)
			}
		}
	}
	return 0
}

func stringFromMap(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return ""
}

// summarizeRoutingCandidatesForTrace 从 candidates JSON 解析条数与得分最高项（供 trace 摘要）
func summarizeRoutingCandidatesForTrace(candidatesJSON json.RawMessage) (int, *traceTopCandidate) {
	if len(candidatesJSON) == 0 {
		return 0, nil
	}
	var items []map[string]any
	if err := json.Unmarshal(candidatesJSON, &items); err != nil || len(items) == 0 {
		return 0, nil
	}
	bestIdx := 0
	bestScore := candidateScoreFromMap(items[0])
	for i := 1; i < len(items); i++ {
		s := candidateScoreFromMap(items[i])
		if s > bestScore {
			bestScore = s
			bestIdx = i
		}
	}
	best := items[bestIdx]
	return len(items), &traceTopCandidate{
		APIKeyID: intFromMap(best, "APIKeyID", "api_key_id", "ApiKeyID"),
		Provider: stringFromMap(best, "Provider", "provider"),
		Model:    stringFromMap(best, "Model", "model"),
		Score:    candidateScoreFromMap(best),
	}
}

// normalizeEffectivePolicySource 将 trace 展示的 policy 来源限定为 env / db / default（无落库或非三者之一时视为 default）
func normalizeEffectivePolicySource(stored string) string {
	switch strings.TrimSpace(stored) {
	case policySourceEnv, policySourceDB, policySourceDefault:
		return strings.TrimSpace(stored)
	default:
		return policySourceDefault
	}
}

func parseRoutingDecisionPayload(raw json.RawMessage) (routingDecisionPayload, error) {
	if len(raw) == 0 {
		return routingDecisionPayload{}, nil
	}
	var wrapped routingDecisionPayload
	if err := json.Unmarshal(raw, &wrapped); err == nil && (len(wrapped.Candidates) > 0 || wrapped.StrategyRuntime.StrategyCode != "" || strings.TrimSpace(wrapped.EffectivePolicySource) != "") {
		return wrapped, nil
	}
	return routingDecisionPayload{}, fmt.Errorf("invalid routing decision payload")
}

func GetAPIProviders(c *gin.Context) {
	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusOK, legacyProviderList())
		return
	}

	rows, err := db.Query(
		`SELECT code, name, api_format FROM model_providers WHERE status = 'active' ORDER BY sort_order ASC`,
	)
	if err != nil {
		c.JSON(http.StatusOK, legacyProviderList())
		return
	}
	defer rows.Close()

	providers := make([]map[string]interface{}, 0)
	for rows.Next() {
		var code, name, apiFormat string
		if scanErr := rows.Scan(&code, &name, &apiFormat); scanErr != nil {
			continue
		}
		providers = append(providers, map[string]interface{}{
			"name":         code,
			"display_name": name,
			"api_format":   apiFormat,
		})
	}
	if len(providers) == 0 {
		providers = legacyProviderList()
	}

	c.JSON(http.StatusOK, providers)
}

func GetAPIUsageStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}
	userIDInt, ok := userID.(int)
	if !ok {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var stats struct {
		TotalRequests int     `json:"total_requests"`
		TotalTokens   int     `json:"total_tokens"`
		TotalCost     float64 `json:"total_cost"`
		AvgLatencyMs  int     `json:"avg_latency_ms"`
	}

	db.QueryRow(
		"SELECT COUNT(*), COALESCE(SUM(input_tokens + output_tokens), 0), COALESCE(SUM(cost), 0), COALESCE(AVG(latency_ms), 0) FROM api_usage_logs WHERE user_id = $1",
		userIDInt,
	).Scan(&stats.TotalRequests, &stats.TotalTokens, &stats.TotalCost, &stats.AvgLatencyMs)

	rows, err := db.Query(
		"SELECT provider, COUNT(*) as count, SUM(cost) as cost FROM api_usage_logs WHERE user_id = $1 GROUP BY provider ORDER BY count DESC",
		userIDInt,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	byProvider := make([]map[string]interface{}, 0)
	for rows.Next() {
		var provider string
		var count int
		var cost float64
		if err := rows.Scan(&provider, &count, &cost); err != nil {
			continue
		}
		byProvider = append(byProvider, map[string]interface{}{
			"provider": provider,
			"count":    count,
			"cost":     cost,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"stats":       stats,
		"by_provider": byProvider,
	})
}

func GetAPIRequestTrace(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}
	userIDInt, ok := userID.(int)
	if !ok {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	requestID := strings.TrimSpace(c.Param("request_id"))
	if requestID == "" {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var usage struct {
		RequestID    string  `json:"request_id"`
		Provider     string  `json:"provider"`
		Model        string  `json:"model"`
		StatusCode   int     `json:"status_code"`
		LatencyMS    int     `json:"latency_ms"`
		InputTokens  int     `json:"input_tokens"`
		OutputTokens int     `json:"output_tokens"`
		Cost         float64 `json:"cost"`
	}
	err := db.QueryRow(
		`SELECT request_id, provider, model, status_code, latency_ms, input_tokens, output_tokens, cost
		 FROM api_usage_logs WHERE request_id = $1 AND user_id = $2 LIMIT 1`,
		requestID, userIDInt,
	).Scan(&usage.RequestID, &usage.Provider, &usage.Model, &usage.StatusCode, &usage.LatencyMS, &usage.InputTokens, &usage.OutputTokens, &usage.Cost)
	if err != nil {
		if err == sql.ErrNoRows {
			middleware.RespondWithError(c, apperrors.NewAppError(
				"TRACE_NOT_FOUND",
				"Trace not found",
				http.StatusNotFound,
				nil,
			))
			return
		}
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var decision struct {
		StrategyUsed          string                  `json:"strategy_used"`
		Candidates            json.RawMessage         `json:"candidates"`
		StrategyRuntime       strategyRuntimeSnapshot `json:"strategy_runtime"`
		SelectedAPIKeyID      *int                    `json:"selected_api_key_id"`
		DecisionLatencyMS     int                     `json:"decision_latency_ms"`
		WasRetry              bool                    `json:"was_retry"`
		RetryCount            int                     `json:"retry_count"`
		CandidatesCount       int                     `json:"candidates_count"`
		TopCandidate          *traceTopCandidate      `json:"top_candidate,omitempty"`
		EffectivePolicySource string                  `json:"effective_policy_source"`
	}
	var selectedID sql.NullInt64
	var decisionRaw json.RawMessage
	rdErr := db.QueryRow(
		`SELECT strategy_used, COALESCE(candidates, '[]'::jsonb), selected_api_key_id, decision_latency_ms, was_retry, retry_count
		 FROM routing_decisions WHERE request_id = $1 AND user_id = $2 ORDER BY created_at DESC LIMIT 1`,
		requestID, userIDInt,
	).Scan(&decision.StrategyUsed, &decisionRaw, &selectedID, &decision.DecisionLatencyMS, &decision.WasRetry, &decision.RetryCount)
	if rdErr == nil && selectedID.Valid {
		val := int(selectedID.Int64)
		decision.SelectedAPIKeyID = &val
	}
	storedPolicySource := ""
	if rdErr == nil {
		parsedPayload, parseErr := parseRoutingDecisionPayload(decisionRaw)
		if parseErr == nil {
			decision.Candidates = parsedPayload.Candidates
			decision.StrategyRuntime = parsedPayload.StrategyRuntime
			storedPolicySource = parsedPayload.EffectivePolicySource
		} else {
			decision.Candidates = decisionRaw
		}
		decision.CandidatesCount, decision.TopCandidate = summarizeRoutingCandidatesForTrace(decision.Candidates)
		decision.EffectivePolicySource = normalizeEffectivePolicySource(storedPolicySource)
	}

	c.JSON(http.StatusOK, gin.H{
		"usage":    usage,
		"decision": decision,
	})
}
