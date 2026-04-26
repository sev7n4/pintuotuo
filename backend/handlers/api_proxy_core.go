package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/billing"
	"github.com/pintuotuo/backend/cache"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/logger"
	"github.com/pintuotuo/backend/metrics"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/models"
	"github.com/pintuotuo/backend/services"
)

type requestPrepareResult struct {
	StrictPricingVID    *int
	TokenBalance        float64
	EstimatedUsage      int64
	PreDeductConfig     *billing.PreDeductConfig
	InputTokensEstimate int
	MerchantID          int
	EntCtx              *services.EntitlementRoutingContext
	BillingEngine       *billing.BillingEngine
}

func validateAndPrepareRequest(c *gin.Context, db *sql.DB, userIDInt int, req APIProxyRequest, requestID string) (*requestPrepareResult, error) {
	pricingResult, err := resolveStrictPricingVersion(db, userIDInt, req.Provider, req.Model)
	if err != nil {
		return nil, apperrors.ErrDatabaseError
	}
	if !pricingResult.Found {
		return nil, apperrors.ErrEntitlementDenied
	}

	tokenBalance, err := getTokenBalance(db, userIDInt)
	if err != nil {
		return nil, apperrors.ErrTokenNotFound
	}

	if tokenBalance <= 0 {
		return nil, apperrors.ErrInsufficientBalance
	}

	usageResult, _ := estimateTokenUsage(req.Messages, req.Provider)

	if !hasSufficientBalance(tokenBalance, usageResult.EstimatedUsage) {
		logger.LogWarn(c.Request.Context(), "api_proxy", "Insufficient balance for pre-deduction", map[string]interface{}{
			"user_id":         userIDInt,
			"balance":         tokenBalance,
			"estimated_usage": usageResult.EstimatedUsage,
			"input_estimate":  usageResult.InputTokensEstimate,
			"multiplier":      usageResult.PreDeductConfig.Multiplier,
			"request_id":      requestID,
		})
		return nil, apperrors.ErrInsufficientBalance
	}

	billingEngine := billing.GetBillingEngine()
	preDeductErr := billingEngine.PreDeductBalance(userIDInt, usageResult.EstimatedUsage, "API call pre-deduct", requestID)
	if preDeductErr != nil {
		logger.LogError(c.Request.Context(), "api_proxy", "Pre-deduction failed", preDeductErr, map[string]interface{}{
			"user_id":         userIDInt,
			"estimated_usage": usageResult.EstimatedUsage,
			"request_id":      requestID,
		})
		if strings.Contains(preDeductErr.Error(), "insufficient balance") {
			return nil, apperrors.ErrInsufficientBalance
		}
		return nil, apperrors.ErrDatabaseError
	}

	logger.LogInfo(c.Request.Context(), "api_proxy", "Pre-deduction successful", map[string]interface{}{
		"user_id":         userIDInt,
		"estimated_usage": usageResult.EstimatedUsage,
		"request_id":      requestID,
	})

	merchantID, merchantErr := resolveMerchantIDByUser(db, userIDInt)
	if merchantErr != nil {
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		return nil, apperrors.ErrDatabaseError
	}

	var entCtx *services.EntitlementRoutingContext
	if services.EntitlementEnforcementStrict() {
		entCtx, err = services.ResolveEntitlementRoutingContext(db, userIDInt, req.Provider, req.Model)
		if err != nil {
			billingEngine.CancelPreDeduct(userIDInt, requestID)
			return nil, apperrors.ErrDatabaseError
		}
		if req.APIKeyID != nil && *req.APIKeyID > 0 && !entCtx.AllowsAPIKey(*req.APIKeyID) {
			billingEngine.CancelPreDeduct(userIDInt, requestID)
			return nil, apperrors.NewAppError(
				"API_KEY_NOT_AUTHORIZED",
				"api_key_id is not allowed for your entitlement",
				http.StatusForbidden,
				nil,
			)
		}
		if req.MerchantSKUID != nil && *req.MerchantSKUID > 0 && !entCtx.AllowsMerchantSKU(*req.MerchantSKUID) {
			billingEngine.CancelPreDeduct(userIDInt, requestID)
			return nil, apperrors.NewAppError(
				"API_KEY_NOT_AUTHORIZED",
				"merchant_sku_id is not allowed for your entitlement",
				http.StatusForbidden,
				nil,
			)
		}
	}

	return &requestPrepareResult{
		StrictPricingVID:    pricingResult.PricingVID,
		TokenBalance:        tokenBalance,
		EstimatedUsage:      usageResult.EstimatedUsage,
		PreDeductConfig:     usageResult.PreDeductConfig,
		InputTokensEstimate: usageResult.InputTokensEstimate,
		MerchantID:          merchantID,
		EntCtx:              entCtx,
		BillingEngine:       billingEngine,
	}, nil
}

