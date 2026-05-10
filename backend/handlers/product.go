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

const productStatusActive = "active"

// ListProducts retrieves product list with pagination
func ListProducts(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("per_page", "20")
	status := c.DefaultQuery("status", productStatusActive)

	pageNum, _ := strconv.Atoi(page)
	perPageNum, _ := strconv.Atoi(perPage)

	if pageNum < 1 {
		pageNum = 1
	}
	if perPageNum < 1 || perPageNum > 100 {
		perPageNum = 20
	}

	ctx := context.Background()
	cacheKey := cache.ProductListKey(pageNum, perPageNum, status)

	// Try cache first
	if cachedList, err := cache.Get(ctx, cacheKey); err == nil {
		var cachedData struct {
			Total   int              `json:"total"`
			Page    int              `json:"page"`
			PerPage int              `json:"per_page"`
			Data    []models.Product `json:"data"`
		}
		if err := json.Unmarshal([]byte(cachedList), &cachedData); err == nil {
			c.JSON(http.StatusOK, cachedData)
			return
		}
	}

	offset := (pageNum - 1) * perPageNum

	db := config.GetDB()

	// Marketplace: one row per SKU (legacy `products` table is deprecated).
	base := listSKUProductsBaseQuery
	var rows *sql.Rows
	var err error
	if status == "" || status == allProductStatus {
		q := base + ` ORDER BY s.created_at DESC LIMIT $1 OFFSET $2`
		rows, err = db.Query(q, perPageNum, offset)
	} else {
		q := base + ` AND s.status = $1 ORDER BY s.created_at DESC LIMIT $2 OFFSET $3`
		rows, err = db.Query(q, status, perPageNum, offset)
	}

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		p, scanErr := productFromSKU(rows)
		if scanErr != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		products = append(products, p)
	}

	var total int
	var countErr error
	if status == "" || status == allProductStatus {
		countErr = db.QueryRow(`SELECT COUNT(*) FROM skus s JOIN spus sp ON s.spu_id = sp.id WHERE s.status = 'active' AND sp.status = 'active' AND (s.stock > 0 OR s.stock = -1)`).Scan(&total)
	} else {
		countErr = db.QueryRow(`SELECT COUNT(*) FROM skus s JOIN spus sp ON s.spu_id = sp.id WHERE s.status = $1 AND sp.status = 'active' AND (s.stock > 0 OR s.stock = -1)`, status).Scan(&total)
	}

	if countErr != nil {
		total = 0
	}

	result := gin.H{
		"total":    total,
		"page":     pageNum,
		"per_page": perPageNum,
		"data":     products,
	}

	// Cache the result
	if resultJSON, err := json.Marshal(result); err == nil {
		cache.Set(ctx, cacheKey, string(resultJSON), cache.ProductListTTL)
	}

	c.JSON(http.StatusOK, result)
}

// GetProductByID retrieves a single product by ID with caching
func GetProductByID(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()

	productID := idToInt(id)
	if productID <= 0 {
		middleware.RespondWithError(c, apperrors.ErrProductNotFound)
		return
	}

	// Try cache first (ignore errors if cache is not available)
	cacheKey := cache.ProductKey(productID)
	if cachedProduct, err := cache.Get(ctx, cacheKey); err == nil {
		var product models.Product
		if err := json.Unmarshal([]byte(cachedProduct), &product); err == nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "success",
				"data":    product,
			})
			return
		}
	}

	// Query database
	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrProductNotFound)
		return
	}

	product, err := getProductBySKUID(db, productID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrProductNotFound)
		return
	}

	// Cache the result
	if productJSON, err := json.Marshal(product); err == nil {
		cache.Set(ctx, cacheKey, string(productJSON), cache.ProductCacheTTL)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    product,
	})
}

// SearchProducts searches for products by query
func SearchProducts(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
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

	ctx := context.Background()
	cacheKey := cache.ProductSearchKey(query, pageNum, perPageNum)

	// Try cache first
	if cachedResults, err := cache.Get(ctx, cacheKey); err == nil {
		var cachedData struct {
			Total   int              `json:"total"`
			Page    int              `json:"page"`
			PerPage int              `json:"per_page"`
			Data    []models.Product `json:"data"`
		}
		if err := json.Unmarshal([]byte(cachedResults), &cachedData); err == nil {
			c.JSON(http.StatusOK, cachedData)
			return
		}
	}

	offset := (pageNum - 1) * perPageNum

	db := config.GetDB()

	searchQuery := "%" + query + "%"
	q := listSKUProductsBaseQuery + ` AND (sp.name ILIKE $1 OR sp.description ILIKE $1 OR s.sku_code ILIKE $1 OR sp.model_name ILIKE $1)
		ORDER BY s.created_at DESC LIMIT $2 OFFSET $3`
	rows, err := db.Query(q, searchQuery, perPageNum, offset)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		p, scanErr := productFromSKU(rows)
		if scanErr != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		products = append(products, p)
	}

	var total int
	db.QueryRow(
		`SELECT COUNT(*) FROM skus s JOIN spus sp ON s.spu_id = sp.id WHERE s.status = 'active' AND sp.status = 'active'
		 AND (s.stock > 0 OR s.stock = -1)
		 AND (sp.name ILIKE $1 OR sp.description ILIKE $1 OR s.sku_code ILIKE $1 OR sp.model_name ILIKE $1)`,
		searchQuery,
	).Scan(&total)

	result := gin.H{
		"total":    total,
		"page":     pageNum,
		"per_page": perPageNum,
		"data":     products,
	}

	// Cache the result
	if resultJSON, err := json.Marshal(result); err == nil {
		cache.Set(ctx, cacheKey, string(resultJSON), cache.SearchResultsTTL)
	}

	c.JSON(http.StatusOK, result)
}

