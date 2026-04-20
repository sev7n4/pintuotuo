#!/usr/bin/env bash
# 在线验收 Tempo + OTel 链路：
# 1) 调用后端接口，提取响应头 X-Trace-ID
# 2) 输出 Grafana Explore 与 Tempo API 查询入口
# 3) 传入 AUTH_TOKEN，或用账号密码/短信自动换取 token 后，校验用户链路是否出现 enduser.id
set -euo pipefail

API_BASE_URL="${API_BASE_URL:-http://127.0.0.1:8080/api/v1}"
GRAFANA_URL="${GRAFANA_URL:-http://127.0.0.1:3001}"
TEMPO_URL="${TEMPO_URL:-http://127.0.0.1:3200}"
TEMPO_DATASOURCE_UID="${TEMPO_DATASOURCE_UID:-tempo}"
REQUEST_TIMEOUT_SECONDS="${REQUEST_TIMEOUT_SECONDS:-20}"

AUTH_TOKEN="${AUTH_TOKEN:-}"
LOGIN_EMAIL="${LOGIN_EMAIL:-}"
LOGIN_PASSWORD="${LOGIN_PASSWORD:-}"
LOGIN_TOTP_CODE="${LOGIN_TOTP_CODE:-}"
SMS_PHONE="${SMS_PHONE:-}"
SMS_CODE="${SMS_CODE:-}"
SMS_SCENE="${SMS_SCENE:-login}"
SMS_AUTO_SEND="${SMS_AUTO_SEND:-true}"

ANON_PATH="${ANON_PATH:-/health/ready}"
AUTH_PATH="${AUTH_PATH:-/users/me}"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

trim_trailing_slash() {
  local url="$1"
  echo "${url%/}"
}

urlencode() {
  python3 -c 'import sys, urllib.parse; print(urllib.parse.quote(sys.argv[1], safe=""))' "$1"
}

build_grafana_explore_url() {
  local trace_id="$1"
  local grafana
  grafana="$(trim_trailing_slash "$GRAFANA_URL")"
  local left_json
  left_json=$(cat <<EOF
{"datasource":"${TEMPO_DATASOURCE_UID}","queries":[{"refId":"A","query":"${trace_id}"}]}
EOF
)
  echo "${grafana}/explore?orgId=1&left=$(urlencode "$left_json")"
}

extract_trace_id() {
  local header_file="$1"
  awk 'BEGIN{IGNORECASE=1} /^X-Trace-ID:/ {gsub("\r","",$2); print $2}' "$header_file" | tail -n 1
}

auto_login_and_get_token() {
  local login_url
  login_url="$(trim_trailing_slash "$API_BASE_URL")/users/login"
  local login_body_file="${TMP_DIR}/login.body"
  local login_resp_file="${TMP_DIR}/login.resp"
  local code

  if [[ -z "$LOGIN_EMAIL" || -z "$LOGIN_PASSWORD" ]]; then
    return 1
  fi

  echo "" >&2
  echo "=== Auto login for JWT ===" >&2
  echo "Request: POST ${login_url} (email/password)" >&2

  python3 - "$LOGIN_EMAIL" "$LOGIN_PASSWORD" "$LOGIN_TOTP_CODE" > "$login_body_file" <<'PY'
import json
import sys

payload = {
    "email": sys.argv[1],
    "password": sys.argv[2],
}
totp = (sys.argv[3] or "").strip()
if totp:
    payload["totp_code"] = totp
print(json.dumps(payload, ensure_ascii=False))
PY

  code=$(curl -sS --max-time "$REQUEST_TIMEOUT_SECONDS" \
    -H "Content-Type: application/json" \
    -o "$login_resp_file" \
    -w "%{http_code}" \
    -X POST \
    --data "@${login_body_file}" \
    "$login_url")

  echo "HTTP: ${code}" >&2
  if [[ "$code" -lt 200 || "$code" -ge 300 ]]; then
    echo "WARN: 自动登录失败，响应前 300 字符：" >&2
    head -c 300 "$login_resp_file" >&2 || true
    echo "" >&2
    return 1
  fi

  python3 - "$login_resp_file" <<'PY'
import json
import sys

path = sys.argv[1]
with open(path, "r", encoding="utf-8") as f:
    data = json.load(f)
token = (((data or {}).get("data") or {}).get("token") or "").strip()
if not token:
    sys.exit(2)
print(token)
PY
}

