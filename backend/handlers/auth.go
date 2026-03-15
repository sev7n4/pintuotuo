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
