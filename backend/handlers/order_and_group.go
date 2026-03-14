package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/models"
)

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

	// Create order
	var order models.Order
	totalPrice := product.Price * float64(req.Quantity)

	err = db.QueryRow(
		"INSERT INTO orders (user_id, product_id, group_id, quantity, unit_price, total_price, status) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, user_id, product_id, group_id, quantity, total_price, status, created_at, updated_at",
		userID, req.ProductID, req.GroupID, req.Quantity, product.Price, totalPrice, "pending",
	).Scan(&order.ID, &order.UserID, &order.ProductID, &order.GroupID, &order.Quantity, &order.TotalPrice, &order.Status, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"ORDER_CREATION_FAILED",
			"Failed to create order",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	c.JSON(http.StatusCreated, order)
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

	db := config.GetDB()

	var order models.Order
	err := db.QueryRow(
		"SELECT id, user_id, product_id, group_id, quantity, total_price, status, created_at, updated_at FROM orders WHERE id = $1 AND user_id = $2",
		id, userID,
	).Scan(&order.ID, &order.UserID, &order.ProductID, &order.GroupID, &order.Quantity, &order.TotalPrice, &order.Status, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrOrderNotFound)
		return
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

	// Check order status
	var status string
	err := db.QueryRow(
		"SELECT status FROM orders WHERE id = $1 AND user_id = $2",
		id, userID,
	).Scan(&status)

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrOrderNotFound)
		return
	}

	if status != "pending" {
		middleware.RespondWithError(c, apperrors.ErrCannotCancelOrder)
		return
	}

	var order models.Order
	err = db.QueryRow(
		"UPDATE orders SET status = $1 WHERE id = $2 AND user_id = $3 RETURNING id, user_id, product_id, group_id, quantity, total_price, status, created_at, updated_at",
		"cancelled", id, userID,
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
		ProductID  int       `json:"product_id" binding:"required"`
		TargetCount int      `json:"target_count" binding:"required,gt=0"`
		Deadline   time.Time `json:"deadline" binding:"required"`
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

	rows, err := db.Query(
		"SELECT id, product_id, creator_id, target_count, current_count, status, deadline, created_at, updated_at FROM groups WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3",
		status, perPageNum, offset,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch groups"})
		return
	}
	defer rows.Close()

	var groups []models.Group
	for rows.Next() {
		var g models.Group
		err := rows.Scan(&g.ID, &g.ProductID, &g.CreatorID, &g.TargetCount, &g.CurrentCount, &g.Status, &g.Deadline, &g.CreatedAt, &g.UpdatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan group"})
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

	db := config.GetDB()

	var group models.Group
	err := db.QueryRow(
		"SELECT id, product_id, creator_id, target_count, current_count, status, deadline, created_at, updated_at FROM groups WHERE id = $1",
		id,
	).Scan(&group.ID, &group.ProductID, &group.CreatorID, &group.TargetCount, &group.CurrentCount, &group.Status, &group.Deadline, &group.CreatedAt, &group.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	c.JSON(http.StatusOK, group)
}

// JoinGroup adds current user to a group
func JoinGroup(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	groupID := c.Param("id")

	db := config.GetDB()

	// Get group info
	var group models.Group
	err := db.QueryRow(
		"SELECT id, product_id, target_count, current_count, status, deadline FROM groups WHERE id = $1",
		groupID,
	).Scan(&group.ID, &group.ProductID, &group.TargetCount, &group.CurrentCount, &group.Status, &group.Deadline)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	if group.Status != "active" {
		c.JSON(http.StatusConflict, gin.H{"error": "Group is not active"})
		return
	}

	if group.CurrentCount >= group.TargetCount {
		c.JSON(http.StatusConflict, gin.H{"error": "Group is already full"})
		return
	}

	if time.Now().After(group.Deadline) {
		c.JSON(http.StatusConflict, gin.H{"error": "Group deadline has passed"})
		return
	}

	// Create order for group
	var product models.Product
	err = db.QueryRow("SELECT price FROM products WHERE id = $1", group.ProductID).Scan(&product.Price)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get product"})
		return
	}

	var orderID int
	err = db.QueryRow(
		"INSERT INTO orders (user_id, product_id, group_id, quantity, unit_price, total_price, status) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		userID, group.ProductID, group.ID, 1, product.Price, product.Price, "pending",
	).Scan(&orderID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	// Add to group members
	_, err = db.Exec(
		"INSERT INTO group_members (group_id, user_id, order_id) VALUES ($1, $2, $3)",
		group.ID, userID, orderID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join group"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update group"})
		return
	}

	c.JSON(http.StatusOK, group)
}

// CancelGroup cancels a group (creator only)
func CancelGroup(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id := c.Param("id")

	db := config.GetDB()

	// Verify creator
	var creatorID int
	err := db.QueryRow("SELECT creator_id FROM groups WHERE id = $1", id).Scan(&creatorID)
	if err != nil || creatorID != userID.(int) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only creator can cancel group"})
		return
	}

	_, err = db.Exec("DELETE FROM groups WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel group"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Group cancelled successfully"})
}

// GetGroupProgress retrieves group progress
func GetGroupProgress(c *gin.Context) {
	id := c.Param("id")

	db := config.GetDB()

	var group models.Group
	err := db.QueryRow(
		"SELECT id, product_id, creator_id, target_count, current_count, status, deadline, created_at, updated_at FROM groups WHERE id = $1",
		id,
	).Scan(&group.ID, &group.ProductID, &group.CreatorID, &group.TargetCount, &group.CurrentCount, &group.Status, &group.Deadline, &group.CreatedAt, &group.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	c.JSON(http.StatusOK, group)
}
