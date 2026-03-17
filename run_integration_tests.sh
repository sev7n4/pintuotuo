#!/bin/bash
# Pintuotuo Integration Test Runner

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}Starting integration tests for Pintuotuo...${NC}"

# Check if test containers are running
if ! docker ps | grep -q "pintuotuo_postgres_test"; then
    echo -e "${RED}Error: Test containers are not running. Run 'docker-compose -f docker-compose.test.yml up -d' first.${NC}"
    exit 1
fi

# Set environment variables for tests (targeting test containers)
export DATABASE_URL=postgresql://pintuotuo:dev_password_123@localhost:5433/pintuotuo_db?sslmode=disable
export REDIS_URL=redis://localhost:6380
export JWT_SECRET=pintuotuo-secret-key-dev
export GIN_MODE=release

# Run tests
echo -e "${BLUE}Running backend integration tests...${NC}"
cd backend && go test -v -count=1 ./tests/integration/...
TEST_EXIT_CODE=$?

if [ $TEST_EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}✓ All integration tests passed!${NC}"
else
    echo -e "${RED}✗ Integration tests failed!${NC}"
    exit $TEST_EXIT_CODE
fi
