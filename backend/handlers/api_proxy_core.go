package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/billing"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/logger"
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
