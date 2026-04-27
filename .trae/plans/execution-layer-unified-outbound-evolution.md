# 三层路由架构执行层统一出站演进计划

> **文档版本**: v1.4
> **创建时间**: 2026-04-26
> **最后更新**: 2026-04-26 21:30

## 版本历史

| 版本 | 日期 | 更新内容 |
|------|------|----------|
| v1.0 | 2026-04-26 | 初始版本：执行层统一出站基础计划 |
| v1.1 | 2026-04-26 | 新增 Phase 2.5 配置驱动路由集成 |
| v1.2 | 2026-04-26 | 新增监控与可观测性、E2E 测试场景 |
| v1.3 | 2026-04-26 | 新增环境变量配置、LiteLLM 配置说明 |
| v1.4 | 2026-04-26 | 修正架构图：双入口（OpenAI 兼容 + 自定义代理） |

---

## 文档目标

将执行层改造为统一的出站入口，整合 Gateway 选择、认证、HTTP 执行等逻辑，实现配置驱动路由

---

## 一、现状总结

### 1.1 当前架构问题

| 问题 | 描述 | 影响 |
|------|------|------|
| 执行层未使用 | `ExecutionLayer.Execute()` 定义但未被主流程调用 | 架构不完整 |
| Gateway 选择独立 | `applyGatewayOverride()` 在 handler 中独立处理 | 职责分散 |
| 出站逻辑散布 | 分布在 `api_proxy.go`、`api_proxy_http.go` | 维护困难 |
| 重试逻辑重复 | `executeProviderRequestWithRetry()` 与 `ExecutionEngine.ExecuteWithRetry()` 功能重叠 | 代码冗余 |

### 1.2 🔴 关键发现：配置驱动架构已存在但未使用

| 组件 | 状态 | 说明 |
|------|------|------|
| **数据库字段** | ✅ 已存在 | `model_providers.route_strategy`, `endpoints` (migration 062) |
| **管理端 API** | ✅ 已存在 | `GetProviderRouteConfigs`, `UpdateProviderRouteConfig` (`route_config.go`) |
| **服务层逻辑** | ✅ 已存在 | `UnifiedRouter.DecideRoute()` (`unified_router.go`) |
| **主流程集成** | ❌ 未使用 | 仍然使用环境变量 `LLM_GATEWAY_ACTIVE` |

#### 数据库配置结构

```sql
-- model_providers 表已存在的字段
route_strategy JSONB  -- 路由策略配置
endpoints JSONB       -- 多端点配置
provider_region VARCHAR(20)  -- 厂商区域

-- 示例配置（已在 migration 062 中初始化）
{
  "route_strategy": {
    "default_mode": "auto",
    "domestic_users": {
      "mode": "litellm",
      "fallback_mode": "proxy",
      "proxy_endpoint": "gaap"
    },
    "overseas_users": {
      "mode": "direct"
    },
    "enterprise_users": {
      "mode": "litellm",
      "fallback_mode": "proxy"
    }
  },
  "endpoints": {
    "direct": {
      "domestic": "",
      "overseas": "https://api.openai.com/v1"
    },
    "litellm": {
      "domestic": "http://litellm-overseas:4000/v1",
      "overseas": "http://litellm-overseas:4000/v1"
    },
    "proxy": {
      "gaap": "https://openai-gaap.example.com",
      "nginx_hk": "https://openai-proxy-hk.example.com"
    }
  }
}
```

#### 已存在的服务层代码

```go
// services/unified_router.go - 已存在
func (r *UnifiedRouter) DecideRoute(
    ctx context.Context,
    providerConfig *ProviderConfig,
    merchantConfig *MerchantConfig,
) (*RouteDecision, error) {
    // 根据商户类型和区域决定路由模式
    // 返回: mode, endpoint, fallback_mode, fallback_endpoint
}
```

#### 环境变量 vs 配置驱动对比

| 特性 | 环境变量方式（当前） | 配置驱动方式（目标） |
|------|---------------------|---------------------|
| **粒度** | 全局单一模式 | 按厂商、按商户类型 |
| **灵活性** | 需重启服务 | 实时生效 |
| **可观测性** | 无 | 管理端可视化配置 |
| **降级策略** | 无 | 支持 fallback_mode |
| **区域感知** | 无 | 支持 domestic/overseas |

#### 商户路由信息来源

```sql
-- merchants 表相关字段
type VARCHAR(20)    -- 商户类型: 'regular', 'enterprise', 'trial'
region VARCHAR(20)  -- 商户区域: 'domestic', 'overseas'

-- 路由决策逻辑
-- 1. 根据 merchant.type 决定用户类型键
--    - enterprise → 'enterprise_users'
--    - 其他 → 'domestic_users' 或 'overseas_users'（根据 region）
-- 2. 从 provider.route_strategy 中查找对应策略
-- 3. 从 provider.endpoints 中选择对应端点
```

### 1.3 优化前后架构对比

#### 1.3.1 当前架构（优化前）

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           当前架构（环境变量驱动）                                │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │                          用户请求入口（双入口）                           │   │
│  │                                                                          │   │
│  │  入口 1: OpenAI 兼容接口（推荐）                                          │   │
│  │  POST /api/v1/openai/v1/chat/completions                                 │   │
│  │  ├── 认证: API Key (ptd_*/ptt_*) 或 JWT                                  │   │
│  │  ├── 请求体: { "model": "gpt-4", "messages": [...] }                     │   │
│  │  └── 自动解析: model → provider + modelName                              │   │
│  │                                                                          │   │
│  │  入口 2: 自定义代理接口                                                   │   │
│  │  POST /api/v1/proxy/chat                                                 │   │
│  │  ├── 认证: JWT                                                           │   │
│  │  ├── 请求体: { "provider": "openai", "model": "gpt-4", ... }             │   │
│  │  └── 显式指定 provider                                                   │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                      │                                          │
│                                      ▼                                          │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │                        认证与请求准备                                    │   │
│  │  ┌───────────────┐  ┌─────────────────┐  ┌───────────────────────────┐  │   │
│  │  │ JWT/API Key   │─▶│ 解析请求体       │─▶│ 余额检查 & 预扣款         │  │   │
│  │  │ 认证          │  │                 │  │                           │  │   │
│  │  └───────────────┘  └─────────────────┘  └───────────────────────────┘  │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                      │                                          │
│                                      ▼                                          │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │                     三层路由（仅选择 API Key）                            │   │
│  │  ┌─────────────┐   ┌─────────────┐   ┌─────────────────────────────┐   │   │
│  │  │  策略层     │──▶│  决策层     │──▶│  执行层（未使用）            │   │   │
│  │  │             │   │             │   │                             │   │   │
│  │  │ SmartRouter │   │ SelectKey   │   │ ExecutionLayer.Execute()    │   │   │
│  │  │ Candidates  │   │ HealthCheck │   │ ❌ 空壳，未被调用           │   │   │
│  │  └─────────────┘   └─────────────┘   └─────────────────────────────┘   │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                      │                                          │
│                                      ▼                                          │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │                     出站逻辑（环境变量驱动）⚠️                            │   │
│  │                                                                          │   │
│  │  ┌─────────────────────────────────────────────────────────────────┐    │   │
│  │  │  LLM_GATEWAY_ACTIVE=litellm  （全局单一模式）                     │    │   │
│  │  │           │                                                     │    │   │
│  │  │           ▼                                                     │    │   │
│  │  │  applyGatewayOverride()                                         │    │   │
│  │  │           │                                                     │    │   │
│  │  │           ▼                                                     │    │   │
│  │  │  所有请求 → LiteLLM Gateway → Provider                          │    │   │
│  │  │           │                                                     │    │   │
│  │  │           ▼                                                     │    │   │
│  │  │  executeProviderRequestWithRetry()                              │    │   │
│  │  └─────────────────────────────────────────────────────────────────┘    │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                      │                                          │
│                                      ▼                                          │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │                          结算与响应                                      │   │
│  │  processResponseAndSettlement()                                         │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                                                                 │
│  ⚠️ 问题：                                                                      │
│  1. 执行层定义但未使用，架构不完整                                              │
│  2. Gateway 选择全局单一，无法按厂商/商户区分                                   │
│  3. 数据库配置（route_strategy, endpoints）未读取                              │
│  4. 无降级策略，无区域感知                                                     │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

