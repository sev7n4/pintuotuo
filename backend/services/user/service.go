package user

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pintuotuo/backend/cache"
)

// Service defines the user service interface
type Service interface {
	// Authentication
	RegisterUser(ctx context.Context, req *RegisterRequest) (*User, error)
	AuthenticateUser(ctx context.Context, email, password string) (*User, error)
	RefreshToken(ctx context.Context, token string) (*User, string, int64, error)

	// Profile Management
	GetUserByID(ctx context.Context, userID int) (*User, error)
	UpdateUserProfile(ctx context.Context, userID int, req *UpdateUserRequest) (*User, error)
	GetCurrentUser(ctx context.Context, userID int) (*User, error)

	// Password Management
	RequestPasswordReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, resetToken, newPassword string) error

	// Account Operations
	DeleteUser(ctx context.Context, userID int) error
	BanUser(ctx context.Context, userID int, reason string) error
}

// service implements the Service interface
type service struct {
	db        *sql.DB
	log       *log.Logger
	jwtSecret []byte
}

// NewService creates a new user service
func NewService(db *sql.DB, logger *log.Logger) Service {
	if logger == nil {
		logger = log.New(os.Stderr, "[UserService] ", log.LstdFlags)
	}

	jwtSecret := []byte(getEnv("JWT_SECRET", "pintuotuo-secret-key-dev"))

	return &service{
		db:        db,
		log:       logger,
		jwtSecret: jwtSecret,
	}
}

// RegisterUser registers a new user
func (s *service) RegisterUser(ctx context.Context, req *RegisterRequest) (*User, error) {
	// Validate input
	if req.Email == "" {
		return nil, ErrInvalidEmail
	}
	if req.Password == "" || len(req.Password) < 6 {
		return nil, ErrPasswordTooShort
	}
	if req.Name == "" || len(req.Name) < 2 {
		return nil, ErrNameTooShort
	}

	// Check if user already exists
	var existingID int
	err := s.db.QueryRowContext(ctx, "SELECT id FROM users WHERE email = $1", req.Email).Scan(&existingID)
	if err == nil {
		return nil, ErrUserAlreadyExists
	}
	if err != sql.ErrNoRows {
		return nil, wrapError("RegisterUser", "checkExists", err)
	}

	// Hash password
	passwordHash := s.hashPassword(req.Password)

	// Create user in transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, wrapError("RegisterUser", "beginTx", err)
	}
	defer tx.Rollback()

	// Insert user
	var user User
	err = tx.QueryRowContext(
		ctx,
		"INSERT INTO users (email, name, password_hash, role, status) VALUES ($1, $2, $3, $4, $5) RETURNING id, email, name, role, status, created_at, updated_at",
		req.Email, req.Name, passwordHash, "user", "active",
	).Scan(&user.ID, &user.Email, &user.Name, &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, wrapError("RegisterUser", "insertUser", err)
	}

	// Initialize token balance
	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO tokens (user_id, balance) VALUES ($1, $2)",
		user.ID, 0,
	)
	if err != nil {
		return nil, wrapError("RegisterUser", "initToken", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, wrapError("RegisterUser", "commitTx", err)
	}

	s.log.Printf("User registered: id=%d, email=%s", user.ID, user.Email)
	return &user, nil
}

// AuthenticateUser authenticates user with email and password
func (s *service) AuthenticateUser(ctx context.Context, email, password string) (*User, error) {
	if email == "" || password == "" {
		return nil, ErrInvalidCredentials
	}

	// Find user by email
	var user User
	var passwordHash string
	err := s.db.QueryRowContext(
		ctx,
		"SELECT id, email, name, role, status, password_hash, created_at, updated_at FROM users WHERE email = $1",
		email,
	).Scan(&user.ID, &user.Email, &user.Name, &user.Role, &user.Status, &passwordHash, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrInvalidCredentials
		}
		return nil, wrapError("AuthenticateUser", "queryUser", err)
	}

	// Check user status
	if user.Status != "active" {
		if user.Status == "banned" {
			return nil, ErrUserBanned
		}
		return nil, ErrInvalidCredentials
	}

	// Verify password
	if !s.verifyPassword(password, passwordHash) {
		return nil, ErrInvalidCredentials
	}

	s.log.Printf("User authenticated: id=%d, email=%s", user.ID, user.Email)
	return &user, nil
}

