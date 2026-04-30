# BYOK路由模式支持开发计划

## 问题背景

### 问题描述
轻量验证、深度验证、立即探测的功能没有根据路由模式（直连、auto、litellm、代理模式）不同均完整实现了对应的能力。

### 当前状态
- ✅ 数据库已有`route_mode`和`route_config`字段（迁移文件074已应用）
- ✅ ExecutionLayer中有完整的路由模式处理逻辑
- ✅ 已有完整的错误处理机制和错误码规范（provider_error_mapper.go）
- ❌ 验证和探测功能没有使用路由模式
- ❌ 只使用了`provider`和`endpoint_url`，完全忽略了`route_mode`字段

### 影响范围
- 后端：`admin_byok_routing.go`、`api_key_validator.go`、`health_checker.go`
- 前端：`AdminByokRouting.tsx`、`adminByokRouting.ts`
- 测试：需要新增单元测试和集成测试

---

## 开发计划

### Phase 1: 后端基础设施（优先级：高）

#### 1.1 扩展API Key模型
**文件**: `backend/models/models.go`

**任务**:
- ✅ 确认`MerchantAPIKey`模型已包含`RouteMode`和`RouteConfig`字段
- 添加辅助方法用于获取路由配置

**代码示例**:
```go
type MerchantAPIKey struct {
    // ... 现有字段 ...
    RouteMode   string                 `json:"route_mode"`
    RouteConfig map[string]interface{} `json:"route_config"`
}

func (k *MerchantAPIKey) GetEndpointForMode() string {
    // 根据route_mode返回对应的endpoint
}
```

#### 1.2 扩展验证服务接口
**文件**: `backend/services/api_key_validator.go`

**重要参考**: 参考`backend/services/execution_layer.go`的实现逻辑，特别是：
- `resolveEndpoint`方法（行231-310）
- `resolveAuthToken`方法（行312-328）
- `determineGatewayMode`方法（行330-340）

**任务**:
- 新增`ValidateAsyncWithRouteMode`方法
- 新增`performVerificationWithRouteMode`方法
- 实现不同路由模式的endpoint解析函数
- 实现不同路由模式的认证Token处理
- **重要**: 所有错误处理必须调用`MapProviderError`进行统一错误映射

**新增方法**:
```go
// 支持路由模式的验证入口
func (v *APIKeyValidator) ValidateAsyncWithRouteMode(
    apiKeyID int, 
    provider, encryptedKey, verificationType, routeMode string, 
    routeConfig map[string]interface{},
    region string  // 新增region参数
) error

// 根据路由模式执行验证
func (v *APIKeyValidator) performVerificationWithRouteMode(
    apiKeyID int, 
    provider, encryptedKey, verificationType, routeMode string, 
    routeConfig map[string]interface{}, 
    region string,  // 新增region参数
    retryCount int
)

// 不同路由模式的endpoint解析
func (v *APIKeyValidator) resolveDirectEndpoint(ctx context.Context, provider string, routeConfig map[string]interface{}, region string) (string, error)
func (v *APIKeyValidator) resolveLitellmEndpoint(ctx context.Context, routeConfig map[string]interface{}, region string) (string, error)
func (v *APIKeyValidator) resolveProxyEndpoint(ctx context.Context, routeConfig map[string]interface{}) (string, error)
func (v *APIKeyValidator) resolveAutoEndpoint(ctx context.Context, provider string, routeConfig map[string]interface{}, region string) (string, error)

// 不同路由模式的认证Token处理
func (v *APIKeyValidator) resolveAuthToken(routeMode string, originalAPIKey string) string
```

**错误处理规范**:
```go
// 所有路由模式的错误处理必须遵循统一规范
func (v *APIKeyValidator) performVerificationWithRouteMode(...) {
    // ... 验证逻辑 ...
    
    if err != nil {
        // 使用统一的错误映射函数
        errInfo := MapProviderError(statusCode, errorCode, errorMsg, headers, err, rawBody)
        
        // 记录详细的验证日志
        logger.LogError(ctx, "verification", "verification failed", err, map[string]interface{}{
            "api_key_id":     apiKeyID,
            "route_mode":     routeMode,
            "error_category": errInfo.Category,
            "error_code":     errInfo.ProviderCode,
            "retryable":      errInfo.Retryable,
        })
        
        // 根据Retryable判断是否重试
        if errInfo.Retryable && retryCount < maxRetries {
            v.performVerificationWithRouteMode(apiKeyID, provider, encryptedKey, 
                verificationType, routeMode, routeConfig, retryCount+1)
        }
    }
}
```

