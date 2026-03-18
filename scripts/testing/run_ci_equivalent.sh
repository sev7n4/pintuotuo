#!/usr/bin/env bash
set -euo pipefail

ROOT="/Users/4seven/pintuotuo"
cd "$ROOT"

docker-compose -f docker-compose.test.yml up -d

timeout=60
while ! docker exec pintuotuo_postgres_test pg_isready -U pintuotuo -d pintuotuo_db >/dev/null 2>&1; do
  ((timeout--)) || { echo "Postgres not ready"; exit 1; }
  sleep 1
done

docker exec pintuotuo_postgres_test psql -U pintuotuo -d pintuotuo_db -f /docker-entrypoint-initdb.d/01-full_schema.sql

export TEST_MODE=true
export DATABASE_URL=postgresql://pintuotuo:dev_password_123@localhost:5433/pintuotuo_db?sslmode=disable
export REDIS_URL=redis://localhost:6380
export JWT_SECRET=pintuotuo-secret-key-dev
export GIN_MODE=release

cd "$ROOT/backend"
go mod download
go mod verify
go test -v -count=1 -p 1 ./...
go test -v -count=1 -p 1 ./tests/integration -timeout 120s

if [ -d "$ROOT/frontend" ]; then
  cd "$ROOT/frontend"
  npm ci
  npm run build || true
  CI=true npm test -- --watchAll=false || true
fi

cd "$ROOT"
docker-compose -f docker-compose.test.yml down

