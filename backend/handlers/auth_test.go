package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	// 测试密码哈希函数
	password := "testpassword"
	hash1 := hashPassword(password)
	hash2 := hashPassword(password)

	// 相同密码应该产生相同的哈希
	assert.Equal(t, hash1, hash2)

	// 不同密码应该产生不同的哈希
	differentPassword := "differentpassword"
	differentHash := hashPassword(differentPassword)
	assert.NotEqual(t, hash1, differentHash)
}

func TestVerifyPassword(t *testing.T) {
	// 测试密码验证函数
	password := "testpassword"
	hash := hashPassword(password)

	// 正确的密码应该验证成功
	assert.True(t, verifyPassword(password, hash))

	// 错误的密码应该验证失败
	wrongPassword := "wrongpassword"
	assert.False(t, verifyPassword(wrongPassword, hash))
}

func TestGenerateToken(t *testing.T) {
	// 测试令牌生成函数
	userID := 1
	email := "test@example.com"
	token := generateToken(userID, email)

	// 生成的令牌应该不为空
	assert.NotEmpty(t, token)

	// 不同用户应该生成不同的令牌
	anotherUserID := 2
	anotherEmail := "another@example.com"
	anotherToken := generateToken(anotherUserID, anotherEmail)
	assert.NotEqual(t, token, anotherToken)
}

func TestGenerateResetToken(t *testing.T) {
	// 测试重置令牌生成函数
	userID := 1
	token := generateResetToken(userID)

	// 生成的令牌应该不为空
	assert.NotEmpty(t, token)

	// 不同用户应该生成不同的令牌
	anotherUserID := 2
	anotherToken := generateResetToken(anotherUserID)
	assert.NotEqual(t, token, anotherToken)
}

func TestGetEnv(t *testing.T) {
	// 测试环境变量获取函数
	// 测试默认值
	defaultValue := "default"
	result := getEnv("NON_EXISTENT_ENV", defaultValue)
	assert.Equal(t, defaultValue, result)
}

func TestLogoutUser(t *testing.T) {
	// 测试登出函数
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// 调用登出函数
	LogoutUser(c)

	// 检查响应状态码
	assert.Equal(t, http.StatusOK, w.Code)

	// 检查响应内容
	assert.Contains(t, w.Body.String(), "Logged out successfully")
}
