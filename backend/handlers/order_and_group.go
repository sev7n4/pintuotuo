package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/cache"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/models"
)

const orderStatusPending = "pending"

// CreateOrder creates a new order
func CreateOrder(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	var req struct {
		ProductID int `json:"product_id" binding:"required"`
		GroupID   int `json:"group_id"`
		Quantity  int `json:"quantity" binding:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	// Get product info
	var product models.Product
	err := db.QueryRow(
		"SELECT id, price, stock FROM products WHERE id = $1",
		req.ProductID,
	).Scan(&product.ID, &product.Price, &product.Stock)

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrProductNotFound)
		return
	}

	if product.Stock < req.Quantity {
		middleware.RespondWithError(c, apperrors.ErrInsufficientStock)
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

	result, err := tx.Exec(
		"UPDATE products SET stock = stock - $1 WHERE id = $2 AND stock >= $1",
		req.Quantity, req.ProductID,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"STOCK_UPDATE_FAILED",
			"Failed to update stock",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		middleware.RespondWithError(c, apperrors.ErrInsufficientStock)
		return
	}

	var order models.Order
	totalPrice := product.Price * float64(req.Quantity)

	err = tx.QueryRow(
		"INSERT INTO orders (user_id, product_id, group_id, quantity, unit_price, total_price, status) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, user_id, product_id, group_id, quantity, unit_price, total_price, status, created_at, updated_at",
		userID, req.ProductID, req.GroupID, req.Quantity, product.Price, totalPrice, orderStatusPending,
	).Scan(&order.ID, &order.UserID, &order.ProductID, &order.GroupID, &order.Quantity, &order.UnitPrice, &order.TotalPrice, &order.Status, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"ORDER_CREATION_FAILED",
			"Failed to create order",
			http.StatusInternalServerError,
			err,
		))
		return
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

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "success",
		"data":    order,
	})
}

// ListOrders lists all orders for current user
func ListOrders(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
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
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	rows, err := db.Query(
		"SELECT id, user_id, product_id, group_id, quantity, total_price, status, created_at, updated_at FROM orders WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3",
		userID, perPageNum, offset,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		err := rows.Scan(&o.ID, &o.UserID, &o.ProductID, &o.GroupID, &o.Quantity, &o.TotalPrice, &o.Status, &o.CreatedAt, &o.UpdatedAt)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		orders = append(orders, o)
	}

	var total int
	db.QueryRow("SELECT COUNT(*) FROM orders WHERE user_id = $1", userID).Scan(&total)

	c.JSON(http.StatusOK, gin.H{
		"total":    total,
		"page":     pageNum,
		"per_page": perPageNum,
		"data":     orders,
	})
}

// GetOrderByID retrieves an order by ID
func GetOrderByID(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	id := c.Param("id")
	orderID, _ := strconv.Atoi(id)
	ctx := context.Background()

	cacheKey := cache.OrderKey(orderID)
	if cachedOrder, err := cache.Get(ctx, cacheKey); err == nil {
		var order models.Order
		if err := json.Unmarshal([]byte(cachedOrder), &order); err == nil {
			if order.UserID == userID.(int) {
				c.JSON(http.StatusOK, order)
				return
			}
		}
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var order models.Order
	err := db.QueryRow(
		"SELECT id, user_id, product_id, group_id, quantity, total_price, status, created_at, updated_at FROM orders WHERE id = $1 AND user_id = $2",
		id, userID,
	).Scan(&order.ID, &order.UserID, &order.ProductID, &order.GroupID, &order.Quantity, &order.TotalPrice, &order.Status, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrOrderNotFound)
		return
	}

	if orderJSON, err := json.Marshal(order); err == nil {
		cache.Set(ctx, cacheKey, string(orderJSON), cache.OrderCacheTTL)
	}

	c.JSON(http.StatusOK, order)
}

// CancelOrder cancels an order
func CancelOrder(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	id := c.Param("id")

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var orderInfo struct {
		status    string
		productID int
		quantity  int
	}
	err := db.QueryRow(
		"SELECT status, product_id, quantity FROM orders WHERE id = $1 AND user_id = $2",
		id, userID,
	).Scan(&orderInfo.status, &orderInfo.productID, &orderInfo.quantity)

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrOrderNotFound)
		return
	}

	if orderInfo.status != orderStatusPending {
		middleware.RespondWithError(c, apperrors.ErrCannotCancelOrder)
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

	var order models.Order
	err = tx.QueryRow(
		"UPDATE orders SET status = $1 WHERE id = $2 AND user_id = $3 RETURNING id, user_id, product_id, group_id, quantity, total_price, status, created_at, updated_at",
		"canceled", id, userID,
	).Scan(&order.ID, &order.UserID, &order.ProductID, &order.GroupID, &order.Quantity, &order.TotalPrice, &order.Status, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"ORDER_CANCEL_FAILED",
			"Failed to cancel order",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	_, err = tx.Exec(
		"UPDATE products SET stock = stock + $1 WHERE id = $2",
		orderInfo.quantity, orderInfo.productID,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"STOCK_RESTORE_FAILED",
			"Failed to restore stock",
			http.StatusInternalServerError,
			err,
		))
		return
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

	cache.Delete(context.Background(), cache.OrderKey(order.ID))

	c.JSON(http.StatusOK, order)
}

// CreateGroup creates a new group purchase
func CreateGroup(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	var req struct {
		ProductID   int       `json:"product_id" binding:"required"`
		TargetCount int       `json:"target_count" binding:"required,gt=0"`
		Deadline    time.Time `json:"deadline" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	// Validate deadline is in future
	if req.Deadline.Before(time.Now()) {
		middleware.RespondWithError(c, apperrors.ErrInvalidGroupData)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	// Verify product exists
	var productID int
	err := db.QueryRow("SELECT id FROM products WHERE id = $1", req.ProductID).Scan(&productID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrProductNotFound)
		return
	}

	var group models.Group
	err = db.QueryRow(
		"INSERT INTO groups (product_id, creator_id, target_count, current_count, status, deadline) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, product_id, creator_id, target_count, current_count, status, deadline, created_at, updated_at",
		req.ProductID, userID, req.TargetCount, 1, "active", req.Deadline,
	).Scan(&group.ID, &group.ProductID, &group.CreatorID, &group.TargetCount, &group.CurrentCount, &group.Status, &group.Deadline, &group.CreatedAt, &group.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"GROUP_CREATION_FAILED",
			"Failed to create group",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	c.JSON(http.StatusCreated, group)
}

// ListGroups lists all active groups
func ListGroups(c *gin.Context) {
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

	offset := (pageNum - 1) * perPageNum

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	rows, err := db.Query(
		"SELECT id, product_id, creator_id, target_count, current_count, status, deadline, created_at, updated_at FROM groups WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3",
		status, perPageNum, offset,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	var groups []models.Group
	for rows.Next() {
		var g models.Group
		err := rows.Scan(&g.ID, &g.ProductID, &g.CreatorID, &g.TargetCount, &g.CurrentCount, &g.Status, &g.Deadline, &g.CreatedAt, &g.UpdatedAt)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		groups = append(groups, g)
	}

	var total int
	db.QueryRow("SELECT COUNT(*) FROM groups WHERE status = $1", status).Scan(&total)

	c.JSON(http.StatusOK, gin.H{
		"total":    total,
		"page":     pageNum,
		"per_page": perPageNum,
		"data":     groups,
	})
}

// GetGroupByID retrieves a group by ID
func GetGroupByID(c *gin.Context) {
	id := c.Param("id")
	groupID, _ := strconv.Atoi(id)
	ctx := context.Background()

	cacheKey := cache.GroupKey(groupID)
	if cachedGroup, err := cache.Get(ctx, cacheKey); err == nil {
		var group models.Group
		if err := json.Unmarshal([]byte(cachedGroup), &group); err == nil {
			c.JSON(http.StatusOK, group)
			return
		}
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var group models.Group
	err := db.QueryRow(
		"SELECT id, product_id, creator_id, target_count, current_count, status, deadline, created_at, updated_at FROM groups WHERE id = $1",
		id,
	).Scan(&group.ID, &group.ProductID, &group.CreatorID, &group.TargetCount, &group.CurrentCount, &group.Status, &group.Deadline, &group.CreatedAt, &group.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrGroupNotFound)
		return
	}

	if groupJSON, err := json.Marshal(group); err == nil {
		cache.Set(ctx, cacheKey, string(groupJSON), 5*time.Minute)
	}

	c.JSON(http.StatusOK, group)
}

// JoinGroup adds current user to a group
func JoinGroup(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	groupID := c.Param("id")

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	// Get group info
	var group models.Group
	err := db.QueryRow(
		"SELECT id, product_id, target_count, current_count, status, deadline FROM groups WHERE id = $1",
		groupID,
	).Scan(&group.ID, &group.ProductID, &group.TargetCount, &group.CurrentCount, &group.Status, &group.Deadline)

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrGroupNotFound)
		return
	}

	if group.Status != "active" {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"GROUP_INACTIVE",
			"Group is not active",
			http.StatusConflict,
			nil,
		))
		return
	}

	if group.CurrentCount >= group.TargetCount {
		middleware.RespondWithError(c, apperrors.ErrGroupFull)
		return
	}

	if time.Now().After(group.Deadline) {
		middleware.RespondWithError(c, apperrors.ErrGroupExpired)
		return
	}

	// Create order for group
	var product models.Product
	err = db.QueryRow("SELECT price FROM products WHERE id = $1", group.ProductID).Scan(&product.Price)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrProductNotFound)
		return
	}

	var orderID int
	err = db.QueryRow(
		"INSERT INTO orders (user_id, product_id, group_id, quantity, unit_price, total_price, status) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		userID, group.ProductID, group.ID, 1, product.Price, product.Price, orderStatusPending,
	).Scan(&orderID)

	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"ORDER_CREATION_FAILED",
			"Failed to create order",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	// Add to group members
	_, err = db.Exec(
		"INSERT INTO group_members (group_id, user_id, order_id) VALUES ($1, $2, $3)",
		group.ID, userID, orderID,
	)

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrAlreadyInGroup)
		return
	}

	// Update group count
	newCount := group.CurrentCount + 1
	newStatus := group.Status
	if newCount >= group.TargetCount {
		newStatus = "completed"
	}

	err = db.QueryRow(
		"UPDATE groups SET current_count = $1, status = $2 WHERE id = $3 RETURNING id, product_id, creator_id, target_count, current_count, status, deadline, created_at, updated_at",
		newCount, newStatus, group.ID,
	).Scan(&group.ID, &group.ProductID, &group.CreatorID, &group.TargetCount, &group.CurrentCount, &group.Status, &group.Deadline, &group.CreatedAt, &group.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"GROUP_UPDATE_FAILED",
			"Failed to update group",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	cache.Delete(context.Background(), cache.GroupKey(group.ID))

	c.JSON(http.StatusOK, group)
}