// CreateProduct is removed: use merchant SKU shelf APIs (POST /merchants/skus).
func CreateProduct(c *gin.Context) {
	c.JSON(http.StatusGone, gin.H{
		"code":    "DEPRECATED",
		"message": "Legacy products API removed. Use POST /api/v1/merchants/skus to put platform SKUs on shelf.",
	})
}

// UpdateProduct is removed: use PUT /merchants/skus/:id.
func UpdateProduct(c *gin.Context) {
	c.JSON(http.StatusGone, gin.H{
		"code":    "DEPRECATED",
		"message": "Legacy products API removed. Use PUT /api/v1/merchants/skus/:id.",
	})
}

// DeleteProduct is removed: use DELETE /merchants/skus/:id.
func DeleteProduct(c *gin.Context) {
	c.JSON(http.StatusGone, gin.H{
		"code":    "DEPRECATED",
		"message": "Legacy products API removed. Use DELETE /api/v1/merchants/skus/:id.",
	})
}

// Helper function to convert string ID to int
func idToInt(id string) int {
	idInt, _ := strconv.Atoi(id)
	return idInt
}

// GetHotProducts retrieves hot/popular products
func GetHotProducts(c *gin.Context) {
	limit := c.DefaultQuery("limit", "10")
	limitNum, _ := strconv.Atoi(limit)
	if limitNum < 1 || limitNum > 50 {
		limitNum = 10
	}

	ctx := context.Background()
	cacheKey := "products:hot:sku:" + strconv.Itoa(limitNum)

	// Try cache first
	if cachedProducts, err := cache.Get(ctx, cacheKey); err == nil {
		var products []models.Product
		if err := json.Unmarshal([]byte(cachedProducts), &products); err == nil {
			c.JSON(http.StatusOK, gin.H{
				"data": products,
			})
			return
		}
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	q := listSKUProductsBaseQuery + ` ORDER BY s.sales_count DESC NULLS LAST, s.created_at DESC LIMIT $1`
	rows, err := db.Query(q, limitNum)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		p, scanErr := productFromSKU(rows)
		if scanErr != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		products = append(products, p)
	}

	// Cache for 5 minutes
	if productsJSON, err := json.Marshal(products); err == nil {
		cache.Set(ctx, cacheKey, string(productsJSON), 5*60)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": products,
	})
}

// GetNewProducts retrieves new arrivals
func GetNewProducts(c *gin.Context) {
	limit := c.DefaultQuery("limit", "10")
	limitNum, _ := strconv.Atoi(limit)
	if limitNum < 1 || limitNum > 50 {
		limitNum = 10
	}

	ctx := context.Background()
	cacheKey := "products:new:sku:" + strconv.Itoa(limitNum)

	// Try cache first
	if cachedProducts, err := cache.Get(ctx, cacheKey); err == nil {
		var products []models.Product
		if err := json.Unmarshal([]byte(cachedProducts), &products); err == nil {
			c.JSON(http.StatusOK, gin.H{
				"data": products,
			})
			return
		}
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	q := listSKUProductsBaseQuery + ` ORDER BY s.created_at DESC LIMIT $1`
	rows, err := db.Query(q, limitNum)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		p, scanErr := productFromSKU(rows)
		if scanErr != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		products = append(products, p)
	}

	// Cache for 5 minutes
	if productsJSON, err := json.Marshal(products); err == nil {
		cache.Set(ctx, cacheKey, string(productsJSON), 5*60)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": products,
	})
}

