package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
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
	"github.com/pintuotuo/backend/models"
	"github.com/pintuotuo/backend/services"
	"github.com/pintuotuo/backend/utils"
)

const orderStatusPending = "pending"
const groupStatusActive = "active"

type createOrderLineItem struct {
	SKUID    int `json:"sku_id" binding:"required,gt=0"`
	Quantity int `json:"quantity" binding:"required,gt=0"`
}

func validateEntitlementPackageOrderLines(tx *sql.Tx, packageID int, reqItems []createOrderLineItem) *apperrors.AppError {
	var st string
	err := tx.QueryRow(`SELECT status FROM entitlement_packages WHERE id = $1`, packageID).Scan(&st)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.NewAppError("ENTITLEMENT_PACKAGE_NOT_FOUND", "套餐不存在", http.StatusBadRequest, nil)
		}
		return apperrors.ErrDatabaseError
	}
	if st != productStatusActive {
		return apperrors.NewAppError("ENTITLEMENT_PACKAGE_INACTIVE", "套餐已下架", http.StatusBadRequest, nil)
	}
	rows, err := tx.Query(`SELECT sku_id, default_quantity FROM entitlement_package_items WHERE package_id = $1`, packageID)
	if err != nil {
		return apperrors.ErrDatabaseError
	}
	defer rows.Close()
	expected := make(map[int]int)
	for rows.Next() {
		var sk, dq int
		if scanErr := rows.Scan(&sk, &dq); scanErr != nil {
			return apperrors.ErrDatabaseError
		}
		expected[sk] = dq
	}
	if err := rows.Err(); err != nil {
		return apperrors.ErrDatabaseError
	}
	if len(expected) == 0 {
		return apperrors.NewAppError("ENTITLEMENT_PACKAGE_EMPTY", "套餐未配置商品", http.StatusBadRequest, nil)
	}
	actual := make(map[int]int)
	for _, it := range reqItems {
		actual[it.SKUID] += it.Quantity
	}
	if len(actual) != len(expected) {
		return apperrors.NewAppError("ENTITLEMENT_PACKAGE_LINE_MISMATCH", "下单明细与套餐不一致，请刷新后重试", http.StatusBadRequest, nil)
	}
	for sk, q := range expected {
		if actual[sk] != q {
			return apperrors.NewAppError("ENTITLEMENT_PACKAGE_LINE_MISMATCH", "下单明细与套餐不一致，请刷新后重试", http.StatusBadRequest, nil)
		}
	}
	return nil
}

