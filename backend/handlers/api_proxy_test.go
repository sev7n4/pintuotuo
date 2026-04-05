package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/config"
	"github.com/stretchr/testify/assert"
)

func TestResolveOpenAICompatModel(t *testing.T) {
	p, m := resolveOpenAICompatModel("openai/gpt-4o")
	assert.Equal(t, "openai", p)
	assert.Equal(t, "gpt-4o", m)

	p, m = resolveOpenAICompatModel("zhipu/glm-4-flash")
	assert.Equal(t, "zhipu", p)
	assert.Equal(t, "glm-4-flash", m)

	p, m = resolveOpenAICompatModel("gpt-3.5-turbo")
	assert.Equal(t, "openai", p)
	assert.Equal(t, "gpt-3.5-turbo", m)

	p, m = resolveOpenAICompatModel("claude-3-sonnet-20240229")
	assert.Equal(t, "anthropic", p)
	assert.Equal(t, "claude-3-sonnet-20240229", m)

	p, m = resolveOpenAICompatModel("gemini-pro")
	assert.Equal(t, "google", p)
	assert.Equal(t, "gemini-pro", m)

	p, m = resolveOpenAICompatModel("glm-4-air")
	assert.Equal(t, "zhipu", p)
	assert.Equal(t, "glm-4-air", m)
}

func TestCalculateTokenCost(t *testing.T) {
	// 测试OpenAI GPT-4 Turbo (InputPrice: 10, OutputPrice: 30 per 1M tokens)
	cost := calculateTokenCost("openai", "gpt-4-turbo-preview", 1000, 1000)
	assert.InDelta(t, 0.04, cost, 0.0001)

	// 测试OpenAI GPT-4 (InputPrice: 30, OutputPrice: 60 per 1M tokens)
	cost = calculateTokenCost("openai", "gpt-4", 1000, 1000)
	assert.InDelta(t, 0.09, cost, 0.0001)

	// 测试OpenAI GPT-3.5 Turbo (InputPrice: 0.5, OutputPrice: 1.5 per 1M tokens)
	cost = calculateTokenCost("openai", "gpt-3.5-turbo", 1000, 1000)
	assert.InDelta(t, 0.002, cost, 0.0001)

	// 测试Anthropic Claude 3 Opus (InputPrice: 15, OutputPrice: 75 per 1M tokens)
	cost = calculateTokenCost("anthropic", "claude-3-opus-20240229", 1000, 1000)
	assert.InDelta(t, 0.09, cost, 0.0001)

	// 测试Anthropic Claude 3 Sonnet (InputPrice: 3, OutputPrice: 15 per 1M tokens)
	cost = calculateTokenCost("anthropic", "claude-3-sonnet-20240229", 1000, 1000)
	assert.InDelta(t, 0.018, cost, 0.0001)

	// 测试Anthropic Claude 3 Haiku (InputPrice: 0.25, OutputPrice: 1.25 per 1M tokens)
	cost = calculateTokenCost("anthropic", "claude-3-haiku-20240307", 1000, 1000)
	assert.InDelta(t, 0.0015, cost, 0.0001)

	// 测试Google AI Gemini Pro (InputPrice: 0.5, OutputPrice: 1.5 per 1M tokens)
	cost = calculateTokenCost("google", "gemini-pro", 1000, 1000)
	assert.InDelta(t, 0.002, cost, 0.0001)

	// 测试默认提供商 (InputPrice: 1, OutputPrice: 2 per 1M tokens)
	cost = calculateTokenCost("unknown", "unknown-model", 1000, 1000)
	assert.InDelta(t, 0.003, cost, 0.0001)
}

func TestGetAPIProviders(t *testing.T) {
	// 创建一个测试gin上下文
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// 调用GetAPIProviders函数
	GetAPIProviders(c)

	// 检查响应状态码
	assert.Equal(t, http.StatusOK, w.Code)

	// 解析响应体
	var providers []map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &providers)
	assert.NoError(t, err)

	// 检查提供商列表
	assert.Len(t, providers, 3)

	// 检查第一个提供商是否是OpenAI
	assert.Equal(t, "openai", providers[0]["name"])
	assert.Equal(t, "OpenAI", providers[0]["display_name"])

	// 检查第二个提供商是否是Anthropic
	assert.Equal(t, "anthropic", providers[1]["name"])
	assert.Equal(t, "Anthropic", providers[1]["display_name"])

	// 检查第三个提供商是否是Google
	assert.Equal(t, "google", providers[2]["name"])
	assert.Equal(t, "Google AI", providers[2]["display_name"])
}

func TestGetAPIUsageStats(t *testing.T) {
	// 跳过测试，因为需要数据库连接
	t.Skip("Skipping test due to database connection requirement")

	// 初始化数据库连接
	err := config.InitDB()
	assert.NoError(t, err)
	defer config.CloseDB()
	db := config.GetDB()

	// 创建测试用户
	var userID int
	err = db.QueryRow("INSERT INTO users (email, password_hash, name, role) VALUES ($1, $2, $3, $4) RETURNING id", "test@example.com", "password", "Test User", "user").Scan(&userID)
	assert.NoError(t, err)

	// 创建测试API使用日志
	_, err = db.Exec("INSERT INTO api_usage_logs (user_id, key_id, request_id, provider, model, method, path, status_code, latency_ms, input_tokens, output_tokens, cost) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)",
		userID, 1, "test-request-id", "openai", "gpt-3.5-turbo", "POST", "/v1/chat/completions", 200, 1000, 100, 50, 0.002)
	assert.NoError(t, err)

	// 创建一个测试gin上下文
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// 设置用户ID
	c.Set("user_id", userID)

	// 调用GetAPIUsageStats函数
	GetAPIUsageStats(c)

	// 检查响应状态码
	assert.Equal(t, http.StatusOK, w.Code)

	// 解析响应体
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// 检查响应结构
	assert.Contains(t, response, "stats")
	assert.Contains(t, response, "by_provider")
}

func TestProxyAPIRequest(t *testing.T) {
	// 跳过测试，因为需要数据库连接和外部API调用
	t.Skip("Skipping test due to database connection and external API dependencies")

	// 创建测试请求体
	requestBody := APIProxyRequest{
		Provider: "openai",
		Model:    "gpt-3.5-turbo",
		Messages: []ChatMessage{
			{Role: "user", Content: "Hello"},
		},
		Stream: false,
	}

	// 序列化请求体
	requestBodyBytes, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	// 创建一个测试gin上下文
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// 设置请求方法和路径
	c.Request, _ = http.NewRequest("POST", "/api/proxy", bytes.NewBuffer(requestBodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	// 设置用户ID
	c.Set("user_id", 1)

	// 调用ProxyAPIRequest函数
	// 注意：这个测试会失败，因为我们没有模拟HTTP客户端和加密解密函数
	// 实际测试中，我们需要使用mock库来模拟这些依赖
	// ProxyAPIRequest(c)

	// 检查响应状态码
	// assert.Equal(t, http.StatusOK, w.Code)
}
