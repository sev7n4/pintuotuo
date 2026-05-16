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

- **国内**：`docker-compose.prod.yml` 中 `litellm` 挂载本目录下 yaml；CI 见 `.github/workflows/deploy-litellm.yml` / `deploy-tencent.yml`。
- **海外（新加坡）**：`docker-compose.overseas.yml` + 服务器本地 `.env`；CI 见 `.github/workflows/deploy-litellm-overseas.yml`（`main` 推送 `deploy/litellm/**` 或手动 `workflow_dispatch`）。
- **不要在容器内注入各厂商 `*_API_KEY`**（BYOK 由后端 `user_config` 注入）。

### 海外 GitHub Secrets

完整配置步骤、失败排查与验收见 **[documentation/ops/litellm-overseas-github-secrets.md](../../documentation/ops/litellm-overseas-github-secrets.md)**。

| Secret | 说明 |
|--------|------|
| `TENCENT_CLOUD_OVERSEAS_SSH_KEY` | 海外机 SSH 私钥（配置在 **Environments → production**） |
| `TENCENT_CLOUD_OVERSEAS_USER` | SSH 用户（如 `ubuntu`） |
| `TENCENT_CLOUD_OVERSEAS_IP` | 海外机 IP |
| `TENCENT_CLOUD_OVERSEAS_LITELLM_DIR` | 部署目录（如 `/opt/pintuotuo-litellm`） |
| `LITELLM_MASTER_KEY` | 与 backend 一致，用于部署后 `/health` 校验 |

首次部署前在目标目录创建 `.env`（含 `LITELLM_MASTER_KEY`），workflow **不会**覆盖该文件。示例见 [`.env.example`](./.env.example)；一键从大陆同步见 [`scripts/bootstrap-overseas-litellm-env.sh`](../../scripts/bootstrap-overseas-litellm-env.sh)。

