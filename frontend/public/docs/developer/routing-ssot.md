# 路由与出站矩阵（用户可见摘要）

> **运维完整版**以仓库内 `deploy/litellm/SSOT_ROUTING.md` 为准；本节面向 C 端开发者做能力说明。

## C 端调用链（概念）

1. 浏览器或服务器使用**平台 JWT**（开发者中心「试打一条」）或 **ptd_ 密钥**（自建服务）请求拼脱脱 API。
2. 平台根据密钥、`model_providers`、商户/权益配置解析 **route_mode**（如 `direct`、`litellm`、`proxy`）。
3. **litellm**：请求发往自部署 LiteLLM，上游厂商 Key 常由 BYOK 注入（商户侧）；C 端用户一般只感知平台计费与权益。
4. **direct**：直连配置的上游 `api_base`（视网络环境可能需代理）。

## OpenAI 兼容 Base

- 统一前缀：`/api/v1/openai/v1`（以前端 `VITE_API_BASE_URL` 为准）
- 与 `GET /tokens/api-usage-guide` 中 `openai_compat_path` 对齐

## 常见厂商 model 前缀（示例）

| 前缀 | 说明 |
|------|------|
| `openai/*` | OpenAI 路由 |
| `anthropic/*` | Anthropic |
| `dashscope/*` | 阿里灵积等 |
| `moonshot/*`、`zai/*`、`deepseek/*` 等 | 以平台目录为准 |

实际上架模型以 **api-usage-guide** 与控制台为准。

## 网络与部署

国内生产环境访问海外上游时，常需 HTTPS 代理；详见仓库 `DEPLOYMENT.md` 与运维 SSOT，**域名分流列表勿重复维护多份**。
