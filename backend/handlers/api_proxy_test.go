package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/models"
	"github.com/pintuotuo/backend/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateTokenCost(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// 无最近绑定价目版本的订单 → 走 live SPU（与历史期望一致）
	for range make([]struct{}, 8) {
		mock.ExpectQuery(`SELECT pricing_version_id FROM orders`).
			WithArgs(1).
			WillReturnError(sql.ErrNoRows)
	}

	cost := calculateTokenCost(db, 1, "openai", "gpt-4-turbo-preview", 1000, 1000)
	assert.InDelta(t, 0.04, cost, 0.0001)

	cost = calculateTokenCost(db, 1, "openai", "gpt-4", 1000, 1000)
	assert.InDelta(t, 0.09, cost, 0.0001)

	cost = calculateTokenCost(db, 1, "openai", "gpt-3.5-turbo", 1000, 1000)
	assert.InDelta(t, 0.002, cost, 0.0001)

	cost = calculateTokenCost(db, 1, "anthropic", "claude-3-opus-20240229", 1000, 1000)
	assert.InDelta(t, 0.09, cost, 0.0001)

	cost = calculateTokenCost(db, 1, "anthropic", "claude-3-sonnet-20240229", 1000, 1000)
	assert.InDelta(t, 0.018, cost, 0.0001)

	cost = calculateTokenCost(db, 1, "anthropic", "claude-3-haiku-20240307", 1000, 1000)
	assert.InDelta(t, 0.0015, cost, 0.0001)

	cost = calculateTokenCost(db, 1, "google", "gemini-pro", 1000, 1000)
	assert.InDelta(t, 0.002, cost, 0.0001)

	cost = calculateTokenCost(db, 1, "unknown", "unknown-model", 1000, 1000)
	assert.InDelta(t, 0.003, cost, 0.0001)

	require.NoError(t, mock.ExpectationsWereMet())
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

func TestSelectAPIKeyForRequest_NonMerchantWithAPIKeyID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	req := APIProxyRequest{
		Provider: "stepfun",
		Model:    "step-1-8k",
		APIKeyID: func() *int { v := 22; return &v }(),
	}

	rows := sqlmock.NewRows([]string{
		"id", "merchant_id", "provider", "api_key_encrypted", "api_secret_encrypted", "quota_limit", "quota_used", "status",
	}).AddRow(22, 4, "stepfun", "enc_key", "enc_secret", nil, 0.0, "active")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, merchant_id, provider, api_key_encrypted, api_secret_encrypted, quota_limit, quota_used, status
			 FROM merchant_api_keys
			 WHERE id = $1 AND provider = $2 AND status = 'active'
			   AND (verified_at IS NOT NULL OR verification_result = 'verified')
			   AND (quota_limit IS NULL OR quota_used < quota_limit) LIMIT 1`)).
		WithArgs(22, "stepfun").
		WillReturnRows(rows)

	var apiKey models.MerchantAPIKey
	err = selectAPIKeyForRequest(db, 13, 0, req, &apiKey)
	require.NoError(t, err)
	assert.Equal(t, 22, apiKey.ID)
	assert.Equal(t, 4, apiKey.MerchantID)
	assert.Equal(t, "stepfun", apiKey.Provider)
	assert.Nil(t, apiKey.QuotaLimit)
	assert.Equal(t, 0.0, apiKey.QuotaUsed)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSelectAPIKeyForRequest_MerchantWithAPIKeyIDBoundedByMerchant(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	req := APIProxyRequest{
		Provider: "stepfun",
		Model:    "step-1-8k",
		APIKeyID: func() *int { v := 22; return &v }(),
	}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, merchant_id, provider, api_key_encrypted, api_secret_encrypted, quota_limit, quota_used, status
			 FROM merchant_api_keys
			 WHERE id = $1 AND provider = $2 AND status = 'active'
			   AND (verified_at IS NOT NULL OR verification_result = 'verified')
			   AND (quota_limit IS NULL OR quota_used < quota_limit) AND merchant_id = $3 LIMIT 1`)).
		WithArgs(22, "stepfun", 99).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, merchant_id, provider, api_key_encrypted, api_secret_encrypted, quota_limit, quota_used, status
				 FROM merchant_api_keys
				 WHERE provider = $1 AND status = 'active'
				   AND merchant_id = $2
				   AND (verified_at IS NOT NULL OR verification_result = 'verified')
				   AND (quota_limit IS NULL OR quota_used < quota_limit)
				 ORDER BY COALESCE((quota_limit - quota_used)::double precision, 1e30::double precision) DESC
				 LIMIT 1`)).
		WithArgs("stepfun", 99).
		WillReturnError(sql.ErrNoRows)

	var apiKey models.MerchantAPIKey
	err = selectAPIKeyForRequest(db, 15, 99, req, &apiKey)
	require.ErrorIs(t, err, sql.ErrNoRows)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTrySelectAPIKeyWithSmartRouter_EmptyProviderSkipsInjection(t *testing.T) {
	pick := trySelectAPIKeyWithSmartRouter(APIProxyRequest{
		Provider: "",
		Model:    "gpt-4",
	}, "balanced")
	assert.Nil(t, pick.APIKeyID)
}