#### 1.3.2 目标架构（优化后）

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           目标架构（配置驱动）                                    │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │                          用户请求入口（双入口）                           │   │
│  │                                                                          │   │
│  │  入口 1: OpenAI 兼容接口（推荐）                                          │   │
│  │  POST /api/v1/openai/v1/chat/completions                                 │   │
│  │  ├── 认证: API Key (ptd_*/ptt_*) 或 JWT                                  │   │
│  │  ├── 请求体: { "model": "gpt-4", "messages": [...] }                     │   │
│  │  └── 自动解析: model → provider + modelName                              │   │
│  │                                                                          │   │
│  │  入口 2: 自定义代理接口                                                   │   │
│  │  POST /api/v1/proxy/chat                                                 │   │
│  │  ├── 认证: JWT                                                           │   │
│  │  ├── 请求体: { "provider": "openai", "model": "gpt-4", ... }             │   │
│  │  └── 显式指定 provider                                                   │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                      │                                          │
│                                      ▼                                          │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │                        认证与请求准备                                    │   │
│  │  ┌───────────────┐  ┌─────────────────┐  ┌───────────────────────────┐  │   │
│  │  │ JWT/API Key   │─▶│ 解析请求体       │─▶│ 余额检查 & 预扣款         │  │   │
│  │  │ 认证          │  │                 │  │                           │  │   │
│  │  └───────────────┘  └─────────────────┘  └───────────────────────────┘  │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                      │                                          │
│                                      ▼                                          │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │                     三层路由（完整架构）                                  │   │
│  │  ┌─────────────┐   ┌─────────────┐   ┌─────────────────────────────┐   │   │
│  │  │  策略层     │──▶│  决策层     │──▶│      执行层                  │   │   │
│  │  │             │   │             │   │                             │   │   │
│  │  │ SmartRouter │   │ SelectKey   │   │ ✅ ExecutionLayer.Execute() │   │   │
│  │  │ Candidates  │   │ HealthCheck │   │    ├── resolveEndpoint()    │   │   │
│  │  │             │   │             │   │    ├── resolveAuthToken()   │   │   │
│  │  │             │   │             │   │    └── ExecuteWithRetry()   │   │   │
│  │  └─────────────┘   └─────────────┘   └─────────────────────────────┘   │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                      │                                          │
│                                      ▼                                          │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │                   配置驱动路由决策 ✅                                     │   │
│  │                                                                          │   │
│  │  ┌─────────────────────────────────────────────────────────────────┐    │   │
│  │  │  数据库配置（唯一真相来源）                                       │    │   │
│  │  │                                                                  │    │   │
│  │  │  merchant_api_keys 表:                                           │    │   │
│  │  │    - api_key_encrypted  ← 商户上传的 API Key                     │    │   │
│  │  │    - route_preference   ← 路由偏好（管理端维护）                  │    │   │
│  │  │    - merchant_region    ← 商户区域                               │    │   │
│  │  │                                                                  │    │   │
│  │  │  model_providers 表:                                             │    │   │
│  │  │    - route_strategy     ← 路由策略（管理端维护）                  │    │   │
│  │  │    - endpoints          ← 多端点配置                             │    │   │
│  │  │    - provider_region    ← 厂商区域                               │    │   │
│  │  └─────────────────────────────────────────────────────────────────┘    │   │
│  │                          │                                               │   │
│  │                          ▼                                               │   │
│  │  ┌─────────────────────────────────────────────────────────────────┐    │   │
│  │  │  resolveRouteDecision(provider, merchant)                        │    │   │
│  │  │           │                                                      │    │   │
│  │  │           ▼                                                      │    │   │
│  │  │  UnifiedRouter.DecideRoute()                                     │    │   │
│  │  │           │                                                      │    │   │
│  │  │           ├── domestic_users + overseas provider → litellm       │    │   │
│  │  │           ├── overseas_users + overseas provider → direct        │    │   │
│  │  │           ├── enterprise_users → litellm + fallback              │    │   │
│  │  │           └── 配置缺失 → 降级到环境变量                           │    │   │
│  │  └─────────────────────────────────────────────────────────────────┘    │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                      │                                          │
│                                      ▼                                          │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │                     执行层统一出站 ✅                                     │   │
│  │                                                                          │   │
│  │  ┌─────────────────────────────────────────────────────────────────┐    │   │
│  │  │  ExecutionLayer.Execute()                                        │    │   │
│  │  │           │                                                      │    │   │
│  │  │           ├── mode=direct   → Provider API                       │    │   │
│  │  │           ├── mode=litellm  → LiteLLM Gateway                    │    │   │
│  │  │           └── mode=proxy    → Custom Proxy                       │    │   │
│  │  │                                                                  │    │   │
│  │  │  ExecutionEngine.ExecuteWithRetry()                              │    │   │
│  │  │           │                                                      │    │   │
│  │  │           └── 重试策略 + 熔断器 + 流式支持                        │    │   │
│  │  └─────────────────────────────────────────────────────────────────┘    │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                      │                                          │
│                                      ▼                                          │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │                          结算与响应                                      │   │
│  │  processResponseAndSettlement()                                         │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                                                                 │
│  ✅ 改进：                                                                      │
│  1. 执行层成为统一出站入口，架构完整                                            │
│  2. 配置驱动，按厂商/商户类型/区域区分路由                                      │
│  3. 数据库配置实时生效，管理端可视化                                            │
│  4. 支持降级策略、区域感知、fallback                                            │
│  5. merchant_api_keys 为唯一真相来源，职责分离                                 │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### 1.4 用户调用大模型业务流程图

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                        用户调用大模型完整流程                                    │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │                              用户发起请求                                │   │
│  │                                                                          │   │
│  │  POST /api/v1/proxy/chat                                                 │   │
│  │  Headers: Authorization: Bearer <JWT>                                    │   │
│  │  Body: { provider: "openai", model: "gpt-4", messages: [...] }          │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                      │                                          │
│                                      ▼                                          │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │                           1. 认证层                                      │   │
│  │  ┌─────────────────────────────────────────────────────────────────┐    │   │
│  │  │  authenticateUser(c) → userID, error                            │    │   │
│  │  │    - JWT 验证                                                    │    │   │
│  │  │    - 从 Token 提取 user_id                                       │    │   │
│  │  └─────────────────────────────────────────────────────────────────┘    │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                      │                                          │
│                                      ▼                                          │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │                         2. 请求准备层                                    │   │
│  │  ┌─────────────────────────────────────────────────────────────────┐    │   │
│  │  │  validateAndPrepareRequest()                                     │    │   │
│  │  │    ├── parseAPIProxyRequest()     解析请求体                     │    │   │
│  │  │    ├── resolveStrictPricingVersion() 定价版本                    │    │   │
│  │  │    ├── getTokenBalance()          获取余额                       │    │   │
│  │  │    ├── estimateTokenUsage()       预估用量                       │    │   │
│  │  │    └── hasSufficientBalance()     余额检查                       │    │   │
│  │  │           │                                                      │    │   │
│  │  │           ▼                                                      │    │   │
│  │  │    billingEngine.PreDeductBalance() 预扣款                       │    │   │
│  │  └─────────────────────────────────────────────────────────────────┘    │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                      │                                          │
│                                      ▼                                          │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │                      3. 三层路由 - 策略层                                │   │
│  │  ┌─────────────────────────────────────────────────────────────────┐    │   │
│  │  │  resolveRoutingSelection()                                       │    │   │
│  │  │    ├── routingStrategyWithSource() 获取路由策略                  │    │   │
│  │  │    │     - 环境变量 LLM_ROUTING_STRATEGY                         │    │   │
│  │  │    │     - 数据库 merchant_routing_policies                      │    │   │
│  │  │    │     - 默认策略                                              │    │   │
│  │  │    │                                                             │    │   │
│  │  │    └── trySelectAPIKeyWithSmartRouter()                          │    │   │
│  │  │          ├── SmartRouter.GetCandidatesWithKeyAllowlist()         │    │   │
│  │  │          ├── 候选 Key 评分排序                                    │    │   │
│  │  │          └── 健康检查 & 熔断器检查                                │    │   │
│  │  └─────────────────────────────────────────────────────────────────┘    │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                      │                                          │
│                                      ▼                                          │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │                      4. 三层路由 - 决策层                                │   │
│  │  ┌─────────────────────────────────────────────────────────────────┐    │   │
│  │  │  selectAPIKeyForRequest()                                        │    │   │
│  │  │    ├── 指定 API Key ID → 直接验证权限                             │    │   │
│  │  │    ├── 指定 Merchant SKU ID → 查询关联 API Key                   │    │   │
│  │  │    └── 自动选择 → 按配额余额排序选择最优 Key                      │    │   │
│  │  │           │                                                      │    │   │
│  │  │           ▼                                                      │    │   │
│  │  │    验证 API Key 状态                                              │    │   │
│  │  │      - status = 'active'                                         │    │   │
│  │  │      - verified_at IS NOT NULL                                   │    │   │
│  │  │      - quota_used < quota_limit                                  │    │   │
│  │  │      - merchant.status IN ('active', 'approved')                 │    │   │
│  │  └─────────────────────────────────────────────────────────────────┘    │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                      │                                          │
│                                      ▼                                          │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │                   5. 配置驱动路由决策 【新增】                            │   │
│  │  ┌─────────────────────────────────────────────────────────────────┐    │   │
│  │  │  getProviderRuntimeConfig()                                      │    │   │
│  │  │    - 读取 route_strategy, endpoints, provider_region            │    │   │
│  │  │           │                                                      │    │   │
│  │  │           ▼                                                      │    │   │
│  │  │  getMerchantRouteInfo()                                          │    │   │
│  │  │    - 读取 merchant.type, merchant.region                        │    │   │
│  │  │           │                                                      │    │   │
│  │  │           ▼                                                      │    │   │
│  │  │  resolveRouteDecision()                                          │    │   │
│  │  │    ┌─────────────────────────────────────────────────────────┐  │    │   │
│  │  │    │  UnifiedRouter.DecideRoute()                             │  │    │   │
│  │  │    │                                                         │  │    │   │
│  │  │    │  商户类型: enterprise                                    │  │    │   │
│  │  │    │    └── enterprise_users → litellm + fallback             │  │    │   │
│  │  │    │                                                         │  │    │   │
│  │  │    │  商户类型: regular + 区域: domestic                      │  │    │   │
│  │  │    │    └── domestic_users → litellm                          │  │    │   │
│  │  │    │                                                         │  │    │   │
│  │  │    │  商户类型: regular + 区域: overseas                      │  │    │   │
│  │  │    │    └── overseas_users → direct                           │  │    │   │
│  │  │    │                                                         │  │    │   │
│  │  │    │  配置缺失 → 降级到环境变量                                │  │    │   │
│  │  │    └─────────────────────────────────────────────────────────┘  │    │   │
│  │  │           │                                                      │    │   │
│  │  │           ▼                                                      │    │   │
│  │  │  RouteDecision { mode, endpoint, fallback_mode, fallback_endpoint }│   │
│  │  └─────────────────────────────────────────────────────────────────┘    │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                      │                                          │
│                                      ▼                                          │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │                      6. 三层路由 - 执行层                                │   │
│  │  ┌─────────────────────────────────────────────────────────────────┐    │   │
│  │  │  ExecutionLayer.Execute() 【统一出站入口】                        │    │   │
│  │  │           │                                                      │    │   │
│  │  │           ├── resolveEndpoint()                                  │    │   │
│  │  │           │     - 根据 mode 选择端点 URL                          │    │   │
│  │  │           │                                                      │    │   │
│  │  │           ├── resolveAuthToken()                                 │    │   │
│  │  │           │     - litellm → LITELLM_MASTER_KEY                   │    │   │
│  │  │           │     - direct/proxy → Provider API Key                │    │   │
│  │  │           │                                                      │    │   │
│  │  │           └── ExecutionEngine.ExecuteWithRetry()                 │    │   │
│  │  │                 ├── 构建 HTTP 请求                                │    │   │
│  │  │                 ├── 执行请求（带重试）                            │    │   │
│  │  │                 │     - MaxRetries: 3                            │    │   │
│  │  │                 │     - InitialDelay: 100ms                      │    │   │
│  │  │                 │     - BackoffFactor: 2.0                       │    │   │
│  │  │                 └── 熔断器检查                                    │    │   │
│  │  └─────────────────────────────────────────────────────────────────┘    │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                      │                                          │
│                    ┌─────────────────┴─────────────────┐                       │
│                    │                                   │                       │
│                    ▼                                   ▼                       │
│        ┌─────────────────────┐             ┌─────────────────────┐             │
│        │    流式响应处理      │             │    非流式响应处理    │             │
│        │  stream = true      │             │  stream = false     │             │
│        │                     │             │                     │             │
│        │  SSE 透传 + 解析    │             │  JSON 解析          │             │
│        └─────────────────────┘             └─────────────────────┘             │
│                    │                                   │                       │
│                    └─────────────────┬─────────────────┘                       │
│                                      │                                          │
│                                      ▼                                          │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │                          7. 结算层                                       │   │
│  │  ┌─────────────────────────────────────────────────────────────────┐    │   │
│  │  │  processResponseAndSettlement()                                  │    │   │
│  │  │    ├── 解析响应                                                   │    │   │
│  │  │    │     - inputTokens = apiResp.Usage.PromptTokens              │    │   │
│  │  │    │     - outputTokens = apiResp.Usage.CompletionTokens         │    │   │
│  │  │    │                                                             │    │   │
│  │  │    ├── Token 计费                                                 │    │   │
│  │  │    │     - cost = calculateTokenCost()                           │    │   │
│  │  │    │     - billingEngine.SettlePreDeduct()                       │    │   │
│  │  │    │                                                             │    │   │
│  │  │    ├── 记录日志                                                   │    │   │
│  │  │    │     - UPDATE merchant_api_keys SET quota_used               │    │   │
│  │  │    │     - INSERT INTO api_usage_logs                            │    │   │
│  │  │    │     - INSERT INTO routing_decisions                         │    │   │
│  │  │    │                                                             │    │   │
│  │  │    └── 清理缓存                                                   │    │   │
│  │  │          - cache.Delete(TokenBalanceKey)                        │    │   │
│  │  │          - cache.Delete(ComputePointBalanceKey)                 │    │   │
│  │  └─────────────────────────────────────────────────────────────────┘    │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                      │                                          │
│                                      ▼                                          │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │                          8. 返回响应                                     │   │
│  │                                                                          │   │
│  │  c.Data(statusCode, "application/json", body)                           │   │
│  │                                                                          │   │
│  │  响应体:                                                                 │   │
│  │  {                                                                       │   │
│  │    "id": "chatcmpl-xxx",                                                 │   │
│  │    "choices": [{ "message": { "content": "..." } }],                    │   │
│  │    "usage": { "prompt_tokens": 100, "completion_tokens": 50 }           │   │
│  │  }                                                                       │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### 1.5 当前调用链

