package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pintuotuo/backend/cache"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/models"
)

var jwtSecret = []byte(getEnv("JWT_SECRET", "pintuotuo-secret-key-dev"))

// RegisterRequest represents user registration data
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name" binding:"required,min=2"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest represents user login data
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	User  *models.User `json:"user"`
	Token string       `json:"token"`
}

// RefreshTokenRequest represents token refresh request
type RefreshTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

// RefreshTokenResponse represents token refresh response
type RefreshTokenResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

// RegisterUser handles user registration
func RegisterUser(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()

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

	// Create user
	var user models.User
	err = db.QueryRow(
		"INSERT INTO users (email, name, password_hash, role, status) VALUES ($1, $2, $3, $4, $5) RETURNING id, email, name, role, created_at, updated_at",
		req.Email, req.Name, passwordHash, "user", "active",
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

	// Generate JWT token
	token := generateToken(user.ID, user.Email)

	c.JSON(http.StatusCreated, gin.H{
		"user":  user,
		"token": token,
	})
}

// LoginUser handles user login
func LoginUser(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()

	// Find user by email
	var user models.User
	var passwordHash string
	err := db.QueryRow(
		"SELECT id, email, name, role, password_hash, created_at, updated_at FROM users WHERE email = $1",
		req.Email,
	).Scan(&user.ID, &user.Email, &user.Name, &user.Role, &passwordHash, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidCredentials)
		return
	}

	// Verify password
	if !verifyPassword(req.Password, passwordHash) {
		middleware.RespondWithError(c, apperrors.ErrInvalidCredentials)
		return
	}

	// Generate JWT token
	token := generateToken(user.ID, user.Email)

	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"token": token,
	})
}

// LogoutUser handles user logout
func LogoutUser(c *gin.Context) {
	// JWT is stateless, so logout is just a client-side operation
	// Server can maintain a blacklist if needed
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// RefreshToken refreshes an expired JWT token
func RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	// Parse and validate the token
	token, err := jwt.ParseWithClaims(req.Token, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	// Extract user ID and email from claims
	userID, ok := claims["user_id"].(float64)
	if !ok {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	email, ok := claims["email"].(string)
	if !ok {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	// Verify user still exists
	db := config.GetDB()
	var user models.User
	err = db.QueryRow(
		"SELECT id, email, name, role, created_at, updated_at FROM users WHERE id = $1 AND status = 'active'",
		int(userID),
	).Scan(&user.ID, &user.Email, &user.Name, &user.Role, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrUserNotFound)
		return
	}

	// Generate new token
	newToken := generateToken(int(userID), email)
	expiresAt := time.Now().Add(time.Hour * 24).Unix()

	c.JSON(http.StatusOK, RefreshTokenResponse{
		Token:     newToken,
		ExpiresAt: expiresAt,
	})
}

// RequestPasswordReset initiates password reset
func RequestPasswordReset(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()

	// Check if user exists
	var userID int
	err := db.QueryRow("SELECT id FROM users WHERE email = $1", req.Email).Scan(&userID)
	if err != nil {
		// Don't reveal if email exists for security
		c.JSON(http.StatusOK, gin.H{
			"message": "If the email exists, a password reset link has been sent",
		})
		return
	}

	// In production, generate a secure reset token and send via email
	// For now, we'll generate a simple reset token stored in cache
	resetToken := fmt.Sprintf("%d-%d", userID, time.Now().Unix())
	ctx := context.Background()
	cacheKey := fmt.Sprintf("password_reset:%s", resetToken)
	cache.Set(ctx, cacheKey, req.Email, 15*time.Minute) // 15-minute expiry

	// In production, send email with reset link containing resetToken
	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset email sent",
		// Only return reset_token in development
		"reset_token": resetToken,
	})
}

