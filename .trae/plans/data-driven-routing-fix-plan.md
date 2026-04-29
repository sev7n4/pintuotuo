# 数据驱动路由修复方案（SSOT 简化版）

## 问题概述

当前路由系统存在以下问题：
1. `merchant_api_keys` 表已包含完整的路由配置字段（`route_mode`, `route_config`, `endpoint_url`, `fallback_endpoint_url`），但代码未使用
2. 路由逻辑依赖环境变量（`LLM_GATEWAY_ACTIVE`），而非数据库配置
3. 多层配置来源（环境变量、厂商策略、Key 配置）导致优先级混乱
4. 违反了 SSOT（Single Source of Truth）设计原则

## 设计原则

**唯一数据源**：`merchant_api_keys.route_mode` - API Key 级别的路由配置

**废弃配置来源**：
- ~~环境变量 `LLM_GATEWAY_ACTIVE`~~
- ~~`model_providers.route_strategy` 厂商级别路由策略~~

**路由模式定义**：
| route_mode | 说明 | 端点来源 |
|------------|------|----------|
| `auto` | 自动（默认直连） | `merchant_api_keys.endpoint_url` 或 `model_providers.api_base_url` |
| `direct` | 直连厂商 | `merchant_api_keys.endpoint_url` 或 `model_providers.api_base_url` |
| `litellm` | LiteLLM 网关 | 环境变量 `LLM_GATEWAY_LITELLM_URL` 或 `merchant_api_keys.endpoint_url` |
| `proxy` | 代理模式 | `merchant_api_keys.fallback_endpoint_url` |

---

## 修复范围和边界清单

### 一、需要修改的代码文件

#### 1. 后端核心文件

| 文件 | 修改内容 | 影响范围 |
|------|----------|----------|
| `backend/handlers/api_proxy.go` | 核心修改文件 | 主代理流程 |
| `backend/handlers/api_proxy_model_fallback.go` | 修改 `resolveProxyAttemptRuntime` | Fallback 流程 |
| `backend/services/execution_layer.go` | 清理环境变量依赖 | Execution Layer |

#### 2. 需要删除的函数/代码

| 文件 | 函数/代码 | 原因 |
|------|-----------|------|
| `backend/handlers/api_proxy.go:967-979` | `applyGatewayOverride()` | 废弃环境变量驱动 |
| `backend/handlers/api_proxy.go:981-993` | `resolveGatewayAuthToken()` | 废弃环境变量驱动 |
| `backend/handlers/api_proxy.go:925-945` | `applyLitellmGatewayRetryCap()` | 废弃环境变量驱动 |
| `backend/handlers/api_proxy_routing.go:43-129` | 整个文件（标记 unused） | 从未使用的冗余代码 |
| `backend/handlers/api_proxy.go:866-871` | 环境变量读取逻辑 | 废弃环境变量驱动 |
| `backend/handlers/api_proxy.go:47-53` | 常量 `llmGatewayLitellm`, `llmGatewayNone` | 废弃常量 |

#### 3. 需要新增的函数

| 文件 | 函数 | 说明 |
|------|------|------|
| `backend/handlers/api_proxy.go` | `resolveRouteMode()` | 从 API Key 配置解析路由模式 |
| `backend/handlers/api_proxy.go` | `resolveEndpointURL()` | 根据路由模式决定端点URL |
| `backend/handlers/api_proxy.go` | `applyAPIKeyRouteConfig()` | 应用 API Key 级别路由配置 |

---

### 二、需要修改的 SQL 查询

#### `selectAPIKeyForRequest` 函数中的 6 处 SQL 查询

| 位置 | 当前查询字段 | 需要新增字段 |
|------|-------------|-------------|
| 行 1102-1109 | 8 个字段 | `endpoint_url`, `fallback_endpoint_url`, `route_mode`, `route_config` |
| 行 1141-1151 | 8 个字段 | 同上 |
| 行 1167-1176 | 8 个字段 | 同上 |
| 行 1192-1199 | 8 个字段 | 同上 |
| 行 1208-1218 | 8 个字段 | 同上 |

