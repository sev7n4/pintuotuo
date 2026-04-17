#!/usr/bin/env bash
# 在已部署主机上执行（项目根目录，已 docker compose --profile llm-gateway up）。
# 用途：确认双轨网关容器存活，并可选验证「平台 Key → LiteLLM」路径（BYOK 需走业务 API，见 Runbook）。
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

echo "=== docker: litellm / oneapi 运行状态 ==="
docker ps --format 'table {{.Names}}\t{{.Status}}' | grep -E 'NAMES|pintuotuo-litellm|pintuotuo-oneapi' || true

for name in pintuotuo-litellm pintuotuo-oneapi; do
  if ! docker ps --format '{{.Names}}' | grep -qx "$name"; then
    echo "ERROR: 容器未运行: $name"
    exit 1
  fi
done

echo "=== LiteLLM GET /health (本机 4000) ==="
if [[ -n "${LITELLM_MASTER_KEY:-}" ]]; then
  curl -sS -o /dev/null -w "HTTP %{http_code} (with Bearer master key)\n" --max-time 10 \
    -H "Authorization: Bearer ${LITELLM_MASTER_KEY}" \
    "http://127.0.0.1:4000/health" || echo "curl failed"
else
  curl -sS -o /dev/null -w "HTTP %{http_code} (no key; 401 is common on newer proxy)\n" --max-time 10 \
    "http://127.0.0.1:4000/health" || echo "curl failed"
fi

echo "=== OneAPI GET / (本机 3002；首次启动 token encoders 可能需 1–3 分钟) ==="
curl -sS -o /dev/null -w "HTTP %{http_code}\n" --max-time 15 "http://127.0.0.1:3002/" || echo "curl failed"

if [[ -n "${LITELLM_MASTER_KEY:-}" ]]; then
  echo "=== 路径 A（LiteLLM）：平台 Master Key → GET /v1/models ==="
  code=$(curl -sS -o /tmp/litellm_models.json -w "%{http_code}" --max-time 15 \
    -H "Authorization: Bearer ${LITELLM_MASTER_KEY}" \
    "http://127.0.0.1:4000/v1/models" || echo "000")
  echo "HTTP $code"
  head -c 300 /tmp/litellm_models.json 2>/dev/null || true
  echo ""
else
  echo "=== 路径 A（LiteLLM）：跳过（未设置 LITELLM_MASTER_KEY）==="
fi

if [[ -n "${ONEAPI_ACCESS_TOKEN:-}" ]]; then
  echo "=== 路径 A（OneAPI）：平台 Access Token → GET /v1/models（与后端 ONEAPI_ACCESS_TOKEN 一致）==="
  code=$(curl -sS -o /tmp/oneapi_models.json -w "%{http_code}" --max-time 15 \
    -H "Authorization: Bearer ${ONEAPI_ACCESS_TOKEN}" \
    "http://127.0.0.1:3002/v1/models" || echo "000")
  echo "HTTP $code"
  head -c 300 /tmp/oneapi_models.json 2>/dev/null || true
  echo ""
else
  echo "=== 路径 A（OneAPI）：跳过（未设置 ONEAPI_ACCESS_TOKEN；请在 OneAPI 控制台创建令牌后写入 .env）==="
fi

echo "=== 路径 B（BYOK）==="
echo "当 LLM_GATEWAY_ACTIVE 为 litellm/oneapi 但未配置平台 Key 时，后端对 OpenAI 格式请求会使用商户库内解密的 Key 作为 Bearer。"
echo "请在业务环境调用 POST /api/v1/proxy/chat 或 POST /api/v1/openai/v1/chat/completions，并携带已授权商户/平台 API Key 做联调。"

echo "=== 完成 ==="
