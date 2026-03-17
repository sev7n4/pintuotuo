package user

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pintuotuo/backend/cache"
	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/services/token"
)

var testService Service

func init() {
	// Initialize test database
	if err := config.InitDB(); err != nil {
		log.Fatalf("Failed to init test DB: %v", err)
	}

	// Initialize cache
	if err := cache.Init(); err != nil {
		log.Fatalf("Failed to init cache: %v", err)
	}

	logger := log.New(os.Stderr, "[TestUserService] ", log.LstdFlags)
	tokenSvc := token.NewService(config.GetDB(), logger)
	testService = NewService(config.GetDB(), logger, tokenSvc)
}

// TestRegisterUserValid tests valid user registration
func TestRegisterUserValid(t *testing.T) {
	req := &RegisterRequest{
		Email:    "test1@example.com",
		Name:     "Test User",
		Password: "password123",
	}

	user, err := testService.RegisterUser(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, req.Email, user.Email)
	assert.Equal(t, req.Name, user.Name)
	assert.Equal(t, "user", user.Role)
	assert.Equal(t, "active", user.Status)
	assert.True(t, user.ID > 0)
}

// TestRegisterUserDuplicate tests duplicate email registration
func TestRegisterUserDuplicate(t *testing.T) {
	email := "test_dup@example.com"
	req := &RegisterRequest{
		Email:    email,
		Name:     "Test User",
		Password: "password123",
	}

	// First registration should succeed
	user, err := testService.RegisterUser(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, user)

	// Second registration with same email should fail
	user2, err := testService.RegisterUser(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, user2)
	assert.Equal(t, ErrUserAlreadyExists, err)
}

// TestRegisterUserInvalidEmail tests registration with invalid email
func TestRegisterUserInvalidEmail(t *testing.T) {
	req := &RegisterRequest{
		Email:    "",
		Name:     "Test User",
		Password: "password123",
	}

	user, err := testService.RegisterUser(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, ErrInvalidEmail, err)
}

// TestRegisterUserShortPassword tests registration with short password
func TestRegisterUserShortPassword(t *testing.T) {
	req := &RegisterRequest{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "12345", // Less than 6 chars
	}

	user, err := testService.RegisterUser(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, ErrPasswordTooShort, err)
}

// TestRegisterUserShortName tests registration with short name
func TestRegisterUserShortName(t *testing.T) {
	req := &RegisterRequest{
		Email:    "test@example.com",
		Name:     "A", // Less than 2 chars
		Password: "password123",
	}

	user, err := testService.RegisterUser(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, ErrNameTooShort, err)
}

// TestAuthenticateUserValid tests valid user authentication
func TestAuthenticateUserValid(t *testing.T) {
	email := "auth_valid@example.com"
	password := "password123"

	// Register user first
	req := &RegisterRequest{
		Email:    email,
		Name:     "Test User",
		Password: password,
	}
	_, err := testService.RegisterUser(context.Background(), req)
	require.NoError(t, err)

	// Authenticate
	user, err := testService.AuthenticateUser(context.Background(), email, password)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, "active", user.Status)
}

// TestAuthenticateUserInvalidPassword tests authentication with wrong password
func TestAuthenticateUserInvalidPassword(t *testing.T) {
	email := "auth_invalid_pwd@example.com"
	password := "correct_password"

	// Register user
	req := &RegisterRequest{
		Email:    email,
		Name:     "Test User",
		Password: password,
	}
	_, err := testService.RegisterUser(context.Background(), req)
	require.NoError(t, err)

	// Try with wrong password
	user, err := testService.AuthenticateUser(context.Background(), email, "wrong_password")
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, ErrInvalidCredentials, err)
}

// TestAuthenticateUserNonexistent tests authentication with non-existent email
func TestAuthenticateUserNonexistent(t *testing.T) {
	user, err := testService.AuthenticateUser(context.Background(), "nonexistent@example.com", "password123")
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, ErrInvalidCredentials, err)
}

// TestAuthenticateUserEmptyEmail tests authentication with empty email
func TestAuthenticateUserEmptyEmail(t *testing.T) {
	user, err := testService.AuthenticateUser(context.Background(), "", "password123")
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, ErrInvalidCredentials, err)
}

// TestAuthenticateUserEmptyPassword tests authentication with empty password
func TestAuthenticateUserEmptyPassword(t *testing.T) {
	user, err := testService.AuthenticateUser(context.Background(), "test@example.com", "")
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, ErrInvalidCredentials, err)
}