#### 1.3 扩展健康检查服务
**文件**: `backend/services/health_checker.go`

**任务**:
- 修改`resolveEndpoint`方法，支持路由模式
- 新增不同路由模式的endpoint解析方法
- **重要**: 所有错误处理必须遵循现有的`MapProviderError`规范

**修改方法**:
```go
func (s *HealthChecker) resolveEndpointWithRouteMode(ctx context.Context, apiKey *models.MerchantAPIKey) (string, error) {
    switch apiKey.RouteMode {
    case "direct":
        return s.resolveDirectEndpoint(ctx, apiKey)
    case "litellm":
        return s.resolveLitellmEndpoint(ctx, apiKey)
    case "proxy":
        return s.resolveProxyEndpoint(ctx, apiKey)
    case "auto":
        return s.resolveAutoEndpoint(ctx, apiKey)
    default:
        return s.resolveDirectEndpoint(ctx, apiKey)
    }
}
```

#### 1.4 修改BYOK路由管理API
**文件**: `backend/handlers/admin_byok_routing.go`

**任务**:
- 修改`LightVerifyBYOK`函数，获取并传递`route_mode`、`route_config`和`region`
- 修改`DeepVerifyBYOK`函数，获取并传递`route_mode`、`route_config`和`region`
- 修改`TriggerBYOKProbe`函数，获取并传递`route_mode`、`route_config`和`region`

**修改示例**:
```go
func LightVerifyBYOK(c *gin.Context) {
    // ... 权限检查和参数验证 ...
    
    var apiKey models.MerchantAPIKey
    err = db.QueryRow(
        `SELECT id, merchant_id, provider, api_key_encrypted, route_mode, route_config, region
         FROM merchant_api_keys
         WHERE id = $1 AND status = 'active'`,
        keyID,
    ).Scan(&apiKey.ID, &apiKey.MerchantID, &apiKey.Provider, &apiKey.APIKeyEncrypted, 
           &apiKey.RouteMode, &apiKey.RouteConfig, &apiKey.Region)
    
    validator := services.GetAPIKeyValidator()
    err = validator.ValidateAsyncWithRouteMode(
        apiKey.ID, apiKey.Provider, apiKey.APIKeyEncrypted, 
        "admin_light", apiKey.RouteMode, apiKey.RouteConfig, apiKey.Region
    )
}
```

---

### Phase 2: 路由模式实现（优先级：高）

#### 2.1 Direct模式实现
**优先级**: 最高（已有基础逻辑）

**任务**:
- 使用现有的`endpoint_url`或`provider`的`api_base_url`
- 直接访问上游API进行验证和探测
- **Region处理**: 根据region选择不同的endpoint（参考execution_layer.go:297-308）
- **Endpoints配置**: 支持从route_config.endpoints.direct.{region}获取URL
- **错误处理**: 使用`MapProviderError`进行错误映射

**实现要点**:
```go
func (v *APIKeyValidator) resolveDirectEndpoint(ctx context.Context, provider string, routeConfig map[string]interface{}, region string) (string, error) {
    // 1. 优先使用route_config中的endpoint_url
    if endpoint, ok := routeConfig["endpoint_url"].(string); ok && endpoint != "" {
        return endpoint, nil
    }
    
    // 2. 从route_config.endpoints.direct.{region}获取
    if endpoints, ok := routeConfig["endpoints"].(map[string]interface{}); ok {
        if directEndpoints, ok := endpoints["direct"].(map[string]interface{}); ok {
            if region == "" {
                region = "overseas"  // 默认海外
            }
            if url, ok := directEndpoints[region].(string); ok && url != "" {
                return url, nil
            }
        }
    }
    
    // 3. 否则从model_providers表获取
    return v.getProviderBaseURL(ctx, provider)
}
```

#### 2.2 LiteLLM模式实现
**优先级**: 高

**任务**:
- 从`route_config`或系统配置获取LiteLLM地址
- 通过LiteLLM网关进行验证和探测
- 需要处理LiteLLM特有的请求格式
- **Region处理**: 根据region选择不同的endpoint（参考execution_layer.go:262-279）
- **Endpoints配置**: 支持从route_config.endpoints.litellm.{region}获取URL
- **认证Token**: 使用LITELLM_MASTER_KEY环境变量（参考execution_layer.go:317-323）
- **错误处理**: 使用`MapProviderError`进行错误映射

