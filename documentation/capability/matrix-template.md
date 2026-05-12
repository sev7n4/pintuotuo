# 厂商 × 端点能力矩阵（模板）（Phase 0.2）

**版本**：1.0  
**日期**：2026-05-12  
**填写说明**：每格填状态码 + 证据（链接或 `capability-probe-output.csv` 行号/时间戳）。勿将 API Key 写入本文件。

## 状态枚举

| 状态 | 含义 |
|------|------|
| `Supported` | 有文档或实测最小请求 **2xx** 且响应体符合预期 |
| `Unsupported` | 文档明确不支持，或实测稳定 **4xx** 且无兼容计划 |
| `Unknown` | 仅有 `GET /v1/models` 或未测 |
| `ViaGatewayOnly` | 仅通过 LiteLLM/代理等网关可达；需注明后端真实 provider |
| `DifferentBase` | 与 chat 不同 base 或路径策略；需 Phase 4 配置 |

## 主表：OpenAI 形状子路径（相对各厂商 OpenAI 兼容根）

**列**与 `backend/services/execution_layer.go` 中 `EndpointType*` 一致。

| 厂商 `code` | chat_completions | embeddings | images_generations | images_variations | images_edits | audio_transcriptions | audio_translations | audio_speech | moderations | responses |
|---------------|------------------|------------|----------------------|-------------------|--------------|----------------------|--------------------|--------------|-------------|-----------|
| openai | | | | | | | | | | |
| deepseek | | | | | | | | | | |
| zhipu | | | | | | | | | | |
| baidu | | | | | | | | | | |
| bytedance | | | | | | | | | | |
| alibaba | | | | | | | | | | |
| minimax | | | | | | | | | | |
| moonshot | | | | | | | | | | |
| stepfun | | | | | | | | | | |
| google | | | | | | | | | | |
| __default__ | | | | | | | | | | |

> **注意**：`baidu` 在种子数据中 `api_format` 为 `baidu`，与上表 OpenAI 子路径可能不一致；探测前请确认实际转发层行为，必要时单独子表。

## 子表 A：Anthropic Messages API（非上表 OpenAI 子路径）

| 能力 | 路径（平台侧） | 状态 | 证据 |
|------|----------------|------|------|
| List models | `GET /api/v1/anthropic/v1/models`（经平台网关） | | |
| Messages | `POST /api/v1/anthropic/v1/messages` | | |

## 子表 B：LiteLLM / 代理入口（若启用）

| 入口 | 后端路由说明 | chat | embeddings | … |
|------|----------------|------|------------|---|
| LLM_GATEWAY / litellm | 写明实际路由到的 provider 列表 | | | |

## P1 范围签字（Phase 0.6）

- **签字日期**：________  
- **P1 必须 Supported 的格子**：（逐条列出，例如 `openai × embeddings`）  
- **明确不在 P1 的格子**：  
