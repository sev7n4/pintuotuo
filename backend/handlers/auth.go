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

var jwtSecret = []byte(getEnv("JWT_SECRET", "pintuotuo-secret-key-dev"))

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
