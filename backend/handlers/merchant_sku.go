package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/cache"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/models"
)

func ListMerchantSKUs(c *gin.Context) {
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
	err := db.QueryRow("SELECT id FROM merchants WHERE user_id = $1 AND status = 'active'", userIDInt).Scan(&merchantID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"MERCHANT_NOT_FOUND",
			"商户不存在或未激活",
			http.StatusNotFound,
			err,
		))
		return
	}

	status := c.DefaultQuery("status", "active")

	ctx := context.Background()
	cacheKey := cache.MerchantSKUsKey(merchantID, status)

	if cachedSKUs, cacheErr := cache.Get(ctx, cacheKey); cacheErr == nil {
		var skus []models.MerchantSKUDetail
		if unmarshalErr := json.Unmarshal([]byte(cachedSKUs), &skus); unmarshalErr == nil {
			c.JSON(http.StatusOK, gin.H{"data": skus})
			return
		}
	}

	query := `SELECT id, merchant_id, sku_id, api_key_id, status, sales_count, total_sales_amount, created_at, updated_at,
		sku_code, sku_type, token_amount, compute_points, retail_price, original_price, valid_days, 
		group_enabled, group_discount_rate, spu_name, model_provider, model_name, model_tier, api_key_name, api_key_provider
		FROM merchant_sku_details WHERE merchant_id = $1`
	args := []interface{}{merchantID}

	if status != "" && status != "all" {
		query += " AND status = $" + strconv.Itoa(len(args)+1)
		args = append(args, status)
	}

	query += " ORDER BY created_at DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	var skus []models.MerchantSKUDetail
	for rows.Next() {
		var s models.MerchantSKUDetail
		var apiKeyID sql.NullInt64
		var apiKeyName, apiKeyProvider sql.NullString
		var tokenAmount sql.NullInt64
		var computePoints, originalPrice, groupDiscountRate sql.NullFloat64

		err := rows.Scan(&s.ID, &s.MerchantID, &s.SKUID, &apiKeyID, &s.Status, &s.SalesCount, &s.TotalSalesAmount, &s.CreatedAt, &s.UpdatedAt,
			&s.SKUCode, &s.SKUType, &tokenAmount, &computePoints, &s.RetailPrice, &originalPrice, &s.ValidDays,
			&s.GroupEnabled, &groupDiscountRate, &s.SPUName, &s.ModelProvider, &s.ModelName, &s.ModelTier, &apiKeyName, &apiKeyProvider)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}

		if apiKeyID.Valid {
			mid := int(apiKeyID.Int64)
			s.APIKeyID = &mid
		}
		if apiKeyName.Valid {
			s.APIKeyName = apiKeyName.String
		}
		if apiKeyProvider.Valid {
			s.APIKeyProvider = apiKeyProvider.String
		}
		if tokenAmount.Valid {
			s.TokenAmount = tokenAmount.Int64
		}
		if computePoints.Valid {
			s.ComputePoints = computePoints.Float64
		}
		if originalPrice.Valid {
			s.OriginalPrice = originalPrice.Float64
		}
		if groupDiscountRate.Valid {
			s.GroupDiscountRate = groupDiscountRate.Float64
		}

		skus = append(skus, s)
	}

	if skusJSON, err := json.Marshal(skus); err == nil {
		cache.Set(ctx, cacheKey, string(skusJSON), cache.MerchantCacheTTL)
	}

	c.JSON(http.StatusOK, gin.H{"data": skus})
}

