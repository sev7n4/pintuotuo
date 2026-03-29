package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/models"
)

func GetCart(c *gin.Context) {
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

	query := `
		SELECT ci.id, ci.product_id, ci.sku_id, ci.spu_id, ci.group_id, ci.quantity,
			   COALESCE(p.id, pv.id) as prod_id,
			   COALESCE(p.merchant_id, pv.merchant_id) as merchant_id,
			   COALESCE(p.name, pv.name) as name,
			   COALESCE(p.description, pv.description) as description,
			   COALESCE(p.price, pv.price) as price,
			   COALESCE(p.original_price, pv.original_price) as original_price,
			   COALESCE(p.stock, pv.stock) as stock,
			   COALESCE(p.sold_count, pv.sold_count) as sold_count,
			   COALESCE(p.category, pv.category) as category,
			   COALESCE(p.status, pv.status) as status,
			   COALESCE(p.created_at, pv.created_at) as created_at,
			   COALESCE(p.updated_at, pv.updated_at) as updated_at
		FROM cart_items ci
		LEFT JOIN products p ON ci.product_id = p.id
		LEFT JOIN products_v2 pv ON ci.sku_id = pv.id
		WHERE ci.user_id = $1
		ORDER BY ci.created_at DESC
	`

	rows, err := db.Query(query, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Failed to fetch cart items",
		})
		return
	}
	defer rows.Close()

	var items []models.CartResponse
	var totalPrice float64

	for rows.Next() {
		var item models.CartResponse
		var product models.Product
		var groupID, skuID, spuID sql.NullInt64

		err := rows.Scan(
			&item.ID, &item.ProductID, &skuID, &spuID, &groupID, &item.Quantity,
			&product.ID, &product.MerchantID, &product.Name, &product.Description,
			&product.Price, &product.OriginalPrice, &product.Stock, &product.SoldCount,
			&product.Category, &product.Status, &product.CreatedAt, &product.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if groupID.Valid {
			item.GroupID = int(groupID.Int64)
		}
		item.Product = product
		items = append(items, item)
		totalPrice += product.Price * float64(item.Quantity)
	}

	if items == nil {
		items = []models.CartResponse{}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": models.CartSummary{
			Items:      items,
			TotalItems: len(items),
			TotalPrice: totalPrice,
		},
	})
}

func AddToCart(c *gin.Context) {
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
		SKUID     int `json:"sku_id"`
		Quantity  int `json:"quantity"`
		GroupID   int `json:"group_id,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
		return
	}

	if req.ProductID == 0 && req.SKUID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "product_id or sku_id is required",
		})
		return
	}

	if req.Quantity <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "quantity must be greater than 0",
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

	var spuID int
	if req.SKUID > 0 {
		err := db.QueryRow(
			`SELECT s.spu_id FROM skus s JOIN spus sp ON s.spu_id = sp.id 
			 WHERE s.id = $1 AND s.status = 'active' AND sp.status = 'active'`,
			req.SKUID,
		).Scan(&spuID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    "SKU_NOT_FOUND",
				"message": "SKU not found or inactive",
			})
			return
		}
	}

	var existingID int
	var existingQty int
	var checkQuery string
	var err error

	if req.SKUID > 0 {
		checkQuery = `SELECT id, quantity FROM cart_items WHERE user_id = $1 AND sku_id = $2 AND (group_id = $3 OR (group_id IS NULL AND $3 = 0))`
		err = db.QueryRow(checkQuery, userID, req.SKUID, req.GroupID).Scan(&existingID, &existingQty)
	} else {
		checkQuery = `SELECT id, quantity FROM cart_items WHERE user_id = $1 AND product_id = $2 AND sku_id IS NULL AND (group_id = $3 OR (group_id IS NULL AND $3 = 0))`
		err = db.QueryRow(checkQuery, userID, req.ProductID, req.GroupID).Scan(&existingID, &existingQty)
	}

	if err == nil {
		updateQuery := `UPDATE cart_items SET quantity = $1, updated_at = NOW() WHERE id = $2`
		_, err = db.Exec(updateQuery, existingQty+req.Quantity, existingID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to update cart item",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "Cart item updated",
			"data": gin.H{
				"id":         existingID,
				"product_id": req.ProductID,
				"sku_id":     req.SKUID,
				"quantity":   existingQty + req.Quantity,
				"group_id":   req.GroupID,
			},
		})
		return
	}

	var insertQuery string
	var newID int

	if req.SKUID > 0 {
		insertQuery = `
			INSERT INTO cart_items (user_id, sku_id, spu_id, group_id, quantity, created_at, updated_at)
			VALUES ($1, $2, $3, NULLIF($4, 0), $5, NOW(), NOW())
			RETURNING id
		`
		err = db.QueryRow(insertQuery, userID, req.SKUID, spuID, req.GroupID, req.Quantity).Scan(&newID)
	} else {
		insertQuery = `
			INSERT INTO cart_items (user_id, product_id, group_id, quantity, created_at, updated_at)
			VALUES ($1, $2, NULLIF($3, 0), $4, NOW(), NOW())
			RETURNING id
		`
		err = db.QueryRow(insertQuery, userID, req.ProductID, req.GroupID, req.Quantity).Scan(&newID)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Failed to add item to cart",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "Item added to cart",
		"data": gin.H{
			"id":         newID,
			"product_id": req.ProductID,
			"sku_id":     req.SKUID,
			"quantity":   req.Quantity,
			"group_id":   req.GroupID,
		},
	})
}

func UpdateCartItem(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "UNAUTHORIZED",
			"message": "User not authenticated",
		})
		return
	}

	itemIDStr := c.Param("id")
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "NOT_FOUND",
			"message": "Cart item not found",
		})
		return
	}

	var req struct {
		Quantity int `json:"quantity"`
	}

	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
		return
	}

	if req.Quantity <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "quantity must be greater than 0",
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

	query := `UPDATE cart_items SET quantity = $1, updated_at = NOW() WHERE id = $2 AND user_id = $3`
	result, err := db.Exec(query, req.Quantity, itemID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Failed to update cart item",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "NOT_FOUND",
			"message": "Cart item not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Cart item updated",
		"data": gin.H{
			"id":       itemID,
			"quantity": req.Quantity,
		},
	})
}

func RemoveFromCart(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "UNAUTHORIZED",
			"message": "User not authenticated",
		})
		return
	}

	itemIDStr := c.Param("id")
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "NOT_FOUND",
			"message": "Cart item not found",
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

	query := `DELETE FROM cart_items WHERE id = $1 AND user_id = $2`
	result, err := db.Exec(query, itemID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Failed to remove cart item",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "NOT_FOUND",
			"message": "Cart item not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Cart item removed",
	})
}

func ClearCart(c *gin.Context) {
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

	query := `DELETE FROM cart_items WHERE user_id = $1`
	_, err := db.Exec(query, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Failed to clear cart",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Cart cleared",
	})
}
