package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/cache"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/models"
)

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

	// Invalidate token balance caches for both users
	ctx := context.Background()
	senderIDInt, _ := userID.(int)
	cache.Delete(ctx, cache.TokenBalanceKey(senderIDInt))
	cache.Delete(ctx, cache.TokenBalanceKey(req.RecipientID))

	c.JSON(http.StatusOK, gin.H{"message": "Transfer successful"})
}
