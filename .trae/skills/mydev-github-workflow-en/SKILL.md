---
name: mydev-github-workflow-en
description: Automated development workflow for bug fixes, features, and code changes with GitHub CI/CD integration.
---

---
name: "mydev-github-workflow-en"
description: |
Automated development workflow for bug fixes, features, and code changes with GitHub CI/CD integration. 
Use when user: (1) reports a bug ("login failed", "returns 401", "there's an error"), (2) requests a feature ("add feature", "implement", "create new"), (3) describes improvements ("optimize", "refactor", "improve"), or (4) provides structured issue input (type/title/description fields).
---

# MyDev GitHub Workflow

## Overview

Automated development workflow: Issue parsing → Plan generation → Branch creation → TDD development → Test verification → CI monitoring → PR creation.

## Quick Start

```
User: Login returns 401 error
→ Execute complete workflow until CI passes and PR is created
```

## Workflow Decision Tree

```
Step 0: Init Check → Git status/unfinished tasks
Step 1: Issue Parsing → Type/Priority/Scope/Issue ID
Step 2: Plan Generation → Read templates → Generate plan docs
Step 3: Branch Creation → git checkout -b {type}/issue-{id}
Step 4: Code Analysis → SearchCodebase → Grep → Read
Step 5: Test Design → TDD Red: Write failing tests
Step 6: Minimal Impl → TDD Green: Write minimal code
Step 7: Refactor → TDD Refactor: Optimize code
Step 8: Local Verify → Run tests + coverage check
Step 9: Code Commit → git commit -m "..."
Step 10: CI Monitor → Monitor complete workflow chain
Step 11: Error Fix → Analyze logs → Fix → Retry (max 5 times)
Step 12: Doc Update → Update tracking docs
Step 13: Create PR → gh pr create
Step 14: Cleanup → Output PR link
```

## Core Steps

### Step 0: Init Check

```bash
# Check if current branch is clean
git status --porcelain  # Has output? → Ask user: stash/commit/ignore
```

### Step 1: Issue Parsing

| Type | Keywords | Branch |
|------|----------|--------|
| Bug | "error", "failed", "exception" | bugfix |
| Feature | "add", "implement", "create" | feature |
| Improvement | "optimize", "refactor", "improve" | enhancement |

**Scope**: API/DB → backend | UI → frontend | Both → both

### Step 2-3: Plan & Branch

```bash
# 1. Switch to main and pull latest
git checkout main && git pull origin main

# 2. Verify main branch is clean
git status --porcelain  # Has output? → Warn user, do not continue

# 3. Create new branch
git checkout -b bugfix/issue-001-{description}
```

### Step 4: Code Analysis

Not found? → Ask user

### Step 5: Test Design (TDD - Red)

**Test Strategy**: backend → unit+integration | frontend → unit+E2E | both → all

**Detailed Guide**: `references/test_guide.md`

### Step 6-7: Implementation & Refactor (TDD - Green/Refactor)

```
Red: Write failing test → Green: Minimal implementation → Refactor: Optimize
```

### Step 8: Local Verification

```bash
cd backend && go test -v -race ./...
cd frontend && npm test -- --coverage
```

**Coverage**: Backend ≥85%, Frontend ≥80%

**Failed?** → Analyze logs → Fix → Re-verify

### Step 9: Code Commit

```bash
git commit -m "fix(auth): resolve login 401 error

- Fix jwtSecret initialization
- Add unit tests

Closes #ISSUE-001"
```

### Step 10: CI Monitoring

**Workflow Chain**: CI/CD → Integration → E2E (sequential trigger)

```bash
# Check all workflow statuses
gh pr view {pr-number} --json statusCheckRollup
```

**Monitoring Process**:
```
1. Wait for CI/CD Pipeline to complete
   ├─ Success → Wait for Integration Tests
   └─ Failure → Step 11 (analyze CI error)

2. Wait for Integration Tests to complete
   ├─ Success → Wait for E2E Tests
   └─ Failure → Step 11 (analyze integration error)

3. Wait for E2E Tests to complete
   ├─ Success → Step 12
   └─ Failure → Step 11 (analyze E2E error)
```

**Error Type Decision**:

| Workflow | Error Type | Error Patterns | Next Step |
|----------|------------|----------------|-----------|
| CI/CD | Compile Error | `undefined`, `type error` | Fix code |
| CI/CD | Lint Error | `errcheck`, `staticcheck` | Fix style |
| CI/CD | Unit Test | `FAIL`, `assertion` | Fix test |
| Integration | API Error | `connection refused`, `500` | Fix API/DB |
| Integration | Data Error | `duplicate`, `foreign key` | Fix data logic |
| E2E | UI Error | `selector not found`, `timeout` | Fix UI/selector |
| E2E | Flow Error | `assertion failed` | Fix user flow |
| Any | Environment | `permission`, `OOM`, `timeout` | Retry once → Ask user |

**Success?** → Step 12

### Step 11: Error Fix

```bash
# Get failed logs
gh run view {run-id} --log-failed

# Analyze error location
grep -i "error\|fail\|panic" logs.txt
```

**Fix Workflow**:
```
1. Identify error type (compile/test/lint/security/env)
2. Locate error position (file:line)
3. Analyze error cause
4. Apply fix
5. Local verify
6. Re-commit
7. Return to Step 10
```

**Max 5 retries** → Request human intervention after

### Step 12-14: Finalize

```
Update docs → Create PR → Output PR link
```

## Failure Handling

| Scenario | Action |
|----------|--------|
| Git failed | Retry once → Ask user |
| Test failed | Analyze logs → Fix |
| CI failed | Retry 5 times → Human intervention |
| Code not found | Ask user |

**Core Principle**: Cannot auto-resolve → Ask user immediately

## Reference Files

See `references/` directory:
- `decision_guide.md` - Decision guidance
- `test_guide.md` - Test guide
- `error_reference.md` - Error reference
- `quick_reference.md` - Command cheat sheet
- `design.md` - Complete design