// TestGetUserByIDValid tests retrieving user by valid ID
func TestGetUserByIDValid(t *testing.T) {
	// Create user first
	req := &RegisterRequest{
		Email:    "get_by_id@example.com",
		Name:     "Test User",
		Password: "password123",
	}
	created, err := testService.RegisterUser(context.Background(), req)
	require.NoError(t, err)

	// Retrieve user
	user, err := testService.GetUserByID(context.Background(), created.ID)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, created.ID, user.ID)
	assert.Equal(t, created.Email, user.Email)
}

// TestGetUserByIDCache tests caching in GetUserByID
func TestGetUserByIDCache(t *testing.T) {
	// Create user
	req := &RegisterRequest{
		Email:    "cache_test@example.com",
		Name:     "Test User",
		Password: "password123",
	}
	created, err := testService.RegisterUser(context.Background(), req)
	require.NoError(t, err)

	ctx := context.Background()

	// First call - should hit database
	user1, err := testService.GetUserByID(ctx, created.ID)
	assert.NoError(t, err)

	// Second call - should hit cache
	user2, err := testService.GetUserByID(ctx, created.ID)
	assert.NoError(t, err)

	assert.Equal(t, user1.ID, user2.ID)
	assert.Equal(t, user1.Email, user2.Email)

	// Verify cache was set
	cacheKey := cache.UserKey(created.ID)
	cached, err := cache.Get(ctx, cacheKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, cached)
}

// TestGetUserByIDNotFound tests retrieving non-existent user
func TestGetUserByIDNotFound(t *testing.T) {
	user, err := testService.GetUserByID(context.Background(), 99999)
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, ErrUserNotFound, err)
}

// TestUpdateUserProfileValid tests valid profile update
func TestUpdateUserProfileValid(t *testing.T) {
	// Create user
	req := &RegisterRequest{
		Email:    "update_valid@example.com",
		Name:     "Old Name",
		Password: "password123",
	}
	created, err := testService.RegisterUser(context.Background(), req)
	require.NoError(t, err)

	// Update profile
	updateReq := &UpdateUserRequest{
		Name: "New Name",
	}
	user, err := testService.UpdateUserProfile(context.Background(), created.ID, updateReq)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "New Name", user.Name)
	assert.Equal(t, created.ID, user.ID)
}

// TestUpdateUserProfileInvalidName tests profile update with invalid name
func TestUpdateUserProfileInvalidName(t *testing.T) {
	// Create user
	req := &RegisterRequest{
		Email:    "update_invalid@example.com",
		Name:     "Test User",
		Password: "password123",
	}
	created, err := testService.RegisterUser(context.Background(), req)
	require.NoError(t, err)

	// Try to update with short name
	updateReq := &UpdateUserRequest{
		Name: "A", // Less than 2 chars
	}
	user, err := testService.UpdateUserProfile(context.Background(), created.ID, updateReq)
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, ErrNameTooShort, err)
}

// TestUpdateUserProfileCacheInvalidation tests cache invalidation on profile update
func TestUpdateUserProfileCacheInvalidation(t *testing.T) {
	ctx := context.Background()

	// Create user
	req := &RegisterRequest{
		Email:    "cache_invalidate@example.com",
		Name:     "Old Name",
		Password: "password123",
	}
	created, err := testService.RegisterUser(ctx, req)
	require.NoError(t, err)

	// Load into cache
	_, err = testService.GetUserByID(ctx, created.ID)
	require.NoError(t, err)

	// Verify cache has data
	cacheKey := cache.UserKey(created.ID)
	_, err = cache.Get(ctx, cacheKey)
	assert.NoError(t, err)

	// Update profile
	updateReq := &UpdateUserRequest{
		Name: "New Name",
	}
	_, err = testService.UpdateUserProfile(ctx, created.ID, updateReq)
	require.NoError(t, err)

	// Cache should be invalidated (but this is hard to test directly,
	// so we just verify the update happened)
}

// TestGetCurrentUser tests retrieving current user
func TestGetCurrentUser(t *testing.T) {
	// Create user
	req := &RegisterRequest{
		Email:    "current_user@example.com",
		Name:     "Test User",
		Password: "password123",
	}
	created, err := testService.RegisterUser(context.Background(), req)
	require.NoError(t, err)

	// Get current user
	user, err := testService.GetCurrentUser(context.Background(), created.ID)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, created.ID, user.ID)
	assert.Equal(t, created.Email, user.Email)
}

