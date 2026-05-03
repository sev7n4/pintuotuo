package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/pintuotuo/backend/cache"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/logger"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/models"
	"github.com/pintuotuo/backend/services"
)

const (
	adminSKUListScopeSellable = "sellable"
	adminSKUListScopeAll      = "all"
)

const modelProviderFallbackCode = "__default__"

func ensureAdmin(c *gin.Context) bool {
	userRole, exists := c.Get("user_role")
	if !exists || userRole != roleAdmin {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"FORBIDDEN",
			"Admin access required",
			http.StatusForbidden,
			nil,
		))
		return false
	}
	return true
}

func ensureMerchant(c *gin.Context) bool {
	userRole, exists := c.Get("user_role")
	if !exists || userRole != roleMerchant {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"FORBIDDEN",
			"Merchant access required",
			http.StatusForbidden,
			nil,
		))
		return false
	}
	return true
}

// loadActiveModelProviders returns model_providers rows with status=active, using the same cache key as admin GetModelProviders.
func loadActiveModelProviders() ([]models.ModelProvider, error) {
	ctx := context.Background()
	cacheKey := "model_providers:all"

	if cachedProviders, err := cache.Get(ctx, cacheKey); err == nil {
		var providers []models.ModelProvider
		if err := json.Unmarshal([]byte(cachedProviders), &providers); err == nil {
			return providers, nil
		}
	}

	db := config.GetDB()
	if db == nil {
		return nil, errors.New("database not available")
	}
	rows, err := db.Query(
		`SELECT id, code, name, api_base_url, api_format, billing_type, cache_enabled, COALESCE(cache_discount_rate, 0),
			compat_prefixes,
			litellm_model_template, litellm_gateway_api_key_env, litellm_gateway_api_base,
			status, sort_order, created_at, updated_at
		 FROM model_providers WHERE status = 'active' ORDER BY sort_order ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []models.ModelProvider
	for rows.Next() {
		var p models.ModelProvider
		var apiBaseURL, billingType sql.NullString
		var litellmTpl, litellmKey, litellmBase sql.NullString
		var compatPfx pq.StringArray
		err := rows.Scan(&p.ID, &p.Code, &p.Name, &apiBaseURL, &p.APIFormat, &billingType, &p.CacheEnabled, &p.CacheDiscount,
			&compatPfx, &litellmTpl, &litellmKey, &litellmBase, &p.Status, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if apiBaseURL.Valid {
			p.APIBaseURL = apiBaseURL.String
		}
		if billingType.Valid {
			p.BillingType = billingType.String
		}
		if litellmTpl.Valid {
			p.LitellmModelTemplate = litellmTpl.String
		}
		if litellmKey.Valid {
			p.LitellmGatewayAPIKeyEnv = litellmKey.String
		}
		if litellmBase.Valid {
			p.LitellmGatewayAPIBase = litellmBase.String
		}
		if len(compatPfx) > 0 {
			p.CompatPrefixes = []string(compatPfx)
		}
		providers = append(providers, p)
	}

	if providersJSON, err := json.Marshal(providers); err == nil {
		cache.Set(ctx, cacheKey, string(providersJSON), 30*60)
	}

	return providers, nil
}

// adminSPUListFilters builds WHERE for admin SPU list (table alias p).
func adminSPUListFilters(status, provider, tier, q string) (where string, args []interface{}) {
	parts := []string{"1=1"}
	args = []interface{}{}
	n := 1
	if status != "" && status != allProductStatus {
		parts = append(parts, fmt.Sprintf("p.status = $%d", n))
		args = append(args, status)
		n++
	}
	if provider != "" {
		parts = append(parts, fmt.Sprintf("p.model_provider = $%d", n))
		args = append(args, provider)
		n++
	}
	if tier != "" {
		parts = append(parts, fmt.Sprintf("p.model_tier = $%d", n))
		args = append(args, tier)
		n++
	}
	if q != "" {
		pat := "%" + q + "%"
		parts = append(parts, fmt.Sprintf("(p.spu_code ILIKE $%d OR p.name ILIKE $%d)", n, n+1))
		args = append(args, pat, pat)
	}
	return strings.Join(parts, " AND "), args
}

func ListSPUs(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}

	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("per_page", "20")
	provider := c.Query("provider")
	tier := c.Query("tier")
	status := c.DefaultQuery("status", "active")
	q := strings.TrimSpace(c.Query("q"))

	pageNum, _ := strconv.Atoi(page)
	perPageNum, _ := strconv.Atoi(perPage)

	if pageNum < 1 {
		pageNum = 1
	}
	if perPageNum < 1 || perPageNum > 100 {
		perPageNum = 20
	}

	ctx := context.Background()
	cacheKey := cache.SPUListKey(pageNum, perPageNum, provider, tier, status, q)

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

	whereClause, baseArgs := adminSPUListFilters(status, provider, tier, q)
	limPos := len(baseArgs) + 1
	offPos := len(baseArgs) + 2
	query := fmt.Sprintf(
		`SELECT p.id, p.spu_code, p.name, p.model_provider, p.model_name, p.model_version, p.model_tier, p.context_window, p.base_compute_points,
			COALESCE(p.provider_input_rate, 0), COALESCE(p.provider_output_rate, 0), p.description, p.status, p.sort_order, p.total_sales_count, COALESCE(p.average_rating, 0),
			(SELECT COUNT(*) FROM skus s WHERE s.spu_id = p.id),
			(SELECT COUNT(*) FROM skus s WHERE s.spu_id = p.id AND s.status = 'active'),
			p.created_at, p.updated_at
		 FROM spus p WHERE %s ORDER BY p.sort_order ASC, p.created_at DESC LIMIT $%d OFFSET $%d`,
		whereClause, limPos, offPos,
	)
	args := append(append([]interface{}{}, baseArgs...), perPageNum, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	var spus []models.SPU
	for rows.Next() {
		var s models.SPU
		err := rows.Scan(&s.ID, &s.SPUCode, &s.Name, &s.ModelProvider, &s.ModelName, &s.ModelVersion, &s.ModelTier, &s.ContextWindow, &s.BaseComputePoints, &s.ProviderInputRate, &s.ProviderOutputRate, &s.Description, &s.Status, &s.SortOrder, &s.TotalSalesCount, &s.AverageRating, &s.SkuCount, &s.ActiveSkuCount, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		spus = append(spus, s)
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM spus p WHERE %s", whereClause)
	var total int
	db.QueryRow(countQuery, baseArgs...).Scan(&total)

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
	if !ensureAdmin(c) {
		return
	}

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
		`SELECT id, spu_code, name, model_provider, model_name, model_version, model_tier, context_window, max_output_tokens, base_compute_points,
			COALESCE(provider_input_rate, 0), COALESCE(provider_output_rate, 0), description, thumbnail_url, status, sort_order, total_sales_count, COALESCE(average_rating, 0),
			(SELECT COUNT(*) FROM skus s WHERE s.spu_id = spus.id),
			(SELECT COUNT(*) FROM skus s WHERE s.spu_id = spus.id AND s.status = 'active'),
			created_at, updated_at FROM spus WHERE id = $1`,
		spuID,
	).Scan(&spu.ID, &spu.SPUCode, &spu.Name, &spu.ModelProvider, &spu.ModelName, &spu.ModelVersion, &spu.ModelTier, &spu.ContextWindow, &spu.MaxOutputTokens, &spu.BaseComputePoints, &spu.ProviderInputRate, &spu.ProviderOutputRate, &spu.Description, &spu.ThumbnailURL, &spu.Status, &spu.SortOrder, &spu.TotalSalesCount, &spu.AverageRating, &spu.SkuCount, &spu.ActiveSkuCount, &spu.CreatedAt, &spu.UpdatedAt)

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
	if !ensureAdmin(c) {
		return
	}

	var req models.SPUCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	if req.BaseComputePoints <= 0 {
		req.BaseComputePoints = 1.0
	}
	if req.Status == "" {
		req.Status = merchantStatusActive
	}

	db := config.GetDB()
	var spu models.SPU
	err := db.QueryRow(
		`INSERT INTO spus (spu_code, name, model_provider, model_name, model_version, model_tier, context_window, max_output_tokens, base_compute_points, provider_input_rate, provider_output_rate, description, thumbnail_url, status, sort_order) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) 
		 RETURNING id, spu_code, name, model_provider, model_name, model_version, model_tier, context_window, base_compute_points, provider_input_rate, provider_output_rate, description, status, sort_order, created_at, updated_at`,
		req.SPUCode, req.Name, req.ModelProvider, req.ModelName, req.ModelVersion, req.ModelTier, req.ContextWindow, req.MaxOutputTokens, req.BaseComputePoints, req.ProviderInputRate, req.ProviderOutputRate, req.Description, req.ThumbnailURL, req.Status, req.SortOrder,
	).Scan(&spu.ID, &spu.SPUCode, &spu.Name, &spu.ModelProvider, &spu.ModelName, &spu.ModelVersion, &spu.ModelTier, &spu.ContextWindow, &spu.BaseComputePoints, &spu.ProviderInputRate, &spu.ProviderOutputRate, &spu.Description, &spu.Status, &spu.SortOrder, &spu.CreatedAt, &spu.UpdatedAt)

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
	if !ensureAdmin(c) {
		return
	}

	id := c.Param("id")
	spuID := idToInt(id)
	if spuID <= 0 {
		middleware.RespondWithError(c, apperrors.ErrProductNotFound)
		return
	}

	var req models.SPUUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()

	var name, modelProvider, modelName, modelVersion, modelTier, description, status interface{}
	var contextWindow, maxOutputTokens interface{}
	var baseComputePoints, providerInputRate, providerOutputRate interface{}
	var sortOrder interface{}

	if req.Name != nil {
		name = *req.Name
	}
	if req.ModelProvider != nil {
		modelProvider = *req.ModelProvider
	}
	if req.ModelName != nil {
		modelName = *req.ModelName
	}
	if req.ModelVersion != nil {
		modelVersion = *req.ModelVersion
	}
	if req.ModelTier != nil {
		modelTier = *req.ModelTier
	}
	if req.ContextWindow != nil {
		contextWindow = *req.ContextWindow
	}
	if req.MaxOutputTokens != nil {
		maxOutputTokens = *req.MaxOutputTokens
	}
	if req.BaseComputePoints != nil {
		baseComputePoints = *req.BaseComputePoints
	}
	if req.Description != nil {
		description = *req.Description
	}
	if req.ProviderInputRate != nil {
		providerInputRate = *req.ProviderInputRate
	}
	if req.ProviderOutputRate != nil {
		providerOutputRate = *req.ProviderOutputRate
	}
	if req.Status != nil {
		status = *req.Status
	}
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}

	var spu models.SPU
	err := db.QueryRow(
		`UPDATE spus SET 
		 name = COALESCE($1, name), 
		 model_provider = COALESCE($2, model_provider),
		 model_name = COALESCE($3, model_name),
		 model_version = COALESCE($4, model_version),
		 model_tier = COALESCE($5, model_tier),
		 context_window = COALESCE($6, context_window),
		 max_output_tokens = COALESCE($7, max_output_tokens),
		 base_compute_points = COALESCE($8, base_compute_points),
		 provider_input_rate = COALESCE($9, provider_input_rate),
		 provider_output_rate = COALESCE($10, provider_output_rate),
		 description = COALESCE($11, description),
		 status = COALESCE($12, status),
		 sort_order = COALESCE($13, sort_order)
		 WHERE id = $14 
		 RETURNING id, spu_code, name, model_provider, model_name, model_version, model_tier, context_window, max_output_tokens, base_compute_points, provider_input_rate, provider_output_rate, description, status, sort_order, created_at, updated_at`,
		name, modelProvider, modelName, modelVersion, modelTier, contextWindow, maxOutputTokens, baseComputePoints, providerInputRate, providerOutputRate, description, status, sortOrder, spuID,
	).Scan(&spu.ID, &spu.SPUCode, &spu.Name, &spu.ModelProvider, &spu.ModelName, &spu.ModelVersion, &spu.ModelTier, &spu.ContextWindow, &spu.MaxOutputTokens, &spu.BaseComputePoints, &spu.ProviderInputRate, &spu.ProviderOutputRate, &spu.Description, &spu.Status, &spu.SortOrder, &spu.CreatedAt, &spu.UpdatedAt)

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
	if !ensureAdmin(c) {
		return
	}

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

// ResolveAdminSKUListScope 解析管理端 SKU 列表 scope（供单测与 ListSKUs 使用）。
// 未传 scope 且未传 status → sellable；传了 status → all；status=all 表示不按 SKU 状态过滤。
func ResolveAdminSKUListScope(scopeQuery, statusQuery string) (scope string, skuStatus string) {
	scope = strings.TrimSpace(scopeQuery)
	explicitStatus := strings.TrimSpace(statusQuery)
	statusMeansAllSKUs := explicitStatus == adminSKUListScopeAll
	if statusMeansAllSKUs {
		explicitStatus = ""
	}
	if scope == "" {
		if explicitStatus == "" && !statusMeansAllSKUs {
			scope = adminSKUListScopeSellable
		} else {
			scope = adminSKUListScopeAll
		}
	}
	if scope != adminSKUListScopeSellable && scope != adminSKUListScopeAll {
		scope = adminSKUListScopeSellable
	}
	return scope, explicitStatus
}

func resolveAdminSKUScope(c *gin.Context) (scope string, skuStatus string) {
	return ResolveAdminSKUListScope(c.Query("scope"), c.Query("status"))
}

// adminSKUListFilters builds WHERE (JOIN spus) for admin SKU list. scope=sellable → SKU+SPU 均在售（与商户可选列表一致）。
func adminSKUListFilters(scope, skuStatus, spuStatus, provider, q string, misaligned bool, spuID, skuType string) (where string, args []interface{}) {
	parts := []string{"1=1"}
	args = []interface{}{}
	n := 1

	if scope == adminSKUListScopeSellable {
		parts = append(parts, "s.status = 'active'", "sp.status = 'active'")
	} else {
		if skuStatus == "active" || skuStatus == "inactive" {
			parts = append(parts, fmt.Sprintf("s.status = $%d", n))
			args = append(args, skuStatus)
			n++
		}
		if spuStatus == "active" || spuStatus == "inactive" {
			parts = append(parts, fmt.Sprintf("sp.status = $%d", n))
			args = append(args, spuStatus)
			n++
		}
	}

	if misaligned {
		parts = append(parts, "s.status = 'active'", "sp.status = 'inactive'")
	}
	if provider != "" {
		parts = append(parts, fmt.Sprintf("sp.model_provider = $%d", n))
		args = append(args, provider)
		n++
	}
	if q != "" {
		pat := "%" + q + "%"
		parts = append(parts, fmt.Sprintf("(s.sku_code ILIKE $%d OR sp.name ILIKE $%d OR sp.spu_code ILIKE $%d)", n, n+1, n+2))
		args = append(args, pat, pat, pat)
		n += 3
	}
	if spuID != "" {
		parts = append(parts, fmt.Sprintf("s.spu_id = $%d", n))
		args = append(args, idToInt(spuID))
		n++
	}
	if skuType != "" {
		parts = append(parts, fmt.Sprintf("s.sku_type = $%d", n))
		args = append(args, skuType)
	}

	return strings.Join(parts, " AND "), args
}

func ListSKUs(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}

	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("per_page", "20")
	spuID := c.Query("spu_id")
	skuType := c.Query("type")
	scope, skuStatus := resolveAdminSKUScope(c)
	spuStatus := strings.TrimSpace(c.Query("spu_status"))
	provider := strings.TrimSpace(c.Query("provider"))
	q := strings.TrimSpace(c.Query("q"))
	misaligned := c.Query("misaligned") == "1" || strings.EqualFold(c.Query("misaligned"), "true")

	pageNum, _ := strconv.Atoi(page)
	perPageNum, _ := strconv.Atoi(perPage)

	if pageNum < 1 {
		pageNum = 1
	}
	if perPageNum < 1 || perPageNum > 100 {
		perPageNum = 20
	}

	misalignedKey := "0"
	if misaligned {
		misalignedKey = "1"
	}
	ctx := context.Background()
	cacheKey := cache.SKUListKey(pageNum, perPageNum, spuID, skuType, scope, skuStatus, spuStatus, provider, q, misalignedKey)

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

	whereClause, baseArgs := adminSKUListFilters(scope, skuStatus, spuStatus, provider, q, misaligned, spuID, skuType)
	limitPos := len(baseArgs) + 1
	offsetPos := len(baseArgs) + 2

	query := `SELECT s.id, s.spu_id, s.sku_code, s.merchant_id, s.sku_type, s.token_amount, s.compute_points, 
		s.subscription_period, s.is_unlimited, s.fair_use_limit, s.tpm_limit, s.rpm_limit, s.concurrent_requests,
		s.valid_days, s.retail_price, s.wholesale_price, s.original_price, s.stock, s.daily_limit,
		s.group_enabled, s.min_group_size, s.max_group_size, s.group_discount_rate,
		COALESCE(s.cost_input_rate, 0), COALESCE(s.cost_output_rate, 0), COALESCE(s.inherit_spu_cost, true),
		s.is_trial, s.trial_duration_days, s.status, s.is_promoted, s.sales_count, s.created_at, s.updated_at,
		sp.name as spu_name, sp.status as spu_status, sp.model_provider, sp.model_name, sp.model_tier
		FROM skus s JOIN spus sp ON s.spu_id = sp.id WHERE ` + whereClause +
		fmt.Sprintf(" ORDER BY s.created_at DESC LIMIT $%d OFFSET $%d", limitPos, offsetPos)

	listArgs := append(append([]interface{}{}, baseArgs...), perPageNum, offset)

	rows, err := db.Query(query, listArgs...)
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
			&s.CostInputRate, &s.CostOutputRate, &s.InheritSPUCost,
			&s.IsTrial, &trialDurationDays, &s.Status, &s.IsPromoted, &s.SalesCount, &s.CreatedAt, &s.UpdatedAt,
			&s.SPUName, &s.SpuStatus, &s.ModelProvider, &s.ModelName, &s.ModelTier)
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

		s.Sellable = s.Status == merchantStatusActive && s.SpuStatus == merchantStatusActive
		skus = append(skus, s)
	}

	countQuery := "SELECT COUNT(*) FROM skus s JOIN spus sp ON s.spu_id = sp.id WHERE " + whereClause
	var total int
	if err := db.QueryRow(countQuery, baseArgs...).Scan(&total); err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

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
	if !ensureAdmin(c) {
		return
	}

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
		COALESCE(s.cost_input_rate, 0), COALESCE(s.cost_output_rate, 0), COALESCE(s.inherit_spu_cost, true),
		s.is_trial, s.trial_duration_days, s.status, s.is_promoted, s.sales_count, s.created_at, s.updated_at,
		sp.name as spu_name, sp.status as spu_status, sp.model_provider, sp.model_name, sp.model_tier
		FROM skus s JOIN spus sp ON s.spu_id = sp.id WHERE s.id = $1`,
		skuID,
	).Scan(&s.ID, &s.SPUID, &s.SKUCode, &merchantID, &s.SKUType, &tokenAmount, &computePoints,
		&subscriptionPeriod, &s.IsUnlimited, &fairUseLimit, &tpmLimit, &rpmLimit, &concurrentReqs,
		&s.ValidDays, &s.RetailPrice, &wholesalePrice, &originalPrice, &s.Stock, &dailyLimit,
		&s.GroupEnabled, &s.MinGroupSize, &s.MaxGroupSize, &groupDiscountRate,
		&s.CostInputRate, &s.CostOutputRate, &s.InheritSPUCost,
		&s.IsTrial, &trialDurationDays, &s.Status, &s.IsPromoted, &s.SalesCount, &s.CreatedAt, &s.UpdatedAt,
		&s.SPUName, &s.SpuStatus, &s.ModelProvider, &s.ModelName, &s.ModelTier)

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

	s.Sellable = s.Status == merchantStatusActive && s.SpuStatus == merchantStatusActive

	if skuJSON, err := json.Marshal(s); err == nil {
		cache.Set(ctx, cacheKey, string(skuJSON), cache.ProductCacheTTL)
	}

	c.JSON(http.StatusOK, gin.H{"data": s})
}

