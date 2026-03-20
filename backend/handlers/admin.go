package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/models"
)

// GetAdminUsers retrieves all users (admin only)
func GetAdminUsers(c *gin.Context) {
	userRole, exists := c.Get("user_role")
	if !exists || userRole != "admin" {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"FORBIDDEN",
			"Admin access required",
			http.StatusForbidden,
			nil,
		))
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	rows, err := db.Query(
		"SELECT id, email, name, role, created_at, updated_at FROM users ORDER BY created_at DESC",
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Email, &user.Name, &user.Role, &user.CreatedAt, &user.UpdatedAt); err != nil {
			continue
		}
		users = append(users, user)
	}

	if users == nil {
		users = []models.User{}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    users,
	})
}

// CreateAdminUser creates a new admin user (admin only)
func CreateAdminUser(c *gin.Context) {
	userRole, exists := c.Get("user_role")
	if !exists || userRole != "admin" {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"FORBIDDEN",
			"Admin access required",
			http.StatusForbidden,
			nil,
		))
		return
	}

	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Name     string `json:"name" binding:"required,min=2"`
		Password string `json:"password" binding:"required,min=6"`
		Role     string `json:"role"`
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

	// Check if user exists
	var existingUser models.User
	err := db.QueryRow(
		"SELECT id FROM users WHERE email = $1",
		req.Email,
	).Scan(&existingUser.ID)
	if err == nil {
		middleware.RespondWithError(c, apperrors.ErrUserAlreadyExists)
		return
	}

	// Hash password
	passwordHash := hashPassword(req.Password)

	// Determine role (default to admin)
	role := "admin"
	if req.Role == "merchant" || req.Role == "user" {
		role = req.Role
	}

	// Create user
	var user models.User
	err = db.QueryRow(
		"INSERT INTO users (email, name, password_hash, role, status) VALUES ($1, $2, $3, $4, $5) RETURNING id, email, name, role, created_at, updated_at",
		req.Email, req.Name, passwordHash, role, "active",
	).Scan(&user.ID, &user.Email, &user.Name, &user.Role, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"USER_CREATION_FAILED",
			"Failed to create user",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	// Create token balance
	_, err = db.Exec(
		"INSERT INTO tokens (user_id, balance) VALUES ($1, $2)",
		user.ID, 0,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"TOKEN_INIT_FAILED",
			"Failed to initialize token balance",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "success",
		"data":    user,
	})
}

// GetAdminStats retrieves platform statistics (admin only)
func GetAdminStats(c *gin.Context) {
	userRole, exists := c.Get("user_role")
	if !exists || userRole != "admin" {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"FORBIDDEN",
			"Admin access required",
			http.StatusForbidden,
			nil,
		))
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var stats struct {
		TotalUsers     int     `json:"total_users"`
		TotalMerchants int     `json:"total_merchants"`
		TotalOrders    int     `json:"total_orders"`
		TotalRevenue   float64 `json:"total_revenue"`
	}

	// Get total users
	db.QueryRow("SELECT COUNT(*) FROM users").Scan(&stats.TotalUsers)

	// Get total merchants
	db.QueryRow("SELECT COUNT(*) FROM merchants").Scan(&stats.TotalMerchants)

	// Get total orders
	db.QueryRow("SELECT COUNT(*) FROM orders").Scan(&stats.TotalOrders)

	// Get total revenue
	db.QueryRow("SELECT COALESCE(SUM(total_price), 0) FROM orders WHERE status = 'completed'").Scan(&stats.TotalRevenue)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    stats,
	})
}
