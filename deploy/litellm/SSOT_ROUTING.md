# LiteLLM 路由与 API Key：分层 SSOT 与维护入口

维护「网关可用模型」成本高，通常是因为 **DB 商品目录、`provider_gateway_map.json`、`litellm_proxy_config.yaml`、容器环境变量、商户 BYOK** 多条链路未对齐。本节约定 **唯一配置驱动入口** 与校验命令。

## 分层（数据流自上而下）

| 层 | 位置 | 作用 |
|----|------|------|
| **1. 商品目录** | PostgreSQL `spus` × `model_providers` | 上架模型、OpenAI 兼容请求体中的 **短名**（`provider_model_id` 或 `model_name`） |
| **2. 网关映射（配置 SSOT）** | [`provider_gateway_map.json`](./provider_gateway_map.json) | 平台 `model_providers.code` → LiteLLM 上游 **`litellm_params.model`**（`litellm_model_template` 含 `{model_id}`）、**`api_key` 环境变量名**、可选 **`api_base`** |
| **3. LiteLLM 运行时** | [`litellm_proxy_config.yaml`](./litellm_proxy_config.yaml) `model_list` | 实际加载的模型表（含手搓 P0 与目录应对照项） |
| **4. 密钥注入** | 宿主机 `.env` → `docker-compose.prod.yml` → `pintuotuo-litellm` | 与映射表 `api_key_env` 一致（`os.environ/XXX`） |
| **5. 直连 / BYOK** | `merchant_api_keys`、SmartRouter | **未**走 `LLM_GATEWAY_ACTIVE=litellm` 时的另一条链路；**不**参与 `litellm-catalog-sync -verify` |

## 维护流程（改厂商或模型时）

1. 改 **`provider_gateway_map.json`**（新增厂商、改 `zai/` vs `openai/`+`api_base` 等）。
2. 运行 **`make litellm-catalog-generate`**（需 `DATABASE_URL`）生成片段，与 `litellm_proxy_config.yaml` 中 `model_list` **合并/对照**。
3. 运行 **`make litellm-catalog-verify`**，确保每个 **active SPU** 在 yaml 中有对应 `model_name`。
4. 在 `.env` 配置映射表中的 `*_API_KEY`，重建 LiteLLM 容器。

## 历史说明

原 `catalog_provider_map.json` 仅支持 `litellm_prefix`，无法表达「阶跃需 `openai/{id}` + 独立 `api_base`」等形态，已由 **`provider_gateway_map.json`** 替代（`litellm_model_template` + 可选 `api_base`）。