**实现要点**:
```go
func (v *APIKeyValidator) resolveLitellmEndpoint(ctx context.Context, routeConfig map[string]interface{}, region string) (string, error) {
    // 1. 从route_config.endpoints.litellm.{region}获取
    if endpoints, ok := routeConfig["endpoints"].(map[string]interface{}); ok {
        if litellmEndpoints, ok := endpoints["litellm"].(map[string]interface{}); ok {
            if region == "" {
                region = "domestic"  // LiteLLM默认国内
            }
            if url, ok := litellmEndpoints[region].(string); ok && url != "" {
                return url, nil
            }
        }
    }
    
    // 2. 从route_config获取base_url
    if baseURL, ok := routeConfig["base_url"].(string); ok && baseURL != "" {
        return baseURL, nil
    }
    
    // 3. 从环境变量获取（使用正确的环境变量名）
    litellmURL := os.Getenv("LLM_GATEWAY_LITELLM_URL")
    if litellmURL != "" {
        return litellmURL + "/v1", nil
    }
    
    return "", fmt.Errorf("LiteLLM endpoint not configured")
}

// 认证Token处理
func (v *APIKeyValidator) resolveAuthToken(routeMode string, originalAPIKey string) string {
    switch routeMode {
    case "litellm":
        masterKey := os.Getenv("LITELLM_MASTER_KEY")
        if masterKey != "" {
            return masterKey  // LiteLLM模式使用master key
        }
        return originalAPIKey
    default:
        return originalAPIKey  // 其他模式使用原始API key
    }
}
```

#### 2.3 Proxy模式实现
**优先级**: 中

**任务**:
- 从`route_config`或系统配置获取代理地址
- 通过代理进行验证和探测
- 需要处理代理特有的请求头和认证
- **Endpoints配置**: 支持从route_config.endpoints.proxy.{type}获取URL（参考execution_layer.go:281-294）
- **错误处理**: 使用`MapProviderError`进行错误映射

**实现要点**:
```go
func (v *APIKeyValidator) resolveProxyEndpoint(ctx context.Context, routeConfig map[string]interface{}) (string, error) {
    // 1. 从route_config.endpoints.proxy获取
    if endpoints, ok := routeConfig["endpoints"].(map[string]interface{}); ok {
        if proxyEndpoints, ok := endpoints["proxy"].(map[string]interface{}); ok {
            // 优先使用gaap类型
            if gaapURL, ok := proxyEndpoints["gaap"].(string); ok && gaapURL != "" {
                return gaapURL, nil
            }
            // 否则使用第一个可用的
            for _, v := range proxyEndpoints {
                if url, ok := v.(string); ok && url != "" {
                    return url, nil
                }
            }
        }
    }
    
    // 2. 从route_config获取proxy_url
    if proxyURL, ok := routeConfig["proxy_url"].(string); ok && proxyURL != "" {
        return proxyURL, nil
    }
    
    return "", fmt.Errorf("Proxy endpoint not configured")
}
```

#### 2.4 Auto模式实现
**优先级**: 中

**任务**:
- 参考`ExecutionLayer`的`determineGatewayMode`逻辑（execution_layer.go:330-340）
- 根据系统配置自动选择最佳路由模式
- 实现降级策略：直连 → LiteLLM → 代理
- **错误处理**: 使用`MapProviderError`进行错误映射

**实现要点**:
```go
func (v *APIKeyValidator) resolveAutoEndpoint(ctx context.Context, provider string, routeConfig map[string]interface{}, region string) (string, error) {
    // 1. 检查系统配置，决定优先级（默认：direct -> litellm -> proxy）
    priority := []string{"direct", "litellm", "proxy"}
    
    // 2. 按优先级尝试
    for _, mode := range priority {
        var endpoint string
        var err error
        
        switch mode {
        case "direct":
            endpoint, err = v.resolveDirectEndpoint(ctx, provider, routeConfig, region)
        case "litellm":
            endpoint, err = v.resolveLitellmEndpoint(ctx, routeConfig, region)
        case "proxy":
            endpoint, err = v.resolveProxyEndpoint(ctx, routeConfig)
        }
        
        if err == nil && endpoint != "" {
            return endpoint, nil
        }
    }
    
    return "", fmt.Errorf("no available endpoint for auto mode")
}
```

