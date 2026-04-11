package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/cache"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/models"
	"github.com/pintuotuo/backend/services"
	"github.com/pintuotuo/backend/utils"
)

const orderStatusPending = "pending"
const groupStatusActive = "active"

// CreateOrder creates a new order
func CreateOrder(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	var req struct {
		ProductID int `json:"product_id"`
		SKUID     int `json:"sku_id"`
		GroupID   int `json:"group_id"`
		Quantity  int `json:"quantity" binding:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	if req.SKUID <= 0 {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"MISSING_SKU",
			"sku_id is required",
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

	var skuID, spuID int
	var retailPrice, wholesalePrice float64
	var stock int
	var skuType string

	var groupEnabled bool
	var minGroupSize, maxGroupSize int
	var groupDiscountRate sql.NullFloat64
	var tokenAmount sql.NullInt64
	var computePoints sql.NullFloat64
	var subscriptionPeriod sql.NullString
	var validDays sql.NullInt64
	var trialDurationDays sql.NullInt64

	err := db.QueryRow(
		`SELECT s.id, s.spu_id, s.retail_price, s.wholesale_price, s.stock, s.sku_type,
			s.group_enabled, s.min_group_size, s.max_group_size, s.group_discount_rate,
			s.token_amount, s.compute_points, s.subscription_period, s.valid_days, s.trial_duration_days
		 FROM skus s JOIN spus sp ON s.spu_id = sp.id 
		 WHERE s.id = $1 AND s.status = 'active' AND sp.status = 'active'`,
		req.SKUID,
	).Scan(&skuID, &spuID, &retailPrice, &wholesalePrice, &stock, &skuType,
		&groupEnabled, &minGroupSize, &maxGroupSize, &groupDiscountRate,
		&tokenAmount, &computePoints, &subscriptionPeriod, &validDays, &trialDurationDays)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrProductNotFound)
		return
	}

	periodStr := ""
	if subscriptionPeriod.Valid {
		periodStr = subscriptionPeriod.String
	}
	trialDays := 0
	if trialDurationDays.Valid {
		trialDays = int(trialDurationDays.Int64)
	}
	tokAmt := int64(0)
	if tokenAmount.Valid {
		tokAmt = tokenAmount.Int64
	}
	cp := 0.0
	if computePoints.Valid {
		cp = computePoints.Float64
	}
	vd := 0
	if validDays.Valid {
		vd = int(validDays.Int64)
	}
	if err = services.ValidateSKUForOrder(skuType, tokAmt, cp, periodStr, vd, trialDays); err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"INVALID_SKU_CONFIG",
			err.Error(),
			http.StatusBadRequest,
			nil,
		))
		return
	}

	if req.GroupID > 0 && !groupEnabled {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"SKU_GROUP_NOT_ENABLED",
			"This SKU does not support group purchase",
			http.StatusBadRequest,
			nil,
		))
		return
	}
	skuID = req.SKUID

	if stock != -1 && stock < req.Quantity {
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

	result, execErr := tx.Exec(
		"UPDATE skus SET stock = stock - $1 WHERE id = $2 AND (stock = -1 OR stock >= $1)",
		req.Quantity, skuID,
	)
	if execErr != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"STOCK_UPDATE_FAILED",
			"Failed to update stock",
			http.StatusInternalServerError,
			execErr,
		))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		middleware.RespondWithError(c, apperrors.ErrInsufficientStock)
		return
	}

	var order models.Order
	unitPrice := retailPrice
	if wholesalePrice > 0 && wholesalePrice < retailPrice {
		unitPrice = wholesalePrice
	}
	totalPrice := unitPrice * float64(req.Quantity)

	var groupID interface{}
	if req.GroupID > 0 {
		groupID = req.GroupID
	} else {
		groupID = nil
	}

	var pid interface{} = nil
	var productID sql.NullInt64
	pv := services.BaselinePricingVersionID(tx)
	var pvArg interface{}
	if pv.Valid {
		pvArg = pv.Int64
	} else {
		pvArg = nil
	}
	var pricingVID sql.NullInt64
	err = tx.QueryRow(
		`INSERT INTO orders (user_id, product_id, sku_id, spu_id, group_id, quantity, unit_price, total_price, status, pricing_version_id) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) 
		 RETURNING id, user_id, product_id, sku_id, spu_id, group_id, quantity, unit_price, total_price, status, pricing_version_id, created_at, updated_at`,
		userID, pid, skuID, spuID, groupID, req.Quantity, unitPrice, totalPrice, orderStatusPending, pvArg,
	).Scan(&order.ID, &order.UserID, &productID, &order.SKUID, &order.SPUID, &order.GroupID, &order.Quantity, &order.UnitPrice, &order.TotalPrice, &order.Status, &pricingVID, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"ORDER_CREATION_FAILED",
			"Failed to create order",
			http.StatusInternalServerError,
			err,
		))
		return
	}
	applyNullOrderProductID(&order, productID)
	if pricingVID.Valid {
		v := int(pricingVID.Int64)
		order.PricingVersionID = &v
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
		`SELECT id, user_id, product_id, sku_id, spu_id, group_id, quantity, unit_price, total_price, status, created_at, updated_at 
		 FROM orders WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
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
		var productID, skuID, spuID sql.NullInt64
		err := rows.Scan(&o.ID, &o.UserID, &productID, &skuID, &spuID, &o.GroupID, &o.Quantity, &o.UnitPrice, &o.TotalPrice, &o.Status, &o.CreatedAt, &o.UpdatedAt)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		applyNullOrderProductID(&o, productID)
		if skuID.Valid {
			o.SKUID = int(skuID.Int64)
		}
		if spuID.Valid {
			o.SPUID = int(spuID.Int64)
		}
		orders = append(orders, o)
	}

	var total int
	db.QueryRow("SELECT COUNT(*) FROM orders WHERE user_id = $1", userID).Scan(&total)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"total":    total,
			"page":     pageNum,
			"per_page": perPageNum,
			"data":     orders,
		},
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
				c.JSON(http.StatusOK, gin.H{
					"code":    0,
					"message": "success",
					"data":    order,
				})
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
	var productID, skuID, spuID sql.NullInt64
	err := db.QueryRow(
		`SELECT id, user_id, product_id, sku_id, spu_id, group_id, quantity, unit_price, total_price, status, created_at, updated_at 
		 FROM orders WHERE id = $1 AND user_id = $2`, id, userID,
	).Scan(&order.ID, &order.UserID, &productID, &skuID, &spuID, &order.GroupID, &order.Quantity, &order.UnitPrice, &order.TotalPrice, &order.Status, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrOrderNotFound)
		return
	}

	applyNullOrderProductID(&order, productID)
	if skuID.Valid {
		order.SKUID = int(skuID.Int64)
	}
	if spuID.Valid {
		order.SPUID = int(spuID.Int64)
	}

	if orderJSON, err := json.Marshal(order); err == nil {
		cache.Set(ctx, cacheKey, string(orderJSON), cache.OrderCacheTTL)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    order,
	})
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
		productID sql.NullInt64
		skuID     sql.NullInt64
		quantity  int
	}
	err := db.QueryRow(
		"SELECT status, product_id, sku_id, quantity FROM orders WHERE id = $1 AND user_id = $2",
		id, userID,
	).Scan(&orderInfo.status, &orderInfo.productID, &orderInfo.skuID, &orderInfo.quantity)

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
	var productID, skuIDNull, spuIDNull sql.NullInt64
	err = tx.QueryRow(
		`UPDATE orders SET status = $1 WHERE id = $2 AND user_id = $3 
		 RETURNING id, user_id, product_id, sku_id, spu_id, group_id, quantity, unit_price, total_price, status, created_at, updated_at`,
		"canceled", id, userID,
	).Scan(&order.ID, &order.UserID, &productID, &skuIDNull, &spuIDNull, &order.GroupID, &order.Quantity, &order.UnitPrice, &order.TotalPrice, &order.Status, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"ORDER_CANCEL_FAILED",
			"Failed to cancel order",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	applyNullOrderProductID(&order, productID)
	if skuIDNull.Valid {
		order.SKUID = int(skuIDNull.Int64)
	}
	if spuIDNull.Valid {
		order.SPUID = int(spuIDNull.Int64)
	}

	if orderInfo.skuID.Valid && orderInfo.skuID.Int64 > 0 {
		_, err = tx.Exec(
			"UPDATE skus SET stock = stock + $1 WHERE id = $2",
			orderInfo.quantity, orderInfo.skuID.Int64,
		)
	}
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

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    order,
	})
}

