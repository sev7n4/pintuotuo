package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
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

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	now := time.Now()
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
		result = append(result, FlashSaleWithProducts{
			FlashSale: sale,
			Skus:      products,
		})
	}

	if len(result) == 0 {
		result = []FlashSaleWithProducts{}
	}

	if data, err := json.Marshal(result); err == nil {
		cache.Set(ctx, cacheKey, string(data), 30)
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

	var sale FlashSale
	err = tx.QueryRow(`
		INSERT INTO flash_sales (name, description, start_time, end_time, status)
		VALUES ($1, $2, $3, $4, 'upcoming')
		RETURNING id, name, description, start_time, end_time, status, created_at, updated_at`,
		req.Name, req.Description, req.StartTime, req.EndTime,
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
		var skuType, modelProvider, modelName, providerModelID string
		err := tx.QueryRow(
			`SELECT s.retail_price, s.sku_type, COALESCE(sp.model_provider, ''), COALESCE(sp.model_name, ''), COALESCE(sp.provider_model_id, '')
			 FROM skus s
			 JOIN spus sp ON s.spu_id = sp.id
			 WHERE s.id = $1 AND s.status = 'active' AND sp.status = 'active'`,
			p.SKUID,
		).Scan(&originalPrice, &skuType, &modelProvider, &modelName, &providerModelID)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrProductNotFound)
			return
		}
		if err := services.ValidateFuelPackBundle([]services.OrderLinePolicyInput{{
			SKUType:         skuType,
			ModelProvider:   modelProvider,
			ModelName:       modelName,
			ProviderModelID: providerModelID,
		}}); err != nil {
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

	cache.Delete(context.Background(), "flash_sales:active")

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    sale,
	})
}
