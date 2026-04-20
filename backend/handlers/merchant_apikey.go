package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/cache"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/logger"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/models"
	"github.com/pintuotuo/backend/services"
	"github.com/pintuotuo/backend/utils"
)

func CreateMerchantAPIKey(c *gin.Context) {
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

	var req struct {
		Name       string   `json:"name" binding:"required"`
		Provider   string   `json:"provider" binding:"required"`
		APIKey     string   `json:"api_key" binding:"required"`
		APISecret  string   `json:"api_secret"`
		QuotaLimit *float64 `json:"quota_limit"`
	}

	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	merchantID, ok := gateMerchantOperational(c, db, userIDInt)
	if !ok {
		return
	}

	apiKeyEncrypted, err := utils.Encrypt(req.APIKey)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"ENCRYPTION_FAILED",
			"Failed to encrypt API key",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	var apiSecretEncrypted string
	if req.APISecret != "" {
		apiSecretEncrypted, err = utils.Encrypt(req.APISecret)
		if err != nil {
			middleware.RespondWithError(c, apperrors.NewAppError(
				"ENCRYPTION_FAILED",
				"Failed to encrypt API secret",
				http.StatusInternalServerError,
				err,
			))
			return
		}
	}

	var quota any
	if req.QuotaLimit != nil && *req.QuotaLimit > 0 {
		quota = *req.QuotaLimit
	} else {
		quota = nil
	}

	var apiKey models.MerchantAPIKey
	var quotaReturned sql.NullFloat64
	err = db.QueryRow(
		`INSERT INTO merchant_api_keys (merchant_id, name, provider, api_key_encrypted, api_secret_encrypted, quota_limit, quota_used, status) 
		 VALUES ($1, $2, $3, $4, $5, $6, 0, 'active') 
		 RETURNING id, merchant_id, name, provider, quota_limit, quota_used, status, created_at, updated_at`,
		merchantID, req.Name, req.Provider, apiKeyEncrypted, apiSecretEncrypted, quota,
	).Scan(&apiKey.ID, &apiKey.MerchantID, &apiKey.Name, &apiKey.Provider, &quotaReturned, &apiKey.QuotaUsed, &apiKey.Status, &apiKey.CreatedAt, &apiKey.UpdatedAt)
	apiKey.QuotaLimit = utils.NullFloat64Ptr(quotaReturned)

	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"API_KEY_CREATE_FAILED",
			"Failed to create API key",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	ctx := context.Background()
	cache.Delete(ctx, cache.MerchantAPIKeysKey(merchantID))

	c.JSON(http.StatusCreated, apiKey)
}

func ListMerchantAPIKeys(c *gin.Context) {
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

	var merchantID int
	err := db.QueryRow("SELECT id FROM merchants WHERE user_id = $1", userIDInt).Scan(&merchantID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"MERCHANT_NOT_FOUND",
			"Merchant not found",
			http.StatusNotFound,
			err,
		))
		return
	}

	ctx := context.Background()
	cacheKey := cache.MerchantAPIKeysKey(merchantID)

	if cachedKeys, cacheErr := cache.Get(ctx, cacheKey); cacheErr == nil {
		var apiKeys []models.MerchantAPIKey
		if unmarshalErr := json.Unmarshal([]byte(cachedKeys), &apiKeys); unmarshalErr == nil {
			c.JSON(http.StatusOK, gin.H{"data": apiKeys})
			return
		}
	}

	rows, err := db.Query(
		`SELECT id, merchant_id, name, provider, quota_limit, quota_used, status, last_used_at, created_at, updated_at 
		 FROM merchant_api_keys WHERE merchant_id = $1 ORDER BY created_at DESC`,
		merchantID,
	)

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	var apiKeys []models.MerchantAPIKey
	for rows.Next() {
		var key models.MerchantAPIKey
		var lastUsedAt sql.NullTime
		var qLim sql.NullFloat64
		err := rows.Scan(&key.ID, &key.MerchantID, &key.Name, &key.Provider, &qLim, &key.QuotaUsed, &key.Status, &lastUsedAt, &key.CreatedAt, &key.UpdatedAt)
		key.QuotaLimit = utils.NullFloat64Ptr(qLim)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		if lastUsedAt.Valid {
			key.LastUsedAt = lastUsedAt.Time
		}
		apiKeys = append(apiKeys, key)
	}

	if apiKeysJSON, err := json.Marshal(apiKeys); err == nil {
		cache.Set(ctx, cacheKey, string(apiKeysJSON), cache.MerchantCacheTTL)
	}

	c.JSON(http.StatusOK, gin.H{"data": apiKeys})
}