```
proxyAPIRequestCore()
    ├── validateAndPrepareRequest()
    ├── resolveRoutingSelection()
    │       └── trySelectAPIKeyWithSmartRouter() → SelectedAPIKeyID
    ├── selectAPIKeyForRequest()
    ├── getProviderRuntimeConfig()         ← 只读取 code, name, api_base_url, api_format
    ├── applyGatewayOverride()             ← Gateway 选择（环境变量驱动）
    ├── executeProviderRequestWithRetry()  ← HTTP 执行（独立）
    └── processResponseAndSettlement()
```

### 1.5 当前调用链

```
proxyAPIRequestCore()
    ├── validateAndPrepareRequest()
    ├── resolveRoutingSelection()
    │       └── trySelectAPIKeyWithSmartRouter() → SelectedAPIKeyID
    ├── selectAPIKeyForRequest()
    ├── getProviderRuntimeConfig()         ← 只读取 code, name, api_base_url, api_format
    ├── applyGatewayOverride()             ← Gateway 选择（环境变量驱动）
    ├── executeProviderRequestWithRetry()  ← HTTP 执行（独立）
    └── processResponseAndSettlement()
```

### 1.6 目标调用链

```
proxyAPIRequestCore()
    ├── validateAndPrepareRequest()
    ├── resolveRoutingSelection()
    │       └── trySelectAPIKeyWithSmartRouter() → SelectedAPIKeyID
    ├── getProviderRuntimeConfig()         ← 扩展：读取 route_strategy + endpoints
    │       │
    │       └── resolveRouteDecision()     ← 【新增】调用 UnifiedRouter.DecideRoute()
    │               │
    │               └── 根据商户类型/区域决定 mode + endpoint
    │
    └── executionLayer.Execute()           ← 统一出站入口
            ├── 使用数据库配置的 endpoint
            ├── 使用对应的 auth token
            └── engine.ExecuteWithRetry()
```

