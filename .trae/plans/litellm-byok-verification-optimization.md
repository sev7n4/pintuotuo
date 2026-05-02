# LiteLLM BYOK 验证优化开发计划

> 版本: v1.1
> 日期: 2026-05-02
> 状态: 待审批

---

## 一、背景与问题

### 1.1 当前架构

项目使用 LiteLLM 作为 LLM 网关，支持 BYOK（Bring Your Own Key）模式，允许商户使用自己的 API Key 访问大模型服务。

```
当前架构（国内部署）:

国内服务器 ──► LiteLLM（国内）──► 上游厂商
     │              │                    │
     │              │                    └── 可能受限制
     │              │
     └──────────────┴──► 直连上游 ──► 可能受限制
```

### 1.2 遇到的问题

在实施 LiteLLM BYOK 模式的深度验证时，发现以下问题：

| 问题 | 描述 | 根因 |
|------|------|------|
| 通配符模型认证失败 | `openai/*` 通配符模型无法正确传递 `api_key` 和 `api_base` | LiteLLM 对 `openai/*` 通配符的 BYOK 支持不完善 |
| `/models` 返回通配符 | LiteLLM `/models` 端点返回 `openai/*` 而不是实际模型列表 | LiteLLM 设计如此，通配符作为模型名称返回 |
| 深度验证超时 | 通过 LiteLLM 发送验证请求时超时 | 认证参数传递机制与 OpenAI 兼容提供商不匹配 |

### 1.3 未来架构规划

```
未来架构（海外 LiteLLM）:

国内服务器 ──► LiteLLM（海外）──► 上游厂商
     │              │                    │
     │              │                    └── 无限制 ✓
     │              │
     └──────────────┴──► 直连上游 ──► 受限制 ✗
```

**关键需求**：验证功能需要通过海外 LiteLLM 节点，避开网络限制。

---

## 二、调研结论

### 2.1 LiteLLM BYOK 支持方式

| 方式 | 配置 | 传递方式 | 适用场景 | 状态 |
|------|------|----------|----------|------|
| `configurable_clientside_auth_params` | 在 `model_list` 中配置 | 请求体 `extra_body` | 动态 API base | ⚠️ 通配符支持不完善 |
| `forward_llm_provider_auth_headers` | 在 `general_settings` 中配置 | 请求头 | Anthropic/Azure/Google | ✅ 可用 |
| `user_config` | 在请求体中传递完整配置 | 请求体 `extra_body` | 完全动态配置 | ✅ 可用 |

### 2.2 关键发现

#### 2.2.1 通配符模型问题

**Issue #13752**: Wildcard entries appear as models in the /models endpoint

```yaml
配置:
  model_list:
    - model_name: "openai/*"
      litellm_params:
        model: "openai/*"
        configurable_clientside_auth_params: ["api_key", "api_base"]

/models 返回:
  [
    {"id": "openai/*"},      ← 通配符作为模型名称返回
    ...
  ]
```

**结论**: LiteLLM 对 `openai/*` 通配符的 BYOK 支持存在兼容性问题。

#### 2.2.2 `/models` 端点行为

| 问题 | 结论 |
|------|------|
| 是否支持 pass-through `/models`？ | ❌ 不支持 |
| `/models` 返回什么？ | 返回 LiteLLM 配置文件中定义的模型列表 |
| 通配符模型如何处理？ | 直接返回通配符名称，不会展开为实际模型 |

#### 2.2.3 `user_config` 方式

根据 LiteLLM 官方文档，可以通过 `extra_body` 传递完整的模型配置：

```python
user_config = {
    'model_list': [
        {
            'model_name': 'dynamic-model',
            'litellm_params': {
                'model': 'openai/stepfun-step-1-8k',
                'api_key': '商户-API-Key',
                'api_base': 'https://api.stepfun.com/v1'
            }
        }
    ]
}

response = client.chat.completions.create(
    model="dynamic-model",
    messages=[...],
    extra_body={"user_config": user_config}
)
```

**结论**: `user_config` 方式是 LiteLLM 官方支持的动态配置方式，可用于深度验证。

---

## 三、需求分析

### 3.1 功能需求

| 功能 | 当前状态 | 目标状态 |
|------|----------|----------|
| 轻量验证 | 直连上游，可能受网络限制 | 直连上游，失败时降级处理 |
| 深度验证 | 通过 LiteLLM，但认证失败 | 通过 LiteLLM + user_config，正常工作 |
| 立即探测 | 通过 LiteLLM，但认证失败 | 通过 LiteLLM + user_config，正常工作 |
| 生产请求 | 通过 LiteLLM，使用 configurable_clientside_auth_params | 评估是否需要改为 user_config |

### 3.2 非功能需求

| 需求 | 描述 |
|------|------|
| 网络兼容 | 支持海外 LiteLLM 节点部署，避开网络限制 |
| 向后兼容 | 不影响现有的直连模式和 auto 模式 |
| 可观测性 | 记录验证过程中的网络状态和错误信息 |
| 降级处理 | 轻量验证失败时，提供合理的降级策略 |

---

## 四、技术方案

