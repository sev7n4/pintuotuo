package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/pintuotuo/backend/cache"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/models"
)

const merchantSKUStatusInactive = "inactive"
const defaultMerchantProfitMargin = 20.0

func merchantSKUKeyConflictAppErr(cause error) *apperrors.AppError {
	return apperrors.NewAppError(
		"MERCHANT_SKU_KEY_CONFLICT",
		"该 API Key 已绑定另一在售 SKU，请更换 Key 或先下架已占用该 Key 的商品",
		http.StatusConflict,
		cause,
	)
}

// hasOtherActiveMerchantSKUForAPIKey 在「将存在一条 active 且绑定该 Key 的 merchant_sku」为真时返回 true。excludeMerchantSKUID=0 表示不排除（用于新建）。
func hasOtherActiveMerchantSKUForAPIKey(db *sql.DB, apiKeyID int, excludeMerchantSKUID int) (bool, error) {
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM merchant_skus
			WHERE api_key_id = $1 AND status = 'active'
			  AND ($2 = 0 OR id <> $2)
		)`, apiKeyID, excludeMerchantSKUID).Scan(&exists)
	return exists, err
}

func resolveSKUDefaultCostBySKUID(db *sql.DB, skuID int) (float64, float64, error) {
	var inputRate, outputRate float64
	err := db.QueryRow(
		`SELECT
			CASE WHEN COALESCE(s.inherit_spu_cost, true) THEN COALESCE(sp.provider_input_rate, 0) ELSE COALESCE(s.cost_input_rate, 0) END,
			CASE WHEN COALESCE(s.inherit_spu_cost, true) THEN COALESCE(sp.provider_output_rate, 0) ELSE COALESCE(s.cost_output_rate, 0) END
		 FROM skus s
		 JOIN spus sp ON sp.id = s.spu_id
		 WHERE s.id = $1`,
		skuID,
	).Scan(&inputRate, &outputRate)
	return inputRate, outputRate, err
}

func normalizeMerchantSKUCost(req models.MerchantSKUCreateRequest, spuInput, spuOutput float64) (float64, float64, float64, bool, *apperrors.AppError) {
	profit := defaultMerchantProfitMargin
	if req.ProfitMargin != nil && *req.ProfitMargin >= 0 {
		profit = *req.ProfitMargin
	}
	if req.CustomPricingEnabled {
		if req.CostInputRate == nil || req.CostOutputRate == nil {
			return 0, 0, 0, false, apperrors.NewAppError(
				"COST_REQUIRED",
				"开启自定义成本时，输入/输出成本为必填",
				http.StatusBadRequest,
				nil,
			)
		}
		return *req.CostInputRate, *req.CostOutputRate, profit, true, nil
	}
	return spuInput, spuOutput, profit, false, nil
}

func normalizeMerchantSKUCostUpdate(req models.MerchantSKUUpdateRequest, currentInput, currentOutput, currentProfit float64, currentCustom bool) (float64, float64, float64, bool, *apperrors.AppError) {
	custom := currentCustom
	if req.CustomPricingEnabled != nil {
		custom = *req.CustomPricingEnabled
	}
	inputRate := currentInput
	outputRate := currentOutput
	profit := currentProfit
	if req.CostInputRate != nil {
		inputRate = *req.CostInputRate
	}
	if req.CostOutputRate != nil {
		outputRate = *req.CostOutputRate
	}
	if req.ProfitMargin != nil && *req.ProfitMargin >= 0 {
		profit = *req.ProfitMargin
	}
	if custom && (inputRate <= 0 || outputRate <= 0) {
		return 0, 0, 0, false, apperrors.NewAppError(
			"COST_REQUIRED",
			"开启自定义成本时，输入/输出成本必须大于 0",
			http.StatusBadRequest,
			nil,
		)
	}
	return inputRate, outputRate, profit, custom, nil
}

func syncMerchantAPIKeyCostByID(db *sql.DB, merchantSKUID int) error {
	var apiKeyID sql.NullInt64
	var inputRate, outputRate, profit float64
	err := db.QueryRow(
		`SELECT api_key_id, COALESCE(cost_input_rate, 0), COALESCE(cost_output_rate, 0), COALESCE(profit_margin, 20)
		 FROM merchant_skus WHERE id = $1`,
		merchantSKUID,
	).Scan(&apiKeyID, &inputRate, &outputRate, &profit)
	if err != nil || !apiKeyID.Valid || apiKeyID.Int64 <= 0 {
		return err
	}
	_, err = db.Exec(
		`UPDATE merchant_api_keys
		 SET cost_input_rate = $1, cost_output_rate = $2, profit_margin = $3, updated_at = CURRENT_TIMESTAMP
		 WHERE id = $4`,
		inputRate, outputRate, profit, apiKeyID.Int64,
	)
	return err
}

func invalidateMerchantSKUCache(ctx context.Context, merchantID int) {
	// invalidate all status-filtered sku list cache
	cache.Delete(ctx, cache.MerchantSKUsKey(merchantID, ""))
	cache.Delete(ctx, cache.MerchantSKUsKey(merchantID, allProductStatus))
	cache.Delete(ctx, cache.MerchantSKUsKey(merchantID, merchantStatusActive))
	cache.Delete(ctx, cache.MerchantSKUsKey(merchantID, merchantSKUStatusInactive))
	// invalidate all available sku filter cache
	_ = cache.InvalidatePatterns(ctx, cache.AvailableSKUsKey(merchantID, "*", "*"))
}

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
	err := db.QueryRow("SELECT id FROM merchants WHERE user_id = $1", userIDInt).Scan(&merchantID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"MERCHANT_NOT_FOUND",
			"商户不存在",
			http.StatusNotFound,
			err,
		))
		return
	}

	status := c.DefaultQuery("status", merchantStatusActive)

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
		group_enabled, group_discount_rate, spu_name, model_provider, model_name, model_tier, api_key_name, api_key_provider,
		COALESCE(cost_input_rate, 0), COALESCE(cost_output_rate, 0), COALESCE(profit_margin, 20), COALESCE(custom_pricing_enabled, false),
		COALESCE(spu_input_rate, 0), COALESCE(spu_output_rate, 0)
		FROM merchant_sku_details WHERE merchant_id = $1`
	args := []interface{}{merchantID}

	if status != "" && status != allProductStatus {
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
			&s.GroupEnabled, &groupDiscountRate, &s.SPUName, &s.ModelProvider, &s.ModelName, &s.ModelTier, &apiKeyName, &apiKeyProvider,
			&s.CostInputRate, &s.CostOutputRate, &s.ProfitMargin, &s.CustomPricing, &s.SPUInputRate, &s.SPUOutputRate)
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
	err := db.QueryRow("SELECT id FROM merchants WHERE user_id = $1", userIDInt).Scan(&merchantID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"MERCHANT_NOT_FOUND",
			"商户不存在",
			http.StatusNotFound,
			err,
		))
		return
	}

	provider := c.Query("provider")
	skuType := c.Query("type")

	query := `SELECT s.id, s.sku_code, s.sku_type, s.token_amount, s.compute_points, s.retail_price, s.original_price, 
		s.valid_days, s.group_enabled, s.group_discount_rate, s.spu_id, sp.name as spu_name, sp.model_provider, sp.model_name, sp.model_tier,
		COALESCE(sp.provider_input_rate, 0), COALESCE(sp.provider_output_rate, 0)
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
			&s.ValidDays, &s.GroupEnabled, &groupDiscountRate, &s.SPUID, &s.SPUName, &s.ModelProvider, &s.ModelName, &s.ModelTier, &s.SPUInputRate, &s.SPUOutputRate)
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

	rows2, err := db.Query("SELECT sku_id FROM merchant_skus WHERE merchant_id = $1 AND status = 'active'", merchantID)
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

	var existingID int
	var existingStatus string
	err = db.QueryRow(
		"SELECT id, status FROM merchant_skus WHERE merchant_id = $1 AND sku_id = $2",
		merchantID, req.SKUID,
	).Scan(&existingID, &existingStatus)
	if err != nil && err != sql.ErrNoRows {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	if err == nil && existingStatus == merchantStatusActive {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"SKU_ALREADY_SELECTED",
			"该SKU已选择",
			http.StatusBadRequest,
			nil,
		))
		return
	}
	if err == nil && existingStatus == merchantSKUStatusInactive {
		var spuInputRate, spuOutputRate float64
		spuInputRate, spuOutputRate, err = resolveSKUDefaultCostBySKUID(db, req.SKUID)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		costInputRate, costOutputRate, profitMargin, customPricing, appErr := normalizeMerchantSKUCost(req, spuInputRate, spuOutputRate)
		if appErr != nil {
			middleware.RespondWithError(c, appErr)
			return
		}
		if req.APIKeyID != nil && *req.APIKeyID > 0 {
			var dup bool
			dup, err = hasOtherActiveMerchantSKUForAPIKey(db, *req.APIKeyID, existingID)
			if err != nil {
				middleware.RespondWithError(c, apperrors.ErrDatabaseError)
				return
			}
			if dup {
				middleware.RespondWithError(c, merchantSKUKeyConflictAppErr(nil))
				return
			}
		}
		var reactivated models.MerchantSKUDetail
		err = db.QueryRow(
			`UPDATE merchant_skus SET status = $1, api_key_id = $2, cost_input_rate = $3, cost_output_rate = $4,
			 profit_margin = $5, custom_pricing_enabled = $6, updated_at = CURRENT_TIMESTAMP
			 WHERE id = $7
			 RETURNING id, merchant_id, sku_id, api_key_id, status, sales_count, total_sales_amount, created_at, updated_at`,
			merchantStatusActive, req.APIKeyID, costInputRate, costOutputRate, profitMargin, customPricing, existingID,
		).Scan(&reactivated.ID, &reactivated.MerchantID, &reactivated.SKUID, &reactivated.APIKeyID, &reactivated.Status,
			&reactivated.SalesCount, &reactivated.TotalSalesAmount, &reactivated.CreatedAt, &reactivated.UpdatedAt)
		if err != nil {
			var pqErr *pq.Error
			if errors.As(err, &pqErr) && pqErr.Code == "23505" {
				middleware.RespondWithError(c, merchantSKUKeyConflictAppErr(err))
				return
			}
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		err = db.QueryRow(
			`SELECT s.sku_code, s.sku_type, s.retail_price, s.valid_days, s.group_enabled, sp.name as spu_name, sp.model_provider, sp.model_name, sp.model_tier
			 FROM skus s JOIN spus sp ON s.spu_id = sp.id WHERE s.id = $1`,
			reactivated.SKUID,
		).Scan(&reactivated.SKUCode, &reactivated.SKUType, &reactivated.RetailPrice, &reactivated.ValidDays, &reactivated.GroupEnabled,
			&reactivated.SPUName, &reactivated.ModelProvider, &reactivated.ModelName, &reactivated.ModelTier)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		if reactivated.APIKeyID != nil && *reactivated.APIKeyID > 0 {
			_ = db.QueryRow("SELECT name, provider FROM merchant_api_keys WHERE id = $1", *reactivated.APIKeyID).Scan(&reactivated.APIKeyName, &reactivated.APIKeyProvider)
		}
		reactivated.CostInputRate = costInputRate
		reactivated.CostOutputRate = costOutputRate
		reactivated.ProfitMargin = profitMargin
		reactivated.CustomPricing = customPricing
		reactivated.SPUInputRate = spuInputRate
		reactivated.SPUOutputRate = spuOutputRate
		_ = syncMerchantAPIKeyCostByID(db, reactivated.ID)
		ctx := context.Background()
		invalidateMerchantSKUCache(ctx, merchantID)
		c.JSON(http.StatusOK, gin.H{"data": reactivated})
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
		var dup bool
		dup, err = hasOtherActiveMerchantSKUForAPIKey(db, *req.APIKeyID, 0)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		if dup {
			middleware.RespondWithError(c, merchantSKUKeyConflictAppErr(nil))
			return
		}
	}

	spuInputRate, spuOutputRate, err := resolveSKUDefaultCostBySKUID(db, req.SKUID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	costInputRate, costOutputRate, profitMargin, customPricing, appErr := normalizeMerchantSKUCost(req, spuInputRate, spuOutputRate)
	if appErr != nil {
		middleware.RespondWithError(c, appErr)
		return
	}

	var merchantSKU models.MerchantSKUDetail
	err = db.QueryRow(
		`INSERT INTO merchant_skus (merchant_id, sku_id, api_key_id, status, cost_input_rate, cost_output_rate, profit_margin, custom_pricing_enabled) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
		 RETURNING id, merchant_id, sku_id, api_key_id, status, sales_count, total_sales_amount, created_at, updated_at`,
		merchantID, req.SKUID, req.APIKeyID, merchantStatusActive, costInputRate, costOutputRate, profitMargin, customPricing,
	).Scan(&merchantSKU.ID, &merchantSKU.MerchantID, &merchantSKU.SKUID, &merchantSKU.APIKeyID, &merchantSKU.Status, &merchantSKU.SalesCount, &merchantSKU.TotalSalesAmount, &merchantSKU.CreatedAt, &merchantSKU.UpdatedAt)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			middleware.RespondWithError(c, merchantSKUKeyConflictAppErr(err))
			return
		}
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
	merchantSKU.CostInputRate = costInputRate
	merchantSKU.CostOutputRate = costOutputRate
	merchantSKU.ProfitMargin = profitMargin
	merchantSKU.CustomPricing = customPricing
	merchantSKU.SPUInputRate = spuInputRate
	merchantSKU.SPUOutputRate = spuOutputRate
	_ = syncMerchantAPIKeyCostByID(db, merchantSKU.ID)

	ctx := context.Background()
	invalidateMerchantSKUCache(ctx, merchantID)

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

	var currentInput, currentOutput, currentProfit float64
	var currentCustom bool
	var curKey sql.NullInt64
	var curStatus string
	err = db.QueryRow(
		`SELECT COALESCE(cost_input_rate, 0), COALESCE(cost_output_rate, 0), COALESCE(profit_margin, 20), COALESCE(custom_pricing_enabled, false),
			api_key_id, status
		 FROM merchant_skus WHERE id = $1 AND merchant_id = $2`,
		id, merchantID,
	).Scan(&currentInput, &currentOutput, &currentProfit, &currentCustom, &curKey, &curStatus)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"MERCHANT_SKU_NOT_FOUND",
			"商户SKU不存在",
			http.StatusNotFound,
			err,
		))
		return
	}
	effKey := curKey
	if req.APIKeyID != nil {
		if *req.APIKeyID <= 0 {
			effKey = sql.NullInt64{}
		} else {
			effKey = sql.NullInt64{Int64: int64(*req.APIKeyID), Valid: true}
		}
	}
	effStatus := curStatus
	if strings.TrimSpace(req.Status) != "" {
		effStatus = strings.TrimSpace(req.Status)
	}
	if effKey.Valid && effKey.Int64 > 0 && effStatus == merchantStatusActive {
		var dup bool
		dup, err = hasOtherActiveMerchantSKUForAPIKey(db, int(effKey.Int64), id)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		if dup {
			middleware.RespondWithError(c, merchantSKUKeyConflictAppErr(nil))
			return
		}
	}
	newInput, newOutput, newProfit, newCustom, appErr := normalizeMerchantSKUCostUpdate(req, currentInput, currentOutput, currentProfit, currentCustom)
	if appErr != nil {
		middleware.RespondWithError(c, appErr)
		return
	}

	var merchantSKU models.MerchantSKUDetail
	err = db.QueryRow(
		`UPDATE merchant_skus SET api_key_id = $1, status = COALESCE($2, status),
		 cost_input_rate = $3, cost_output_rate = $4, profit_margin = $5, custom_pricing_enabled = $6,
		 updated_at = CURRENT_TIMESTAMP 
		 WHERE id = $7 AND merchant_id = $8
		 RETURNING id, merchant_id, sku_id, api_key_id, status, sales_count, total_sales_amount, created_at, updated_at`,
		req.APIKeyID, req.Status, newInput, newOutput, newProfit, newCustom, id, merchantID,
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
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			middleware.RespondWithError(c, merchantSKUKeyConflictAppErr(err))
			return
		}
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
	merchantSKU.CostInputRate = newInput
	merchantSKU.CostOutputRate = newOutput
	merchantSKU.ProfitMargin = newProfit
	merchantSKU.CustomPricing = newCustom
	_ = syncMerchantAPIKeyCostByID(db, merchantSKU.ID)

	ctx := context.Background()
	invalidateMerchantSKUCache(ctx, merchantID)

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
	invalidateMerchantSKUCache(ctx, merchantID)

	c.JSON(http.StatusOK, gin.H{"message": "SKU已下架"})
}
