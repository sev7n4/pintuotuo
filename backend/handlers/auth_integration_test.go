package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/models"
)

func setupIntegrationTestDB(t *testing.T) *sql.DB {
	db := config.GetDB()
	if db == nil {
		t.Skip("Database not available for integration tests")
	}
	return db
}

func cleanupTestUser(db *sql.DB, email string) {
	db.Exec("DELETE FROM users WHERE email = $1", email)
}

func TestIntegration_MerchantRegistrationFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupIntegrationTestDB(t)

	testEmail := fmt.Sprintf("test_merchant_%d@example.com", time.Now().UnixNano())
	defer cleanupTestUser(db, testEmail)

	t.Run("merchant registration creates user and merchant records", func(t *testing.T) {
		t.Setenv("MERCHANT_REGISTER_MODE", "open")
		router := gin.New()
		router.POST("/users/register", RegisterUser)

		registerReq := RegisterRequest{
			Email:    testEmail,
			Name:     "Test Merchant",
			Password: "password123",
			Role:     "merchant",
		}
		body, _ := json.Marshal(registerReq)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/users/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code, "Merchant registration should return 201")

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, float64(0), response["code"], "Response code should be 0")
		assert.Equal(t, "success", response["message"], "Response message should be success")

		data := response["data"].(map[string]interface{})
		user := data["user"].(map[string]interface{})
		assert.Equal(t, "merchant", user["role"], "User role should be merchant")
		assert.Equal(t, testEmail, user["email"], "User email should match")

		token := data["token"].(string)
		assert.NotEmpty(t, token, "Token should be returned")

		var dbUser models.User
		err = db.QueryRow(
			"SELECT id, email, name, role FROM users WHERE email = $1",
			testEmail,
		).Scan(&dbUser.ID, &dbUser.Email, &dbUser.Name, &dbUser.Role)
		require.NoError(t, err)
		assert.Equal(t, "merchant", dbUser.Role, "DB user role should be merchant")

		var merchantExists bool
		err = db.QueryRow(
			"SELECT EXISTS(SELECT 1 FROM merchants WHERE user_id = $1)",
			dbUser.ID,
		).Scan(&merchantExists)
		require.NoError(t, err)
		assert.True(t, merchantExists, "Merchant record should exist for merchant user")
	})

	t.Run("user registration does not create merchant record", func(t *testing.T) {
		userEmail := fmt.Sprintf("test_user_%d@example.com", time.Now().UnixNano())
		defer cleanupTestUser(db, userEmail)

		router := gin.New()
		router.POST("/users/register", RegisterUser)

		registerReq := RegisterRequest{
			Email:    userEmail,
			Name:     "Test User",
			Password: "password123",
			Role:     "user",
		}
		body, _ := json.Marshal(registerReq)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/users/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code, "User registration should return 201")

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		data := response["data"].(map[string]interface{})
		user := data["user"].(map[string]interface{})
		assert.Equal(t, "user", user["role"], "User role should be user")

		var dbUser models.User
		err = db.QueryRow(
			"SELECT id, role FROM users WHERE email = $1",
			userEmail,
		).Scan(&dbUser.ID, &dbUser.Role)
		require.NoError(t, err)
		assert.Equal(t, "user", dbUser.Role, "DB user role should be user")

		var merchantExists bool
		err = db.QueryRow(
			"SELECT EXISTS(SELECT 1 FROM merchants WHERE user_id = $1)",
			dbUser.ID,
		).Scan(&merchantExists)
		require.NoError(t, err)
		assert.False(t, merchantExists, "Merchant record should NOT exist for regular user")
	})

	t.Run("registration without role defaults to user", func(t *testing.T) {
		noRoleEmail := fmt.Sprintf("test_norole_%d@example.com", time.Now().UnixNano())
		defer cleanupTestUser(db, noRoleEmail)

		router := gin.New()
		router.POST("/users/register", RegisterUser)

		registerReq := RegisterRequest{
			Email:    noRoleEmail,
			Name:     "Test No Role",
			Password: "password123",
		}
		body, _ := json.Marshal(registerReq)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/users/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code, "Registration without role should return 201")

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		data := response["data"].(map[string]interface{})
		user := data["user"].(map[string]interface{})
		assert.Equal(t, "user", user["role"], "Default role should be user")
	})

	t.Run("duplicate email registration fails", func(t *testing.T) {
		router := gin.New()
		router.POST("/users/register", RegisterUser)

		registerReq := RegisterRequest{
			Email:    testEmail,
			Name:     "Duplicate Test",
			Password: "password123",
			Role:     "user",
		}
		body, _ := json.Marshal(registerReq)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/users/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code, "Duplicate registration should return 409")
	})

	t.Run("invalid role defaults to user", func(t *testing.T) {
		invalidRoleEmail := fmt.Sprintf("test_invalidrole_%d@example.com", time.Now().UnixNano())
		defer cleanupTestUser(db, invalidRoleEmail)

		router := gin.New()
		router.POST("/users/register", RegisterUser)

		registerReq := RegisterRequest{
			Email:    invalidRoleEmail,
			Name:     "Test Invalid Role",
			Password: "password123",
			Role:     "superadmin",
		}
		body, _ := json.Marshal(registerReq)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/users/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code, "Invalid role should default to user")

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		data := response["data"].(map[string]interface{})
		user := data["user"].(map[string]interface{})
		assert.Equal(t, "user", user["role"], "Invalid role should default to user")
	})
}

