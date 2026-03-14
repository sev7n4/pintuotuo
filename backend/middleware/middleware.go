package middleware

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
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
					"error": "internal server error",
					"message": "An unexpected error occurred",
				})
				c.Abort()
			}
		}()
		c.Next()
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

// AuthMiddleware validates JWT token (to be implemented)
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(401, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		// Validate token (placeholder)
		// In real implementation, validate JWT token
		userID := c.GetInt("user_id")
		fmt.Printf("Authenticated user: %d\n", userID)

		c.Next()
	}
}

// RateLimitMiddleware limits request rate (to be implemented)
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implement rate limiting using Redis or in-memory store
		// For now, just pass through
		c.Next()
	}
}
