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
	"sort"
	"strconv"
	"strings"
	"sync"
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
	"github.com/pintuotuo/backend/tracing"
	"github.com/pintuotuo/backend/utils"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var (
	proxyHTTPOnce         sync.Once
	proxyHTTPRoundTripper http.RoundTripper
)

func proxyHTTPClient(timeout time.Duration) *http.Client {
	proxyHTTPOnce.Do(func() {
		proxyHTTPRoundTripper = otelhttp.NewTransport(http.DefaultTransport)
	})
	return &http.Client{Transport: proxyHTTPRoundTripper, Timeout: timeout}
}

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
	Code           string
	Name           string
	APIBaseURL     string
	APIFormat      string
	ProviderRegion string
	RouteStrategy  map[string]interface{}
	Endpoints      map[string]interface{}
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

const (
	ImageDetailLow  = "low"
	ImageDetailHigh = "high"
	ImageDetailAuto = "auto"
)

type ImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

type ContentPart struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

type MessageContent struct {
	Text  string        `json:"-"`
	Parts []ContentPart `json:"-"`
}

func (mc *MessageContent) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		mc.Text = str
		mc.Parts = nil
		return nil
	}

	var parts []ContentPart
	if err := json.Unmarshal(data, &parts); err == nil {
		mc.Text = ""
		if len(parts) == 0 {
			mc.Parts = nil
		} else {
			mc.Parts = parts
		}
		return nil
	}

	return json.Unmarshal(data, &str)
}

func (mc MessageContent) MarshalJSON() ([]byte, error) {
	if mc.Parts != nil {
		return json.Marshal(mc.Parts)
	}
	return json.Marshal(mc.Text)
}

type ChatMessage struct {
	Role    string         `json:"role"`
	Content MessageContent `json:"content"`
	Name    string         `json:"name,omitempty"`
}

func estimateImageTokensWithSize(detail string, width, height int) int {
	switch detail {
	case ImageDetailLow:
		return 85
	case ImageDetailHigh:
		if width <= 0 || height <= 0 {
			return 765
		}
		tiles := ((width + 511) / 512) * ((height + 511) / 512)
		return 85 + tiles*170 + 255
	default:
		if width > 0 && height > 0 {
			tiles := ((width + 511) / 512) * ((height + 511) / 512)
			return 85 + tiles*170 + 255
		}
		return 765
	}
}