func sqlNullableString(s string) interface{} {
	t := strings.TrimSpace(s)
	if t == "" {
		return nil
	}
	return t
}

// validateSKUCreateRequest checks type-specific required fields before INSERT.
func validateSKUCreateRequest(req *models.SKUCreateRequest) *apperrors.AppError {
	switch req.SKUType {
	case "token_pack":
		if req.TokenAmount <= 0 {
			return apperrors.NewAppError("TOKEN_AMOUNT_REQUIRED", "Token包须填写大于 0 的 Token 数量", http.StatusBadRequest, nil)
		}
		// compute_points 可选：履约仅使用 token_amount；算力点字段仅作展示/兼容，可为 0 或不填
	case "subscription":
		p := strings.TrimSpace(req.SubscriptionPeriod)
		if p == "" {
			return apperrors.NewAppError("SUBSCRIPTION_PERIOD_REQUIRED", "订阅套餐必须选择订阅周期", http.StatusBadRequest, nil)
		}
		if p != "monthly" && p != "quarterly" && p != "yearly" {
			return apperrors.NewAppError("INVALID_SUBSCRIPTION_PERIOD", "订阅周期必须是 monthly、quarterly 或 yearly", http.StatusBadRequest, nil)
		}
	case "concurrent":
		if req.ConcurrentReqs <= 0 {
			return apperrors.NewAppError("CONCURRENT_REQUESTS_REQUIRED", "并发套餐须填写大于 0 的并发请求数", http.StatusBadRequest, nil)
		}
	case "trial":
		// 试用套餐：订阅周期等可为空，由数据库存 NULL
	case "compute_points":
		if req.ComputePoints <= 0 {
			return apperrors.NewAppError("COMPUTE_POINTS_REQUIRED", "算力点套餐须填写大于 0 的算力点", http.StatusBadRequest, nil)
		}
	default:
		return apperrors.NewAppError("INVALID_SKU_TYPE", "不支持的 SKU 类型", http.StatusBadRequest, nil)
	}
	return nil
}

