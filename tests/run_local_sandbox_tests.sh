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

export TEST_MODE=true
export DATABASE_URL=postgresql://pintuotuo:dev_password_123@localhost:5433/pintuotuo_db?sslmode=disable
export REDIS_URL=redis://localhost:6380
export JWT_SECRET=pintuotuo-secret-key-dev
export GIN_MODE=release

cd "$ROOT/backend"
go clean -testcache
go test -v -count=1 -p 1 ./...
go test -v -count=1 -p 1 ./tests/integration

if [ -d "$ROOT/frontend" ]; then
  cd "$ROOT/frontend"
  npm ci
  CI=true npm test -- --watchAll=false || true

  # Start frontend server for E2E tests
  echo "Starting frontend server for E2E tests..."
  npm run dev &
  FRONTEND_PID=$!

  # Wait for the frontend server to be ready
  echo "Waiting for frontend server to be ready..."
  sleep 10

  # Run E2E tests
  echo "Running E2E tests..."
  # Use system Chrome (configured in playwright.config.ts) to avoid downloading browsers
  npx playwright test

  # Stop the frontend server
  echo "Stopping frontend server..."
  kill $FRONTEND_PID
fi

cd "$ROOT"
docker-compose -f docker-compose.test.yml down


