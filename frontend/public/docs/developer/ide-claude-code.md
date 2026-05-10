# Claude Code（Anthropic 协议 + `cc switch`）

Claude Code 使用 **Anthropic Messages API**。将上游从 `api.anthropic.com` 换到拼脱脱的 **`/api/v1/anthropic/v1`**，并把 **模型 ID** 换成 **`api-usage-guide` 中有权益**的 `provider/model`（如部分编程套餐为 `alibaba/...`），即可在 **strict** 下使用。

## 核心环境变量

| 变量 | 说明 |
|------|------|
| `ANTHROPIC_BASE_URL` | `https://YOUR_ORIGIN/api/v1/anthropic/v1`（若客户端拼接出现 `/v1/v1`，去掉多余段）。 |
| `ANTHROPIC_AUTH_TOKEN` 或 `ANTHROPIC_API_KEY` | `ptd_...` 平台密钥。 |
| `ANTHROPIC_MODEL` | 默认模型，须为权益内 `provider/model`。 |
| `ANTHROPIC_DEFAULT_SONNET_MODEL` 等 | 各档位模型同样须为权益内 ID。 |

键名以 **Claude Code 当前版本文档**为准。

## `cc switch` / Profile JSON

**`cc switch`** 用于在多套环境间切换。典型做法：导入 **JSON profile**，在 `env` 中写入上表变量。

**示例（占位符，勿提交真实密钥）：**

```json
{
  "effortLevel": "high",
  "env": {
    "ANTHROPIC_AUTH_TOKEN": "ptd_YOUR_KEY_HERE",
    "ANTHROPIC_BASE_URL": "https://YOUR_ORIGIN/api/v1/anthropic/v1",
    "ANTHROPIC_DEFAULT_HAIKU_MODEL": "alibaba/qwen3.6-plus",
    "ANTHROPIC_DEFAULT_OPUS_MODEL": "alibaba/MiniMax-M2.5",
    "ANTHROPIC_DEFAULT_SONNET_MODEL": "alibaba/kimi-k2.6",
    "ANTHROPIC_MODEL": "alibaba/glm-5.1",
    "ANTHROPIC_REASONING_MODEL": "alibaba/deepseek-v4-pro"
  },
  "includeCoAuthoredBy": false,
  "model": "alibaba/glm-5.1"
}
```

- `alibaba/...` 仅为示例，**必须**换成你 **`api-usage-guide`** 中的值。  
- 密钥泄露请立即在 **密钥与安全** 中轮换。

## 验证

- 开发者中心 **快速开始 → 浏览器内试打**（需已购权益）。  
- 或对 OpenAI 兼容路径执行 `curl`（见 `openai.md`）。