---

## 二、演进阶段规划

### Phase 1: 基础设施准备（低风险）

**目标**：增强 ExecutionLayer 和 ExecutionEngine 能力，不改变现有流程

#### 1.1 扩展 ExecutionProviderConfig

```go
// services/execution_layer.go
type ExecutionProviderConfig struct {
    Code           string
    Name           string
    APIBaseURL     string
    APIFormat      string
    GatewayMode    string                  // 新增: "direct" | "litellm" | "proxy"
    ProviderRegion string                  // 新增: "domestic" | "overseas"
    RouteStrategy  map[string]interface{}  // 新增: 路由策略配置
    Endpoints      map[string]interface{}  // 新增: 多端点配置
}
```

#### 1.2 扩展 ExecutionInput

```go
// services/execution_engine.go
type ExecutionInput struct {
    Provider       string
    Model          string
    APIKey         string
    EndpointURL    string
    RequestFormat  string
    RequestBody    []byte
    Headers        map[string]string
    Messages       []Message
    Stream         bool
    Options        json.RawMessage
    // 新增
    OriginalAPIKey  string  // 原始 Provider Key
    GatewayMode     string  // Gateway 模式
    ProviderBaseURL string  // 原始 Provider URL
    FallbackURL     string  // 降级端点 URL
}
```

#### 1.3 新增单元测试

- `TestExecutionLayer_ResolveEndpoint`
- `TestExecutionLayer_ResolveAuthToken`
- `TestExecutionLayer_GatewayMode`

**预计工作量**：2-3 小时
**风险等级**：低（不改变现有流程）

---

### Phase 2: 执行层能力增强（中风险）

**目标**：让 ExecutionLayer 能够完整处理出站请求

#### 2.1 增强 ExecutionLayer.Execute()

```go
func (l *ExecutionLayer) Execute(ctx context.Context, input *ExecutionLayerInput) (*ExecutionLayerOutput, error) {
    startTime := time.Now()

    // 1. 解析 Gateway 模式（优先使用数据库配置，降级到环境变量）
    gatewayMode := l.determineGatewayMode(input.ProviderConfig)
    input.ProviderConfig.GatewayMode = gatewayMode

    // 2. 解析端点 URL
    endpointURL := l.resolveEndpoint(input.ProviderConfig)

    // 3. 解析认证 Token
    authToken := l.resolveAuthToken(input.ProviderConfig, input.DecryptedAPIKey)

    // 4. 构建执行输入
    execInput := &ExecutionInput{
        Provider:        input.ProviderConfig.Code,
        Model:           l.resolveModel(input),
        APIKey:          authToken,
        EndpointURL:     endpointURL,
        RequestFormat:   input.ProviderConfig.APIFormat,
        RequestBody:     input.RequestBody,
        Messages:        input.Messages,
        Stream:          input.Stream,
        Options:         input.Options,
        OriginalAPIKey:  input.DecryptedAPIKey,
        GatewayMode:     gatewayMode,
        ProviderBaseURL: input.ProviderConfig.APIBaseURL,
    }

    // 5. 执行请求
    result, err := l.engine.ExecuteWithRetry(ctx, execInput)

    // 6. 记录结果
    if input.RoutingDecision != nil {
        l.recordExecutionResult(input.RoutingDecision, result)
    }

    return &ExecutionLayerOutput{
        Result:     result,
        Decision:   input.RoutingDecision,
        DurationMs: int(time.Since(startTime).Milliseconds()),
    }, nil
}
```

#### 2.2 增强 ExecutionEngine

```go
// 新增流式支持
func (e *ExecutionEngine) ExecuteStream(ctx context.Context, input *ExecutionInput) (*http.Response, error) {
    req, err := e.buildHTTPRequest(ctx, input)
    if err != nil {
        return nil, err
    }
    req.Header.Set("Accept", "text/event-stream")
    return e.httpClient.Do(req)
}

// 新增重试策略配置
func (e *ExecutionEngine) WithRetryPolicy(policy *RetryPolicy) *ExecutionEngine {
    e.retryPolicy = policy
    return e
}
```

**预计工作量**：4-6 小时
**风险等级**：中（需要充分测试）

---

### Phase 2.5: 🔴 配置驱动路由集成（关键步骤）

**目标**：打通数据库配置与主流程，实现配置驱动路由

#### 2.5.1 扩展 getProviderRuntimeConfig()

```go
// handlers/api_proxy_http.go
type providerRuntimeConfig struct {
    Code           string
    Name           string
    APIBaseURL     string
    APIFormat      string
    // 新增
    ProviderRegion string
    RouteStrategy  map[string]interface{}
    Endpoints      map[string]interface{}
}

func getProviderRuntimeConfig(db *sql.DB, providerCode string) (providerRuntimeConfig, error) {
    var cfg providerRuntimeConfig
    var routeStrategyJSON, endpointsJSON []byte
    
    err := db.QueryRow(
        `SELECT code, name, COALESCE(api_base_url, ''), api_format,
                COALESCE(provider_region, 'domestic'),
                COALESCE(route_strategy, '{}'::jsonb),
                COALESCE(endpoints, '{}'::jsonb)
         FROM model_providers
         WHERE code = $1 AND status = 'active'
         LIMIT 1`,
        providerCode,
    ).Scan(&cfg.Code, &cfg.Name, &cfg.APIBaseURL, &cfg.APIFormat,
           &cfg.ProviderRegion, &routeStrategyJSON, &endpointsJSON)
    
    if err != nil {
        return cfg, err
    }
    
    json.Unmarshal(routeStrategyJSON, &cfg.RouteStrategy)
    json.Unmarshal(endpointsJSON, &cfg.Endpoints)
    
    return cfg, nil
}
```

#### 2.5.2 新增 resolveRouteDecision()

```go
// handlers/api_proxy_routing.go
func resolveRouteDecision(
    providerCfg providerRuntimeConfig,
    merchantID int,
    merchantType string,
    merchantRegion string,
) *services.RouteDecision {
    // 如果数据库配置为空，降级到环境变量
    if providerCfg.RouteStrategy == nil || len(providerCfg.RouteStrategy) == 0 {
        return &services.RouteDecision{
            Mode:     determineGatewayModeFromEnv(),
            Endpoint: resolveEndpointFromEnv(),
            Reason:   "fallback to environment variable",
        }
    }
    
    // 使用 UnifiedRouter 决策
    router := services.NewUnifiedRouter(nil)
    decision, err := router.DecideRoute(
        context.Background(),
        &services.ProviderConfig{
            Code:           providerCfg.Code,
            ProviderRegion: providerCfg.ProviderRegion,
            RouteStrategy:  providerCfg.RouteStrategy,
            Endpoints:      providerCfg.Endpoints,
        },
        &services.MerchantConfig{
            ID:     merchantID,
            Type:   merchantType,
            Region: merchantRegion,
        },
    )
    
    if err != nil {
        return &services.RouteDecision{
            Mode:     "direct",
            Endpoint: providerCfg.APIBaseURL,
            Reason:   "error in route decision, fallback to direct",
        }
    }
    
    return decision
}

func determineGatewayModeFromEnv() string {
    active := strings.TrimSpace(strings.ToLower(os.Getenv("LLM_GATEWAY_ACTIVE")))
    if active == "" || active == "none" {
        return "direct"
    }
    return active
}

func resolveEndpointFromEnv() string {
    mode := determineGatewayModeFromEnv()
    switch mode {
    case "litellm":
        if base := os.Getenv("LLM_GATEWAY_LITELLM_URL"); base != "" {
            return strings.TrimRight(base, "/") + "/v1"
        }
    case "proxy":
        return os.Getenv("LLM_GATEWAY_PROXY_URL")
    }
    return ""
}

func getMerchantRouteInfo(db *sql.DB, merchantID int) (merchantType string, merchantRegion string) {
    merchantType = "regular"
    merchantRegion = "domestic"
    
    if merchantID <= 0 {
        return
    }
    
    var mType, mRegion sql.NullString
    err := db.QueryRow(
        `SELECT COALESCE(type, 'regular'), COALESCE(region, 'domestic')
         FROM merchants WHERE id = $1 LIMIT 1`,
        merchantID,
    ).Scan(&mType, &mRegion)
    
    if err == nil {
        if mType.Valid {
            merchantType = mType.String
        }
        if mRegion.Valid {
            merchantRegion = mRegion.String
        }
    }
    
    return
}
```