// RefreshToken refreshes an expired JWT token
func (s *service) RefreshToken(ctx context.Context, tokenString string) (*User, string, int64, error) {
	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, "", 0, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, "", 0, ErrInvalidToken
	}

	// Extract claims
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return nil, "", 0, ErrInvalidToken
	}
	userID := int(userIDFloat)

	email, ok := claims["email"].(string)
	if !ok {
		return nil, "", 0, ErrInvalidToken
	}

	// Verify user still exists and is active
	var user User
	err = s.db.QueryRowContext(
		ctx,
		"SELECT id, email, name, role, status, created_at, updated_at FROM users WHERE id = $1 AND status = 'active'",
		userID,
	).Scan(&user.ID, &user.Email, &user.Name, &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, "", 0, ErrUserNotFound
		}
		return nil, "", 0, wrapError("RefreshToken", "queryUser", err)
	}

	// Generate new token
	newToken := s.generateToken(userID, email)
	expiresAt := time.Now().Add(24 * time.Hour).Unix()

	s.log.Printf("Token refreshed: user_id=%d", userID)
	return &user, newToken, expiresAt, nil
}

// GetUserByID retrieves a user by ID with caching
func (s *service) GetUserByID(ctx context.Context, userID int) (*User, error) {
	// Try cache first
	cacheKey := cache.UserKey(userID)
	if cachedData, err := cache.Get(ctx, cacheKey); err == nil {
		var user User
		if err := json.Unmarshal([]byte(cachedData), &user); err == nil {
			return &user, nil
		}
	}

	// Query database
	var user User
	err := s.db.QueryRowContext(
		ctx,
		"SELECT id, email, name, role, status, created_at, updated_at FROM users WHERE id = $1",
		userID,
	).Scan(&user.ID, &user.Email, &user.Name, &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, wrapError("GetUserByID", "queryUser", err)
	}

	// Cache result
	if userData, err := json.Marshal(user); err == nil {
		_ = cache.Set(ctx, cacheKey, string(userData), cache.UserCacheTTL)
	}

	return &user, nil
}

// UpdateUserProfile updates user profile
func (s *service) UpdateUserProfile(ctx context.Context, userID int, req *UpdateUserRequest) (*User, error) {
	if req.Name == "" || len(req.Name) < 2 {
		return nil, ErrNameTooShort
	}

	var user User
	err := s.db.QueryRowContext(
		ctx,
		"UPDATE users SET name = $1, updated_at = NOW() WHERE id = $2 RETURNING id, email, name, role, status, created_at, updated_at",
		req.Name, userID,
	).Scan(&user.ID, &user.Email, &user.Name, &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, wrapError("UpdateUserProfile", "updateUser", err)
	}

	// Invalidate cache
	_ = cache.Delete(ctx, cache.UserKey(userID))

	s.log.Printf("User profile updated: id=%d", userID)
	return &user, nil
}

// GetCurrentUser retrieves the current user (same as GetUserByID)
func (s *service) GetCurrentUser(ctx context.Context, userID int) (*User, error) {
	return s.GetUserByID(ctx, userID)
}

