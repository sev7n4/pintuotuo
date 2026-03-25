package handlers

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/models"
	"github.com/pintuotuo/backend/services/token"
	"github.com/pintuotuo/backend/services/user"
)

var jwtSecret []byte

func init() {
	jwtSecret = []byte(getEnv("JWT_SECRET", "pintuotuo-secret-key-dev"))
}

// Initialize user service
var userService user.Service

func init() {
	logger := log.New(os.Stderr, "[AuthHandler] ", log.LstdFlags)
	tokenSvc := token.NewService(config.GetDB(), logger)
	userService = user.NewService(config.GetDB(), logger, tokenSvc)
}

// AuthResponse represents authentication response
type AuthResponse struct {
	User  *models.User `json:"user"`
	Token string       `json:"token"`
}

// RefreshTokenResponse represents token refresh response
type RefreshTokenResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

// RegisterUser handles user registration
func RegisterUser(c *gin.Context) {
	var req user.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	// Call service
	usr, err := userService.RegisterUser(c.Request.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.RespondWithError(c, appErr)
		} else {
			middleware.RespondWithError(c, apperrors.NewAppError(
				"REGISTRATION_FAILED",
				"Failed to register user",
				http.StatusInternalServerError,
				err,
			))
		}
		return
	}

	// Map service user to model
	userModel := mapServiceUserToModel(usr)

	// Generate token
	token := generateToken(usr.ID, usr.Email)

	// Generate JWT token
	token := generateToken(user.ID, user.Email, user.Role)
	
	c.JSON(http.StatusCreated, gin.H{
		"user":  userModel,
		"token": token,
	})
}

// LoginUser handles user login
func LoginUser(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	// Call service
	usr, err := userService.AuthenticateUser(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.RespondWithError(c, appErr)
		} else {
			middleware.RespondWithError(c, apperrors.ErrInvalidCredentials)
		}
		return
	}

	// Map service user to model
	userModel := mapServiceUserToModel(usr)

	// Generate JWT token
	token := generateToken(usr.ID, usr.Email)

	c.JSON(http.StatusOK, gin.H{
		"user":  userModel,
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
	var req struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	// Call service
	_, newToken, expiresAt, err := userService.RefreshToken(c.Request.Context(), req.Token)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.RespondWithError(c, appErr)
		} else {
			middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":      newToken,
		"expires_at": expiresAt,
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

	// Call service
	err := userService.RequestPasswordReset(c.Request.Context(), req.Email)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.RespondWithError(c, appErr)
		} else {
			// Don't expose error details for security
		}
		// Always return success response
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "If the email exists, a password reset link has been sent",
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

	// Call service
	err := userService.ResetPassword(c.Request.Context(), req.ResetToken, req.NewPassword)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.RespondWithError(c, appErr)
		} else {
			middleware.RespondWithError(c, apperrors.NewAppError(
				"PASSWORD_RESET_FAILED",
				"Failed to reset password",
				http.StatusInternalServerError,
				err,
			))
		}
		return
	}

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

	userIDInt, ok := userID.(int)
	if !ok {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	// Call service
	usr, err := userService.GetCurrentUser(c.Request.Context(), userIDInt)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.RespondWithError(c, appErr)
		} else {
			middleware.RespondWithError(c, apperrors.ErrUserNotFound)
		}
		return
	}

	// Map service user to model
	userModel := mapServiceUserToModel(usr)
	c.JSON(http.StatusOK, userModel)
}

// UpdateCurrentUser updates current user profile
func UpdateCurrentUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	var req user.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	userIDInt, _ := userID.(int)

	// Call service
	usr, err := userService.UpdateUserProfile(c.Request.Context(), userIDInt, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.RespondWithError(c, appErr)
		} else {
			middleware.RespondWithError(c, apperrors.NewAppError(
				"USER_UPDATE_FAILED",
				"Failed to update user",
				http.StatusInternalServerError,
				err,
			))
		}
		return
	}

	// Map service user to model
	userModel := mapServiceUserToModel(usr)
	c.JSON(http.StatusOK, userModel)
}