#### 2.5.3 新增路由配置缓存

```go
// services/route_cache.go
package services

import (
    "sync"
    "time"
)

type RouteCache struct {
    cache map[string]*CachedRouteConfig
    mu    sync.RWMutex
    ttl   time.Duration
}

type CachedRouteConfig struct {
    Config    providerRuntimeConfig
    ExpiredAt time.Time
}

func NewRouteCache(ttl time.Duration) *RouteCache {
    return &RouteCache{
        cache: make(map[string]*CachedRouteConfig),
        ttl:   ttl,
    }
}

func (c *RouteCache) Get(providerCode string) (*providerRuntimeConfig, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    if cached, ok := c.cache[providerCode]; ok {
        if time.Now().Before(cached.ExpiredAt) {
            return &cached.Config, true
        }
    }
    return nil, false
}

func (c *RouteCache) Set(providerCode string, cfg providerRuntimeConfig) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    c.cache[providerCode] = &CachedRouteConfig{
        Config:    cfg,
        ExpiredAt: time.Now().Add(c.ttl),
    }
}

// Invalidate 使指定 provider 的缓存失效
func (c *RouteCache) Invalidate(providerCode string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    delete(c.cache, providerCode)
}

// InvalidateAll 使所有缓存失效
func (c *RouteCache) InvalidateAll() {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    c.cache = make(map[string]*CachedRouteConfig)
}

// 全局缓存实例
var globalRouteCache *RouteCache
var routeCacheOnce sync.Once

func GetRouteCache() *RouteCache {
    routeCacheOnce.Do(func() {
        globalRouteCache = NewRouteCache(5 * time.Minute) // 默认 5 分钟 TTL
    })
    return globalRouteCache
}
```

#### 2.5.3.1 缓存失效集成

```go
// handlers/route_config.go - 在管理端更新配置时使缓存失效
func UpdateProviderRouteConfig(c *gin.Context) {
    // ... 更新数据库逻辑 ...
    
    // 使缓存失效
    services.GetRouteCache().Invalidate(providerCode)
    
    c.JSON(http.StatusOK, gin.H{"message": "Route config updated"})
}
```

#### 2.5.4 修改主流程集成配置驱动

```go
// handlers/api_proxy.go
func proxyAPIRequestCore(...) {
    // ... 前置逻辑 ...
    
    // 获取完整的 Provider 配置（包含路由策略）
    providerCfg, err := getProviderRuntimeConfig(db, req.Provider)
    
    // 获取商户信息
    merchantType, merchantRegion := getMerchantRouteInfo(db, merchantID)
    
    // 【新增】配置驱动路由决策
    routeDecision := resolveRouteDecision(providerCfg, merchantID, merchantType, merchantRegion)
    
    // 使用决策结果
    endpointURL := routeDecision.Endpoint
    if endpointURL == "" {
        endpointURL = providerCfg.APIBaseURL  // 降级
    }
    
    // 构建 ExecutionLayerInput
    execLayerInput := &services.ExecutionLayerInput{
        ProviderConfig: &services.ExecutionProviderConfig{
            Code:           providerCfg.Code,
            APIBaseURL:     endpointURL,
            APIFormat:      providerCfg.APIFormat,
            GatewayMode:    routeDecision.Mode,
            RouteStrategy:  providerCfg.RouteStrategy,
            Endpoints:      providerCfg.Endpoints,
        },
        // ...
    }
    
    // 调用 ExecutionLayer
    // ...
}
```

**预计工作量**：4-5 小时
**风险等级**：中（核心改动）

---

### Phase 3: 主流程切换（高风险）

**目标**：将主流程从直接调用改为通过 ExecutionLayer 调用

#### 3.1 灰度开关

```go
// 通过环境变量控制是否使用新流程
func shouldUseExecutionLayer() bool {
    enabled := strings.TrimSpace(strings.ToLower(os.Getenv("USE_EXECUTION_LAYER")))
    return enabled == "true" || enabled == "1"
}

// 控制是否使用配置驱动路由
func shouldUseConfigDrivenRouting() bool {
    enabled := strings.TrimSpace(strings.ToLower(os.Getenv("USE_CONFIG_DRIVEN_ROUTING")))
    return enabled == "true" || enabled == "1"
}
```

#### 3.2 监控指标

- 新旧流程成功率对比
- 延迟对比
- 错误类型分布
- 配置驱动 vs 环境变量决策对比

**预计工作量**：6-8 小时
**风险等级**：高（需要灰度发布和监控）

---

### Phase 4: 清理与优化（低风险）

**目标**：移除旧代码，优化架构

#### 4.1 移除冗余代码

- 移除 `executeProviderRequestWithRetry()`（已被 ExecutionEngine 替代）
- 移除 `applyGatewayOverride()`（已集成到配置驱动路由）
- 移除 `resolveGatewayAuthToken()`（已集成到 ExecutionLayer）

#### 4.2 更新文档

- 更新架构文档
- 更新管理端配置说明
- 更新运维手册

**预计工作量**：2-3 小时
**风险等级**：低

---

## 三、详细任务清单

### Phase 1: 基础设施准备

| ID | 任务 | 文件 | 优先级 | 预计时间 |
|----|------|------|--------|----------|
| P1.1 | 扩展 ExecutionProviderConfig 结构体 | `services/execution_layer.go` | P0 | 0.5h |
| P1.2 | 扩展 ExecutionInput 结构体 | `services/execution_engine.go` | P0 | 0.5h |
| P1.3 | 新增 resolveEndpoint 函数 | `services/execution_layer.go` | P0 | 1h |
| P1.4 | 新增 resolveAuthToken 函数 | `services/execution_layer.go` | P0 | 1h |
| P1.5 | 新增 determineGatewayMode 函数 | `services/execution_layer.go` | P0 | 0.5h |
| P1.6 | 单元测试：resolveEndpoint | `services/execution_layer_test.go` | P1 | 0.5h |
| P1.7 | 单元测试：resolveAuthToken | `services/execution_layer_test.go` | P1 | 0.5h |
| P1.8 | 单元测试：determineGatewayMode | `services/execution_layer_test.go` | P1 | 0.5h |

### Phase 2: 执行层能力增强

| ID | 任务 | 文件 | 优先级 | 预计时间 |
|----|------|------|--------|----------|
| P2.1 | 增强 ExecutionLayer.Execute() | `services/execution_layer.go` | P0 | 2h |
| P2.2 | 新增 ExecutionEngine.ExecuteStream() | `services/execution_engine.go` | P0 | 1.5h |
| P2.3 | 新增 ExecutionEngine.WithRetryPolicy() | `services/execution_engine.go` | P1 | 0.5h |
| P2.4 | 集成测试：Execute Direct | `services/execution_layer_test.go` | P0 | 1h |
| P2.5 | 集成测试：Execute Litellm | `services/execution_layer_test.go` | P0 | 1h |
| P2.6 | 集成测试：Execute Stream | `services/execution_layer_test.go` | P0 | 1h |

### Phase 2.5: 🔴 配置驱动路由集成

| ID | 任务 | 文件 | 优先级 | 预计时间 |
|----|------|------|--------|----------|
| P2.5.1 | 扩展 getProviderRuntimeConfig 读取路由配置 | `handlers/api_proxy_http.go` | P0 | 1h |
| P2.5.2 | 新增 resolveRouteDecision 函数 | `handlers/api_proxy_routing.go` | P0 | 1h |
| P2.5.3 | 新增 resolveEndpointFromEnv 函数 | `handlers/api_proxy_routing.go` | P0 | 0.5h |
| P2.5.4 | 新增 getMerchantRouteInfo 函数 | `handlers/api_proxy_routing.go` | P0 | 0.5h |
| P2.5.5 | 新增 RouteCache 缓存层（含失效机制） | `services/route_cache.go` | P1 | 1h |
| P2.5.6 | 集成缓存失效到管理端 API | `handlers/route_config.go` | P1 | 0.5h |
| P2.5.7 | 修改主流程集成配置驱动 | `handlers/api_proxy.go` | P0 | 1.5h |
| P2.5.8 | 集成测试：配置驱动路由 | `handlers/api_proxy_test.go` | P0 | 1h |
| P2.5.9 | 集成测试：降级到环境变量 | `handlers/api_proxy_test.go` | P1 | 0.5h |
| P2.5.10 | 集成测试：缓存失效机制 | `services/route_cache_test.go` | P1 | 0.5h |