### 4.1 整体架构

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        完整架构（验证 + 生产）                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  国内服务器                                                                  │
│       │                                                                     │
│       │                                                                     │
│       ├──► 轻量验证 ──► 直连上游厂商 /models                                 │
│       │       │                   │                                         │
│       │       │                   ├── 成功：返回实际模型列表                  │
│       │       │                   └── 失败：降级处理，标记网络问题             │
│       │                                                                     │
│       │                                                                     │
│       ├──► 深度验证 ──► LiteLLM（海外）──► 上游厂商 /chat/completions        │
│       │       │                   │                                         │
│       │       │                   └── 使用 user_config 传递商户凭证          │
│       │                                                                     │
│       │                                                                     │
│       ├──► 立即探测 ──► LiteLLM（海外）──► 上游厂商 /chat/completions        │
│       │       │                   │                                         │
│       │       │                   └── 使用 user_config 传递商户凭证          │
│       │                                                                     │
│       │                                                                     │
│       └──► 生产请求 ──► LiteLLM（海外）──► 上游厂商 /chat/completions        │
│               │                   │                                         │
│               │                   └── 使用 configurable_clientside_auth     │
│               │                       或 user_config 传递商户凭证            │
│                                                                             │
│                                                                             │
│  上游厂商（OpenAI、Anthropic、StepFun 等）                                   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 4.2 生产请求实现评估

#### 4.2.1 当前实现

当前生产请求在 LiteLLM 模式下的实现（`api_proxy.go`）：

```go
// 流式请求和非流式请求都使用相同的逻辑
routeMode := resolveRouteMode(&pk)
if routeMode == routeModeLitellm {
    // 1. 模型名称添加 openai/ 前缀
    litellmModel := "openai/" + att.model
    rb["model"] = litellmModel
    
    // 2. 在请求体中传递 api_key 和 api_base
    rb["api_key"] = dk
    rb["api_base"] = apiBaseForUser
}

// 3. 认证使用 LITELLM_MASTER_KEY
authToken := resolveAuthTokenFromRouteMode(resolveRouteMode(&pk), dk)
hreq.Header.Set("Authorization", "Bearer "+authToken)
```

#### 4.2.2 问题分析

| 问题 | 影响 | 严重程度 |
|------|------|----------|
| 使用 `configurable_clientside_auth_params` 方式 | 通配符模型可能无法正确传递认证参数 | 高 |
| 模型名称使用 `openai/` 前缀 | 与 LiteLLM 配置中的通配符模型匹配 | 中 |
| 认证使用 `LITELLM_MASTER_KEY` | 正确，LiteLLM 需要代理认证 | ✓ 正常 |

#### 4.2.3 修改建议

**建议**：将生产请求也改为使用 `user_config` 方式，与深度验证保持一致。

```go
// 修改后的实现
routeMode := resolveRouteMode(&pk)
if routeMode == routeModeLitellm {
    // 1. 构建 user_config
    userConfig := map[string]interface{}{
        "model_list": []map[string]interface{}{
            {
                "model_name": "proxy-model",
                "litellm_params": map[string]interface{}{
                    "model":    "openai/" + att.model,
                    "api_key":  dk,
                    "api_base": apiBaseForUser,
                },
            },
        },
    }
    
    // 2. 设置模型名称和 user_config
    rb["model"] = "proxy-model"
    rb["user_config"] = userConfig
    
    // 3. 移除原来的 api_key 和 api_base
    // (不再需要，已在 user_config 中传递)
}
```

#### 4.2.4 修改影响评估

| 影响项 | 描述 | 风险 |
|--------|------|------|
| 向后兼容 | 不影响直连模式和 auto 模式 | 低 |
| 功能等价 | `user_config` 方式功能更完整 | 低 |
| 性能影响 | 请求体略大，影响可忽略 | 低 |
| 测试范围 | 需要重新测试所有 LiteLLM 模式的生产请求 | 中 |

### 4.3 轻量验证实现

```go
// 轻量验证 - 直连上游厂商
func (v *APIKeyValidator) lightVerify(ctx context.Context, provider, apiKey string) (*VerificationResult, error) {
    // 1. 获取上游厂商 base URL
    upstreamBaseURL := v.getProviderBaseURL(provider)
    
    // 2. 直连探测模型列表
    modelsURL := upstreamBaseURL + "/models"
    probe, err := ProbeProviderModels(ctx, client, modelsURL, apiKey)
    
    if err != nil || !probe.Success {
        // 3. 失败时降级处理
        return &VerificationResult{
            ModelsFound:     getPredefinedModels(provider),
            ConnectionTest:  false,
            ConnectionError: "无法连接上游厂商，可能是网络限制",
        }, nil
    }
    
    // 4. 成功时返回实际模型列表
    return &VerificationResult{
        ModelsFound:    probe.Models,
        ConnectionTest: true,
    }, nil
}
```

### 4.4 深度验证实现