// CancelGroup cancels a group (creator only)
func CancelGroup(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	id := c.Param("id")

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	// Verify creator
	var creatorID int
	err := db.QueryRow("SELECT creator_id FROM groups WHERE id = $1", id).Scan(&creatorID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrGroupNotFound)
		return
	}

	userIDInt, ok := userID.(int)
	if !ok || creatorID != userIDInt {
		middleware.RespondWithError(c, apperrors.ErrForbidden)
		return
	}

	_, err = db.Exec("DELETE FROM groups WHERE id = $1", id)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"GROUP_DELETE_FAILED",
			"Failed to cancel group",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	groupID, _ := strconv.Atoi(id)
	cache.Delete(context.Background(), cache.GroupKey(groupID))

	c.JSON(http.StatusOK, gin.H{"message": "Group canceled successfully"})
}

// GetGroupProgress retrieves group progress
func GetGroupProgress(c *gin.Context) {
	id := c.Param("id")

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var group models.Group
	err := db.QueryRow(
		"SELECT id, product_id, creator_id, target_count, current_count, status, deadline, created_at, updated_at FROM groups WHERE id = $1",
		id,
	).Scan(&group.ID, &group.ProductID, &group.CreatorID, &group.TargetCount, &group.CurrentCount, &group.Status, &group.Deadline, &group.CreatedAt, &group.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrGroupNotFound)
		return
	}

	c.JSON(http.StatusOK, group)
}