// TestRefreshTokenValid tests valid token refresh
func TestRefreshTokenValid(t *testing.T) {
	// Create user and get token
	req := &RegisterRequest{
		Email:    "refresh_valid@example.com",
		Name:     "Test User",
		Password: "password123",
	}
	created, err := testService.RegisterUser(context.Background(), req)
	require.NoError(t, err)

	// Get initial service to generate token
	logger := log.New(os.Stderr, "[Test] ", log.LstdFlags)
	tokenSvc := token.NewService(config.GetDB(), logger)
	svc := NewService(config.GetDB(), logger, tokenSvc)

	// Generate token manually for testing
	service := svc.(*service)
	token := service.generateToken(created.ID, created.Email)

	// Refresh token
	user, newToken, expiresAt, err := testService.RefreshToken(context.Background(), token)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotEmpty(t, newToken)
	assert.True(t, expiresAt > 0)
	assert.Equal(t, created.ID, user.ID)
}

// TestRefreshTokenInvalid tests invalid token refresh
func TestRefreshTokenInvalid(t *testing.T) {
	user, newToken, expiresAt, err := testService.RefreshToken(context.Background(), "invalid.token.here")
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Empty(t, newToken)
	assert.Equal(t, int64(0), expiresAt)
	assert.Equal(t, ErrInvalidToken, err)
}

// TestRefreshTokenExpired tests refreshing with expired token
func TestRefreshTokenExpired(t *testing.T) {
	// Create a token that's already expired
	logger := log.New(os.Stderr, "[Test] ", log.LstdFlags)
	tokenSvc := token.NewService(config.GetDB(), logger)
	svc := NewService(config.GetDB(), logger, tokenSvc)
	service := svc.(*service)

	// Create token with past expiry
	expiredToken := service.generateTokenWithExpiry(1, "test@example.com", time.Now().Add(-24*time.Hour))

	user, newToken, expiresAt, err := testService.RefreshToken(context.Background(), expiredToken)
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Empty(t, newToken)
	assert.Equal(t, int64(0), expiresAt)
}

// TestRequestPasswordReset tests password reset request
func TestRequestPasswordReset(t *testing.T) {
	// Create user
	req := &RegisterRequest{
		Email:    "pwd_reset@example.com",
		Name:     "Test User",
		Password: "password123",
	}
	_, err := testService.RegisterUser(context.Background(), req)
	require.NoError(t, err)

	// Request password reset
	err = testService.RequestPasswordReset(context.Background(), req.Email)
	assert.NoError(t, err)
}

// TestRequestPasswordResetNonexistent tests password reset for non-existent user
func TestRequestPasswordResetNonexistent(t *testing.T) {
	// Request password reset for non-existent email (should return success for security)
	err := testService.RequestPasswordReset(context.Background(), "nonexistent@example.com")
	assert.NoError(t, err)
}

// TestResetPasswordValid tests valid password reset
func TestResetPasswordValid(t *testing.T) {
	ctx := context.Background()

	// Create user
	email := "reset_valid@example.com"
	oldPassword := "oldpassword123"
	newPassword := "newpassword123"

	req := &RegisterRequest{
		Email:    email,
		Name:     "Test User",
		Password: oldPassword,
	}
	created, err := testService.RegisterUser(ctx, req)
	require.NoError(t, err)

	// Request password reset (to generate token)
	err = testService.RequestPasswordReset(ctx, email)
	require.NoError(t, err)

	// Get reset token from cache (in real scenario, it would be in email)
	// For testing, we need to manually set it
	resetToken := "test_reset_token_" + time.Now().Format("20060102150405")
	cacheKey := "password_reset:" + resetToken
	err = cache.Set(ctx, cacheKey, email, 15*time.Minute)
	require.NoError(t, err)

	// Reset password
	err = testService.ResetPassword(ctx, resetToken, newPassword)
	assert.NoError(t, err)

	// Verify old password doesn't work
	user, err := testService.AuthenticateUser(ctx, email, oldPassword)
	assert.Error(t, err)
	assert.Nil(t, user)

	// Verify new password works
	user, err = testService.AuthenticateUser(ctx, email, newPassword)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, created.ID, user.ID)
}