func CreateSKU(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}

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
		req.Status = merchantStatusActive
	}
	if req.Stock == 0 {
		req.Stock = -1
	}
	if req.ValidDays == 0 {
		req.ValidDays = 365
	}

	if vErr := validateSKUCreateRequest(&req); vErr != nil {
		middleware.RespondWithError(c, vErr)
		return
	}

	db := config.GetDB()

	var spuExists bool
	err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM spus WHERE id = $1)`, req.SPUID).Scan(&spuExists)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"SPU_QUERY_FAILED",
			"查询SPU失败",
			http.StatusInternalServerError,
			err,
		))
		return
	}
	if !spuExists {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"SPU_NOT_FOUND",
			"SPU不存在",
			http.StatusBadRequest,
			nil,
		))
		return
	}

	inheritSPUCost := req.InheritSPUCost
	if !req.InheritSPUCost && req.CostInputRate == 0 && req.CostOutputRate == 0 {
		// 未显式传值时，保持默认继承
		inheritSPUCost = true
	}
	costInputRate := req.CostInputRate
	costOutputRate := req.CostOutputRate
	if inheritSPUCost {
		err = db.QueryRow(
			`SELECT COALESCE(provider_input_rate, 0), COALESCE(provider_output_rate, 0) FROM spus WHERE id = $1`,
			req.SPUID,
		).Scan(&costInputRate, &costOutputRate)
		if err != nil {
			middleware.RespondWithError(c, apperrors.NewAppError(
				"SPU_QUERY_FAILED",
				"读取SPU参考成本失败",
				http.StatusInternalServerError,
				err,
			))
			return
		}
	}
	if !inheritSPUCost && (costInputRate <= 0 || costOutputRate <= 0) {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"SKU_COST_REQUIRED",
			"关闭继承后，SKU输入/输出成本必须大于0",
			http.StatusBadRequest,
			nil,
		))
		return
	}

	subscriptionPeriod := sqlNullableString(req.SubscriptionPeriod)

	var sku models.SKU
	err = db.QueryRow(
		`INSERT INTO skus (spu_id, sku_code, sku_type, token_amount, compute_points, 
		 subscription_period, is_unlimited, fair_use_limit, tpm_limit, rpm_limit, concurrent_requests,
		 valid_days, retail_price, wholesale_price, original_price, stock, daily_limit,
		 group_enabled, min_group_size, max_group_size, group_discount_rate,
		 cost_input_rate, cost_output_rate, inherit_spu_cost, is_trial, trial_duration_days, status, is_promoted) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28) 
		 RETURNING id, spu_id, sku_code, sku_type, retail_price, stock, status, cost_input_rate, cost_output_rate, inherit_spu_cost, created_at, updated_at`,
		req.SPUID, req.SKUCode, req.SKUType, req.TokenAmount, req.ComputePoints,
		subscriptionPeriod, req.IsUnlimited, req.FairUseLimit, req.TPMLimit, req.RPMLimit, req.ConcurrentReqs,
		req.ValidDays, req.RetailPrice, req.WholesalePrice, req.OriginalPrice, req.Stock, req.DailyLimit,
		req.GroupEnabled, req.MinGroupSize, req.MaxGroupSize, req.GroupDiscountRate,
		costInputRate, costOutputRate, inheritSPUCost, req.IsTrial, req.TrialDurationDays, req.Status, req.IsPromoted,
	).Scan(&sku.ID, &sku.SPUID, &sku.SKUCode, &sku.SKUType, &sku.RetailPrice, &sku.Stock, &sku.Status, &sku.CostInputRate, &sku.CostOutputRate, &sku.InheritSPUCost, &sku.CreatedAt, &sku.UpdatedAt)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23505":
				middleware.RespondWithError(c, apperrors.NewAppError(
					"SKU_CODE_EXISTS",
					"SKU编码已存在，请更换其他编码",
					http.StatusConflict,
					err,
				))
				return
			case "23514":
				middleware.RespondWithError(c, apperrors.NewAppErrorWithDetails(
					"SKU_CONSTRAINT_VIOLATION",
					"创建失败：数据不符合数据库规则（例如 SKU 类型与订阅周期、并发数不匹配）",
					http.StatusBadRequest,
					err,
					map[string]string{
						"constraint": pqErr.Constraint,
						"hint":       "非订阅类 SKU 请勿提交空的订阅周期；订阅类必须选择 monthly / quarterly / yearly",
					},
				))
				return
			}
		}
		middleware.RespondWithError(c, apperrors.NewAppError(
			"SKU_CREATION_FAILED",
			"创建SKU失败，请稍后重试或查看服务日志",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	ctx := context.Background()
	cache.Delete(ctx, cache.SPUKey(req.SPUID))
	cache.InvalidatePatterns(ctx, "skus:list:*")
	cache.InvalidatePatterns(ctx, "spus:list:*")

	c.JSON(http.StatusCreated, gin.H{"data": sku})
}

func UpdateSKU(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}

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
	var spuID int
	var dbInheritSPUCost bool
	var dbCostInputRate, dbCostOutputRate float64
	if err := db.QueryRow("SELECT spu_id, inherit_spu_cost, COALESCE(cost_input_rate, 0), COALESCE(cost_output_rate, 0) FROM skus WHERE id = $1", skuID).Scan(&spuID, &dbInheritSPUCost, &dbCostInputRate, &dbCostOutputRate); err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"SKU_NOT_FOUND",
			"SKU不存在",
			http.StatusNotFound,
			err,
		))
		return
	}

	var spuInputRate, spuOutputRate float64
	if err := db.QueryRow(
		"SELECT COALESCE(provider_input_rate, 0), COALESCE(provider_output_rate, 0) FROM spus WHERE id = $1",
		spuID,
	).Scan(&spuInputRate, &spuOutputRate); err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"SPU_QUERY_FAILED",
			"读取SPU参考成本失败",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	inherit, costInputRate, costOutputRate := resolveSKUUpdateInherit(req, dbInheritSPUCost, dbCostInputRate, dbCostOutputRate, spuInputRate, spuOutputRate)

	if !inherit && (costInputRate <= 0 || costOutputRate <= 0) {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"SKU_COST_REQUIRED",
			"关闭继承后，SKU输入/输出成本必须大于0",
			http.StatusBadRequest,
			nil,
		))
		return
	}
	var sku models.SKU
	subscriptionPeriod := strings.TrimSpace(req.SubscriptionPeriod)
	err := db.QueryRow(
		// 价格/折扣参数须显式 ::numeric：否则在扩展查询中「$n > 0」会与整型字面量 0 比较，导致小数绑定失败（pq: invalid input syntax for type integer: "9.9"）。
		`UPDATE skus SET 
		 retail_price = CASE WHEN $1::numeric > 0 THEN $1::numeric ELSE retail_price END,
		 wholesale_price = CASE WHEN $2::numeric > 0 THEN $2::numeric ELSE wholesale_price END,
		 original_price = CASE WHEN $3::numeric > 0 THEN $3::numeric ELSE original_price END,
		 token_amount = CASE WHEN $4 > 0 THEN $4 ELSE token_amount END,
		 compute_points = CASE WHEN $5::numeric > 0 THEN $5::numeric ELSE compute_points END,
		 subscription_period = COALESCE(NULLIF($6::text, ''), subscription_period),
		 is_unlimited = $7,
		 fair_use_limit = CASE WHEN $8 > 0 THEN $8 ELSE fair_use_limit END,
		 tpm_limit = CASE WHEN $9 > 0 THEN $9 ELSE tpm_limit END,
		 rpm_limit = CASE WHEN $10 > 0 THEN $10 ELSE rpm_limit END,
		 concurrent_requests = CASE WHEN $11 > 0 THEN $11 ELSE concurrent_requests END,
		 valid_days = CASE WHEN $12 > 0 THEN $12 ELSE valid_days END,
		 stock = CASE WHEN $13 != 0 THEN $13 ELSE stock END,
		 daily_limit = CASE WHEN $14 > 0 THEN $14 ELSE daily_limit END,
		 group_enabled = $15,
		 min_group_size = CASE WHEN $16 > 0 THEN $16 ELSE min_group_size END,
		 max_group_size = CASE WHEN $17 > 0 THEN $17 ELSE max_group_size END,
		 group_discount_rate = CASE WHEN $18::numeric > 0 THEN $18::numeric ELSE group_discount_rate END,
		 status = COALESCE(NULLIF($19, ''), status),
		 is_promoted = $20,
		 cost_input_rate = $21,
		 cost_output_rate = $22,
		 inherit_spu_cost = $23
		 WHERE id = $24 
		 RETURNING id, spu_id, sku_code, sku_type, retail_price, stock, status, valid_days, cost_input_rate, cost_output_rate, inherit_spu_cost, created_at, updated_at`,
		req.RetailPrice, req.WholesalePrice, req.OriginalPrice,
		req.TokenAmount, req.ComputePoints,
		subscriptionPeriod,
		req.IsUnlimited, req.FairUseLimit, req.TPMLimit, req.RPMLimit, req.ConcurrentReqs,
		req.ValidDays, req.Stock, req.DailyLimit,
		req.GroupEnabled, req.MinGroupSize, req.MaxGroupSize, req.GroupDiscountRate,
		req.Status, req.IsPromoted, costInputRate, costOutputRate, inherit, skuID,
	).Scan(&sku.ID, &sku.SPUID, &sku.SKUCode, &sku.SKUType, &sku.RetailPrice, &sku.Stock, &sku.Status, &sku.ValidDays, &sku.CostInputRate, &sku.CostOutputRate, &sku.InheritSPUCost, &sku.CreatedAt, &sku.UpdatedAt)

	if err != nil {
		log.Printf("[UpdateSKU] sku_id=%d: %v", skuID, err)
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
	cache.Delete(ctx, cache.SPUKey(sku.SPUID))
	cache.InvalidatePatterns(ctx, "skus:list:*")
	cache.InvalidatePatterns(ctx, "spus:list:*")

	c.JSON(http.StatusOK, gin.H{"data": sku})
}

func resolveSKUUpdateInherit(req models.SKUUpdateRequest, dbInherit bool, dbInputRate, dbOutputRate, spuInputRate, spuOutputRate float64) (inherit bool, inputRate, outputRate float64) {
	if req.InheritSPUCost == nil {
		return dbInherit, dbInputRate, dbOutputRate
	}
	if *req.InheritSPUCost {
		return true, spuInputRate, spuOutputRate
	}
	reqInputRate := req.CostInputRate
	reqOutputRate := req.CostOutputRate
	if reqInputRate <= 0 {
		reqInputRate = dbInputRate
	}
	if reqOutputRate <= 0 {
		reqOutputRate = dbOutputRate
	}
	return false, reqInputRate, reqOutputRate
}

func DeleteSKU(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}

	id := c.Param("id")
	skuID := idToInt(id)
	if skuID <= 0 {
		middleware.RespondWithError(c, apperrors.ErrProductNotFound)
		return
	}

	db := config.GetDB()
	var spuIDForCache int
	err := db.QueryRow("DELETE FROM skus WHERE id = $1 RETURNING spu_id", skuID).Scan(&spuIDForCache)
	if err != nil {
		if err == sql.ErrNoRows {
			middleware.RespondWithError(c, apperrors.ErrProductNotFound)
			return
		}
		middleware.RespondWithError(c, apperrors.NewAppError(
			"SKU_DELETE_FAILED",
			"删除SKU失败",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	ctx := context.Background()
	cache.Delete(ctx, cache.SKUKey(skuID))
	cache.Delete(ctx, cache.SPUKey(spuIDForCache))
	cache.InvalidatePatterns(ctx, "skus:list:*")
	cache.InvalidatePatterns(ctx, "spus:list:*")

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
	// Single retail ledger: same balances as GET /tokens/balance (046 merged legacy compute_point_accounts).
	var balance, totalEarned, totalUsed float64
	err := db.QueryRow(
		`SELECT COALESCE(balance, 0), COALESCE(total_earned, 0), COALESCE(total_used, 0) FROM tokens WHERE user_id = $1`,
		userIDInt,
	).Scan(&balance, &totalEarned, &totalUsed)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusOK, gin.H{"data": models.ComputePointBalanceResponse{
			Balance:      0,
			TotalEarned:  0,
			TotalUsed:    0,
			TotalExpired: 0,
		}})
		return
	}
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": models.ComputePointBalanceResponse{
		Balance:      balance,
		TotalEarned:  totalEarned,
		TotalUsed:    totalUsed,
		TotalExpired: 0,
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
		`SELECT id, user_id, type, amount, order_id, COALESCE(reason, ''), created_at 
		 FROM token_transactions WHERE user_id = $1 
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
		var orderID sql.NullInt64
		var reason string

		err := rows.Scan(&t.ID, &t.UserID, &t.Type, &t.Amount, &orderID, &reason, &t.CreatedAt)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}

		t.BalanceAfter = 0
		if orderID.Valid {
			oid := int(orderID.Int64)
			t.OrderID = &oid
		}
		t.Description = reason

		transactions = append(transactions, t)
	}

	var total int
	db.QueryRow("SELECT COUNT(*) FROM token_transactions WHERE user_id = $1", userIDInt).Scan(&total)

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
		 us.status, us.auto_renew, us.pricing_version_id, us.entitlement_anchor_at, us.created_at, us.updated_at,
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
		var pvID sql.NullInt64
		var anchorAt sql.NullTime
		err := rows.Scan(&s.ID, &s.UserID, &s.SKUID, &s.StartDate, &s.EndDate, &s.UsedTokens, &s.UsedComputePoints,
			&s.Status, &s.AutoRenew, &pvID, &anchorAt, &s.CreatedAt, &s.UpdatedAt, &s.SKUCode, &s.SPUName, &s.RetailPrice)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		if pvID.Valid {
			v := int(pvID.Int64)
			s.PricingVersionID = &v
		}
		if anchorAt.Valid {
			t := anchorAt.Time
			s.EntitlementAnchorAt = &t
		}
		subscriptions = append(subscriptions, s)
	}

	c.JSON(http.StatusOK, gin.H{"data": subscriptions})
}

