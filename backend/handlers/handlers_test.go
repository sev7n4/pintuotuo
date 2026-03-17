package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/cache"
	"github.com/pintuotuo/backend/config"
	"github.com/stretchr/testify/assert"
)

func init() {
	// Initialize test database
	_ = config.InitDB()
	// Initialize cache
	_ = cache.Init()
}

// TestUserRegistration tests user registration endpoint
func TestUserRegistration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	// Setup route
	router.POST("/users/register", RegisterUser)

	tests := []struct {
		name           string
		payload        map[string]string
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "Valid registration",
			payload: map[string]string{
				"email":    "test@example.com",
				"name":     "Test User",
				"password": "password123",
			},
			expectedStatus: http.StatusCreated,
			expectedError:  false,
		},
		{
			name: "Missing email",
			payload: map[string]string{
				"name":     "Test User",
				"password": "password123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "Short password",
			payload: map[string]string{
				"email":    "test@example.com",
				"name":     "Test User",
				"password": "pass",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/users/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError {
				var response map[string]interface{}
				_ = json.Unmarshal(w.Body.Bytes(), &response)
				assert.NotNil(t, response["error"])
			}
		})
	}
}

// TestLoginUser tests user login endpoint
func TestLoginUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	router.POST("/users/login", LoginUser)

	tests := []struct {
		name           string
		payload        map[string]string
		expectedStatus int
	}{
		{
			name: "Invalid credentials",
			payload: map[string]string{
				"email":    "nonexistent@example.com",
				"password": "password123",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Missing email",
			payload: map[string]string{
				"password": "password123",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/users/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestGetProductByID tests product retrieval
func TestGetProductByID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	router.GET("/products/:id", GetProductByID)

	tests := []struct {
		name           string
		productID      string
		expectedStatus int
	}{
		{
			name:           "Invalid product ID",
			productID:      "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Non-existent product",
			productID:      "99999",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/products/"+tt.productID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// Note: Benchmark tests below reference internal service methods
// and are kept as examples for future implementation
/*
// BenchmarkHashPassword benchmarks password hashing
func BenchmarkHashPassword(b *testing.B) {
	password := "test_password_123"
	for i := 0; i < b.N; i++ {
		hashPassword(password)
	}
}

// BenchmarkGenerateToken benchmarks token generation
func BenchmarkGenerateToken(b *testing.B) {
	userID := 1
	email := "test@example.com"
	for i := 0; i < b.N; i++ {
		generateToken(userID, email)
	}
}
*/
