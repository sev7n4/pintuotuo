package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/cache"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/logger"
	"github.com/pintuotuo/backend/metrics"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/services"
)

type FlashSale struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type FlashSaleProduct struct {
	ID            int     `json:"id"`
	FlashSaleID   int     `json:"flash_sale_id"`
	SKUID         int     `json:"sku_id"`
	ProductName   string  `json:"product_name"`
	FlashPrice    float64 `json:"flash_price"`
	OriginalPrice float64 `json:"original_price"`
	StockLimit    int     `json:"stock_limit"`
	StockSold     int     `json:"stock_sold"`
	PerUserLimit  int     `json:"per_user_limit"`
	Discount      int     `json:"discount"`
}

type FlashSaleWithProducts struct {
	FlashSale
	Skus []FlashSaleProduct `json:"skus"`
}

func GetActiveFlashSales(c *gin.Context) {
	ctx := context.Background()
	cacheKey := "flash_sales:active"

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	now := time.Now()
	if changed, err := services.PromoteFlashSaleStatuses(db, now); err != nil {
		logger.LogWarn(ctx, "flash_sale", "promote flash sale statuses failed", map[string]interface{}{"error": err.Error()})
	} else if changed {
		cache.Delete(ctx, cacheKey)
	}

	if cachedData, err := cache.Get(ctx, cacheKey); err == nil {
		var sales []FlashSaleWithProducts
		if err := json.Unmarshal([]byte(cachedData), &sales); err == nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "success",
				"data":    sales,
			})
			return
		}
	}

	rows, err := db.Query(`
		SELECT id, name, description, start_time, end_time, status, created_at, updated_at 
		FROM flash_sales 
		WHERE status = 'active' AND start_time <= $1 AND end_time > $1
		ORDER BY start_time ASC`,
		now,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	var sales []FlashSale
	for rows.Next() {
		var s FlashSale
		if err := rows.Scan(&s.ID, &s.Name, &s.Description, &s.StartTime, &s.EndTime, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		sales = append(sales, s)
	}

	var result []FlashSaleWithProducts
	for _, sale := range sales {
		products, err := getFlashSaleProducts(db, sale.ID)
		if err != nil {
			continue
		}
		if len(products) == 0 {
			continue
		}
		result = append(result, FlashSaleWithProducts{
			FlashSale: sale,
			Skus:      products,
		})
	}

	if len(result) == 0 {
		result = []FlashSaleWithProducts{}
	}

	if data, err := json.Marshal(result); err == nil {
		// 较短 TTL：多实例下降低读到过期列表的概率；状态推进另由定时任务与 promote 兜底。
		cache.Set(ctx, cacheKey, string(data), 15)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    result,
	})
}

// GetUpcomingFlashSales 返回尚未开始且仍有效的秒杀（含仍有库存的 SKU），用于卖场「即将开始」展示。
func GetUpcomingFlashSales(c *gin.Context) {
	ctx := context.Background()
	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	now := time.Now()
	if _, err := services.PromoteFlashSaleStatuses(db, now); err != nil {
		logger.LogWarn(ctx, "flash_sale", "promote before upcoming list failed", map[string]interface{}{"error": err.Error()})
	}

	rows, err := db.Query(`
		SELECT id, name, description, start_time, end_time, status, created_at, updated_at
		FROM flash_sales
		WHERE status = 'upcoming' AND start_time > $1 AND end_time > $1
		ORDER BY start_time ASC`,
		now,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	var sales []FlashSale
	for rows.Next() {
		var s FlashSale
		if err := rows.Scan(&s.ID, &s.Name, &s.Description, &s.StartTime, &s.EndTime, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		sales = append(sales, s)
	}

	var result []FlashSaleWithProducts
	for _, sale := range sales {
		products, err := getFlashSaleProducts(db, sale.ID)
		if err != nil {
			continue
		}
		if len(products) == 0 {
			continue
		}
		result = append(result, FlashSaleWithProducts{
			FlashSale: sale,
			Skus:      products,
		})
	}
	if len(result) == 0 {
		result = []FlashSaleWithProducts{}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    result,
	})
}

func GetFlashSaleProducts(c *gin.Context) {
	saleID := c.Param("id")
	saleIDInt, err := strconv.Atoi(saleID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	products, err := getFlashSaleProducts(db, saleIDInt)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    products,
	})
}

