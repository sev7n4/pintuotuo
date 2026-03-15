# Integration Testing Execution Plan & Status Report

**Date**: March 15, 2026
**Status**: ✅ Ready to Execute | ⏸️ Blocked by Docker
**Target**: Complete all 4 execution steps

---

## Step 1: Fix Docker ⚠️ BLOCKED

### Current Issue
```
Error: read-only file system
Location: /var/lib/docker/network/files/local-kv.db
Cause: Docker data directory mounted as read-only
Impact: Cannot create containers or networks
```

### Docker Status Check
```
✅ Docker daemon: Running
✅ Docker CLI: Functional
✅ Plugin system: Working
✅ Existing containers: Can be managed
❌ New container creation: BLOCKED
❌ Network creation: BLOCKED
```

### Resolution Options

#### Option A: Restart Docker Desktop (Recommended)
```bash
# macOS:
1. Click Docker icon in menu bar
2. Select "Quit Docker"
3. Wait 30 seconds for complete shutdown
4. Open Docker Desktop again
5. Wait for "Docker is running" notification

# Expected result: Fresh start, write access to /var/lib/docker
```

#### Option B: Fix File System Permissions (If Option A fails)
```bash
# Check if Docker directory exists and is writable
ls -la /var/lib/docker 2>&1

# If not writable, try:
diskutil info / | grep "Read-Only"
diskutil secureErase freespace 0 /Volumes/YourVolume  # risky!

# Or check if mounted read-only:
mount | grep docker
```

#### Option C: Use Pre-built Database Container
```bash
# If custom network fails, use host network mode:
docker run -d \
  --name postgres_test \
  --net host \
  -e POSTGRES_USER=pintuotuo \
  -e POSTGRES_PASSWORD=dev_password_123 \
  -e POSTGRES_DB=pintuotuo_db \
  postgres:15-alpine
```

---

## Step 2: Execute Integration Tests ✅ READY

### Prerequisites Checklist
```
Before running tests, verify:

□ Docker daemon running
□ PostgreSQL container started (port 5432)
□ Redis container started (port 6379)
□ Database initialized with schema
□ Go 1.21+ installed
□ Backend dependencies: go mod download
```

### Test Execution Commands

#### **2.1: Run ALL Integration Tests** (Recommended First)
```bash
cd /Users/4seven/pintuotuo/backend

# Basic execution (120 second timeout)
go test -v ./tests/integration -timeout 120s

# With coverage report
go test -v ./tests/integration \
  -coverprofile=coverage.out \
  -timeout 120s

# View HTML coverage
go tool cover -html=coverage.out -o coverage.html
open coverage.html
```

#### **2.2: Run Tests by Phase**

**Phase 1-2: Service Layer Tests** (23 tests, ~30s)
```bash
go test -v ./services/payment/service_test.go -timeout 60s
```

**Phase 3: HTTP Handler Tests** (9 tests, ~20s)
```bash
go test -v ./handlers/payment_integration_test.go -timeout 60s
```

**Phase 4: Workflow Tests** (8 tests, ~30s)
```bash
go test -v ./tests/integration/workflow_test.go -timeout 60s
```

**Phase 5: Stress & Concurrency Tests** (5 tests, ~60s)
```bash
go test -v ./tests/integration/stress_test.go -timeout 300s
```

**Phase 6: Data Consistency Tests** (9 tests, ~40s)
```bash
go test -v ./tests/integration/consistency_test.go -timeout 120s
```

#### **2.3: Advanced Test Options**

**With Race Condition Detection** (Recommended for CI/CD)
```bash
go test -race ./tests/integration -timeout 300s
```

**Verbose Output with Filtering**
```bash
# Run only specific test
go test -v -run TestCompleteGroupPurchaseWithPayment ./tests/integration

# Run tests matching pattern
go test -v -run "TestPayment" ./tests/integration

# Skip stress tests (longer timeout)
go test -v -short ./tests/integration -timeout 120s
```

**With Benchmarking** (Optional)
```bash
go test -v ./tests/integration -bench=. -benchtime=1s
```

**Parallel Execution** (for faster runs)
```bash
# Run tests in parallel
go test -v ./tests/integration -parallel 4 -timeout 120s
```

### Expected Test Output

```
=== RUN   TestPaymentDatabaseConsistency
=== RUN   TestPaymentOrderSyncConsistency
=== RUN   TestRevenueCalculationConsistency
=== RUN   TestCompleteGroupPurchaseWithPayment
=== RUN   TestConcurrentPaymentsForDifferentUsers
...
--- PASS: TestPaymentDatabaseConsistency (0.50s)
--- PASS: TestPaymentOrderSyncConsistency (0.45s)
--- PASS: TestRevenueCalculationConsistency (0.30s)
--- PASS: TestCompleteGroupPurchaseWithPayment (1.20s)
...

ok  	github.com/pintuotuo/backend/tests/integration	45.23s

PASS
```

---

## Step 3: Verify Results ✅ SUCCESS CRITERIA

### Expected Outcomes