---

### 三、需要清理的环境变量

#### 1. 完全废弃的环境变量

| 环境变量 | 当前使用位置 | 处理方式 |
|----------|-------------|----------|
| `LLM_GATEWAY_ACTIVE` | 7 处代码引用 | 删除所有引用 |
| `LLM_GATEWAY_PROXY_URL` | 2 处代码引用 | 删除所有引用 |
| `LLM_GATEWAY_PROXY_TOKEN` | 2 处代码引用 | 删除所有引用 |
| `USE_CONFIG_DRIVEN_ROUTING` | 2 处 yml 引用 | 删除（已无意义） |
| `USE_EXECUTION_LAYER` | 2 处 yml 引用 | 保留（Phase 3 灰度开关） |

#### 2. 保留的环境变量

| 环境变量 | 用途 | 说明 |
|----------|------|------|
| `LLM_GATEWAY_LITELLM_URL` | LiteLLM 网关地址 | 当 `route_mode=litellm` 且无自定义端点时使用 |
| `LITELLM_MASTER_KEY` | LiteLLM 鉴权 | 当 `route_mode=litellm` 时使用 |
| `API_PROXY_LITELLM_MAX_RETRIES` | 重试上限 | 当 `route_mode=litellm` 时使用 |

---

### 四、需要修改的配置文件

#### 1. Docker Compose 文件

| 文件 | 需要删除的配置 |
|------|---------------|
| `docker-compose.prod.yml:66` | `LLM_GATEWAY_ACTIVE=${LLM_GATEWAY_ACTIVE:-none}` |
| `docker-compose.prod.yml:69` | `USE_CONFIG_DRIVEN_ROUTING=${USE_CONFIG_DRIVEN_ROUTING:-false}` |
| `docker-compose.yml:23-24` | 同上 |

#### 2. 环境变量示例文件

| 文件 | 需要删除的配置 |
|------|---------------|
| `.env.example:91-117` | 整个 LLM 聚合网关配置区块（保留 `LLM_GATEWAY_LITELLM_URL` 和 `LITELLM_MASTER_KEY`） |

---

### 五、前端相关文件（本次不修改）

以下文件涉及 `model_providers.route_strategy`，但本次修复**仅影响后端路由执行层**，前端管理界面可保留：

| 文件 | 说明 |
|------|------|
| `frontend/src/components/admin/ProviderConfigForm.tsx` | 厂商路由策略配置表单 |
| `frontend/src/components/admin/RouteStrategyConfig.tsx` | 路由策略配置组件 |
| `frontend/src/pages/admin/AdminModelProviders.tsx` | 厂商管理页面 |
| `frontend/src/pages/admin/AdminByokRouting.tsx` | BYOK 路由管理页面（已正确使用 `route_mode`） |

**说明**：前端 `route_strategy` 配置界面可保留用于未来扩展，但后端执行层不再使用该配置。

---

### 六、文档需要更新

| 文件 | 需要更新 |
|------|----------|
| `deploy/litellm/README.md` | 更新路由配置说明 |
| `deploy/litellm/SSOT_ROUTING.md` | 更新 SSOT 说明 |

---

### 七、测试文件需要更新

| 文件 | 需要修改 |
|------|----------|
| `backend/services/execution_layer_test.go` | 移除废弃环境变量的测试用例 |

---

## 详细修改计划

### 阶段 1：扩展数据查询层

#### 1.1 修改 `scanMerchantAPIKeyQuotaRow` 函数

**位置**: `handlers/api_proxy.go:1091-1098`

**当前代码**:
```go
func scanMerchantAPIKeyQuotaRow(row *sql.Row, apiKey *models.MerchantAPIKey) error {
    var qLim sql.NullFloat64
    if err := row.Scan(&apiKey.ID, &apiKey.MerchantID, &apiKey.Provider, &apiKey.APIKeyEncrypted, &apiKey.APISecretEncrypted, &qLim, &apiKey.QuotaUsed, &apiKey.Status); err != nil {
        return err
    }
    apiKey.QuotaLimit = utils.NullFloat64Ptr(qLim)
    return nil
}
```

