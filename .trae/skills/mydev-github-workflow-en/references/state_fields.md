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
