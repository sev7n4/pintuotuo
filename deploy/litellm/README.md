# LiteLLM 与本项目对齐说明

本文档落地「网关能力分层」：LiteLLM 负责多厂商统一出口与可配置路由/容错；业务侧负责账户、权益、预扣费与目录（SSOT）。

**维护入口（厂商 / 模型 / 密钥从哪来）**：先读 [**SSOT_ROUTING.md**](./SSOT_ROUTING.md)；**LiteLLM 网关映射**以 DB **`model_providers.litellm_*`** 为主，可选 [`provider_gateway_map.json`](./provider_gateway_map.json) 作 `-map` 覆盖；再用 `make litellm-catalog-verify` / `make litellm-catalog-generate` 对齐 [`litellm_proxy_config.yaml`](./litellm_proxy_config.yaml)。

## 运行时环境

- Docker：`docker-compose.prod.yml` 中 `litellm` 服务；配置挂载 [`litellm_proxy_config.yaml`](./litellm_proxy_config.yaml)。
- 后端：`LLM_GATEWAY_ACTIVE=litellm`、`LLM_GATEWAY_LITELLM_URL`（指向网关 `/v1` 根）、`LITELLM_MASTER_KEY`（与网关一致）。见 [`backend/handlers/api_proxy.go`](../../backend/handlers/api_proxy.go) 中 `applyGatewayOverride` / `resolveGatewayAuthToken`。
- **`API_PROXY_LITELLM_MAX_RETRIES`**（默认 `1`）：在启用 LiteLLM 网关时限制业务层 `api_proxy` 的 HTTP `MaxRetries`，避免与网关 `router_settings.num_retries` 叠加导致重试风暴。见 `applyLitellmGatewayRetryCap`。

## 目录与 model_list（SSOT）

- 校验/生成：见 [`backend/cmd/litellm-catalog-sync/main.go`](../../backend/cmd/litellm-catalog-sync/main.go) 顶部用法；Makefile 目标 `litellm-catalog-verify`、`litellm-catalog-generate`。
- CI：`.github/workflows/integration-tests.yml` 中对 `litellm_proxy_config.yaml` 的 soft verify（映射来自已迁移的 `model_providers`）。
- **Fallback 引用**：`router_settings` 内 `fallbacks` / `context_window_fallbacks` / `content_policy_fallbacks` 中出现的 `model_name` 必须已在 `model_list` 中定义（`litellm-catalog-sync -verify` 会校验；`#` 注释行不参与解析）。

## 请求关联（可观测性）

- 后端在转发至上游（含 LiteLLM）时设置 **`X-Request-ID`**，并透传 **`traceparent` / `tracestate` / `baggage`**（若入口已带），便于将应用日志、网关日志与 Prometheus 指标按同一次调用对齐。
- Prometheus 对 `litellm:4000` 的抓取为**聚合序列**；单次请求与后端日志对齐依赖上述头。见 [`deploy/prometheus/prometheus.yml`](../../deploy/prometheus/prometheus.yml) 中 `litellm` job 注释。

## 错误与重试：分层职责（R1）

| 现象 | 优先排查层 | 说明 |
|------|------------|------|
| 429 / 上游 5xx / 连接失败 | LiteLLM 网关日志、`num_retries`、fallback 是否触发 | 网关对厂商侧重试与降级 |
| 业务层再次重试同一 URL | 后端 `executeProviderRequestWithRetry` + `API_PROXY_LITELLM_MAX_RETRIES` | 与网关叠加时注意总延迟 |
| 无可用商户 Key / 403 / 未验证 | **SmartRouter**、商户 Key 健康与配额 | 与厂商路由独立 |
| 余额不足、权益拒绝 | 平台计费与 `Entitlement` | 网关不知情 |
| 响应 `model` 与请求不一致 | 网关 **fallback** 或厂商别名 | 后端会打 `upstream response model differs from request` 日志（非流式）；对账时留意 |

## 网关与厂商侧排查（400 / `no healthy deployments` / Invalid model）

业务侧 `api_proxy` 在 `LLM_GATEWAY_ACTIVE=litellm` 时把请求转到 **LiteLLM**，请求体里的 `model` 为「目录短名」（例如 `glm-5`、`step-1-8k`），须与 [`litellm_proxy_config.yaml`](./litellm_proxy_config.yaml) 中 **`model_name`** 一致；上游真实路由由 `litellm_params.model` + 各厂商 `api_key` 决定。

