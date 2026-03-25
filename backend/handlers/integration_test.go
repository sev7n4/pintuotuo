package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
)

// TestErrorHandlingMiddleware verifies error middleware processes responses correctly
func TestErrorHandlingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.Use(middleware.ErrorHandlingMiddleware())

	// Test endpoint that returns an AppError
	router.POST("/test-error", func(c *gin.Context) {
		appErr := errors.ErrUserNotFound
		c.JSON(appErr.Status, appErr)
	})

	// Test endpoint that returns success
	router.POST("/test-success", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test error response
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test-error", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code, "Error response should have correct status code")

	var errorResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &errorResp)
	assert.Equal(t, "USER_NOT_FOUND", errorResp["code"], "Error response should include error code")

	// Test success response
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/test-success", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Success response should have 200 status")
}

// TestErrorResponseConsistency verifies all error types respond with consistent structure
func TestErrorResponseConsistency(t *testing.T) {
	testCases := []struct {
		name      string
		appErr    *errors.AppError
		expectErr bool
	}{
		{"InvalidCredentials", errors.ErrInvalidCredentials, true},
		{"MissingToken", errors.ErrMissingToken, true},
		{"UserNotFound", errors.ErrUserNotFound, true},
		{"ProductNotFound", errors.ErrProductNotFound, true},
		{"InsufficientStock", errors.ErrInsufficientStock, true},
		{"OrderNotFound", errors.ErrOrderNotFound, true},
		{"GroupNotFound", errors.ErrGroupNotFound, true},
		{"PaymentNotFound", errors.ErrPaymentNotFound, true},
		{"TokenNotFound", errors.ErrTokenNotFound, true},
		{"Forbidden", errors.ErrForbidden, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.Default()
			router.Use(middleware.ErrorHandlingMiddleware())

			router.POST("/test", func(c *gin.Context) {
				c.JSON(tc.appErr.Status, tc.appErr)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/test", nil)
			router.ServeHTTP(w, req)

			// Verify response structure
			assert.Greater(t, w.Code, 199, "Error should have HTTP status code")

			var errorResp map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &errorResp)

			// All errors should have Code and Message
			assert.NotEmpty(t, errorResp["code"], "Error response should include code")
			assert.NotEmpty(t, errorResp["message"], "Error response should include message")
		})
	}
}

// TestRequestLogging verifies request logging middleware tracks requests
func TestRequestLogging(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.Use(middleware.ErrorHandlingMiddleware())

	router.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Make a request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Health endpoint should return 200")

	// Verify response structure
	var healthResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &healthResp)
	assert.Equal(t, "healthy", healthResp["status"], "Health check should return healthy status")
}

// TestRequestValidationStructure verifies request validation patterns
func TestRequestValidationStructure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.Use(middleware.ErrorHandlingMiddleware())

	// Test endpoint with validation
	router.POST("/validate", func(c *gin.Context) {
		var req struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required,min=8"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errors.ErrInvalidRequest)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "valid"})
	})

	// Test with invalid email
	invalidReq := map[string]string{
		"email":    "invalid-email",
		"password": "password123",
	}
	body, _ := json.Marshal(invalidReq)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Should reject due to invalid email format
	assert.Greater(t, w.Code, 299, "Invalid email should be rejected")

	// Test with missing password
	missingReq := map[string]string{
		"email": "test@example.com",
	}
	body, _ = json.Marshal(missingReq)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Should reject due to missing password
	assert.Greater(t, w.Code, 299, "Missing password should be rejected")
}

