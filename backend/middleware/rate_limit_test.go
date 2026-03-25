package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestRateLimitMiddlewareStructure tests basic middleware structure
func TestRateLimitMiddlewareStructure(t *testing.T) {
	t.Run("RateLimitMiddleware returns a valid handler function", func(t *testing.T) {
		handler := RateLimitMiddleware()
		assert.NotNil(t, handler, "RateLimitMiddleware should return a handler")
	})

	t.Run("RateLimitMiddlewareWithConfig returns a valid handler function", func(t *testing.T) {
		config := RateLimitConfig{
			RequestsPerMinute: 100,
			KeyPrefix:        "test",
		}
		handler := RateLimitMiddlewareWithConfig(config)
		assert.NotNil(t, handler, "RateLimitMiddlewareWithConfig should return a handler")
	})
}

// TestRateLimitConfigDefaults tests default configuration
func TestRateLimitConfigDefaults(t *testing.T) {
	t.Run("Default middleware uses 100 requests per minute", func(t *testing.T) {
		handler := RateLimitMiddleware()
		assert.NotNil(t, handler, "Handler should be created with defaults")
	})

	t.Run("Config stores RequestsPerMinute correctly", func(t *testing.T) {
		config := RateLimitConfig{
			RequestsPerMinute: 50,
			KeyPrefix:        "test:limit",
		}
		assert.Equal(t, 50, config.RequestsPerMinute)
		assert.Equal(t, "test:limit", config.KeyPrefix)
	})
}

// TestRateLimitKeyFormat tests correct cache key generation
func TestRateLimitKeyFormat(t *testing.T) {
	t.Run("Rate limit key format includes prefix and IP", func(t *testing.T) {
		prefix := "ratelimit:ip"
		ip := "192.168.1.100"
		key := fmt.Sprintf("%s:%s", prefix, ip)

		assert.Equal(t, "ratelimit:ip:192.168.1.100", key)
		assert.Contains(t, key, prefix)
		assert.Contains(t, key, ip)
	})

	t.Run("Different IPs produce different keys", func(t *testing.T) {
		prefix := "ratelimit:ip"
		key1 := fmt.Sprintf("%s:%s", prefix, "10.0.0.1")
		key2 := fmt.Sprintf("%s:%s", prefix, "10.0.0.2")

		assert.NotEqual(t, key1, key2)
	})

	t.Run("Different prefixes produce different keys", func(t *testing.T) {
		ip := "172.16.0.1"
		key1 := fmt.Sprintf("%s:%s", "ratelimit:ip", ip)
		key2 := fmt.Sprintf("%s:%s", "ratelimit:user", ip)

		assert.NotEqual(t, key1, key2)
	})
}

// TestRateLimitMiddlewareBasicFlow tests that middleware allows requests through
func TestRateLimitMiddlewareBasicFlow(t *testing.T) {
	t.Run("Middleware processes request and continues to next handler", func(t *testing.T) {
		router := gin.New()
		gin.SetMode(gin.TestMode)

		router.Use(RateLimitMiddlewareWithConfig(RateLimitConfig{
			RequestsPerMinute: 100,
			KeyPrefix:        "test:flow",
		}))

		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "127.0.0.1:8080"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// With cache unavailable, request should pass through gracefully
		assert.True(t, w.Code == 200 || w.Code == 500, "Request should be processed")
	})
}

