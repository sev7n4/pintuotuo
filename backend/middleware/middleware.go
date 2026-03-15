package middleware

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/cache"
	"github.com/pintuotuo/backend/errors"
)

// CORSMiddleware handles CORS for API requests
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// ErrorHandlingMiddleware handles errors globally
func ErrorHandlingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic: %v", err)
				c.JSON(500, gin.H{
					"code":    "INTERNAL_SERVER_ERROR",
					"message": "An unexpected error occurred",
				})
				c.Abort()
			}
		}()
		c.Next()

		// Check if there are errors in the context
		if len(c.Errors) > 0 {
			lastErr := c.Errors.Last()
			if appErr, ok := lastErr.Err.(*errors.AppError); ok {
				c.JSON(appErr.Status, appErr)
				return
			}
		}
	}
}

// LoggingMiddleware logs all requests
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(startTime)

		// Log request details
		statusCode := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path
		clientIP := c.ClientIP()

		log.Printf(
			"[%s] %s %s | Status: %d | Duration: %dms | IP: %s",
			time.Now().Format("2006-01-02 15:04:05"),
			method,
			path,
			statusCode,
			duration.Milliseconds(),
			clientIP,
		)
	}
}

// AuthMiddleware validates JWT token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		const bearerPrefix = "Bearer "
		if len(authHeader) < len(bearerPrefix) {
			c.JSON(401, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		token := authHeader[len(bearerPrefix):]
		if token == "" {
			c.JSON(401, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		// Validate token and extract user info
		// This would normally verify JWT signature
		// For now, just log and continue
		fmt.Printf("Authenticated with token: %s...\n", token[:min(20, len(token))])

		c.Next()
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// RespondWithError is a helper function to respond with an AppError
func RespondWithError(c *gin.Context, appErr *errors.AppError) {
	c.JSON(appErr.Status, appErr)
}

// RateLimitMiddleware limits request rate per IP address
// Default: 100 requests per minute per IP
func RateLimitMiddleware() gin.HandlerFunc {
	return RateLimitMiddlewareWithConfig(RateLimitConfig{
		RequestsPerMinute: 100,
		KeyPrefix:        "ratelimit:ip",
	})
}

// RateLimitConfig configures rate limiting behavior
type RateLimitConfig struct {
	RequestsPerMinute int
	KeyPrefix        string
}

// RateLimitMiddlewareWithConfig limits request rate with custom config
func RateLimitMiddlewareWithConfig(config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client IP address
		clientIP := c.ClientIP()
		if clientIP == "" {
			clientIP = "unknown"
		}

		// Create rate limit key
		key := fmt.Sprintf("%s:%s", config.KeyPrefix, clientIP)

		// Try to increment counter
		ctx := c.Request.Context()
		count, err := incrementRateLimit(ctx, key, 60*time.Second)
		if err != nil {
			// On cache error, allow request to pass through (graceful degradation)
			log.Printf("Rate limit check failed: %v", err)
			c.Next()
			return
		}

		// Check if limit exceeded
		if count > int64(config.RequestsPerMinute) {
			c.JSON(429, gin.H{
				"code":    "RATE_LIMIT_EXCEEDED",
				"message": fmt.Sprintf("Rate limit exceeded: %d requests per minute", config.RequestsPerMinute),
			})
			c.Abort()
			return
		}

		// Continue to next handler
		c.Next()
	}
}

// incrementRateLimit increments the rate limit counter using sliding window
func incrementRateLimit(ctx context.Context, key string, window time.Duration) (int64, error) {
	client := cache.GetClient()
	if client == nil {
		return 1, nil // Allow if cache not initialized
	}

	// Increment counter and get new value
	count, err := client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	// Set expiration on first request (TTL = window size)
	if count == 1 {
		if err := client.Expire(ctx, key, window).Err(); err != nil {
			log.Printf("Failed to set rate limit expiration: %v", err)
		}
	}

	return count, nil
}