// ResetPassword resets user password with reset token
func ResetPassword(c *gin.Context) {
	var req struct {
		ResetToken  string `json:"reset_token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	// Verify reset token exists in cache
	ctx := context.Background()
	cacheKey := fmt.Sprintf("password_reset:%s", req.ResetToken)
	email, err := cache.Get(ctx, cacheKey)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"INVALID_RESET_TOKEN",
			"Password reset token is invalid or expired",
			http.StatusBadRequest,
			err,
		))
		return
	}

	db := config.GetDB()

	// Find user by email
	var userID int
	err = db.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&userID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrUserNotFound)
		return
	}

	// Hash new password
	passwordHash := hashPassword(req.NewPassword)

	// Update password
	_, err = db.Exec(
		"UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2",
		passwordHash, userID,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"PASSWORD_UPDATE_FAILED",
			"Failed to update password",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	// Invalidate reset token
	cache.Delete(ctx, cacheKey)

	// Invalidate user cache to force fresh data on next login
	cache.Delete(ctx, cache.UserKey(userID))

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset successfully",
	})
}

// GetCurrentUser retrieves current authenticated user
func GetCurrentUser(c *gin.Context) {
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
	cacheKey := cache.UserKey(userIDInt)
	if cachedUser, err := cache.Get(ctx, cacheKey); err == nil {
		var user models.User
		if err := json.Unmarshal([]byte(cachedUser), &user); err == nil {
			c.JSON(http.StatusOK, user)
			return
		}
	}

	db := config.GetDB()

	var user models.User
	err := db.QueryRow(
		"SELECT id, email, name, role, created_at, updated_at FROM users WHERE id = $1",
		userIDInt,
	).Scan(&user.ID, &user.Email, &user.Name, &user.Role, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrUserNotFound)
		return
	}

	// Cache the result
	if userJSON, err := json.Marshal(user); err == nil {
		cache.Set(ctx, cacheKey, string(userJSON), cache.UserCacheTTL)
	}

	c.JSON(http.StatusOK, user)
}

// UpdateCurrentUser updates current user profile
func UpdateCurrentUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()

	var user models.User
	err := db.QueryRow(
		"UPDATE users SET name = $1 WHERE id = $2 RETURNING id, email, name, role, created_at, updated_at",
		req.Name, userID,
	).Scan(&user.ID, &user.Email, &user.Name, &user.Role, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"USER_UPDATE_FAILED",
			"Failed to update user",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	// Invalidate user cache
	ctx := context.Background()
	userIDInt, _ := userID.(int)
	cache.Delete(ctx, cache.UserKey(userIDInt))

	c.JSON(http.StatusOK, user)
}

// GetUserByID retrieves user by ID
func GetUserByID(c *gin.Context) {
	id := c.Param("id")

	ctx := context.Background()
	idInt := idToInt(id)

	// Try cache first
	cacheKey := cache.UserKey(idInt)
	if cachedUser, err := cache.Get(ctx, cacheKey); err == nil {
		var user models.User
		if err := json.Unmarshal([]byte(cachedUser), &user); err == nil {
			c.JSON(http.StatusOK, user)
			return
		}
	}

	db := config.GetDB()

	var user models.User
	err := db.QueryRow(
		"SELECT id, email, name, role, created_at, updated_at FROM users WHERE id = $1",
		id,
	).Scan(&user.ID, &user.Email, &user.Name, &user.Role, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrUserNotFound)
		return
	}

	// Cache the result
	if userJSON, err := json.Marshal(user); err == nil {
		cache.Set(ctx, cacheKey, string(userJSON), cache.UserCacheTTL)
	}

	c.JSON(http.StatusOK, user)
}

// UpdateUser updates user by ID (admin only)
func UpdateUser(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Name string `json:"name"`
		Role string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()

	var user models.User
	err := db.QueryRow(
		"UPDATE users SET name = $1, role = $2 WHERE id = $3 RETURNING id, email, name, role, created_at, updated_at",
		req.Name, req.Role, id,
	).Scan(&user.ID, &user.Email, &user.Name, &user.Role, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"USER_UPDATE_FAILED",
			"Failed to update user",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	// Invalidate user cache
	ctx := context.Background()
	cache.Delete(ctx, cache.UserKey(idToInt(id)))

	c.JSON(http.StatusOK, user)
}

// Helper functions

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password + string(jwtSecret)))
	return fmt.Sprintf("%x", hash)
}

func verifyPassword(password, hash string) bool {
	return hashPassword(password) == hash
}

func generateToken(userID int, email string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, _ := token.SignedString(jwtSecret)
	return tokenString
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
