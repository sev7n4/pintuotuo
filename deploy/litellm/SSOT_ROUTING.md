# LiteLLM 路由与 API Key：分层 SSOT 与维护入口

维护「网关可用模型」成本高，通常是因为 **DB 商品目录、`litellm_proxy_config.yaml`、容器环境变量、商户 BYOK** 多条链路未对齐。本节约定 **唯一配置驱动入口** 与校验命令。

## 分层（数据流自上而下）

| 层 | 位置 | 作用 |
|----|------|------|
| **1. 商品目录** | PostgreSQL `spus` × `model_providers` | 上架模型、OpenAI 兼容请求体中的 **短名**（`provider_model_id` 或 `model_name`） |
| **2. 网关映射（LiteLLM SSOT）** | `model_providers` 列 **`litellm_model_template`**（可含 `{model_id}`）、**`litellm_gateway_api_key_env`**、可选 **`litellm_gateway_api_base`** | 平台 `code` → LiteLLM **`litellm_params.model`**、**`api_key: os.environ/…`**、可选 **`api_base`**；迁移见 `backend/migrations/057_model_providers_litellm_gateway.sql` |
| **2b. 可选 JSON 覆盖** | [`provider_gateway_map.json`](./provider_gateway_map.json) | 与 DB **合并**：`litellm-catalog-sync -map <path>` 时 **同名 `code` 以文件为准**（应急、本地对照）；**不设 `-map` 时仅以 DB 为准** |
| **3. LiteLLM 运行时** | [`litellm_proxy_config.yaml`](./litellm_proxy_config.yaml) `model_list` | 实际加载的模型表（含手搓 P0 与目录应对照项） |
| **4. 密钥注入** | 宿主机 `.env` → `docker-compose.prod.yml` → `pintuotuo-litellm` | 与 **`litellm_gateway_api_key_env`** / JSON `api_key_env` 一致（`os.environ/XXX`） |
| **5. 直连 / BYOK** | `merchant_api_keys`、SmartRouter | **未**走 `LLM_GATEWAY_ACTIVE=litellm` 时的另一条链路；**不**参与 `litellm-catalog-sync -verify` |

## 维护流程（改厂商或模型时）

1. 在 **`model_providers`** 更新对应行的 **`litellm_*`** 字段（或通过管理端 API 写入）；必要时用 **`-map`** 临时覆盖验证。
2. 运行 **`make litellm-catalog-generate`**（需 `DATABASE_URL`）生成片段，与 `litellm_proxy_config.yaml` 中 `model_list` **合并/对照**。
3. 运行 **`make litellm-catalog-verify`**，确保每个 **active SPU** 在 yaml 中有对应 `model_name`。
4. 在 `.env` 配置各厂商 `*_API_KEY`，重建 LiteLLM 容器。

### 合并 `litellm-catalog-generate` 输出

`make litellm-catalog-generate` 写入（默认）**`deploy/litellm/generated_model_list.fragment.yaml`**（gitignore，不落库）。用途：

- **对照**：将片段与 `litellm_proxy_config.yaml` 里现有 `model_list` 条目逐条比对；缺则 **手工追加**到 yaml（保持缩进与注释风格）。
- **不自动覆盖全文件**：当前仓库仍保留手搓 **P0 全量**与 **router_settings**；生成物只覆盖「DB 里 active SPU ∩ 映射表中有 code」的推导结果，二者可能不完全相等（例如 P0 多出来的模型）。

合并后务必再跑 **`make litellm-catalog-verify`**（需 `DATABASE_URL`）。

## 历史说明

原 `catalog_provider_map.json` 仅支持 `litellm_prefix`，无法表达「阶跃需 `openai/{id}` + 独立 `api_base`」等形态，已由 **`litellm_model_template` + 可选 `api_base`** 表达。网关映射已 **字段化进 `model_providers`**；`provider_gateway_map.json` 保留为 **可选合并覆盖** 与文档对照。
