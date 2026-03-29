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

func ListSPUs(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("per_page", "20")
	provider := c.Query("provider")
	tier := c.Query("tier")
	status := c.DefaultQuery("status", "active")

	pageNum, _ := strconv.Atoi(page)
	perPageNum, _ := strconv.Atoi(perPage)

	if pageNum < 1 {
		pageNum = 1
	}
	if perPageNum < 1 || perPageNum > 100 {
		perPageNum = 20
	}

	ctx := context.Background()
	cacheKey := cache.SPUListKey(pageNum, perPageNum, provider, tier, status)

	if cachedList, err := cache.Get(ctx, cacheKey); err == nil {
		var cachedData struct {
			Total   int          `json:"total"`
			Page    int          `json:"page"`
			PerPage int          `json:"per_page"`
			Data    []models.SPU `json:"data"`
		}
		if err := json.Unmarshal([]byte(cachedList), &cachedData); err == nil {
			c.JSON(http.StatusOK, cachedData)
			return
		}
	}

	offset := (pageNum - 1) * perPageNum
	db := config.GetDB()

	query := "SELECT id, spu_code, name, model_provider, model_name, model_version, model_tier, context_window, base_compute_points, description, status, sort_order, total_sales_count, COALESCE(average_rating, 0), created_at, updated_at FROM spus WHERE 1=1"
	args := []interface{}{}
	argPos := 1

	if status != "" && status != "all" {
		query += " AND status = $" + strconv.Itoa(argPos)
		args = append(args, status)
		argPos++
	}
	if provider != "" {
		query += " AND model_provider = $" + strconv.Itoa(argPos)
		args = append(args, provider)
		argPos++
	}
	if tier != "" {
		query += " AND model_tier = $" + strconv.Itoa(argPos)
		args = append(args, tier)
		argPos++
	}

	query += " ORDER BY sort_order ASC, created_at DESC LIMIT $" + strconv.Itoa(argPos) + " OFFSET $" + strconv.Itoa(argPos+1)
	args = append(args, perPageNum, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	var spus []models.SPU
	for rows.Next() {
		var s models.SPU
		err := rows.Scan(&s.ID, &s.SPUCode, &s.Name, &s.ModelProvider, &s.ModelName, &s.ModelVersion, &s.ModelTier, &s.ContextWindow, &s.BaseComputePoints, &s.Description, &s.Status, &s.SortOrder, &s.TotalSalesCount, &s.AverageRating, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		spus = append(spus, s)
	}

	countQuery := "SELECT COUNT(*) FROM spus WHERE 1=1"
	countArgs := []interface{}{}

	if status != "" && status != "all" {
		countQuery += " AND status = $" + strconv.Itoa(len(countArgs)+1)
		countArgs = append(countArgs, status)
	}
	if provider != "" {
		countQuery += " AND model_provider = $" + strconv.Itoa(len(countArgs)+1)
		countArgs = append(countArgs, provider)
	}
	if tier != "" {
		countQuery += " AND model_tier = $" + strconv.Itoa(len(countArgs)+1)
		countArgs = append(countArgs, tier)
	}

	var total int
	db.QueryRow(countQuery, countArgs...).Scan(&total)

	result := gin.H{
		"total":    total,
		"page":     pageNum,
		"per_page": perPageNum,
		"data":     spus,
	}

	if resultJSON, err := json.Marshal(result); err == nil {
		cache.Set(ctx, cacheKey, string(resultJSON), cache.ProductListTTL)
	}

	c.JSON(http.StatusOK, result)
}

func GetSPUByID(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()

	spuID := idToInt(id)
	if spuID <= 0 {
		middleware.RespondWithError(c, apperrors.ErrProductNotFound)
		return
	}

	cacheKey := cache.SPUKey(spuID)
	if cachedSPU, err := cache.Get(ctx, cacheKey); err == nil {
		var spu models.SPU
		if err := json.Unmarshal([]byte(cachedSPU), &spu); err == nil {
			c.JSON(http.StatusOK, gin.H{"data": spu})
			return
		}
	}

	db := config.GetDB()
	var spu models.SPU
	err := db.QueryRow(
		"SELECT id, spu_code, name, model_provider, model_name, model_version, model_tier, context_window, max_output_tokens, base_compute_points, description, thumbnail_url, status, sort_order, total_sales_count, COALESCE(average_rating, 0), created_at, updated_at FROM spus WHERE id = $1",
		spuID,
	).Scan(&spu.ID, &spu.SPUCode, &spu.Name, &spu.ModelProvider, &spu.ModelName, &spu.ModelVersion, &spu.ModelTier, &spu.ContextWindow, &spu.MaxOutputTokens, &spu.BaseComputePoints, &spu.Description, &spu.ThumbnailURL, &spu.Status, &spu.SortOrder, &spu.TotalSalesCount, &spu.AverageRating, &spu.CreatedAt, &spu.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrProductNotFound)
		return
	}

	if spuJSON, err := json.Marshal(spu); err == nil {
		cache.Set(ctx, cacheKey, string(spuJSON), cache.ProductCacheTTL)
	}

	c.JSON(http.StatusOK, gin.H{"data": spu})
}

func CreateSPU(c *gin.Context) {
	var req models.SPUCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	if req.BaseComputePoints <= 0 {
		req.BaseComputePoints = 1.0
	}
	if req.Status == "" {
		req.Status = "active"
	}

	db := config.GetDB()
	var spu models.SPU
	err := db.QueryRow(
		`INSERT INTO spus (spu_code, name, model_provider, model_name, model_version, model_tier, context_window, max_output_tokens, base_compute_points, description, thumbnail_url, status, sort_order) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) 
		 RETURNING id, spu_code, name, model_provider, model_name, model_version, model_tier, context_window, base_compute_points, description, status, sort_order, created_at, updated_at`,
		req.SPUCode, req.Name, req.ModelProvider, req.ModelName, req.ModelVersion, req.ModelTier, req.ContextWindow, req.MaxOutputTokens, req.BaseComputePoints, req.Description, req.ThumbnailURL, req.Status, req.SortOrder,
	).Scan(&spu.ID, &spu.SPUCode, &spu.Name, &spu.ModelProvider, &spu.ModelName, &spu.ModelVersion, &spu.ModelTier, &spu.ContextWindow, &spu.BaseComputePoints, &spu.Description, &spu.Status, &spu.SortOrder, &spu.CreatedAt, &spu.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"SPU_CREATION_FAILED",
			"创建SPU失败",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	ctx := context.Background()
	cache.InvalidatePatterns(ctx, "spus:list:*")

	c.JSON(http.StatusCreated, gin.H{"data": spu})
}

func UpdateSPU(c *gin.Context) {
	id := c.Param("id")
	spuID := idToInt(id)
	if spuID <= 0 {
		middleware.RespondWithError(c, apperrors.ErrProductNotFound)
		return
	}

	var req models.SPUCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()
	var spu models.SPU
	err := db.QueryRow(
		`UPDATE spus SET 
		 name = COALESCE(NULLIF($1, ''), name), 
		 model_provider = COALESCE(NULLIF($2, ''), model_provider),
		 model_name = COALESCE(NULLIF($3, ''), model_name),
		 model_version = COALESCE(NULLIF($4, ''), model_version),
		 model_tier = COALESCE(NULLIF($5, ''), model_tier),
		 context_window = CASE WHEN $6 > 0 THEN $6 ELSE context_window END,
		 base_compute_points = CASE WHEN $7 > 0 THEN $7 ELSE base_compute_points END,
		 description = COALESCE(NULLIF($8, ''), description),
		 status = COALESCE(NULLIF($9, ''), status),
		 sort_order = $10
		 WHERE id = $11 
		 RETURNING id, spu_code, name, model_provider, model_name, model_version, model_tier, context_window, base_compute_points, description, status, sort_order, created_at, updated_at`,
		req.Name, req.ModelProvider, req.ModelName, req.ModelVersion, req.ModelTier, req.ContextWindow, req.BaseComputePoints, req.Description, req.Status, req.SortOrder, spuID,
	).Scan(&spu.ID, &spu.SPUCode, &spu.Name, &spu.ModelProvider, &spu.ModelName, &spu.ModelVersion, &spu.ModelTier, &spu.ContextWindow, &spu.BaseComputePoints, &spu.Description, &spu.Status, &spu.SortOrder, &spu.CreatedAt, &spu.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"SPU_UPDATE_FAILED",
			"更新SPU失败",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	ctx := context.Background()
	cache.Delete(ctx, cache.SPUKey(spuID))
	cache.InvalidatePatterns(ctx, "spus:list:*")

	c.JSON(http.StatusOK, gin.H{"data": spu})
}

func DeleteSPU(c *gin.Context) {
	id := c.Param("id")
	spuID := idToInt(id)
	if spuID <= 0 {
		middleware.RespondWithError(c, apperrors.ErrProductNotFound)
		return
	}

	db := config.GetDB()
	result, err := db.Exec("DELETE FROM spus WHERE id = $1", spuID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"SPU_DELETE_FAILED",
			"删除SPU失败",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		middleware.RespondWithError(c, apperrors.ErrProductNotFound)
		return
	}

	ctx := context.Background()
	cache.Delete(ctx, cache.SPUKey(spuID))
	cache.InvalidatePatterns(ctx, "spus:list:*")
	cache.InvalidatePatterns(ctx, "skus:list:*")

	c.JSON(http.StatusOK, gin.H{"message": "SPU deleted successfully"})
}

func ListSKUs(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("per_page", "20")
	spuID := c.Query("spu_id")
	skuType := c.Query("type")
	status := c.DefaultQuery("status", "active")

	pageNum, _ := strconv.Atoi(page)
	perPageNum, _ := strconv.Atoi(perPage)

	if pageNum < 1 {
		pageNum = 1
	}
	if perPageNum < 1 || perPageNum > 100 {
		perPageNum = 20
	}

	ctx := context.Background()
	cacheKey := cache.SKUListKey(pageNum, perPageNum, spuID, skuType, status)

	if cachedList, err := cache.Get(ctx, cacheKey); err == nil {
		var cachedData struct {
			Total   int                 `json:"total"`
			Page    int                 `json:"page"`
			PerPage int                 `json:"per_page"`
			Data    []models.SKUWithSPU `json:"data"`
		}
		if err := json.Unmarshal([]byte(cachedList), &cachedData); err == nil {
			c.JSON(http.StatusOK, cachedData)
			return
		}
	}

	offset := (pageNum - 1) * perPageNum
	db := config.GetDB()

	query := `SELECT s.id, s.spu_id, s.sku_code, s.merchant_id, s.sku_type, s.token_amount, s.compute_points, 
		s.subscription_period, s.is_unlimited, s.fair_use_limit, s.tpm_limit, s.rpm_limit, s.concurrent_requests,
		s.valid_days, s.retail_price, s.wholesale_price, s.original_price, s.stock, s.daily_limit,
		s.group_enabled, s.min_group_size, s.max_group_size, s.group_discount_rate,
		s.is_trial, s.trial_duration_days, s.status, s.is_promoted, s.sales_count, s.created_at, s.updated_at,
		sp.name as spu_name, sp.model_provider, sp.model_name, sp.model_tier
		FROM skus s JOIN spus sp ON s.spu_id = sp.id WHERE 1=1`
	args := []interface{}{}
	argPos := 1

	if status != "" && status != "all" {
		query += " AND s.status = $" + strconv.Itoa(argPos)
		args = append(args, status)
		argPos++
	}
	if spuID != "" {
		query += " AND s.spu_id = $" + strconv.Itoa(argPos)
		args = append(args, idToInt(spuID))
		argPos++
	}
	if skuType != "" {
		query += " AND s.sku_type = $" + strconv.Itoa(argPos)
		args = append(args, skuType)
		argPos++
	}

	query += " ORDER BY s.created_at DESC LIMIT $" + strconv.Itoa(argPos) + " OFFSET $" + strconv.Itoa(argPos+1)
	args = append(args, perPageNum, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	var skus []models.SKUWithSPU
	for rows.Next() {
		var s models.SKUWithSPU
		var merchantID sql.NullInt64
		var tokenAmount sql.NullInt64
		var computePoints sql.NullFloat64
		var subscriptionPeriod sql.NullString
		var fairUseLimit sql.NullInt64
		var tpmLimit, rpmLimit, concurrentReqs sql.NullInt64
		var wholesalePrice, originalPrice sql.NullFloat64
		var dailyLimit sql.NullInt64
		var groupDiscountRate sql.NullFloat64
		var trialDurationDays sql.NullInt64

		err := rows.Scan(&s.ID, &s.SPUID, &s.SKUCode, &merchantID, &s.SKUType, &tokenAmount, &computePoints,
			&subscriptionPeriod, &s.IsUnlimited, &fairUseLimit, &tpmLimit, &rpmLimit, &concurrentReqs,
			&s.ValidDays, &s.RetailPrice, &wholesalePrice, &originalPrice, &s.Stock, &dailyLimit,
			&s.GroupEnabled, &s.MinGroupSize, &s.MaxGroupSize, &groupDiscountRate,
			&s.IsTrial, &trialDurationDays, &s.Status, &s.IsPromoted, &s.SalesCount, &s.CreatedAt, &s.UpdatedAt,
			&s.SPUName, &s.ModelProvider, &s.ModelName, &s.ModelTier)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}

		if merchantID.Valid {
			mid := int(merchantID.Int64)
			s.MerchantID = &mid
		}
		if tokenAmount.Valid {
			s.TokenAmount = tokenAmount.Int64
		}
		if computePoints.Valid {
			s.ComputePoints = computePoints.Float64
		}
		if subscriptionPeriod.Valid {
			s.SubscriptionPeriod = subscriptionPeriod.String
		}
		if fairUseLimit.Valid {
			s.FairUseLimit = fairUseLimit.Int64
		}
		if tpmLimit.Valid {
			s.TPMLimit = int(tpmLimit.Int64)
		}
		if rpmLimit.Valid {
			s.RPMLimit = int(rpmLimit.Int64)
		}
		if concurrentReqs.Valid {
			s.ConcurrentReqs = int(concurrentReqs.Int64)
		}
		if wholesalePrice.Valid {
			s.WholesalePrice = wholesalePrice.Float64
		}
		if originalPrice.Valid {
			s.OriginalPrice = originalPrice.Float64
		}
		if dailyLimit.Valid {
			s.DailyLimit = int(dailyLimit.Int64)
		}
		if groupDiscountRate.Valid {
			s.GroupDiscountRate = groupDiscountRate.Float64
		}
		if trialDurationDays.Valid {
			s.TrialDurationDays = int(trialDurationDays.Int64)
		}

		skus = append(skus, s)
	}

	countQuery := "SELECT COUNT(*) FROM skus s WHERE 1=1"
	countArgs := []interface{}{}

	if status != "" && status != "all" {
		countQuery += " AND s.status = $" + strconv.Itoa(len(countArgs)+1)
		countArgs = append(countArgs, status)
	}
	if spuID != "" {
		countQuery += " AND s.spu_id = $" + strconv.Itoa(len(countArgs)+1)
		countArgs = append(countArgs, idToInt(spuID))
	}
	if skuType != "" {
		countQuery += " AND s.sku_type = $" + strconv.Itoa(len(countArgs)+1)
		countArgs = append(countArgs, skuType)
	}

	var total int
	db.QueryRow(countQuery, countArgs...).Scan(&total)

	result := gin.H{
		"total":    total,
		"page":     pageNum,
		"per_page": perPageNum,
		"data":     skus,
	}

	if resultJSON, err := json.Marshal(result); err == nil {
		cache.Set(ctx, cacheKey, string(resultJSON), cache.ProductListTTL)
	}

	c.JSON(http.StatusOK, result)
}

func GetSKUByID(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()

	skuID := idToInt(id)
	if skuID <= 0 {
		middleware.RespondWithError(c, apperrors.ErrProductNotFound)
		return
	}

	cacheKey := cache.SKUKey(skuID)
	if cachedSKU, err := cache.Get(ctx, cacheKey); err == nil {
		var sku models.SKUWithSPU
		if err := json.Unmarshal([]byte(cachedSKU), &sku); err == nil {
			c.JSON(http.StatusOK, gin.H{"data": sku})
			return
		}
	}

	db := config.GetDB()
	var s models.SKUWithSPU
	var merchantID sql.NullInt64
	var tokenAmount sql.NullInt64
	var computePoints sql.NullFloat64
	var subscriptionPeriod sql.NullString
	var fairUseLimit sql.NullInt64
	var tpmLimit, rpmLimit, concurrentReqs sql.NullInt64
	var wholesalePrice, originalPrice sql.NullFloat64
	var dailyLimit sql.NullInt64
	var groupDiscountRate sql.NullFloat64
	var trialDurationDays sql.NullInt64

	err := db.QueryRow(
		`SELECT s.id, s.spu_id, s.sku_code, s.merchant_id, s.sku_type, s.token_amount, s.compute_points, 
		s.subscription_period, s.is_unlimited, s.fair_use_limit, s.tpm_limit, s.rpm_limit, s.concurrent_requests,
		s.valid_days, s.retail_price, s.wholesale_price, s.original_price, s.stock, s.daily_limit,
		s.group_enabled, s.min_group_size, s.max_group_size, s.group_discount_rate,
		s.is_trial, s.trial_duration_days, s.status, s.is_promoted, s.sales_count, s.created_at, s.updated_at,
		sp.name as spu_name, sp.model_provider, sp.model_name, sp.model_tier
		FROM skus s JOIN spus sp ON s.spu_id = sp.id WHERE s.id = $1`,
		skuID,
	).Scan(&s.ID, &s.SPUID, &s.SKUCode, &merchantID, &s.SKUType, &tokenAmount, &computePoints,
		&subscriptionPeriod, &s.IsUnlimited, &fairUseLimit, &tpmLimit, &rpmLimit, &concurrentReqs,
		&s.ValidDays, &s.RetailPrice, &wholesalePrice, &originalPrice, &s.Stock, &dailyLimit,
		&s.GroupEnabled, &s.MinGroupSize, &s.MaxGroupSize, &groupDiscountRate,
		&s.IsTrial, &trialDurationDays, &s.Status, &s.IsPromoted, &s.SalesCount, &s.CreatedAt, &s.UpdatedAt,
		&s.SPUName, &s.ModelProvider, &s.ModelName, &s.ModelTier)

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrProductNotFound)
		return
	}

	if merchantID.Valid {
		mid := int(merchantID.Int64)
		s.MerchantID = &mid
	}
	if tokenAmount.Valid {
		s.TokenAmount = tokenAmount.Int64
	}
	if computePoints.Valid {
		s.ComputePoints = computePoints.Float64
	}
	if subscriptionPeriod.Valid {
		s.SubscriptionPeriod = subscriptionPeriod.String
	}
	if fairUseLimit.Valid {
		s.FairUseLimit = fairUseLimit.Int64
	}
	if tpmLimit.Valid {
		s.TPMLimit = int(tpmLimit.Int64)
	}
	if rpmLimit.Valid {
		s.RPMLimit = int(rpmLimit.Int64)
	}
	if concurrentReqs.Valid {
		s.ConcurrentReqs = int(concurrentReqs.Int64)
	}
	if wholesalePrice.Valid {
		s.WholesalePrice = wholesalePrice.Float64
	}
	if originalPrice.Valid {
		s.OriginalPrice = originalPrice.Float64
	}
	if dailyLimit.Valid {
		s.DailyLimit = int(dailyLimit.Int64)
	}
	if groupDiscountRate.Valid {
		s.GroupDiscountRate = groupDiscountRate.Float64
	}
	if trialDurationDays.Valid {
		s.TrialDurationDays = int(trialDurationDays.Int64)
	}

	if skuJSON, err := json.Marshal(s); err == nil {
		cache.Set(ctx, cacheKey, string(skuJSON), cache.ProductCacheTTL)
	}

	c.JSON(http.StatusOK, gin.H{"data": s})
}

func CreateSKU(c *gin.Context) {
	var req models.SKUCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	if req.RetailPrice <= 0 {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"INVALID_PRICE",
			"价格必须大于0",
			http.StatusBadRequest,
			nil,
		))
		return
	}

	if req.Status == "" {
		req.Status = "active"
	}
	if req.Stock == 0 {
		req.Stock = -1
	}
	if req.ValidDays == 0 {
		req.ValidDays = 365
	}

	db := config.GetDB()

	var merchantID int
	err := db.QueryRow(`SELECT merchant_id FROM spus WHERE id = $1 AND status = 'active'`, req.SPUID).Scan(&merchantID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"SPU_NOT_FOUND",
			"SPU不存在或已下架",
			http.StatusBadRequest,
			err,
		))
		return
	}

	var sku models.SKU
	err = db.QueryRow(
		`INSERT INTO skus (spu_id, sku_code, merchant_id, sku_type, token_amount, compute_points, 
		 subscription_period, is_unlimited, fair_use_limit, tpm_limit, rpm_limit, concurrent_requests,
		 valid_days, retail_price, wholesale_price, original_price, stock, daily_limit,
		 group_enabled, min_group_size, max_group_size, group_discount_rate,
		 is_trial, trial_duration_days, status, is_promoted) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26) 
		 RETURNING id, spu_id, sku_code, sku_type, retail_price, stock, status, created_at, updated_at`,
		req.SPUID, req.SKUCode, merchantID, req.SKUType, req.TokenAmount, req.ComputePoints,
		req.SubscriptionPeriod, req.IsUnlimited, req.FairUseLimit, req.TPMLimit, req.RPMLimit, req.ConcurrentReqs,
		req.ValidDays, req.RetailPrice, req.WholesalePrice, req.OriginalPrice, req.Stock, req.DailyLimit,
		req.GroupEnabled, req.MinGroupSize, req.MaxGroupSize, req.GroupDiscountRate,
		req.IsTrial, req.TrialDurationDays, req.Status, req.IsPromoted,
	).Scan(&sku.ID, &sku.SPUID, &sku.SKUCode, &sku.SKUType, &sku.RetailPrice, &sku.Stock, &sku.Status, &sku.CreatedAt, &sku.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"SKU_CREATION_FAILED",
			"创建SKU失败",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	ctx := context.Background()
	cache.InvalidatePatterns(ctx, "skus:list:*")

	c.JSON(http.StatusCreated, gin.H{"data": sku})
}