func estimateInputTokens(messages []ChatMessage) int {
	totalChars := 0
	for _, msg := range messages {
		totalChars += len(msg.Role)
		if msg.Content.Text != "" {
			totalChars += len(msg.Content.Text)
		}
		for _, part := range msg.Content.Parts {
			if part.Type == "text" {
				totalChars += len(part.Text)
			} else if part.Type == "image_url" {
				detail := ""
				if part.ImageURL != nil {
					detail = part.ImageURL.Detail
				}
				totalChars += estimateImageTokensWithSize(detail, 0, 0)
			}
		}
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
		// PreDeductBalance wraps many failures; only true short-balance cases should map to 409.
		if strings.Contains(preDeductErr.Error(), "insufficient balance") {
			middleware.RespondWithError(c, apperrors.ErrInsufficientBalance)
		} else {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		}
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

	var entCtx *services.EntitlementRoutingContext
	if services.EntitlementEnforcementStrict() {
		var entErr error
		entCtx, entErr = services.ResolveEntitlementRoutingContext(db, userIDInt, req.Provider, req.Model)
		if entErr != nil {
			billingEngine.CancelPreDeduct(userIDInt, requestID)
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		if req.APIKeyID != nil && *req.APIKeyID > 0 && !entCtx.AllowsAPIKey(*req.APIKeyID) {
			billingEngine.CancelPreDeduct(userIDInt, requestID)
			middleware.RespondWithError(c, apperrors.NewAppError(
				"API_KEY_NOT_AUTHORIZED",
				"api_key_id is not allowed for your entitlement",
				http.StatusForbidden,
				nil,
			))
			return
		}
		if req.MerchantSKUID != nil && *req.MerchantSKUID > 0 && !entCtx.AllowsMerchantSKU(*req.MerchantSKUID) {
			billingEngine.CancelPreDeduct(userIDInt, requestID)
			middleware.RespondWithError(c, apperrors.NewAppError(
				"API_KEY_NOT_AUTHORIZED",
				"merchant_sku_id is not allowed for your entitlement",
				http.StatusForbidden,
				nil,
			))
			return
		}
	}

	keyFilter := entitlementKeyFilterForRouter(services.EntitlementEnforcementStrict(), entCtx)

	selectedStrategy := "legacy_fallback"
	effectivePolicySource := ""
	var smartCandidatesJSON []byte
	var currentRoutingDecision *services.RoutingDecision
	decisionStart := time.Now()
	if req.APIKeyID == nil && req.MerchantSKUID == nil && shouldUseSmartRouting(userIDInt, requestID) {
		strategyCode, policySrc := routingStrategyWithSource()
		if smartReq := trySelectAPIKeyWithSmartRouter(req, strategyCode, keyFilter, requestID, merchantID); smartReq.APIKeyID != nil {
			req.APIKeyID = smartReq.APIKeyID
			if entCtx != nil {
				if msid, ok := entCtx.MerchantSKUForAPIKey(*req.APIKeyID); ok && req.MerchantSKUID == nil {
					req.MerchantSKUID = &msid
				}
			}
			selectedStrategy = strategyCode
			smartCandidatesJSON = smartReq.CandidatesJSON
			effectivePolicySource = policySrc
			currentRoutingDecision = smartReq.RoutingDecision
		}
	}
	if req.APIKeyID == nil && req.MerchantSKUID == nil && services.EntitlementEnforcementStrict() && entCtx != nil && len(entCtx.AllowedAPIKeyIDs) > 0 {
		if pick, msid := pickDeterministicEntitledKey(entCtx); pick > 0 {
			p := pick
			req.APIKeyID = &p
			if req.MerchantSKUID == nil && msid > 0 {
				req.MerchantSKUID = &msid
			}
		}
	}

	var apiKey models.MerchantAPIKey
	err = selectAPIKeyForRequest(db, userIDInt, merchantID, req, &apiKey, entCtx)
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
	providerCfg = applyAPIKeyRouteConfig(providerCfg, &apiKey)
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

	strategySnapshot := buildStrategyRuntimeSnapshot(selectedStrategy)
	applyCircuitBreakerConfig(apiKey.ID, strategySnapshot)

	if req.Stream {
		if providerCfg.APIFormat != apiFormatOpenAI {
			billingEngine.CancelPreDeduct(userIDInt, requestID)
			middleware.RespondWithError(c, apperrors.NewAppError(
				"STREAMING_NOT_SUPPORTED",
				"Streaming is only supported for OpenAI-compatible providers",
				http.StatusBadRequest,
				nil,
			))
			return
		}
		streamClient := proxyHTTPClient(15 * time.Minute)
		retryPolicyStream := buildRetryPolicy(strategySnapshot)
		retryPolicyStream = applyLitellmGatewayRetryCap(retryPolicyStream, resolveRouteModeWithProvider(&apiKey, providerCfg.ProviderRegion))
		streamAttempts := buildProxyCatalogAttempts(c.Request.Context(), db, req)
		var retryCountStream int
		for attemptIdx, att := range streamAttempts {
			traceSpan.SetRoute(att.provider, att.model)

			pk, dk, pcfg, skip, fatalErr := resolveProxyAttemptRuntime(
				c.Request.Context(),
				db,
				userIDInt,
				merchantID,
				req,
				att,
				apiKey,
				decryptedKey,
				providerCfg,
				entCtx,
				requestID,
			)
			if fatalErr != nil {
				billingEngine.CancelPreDeduct(userIDInt, requestID)
				middleware.RespondWithError(c, apperrors.ErrDatabaseError)
				return
			}
			if skip {
				continue
			}

			if pcfg.APIFormat != apiFormatOpenAI {
				if attemptIdx < len(streamAttempts)-1 {
					continue
				}
				billingEngine.CancelPreDeduct(userIDInt, requestID)
				middleware.RespondWithError(c, apperrors.NewAppError(
					"STREAMING_NOT_SUPPORTED",
					"Streaming is only supported for OpenAI-compatible providers",
					http.StatusBadRequest,
					nil,
				))
				return
			}

			base := strings.TrimRight(pcfg.APIBaseURL, "/")
			if base == "" {
				if attemptIdx < len(streamAttempts)-1 {
					continue
				}
				billingEngine.CancelPreDeduct(userIDInt, requestID)
				middleware.RespondWithError(c, apperrors.NewAppError(
					"UNSUPPORTED_PROVIDER",
					fmt.Sprintf("Provider %s is missing api_base_url", att.provider),
					http.StatusBadRequest,
					nil,
				))
				return
			}

			ep := fmt.Sprintf("%s/chat/completions", base)
			rb := map[string]interface{}{
				"model":    att.model,
				"messages": req.Messages,
				"stream":   true,
			}
			if req.Options != nil {
				var options map[string]interface{}
				if unmarshalErr := json.Unmarshal(req.Options, &options); unmarshalErr == nil {
					for k, v := range options {
						rb[k] = v
					}
				}
			}
			routeMode := resolveRouteModeWithProvider(&pk, pcfg.ProviderRegion)
			if routeMode == routeModeLitellm {
				litellmTemplate, tmplErr := getProviderLitellmTemplate(db, att.provider)
				if tmplErr != nil {
					litellmTemplate = att.model
				}
				litellmModel := litellmTemplate
				if strings.Contains(litellmTemplate, "{model_id}") {
					litellmModel = strings.ReplaceAll(litellmTemplate, "{model_id}", att.model)
				} else if litellmTemplate == "" {
					litellmModel = "openai/" + att.model
				}
				rb["model"] = litellmModel
				apiBaseForUser := pk.EndpointURL
				if apiBaseForUser == "" {
					apiBaseForUser = pcfg.APIBaseURL
				}
				rb["api_key"] = dk
				rb["api_base"] = apiBaseForUser
			}
			jb, mErr := json.Marshal(rb)
			if mErr != nil {
				billingEngine.CancelPreDeduct(userIDInt, requestID)
				middleware.RespondWithError(c, apperrors.NewAppError(
					"REQUEST_BUILD_FAILED",
					"Failed to build request body",
					http.StatusInternalServerError,
					mErr,
				))
				return
			}
			hreq, hErr := http.NewRequestWithContext(c.Request.Context(), "POST", ep, bytes.NewBuffer(jb))
			if hErr != nil {
				billingEngine.CancelPreDeduct(userIDInt, requestID)
				middleware.RespondWithError(c, apperrors.NewAppError(
					"REQUEST_CREATE_FAILED",
					"Failed to create request",
					http.StatusInternalServerError,
					hErr,
				))
				return
			}
			hreq.GetBody = func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader(jb)), nil
			}
			hreq.Header.Set("Content-Type", "application/json")
			authToken := resolveAuthTokenFromRouteMode(resolveRouteModeWithProvider(&pk, pcfg.ProviderRegion), dk)
			hreq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
			hreq.Header.Set("Accept", "text/event-stream")
			applyProxyUpstreamHeaders(c, hreq, requestID)

			r, rc, execErr := executeProviderRequestWithRetry(streamClient, hreq, retryPolicyStream)
			retryCountStream += rc

			if execErr != nil {
				info := providerInfoFromUpstreamFailure(0, nil, nil, execErr)
				services.GetSmartRouter().RecordRequestResult(pk.ID, false)
				recordHealthCheckerProxyOutcome(c, pk.ID, false, startTime)
				if attemptIdx < len(streamAttempts)-1 && services.SuggestModelFallbackAfterFailure(info) {
					logger.LogInfo(c.Request.Context(), "api_proxy", "stream model fallback after transport error", map[string]interface{}{
						"request_id": requestID, "attempt": att.provider + "/" + att.model,
					})
					continue
				}
				traceSpan.SetStatusCode(http.StatusBadGateway)
				traceSpan.SetErrorCode("API_STREAM_REQUEST_FAILED")
				billingEngine.CancelPreDeduct(userIDInt, requestID)
				middleware.RespondWithError(c, apperrors.NewAppError(
					"API_REQUEST_FAILED",
					"Failed to send streaming request to provider",
					http.StatusBadGateway,
					execErr,
				))
				return
			}

			if r.StatusCode != http.StatusOK {
				bRead, _ := io.ReadAll(io.LimitReader(r.Body, upstreamErrorBodyPeek))
				_ = r.Body.Close()
				info := providerInfoFromUpstreamFailure(r.StatusCode, bRead, r.Header, nil)
				services.GetSmartRouter().RecordRequestResult(pk.ID, false)
				recordHealthCheckerProxyOutcome(c, pk.ID, false, startTime)
				if attemptIdx < len(streamAttempts)-1 && services.SuggestModelFallbackAfterFailure(info) {
					logger.LogInfo(c.Request.Context(), "api_proxy", "stream model fallback after upstream HTTP error", map[string]interface{}{
						"request_id": requestID, "status": r.StatusCode, "attempt": att.provider + "/" + att.model,
					})
					continue
				}
				traceSpan.SetStatusCode(r.StatusCode)
				billingEngine.CancelPreDeduct(userIDInt, requestID)
				c.Data(r.StatusCode, "application/json", bRead)
				return
			}

			executeProxyChatCompletionStreamFromUpstream(c, r, requestID, userIDInt, req, att.provider, att.model, requestPath, startTime, db,
				billingEngine, pk, merchantID, strictPricingVID, selectedStrategy, smartCandidatesJSON, effectivePolicySource,
				decisionStart, traceSpan, strategySnapshot, retryCountStream)
			return
		}

		traceSpan.SetStatusCode(http.StatusBadGateway)
		traceSpan.SetErrorCode("API_REQUEST_FAILED")
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		middleware.RespondWithError(c, apperrors.NewAppError(
			"API_REQUEST_FAILED",
			"Failed to complete streaming request via model fallback chain",
			http.StatusBadGateway,
			nil,
		))
		return
	}

	client := proxyHTTPClient(60 * time.Second)
	retryPolicy := buildRetryPolicy(strategySnapshot)
	retryPolicy = applyLitellmGatewayRetryCap(retryPolicy, resolveRouteModeWithProvider(&apiKey, providerCfg.ProviderRegion))
	attempts := buildProxyCatalogAttempts(c.Request.Context(), db, req)

	var (
		resp            *http.Response
		body            []byte
		retryCountTotal int
		billProv        string
		billModel       string
		winningKey      models.MerchantAPIKey
	)

	for attemptIdx, att := range attempts {
		traceSpan.SetRoute(att.provider, att.model)

		pk, dk, pcfg, skip, fatalErr := resolveProxyAttemptRuntime(
			c.Request.Context(),
			db,
			userIDInt,
			merchantID,
			req,
			att,
			apiKey,
			decryptedKey,
			providerCfg,
			entCtx,
			requestID,
		)
		if fatalErr != nil {
			billingEngine.CancelPreDeduct(userIDInt, requestID)
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		if skip {
			continue
		}

		base := strings.TrimRight(pcfg.APIBaseURL, "/")
		if base == "" {
			if attemptIdx < len(attempts)-1 {
				continue
			}
			billingEngine.CancelPreDeduct(userIDInt, requestID)
			middleware.RespondWithError(c, apperrors.NewAppError(
				"UNSUPPORTED_PROVIDER",
				fmt.Sprintf("Provider %s is missing api_base_url", att.provider),
				http.StatusBadRequest,
				nil,
			))
			return
		}

		ep := fmt.Sprintf("%s/chat/completions", base)
		if pcfg.APIFormat == providerAnthropic {
			ep = fmt.Sprintf("%s/messages", base)
		}

		rb := map[string]interface{}{
			"model":    att.model,
			"messages": req.Messages,
			"stream":   req.Stream,
		}
		if req.Options != nil {
			var options map[string]interface{}
			if unmarshalErr := json.Unmarshal(req.Options, &options); unmarshalErr == nil {
				for k, v := range options {
					rb[k] = v
				}
			}
		}
		routeMode := resolveRouteModeWithProvider(&pk, pcfg.ProviderRegion)
		if routeMode == routeModeLitellm {
			litellmTemplate, tmplErr := getProviderLitellmTemplate(db, att.provider)
			if tmplErr != nil {
				litellmTemplate = att.model
			}
			litellmModel := litellmTemplate
			if strings.Contains(litellmTemplate, "{model_id}") {
				litellmModel = strings.ReplaceAll(litellmTemplate, "{model_id}", att.model)
			} else if litellmTemplate == "" {
				litellmModel = "openai/" + att.model
			}
			apiBaseForUser := pk.EndpointURL
			if apiBaseForUser == "" {
				apiBaseForUser = pcfg.APIBaseURL
			}
			rb["model"] = litellmModel
			rb["api_key"] = dk
			rb["api_base"] = apiBaseForUser
		}
		jb, mErr := json.Marshal(rb)
		if mErr != nil {
			billingEngine.CancelPreDeduct(userIDInt, requestID)
			middleware.RespondWithError(c, apperrors.NewAppError(
				"REQUEST_BUILD_FAILED",
				"Failed to build request body",
				http.StatusInternalServerError,
				mErr,
			))
			return
		}
		hreq, hErr := http.NewRequestWithContext(c.Request.Context(), "POST", ep, bytes.NewBuffer(jb))
		if hErr != nil {
			billingEngine.CancelPreDeduct(userIDInt, requestID)
			middleware.RespondWithError(c, apperrors.NewAppError(
				"REQUEST_CREATE_FAILED",
				"Failed to create request",
				http.StatusInternalServerError,
				hErr,
			))
			return
		}
		hreq.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(jb)), nil
		}
		hreq.Header.Set("Content-Type", "application/json")
		switch pcfg.APIFormat {
		case providerAnthropic:
			hreq.Header.Set("x-api-key", dk)
			hreq.Header.Set("anthropic-version", "2023-06-01")
		default:
			authToken := resolveAuthTokenFromRouteMode(resolveRouteModeWithProvider(&pk, pcfg.ProviderRegion), dk)
			hreq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
		}
		applyProxyUpstreamHeaders(c, hreq, requestID)

		r, rc, execErr := executeProviderRequestWithRetry(client, hreq, retryPolicy)
		retryCountTotal += rc

		if execErr != nil {
			info := providerInfoFromUpstreamFailure(0, nil, nil, execErr)
			services.GetSmartRouter().RecordRequestResult(pk.ID, false)
			recordHealthCheckerProxyOutcome(c, pk.ID, false, startTime)
			if attemptIdx < len(attempts)-1 && services.SuggestModelFallbackAfterFailure(info) {
				logger.LogInfo(c.Request.Context(), "api_proxy", "model fallback after transport error", map[string]interface{}{
					"request_id": requestID, "attempt": att.provider + "/" + att.model,
				})
				continue
			}
			traceSpan.SetStatusCode(http.StatusBadGateway)
			traceSpan.SetErrorCode("API_REQUEST_FAILED")
			billingEngine.CancelPreDeduct(userIDInt, requestID)
			middleware.RespondWithError(c, apperrors.NewAppError(
				"API_REQUEST_FAILED",
				"Failed to send request to provider",
				http.StatusBadGateway,
				execErr,
			))
			return
		}

		bRead, readErr := io.ReadAll(r.Body)
		_ = r.Body.Close()
		if readErr != nil {
			billingEngine.CancelPreDeduct(userIDInt, requestID)
			recordHealthCheckerProxyOutcome(c, pk.ID, false, startTime)
			middleware.RespondWithError(c, apperrors.NewAppError(
				"RESPONSE_READ_FAILED",
				"Failed to read response",
				http.StatusInternalServerError,
				readErr,
			))
			return
		}

		if chatCompletionJSONIndicatesProxySuccess(r.StatusCode, bRead) {
			resp = r
			body = bRead
			billProv = att.provider
			billModel = att.model
			winningKey = pk
			proxyTransportOK := r.StatusCode < http.StatusInternalServerError && r.StatusCode != http.StatusTooManyRequests
			services.GetSmartRouter().RecordRequestResult(pk.ID, proxyTransportOK)
			recordHealthCheckerProxyOutcome(c, pk.ID, true, startTime)
			break
		}

		info := providerInfoFromUpstreamFailure(r.StatusCode, bRead, r.Header, nil)
		services.GetSmartRouter().RecordRequestResult(pk.ID, false)
		recordHealthCheckerProxyOutcome(c, pk.ID, false, startTime)
		if attemptIdx < len(attempts)-1 && services.SuggestModelFallbackAfterFailure(info) {
			logger.LogInfo(c.Request.Context(), "api_proxy", "model fallback after upstream HTTP error", map[string]interface{}{
				"request_id": requestID, "status": r.StatusCode, "attempt": att.provider + "/" + att.model,
			})
			continue
		}

		traceSpan.SetStatusCode(r.StatusCode)
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		c.Data(r.StatusCode, "application/json", bRead)
		return
	}

	if resp == nil || body == nil {
		traceSpan.SetStatusCode(http.StatusBadGateway)
		traceSpan.SetErrorCode("API_REQUEST_FAILED")
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		middleware.RespondWithError(c, apperrors.NewAppError(
			"API_REQUEST_FAILED",
			"Failed to complete request via model fallback chain",
			http.StatusBadGateway,
			nil,
		))
		return
	}

	retryCount := retryCountTotal

	var apiResp APIProxyResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		traceSpan.SetStatusCode(resp.StatusCode)
		traceSpan.SetErrorCode("UNMARSHAL_PROXY_RESPONSE_FAILED")
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		c.Data(resp.StatusCode, "application/json", body)
		return
	}

	if sm := strings.TrimSpace(apiResp.Model); sm != "" && sm != strings.TrimSpace(billModel) {
		logger.LogInfo(c.Request.Context(), "api_proxy", "upstream response model differs from request (e.g. gateway fallback)", map[string]interface{}{
			"request_id": requestID, "requested_model": billModel, "response_model": sm,
		})
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
		cost, cerr = calculateTokenCost(db, userIDInt, billProv, billModel, inputTokens, outputTokens, strictPricingVID)
		if cerr != nil {
			billingEngine.CancelPreDeduct(userIDInt, requestID)
			logger.LogError(context.Background(), "api_proxy", "Token cost resolution failed", cerr, map[string]interface{}{
				"user_id": userIDInt, "provider": billProv, "model": billModel, "request_id": requestID,
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
		billReq := req
		billReq.Provider = billProv
		billReq.Model = billModel
		logMerchantSKUID, logProcurementCNY := resolveMerchantSKUProcurementForLog(db, billReq, winningKey.ID, merchantID, inputTokens, outputTokens)
		tx, err := db.Begin()
		if err == nil {
			_, updateErr := tx.Exec(
				"UPDATE merchant_api_keys SET quota_used = quota_used + $1, last_used_at = $2 WHERE id = $3",
				cost, time.Now(), winningKey.ID,
			)
			err = updateErr
			if err == nil {
				_, err = tx.Exec(
					"INSERT INTO api_usage_logs (user_id, key_id, request_id, provider, model, method, path, status_code, latency_ms, input_tokens, output_tokens, cost, token_usage, merchant_sku_id, procurement_cost_cny) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)",
					userIDInt, winningKey.ID, requestID, billProv, billModel, "POST", requestPath, resp.StatusCode, latency, inputTokens, outputTokens, cost, tokenUsage, nullInt64Arg(logMerchantSKUID), nullFloat64Arg(logProcurementCNY),
				)
			}

			if err != nil {
				tx.Rollback()
				logger.LogError(context.Background(), "api_proxy", "Transaction rollback", err, map[string]interface{}{
					"user_id":     userIDInt,
					"provider":    billProv,
					"model":       billModel,
					"cost":        cost,
					"token_usage": tokenUsage,
					"request_id":  requestID,
				})
			} else {
				tx.Commit()
				logger.LogInfo(context.Background(), "api_proxy", "API request completed", map[string]interface{}{
					"user_id":       userIDInt,
					"provider":      billProv,
					"model":         billModel,
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

	if currentRoutingDecision != nil {
		pipeline := services.NewThreeLayerRoutingPipeline()

		gatewayMode := resolveRouteModeWithProvider(&apiKey, providerCfg.ProviderRegion)

		execInput := &services.ExecutionLayerInputData{
			GatewayMode:   gatewayMode,
			EndpointURL:   fmt.Sprintf("%s/chat/completions", strings.TrimRight(providerCfg.APIBaseURL, "/")),
			EndpointType:  services.EndpointTypeChatCompletions,
			AuthMethod:    "bearer",
			ResolvedModel: billModel,
			RequestFormat: providerCfg.APIFormat,
		}
		pipeline.RecordExecutionInput(currentRoutingDecision, execInput)

		execSuccess := resp.StatusCode >= 200 && resp.StatusCode < 300
		execLatency := int(time.Since(decisionStart).Milliseconds())
		var execErrMsg string
		if !execSuccess {
			execErrMsg = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}

		execResult := &services.ExecutionLayerResultData{
			Success:      execSuccess,
			StatusCode:   resp.StatusCode,
			LatencyMs:    execLatency,
			ErrorMessage: execErrMsg,
			Model:        currentRoutingDecision.SelectedModel,
			Provider:     currentRoutingDecision.SelectedProvider,
			ActualModel:  apiResp.Model,
			InputTokens:  apiResp.Usage.PromptTokens,
			OutputTokens: apiResp.Usage.CompletionTokens,
		}
		if len(apiResp.Choices) > 0 {
			execResult.FinishReason = apiResp.Choices[0].FinishReason
		}
		pipeline.RecordExecutionResultExtended(currentRoutingDecision, execResult)

		engine := services.NewUnifiedRoutingEngine()
		if logErr := engine.LogDecision(c.Request.Context(), currentRoutingDecision); logErr != nil {
			logger.LogWarn(c.Request.Context(), "api_proxy", "Failed to log routing decision", map[string]interface{}{
				"request_id": requestID,
				"error":      logErr.Error(),
			})
		}
	} else {
		decisionPayload := buildRoutingDecisionPayload(smartCandidatesJSON, strategySnapshot, effectivePolicySource)
		_ = insertRoutingDecision(db, requestID, userIDInt, req, selectedStrategy, decisionPayload, winningKey.ID, int(time.Since(decisionStart).Milliseconds()), retryCount)
	}

	traceSpan.SetStatusCode(resp.StatusCode)
	c.Data(resp.StatusCode, "application/json", body)
}

// applyProxyUpstreamHeaders 将本请求的追踪关联信息传给上游（含 LiteLLM），便于与网关 /metrics、日志对齐定界。
// 见 deploy/litellm/README.md 与 LiteLLM reliability 文档。
// applyLitellmGatewayRetryCap 在走 LiteLLM 网关时限制业务层 HTTP 重试次数，避免与网关 router num_retries 叠加放大。
// 环境变量 API_PROXY_LITELLM_MAX_RETRIES 表示 MaxRetries 上限（默认 1，即最多 2 次出站尝试）。
func applyLitellmGatewayRetryCap(policy *services.RetryPolicy, routeMode string) *services.RetryPolicy {
	if policy == nil {
		return policy
	}
	if routeMode != routeModeLitellm {
		return policy
	}
	cap := 1
	if v := strings.TrimSpace(os.Getenv("API_PROXY_LITELLM_MAX_RETRIES")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			cap = n
		}
	}
	if policy.MaxRetries > cap {
		p := *policy
		p.MaxRetries = cap
		return &p
	}
	return policy
}

func applyProxyUpstreamHeaders(c *gin.Context, httpReq *http.Request, requestID string) {
	if strings.TrimSpace(requestID) != "" {
		httpReq.Header.Set("X-Request-ID", requestID)
	}
	if tracing.Enabled() {
		for _, h := range []string{"tracestate", "baggage"} {
			if v := strings.TrimSpace(c.GetHeader(h)); v != "" {
				httpReq.Header.Set(h, v)
			}
		}
		return
	}
	for _, h := range []string{"traceparent", "tracestate", "baggage"} {
		if v := strings.TrimSpace(c.GetHeader(h)); v != "" {
			httpReq.Header.Set(h, v)
		}
	}
}

func resolveRouteModeWithProvider(apiKey *models.MerchantAPIKey, providerRegion string) string {
	if apiKey == nil {
		return routeModeDirect
	}
	mode := services.ResolveRouteModeWithProvider(apiKey.RouteMode, providerRegion)
	if mode == services.GatewayModeDirect {
		return routeModeDirect
	}
	if mode == services.GatewayModeLitellm {
		return routeModeLitellm
	}
	if mode == services.GatewayModeProxy {
		return routeModeProxy
	}
	return routeModeDirect
}

func applyAPIKeyRouteConfig(cfg providerRuntimeConfig, apiKey *models.MerchantAPIKey) providerRuntimeConfig {
	execCfg := &services.ExecutionProviderConfig{
		Code:           cfg.Code,
		Name:           cfg.Name,
		APIBaseURL:     cfg.APIBaseURL,
		APIFormat:      cfg.APIFormat,
		ProviderRegion: cfg.ProviderRegion,
		RouteStrategy:  cfg.RouteStrategy,
		Endpoints:      cfg.Endpoints,
	}
	if apiKey != nil {
		services.ApplyBYOKConfig(execCfg, apiKey.EndpointURL, apiKey.RouteMode, apiKey.RouteConfig, apiKey.FallbackEndpointURL)
		execCfg.GatewayMode = services.ResolveRouteModeWithProvider(apiKey.RouteMode, cfg.ProviderRegion)
	}
	endpointURL := services.ResolveEndpoint(execCfg)
	if endpointURL != "" {
		cfg.APIBaseURL = endpointURL
	}
	return cfg
}

func resolveAuthTokenFromRouteMode(routeMode string, fallbackToken string) string {
	if routeMode == routeModeLitellm {
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
	var routeStrategyJSON, endpointsJSON []byte
	err := db.QueryRow(
		`SELECT code, name, COALESCE(api_base_url, ''), api_format,
				COALESCE(provider_region, 'domestic'),
				COALESCE(route_strategy, '{}'::jsonb),
				COALESCE(endpoints, '{}'::jsonb)
		 FROM model_providers
		 WHERE code = $1 AND status = 'active'
		 LIMIT 1`,
		providerCode,
	).Scan(&cfg.Code, &cfg.Name, &cfg.APIBaseURL, &cfg.APIFormat,
		&cfg.ProviderRegion, &routeStrategyJSON, &endpointsJSON)
	if err != nil {
		return cfg, err
	}
	if len(routeStrategyJSON) > 0 {
		if err := json.Unmarshal(routeStrategyJSON, &cfg.RouteStrategy); err != nil {
			cfg.RouteStrategy = make(map[string]interface{})
		}
	}
	if len(endpointsJSON) > 0 {
		if err := json.Unmarshal(endpointsJSON, &cfg.Endpoints); err != nil {
			cfg.Endpoints = make(map[string]interface{})
		}
	}
	return cfg, nil
}

func getProviderLitellmTemplate(db *sql.DB, providerCode string) (string, error) {
	var template string
	err := db.QueryRow(
		`SELECT COALESCE(litellm_model_template, '') 
		 FROM model_providers 
		 WHERE code = $1 AND status = 'active' 
		 LIMIT 1`,
		providerCode,
	).Scan(&template)
	if err != nil {
		return "", err
	}
	return template, nil
}

// entitlementKeyFilterForRouter: nil = no filter (legacy); strict with no keys = empty slice (no SmartRouter pool).
func entitlementKeyFilterForRouter(strict bool, ent *services.EntitlementRoutingContext) []int {
	if !strict {
		return nil
	}
	if ent == nil || len(ent.AllowedAPIKeyIDs) == 0 {
		return []int{}
	}
	out := make([]int, 0, len(ent.AllowedAPIKeyIDs))
	for id := range ent.AllowedAPIKeyIDs {
		out = append(out, id)
	}
	sort.Ints(out)
	return out
}

func pickDeterministicEntitledKey(ent *services.EntitlementRoutingContext) (apiKeyID int, merchantSKUID int) {
	if ent == nil || len(ent.AllowedAPIKeyIDs) == 0 {
		return 0, 0
	}
	minK := 0
	for k := range ent.AllowedAPIKeyIDs {
		if minK == 0 || k < minK {
			minK = k
		}
	}
	msid, _ := ent.MerchantSKUForAPIKey(minK)
	return minK, msid
}

func scanMerchantAPIKeyQuotaRow(row *sql.Row, apiKey *models.MerchantAPIKey) error {
	var qLim sql.NullFloat64
	var endpointURL, fallbackEndpointURL, routeMode sql.NullString
	var routeConfigBytes []byte
	if err := row.Scan(
		&apiKey.ID, &apiKey.MerchantID, &apiKey.Provider, &apiKey.APIKeyEncrypted, &apiKey.APISecretEncrypted,
		&qLim, &apiKey.QuotaUsed, &apiKey.Status,
		&endpointURL, &fallbackEndpointURL, &routeMode, &routeConfigBytes,
	); err != nil {
		return err
	}
	apiKey.QuotaLimit = utils.NullFloat64Ptr(qLim)
	if endpointURL.Valid {
		apiKey.EndpointURL = endpointURL.String
	}
	if fallbackEndpointURL.Valid {
		apiKey.FallbackEndpointURL = fallbackEndpointURL.String
	}
	if routeMode.Valid {
		apiKey.RouteMode = routeMode.String
	}
	if len(routeConfigBytes) > 0 {
		_ = json.Unmarshal(routeConfigBytes, &apiKey.RouteConfig)
	}
	return nil
}

func selectAPIKeyForRequest(db *sql.DB, userID, merchantID int, req APIProxyRequest, apiKey *models.MerchantAPIKey, ent *services.EntitlementRoutingContext) error {
	if req.APIKeyID != nil && *req.APIKeyID > 0 {
		keyPick := `SELECT mak.id, mak.merchant_id, mak.provider, mak.api_key_encrypted, mak.api_secret_encrypted, mak.quota_limit, mak.quota_used, mak.status,
			 COALESCE(mak.endpoint_url, '') as endpoint_url,
			 COALESCE(mak.fallback_endpoint_url, '') as fallback_endpoint_url,
			 COALESCE(mak.route_mode, 'auto') as route_mode,
			 COALESCE(mak.route_config, '{}'::jsonb) as route_config
			 FROM merchant_api_keys mak
			 INNER JOIN merchants m ON m.id = mak.merchant_id
			 WHERE mak.id = $1 AND mak.provider = $2 AND mak.status = 'active'
			   AND (mak.verified_at IS NOT NULL OR mak.verification_result = 'verified')
			   AND (mak.quota_limit IS NULL OR mak.quota_used < mak.quota_limit)
			   AND m.status IN ('active', 'approved')
			   AND m.lifecycle_status <> 'suspended'`
		if merchantID <= 0 {
			if ent != nil && ent.AllowsAPIKey(*req.APIKeyID) {
				return scanMerchantAPIKeyQuotaRow(
					db.QueryRow(keyPick+` LIMIT 1`, *req.APIKeyID, req.Provider),
					apiKey,
				)
			}
			keyPick += ` AND m.user_id = $3`
			return scanMerchantAPIKeyQuotaRow(
				db.QueryRow(keyPick+` LIMIT 1`, *req.APIKeyID, req.Provider, userID),
				apiKey,
			)
		}
		keyPick += ` AND mak.merchant_id = $3 LIMIT 1`
		err := scanMerchantAPIKeyQuotaRow(
			db.QueryRow(keyPick, *req.APIKeyID, req.Provider, merchantID),
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
			if ent != nil && ent.AllowsMerchantSKU(*req.MerchantSKUID) {
				err := scanMerchantAPIKeyQuotaRow(
					db.QueryRow(
						`SELECT mak.id, mak.merchant_id, mak.provider, mak.api_key_encrypted, mak.api_secret_encrypted, mak.quota_limit, mak.quota_used, mak.status,
						 COALESCE(mak.endpoint_url, '') as endpoint_url,
						 COALESCE(mak.fallback_endpoint_url, '') as fallback_endpoint_url,
						 COALESCE(mak.route_mode, 'auto') as route_mode,
						 COALESCE(mak.route_config, '{}'::jsonb) as route_config
						 FROM merchant_skus ms
						 JOIN merchant_api_keys mak ON mak.id = ms.api_key_id
						 JOIN merchants m ON m.id = ms.merchant_id
						 WHERE ms.id = $1 AND ms.status = 'active'
						   AND mak.provider = $2 AND mak.status = 'active'
						   AND (mak.verified_at IS NOT NULL OR mak.verification_result = 'verified')
						   AND m.status IN ('active', 'approved')
						   AND m.lifecycle_status <> 'suspended'
						   AND (mak.quota_limit IS NULL OR mak.quota_used < mak.quota_limit)
						 LIMIT 1`,
						*req.MerchantSKUID, req.Provider,
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
			return sql.ErrNoRows
		}
		err := scanMerchantAPIKeyQuotaRow(
			db.QueryRow(
				`SELECT mak.id, mak.merchant_id, mak.provider, mak.api_key_encrypted, mak.api_secret_encrypted, mak.quota_limit, mak.quota_used, mak.status,
				 COALESCE(mak.endpoint_url, '') as endpoint_url,
				 COALESCE(mak.fallback_endpoint_url, '') as fallback_endpoint_url,
				 COALESCE(mak.route_mode, 'auto') as route_mode,
				 COALESCE(mak.route_config, '{}'::jsonb) as route_config
				 FROM merchant_skus ms
				 JOIN merchant_api_keys mak ON mak.id = ms.api_key_id
				 JOIN merchants m ON m.id = ms.merchant_id
				 WHERE ms.id = $1 AND ms.status = 'active'
				   AND ms.merchant_id = $2
				   AND m.user_id = $3
				   AND mak.provider = $4 AND mak.status = 'active'
				   AND (mak.verified_at IS NOT NULL OR mak.verification_result = 'verified')
				   AND m.lifecycle_status <> 'suspended'`,
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
				`SELECT id, merchant_id, provider, api_key_encrypted, api_secret_encrypted, quota_limit, quota_used, status,
				 COALESCE(endpoint_url, '') as endpoint_url,
				 COALESCE(fallback_endpoint_url, '') as fallback_endpoint_url,
				 COALESCE(route_mode, 'auto') as route_mode,
				 COALESCE(route_config, '{}'::jsonb) as route_config
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
			`SELECT mak.id, mak.merchant_id, mak.provider, mak.api_key_encrypted, mak.api_secret_encrypted, mak.quota_limit, mak.quota_used, mak.status,
			 COALESCE(mak.endpoint_url, '') as endpoint_url,
			 COALESCE(mak.fallback_endpoint_url, '') as fallback_endpoint_url,
			 COALESCE(mak.route_mode, 'auto') as route_mode,
			 COALESCE(mak.route_config, '{}'::jsonb) as route_config
			 FROM merchant_api_keys mak
			 INNER JOIN merchants m ON m.id = mak.merchant_id
			 WHERE mak.provider = $1 AND mak.status = 'active'
			   AND m.user_id = $2
			   AND m.status IN ('active', 'approved')
			   AND m.lifecycle_status <> 'suspended'
			   AND (mak.verified_at IS NOT NULL OR mak.verification_result = 'verified')
			   AND (mak.quota_limit IS NULL OR mak.quota_used < mak.quota_limit)
			 ORDER BY COALESCE((mak.quota_limit - mak.quota_used)::double precision, 1e30::double precision) DESC
			 LIMIT 1`,
			req.Provider, userID,
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
	err := db.QueryRow("SELECT id FROM merchants WHERE user_id = $1 AND "+sqlMerchantOperational+" LIMIT 1", userID).Scan(&merchantID)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return merchantID, nil
}

type smartRoutingPick struct {
	APIKeyID        *int
	CandidatesJSON  []byte
	RoutingDecision *services.RoutingDecision
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

func trySelectAPIKeyWithSmartRouter(req APIProxyRequest, strategyCode string, keyFilter []int, requestID string, merchantID int) smartRoutingPick {
	if strings.TrimSpace(req.Provider) == "" {
		return smartRoutingPick{}
	}
	if keyFilter != nil && len(keyFilter) == 0 {
		return smartRoutingPick{}
	}

	pipeline := services.NewThreeLayerRoutingPipeline()
	routingReq := &services.RoutingRequest{
		RequestID:     requestID,
		MerchantID:    merchantID,
		Model:         req.Model,
		Provider:      req.Provider,
		EndpointType:  services.EndpointTypeChatCompletions,
		AllowedKeyIDs: keyFilter,
		RequestBody:   map[string]interface{}{"messages": req.Messages},
	}

	decision, err := pipeline.Execute(context.Background(), routingReq)
	if err != nil || decision == nil || decision.SelectedAPIKeyID == 0 {
		return smartRoutingPick{}
	}

	var candidatesJSON []byte
	if len(decision.DecisionLayerCandidates) > 0 {
		candidatesJSON, _ = json.Marshal(map[string]interface{}{
			"candidates": decision.DecisionLayerCandidates,
		})
	}

	return smartRoutingPick{
		APIKeyID:        &decision.SelectedAPIKeyID,
		CandidatesJSON:  candidatesJSON,
		RoutingDecision: decision,
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

const upstreamErrorBodyPeek = 8192

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
		if err != nil {
			info := services.MapProviderError(0, "", err.Error(), nil, err, "")
			if !info.Retryable || i >= policy.MaxRetries {
				return nil, retryCount, err
			}
			retryCount++
			time.Sleep(policy.DelayForAttempt(i))
			continue
		}

		if resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode < http.StatusInternalServerError {
			return resp, retryCount, nil
		}

		b, _ := io.ReadAll(io.LimitReader(resp.Body, upstreamErrorBodyPeek))
		_ = resp.Body.Close()
		status := resp.StatusCode
		headers := resp.Header
		retryable := services.HTTPUpstreamRetryable(status, b, headers)
		if !retryable || i >= policy.MaxRetries {
			resp.Body = io.NopCloser(bytes.NewReader(b))
			return resp, retryCount, nil
		}
		retryCount++
		time.Sleep(policy.DelayForAttempt(i))
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