func GetModelProviders(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}

	providers, err := loadActiveModelProviders()
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": providers})
}

// GetMerchantModelProviders exposes the same active-only model_providers list as GetModelProviders for merchant API key forms (provider dropdown). Requires merchant role; does not expose inactive providers.
func GetMerchantModelProviders(c *gin.Context) {
	if !ensureMerchant(c) {
		return
	}

	providers, err := loadActiveModelProviders()
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": providers})
}

// ListAllModelProviders returns every model_providers row (any status) for admin maintenance.
// Not cached; use GetModelProviders for the active-only cached list used by dropdowns.
func ListAllModelProviders(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	rows, err := db.Query(
		`SELECT id, code, name, api_base_url, api_format, billing_type, cache_enabled, COALESCE(cache_discount_rate, 0),
			compat_prefixes,
			litellm_model_template, litellm_gateway_api_key_env, litellm_gateway_api_base,
			COALESCE(provider_region, 'domestic'), COALESCE(route_strategy, '{}'::jsonb), COALESCE(endpoints, '{}'::jsonb),
			status, sort_order, created_at, updated_at
		 FROM model_providers ORDER BY sort_order ASC, id ASC`,
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
		var litellmTpl, litellmKey, litellmBase sql.NullString
		var compatPfx pq.StringArray
		var providerRegion sql.NullString
		var routeStrategyJSON, endpointsJSON []byte
		err := rows.Scan(&p.ID, &p.Code, &p.Name, &apiBaseURL, &p.APIFormat, &billingType, &p.CacheEnabled, &p.CacheDiscount,
			&compatPfx, &litellmTpl, &litellmKey, &litellmBase, &providerRegion, &routeStrategyJSON, &endpointsJSON,
			&p.Status, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt)
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
		if litellmTpl.Valid {
			p.LitellmModelTemplate = litellmTpl.String
		}
		if litellmKey.Valid {
			p.LitellmGatewayAPIKeyEnv = litellmKey.String
		}
		if litellmBase.Valid {
			p.LitellmGatewayAPIBase = litellmBase.String
		}
		if len(compatPfx) > 0 {
			p.CompatPrefixes = []string(compatPfx)
		}
		if providerRegion.Valid {
			p.ProviderRegion = providerRegion.String
		}
		if len(routeStrategyJSON) > 0 {
			json.Unmarshal(routeStrategyJSON, &p.RouteStrategy)
		}
		if len(endpointsJSON) > 0 {
			json.Unmarshal(endpointsJSON, &p.Endpoints)
		}
		providers = append(providers, p)
	}

	c.JSON(http.StatusOK, gin.H{"data": providers})
}