type routingSelectionResult struct {
	APIKeyID               *int
	MerchantSKUID          *int
	SelectedStrategy       string
	EffectivePolicySource  string
	SmartCandidatesJSON    []byte
	CurrentRoutingDecision *services.RoutingDecision
}

func resolveRoutingSelection(req APIProxyRequest, userIDInt int, requestID string, merchantID int, entCtx *services.EntitlementRoutingContext) routingSelectionResult {
	result := routingSelectionResult{
		SelectedStrategy: "legacy_fallback",
	}

	keyFilter := entitlementKeyFilterForRouter(services.EntitlementEnforcementStrict(), entCtx)

	if req.APIKeyID == nil && req.MerchantSKUID == nil && shouldUseSmartRouting(userIDInt, requestID) {
		strategyCode, policySrc := routingStrategyWithSource()
		if smartReq := trySelectAPIKeyWithSmartRouter(req, strategyCode, keyFilter, requestID, merchantID); smartReq.APIKeyID != nil {
			result.APIKeyID = smartReq.APIKeyID
			if entCtx != nil {
				if msid, ok := entCtx.MerchantSKUForAPIKey(*result.APIKeyID); ok && req.MerchantSKUID == nil {
					result.MerchantSKUID = &msid
				}
			}
			result.SelectedStrategy = strategyCode
			result.SmartCandidatesJSON = smartReq.CandidatesJSON
			result.EffectivePolicySource = policySrc
			result.CurrentRoutingDecision = smartReq.RoutingDecision
			return result
		}
	}

	if req.APIKeyID == nil && req.MerchantSKUID == nil && services.EntitlementEnforcementStrict() && entCtx != nil && len(entCtx.AllowedAPIKeyIDs) > 0 {
		if pick, msid := pickDeterministicEntitledKey(entCtx); pick > 0 {
			result.APIKeyID = &pick
			if req.MerchantSKUID == nil && msid > 0 {
				result.MerchantSKUID = &msid
			}
		}
	}

	return result
}

func handleResponseAndSettlement(
	c *gin.Context,
	db *sql.DB,
	req APIProxyRequest,
	resp *http.Response,
	body []byte,
	requestID string,
	userIDInt int,
	billProv string,
	billModel string,
	requestPath string,
	startTime time.Time,
	winningKey models.MerchantAPIKey,
	merchantID int,
	strictPricingVID *int,
	selectedStrategy string,
	smartCandidatesJSON []byte,
	effectivePolicySource string,
	decisionStart time.Time,
	traceSpan *services.LLMTraceSpan,
	strategySnapshot strategyRuntimeSnapshot,
	retryCount int,
	billingEngine *billing.BillingEngine,
	currentRoutingDecision *services.RoutingDecision,
	providerCfg providerRuntimeConfig,
) {
	latency := int(time.Since(startTime).Milliseconds())

	var inputTokens, outputTokens int
	var tokenUsage int64
	var cost float64

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

		gatewayMode := "direct"
		active := strings.TrimSpace(strings.ToLower(os.Getenv("LLM_GATEWAY_ACTIVE")))
		if active == llmGatewayLitellm {
			gatewayMode = "litellm"
		} else if active == "proxy" {
			gatewayMode = "proxy"
		}

		execInput := &services.ExecutionLayerInputData{
			GatewayMode:   gatewayMode,
			EndpointURL:   fmt.Sprintf("%s/chat/completions", strings.TrimRight(providerCfg.APIBaseURL, "/")),
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

func handlePrepareError(c *gin.Context, err error, userIDInt int, requestID string, billingEngine *billing.BillingEngine) bool {
	if err == nil {
		return false
	}
	if billingEngine != nil {
		billingEngine.CancelPreDeduct(userIDInt, requestID)
	}
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		middleware.RespondWithError(c, appErr)
	} else {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
	}
	return true
}
