package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/config"
)

type RechargeRequest struct {
	Amount float64 `json:"amount"`
	Method string  `json:"method"`
}

type RechargeOrder struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	Amount        float64   `json:"amount"`
	PaymentMethod string    `json:"payment_method"`
	PaymentID     int       `json:"payment_id,omitempty"`
	Status        string    `json:"status"`
	OutTradeNo    string    `json:"out_trade_no"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type RechargeCallbackRequest struct {
	PaymentID     int    `json:"payment_id"`
	OutTradeNo    string `json:"out_trade_no"`
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
}

func CreateRechargeOrder(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "UNAUTHORIZED",
			"message": "User not authenticated",
		})
		return
	}

	var req RechargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
		return
	}

	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "Amount must be greater than 0",
		})
		return
	}

	validMethods := map[string]bool{"alipay": true, "wechat": true, "balance": true}
	if !validMethods[req.Method] {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "Invalid payment method. Supported: alipay, wechat, balance",
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

	outTradeNo := fmt.Sprintf("RCH%d%d", time.Now().UnixNano(), userID.(int))

	var rechargeID int
	query := `
		INSERT INTO recharge_orders (user_id, amount, payment_method, status, out_trade_no, created_at, updated_at)
		VALUES ($1, $2, $3, 'pending', $4, NOW(), NOW())
		RETURNING id
	`
	err := db.QueryRow(query, userID, req.Amount, req.Method, outTradeNo).Scan(&rechargeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Failed to create recharge order",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "Recharge order created",
		"data": RechargeOrder{
			ID:            rechargeID,
			UserID:        userID.(int),
			Amount:        req.Amount,
			PaymentMethod: req.Method,
			Status:        "pending",
			OutTradeNo:    outTradeNo,
		},
	})
}

func GetRechargeOrders(c *gin.Context) {
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
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	var total int
	err := db.QueryRow("SELECT COUNT(*) FROM recharge_orders WHERE user_id = $1", userID).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Failed to count recharge orders",
		})
		return
	}

	query := `
		SELECT id, user_id, amount, payment_method, payment_id, status, out_trade_no, created_at, updated_at
		FROM recharge_orders
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := db.Query(query, userID, perPage, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Failed to fetch recharge orders",
		})
		return
	}
	defer rows.Close()

	var orders []RechargeOrder
	for rows.Next() {
		var order RechargeOrder
		var paymentID sql.NullInt64
		err := rows.Scan(
			&order.ID, &order.UserID, &order.Amount, &order.PaymentMethod,
			&paymentID, &order.Status, &order.OutTradeNo, &order.CreatedAt, &order.UpdatedAt,
		)
		if err != nil {
			continue
		}
		if paymentID.Valid {
			order.PaymentID = int(paymentID.Int64)
		}
		orders = append(orders, order)
	}

	if orders == nil {
		orders = []RechargeOrder{}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"total":    total,
			"page":     page,
			"per_page": perPage,
			"data":     orders,
		},
	})
}

func GetRechargeOrder(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "UNAUTHORIZED",
			"message": "User not authenticated",
		})
		return
	}

	orderIDStr := c.Param("id")
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "Invalid order ID",
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
		SELECT id, user_id, amount, payment_method, payment_id, status, out_trade_no, created_at, updated_at
		FROM recharge_orders
		WHERE id = $1 AND user_id = $2
	`
	var order RechargeOrder
	var paymentID sql.NullInt64
	err = db.QueryRow(query, orderID, userID).Scan(
		&order.ID, &order.UserID, &order.Amount, &order.PaymentMethod,
		&paymentID, &order.Status, &order.OutTradeNo, &order.CreatedAt, &order.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "NOT_FOUND",
			"message": "Recharge order not found",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Failed to fetch recharge order",
		})
		return
	}

	if paymentID.Valid {
		order.PaymentID = int(paymentID.Int64)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    order,
	})
}

func HandleRechargeCallback(c *gin.Context) {
	var req RechargeCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
		return
	}

	if req.PaymentID == 0 && req.OutTradeNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "payment_id or out_trade_no is required",
		})
		return
	}

	if req.Status == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "status is required",
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

	var order RechargeOrder
	var query string
	var args []interface{}

	if req.PaymentID > 0 {
		query = `
			SELECT id, user_id, amount, payment_method, payment_id, status, out_trade_no, created_at, updated_at
			FROM recharge_orders
			WHERE payment_id = $1
		`
		args = []interface{}{req.PaymentID}
	} else {
		query = `
			SELECT id, user_id, amount, payment_method, payment_id, status, out_trade_no, created_at, updated_at
			FROM recharge_orders
			WHERE out_trade_no = $1
		`
		args = []interface{}{req.OutTradeNo}
	}

	var paymentID sql.NullInt64
	err := db.QueryRow(query, args...).Scan(
		&order.ID, &order.UserID, &order.Amount, &order.PaymentMethod,
		&paymentID, &order.Status, &order.OutTradeNo, &order.CreatedAt, &order.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "NOT_FOUND",
			"message": "Recharge order not found",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DATABASE_ERROR",
			"message": "Failed to fetch recharge order",
		})
		return
	}

	if order.Status != "pending" {
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "Order already processed",
			"data":    order,
		})
		return
	}

	if req.Status == "success" {
		tx, err := db.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to start transaction",
			})
			return
		}
		defer tx.Rollback()

		_, err = tx.Exec(
			"UPDATE recharge_orders SET status = 'success', updated_at = NOW() WHERE id = $1",
			order.ID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to update recharge order",
			})
			return
		}

		_, err = tx.Exec(
			"UPDATE tokens SET balance = balance + $1, updated_at = NOW() WHERE user_id = $2",
			order.Amount, order.UserID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to update token balance",
			})
			return
		}

		_, err = tx.Exec(
			"INSERT INTO token_transactions (user_id, amount, type, description, created_at) VALUES ($1, $2, 'recharge', $3, NOW())",
			order.UserID, order.Amount, fmt.Sprintf("Recharge order #%d", order.ID),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to create transaction record",
			})
			return
		}

		if err = tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to commit transaction",
			})
			return
		}

		order.Status = "success"
	} else if req.Status == "failed" {
		_, err = db.Exec(
			"UPDATE recharge_orders SET status = 'failed', updated_at = NOW() WHERE id = $1",
			order.ID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to update recharge order",
			})
			return
		}
		order.Status = "failed"
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    order,
	})
}

func GetRechargePackages(c *gin.Context) {
	packages := []gin.H{
		{"id": 1, "amount": 100, "bonus": 0, "description": "100 Tokens"},
		{"id": 2, "amount": 500, "bonus": 50, "description": "500 + 50 Bonus"},
		{"id": 3, "amount": 1000, "bonus": 150, "description": "1000 + 150 Bonus"},
		{"id": 4, "amount": 5000, "bonus": 1000, "description": "5000 + 1000 Bonus"},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    packages,
	})
}