func UpdateMerchantAPIKey(c *gin.Context) {
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

	keyIDStr := c.Param("id")
	keyID, err := strconv.Atoi(keyIDStr)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	bodyBytes, readErr := io.ReadAll(c.Request.Body)
	if readErr != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}
	var patch map[string]json.RawMessage
	if patchErr := json.Unmarshal(bodyBytes, &patch); patchErr != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	var name, status string
	if raw, has := patch["name"]; has {
		_ = json.Unmarshal(raw, &name)
	}
	if raw, has := patch["status"]; has {
		_ = json.Unmarshal(raw, &status)
	}

	patchQuota := false
	unlimitedQuota := false
	quotaVal := 0.0
	if raw, has := patch["quota_limit"]; has {
		patchQuota = true
		if strings.TrimSpace(string(raw)) == "null" {
			unlimitedQuota = true
		} else {
			if qErr := json.Unmarshal(raw, &quotaVal); qErr != nil {
				middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
				return
			}
			if quotaVal <= 0 {
				unlimitedQuota = true
			}
		}
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	merchantID, ok := gateMerchantOperational(c, db, userIDInt)
	if !ok {
		return
	}

	var apiKey models.MerchantAPIKey
	var lastUsedAt sql.NullTime
	var quotaAfter sql.NullFloat64
	err = db.QueryRow(
		`UPDATE merchant_api_keys SET 
		 name = COALESCE(NULLIF($1, ''), name),
		 status = COALESCE(NULLIF($2, ''), status),
		 quota_limit = CASE
		   WHEN NOT $3::bool THEN quota_limit
		   WHEN $4::bool THEN NULL
		   ELSE $5::numeric
		 END,
		 updated_at = CURRENT_TIMESTAMP
		 WHERE id = $6 AND merchant_id = $7
		 RETURNING id, merchant_id, name, provider, quota_limit, quota_used, status, last_used_at, created_at, updated_at`,
		name, status, patchQuota, unlimitedQuota, quotaVal, keyID, merchantID,
	).Scan(&apiKey.ID, &apiKey.MerchantID, &apiKey.Name, &apiKey.Provider, &quotaAfter, &apiKey.QuotaUsed, &apiKey.Status, &lastUsedAt, &apiKey.CreatedAt, &apiKey.UpdatedAt)
	apiKey.QuotaLimit = utils.NullFloat64Ptr(quotaAfter)

	if err == sql.ErrNoRows {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"API_KEY_NOT_FOUND",
			"API key not found",
			http.StatusNotFound,
			err,
		))
		return
	} else if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	if lastUsedAt.Valid {
		apiKey.LastUsedAt = lastUsedAt.Time
	}

	ctx := context.Background()
	cache.Delete(ctx, cache.MerchantAPIKeysKey(merchantID))

	c.JSON(http.StatusOK, apiKey)
}

func DeleteMerchantAPIKey(c *gin.Context) {
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

	keyIDStr := c.Param("id")
	keyID, err := strconv.Atoi(keyIDStr)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	merchantID, ok := gateMerchantOperational(c, db, userIDInt)
	if !ok {
		return
	}

	result, err := db.Exec("DELETE FROM merchant_api_keys WHERE id = $1 AND merchant_id = $2", keyID, merchantID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"API_KEY_NOT_FOUND",
			"API key not found",
			http.StatusNotFound,
			nil,
		))
		return
	}

	ctx := context.Background()
	cache.Delete(ctx, cache.MerchantAPIKeysKey(merchantID))

	c.JSON(http.StatusOK, gin.H{"message": "API key deleted successfully"})
}