---

### Phase 3: 前端展示优化（优先级：中）

#### 3.1 验证结果展示优化
**文件**: `frontend/src/pages/admin/AdminByokRouting.tsx`

**任务**:
- 在验证结果模态框中显示路由模式
- 显示使用的endpoint地址
- 根据路由模式显示不同的验证详情
- 显示错误分类和错误码

**实现要点**:
```typescript
<Descriptions.Item label="验证模式">
  <Tag color={getRouteModeColor(verificationResult.route_mode)}>
    {getRouteModeLabel(verificationResult.route_mode)}
  </Tag>
</Descriptions.Item>

<Descriptions.Item label="验证端点">
  <Text code>{verificationResult.endpoint_used}</Text>
</Descriptions.Item>

<Descriptions.Item label="错误分类">
  <Tag color={getErrorCategoryColor(verificationResult.error_category)}>
    {verificationResult.error_category}
  </Tag>
</Descriptions.Item>

<Descriptions.Item label="错误码">
  <Text code>{verificationResult.error_code}</Text>
</Descriptions.Item>
```

#### 3.2 服务层扩展
**文件**: `frontend/src/services/adminByokRouting.ts`

**任务**:
- 扩展`VerificationResult`接口，包含路由模式信息和错误分类
- 添加路由模式相关的辅助函数
- 添加错误分类相关的辅助函数

**实现要点**:
```typescript
export interface VerificationResult {
  // ... 现有字段 ...
  route_mode: 'direct' | 'litellm' | 'proxy' | 'auto';
  endpoint_used: string;
  error_category: string;
  error_code: string;
  retryable: boolean;
}

export function getRouteModeColor(mode: string): string {
  const colorMap: Record<string, string> = {
    'direct': 'green',
    'litellm': 'blue',
    'proxy': 'orange',
    'auto': 'purple',
  };
  return colorMap[mode] || 'default';
}

export function getRouteModeLabel(mode: string): string {
  const labelMap: Record<string, string> = {
    'direct': '直连',
    'litellm': 'LiteLLM',
    'proxy': '代理',
    'auto': '自动',
  };
  return labelMap[mode] || mode;
}

export function getErrorCategoryColor(category: string): string {
  const colorMap: Record<string, string> = {
    'AUTH_INVALID_KEY': 'red',
    'AUTH_PERMISSION_DENIED': 'red',
    'QUOTA_INSUFFICIENT': 'orange',
    'RATE_LIMITED': 'orange',
    'MODEL_NOT_FOUND': 'yellow',
    'SERVICE_UNAVAILABLE': 'red',
    'NETWORK_TIMEOUT': 'red',
    'UNKNOWN': 'default',
  };
  return colorMap[category] || 'default';
}
```

---

### Phase 4: 测试与验证（优先级：高）

#### 4.1 单元测试
**文件**: `backend/services/api_key_validator_test.go`

**任务**:
- 为每种路由模式编写单元测试
- 测试endpoint解析逻辑
- **测试错误处理逻辑，验证错误分类的正确性**
- **测试重试机制**

**测试用例**:
```go
func TestResolveDirectEndpoint(t *testing.T) {
    // 测试直连模式endpoint解析
}

func TestResolveLitellmEndpoint(t *testing.T) {
    // 测试LiteLLM模式endpoint解析
}

func TestResolveProxyEndpoint(t *testing.T) {
    // 测试代理模式endpoint解析
}

func TestResolveAutoEndpoint(t *testing.T) {
    // 测试自动模式endpoint解析和降级策略
}

func TestVerificationErrorMapping(t *testing.T) {
    // 测试错误映射是否正确
    // 验证不同路由模式下的错误分类是否符合规范
    result := performVerificationWithRouteMode(...)
    
    // 验证错误分类
    if result.ErrorCategory != errorCategoryAuthInvalidKey {
        t.Errorf("expected category %s, got %s", errorCategoryAuthInvalidKey, result.ErrorCategory)
    }
    
    // 验证重试标志
    if result.Retryable != expectedRetryable {
        t.Errorf("expected retryable %v, got %v", expectedRetryable, result.Retryable)
    }
}
```

#### 4.2 集成测试
**文件**: `backend/handlers/admin_byok_routing_test.go`