auto_sms_login_and_get_token() {
  local send_url login_url
  send_url="$(trim_trailing_slash "$API_BASE_URL")/users/sms/send"
  login_url="$(trim_trailing_slash "$API_BASE_URL")/users/sms/login"
  local send_body_file="${TMP_DIR}/sms_send.body"
  local send_resp_file="${TMP_DIR}/sms_send.resp"
  local login_body_file="${TMP_DIR}/sms_login.body"
  local login_resp_file="${TMP_DIR}/sms_login.resp"
  local send_code login_code code_to_use

  if [[ -z "$SMS_PHONE" ]]; then
    return 1
  fi

  code_to_use="$SMS_CODE"
  if [[ -z "$code_to_use" && "${SMS_AUTO_SEND,,}" == "true" ]]; then
    echo "" >&2
    echo "=== Auto SMS send ===" >&2
    echo "Request: POST ${send_url}" >&2

    python3 - "$SMS_PHONE" "$SMS_SCENE" > "$send_body_file" <<'PY'
import json
import sys
print(json.dumps({"phone": sys.argv[1], "scene": sys.argv[2]}, ensure_ascii=False))
PY

    send_code=$(curl -sS --max-time "$REQUEST_TIMEOUT_SECONDS" \
      -H "Content-Type: application/json" \
      -o "$send_resp_file" \
      -w "%{http_code}" \
      -X POST \
      --data "@${send_body_file}" \
      "$send_url")
    echo "HTTP: ${send_code}" >&2

    if [[ "$send_code" -lt 200 || "$send_code" -ge 300 ]]; then
      echo "WARN: 自动发短信失败，响应前 300 字符：" >&2
      head -c 300 "$send_resp_file" >&2 || true
      echo "" >&2
      return 1
    fi

    code_to_use="$(python3 - "$send_resp_file" <<'PY'
import json
import sys
path = sys.argv[1]
with open(path, "r", encoding="utf-8") as f:
    data = json.load(f)
print((data.get("debug_code") or "").strip())
PY
)"
    if [[ -z "$code_to_use" ]]; then
      echo "WARN: 未在发码响应里拿到 debug_code，请手动设置 SMS_CODE 后重试。" >&2
      return 1
    fi
  fi

  if [[ -z "$code_to_use" ]]; then
    return 1
  fi

  echo "" >&2
  echo "=== Auto SMS login for JWT ===" >&2
  echo "Request: POST ${login_url}" >&2

  python3 - "$SMS_PHONE" "$code_to_use" > "$login_body_file" <<'PY'
import json
import sys
print(json.dumps({"phone": sys.argv[1], "code": sys.argv[2]}, ensure_ascii=False))
PY

  login_code=$(curl -sS --max-time "$REQUEST_TIMEOUT_SECONDS" \
    -H "Content-Type: application/json" \
    -o "$login_resp_file" \
    -w "%{http_code}" \
    -X POST \
    --data "@${login_body_file}" \
    "$login_url")
  echo "HTTP: ${login_code}" >&2

  if [[ "$login_code" -lt 200 || "$login_code" -ge 300 ]]; then
    echo "WARN: 短信登录失败，响应前 300 字符：" >&2
    head -c 300 "$login_resp_file" >&2 || true
    echo "" >&2
    return 1
  fi

  python3 - "$login_resp_file" <<'PY'
import json
import sys

path = sys.argv[1]
with open(path, "r", encoding="utf-8") as f:
    data = json.load(f)
token = (((data or {}).get("data") or {}).get("token") or "").strip()
if not token:
    sys.exit(2)
print(token)
PY
}

call_api() {
  local name="$1"
  local path="$2"
  local auth_header="${3:-}"
  local header_file="${TMP_DIR}/${name}.headers"
  local body_file="${TMP_DIR}/${name}.body"
  local code

  local url
  url="$(trim_trailing_slash "$API_BASE_URL")${path}"
  echo ""
  echo "=== ${name} ==="
  echo "Request: GET ${url}"

  if [[ -n "$auth_header" ]]; then
    code=$(curl -sS --max-time "$REQUEST_TIMEOUT_SECONDS" -D "$header_file" -o "$body_file" \
      -H "$auth_header" -w "%{http_code}" "$url")
  else
    code=$(curl -sS --max-time "$REQUEST_TIMEOUT_SECONDS" -D "$header_file" -o "$body_file" \
      -w "%{http_code}" "$url")
  fi

  local trace_id
  trace_id="$(extract_trace_id "$header_file")"
  echo "HTTP: ${code}"
  if [[ -n "$trace_id" ]]; then
    echo "X-Trace-ID: ${trace_id}"
  else
    echo "X-Trace-ID: <missing>"
  fi

  if [[ "${code}" -ge 400 ]]; then
    echo "Response (first 300 chars):"
    head -c 300 "$body_file" || true
    echo ""
  fi

  if [[ -n "$trace_id" ]]; then
    local explore_url
    explore_url="$(build_grafana_explore_url "$trace_id")"
    echo "Grafana Explore: ${explore_url}"
    echo "Tempo API: $(trim_trailing_slash "$TEMPO_URL")/api/traces/${trace_id}"
  fi

  echo "$trace_id"
}

