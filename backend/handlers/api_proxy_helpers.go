package handlers

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
)

var (
	ErrUserNotAuthenticated = errors.New("user not authenticated")
	ErrInvalidUserIDType    = errors.New("invalid user_id type")
)

func authenticateUser(c *gin.Context) (int, error) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		return 0, ErrUserNotAuthenticated
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
		return 0, fmt.Errorf("%w: got %T", ErrInvalidUserIDType, userIDVal)
	}

	return userID, nil
}

func parseAPIProxyRequest(c *gin.Context) (*APIProxyRequest, error) {
	var req APIProxyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, fmt.Errorf("invalid request body: %w", err)
	}

	if req.Model == "" {
		return nil, errors.New("model is required")
	}

	return &req, nil
}

func getTokenBalance(db *sql.DB, userID int) (float64, error) {
	if db == nil {
		return 0, errors.New("database connection is nil")
	}

	var balance float64
	err := db.QueryRow("SELECT balance FROM tokens WHERE user_id = $1", userID).Scan(&balance)
	if err != nil {
		return 0, fmt.Errorf("failed to get token balance: %w", err)
	}

	return balance, nil
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