// CreateGroup creates a new group purchase
func CreateGroup(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	var req struct {
		ProductID   int       `json:"product_id"`
		SKUID       int       `json:"sku_id"`
		TargetCount int       `json:"target_count" binding:"required,gt=0"`
		Deadline    time.Time `json:"deadline" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	if req.SKUID <= 0 {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"MISSING_SKU",
			"sku_id is required",
			http.StatusBadRequest,
			nil,
		))
		return
	}

	if req.Deadline.Before(time.Now()) {
		middleware.RespondWithError(c, apperrors.ErrInvalidGroupData)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var skuID, spuID int
	var retailPrice float64
	var groupEnabled bool
	var minGroupSize, maxGroupSize int
	var groupDiscountRate sql.NullFloat64

	err := db.QueryRow(
		`SELECT s.id, s.spu_id, s.retail_price, s.group_enabled, s.min_group_size, s.max_group_size, s.group_discount_rate
		 FROM skus s JOIN spus sp ON s.spu_id = sp.id 
		 WHERE s.id = $1 AND s.status = 'active' AND sp.status = 'active'`,
		req.SKUID,
	).Scan(&skuID, &spuID, &retailPrice, &groupEnabled, &minGroupSize, &maxGroupSize, &groupDiscountRate)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrProductNotFound)
		return
	}

	if !groupEnabled {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"SKU_GROUP_NOT_ENABLED",
			"This SKU does not support group purchase",
			http.StatusBadRequest,
			nil,
		))
		return
	}

	if req.TargetCount < minGroupSize || req.TargetCount > maxGroupSize {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"INVALID_GROUP_SIZE",
			fmt.Sprintf("Group size must be between %d and %d", minGroupSize, maxGroupSize),
			http.StatusBadRequest,
			nil,
		))
		return
	}
	skuID = req.SKUID

	var group models.Group
	var nilPID interface{} = nil
	var productID sql.NullInt64
	err = db.QueryRow(
		`INSERT INTO groups (product_id, sku_id, spu_id, creator_id, target_count, current_count, status, deadline) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
		 RETURNING id, product_id, sku_id, spu_id, creator_id, target_count, current_count, status, deadline, created_at, updated_at`,
		nilPID, skuID, spuID, userID, req.TargetCount, 1, groupStatusActive, req.Deadline,
	).Scan(&group.ID, &productID, &group.SKUID, &group.SPUID, &group.CreatorID, &group.TargetCount, &group.CurrentCount, &group.Status, &group.Deadline, &group.CreatedAt, &group.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"GROUP_CREATION_FAILED",
			"Failed to create group",
			http.StatusInternalServerError,
			err,
		))
		return
	}
	applyNullProductID(&group, productID)

	var orderID int
	groupPrice := retailPrice * (1 - utils.NormalizeGroupDiscountRateNull(groupDiscountRate))
	pvCreate := services.BaselinePricingVersionID(db)
	var pvCreateArg interface{}
	if pvCreate.Valid {
		pvCreateArg = pvCreate.Int64
	} else {
		pvCreateArg = nil
	}

	err = db.QueryRow(
		`INSERT INTO orders (user_id, product_id, sku_id, spu_id, group_id, quantity, unit_price, total_price, status, pricing_version_id) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`,
		userID, nilPID, skuID, spuID, group.ID, 1, groupPrice, groupPrice, orderStatusPending, pvCreateArg,
	).Scan(&orderID)

	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"ORDER_CREATION_FAILED",
			"Failed to create order for group creator",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	_, err = db.Exec(
		"INSERT INTO group_members (group_id, user_id, order_id) VALUES ($1, $2, $3)",
		group.ID, userID, orderID,
	)

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrAlreadyInGroup)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"group":    group,
			"order_id": orderID,
		},
	})
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
		`SELECT id, product_id, sku_id, spu_id, creator_id, target_count, current_count, status, deadline, created_at, updated_at 
		 FROM groups WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
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
		var productID, skuID, spuID sql.NullInt64
		err := rows.Scan(&g.ID, &productID, &skuID, &spuID, &g.CreatorID, &g.TargetCount, &g.CurrentCount, &g.Status, &g.Deadline, &g.CreatedAt, &g.UpdatedAt)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		applyNullProductID(&g, productID)
		if skuID.Valid {
			g.SKUID = int(skuID.Int64)
		}
		if spuID.Valid {
			g.SPUID = int(spuID.Int64)
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

// GetGroupsBySKU retrieves active groups for a specific SKU (catalog :id is SKU id).
func GetGroupsBySKU(c *gin.Context) {
	productID := c.Param("id")
	productIDNum, err := strconv.Atoi(productID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	rows, err := db.Query(
		`SELECT id, product_id, sku_id, spu_id, creator_id, target_count, current_count, status, deadline, created_at, updated_at 
		 FROM groups WHERE sku_id = $1 AND status = $2 AND deadline > NOW() ORDER BY created_at DESC`,
		productIDNum, "active",
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	var groups []models.Group
	for rows.Next() {
		var g models.Group
		var productID, skuID, spuID sql.NullInt64
		err := rows.Scan(&g.ID, &productID, &skuID, &spuID, &g.CreatorID, &g.TargetCount, &g.CurrentCount, &g.Status, &g.Deadline, &g.CreatedAt, &g.UpdatedAt)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		applyNullProductID(&g, productID)
		if skuID.Valid {
			g.SKUID = int(skuID.Int64)
		}
		if spuID.Valid {
			g.SPUID = int(spuID.Int64)
		}
		groups = append(groups, g)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    groups,
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
	var productID, skuID, spuID sql.NullInt64
	err := db.QueryRow(
		`SELECT id, product_id, sku_id, spu_id, creator_id, target_count, current_count, status, deadline, created_at, updated_at 
		 FROM groups WHERE id = $1`,
		id,
	).Scan(&group.ID, &productID, &skuID, &spuID, &group.CreatorID, &group.TargetCount, &group.CurrentCount, &group.Status, &group.Deadline, &group.CreatedAt, &group.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrGroupNotFound)
		return
	}

	applyNullProductID(&group, productID)
	if skuID.Valid {
		group.SKUID = int(skuID.Int64)
	}
	if spuID.Valid {
		group.SPUID = int(spuID.Int64)
	}

	if groupJSON, err := json.Marshal(group); err == nil {
		cache.Set(ctx, cacheKey, string(groupJSON), 5*time.Minute)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    group,
	})
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

	var group models.Group
	var productID, skuID, spuID sql.NullInt64
	err := db.QueryRow(
		`SELECT id, product_id, sku_id, spu_id, target_count, current_count, status, deadline 
		 FROM groups WHERE id = $1`,
		groupID,
	).Scan(&group.ID, &productID, &skuID, &spuID, &group.TargetCount, &group.CurrentCount, &group.Status, &group.Deadline)

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrGroupNotFound)
		return
	}

	applyNullProductID(&group, productID)
	if skuID.Valid {
		group.SKUID = int(skuID.Int64)
	}
	if spuID.Valid {
		group.SPUID = int(spuID.Int64)
	}

	if group.Status != groupStatusActive {
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

	if group.SKUID <= 0 {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"GROUP_MISSING_SKU",
			"group has no sku_id",
			http.StatusBadRequest,
			nil,
		))
		return
	}
	var retailPrice float64
	var groupDiscountRate sql.NullFloat64
	err = db.QueryRow(
		`SELECT s.retail_price, s.group_discount_rate 
		 FROM skus s WHERE s.id = $1 AND s.status = 'active'`,
		group.SKUID,
	).Scan(&retailPrice, &groupDiscountRate)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrProductNotFound)
		return
	}
	unitPrice := retailPrice * (1 - utils.NormalizeGroupDiscountRateNull(groupDiscountRate))

	var nilPID interface{} = nil
	var orderID int
	pvJoin := services.BaselinePricingVersionID(db)
	var pvJoinArg interface{}
	if pvJoin.Valid {
		pvJoinArg = pvJoin.Int64
	} else {
		pvJoinArg = nil
	}
	err = db.QueryRow(
		`INSERT INTO orders (user_id, product_id, sku_id, spu_id, group_id, quantity, unit_price, total_price, status, pricing_version_id) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`,
		userID, nilPID, group.SKUID, group.SPUID, group.ID, 1, unitPrice, unitPrice, orderStatusPending, pvJoinArg,
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

	_, err = db.Exec(
		"INSERT INTO group_members (group_id, user_id, order_id) VALUES ($1, $2, $3)",
		group.ID, userID, orderID,
	)

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrAlreadyInGroup)
		return
	}

	newCount := group.CurrentCount + 1
	newStatus := group.Status
	if newCount >= group.TargetCount {
		newStatus = "completed"
	}

	var productIDUpdate, skuIDUpdate, spuIDUpdate sql.NullInt64
	err = db.QueryRow(
		`UPDATE groups SET current_count = $1, status = $2 WHERE id = $3 
		 RETURNING id, product_id, sku_id, spu_id, creator_id, target_count, current_count, status, deadline, created_at, updated_at`,
		newCount, newStatus, group.ID,
	).Scan(&group.ID, &productIDUpdate, &skuIDUpdate, &spuIDUpdate, &group.CreatorID, &group.TargetCount, &group.CurrentCount, &group.Status, &group.Deadline, &group.CreatedAt, &group.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"GROUP_UPDATE_FAILED",
			"Failed to update group",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	applyNullProductID(&group, productIDUpdate)
	if skuIDUpdate.Valid {
		group.SKUID = int(skuIDUpdate.Int64)
	}
	if spuIDUpdate.Valid {
		group.SPUID = int(spuIDUpdate.Int64)
	}

	cache.Delete(context.Background(), cache.GroupKey(group.ID))

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"group":    group,
			"order_id": orderID,
		},
	})
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
	var productID, skuID, spuID sql.NullInt64
	err := db.QueryRow(
		`SELECT id, product_id, sku_id, spu_id, creator_id, target_count, current_count, status, deadline, created_at, updated_at 
		 FROM groups WHERE id = $1`,
		id,
	).Scan(&group.ID, &productID, &skuID, &spuID, &group.CreatorID, &group.TargetCount, &group.CurrentCount, &group.Status, &group.Deadline, &group.CreatedAt, &group.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrGroupNotFound)
		return
	}

	applyNullProductID(&group, productID)
	if skuID.Valid {
		group.SKUID = int(skuID.Int64)
	}
	if spuID.Valid {
		group.SPUID = int(spuID.Int64)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    group,
	})
}