**修改后**:
```go
func scanMerchantAPIKeyQuotaRow(row *sql.Row, apiKey *models.MerchantAPIKey) error {
    var qLim sql.NullFloat64
    var endpointURL, fallbackEndpointURL, routeMode sql.NullString
    var routeConfigBytes []byte
    if err := row.Scan(
        &apiKey.ID, &apiKey.MerchantID, &apiKey.Provider, &apiKey.APIKeyEncrypted, &apiKey.APISecretEncrypted,
        &qLim, &apiKey.QuotaUsed, &apiKey.Status,
        &endpointURL, &fallbackEndpointURL, &routeMode, &routeConfigBytes,
    ); err != nil {
        return err
    }
    apiKey.QuotaLimit = utils.NullFloat64Ptr(qLim)
    if endpointURL.Valid {
        apiKey.EndpointURL = endpointURL.String
    }
    if fallbackEndpointURL.Valid {
        apiKey.FallbackEndpointURL = fallbackEndpointURL.String
    }
    if routeMode.Valid {
        apiKey.RouteMode = routeMode.String
    }
    if len(routeConfigBytes) > 0 {
        _ = json.Unmarshal(routeConfigBytes, &apiKey.RouteConfig)
    }
    return nil
}
```

#### 1.2 修改 `selectAPIKeyForRequest` 函数中的 SQL 查询

**位置**: `handlers/api_proxy.go:1100-1223`

**需要修改的 SQL 查询**（共 6 处）：

**修改示例**（以第一处为例）:
```sql
-- 当前
SELECT mak.id, mak.merchant_id, mak.provider, mak.api_key_encrypted, mak.api_secret_encrypted, mak.quota_limit, mak.quota_used, mak.status

-- 修改后
SELECT mak.id, mak.merchant_id, mak.provider, mak.api_key_encrypted, mak.api_secret_encrypted, mak.quota_limit, mak.quota_used, mak.status,
       COALESCE(mak.endpoint_url, '') as endpoint_url,
       COALESCE(mak.fallback_endpoint_url, '') as fallback_endpoint_url,
       COALESCE(mak.route_mode, 'auto') as route_mode,
       COALESCE(mak.route_config, '{}'::jsonb) as route_config
```

---

### 阶段 2：重构路由决策层

#### 2.1 新增 `resolveRouteMode` 函数

**位置**: `handlers/api_proxy.go` (新增函数)

```go
// resolveRouteMode 从 API Key 配置解析路由模式
// SSOT: 仅从 merchant_api_keys.route_mode 获取配置
func resolveRouteMode(apiKey *models.MerchantAPIKey) string {
    if apiKey == nil {
        return routeModeDirect
    }
    
    mode := strings.TrimSpace(strings.ToLower(apiKey.RouteMode))
    switch mode {
    case routeModeDirect, routeModeLitellm, routeModeProxy:
        return mode
    case routeModeAuto, "":
        return routeModeDirect
    default:
        return routeModeDirect
    }
}
```

#### 2.2 新增 `resolveEndpointURL` 函数

```go
// resolveEndpointURL 根据路由模式决定最终的端点URL
// SSOT: 仅使用 merchant_api_keys 中的配置
func resolveEndpointURL(routeMode string, apiKey *models.MerchantAPIKey, providerBaseURL string) string {
    switch routeMode {
    case routeModeDirect:
        if apiKey != nil && apiKey.EndpointURL != "" {
            return strings.TrimRight(apiKey.EndpointURL, "/")
        }
        return strings.TrimRight(providerBaseURL, "/")
        
    case routeModeLitellm:
        if apiKey != nil && apiKey.EndpointURL != "" {
            return strings.TrimRight(apiKey.EndpointURL, "/")
        }
        if base := strings.TrimSpace(os.Getenv("LLM_GATEWAY_LITELLM_URL")); base != "" {
            return strings.TrimRight(base, "/") + "/v1"
        }
        return ""
        
    case routeModeProxy:
        if apiKey != nil && apiKey.FallbackEndpointURL != "" {
            return strings.TrimRight(apiKey.FallbackEndpointURL, "/")
        }
        return ""
        
    default:
        return strings.TrimRight(providerBaseURL, "/")
    }
}
```

