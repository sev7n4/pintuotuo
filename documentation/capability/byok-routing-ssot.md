# BYOK 路由：单一事实来源（SSOT）与代码路径

**SSOT 表**：`merchant_api_keys`  
**字段**：`route_mode`、`region`、`endpoint_url`、`fallback_endpoint_url`、`route_config`（JSONB）

`model_providers.endpoints` / `api_base_url` 表示**厂商级默认**（迁移种子、运营回填）。在 **ExecutionLayer** 中，仅当密钥侧未解析出出站基址时才作为回退使用。

---

## 1. 解析优先级（ExecutionLayer / capability-probe）

对 `services.ExecutionProviderConfig` 调用 `ConfigureGatewayMode` 后 `ResolveEndpoint` / `ResolveEndpointByType` 时：

1. **商户密钥（SSOT）** — `models.ResolveBYOKRoutingEndpoint(route_config, gatewayMode, region, endpoint_url 列)`  
   语义与 `MerchantAPIKey.GetEndpointForMode` 一致：`direct` 下 `endpoint_url` 列优先，其次 `route_config.endpoints.direct[region]`；`litellm` 用 `route_config.endpoints.litellm[region]` 等。
2. **`fallback_endpoint_url` 列** — 当第 1 步无结果且需要代理回退等场景。
3. **`model_providers.endpoints`** — 密钥未配置时的目录级默认（含 `litellm` / `direct` / `proxy` 分块）。

---

## 2. OpenAI 兼容代理（`handlers/api_proxy`）

`resolveEndpointURL` + `buildLitellmUserConfig` 路径：

- **direct / proxy**：出站 URL 仍来自**密钥行**（列 + `route_config`），与上表一致。
- **litellm**：出站 HTTP **必须**打到进程环境 **`LLM_GATEWAY_LITELLM_URL`**（+`/v1`），Bearer 为 **`LITELLM_MASTER_KEY`**；厂商 `api_base` 与商户 sk 通过 **user_config** 注入 LiteLLM，避免把 master key 误发到上游。

详见 `resolveEndpointURL` 注释与 `buildLitellmUserConfig`。

---

## 3. 健康检查 / 额度探测

`HealthChecker` 使用同一商户行解析（`resolveLitellmEndpoint` 等），并对历史占位主机名做 **`NormalizeLegacyLitellmGatewayBaseURL`**：

| URL 子串 | 映射到 |
|----------|--------|
| `litellm-domestic:` | `LLM_GATEWAY_LITELLM_URL` + `/v1` |
| `litellm-overseas:` | `LLM_GATEWAY_LITELLM_URL_OVERSEAS` + `/v1`，未设置则回退到 `LLM_GATEWAY_LITELLM_URL` |

显式写在 `route_config` 或 `model_providers` 中的完整 URL **不会被改写**。

---

## 4. 国内 / 海外双 LiteLLM 节点

- **国内**：`LLM_GATEWAY_LITELLM_URL`（默认 `http://litellm:4000`）。
- **海外**：`LLM_GATEWAY_LITELLM_URL_OVERSEAS`（可选）；未配置时海外占位仍回落到国内基址（单节点阶段）。

密钥侧通过 **`merchant_api_keys.region`** + `route_config.endpoints.litellm.domestic|overseas` 选择逻辑节点；`ExecutionProviderConfig.BYOKRegion` 与之一致。

---

## 5. 相关代码入口

| 能力 | 入口 |
|------|------|
| SSOT 解析封装 | `models.ResolveBYOKRoutingEndpoint`、`MerchantAPIKey.GetEndpointForMode` |
| 执行层出站 | `services.ResolveEndpoint`、`ResolveEndpointByType`、`ExecutionLayer.resolveEndpoint` |
| 网关模式 | `ConfigureGatewayMode`、`determineGatewayMode` |
| 代理 HTTP | `handlers.resolveEndpointURL`、`applyAPIKeyRouteConfig` |
| 健康 / 探测基址 | `HealthChecker.resolveEndpointWithRouteMode`、`ResolveMerchantAPIKeyUpstreamBase` |
