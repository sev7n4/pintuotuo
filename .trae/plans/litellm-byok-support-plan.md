# Litellm BYOK支持改进计划

## 1. 需求背景

### 1.1 当前问题

在三层路由架构中，ExecutionLayer作为统一出站入口，当网关模式设置为`litellm`时，存在以下问题：

1. **API Key未正确传递**：当前实现只使用`LITELLM_MASTER_KEY`认证Litellm代理，但未将merchant的实际API key传递给Litellm
2. **BYOK模式不支持**：Litellm的BYOK（Bring Your Own Key）机制未被正确利用
3. **多租户场景受限**：无法实现每个merchant使用自己的API key调用provider

### 1.2 目标

实现Litellm BYOK模式支持，确保：
- merchant的实际API key能够正确传递给Litellm
- Litellm能够使用传递的API key调用上游provider
- 保持向后兼容性，不影响现有direct模式

## 2. 技术方案设计

### 2.1 架构设计

```
用户请求 → API Proxy → ExecutionLayer.Execute()
                         ↓
                    determineGatewayMode() → "litellm"
                         ↓
                    resolveEndpoint() → "http://litellm:4000/v1"
                         ↓
                    resolveAuthToken() → LITELLM_MASTER_KEY (认证Litellm)
                         ↓
                    buildHTTPRequest() → 添加 x-api-key header (传递merchant API key)
                         ↓
                    POST http://litellm:4000/v1/chat/completions
                    Headers:
                      Authorization: Bearer {LITELLM_MASTER_KEY}
                      x-api-key: {merchant的实际API key}
```

### 2.2 代码修改点

#### 2.2.1 后端修改

**文件**: `backend/services/execution_engine.go`

修改`buildHTTPRequest`方法，在Litellm模式下添加provider API key header：

```go
func (e *ExecutionEngine) buildHTTPRequest(ctx context.Context, input *ExecutionInput) (*http.Request, error) {
    // ... 现有代码 ...
    
    req.Header.Set("Content-Type", "application/json")
    
    switch input.RequestFormat {
    case modelProviderAnthropic:
        req.Header.Set("x-api-key", input.APIKey)
        req.Header.Set("anthropic-version", "2023-06-01")
    default:
        req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", input.APIKey))
    }
    
    // 新增: Litellm BYOK支持
    if input.GatewayMode == GatewayModeLitellm && input.OriginalAPIKey != "" {
        req.Header.Set("x-api-key", input.OriginalAPIKey)
    }
    
    // ... 现有代码 ...
}
```

#### 2.2.2 Litellm配置修改

**文件**: 部署环境配置

需要确保Litellm配置启用BYOK：

```yaml
general_settings:
  forward_client_headers_to_llm_api: true
  forward_llm_provider_auth_headers: true
```

### 2.3 Header传递规则

| Header | 用途 | 来源 |
|--------|------|------|
| `Authorization` | Litellm代理认证 | `LITELLM_MASTER_KEY` |
| `x-api-key` | Provider API key (BYOK) | merchant的实际API key |
| `Content-Type` | 请求内容类型 | `application/json` |

### 2.4 兼容性考虑

1. **Direct模式**：不受影响，继续使用原有逻辑
2. **Proxy模式**：不受影响，继续使用原有逻辑
3. **Litellm模式**：
   - 如果`LITELLM_MASTER_KEY`未配置，fallback到使用merchant API key
   - 如果merchant API key为空，返回错误

## 3. 实施步骤

### Phase 1: 代码修改 (预计1小时)

| 步骤 | 任务 | 文件 | 说明 |
|------|------|------|------|
| 1.1 | 修改buildHTTPRequest | `backend/services/execution_engine.go` | 添加Litellm BYOK header支持 |
| 1.2 | 添加单元测试 | `backend/services/execution_engine_test.go` | 测试Litellm模式下的header传递 |
| 1.3 | 运行测试验证 | - | 确保现有测试通过 |

### Phase 2: 部署配置 (预计30分钟)

| 步骤 | 任务 | 说明 |
|------|------|------|
| 2.1 | 更新Litellm配置 | 启用`forward_llm_provider_auth_headers` |
| 2.2 | 重启Litellm服务 | 应用新配置 |
| 2.3 | 重启后端服务 | 应用代码修改 |

### Phase 3: 集成验证 (预计30分钟)

| 步骤 | 任务 | 验证内容 |
|------|------|----------|
| 3.1 | API调用测试 | 使用测试账号调用step-1-8k模型 |
| 3.2 | Header传递验证 | 确认x-api-key正确传递 |
| 3.3 | 日志检查 | 确认请求成功到达provider |

## 4. 验证方案

### 4.1 单元测试

```go
func TestExecutionEngine_LitellmBYOK(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 验证Authorization header是master key
        auth := r.Header.Get("Authorization")
        assert.Equal(t, "Bearer sk-master-key", auth)
        
        // 验证x-api-key是原始API key
        apiKey := r.Header.Get("x-api-key")
        assert.Equal(t, "sk-original-key", apiKey)
        
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
    }))
    defer server.Close()
    
    engine := NewExecutionEngine()
    input := &ExecutionInput{
        Provider:       "openai",
        Model:          "gpt-4",
        APIKey:         "sk-master-key",
        OriginalAPIKey: "sk-original-key",
        EndpointURL:    server.URL,
        RequestFormat:  "openai",
        GatewayMode:    GatewayModeLitellm,
        Messages:       []Message{{Role: "user", Content: "Hello"}},
    }
    
    result, err := engine.Execute(context.Background(), input)
    require.NoError(t, err)
    assert.True(t, result.Success)
}
```

### 4.2 集成测试

```bash
# 1. 登录获取token
curl -X POST "http://119.29.173.89:8080/api/v1/users/login" \
  -H "Content-Type: application/json" \
  -d '{"email": "user100@163.com", "password": "111111"}'

# 2. 调用API
curl -X POST "http://119.29.173.89:8080/api/v1/openai/v1/chat/completions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {token}" \
  -d '{
    "model": "stepfun/step-1-8k",
    "messages": [{"role": "user", "content": "你好"}],
    "max_tokens": 50
  }'

# 3. 检查Litellm日志确认header传递
docker logs litellm --tail 50 | grep "x-api-key"
```

## 5. 风险评估

| 风险 | 等级 | 缓解措施 |
|------|------|----------|
| 现有功能回归 | 低 | 完整的单元测试覆盖 |
| Litellm配置不兼容 | 低 | 检查Litellm版本 >= 1.82 |
| Header冲突 | 低 | x-api-key header优先级明确 |

## 6. 回滚方案

如果出现问题，可以快速回滚：

1. **代码回滚**：移除`buildHTTPRequest`中的Litellm BYOK代码
2. **配置回滚**：禁用`forward_llm_provider_auth_headers`
3. **服务重启**：重启后端和Litellm服务

## 7. 后续优化

1. **监控告警**：添加Litellm BYOK调用成功率监控
2. **日志增强**：记录API key传递过程的关键日志
3. **文档更新**：更新架构文档，说明BYOK机制
