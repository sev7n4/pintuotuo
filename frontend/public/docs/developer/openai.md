# C 端 OpenAI 兼容接入

## Base URL

将客户端 `baseURL`（或 curl 主机前缀）设为：

- 形如 `{API_ORIGIN}{VITE_API_BASE_URL 去掉末尾斜杠}/openai/v1`
- 默认本地开发：`http://localhost:5173/api/v1/openai/v1`（经 Vite 代理到后端）
- 生产：以部署的 `VITE_API_BASE_URL` 为准，路径后缀恒为 `/openai/v1`

完整 Chat Completions 路径：`POST {baseURL}/chat/completions`

## 鉴权

- Header：`Authorization: Bearer <平台密钥>`
- 平台密钥前缀一般为 `ptd_`，在「开发者中心 → 密钥与安全」或「我的 Token」创建
- **不是** OpenAI 官方 `sk-...`；厂商密钥由平台/商户侧配置（BYOK 走商户工作台）

## 请求体

- `model`：使用 `GET /tokens/api-usage-guide` 返回的 `provider/model` 或与权益一致的写法
- `messages`：标准 OpenAI 角色数组
- `stream`：对 **OpenAI 兼容** 路径可传 `stream: true` 使用 SSE（与后端 `api_proxy` 一致）
- `max_tokens`：建议在试调用时保持较小值以控制成本

## 与 OpenAI 官方 SDK（TypeScript）

```ts
import OpenAI from 'openai';

const client = new OpenAI({
  apiKey: process.env.PTD_KEY!,
  baseURL: 'https://你的域名/api/v1/openai/v1',
  dangerouslyAllowBrowser: false,
});
```

## 进阶

- Anthropic 兼容路径见 `GET /tokens/api-usage-guide` 返回的 `anthropic_compat_path`（若配置）
- **Claude Code、Cursor、Codex 等 IDE/CLI**：见同目录分册（如 [`ide-overview.md`](./ide-overview.md)）或开发者中心 **「IDE 与 CLI 接入」**（多 Tab）
- 多轮对话、**tools / 多模态 messages** 等：请求体会**原样转发**至上游 OpenAI 兼容接口（平台会解析并重写 `model`；走 LiteLLM 时注入 `user_config` 并覆盖客户端同名键）。计费与权益仍以平台规则为准。

## 路线图（可选）

- 浏览器内多轮 Playground、更强配额与审计：见产品排期