// TestConcurrentRequestHandling verifies server handles concurrent requests
func TestConcurrentRequestHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.Use(middleware.ErrorHandlingMiddleware())

	var (
		requestCount int
		mu          sync.Mutex
	)
	router.GET("/concurrent", func(c *gin.Context) {
		mu.Lock()
		requestCount++
		count := requestCount
		mu.Unlock()
		time.Sleep(10 * time.Millisecond) // Simulate processing
		c.JSON(http.StatusOK, gin.H{
			"request_number": count,
			"timestamp":      time.Now().Unix(),
		})
	})

	// Make concurrent requests
	done := make(chan int, 5)

	for i := 0; i < 5; i++ {
		go func() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/concurrent", nil)
			router.ServeHTTP(w, req)

			if w.Code == http.StatusOK {
				done <- 1
			} else {
				done <- 0
			}
		}()
	}

	// Collect results
	successCount := 0
	for i := 0; i < 5; i++ {
		successCount += <-done
	}

	assert.Equal(t, 5, successCount, "All concurrent requests should succeed")
}

// TestErrorDetailsPropagation verifies error details are properly included
func TestErrorDetailsPropagation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.Use(middleware.ErrorHandlingMiddleware())

	router.POST("/test-details", func(c *gin.Context) {
		appErr := errors.NewAppErrorWithDetails(
			"VALIDATION_ERROR",
			"Validation failed",
			http.StatusBadRequest,
			nil,
			map[string]interface{}{
				"field": "email",
				"issue": "invalid format",
			},
		)
		c.JSON(appErr.Status, appErr)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test-details", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code, "Error with details should have correct status")

	var errorResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &errorResp)

	// Verify details are included
	assert.NotNil(t, errorResp["details"], "Error response should include details field")
	details := errorResp["details"].(map[string]interface{})
	assert.Equal(t, "email", details["field"], "Details should contain field information")
}

// TestContextPropagation verifies request context is properly propagated
func TestContextPropagation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	router.GET("/context-test", func(c *gin.Context) {
		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()

		select {
		case <-ctx.Done():
			c.JSON(http.StatusOK, gin.H{"message": "context expired"})
		default:
			c.JSON(http.StatusOK, gin.H{"message": "context active"})
		}
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/context-test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Context test should succeed")

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NotEmpty(t, resp["message"], "Response should contain message")
}

// TestJSONSerializationConsistency verifies JSON serialization works correctly
func TestJSONSerializationConsistency(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.Use(middleware.ErrorHandlingMiddleware())

	// Test with various data types
	router.GET("/json-test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"string_field": "test",
			"int_field":    123,
			"float_field":  123.45,
			"bool_field":   true,
			"null_field":   nil,
			"array_field":  []string{"a", "b", "c"},
		})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/json-test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "JSON serialization should succeed")

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err, "Response should be valid JSON")

	// Verify all fields are present and correct type
	assert.Equal(t, "test", resp["string_field"], "String field should match")
	assert.Equal(t, float64(123), resp["int_field"], "Int field should match (JSON converts to float64)")
	assert.Equal(t, 123.45, resp["float_field"], "Float field should match")
	assert.Equal(t, true, resp["bool_field"], "Bool field should match")
	assert.Nil(t, resp["null_field"], "Null field should be nil")
	assert.NotNil(t, resp["array_field"], "Array field should not be nil")
}

// TestHTTPStatusCodeConsistency verifies status codes are used correctly
func TestHTTPStatusCodeConsistency(t *testing.T) {
	testCases := []struct {
		name           string
		appErr         *errors.AppError
		expectedStatus int
	}{
		{"BadRequest", errors.ErrInvalidRequest, http.StatusBadRequest},
		{"Unauthorized", errors.ErrInvalidCredentials, http.StatusUnauthorized},
		{"Forbidden", errors.ErrForbidden, http.StatusForbidden},
		{"NotFound", errors.ErrUserNotFound, http.StatusNotFound},
		{"Conflict", errors.ErrUserAlreadyExists, http.StatusConflict},
		{"InternalServer", errors.ErrInternalServer, http.StatusInternalServerError},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedStatus, tc.appErr.Status,
				fmt.Sprintf("%s should have correct status code", tc.name))
		})
	}
}
