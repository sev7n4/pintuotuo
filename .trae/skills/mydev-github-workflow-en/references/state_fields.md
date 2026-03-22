# Workflow State Field Reference

> This document explains the meaning and update timing of each field in `scripts/workflow_state.json`.

## Field Descriptions

### Top-Level Fields

| Field | Type | Description | Update Timing |
|-------|------|-------------|---------------|
| `version` | string | State file version | Only on upgrades |
| `lastUpdated` | string | Last update time (ISO 8601) | Every update |
| `currentIssue` | object/null | Currently processing issue | Start/end of task |
| `activeBranch` | string/null | Current working branch | Branch creation/deletion |

### currentIssue Object

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Issue ID, e.g., "ISSUE-001" |
| `type` | string | Issue type: bug/feature/enhancement |
| `title` | string | Issue title |
| `branch` | string | Branch name |

### workflowState Object

| Field | Type | Description | Possible Values |
|-------|------|-------------|-----------------|
| `stage` | string | Current stage | See stage list below |
| `issueId` | string/null | Associated Issue ID | ISSUE-XXX |
| `branchName` | string/null | Current branch name | bugfix/feature/... |
| `planFile` | string/null | Plan file path | assets/plans/... |
| `tasksFile` | string/null | Tasks file path | assets/tasks/... |
| `startedAt` | string/null | Start time | ISO 8601 |
| `lastCheckpoint` | string/null | Last checkpoint time | ISO 8601 |
| `retryCount` | number | Retry count | 0-5 |

### Possible stage Values

| Value | Description | Corresponding Step |
|-------|-------------|-------------------|
| `idle` | Idle state | No task |
| `parsing` | Parsing issue | Step 1 |
| `planning` | Generating plan | Step 2 |
| `branching` | Creating branch | Step 3 |
| `analyzing` | Analyzing code | Step 4 |
| `implementing` | Implementing code | Step 5-6 |
| `verifying` | Local verification | Step 7-8 |
| `committing` | Committing code | Step 8 |
| `ci_monitoring` | Monitoring CI | Step 9 |
| `fixing` | Fixing errors | Step 10 |
| `documenting` | Updating docs | Step 11 |
| `pr_created` | PR created | Step 12 |
| `completed` | Completed | Step 13 |

### statistics Object

| Field | Type | Description |
|-------|------|-------------|
| `totalIssues` | number | Total issues |
| `resolvedIssues` | number | Resolved issues |
| `totalWorkflows` | number | Workflow runs |
| `successfulWorkflows` | number | Successful workflows |
| `failedWorkflows` | number | Failed workflows |

## Update Examples

### Starting New Task

```json
{
  "currentIssue": {
    "id": "ISSUE-001",
    "type": "bug",
    "title": "Login returns 401 error",
    "branch": "bugfix/issue-001-login-401"
  },
  "activeBranch": "bugfix/issue-001-login-401",
  "workflowState": {
    "stage": "parsing",
    "issueId": "ISSUE-001",
    "branchName": null,
    "planFile": null,
    "tasksFile": null,
    "startedAt": "2026-03-21T10:00:00Z",
    "lastCheckpoint": "2026-03-21T10:00:00Z",
    "retryCount": 0
  }
}
```

### After Branch Creation

```json
{
  "workflowState": {
    "stage": "branching",
    "branchName": "bugfix/issue-001-login-401",
    "planFile": "assets/plans/2026-03-21_issue_001_plan.md",
    "tasksFile": "assets/tasks/2026-03-21_issue_001_tasks.md"
  }
}
```

### After CI Failure

```json
{
  "workflowState": {
    "stage": "fixing",
    "retryCount": 1
  }
}
```

### After Task Completion

```json
{
  "currentIssue": null,
  "activeBranch": null,
  "workflowState": {
    "stage": "completed",
    "issueId": null,
    "branchName": null,
    "planFile": null,
    "tasksFile": null,
    "startedAt": null,
    "lastCheckpoint": null,
    "retryCount": 0
  },
  "statistics": {
    "totalIssues": 1,
    "resolvedIssues": 1,
    "totalWorkflows": 1,
    "successfulWorkflows": 1,
    "failedWorkflows": 0
  }
}
```

---

## Simplified State Tracking

> For SKILL.md core flow

### State File Structure

```json
{
  "current_step": 0,
  "current_fix_cases": ["PROD-002", "SET-002"],
  "retry_count": 0,
  "pr_number": null,
  "merged": false,
  "ci_status": {
    "cicd": "pending",
    "integration": "pending",
    "e2e": "pending"
  }
}
```

### Write Timing

| Step | Write Content | Description |
|------|---------------|-------------|
| Step 0 | Initialize state file | Reset all fields |
| Step 1 | `current_fix_cases` | Test cases to fix this time |
| Step 2-3 | `current_step` | Current step |
| Step 10.1 | `ci_status.cicd` | CI/CD status |
| Step 10.2 | `ci_status.integration` | Integration status |
| Step 10.3 | `ci_status.e2e` | E2E status |
| Step 11 | `retry_count` | Retry count |
| Step 12 | `pr_number` | PR number |
| Step 13 | `merged: true` | Merge status |

---

## Test Case State Tracking

> Field descriptions for `scripts/test_cases_state.json`

### State File Structure

```json
{
  "version": "1.0",
  "lastUpdated": "2026-03-22T10:00:00Z",
  "currentIssue": {
    "id": "ISSUE-001",
    "type": "bug",
    "module": "product",
    "branch": "bugfix/issue-001-product-price"
  },
  "testCases": {
    "unit": [],
    "integration": [],
    "e2e": []
  },
  "statistics": {
    "total": 0,
    "passed": 0,
    "failed": 0,
    "pending": 0
  }
}
```

### currentIssue Fields

| Field | Type | Description | Source |
|-------|------|-------------|--------|
| `id` | string/null | Issue ID | From workflow_state.json |
| `type` | string/null | Issue type | bug/feature/enhancement |
| `module` | string/null | Module | Inferred from code analysis |
| `branch` | string/null | Branch name | From workflow_state.json |

### testCase Object

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Case ID, e.g., UT-PROD-001 |
| `name` | string | Test name |
| `feature` | string | Feature name |
| `status` | string | Status: draft/ready/running/passed/failed/archived |
| `designedAt` | string | Design time |
| `testedAt` | string/null | Test time |
| `file` | string | Test file path |
| `archivedTo` | string/null | Archive doc path (E2E only) |

### Write Timing

| Step | Write Content | Description |
|------|---------------|-------------|
| Step 5.1 | `currentIssue` | Sync from workflow_state.json |
| Step 5.3 | `testCases.*.status=draft` | Test case design |
| Step 5.4 | `testCases.*.status=ready` | Failing test written |
| Step 6.2 | `testCases.*.status=running` | Test executing |
| Step 6.3 | `testCases.*.status=passed/failed` | Test result |
| Step 7.4 | `testCases.*.status=archived` | Archive complete |

### Relationship with workflow_state.json

```
workflow_state.json          test_cases_state.json
├─ current_fix_cases    →    Used to determine test scope
├─ pr_number            ←    Statistics summary
└─ ci_status.e2e        ←    E2E case status verification
```
