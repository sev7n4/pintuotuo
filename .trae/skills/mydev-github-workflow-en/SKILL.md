---
name: "mydev-github-workflow-en"
description: |
Automated development workflow for bug fixes, features, and code changes with GitHub CI/CD integration.
Use when user: (1) reports a bug ("login fails", "returns 401", "there's an error"), (2) requests a feature ("add feature", "implement a", "new"), (3) describes improvements ("optimize", "refactor", "improve"), or (4) provides structured issue input (type/title/description fields).
---

# MyDev GitHub Workflow

## Overview

Automated development workflow that implements a complete closed loop from issue input to code merge. Integrated with GitHub CI/CD, automatically executes: Issue Parsing → Plan Generation → Branch Creation → Code Implementation → Test Writing → Local Verification → Commit & Push → CI Monitoring → Documentation Update.

## Quick Start

```
User: Login function returns 401 error

→ Automatically executes the complete workflow until CI passes and PR is created
```

```
User:
Type: feature
Title: Add product favorite feature
Description: Users can favorite products they like
Priority: medium
Scope: both

→ Automatically executes the development workflow
```

## When to Use

| Trigger Scenario | User Phrases |
|------------------|--------------|
| Bug Fix | "there's a bug", "login fails", "returns 401 error", "something's wrong" |
| New Feature | "add feature", "implement a", "new", "develop" |
| Code Improvement | "optimize", "refactor", "improve", "enhance" |
| Structured Input | Provides type/title/description fields |

## Workflow Decision Tree

```
Step 0: Initialization Check
    ↓
    ├─ Git has uncommitted changes? → Ask user: stash/commit/ignore
    └─ Has unfinished task? → Ask user: continue/abandon
    ↓
Step 1: Issue Parsing ──→ Determine Type
    ↓                      ├─ bug → bugfix/issue-{id}
    ↓                      ├─ feature → feature/issue-{id}
    ↓                      └─ enhancement → enhancement/issue-{id}
    ↓
Step 2: Plan Generation ──→ Read templates → Generate plan and task docs
    ↓
Step 3: Branch Creation ──→ git checkout -b {branch}
    ↓                        └─ Push failed? → git pull --rebase → retry
    ↓
Step 4: Code Analysis ──→ SearchCodebase → Grep → Read
    ↓                       └─ Not found? → Ask user
    ↓
Step 5: Test Design (TDD - Red) ──→ Write failing tests first
    ↓                                  ├─ Bug: Write reproduction test
    ↓                                  └─ Feature: Write acceptance test
    ↓
Step 6: Minimal Implementation (TDD - Green) ──→ Write minimal code to pass
    ↓
Step 7: Refactor (TDD - Refactor) ──→ Optimize under test protection
    ↓
Step 8: Local Verification ──→ Run tests + coverage check
    ↓                            └─ Failed? → Fix → Re-verify
    ↓
Step 9: Code Commit ──→ git commit -m "..."
    ↓                     └─ Push failed? → git pull --rebase → retry
    ↓
Step 10: CI Monitoring ──→ gh run watch
    ↓                        ├─ Success → Step 12
    ↓                        └─ Failed → Step 11 (max 5 retries)
    ↓
Step 11: Error Fix ──→ Analyze logs → Fix → Re-commit
    ↓
Step 12: Documentation Update ──→ Update tracking docs
    ↓
Step 13: Create PR ──→ gh pr create
    ↓
Step 14: Cleanup ──→ Update state → Output PR link
    ↓
Complete
```

## Core Steps

### Step 0: Initialization Check

**Purpose**: Ensure clean environment to avoid mid-process failures

**Check Items**:
```bash
# Check Git status
git status --porcelain
# Has output? → Ask user for handling

# Check unfinished tasks
# Read workflow_state.json
# stage != "idle" && stage != "completed"? → Ask user
```

**Decision Point**: Has issues? → Ask user, do not auto-handle

### Step 1: Issue Parsing

**Determine Type**:
- Contains "bug", "error", "fail", "exception" → `bug`
- Contains "add", "implement", "new", "develop" → `feature`
- Contains "optimize", "refactor", "improve", "enhance" → `enhancement`

**Evaluate Priority**:
- Affects core functionality/security → `high`
- Affects secondary features → `medium`
- Optimization improvements → `low`

**Determine Scope**:
- Involves API/database → `backend`
- Involves UI/interaction → `frontend`
- Both → `both`

**Generate Issue ID**: Read `scripts/workflow_state.json` to get next sequence number

### Step 2: Plan Generation

```
Read: assets/templates/plan_template.md
      assets/templates/tasks_template.md
Generate: assets/plans/{YYYY-MM-DD}_issue_{id}_plan.md
          assets/tasks/{YYYY-MM-DD}_issue_{id}_tasks.md
```

### Step 3: Branch Creation

```bash
git checkout main && git pull origin main
git fetch origin main
git log HEAD..origin/main --oneline  # Check for new commits
# Has new commits? → git rebase origin/main
git checkout -b bugfix/issue-001-login-401
git push -u origin bugfix/issue-001-login-401
```

