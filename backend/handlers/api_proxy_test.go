package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestShouldUseSmartRouting(t *testing.T) {
	t.Setenv("SMART_ROUTING_ENABLE", "true")
	t.Setenv("SMART_ROUTING_GRAY_PERCENT", "0")
	assert.False(t, shouldUseSmartRouting(1, "req-1"))

	t.Setenv("SMART_ROUTING_GRAY_PERCENT", "100")
	assert.True(t, shouldUseSmartRouting(1, "req-1"))

	t.Setenv("SMART_ROUTING_ENABLE", "false")
	assert.False(t, shouldUseSmartRouting(1, "req-1"))
}

func TestBuildRetryPolicy(t *testing.T) {
	policy := buildRetryPolicy(buildStrategyRuntimeSnapshot(""))
	assert.NotNil(t, policy)
	assert.Equal(t, services.DefaultRetryPolicy.MaxRetries, policy.MaxRetries)

	known := buildRetryPolicy(buildStrategyRuntimeSnapshot("balanced"))
	assert.NotNil(t, known)
	assert.Equal(t, 3, known.MaxRetries)
	assert.Equal(t, 1000*time.Millisecond, known.InitialDelay)

	unknown := buildRetryPolicy(buildStrategyRuntimeSnapshot("not_exists"))
	assert.NotNil(t, unknown)
	assert.Equal(t, services.DefaultRetryPolicy.MaxRetries, unknown.MaxRetries)
}

func TestApplyCircuitBreakerConfig(t *testing.T) {
	apiKeyID := 10001
	applyCircuitBreakerConfig(apiKeyID, buildStrategyRuntimeSnapshot(""))
	cb := services.GetCircuitBreaker(apiKeyID)
	assert.NotNil(t, cb)
	assert.Equal(t, services.CircuitStateOpen, cb.GetState())

	applyCircuitBreakerConfig(apiKeyID, buildStrategyRuntimeSnapshot("balanced"))
	assert.NotNil(t, cb)
}

func TestBuildRoutingDecisionPayload(t *testing.T) {
	candidates := []byte(`[{"api_key_id":1,"score":0.9}]`)
	snapshot := buildStrategyRuntimeSnapshot("balanced")
	payload := buildRoutingDecisionPayload(candidates, snapshot, "")
	assert.NotEmpty(t, payload)

	var parsed map[string]any
	err := json.Unmarshal(payload, &parsed)
	assert.NoError(t, err)
	assert.Contains(t, parsed, "candidates")
	assert.Contains(t, parsed, "strategy_runtime")
	assert.NotContains(t, parsed, "effective_policy_source")

	withSrc := buildRoutingDecisionPayload(candidates, snapshot, policySourceEnv)
	var parsed2 map[string]any
	err = json.Unmarshal(withSrc, &parsed2)
	assert.NoError(t, err)
	assert.Equal(t, policySourceEnv, parsed2["effective_policy_source"])
}

func TestParseRoutingDecisionPayload(t *testing.T) {
	snapshot := buildStrategyRuntimeSnapshot("balanced")
	raw := buildRoutingDecisionPayload([]byte(`[{"api_key_id":1}]`), snapshot, "db")
	parsed, err := parseRoutingDecisionPayload(raw)
	assert.NoError(t, err)
	assert.NotEmpty(t, parsed.Candidates)
	assert.Equal(t, "balanced", parsed.StrategyRuntime.StrategyCode)
	assert.Equal(t, "db", parsed.EffectivePolicySource)

	legacyRaw := json.RawMessage(`[{"api_key_id":2}]`)
	_, legacyErr := parseRoutingDecisionPayload(legacyRaw)
	assert.Error(t, legacyErr)
}

func TestRoutingStrategyWithSource(t *testing.T) {
	t.Setenv("SMART_ROUTING_STRATEGY", "cost_first")
	code, src := routingStrategyWithSource()
	assert.Equal(t, "cost_first", code)
	assert.Equal(t, policySourceEnv, src)

	_ = os.Unsetenv("SMART_ROUTING_STRATEGY")
	code, src = routingStrategyWithSource()
	assert.NotEmpty(t, code)
	assert.Contains(t, []string{policySourceDB, policySourceDefault}, src)
}

func TestSummarizeRoutingCandidatesForTrace(t *testing.T) {
	raw := json.RawMessage(`[{"APIKeyID":1,"Provider":"openai","Model":"gpt-4","Score":0.5},{"api_key_id":2,"provider":"anthropic","score":0.9}]`)
	n, top := summarizeRoutingCandidatesForTrace(raw)
	assert.Equal(t, 2, n)
	require.NotNil(t, top)
	assert.Equal(t, 2, top.APIKeyID)
	assert.Equal(t, "anthropic", top.Provider)
	assert.InDelta(t, 0.9, top.Score, 0.0001)
}

func TestNormalizeEffectivePolicySource(t *testing.T) {
	assert.Equal(t, policySourceEnv, normalizeEffectivePolicySource(policySourceEnv))
	assert.Equal(t, policySourceDB, normalizeEffectivePolicySource(policySourceDB))
	assert.Equal(t, policySourceDefault, normalizeEffectivePolicySource(policySourceDefault))
	assert.Equal(t, policySourceDefault, normalizeEffectivePolicySource(""))
	assert.Equal(t, policySourceDefault, normalizeEffectivePolicySource("legacy"))
}
