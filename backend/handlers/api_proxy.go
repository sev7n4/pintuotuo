package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pintuotuo/backend/billing"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/logger"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/models"
	"github.com/pintuotuo/backend/services"
	"github.com/pintuotuo/backend/utils"
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

//nolint:unused // Will be used in Phase 3 Part 2
func shouldUseExecutionLayer() bool {
	val := os.Getenv("USE_EXECUTION_LAYER")
	return val == envTrue || val == "1"
}

//nolint:unused // Will be used in Phase 3 Part 2
func shouldUseConfigDrivenRouting() bool {
	val := os.Getenv("USE_CONFIG_DRIVEN_ROUTING")
	return val == envTrue || val == "1"
}

//nolint:unused // Will be fully implemented in Phase 4
func executeViaExecutionLayer(
	c *gin.Context,
	db *sql.DB,
	userIDInt int,
	merchantID int,
	req APIProxyRequest,
	providerCfg providerRuntimeConfig,
	decryptedKey string,
	requestID string,
	traceSpan *services.LLMTraceSpan,
) (*services.ExecutionLayerOutput, error) {
	engine := services.NewExecutionEngine()
	layer := services.NewExecutionLayer(db, engine)

	routeDecision, err := resolveRouteDecision(db, &providerCfg, merchantID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve route decision: %w", err)
	}

	execProviderCfg := &services.ExecutionProviderConfig{
		Code:           providerCfg.Code,
		Name:           providerCfg.Name,
		APIBaseURL:     providerCfg.APIBaseURL,
		APIFormat:      providerCfg.APIFormat,
		GatewayMode:    routeDecision.Mode,
		ProviderRegion: providerCfg.ProviderRegion,
		RouteStrategy:  providerCfg.RouteStrategy,
		Endpoints:      providerCfg.Endpoints,
	}

	messages := make([]services.Message, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = services.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	routingDecision := &services.RoutingDecision{
		RequestID:        requestID,
		MerchantID:       merchantID,
		Model:            req.Model,
		Provider:         req.Provider,
		SelectedProvider: providerCfg.Code,
		SelectedModel:    req.Model,
		RoutingMode:      routeDecision.Mode,
		Timestamp:        time.Now(),
	}

	input := &services.ExecutionLayerInput{
		ProviderConfig:  execProviderCfg,
		DecryptedAPIKey: decryptedKey,
		Messages:        messages,
		Stream:          req.Stream,
		Options:         req.Options,
		RoutingDecision: routingDecision,
	}

	return layer.Execute(c.Request.Context(), input)
}

func processExecutionLayerOutput(
	c *gin.Context,
	db *sql.DB,
	req APIProxyRequest,
	output *services.ExecutionLayerOutput,
	requestID string,
	userIDInt int,
	merchantID int,
	startTime time.Time,
	billingEngine *billing.BillingEngine,
	traceSpan *services.LLMTraceSpan,
) {
	result := output.Result
	decision := output.Decision

	traceSpan.SetStatusCode(result.StatusCode)

	var promptTokens, completionTokens int
	if result.Usage != nil {
		promptTokens = result.Usage.PromptTokens
		completionTokens = result.Usage.CompletionTokens
	}

	billProv := result.Provider
	if billProv == "" {
		billProv = req.Provider
	}
	billModel := result.ActualModel
	if billModel == "" {
		billModel = req.Model
	}

	tokenUsage := int64(promptTokens + completionTokens)
	if tokenUsage > 0 {
		if settleErr := billingEngine.SettlePreDeduct(userIDInt, requestID, tokenUsage); settleErr != nil {
			logger.LogError(c.Request.Context(), "api_proxy", "Settle pre-deduct failed", settleErr, map[string]interface{}{
				"user_id":     userIDInt,
				"token_usage": tokenUsage,
				"request_id":  requestID,
			})
		}
	} else {
		billingEngine.CancelPreDeduct(userIDInt, requestID)
	}

	if decision != nil {
		decision.ExecutionSuccess = result.Success
		decision.ExecutionStatusCode = result.StatusCode
		decision.ExecutionLatencyMs = result.LatencyMs
		decision.SelectedProvider = billProv
		decision.SelectedModel = billModel
		if result.Usage != nil {
			decision.EstimatedInputTokens = float64(promptTokens)
			decision.EstimatedOutputTokens = float64(completionTokens)
		}
		decision.DecisionResult = string(services.DecisionResultSuccess)
		decision.DecisionDurationMs = int(time.Since(startTime).Milliseconds())
	}

	services.RouteDecisionCounter.WithLabelValues(
		billProv,
		billModel,
		"execution_layer",
		"success",
	).Inc()

	services.ExecutionLayerLatency.WithLabelValues(
		billProv,
		billModel,
	).Observe(float64(result.LatencyMs))

	c.Data(result.StatusCode, "application/json", result.ResponseBody)
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

	if shouldUseExecutionLayer() {
		output, err := executeViaExecutionLayer(
			c, db, userIDInt, merchantID, req, providerCfg,
			decryptedKey, requestID, traceSpan,
		)
		if err != nil {
			logger.LogError(c.Request.Context(), "api_proxy", "ExecutionLayer failed", err, map[string]interface{}{
				"request_id": requestID,
				"provider":   req.Provider,
				"model":      req.Model,
			})
			billingEngine.CancelPreDeduct(userIDInt, requestID)
			traceSpan.SetStatusCode(http.StatusBadGateway)
			traceSpan.SetErrorCode("EXECUTION_LAYER_FAILED")
			middleware.RespondWithError(c, apperrors.NewAppError(
				"EXECUTION_LAYER_FAILED",
				fmt.Sprintf("Execution layer error: %v", err),
				http.StatusBadGateway,
				err,
			))
			return
		}

		if output.Result == nil {
			billingEngine.CancelPreDeduct(userIDInt, requestID)
			traceSpan.SetStatusCode(http.StatusInternalServerError)
			traceSpan.SetErrorCode("EXECUTION_LAYER_NO_RESULT")
			middleware.RespondWithError(c, apperrors.NewAppError(
				"EXECUTION_LAYER_NO_RESULT",
				"Execution layer returned no result",
				http.StatusInternalServerError,
				nil,
			))
			return
		}

		if !output.Result.Success {
			logger.LogWarn(c.Request.Context(), "api_proxy", "ExecutionLayer request failed", map[string]interface{}{
				"request_id":    requestID,
				"status_code":   output.Result.StatusCode,
				"error_message": output.Result.ErrorMessage,
			})
			billingEngine.CancelPreDeduct(userIDInt, requestID)
			traceSpan.SetStatusCode(output.Result.StatusCode)
			c.Data(output.Result.StatusCode, "application/json", output.Result.ResponseBody)
			return
		}

		processExecutionLayerOutput(c, db, req, output, requestID, userIDInt, merchantID,
			startTime, billingEngine, traceSpan)
		return
	}

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