### Phase 3: 主流程切换

| ID | 任务 | 文件 | 优先级 | 预计时间 |
|----|------|------|--------|----------|
| P3.1 | 新增 shouldUseExecutionLayer 开关 | `handlers/api_proxy.go` | P0 | 0.5h |
| P3.2 | 新增 shouldUseConfigDrivenRouting 开关 | `handlers/api_proxy.go` | P0 | 0.5h |
| P3.3 | 新增 executeViaExecutionLayer 函数 | `handlers/api_proxy.go` | P0 | 2h |
| P3.4 | 修改 proxyAPIRequestCore 支持双流程 | `handlers/api_proxy.go` | P0 | 2h |
| P3.5 | 新增监控指标 | `handlers/api_proxy.go` | P1 | 1h |
| P3.6 | 灰度发布配置 | `.env`, `docker-compose.yml` | P0 | 0.5h |
| P3.7 | 生产环境灰度验证 | - | P0 | 2h |
| P3.8 | E2E 测试：配置驱动路由场景 | `e2e/routing_test.go` | P0 | 1.5h |
| P3.9 | E2E 测试：降级机制验证 | `e2e/routing_test.go` | P1 | 1h |
| P3.10 | 性能基准测试 | `services/execution_layer_test.go` | P1 | 1h |

### Phase 4: 清理与优化

| ID | 任务 | 文件 | 优先级 | 预计时间 |
|----|------|------|--------|----------|
| P4.1 | 移除 executeProviderRequestWithRetry | `handlers/api_proxy_http.go` | P1 | 0.5h |
| P4.2 | 移除 applyGatewayOverride | `handlers/api_proxy_http.go` | P1 | 0.5h |
| P4.3 | 移除 resolveGatewayAuthToken | `handlers/api_proxy_http.go` | P1 | 0.5h |
| P4.4 | 更新架构文档 | `docs/llm-request-flow.md` | P2 | 0.5h |
| P4.5 | 更新管理端配置说明 | `docs/` | P2 | 0.5h |
| P4.6 | 移除灰度开关，强制新流程 | `handlers/api_proxy.go` | P0 | 0.5h |
| P4.7 | 管理端前端：路由配置 JSON 编辑器 | `frontend/src/pages/` | P2 | 2h |
| P4.8 | 管理端前端：配置验证功能 | `frontend/src/pages/` | P2 | 1h |
| P4.9 | 管理端前端：配置预览功能 | `frontend/src/pages/` | P2 | 1h |

---

## 四、监控与可观测性

### 4.1 核心监控指标

```go
// services/route_metrics.go
package services

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // 路由决策计数器
    RouteDecisionCounter = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "route_decision_total",
            Help: "Total number of route decisions",
        },
        []string{"provider", "mode", "merchant_type", "region"},
    )
    
    // 执行层延迟直方图
    ExecutionLayerLatency = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "execution_layer_latency_ms",
            Help:    "Execution layer latency in milliseconds",
            Buckets: []float64{10, 50, 100, 200, 500, 1000, 2000},
        },
        []string{"provider", "mode", "success"},
    )
    
    // 配置驱动 vs 环境变量决策对比
    RouteDecisionSourceCounter = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "route_decision_source_total",
            Help: "Route decision source comparison",
        },
        []string{"source"}, // "config_driven", "env_fallback", "direct_fallback"
    )
    
    // 缓存命中率
    RouteCacheResultCounter = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "route_cache_result_total",
            Help: "Route cache hit/miss count",
        },
        []string{"result"}, // "hit", "miss"
    )
    
    // 执行层错误计数
    ExecutionLayerErrors = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "execution_layer_errors_total",
            Help: "Execution layer error count",
        },
        []string{"provider", "mode", "error_type"},
    )
)
```

### 4.2 指标埋点位置

```go
// 在 resolveRouteDecision 中埋点
func resolveRouteDecision(...) *services.RouteDecision {
    startTime := time.Now()
    
    decision := &services.RouteDecision{...}
    
    // 记录路由决策
    services.RouteDecisionCounter.WithLabelValues(
        providerCfg.Code,
        decision.Mode,
        merchantType,
        merchantRegion,
    ).Inc()
    
    // 记录决策来源
    source := "config_driven"
    if decision.Reason == "fallback to environment variable" {
        source = "env_fallback"
    } else if decision.Reason == "fallback to direct" {
        source = "direct_fallback"
    }
    services.RouteDecisionSourceCounter.WithLabelValues(source).Inc()
    
    return decision
}

// 在 ExecutionLayer.Execute 中埋点
func (l *ExecutionLayer) Execute(ctx context.Context, input *ExecutionLayerInput) (*ExecutionLayerOutput, error) {
    startTime := time.Now()
    
    result, err := l.engine.ExecuteWithRetry(ctx, execInput)
    
    // 记录延迟
    latency := time.Since(startTime).Milliseconds()
    services.ExecutionLayerLatency.WithLabelValues(
        input.ProviderConfig.Code,
        input.ProviderConfig.GatewayMode,
        fmt.Sprintf("%v", err == nil),
    ).Observe(float64(latency))
    
    // 记录错误
    if err != nil {
        services.ExecutionLayerErrors.WithLabelValues(
            input.ProviderConfig.Code,
            input.ProviderConfig.GatewayMode,
            categorizeError(err),
        ).Inc()
    }
    
    return result, err
}

// 在 RouteCache 中埋点
func (c *RouteCache) Get(providerCode string) (*providerRuntimeConfig, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    if cached, ok := c.cache[providerCode]; ok {
        if time.Now().Before(cached.ExpiredAt) {
            services.RouteCacheResultCounter.WithLabelValues("hit").Inc()
            return &cached.Config, true
        }
    }
    services.RouteCacheResultCounter.WithLabelValues("miss").Inc()
    return nil, false
}
```

### 4.3 告警规则

```yaml
# prometheus/alerts.yml
groups:
  - name: routing_alerts
    rules:
      - alert: ExecutionLayerHighErrorRate
        expr: |
          sum(rate(execution_layer_errors_total[5m])) 
          / sum(rate(execution_layer_latency_ms_count[5m])) > 0.01
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "执行层错误率过高"
          description: "执行层错误率超过 1%，当前值: {{ $value }}"
      
      - alert: RouteDecisionHighFallbackRate
        expr: |
          sum(rate(route_decision_source_total{source!="config_driven"}[5m])) 
          / sum(rate(route_decision_source_total[5m])) > 0.1
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "路由决策降级率过高"
          description: "降级到环境变量的比例超过 10%"
      
      - alert: RouteCacheLowHitRate
        expr: |
          sum(rate(route_cache_result_total{result="hit"}[5m])) 
          / sum(rate(route_cache_result_total[5m])) < 0.8
        for: 15m
        labels:
          severity: warning
        annotations:
          summary: "路由缓存命中率过低"
          description: "缓存命中率低于 80%"
      
      - alert: ExecutionLayerHighLatency
        expr: |
          histogram_quantile(0.99, sum(rate(execution_layer_latency_ms_bucket[5m])) by (le)) > 200
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "执行层延迟过高"
          description: "P99 延迟超过 200ms"
```

### 4.4 Grafana Dashboard

```json
{
  "title": "三层路由架构监控",
  "panels": [
    {
      "title": "路由决策分布",
      "type": "piechart",
      "targets": [
        {
          "expr": "sum by (mode) (route_decision_total)"
        }
      ]
    },
    {
      "title": "执行层延迟 P99",
      "type": "graph",
      "targets": [
        {
          "expr": "histogram_quantile(0.99, sum(rate(execution_layer_latency_ms_bucket[5m])) by (le))"
        }
      ]
    },
    {
      "title": "决策来源对比",
      "type": "graph",
      "targets": [
        {
          "expr": "sum by (source) (rate(route_decision_source_total[5m]))"
        }
      ]
    },
    {
      "title": "缓存命中率",
      "type": "stat",
      "targets": [
        {
          "expr": "sum(rate(route_cache_result_total{result=\"hit\"}[5m])) / sum(rate(route_cache_result_total[5m]))"
        }
      ]
    }
  ]
}
```

