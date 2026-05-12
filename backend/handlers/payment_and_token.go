package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/billing"
	"github.com/pintuotuo/backend/cache"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/models"
)

const PaymentStatusSuccess = "success"

const paymentStatusPending = "pending"

// InitiatePayment initiates a payment for an order
func InitiatePayment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	var req struct {
		OrderID int    `json:"order_id" binding:"required"`
		Method  string `json:"method" binding:"required,oneof=alipay wechat"`
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

	// Verify order belongs to user
	var order models.Order
	err := db.QueryRow(
		"SELECT id, user_id, total_price, status FROM orders WHERE id = $1 AND user_id = $2",
		req.OrderID, userID,
	).Scan(&order.ID, &order.UserID, &order.TotalPrice, &order.Status)

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrOrderNotFound)
		return
	}

	if order.Status != paymentStatusPending {
		middleware.RespondWithError(c, apperrors.ErrOrderAlreadyPaid)
		return
	}

	var payment models.Payment
	err = db.QueryRow(
		"INSERT INTO payments (order_id, user_id, amount, pay_method, status) VALUES ($1, $2, $3, $4, $5) RETURNING id, order_id, amount, pay_method, status, created_at, updated_at",
		req.OrderID, userID, order.TotalPrice, req.Method, paymentStatusPending,
	).Scan(&payment.ID, &payment.OrderID, &payment.Amount, &payment.PayMethod, &payment.Status, &payment.CreatedAt, &payment.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"PAYMENT_CREATION_FAILED",
			"Failed to create payment",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "success",
		"data":    payment,
	})
}

// GetPaymentByID retrieves a payment by ID
func GetPaymentByID(c *gin.Context) {
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

	var payment models.Payment
	err := db.QueryRow(
		"SELECT id, order_id, amount, pay_method, status, created_at, updated_at FROM payments WHERE id = $1 AND user_id = $2",
		id, userID,
	).Scan(&payment.ID, &payment.OrderID, &payment.Amount, &payment.PayMethod, &payment.Status, &payment.CreatedAt, &payment.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrPaymentNotFound)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    payment,
	})
}

// RefundPayment processes a refund for a payment
func RefundPayment(c *gin.Context) {
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

	// Get payment details
	var payment models.Payment
	err := db.QueryRow(
		"SELECT id, order_id, amount, pay_method, status FROM payments WHERE id = $1 AND user_id = $2",
		id, userID,
	).Scan(&payment.ID, &payment.OrderID, &payment.Amount, &payment.PayMethod, &payment.Status)

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrPaymentNotFound)
		return
	}

	if payment.Status != PaymentStatusSuccess {
		middleware.RespondWithError(c, apperrors.ErrPaymentAlreadyProcessed)
		return
	}

	// Update payment status
	err = db.QueryRow(
		"UPDATE payments SET status = $1 WHERE id = $2 RETURNING id, order_id, amount, pay_method, status, created_at, updated_at",
		"refunded", id,
	).Scan(&payment.ID, &payment.OrderID, &payment.Amount, &payment.PayMethod, &payment.Status, &payment.CreatedAt, &payment.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"REFUND_FAILED",
			"Failed to process refund",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    payment,
	})
}

