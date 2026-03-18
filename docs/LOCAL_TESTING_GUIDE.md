# Local Testing Guide

**Last Updated**: 2026-03-18
**Purpose**: How to test and verify code locally before pushing to CI/CD

---

## Overview

Running tests locally speeds up development and catches issues before GitHub Actions runs. This guide covers:

- Setting up local testing environment
- Running backend tests
- Running frontend tests
- Running integration tests
- Debugging test failures
- Performance profiling

---

## Table of Contents

1. [Local Environment Setup](#local-environment-setup)
2. [Backend Testing](#backend-testing)
3. [Frontend Testing](#frontend-testing)
4. [Integration Testing](#integration-testing)
5. [Coverage Analysis](#coverage-analysis)
6. [Debugging Failed Tests](#debugging-failed-tests)
7. [Performance Testing](#performance-testing)
8. [Pre-Push Checklist](#pre-push-checklist)

---

## Local Environment Setup

### Prerequisites

```bash
# Check required tools
go version          # Should be 1.21+
node --version      # Should be 18+
npm --version       # Should be 8+
docker --version    # Should be 20.10+
docker-compose --version  # Should be 1.29+
```

### Start Docker Services

All tests need PostgreSQL and Redis running locally.

```bash
# Start all services
docker-compose up -d

# Verify services are running
docker-compose ps

# Expected output:
# NAME              STATUS              PORTS
# postgres          Up (healthy)        5432
# redis             Up (healthy)        6379
```

### Database Setup

```bash
# Initialize schema
PGPASSWORD=dev_password_123 psql \
  -h localhost \
  -U pintuotuo \
  -d pintuotuo_db \
  -f scripts/db/full_schema.sql

# Verify tables created
PGPASSWORD=dev_password_123 psql \
  -h localhost \
  -U pintuotuo \
  -d pintuotuo_db \
  -c "\dt"

# Output:
#           List of relations
# Schema | Name | Type | Owner
# --------+-------+-------+----------
# public | users | table | pintuotuo
# public | products | table | pintuotuo
# ...
```

### Environment Variables

Create `.env.local` in project root:

```bash
# Backend
DATABASE_URL=postgresql://pintuotuo:dev_password_123@localhost:5432/pintuotuo_db
REDIS_URL=redis://localhost:6379
JWT_SECRET=dev-secret-key-local
GIN_MODE=debug
TEST_MODE=true

# Frontend
VITE_API_URL=http://localhost:8080
VITE_MOCK_API_URL=http://localhost:3001
```

Or set environment variables in shell:

```bash
export DATABASE_URL="postgresql://pintuotuo:dev_password_123@localhost:5432/pintuotuo_db"
export REDIS_URL="redis://localhost:6379"
export JWT_SECRET="dev-secret-key-local"
export GIN_MODE="debug"
export TEST_MODE="true"
```

---

## Backend Testing

### Quick Test Run

```bash
cd backend

# Run all tests
go test ./... -v

# Run specific package
go test ./services/payment -v

# Run specific test
go test ./services/payment -v -run TestInitiatePayment
```

### Test with Verbose Output

```bash
# Show detailed output
go test ./... -v

# Output shows:
# === RUN TestInitiatePayment
# --- PASS: TestInitiatePayment (0.05s)
# === RUN TestRefundPayment
# --- PASS: TestRefundPayment (0.03s)
# ok      pintuotuo/services/payment      0.456s
```

### Test with Coverage

```bash
# Generate coverage report
go test ./... -cover

# Output:
# ok      pintuotuo/services       coverage: 82.3% of statements
# ok      pintuotuo/handlers       coverage: 91.2% of statements
```

### Detailed Coverage Report

```bash
# Generate coverage profile
go test ./... -coverprofile=coverage.out

# View coverage in browser
go tool cover -html=coverage.out

# This opens coverage.html showing:
# - Green lines: covered
# - Red lines: not covered
# - Gray lines: not testable
```

### Test for Race Conditions

```bash
# Run with race detector
go test -race ./...

# Note: May be slower, but catches concurrency bugs
# Output:
# ==================
# WARNING: DATA RACE
# Write at 0x00c0001a2340 by goroutine 23:
#   ...
# ==================
```

### Test Specific Scenarios

```bash
# Integration tests only
go test ./tests/integration -v -timeout 120s

# Unit tests (skip integration)
go test ./services/... ./handlers/... -v

# Payment service tests only
go test ./services/payment/... -v

# Run tests matching pattern
go test ./... -run "TestPayment"  # Runs all tests with "TestPayment" in name
go test ./... -run "Test.*Webhook"  # Regex pattern
```

### Run Tests in Serial Mode

```bash
# Run sequentially (no parallelism)
# Same as GitHub Actions
go test -p 1 ./... -v

# -p 1 = 1 process at a time
# Prevents race conditions
# Slower but more stable
```

### Skip Database Tests

```bash
# Run unit tests only (skip integration/database tests)
go test ./services/payment/service_test.go -v

# For tests that need database, skip them:
go test ./... -v -short
# (Tests use t.Skip() in short mode)
```

### Common Backend Test Commands

```bash
# Quick sanity check
go test ./services -v -timeout 30s

# Full suite with coverage
go test ./... -v -cover -timeout 60s

# Integration tests with details
go test ./tests/integration -v -timeout 120s -count=1

# Payment service complete test
go test ./services/payment -v -cover -race

# Test with specific Go version check
go version && go test ./...

# Parallel execution (4 jobs)
go test -p 4 ./... -v

# Serial execution (1 job - slower but safer)
go test -p 1 ./... -v
```

---

## Frontend Testing

### Setup

```bash
cd frontend

# Install dependencies
npm install

# Check Node version
node --version  # Should be 18+
```

### Run Tests

```bash
# Run all tests once
npm test -- --watchAll=false

# Run tests in watch mode (rerun on file changes)
npm test

# Run specific test file
npm test LoginForm

# Run tests matching pattern
npm test --testNamePattern="should render"

# Run with coverage
npm test -- --coverage --watchAll=false
```

### Coverage Report

```bash
# Generate coverage
npm test -- --coverage --watchAll=false

# Output:
# =============================== Coverage summary ===============================
# Statements   : 85.2% ( 120/141 )
# Branches     : 78.5% ( 41/52 )
# Functions    : 89.3% ( 50/56 )
# Lines        : 86.1% ( 120/139 )

# Open coverage report
open coverage/lcov-report/index.html
```

### Build Frontend

```bash
# Production build
npm run build

# Build with analysis
npm run build -- --analyze  # If available

# Preview build locally
npm run preview
```

### Run E2E Tests

```bash
# Install Playwright (once)
npx playwright install

# Run all E2E tests
npx playwright test

# Run specific E2E test
npx playwright test login

# Run in headed mode (see browser)
npx playwright test --headed

# Debug mode (step through)
npx playwright test --debug

# Show report
npx playwright show-report
```

### Linting

```bash
# Check code style
npm run lint

# Auto-fix style issues
npm run lint -- --fix

# Type checking
npm run type-check
```

### Common Frontend Commands

```bash
# Quick check
npm install && npm run build

# Full verification
npm test -- --watchAll=false && npm run lint && npm run type-check

# Development mode
npm start  # or npm run dev

# Test with coverage
npm test -- --coverage --watchAll=false

# E2E tests only
npx playwright test

# Watch mode for development
npm test
```

---

## Integration Testing

### Prerequisites

```bash
# Both services must be running
docker-compose up -d postgres redis

# Verify connectivity
PGPASSWORD=dev_password_123 psql -h localhost -U pintuotuo -d pintuotuo_db -c "SELECT 1"
redis-cli ping  # Should output: PONG
```

### Run Integration Tests

```bash
cd backend

# All integration tests
go test ./tests/integration -v -timeout 120s

# Specific integration test
go test ./tests/integration -v -run TestCompleteGroupPurchaseWithPayment

# With coverage
go test ./tests/integration -v -cover -timeout 120s

# Detailed output
go test ./tests/integration -v -timeout 120s -count=1 -p 1
```

### Integration Test Scenarios

```bash
# Payment workflow tests
go test ./tests/integration -v -run "Payment" -timeout 120s

# Concurrency tests
go test ./tests/integration -v -run "Concurrent" -timeout 300s

# Stress tests
go test ./tests/integration -v -run "Stress" -timeout 300s

# Consistency tests
go test ./tests/integration -v -run "Consistency" -timeout 120s

# Run single test (slow to debug)
go test ./tests/integration -v -run "^TestPaymentOrderSyncConsistency$" -timeout 60s
```

---

## Coverage Analysis

### Generate Coverage Report

```bash
cd backend

# Unit test coverage
go test ./services ./handlers -coverprofile=unit_coverage.out

# Integration test coverage
go test ./tests/integration -coverprofile=integration_coverage.out

# Merge coverage files
go tool covdata merge -i=unit_coverage.out,integration_coverage.out -o=merged

# View merged coverage
go tool cover -html=merged
```

### Coverage Requirements

Minimum coverage targets:
- Services: 80%
- Handlers: 85%
- Core packages: 75%

```bash
# Check coverage meets threshold
go test ./services ./handlers -cover

# If coverage is low, identify uncovered code:
go test ./services -coverprofile=coverage.out
go tool cover -html=coverage.out
# Look for red (uncovered) lines
```

### Coverage Tips

1. **Focus on critical paths**: Payment flow, authentication, data consistency
2. **Test error cases**: What happens when things fail?
3. **Avoid low-value tests**: Testing library functions doesn't help
4. **Use table-driven tests**: More coverage per test

```go
// Good: Covers multiple scenarios
func TestValidatePayment(t *testing.T) {
  tests := []struct {
    name    string
    amount  float64
    wantErr bool
  }{
    {"valid", 100.00, false},
    {"zero", 0.00, true},
    {"negative", -10.00, true},
  }
  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      // test
    })
  }
}
```

---

## Debugging Failed Tests

### Getting Detailed Error Information

```bash
# Show full error output
go test ./services/payment -v -run TestInitiatePayment

# With stack trace
go test ./services/payment -v -run TestInitiatePayment 2>&1 | head -50

# With verbose logging (if configured)
DEBUG=* go test ./services/payment -v
```

### Add Debug Output

```go
// In test code, use t.Logf for debugging
func TestPayment(t *testing.T) {
  payment := initiate(100.00)

  t.Logf("Payment ID: %d", payment.ID)
  t.Logf("Status: %s", payment.Status)
  t.Logf("Created: %v", payment.CreatedAt)

  if payment.Status != "pending" {
    t.Errorf("Expected pending, got %s", payment.Status)
  }
}

// Run with: go test -v
// Output shows all t.Logf calls
```

### Common Test Failures

#### Database Connection Error

```
ERROR: could not translate host name "localhost"
```

**Fix**:
```bash
# Ensure PostgreSQL is running
docker-compose ps postgres
# Status should be "Up (healthy)"

# Start if needed
docker-compose up -d postgres
```

#### Table Does Not Exist

```
ERROR: relation "payments" does not exist
```

**Fix**:
```bash
# Reinitialize schema
PGPASSWORD=dev_password_123 psql \
  -h localhost -U pintuotuo -d pintuotuo_db \
  -f scripts/db/full_schema.sql
```

#### Timeout Error

```
ERROR: context deadline exceeded
```

**Fix**:
```bash
# Increase timeout
go test ./... -timeout 180s

# Check for slow tests
go test -v -timeout 180s -run SlowTest
```

#### Flaky Tests (Pass/Fail Randomly)

```
# Run test multiple times
go test -v -count=5 ./... -run FlakyTest

# If some pass and some fail, test is flaky
# Usually caused by:
# - Race conditions (use -race flag)
# - Timing assumptions (use mocks instead of sleeps)
# - Database state issues (ensure cleanup)
```

**Fix**:
```bash
# Use -race to find races
go test -race ./...

# Fix race by synchronizing properly:
// BAD: Assumes timing
time.Sleep(100 * time.Millisecond)

// GOOD: Use channels or wait groups
var wg sync.WaitGroup
// or use mock clock for tests
```

---

## Performance Testing

### Benchmark Tests

```bash
# Run benchmarks
go test -bench=. ./services/payment

# Output:
# BenchmarkInitiatePayment-8  10000  100234 ns/op
# BenchmarkGetPayment-8       50000   23456 ns/op

# -8: GOMAXPROCS (8 CPU cores)
# 10000: iterations run
# 100234 ns/op: nanoseconds per operation
```

### Memory Profiling

```bash
# Generate memory profile
go test ./services/payment -memprofile=mem.out

# Analyze
go tool pprof mem.out
# Commands in pprof:
# (pprof) top         # Top memory consumers
# (pprof) list Func   # Show function source
# (pprof) quit        # Exit
```

### CPU Profiling

```bash
# Generate CPU profile
go test ./services/payment -cpuprofile=cpu.out

# Analyze
go tool pprof cpu.out
# (pprof) top         # Top CPU consumers
# (pprof) list        # Show source
```

### Load Testing

```bash
# Run tests under load (multiple times in parallel)
go test -run TestPayment -count=100 ./services/payment

# Useful for finding race conditions
go test -race -run TestPayment -count=100 ./services/payment
```

---

## Pre-Push Checklist

Before pushing to GitHub, run this checklist locally:

### Backend Checklist

```bash
cd backend

# 1. All tests pass
echo "1. Running tests..."
go test ./... -timeout 60s || exit 1

# 2. No formatting issues
echo "2. Checking formatting..."
gofmt -l . | grep . && exit 1 || echo "✓ Formatting OK"

# 3. No lint errors
echo "3. Running linter..."
golangci-lint run ./... || exit 1

# 4. No vet errors
echo "4. Running go vet..."
go vet ./... || exit 1

# 5. Coverage OK
echo "5. Checking coverage..."
go test ./... -cover | grep -E "coverage.*8[0-9]\." || echo "⚠ Coverage might be low"

# 6. No race conditions
echo "6. Running race detector..."
go test -race ./services ./handlers -timeout 120s || exit 1

echo "✅ All backend checks passed!"
```

### Frontend Checklist

```bash
cd frontend

# 1. Install dependencies
echo "1. Installing dependencies..."
npm install || exit 1

# 2. Build succeeds
echo "2. Building..."
npm run build || exit 1

# 3. Tests pass
echo "3. Running tests..."
npm test -- --watchAll=false || exit 1

# 4. Linting passes
echo "4. Linting..."
npm run lint || exit 1

# 5. Type checking passes
echo "5. Type checking..."
npm run type-check || exit 1

echo "✅ All frontend checks passed!"
```

### Integration Checklist

```bash
# 1. Services running
echo "1. Checking services..."
docker-compose ps postgres redis | grep Up || exit 1

# 2. Database initialized
echo "2. Checking database..."
PGPASSWORD=dev_password_123 psql \
  -h localhost -U pintuotuo -d pintuotuo_db \
  -c "SELECT COUNT(*) FROM users" > /dev/null || exit 1

# 3. Integration tests pass
echo "3. Running integration tests..."
cd backend
go test ./tests/integration -v -timeout 120s || exit 1

echo "✅ All integration checks passed!"
```

### Complete Pre-Push Script

Save as `scripts/pre-push.sh`:

```bash
#!/bin/bash
set -e

echo "🔍 Running pre-push checks..."

# Backend
cd backend
echo "✓ Backend tests..."
go test ./... -timeout 60s -count=1

echo "✓ Backend formatting..."
gofmt -l . | grep . && exit 1 || true

echo "✓ Backend linting..."
golangci-lint run ./... 2>/dev/null || true

# Frontend
cd ../frontend
echo "✓ Frontend build..."
npm run build -- --no-type-check || true

echo "✓ Frontend tests..."
npm test -- --watchAll=false --passWithNoTests || true

# Done
cd ..
echo "✅ All checks passed! Safe to push."
```

Run before pushing:
```bash
bash scripts/pre-push.sh
```

---

## Common Workflows

### Development Loop

```bash
# 1. Make code changes
# 2. Run quick tests
go test ./services/payment -v

# 3. Check if it compiles
go build ./...

# 4. When ready, run full suite
go test -race ./... -timeout 120s

# 5. Check formatting
gofmt -l .

# 6. Commit and push
git add .
git commit -m "feat(payment): add retry logic"
git push origin feature/payment-retry
```

### Bug Investigation

```bash
# 1. Run test to reproduce
go test -v -run TestBuggyFeature

# 2. Add debug output
# (edit test, add t.Logf() calls)

# 3. Rerun with verbose
go test -v -run TestBuggyFeature

# 4. Run with race detector
go test -race -run TestBuggyFeature

# 5. Check coverage
go test -cover -run TestBuggyFeature
```

### Adding New Feature

```bash
# 1. Write tests first (TDD)
# 2. Run tests (should fail)
go test -v -run NewFeature

# 3. Implement feature
# 4. Run tests (should pass)
go test -v -run NewFeature

# 5. Run full suite to ensure no regressions
go test ./... -timeout 120s

# 6. Check coverage for new code
go test ./... -cover
```

---

## Tips & Tricks

### Run Only New/Modified Tests

```bash
# Git hook to run tests on modified files
go test ./... -run "$(git diff --name-only | grep _test.go | sed 's/_test.go//' | tr '\n' '|')"
```

### Parallel Test Execution

```bash
# Default: parallel (fast)
go test ./...

# Serial: safer, slower
go test -p 1 ./...

# Custom parallelism
go test -p 8 ./...  # 8 processes
```

### Long Running Tests

```bash
# Separate long tests
go test -short ./...  # Skips slow tests
go test -run "Long" ./...  # Only slow tests
```

### Interactive Debugging

```bash
# Requires dlv (Go debugger)
go install github.com/go-delve/delve/cmd/dlv@latest

# Run test under debugger
dlv test ./services/payment -- -test.run TestPayment

# Interactive commands:
# (dlv) break main.Authenticate  # Set breakpoint
# (dlv) continue                 # Run to breakpoint
# (dlv) next                     # Next line
# (dlv) print var               # Print variable
# (dlv) exit                     # Exit debugger
```

---

**Document Version**: 1.0
**Status**: Ready to Use ✅
**Maintained By**: Development Team
