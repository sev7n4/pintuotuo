package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pintuotuo/backend/errors"
)

type RateLimitData struct {
	Count     int
	ResetTime time.Time
}

var (
	rateLimitStore = map[string]RateLimitData{}
	rateLimitMu    sync.Mutex
)

func isProductionEnv() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("APP_ENV")), "production")
}

func CORSMiddleware() gin.HandlerFunc {
	allowList := parseAllowedOrigins()
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowOrigin := ""
		if origin != "" {
			if len(allowList) > 0 {
				if allowList[origin] {
					allowOrigin = origin
				}
			} else if !isProductionEnv() {
				// 非生产且未配置白名单时反射 Origin，避免本地与 CI（Vite 端口 ≠ API 端口）跨域被浏览器拦截
				allowOrigin = origin
			}
		}
		if allowOrigin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", allowOrigin)
			c.Writer.Header().Set("Vary", "Origin")
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}
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

func ValidateSecurityConfig() error {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("APP_ENV")), "production") {
		if strings.TrimSpace(os.Getenv("JWT_SECRET")) == "" {
			return fmt.Errorf("JWT_SECRET is required in production")
		}
		if len(parseAllowedOrigins()) == 0 {
			return fmt.Errorf("CORS_ALLOWED_ORIGINS is required in production")
		}
	}
	return nil
}

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
	body := gin.H{
		"error":   appErr.Message,
		"code":    appErr.Code,
		"message": appErr.Message,
	}
	if appErr.Details != nil {
		body["details"] = appErr.Details
	}
	c.JSON(appErr.Status, body)
}

func RateLimitMiddleware() gin.HandlerFunc {
	limit := mustParseInt(getEnv("RATE_LIMIT_PER_MINUTE", "120"), 120)
	window := time.Minute
	return func(c *gin.Context) {
		userKey := c.ClientIP()
		if userID, exists := c.Get("user_id"); exists {
			userKey = fmt.Sprintf("u:%v", userID)
		}
		key := fmt.Sprintf("%s:%s", userKey, c.FullPath())
		now := time.Now()

		rateLimitMu.Lock()
		entry, exists := rateLimitStore[key]
		if !exists || now.After(entry.ResetTime) {
			entry = RateLimitData{
				Count:     0,
				ResetTime: now.Add(window),
			}
		}
		entry.Count++
		rateLimitStore[key] = entry
		rateLimitMu.Unlock()

		remaining := limit - entry.Count
		if remaining < 0 {
			remaining = 0
		}
		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(entry.ResetTime.Unix(), 10))

		if entry.Count > limit {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate limit exceeded",
				"message": "too many requests",
				"code":    "RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

func parseAllowedOrigins() map[string]bool {
	origins := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if origins == "" {
		return map[string]bool{}
	}
	out := map[string]bool{}
	for _, origin := range strings.Split(origins, ",") {
		o := strings.TrimSpace(origin)
		if o != "" {
			out[o] = true
		}
	}
	return out
}

func mustParseInt(v string, fallback int) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