// CreateOrder creates a new order
func CreateOrder(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	var req struct {
		EntitlementPackageID *int                  `json:"entitlement_package_id"`
		Items                []createOrderLineItem `json:"items" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}
	if req.EntitlementPackageID != nil && *req.EntitlementPackageID <= 0 {
		middleware.RespondWithError(c, apperrors.NewAppError("INVALID_ENTITLEMENT_PACKAGE", "套餐参数无效", http.StatusBadRequest, nil))
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
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

	if req.EntitlementPackageID != nil {
		if ae := validateEntitlementPackageOrderLines(tx, *req.EntitlementPackageID, req.Items); ae != nil {
			middleware.RespondWithError(c, ae)
			return
		}
	}

	type pendingItem struct {
		item        models.OrderItem
		tokenAmount sql.NullInt64
		cpAmount    sql.NullFloat64
	}
	items := make([]pendingItem, 0, len(req.Items))
	policyLines := make([]services.OrderLinePolicyInput, 0, len(req.Items))
	totalQty := 0
	totalPrice := 0.0

	pv := services.BaselinePricingVersionID(tx)
	var pvArg interface{}
	if pv.Valid {
		pvArg = pv.Int64
	}

	for _, in := range req.Items {
		var skuID, spuID int
		var retailPrice, wholesalePrice float64
		var stock int
		var skuType string
		var tokenAmount sql.NullInt64
		var computePoints sql.NullFloat64
		var subscriptionPeriod sql.NullString
		var validDays sql.NullInt64
		var trialDurationDays sql.NullInt64

		var modelProvider, modelName, providerModelID sql.NullString
		err = tx.QueryRow(
			`SELECT s.id, s.spu_id, s.retail_price, s.wholesale_price, s.stock, s.sku_type,
				s.token_amount, s.compute_points, s.subscription_period, s.valid_days, s.trial_duration_days,
				sp.model_provider, sp.model_name, sp.provider_model_id
			 FROM skus s JOIN spus sp ON s.spu_id = sp.id
			 WHERE s.id = $1 AND s.status = 'active' AND sp.status = 'active'`,
			in.SKUID,
		).Scan(
			&skuID, &spuID, &retailPrice, &wholesalePrice, &stock, &skuType,
			&tokenAmount, &computePoints, &subscriptionPeriod, &validDays, &trialDurationDays,
			&modelProvider, &modelName, &providerModelID,
		)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				middleware.RespondWithError(c, apperrors.NewAppError(
					"ORDER_LINE_SKU_UNAVAILABLE",
					"包含不可售或已下架的商品，请刷新页面后重试",
					http.StatusBadRequest,
					nil,
				))
				return
			}
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
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

		result, execErr := tx.Exec(
			services.SQLReserveSKUStockForOrder,
			in.Quantity, skuID,
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

		unitPrice := retailPrice
		if wholesalePrice > 0 && wholesalePrice < retailPrice {
			unitPrice = wholesalePrice
		}
		lineTotal := unitPrice * float64(in.Quantity)

		item := models.OrderItem{
			SKUID:      skuID,
			SPUID:      spuID,
			Quantity:   in.Quantity,
			UnitPrice:  unitPrice,
			TotalPrice: lineTotal,
			SKUType:    skuType,
		}
		if tokenAmount.Valid {
			t := tokenAmount.Int64
			item.TokenAmount = &t
		}
		if computePoints.Valid {
			cpVal := computePoints.Float64
			item.ComputePoints = &cpVal
		}
		if pv.Valid {
			pvInt := int(pv.Int64)
			item.PricingVersionID = &pvInt
		}
		items = append(items, pendingItem{item: item, tokenAmount: tokenAmount, cpAmount: computePoints})
		policyLines = append(policyLines, services.OrderLinePolicyInput{
			SKUType:         skuType,
			ModelProvider:   modelProvider.String,
			ModelName:       modelName.String,
			ProviderModelID: providerModelID.String,
		})
		totalQty += in.Quantity
		totalPrice += lineTotal
	}
	if err = services.ValidateFuelPackBundle(policyLines); err != nil {
		metrics.RecordFuelPackRestriction("create_order", "FUEL_PACK_PURCHASE_RESTRICTED")
		logger.LogWarn(c.Request.Context(), "fuel_pack_policy", "Blocked token_pack-only create order", map[string]interface{}{
			"user_id": userID,
			"source":  "create_order",
			"code":    "FUEL_PACK_PURCHASE_RESTRICTED",
		})
		middleware.RespondWithError(c, apperrors.NewAppError(
			"FUEL_PACK_PURCHASE_RESTRICTED",
			err.Error(),
			http.StatusBadRequest,
			nil,
		))
		return
	}

	var order models.Order
	var productID sql.NullInt64
	var skuID sql.NullInt64
	var spuID sql.NullInt64
	var pricingVID sql.NullInt64
	var epID sql.NullInt64
	var epArg interface{}
	if req.EntitlementPackageID != nil {
		epArg = *req.EntitlementPackageID
	}
	err = tx.QueryRow(
		`INSERT INTO orders (user_id, product_id, sku_id, spu_id, group_id, quantity, unit_price, total_price, status, pricing_version_id, entitlement_package_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 RETURNING id, user_id, product_id, sku_id, spu_id, group_id, quantity, unit_price, total_price, status, pricing_version_id, entitlement_package_id, created_at, updated_at`,
		userID, nil, nil, nil, nil, totalQty, 0, totalPrice, orderStatusPending, pvArg, epArg,
	).Scan(&order.ID, &order.UserID, &productID, &skuID, &spuID, &order.GroupID, &order.Quantity, &order.UnitPrice, &order.TotalPrice, &order.Status, &pricingVID, &epID, &order.CreatedAt, &order.UpdatedAt)
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
	if epID.Valid {
		v := int(epID.Int64)
		order.EntitlementPackageID = &v
	}

	order.Items = make([]models.OrderItem, 0, len(items))
	for _, pi := range items {
		item := pi.item
		var itemID int
		var itemPV sql.NullInt64
		var createdAt, updatedAt time.Time
		var fulfilledAt sql.NullTime
		err = tx.QueryRow(
			`INSERT INTO order_items (order_id, sku_id, spu_id, quantity, unit_price, total_price, pricing_version_id, sku_type, token_amount, compute_points)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
			 RETURNING id, pricing_version_id, fulfilled_at, created_at, updated_at`,
			order.ID, item.SKUID, item.SPUID, item.Quantity, item.UnitPrice, item.TotalPrice, pvArg, item.SKUType, pi.tokenAmount, pi.cpAmount,
		).Scan(&itemID, &itemPV, &fulfilledAt, &createdAt, &updatedAt)
		if err != nil {
			middleware.RespondWithError(c, apperrors.NewAppError(
				"ORDER_ITEM_CREATION_FAILED",
				"Failed to create order item",
				http.StatusInternalServerError,
				err,
			))
			return
		}
		item.ID = itemID
		item.OrderID = order.ID
		item.CreatedAt = createdAt
		item.UpdatedAt = updatedAt
		if itemPV.Valid {
			v := int(itemPV.Int64)
			item.PricingVersionID = &v
		}
		if fulfilledAt.Valid {
			t := fulfilledAt.Time
			item.FulfilledAt = &t
		}
		order.Items = append(order.Items, item)
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

func loadOrderItems(db *sql.DB, orderID int) ([]models.OrderItem, error) {
	rows, err := db.Query(
		`SELECT oi.id, oi.order_id, oi.sku_id, oi.spu_id, oi.quantity, oi.unit_price, oi.total_price,
		        oi.sku_type, COALESCE(sp.name, ''), COALESCE(s.sku_code, ''),
		        oi.token_amount, oi.compute_points, oi.fulfilled_at, oi.pricing_version_id, oi.created_at, oi.updated_at
		   FROM order_items oi
		   LEFT JOIN spus sp ON oi.spu_id = sp.id
		   LEFT JOIN skus s ON oi.sku_id = s.id
		  WHERE oi.order_id = $1
		  ORDER BY oi.id ASC`,
		orderID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.OrderItem, 0)
	for rows.Next() {
		var item models.OrderItem
		var tokenAmount sql.NullInt64
		var computePoints sql.NullFloat64
		var fulfilledAt sql.NullTime
		var pricingVID sql.NullInt64
		if scanErr := rows.Scan(
			&item.ID, &item.OrderID, &item.SKUID, &item.SPUID, &item.Quantity, &item.UnitPrice, &item.TotalPrice,
			&item.SKUType, &item.SPUName, &item.SKUCode,
			&tokenAmount, &computePoints, &fulfilledAt, &pricingVID, &item.CreatedAt, &item.UpdatedAt,
		); scanErr != nil {
			return nil, scanErr
		}
		if tokenAmount.Valid {
			v := tokenAmount.Int64
			item.TokenAmount = &v
		}
		if computePoints.Valid {
			v := computePoints.Float64
			item.ComputePoints = &v
		}
		if fulfilledAt.Valid {
			t := fulfilledAt.Time
			item.FulfilledAt = &t
		}
		if pricingVID.Valid {
			v := int(pricingVID.Int64)
			item.PricingVersionID = &v
		}
		items = append(items, item)
	}

	return items, nil
}

func enrichOrderGroupStatus(db *sql.DB, o *models.Order) {
	if o == nil || o.GroupID == nil {
		return
	}
	gid, ok := groupIDToInt(o.GroupID)
	if !ok || gid <= 0 {
		return
	}
	var st sql.NullString
	if err := db.QueryRow(`SELECT status FROM groups WHERE id = $1`, gid).Scan(&st); err == nil && st.Valid {
		o.GroupStatus = st.String
	}
}

func groupIDToInt(v interface{}) (int, bool) {
	if v == nil {
		return 0, false
	}
	switch x := v.(type) {
	case int:
		return x, true
	case int64:
		return int(x), true
	case float64:
		return int(x), true
	default:
		return 0, false
	}
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
		`SELECT o.id, o.user_id, o.product_id, o.group_id, o.quantity, o.unit_price, o.total_price, o.status, o.pricing_version_id, o.entitlement_package_id, o.created_at, o.updated_at,
		        g.status AS group_status
		   FROM orders o
		   LEFT JOIN groups g ON o.group_id = g.id
		  WHERE o.user_id = $1
		  ORDER BY o.created_at DESC
		  LIMIT $2 OFFSET $3`,
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
		var productID sql.NullInt64
		var pricingVID sql.NullInt64
		var epID sql.NullInt64
		var groupSt sql.NullString
		err := rows.Scan(&o.ID, &o.UserID, &productID, &o.GroupID, &o.Quantity, &o.UnitPrice, &o.TotalPrice, &o.Status, &pricingVID, &epID, &o.CreatedAt, &o.UpdatedAt, &groupSt)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		applyNullOrderProductID(&o, productID)
		if pricingVID.Valid {
			v := int(pricingVID.Int64)
			o.PricingVersionID = &v
		}
		if epID.Valid {
			v := int(epID.Int64)
			o.EntitlementPackageID = &v
		}
		if groupSt.Valid {
			o.GroupStatus = groupSt.String
		}
		o.Items, err = loadOrderItems(db, o.ID)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
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
				db := config.GetDB()
				if db != nil {
					enrichOrderGroupStatus(db, &order)
				}
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
	var productID sql.NullInt64
	var pricingVID sql.NullInt64
	var epID sql.NullInt64
	var groupSt sql.NullString
	err := db.QueryRow(
		`SELECT o.id, o.user_id, o.product_id, o.group_id, o.quantity, o.unit_price, o.total_price, o.status, o.pricing_version_id, o.entitlement_package_id, o.created_at, o.updated_at,
		        g.status AS group_status
		   FROM orders o
		   LEFT JOIN groups g ON o.group_id = g.id
		  WHERE o.id = $1 AND o.user_id = $2`,
		id, userID,
	).Scan(&order.ID, &order.UserID, &productID, &order.GroupID, &order.Quantity, &order.UnitPrice, &order.TotalPrice, &order.Status, &pricingVID, &epID, &order.CreatedAt, &order.UpdatedAt, &groupSt)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrOrderNotFound)
		return
	}

	applyNullOrderProductID(&order, productID)
	if pricingVID.Valid {
		v := int(pricingVID.Int64)
		order.PricingVersionID = &v
	}
	if epID.Valid {
		v := int(epID.Int64)
		order.EntitlementPackageID = &v
	}
	if groupSt.Valid {
		order.GroupStatus = groupSt.String
	}
	order.Items, err = loadOrderItems(db, order.ID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
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
		status string
	}
	err := db.QueryRow(
		"SELECT status FROM orders WHERE id = $1 AND user_id = $2",
		id, userID,
	).Scan(&orderInfo.status)

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

	_, err = tx.Exec(
		services.SQLRestoreSKUStockFromOrderItems,
		order.ID,
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
	var skuType string
	var tokenAmount sql.NullInt64
	var computePoints sql.NullFloat64
	var groupEnabled bool
	var minGroupSize, maxGroupSize int
	var groupDiscountRate sql.NullFloat64
	var modelProvider, modelName, providerModelID sql.NullString

	err := db.QueryRow(
		`SELECT s.id, s.spu_id, s.retail_price, s.sku_type, s.token_amount, s.compute_points,
		        s.group_enabled, s.min_group_size, s.max_group_size, s.group_discount_rate,
		        sp.model_provider, sp.model_name, sp.provider_model_id
		 FROM skus s JOIN spus sp ON s.spu_id = sp.id 
		 WHERE s.id = $1 AND s.status = 'active' AND sp.status = 'active'`,
		req.SKUID,
	).Scan(&skuID, &spuID, &retailPrice, &skuType, &tokenAmount, &computePoints, &groupEnabled, &minGroupSize, &maxGroupSize, &groupDiscountRate, &modelProvider, &modelName, &providerModelID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrProductNotFound)
		return
	}
	if err = services.ValidateFuelPackBundle([]services.OrderLinePolicyInput{{
		SKUType:         skuType,
		ModelProvider:   modelProvider.String,
		ModelName:       modelName.String,
		ProviderModelID: providerModelID.String,
	}}); err != nil {
		metrics.RecordFuelPackRestriction("create_group", "FUEL_PACK_PURCHASE_RESTRICTED")
		logger.LogWarn(c.Request.Context(), "fuel_pack_policy", "Blocked token_pack-only group create", map[string]interface{}{
			"user_id": userID,
			"source":  "create_group",
			"code":    "FUEL_PACK_PURCHASE_RESTRICTED",
			"sku_id":  req.SKUID,
		})
		middleware.RespondWithError(c, apperrors.NewAppError(
			"FUEL_PACK_PURCHASE_RESTRICTED",
			"加油包不支持单独发起拼团，请选择带模型的商品",
			http.StatusBadRequest,
			nil,
		))
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
		userID, nilPID, nil, nil, group.ID, 1, 0, groupPrice, orderStatusPending, pvCreateArg,
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
		`INSERT INTO order_items (order_id, sku_id, spu_id, quantity, unit_price, total_price, pricing_version_id, sku_type, token_amount, compute_points)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		orderID, skuID, spuID, 1, groupPrice, groupPrice, pvCreateArg, skuType, tokenAmount, computePoints,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"ORDER_ITEM_CREATION_FAILED",
			"Failed to create order item for group creator",
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

// ListGroups lists groups. Query scope:
//   - all (default): all groups with given status (参团广场 / 全站热团)
//   - mine_created: current user as creator（我发起的）
//   - mine_joined: user in group_members but not creator（我跟的团，不含自己发起的）
//   - mine_involved: creator OR member in group_members（发起+参团一条时间线）
func ListGroups(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("per_page", "20")
	status := c.DefaultQuery("status", "active")
	scope := strings.ToLower(strings.TrimSpace(c.DefaultQuery("scope", "all")))
	if scope != "all" && scope != "mine_created" && scope != "mine_joined" && scope != "mine_involved" {
		scope = "all"
	}

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

	var uid int
	if scope != "all" {
		userIDRaw, exists := c.Get("user_id")
		if !exists {
			middleware.RespondWithError(c, apperrors.ErrInvalidToken)
			return
		}
		var ok bool
		uid, ok = userIDRaw.(int)
		if !ok {
			if f, ok2 := userIDRaw.(float64); ok2 {
				uid = int(f)
			} else {
				middleware.RespondWithError(c, apperrors.ErrInvalidToken)
				return
			}
		}
	}

	appendGroup := func(rows *sql.Rows) ([]models.Group, error) {
		var groups []models.Group
		for rows.Next() {
			var g models.Group
			var productID, skuID, spuID sql.NullInt64
			if err := rows.Scan(&g.ID, &productID, &skuID, &spuID, &g.CreatorID, &g.TargetCount, &g.CurrentCount, &g.Status, &g.Deadline, &g.CreatedAt, &g.UpdatedAt); err != nil {
				return nil, err
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
		return groups, rows.Err()
	}

	var (
		rows  *sql.Rows
		err   error
		total int
	)

	switch scope {
	case "mine_created":
		rows, err = db.Query(
			`SELECT id, product_id, sku_id, spu_id, creator_id, target_count, current_count, status, deadline, created_at, updated_at
			 FROM groups WHERE status = $1 AND creator_id = $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`,
			status, uid, perPageNum, offset,
		)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		defer rows.Close()
		if err = db.QueryRow(
			`SELECT COUNT(*) FROM groups WHERE status = $1 AND creator_id = $2`,
			status, uid,
		).Scan(&total); err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
	case "mine_joined":
		rows, err = db.Query(
			`SELECT g.id, g.product_id, g.sku_id, g.spu_id, g.creator_id, g.target_count, g.current_count, g.status, g.deadline, g.created_at, g.updated_at
			 FROM groups g
			 WHERE g.status = $1
			   AND g.id IN (SELECT gm.group_id FROM group_members gm WHERE gm.user_id = $2)
			   AND g.creator_id <> $2
			 ORDER BY g.created_at DESC
			 LIMIT $3 OFFSET $4`,
			status, uid, perPageNum, offset,
		)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		defer rows.Close()
		if err = db.QueryRow(
			`SELECT COUNT(*) FROM groups g
			 WHERE g.status = $1
			   AND g.id IN (SELECT group_id FROM group_members WHERE user_id = $2)
			   AND g.creator_id <> $2`,
			status, uid,
		).Scan(&total); err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
	case "mine_involved":
		rows, err = db.Query(
			`SELECT g.id, g.product_id, g.sku_id, g.spu_id, g.creator_id, g.target_count, g.current_count, g.status, g.deadline, g.created_at, g.updated_at
			 FROM groups g
			 WHERE g.status = $1
			   AND (g.creator_id = $2 OR g.id IN (SELECT gm.group_id FROM group_members gm WHERE gm.user_id = $2))
			 ORDER BY g.created_at DESC
			 LIMIT $3 OFFSET $4`,
			status, uid, perPageNum, offset,
		)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		defer rows.Close()
		if err = db.QueryRow(
			`SELECT COUNT(*) FROM groups g
			 WHERE g.status = $1
			   AND (g.creator_id = $2 OR g.id IN (SELECT group_id FROM group_members WHERE user_id = $2))`,
			status, uid,
		).Scan(&total); err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
	default:
		rows, err = db.Query(
			`SELECT id, product_id, sku_id, spu_id, creator_id, target_count, current_count, status, deadline, created_at, updated_at
			 FROM groups WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
			status, perPageNum, offset,
		)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		defer rows.Close()
		if err = db.QueryRow("SELECT COUNT(*) FROM groups WHERE status = $1", status).Scan(&total); err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
	}

	groups, err := appendGroup(rows)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

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
	var skuType string
	var tokenAmount sql.NullInt64
	var computePoints sql.NullFloat64
	var groupDiscountRate sql.NullFloat64
	var modelProvider, modelName, providerModelID sql.NullString
	err = db.QueryRow(
		`SELECT s.retail_price, s.sku_type, s.token_amount, s.compute_points, s.group_discount_rate,
		        sp.model_provider, sp.model_name, sp.provider_model_id
		 FROM skus s
		 JOIN spus sp ON s.spu_id = sp.id
		 WHERE s.id = $1 AND s.status = 'active' AND sp.status = 'active'`,
		group.SKUID,
	).Scan(&retailPrice, &skuType, &tokenAmount, &computePoints, &groupDiscountRate, &modelProvider, &modelName, &providerModelID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrProductNotFound)
		return
	}
	if err = services.ValidateFuelPackBundle([]services.OrderLinePolicyInput{{
		SKUType:         skuType,
		ModelProvider:   modelProvider.String,
		ModelName:       modelName.String,
		ProviderModelID: providerModelID.String,
	}}); err != nil {
		metrics.RecordFuelPackRestriction("join_group", "FUEL_PACK_PURCHASE_RESTRICTED")
		logger.LogWarn(c.Request.Context(), "fuel_pack_policy", "Blocked token_pack-only join group", map[string]interface{}{
			"user_id": userID,
			"source":  "join_group",
			"code":    "FUEL_PACK_PURCHASE_RESTRICTED",
			"sku_id":  group.SKUID,
		})
		middleware.RespondWithError(c, apperrors.NewAppError(
			"FUEL_PACK_PURCHASE_RESTRICTED",
			"加油包不支持单独参团，请选择带模型的商品",
			http.StatusBadRequest,
			nil,
		))
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
		userID, nilPID, nil, nil, group.ID, 1, 0, unitPrice, orderStatusPending, pvJoinArg,
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
		`INSERT INTO order_items (order_id, sku_id, spu_id, quantity, unit_price, total_price, pricing_version_id, sku_type, token_amount, compute_points)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		orderID, group.SKUID, group.SPUID, 1, unitPrice, unitPrice, pvJoinArg, skuType, tokenAmount, computePoints,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"ORDER_ITEM_CREATION_FAILED",
			"Failed to create order item",
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