var modelProviderCodeRegexp = regexp.MustCompile(`^[a-z][a-z0-9_]{0,48}$`)

// CreateModelProvider inserts a row into model_providers (admin only).
func CreateModelProvider(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}

	var req struct {
		Code                    string   `json:"code" binding:"required"`
		Name                    string   `json:"name" binding:"required"`
		APIBaseURL              string   `json:"api_base_url"`
		APIFormat               string   `json:"api_format"`
		BillingType             string   `json:"billing_type"`
		Status                  string   `json:"status"`
		SortOrder               int      `json:"sort_order"`
		CompatPrefixes          []string `json:"compat_prefixes"`
		LitellmModelTemplate    string   `json:"litellm_model_template"`
		LitellmGatewayAPIKeyEnv string   `json:"litellm_gateway_api_key_env"`
		LitellmGatewayAPIBase   string   `json:"litellm_gateway_api_base"`
	}
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	code := strings.ToLower(strings.TrimSpace(req.Code))
	name := strings.TrimSpace(req.Name)
	if code == "" || name == "" || !modelProviderCodeRegexp.MatchString(code) {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"INVALID_MODEL_PROVIDER_CODE",
			"code must be lowercase [a-z][a-z0-9_]{0,48}",
			http.StatusBadRequest,
			nil,
		))
		return
	}
	if code == modelProviderFallbackCode {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"MODEL_PROVIDER_RESERVED_CODE",
			"code __default__ is reserved for fallback provider and cannot be created from this endpoint",
			http.StatusBadRequest,
			nil,
		))
		return
	}

	apiFormat := strings.TrimSpace(req.APIFormat)
	if apiFormat == "" {
		apiFormat = "openai"
	}
	billingType := strings.TrimSpace(req.BillingType)
	if billingType == "" {
		billingType = "flat"
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = merchantStatusActive
	}
	if status != merchantStatusActive && status != merchantSKUStatusInactive {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	apiBase := strings.TrimSpace(req.APIBaseURL)
	var apiBaseArg interface{}
	if apiBase != "" {
		apiBaseArg = apiBase
	}

	litellmTpl := strings.TrimSpace(req.LitellmModelTemplate)
	var litellmTplArg interface{}
	if litellmTpl != "" {
		litellmTplArg = litellmTpl
	}
	litellmKey := strings.TrimSpace(req.LitellmGatewayAPIKeyEnv)
	var litellmKeyArg interface{}
	if litellmKey != "" {
		litellmKeyArg = litellmKey
	}
	litellmBase := strings.TrimSpace(req.LitellmGatewayAPIBase)
	var litellmBaseArg interface{}
	if litellmBase != "" {
		litellmBaseArg = litellmBase
	}

	normalizedPrefixes, nErr := services.NormalizeCompatPrefixes(req.CompatPrefixes)
	if nErr != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"INVALID_COMPAT_PREFIXES",
			nErr.Error(),
			http.StatusBadRequest,
			nil,
		))
		return
	}
	if normalizedPrefixes == nil {
		normalizedPrefixes = []string{}
	}
	compatArg := pq.Array(normalizedPrefixes)

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var p models.ModelProvider
	var compatScan pq.StringArray
	var providerRegion sql.NullString
	var routeStrategyJSON, endpointsJSON []byte
	err := db.QueryRow(
		`INSERT INTO model_providers (code, name, api_base_url, api_format, billing_type, status, sort_order, compat_prefixes,
			litellm_model_template, litellm_gateway_api_key_env, litellm_gateway_api_base,
			provider_region, route_strategy, endpoints)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13::jsonb, $14::jsonb)
		 RETURNING id, code, name, COALESCE(api_base_url, ''), api_format, COALESCE(billing_type, ''), cache_enabled, COALESCE(cache_discount_rate, 0), compat_prefixes,
			COALESCE(litellm_model_template, ''), COALESCE(litellm_gateway_api_key_env, ''), COALESCE(litellm_gateway_api_base, ''),
			COALESCE(provider_region, 'domestic'), COALESCE(route_strategy, '{}'::jsonb), COALESCE(endpoints, '{}'::jsonb),
			status, sort_order, created_at, updated_at`,
		code, name, apiBaseArg, apiFormat, billingType, status, req.SortOrder, compatArg, litellmTplArg, litellmKeyArg, litellmBaseArg,
		"domestic", "{}", "{}",
	).Scan(&p.ID, &p.Code, &p.Name, &p.APIBaseURL, &p.APIFormat, &p.BillingType, &p.CacheEnabled, &p.CacheDiscount, &compatScan,
		&p.LitellmModelTemplate, &p.LitellmGatewayAPIKeyEnv, &p.LitellmGatewayAPIBase,
		&providerRegion, &routeStrategyJSON, &endpointsJSON,
		&p.Status, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		log.Printf("[CreateModelProvider] INSERT failed: %v", err)
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			middleware.RespondWithError(c, apperrors.NewAppError(
				"MODEL_PROVIDER_CODE_EXISTS",
				"Model provider code already exists",
				http.StatusConflict,
				err,
			))
			return
		}
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	if len(compatScan) > 0 {
		p.CompatPrefixes = []string(compatScan)
	}
	if providerRegion.Valid {
		p.ProviderRegion = providerRegion.String
	}
	if len(routeStrategyJSON) > 0 {
		json.Unmarshal(routeStrategyJSON, &p.RouteStrategy)
	}
	if len(endpointsJSON) > 0 {
		json.Unmarshal(endpointsJSON, &p.Endpoints)
	}

	cache.Delete(context.Background(), "model_providers:all")

	c.JSON(http.StatusCreated, gin.H{"data": p})
}

