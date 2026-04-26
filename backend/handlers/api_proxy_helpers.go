package handlers

import (
	"database/sql"
	"fmt"

	"github.com/gin-gonic/gin"
)

func authenticateUser(c *gin.Context) (int, error) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		return 0, fmt.Errorf("user not authenticated")
	}

	var userID int
	switch v := userIDVal.(type) {
	case int:
		userID = v
	case int64:
		userID = int(v)
	case float64:
		userID = int(v)
	default:
		return 0, fmt.Errorf("invalid user_id type")
	}

	return userID, nil
}

func parseRequest(c *gin.Context) (*APIProxyRequest, error) {
	var req APIProxyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, fmt.Errorf("invalid request body: %w", err)
	}

	if req.Model == "" {
		return nil, fmt.Errorf("model is required")
	}

	return &req, nil
}

func hasMinimumBalance(db *sql.DB, userID int) bool {
	if db == nil {
		return true
	}

	var balance float64
	err := db.QueryRow(
		`SELECT COALESCE(balance, 0) FROM users WHERE id = $1`,
		userID,
	).Scan(&balance)

	if err != nil {
		return false
	}

	return balance > 0
}

func calculateEstimatedCost(inputPrice, outputPrice float64, estimatedInputTokens, estimatedOutputTokens int) float64 {
	inputCost := inputPrice * float64(estimatedInputTokens) / 1000.0
	outputCost := outputPrice * float64(estimatedOutputTokens) / 1000.0
	return inputCost + outputCost
}

func calculateActualCost(inputPrice, outputPrice float64, actualInputTokens, actualOutputTokens int) float64 {
	inputCost := inputPrice * float64(actualInputTokens) / 1000.0
	outputCost := outputPrice * float64(actualOutputTokens) / 1000.0
	return inputCost + outputCost
}

func nullInt64Arg(n sql.NullInt64) interface{} {
	if n.Valid {
		return n.Int64
	}
	return nil
}

func nullFloat64Arg(n sql.NullFloat64) interface{} {
	if n.Valid {
		return n.Float64
	}
	return nil
}