func UpdateSKU(c *gin.Context) {
	id := c.Param("id")
	skuID := idToInt(id)
	if skuID <= 0 {
		middleware.RespondWithError(c, apperrors.ErrProductNotFound)
		return
	}

	var req models.SKUUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()
	var sku models.SKU
	err := db.QueryRow(
		`UPDATE skus SET 
		 retail_price = CASE WHEN $1 > 0 THEN $1 ELSE retail_price END,
		 wholesale_price = CASE WHEN $2 > 0 THEN $2 ELSE wholesale_price END,
		 original_price = CASE WHEN $3 > 0 THEN $3 ELSE original_price END,
		 stock = CASE WHEN $4 != 0 THEN $4 ELSE stock END,
		 daily_limit = CASE WHEN $5 > 0 THEN $5 ELSE daily_limit END,
		 group_enabled = $6,
		 min_group_size = CASE WHEN $7 > 0 THEN $7 ELSE min_group_size END,
		 max_group_size = CASE WHEN $8 > 0 THEN $8 ELSE max_group_size END,
		 group_discount_rate = CASE WHEN $9 > 0 THEN $9 ELSE group_discount_rate END,
		 status = COALESCE(NULLIF($10, ''), status),
		 is_promoted = $11
		 WHERE id = $12 
		 RETURNING id, spu_id, sku_code, sku_type, retail_price, stock, status, created_at, updated_at`,
		req.RetailPrice, req.WholesalePrice, req.OriginalPrice, req.Stock, req.DailyLimit,
		req.GroupEnabled, req.MinGroupSize, req.MaxGroupSize, req.GroupDiscountRate,
		req.Status, req.IsPromoted, skuID,
	).Scan(&sku.ID, &sku.SPUID, &sku.SKUCode, &sku.SKUType, &sku.RetailPrice, &sku.Stock, &sku.Status, &sku.CreatedAt, &sku.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"SKU_UPDATE_FAILED",
			"更新SKU失败",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	ctx := context.Background()
	cache.Delete(ctx, cache.SKUKey(skuID))
	cache.InvalidatePatterns(ctx, "skus:list:*")

	c.JSON(http.StatusOK, gin.H{"data": sku})
}

func DeleteSKU(c *gin.Context) {
	id := c.Param("id")
	skuID := idToInt(id)
	if skuID <= 0 {
		middleware.RespondWithError(c, apperrors.ErrProductNotFound)
		return
	}

	db := config.GetDB()
	result, err := db.Exec("DELETE FROM skus WHERE id = $1", skuID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"SKU_DELETE_FAILED",
			"删除SKU失败",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		middleware.RespondWithError(c, apperrors.ErrProductNotFound)
		return
	}

	ctx := context.Background()
	cache.Delete(ctx, cache.SKUKey(skuID))
	cache.InvalidatePatterns(ctx, "skus:list:*")

	c.JSON(http.StatusOK, gin.H{"message": "SKU deleted successfully"})
}

func GetComputePointBalance(c *gin.Context) {
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
	var account models.ComputePointAccount
	err := db.QueryRow(
		"SELECT id, user_id, balance, total_earned, total_used, COALESCE(total_expired, 0), created_at, updated_at FROM compute_point_accounts WHERE user_id = $1",
		userIDInt,
	).Scan(&account.ID, &account.UserID, &account.Balance, &account.TotalEarned, &account.TotalUsed, &account.TotalExpired, &account.CreatedAt, &account.UpdatedAt)

	if err == sql.ErrNoRows {
		_, err = db.Exec("INSERT INTO compute_point_accounts (user_id, balance) VALUES ($1, 0)", userIDInt)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		account = models.ComputePointAccount{
			UserID:  userIDInt,
			Balance: 0,
		}
	} else if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": models.ComputePointBalanceResponse{
		Balance:      account.Balance,
		TotalEarned:  account.TotalEarned,
		TotalUsed:    account.TotalUsed,
		TotalExpired: account.TotalExpired,
	}})
}

