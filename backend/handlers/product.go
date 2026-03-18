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

// ListProducts retrieves product list with pagination
func ListProducts(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("per_page", "20")
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
	cacheKey := cache.ProductListKey(pageNum, perPageNum, status)

	// Try cache first
	if cachedList, err := cache.Get(ctx, cacheKey); err == nil {
		var cachedData struct {
			Total   int                `json:"total"`
			Page    int                `json:"page"`
			PerPage int                `json:"per_page"`
			Data    []models.Product   `json:"data"`
		}
		if err := json.Unmarshal([]byte(cachedList), &cachedData); err == nil {
			c.JSON(http.StatusOK, cachedData)
			return
		}
	}

	offset := (pageNum - 1) * perPageNum

	db := config.GetDB()

	// Build query with proper parameter handling for PostgreSQL
	var rows *sql.Rows
	var err error
	if status == "" || status == "all" {
		rows, err = db.Query(
			"SELECT id, merchant_id, name, description, price, stock, status, created_at, updated_at FROM products ORDER BY created_at DESC LIMIT $1 OFFSET $2",
			perPageNum, offset,
		)
	} else {
		rows, err = db.Query(
			"SELECT id, merchant_id, name, description, price, stock, status, created_at, updated_at FROM products WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3",
			status, perPageNum, offset,
		)
	}

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		err := rows.Scan(&p.ID, &p.MerchantID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.Status, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		products = append(products, p)
	}

	// Get total count
	var total int
	var countErr error
	if status == "" || status == "all" {
		countErr = db.QueryRow("SELECT COUNT(*) FROM products").Scan(&total)
	} else {
		countErr = db.QueryRow("SELECT COUNT(*) FROM products WHERE status = $1", status).Scan(&total)
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
			c.JSON(http.StatusOK, product)
			return
		}
	}

	// Query database
	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrProductNotFound)
		return
	}

	var product models.Product
	err := db.QueryRow(
		"SELECT id, merchant_id, name, description, price, stock, status, created_at, updated_at FROM products WHERE id = $1",
		productID,
	).Scan(&product.ID, &product.MerchantID, &product.Name, &product.Description, &product.Price, &product.Stock, &product.Status, &product.CreatedAt, &product.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrProductNotFound)
		return
	}

	// Cache the result
	if productJSON, err := json.Marshal(product); err == nil {
		cache.Set(ctx, cacheKey, string(productJSON), cache.ProductCacheTTL)
	}

	c.JSON(http.StatusOK, product)
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
			Total   int                `json:"total"`
			Page    int                `json:"page"`
			PerPage int                `json:"per_page"`
			Data    []models.Product   `json:"data"`
		}
		if err := json.Unmarshal([]byte(cachedResults), &cachedData); err == nil {
			c.JSON(http.StatusOK, cachedData)
			return
		}
	}

	offset := (pageNum - 1) * perPageNum

	db := config.GetDB()

	// Search in product name and description
	searchQuery := "%" + query + "%"
	rows, err := db.Query(
		"SELECT id, merchant_id, name, description, price, stock, status, created_at, updated_at FROM products WHERE status = 'active' AND (name ILIKE $1 OR description ILIKE $1) ORDER BY created_at DESC LIMIT $2 OFFSET $3",
		searchQuery, perPageNum, offset,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		err := rows.Scan(&p.ID, &p.MerchantID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.Status, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		products = append(products, p)
	}

	// Get total count
	var total int
	db.QueryRow(
		"SELECT COUNT(*) FROM products WHERE status = 'active' AND (name ILIKE $1 OR description ILIKE $1)",
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

// CreateProduct creates a new product (merchant only)
func CreateProduct(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	var req struct {
		Name            string  `json:"name" binding:"required"`
		Description     string  `json:"description"`
		Price           float64 `json:"price" binding:"required,gt=0"`
		OriginalPrice   float64 `json:"original_price"`
		Stock           int     `json:"stock" binding:"required,gte=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidProductData)
		return
	}

	db := config.GetDB()

	// Verify user is a merchant
	var role string
	err := db.QueryRow("SELECT role FROM users WHERE id = $1", userID).Scan(&role)
	if err != nil || role != "merchant" {
		middleware.RespondWithError(c, apperrors.ErrMerchantOnly)
		return
	}

	var product models.Product
	err = db.QueryRow(
		"INSERT INTO products (merchant_id, name, description, price, original_price, stock, status) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, merchant_id, name, description, price, stock, status, created_at, updated_at",
		userID, req.Name, req.Description, req.Price, req.OriginalPrice, req.Stock, "active",
	).Scan(&product.ID, &product.MerchantID, &product.Name, &product.Description, &product.Price, &product.Stock, &product.Status, &product.CreatedAt, &product.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"PRODUCT_CREATION_FAILED",
			"Failed to create product",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	c.JSON(http.StatusCreated, product)
}

// UpdateProduct updates a product (merchant only)
func UpdateProduct(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	id := c.Param("id")

	var req struct {
		Name          string  `json:"name"`
		Description   string  `json:"description"`
		Price         float64 `json:"price"`
		OriginalPrice float64 `json:"original_price"`
		Stock         int     `json:"stock"`
		Status        string  `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()

	// Verify ownership
	var merchantID int
	err := db.QueryRow("SELECT merchant_id FROM products WHERE id = $1", id).Scan(&merchantID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrProductNotFound)
		return
	}

	userIDInt, ok := userID.(int)
	if !ok || merchantID != userIDInt {
		middleware.RespondWithError(c, apperrors.ErrForbidden)
		return
	}

	var product models.Product
	err = db.QueryRow(
		"UPDATE products SET name = COALESCE(NULLIF($1, ''), name), description = COALESCE(NULLIF($2, ''), description), price = CASE WHEN $3 > 0 THEN $3 ELSE price END, original_price = CASE WHEN $4 > 0 THEN $4 ELSE original_price END, stock = CASE WHEN $5 >= 0 THEN $5 ELSE stock END, status = COALESCE(NULLIF($6, ''), status) WHERE id = $7 RETURNING id, merchant_id, name, description, price, stock, status, created_at, updated_at",
		req.Name, req.Description, req.Price, req.OriginalPrice, req.Stock, req.Status, id,
	).Scan(&product.ID, &product.MerchantID, &product.Name, &product.Description, &product.Price, &product.Stock, &product.Status, &product.CreatedAt, &product.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"PRODUCT_UPDATE_FAILED",
			"Failed to update product",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	// Invalidate cache
	ctx := context.Background()
	cache.Delete(ctx, cache.ProductKey(idToInt(id)))
	cache.InvalidatePatterns(ctx, "products:list:*")
	cache.InvalidatePatterns(ctx, "products:search:*")

	c.JSON(http.StatusOK, product)
}

// DeleteProduct deletes a product (merchant only)
func DeleteProduct(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	id := c.Param("id")

	db := config.GetDB()

	// Verify ownership
	var merchantID int
	err := db.QueryRow("SELECT merchant_id FROM products WHERE id = $1", id).Scan(&merchantID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrProductNotFound)
		return
	}

	userIDInt, ok := userID.(int)
	if !ok || merchantID != userIDInt {
		middleware.RespondWithError(c, apperrors.ErrForbidden)
		return
	}

	_, err = db.Exec("DELETE FROM products WHERE id = $1", id)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"PRODUCT_DELETE_FAILED",
			"Failed to delete product",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	// Invalidate cache
	ctx := context.Background()
	cache.Delete(ctx, cache.ProductKey(idToInt(id)))
	cache.InvalidatePatterns(ctx, "products:list:*")
	cache.InvalidatePatterns(ctx, "products:search:*")

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}

// Helper function to convert string ID to int
func idToInt(id string) int {
	idInt, _ := strconv.Atoi(id)
	return idInt
}