func PatchModelProvider(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	var req struct {
		Name                    *string   `json:"name"`
		APIBaseURL              *string   `json:"api_base_url"`
		APIFormat               *string   `json:"api_format"`
		BillingType             *string   `json:"billing_type"`
		Status                  *string   `json:"status"`
		SortOrder               *int      `json:"sort_order"`
		CompatPrefixes          *[]string `json:"compat_prefixes"`
		LitellmModelTemplate    *string   `json:"litellm_model_template"`
		LitellmGatewayAPIKeyEnv *string   `json:"litellm_gateway_api_key_env"`
		LitellmGatewayAPIBase   *string   `json:"litellm_gateway_api_base"`
	}
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	if req.Status != nil {
		if *req.Status != merchantStatusActive && *req.Status != merchantSKUStatusInactive {
			middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
			return
		}
	}

	compatSet := false
	compatArg := pq.Array([]string{})
	if req.CompatPrefixes != nil {
		normalized, nErr := services.NormalizeCompatPrefixes(*req.CompatPrefixes)
		if nErr != nil {
			middleware.RespondWithError(c, apperrors.NewAppError(
				"INVALID_COMPAT_PREFIXES",
				nErr.Error(),
				http.StatusBadRequest,
				nil,
			))
			return
		}
		compatSet = true
		compatArg = pq.Array(normalized)
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var existingCode string
	if err = db.QueryRow(`SELECT code FROM model_providers WHERE id = $1`, id).Scan(&existingCode); err == sql.ErrNoRows {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"MODEL_PROVIDER_NOT_FOUND",
			"Model provider not found",
			http.StatusNotFound,
			err,
		))
		return
	} else if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	if existingCode == modelProviderFallbackCode {
		allowOnlyStatusToggle := req.Status != nil &&
			req.Name == nil &&
			req.APIBaseURL == nil &&
			req.APIFormat == nil &&
			req.BillingType == nil &&
			req.SortOrder == nil &&
			req.CompatPrefixes == nil &&
			req.LitellmModelTemplate == nil &&
			req.LitellmGatewayAPIKeyEnv == nil &&
			req.LitellmGatewayAPIBase == nil
		if !allowOnlyStatusToggle {
			middleware.RespondWithError(c, apperrors.NewAppError(
				"MODEL_PROVIDER_READONLY",
				"fallback provider __default__ only allows status toggle",
				http.StatusForbidden,
				nil,
			))
			return
		}
	}

	var p models.ModelProvider
	var compatScan pq.StringArray
	err = db.QueryRow(
		`UPDATE model_providers SET
			name = COALESCE($1, name),
			api_base_url = COALESCE($2, api_base_url),
			api_format = COALESCE($3, api_format),
			billing_type = COALESCE($4, billing_type),
			status = COALESCE($5, status),
			sort_order = COALESCE($6, sort_order),
			compat_prefixes = CASE WHEN $7::boolean THEN COALESCE($8, '{}'::text[]) ELSE compat_prefixes END,
			litellm_model_template = COALESCE($9, litellm_model_template),
			litellm_gateway_api_key_env = COALESCE($10, litellm_gateway_api_key_env),
			litellm_gateway_api_base = COALESCE($11, litellm_gateway_api_base),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $12
		RETURNING id, code, name, COALESCE(api_base_url, ''), api_format, COALESCE(billing_type, ''), cache_enabled, COALESCE(cache_discount_rate, 0), compat_prefixes,
			COALESCE(litellm_model_template, ''), COALESCE(litellm_gateway_api_key_env, ''), COALESCE(litellm_gateway_api_base, ''),
			status, sort_order, created_at, updated_at`,
		req.Name, req.APIBaseURL, req.APIFormat, req.BillingType, req.Status, req.SortOrder, compatSet, compatArg,
		req.LitellmModelTemplate, req.LitellmGatewayAPIKeyEnv, req.LitellmGatewayAPIBase, id,
	).Scan(&p.ID, &p.Code, &p.Name, &p.APIBaseURL, &p.APIFormat, &p.BillingType, &p.CacheEnabled, &p.CacheDiscount, &compatScan,
		&p.LitellmModelTemplate, &p.LitellmGatewayAPIKeyEnv, &p.LitellmGatewayAPIBase, &p.Status, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"MODEL_PROVIDER_NOT_FOUND",
			"Model provider not found",
			http.StatusNotFound,
			err,
		))
		return
	}
	if err != nil {
		logger.LogError(c.Request.Context(), "model_provider", "Failed to update model provider", err, nil)
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	if len(compatScan) > 0 {
		p.CompatPrefixes = []string(compatScan)
	}

	cache.Delete(context.Background(), "model_providers:all")

	c.JSON(http.StatusOK, gin.H{"data": p})
}