func GetMerchantAPIKeyUsage(c *gin.Context) {
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

	var merchantID int
	err := db.QueryRow("SELECT id FROM merchants WHERE user_id = $1", userIDInt).Scan(&merchantID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"MERCHANT_NOT_FOUND",
			"Merchant not found",
			http.StatusNotFound,
			err,
		))
		return
	}

	rows, err := db.Query(
		`SELECT id, name, provider, quota_limit, quota_used, 
		 CASE WHEN quota_limit IS NOT NULL AND quota_limit > 0 THEN (quota_used / quota_limit * 100) ELSE 0 END as usage_percentage
		 FROM merchant_api_keys WHERE merchant_id = $1 AND status = 'active'`,
		merchantID,
	)

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	type UsageInfo struct {
		ID              int      `json:"id"`
		Name            string   `json:"name"`
		Provider        string   `json:"provider"`
		QuotaLimit      *float64 `json:"quota_limit"`
		QuotaUsed       float64  `json:"quota_used"`
		UsagePercentage float64  `json:"usage_percentage"`
	}

	var usageList []UsageInfo
	for rows.Next() {
		var usage UsageInfo
		var qLim sql.NullFloat64
		err := rows.Scan(&usage.ID, &usage.Name, &usage.Provider, &qLim, &usage.QuotaUsed, &usage.UsagePercentage)
		usage.QuotaLimit = utils.NullFloat64Ptr(qLim)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		usageList = append(usageList, usage)
	}

	c.JSON(http.StatusOK, gin.H{"data": usageList})
}

func RequestSettlement(c *gin.Context) {
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

	merchantID, ok := gateMerchantOperational(c, db, userIDInt)
	if !ok {
		return
	}

	var pendingSettlement int
	db.QueryRow("SELECT COUNT(*) FROM merchant_settlements WHERE merchant_id = $1 AND status IN ('pending', 'processing')", merchantID).Scan(&pendingSettlement)
	if pendingSettlement > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "已有待处理的结算申请"})
		return
	}

	var totalSales float64
	db.QueryRow(
		`SELECT COALESCE(SUM(o.total_price), 0) FROM orders o 
		 JOIN merchant_skus ms ON ms.sku_id = o.sku_id AND ms.merchant_id = $1
		 WHERE o.status = 'completed' AND o.updated_at > 
		 COALESCE((SELECT MAX(period_end) FROM merchant_settlements WHERE merchant_id = $1), '1970-01-01')`,
		merchantID,
	).Scan(&totalSales)

	if totalSales < 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "可结算金额不足100元"})
		return
	}

	platformFeeRate := 0.05
	platformFee := totalSales * platformFeeRate
	settlementAmount := totalSales - platformFee

	now := time.Now()
	var settlement models.MerchantSettlement
	err := db.QueryRow(
		`INSERT INTO merchant_settlements (merchant_id, period_start, period_end, total_sales, platform_fee, settlement_amount, status) 
		 VALUES ($1, $2, $3, $4, $5, $6, 'pending') 
		 RETURNING id, merchant_id, period_start, period_end, total_sales, platform_fee, settlement_amount, status, created_at, updated_at`,
		merchantID, now.AddDate(0, 0, -30), now, totalSales, platformFee, settlementAmount,
	).Scan(&settlement.ID, &settlement.MerchantID, &settlement.PeriodStart, &settlement.PeriodEnd, &settlement.TotalSales, &settlement.PlatformFee, &settlement.SettlementAmount, &settlement.Status, &settlement.CreatedAt, &settlement.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"SETTLEMENT_CREATE_FAILED",
			"Failed to create settlement request",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	c.JSON(http.StatusCreated, settlement)
}

func GetSettlementDetail(c *gin.Context) {
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

	settlementIDStr := c.Param("id")
	settlementID, err := strconv.Atoi(settlementIDStr)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var merchantID int
	err = db.QueryRow("SELECT id FROM merchants WHERE user_id = $1", userIDInt).Scan(&merchantID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"MERCHANT_NOT_FOUND",
			"Merchant not found",
			http.StatusNotFound,
			err,
		))
		return
	}

	var settlement models.MerchantSettlement
	var settledAt sql.NullTime
	err = db.QueryRow(
		`SELECT id, merchant_id, period_start, period_end, total_sales, platform_fee, settlement_amount, status, settled_at, created_at, updated_at 
		 FROM merchant_settlements WHERE id = $1 AND merchant_id = $2`,
		settlementID, merchantID,
	).Scan(&settlement.ID, &settlement.MerchantID, &settlement.PeriodStart, &settlement.PeriodEnd, &settlement.TotalSales, &settlement.PlatformFee, &settlement.SettlementAmount, &settlement.Status, &settledAt, &settlement.CreatedAt, &settlement.UpdatedAt)

	if err == sql.ErrNoRows {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"SETTLEMENT_NOT_FOUND",
			"Settlement not found",
			http.StatusNotFound,
			err,
		))
		return
	} else if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	if settledAt.Valid {
		settlement.SettledAt = &settledAt.Time
	}

	c.JSON(http.StatusOK, settlement)
}