func getFlashSaleProducts(db *sql.DB, saleID int) ([]FlashSaleProduct, error) {
	rows, err := db.Query(`
		SELECT fsp.id, fsp.flash_sale_id, fsp.sku_id, sp.name || ' · ' || s.sku_code, fsp.flash_price, fsp.original_price, 
		       fsp.stock_limit, fsp.stock_sold, fsp.per_user_limit
		FROM flash_sale_products fsp
		JOIN skus s ON fsp.sku_id = s.id
		JOIN spus sp ON s.spu_id = sp.id
		WHERE fsp.flash_sale_id = $1 AND fsp.stock_limit > fsp.stock_sold`,
		saleID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []FlashSaleProduct
	for rows.Next() {
		var p FlashSaleProduct
		if err := rows.Scan(&p.ID, &p.FlashSaleID, &p.SKUID, &p.ProductName, &p.FlashPrice, &p.OriginalPrice, &p.StockLimit, &p.StockSold, &p.PerUserLimit); err != nil {
			return nil, err
		}
		if p.OriginalPrice > 0 {
			p.Discount = int((1 - p.FlashPrice/p.OriginalPrice) * 100)
		}
		products = append(products, p)
	}

	return products, nil
}

func getFlashSaleProductsAll(db *sql.DB, saleID int) ([]FlashSaleProduct, error) {
	rows, err := db.Query(`
		SELECT fsp.id, fsp.flash_sale_id, fsp.sku_id, sp.name || ' · ' || s.sku_code, fsp.flash_price, fsp.original_price, 
		       fsp.stock_limit, fsp.stock_sold, fsp.per_user_limit
		FROM flash_sale_products fsp
		JOIN skus s ON fsp.sku_id = s.id
		JOIN spus sp ON s.spu_id = sp.id
		WHERE fsp.flash_sale_id = $1`,
		saleID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []FlashSaleProduct
	for rows.Next() {
		var p FlashSaleProduct
		if err := rows.Scan(&p.ID, &p.FlashSaleID, &p.SKUID, &p.ProductName, &p.FlashPrice, &p.OriginalPrice, &p.StockLimit, &p.StockSold, &p.PerUserLimit); err != nil {
			return nil, err
		}
		if p.OriginalPrice > 0 {
			p.Discount = int((1 - p.FlashPrice/p.OriginalPrice) * 100)
		}
		products = append(products, p)
	}

	return products, nil
}

func requireFlashSaleAdmin(c *gin.Context, db *sql.DB) bool {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return false
	}
	var role string
	if err := db.QueryRow("SELECT COALESCE(role, '') FROM users WHERE id = $1", userID).Scan(&role); err != nil || role != "admin" {
		middleware.RespondWithError(c, apperrors.ErrForbidden)
		return false
	}
	return true
}

// AdminListFlashSales 管理端：列出秒杀活动（含历史），按 id 倒序。
func AdminListFlashSales(c *gin.Context) {
	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	if !requireFlashSaleAdmin(c, db) {
		return
	}

	rows, err := db.Query(`
		SELECT id, name, description, start_time, end_time, status, created_at, updated_at
		FROM flash_sales
		ORDER BY id DESC
		LIMIT 500`)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	var list []FlashSale
	for rows.Next() {
		var s FlashSale
		if err := rows.Scan(&s.ID, &s.Name, &s.Description, &s.StartTime, &s.EndTime, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		list = append(list, s)
	}
	if list == nil {
		list = []FlashSale{}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    list,
	})
}

// AdminGetFlashSaleDetail 管理端：单场详情（含售罄行，用于预览与核对）。
func AdminGetFlashSaleDetail(c *gin.Context) {
	saleID := c.Param("id")
	saleIDInt, err := strconv.Atoi(saleID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	if !requireFlashSaleAdmin(c, db) {
		return
	}

	var sale FlashSale
	err = db.QueryRow(`
		SELECT id, name, description, start_time, end_time, status, created_at, updated_at
		FROM flash_sales WHERE id = $1`,
		saleIDInt,
	).Scan(&sale.ID, &sale.Name, &sale.Description, &sale.StartTime, &sale.EndTime, &sale.Status, &sale.CreatedAt, &sale.UpdatedAt)
	if err == sql.ErrNoRows {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"FLASH_SALE_NOT_FOUND",
			"秒杀活动不存在",
			http.StatusNotFound,
			nil,
		))
		return
	}
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	skus, err := getFlashSaleProductsAll(db, saleIDInt)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": FlashSaleWithProducts{
			FlashSale: sale,
			Skus:      skus,
		},
	})
}

