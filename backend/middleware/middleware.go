package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/cache"
	"github.com/pintuotuo/backend/errors"
)

type RateLimitData struct {
	Count     int
	ResetTime time.Time
}

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

func ErrorHandlingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
				})
			}
		}()
		c.Next()
	}
}

func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		duration := time.Since(startTime)
		statusCode := c.Writer.Status()

		logMsg := fmt.Sprintf("[%s] %s %s | Status: %d | Duration: %v | IP: %s\n",
			time.Now().Format("2006-01-02 15:04:05"),
			method,
			path,
			statusCode,
			duration,
			c.ClientIP(),
		)
		fmt.Print(logMsg)
		os.Stdout.Sync()
	}
}

var jwtSecret = []byte(getEnv("JWT_SECRET", "pintuotuo-secret-key-dev"))

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			c.JSON(401, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, bearerPrefix)
		if tokenString == "" {
			c.JSON(401, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		})

		if err != nil {
			c.JSON(401, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			userIDFloat, ok := claims["user_id"].(float64)
			if !ok {
				c.JSON(401, gin.H{"error": "invalid token claims"})
				c.Abort()
				return
			}
			userID := int(userIDFloat)
			c.Set("user_id", userID)

			if email, ok := claims["email"].(string); ok {
				c.Set("email", email)
			}

			if role, ok := claims["role"].(string); ok {
				c.Set("user_role", role)
			}

			c.Next()
		} else {
			c.JSON(401, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}
	}
}

func getEnv(key, defaultValue string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return defaultValue
}

func RespondWithError(c *gin.Context, appErr *errors.AppError) {
	c.JSON(appErr.Status, gin.H{
		"error":   appErr.Message,
		"code":    appErr.Code,
		"message": appErr.Message,
	})
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