| 现象 | 含义 | 运维动作 |
|------|------|----------|
| `Invalid model name ... model=step-1-8k` | 网关 **model_list 未注册** 该短名，或名称与目录不一致 | 在 yaml 增加对应 `model_name` 与 `litellm_params`；`litellm-catalog-sync -verify` 校验目录与 yaml；重建 `pintuotuo-litellm` 使配置生效 |
| `no healthy deployments for glm-5` / `glm-5.1` | LiteLLM 认为该 deployment **不健康**（常见：缺 Key、Key 无效、上游不可达、健康探测失败） | 在宿主机 `.env` 配置 **`ZAI_API_KEY`**（与 yaml 中 `zai/glm-*` 一致），`docker compose ... up -d litellm` 注入环境；查容器日志与 `GET http://litellm:4000/health`（需 master key） |
| 智谱 **直连** vs **网关** | DB `model_providers.zhipu` 的默认 `api_base_url` 指向 BigModel 旧网关；**LiteLLM 条目使用 `zai/glm-*` + `ZAI_API_KEY`** | 走 LiteLLM 时以 **yaml + ZAI_API_KEY** 为准，勿与直连 URL 混用 |
| StepFun `step-1-8k` | 需 **`STEPFUN_API_KEY`** + 上表 yaml 中 `openai/step-1-8k` 与 `api_base: https://api.stepfun.com/v1` | 若厂商已下线/改名该 model id，需同步改 **SPU `provider_model_id`** 与网关 `model_name` |

验证（生产主机，勿在日志中打印完整 Key）：

```bash
docker exec pintuotuo-litellm sh -c 'test -n "$ZAI_API_KEY" && echo ZAI_API_KEY=set || echo ZAI_API_KEY=missing'
docker exec pintuotuo-litellm sh -c 'test -n "$STEPFUN_API_KEY" && echo STEPFUN_API_KEY=set || echo STEPFUN_API_KEY=missing'
```

## Router / Fallback 模板（F1）

- `router_settings` 与可选 `fallbacks` / `context_window_fallbacks` / `content_policy_fallbacks` 见 [`litellm_proxy_config.yaml`](./litellm_proxy_config.yaml) 内注释及 [Reliability](https://docs.litellm.ai/docs/proxy/reliability)。
- 启用前：运行 `make litellm-catalog-verify`；确认链路上所有 `model_name` 已在 `model_list` 且与 **`model_providers` 中 LiteLLM 映射**（及可选 `-map` 文件）一致。

## 流式（S0–S3）

- **范围（S0）**：当前 **`POST .../openai/v1/chat/completions`**（OpenAI 兼容）在 **`stream: true`** 时走 SSE 透传；仅 **OpenAI 兼容**提供商（`api_format=openai`）。Anthropic `/messages` 等仍返回「仅支持 OpenAI 兼容流式」类错误。
- **协议（S1）**：上游 `Accept: text/event-stream`；响应头 `X-Accel-Buffering: no` 便于 Nginx 不缓冲 SSE。
- **计费（S2）**：优先解析流内 **`usage`** 块；若无则按**流字节粗估** output，可能与实际账单有偏差，适合可接受近似场景。
- **入口与 Nginx（S3）**：若经 Nginx 反代，建议对该 location 配置 `proxy_buffering off`、`proxy_read_timeout` 足够长（如 15m 级），与后端流式超时一致。

## 跨模型 fallback 与计费（R3）

- 若启用跨模型 **fallback**，上游返回的 **`model` 字段可能与请求不一致**；平台按**请求的 `model` 与价目**扣费，对账时请结合网关日志与上述 **model 差异** 日志。
- 需要「按实际服务模型计费」时须单独产品设计（本迭代未改价目解析逻辑）。

## 不建议在网关重复建设的能力

- 多租户计费和订单级价目版本：保持现有后端与数据库模型。
- 若启用 LiteLLM Virtual Keys / 团队预算，需单独设计与本系统 API Key、商户配额的双向映射，避免两套真相。

## 部署回滚（F3）

1. 修改 [`litellm_proxy_config.yaml`](./litellm_proxy_config.yaml)（例如注释掉 `fallbacks` 块或恢复上一版）。
2. 在部署主机：`cd` 至项目目录，`docker compose -f docker-compose.prod.yml -f docker-compose.prod.images.yml --profile llm-gateway up -d`（与 CI 一致）。
3. 确认 `pintuotuo-litellm` 与 `pintuotuo-backend` 健康；必要时查看网关容器日志中 model 路由与报错。