func VerifyMerchantAPIKey(c *gin.Context) {
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

	keyIDStr := c.Param("id")
	keyID, err := strconv.Atoi(keyIDStr)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	merchantID, ok := gateMerchantOperational(c, db, userIDInt)
	if !ok {
		return
	}

	var apiKey models.MerchantAPIKey
	err = db.QueryRow(
		`SELECT id, merchant_id, provider, api_key_encrypted
		 FROM merchant_api_keys
		 WHERE id = $1 AND merchant_id = $2 AND status = 'active'`,
		keyID, merchantID,
	).Scan(&apiKey.ID, &apiKey.MerchantID, &apiKey.Provider, &apiKey.APIKeyEncrypted)

	if err == sql.ErrNoRows {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"API_KEY_NOT_FOUND",
			"API key not found",
			http.StatusNotFound,
			err,
		))
		return
	} else if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	validator := services.GetAPIKeyValidator()
	var req struct {
		VerificationMode string `json:"verification_mode"` // light(default) / deep
	}
	_ = c.ShouldBindJSON(&req)
	verificationType := "manual"
	if req.VerificationMode == "deep" {
		verificationType = "manual_deep"
	}

	err = validator.ValidateAsync(apiKey.ID, apiKey.Provider, apiKey.APIKeyEncrypted, verificationType)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"VERIFICATION_FAILED",
			"Failed to start verification",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":            "Verification started",
		"api_key_id":         apiKey.ID,
		"verification_mode":  req.VerificationMode,
		"verification_type":  verificationType,
		"quota_probe_policy": "light(default), optional deep probe for supported providers",
	})
}

// TriggerMerchantAPIKeyHealthCheck forces one immediate health probe for merchant-owned key.
func TriggerMerchantAPIKeyHealthCheck(c *gin.Context) {
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

	keyIDStr := c.Param("id")
	keyID, err := strconv.Atoi(keyIDStr)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	merchantID, ok := gateMerchantOperational(c, db, userIDInt)
	if !ok {
		return
	}

	var existsKey bool
	err = db.QueryRow(
		`SELECT EXISTS(
			SELECT 1 FROM merchant_api_keys
			WHERE id = $1 AND merchant_id = $2 AND status = 'active'
		)`,
		keyID, merchantID,
	).Scan(&existsKey)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	if !existsKey {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"API_KEY_NOT_FOUND",
			"API key not found",
			http.StatusNotFound,
			nil,
		))
		return
	}

	go func() {
		_ = services.GetHealthScheduler().TriggerImmediateCheck(keyID)
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"message":    "Immediate health check triggered",
		"api_key_id": keyID,
	})
}

func GetMerchantAPIKeyVerification(c *gin.Context) {
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

	keyIDStr := c.Param("id")
	keyID, err := strconv.Atoi(keyIDStr)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	merchantID, ok := gateMerchantOperational(c, db, userIDInt)
	if !ok {
		return
	}

	var apiKey models.MerchantAPIKey
	var verificationResult, verificationMsg sql.NullString
	var verifiedAt sql.NullTime
	var modelsJSON []byte
	err = db.QueryRow(
		`SELECT id, merchant_id, provider, verification_result, verified_at, models_supported, verification_message
		 FROM merchant_api_keys
		 WHERE id = $1 AND merchant_id = $2`,
		keyID, merchantID,
	).Scan(&apiKey.ID, &apiKey.MerchantID, &apiKey.Provider, &verificationResult, &verifiedAt, &modelsJSON, &verificationMsg)

	if err == sql.ErrNoRows {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"API_KEY_NOT_FOUND",
			"API key not found",
			http.StatusNotFound,
			err,
		))
		return
	} else if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	if verificationResult.Valid {
		apiKey.VerificationResult = verificationResult.String
	}
	if verificationMsg.Valid {
		apiKey.VerificationMsg = verificationMsg.String
	}
	if verifiedAt.Valid {
		t := verifiedAt.Time
		apiKey.VerifiedAt = &t
	}
	if len(modelsJSON) > 0 {
		_ = json.Unmarshal(modelsJSON, &apiKey.ModelsSupported)
	}

	validator := services.GetAPIKeyValidator()
	history, err := validator.GetVerificationHistory(keyID, 10)
	if err != nil {
		logger.LogError(context.Background(), "merchant_apikey", "Failed to get verification history", err, map[string]interface{}{
			"api_key_id": keyID,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"api_key": apiKey,
		"history": history,
	})
}