func ListPublicSKUs(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("per_page", "20")
	spuID := strings.TrimSpace(c.Query("spu_id"))
	skuType := strings.TrimSpace(c.Query("type"))
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		q = strings.TrimSpace(c.Query("search"))
	}
	provider := strings.TrimSpace(c.Query("provider"))
	tier := strings.TrimSpace(c.Query("tier"))
	modelName := strings.TrimSpace(c.Query("model_name"))
	category := strings.TrimSpace(c.Query("category"))
	groupEnabled := strings.TrimSpace(c.Query("group_enabled"))
	priceMinStr := strings.TrimSpace(c.Query("price_min"))
	priceMaxStr := strings.TrimSpace(c.Query("price_max"))
	validMinStr := strings.TrimSpace(c.Query("valid_days_min"))
	validMaxStr := strings.TrimSpace(c.Query("valid_days_max"))
	sortParam := strings.TrimSpace(c.Query("sort"))
	scenario := strings.TrimSpace(c.Query("scenario"))

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

	where := "s.status = 'active' AND sp.status = 'active'"
	args := []interface{}{}
	argPos := 1

	if spuID != "" {
		where += fmt.Sprintf(" AND s.spu_id = $%d", argPos)
		args = append(args, idToInt(spuID))
		argPos++
	}
	if skuType != "" {
		where += fmt.Sprintf(" AND s.sku_type = $%d", argPos)
		args = append(args, skuType)
		argPos++
	}
	if provider != "" {
		where += fmt.Sprintf(" AND sp.model_provider = $%d", argPos)
		args = append(args, provider)
		argPos++
	}
	if tier != "" {
		where += fmt.Sprintf(" AND sp.model_tier = $%d", argPos)
		args = append(args, tier)
		argPos++
	}
	if modelName != "" {
		pattern := "%" + modelName + "%"
		where += fmt.Sprintf(" AND sp.model_name ILIKE $%d", argPos)
		args = append(args, pattern)
		argPos++
	}
	if category != "" {
		pattern := "%" + category + "%"
		where += fmt.Sprintf(" AND sp.name ILIKE $%d", argPos)
		args = append(args, pattern)
		argPos++
	}
	if q != "" {
		pattern := "%" + q + "%"
		where += fmt.Sprintf(" AND (sp.name ILIKE $%d OR sp.model_name ILIKE $%d OR s.sku_code ILIKE $%d)", argPos, argPos+1, argPos+2)
		args = append(args, pattern, pattern, pattern)
		argPos += 3
	}
	if groupEnabled == "true" || groupEnabled == "1" {
		where += " AND s.group_enabled = true"
	}
	if priceMinStr != "" {
		if v, err := strconv.ParseFloat(priceMinStr, 64); err == nil {
			where += fmt.Sprintf(" AND s.retail_price >= $%d", argPos)
			args = append(args, v)
			argPos++
		}
	}
	if priceMaxStr != "" {
		if v, err := strconv.ParseFloat(priceMaxStr, 64); err == nil {
			where += fmt.Sprintf(" AND s.retail_price <= $%d", argPos)
			args = append(args, v)
			argPos++
		}
	}
	if validMinStr != "" {
		if v, err := strconv.Atoi(validMinStr); err == nil {
			where += fmt.Sprintf(" AND s.valid_days >= $%d", argPos)
			args = append(args, v)
			argPos++
		}
	}
	if validMaxStr != "" {
		if v, err := strconv.Atoi(validMaxStr); err == nil {
			where += fmt.Sprintf(" AND s.valid_days <= $%d", argPos)
			args = append(args, v)
			argPos++
		}
	}
	if scenario != "" {
		where += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM spu_scenarios ss JOIN usage_scenarios us ON ss.scenario_id = us.id WHERE ss.spu_id = s.spu_id AND us.code = $%d)", argPos)
		args = append(args, scenario)
		argPos++
	}

	orderClause := "s.is_promoted DESC, s.sales_count DESC, s.created_at DESC"
	switch sortParam {
	case "hot":
		orderClause = "s.sales_count DESC, s.created_at DESC"
	case "new":
		orderClause = "s.created_at DESC"
	case "price_asc":
		orderClause = "s.retail_price ASC NULLS LAST, s.id ASC"
	case "price_desc":
		orderClause = "s.retail_price DESC NULLS LAST, s.id DESC"
	}

	selectCols := `SELECT s.id, s.spu_id, s.sku_code, s.sku_type, s.token_amount, s.compute_points, 
		s.subscription_period, s.is_unlimited, s.valid_days, s.retail_price, s.original_price, 
		s.group_enabled, s.min_group_size, s.max_group_size, s.group_discount_rate,
		s.is_trial, s.trial_duration_days, s.is_promoted, s.sales_count,
		sp.name as spu_name, sp.model_provider, sp.model_name, sp.model_tier,
		sp.total_sales_count, sp.average_rating`

	query := fmt.Sprintf(`%s
		FROM skus s JOIN spus sp ON s.spu_id = sp.id WHERE %s
		ORDER BY %s LIMIT $%d OFFSET $%d`,
		selectCols, where, orderClause, argPos, argPos+1)
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
		var spTotalSales sql.NullInt64
		var spAvgRating sql.NullFloat64

		err := rows.Scan(&s.ID, &s.SPUID, &s.SKUCode, &s.SKUType, &tokenAmount, &computePoints,
			&subscriptionPeriod, &s.IsUnlimited, &s.ValidDays, &s.RetailPrice, &originalPrice,
			&s.GroupEnabled, &s.MinGroupSize, &s.MaxGroupSize, &groupDiscountRate,
			&s.IsTrial, &trialDurationDays, &s.IsPromoted, &s.SalesCount,
			&s.SPUName, &s.ModelProvider, &s.ModelName, &s.ModelTier,
			&spTotalSales, &spAvgRating)
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
		if spTotalSales.Valid {
			s.SPUTotalSalesCount = spTotalSales.Int64
		}
		if spAvgRating.Valid {
			r := spAvgRating.Float64
			s.SPUAverageRating = &r
		}

		skus = append(skus, s)
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM skus s JOIN spus sp ON s.spu_id = sp.id WHERE %s", where)
	var total int
	if err := db.QueryRow(countQuery, args[:len(args)-2]...).Scan(&total); err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

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

	var spTotalSales sql.NullInt64
	var spAvgRating sql.NullFloat64

	err := db.QueryRow(
		`SELECT s.id, s.spu_id, s.sku_code, s.sku_type, s.token_amount, s.compute_points, 
		s.subscription_period, s.is_unlimited, s.valid_days, s.retail_price, s.original_price, 
		s.group_enabled, s.min_group_size, s.max_group_size, s.group_discount_rate,
		s.is_trial, s.trial_duration_days, s.is_promoted, s.sales_count,
		sp.name as spu_name, sp.model_provider, sp.model_name, sp.model_tier,
		sp.total_sales_count, sp.average_rating
		FROM skus s JOIN spus sp ON s.spu_id = sp.id 
		WHERE s.id = $1 AND s.status = 'active' AND sp.status = 'active'`,
		skuID,
	).Scan(&s.ID, &s.SPUID, &s.SKUCode, &s.SKUType, &tokenAmount, &computePoints,
		&subscriptionPeriod, &s.IsUnlimited, &s.ValidDays, &s.RetailPrice, &originalPrice,
		&s.GroupEnabled, &s.MinGroupSize, &s.MaxGroupSize, &groupDiscountRate,
		&s.IsTrial, &trialDurationDays, &s.IsPromoted, &s.SalesCount,
		&s.SPUName, &s.ModelProvider, &s.ModelName, &s.ModelTier,
		&spTotalSales, &spAvgRating)

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
	if spTotalSales.Valid {
		s.SPUTotalSalesCount = spTotalSales.Int64
	}
	if spAvgRating.Valid {
		r := spAvgRating.Float64
		s.SPUAverageRating = &r
	}

	c.JSON(http.StatusOK, gin.H{"data": s})
}

