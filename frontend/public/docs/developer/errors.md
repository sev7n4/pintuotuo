# HTTP 状态与鉴权（C 端代理）

本文档描述**拼脱脱 OpenAI / Anthropic 兼容代理**（路径含 `/openai/`、`/anthropic/`）与**平台登录 JWT**并存时的语义，便于区分「登录过期」与「上游/BYOK 失败」。

## 前端 axios 行为（`frontend/src/services/api.ts`）

| 条件 | 行为 |
|------|------|
| 响应 **401** 且 URL **不含** `/openai/`、`/anthropic/`，且不是登录接口 | 清除本地 token，跳转登录页 |
| 响应 **401** 且 URL **含** `/openai/` 或 `/anthropic/` | **不清除** JWT（避免把上游鉴权失败当成平台登出） |

## 后端映射（`route_mode = litellm`）

当 API 密钥走 LiteLLM 网关时，`mapLLMProxyHTTPStatusForClient` 会将上游部分鉴权类错误映射为 **502**，与前端上述豁免形成纵深：

| 上游 HTTP | 客户端看到的 HTTP | 说明 |
|-----------|-------------------|------|
| 401、403 | **502**（仅 litellm 模式） | 多为上游 Key/Base 或网关鉴权问题，**不应**当作平台 JWT 失效 |
| 其他 | 视情况透传或业务映射 | 以实际响应体 `error.message` 为准 |

非 `litellm` 的直连等模式可能**透传**上游 401/403，此时仍受 axios「路径豁免」保护，不会误清 JWT。

## 用户建议动作

| 现象 | 建议 |
|------|------|
| 502 + 调用模型代理 | 检查权益是否包含该 `model`；检查套餐是否有效；仍失败联系客服并带 `request_id`（响应头或日志） |
| 401 在 `/openai/` 上但页面未跳登录 | 预期行为：请检查平台 `ptd_` 密钥是否正确、是否启用 |
| 平台页面被踢回登录 | 多为非代理接口 401，请重新登录 |

## 维护

若修改 `mapLLMProxyHTTPStatusForClient` 或 axios 拦截器，请同步更新本文件与 E2E 断言。
