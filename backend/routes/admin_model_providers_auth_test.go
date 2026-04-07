package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pintuotuo/backend/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 与 middleware.AuthMiddleware / handlers.generateToken 默认开发密钥一致
const testJWTSecret = "pintuotuo-secret-key-dev"

func testAccessToken(t *testing.T, userID int, email, role string) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": float64(userID),
		"email":   email,
		"role":    role,
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	s, err := tok.SignedString([]byte(testJWTSecret))
	require.NoError(t, err)
	return s
}

// TestAdminModelProvidersHTTPAuth 验证 /api/v1/admin/model-providers/* 鉴权链：JWT → 管理员角色
func TestAdminModelProvidersHTTPAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	v1 := r.Group("/api/v1")
	RegisterAdminRoutes(v1)

	t.Run("GET /all 无 Authorization 返回 401", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/model-providers/all", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code, "body=%s", w.Body.String())
	})

	t.Run("GET /all Bearer 非法 JWT 返回 401", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/model-providers/all", nil)
		req.Header.Set("Authorization", "Bearer not-a-valid-jwt")
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("GET /all 普通用户 JWT 返回 403", func(t *testing.T) {
		token := testAccessToken(t, 1, "user@example.com", "user")
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/model-providers/all", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code, "管理员接口应对非 admin 角色拒绝: body=%s", w.Body.String())
	})

	t.Run("GET /all 商户角色 JWT 返回 403", func(t *testing.T) {
		token := testAccessToken(t, 2, "m@example.com", "merchant")
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/model-providers/all", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("PATCH /:id 无 Token 返回 401", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/model-providers/1", nil)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("PATCH /:id 普通用户返回 403", func(t *testing.T) {
		token := testAccessToken(t, 1, "user@example.com", "user")
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/model-providers/1", http.NoBody)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("POST /model-providers 无 Token 返回 401", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/model-providers", http.NoBody)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code, "body=%s", w.Body.String())
	})

	t.Run("POST /model-providers 普通用户返回 403", func(t *testing.T) {
		token := testAccessToken(t, 1, "user@example.com", "user")
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/model-providers", strings.NewReader(`{"code":"x","name":"y"}`))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code, "body=%s", w.Body.String())
	})

	t.Run("GET /all admin 角色通过鉴权（有库则 200，无库则 500）", func(t *testing.T) {
		token := testAccessToken(t, 99, "admin@example.com", "admin")
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/model-providers/all", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)
		if config.GetDB() == nil {
			assert.Equal(t, http.StatusInternalServerError, w.Code, "无 DB 时应为业务层错误，但不应 401/403: %s", w.Body.String())
		} else {
			assert.Equal(t, http.StatusOK, w.Code, w.Body.String())
		}
	})
}