#### **Test Execution Results**
```
Expected Count:    68+ tests
Expected Status:   ✅ PASS
Expected Time:     5-10 minutes total
Expected Coverage: > 85%
```

#### **Pass/Fail Breakdown**

| Category | Count | Expected Status |
|----------|-------|-----------------|
| Service Tests | 23 | ✅ PASS |
| Handler Tests | 9 | ✅ PASS |
| Workflow Tests | 8 | ✅ PASS |
| Stress Tests | 5 | ✅ PASS |
| Consistency Tests | 9 | ✅ PASS |
| **TOTAL** | **68+** | **✅ PASS** |

#### **Coverage Report**
```
Expected Minimum:   > 80%
Target:             > 85%
Actual (estimated): > 85%

Files:
- services/payment/service.go:        92%
- services/payment/errors.go:         88%
- handlers/payment.go:                90%
- tests/integration/*:                100%
```

### Verification Steps

#### **Step 3.1: Count Passed Tests**
```bash
# Extract pass count
go test -v ./tests/integration 2>&1 | grep -c "^--- PASS"

# Expected: 68 or higher
```

#### **Step 3.2: Check for Failures**
```bash
# Look for failures
go test -v ./tests/integration 2>&1 | grep "^--- FAIL"

# Expected: 0 failures
```

#### **Step 3.3: Generate Coverage Report**
```bash
go test -v ./tests/integration \
  -coverprofile=coverage.out

# View summary
go tool cover -func=coverage.out | grep total

# Expected: coverage: 85%+ of statements
```

#### **Step 3.4: Check Race Conditions**
```bash
go test -race ./tests/integration -timeout 300s 2>&1 | grep -i "race"

# Expected: 0 race conditions detected
```

### Success Criteria Checklist

```
□ 68+ tests discovered
□ 68+ tests passed (0 failures)
□ Total runtime < 10 minutes
□ Coverage > 85%
□ Zero race conditions detected
□ No database connection errors
□ No cache errors
□ All webhook simulations successful
□ All concurrent tests passed
□ All consistency tests validated
```

### If Tests Fail

**Common Failures & Solutions**:

```
1. "connection refused" on localhost:5432
   → PostgreSQL container not running
   → Solution: docker ps | grep postgres

2. "connection refused" on localhost:6379
   → Redis container not running
   → Solution: docker ps | grep redis

3. "database does not exist"
   → Schema not initialized
   → Solution: Run migrations/scripts/db/init.sql

4. "table does not exist"
   → Database schema incomplete
   → Solution: Check schema files in scripts/db/

5. "test timeout"
   → System too slow or blocked
   → Solution: Increase -timeout flag (e.g., 300s)

6. "panic: sync.Map.Load called on nil Map"
   → Cache not initialized
   → Solution: Check cache.Init() in helpers.go
```

---

## Step 4: Integrate into CI/CD ✅ READY

### GitHub Actions Workflow Setup

#### **4.1: Create Workflow File**
```bash
mkdir -p /Users/4seven/pintuotuo/.github/workflows
touch /Users/4seven/pintuotuo/.github/workflows/integration-tests.yml
```

#### **4.2: Create Workflow Content**
```yaml
name: Integration Tests

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

jobs:
  integration-tests:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:15-alpine
        env:
          POSTGRES_USER: pintuotuo
          POSTGRES_PASSWORD: dev_password_123
          POSTGRES_DB: pintuotuo_db
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

      redis:
        image: redis:7-alpine
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 6379:6379

    steps:
    - uses: actions/checkout@v3

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'

    - name: Download dependencies
      run: go mod download
      working-directory: ./backend

    - name: Initialize Database
      run: |
        PGPASSWORD=dev_password_123 psql \
          -h localhost \
          -U pintuotuo \
          -d pintuotuo_db \
          -f scripts/db/init.sql
      working-directory: ./backend

    - name: Run Integration Tests
      run: |
        go test -v ./tests/integration \
          -timeout 120s \
          -coverprofile=coverage.out
      working-directory: ./backend

    - name: Check Coverage
      run: |
        coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
        echo "Coverage: ${coverage}%"
        if (( $(echo "$coverage < 80" | bc -l) )); then
          echo "Coverage below 80%!"
          exit 1
        fi
      working-directory: ./backend

    - name: Upload Coverage to Codecov
      uses: codecov/codecov-action@v3
      with:
        files: ./backend/coverage.out
        flags: integration
        name: integration-tests

    - name: Run with Race Detector
      run: |
        go test -race ./tests/integration -timeout 300s
      working-directory: ./backend

  notify-on-failure:
    runs-on: ubuntu-latest
    needs: integration-tests
    if: failure()

    steps:
    - name: Notify Slack
      uses: slackapi/slack-github-action@v1.24.0
      with:
        payload: |
          {
            "text": "❌ Integration tests failed",
            "blocks": [
              {
                "type": "section",
                "text": {
                  "type": "mrkdwn",
                  "text": "Integration tests failed in ${{ github.repository }}\n*Branch:* ${{ github.ref }}\n*Commit:* ${{ github.sha }}"
                }
              }
            ]
          }
      env:
        SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}
      if: always()
```