#### 2.3 新增 `applyAPIKeyRouteConfig` 函数

```go
// applyAPIKeyRouteConfig 应用 API Key 级别的路由配置
// SSOT: 完全基于 merchant_api_keys 表的配置
func applyAPIKeyRouteConfig(cfg providerRuntimeConfig, apiKey *models.MerchantAPIKey) providerRuntimeConfig {
    routeMode := resolveRouteMode(apiKey)
    endpointURL := resolveEndpointURL(routeMode, apiKey, cfg.APIBaseURL)
    
    if endpointURL != "" {
        cfg.APIBaseURL = endpointURL
    }
    
    return cfg
}
```

---

### 阶段 3：删除废弃代码

#### 3.1 删除 `applyGatewayOverride` 函数

**位置**: `handlers/api_proxy.go:967-979`

#### 3.2 删除 `resolveGatewayAuthToken` 函数

**位置**: `handlers/api_proxy.go:981-993`

#### 3.3 删除 `applyLitellmGatewayRetryCap` 函数

**位置**: `handlers/api_proxy.go:925-945`

#### 3.4 删除整个 `api_proxy_routing.go` 文件

**位置**: `handlers/api_proxy_routing.go`（整个文件标记为 unused）

#### 3.5 删除废弃常量

**位置**: `handlers/api_proxy.go:47-53`

```go
// 删除以下常量
llmGatewayLitellm = "litellm"
llmGatewayNone    = "none"
```

---

### 阶段 4：修改调用链

#### 4.1 修改主代理流程

**位置**: `handlers/api_proxy.go:375`

**当前代码**:
```go
providerCfg = applyGatewayOverride(providerCfg)
```

**修改后**:
```go
providerCfg = applyAPIKeyRouteConfig(providerCfg, &apiKey)
```

#### 4.2 修改 `resolveProxyAttemptRuntime` 函数

**位置**: `handlers/api_proxy_model_fallback.go:118`

**当前代码**:
```go
pcfg = applyGatewayOverride(pcfg)
```

**修改后**:
```go
pcfg = applyAPIKeyRouteConfig(pcfg, &pk)
```

#### 4.3 修改鉴权逻辑

**位置**: `handlers/api_proxy.go:503` 和 `handlers/api_proxy.go:672`

**当前代码**:
```go
authToken := resolveGatewayAuthToken(pcfg, dk)
```

**修改后**:
```go
authToken := resolveAuthTokenFromRouteMode(routeMode, dk)
```

新增函数：
```go
func resolveAuthTokenFromRouteMode(routeMode string, fallbackToken string) string {
    if routeMode == routeModeLitellm {
        if token := strings.TrimSpace(os.Getenv("LITELLM_MASTER_KEY")); token != "" {
            return token
        }
    }
    return fallbackToken
}
```

---

### 阶段 5：清理配置文件

#### 5.1 修改 `docker-compose.prod.yml`

删除以下环境变量：
- `LLM_GATEWAY_ACTIVE`
- `USE_CONFIG_DRIVEN_ROUTING`

#### 5.2 修改 `docker-compose.yml`

同上

#### 5.3 修改 `.env.example`

更新 LLM 聚合网关配置区块说明

---

### 阶段 6：清理 Execution Layer

#### 6.1 修改 `execution_layer.go`

**位置**: `services/execution_layer.go:341-359`

