# GitHub Workflows - Visual Diagrams & Flow Charts

**Last Updated**: 2026-03-18
**Purpose**: Visual representation of workflow execution flows

---

## Table of Contents

1. [Overall CI/CD Pipeline Flow](#overall-cicd-pipeline-flow)
2. [Job Dependencies](#job-dependencies)
3. [Branch-based Triggers](#branch-based-triggers)
4. [Service Dependencies](#service-dependencies)
5. [Detailed Workflow Diagrams](#detailed-workflow-diagrams)
6. [Test Execution Flow](#test-execution-flow)
7. [Deployment Pipeline](#deployment-pipeline)

---

## Overall CI/CD Pipeline Flow

```
┌──────────────────────────────────────────────────────────────┐
│                     Developer Pushes Code                    │
│              git commit && git push origin <branch>           │
└──────────────┬───────────────────────────────────────────────┘
               │
               ↓
┌──────────────────────────────────────────────────────────────┐
│           GitHub Detects Push / PR Created                   │
│          Workflows triggered based on branch + file changes  │
└──────────────┬───────────────────────────────────────────────┘
               │
        ┌──────┴──────────────────────────────┐
        │                                     │
        ↓                                     ↓
    ┌────────────────────┐         ┌──────────────────────┐
    │  CI Pipeline       │         │ Integration Tests    │
    │  ci-pipeline.yml   │         │ test.yml             │
    │  (5-7 minutes)     │         │ (3-5 minutes)        │
    │                    │         │                      │
    │ ├─ Backend Tests   │         │ ├─ Run Tests         │
    │ │  ├─ Schema init  │         │ │  ├─ Schema init    │
    │ │  └─ Unit tests   │         │ │  ├─ 22+ tests      │
    │ │                  │         │ │  └─ Coverage check  │
    │ └─ Frontend Tests  │         │                      │
    │    ├─ Build        │         │ └─ Build Docker      │
    │    └─ Jest + E2E   │         │    └─ Image created  │
    │                    │         │                      │
    └─────────┬──────────┘         └──────────┬───────────┘
              │                               │
              │ ❌ FAIL                       │ ❌ FAIL
              ├──────────────┬────────────────┤
              │              │                │
              │        ✅ PASS              │
              │              │                │
              │        ┌─────▼────────┐      │
              │        │ Integration  │      │
              │        │ Tests -      │      │
              │        │ Payment      │      │
              │        │ (6-8 min)    │      │
              │        │              │      │
              │        │ ├─ Service   │      │
              │        │ │  tests     │      │
              │        │ ├─ Handler   │      │
              │        │ │  tests     │      │
              │        │ ├─ Code      │      │
              │        │ │  quality   │      │
              │        │ └─ Slack     │      │
              │        │    notify    │      │
              │        └─────┬────────┘      │
              │              │                │
              ↓        ✅ PASS│               ↓
        ┌──────────┐   │   ┌──────────┐
        │  REPORT  │◄──┘   │  REPORT  │
        │ ❌ ERROR │       │ ❌ ERROR │
        │ ✅ PASS  │       │ ✅ PASS  │
        └──────────┘   ┌───┴──────────┘
                       │
                ┌──────▼──────────────────┐
                │ Check branch type       │
                └──────┬───────────────────┘
                       │
           ┌───────────┼───────────────┐
           │           │               │
       develop       main/master    feature/*
           │           │               │
           ↓           ↓               ↓
    ┌────────────┐  ✅ Ready    ┌──────────┐
    │  Deploy to │     for     │ PR can   │
    │  Staging   │   Manual    │ be       │
    │            │   Deploy    │ merged   │
    │ ├─ SSH     │             └──────────┘
    │ │  deploy  │
    │ ├─ Health  │
    │ │  check   │
    │ ├─ Smoke   │
    │ │  tests   │
    │ └─ Slack   │
    │    notify  │
    └────┬───────┘
         │
         ↓
   ┌──────────────┐
   │ Deployment   │
   │ Complete ✅  │
   └──────────────┘
```

---

## Job Dependencies

### CI Pipeline Dependencies

```
ci-pipeline.yml
│
├─ Backend Tests (5-6 min)
│  ├─ Checkout
│  ├─ Setup Go 1.21
│  ├─ Install deps
│  ├─ Install psql
│  ├─ Init schema
│  └─ Run tests ←─┐
│                 ├─ Both run in parallel
└─ Frontend Tests └─ (no dependencies)
   ├─ Checkout
   ├─ Setup Node 18
   ├─ npm install
   ├─ npm build
   ├─ npm test
   ├─ playwright install
   └─ playwright test
```

**Execution**: Backend and Frontend tests run in parallel
**Total Time**: ~5-7 minutes (max of both, not sum)

---

### Integration Tests Dependencies

```
test.yml
│
├─ Run Integration Tests (3-5 min) ─┐
│  ├─ Checkout                      │
│  ├─ Setup Go                      │
│  ├─ Download deps                 │
│  ├─ Install psql                  │
│  ├─ Init schema                   │
│  ├─ Run tests (22+ cases)         │
│  └─ Check coverage                │
│                                   ├─ Both run after
│                                   │ test completes
└─ Build Docker Image (2 min) ◄─────┤
   ├─ Checkout
   ├─ Setup Go
   ├─ Build binary
   ├─ Setup Docker
   ├─ Login Docker Hub
   ├─ Build image
   └─ Build summary
```

**Dependency**: Build Docker Image depends on ✅ Run Integration Tests (needs: test)
**Failure**: If tests fail, Docker build doesn't run

---

### Payment Service Integration Tests Dependencies

```
integration-tests.yml
│
├─ Run Integration Tests (4-5 min) ──┐
│  ├─ Checkout                       │
│  ├─ Setup Go                       │
│  ├─ Download deps                  │
│  ├─ Install psql                   │
│  ├─ Init schema                    │
│  ├─ Integration tests              │
│  │  ├─ Phase 1: Service layer     │
│  │  ├─ Phase 2: HTTP handlers     │
│  │  ├─ Phase 3: Workflows         │
│  │  ├─ Phase 4: Stress tests      │
│  │  └─ Phase 5: Data consistency  │
│  ├─ Coverage report                │
│  └─ Check coverage threshold       │
│                                    │
├─ Test HTTP Handlers (1-2 min) ─────┤
│  ├─ Checkout                       │
│  ├─ Setup Go                       │
│  └─ Run handlers tests             │ All 3 run in parallel
│                                    │
├─ Code Quality Checks (2-3 min) ────┤
│  ├─ Checkout                       │
│  ├─ Setup Go                       │
│  ├─ golangci-lint                  │
│  ├─ go vet                         │
│  └─ gofmt check                    │
│                                    │
└─ Notify Results (depends on all) ◄─┘
   ├─ Check results
   ├─ Slack notify (if failed)
   └─ Slack notify (if success)
```

**Execution**: All 3 test jobs run in parallel, then notification job runs
**Time**: ~6-8 minutes (parallel + notification)

---

## Branch-based Triggers

```
Developer commits to different branches
│
├─ Push to feature/* (PR created)
│  │
│  └─ Workflows Triggered:
│     ├─ CI Pipeline ✅ (run tests)
│     ├─ Integration Tests ✅ (run tests)
│     ├─ Payment Tests ✅ (run tests)
│     └─ Deploy to Staging ❌ (skip - not develop)
│
├─ Push to develop
│  │
│  └─ Workflows Triggered:
│     ├─ CI Pipeline ✅ (run tests)
│     ├─ Integration Tests ✅ (run tests)
│     ├─ Payment Tests ✅ (run tests)
│     └─ Deploy to Staging ✅ (auto deploy if ✅)
│
├─ Push to main
│  │
│  └─ Workflows Triggered:
│     ├─ CI Pipeline ✅ (run tests)
│     ├─ Integration Tests ✅ (run tests)
│     ├─ Payment Tests ✅ (run tests)
│     └─ Deploy to Staging ❌ (skip - not develop)
│
└─ Push to master
   │
   └─ Workflows Triggered:
      ├─ CI Pipeline ✅ (run tests)
      ├─ Integration Tests ✅ (run tests)
      ├─ Payment Tests ✅ (run tests)
      └─ Deploy to Staging ❌ (skip - not develop)
```

### Trigger Rules

| Branch | CI | Integration | Payment Tests | Deploy Staging |
|--------|----|----|---|---|
| develop | ✅ | ✅ | ✅ | ✅ (if tests pass) |
| main | ✅ | ✅ | ✅ | ❌ |
| master | ✅ | ✅ | ✅ | ❌ |
| feature/* | ✅ (PR) | ✅ (PR) | ✅ (PR) | ❌ |
| bugfix/* | ✅ (PR) | ✅ (PR) | ✅ (PR) | ❌ |

---

## Service Dependencies

```
┌────────────────────────────────────────────────────────────┐
│                    Workflow Runtime                        │
└────────────────────────────────────────────────────────────┘
        │
        ├─────────────────────┬───────────────────────┐
        │                     │                       │
        ↓                     ↓                       ↓
    ┌────────┐           ┌────────┐            ┌─────────┐
    │ GitHub │           │ Docker │            │ System  │
    │ Actions│           │ Engine │            │ Tools   │
    │ Runner │           │        │            │         │
    │        │           │        │            │         │
    │ • CPU │           │ • Imgs │            │ • apt   │
    │ • Mem │           │ • Nets │            │ • curl  │
    │ • Disk│           │ • Vols │            │ • psql  │
    └────────┘           └────────┘            └─────────┘
        │                     │                       │
        │   ┌─────────────────┴───────────────────┬───┘
        │   │                                     │
        └───┼─────────────────────────────────────┼────┐
            │                                     │    │
            ↓                                     ↓    │
        ┌─────────────────┐           ┌────────────────┴──┐
        │ Test Databases  │           │ System Services  │
        │                 │           │                  │
        │ PostgreSQL 15   │──┐        │ • SSH daemon     │
        │ ├─ Schema       │  │        │ • Docker daemon  │
        │ ├─ Tables       │  │        │ • Network        │
        │ └─ Indexes      │  │        └──────────────────┘
        │                 │  │
        │ Redis 7         │  │
        │ ├─ Cache        │  │
        │ └─ Session      │  │
        └────────┬────────┘  │
                 │           │
            ┌────┴───┬──────┐ │
            ↓        ↓      ↓ ↓
        ┌──────┐  ┌──┐  ┌───────────┐
        │ App  │  │Go│  │ Tests     │
        │ Code │  │  │  │ ├─ Unit   │
        │      │  │  │  │ ├─ Integ  │
        │ • Go │  │  │  │ └─ E2E    │
        │ • JS │  │  │  │           │
        │ • TS │  │1.21│ ├─ Stress  │
        └──────┘  └──┘  │ └─ Smoke  │
                        └───────────┘
```

### Service Initialization Order

```
Workflow Start
    │
    ├─ Step 1: Checkout code (from GitHub)
    │
    ├─ Step 2: Start Docker services
    │  ├─ PostgreSQL 15
    │  │  └─ Wait for healthy (pg_isready)
    │  │
    │  └─ Redis 7
    │     └─ Wait for healthy (redis-cli ping)
    │
    ├─ Step 3: Install system tools
    │  ├─ apt-get update
    │  ├─ Install postgresql-client
    │  └─ (psql command now available)
    │
    ├─ Step 4: Initialize database schema
    │  ├─ Run full_schema.sql
    │  │  ├─ Create users table
    │  │  ├─ Create products table
    │  │  ├─ Create orders table
    │  │  ├─ Create payments table
    │  │  ├─ Create groups table
    │  │  ├─ Create tokens table
    │  │  └─ Create indexes
    │  │
    │  └─ Verify schema (tables exist)
    │
    └─ Step 5: Run tests
       ├─ Connect to PostgreSQL
       ├─ Connect to Redis
       ├─ Run test cases
       └─ Collect results
```

---

## Detailed Workflow Diagrams

### CI Pipeline - Backend Tests Detail

```
Backend Tests Job
├─ START
│  └─ Runner initialized on ubuntu-latest
│
├─ CHECKOUT
│  └─ Pull code from GitHub
│
├─ SETUP GO
│  ├─ Install Go 1.21
│  ├─ Setup go cache
│  └─ Cache go modules
│
├─ INSTALL DEPS
│  ├─ go mod download
│  │  └─ Download all dependencies
│  └─ go mod verify
│     └─ Verify go.sum checksums
│
├─ INSTALL POSTGRES CLIENT
│  ├─ sudo apt-get update
│  └─ sudo apt-get install postgresql-client
│     └─ psql command now available
│
├─ INIT DATABASE SCHEMA
│  ├─ Wait for PostgreSQL ready
│  │  └─ pg_isready check (retry loop)
│  │
│  └─ Load schema
│     ├─ psql -f full_schema.sql
│     │  ├─ CREATE TABLE users
│     │  ├─ CREATE TABLE products
│     │  ├─ CREATE TABLE orders
│     │  ├─ CREATE TABLE payments
│     │  ├─ CREATE TABLE groups
│     │  ├─ CREATE TABLE tokens
│     │  └─ CREATE INDEXes
│     │
│     └─ Verify schema loaded
│
├─ RUN TESTS (Critical step)
│  ├─ Environment:
│  │  ├─ DATABASE_URL: postgresql://...
│  │  ├─ REDIS_URL: redis://...
│  │  ├─ JWT_SECRET: dev_key
│  │  ├─ GIN_MODE: release
│  │  └─ TEST_MODE: true
│  │
│  ├─ Command:
│  │  └─ go test -v ./... -count=1 -p 1
│  │
│  ├─ Test execution:
│  │  ├─ Discover all *_test.go files
│  │  ├─ Run tests serially (-p 1)
│  │  ├─ No caching (-count=1)
│  │  ├─ Verbose output (-v)
│  │  └─ Each package independently
│  │
│  └─ Possible results:
│     ├─ ✅ PASS - All tests passed
│     │  └─ Continue to next step
│     │
│     ├─ ❌ FAIL - Test failed
│     │  ├─ Output failed test name
│     │  ├─ Show failure reason
│     │  └─ Stop job (exit 1)
│     │
│     └─ ⚠️ PANIC - Unrecovered panic
│        └─ Stop job immediately
│
└─ END
   ├─ Job status: PASS or FAIL
   └─ Execution time: ~5-6 minutes
```

### Integration Tests - Complete Flow

```
Integration Tests (test.yml)
│
├─ Parallel Job 1: Run Integration Tests
│  │
│  ├─ Setup (PostgreSQL + Redis)
│  │  ├─ postgres:15-alpine ← Health check enabled
│  │  ├─ redis:7-alpine     ← Health check enabled
│  │  └─ Wait for both ready
│  │
│  ├─ Checkout + Setup Go
│  │
│  ├─ Initialize Schema
│  │  └─ psql -f full_schema.sql
│  │
│  ├─ Run Integration Tests (Main)
│  │  └─ go test ./tests/integration -timeout 120s
│  │     │
│  │     ├─ TEST: PaymentDatabaseConsistency
│  │     │  └─ Verify payment data integrity
│  │     │
│  │     ├─ TEST: PaymentOrderSyncConsistency
│  │     │  └─ Order and payment sync
│  │     │
│  │     ├─ TEST: ConcurrentPaymentsForSameOrder
│  │     │  └─ Race condition test
│  │     │
│  │     ├─ TEST: CompleteGroupPurchaseWithPayment
│  │     │  └─ Full workflow test
│  │     │
│  │     ├─ TEST: HighConcurrencyPaymentInitiation
│  │     │  └─ 100+ concurrent payments
│  │     │
│  │     └─ ... 17 more tests
│  │
│  └─ Check Coverage
│     ├─ go test -cover ./services/... ./handlers/...
│     ├─ Generate coverage report
│     └─ Verify threshold (>80%)
│
└─ Parallel Job 2: Build Docker Image
   │
   ├─ Checkout code
   ├─ Setup Go
   ├─ Build Linux binary
   │  └─ CGO_ENABLED=0 go build
   │
   ├─ Setup Docker Buildx
   ├─ Login Docker Hub
   ├─ Build Docker image
   │  ├─ Dockerfile.alpine (minimal size)
   │  ├─ Tag as :latest
   │  └─ Tag with commit SHA
   │
   └─ Build summary
      └─ Report to GitHub
```

---

## Test Execution Flow

### Single Test Execution

```
Test: func TestInitiatePayment(t *testing.T)
│
├─ Setup Phase
│  ├─ Create test database connection
│  ├─ Connect to Redis test instance
│  ├─ Create test user
│  ├─ Create test product
│  └─ Create test order
│
├─ Execution Phase
│  ├─ Call service.InitiatePayment()
│  │  ├─ Validate order exists
│  │  ├─ Check order status
│  │  ├─ Create payment record
│  │  ├─ Save to database
│  │  └─ Cache payment (Redis)
│  │
│  └─ Receive response
│     ├─ Payment ID
│     ├─ Payment URL
│     └─ Status
│
├─ Assertion Phase
│  ├─ Assert response != nil
│  ├─ Assert payment.Status == "pending"
│  ├─ Assert payment.Amount == order.TotalPrice
│  ├─ Assert payment.Method == "alipay"
│  └─ Assert database has record
│
├─ Cleanup Phase (defer)
│  ├─ Delete test payment
│  ├─ Delete test order
│  ├─ Delete test product
│  └─ Delete test user
│
└─ Result
   ├─ ✅ PASS - All assertions passed
   ├─ ❌ FAIL - Assertion failed
   │  └─ Error: expected X, got Y
   │
   └─ Time: 0.05s
```

### Parallel Test Execution

```
Multiple tests run concurrently
│
├─ Pool of workers created (-p N)
│  ├─ Worker 1 ── Test A
│  ├─ Worker 2 ── Test B
│  ├─ Worker 3 ── Test C
│  └─ Worker 4 ── Test D
│
├─ Each worker:
│  ├─ Gets one test package
│  ├─ Runs all tests in package
│  ├─ Reports results
│  └─ Gets next package
│
├─ With -p 1:
│  └─ Only 1 worker
│     ├─ Tests run serially
│     ├─ No parallel overhead
│     └─ Slower but safer
│
└─ With -p 4:
   └─ 4 workers
      ├─ Tests run in parallel
      ├─ 4x faster (potentially)
      └─ Race conditions possible
```

---

## Deployment Pipeline

### Staging Deployment Flow

```
Push to develop
│
└─ Workflows triggered
   │
   ├─ CI Pipeline ✅
   │
   ├─ Integration Tests ✅
   │
   ├─ Payment Service Tests ✅
   │
   └─ Deploy to Staging (if all ✅)
      │
      ├─ Build Docker image
      │  ├─ Compile Go binary
      │  ├─ Build Docker image
      │  └─ Push to Docker Hub
      │     └─ Tag: pintuotuo/backend:staging
      │
      ├─ Deploy via SSH
      │  ├─ SSH into staging server
      │  │  └─ Using STAGING_SSH_KEY secret
      │  │
      │  ├─ Pull latest Docker image
      │  │  └─ docker-compose pull
      │  │
      │  ├─ Start services
      │  │  └─ docker-compose up -d
      │  │     ├─ backend service
      │  │     ├─ database service
      │  │     └─ cache service
      │  │
      │  └─ Verify deployment
      │     ├─ Check container status
      │     ├─ Check service health
      │     └─ Verify API responds
      │
      ├─ Run smoke tests
      │  ├─ curl /api/v1/health
      │  ├─ curl /api/v1/products
      │  └─ curl /api/v1/users
      │
      ├─ Health check
      │  ├─ Database connection
      │  ├─ Redis connection
      │  ├─ API response time
      │  └─ All services ready
      │
      ├─ Comment on PR
      │  └─ "✅ Deployed to staging"
      │
      └─ Slack notification
         ├─ Channel: #engineering
         ├─ Message: Deployment complete
         └─ Link: staging.pintuotuo.com
```

### Production Deployment (Manual)

```
Main branch ready
│
└─ Create GitHub Release
   │
   ├─ All workflows already passed ✅
   │
   ├─ Manual approval step
   │  └─ Tech lead clicks "Deploy to Production"
   │
   ├─ Production deployment job
   │  ├─ SSH into production server
   │  ├─ Pull latest image
   │  ├─ Stop current deployment
   │  ├─ Start new deployment
   │  ├─ Health checks
   │  └─ Smoke tests
   │
   └─ Slack notification
      ├─ #releases channel
      └─ Announce production deployment
```

---

## Decision Trees

### "Should my workflow run?"

```
Workflow Triggered
│
├─ Is code on main, master, or develop?
│  ├─ YES → All CI workflows run ✅
│  ├─ NO ──→ Only PR checks run
│  │
│  └─ Is code from a PR?
│     ├─ YES → Run CI + Integration tests
│     └─ NO → Run all workflows
│
├─ After tests pass...
│  │
│  ├─ Is branch develop?
│  │  ├─ YES → Deploy to staging
│  │  └─ NO → Skip staging deploy
│  │
│  ├─ Is branch main?
│  │  ├─ YES → Ready for manual production deploy
│  │  └─ NO → (skip)
│  │
│  └─ Is branch master?
│     ├─ YES → Ready for manual production deploy
│     └─ NO → (skip)
```

### "Why did my workflow fail?"

```
Workflow Failed
│
├─ Which job failed?
│  │
│  ├─ Backend Tests?
│  │  ├─ Database initialization
│  │  ├─ Connection string
│  │  ├─ Test assertion
│  │  └─ Dependency import
│  │
│  ├─ Frontend Tests?
│  │  ├─ npm install
│  │  ├─ Build error
│  │  ├─ Jest test failure
│  │  └─ Playwright test
│  │
│  ├─ Docker Build?
│  │  ├─ File missing
│  │  ├─ Build context
│  │  ├─ Authentication
│  │  └─ Docker Hub quota
│  │
│  └─ Deployment?
│     ├─ SSH authentication
│     ├─ Network connectivity
│     ├─ Docker compose
│     └─ Health check
│
└─ Check the logs!
   └─ Click job → scroll to failed step
```

---

**Document Version**: 1.0
**Status**: Complete ✅
**Purpose**: Visual reference for workflow architecture