// HandleAlipayCallback handles Alipay payment callback
func HandleAlipayCallback(c *gin.Context) {
	var req struct {
		PaymentID     int     `json:"payment_id" binding:"required"`
		Status        string  `json:"status" binding:"required"`
		Amount        float64 `json:"amount" binding:"required"`
		TransactionID string  `json:"transaction_id"`
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

	// Verify payment exists
	var paymentStatus string
	err := db.QueryRow("SELECT status FROM payments WHERE id = $1", req.PaymentID).Scan(&paymentStatus)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrPaymentNotFound)
		return
	}

	// Update payment status
	var payMethod string
	err = db.QueryRow(
		"UPDATE payments SET status = $1, transaction_id = $2 WHERE id = $3 RETURNING pay_method",
		req.Status, req.TransactionID, req.PaymentID,
	).Scan(&payMethod)

	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"PAYMENT_UPDATE_FAILED",
			"Failed to update payment",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	// If payment successful, update order status
	if req.Status == PaymentStatusSuccess {
		_, err := db.Exec(
			"UPDATE orders SET status = $1 WHERE id = (SELECT order_id FROM payments WHERE id = $2)",
			"paid", req.PaymentID,
		)
		if err != nil {
			middleware.RespondWithError(c, apperrors.NewAppError(
				"ORDER_UPDATE_FAILED",
				"Failed to update order",
				http.StatusInternalServerError,
				err,
			))
			return
		}
		var orderID int
		if err := db.QueryRow("SELECT order_id FROM payments WHERE id = $1", req.PaymentID).Scan(&orderID); err == nil {
			ApplyReferralRewardForPaidOrder(orderID)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Callback processed"})
}

// HandleWechatCallback handles WeChat payment callback
func HandleWechatCallback(c *gin.Context) {
	var req struct {
		PaymentID     int     `json:"payment_id" binding:"required"`
		Status        string  `json:"status" binding:"required"`
		Amount        float64 `json:"amount" binding:"required"`
		TransactionID string  `json:"transaction_id"`
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

	// Verify payment exists
	var paymentStatus string
	err := db.QueryRow("SELECT status FROM payments WHERE id = $1", req.PaymentID).Scan(&paymentStatus)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrPaymentNotFound)
		return
	}

	// Update payment status
	var payMethod string
	err = db.QueryRow(
		"UPDATE payments SET status = $1, transaction_id = $2 WHERE id = $3 RETURNING pay_method",
		req.Status, req.TransactionID, req.PaymentID,
	).Scan(&payMethod)

	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"PAYMENT_UPDATE_FAILED",
			"Failed to update payment",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	// If payment successful, update order status
	if req.Status == PaymentStatusSuccess {
		_, err := db.Exec(
			"UPDATE orders SET status = $1 WHERE id = (SELECT order_id FROM payments WHERE id = $2)",
			"paid", req.PaymentID,
		)
		if err != nil {
			middleware.RespondWithError(c, apperrors.NewAppError(
				"ORDER_UPDATE_FAILED",
				"Failed to update order",
				http.StatusInternalServerError,
				err,
			))
			return
		}
		var orderID int
		if err := db.QueryRow("SELECT order_id FROM payments WHERE id = $1", req.PaymentID).Scan(&orderID); err == nil {
			ApplyReferralRewardForPaidOrder(orderID)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Callback processed"})
}

// GetTokenBalance retrieves user token balance
func GetTokenBalance(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	ctx := context.Background()
	userIDInt, ok := userID.(int)
	if !ok {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	// Try cache first
	cacheKey := cache.TokenBalanceKey(userIDInt)
	if cachedToken, err := cache.Get(ctx, cacheKey); err == nil {
		var token models.Token
		if err := json.Unmarshal([]byte(cachedToken), &token); err == nil {
			c.JSON(http.StatusOK, token)
			return
		}
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var token models.Token
	err := db.QueryRow(
		"SELECT id, user_id, balance, created_at, updated_at FROM tokens WHERE user_id = $1",
		userIDInt,
	).Scan(&token.ID, &token.UserID, &token.Balance, &token.CreatedAt, &token.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrTokenNotFound)
		return
	}

	// Cache the result
	if tokenJSON, err := json.Marshal(token); err == nil {
		cache.Set(ctx, cacheKey, string(tokenJSON), cache.TokenBalanceTTL)
	}

	c.JSON(http.StatusOK, token)
}

// GetTokenConsumption retrieves user token consumption history
func GetTokenConsumption(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	rows, err := db.Query(
		"SELECT id, type, amount, reason, created_at FROM token_transactions WHERE user_id = $1 ORDER BY created_at DESC LIMIT 100",
		userID,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	var transactions []map[string]interface{}
	for rows.Next() {
		var id int
		var txType, reason string
		var amount float64
		var createdAt interface{}

		err := rows.Scan(&id, &txType, &amount, &reason, &createdAt)
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}

		transactions = append(transactions, map[string]interface{}{
			"id":         id,
			"type":       txType,
			"amount":     amount,
			"reason":     reason,
			"created_at": createdAt,
		})
	}

	c.JSON(http.StatusOK, transactions)
}

// TransferTokens transfers tokens between users
func TransferTokens(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	senderIDInt, ok := userID.(int)
	if !ok {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	var req struct {
		RecipientID    int     `json:"recipient_id"`
		RecipientEmail string  `json:"recipient_email"`
		Amount         float64 `json:"amount" binding:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	email := strings.TrimSpace(req.RecipientEmail)
	if email == "" && req.RecipientID <= 0 {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"INVALID_REQUEST",
			"请填写接收方注册邮箱或数字用户ID",
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

	var recipientID int
	var err error

	switch {
	case email != "" && req.RecipientID != 0:
		middleware.RespondWithError(c, apperrors.NewAppError(
			"INVALID_REQUEST",
			"请只填写接收方注册邮箱或用户ID其中一项",
			http.StatusBadRequest,
			nil,
		))
		return
	case email != "":
		err = db.QueryRow(
			`SELECT id FROM users WHERE LOWER(email) = $1 AND status = 'active'`,
			strings.ToLower(email),
		).Scan(&recipientID)
		if err == sql.ErrNoRows {
			middleware.RespondWithError(c, apperrors.ErrUserNotFound)
			return
		}
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
	case req.RecipientID > 0:
		err = db.QueryRow(
			`SELECT id FROM users WHERE id = $1 AND status = 'active'`,
			req.RecipientID,
		).Scan(&recipientID)
		if err == sql.ErrNoRows {
			middleware.RespondWithError(c, apperrors.ErrUserNotFound)
			return
		}
		if err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
	default:
		middleware.RespondWithError(c, apperrors.NewAppError(
			"INVALID_REQUEST",
			"请填写接收方注册邮箱或数字用户ID",
			http.StatusBadRequest,
			nil,
		))
		return
	}

	if recipientID == senderIDInt {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"INVALID_REQUEST",
			"不能向自己的账户转账",
			http.StatusBadRequest,
			nil,
		))
		return
	}

	tx, err := db.Begin()
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer tx.Rollback()

	var senderBalance float64
	err = tx.QueryRow(
		`SELECT balance FROM tokens WHERE user_id = $1 FOR UPDATE`,
		senderIDInt,
	).Scan(&senderBalance)
	if err == sql.ErrNoRows {
		middleware.RespondWithError(c, apperrors.ErrTokenNotFound)
		return
	}
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	if err = billing.ForfeitExpiredLots(tx, senderIDInt); err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	err = tx.QueryRow(`SELECT balance FROM tokens WHERE user_id = $1`, senderIDInt).Scan(&senderBalance)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	if senderBalance < req.Amount {
		middleware.RespondWithError(c, apperrors.ErrInsufficientBalance)
		return
	}

	if err = billing.DebitLotsFIFO(tx, senderIDInt, req.Amount, false); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInsufficientBalance)
		return
	}

	if err = billing.CreditLegacyLot(tx, recipientID, req.Amount, "transfer_in"); err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"TOKEN_TRANSFER_FAILED",
			"Failed to transfer tokens",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	_, err = tx.Exec(
		`INSERT INTO token_transactions (user_id, type, amount, reason) VALUES ($1, $2, $3, $4)`,
		senderIDInt, "transfer", -req.Amount, "Transfer to user",
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"TOKEN_TRANSFER_FAILED",
			"Failed to record sender transaction",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	_, err = tx.Exec(
		`INSERT INTO token_transactions (user_id, type, amount, reason) VALUES ($1, $2, $3, $4)`,
		recipientID, "transfer", req.Amount, "Transfer from user",
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"TOKEN_TRANSFER_FAILED",
			"Failed to record recipient transaction",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	if err = tx.Commit(); err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"TOKEN_TRANSFER_FAILED",
			"Failed to commit transfer",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	ctx := context.Background()
	cache.Delete(ctx, cache.TokenBalanceKey(senderIDInt))
	cache.Delete(ctx, cache.TokenBalanceKey(recipientID))

	c.JSON(http.StatusOK, gin.H{"message": "Transfer successful"})
}

// GetTokenLots returns non-empty token lots for FIFO / expiry display (加油包批次).
func GetTokenLots(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}
	uid, ok := userID.(int)
	if !ok {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}
	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	rows, err := db.Query(`
		SELECT id, remaining_amount, expires_at, lot_type, order_item_id, created_at
		  FROM token_lots
		 WHERE user_id = $1 AND remaining_amount > 0
		 ORDER BY expires_at ASC NULLS LAST, id ASC`,
		uid,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	var out []gin.H
	for rows.Next() {
		var id int
		var rem float64
		var exp sql.NullTime
		var lotType string
		var oi sql.NullInt64
		var createdAt interface{}
		if err = rows.Scan(&id, &rem, &exp, &lotType, &oi, &createdAt); err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		row := gin.H{
			"id":               id,
			"remaining_amount": rem,
			"lot_type":         lotType,
			"created_at":       createdAt,
			"expires_at":       nil,
			"order_item_id":    nil,
		}
		if exp.Valid {
			row["expires_at"] = exp.Time.UTC().Format(time.RFC3339)
		}
		if oi.Valid {
			row["order_item_id"] = oi.Int64
		}
		out = append(out, row)
	}
	if err = rows.Err(); err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	if out == nil {
		out = []gin.H{}
	}
	c.JSON(http.StatusOK, gin.H{"data": out})
}