// GetUserByID retrieves user by ID
func GetUserByID(c *gin.Context) {
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	// Call service
	usr, err := userService.GetUserByID(c.Request.Context(), idInt)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.RespondWithError(c, appErr)
		} else {
			middleware.RespondWithError(c, apperrors.ErrUserNotFound)
		}
		return
	}

	// Map service user to model
	userModel := mapServiceUserToModel(usr)
	c.JSON(http.StatusOK, userModel)
}

// UpdateUser updates user by ID (admin only)
func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	var req user.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	// Call service
	usr, err := userService.UpdateUserProfile(c.Request.Context(), idInt, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.RespondWithError(c, appErr)
		} else {
			middleware.RespondWithError(c, apperrors.NewAppError(
				"USER_UPDATE_FAILED",
				"Failed to update user",
				http.StatusInternalServerError,
				err,
			))
		}
		return
	}

	// Map service user to model
	userModel := mapServiceUserToModel(usr)
	c.JSON(http.StatusOK, userModel)
}

// RefreshToken refreshes the JWT token for authenticated user
func RefreshToken(c *gin.Context) {
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

	// Verify user still exists and is active
	var user models.User
	err := db.QueryRow(
		"SELECT id, email, name, role, created_at, updated_at FROM users WHERE id = $1 AND status = 'active'",
		userID,
	).Scan(&user.ID, &user.Email, &user.Name, &user.Role, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrUserNotFound)
		return
	}

	// Generate new JWT token
	token := generateToken(user.ID, user.Email, user.Role)

	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"token": token,
	})
}

// RequestPasswordResetRequest represents password reset request
type RequestPasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest represents password reset with token
type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

// RequestPasswordReset handles password reset request
func RequestPasswordReset(c *gin.Context) {
	var req RequestPasswordResetRequest
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
	var userID int
	err := db.QueryRow(
		"SELECT id FROM users WHERE email = $1 AND status = 'active'",
		req.Email,
	).Scan(&userID)

	if err != nil {
		// Don't reveal if user exists or not for security
		c.JSON(http.StatusOK, gin.H{
			"message": "If the email exists, a reset link has been sent",
		})
		return
	}

	// Generate reset token (valid for 1 hour)
	resetToken := generateResetToken(userID)

	// Store reset token in database
	_, err = db.Exec(
		`INSERT INTO password_reset_tokens (user_id, token, expires_at) 
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id) DO UPDATE SET token = $2, expires_at = $3, created_at = NOW()`,
		userID, resetToken, time.Now().Add(time.Hour),
	)

	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"RESET_TOKEN_FAILED",
			"Failed to generate reset token",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	// In production, send email with reset link
	// For demo, log the reset token
	fmt.Printf("[Password Reset] Email: %s, Reset Token: %s\n", req.Email, resetToken)

	c.JSON(http.StatusOK, gin.H{
		"message": "If the email exists, a reset link has been sent",
		// Only in development - remove in production
		"debug_token": resetToken,
	})
}

// ResetPassword handles password reset with token
func ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	// Verify reset token
	var userID int
	var expiresAt time.Time
	err := db.QueryRow(
		"SELECT user_id, expires_at FROM password_reset_tokens WHERE token = $1",
		req.Token,
	).Scan(&userID, &expiresAt)

	if err != nil || time.Now().After(expiresAt) {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"INVALID_RESET_TOKEN",
			"Invalid or expired reset token",
			http.StatusBadRequest,
			err,
		))
		return
	}

	// Hash new password
	passwordHash := hashPassword(req.Password)

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

	// Delete used reset token
	_, _ = db.Exec("DELETE FROM password_reset_tokens WHERE token = $1", req.Token)

	// Invalidate user cache
	ctx := context.Background()
	cache.Delete(ctx, cache.UserKey(userID))

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

func generateResetToken(userID int) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"type":    "password_reset",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})

	tokenString, _ := token.SignedString(jwtSecret)
	return tokenString
}

// Helper functions

func generateToken(userID int, email string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, _ := token.SignedString(jwtSecret)
	return tokenString
}

func mapServiceUserToModel(u *user.User) *models.User {
	return &models.User{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
