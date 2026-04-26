package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
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

type responseSettlementParams struct {
	C                      *gin.Context
	DB                     *sql.DB
	Req                    APIProxyRequest
	Resp                   *http.Response
	Body                   []byte
	RequestID              string
	UserIDInt              int
	BillProv               string
	BillModel              string
	RequestPath            string
	StartTime              time.Time
	WinningKey             models.MerchantAPIKey
	MerchantID             int
	StrictPricingVID       *int
	SelectedStrategy       string
	SmartCandidatesJSON    []byte
	EffectivePolicySource  string
	DecisionStart          time.Time
	TraceSpan              *services.LLMTraceSpan
	StrategySnapshot       strategyRuntimeSnapshot
	RetryCount             int
	BillingEngine          *billing.BillingEngine
	CurrentRoutingDecision *services.RoutingDecision
	ProviderCfg            providerRuntimeConfig
}

func processResponseAndSettlement(p responseSettlementParams) {
	latency := int(time.Since(p.StartTime).Milliseconds())

	var apiResp APIProxyResponse
	if err := json.Unmarshal(p.Body, &apiResp); err != nil {
		p.TraceSpan.SetStatusCode(p.Resp.StatusCode)
		p.TraceSpan.SetErrorCode("UNMARSHAL_PROXY_RESPONSE_FAILED")
		p.BillingEngine.CancelPreDeduct(p.UserIDInt, p.RequestID)
		p.C.Data(p.Resp.StatusCode, "application/json", p.Body)
		return
	}

	if sm := strings.TrimSpace(apiResp.Model); sm != "" && sm != strings.TrimSpace(p.BillModel) {
		logger.LogInfo(p.C.Request.Context(), "api_proxy", "upstream response model differs from request (e.g. gateway fallback)", map[string]interface{}{
			"request_id": p.RequestID, "requested_model": p.BillModel, "response_model": sm,
		})
	}

	var inputTokens, outputTokens int
	var tokenUsage int64
	var cost float64

	if apiResp.Usage.TotalTokens > 0 {
		inputTokens = apiResp.Usage.PromptTokens
		outputTokens = apiResp.Usage.CompletionTokens
		tokenUsage = p.BillingEngine.CalculateTokenUsage(inputTokens, outputTokens)
		var cerr error
		cost, cerr = calculateTokenCost(p.DB, p.UserIDInt, p.BillProv, p.BillModel, inputTokens, outputTokens, p.StrictPricingVID)
		if cerr != nil {
			p.BillingEngine.CancelPreDeduct(p.UserIDInt, p.RequestID)
			logger.LogError(context.Background(), "api_proxy", "Token cost resolution failed", cerr, map[string]interface{}{
				"user_id": p.UserIDInt, "provider": p.BillProv, "model": p.BillModel, "request_id": p.RequestID,
			})
			middleware.RespondWithError(p.C, apperrors.ErrPricingSnapshotMiss)
			return
		}
	}

	if tokenUsage > 0 {
		settleErr := p.BillingEngine.SettlePreDeduct(p.UserIDInt, p.RequestID, tokenUsage)
		if settleErr != nil {
			logger.LogError(context.Background(), "api_proxy", "Settle pre-deduct failed", settleErr, map[string]interface{}{
				"user_id":     p.UserIDInt,
				"token_usage": tokenUsage,
				"request_id":  p.RequestID,
			})
		}
	} else {
		p.BillingEngine.CancelPreDeduct(p.UserIDInt, p.RequestID)
	}

	if cost > 0 {
		billReq := p.Req
		billReq.Provider = p.BillProv
		billReq.Model = p.BillModel
		logMerchantSKUID, logProcurementCNY := resolveMerchantSKUProcurementForLog(p.DB, billReq, p.WinningKey.ID, p.MerchantID, inputTokens, outputTokens)
		tx, err := p.DB.Begin()
		if err == nil {
			_, updateErr := tx.Exec(
				"UPDATE merchant_api_keys SET quota_used = quota_used + $1, last_used_at = $2 WHERE id = $3",
				cost, time.Now(), p.WinningKey.ID,
			)
			err = updateErr
			if err == nil {
				_, err = tx.Exec(
					"INSERT INTO api_usage_logs (user_id, key_id, request_id, provider, model, method, path, status_code, latency_ms, input_tokens, output_tokens, cost, token_usage, merchant_sku_id, procurement_cost_cny) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)",
					p.UserIDInt, p.WinningKey.ID, p.RequestID, p.BillProv, p.BillModel, "POST", p.RequestPath, p.Resp.StatusCode, latency, inputTokens, outputTokens, cost, tokenUsage, nullInt64Arg(logMerchantSKUID), nullFloat64Arg(logProcurementCNY),
				)
			}

			if err != nil {
				tx.Rollback()
				logger.LogError(context.Background(), "api_proxy", "Transaction rollback", err, map[string]interface{}{
					"user_id":     p.UserIDInt,
					"provider":    p.BillProv,
					"model":       p.BillModel,
					"cost":        cost,
					"token_usage": tokenUsage,
					"request_id":  p.RequestID,
				})
			} else {
				tx.Commit()
				logger.LogInfo(context.Background(), "api_proxy", "API request completed", map[string]interface{}{
					"user_id":       p.UserIDInt,
					"provider":      p.BillProv,
					"model":         p.BillModel,
					"input_tokens":  inputTokens,
					"output_tokens": outputTokens,
					"token_usage":   tokenUsage,
					"cost":          cost,
					"latency_ms":    latency,
					"status_code":   p.Resp.StatusCode,
					"request_id":    p.RequestID,
				})

				metrics.RecordOrderCreation("completed", int64(cost*100), "USD")
			}
		}

		ctx := context.Background()
		cache.Delete(ctx, cache.TokenBalanceKey(p.UserIDInt))
		cache.Delete(ctx, cache.ComputePointBalanceKey(p.UserIDInt))
	}

	if p.CurrentRoutingDecision != nil {
		pipeline := services.NewThreeLayerRoutingPipeline()

		gatewayMode := "direct"
		active := strings.TrimSpace(strings.ToLower(os.Getenv("LLM_GATEWAY_ACTIVE")))
		if active == llmGatewayLitellm {
			gatewayMode = "litellm"
		} else if active == llmGatewayProxy {
			gatewayMode = llmGatewayProxy
		}

		execInput := &services.ExecutionLayerInputData{
			GatewayMode:   gatewayMode,
			EndpointURL:   fmt.Sprintf("%s/chat/completions", strings.TrimRight(p.ProviderCfg.APIBaseURL, "/")),
			AuthMethod:    "bearer",
			ResolvedModel: p.BillModel,
			RequestFormat: p.ProviderCfg.APIFormat,
		}
		pipeline.RecordExecutionInput(p.CurrentRoutingDecision, execInput)

		execSuccess := p.Resp.StatusCode >= 200 && p.Resp.StatusCode < 300
		execLatency := int(time.Since(p.DecisionStart).Milliseconds())
		var execErrMsg string
		if !execSuccess {
			execErrMsg = fmt.Sprintf("HTTP %d", p.Resp.StatusCode)
		}

		execResult := &services.ExecutionLayerResultData{
			Success:      execSuccess,
			StatusCode:   p.Resp.StatusCode,
			LatencyMs:    execLatency,
			ErrorMessage: execErrMsg,
			Model:        p.CurrentRoutingDecision.SelectedModel,
			Provider:     p.CurrentRoutingDecision.SelectedProvider,
			ActualModel:  apiResp.Model,
			InputTokens:  apiResp.Usage.PromptTokens,
			OutputTokens: apiResp.Usage.CompletionTokens,
		}
		if len(apiResp.Choices) > 0 {
			execResult.FinishReason = apiResp.Choices[0].FinishReason
		}
		pipeline.RecordExecutionResultExtended(p.CurrentRoutingDecision, execResult)

		engine := services.NewUnifiedRoutingEngine()
		if logErr := engine.LogDecision(p.C.Request.Context(), p.CurrentRoutingDecision); logErr != nil {
			logger.LogWarn(p.C.Request.Context(), "api_proxy", "Failed to log routing decision", map[string]interface{}{
				"request_id": p.RequestID,
				"error":      logErr.Error(),
			})
		}
	} else {
		decisionPayload := buildRoutingDecisionPayload(p.SmartCandidatesJSON, p.StrategySnapshot, p.EffectivePolicySource)
		_ = insertRoutingDecision(p.DB, p.RequestID, p.UserIDInt, p.Req, p.SelectedStrategy, decisionPayload, p.WinningKey.ID, int(time.Since(p.DecisionStart).Milliseconds()), p.RetryCount)
	}

	p.TraceSpan.SetStatusCode(p.Resp.StatusCode)
	p.C.Data(p.Resp.StatusCode, "application/json", p.Body)
}
