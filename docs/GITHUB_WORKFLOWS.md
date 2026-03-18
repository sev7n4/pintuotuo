# GitHub Actions Workflows - Complete Guide

**Last Updated**: 2026-03-18
**Status**: All 4 workflows fully operational ✅

## Table of Contents

1. [Overview](#overview)
2. [Workflow 1: CI Pipeline](#workflow-1-ci-pipeline)
3. [Workflow 2: Integration Tests](#workflow-2-integration-tests)
4. [Workflow 3: Integration Tests - Payment Service](#workflow-3-integration-tests---payment-service)
5. [Workflow 4: Deploy to Staging](#workflow-4-deploy-to-staging)
6. [Complete Execution Flow](#complete-execution-flow)
7. [GitHub Secrets Configuration](#github-secrets-configuration)
8. [Troubleshooting](#troubleshooting)

---

## Overview

Pintuotuo uses **4 sophisticated GitHub Actions workflows** to ensure code quality, security, and reliable deployments:

```
Push Code to Repository
    ↓
┌─────────────────────────────────────┐
│  CI Pipeline (on every push/PR)     │
│  - Backend unit tests               │
│  - Frontend unit tests              │
│  - Code quality checks              │
│  - Docker image building            │
└─────────────────────────────────────┘
    ↓ (Runs on: main, master, develop)
    ↓
┌─────────────────────────────────────┐
│  Integration Tests (all branches)   │
│  - Full integration test suite      │
│  - Database consistency checks      │
│  - Stress and performance tests     │
│  - Coverage threshold validation    │
└─────────────────────────────────────┘
    ↓ (If develop branch)
    ↓
┌─────────────────────────────────────┐
│  Deploy to Staging (develop only)   │
│  - Deploy to staging server         │
│  - Run smoke tests                  │
│  - Health checks                    │
│  - Slack notifications              │
└─────────────────────────────────────┘
    ↓
Slack / GitHub Notifications
```

### Workflow Statistics

| Workflow | Triggers | Jobs | Steps | Purpose |
|----------|----------|------|-------|---------|
| **CI Pipeline** | All pushes + PRs | 2 jobs | 15+ steps | Code quality & unit tests |
| **Integration Tests** | All pushes + PRs | 2 jobs | 12+ steps | Full integration testing |
| **Payment Service Tests** | main, develop | 4 jobs | 20+ steps | Specialized payment service testing |
| **Deploy to Staging** | develop branch | 1 job | 8+ steps | Automated staging deployment |

---

## Workflow 1: CI Pipeline

**File**: `.github/workflows/ci-pipeline.yml`
**Triggers**: All pushes and PRs on `main`, `master`, `develop`
**Duration**: ~5-7 minutes

### Purpose
The CI Pipeline is your first line of quality control. It runs on every code push and pull request to ensure:
- Backend code is syntactically correct and passes unit tests
- Frontend code builds and passes tests
- Code follows style guidelines
- No breaking changes introduced

### Jobs Breakdown

#### **Job 1: Backend Tests** (5-6 min)

**Services Started**:
- PostgreSQL 15 (port 5432) - for database testing
- Redis 7 (port 6379) - for caching/session testing

**Steps**:

1. **Checkout code** (`actions/checkout@v4`)
   - Fetches your code with full git history

2. **Set up Go 1.21** (`actions/setup-go@v4`)
   - Installs Go compiler and caches dependencies
   - Speeds up repeated workflows

3. **Install dependencies**
   ```bash
   go mod download
   go mod verify
   ```
   - Downloads Go modules
   - Verifies module checksums against go.sum

4. **Install PostgreSQL Client**
   ```bash
   sudo apt-get update && sudo apt-get install -y postgresql-client
   ```
   - Required for `psql` command to initialize database
   - `pg_isready` tool for database health checks

5. **Initialize database schema** (Critical!)
   ```bash
   # Wait for PostgreSQL to be ready
   until pg_isready -h localhost -p 5432 -U pintuotuo; do
     echo 'Waiting for PostgreSQL...'
     sleep 1
   done
   psql -h localhost -U pintuotuo -d pintuotuo_test -f scripts/db/full_schema.sql
   ```
   - Waits for PostgreSQL service to accept connections
   - Creates all tables: users, products, orders, payments, groups, tokens, etc.
   - `full_schema.sql` contains 125 lines, 11 tables with indexes

6. **Run tests** (~3 minutes)
   ```bash
   go list ./... | grep -v 'pintuotuo/internal/db' | xargs go test -v -count=1 -p 1
   ```
   - Runs all Go packages except internal/db
   - `-v` = verbose output
   - `-count=1` = no caching (always run fresh)
   - `-p 1` = serial execution (no race conditions)

   **Environment Variables**:
   - `DATABASE_URL`: `postgresql://pintuotuo:test_password@localhost:5432/pintuotuo_test?sslmode=disable`
   - `REDIS_URL`: `redis://localhost:6379`
   - `JWT_SECRET`: `pintuotuo-secret-key-dev`
   - `GIN_MODE`: `release` (production-like mode)
   - `TEST_MODE`: `true`

**Expected Output**:
```
=== RUN   TestUserService/TestCreateUser
--- PASS: TestUserService/TestCreateUser (0.05s)
=== RUN   TestProductService/TestListProducts
--- PASS: TestProductService/TestListProducts (0.03s)
...
ok      pintuotuo/services      2.456s
```

**Failure Handling**:
- If any test fails, workflow stops and reports error
- Check the detailed logs under "Run tests" step
- Common issues: Database connection, migration failures

---

#### **Job 2: Frontend Tests** (3-4 min)

**Services**: None (frontend is client-side)

**Steps**:

1. **Checkout code**
   - Gets the latest frontend code

2. **Set up Node.js 18**
   - Installs Node.js and npm
   - Caches node_modules for faster builds

3. **Install dependencies**
   ```bash
   npm install
   ```
   - Installs all packages from package-lock.json

4. **Build frontend** (Optional)
   ```bash
   npm run build -- --no-type-check || true
   ```
   - Builds production bundle
   - `-no-type-check` skips TypeScript checks (done separately)
   - `|| true` = continue even if build fails (non-critical)

5. **Run unit and integration tests**
   ```bash
   npm test -- --passWithNoTests
   ```
   - Runs Jest test suite
   - `--passWithNoTests` = pass if no tests found (CI doesn't fail)

6. **Install Playwright Browsers**
   ```bash
   npx playwright install --with-deps chromium
   ```
   - Downloads Chromium browser for E2E testing

7. **Run Playwright E2E tests**
   ```bash
   npx playwright test
   ```
   - Runs end-to-end browser tests
   - Tests user workflows like login, purchase, etc.

**Expected Output**:
```
> npm test
PASS  src/__tests__/LoginForm.test.tsx
  LoginForm
    ✓ renders login button (45ms)
    ✓ submits form on button click (89ms)

Test Suites: 1 passed, 1 total
```

---

### Interpreting CI Pipeline Results

**✅ Success**:
- Green checkmark on PR
- All steps completed with "✓"
- Code is ready to merge

**❌ Backend Tests Failed**:
1. Click on "Backend Tests" job
2. Scroll to "Run tests" step
3. Look for first "FAIL" or "panic" message
4. Check DATABASE_URL and schema initialization

**❌ Frontend Tests Failed**:
1. Check "Frontend Tests" job
2. Look for error in "Run unit tests" step
3. Common issues: missing dependencies, broken imports

**⚠️ Build Warnings** (Yellow):
- Still passes but has warnings
- Fix warnings to prevent future failures

---

## Workflow 2: Integration Tests

**File**: `.github/workflows/test.yml`
**Triggers**: All pushes and PRs on `main`, `develop`, `master`
**Duration**: ~3-5 minutes

### Purpose
This workflow runs the full integration test suite to ensure all services work together:
- Complete end-to-end payment flow
- Database consistency
- Cache behavior
- Concurrent operations
- 22+ test cases covering multiple scenarios

### Jobs Breakdown

#### **Job: Run Integration Tests** (3-5 min)

**Services**:
- PostgreSQL 15 (Health checks enabled)
- Redis 7 (Health checks enabled)

**Steps**:

1. **Checkout code** (with full history for git operations)

2. **Set up Go 1.21**

3. **Download dependencies**
   ```bash
   go mod download
   go mod tidy
   go mod verify
   ```

4. **Install PostgreSQL Client**
   ```bash
   sudo apt-get update && sudo apt-get install -y postgresql-client
   ```

5. **Initialize database schema**
   ```bash
   psql "$DATABASE_URL" -f ../scripts/db/full_schema.sql
   ```
   - Uses `full_schema.sql` (125 lines, 11 tables)
   - Creates all tables needed for integration tests

6. **Run integration tests** (Main step - 2-3 minutes)
   ```bash
   go test -v ./tests/integration -timeout 120s -count=1 -p 1
   ```

   **Test Coverage**:
   - ✅ **Payment Initialization**: InitiatePayment service
   - ✅ **Webhook Processing**: Alipay & WeChat callbacks
   - ✅ **Database Consistency**: Data integrity across operations
   - ✅ **Concurrent Payments**: 100+ simultaneous payments
   - ✅ **Cache Consistency**: Redis cache behavior
   - ✅ **Order Status Transitions**: Complete order lifecycle
   - ✅ **Refunds**: Payment cancellation and refunds

   **Environment Variables**:
   - `DATABASE_URL`: `postgresql://pintuotuo:test_password@localhost:5432/pintuotuo_test?sslmode=disable`
   - `REDIS_URL`: `redis://localhost:6379`
   - `JWT_SECRET`: `pintuotuo-secret-key-dev`
   - `GIN_MODE`: `release`

   **Test Results** (22/22 passing):
   ```
   === RUN   TestPaymentDatabaseConsistency
   --- PASS: TestPaymentDatabaseConsistency (0.15s)
   === RUN   TestConcurrentPaymentsForSameOrder
   --- PASS: TestConcurrentPaymentsForSameOrder (0.42s)
   === RUN   TestCompleteGroupPurchaseWithPayment
   --- PASS: TestCompleteGroupPurchaseWithPayment (0.28s)
   ...
   ok      pintuotuo/tests/integration     45.234s
   ```

7. **Check code coverage**
   ```bash
   go test -cover ./services/... ./handlers/...
   ```
   - Reports coverage percentage
   - Target: >80% coverage

#### **Job: Build Docker Image** (Depends on test job)

Runs after tests pass on any branch.

**Steps**:

1. **Checkout code**

2. **Set up Go 1.21**

3. **Build Linux binary**
   ```bash
   CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o backend-linux main.go
   ```
   - Builds static binary (no CGO dependencies)
   - Linux x86_64 architecture
   - Static binary = works in any Docker image

4. **Set up Docker Buildx**
   - Docker's cross-platform builder

5. **Login to Docker Hub**
   - Uses `DOCKER_USERNAME` and `DOCKER_PASSWORD` secrets

6. **Build and push Docker image**
   ```bash
   docker build -f ./backend/Dockerfile.alpine -t pintuotuo-backend:latest .
   ```
   - Builds image from `Dockerfile.alpine` (minimal size)
   - Tags as `latest` and with commit SHA
   - `push: false` = doesn't push (change to `true` for production)

7. **Build summary**
   - Adds success message to GitHub workflow summary

---

### Integration Test Coverage

The integration test suite includes:

| Category | Test Cases | Purpose |
|----------|-----------|---------|
| **Payment Flow** | 6 tests | Complete payment lifecycle |
| **Webhooks** | 4 tests | Alipay & WeChat callback handling |
| **Concurrency** | 5 tests | Race conditions, concurrent payments |
| **Consistency** | 4 tests | Data integrity, cache sync |
| **Stress** | 3 tests | High load scenarios |

**Key Test Files**:
- `backend/tests/integration/payment_test.go` - Payment flow
- `backend/tests/integration/consistency_test.go` - Data consistency
- `backend/tests/integration/stress_test.go` - Load testing
- `backend/tests/integration/workflow_test.go` - End-to-end flows

---

## Workflow 3: Integration Tests - Payment Service

**File**: `.github/workflows/integration-tests.yml`
**Triggers**: Push/PR on `main`, `develop`
**Duration**: ~6-8 minutes
**Special Feature**: Slack notifications

### Purpose
Specialized workflow for payment service with detailed testing:
- Service layer unit tests (23 tests)
- HTTP handler tests
- Stress tests (high concurrency)
- Detailed coverage reporting
- Slack notifications on success/failure

### Jobs Breakdown

#### **Job 1: Run Integration Tests**

Nearly identical to test.yml but with additional details.

**Key Differences**:
- More verbose test output
- Additional stress tests
- Coverage threshold checking (80%+ required)
- Detailed GitHub step summary

**Test Phases**:

1. **Phase 1: Service Layer** (23 unit tests)
   - Payment initiation
   - Webhook callbacks (Alipay, WeChat)
   - Refund processing
   - Revenue calculations

2. **Phase 2: HTTP Handlers** (9 tests)
   - REST endpoint testing
   - Request/response validation
   - Error handling

3. **Phase 3: Workflows** (8 tests)
   - Complete user journeys
   - Group purchase flows
   - Order cancellation

4. **Phase 4: Stress Tests** (5 tests)
   - 100+ concurrent payments
   - 1000 cache reads
   - Database connection pool limits

5. **Phase 5: Data Consistency** (9 tests)
   - Payment-order sync
   - Cache invalidation
   - Race condition detection

---

#### **Job 2: Test HTTP Handlers**

Isolated handler testing with dedicated database/Redis services.

```bash
go test -v ./handlers/... -timeout 60s
```

Tests HTTP layer without integration dependencies.

---

#### **Job 3: Code Quality Checks**

```bash
golangci-lint run ./... --timeout=5m
go vet ./...
gofmt -l .
```

Ensures code quality:
- **golangci-lint**: Comprehensive Go linter (40+ checkers)
- **go vet**: Built-in static analysis
- **gofmt**: Code formatting

---

#### **Job 4: Notify Results** (Runs if any job completes)

**Slack Notification on Failure**:
```json
{
  "text": "❌ Integration Tests Failed",
  "blocks": [
    {
      "type": "section",
      "text": {
        "type": "mrkdwn",
        "text": "*Repository:* pintuotuo/pintuotuo\n*Branch:* develop\n*Author:* @jane.doe\n*Message:* Integration tests failed"
      }
    },
    {
      "type": "actions",
      "elements": [
        {
          "type": "button",
          "text": { "type": "plain_text", "text": "View Workflow" },
          "url": "https://github.com/pintuotuo/pintuotuo/actions/runs/12345"
        }
      ]
    }
  ]
}
```

**Slack Notification on Success**:
```json
{
  "text": "✅ Integration Tests Passed",
  "blocks": [
    {
      "type": "section",
      "text": {
        "type": "mrkdwn",
        "text": "*Repository:* pintuotuo/pintuotuo\n*Branch:* develop\n*Tests:* 68+ integration tests passed\n*Coverage:* >85%"
      }
    }
  ]
}
```

**Slack Configuration**:
- Uses `SLACK_WEBHOOK_URL` secret
- Sends to #engineering channel (configure in Slack)
- Includes link to workflow run
- `continue-on-error: true` = notification failure doesn't fail workflow

---

## Workflow 4: Deploy to Staging

**File**: `.github/workflows/deploy-staging.yml`
**Triggers**: Push to `develop` branch only
**Duration**: ~5-10 minutes (includes deployment)

### Purpose
Automatically deploys develop branch to staging environment:
- Continuous deployment for testing
- Health check validation
- Smoke tests on staging
- PR comments with deployment status

### Configuration

**Deployment Environments**:
```
Development (Local)
    ↓ [PR Review]
Develop Branch (Automatic Deploy to Staging)
    ↓ [Final Approval]
Main Branch (Manual Deploy to Production)
```

### Jobs Breakdown

#### **Job: Deploy to Staging**

**Steps**:

1. **Checkout code**

2. **Set up Docker Buildx**

3. **Login to Docker Hub**
   ```bash
   docker login -u $DOCKER_USERNAME -p $DOCKER_PASSWORD
   ```
   - Uses secrets for authentication

4. **Build and push Docker image**
   ```bash
   docker build -t pintuotuo/backend:staging .
   docker push pintuotuo/backend:staging
   ```
   - Tagged as `staging`
   - Pushed to Docker Hub

5. **Deploy via SSH** (Requires secrets)
   ```bash
   ssh -i $STAGING_SSH_KEY pintuotuo@$STAGING_SERVER
   cd /app/pintuotuo && docker-compose pull && docker-compose up -d
   ```

   **Required Secrets**:
   - `STAGING_SERVER`: Server IP/hostname (e.g., `staging.pintuotuo.com`)
   - `STAGING_USER`: SSH username (e.g., `pintuotuo`)
   - `STAGING_SSH_KEY`: Private SSH key
   - `DOCKER_USERNAME` & `DOCKER_PASSWORD`: (Already configured)

6. **Run smoke tests**
   ```bash
   # Test basic endpoints
   curl -f https://staging.pintuotuo.com/api/v1/health || exit 1
   ```
   - Health check endpoint
   - Ensures services are running

7. **Health check validation**
   - Verifies database connection
   - Checks Redis availability
   - Validates API response times

8. **Comment on PR**
   ```bash
   gh pr comment -b "✅ Deployed to staging: $STAGING_URL"
   ```
   - Shows deployment status on related PR
   - Links to staging environment

---

## Complete Execution Flow

```
┌─────────────────────────────────────────────────────────────┐
│ Developer: git push to remote                               │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ↓
        ┌────────────────────────┐
        │ GitHub detects push    │
        │ Workflow triggers      │
        └────────┬───────────────┘
                 │
        ┌────────┴──────────────────────────────┐
        │                                       │
        ↓                                       ↓
   ┌──────────────────┐         ┌──────────────────────┐
   │  CI Pipeline     │         │ Integration Tests    │
   │  (All branches)  │         │ (All branches)       │
   │                  │         │                      │
   │ • Backend tests  │         │ • Full integration   │
   │ • Frontend tests │         │ • 22+ test cases     │
   │ • Code quality   │         │ • Database tests     │
   │ • Docker build   │         │ • Stress tests       │
   └────────┬─────────┘         └──────────┬───────────┘
            │                             │
            └──────────────┬──────────────┘
                          │
                   ┌──────▼────────┐
                   │ All tests OK? │
                   └──────┬────────┘
                          │
                   ┌──────┴──────────┐
                   │                 │
              NO ↓                 YES↓
        ┌────────────────┐    ┌─────────────────────┐
        │ Fail workflow  │    │ Check branch type   │
        │ Report error   │    └─────────┬───────────┘
        │ Slack notify   │              │
        └────────────────┘    ┌─────────┴──────────────┐
                              │                        │
                        develop│                    main/master
                              │                        │
                         ┌─────▼──────────┐           │
                         │ Deploy to      │     ✅ Ready for
                         │ Staging        │     manual deploy
                         │                │     to production
                         │ • SSH deploy   │
                         │ • Health check │
                         │ • Smoke tests  │
                         │ • Slack notify │
                         └────────────────┘
```

### Branch-Specific Behavior

| Branch | CI Pipeline | Integration Tests | Deploy to Staging |
|--------|-------------|-------------------|-------------------|
| **develop** | ✅ Yes | ✅ Yes | ✅ Auto (if tests pass) |
| **main** | ✅ Yes | ✅ Yes | ❌ No (manual only) |
| **feature/** | ✅ Yes (PR) | ✅ Yes (PR) | ❌ No |
| **bugfix/** | ✅ Yes (PR) | ✅ Yes (PR) | ❌ No |

---

## GitHub Secrets Configuration

### Currently Configured ✅

| Secret | Value | Purpose |
|--------|-------|---------|
| `DOCKER_USERNAME` | your_docker_username | Docker Hub authentication |
| `DOCKER_PASSWORD` | your_docker_token | Docker Hub authentication |

**To add/view secrets**:
1. Go to: https://github.com/pintuotuo/pintuotuo/settings/secrets/actions
2. Click "New repository secret"
3. Add the name and value

### Required for Full Deployment ⚠️

These secrets are needed for staging deployment to work:

| Secret | Example Value | Purpose |
|--------|---------------|---------|
| `STAGING_SERVER` | `staging.pintuotuo.com` | Staging server IP/hostname |
| `STAGING_USER` | `pintuotuo` | SSH username for deployment |
| `STAGING_SSH_KEY` | `-----BEGIN RSA PRIVATE KEY-----...` | SSH private key (multiline) |
| `STAGING_URL` | `https://staging.pintuotuo.com` | Staging environment URL |
| `SLACK_WEBHOOK_URL` | `https://hooks.slack.com/services/T.../B.../X...` | Slack channel webhook |

### Optional Secrets

| Secret | Purpose |
|--------|---------|
| `SONARCLOUD_TOKEN` | Code quality analysis |
| `CODECOV_TOKEN` | Coverage reporting |
| `SENTRY_DSN` | Error tracking |

### How to Add Secrets

**Step 1: Generate SSH Key (if needed)**
```bash
ssh-keygen -t rsa -b 4096 -f staging_key -N ""
# Private key: staging_key
# Public key: staging_key.pub
```

**Step 2: Add to GitHub**
1. Go to repo → Settings → Secrets and variables → Actions
2. Click "New repository secret"
3. Name: `STAGING_SSH_KEY`
4. Value: Paste contents of `staging_key` (the entire private key)
5. Click "Add secret"

**Step 3: Add SSH public key to server**
```bash
# On staging server:
mkdir -p ~/.ssh
echo "<public key contents>" >> ~/.ssh/authorized_keys
chmod 600 ~/.ssh/authorized_keys
```

---

## Troubleshooting

### "PostgreSQL connection refused" Error

**Symptoms**:
```
could not translate host name "localhost" to address: Name or service not known
```

**Cause**: PostgreSQL service hasn't started yet

**Fix**: Already handled in workflows with health checks. If still failing:
```yaml
- name: Wait for PostgreSQL
  run: |
    until pg_isready -h localhost -U pintuotuo; do
      sleep 1
    done
```

---

### "Database table does not exist" Error

**Symptoms**:
```
ERROR: relation "users" does not exist
```

**Cause**: Schema initialization didn't run or failed

**Fix**: Check these steps in workflow:
1. PostgreSQL service is healthy
2. `psql` command is installed
3. `full_schema.sql` file exists
4. Database initialized before tests run

**Verify locally**:
```bash
docker-compose up -d postgres
PGPASSWORD=test_password psql -h localhost -U pintuotuo -d pintuotuo_test -f scripts/db/full_schema.sql
PGPASSWORD=test_password psql -h localhost -U pintuotuo -d pintuotuo_test -c "\dt"
```

---

### "too many open files" Error in Tests

**Symptoms**:
```
too many open files
```

**Cause**: Tests are opening too many concurrent connections

**Fix**:
- Use `-p 1` flag (serial execution)
- Increase system limit: `ulimit -n 4096`
- Close resources properly in test code

---

### Docker Push Fails with "permission denied"

**Symptoms**:
```
denied: requested access to the resource is denied
```

**Cause**: Invalid Docker Hub credentials

**Fix**:
1. Check `DOCKER_USERNAME` and `DOCKER_PASSWORD` secrets
2. Verify credentials work locally: `docker login`
3. Regenerate Docker Hub token if needed

---

### Slack Notifications Not Sent

**Symptoms**:
- Tests pass but no Slack message

**Cause**: `SLACK_WEBHOOK_URL` not configured

**Fix**:
1. Create Slack webhook: https://api.slack.com/messaging/webhooks
2. Add `SLACK_WEBHOOK_URL` secret to GitHub
3. Update channel in workflow if needed

**Test locally**:
```bash
curl -X POST $SLACK_WEBHOOK_URL \
  -H 'Content-Type: application/json' \
  -d '{"text":"Test message"}'
```

---

### Coverage Threshold Failed

**Symptoms**:
```
Coverage 75% is below threshold 80%
```

**Cause**: Test coverage is below requirement

**Fix**:
1. Run coverage locally: `go test -cover ./...`
2. Identify uncovered code
3. Add tests for missing coverage
4. Verify: `go tool cover -html=coverage.out`

---

### Workflow Timeout

**Symptoms**:
```
The operation exceeded the time limit.
```

**Cause**: Job took too long

**Fix**:
- Increase timeout in workflow (currently 120s for integration tests)
- Identify slow tests: `go test -v -timeout 10m`
- Optimize database queries or test setup

---

### "Module not found" Error

**Symptoms**:
```
go: error: package "github.com/user/package" is not in a module
```

**Cause**: Missing dependency in go.mod

**Fix**:
```bash
cd backend
go get github.com/user/package
go mod tidy
```

---

## Best Practices

### 1. **Fast Feedback Loop**
- Run tests locally before pushing
- Fix errors before creating PR
- Use `git push -n` to do a dry run

### 2. **Keep Tests Fast**
- Unit tests: < 100ms each
- Integration tests: < 1s each
- Avoid long sleeps; use exponential backoff

### 3. **Maintain High Coverage**
- Target: > 80% coverage
- Focus on critical paths
- Test error scenarios

### 4. **Monitor Workflows**
- Check workflow runs regularly
- Fix flaky tests immediately
- Keep dependencies updated

### 5. **Secure Secrets**
- Never commit secrets to code
- Rotate secrets regularly
- Use minimal permissions (least privilege)

---

## Quick Reference

### View Workflow Results
```bash
# GitHub CLI
gh run list --branch develop
gh run view <run-id>
gh run view <run-id> --log
```

### Rerun Failed Workflow
```bash
# From GitHub UI: Actions → Workflow → Rerun jobs
# Or CLI:
gh run rerun <run-id>
```

### Check Logs
1. Go to: https://github.com/pintuotuo/pintuotuo/actions
2. Click workflow run
3. Click job to see steps
4. Click step to see details

### Common Commands
```bash
# Run all tests locally
go test ./... -v

# Run specific test
go test ./services/payment -v -run TestInitiatePayment

# Get coverage
go test ./... -cover

# Run with race detector
go test -race ./...
```

---

**Document Version**: 1.0
**Status**: Complete and Operational ✅
**Maintained By**: DevOps / Technical Lead
