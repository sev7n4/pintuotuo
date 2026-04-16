#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
CONFIG_PATH="$ROOT_DIR/evals/promptfoo/promptfooconfig.yaml"
REPORT_DIR="$ROOT_DIR/evals/promptfoo/reports"

mkdir -p "$REPORT_DIR"

if ! command -v npx >/dev/null 2>&1; then
  echo "npx is required to run promptfoo"
  exit 1
fi

if [ -z "${PROMPTFOO_BASE_URL:-}" ] || [ -z "${PROMPTFOO_API_KEY:-}" ] || [ -z "${PROMPTFOO_MODEL:-}" ]; then
  echo "PROMPTFOO_BASE_URL, PROMPTFOO_API_KEY, PROMPTFOO_MODEL are required"
  exit 1
fi

npx -y promptfoo@latest eval \
  -c "$CONFIG_PATH" \
  --output "$REPORT_DIR/promptfoo-report.json"

echo "Prompt regression completed: $REPORT_DIR/promptfoo-report.json"
