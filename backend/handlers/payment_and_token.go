package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/models"
)

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

	if order.Status != "pending" {
		middleware.RespondWithError(c, apperrors.ErrOrderAlreadyPaid)
		return
	}

	var payment models.Payment
	err = db.QueryRow(
		"INSERT INTO payments (order_id, user_id, amount, method, status) VALUES ($1, $2, $3, $4, $5) RETURNING id, order_id, amount, method, status, created_at, updated_at",
		req.OrderID, userID, order.TotalPrice, req.Method, "pending",
	).Scan(&payment.ID, &payment.OrderID, &payment.Amount, &payment.Method, &payment.Status, &payment.CreatedAt, &payment.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"PAYMENT_CREATION_FAILED",
			"Failed to create payment",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	c.JSON(http.StatusCreated, payment)
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

	var payment models.Payment
	err := db.QueryRow(
		"SELECT id, order_id, amount, method, status, created_at, updated_at FROM payments WHERE id = $1 AND user_id = $2",
		id, userID,
	).Scan(&payment.ID, &payment.OrderID, &payment.Amount, &payment.Method, &payment.Status, &payment.CreatedAt, &payment.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrPaymentNotFound)
		return
	}

	c.JSON(http.StatusOK, payment)
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

	// Get payment details
	var payment models.Payment
	err := db.QueryRow(
		"SELECT id, order_id, amount, method, status FROM payments WHERE id = $1 AND user_id = $2",
		id, userID,
	).Scan(&payment.ID, &payment.OrderID, &payment.Amount, &payment.Method, &payment.Status)

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrPaymentNotFound)
		return
	}

	if payment.Status != "success" {
		middleware.RespondWithError(c, apperrors.ErrPaymentAlreadyProcessed)
		return
	}

	// Update payment status
	err = db.QueryRow(
		"UPDATE payments SET status = $1 WHERE id = $2 RETURNING id, order_id, amount, method, status, created_at, updated_at",
		"refunded", id,
	).Scan(&payment.ID, &payment.OrderID, &payment.Amount, &payment.Method, &payment.Status, &payment.CreatedAt, &payment.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"REFUND_FAILED",
			"Failed to process refund",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	c.JSON(http.StatusOK, payment)
}

// HandleAlipayCallback handles Alipay payment callback
func HandleAlipayCallback(c *gin.Context) {
	var req struct {
		PaymentID   int    `json:"payment_id" binding:"required"`
		Status      string `json:"status" binding:"required"`
		Amount      float64 `json:"amount" binding:"required"`
		TransactionID string `json:"transaction_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()

	// Verify payment exists
	var paymentStatus string
	err := db.QueryRow("SELECT status FROM payments WHERE id = $1", req.PaymentID).Scan(&paymentStatus)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrPaymentNotFound)
		return
	}

	// Update payment status
	var method string
	err = db.QueryRow(
		"UPDATE payments SET status = $1, transaction_id = $2 WHERE id = $3 RETURNING method",
		req.Status, req.TransactionID, req.PaymentID,
	).Scan(&method)

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
	if req.Status == "success" {
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
	}

	c.JSON(http.StatusOK, gin.H{"message": "Callback processed"})
}

// HandleWechatCallback handles WeChat payment callback
func HandleWechatCallback(c *gin.Context) {
	var req struct {
		PaymentID   int    `json:"payment_id" binding:"required"`
		Status      string `json:"status" binding:"required"`
		Amount      float64 `json:"amount" binding:"required"`
		TransactionID string `json:"transaction_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()

	// Verify payment exists
	var paymentStatus string
	err := db.QueryRow("SELECT status FROM payments WHERE id = $1", req.PaymentID).Scan(&paymentStatus)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrPaymentNotFound)
		return
	}

	// Update payment status
	var method string
	err = db.QueryRow(
		"UPDATE payments SET status = $1, transaction_id = $2 WHERE id = $3 RETURNING method",
		req.Status, req.TransactionID, req.PaymentID,
	).Scan(&method)

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
	if req.Status == "success" {
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

	db := config.GetDB()

	var token models.Token
	err := db.QueryRow(
		"SELECT id, user_id, balance, created_at, updated_at FROM tokens WHERE user_id = $1",
		userID,
	).Scan(&token.ID, &token.UserID, &token.Balance, &token.CreatedAt, &token.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrTokenNotFound)
		return
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

	var req struct {
		RecipientID int     `json:"recipient_id" binding:"required"`
		Amount      float64 `json:"amount" binding:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()

	// Get sender balance
	var senderBalance float64
	err := db.QueryRow("SELECT balance FROM tokens WHERE user_id = $1", userID).Scan(&senderBalance)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrTokenNotFound)
		return
	}

	if senderBalance < req.Amount {
		middleware.RespondWithError(c, apperrors.ErrInsufficientBalance)
		return
	}

	// Transfer tokens
	_, err = db.Exec(
		"UPDATE tokens SET balance = balance - $1 WHERE user_id = $2",
		req.Amount, userID,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"TOKEN_TRANSFER_FAILED",
			"Failed to transfer tokens",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	_, err = db.Exec(
		"UPDATE tokens SET balance = balance + $1 WHERE user_id = $2",
		req.Amount, req.RecipientID,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"TOKEN_TRANSFER_FAILED",
			"Failed to transfer tokens",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	// Record transaction
	_, err = db.Exec(
		"INSERT INTO token_transactions (user_id, type, amount, reason) VALUES ($1, $2, $3, $4)",
		userID, "transfer", -req.Amount, "Transfer to user",
	)

	_, err = db.Exec(
		"INSERT INTO token_transactions (user_id, type, amount, reason) VALUES ($1, $2, $3, $4)",
		req.RecipientID, "transfer", req.Amount, "Transfer from user",
	)

	c.JSON(http.StatusOK, gin.H{"message": "Transfer successful"})
}