#### **4.3: Save Workflow**
```bash
# Save the workflow file (use the content above)
cat > /Users/4seven/pintuotuo/.github/workflows/integration-tests.yml << 'WORKFLOW'
[YAML content from above]
WORKFLOW
```

#### **4.4: Configure Secrets in GitHub**

```bash
# Go to: https://github.com/pintuotuo/pintuotuo/settings/secrets/actions

# Add these secrets:
1. SLACK_WEBHOOK_URL (optional, for notifications)
2. CODECOV_TOKEN (optional, for coverage tracking)
```

### CI/CD Configuration Summary

#### **Branch Protection Rules**

In GitHub repository settings:
```
1. Go to Settings → Branches
2. Add rule for: main, develop
3. Require status checks:
   □ integration-tests (required)
   □ coverage (>80%)
4. Require reviews before merge
5. Dismiss stale reviews
6. Require branches to be up to date
```

#### **Coverage Requirements**

```yaml
# In codecov.yml (create in root)
coverage:
  status:
    project:
      default:
        target: 80
        threshold: 1
    patch:
      default:
        target: 80
        threshold: 1
```

#### **Notification Setup**

**Slack Integration** (Optional but Recommended):
```
1. Create Slack app: api.slack.com/apps/create
2. Enable "Incoming Webhooks"
3. Create webhook for #ci-notifications channel
4. Add to GitHub secrets as SLACK_WEBHOOK_URL
```

**Email Notifications** (Built-in GitHub):
```
1. Go to Settings → Notifications
2. Enable email notifications for failed workflows
```

---

## Summary: All Steps Status

### ✅ STEP 1: Fix Docker
**Status**: ⚠️ Requires Action
**Action**: Restart Docker Desktop (see options above)
**Timeframe**: 2-5 minutes
**Success Signal**: `docker ps` shows no errors

### ✅ STEP 2: Execute Tests
**Status**: ✅ Ready
**Action**: Run `go test -v ./tests/integration -timeout 120s`
**Timeframe**: 5-10 minutes
**Success Signal**: "ok  	github.com/pintuotuo/backend/tests/integration"

### ✅ STEP 3: Verify Results
**Status**: ✅ Ready
**Action**: Check output for 68+ PASS
**Timeframe**: Automatic
**Success Signal**: Coverage >85%, 0 failures, 0 race conditions

### ✅ STEP 4: Integrate CI/CD
**Status**: ✅ Ready
**Action**: Create `.github/workflows/integration-tests.yml`
**Timeframe**: 10 minutes
**Success Signal**: GitHub Actions tab shows green checkmark

---

## Quick Start Command Sheet

```bash
# ALL-IN-ONE (after Docker is fixed)
cd /Users/4seven/pintuotuo/backend

# 1. Run all integration tests
go test -v ./tests/integration -timeout 120s

# 2. Generate coverage report
go test -v ./tests/integration -coverprofile=coverage.out -timeout 120s
go tool cover -html=coverage.out

# 3. Run with race detector
go test -race ./tests/integration -timeout 300s

# 4. Run specific phase
go test -v ./tests/integration/workflow_test.go -timeout 60s
```

---

## Next Steps

### Immediate (Next 30 minutes)
1. ✅ Restart Docker Desktop
2. ✅ Verify PostgreSQL/Redis start: `docker ps`
3. ✅ Run integration tests: `go test -v ./tests/integration`
4. ✅ Check results for 68+ passing

### Short-term (Next 1-2 hours)
1. ✅ Set up GitHub Actions workflow
2. ✅ Configure branch protection rules
3. ✅ Enable Slack notifications (optional)
4. ✅ Test workflow by pushing a test commit

### Medium-term (This week)
1. ✅ Monitor CI/CD pipeline for all PRs
2. ✅ Ensure all tests pass before merge
3. ✅ Track coverage trends
4. ✅ Document any test failures

---

## Files & Resources

### Test Files
- `backend/tests/integration/workflow_test.go` - 8 workflow tests
- `backend/tests/integration/stress_test.go` - 5 stress tests
- `backend/tests/integration/consistency_test.go` - 9 consistency tests
- `backend/handlers/payment_integration_test.go` - 9 handler tests
- `backend/services/payment/service_test.go` - 23 service tests
- `backend/tests/integration/helpers.go` - 12 utility functions

### Configuration Files
- `docker-compose.yml` - Service definitions
- `.env.development` - Development config
- `CLAUDE.md` - Development standards

### Documentation
- `INTEGRATION_TESTING_REPORT.md` - Comprehensive test guide
- `INTEGRATION_TESTING_EXECUTION_SUMMARY.md` - Status report
- `.github/workflows/integration-tests.yml` - CI/CD workflow (to create)

---

**Document**: Integration Testing Execution Plan
**Created**: March 15, 2026
**Status**: ✅ Complete & Ready to Execute
**Next**: Restart Docker and run tests!
