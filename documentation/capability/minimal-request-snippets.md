# 最小请求样例（Phase 0.5 附录）

**用途**：与 `scripts/capability_upstream_probe.sh` 及手工 `curl` 复现保持一致；**模型名**请替换为环境中 `GET /v1/models` 返回且你有权使用的 id。

**约定**：`$BASE` 为 OpenAI 兼容根（例如 `https://api.openai.com/v1`）；`$KEY` 为 Bearer Token（勿提交到仓库）。

## GET /v1/models

```bash
curl -sS -o /tmp/models.json -w "%{http_code}\n" \
  -H "Authorization: Bearer $KEY" \
  -H "Content-Type: application/json" \
  "$BASE/models"
```

## POST /v1/embeddings

```bash
curl -sS -o /tmp/embed.json -w "%{http_code}\n" \
  -H "Authorization: Bearer $KEY" -H "Content-Type: application/json" \
  -d '{"model":"text-embedding-3-small","input":"ping"}' \
  "$BASE/embeddings"
```

## POST /v1/moderations

```bash
curl -sS -o /tmp/mod.json -w "%{http_code}\n" \
  -H "Authorization: Bearer $KEY" -H "Content-Type: application/json" \
  -d '{"model":"omni-moderation-latest","input":"ping"}' \
  "$BASE/moderations"
```

## POST /v1/images/generations（可能计费，慎用）

```bash
curl -sS -o /tmp/img.json -w "%{http_code}\n" \
  -H "Authorization: Bearer $KEY" -H "Content-Type: application/json" \
  -d '{"model":"dall-e-3","prompt":"solid blue square","n":1,"size":"1024x1024"}' \
  "$BASE/images/generations"
```

## POST /v1/audio/speech（可能计费）

```bash
curl -sS -o /tmp/speech.bin -w "%{http_code}\n" \
  -H "Authorization: Bearer $KEY" -H "Content-Type: application/json" \
  -d '{"model":"tts-1","input":"ping","voice":"alloy"}' \
  "$BASE/audio/speech"
```

## POST /v1/audio/transcriptions（需上传文件）

```bash
curl -sS -o /tmp/tr.json -w "%{http_code}\n" \
  -H "Authorization: Bearer $KEY" \
  -F file="@./samples/hello.wav" -F model="whisper-1" \
  "$BASE/audio/transcriptions"
```

## POST /v1/responses

以各环境支持的模型与参数为准；示例（可能因模型而异）：

```bash
curl -sS -o /tmp/resp.json -w "%{http_code}\n" \
  -H "Authorization: Bearer $KEY" -H "Content-Type: application/json" \
  -d '{"model":"gpt-4o-mini","input":[{"type":"message","role":"user","content":"ping"}]}' \
  "$BASE/responses"
```

> **Anthropic**：使用 `x-api-key` 与 `anthropic-version`；路径为 Anthropic 官方 Messages API，**不**使用上表 `$BASE` 拼接方式；参见平台 `RegisterAnthropicCompatRoutes`。
