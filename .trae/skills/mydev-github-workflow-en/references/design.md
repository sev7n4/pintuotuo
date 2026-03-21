# Complete Design Specification

> This document provides complete design specification and technical architecture for the MyDev GitHub Workflow skill.

## 1. System Overview

### 1.1 Design Goals

Design an automated development workflow system perfectly integrated with GitHub CI/CD, implementing a complete closed-loop mechanism from issue input to code merge.

### 1.2 Core Capabilities

- **Issue Parsing**: Automatically parse user input, extract issue type, priority, scope
- **Plan Generation**: Automatically generate plan documents and task lists
- **Branch Management**: Create branches per conventions
- **Code Analysis**: Locate code that needs modification
- **Code Implementation**: Modify and improve code
- **Test Writing**: Write unit tests, integration tests, E2E tests
- **Local Verification**: Run full test suite
- **CI Monitoring**: Monitor GitHub workflow execution
- **Error Fix Loop**: Automatically fix failures, max 5 retries
- **Documentation Update**: Update tracking documents

## 2. Technical Architecture

### 2.1 Three-Layer Progressive Disclosure

```
Layer 1: Metadata (always loaded)
├── name: "mydev-github-workflow-en"
└── description: ~100 words

Layer 2: SKILL.md (loaded when triggered)
├── Overview
├── Quick Start
├── When to Use
├── Workflow Decision Tree
├── Core Steps (14 steps)
├── Failure Handling Principles
├── Human Intervention Conditions
└── Reference Files

Layer 3: Resources (loaded on demand)
├── references/
│   ├── decision_guide.md
│   ├── error_reference.md
│   ├── design.md
│   ├── quick_reference.md
│   ├── state_fields.md
│   ├── issue_tracking.md
│   └── workflow_history.md
├── assets/
│   ├── templates/
│   └── plans/ (runtime)
└── scripts/
    └── workflow_state.json
```

### 2.2 Tool Stack

| Component | Technology | Description |
|-----------|------------|-------------|
| AI Engine | Trae IDE + GLM-5 | Intelligent analysis, code generation, issue diagnosis |
| Code Operations | Trae Built-in Tools | Read/Write/SearchReplace/DeleteFile |
| Code Search | SearchCodebase + Grep | Semantic search and regex search |
| Command Execution | RunCommand | Git operations, test runs, builds |
| GitHub Interaction | GitHub CLI (gh) | Branch management, PR creation, workflow monitoring |
| GitHub API | REST API | Workflow status query, log retrieval |

## 3. Workflow Steps Detail

### Step 0: Initialization Check

**Purpose**: Ensure clean environment

**Check Items**:
1. Git status check
2. Unfinished task check
3. Network connectivity check

**Output**: Ready to proceed or ask user

### Step 1: Issue Parsing

**Input**: User natural language or structured input

**Processing**:
1. Analyze user input content
2. Determine issue type (bug/feature/enhancement)
3. Evaluate priority (high/medium/low)
4. Determine scope (backend/frontend/both)
5. Generate Issue ID

**Output**: Structured issue object

### Step 2: Plan Generation

**Input**: Structured issue object

**Processing**:
1. Read plan template
2. Read tasks template
3. Fill templates based on analysis
4. Generate plan and tasks files

**Output**: 
- `assets/plans/{date}_issue_{id}_plan.md`
- `assets/tasks/{date}_issue_{id}_tasks.md`

### Step 3: Branch Creation

**Processing**:
1. Checkout main and pull
2. Fetch and check for new commits
3. Rebase if needed
4. Create new branch
5. Push to remote

**Branch Naming**: `{type}/issue-{id}-{description}`

### Step 4: Code Analysis

**Tools**: SearchCodebase, Grep, Read

**Processing**:
1. Semantic search for related code
2. Keyword search for specific patterns
3. Read relevant files
4. Analyze code dependencies

**Output**: Affected files list, modification plan

### Step 5: Code Implementation

**Tools**: SearchReplace, Write

**Principles**:
- Minimize modification scope
- Maintain consistent code style
- Add necessary error handling
- Follow project conventions

### Step 6: Test Writing

**Test Types**:

