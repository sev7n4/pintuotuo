#!/usr/bin/env bash
# 在本机执行：SSH 到腾讯云后查 model_providers.code 与 spus.model_provider，用于核对前端 providerBrand 映射是否全覆盖。
# 用法：bash scripts/audit-prod-model-providers.sh
# 依赖：~/.ssh/tencent_cloud_deploy 可登录 root@119.29.173.89；服务器上 Docker 内有 Postgres 容器。

set -euo pipefail
KEY="${SSH_KEY:-$HOME/.ssh/tencent_cloud_deploy}"
HOST="${DEPLOY_HOST:-119.29.173.89}"
USER="${DEPLOY_USER:-root}"

ssh -i "$KEY" -o BatchMode=yes -o StrictHostKeyChecking=accept-new -o ConnectTimeout=25 "$USER@$HOST" 'bash -s' <<'REMOTE'
set -euo pipefail
echo "=== hostname ==="
hostname

echo ""
echo "=== postgres containers (name contains postgres) ==="
docker ps --format '{{.Names}}' | grep -i postgres || true

PGC=$(docker ps --format '{{.Names}}' | grep -i postgres | head -1 || true)
if [ -z "$PGC" ]; then
  echo "ERROR: no container name matching postgres; run docker ps manually."
  exit 1
fi
echo "Using container: $PGC"

echo ""
echo "=== model_providers.code (DISTINCT) ==="
docker exec -i "$PGC" psql -U pintuotuo -d pintuotuo_db -c \
  "SELECT DISTINCT code FROM model_providers ORDER BY 1;"

echo ""
echo "=== spus.model_provider (DISTINCT, non-empty) ==="
docker exec -i "$PGC" psql -U pintuotuo -d pintuotuo_db -c \
  "SELECT DISTINCT model_provider FROM spus WHERE model_provider IS NOT NULL AND btrim(model_provider) <> '' ORDER BY 1;"

echo ""
echo "=== spus.model_provider NOT IN model_providers.code ==="
docker exec -i "$PGC" psql -U pintuotuo -d pintuotuo_db -c \
  "SELECT DISTINCT s.model_provider
   FROM spus s
   WHERE s.model_provider IS NOT NULL AND btrim(s.model_provider) <> ''
     AND NOT EXISTS (
       SELECT 1 FROM model_providers mp WHERE mp.code = s.model_provider
     )
   ORDER BY 1;"
REMOTE