type SPUScenarioResponse struct {
	ID          int    `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	IsLinked    bool   `json:"is_linked"`
	IsPrimary   bool   `json:"is_primary"`
}

func GetSPUScenarios(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}

	id := c.Param("id")
	spuID := idToInt(id)
	if spuID <= 0 {
		middleware.RespondWithError(c, apperrors.ErrProductNotFound)
		return
	}

	db := config.GetDB()
	query := `
		SELECT us.id, us.code, us.name, us.description, 
		       CASE WHEN ss.spu_id IS NOT NULL THEN true ELSE false END as is_linked,
		       COALESCE(ss.is_primary, false) as is_primary
		FROM usage_scenarios us
		LEFT JOIN spu_scenarios ss ON us.id = ss.scenario_id AND ss.spu_id = $1
		WHERE us.status = 'active'
		ORDER BY us.sort_order ASC
	`

	rows, err := db.Query(query, spuID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	var scenarios []SPUScenarioResponse
	for rows.Next() {
		var s SPUScenarioResponse
		var desc sql.NullString
		if err := rows.Scan(&s.ID, &s.Code, &s.Name, &desc, &s.IsLinked, &s.IsPrimary); err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		if desc.Valid {
			s.Description = desc.String
		}
		scenarios = append(scenarios, s)
	}

	c.JSON(http.StatusOK, gin.H{"scenarios": scenarios})
}

type UpdateSPUScenariosRequest struct {
	ScenarioIDs []int `json:"scenario_ids"`
	PrimaryID   int   `json:"primary_id,omitempty"`
}

func UpdateSPUScenarios(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}

	id := c.Param("id")
	spuID := idToInt(id)
	if spuID <= 0 {
		middleware.RespondWithError(c, apperrors.ErrProductNotFound)
		return
	}

	var req UpdateSPUScenariosRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM spu_scenarios WHERE spu_id = $1", spuID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	for _, scenarioID := range req.ScenarioIDs {
		isPrimary := scenarioID == req.PrimaryID
		_, err = tx.Exec(
			"INSERT INTO spu_scenarios (spu_id, scenario_id, is_primary) VALUES ($1, $2, $3)",
			spuID, scenarioID, isPrimary,
		)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	cache.InvalidatePatterns(ctx, "spus:list:*")
	cache.InvalidatePatterns(ctx, "catalog:scenarios:*")

	c.JSON(http.StatusOK, gin.H{"message": "SPU scenarios updated successfully"})
}
