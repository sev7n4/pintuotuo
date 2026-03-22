---
name: "mydev-github-workflow-en"
description: |
  Automated development workflow with complete CI chain monitoring (CI/CD→Integration→E2E).
  Trigger: (1)report bug ("login failed","returns 401","there's an error") (2)request feature ("add feature","implement","create new") (3)code improvement ("optimize","refactor","improve")
---

# MyDev GitHub Workflow

## ⚠️ HARD CONSTRAINTS

```
✅ Phase 1: PR Verification Stage (PR branch)
   └─ Develop → Test → CI Chain Verify → Pass/Retry

✅ Phase 2: Merge & Deploy Stage
   └─ Merge → Deploy

[1] DO NOT ignore current_fix_cases failure! Must fix before merge!
[2] DO NOT skip state tracking! Each step MUST update state files!
[3] DO NOT skip local verification! Must pass tests + type check before commit!
[4] DO NOT skip CI monitoring! Must complete CI/CD → Integration → E2E full chain!
[5] DO NOT merge PR before CI verification passes! Must meet merge conditions!
[6] If code logic modification is involved, DO NOT skip TDD flow (Step 5-7)!
```

**Retry Limit**: Max 5 iterations, then request human intervention

---

## Workflow Chain

```
Step 0: Init Check
   ↓
Step 1: Issue Parsing ⚠️ [Constraint 2]
   ↓
Step 2-3: Branch Creation
   ↓
Step 4: Code Analysis
   ↓
Step 5-7: TDD Flow ⚠️ [Constraint 6]
   ↓
Step 8: Local Verification ⚠️ [Constraint 3]
   ↓
Step 9: git push ──────────────────────┐
   ↓                                   │ Push Trigger
Step 10: CI Monitor ⚠️ [Constraint 4] ── CI/CD + Integration
   ├─ Failed → Step 13 → Step 4/6      │
   └─ Passed                           │
   ↓                                   │
Step 11: Create PR ────────────────────┤
   ↓                                   │ PR Trigger
Step 12: E2E Monitor ⚠️ [Constraint 1] ── current_fix_cases
   ├─ Failed → Step 13 → Step 4/6      │
   └─ Passed                           │
   ↓                                   │
Step 13: Error Fix ⚠️ [Constraint 2]   │
   ↓                                   │
Step 14: Merge PR ⚠️ [Constraint 5]    │
   ↓                                   │
Step 15: Deploy Monitor (main branch)  │
   ↓                                   │
Step 16: Cleanup Output                │
```

**Trigger Logic**:
- Push Trigger: CI/CD(Full) + Integration(Full)
- PR Trigger: E2E(current_fix_cases)

**Merge Conditions**:
- CI/CD Pipeline: Full pass ✅
- Integration Tests: Full pass ✅
- E2E Tests: current_fix_cases pass ✅ (other cases can fail)

---

## Step 0: Init Check

```bash
git status --porcelain
```

**Check Items**:
- Is working directory clean
- Is on main branch

**State Write**: Initialize `scripts/workflow_state.json`

---

## Step 1: Issue Parsing

**Input**: User described issue/requirement

**Output**: 
- `current_fix_cases` - List of test case IDs to fix
- Issue type determination

**State Write**: `workflow_state.json.current_fix_cases`

---

## Step 2-3: Branch Creation

```bash
git checkout main
git pull origin main
git checkout -b {type}/issue-{id}
```

**Branch Naming**: `{type}/issue-{id}`
- type: bugfix / feature / enhancement
- id: ISSUE-XXX

---

## Step 4: Code Analysis

**Toolchain**: SearchCodebase → Grep → Read

**Analysis Content**:
- Locate problem code
- Understand context
- Determine modification scope

---

## Step 5-7: TDD Flow

**Detailed Guide**: `references/test_lifecycle.md`

| Step | TDD Phase | State Write |
|------|-----------|-------------|
| Step 5 | Red - Design failing test | `test_cases_state.json` |
| Step 6 | Green - Minimal implementation | `test_cases_state.json` |
| Step 7 | Refactor - Refactor and archive | `test_cases_state.json` |

**Template Reference**: `references/test_case_templates.md`

---

## Step 9: Code Push

```bash
git add .
git commit -m "{type}: {description}"
git push -u origin {branch-name}
```

**Trigger**: Push to remote branch triggers CI/CD + Integration

**State Write**: `workflow_state.json.ci_status.cicd = running`

---

## Step 10: CI Monitor (Push Trigger)

**Monitor Order**: CI/CD → Integration

| Stage | Trigger | Verification | Failure Handling |
|-------|---------|--------------|------------------|
| 10.1 CI/CD | Push | Full pass | Step 13 → Step 6 |
| 10.2 Integration | Push | Full pass | Step 13 → Step 4 |

**Detailed Scripts**: `references/monitor_scripts.md`

---

## Step 11: Create PR

```bash
gh pr create --title "{title}" --body "{description}"
```

**Trigger**: PR creation triggers E2E tests

**State Write**: `workflow_state.json.pr_number`

---

## Step 12: E2E Monitor (PR Trigger)

**Trigger**: PR creation/update

| Stage | Verification | Failure Handling |
|-------|--------------|------------------|
| E2E ⚠️ [Constraint 1] | current_fix_cases pass | Step 13 → Step 4/6 |

**Note**: Only monitor `current_fix_cases`, other case failures can be ignored

**Detailed Scripts**: `references/monitor_scripts.md`

---

## Step 13: Error Fix

```bash
gh run view {run-id} --log-failed
```

| Error Type | Return Step |
|------------|-------------|
| Code error (compile/Lint/unit test) | Step 6 |
| Requirement misunderstanding (integration/E2E design issue) | Step 4 |
| Environment issue | Retry current Step |

**Detailed Logic**: `references/error_reference.md`

---

## Step 14: Merge Conditions

- CI/CD Pipeline: Full pass ✅
- Integration Tests: Full pass ✅
- E2E Tests: current_fix_cases pass ✅

```bash
gh pr merge {pr-number} --merge --delete-branch
```

**State Write**: `workflow_state.json.merged = true`

---

## Step 15: Deploy Monitor

**Trigger**: Auto-triggered after PR merge

```bash
gh run list --workflow="deploy.yml" --limit 1
```

**State Write**: `workflow_state.json.ci_status.deploy = running`

---

## Step 16: Output Requirements

```
✅ Workflow Complete | PR: #{pr-number} | Branch: {branch} → main
```
