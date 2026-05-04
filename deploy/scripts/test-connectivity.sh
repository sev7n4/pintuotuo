#!/usr/bin/env bash
set -euo pipefail

PROXY_HOST="${PROXY_HOST:-127.0.0.1}"
PROXY_PORT="${PROXY_PORT:-7890}"
PROXY_URL="http://${PROXY_HOST}:${PROXY_PORT}"

API_KEY="${API_KEY:-}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m'

endpoints=(
    "OpenAI|https://api.openai.com/v1/models"
    "Anthropic|https://api.anthropic.com/v1/models"
    "Google|https://generativelanguage.googleapis.com/v1/models"
    "OpenRouter|https://openrouter.ai/api/v1/models"
)

echo "=== Connectivity Test ==="
echo "Proxy: ${PROXY_URL}"
echo ""

test_endpoint() {
    local name="$1"
    local url="$2"
    local mode="$3"

    local start_time end_time duration http_code

    start_time=$(date +%s%N 2>/dev/null || python3 -c "import time; print(int(time.time()*1e9))")

    if [ "$mode" = "proxy" ]; then
        if [ -n "$API_KEY" ]; then
            http_code=$(curl -s -o /dev/null -w "%{http_code}" \
                -x "$PROXY_URL" --connect-timeout 10 --max-time 15 \
                -H "Authorization: Bearer ${API_KEY}" \
                "$url" 2>/dev/null || echo "000")
        else
            http_code=$(curl -s -o /dev/null -w "%{http_code}" \
                -x "$PROXY_URL" --connect-timeout 10 --max-time 15 \
                "$url" 2>/dev/null || echo "000")
        fi
    else
        if [ -n "$API_KEY" ]; then
            http_code=$(curl -s -o /dev/null -w "%{http_code}" \
                --connect-timeout 5 --max-time 10 \
                -H "Authorization: Bearer ${API_KEY}" \
                "$url" 2>/dev/null || echo "000")
        else
            http_code=$(curl -s -o /dev/null -w "%{http_code}" \
                --connect-timeout 5 --max-time 10 \
                "$url" 2>/dev/null || echo "000")
        fi
    fi

    end_time=$(date +%s%N 2>/dev/null || python3 -c "import time; print(int(time.time()*1e9))")
    duration=$(( (end_time - start_time) / 1000000 ))

    local status
    if [ "$http_code" = "000" ]; then
        status="${RED}TIMEOUT${NC}"
    elif [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        status="${GREEN}OK (${http_code})${NC}"
    elif [ "$http_code" -ge 401 ]; then
        status="${YELLOW}AUTH (${http_code})${NC}"
    else
        status="${RED}FAIL (${http_code})${NC}"
    fi

    printf "%-12s %-8s %s (%dms)\n" "$name" "$mode" "$status" "$duration"
}

echo "--- Via Proxy ---"
for ep in "${endpoints[@]}"; do
    IFS='|' read -r name url <<< "$ep"
    test_endpoint "$name" "$url" "proxy"
done

echo ""
echo "--- Direct (expected to timeout for overseas) ---"
for ep in "${endpoints[@]}"; do
    IFS='|' read -r name url <<< "$ep"
    test_endpoint "$name" "$url" "direct"
done

echo ""
echo "=== Internal Services (should be DIRECT) ---"
printf "%-12s %-8s " "PostgreSQL" "direct"
if nc -z localhost 5432 2>/dev/null; then
    echo -e "${GREEN}OK${NC}"
else
    echo -e "${YELLOW}NOT REACHABLE (may be in Docker)${NC}"
fi

printf "%-12s %-8s " "Redis" "direct"
if nc -z localhost 6379 2>/dev/null; then
    echo -e "${GREEN}OK${NC}"
else
    echo -e "${YELLOW}NOT REACHABLE (may be in Docker)${NC}"
fi