**任务**:
- 测试完整的验证流程
- 测试不同路由模式下的验证结果
- 测试健康检查功能
- **验证错误信息持久化到api_key_health_history表**

**测试用例**:
```go
func TestLightVerifyBYOK_WithDirectMode(t *testing.T) {
    // 测试直连模式轻量验证
}

func TestDeepVerifyBYOK_WithLitellmMode(t *testing.T) {
    // 测试LiteLLM模式深度验证
}

func TestTriggerBYOKProbe_WithProxyMode(t *testing.T) {
    // 测试代理模式立即探测
}

func TestVerificationErrorPersistence(t *testing.T) {
    // 测试错误信息持久化
    // 验证error_category、provider_error_code等字段是否正确保存
}
```

#### 4.3 E2E测试
**文件**: `tests/e2e/byok_routing_test.go`

**任务**:
- 测试完整的用户流程
- 测试不同路由模式下的端到端功能
- 验证前端展示正确性
- **验证错误信息的展示**

---

### Phase 5: 配置与文档（优先级：中）

#### 5.1 环境变量配置
**重要**: 使用与ExecutionLayer一致的环境变量命名

**环境变量列表**:
```bash
# LiteLLM配置
LLM_GATEWAY_LITELLM_URL=http://litellm:4000  # LiteLLM网关地址
LITELLM_MASTER_KEY=sk-litellm-master-key     # LiteLLM认证密钥

# 代理配置（可选）
# PROXY_URL=http://proxy:8080  # 代理服务器地址（暂未使用）
```

**配置优先级**:
1. route_config中的endpoints配置
2. route_config中的base_url配置
3. 环境变量配置
4. model_providers表中的api_base_url

#### 5.2 数据库迁移
**文件**: `backend/migrations/075_byok_route_mode_support.sql`

**任务**:
- ✅ ~~确保所有必要字段存在~~（已在074迁移中完成）
- ✅ ~~添加必要的索引~~（已在074迁移中完成）
- **扩展api_key_verifications表**，添加路由模式相关字段

**迁移内容**:
```sql
-- 扩展api_key_verifications表，添加路由模式相关字段
ALTER TABLE api_key_verifications 
  ADD COLUMN IF NOT EXISTS route_mode VARCHAR(20),
  ADD COLUMN IF NOT EXISTS endpoint_used VARCHAR(512),
  ADD COLUMN IF NOT EXISTS error_category VARCHAR(64);

-- 添加注释
COMMENT ON COLUMN api_key_verifications.route_mode IS '验证时使用的路由模式: direct/litellm/proxy/auto';
COMMENT ON COLUMN api_key_verifications.endpoint_used IS '验证时实际使用的endpoint地址';
COMMENT ON COLUMN api_key_verifications.error_category IS '错误分类（参考provider_error_mapper.go）';

-- 添加索引
CREATE INDEX IF NOT EXISTS idx_api_key_verifications_route_mode ON api_key_verifications(route_mode);
CREATE INDEX IF NOT EXISTS idx_api_key_verifications_error_category ON api_key_verifications(error_category);
```

#### 5.3 文档更新
**任务**:
- 更新API文档
- 更新开发文档
- 更新部署文档

**文档内容**:
- 路由模式说明
- 配置指南
- 错误处理规范
- 故障排查指南

---

## 错误处理规范

### 统一错误分类
所有路由模式必须遵循现有的错误分类体系（`provider_error_mapper.go`）：

| 错误分类 | 错误码 | 说明 | 是否可重试 |
|---------|--------|------|-----------|
| AUTH_INVALID_KEY | invalid_api_key | API密钥无效 | ❌ |
| AUTH_PERMISSION_DENIED | permission_denied | 权限被拒绝 | ❌ |
| QUOTA_INSUFFICIENT | insufficient_quota | 配额不足 | ❌ |
| RATE_LIMITED | rate_limit_exceeded | 速率限制 | ✅ |
| MODEL_NOT_FOUND | model_not_found | 模型不存在 | ❌ |
| SERVICE_UNAVAILABLE | service_unavailable | 服务不可用 | ✅ |
| NETWORK_TIMEOUT | timeout | 网络超时 | ✅ |
| UNKNOWN | unknown | 未知错误 | 视情况 |

### 错误持久化
所有验证和探测的错误信息必须持久化到`api_key_health_history`表：