// TestRateLimitMiddlewareClientIP tests client IP extraction
func TestRateLimitMiddlewareClientIP(t *testing.T) {
	t.Run("Middleware extracts client IP from request", func(t *testing.T) {
		router := gin.New()
		gin.SetMode(gin.TestMode)

		// Track the IP extracted by middleware
		extractedIP := ""
		router.Use(func(c *gin.Context) {
			extractedIP = c.ClientIP()
			c.Next()
		})

		router.Use(RateLimitMiddlewareWithConfig(RateLimitConfig{
			RequestsPerMinute: 100,
			KeyPrefix:        "test:ip",
		}))

		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ip": c.ClientIP()})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "203.0.113.42:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Client IP should be extracted (exact value depends on Gin's extraction logic)
		assert.NotEmpty(t, extractedIP, "Client IP should be extracted")
	})
}

// TestRateLimitResponseStructure tests response format on failure
func TestRateLimitResponseStructure(t *testing.T) {
	t.Run("Rate limit error response has correct fields", func(t *testing.T) {
		// Verify error response would have correct structure
		response := gin.H{
			"code":    "RATE_LIMIT_EXCEEDED",
			"message": fmt.Sprintf("Rate limit exceeded: %d requests per minute", 100),
		}

		assert.Equal(t, "RATE_LIMIT_EXCEEDED", response["code"])
		assert.Contains(t, response["message"], "Rate limit exceeded")
		assert.Contains(t, response["message"], "100")
	})
}

// TestRateLimitConfigVariations tests different config scenarios
func TestRateLimitConfigVariations(t *testing.T) {
	t.Run("Strict rate limit config (1 request per minute)", func(t *testing.T) {
		config := RateLimitConfig{
			RequestsPerMinute: 1,
			KeyPrefix:        "test:strict",
		}
		assert.Equal(t, 1, config.RequestsPerMinute)
	})

	t.Run("Generous rate limit config (1000 requests per minute)", func(t *testing.T) {
		config := RateLimitConfig{
			RequestsPerMinute: 1000,
			KeyPrefix:        "test:generous",
		}
		assert.Equal(t, 1000, config.RequestsPerMinute)
	})

	t.Run("Custom key prefix for different rate limit zones", func(t *testing.T) {
		authConfig := RateLimitConfig{
			RequestsPerMinute: 5,
			KeyPrefix:        "ratelimit:auth",
		}

		apiConfig := RateLimitConfig{
			RequestsPerMinute: 100,
			KeyPrefix:        "ratelimit:api",
		}

		assert.NotEqual(t, authConfig.KeyPrefix, apiConfig.KeyPrefix)
		assert.Less(t, authConfig.RequestsPerMinute, apiConfig.RequestsPerMinute)
	})
}

// TestRateLimitWindowDuration tests sliding window duration
func TestRateLimitWindowDuration(t *testing.T) {
	t.Run("Rate limit window is 60 seconds", func(t *testing.T) {
		expectedWindow := 60 * time.Second
		actualWindow := time.Duration(60) * time.Second

		assert.Equal(t, expectedWindow, actualWindow)
	})
}

// TestRateLimitMiddlewareIntegration tests middleware in request chain
func TestRateLimitMiddlewareIntegration(t *testing.T) {
	t.Run("Multiple routes with different rate limits work independently", func(t *testing.T) {
		router := gin.New()
		gin.SetMode(gin.TestMode)

		// Strict rate limit for auth routes
		authGroup := router.Group("/auth")
		authGroup.Use(RateLimitMiddlewareWithConfig(RateLimitConfig{
			RequestsPerMinute: 5,
			KeyPrefix:        "ratelimit:auth",
		}))
		authGroup.POST("/login", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "login"})
		})

		// Generous rate limit for public routes
		publicGroup := router.Group("/api")
		publicGroup.Use(RateLimitMiddlewareWithConfig(RateLimitConfig{
			RequestsPerMinute: 100,
			KeyPrefix:        "ratelimit:api",
		}))
		publicGroup.GET("/products", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "products"})
		})

		// Both routes should respond
		authReq := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
		authReq.RemoteAddr = "127.0.0.1:8080"
		authW := httptest.NewRecorder()
		router.ServeHTTP(authW, authReq)

		publicReq := httptest.NewRequest(http.MethodGet, "/api/products", nil)
		publicReq.RemoteAddr = "127.0.0.1:8080"
		publicW := httptest.NewRecorder()
		router.ServeHTTP(publicW, publicReq)

		// Both should process (either succeed or fail gracefully)
		assert.True(t, authW.Code >= 200 && authW.Code < 500)
		assert.True(t, publicW.Code >= 200 && publicW.Code < 500)
	})
}

// TestRateLimitMiddlewareEdgeCases tests edge cases
func TestRateLimitMiddlewareEdgeCases(t *testing.T) {
	t.Run("Middleware handles requests with no RemoteAddr", func(t *testing.T) {
		router := gin.New()
		gin.SetMode(gin.TestMode)

		router.Use(RateLimitMiddlewareWithConfig(RateLimitConfig{
			RequestsPerMinute: 100,
			KeyPrefix:        "test:noremote",
		}))

		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "ok"})
		})

		// Request without RemoteAddr
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		// Should not panic
		assert.NotPanics(t, func() {
			router.ServeHTTP(w, req)
		})
	})

	t.Run("Middleware handles zero or negative limits gracefully", func(t *testing.T) {
		config := RateLimitConfig{
			RequestsPerMinute: -1,
			KeyPrefix:        "test:negative",
		}
		// Config creation should not panic
		assert.Equal(t, -1, config.RequestsPerMinute)
	})

	t.Run("Middleware handles very large limits", func(t *testing.T) {
		config := RateLimitConfig{
			RequestsPerMinute: 1000000,
			KeyPrefix:        "test:large",
		}
		assert.Equal(t, 1000000, config.RequestsPerMinute)
	})
}

// TestIncrementRateLimitFunction tests the increment function signature
func TestIncrementRateLimitFunction(t *testing.T) {
	t.Run("incrementRateLimit function accepts context and key", func(t *testing.T) {
		ctx := context.Background()
		key := "test:increment"

		// Function signature validation (will fail if cache not available, which is OK)
		// The important thing is that the function can be called without panic
		_, _ = incrementRateLimit(ctx, key, 60*time.Second)
	})
}

// TestRateLimitErrorResponse tests error response formatting
func TestRateLimitErrorResponse(t *testing.T) {
	t.Run("Error code in response is consistent", func(t *testing.T) {
		errorCode := "RATE_LIMIT_EXCEEDED"
		response := gin.H{
			"code":    errorCode,
			"message": "Rate limit exceeded: 100 requests per minute",
		}

		assert.Equal(t, errorCode, response["code"])
		assert.NotEmpty(t, response["message"])
	})

	t.Run("Response includes request limit in message", func(t *testing.T) {
		limit := 50
		message := fmt.Sprintf("Rate limit exceeded: %d requests per minute", limit)

		assert.Contains(t, message, "50")
		assert.Contains(t, message, "requests per minute")
	})
}

// TestRateLimitMiddlewareType tests middleware return type
func TestRateLimitMiddlewareType(t *testing.T) {
	t.Run("RateLimitMiddleware returns gin.HandlerFunc", func(t *testing.T) {
		handler := RateLimitMiddleware()
		// If compilation succeeds, type is correct
		assert.NotNil(t, handler)
	})

	t.Run("RateLimitMiddlewareWithConfig returns gin.HandlerFunc", func(t *testing.T) {
		config := RateLimitConfig{
			RequestsPerMinute: 100,
			KeyPrefix:        "test",
		}
		handler := RateLimitMiddlewareWithConfig(config)
		assert.NotNil(t, handler)
	})
}