func GetAvailableSKUs(c *gin.Context) {
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
	err := db.QueryRow("SELECT id FROM merchants WHERE user_id = $1 AND status = 'active'", userIDInt).Scan(&merchantID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"MERCHANT_NOT_FOUND",
			"商户不存在或未激活",
			http.StatusNotFound,
			err,
		))
		return
	}

	provider := c.Query("provider")
	skuType := c.Query("type")

	query := `SELECT s.id, s.sku_code, s.sku_type, s.token_amount, s.compute_points, s.retail_price, s.original_price, 
		s.valid_days, s.group_enabled, s.group_discount_rate, s.spu_id, sp.name as spu_name, sp.model_provider, sp.model_name, sp.model_tier
		FROM skus s JOIN spus sp ON s.spu_id = sp.id 
		WHERE s.status = 'active' AND sp.status = 'active'`
	args := []interface{}{}

	if provider != "" {
		query += " AND sp.model_provider = $" + strconv.Itoa(len(args)+1)
		args = append(args, provider)
	}
	if skuType != "" {
		query += " AND s.sku_type = $" + strconv.Itoa(len(args)+1)
		args = append(args, skuType)
	}

	query += " ORDER BY sp.model_tier, s.retail_price ASC"

	rows, err := db.Query(query, args...)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	var skus []models.AvailableSKU
	for rows.Next() {
		var s models.AvailableSKU
		var tokenAmount sql.NullInt64
		var computePoints, originalPrice, groupDiscountRate sql.NullFloat64

		scanErr := rows.Scan(&s.ID, &s.SKUCode, &s.SKUType, &tokenAmount, &computePoints, &s.RetailPrice, &originalPrice,
			&s.ValidDays, &s.GroupEnabled, &groupDiscountRate, &s.SPUID, &s.SPUName, &s.ModelProvider, &s.ModelName, &s.ModelTier)
		if scanErr != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}

		if tokenAmount.Valid {
			s.TokenAmount = tokenAmount.Int64
		}
		if computePoints.Valid {
			s.ComputePoints = computePoints.Float64
		}
		if originalPrice.Valid {
			s.OriginalPrice = originalPrice.Float64
		}
		if groupDiscountRate.Valid {
			s.GroupDiscountRate = groupDiscountRate.Float64
		}

		skus = append(skus, s)
	}

	rows2, err := db.Query("SELECT sku_id FROM merchant_skus WHERE merchant_id = $1", merchantID)
	if err == nil {
		defer rows2.Close()
		selectedMap := make(map[int]bool)
		for rows2.Next() {
			var skuID int
			if scanErr := rows2.Scan(&skuID); scanErr == nil {
				selectedMap[skuID] = true
			}
		}
		for i := range skus {
			skus[i].IsSelected = selectedMap[skus[i].ID]
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": skus})
}

func CreateMerchantSKU(c *gin.Context) {
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

	var req models.MerchantSKUCreateRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var merchantID int
	err := db.QueryRow("SELECT id FROM merchants WHERE user_id = $1 AND status = 'active'", userIDInt).Scan(&merchantID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"MERCHANT_NOT_FOUND",
			"商户不存在或未激活",
			http.StatusNotFound,
			err,
		))
		return
	}

	var skuExists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM skus WHERE id = $1 AND status = 'active')", req.SKUID).Scan(&skuExists)
	if err != nil || !skuExists {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"SKU_NOT_FOUND",
			"SKU不存在或已下架",
			http.StatusBadRequest,
			err,
		))
		return
	}

	var alreadyExists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM merchant_skus WHERE merchant_id = $1 AND sku_id = $2)", merchantID, req.SKUID).Scan(&alreadyExists)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	if alreadyExists {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"SKU_ALREADY_SELECTED",
			"该SKU已选择",
			http.StatusBadRequest,
			nil,
		))
		return
	}

	if req.APIKeyID != nil && *req.APIKeyID > 0 {
		var apiKeyExists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM merchant_api_keys WHERE id = $1 AND merchant_id = $2 AND status = 'active')", *req.APIKeyID, merchantID).Scan(&apiKeyExists)
		if err != nil || !apiKeyExists {
			middleware.RespondWithError(c, apperrors.NewAppError(
				"API_KEY_NOT_FOUND",
				"API Key不存在或不属于当前商户",
				http.StatusBadRequest,
				err,
			))
			return
		}
	}

	var merchantSKU models.MerchantSKUDetail
	err = db.QueryRow(
		`INSERT INTO merchant_skus (merchant_id, sku_id, api_key_id, status) 
		 VALUES ($1, $2, $3, 'active') 
		 RETURNING id, merchant_id, sku_id, api_key_id, status, sales_count, total_sales_amount, created_at, updated_at`,
		merchantID, req.SKUID, req.APIKeyID,
	).Scan(&merchantSKU.ID, &merchantSKU.MerchantID, &merchantSKU.SKUID, &merchantSKU.APIKeyID, &merchantSKU.Status, &merchantSKU.SalesCount, &merchantSKU.TotalSalesAmount, &merchantSKU.CreatedAt, &merchantSKU.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"MERCHANT_SKU_CREATE_FAILED",
			"创建商户SKU失败",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	err = db.QueryRow(
		`SELECT s.sku_code, s.sku_type, s.retail_price, s.valid_days, s.group_enabled, sp.name as spu_name, sp.model_provider, sp.model_name, sp.model_tier
		 FROM skus s JOIN spus sp ON s.spu_id = sp.id WHERE s.id = $1`,
		merchantSKU.SKUID,
	).Scan(&merchantSKU.SKUCode, &merchantSKU.SKUType, &merchantSKU.RetailPrice, &merchantSKU.ValidDays, &merchantSKU.GroupEnabled, &merchantSKU.SPUName, &merchantSKU.ModelProvider, &merchantSKU.ModelName, &merchantSKU.ModelTier)

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	if merchantSKU.APIKeyID != nil && *merchantSKU.APIKeyID > 0 {
		err = db.QueryRow("SELECT name, provider FROM merchant_api_keys WHERE id = $1", *merchantSKU.APIKeyID).Scan(&merchantSKU.APIKeyName, &merchantSKU.APIKeyProvider)
		if err != nil {
			merchantSKU.APIKeyName = ""
			merchantSKU.APIKeyProvider = ""
		}
	}

	ctx := context.Background()
	cache.Delete(ctx, cache.MerchantSKUsKey(merchantID, ""))

	c.JSON(http.StatusCreated, gin.H{"data": merchantSKU})
}