func GetComputePointTransactions(c *gin.Context) {
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

	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("per_page", "20")

	pageNum, _ := strconv.Atoi(page)
	perPageNum, _ := strconv.Atoi(perPage)

	if pageNum < 1 {
		pageNum = 1
	}
	if perPageNum < 1 || perPageNum > 100 {
		perPageNum = 20
	}

	offset := (pageNum - 1) * perPageNum
	db := config.GetDB()

	rows, err := db.Query(
		`SELECT id, user_id, type, amount, balance_after, order_id, sku_id, description, created_at 
		 FROM compute_point_transactions WHERE user_id = $1 
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userIDInt, perPageNum, offset,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	var transactions []models.ComputePointTransaction
	for rows.Next() {
		var t models.ComputePointTransaction
		var orderID, skuID sql.NullInt64
		var description sql.NullString

		err := rows.Scan(&t.ID, &t.UserID, &t.Type, &t.Amount, &t.BalanceAfter, &orderID, &skuID, &description, &t.CreatedAt)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}

		if orderID.Valid {
			oid := int(orderID.Int64)
			t.OrderID = &oid
		}
		if skuID.Valid {
			sid := int(skuID.Int64)
			t.SKUID = &sid
		}
		if description.Valid {
			t.Description = description.String
		}

		transactions = append(transactions, t)
	}

	var total int
	db.QueryRow("SELECT COUNT(*) FROM compute_point_transactions WHERE user_id = $1", userIDInt).Scan(&total)

	c.JSON(http.StatusOK, gin.H{
		"total":    total,
		"page":     pageNum,
		"per_page": perPageNum,
		"data":     transactions,
	})
}

func GetUserSubscriptions(c *gin.Context) {
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
	rows, err := db.Query(
		`SELECT us.id, us.user_id, us.sku_id, us.start_date, us.end_date, us.used_tokens, us.used_compute_points, 
		 us.status, us.auto_renew, us.created_at, us.updated_at,
		 s.sku_code, sp.name as spu_name, s.retail_price
		 FROM user_subscriptions us 
		 JOIN skus s ON us.sku_id = s.id 
		 JOIN spus sp ON s.spu_id = sp.id
		 WHERE us.user_id = $1 AND us.status = 'active'
		 ORDER BY us.end_date ASC`,
		userIDInt,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	var subscriptions []models.UserSubscriptionWithSKU
	for rows.Next() {
		var s models.UserSubscriptionWithSKU
		err := rows.Scan(&s.ID, &s.UserID, &s.SKUID, &s.StartDate, &s.EndDate, &s.UsedTokens, &s.UsedComputePoints,
			&s.Status, &s.AutoRenew, &s.CreatedAt, &s.UpdatedAt, &s.SKUCode, &s.SPUName, &s.RetailPrice)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		subscriptions = append(subscriptions, s)
	}

	c.JSON(http.StatusOK, gin.H{"data": subscriptions})
}

func GetModelProviders(c *gin.Context) {
	ctx := context.Background()
	cacheKey := "model_providers:all"

	if cachedProviders, err := cache.Get(ctx, cacheKey); err == nil {
		var providers []models.ModelProvider
		if err := json.Unmarshal([]byte(cachedProviders), &providers); err == nil {
			c.JSON(http.StatusOK, gin.H{"data": providers})
			return
		}
	}

	db := config.GetDB()
	rows, err := db.Query(
		"SELECT id, code, name, api_base_url, api_format, billing_type, cache_enabled, COALESCE(cache_discount_rate, 0), status, sort_order, created_at, updated_at FROM model_providers WHERE status = 'active' ORDER BY sort_order ASC",
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	var providers []models.ModelProvider
	for rows.Next() {
		var p models.ModelProvider
		var apiBaseURL, billingType sql.NullString
		err := rows.Scan(&p.ID, &p.Code, &p.Name, &apiBaseURL, &p.APIFormat, &billingType, &p.CacheEnabled, &p.CacheDiscount, &p.Status, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		if apiBaseURL.Valid {
			p.APIBaseURL = apiBaseURL.String
		}
		if billingType.Valid {
			p.BillingType = billingType.String
		}
		providers = append(providers, p)
	}

	if providersJSON, err := json.Marshal(providers); err == nil {
		cache.Set(ctx, cacheKey, string(providersJSON), 30*60)
	}

	c.JSON(http.StatusOK, gin.H{"data": providers})
}

func ListPublicSKUs(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("per_page", "20")
	spuID := c.Query("spu_id")
	skuType := c.Query("type")

	pageNum, _ := strconv.Atoi(page)
	perPageNum, _ := strconv.Atoi(perPage)

	if pageNum < 1 {
		pageNum = 1
	}
	if perPageNum < 1 || perPageNum > 100 {
		perPageNum = 20
	}

	offset := (pageNum - 1) * perPageNum
	db := config.GetDB()

	query := `SELECT s.id, s.spu_id, s.sku_code, s.sku_type, s.token_amount, s.compute_points, 
		s.subscription_period, s.is_unlimited, s.valid_days, s.retail_price, s.original_price, 
		s.group_enabled, s.min_group_size, s.max_group_size, s.group_discount_rate,
		s.is_trial, s.trial_duration_days, s.is_promoted, s.sales_count,
		sp.name as spu_name, sp.model_provider, sp.model_name, sp.model_tier
		FROM skus s JOIN spus sp ON s.spu_id = sp.id WHERE s.status = 'active' AND sp.status = 'active'`
	args := []interface{}{}
	argPos := 1

	if spuID != "" {
		query += " AND s.spu_id = $" + strconv.Itoa(argPos)
		args = append(args, idToInt(spuID))
		argPos++
	}
	if skuType != "" {
		query += " AND s.sku_type = $" + strconv.Itoa(argPos)
		args = append(args, skuType)
		argPos++
	}

	query += " ORDER BY s.is_promoted DESC, s.sales_count DESC, s.created_at DESC LIMIT $" + strconv.Itoa(argPos) + " OFFSET $" + strconv.Itoa(argPos+1)
	args = append(args, perPageNum, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	var skus []models.SKUWithSPU
	for rows.Next() {
		var s models.SKUWithSPU
		var tokenAmount sql.NullInt64
		var computePoints sql.NullFloat64
		var subscriptionPeriod sql.NullString
		var originalPrice sql.NullFloat64
		var groupDiscountRate sql.NullFloat64
		var trialDurationDays sql.NullInt64

		err := rows.Scan(&s.ID, &s.SPUID, &s.SKUCode, &s.SKUType, &tokenAmount, &computePoints,
			&subscriptionPeriod, &s.IsUnlimited, &s.ValidDays, &s.RetailPrice, &originalPrice,
			&s.GroupEnabled, &s.MinGroupSize, &s.MaxGroupSize, &groupDiscountRate,
			&s.IsTrial, &trialDurationDays, &s.IsPromoted, &s.SalesCount,
			&s.SPUName, &s.ModelProvider, &s.ModelName, &s.ModelTier)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}

		if tokenAmount.Valid {
			s.TokenAmount = tokenAmount.Int64
		}
		if computePoints.Valid {
			s.ComputePoints = computePoints.Float64
		}
		if subscriptionPeriod.Valid {
			s.SubscriptionPeriod = subscriptionPeriod.String
		}
		if originalPrice.Valid {
			s.OriginalPrice = originalPrice.Float64
		}
		if groupDiscountRate.Valid {
			s.GroupDiscountRate = groupDiscountRate.Float64
		}
		if trialDurationDays.Valid {
			s.TrialDurationDays = int(trialDurationDays.Int64)
		}

		skus = append(skus, s)
	}

	var total int
	db.QueryRow("SELECT COUNT(*) FROM skus s JOIN spus sp ON s.spu_id = sp.id WHERE s.status = 'active' AND sp.status = 'active'").Scan(&total)

	c.JSON(http.StatusOK, gin.H{
		"total":    total,
		"page":     pageNum,
		"per_page": perPageNum,
		"data":     skus,
	})
}

func GetPublicSKUByID(c *gin.Context) {
	id := c.Param("id")
	skuID := idToInt(id)
	if skuID <= 0 {
		middleware.RespondWithError(c, apperrors.ErrProductNotFound)
		return
	}

	db := config.GetDB()
	var s models.SKUWithSPU
	var tokenAmount sql.NullInt64
	var computePoints sql.NullFloat64
	var subscriptionPeriod sql.NullString
	var originalPrice sql.NullFloat64
	var groupDiscountRate sql.NullFloat64
	var trialDurationDays sql.NullInt64

	err := db.QueryRow(
		`SELECT s.id, s.spu_id, s.sku_code, s.sku_type, s.token_amount, s.compute_points, 
		s.subscription_period, s.is_unlimited, s.valid_days, s.retail_price, s.original_price, 
		s.group_enabled, s.min_group_size, s.max_group_size, s.group_discount_rate,
		s.is_trial, s.trial_duration_days, s.is_promoted, s.sales_count,
		sp.name as spu_name, sp.model_provider, sp.model_name, sp.model_tier
		FROM skus s JOIN spus sp ON s.spu_id = sp.id 
		WHERE s.id = $1 AND s.status = 'active' AND sp.status = 'active'`,
		skuID,
	).Scan(&s.ID, &s.SPUID, &s.SKUCode, &s.SKUType, &tokenAmount, &computePoints,
		&subscriptionPeriod, &s.IsUnlimited, &s.ValidDays, &s.RetailPrice, &originalPrice,
		&s.GroupEnabled, &s.MinGroupSize, &s.MaxGroupSize, &groupDiscountRate,
		&s.IsTrial, &trialDurationDays, &s.IsPromoted, &s.SalesCount,
		&s.SPUName, &s.ModelProvider, &s.ModelName, &s.ModelTier)

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrProductNotFound)
		return
	}

	if tokenAmount.Valid {
		s.TokenAmount = tokenAmount.Int64
	}
	if computePoints.Valid {
		s.ComputePoints = computePoints.Float64
	}
	if subscriptionPeriod.Valid {
		s.SubscriptionPeriod = subscriptionPeriod.String
	}
	if originalPrice.Valid {
		s.OriginalPrice = originalPrice.Float64
	}
	if groupDiscountRate.Valid {
		s.GroupDiscountRate = groupDiscountRate.Float64
	}
	if trialDurationDays.Valid {
		s.TrialDurationDays = int(trialDurationDays.Int64)
	}

	c.JSON(http.StatusOK, gin.H{"data": s})
}
