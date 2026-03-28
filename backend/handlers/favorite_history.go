package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/models"
)

func GetFavorites(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "UNAUTHORIZED",
			"message": "User not authenticated",
		})
		return
	}

	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Database connection error",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	countQuery := `SELECT COUNT(*) FROM favorites WHERE user_id = $1`
	var total int
	err := db.QueryRow(countQuery, userID).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Failed to count favorites",
		})
		return
	}

	query := `
		SELECT f.id, f.product_id, f.created_at,
			   p.id, p.merchant_id, p.name, p.description, p.price, p.original_price, 
			   p.stock, p.sold_count, p.category, p.status, p.created_at, p.updated_at
		FROM favorites f
		JOIN products p ON f.product_id = p.id
		WHERE f.user_id = $1
		ORDER BY f.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := db.Query(query, userID, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Failed to fetch favorites",
		})
		return
	}
	defer rows.Close()

	var items []models.FavoriteResponse
	for rows.Next() {
		var item models.FavoriteResponse
		var product models.Product
		var createdAt sql.NullTime

		err := rows.Scan(
			&item.ID, &item.ProductID, &createdAt,
			&product.ID, &product.MerchantID, &product.Name, &product.Description,
			&product.Price, &product.OriginalPrice, &product.Stock, &product.SoldCount,
			&product.Category, &product.Status, &product.CreatedAt, &product.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if createdAt.Valid {
			item.CreatedAt = createdAt.Time.Format("2006-01-02T15:04:05Z07:00")
		}
		item.Product = product
		items = append(items, item)
	}

	if items == nil {
		items = []models.FavoriteResponse{}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items":      items,
			"total":      total,
			"page":       page,
			"page_size":  pageSize,
			"total_page": (total + pageSize - 1) / pageSize,
		},
	})
}

func AddFavorite(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "UNAUTHORIZED",
			"message": "User not authenticated",
		})
		return
	}

	var req struct {
		ProductID int `json:"product_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
		return
	}

	if req.ProductID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "product_id is required",
		})
		return
	}

	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Database connection error",
		})
		return
	}

	var productExists bool
	err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)`, req.ProductID).Scan(&productExists)
	if err != nil || !productExists {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "NOT_FOUND",
			"message": "Product not found",
		})
		return
	}

	var existingID int
	err = db.QueryRow(`SELECT id FROM favorites WHERE user_id = $1 AND product_id = $2`, userID, req.ProductID).Scan(&existingID)
	if err == nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "Already in favorites",
			"data": gin.H{
				"id":         existingID,
				"product_id": req.ProductID,
			},
		})
		return
	}

	insertQuery := `
		INSERT INTO favorites (user_id, product_id, created_at)
		VALUES ($1, $2, NOW())
		RETURNING id
	`
	var newID int
	err = db.QueryRow(insertQuery, userID, req.ProductID).Scan(&newID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Failed to add favorite",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "Added to favorites",
		"data": gin.H{
			"id":         newID,
			"product_id": req.ProductID,
		},
	})
}

func RemoveFavorite(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "UNAUTHORIZED",
			"message": "User not authenticated",
		})
		return
	}

	productIDStr := c.Param("product_id")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "Invalid product_id",
		})
		return
	}

	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Database connection error",
		})
		return
	}

	query := `DELETE FROM favorites WHERE user_id = $1 AND product_id = $2`
	result, err := db.Exec(query, userID, productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Failed to remove favorite",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "NOT_FOUND",
			"message": "Favorite not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Removed from favorites",
	})
}

func CheckFavorite(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "UNAUTHORIZED",
			"message": "User not authenticated",
		})
		return
	}

	productIDStr := c.Param("product_id")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "Invalid product_id",
		})
		return
	}

	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Database connection error",
		})
		return
	}

	var isFavorite bool
	err = db.QueryRow(`SELECT EXISTS(SELECT 1 FROM favorites WHERE user_id = $1 AND product_id = $2)`, userID, productID).Scan(&isFavorite)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Failed to check favorite",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"is_favorite": isFavorite,
		},
	})
}

func GetBrowseHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "UNAUTHORIZED",
			"message": "User not authenticated",
		})
		return
	}

	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Database connection error",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	countQuery := `SELECT COUNT(*) FROM browse_history WHERE user_id = $1`
	var total int
	err := db.QueryRow(countQuery, userID).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Failed to count browse history",
		})
		return
	}

	query := `
		SELECT bh.id, bh.product_id, bh.view_count, bh.updated_at,
			   p.id, p.merchant_id, p.name, p.description, p.price, p.original_price, 
			   p.stock, p.sold_count, p.category, p.status, p.created_at, p.updated_at
		FROM browse_history bh
		JOIN products p ON bh.product_id = p.id
		WHERE bh.user_id = $1
		ORDER BY bh.updated_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := db.Query(query, userID, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Failed to fetch browse history",
		})
		return
	}
	defer rows.Close()

	var items []models.BrowseHistoryResponse
	for rows.Next() {
		var item models.BrowseHistoryResponse
		var product models.Product
		var updatedAt sql.NullTime

		err := rows.Scan(
			&item.ID, &item.ProductID, &item.ViewCount, &updatedAt,
			&product.ID, &product.MerchantID, &product.Name, &product.Description,
			&product.Price, &product.OriginalPrice, &product.Stock, &product.SoldCount,
			&product.Category, &product.Status, &product.CreatedAt, &product.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if updatedAt.Valid {
			item.ViewedAt = updatedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		}
		item.Product = product
		items = append(items, item)
	}

	if items == nil {
		items = []models.BrowseHistoryResponse{}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items":      items,
			"total":      total,
			"page":       page,
			"page_size":  pageSize,
			"total_page": (total + pageSize - 1) / pageSize,
		},
	})
}

func AddBrowseHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "UNAUTHORIZED",
			"message": "User not authenticated",
		})
		return
	}

	var req struct {
		ProductID int `json:"product_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
		return
	}

	if req.ProductID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "product_id is required",
		})
		return
	}

	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Database connection error",
		})
		return
	}

	var productExists bool
	err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)`, req.ProductID).Scan(&productExists)
	if err != nil || !productExists {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "NOT_FOUND",
			"message": "Product not found",
		})
		return
	}

	upsertQuery := `
		INSERT INTO browse_history (user_id, product_id, view_count, created_at, updated_at)
		VALUES ($1, $2, 1, NOW(), NOW())
		ON CONFLICT (user_id, product_id) 
		DO UPDATE SET view_count = browse_history.view_count + 1, updated_at = NOW()
	`
	_, err = db.Exec(upsertQuery, userID, req.ProductID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Failed to add browse history",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Browse history recorded",
	})
}

func ClearBrowseHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "UNAUTHORIZED",
			"message": "User not authenticated",
		})
		return
	}

	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Database connection error",
		})
		return
	}

	query := `DELETE FROM browse_history WHERE user_id = $1`
	_, err := db.Exec(query, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Failed to clear browse history",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Browse history cleared",
	})
}

func RemoveBrowseHistoryItem(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "UNAUTHORIZED",
			"message": "User not authenticated",
		})
		return
	}

	productIDStr := c.Param("product_id")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "Invalid product_id",
		})
		return
	}

	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Database connection error",
		})
		return
	}

	query := `DELETE FROM browse_history WHERE user_id = $1 AND product_id = $2`
	result, err := db.Exec(query, userID, productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Failed to remove browse history item",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "NOT_FOUND",
			"message": "Browse history item not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Browse history item removed",
	})
}