func UpdateMerchantSKU(c *gin.Context) {
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

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	var req models.MerchantSKUUpdateRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
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
			"商户不存在",
			http.StatusNotFound,
			err,
		))
		return
	}

	if req.APIKeyID != nil && *req.APIKeyID > 0 {
		var apiKeyExists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM merchant_api_keys WHERE id = $1 AND merchant_id = $2 AND status = 'active')", *req.APIKeyID, merchantID).Scan(&apiKeyExists)
		if err != nil || !apiKeyExists {
			middleware.RespondWithError(c, apperrors.NewAppError(
				"API_KEY_NOT_FOUND",
				"API Key不存在或不属于当前商户",
				http.StatusBadRequest,
				err,
			))
			return
		}
	}

	var merchantSKU models.MerchantSKUDetail
	err = db.QueryRow(
		`UPDATE merchant_skus SET api_key_id = $1, status = COALESCE($2, status), updated_at = CURRENT_TIMESTAMP 
		 WHERE id = $3 AND merchant_id = $4 
		 RETURNING id, merchant_id, sku_id, api_key_id, status, sales_count, total_sales_amount, created_at, updated_at`,
		req.APIKeyID, req.Status, id, merchantID,
	).Scan(&merchantSKU.ID, &merchantSKU.MerchantID, &merchantSKU.SKUID, &merchantSKU.APIKeyID, &merchantSKU.Status, &merchantSKU.SalesCount, &merchantSKU.TotalSalesAmount, &merchantSKU.CreatedAt, &merchantSKU.UpdatedAt)

	if err == sql.ErrNoRows {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"MERCHANT_SKU_NOT_FOUND",
			"商户SKU不存在",
			http.StatusNotFound,
			err,
		))
		return
	} else if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	err = db.QueryRow(
		`SELECT s.sku_code, s.sku_type, s.retail_price, s.valid_days, s.group_enabled, sp.name as spu_name, sp.model_provider, sp.model_name, sp.model_tier
		 FROM skus s JOIN spus sp ON s.spu_id = sp.id WHERE s.id = $1`,
		merchantSKU.SKUID,
	).Scan(&merchantSKU.SKUCode, &merchantSKU.SKUType, &merchantSKU.RetailPrice, &merchantSKU.ValidDays, &merchantSKU.GroupEnabled, &merchantSKU.SPUName, &merchantSKU.ModelProvider, &merchantSKU.ModelName, &merchantSKU.ModelTier)

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	if merchantSKU.APIKeyID != nil && *merchantSKU.APIKeyID > 0 {
		err = db.QueryRow("SELECT name, provider FROM merchant_api_keys WHERE id = $1", *merchantSKU.APIKeyID).Scan(&merchantSKU.APIKeyName, &merchantSKU.APIKeyProvider)
		if err != nil {
			merchantSKU.APIKeyName = ""
			merchantSKU.APIKeyProvider = ""
		}
	}

	ctx := context.Background()
	cache.Delete(ctx, cache.MerchantSKUsKey(merchantID, ""))

	c.JSON(http.StatusOK, gin.H{"data": merchantSKU})
}

func DeleteMerchantSKU(c *gin.Context) {
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

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
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
			"商户不存在",
			http.StatusNotFound,
			err,
		))
		return
	}

	result, err := db.Exec("DELETE FROM merchant_skus WHERE id = $1 AND merchant_id = $2", id, merchantID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"MERCHANT_SKU_NOT_FOUND",
			"商户SKU不存在",
			http.StatusNotFound,
			nil,
		))
		return
	}

	ctx := context.Background()
	cache.Delete(ctx, cache.MerchantSKUsKey(merchantID, ""))

	c.JSON(http.StatusOK, gin.H{"message": "SKU已下架"})
}