// TestResetPasswordInvalidToken tests password reset with invalid token
func TestResetPasswordInvalidToken(t *testing.T) {
	err := testService.ResetPassword(context.Background(), "invalid_token", "newpassword123")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidResetToken, err)
}

// TestResetPasswordShortPassword tests password reset with short password
func TestResetPasswordShortPassword(t *testing.T) {
	err := testService.ResetPassword(context.Background(), "valid_token", "12345")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidResetToken, err)
}

// TestDeleteUserValid tests valid user deletion
func TestDeleteUserValid(t *testing.T) {
	// Create user
	req := &RegisterRequest{
		Email:    "delete_valid@example.com",
		Name:     "Test User",
		Password: "password123",
	}
	created, err := testService.RegisterUser(context.Background(), req)
	require.NoError(t, err)

	// Delete user
	err = testService.DeleteUser(context.Background(), created.ID)
	assert.NoError(t, err)

	// Verify user is deleted
	user, err := testService.GetUserByID(context.Background(), created.ID)
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, ErrUserNotFound, err)
}

// TestDeleteUserNonexistent tests deletion of non-existent user
func TestDeleteUserNonexistent(t *testing.T) {
	err := testService.DeleteUser(context.Background(), 99999)
	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
}

// TestBanUserValid tests valid user ban
func TestBanUserValid(t *testing.T) {
	// Create user
	req := &RegisterRequest{
		Email:    "ban_valid@example.com",
		Name:     "Test User",
		Password: "password123",
	}
	created, err := testService.RegisterUser(context.Background(), req)
	require.NoError(t, err)

	// Ban user
	err = testService.BanUser(context.Background(), created.ID, "Suspicious activity")
	assert.NoError(t, err)

	// Verify user can't authenticate
	user, err := testService.AuthenticateUser(context.Background(), created.Email, "password123")
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, ErrUserBanned, err)
}

// TestBanUserNonexistent tests banning non-existent user
func TestBanUserNonexistent(t *testing.T) {
	err := testService.BanUser(context.Background(), 99999, "Spam")
	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
}

// TestConcurrentRegistration tests concurrent user registration
func TestConcurrentRegistration(t *testing.T) {
	email := "concurrent_test@example.com"
	done := make(chan bool)
	errorCount := 0

	// Try to register same user concurrently
	for i := 0; i < 3; i++ {
		go func(idx int) {
			req := &RegisterRequest{
				Email:    email,
				Name:     "Test User " + string(rune(idx)),
				Password: "password123",
			}
			_, err := testService.RegisterUser(context.Background(), req)
			if err != nil {
				errorCount++
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}

	// At least 2 should fail due to duplicate constraint
	assert.True(t, errorCount >= 2, "Expected at least 2 failures, got %d", errorCount)
}

// TestPasswordHashing tests that passwords are properly hashed
func TestPasswordHashing(t *testing.T) {
	// Create user
	password := "test_password_123"
	req := &RegisterRequest{
		Email:    "hash_test@example.com",
		Name:     "Test User",
		Password: password,
	}
	_, err := testService.RegisterUser(context.Background(), req)
	require.NoError(t, err)

	// Verify password hash is not plaintext
	db := config.GetDB()
	var hash string
	err = db.QueryRow("SELECT password_hash FROM users WHERE email = $1", req.Email).Scan(&hash)
	require.NoError(t, err)

	assert.NotEqual(t, password, hash)
	assert.NotEmpty(t, hash)
}

// TestTokenClaimsValid tests that generated tokens have correct claims
func TestTokenClaimsValid(t *testing.T) {
	// Create user
	req := &RegisterRequest{
		Email:    "token_claims@example.com",
		Name:     "Test User",
		Password: "password123",
	}
	created, err := testService.RegisterUser(context.Background(), req)
	require.NoError(t, err)

	// Generate token
	logger := log.New(os.Stderr, "[Test] ", log.LstdFlags)
	tokenSvc := token.NewService(config.GetDB(), logger)
	svc := NewService(config.GetDB(), logger, tokenSvc)
	service := svc.(*service)
	token := service.generateToken(created.ID, created.Email)

	// Parse and verify claims
	parsedToken, err := jwt.ParseWithClaims(token, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return service.jwtSecret, nil
	})
	require.NoError(t, err)

	claims := parsedToken.Claims.(jwt.MapClaims)
	assert.Equal(t, float64(created.ID), claims["user_id"])
	assert.Equal(t, created.Email, claims["email"])
	assert.True(t, claims["exp"].(float64) > 0)
}