check_enduser_id_in_tempo() {
  local trace_id="$1"
  local tempo_api
  tempo_api="$(trim_trailing_slash "$TEMPO_URL")/api/traces/${trace_id}"
  echo ""
  echo "检查 Tempo Trace 属性（enduser.id）: ${tempo_api}"

  if ! curl -sS --max-time "$REQUEST_TIMEOUT_SECONDS" "$tempo_api" > "${TMP_DIR}/trace.json"; then
    echo "WARN: 无法访问 Tempo API，请手动在 Grafana Explore 查询该 Trace ID。"
    return 0
  fi

  if python3 - "${TMP_DIR}/trace.json" <<'PY'
import json
import sys

path = sys.argv[1]
with open(path, "r", encoding="utf-8") as f:
    data = json.load(f)

found = False
values = set()

def walk(obj):
    global found
    if isinstance(obj, dict):
        key = obj.get("key")
        if key == "enduser.id":
            found = True
            v = obj.get("value", {})
            for vk in ("stringValue", "intValue", "doubleValue", "boolValue"):
                if vk in v:
                    values.add(str(v[vk]))
        for item in obj.values():
            walk(item)
    elif isinstance(obj, list):
        for item in obj:
            walk(item)

walk(data)
if found:
    print("FOUND enduser.id:", ",".join(sorted(values)) if values else "(no value)")
    sys.exit(0)
print("NOT_FOUND")
sys.exit(2)
PY
  then
    echo "PASS: 已在 trace 中发现 enduser.id"
  else
    echo "WARN: 未在该 trace 中发现 enduser.id（请确认请求走了鉴权中间件且 token 对应有效用户）。"
  fi
}

echo "Tracing online verification"
echo "- API_BASE_URL: ${API_BASE_URL}"
echo "- GRAFANA_URL:  ${GRAFANA_URL}"
echo "- TEMPO_URL:    ${TEMPO_URL}"

if [[ -z "$AUTH_TOKEN" ]]; then
  if token_from_login="$(auto_login_and_get_token)"; then
    AUTH_TOKEN="$token_from_login"
    echo "Auto login: success（已自动获取 AUTH_TOKEN）"
  elif token_from_sms_login="$(auto_sms_login_and_get_token)"; then
    AUTH_TOKEN="$token_from_sms_login"
    echo "Auto SMS login: success（已自动获取 AUTH_TOKEN）"
  else
    echo "Auto login: skipped/failed（未提供可用参数或登录失败）"
  fi
fi

anon_trace_id="$(call_api "Anonymous request" "$ANON_PATH")"
if [[ -z "$anon_trace_id" ]]; then
  echo ""
  echo "FAIL: 匿名请求未返回 X-Trace-ID。请检查后端 TracingResponseHeaders 中间件与 OTel 初始化。"
  exit 1
fi

if [[ -n "$AUTH_TOKEN" ]]; then
  auth_trace_id="$(call_api "Authenticated request" "$AUTH_PATH" "Authorization: Bearer ${AUTH_TOKEN}")"
  if [[ -z "$auth_trace_id" ]]; then
    echo ""
    echo "WARN: 鉴权请求未返回 X-Trace-ID。"
    exit 0
  fi
  check_enduser_id_in_tempo "$auth_trace_id"
else
  echo ""
  echo "未拿到 AUTH_TOKEN，已跳过用户链路（enduser.id）校验。"
  echo "可用任一方式继续："
  echo "1) AUTH_TOKEN=<jwt> bash scripts/verify_tracing_online.sh"
  echo "2) LOGIN_EMAIL=<email> LOGIN_PASSWORD=<password> [LOGIN_TOTP_CODE=<code>] bash scripts/verify_tracing_online.sh"
  echo "3) SMS_PHONE=<cn_phone> [SMS_CODE=<otp>] bash scripts/verify_tracing_online.sh"
  echo "   - 若未提供 SMS_CODE 且 SMS_AUTO_SEND=true，会先请求 /users/sms/send 并尝试读取 debug_code（MOCK_SMS 常见）。"
fi

echo ""
echo "Done."
