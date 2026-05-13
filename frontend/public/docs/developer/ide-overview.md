# 总览：接入拼脱脱

> 契约以后端 **`GET /tokens/api-usage-guide`** 的 `openai_compat_path`、`anthropic_compat_path` 与 **`items`** 为准。生产请使用 **HTTPS** 与自有域名；下文用 `https://YOUR_ORIGIN` 占位。

## 通用准备

1. **平台 API Key**  
   - **开发者中心 → 密钥与安全** 创建，前缀一般为 `ptd_`。  
   - 用于 `Authorization: Bearer <ptd_...>`（或各工具等价的 API Key 输入框）。

2. **可用模型（与权益一致）**  
   - 登录后请求 **`GET /api/v1/tokens/api-usage-guide`**，查看 `items` 中的 `provider_code`、`model_example` / `provider_slash_example`。  
   - 若环境为 **`ENTITLEMENT_ENFORCEMENT=strict`**，**`model` 必须落在该列表**（含 `provider/model`），否则会 **403**。

3. **Base URL（与后端路由一致）**

| 协议 | Base URL（拼接到 `YOUR_ORIGIN` 后） | 典型请求 |
|------|----------------------------------------|----------|
| OpenAI 兼容 | `https://YOUR_ORIGIN/api/v1/openai/v1` | `POST …/chat/completions` |
| Anthropic 兼容 | `https://YOUR_ORIGIN/api/v1/anthropic/v1` | `POST …/messages` |

两条路径在 **IDE 场景**下均为「除 `model`（及 LiteLLM 的 `user_config`）外 **请求体原样出站**」；Anthropic 路由要求目录中该模型的 **`api_format` 为 anthropic**，否则会在 fallback 中跳过直至失败。

本地开发时 `YOUR_ORIGIN` 常为 `http://localhost:5173`（经 Vite 代理，与 `VITE_API_BASE_URL` 一致）。

4. **余额**  
   - 调用扣 **Token 余额**，与「我的 Token」一致。

## 排障与相关文档

| 文档 | 内容 |
|------|------|
| [`errors.md`](./errors.md) | HTTP 状态与鉴权 |
| [`routing-ssot.md`](./routing-ssot.md) | 路由 / LiteLLM 摘要 |
| [`openai.md`](./openai.md) | OpenAI 路径与 curl |

若 **403**，优先核对 **`model` 是否与 `api-usage-guide` 一致**、订单/订阅是否在有效期内。

## 版本说明

第三方工具 UI 与变量名更新较快；以**各工具官方文档**为主，本文描述与拼脱脱后端的**路径、`ptd_`、`provider/model`、strict** 契约。