func TestIntegration_LoginReturnsCorrectRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupIntegrationTestDB(t)

	t.Run("merchant login returns merchant role", func(t *testing.T) {
		merchantEmail := fmt.Sprintf("test_login_merchant_%d@example.com", time.Now().UnixNano())
		defer cleanupTestUser(db, merchantEmail)

		registerRouter := gin.New()
		registerRouter.POST("/users/register", RegisterUser)

		registerReq := RegisterRequest{
			Email:    merchantEmail,
			Name:     "Login Test Merchant",
			Password: "password123",
			Role:     "merchant",
		}
		body, _ := json.Marshal(registerReq)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/users/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		registerRouter.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code, "Registration should succeed")

		loginRouter := gin.New()
		loginRouter.POST("/users/login", LoginUser)

		loginReq := LoginRequest{
			Email:    merchantEmail,
			Password: "password123",
		}
		loginBody, _ := json.Marshal(loginReq)

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("POST", "/users/login", bytes.NewBuffer(loginBody))
		req.Header.Set("Content-Type", "application/json")
		loginRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Login should return 200")

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		data := response["data"].(map[string]interface{})
		user := data["user"].(map[string]interface{})
		assert.Equal(t, "merchant", user["role"], "Logged in user role should be merchant")
	})

	t.Run("admin login returns admin role", func(t *testing.T) {
		adminEmail := fmt.Sprintf("test_login_admin_%d@example.com", time.Now().UnixNano())
		defer cleanupTestUser(db, adminEmail)

		registerRouter := gin.New()
		registerRouter.POST("/users/register", RegisterUser)

		registerReq := RegisterRequest{
			Email:    adminEmail,
			Name:     "Login Test Admin",
			Password: "password123",
			Role:     "admin",
		}
		body, _ := json.Marshal(registerReq)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/users/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		registerRouter.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code, "Registration should succeed")

		loginRouter := gin.New()
		loginRouter.POST("/users/login", LoginUser)

		loginReq := LoginRequest{
			Email:    adminEmail,
			Password: "password123",
		}
		loginBody, _ := json.Marshal(loginReq)

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("POST", "/users/login", bytes.NewBuffer(loginBody))
		req.Header.Set("Content-Type", "application/json")
		loginRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Login should return 200")

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		data := response["data"].(map[string]interface{})
		user := data["user"].(map[string]interface{})
		assert.Equal(t, "admin", user["role"], "Logged in user role should be admin")
	})
}