**Failure Handling**: Push failed? → `git pull --rebase origin main` → Retry push

### Step 4: Code Analysis

```
SearchCodebase: "login authentication logic"
Grep: "func.*Login|auth.*handler"
Read: backend/handlers/auth.go
```

**Decision Point**: Cannot find code? → Ask user for more information

### Step 5: Test Design (TDD - Red)

**Principle**: Write failing tests first to clarify expected behavior

**Test Strategy Decision Tree**:
```
Feature Impact Scope
    │
    ├─ backend (API only)
    │   └─ Unit Tests + Integration Tests
    │
    ├─ frontend (UI only)
    │   └─ Unit Tests + E2E Tests
    │
    └─ both (Full-stack)
        └─ Unit Tests + Integration Tests + E2E Tests
```

**Bug Fix Process**:
```
1. Analyze bug symptoms
2. Write unit test to reproduce bug (test should fail)
3. If involves API interaction → Write integration test
4. If involves user flow → Write E2E test
5. Confirm all tests fail = bug is captured
```

**New Feature Process**:
```
1. Define acceptance criteria
2. Determine test types based on impact scope:
   - both → E2E Tests + Integration Tests + Unit Tests
   - backend → Integration Tests + Unit Tests
   - frontend → E2E Tests + Unit Tests
3. Write tests in order:
   a. E2E Tests (acceptance tests, define user perspective)
   b. Integration Tests (API contract tests)
   c. Unit Tests (function behavior definition)
4. All tests should fail (feature not implemented)
```

**Test File Locations**:
| Test Type | Backend Location | Frontend Location |
|-----------|------------------|-------------------|
| Unit Tests | `backend/handlers/*_test.go` | `frontend/src/**/*.test.ts` |
| Integration Tests | `backend/integration/*_test.go` | - |
| E2E Tests | - | `frontend/e2e/*.spec.ts` |

**Test Naming Convention**:
```
Unit Tests: Test{FunctionName}_{Scenario}_{ExpectedResult}
Integration Tests: Test{Feature}Integration_{Scenario}
E2E Tests: {Feature} - {User Action} - {Expected Result}

Examples:
- TestUploadAvatar_ValidImage_ReturnsURL
- TestUserAuthIntegration_CompleteFlow
- E2E: Avatar - User uploads image - Avatar displayed
```

**E2E Test Template**:
```typescript
// frontend/e2e/avatar.spec.ts
import { test, expect } from '@playwright/test'

test.describe('Avatar Upload', () => {
  test.beforeEach(async ({ page }) => {
    // Login
    await page.goto('/login')
    await page.fill('[name="email"]', 'test@example.com')
    await page.fill('[name="password"]', 'password')
    await page.click('button[type="submit"]')
    await page.waitForURL('/products')
  })

  test('should upload avatar successfully', async ({ page }) => {
    await page.goto('/profile')
    
    // Upload avatar
    const fileInput = page.locator('input[type="file"]')
    await fileInput.setInputFiles('test/fixtures/test-avatar.jpg')
    
    // Verify upload success
    await expect(page.locator('.ant-message-success')).toBeVisible()
    await expect(page.locator('.avatar img')).toHaveAttribute('src', /uploads\/avatars/)
  })

  test('should reject invalid file type', async ({ page }) => {
    await page.goto('/profile')
    
    const fileInput = page.locator('input[type="file"]')
    await fileInput.setInputFiles('test/fixtures/test.txt')
    
    await expect(page.locator('.ant-message-error')).toBeVisible()
  })

  test('should reject file larger than 2MB', async ({ page }) => {
    await page.goto('/profile')
    
    const fileInput = page.locator('input[type="file"]')
    await fileInput.setInputFiles('test/fixtures/large-image.jpg')
    
    await expect(page.locator('.ant-message-error')).toContainText('2MB')
  })
})
```

**Reference**: `references/test_design_guide.md`

### Step 6: Minimal Implementation (TDD - Green)

**Principle**: Write minimal code to pass tests

```
1. Implement minimal functionality to satisfy tests
2. Do not add features not covered by tests
3. Code can be ugly, but must pass tests
```

**Implementation Methods**:
- SearchReplace: Precise modification of existing files
- Write: Create new files
- Maintain consistent style

### Step 7: Refactor (TDD - Refactor)

**Principle**: Optimize code under test protection

```
1. Ensure all tests pass
2. Refactor code structure
3. Run tests to confirm no functionality broken
4. Repeat until satisfied
```

**Refactoring Checklist**:
- [ ] Eliminate duplicate code
- [ ] Extract functions/methods
- [ ] Improve naming
- [ ] Simplify conditional logic

**Reference**: `references/code_quality_guide.md`

### Step 8: Local Verification

**Verification Steps** (execute in order):

```bash
# 1. Backend unit tests
cd backend && go test -v -race -coverprofile=coverage.out ./...

# 2. Backend integration tests (if any)
cd backend && go test -v -run Integration -race ./...

# 3. Frontend unit tests
cd frontend && npm test -- --coverage --watchAll=false

# 4. Frontend E2E tests (if feature involves UI)
cd frontend && npm run test:e2e
```

