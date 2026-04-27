package handlers

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/billing"
	"github.com/pintuotuo/backend/services"
)

var (
	ErrUserNotAuthenticated = errors.New("user not authenticated")
	ErrInvalidUserIDType    = errors.New("invalid user_id type")
)

func authenticateUser(c *gin.Context) (int, error) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		return 0, ErrUserNotAuthenticated
	}

	var userID int
	switch v := userIDVal.(type) {
	case int:
		userID = v
	case int64:
		userID = int(v)
	case float64:
		userID = int(v)
	default:
		return 0, fmt.Errorf("%w: got %T", ErrInvalidUserIDType, userIDVal)
	}

	return userID, nil
}

func parseAPIProxyRequest(c *gin.Context) (*APIProxyRequest, error) {
	var req APIProxyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, fmt.Errorf("invalid request body: %w", err)
	}

	if req.Model == "" {
		return nil, errors.New("model is required")
	}

	return &req, nil
}

func getTokenBalance(db *sql.DB, userID int) (float64, error) {
	if db == nil {
		return 0, errors.New("database connection is nil")
	}

	var balance float64
	err := db.QueryRow("SELECT balance FROM tokens WHERE user_id = $1", userID).Scan(&balance)
	if err != nil {
		return 0, fmt.Errorf("failed to get token balance: %w", err)
	}

	return balance, nil
}

type StrictPricingResult struct {
	PricingVID *int
	Found      bool
}

func resolveStrictPricingVersion(db *sql.DB, userID int, provider, model string) (*StrictPricingResult, error) {
	if !services.EntitlementEnforcementStrict() {
		return &StrictPricingResult{PricingVID: nil, Found: true}, nil
	}

	vid, _, ok, entErr := services.ResolveChosenPricingVersion(db, userID, provider, model)
	if entErr != nil {
		return nil, fmt.Errorf("failed to resolve pricing version: %w", entErr)
	}
	if !ok {
		return &StrictPricingResult{PricingVID: nil, Found: false}, nil
	}

	return &StrictPricingResult{PricingVID: &vid, Found: true}, nil
}

type EstimatedUsageResult struct {
	InputTokensEstimate int
	EstimatedUsage      int64
	PreDeductConfig     *billing.PreDeductConfig
}

func estimateTokenUsage(messages []ChatMessage, provider string) (*EstimatedUsageResult, error) {
	inputTokensEstimate := estimateInputTokens(messages)
	billingEngine := billing.GetBillingEngine()
	preDeductConfig := billingEngine.GetPreDeductConfig(0, 0, provider)
	estimatedUsage := billingEngine.EstimateTokenUsage(inputTokensEstimate, preDeductConfig)

	return &EstimatedUsageResult{
		InputTokensEstimate: inputTokensEstimate,
		EstimatedUsage:      estimatedUsage,
		PreDeductConfig:     preDeductConfig,
	}, nil
}

func hasSufficientBalance(balance float64, estimatedUsage int64) bool {
	return balance >= float64(estimatedUsage)
}