// RequestPasswordReset initiates password reset
func (s *service) RequestPasswordReset(ctx context.Context, email string) error {
	if email == "" {
		return ErrInvalidEmail
	}

	// Check if user exists (don't reveal if exists for security)
	var userID int
	err := s.db.QueryRowContext(ctx, "SELECT id FROM users WHERE email = $1", email).Scan(&userID)
	if err == sql.ErrNoRows {
		// Return success even if user doesn't exist (security)
		return nil
	}
	if err != nil {
		return wrapError("RequestPasswordReset", "queryUser", err)
	}

	// Generate reset token
	resetToken := fmt.Sprintf("%d-%d", userID, time.Now().Unix())
	cacheKey := fmt.Sprintf("password_reset:%s", resetToken)

	// Store reset token in cache (15-minute expiry)
	if err := cache.Set(ctx, cacheKey, email, 15*time.Minute); err != nil {
		return wrapError("RequestPasswordReset", "cacheToken", err)
	}

	// In production: send email with reset link
	s.log.Printf("Password reset requested for user: id=%d, email=%s", userID, email)
	return nil
}

// ResetPassword resets user password with reset token
func (s *service) ResetPassword(ctx context.Context, resetToken, newPassword string) error {
	if resetToken == "" || newPassword == "" || len(newPassword) < 6 {
		return ErrInvalidResetToken
	}

	// Verify reset token in cache
	cacheKey := fmt.Sprintf("password_reset:%s", resetToken)
	email, err := cache.Get(ctx, cacheKey)
	if err != nil {
		return ErrInvalidResetToken
	}

	// Find user by email
	var userID int
	err = s.db.QueryRowContext(ctx, "SELECT id FROM users WHERE email = $1", email).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrUserNotFound
		}
		return wrapError("ResetPassword", "queryUser", err)
	}

	// Hash new password
	passwordHash := s.hashPassword(newPassword)

	// Update password
	_, err = s.db.ExecContext(
		ctx,
		"UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2",
		passwordHash, userID,
	)
	if err != nil {
		return wrapError("ResetPassword", "updatePassword", err)
	}

	// Invalidate reset token and user cache
	_ = cache.Delete(ctx, cacheKey)
	_ = cache.Delete(ctx, cache.UserKey(userID))

	s.log.Printf("Password reset for user: id=%d", userID)
	return nil
}

// DeleteUser deletes a user account
func (s *service) DeleteUser(ctx context.Context, userID int) error {
	result, err := s.db.ExecContext(
		ctx,
		"DELETE FROM users WHERE id = $1",
		userID,
	)
	if err != nil {
		return wrapError("DeleteUser", "deleteUser", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return wrapError("DeleteUser", "rowsAffected", err)
	}

	if affected == 0 {
		return ErrUserNotFound
	}

	// Invalidate cache
	_ = cache.Delete(ctx, cache.UserKey(userID))

	s.log.Printf("User deleted: id=%d", userID)
	return nil
}

// BanUser bans a user account
func (s *service) BanUser(ctx context.Context, userID int, reason string) error {
	result, err := s.db.ExecContext(
		ctx,
		"UPDATE users SET status = 'banned', updated_at = NOW() WHERE id = $1",
		userID,
	)
	if err != nil {
		return wrapError("BanUser", "updateStatus", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return wrapError("BanUser", "rowsAffected", err)
	}

	if affected == 0 {
		return ErrUserNotFound
	}

	// Invalidate cache
	_ = cache.Delete(ctx, cache.UserKey(userID))

	s.log.Printf("User banned: id=%d, reason=%s", userID, reason)
	return nil
}

// Helper functions

func (s *service) hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password + string(s.jwtSecret)))
	return fmt.Sprintf("%x", hash)
}

func (s *service) verifyPassword(password, hash string) bool {
	return s.hashPassword(password) == hash
}

func (s *service) generateToken(userID int, email string) string {
	return s.generateTokenWithExpiry(userID, email, time.Now().Add(24*time.Hour))
}

func (s *service) generateTokenWithExpiry(userID int, email string, expiresAt time.Time) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     expiresAt.Unix(),
	})

	tokenString, _ := token.SignedString(s.jwtSecret)
	return tokenString
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