// GetCategories retrieves product categories
func GetCategories(c *gin.Context) {
	ctx := context.Background()
	cacheKey := "categories:sku:all"

	// Try cache first
	if cachedCategories, err := cache.Get(ctx, cacheKey); err == nil {
		var categories []map[string]interface{}
		if err := json.Unmarshal([]byte(cachedCategories), &categories); err == nil {
			c.JSON(http.StatusOK, gin.H{
				"data": categories,
			})
			return
		}
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	rows, err := db.Query(
		`SELECT COALESCE(sp.model_tier, '未分类') AS category, COUNT(*) AS count
		 FROM skus s JOIN spus sp ON s.spu_id = sp.id
		 WHERE s.status = 'active' AND sp.status = 'active'
		 GROUP BY sp.model_tier
		 ORDER BY count DESC`,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	var categories []map[string]interface{}
	for rows.Next() {
		var category string
		var count int
		err := rows.Scan(&category, &count)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		categories = append(categories, map[string]interface{}{
			"name":  category,
			"count": count,
		})
	}

	// Cache for 30 minutes
	if categoriesJSON, err := json.Marshal(categories); err == nil {
		cache.Set(ctx, cacheKey, string(categoriesJSON), 30*60)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": categories,
	})
}

// GetHomeData retrieves all data needed for homepage
func GetHomeData(c *gin.Context) {
	ctx := context.Background()
	cacheKey := "home:data:sku:v3"

	// Try cache first
	if cachedData, err := cache.Get(ctx, cacheKey); err == nil {
		var homeData map[string]interface{}
		if err := json.Unmarshal([]byte(cachedData), &homeData); err == nil {
			c.JSON(http.StatusOK, homeData)
			return
		}
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	hotRows, err := db.Query(
		listSKUProductsBaseQuery + ` ORDER BY s.sales_count DESC NULLS LAST, s.created_at DESC LIMIT 10`,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer hotRows.Close()

	var hotProducts []models.Product
	for hotRows.Next() {
		p, scanErr := productFromSKU(hotRows)
		if scanErr != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		hotProducts = append(hotProducts, p)
	}

	newRows, err := db.Query(
		listSKUProductsBaseQuery + ` ORDER BY s.created_at DESC LIMIT 10`,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer newRows.Close()

	var newProducts []models.Product
	for newRows.Next() {
		p, scanErr2 := productFromSKU(newRows)
		if scanErr2 != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		newProducts = append(newProducts, p)
	}

	catRows, err := db.Query(
		`SELECT COALESCE(sp.model_tier, '未分类') AS category, COUNT(*) AS count
		 FROM skus s JOIN spus sp ON s.spu_id = sp.id
		 WHERE s.status = 'active' AND sp.status = 'active'
		 GROUP BY sp.model_tier
		 ORDER BY count DESC`,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer catRows.Close()

	var categories []map[string]interface{}
	for catRows.Next() {
		var category string
		var count int
		err := catRows.Scan(&category, &count)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		categories = append(categories, map[string]interface{}{
			"name":  category,
			"count": count,
		})
	}

	var scenarioCategories []map[string]interface{}
	scenarioRows, errSc := db.Query(`
		SELECT us.code, us.name, COUNT(DISTINCT s.id)::bigint AS cnt
		FROM usage_scenarios us
		INNER JOIN spu_scenarios ss ON us.id = ss.scenario_id
		INNER JOIN spus sp ON ss.spu_id = sp.id AND sp.status = 'active'
		INNER JOIN skus s ON s.spu_id = sp.id AND s.status = 'active' AND (s.stock > 0 OR s.stock = -1)
		WHERE us.status = 'active'
		GROUP BY us.id, us.code, us.name, us.sort_order
		HAVING COUNT(DISTINCT s.id) > 0
		ORDER BY us.sort_order ASC
	`)
	if errSc == nil {
		defer scenarioRows.Close()
		for scenarioRows.Next() {
			var code, name string
			var cnt int64
			if err := scenarioRows.Scan(&code, &name, &cnt); err != nil {
				break
			}
			scenarioCategories = append(scenarioCategories, map[string]interface{}{
				"code":  code,
				"name":  name,
				"count": int(cnt),
			})
		}
	}

	// Get banners (placeholder for now)
	banners := []map[string]interface{}{
		{"id": 1, "title": "新人专享", "image": "/banners/newuser.png", "link": "/catalog?category=newuser"},
		{"id": 2, "title": "限时特惠", "image": "/banners/sale.png", "link": "/catalog?category=sale"},
		{"id": 3, "title": "热门推荐", "image": "/banners/hot.png", "link": "/catalog?sort=hot"},
	}

	homeData := map[string]interface{}{
		"banners":             banners,
		"hot":                 hotProducts,
		"new":                 newProducts,
		"categories":          categories,
		"scenario_categories": scenarioCategories,
	}

	// Cache for 5 minutes
	if homeDataJSON, err := json.Marshal(homeData); err == nil {
		cache.Set(ctx, cacheKey, string(homeDataJSON), 5*60)
	}

	c.JSON(http.StatusOK, homeData)
}