**Verification Checklist**:
- [ ] Backend unit tests pass
- [ ] Backend integration tests pass (if any)
- [ ] Frontend unit tests pass
- [ ] Frontend E2E tests pass (if any)
- [ ] Coverage meets requirements

**Coverage Requirements**:
| Layer | Minimum Coverage |
|-------|------------------|
| Backend Core | ≥85% |
| Backend API | ≥80% |
| Frontend | ≥80% |

**E2E Test Prerequisites**:
- Backend server running on localhost:8080
- Frontend server running on localhost:5173
- Test database initialized

**Decision Point**: Failed? → Analyze logs → Fix → Re-verify

### Step 9: Code Commit

```bash
git add .
git commit -m "fix(auth): resolve login 401 error

- Fix jwtSecret initialization
- Add unit tests

Closes #ISSUE-001"
git push origin bugfix/issue-001-login-401
```

**Failure Handling**: Push failed? → `git pull --rebase origin {branch}` → Resolve conflicts → Retry push

### Step 10: CI Monitoring

**Workflow Trigger Chain**:
```
PR Created/Push
    ↓
┌─────────────────┐
│ CI/CD Pipeline  │ ← First triggered
└────────┬────────┘
         │ workflow_run: completed (success)
         ↓
┌─────────────────┐
│ Integration     │ ← Auto triggered
│ Tests           │
└────────┬────────┘
         │ workflow_run: completed (success)
         ↓
┌─────────────────┐
│ E2E Tests       │ ← Auto triggered
└─────────────────┘
```

**Monitoring Steps**:

```bash
# 1. Monitor CI/CD Pipeline
gh run list --workflow="CI/CD Pipeline" --branch=bugfix/issue-001-login-401 --limit=1
gh run watch {run-id}

# 2. Wait for Integration Tests to trigger and monitor
sleep 10
gh run list --workflow="Integration Tests" --limit=1
gh run watch {integration-run-id}

# 3. Wait for E2E Tests to trigger and monitor
sleep 10
gh run list --workflow="E2E Tests" --limit=1
gh run watch {e2e-run-id}
```

**Decision Point**:
- All workflows succeed → Step 12
- Any workflow fails → Step 11

**Complete Check Command**:
```bash
# Check all workflow statuses
gh pr view {pr-number} --json statusCheckRollup
```

### Step 11: Error Fix Loop

```
1. gh run view {run-id} --log-failed
2. Analyze error (see references/error_reference.md)
3. Locate problematic code
4. Apply fix
5. Local verification
6. Re-commit
7. Return to Step 10
```

**Termination Condition**: Success OR 5 retries → Request human intervention

### Step 12: Documentation Update

Append updates to:
- `references/issue_tracking.md`
- `references/workflow_history.md`

### Step 13: Create PR

```bash
gh pr create --title "fix(auth): resolve login 401 error" --body "..."
```

### Step 14: Cleanup

```
1. Update workflow_state.json: stage="completed"
2. Update statistics
3. Output PR link to user
```

## Failure Handling Principles

| Failure Type | Handling Method |
|--------------|-----------------|
| Git operation failed | Auto retry once, ask user if failed again |
| Test failed | Analyze logs, fix and re-verify |
| CI failed | Max 5 retries, request human intervention if exceeded |
| Cannot locate code | Ask user for more information |

**Core Principle**: When encountering issues that cannot be auto-resolved, immediately ask user, do not guess

## Human Intervention Conditions

| Condition | Reason |
|-----------|--------|
| 5 retries still failing | May have issues AI cannot resolve |
| Cannot locate code | Need more information from user |
| Affects >5 files | Impact scope too large, need confirmation |
| Database migration | Data security risk |
| Security-sensitive operations | Need permission confirmation |
| Git conflicts | Need user to decide which version to keep |

## Reference Files

### Reference Documents (references/)

| File | Purpose |
|------|---------|
| `decision_guide.md` | Decision guidance: type judgment, priority assessment, test strategy |
| `error_reference.md` | Error reference: common errors and handling methods |
| `design.md` | Complete design specification and technical architecture |
| `quick_reference.md` | Command cheat sheet, branch conventions, commit conventions |
| `state_fields.md` | State field documentation: workflow_state.json field meanings |
| `issue_tracking.md` | Issue tracking document (runtime update) |
| `workflow_history.md` | Workflow history record (runtime update) |
| `test_design_guide.md` | TDD test design guide: Red-Green-Refactor, test patterns |
| `code_quality_guide.md` | Code quality guide: SOLID principles, naming conventions, security practices |

### Template Files (assets/templates/)

| File | Purpose |
|------|---------|
| `plan_template.md` | Plan document template |
| `tasks_template.md` | Task list template |
| `pr_template.md` | PR description template |
| `bug_report.md` | Bug report template |
| `feature_request.md` | Feature request template |

### State Management (scripts/)

| File | Purpose |
|------|---------|
| `workflow_state.json` | Workflow state cache, records current progress and statistics |

## Error Handling

See `references/error_reference.md` for details
