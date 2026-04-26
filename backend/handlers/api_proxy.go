package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/logger"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/models"
	"github.com/pintuotuo/backend/services"
	"github.com/pintuotuo/backend/utils"
)

type providerRuntimeConfig struct {
	Code       string
	Name       string
	APIBaseURL string
	APIFormat  string
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

	userIDInt, err := authenticateUser(c)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	req, err := parseAPIProxyRequest(c)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	proxyAPIRequestCore(c, userIDInt, requestID, startTime, *req, c.Request.URL.Path)
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

	prepareResult, err := validateAndPrepareRequest(c, db, userIDInt, req, requestID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			middleware.RespondWithError(c, appErr)
		} else {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		}
		return
	}

	strictPricingVID := prepareResult.StrictPricingVID
	merchantID := prepareResult.MerchantID
	entCtx := prepareResult.EntCtx
	billingEngine := prepareResult.BillingEngine

	routingResult := resolveRoutingSelection(req, userIDInt, requestID, merchantID, entCtx)
	if routingResult.APIKeyID != nil {
		req.APIKeyID = routingResult.APIKeyID
	}
	if routingResult.MerchantSKUID != nil {
		req.MerchantSKUID = routingResult.MerchantSKUID
	}
	selectedStrategy := routingResult.SelectedStrategy
	effectivePolicySource := routingResult.EffectivePolicySource
	smartCandidatesJSON := routingResult.SmartCandidatesJSON
	currentRoutingDecision := routingResult.CurrentRoutingDecision
	decisionStart := time.Now()

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
		retryPolicyStream = applyLitellmGatewayRetryCap(retryPolicyStream)
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
			authToken := resolveGatewayAuthToken(pcfg, dk)
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
	retryPolicy = applyLitellmGatewayRetryCap(retryPolicy)
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
			authToken := resolveGatewayAuthToken(pcfg, dk)
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

	processResponseAndSettlement(responseSettlementParams{
		C:                      c,
		DB:                     db,
		Req:                    req,
		Resp:                   resp,
		Body:                   body,
		RequestID:              requestID,
		UserIDInt:              userIDInt,
		BillProv:               billProv,
		BillModel:              billModel,
		RequestPath:            requestPath,
		StartTime:              startTime,
		WinningKey:             winningKey,
		MerchantID:             merchantID,
		StrictPricingVID:       strictPricingVID,
		SelectedStrategy:       selectedStrategy,
		SmartCandidatesJSON:    smartCandidatesJSON,
		EffectivePolicySource:  effectivePolicySource,
		DecisionStart:          decisionStart,
		TraceSpan:              traceSpan,
		StrategySnapshot:       strategySnapshot,
		RetryCount:             retryCountTotal,
		BillingEngine:          billingEngine,
		CurrentRoutingDecision: currentRoutingDecision,
		ProviderCfg:            providerCfg,
	})
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