// AdminPatchFlashSale 仅允许「待开始」场次修改名称、描述与时间段；会重新校验 SKU 时间冲突。
func AdminPatchFlashSale(c *gin.Context) {
	saleIDInt, convErr := strconv.Atoi(c.Param("id"))
	if convErr != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	var body struct {
		Name        *string    `json:"name"`
		Description *string    `json:"description"`
		StartTime   *time.Time `json:"start_time"`
		EndTime     *time.Time `json:"end_time"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	if !requireFlashSaleAdmin(c, db) {
		return
	}

	var cur FlashSale
	err := db.QueryRow(`
		SELECT id, name, description, start_time, end_time, status, created_at, updated_at
		FROM flash_sales WHERE id = $1`,
		saleIDInt,
	).Scan(&cur.ID, &cur.Name, &cur.Description, &cur.StartTime, &cur.EndTime, &cur.Status, &cur.CreatedAt, &cur.UpdatedAt)
	if err == sql.ErrNoRows {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"FLASH_SALE_NOT_FOUND",
			"秒杀活动不存在",
			http.StatusNotFound,
			nil,
		))
		return
	}
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	if cur.Status != "upcoming" {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"FLASH_SALE_NOT_EDITABLE",
			"仅「待开始」场次可修改基本信息与时间",
			http.StatusBadRequest,
			nil,
		))
		return
	}

	if body.Name == nil && body.Description == nil && body.StartTime == nil && body.EndTime == nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"NO_FIELDS",
			"请至少提交一个要修改的字段",
			http.StatusBadRequest,
			nil,
		))
		return
	}

	name := cur.Name
	if body.Name != nil && strings.TrimSpace(*body.Name) != "" {
		name = strings.TrimSpace(*body.Name)
	}
	desc := cur.Description
	if body.Description != nil {
		desc = *body.Description
	}
	st := cur.StartTime
	if body.StartTime != nil {
		st = *body.StartTime
	}
	en := cur.EndTime
	if body.EndTime != nil {
		en = *body.EndTime
	}
	if !en.After(st) {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"INVALID_TIME",
			"结束时间必须晚于开始时间",
			http.StatusBadRequest,
			nil,
		))
		return
	}

	tx, err := db.Begin()
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer tx.Rollback()

	skuRows, err := tx.Query(`SELECT sku_id FROM flash_sale_products WHERE flash_sale_id = $1`, saleIDInt)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	var skuIDs []int
	for skuRows.Next() {
		var sid int
		if scanErr := skuRows.Scan(&sid); scanErr != nil {
			skuRows.Close()
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		skuIDs = append(skuIDs, sid)
	}
	skuRows.Close()

	for _, sk := range skuIDs {
		var overlapCount int
		err = tx.QueryRow(`
			SELECT COUNT(*) FROM flash_sale_products fsp
			INNER JOIN flash_sales fs ON fs.id = fsp.flash_sale_id
			WHERE fsp.sku_id = $1
			  AND fs.id <> $4
			  AND fs.status IN ('upcoming', 'active')
			  AND fs.start_time < $3 AND fs.end_time > $2`,
			sk, st, en, saleIDInt,
		).Scan(&overlapCount)
		if err != nil {
			middleware.RespondWithError(c, apperrors.NewAppError(
				"FLASH_SALE_OVERLAP_CHECK_FAILED",
				"检查场次冲突失败",
				http.StatusInternalServerError,
				err,
			))
			return
		}
		if overlapCount > 0 {
			middleware.RespondWithError(c, apperrors.NewAppError(
				"FLASH_SALE_SKU_TIME_OVERLAP",
				"该 SKU 在相同时间段内已有未结束或未取消的秒杀场次，请调整时间或 SKU",
				http.StatusBadRequest,
				nil,
			))
			return
		}
	}

	var sale FlashSale
	err = tx.QueryRow(`
		UPDATE flash_sales
		   SET name = $1, description = $2, start_time = $3, end_time = $4, updated_at = CURRENT_TIMESTAMP
		 WHERE id = $5 AND status = 'upcoming'
		 RETURNING id, name, description, start_time, end_time, status, created_at, updated_at`,
		name, desc, st, en, saleIDInt,
	).Scan(&sale.ID, &sale.Name, &sale.Description, &sale.StartTime, &sale.EndTime, &sale.Status, &sale.CreatedAt, &sale.UpdatedAt)
	if err == sql.ErrNoRows {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"FLASH_SALE_NOT_EDITABLE",
			"场次状态已变更，请刷新后重试",
			http.StatusConflict,
			nil,
		))
		return
	}
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	if err := tx.Commit(); err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	opUID, _ := intFromGinUserID(c)
	logger.LogInfo(c.Request.Context(), "flash_sale_admin", "flash sale metadata patched", map[string]interface{}{
		"operator_user_id": opUID,
		"flash_sale_id":    saleIDInt,
		"name":             name,
	})

	cache.Delete(context.Background(), "flash_sales:active")

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    sale,
	})
}

func CreateFlashSale(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	var req struct {
		Name        string    `json:"name" binding:"required"`
		Description string    `json:"description"`
		StartTime   time.Time `json:"start_time" binding:"required"`
		EndTime     time.Time `json:"end_time" binding:"required"`
		Skus        []struct {
			SKUID        int     `json:"sku_id" binding:"required"`
			FlashPrice   float64 `json:"flash_price" binding:"required"`
			StockLimit   int     `json:"stock_limit" binding:"required"`
			PerUserLimit int     `json:"per_user_limit"`
		} `json:"skus" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	if req.EndTime.Before(req.StartTime) {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"INVALID_TIME",
			"结束时间必须晚于开始时间",
			http.StatusBadRequest,
			nil,
		))
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var role string
	if err := db.QueryRow("SELECT COALESCE(role, '') FROM users WHERE id = $1", userID).Scan(&role); err != nil || role != "admin" {
		middleware.RespondWithError(c, apperrors.ErrForbidden)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"TRANSACTION_START_FAILED",
			"Failed to start transaction",
			http.StatusInternalServerError,
			err,
		))
		return
	}
	defer tx.Rollback()

	if _, perr := services.PromoteFlashSaleStatuses(tx, time.Now()); perr != nil {
		logger.LogWarn(c.Request.Context(), "flash_sale", "promote in create flash tx failed", map[string]interface{}{"error": perr.Error()})
	}

	now := time.Now()
	saleStatus := "upcoming"
	if !req.StartTime.After(now) && req.EndTime.After(now) {
		saleStatus = productStatusActive
	}

	var sale FlashSale
	err = tx.QueryRow(`
		INSERT INTO flash_sales (name, description, start_time, end_time, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, description, start_time, end_time, status, created_at, updated_at`,
		req.Name, req.Description, req.StartTime, req.EndTime, saleStatus,
	).Scan(&sale.ID, &sale.Name, &sale.Description, &sale.StartTime, &sale.EndTime, &sale.Status, &sale.CreatedAt, &sale.UpdatedAt)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"FLASH_SALE_CREATE_FAILED",
			"Failed to create flash sale",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	for _, p := range req.Skus {
		var originalPrice float64
		var skuStock int
		var skuType, modelProvider, modelName, providerModelID string
		err := tx.QueryRow(
			`SELECT s.retail_price, s.stock, s.sku_type, COALESCE(sp.model_provider, ''), COALESCE(sp.model_name, ''), COALESCE(sp.provider_model_id, '')
			 FROM skus s
			 JOIN spus sp ON s.spu_id = sp.id
			 WHERE s.id = $1 AND s.status = 'active' AND sp.status = 'active'`,
			p.SKUID,
		).Scan(&originalPrice, &skuStock, &skuType, &modelProvider, &modelName, &providerModelID)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrProductNotFound)
			return
		}
		validateErr := services.ValidateFuelPackBundle([]services.OrderLinePolicyInput{{
			SKUType:         skuType,
			ModelProvider:   modelProvider,
			ModelName:       modelName,
			ProviderModelID: providerModelID,
		}})
		if validateErr != nil {
			metrics.RecordFuelPackRestriction("admin_flash_sale", "FUEL_PACK_PURCHASE_RESTRICTED")
			logger.LogWarn(c.Request.Context(), "fuel_pack_policy", "Blocked flash sale sku without model entitlement", map[string]interface{}{
				"source": "admin_flash_sale",
				"sku_id": p.SKUID,
				"code":   "FUEL_PACK_PURCHASE_RESTRICTED",
			})
			middleware.RespondWithError(c, apperrors.NewAppError(
				"FUEL_PACK_PURCHASE_RESTRICTED",
				"加油包不可单独购买，秒杀活动仅支持带模型商品",
				http.StatusBadRequest,
				nil,
			))
			return
		}

		if skuStock >= 0 && p.StockLimit > skuStock {
			middleware.RespondWithError(c, apperrors.NewAppError(
				"FLASH_SALE_STOCK_EXCEEDS_SKU",
				"秒杀库存不能超过 SKU 当前可售库存",
				http.StatusBadRequest,
				nil,
			))
			return
		}

		var overlapCount int
		err = tx.QueryRow(`
			SELECT COUNT(*) FROM flash_sale_products fsp
			INNER JOIN flash_sales fs ON fs.id = fsp.flash_sale_id
			WHERE fsp.sku_id = $1
			  AND fs.status IN ('upcoming', 'active')
			  AND fs.start_time < $3 AND fs.end_time > $2`,
			p.SKUID, req.StartTime, req.EndTime,
		).Scan(&overlapCount)
		if err != nil {
			middleware.RespondWithError(c, apperrors.NewAppError(
				"FLASH_SALE_OVERLAP_CHECK_FAILED",
				"检查场次冲突失败",
				http.StatusInternalServerError,
				err,
			))
			return
		}
		if overlapCount > 0 {
			middleware.RespondWithError(c, apperrors.NewAppError(
				"FLASH_SALE_SKU_TIME_OVERLAP",
				"该 SKU 在相同时间段内已有未结束或未取消的秒杀场次，请调整时间或 SKU",
				http.StatusBadRequest,
				nil,
			))
			return
		}

		_, err = tx.Exec(`
			INSERT INTO flash_sale_products (flash_sale_id, sku_id, flash_price, original_price, stock_limit, per_user_limit)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			sale.ID, p.SKUID, p.FlashPrice, originalPrice, p.StockLimit, p.PerUserLimit,
		)
		if err != nil {
			middleware.RespondWithError(c, apperrors.NewAppError(
				"FLASH_SALE_PRODUCT_CREATE_FAILED",
				"Failed to add product to flash sale",
				http.StatusInternalServerError,
				err,
			))
			return
		}
	}

	if err := tx.Commit(); err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"TRANSACTION_COMMIT_FAILED",
			"Failed to commit transaction",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	cache.Delete(context.Background(), "flash_sales:active")

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "success",
		"data":    sale,
	})
}

func UpdateFlashSaleStatus(c *gin.Context) {
	saleID := c.Param("id")
	saleIDInt, convErr := strconv.Atoi(saleID)
	if convErr != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	validStatuses := map[string]bool{"upcoming": true, "active": true, "ended": true, "canceled": true}
	if !validStatuses[req.Status] {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"INVALID_STATUS",
			"Invalid status value",
			http.StatusBadRequest,
			nil,
		))
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	if !requireFlashSaleAdmin(c, db) {
		return
	}

	var sale FlashSale
	err := db.QueryRow(`
		UPDATE flash_sales SET status = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
		RETURNING id, name, description, start_time, end_time, status, created_at, updated_at`,
		req.Status, saleIDInt,
	).Scan(&sale.ID, &sale.Name, &sale.Description, &sale.StartTime, &sale.EndTime, &sale.Status, &sale.CreatedAt, &sale.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"FLASH_SALE_UPDATE_FAILED",
			"Failed to update flash sale",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	opUID, _ := intFromGinUserID(c)
	logger.LogInfo(c.Request.Context(), "flash_sale_admin", "flash sale status updated", map[string]interface{}{
		"operator_user_id": opUID,
		"flash_sale_id":    saleIDInt,
		"new_status":       req.Status,
	})

	cache.Delete(context.Background(), "flash_sales:active")

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    sale,
	})
}
