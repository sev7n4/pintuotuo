package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/models"
)

// ListProducts retrieves product list with pagination
func ListProducts(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("per_page", "20")
	status := c.DefaultQuery("status", "")

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

	// Build query
	query := "SELECT id, merchant_id, name, description, price, stock, status, created_at, updated_at FROM products WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3"
	if status == "" {
		query = "SELECT id, merchant_id, name, description, price, stock, status, created_at, updated_at FROM products WHERE status = 'active' ORDER BY created_at DESC LIMIT $2 OFFSET $3"
	}

	rows, err := db.Query(query, status, perPageNum, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		err := rows.Scan(&p.ID, &p.MerchantID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.Status, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan product"})
			return
		}
		products = append(products, p)
	}

	// Get total count
	var total int
	countQuery := "SELECT COUNT(*) FROM products WHERE status = $1"
	if status == "" {
		countQuery = "SELECT COUNT(*) FROM products WHERE status = 'active'"
	}
	db.QueryRow(countQuery, status).Scan(&total)

	c.JSON(http.StatusOK, gin.H{
		"total":    total,
		"page":     pageNum,
		"per_page": perPageNum,
		"data":     products,
	})
}

// GetProductByID retrieves a single product by ID
func GetProductByID(c *gin.Context) {
	id := c.Param("id")

	db := config.GetDB()

	var product models.Product
	err := db.QueryRow(
		"SELECT id, merchant_id, name, description, price, stock, status, created_at, updated_at FROM products WHERE id = $1",
		id,
	).Scan(&product.ID, &product.MerchantID, &product.Name, &product.Description, &product.Price, &product.Stock, &product.Status, &product.CreatedAt, &product.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, product)
}

// SearchProducts searches for products by query
func SearchProducts(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query required"})
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

	// Search in product name and description
	searchQuery := "%" + query + "%"
	rows, err := db.Query(
		"SELECT id, merchant_id, name, description, price, stock, status, created_at, updated_at FROM products WHERE status = 'active' AND (name ILIKE $1 OR description ILIKE $1) ORDER BY created_at DESC LIMIT $2 OFFSET $3",
		searchQuery, perPageNum, offset,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search products"})
		return
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		err := rows.Scan(&p.ID, &p.MerchantID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.Status, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan product"})
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

	c.JSON(http.StatusOK, gin.H{
		"total":    total,
		"page":     pageNum,
		"per_page": perPageNum,
		"data":     products,
	})
}

// CreateProduct creates a new product (merchant only)
func CreateProduct(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := config.GetDB()

	// Verify user is a merchant
	var role string
	err := db.QueryRow("SELECT role FROM users WHERE id = $1", userID).Scan(&role)
	if err != nil || role != "merchant" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only merchants can create products"})
		return
	}

	var product models.Product
	err = db.QueryRow(
		"INSERT INTO products (merchant_id, name, description, price, original_price, stock, status) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, merchant_id, name, description, price, stock, status, created_at, updated_at",
		userID, req.Name, req.Description, req.Price, req.OriginalPrice, req.Stock, "active",
	).Scan(&product.ID, &product.MerchantID, &product.Name, &product.Description, &product.Price, &product.Stock, &product.Status, &product.CreatedAt, &product.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}

	c.JSON(http.StatusCreated, product)
}

// UpdateProduct updates a product (merchant only)
func UpdateProduct(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := config.GetDB()

	// Verify ownership
	var merchantID int
	err := db.QueryRow("SELECT merchant_id FROM products WHERE id = $1", id).Scan(&merchantID)
	if err != nil || merchantID != userID.(int) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own products"})
		return
	}

	var product models.Product
	err = db.QueryRow(
		"UPDATE products SET name = COALESCE(NULLIF($1, ''), name), description = COALESCE(NULLIF($2, ''), description), price = CASE WHEN $3 > 0 THEN $3 ELSE price END, original_price = CASE WHEN $4 > 0 THEN $4 ELSE original_price END, stock = CASE WHEN $5 >= 0 THEN $5 ELSE stock END, status = COALESCE(NULLIF($6, ''), status) WHERE id = $7 RETURNING id, merchant_id, name, description, price, stock, status, created_at, updated_at",
		req.Name, req.Description, req.Price, req.OriginalPrice, req.Stock, req.Status, id,
	).Scan(&product.ID, &product.MerchantID, &product.Name, &product.Description, &product.Price, &product.Stock, &product.Status, &product.CreatedAt, &product.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		return
	}

	c.JSON(http.StatusOK, product)
}

// DeleteProduct deletes a product (merchant only)
func DeleteProduct(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id := c.Param("id")

	db := config.GetDB()

	// Verify ownership
	var merchantID int
	err := db.QueryRow("SELECT merchant_id FROM products WHERE id = $1", id).Scan(&merchantID)
	if err != nil || merchantID != userID.(int) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own products"})
		return
	}

	_, err = db.Exec("DELETE FROM products WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}