```sql
-- 错误字段
status_code INT,                    -- HTTP状态码
provider_error_code VARCHAR(128),   -- 上游错误码
provider_request_id VARCHAR(128),   -- 上游请求ID
endpoint_used VARCHAR(512),         -- 使用的endpoint
error_category VARCHAR(64),         -- 错误分类
raw_error_excerpt TEXT              -- 错误摘要
```

### 重试机制
- 使用`ProviderErrorInfo.Retryable`字段判断是否重试
- 最大重试次数：3次
- 重试间隔：指数退避（1s, 2s, 4s）

---

## 实施建议

### 开发顺序
1. **Phase 1**: 后端基础设施（1-2天）
2. **Phase 2**: 路由模式实现（2-3天）
   - 先实现Direct模式（半天）
   - 再实现LiteLLM模式（1天）
   - 最后实现Proxy和Auto模式（1-1.5天）
3. **Phase 4**: 测试与验证（1-2天）
4. **Phase 3**: 前端展示优化（0.5-1天）
5. **Phase 5**: 配置与文档（0.5-1天）

### 风险控制
1. **向后兼容**: 确保现有API Key的验证功能不受影响
2. **降级策略**: Auto模式必须有明确的降级策略
3. **错误处理**: 所有路由模式使用统一的错误处理机制
4. **性能影响**: 避免路由模式判断影响验证性能

### 验收标准
1. ✅ 所有单元测试通过
2. ✅ 所有集成测试通过
3. ✅ E2E测试通过
4. ✅ 四种路由模式都能正常工作
5. ✅ 错误分类和错误码符合规范
6. ✅ 前端正确显示路由模式信息和错误信息
7. ✅ 文档更新完整

---

## 分支策略

### 主分支
- `feature/byok-route-mode-support`: 主要开发分支

### 子分支（可选）
- `feature/byok-direct-mode`: Direct模式实现
- `feature/byok-litellm-mode`: LiteLLM模式实现
- `feature/byok-proxy-mode`: Proxy模式实现
- `feature/byok-auto-mode`: Auto模式实现

### PR策略
- 建议按Phase提交PR，每个Phase一个PR
- 或者按路由模式提交PR，每个模式一个PR
- 最后合并到主分支

---

## 时间估算

| Phase | 任务 | 预计时间 | 优先级 |
|-------|------|----------|--------|
| Phase 1 | 后端基础设施 | 1-2天 | 高 |
| Phase 2 | 路由模式实现 | 2-3天 | 高 |
| Phase 3 | 前端展示优化 | 0.5-1天 | 中 |
| Phase 4 | 测试与验证 | 1-2天 | 高 |
| Phase 5 | 配置与文档 | 0.5-1天 | 中 |
| **总计** | | **5-9天** | |

---

## 成功指标

1. **功能完整性**: 四种路由模式都能正常工作
2. **代码质量**: 测试覆盖率 > 80%
3. **性能**: 验证延迟增加 < 10%
4. **用户体验**: 前端展示清晰，错误信息明确
5. **可维护性**: 代码结构清晰，文档完整
6. **错误处理**: 所有路由模式使用统一的错误处理机制

---

**计划创建时间**: 2026-04-29
**计划版本**: v1.2
**计划状态**: 已更新（基于代码深度审视）

---

## 更新日志

### v1.2 (2026-04-29)
**更新内容**:
1. **Phase 1.2**: 添加认证Token处理逻辑，明确参考ExecutionLayer实现
2. **Phase 1.4**: 添加region参数的获取和传递
3. **Phase 2**: 所有路由模式实现中添加region处理和endpoints配置结构
4. **Phase 5.1**: 明确环境变量命名，使用与ExecutionLayer一致的命名规范
5. **Phase 5.2**: 扩展api_key_verifications表，添加路由模式相关字段

**关键改进**:
- 认证Token处理：LiteLLM模式使用LITELLM_MASTER_KEY
- Region处理：支持根据region选择不同的endpoint
- Endpoints配置：支持从route_config.endpoints.{mode}.{region}获取URL
- 环境变量统一：使用LLM_GATEWAY_LITELLM_URL替代LITELLM_URL
- 数据持久化：扩展验证结果表，记录路由模式和错误分类

### v1.1 (2026-04-29)
**更新内容**:
1. 明确数据库字段和索引已存在
2. 补充错误处理规范
3. 添加错误持久化要求
4. 更新验收标准

### v1.0 (2026-04-29)
**初始版本**: 基于问题分析创建开发计划