---

## 五、端到端测试场景

### 5.1 E2E 测试用例

```go
// e2e/routing_test.go
package e2e

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// 场景 1: 国内普通用户 → LiteLLM
func TestE2E_DomesticRegularUser_RouteToLiteLLM(t *testing.T) {
    // 设置测试环境
    setup := setupTestEnvironment(t)
    defer setup.Teardown()
    
    // 配置商户为 domestic + regular
    setup.ConfigureMerchant(1, "regular", "domestic")
    
    // 配置 Provider 路由策略
    setup.ConfigureProvider("openai", map[string]interface{}{
        "route_strategy": map[string]interface{}{
            "domestic_users": map[string]interface{}{
                "mode": "litellm",
            },
        },
        "endpoints": map[string]interface{}{
            "litellm": map[string]interface{}{
                "domestic": "http://litellm:4000/v1",
            },
        },
    })
    
    // 发送请求
    resp, err := setup.SendProxyRequest(&ProxyRequest{
        Provider: "openai",
        Model:    "gpt-4",
        Messages: []Message{{Role: "user", Content: "test"}},
    })
    
    require.NoError(t, err)
    assert.Equal(t, 200, resp.StatusCode)
    
    // 验证路由到 LiteLLM
    assert.True(t, setup.WasRequestRoutedTo("litellm"))
}

// 场景 2: 海外普通用户 → Direct
func TestE2E_OverseasRegularUser_RouteToDirect(t *testing.T) {
    setup := setupTestEnvironment(t)
    defer setup.Teardown()
    
    setup.ConfigureMerchant(1, "regular", "overseas")
    setup.ConfigureProvider("openai", map[string]interface{}{
        "route_strategy": map[string]interface{}{
            "overseas_users": map[string]interface{}{
                "mode": "direct",
            },
        },
        "endpoints": map[string]interface{}{
            "direct": map[string]interface{}{
                "overseas": "https://api.openai.com/v1",
            },
        },
    })
    
    resp, err := setup.SendProxyRequest(&ProxyRequest{
        Provider: "openai",
        Model:    "gpt-4",
        Messages: []Message{{Role: "user", Content: "test"}},
    })
    
    require.NoError(t, err)
    assert.True(t, setup.WasRequestRoutedTo("direct"))
}

// 场景 3: 企业用户 → LiteLLM + Fallback
func TestE2E_EnterpriseUser_FallbackToProxy(t *testing.T) {
    setup := setupTestEnvironment(t)
    defer setup.Teardown()
    
    setup.ConfigureMerchant(1, "enterprise", "domestic")
    setup.ConfigureProvider("openai", map[string]interface{}{
        "route_strategy": map[string]interface{}{
            "enterprise_users": map[string]interface{}{
                "mode":          "litellm",
                "fallback_mode": "proxy",
            },
        },
        "endpoints": map[string]interface{}{
            "litellm": map[string]interface{}{
                "domestic": "http://litellm:4000/v1",
            },
            "proxy": map[string]interface{}{
                "gaap": "https://proxy.example.com",
            },
        },
    })
    
    // 模拟 LiteLLM 故障
    setup.SimulateFailure("litellm", http.StatusServiceUnavailable)
    
    resp, err := setup.SendProxyRequest(&ProxyRequest{
        Provider: "openai",
        Model:    "gpt-4",
        Messages: []Message{{Role: "user", Content: "test"}},
    })
    
    require.NoError(t, err)
    // 验证降级到 proxy
    assert.True(t, setup.WasRequestRoutedTo("proxy"))
}

// 场景 4: 配置缺失 → 环境变量降级
func TestE2E_ConfigMissing_FallbackToEnv(t *testing.T) {
    setup := setupTestEnvironment(t)
    defer setup.Teardown()
    
    // 清空数据库配置
    setup.ClearProviderConfig("openai")
    
    // 设置环境变量
    setup.SetEnv("LLM_GATEWAY_ACTIVE", "litellm")
    setup.SetEnv("LLM_GATEWAY_LITELLM_URL", "http://litellm:4000")
    
    resp, err := setup.SendProxyRequest(&ProxyRequest{
        Provider: "openai",
        Model:    "gpt-4",
        Messages: []Message{{Role: "user", Content: "test"}},
    })
    
    require.NoError(t, err)
    // 验证使用环境变量配置
    assert.True(t, setup.WasRequestRoutedTo("litellm"))
    assert.Equal(t, "env_fallback", setup.GetDecisionSource())
}

// 场景 5: 流式请求
func TestE2E_StreamingRequest(t *testing.T) {
    setup := setupTestEnvironment(t)
    defer setup.Teardown()
    
    setup.ConfigureMerchant(1, "regular", "domestic")
    setup.ConfigureProvider("openai", map[string]interface{}{
        "route_strategy": map[string]interface{}{
            "domestic_users": map[string]interface{}{"mode": "litellm"},
        },
    })
    
    // 发送流式请求
    stream, err := setup.SendStreamingRequest(&ProxyRequest{
        Provider: "openai",
        Model:    "gpt-4",
        Messages: []Message{{Role: "user", Content: "test"}},
        Stream:   true,
    })
    
    require.NoError(t, err)
    
    // 验证 SSE 响应
    chunks := 0
    for chunk := range stream {
        chunks++
        assert.NotEmpty(t, chunk.Data)
    }
    assert.Greater(t, chunks, 0)
}

// 场景 6: 高并发
func TestE2E_HighConcurrency(t *testing.T) {
    setup := setupTestEnvironment(t)
    defer setup.Teardown()
    
    setup.ConfigureMerchant(1, "regular", "domestic")
    setup.ConfigureProvider("openai", map[string]interface{}{
        "route_strategy": map[string]interface{}{
            "domestic_users": map[string]interface{}{"mode": "litellm"},
        },
    })
    
    // 并发 100 个请求
    concurrency := 100
    errors := make(chan error, concurrency)
    latencies := make(chan time.Duration, concurrency)
    
    for i := 0; i < concurrency; i++ {
        go func() {
            start := time.Now()
            _, err := setup.SendProxyRequest(&ProxyRequest{
                Provider: "openai",
                Model:    "gpt-4",
                Messages: []Message{{Role: "user", Content: "test"}},
            })
            latencies <- time.Since(start)
            errors <- err
        }()
    }
    
    // 统计结果
    errorCount := 0
    totalLatency := time.Duration(0)
    for i := 0; i < concurrency; i++ {
        if err := <-errors; err != nil {
            errorCount++
        }
        totalLatency += <-latencies
    }
    
    successRate := float64(concurrency-errorCount) / float64(concurrency)
    avgLatency := totalLatency / time.Duration(concurrency)
    
    assert.GreaterOrEqual(t, successRate, 0.99, "成功率应 >= 99%")
    assert.Less(t, avgLatency, 200*time.Millisecond, "平均延迟应 < 200ms")
}
```

### 5.2 性能基准测试

```go
// services/execution_layer_bench_test.go
package services

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
)

// 基准测试：路由决策延迟
func BenchmarkResolveRouteDecision(b *testing.B) {
    providerCfg := &providerRuntimeConfig{
        Code:           "openai",
        ProviderRegion: "overseas",
        RouteStrategy: map[string]interface{}{
            "domestic_users": map[string]interface{}{"mode": "litellm"},
            "overseas_users": map[string]interface{}{"mode": "direct"},
        },
        Endpoints: map[string]interface{}{
            "direct":  map[string]interface{}{"overseas": "https://api.openai.com/v1"},
            "litellm": map[string]interface{}{"domestic": "http://litellm:4000/v1"},
        },
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        resolveRouteDecision(*providerCfg, 1, "regular", "domestic")
    }
}

// 基准测试：执行层延迟
func BenchmarkExecutionLayer_Execute(b *testing.B) {
    // 模拟上游服务器
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"id":"test","choices":[],"usage":{"prompt_tokens":10,"completion_tokens":5}}`))
    }))
    defer server.Close()
    
    layer := NewExecutionLayer(nil, NewExecutionEngine())
    
    input := &ExecutionLayerInput{
        ProviderConfig: &ExecutionProviderConfig{
            Code:        "openai",
            APIBaseURL:  server.URL,
            APIFormat:   "openai",
            GatewayMode: "direct",
        },
        DecryptedAPIKey: "test-key",
        RequestBody:      []byte(`{"model":"gpt-4","messages":[{"role":"user","content":"test"}]}`),
        Stream:           false,
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        layer.Execute(context.Background(), input)
    }
}

