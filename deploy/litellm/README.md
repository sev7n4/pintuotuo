# LiteLLM 与本项目对齐说明

运维向 **路由 / 出站 / 状态码** 单一说明见 [SSOT_ROUTING.md](./SSOT_ROUTING.md)；C 端摘要见前端 `public/docs/developer/routing-ssot.md`。

## 运行时两条主线（互补）

### 1. 同事 PR #475：`user_config` BYOK（route_mode = `litellm`）

- 业务 `**api_proxy**` 在流式/非流式出站 body 中注入 `**user_config**`（内含 `model_list` → `litellm_params.model` / `api_key` / `api_base`），与 `**probeQuotaViaLitellmUserConfig**` 探测路径一致。
- **上游 `api_base`**：`ResolveLitellmUpstreamBaseURL` **优先**使用 `**model_providers.api_base_url`**，其次 `**litellm_gateway_api_base`**，与校验逻辑对齐。
- **网关鉴权**：请求头 `**Authorization: Bearer $LITELLM_MASTER_KEY`**（见 `resolveAuthTokenFromRouteMode`）。

### 2. 通配 `model_list` + catalog-sync（本分支补强）

- 默认 `**[litellm_proxy_config.yaml](./litellm_proxy_config.yaml)`**：`model_name: "*"` 与常见 `**provider/*`** + `**configurable_clientside_auth_params**`，上架新模型通常**不必**改网关文件。
- `**make litellm-catalog-verify`**：`yaml` 若含 `**model_name: '*'`**，跳过「目录 ↔ yaml 逐项对齐」；显式列表模式下仍会校验。
- `**make litellm-catalog-assemble**`：可按 DB 生成 **显式** `model_list`（审计/合规）；产物为 BYOK 片段（`configurable_clientside_auth_params`），**不**写 `os.environ/*` 厂商 Key。

## 前端与 HTTP 状态

- `**frontend/src/services/api.ts`**：对 `**/openai/`**、`**/proxy/**` 的 **401** 不清 token（上游 BYOK 失效不误伤 JWT）。
- `**mapLLMProxyHTTPStatusForClient`**：`route_mode=litellm` 时将上游 **401/403** 映射为 **502**，与上述前端策略形成纵深防御。

## 常用命令

```bash
make litellm-catalog-verify
# make litellm-catalog-verify-soft   # 仅警告
make litellm-catalog-generate
make litellm-catalog-assemble      # 需 DATABASE_URL；可选 LITELLM_CATALOG_MAP
make probe-litellm                 # 通配 yaml 下可能无可探测条目，脚本会提示并退出 0
```

## 部署

- Docker：`docker-compose.prod.yml` 中 `litellm` 挂载本目录下 yaml；**不要在容器内注入各厂商 `*_API_KEY`**（BYOK 由后端注入）。

