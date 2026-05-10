# SSOT：路由与出站（LiteLLM / 直连 / 代理）

本文档为运维与排障的**单一事实来源（SSOT）**之一；与 `deploy/litellm/README.md`、`DEPLOYMENT.md` 交叉引用时请**只在一处维护具体域名/IP**，其他文档链接至此。

## 两条运行时主线

### 1. C 端 / 通用：`route_mode = litellm` + BYOK（`user_config`）

- 业务后端在代理请求体中注入 `user_config` → `model_list` → `litellm_params`（`model`、`api_key`、`api_base`）。
- 对齐代码：`backend/handlers/api_proxy.go`（`buildLitellmUserConfig`）、`backend/services/api_key_validator.go`（`probeQuotaViaLitellmUserConfig`）。
- 网关鉴权：请求 LiteLLM 时使用 `Authorization: Bearer $LITELLM_MASTER_KEY`（见 `resolveAuthTokenFromRouteMode`）。
- 配置文件：`deploy/litellm/litellm_proxy_config.yaml` 中通配 `model_name` + `configurable_clientside_auth_params`。

### 2. 目录与网关 YAML 校验

- `make litellm-catalog-verify` / `litellm-catalog-assemble`：见 `deploy/litellm/README.md`。
- 通配 `*` 模式下可不逐项对齐 DB 目录；显式列表用于审计/合规。

## route_mode 矩阵（简）

| route_mode | 典型出站 | 鉴权要点 |
|------------|----------|----------|
| `litellm` | LiteLLM Proxy | Master Key + 请求内 BYOK |
| `direct` | `model_providers.api_base_url` | 上游格式对应 Header（OpenAI Bearer / Anthropic x-api-key 等） |
| `proxy` | 经 HTTP(S) 代理访问上述端点 | 宿主机代理、Clash/Mihomo 分流 |

区域策略（国内用户访问海外 provider）见 `backend/services/unified_router.go` 与 DB `route_strategy`。

## 客户端可见 HTTP 行为

- `route_mode=litellm` 时上游 401/403 映射为 **502**，避免 SPA 将上游鉴权失败当作平台 JWT 失效；与 `frontend/src/services/api.ts` 对 `/openai/`、`/proxy/` 的 401 豁免一致。
- C 端可读摘要：`frontend/public/docs/developer/errors.md`。

## 变更流程

1. 改路由或映射逻辑：更新本文件 + `errors.md`（若影响用户可见语义）+ 相关 E2E。
2. 改代理域名表：只改 `DEPLOYMENT.md` 或 Clash 配置一处权威，其余文件链接。
