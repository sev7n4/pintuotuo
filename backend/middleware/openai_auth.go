package middleware

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/tracing"
	"github.com/pintuotuo/backend/utils"
)

// APIKeyOrJWTAuthMiddleware allows either a platform user API key (ptd_* / ptt_*) or a JWT,
// for OpenAI-compatible endpoints used by external clients (OpenAI SDK, IDE plugins, etc.).
func APIKeyOrJWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string
		authHeader := c.GetHeader("Authorization")
		const bearerPrefix = "Bearer "
		if authHeader != "" {
			if !strings.HasPrefix(authHeader, bearerPrefix) {
				respondOpenAIAuthError(c, http.StatusUnauthorized, "Authorization must use Bearer scheme")
				c.Abort()
				return
			}
			tokenString = strings.TrimSpace(strings.TrimPrefix(authHeader, bearerPrefix))
		}
		if tokenString == "" {
			tokenString = strings.TrimSpace(c.GetHeader("x-api-key"))
		}
		if tokenString == "" {
			respondOpenAIAuthError(c, http.StatusUnauthorized, "Missing credentials (Authorization: Bearer or x-api-key)")
			c.Abort()
			return
		}

		if strings.HasPrefix(tokenString, "ptd_") || strings.HasPrefix(tokenString, "ptt_") {
			db := config.GetDB()
			if db == nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error": gin.H{
						"message": "Service unavailable",
						"type":    "api_error",
					},
				})
				c.Abort()
				return
			}

			keyHash := utils.HashUserAPIKey(tokenString)
			var userID int
			err := db.QueryRow(
				`SELECT user_id FROM api_keys WHERE key_hash = $1 AND status = 'active' LIMIT 1`,
				keyHash,
			).Scan(&userID)
			if err != nil {
				if err != sql.ErrNoRows {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": gin.H{
							"message": "Failed to validate API key",
							"type":    "api_error",
						},
					})
					c.Abort()
					return
				}
				respondOpenAIAuthError(c, http.StatusUnauthorized, "Incorrect API key provided")
				c.Abort()
				return
			}

			c.Set("user_id", userID)
			c.Next()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		})

		if err != nil {
			respondOpenAIAuthError(c, http.StatusUnauthorized, "Invalid authentication token")
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			userIDFloat, ok := claims["user_id"].(float64)
			if !ok {
				respondOpenAIAuthError(c, http.StatusUnauthorized, "Invalid token claims")
				c.Abort()
				return
			}
			userID := int(userIDFloat)
			c.Set("user_id", userID)
			tracing.SetEndUserID(c.Request.Context(), userID)

			if email, ok := claims["email"].(string); ok {
				c.Set("email", email)
			}
			if role, ok := claims["role"].(string); ok {
				c.Set("user_role", role)
			}

			c.Next()
			return
		}

		respondOpenAIAuthError(c, http.StatusUnauthorized, "Invalid authentication token")
		c.Abort()
	}
}

func respondOpenAIAuthError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"error": gin.H{
			"message": message,
			"type":    "invalid_request_error",
			"code":    "invalid_api_key",
		},
	})
}