| Type | Location | Coverage Target |
|------|----------|-----------------|
| Unit Tests | `backend/{module}_test.go` | ≥85% |
| Integration Tests | `backend/{module}_integration_test.go` | Core flows |
| E2E Tests | `frontend/e2e/{feature}.spec.ts` | User scenarios |

### Step 7: Local Verification

**Commands**:
```bash
# Backend
cd backend && go test -v -race -coverprofile=coverage.out ./...

# Frontend
cd frontend && npm test -- --coverage --watchAll=false
```

**Decision**: Pass → continue, Fail → fix and retry

### Step 8: Code Commit

**Commit Format**: `<type>(<scope>): <subject>`

**Processing**:
1. Stage changes
2. Generate commit message
3. Commit with footer
4. Push to remote

### Step 9: CI Monitoring

**Commands**:
```bash
gh run list --branch={branch} --limit=1
gh run watch {run-id}
```

**Workflow Chain**: CI/CD → Integration Tests → E2E Tests

### Step 10: Error Fix Loop

**Processing**:
1. Get failure logs
2. Analyze error type
3. Locate problematic code
4. Apply fix
5. Local verification
6. Re-commit
7. Return to Step 9

**Termination**: Success or max 5 retries

### Step 11: Documentation Update

**Update**:
- `references/issue_tracking.md` - Add issue record
- `references/workflow_history.md` - Add execution record

### Step 12: Create PR

**Command**: `gh pr create`

### Step 13: Cleanup

**Processing**:
1. Update workflow_state.json
2. Update statistics
3. Output PR link

## 4. State Management

### 4.1 State File Structure

```json
{
  "version": "1.0",
  "lastUpdated": "ISO 8601 timestamp",
  "currentIssue": { ... },
  "activeBranch": "branch name or null",
  "workflowState": {
    "stage": "current stage",
    "issueId": "ISSUE-XXX",
    "branchName": "branch name",
    "planFile": "path to plan",
    "tasksFile": "path to tasks",
    "startedAt": "ISO 8601",
    "lastCheckpoint": "ISO 8601",
    "retryCount": 0
  },
  "statistics": { ... }
}
```

### 4.2 Stage Values

| Stage | Description |
|-------|-------------|
| idle | No active task |
| parsing | Parsing issue |
| planning | Generating plan |
| branching | Creating branch |
| analyzing | Analyzing code |
| implementing | Implementing code |
| verifying | Local verification |
| committing | Committing code |
| ci_monitoring | Monitoring CI |
| fixing | Fixing errors |
| documenting | Updating docs |
| pr_created | PR created |
| completed | Task completed |

## 5. Error Handling

### 5.1 Error Types

| Type | Handling |
|------|----------|
| Git operation failed | Auto retry once, ask user if failed again |
| Test failed | Analyze logs, fix and re-verify |
| CI failed | Max 5 retries, request human intervention if exceeded |
| Cannot locate code | Ask user for more information |

### 5.2 Human Intervention Conditions

- 5 retries still failing
- Cannot locate code
- Affects >5 files
- Database migration
- Security-sensitive operations
- Git conflicts

## 6. Quality Standards

### 6.1 Code Quality

- Code style check passed
- Static analysis passed
- Security scan passed
- No compilation warnings

### 6.2 Test Quality

- Backend unit test coverage ≥85%
- Frontend unit test coverage ≥80%
- Integration tests cover core flows
- E2E tests cover user scenarios

## 7. File Structure

```
.trae/skills/mydev-github-workflow-en/
├── SKILL.md                    # Core skill definition (Layer 2)
├── references/                 # Reference documents (Layer 3)
│   ├── decision_guide.md       # Decision guidance
│   ├── error_reference.md      # Error reference
│   ├── design.md               # This document
│   ├── quick_reference.md      # Quick reference
│   ├── state_fields.md         # State field documentation
│   ├── issue_tracking.md       # Issue tracking (runtime update)
│   └── workflow_history.md     # Workflow history (runtime update)
├── assets/
│   ├── templates/              # Template files
│   │   ├── plan_template.md
│   │   ├── tasks_template.md
│   │   ├── pr_template.md
│   │   ├── bug_report.md
│   │   └── feature_request.md
│   ├── plans/                  # Generated plans (runtime)
│   └── tasks/                  # Generated tasks (runtime)
└── scripts/
    └── workflow_state.json     # State management
```