**当前代码**:
```go
func (l *ExecutionLayer) determineGatewayMode(cfg *ExecutionProviderConfig) string {
    if cfg == nil {
        return GatewayModeDirect
    }

    if cfg.BYOKRouteMode != "" && cfg.BYOKRouteMode != "auto" {
        return cfg.BYOKRouteMode
    }

    if cfg.GatewayMode != "" {
        return cfg.GatewayMode
    }

    envMode := os.Getenv("LLM_GATEWAY_ACTIVE")
    if envMode != "" && envMode != "none" {
        return envMode
    }

    return GatewayModeDirect
}
```

**修改后**:
```go
func (l *ExecutionLayer) determineGatewayMode(cfg *ExecutionProviderConfig) string {
    if cfg == nil {
        return GatewayModeDirect
    }

    if cfg.BYOKRouteMode != "" && cfg.BYOKRouteMode != "auto" {
        return cfg.BYOKRouteMode
    }

    return GatewayModeDirect
}
```

---

## 数据迁移

### 迁移脚本

如果当前环境使用 `LLM_GATEWAY_ACTIVE=litellm`，需要将所有 API Key 的 `route_mode` 设置为 `litellm`：

```sql
-- 将所有现有 API Key 的 route_mode 设置为当前环境变量对应的值
-- 如果 LLM_GATEWAY_ACTIVE=litellm，则执行：
UPDATE merchant_api_keys 
SET route_mode = 'litellm' 
WHERE route_mode IS NULL OR route_mode = 'auto';

-- 如果 LLM_GATEWAY_ACTIVE 为空或 none，则执行：
UPDATE merchant_api_keys 
SET route_mode = 'direct' 
WHERE route_mode IS NULL OR route_mode = 'auto';
```

---

## 测试验证

### 测试场景

1. **直连模式**
   - 设置 `merchant_api_keys.route_mode = 'direct'`
   - 验证请求直接发送到厂商端点

2. **LiteLLM 网关模式**
   - 设置 `merchant_api_keys.route_mode = 'litellm'`
   - 验证请求通过 LiteLLM 网关

3. **代理模式**
   - 设置 `merchant_api_keys.route_mode = 'proxy'`
   - 设置 `merchant_api_keys.fallback_endpoint_url`
   - 验证请求发送到代理端点

4. **自定义端点**
   - 设置 `merchant_api_keys.endpoint_url`
   - 验证请求发送到自定义端点

5. **默认值**
   - `route_mode` 为空或 `auto`
   - 验证使用直连模式

### 验证命令

```bash
# 部署后验证
curl -X POST https://api.example.com/v1/chat/completions \
  -H "Authorization: Bearer <user_token>" \
  -H "Content-Type: application/json" \
  -d '{"model": "step-1-8K", "messages": [{"role": "user", "content": "hello"}]}'
```

---

## 风险评估

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| SQL 查询变更 | 中 | 保持向后兼容，新增字段使用 `COALESCE` |
| 环境变量废弃 | 低 | 迁移脚本将环境变量配置转换为数据库配置 |
| 默认值处理 | 低 | 空值和 `auto` 统一使用 `direct` |
| 删除未使用代码 | 低 | `api_proxy_routing.go` 从未被调用 |

---

## 实施步骤

1. ✅ 分析代码，确认问题范围
2. ✅ 排查所有相关代码和配置
3. ⬜ 添加路由模式常量定义
4. ⬜ 修改 `scanMerchantAPIKeyQuotaRow` 函数
5. ⬜ 修改 `selectAPIKeyForRequest` 中的 SQL 查询
6. ⬜ 新增 `resolveRouteMode`、`resolveEndpointURL`、`applyAPIKeyRouteConfig` 函数
7. ⬜ 删除废弃函数和代码
8. ⬜ 修改主代理流程调用
9. ⬜ 修改 `resolveProxyAttemptRuntime` 函数
10. ⬜ 清理 `execution_layer.go`
11. ⬜ 清理配置文件
12. ⬜ 编译验证
13. ⬜ 执行数据迁移
14. ⬜ 部署测试