// 基准测试：缓存读取
func BenchmarkRouteCache_Get(b *testing.B) {
    cache := NewRouteCache(5 * time.Minute)
    cfg := providerRuntimeConfig{Code: "openai"}
    cache.Set("openai", cfg)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        cache.Get("openai")
    }
}

// 基准测试：缓存写入
func BenchmarkRouteCache_Set(b *testing.B) {
    cache := NewRouteCache(5 * time.Minute)
    cfg := providerRuntimeConfig{Code: "openai"}
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        cache.Set("openai", cfg)
    }
}
```

### 5.3 性能目标

| 指标 | 目标值 | 说明 |
|------|--------|------|
| 路由决策延迟 P99 | < 5ms | 不含数据库查询 |
| 执行层延迟 P99 | < 10ms | 不含上游响应 |
| 缓存命中率 | > 90% | 稳定运行后 |
| 并发成功率 | > 99% | 100 并发请求 |
| 平均延迟 | < 200ms | 含上游响应 |

---

## 六、风险与缓解措施

### 风险矩阵

| 风险 | 可能性 | 影响 | 缓解措施 |
|------|--------|------|----------|
| 新流程性能下降 | 中 | 高 | 灰度发布，性能对比监控 |
| 流式响应异常 | 中 | 高 | 充分测试流式场景 |
| Gateway 切换失败 | 低 | 高 | 保留回滚开关 |
| 认证 Token 错误 | 低 | 高 | 单元测试覆盖所有场景 |
| 配置驱动决策错误 | 中 | 高 | 降级到环境变量机制 |
| 数据库配置缺失 | 低 | 中 | 默认值 + 环境变量降级 |

### 回滚方案

```bash
# 随时可以通过环境变量回滚
USE_EXECUTION_LAYER=false           # 回滚到旧流程
USE_CONFIG_DRIVEN_ROUTING=false     # 回滚到环境变量驱动
```

---

## 七、验收标准

### Phase 1 验收

- [ ] 所有新增函数有单元测试覆盖
- [ ] 测试覆盖率 > 80%
- [ ] CI 通过

### Phase 2 验收

- [ ] ExecutionLayer 能完整处理 Direct/Litellm/Proxy 三种模式
- [ ] 流式响应正常工作
- [ ] 集成测试通过

### Phase 2.5 验收 🔴

- [ ] 数据库配置能正确读取
- [ ] UnifiedRouter.DecideRoute() 返回正确决策
- [ ] 配置缺失时降级到环境变量
- [ ] 管理端修改配置后实时生效
- [ ] 缓存机制正常工作

### Phase 3 验收

- [ ] 灰度发布无异常
- [ ] 新旧流程成功率差异 < 0.1%
- [ ] 延迟差异 < 5ms
- [ ] 监控指标正常

### Phase 4 验收

- [ ] 旧代码已移除
- [ ] 文档已更新
- [ ] 生产环境稳定运行 7 天

---

## 八、时间规划

| 阶段 | 预计时间 | 开始日期 | 结束日期 |
|------|----------|----------|----------|
| Phase 1 | 5h | - | - |
| Phase 2 | 7h | - | - |
| **Phase 2.5** | **8h** | - | - |
| Phase 3 | **11.5h** | - | - |
| Phase 4 | **7h** | - | - |
| **总计** | **38.5h** | - | - |

建议分 5-6 个迭代完成，每个迭代 1-2 天。

---

## 九、依赖关系

```
Phase 1 (基础设施)
    │
    ├── P1.1-P1.5 (结构体扩展)
    │       │
    │       └── P1.6-P1.8 (单元测试)
    │
    ▼
Phase 2 (能力增强)
    │
    ├── P2.1 (Execute 增强)
    │       │
    │       ├── P2.2 (ExecuteStream)
    │       │
    │       └── P2.4-P2.6 (集成测试)
    │
    ▼
Phase 2.5 (配置驱动路由) 🔴
    │
    ├── P2.5.1 (扩展 getProviderRuntimeConfig)
    │       │
    │       ├── P2.5.2 (resolveRouteDecision)
    │       │       │
    │       │       └── P2.5.3 (resolveEndpointFromEnv)
    │       │
    │       ├── P2.5.4 (getMerchantRouteInfo)
    │       │
    │       ├── P2.5.5 (RouteCache 缓存层)
    │       │       │
    │       │       └── P2.5.6 (集成缓存失效)
    │       │
    │       └── P2.5.7 (主流程集成)
    │               │
    │               ├── P2.5.8 (测试：配置驱动)
    │               │
    │               ├── P2.5.9 (测试：降级)
    │               │
    │               └── P2.5.10 (测试：缓存失效)
    │
    ▼
Phase 3 (主流程切换)
    │
    ├── P3.1-P3.2 (灰度开关)
    │       │
    │       ├── P3.3 (新流程函数)
    │       │
    │       ├── P3.4 (双流程支持)
    │       │
    │       └── P3.6-P3.7 (灰度发布)
    │
    ▼
Phase 4 (清理优化)
    │
    ├── P4.1-P4.3 (移除旧代码)
    │
    ├── P4.4-P4.5 (文档更新)
    │
    └── P4.6 (强制新流程)
```

---

## 十、管理端配置说明

### 已存在的管理端 API

| API | 方法 | 说明 |
|-----|------|------|
| `/api/v1/admin/route-config/providers` | GET | 获取所有厂商路由配置 |
| `/api/v1/admin/route-config/providers/:code` | GET | 获取单个厂商路由配置 |
| `/api/v1/admin/route-config/providers/:code` | PUT | 更新厂商路由配置 |
| `/api/v1/admin/route-config/merchants` | GET | 获取商户路由配置 |
| `/api/v1/admin/route-config/merchants/:id` | PUT | 更新商户路由配置 |
| `/api/v1/admin/route-config/test` | POST | 测试路由决策 |

### 配置示例

```json
{
  "provider_region": "overseas",
  "route_strategy": {
    "default_mode": "auto",
    "domestic_users": {
      "mode": "litellm",
      "fallback_mode": "proxy"
    },
    "overseas_users": {
      "mode": "direct"
    }
  },
  "endpoints": {
    "direct": {
      "overseas": "https://api.openai.com/v1"
    },
    "litellm": {
      "domestic": "http://litellm:4000/v1"
    },
    "proxy": {
      "gaap": "https://openai-gaap.example.com"
    }
  }
}
```

### 环境变量配置

#### .env 新增变量

```bash
# === 执行层统一出站开关 ===
# 灰度期间通过这两个开关控制新旧流程
USE_EXECUTION_LAYER=false           # 是否使用新流程（默认 false）
USE_CONFIG_DRIVEN_ROUTING=false     # 是否使用配置驱动路由（默认 false）

# === 保留：降级选项 ===
# 当 USE_CONFIG_DRIVEN_ROUTING=false 或数据库配置缺失时使用
LLM_GATEWAY_ACTIVE=litellm          # none | litellm
LLM_GATEWAY_LITELLM_URL=http://litellm:4000
LITELLM_MASTER_KEY=your_key_here
```

#### 配置优先级

```
1. 数据库配置 (model_providers.route_strategy + endpoints)
      ↓ 配置缺失或无效
2. 环境变量 LLM_GATEWAY_ACTIVE
      ↓ 未配置
3. 直连 Provider (direct)
```

#### LiteLLM 配置文件

`deploy/litellm/litellm_proxy_config.yaml` **无需修改**：
- 该文件是 LiteLLM 网关自身的配置
- 定义 `model_list` 和 `router_settings`
- 后端通过 `LLM_GATEWAY_LITELLM_URL` 调用
- 与本次后端路由决策演进无关

---

## 十一、后续优化方向

完成本次演进后，可以考虑：

1. **策略层增强**：支持更多路由策略（成本优先、延迟优先等）
2. **决策层优化**：引入机器学习模型预测最优 Key
3. **执行层扩展**：支持更多 Gateway（如 AWS Bedrock、Azure OpenAI）
4. **可观测性**：集成 OpenTelemetry 全链路追踪
5. **动态配置**：管理端实时修改配置，无需重启服务
6. **A/B 测试**：支持路由策略 A/B 测试