```go
// 深度验证 - 通过 LiteLLM + user_config
func (v *APIKeyValidator) deepVerifyViaLitellm(
    ctx context.Context,
    litellmEndpoint string,
    provider string,
    model string,
    apiKey string,
    apiBase string,
    litellmMasterKey string,
) (bool, string, string) {
    
    // 1. 构建 user_config
    userConfig := map[string]interface{}{
        "model_list": []map[string]interface{}{
            {
                "model_name": "probe-model",
                "litellm_params": map[string]interface{}{
                    "model":    "openai/" + model,
                    "api_key":  apiKey,
                    "api_base": apiBase,
                },
            },
        },
    }
    
    // 2. 构建请求体
    body := map[string]interface{}{
        "model":      "probe-model",
        "messages":   []map[string]string{{"role": "user", "content": "ping"}},
        "max_tokens": 1,
        "user_config": userConfig,
    }
    
    // 3. 发送请求到 LiteLLM
    chatEndpoint := litellmEndpoint + "/chat/completions"
    req, _ := http.NewRequest("POST", chatEndpoint, jsonBody)
    req.Header.Set("Authorization", "Bearer "+litellmMasterKey)
    req.Header.Set("Content-Type", "application/json")
    
    // 4. 处理响应
    resp, err := client.Do(req)
    // ...
}
```

---

## 五、开发任务

### 5.1 后端任务

| 序号 | 任务 | 优先级 | 预估工时 |
|------|------|--------|----------|
| 1 | 修改轻量验证逻辑，添加降级处理 | 高 | 2h |
| 2 | 实现深度验证通过 LiteLLM + user_config | 高 | 4h |
| 3 | 修改立即探测逻辑，使用 LiteLLM + user_config | 高 | 2h |
| 4 | **修改生产请求，使用 user_config 方式** | 高 | 3h |
| 5 | 添加网络状态记录和错误信息 | 中 | 1h |
| 6 | 单元测试 | 高 | 2h |
| 7 | 集成测试 | 高 | 2h |

### 5.2 配置任务

| 序号 | 任务 | 优先级 | 预估工时 |
|------|------|--------|----------|
| 1 | 更新 LiteLLM 配置文件 | 中 | 0.5h |
| 2 | 添加环境变量配置 | 中 | 0.5h |

### 5.3 文档任务

| 序号 | 任务 | 优先级 | 预估工时 |
|------|------|--------|----------|
| 1 | 更新 API 文档 | 低 | 1h |
| 2 | 更新部署文档 | 低 | 1h |

---

## 六、测试计划

### 6.1 单元测试

| 测试项 | 描述 |
|--------|------|
| TestLightVerifyDirect | 测试轻量验证直连成功场景 |
| TestLightVerifyFallback | 测试轻量验证降级场景 |
| TestDeepVerifyViaLitellm | 测试深度验证通过 LiteLLM |
| TestProbeViaLitellm | 测试立即探测通过 LiteLLM |
| **TestProductionRequestViaLitellm** | **测试生产请求通过 LiteLLM + user_config** |

### 6.2 集成测试

| 测试项 | 描述 |
|--------|------|
| TestByokRoutingLitellmMode | 测试 BYOK 路由 LiteLLM 模式完整流程 |
| TestByokRoutingDirectMode | 测试 BYOK 路由直连模式完整流程 |
| **TestProductionProxyLitellmMode** | **测试生产代理 LiteLLM 模式完整流程** |

### 6.3 E2E 测试

| 测试项 | 描述 |
|--------|------|
| TestByokLightVerify | E2E 测试轻量验证 |
| TestByokDeepVerify | E2E 测试深度验证 |
| **TestProductionChatCompletions** | **E2E 测试生产请求聊天完成** |

---

## 七、风险评估

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| LiteLLM user_config 方式不稳定 | 高 | 低 | 充分测试，保留降级方案 |
| 海外 LiteLLM 节点延迟 | 中 | 中 | 添加超时配置，优化请求处理 |
| 网络问题导致验证失败 | 中 | 中 | 添加重试机制，记录详细错误信息 |
| **生产请求修改引入回归** | **高** | **中** | **充分测试，分阶段发布** |

---

## 八、里程碑

| 里程碑 | 内容 | 预计完成时间 |
|--------|------|--------------|
| M1 | 后端开发完成 | Day 1 |
| M2 | 测试完成 | Day 2 |
| M3 | 部署验证 | Day 3 |

---

## 九、参考资料

1. [LiteLLM Pass-Through Endpoints](https://docs.litellm.ai/docs/pass_through/intro)
2. [LiteLLM Client-Side Auth](https://docs.litellm.ai/docs/proxy/clientside_auth)
3. [LiteLLM Issue #13752: Wildcard entries in /models](https://github.com/BerriAI/litellm/issues/13752)
4. [LiteLLM Forward Client Headers](https://docs.litellm.ai/docs/proxy/forward_client_headers)

---

**审批记录**

| 日期 | 审批人 | 状态 | 备注 |
|------|--------|------|------|
| 2026-05-02 | - | 待审批 | - |

**变更记录**

| 版本 | 日期 | 变更内容 |
|------|------|----------|
| v1.0 | 2026-05-02 | 初始版本 |
| v1.1 | 2026-05-02 | 补充生产请求架构和评估，增加生产请求修改任务 |
