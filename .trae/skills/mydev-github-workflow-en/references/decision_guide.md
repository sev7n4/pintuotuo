# Decision Guide

> This document provides decision references during workflow execution to help AI make correct judgments.

## Issue Type Determination

### Decision Flow

```
User Input
    ↓
Contains "bug", "error", "fail", "exception"?
    ├── Yes → bug
    └── No ↓
Contains "add", "implement", "new", "develop"?
    ├── Yes → feature
    └── No ↓
Contains "optimize", "refactor", "improve", "enhance"?
    ├── Yes → enhancement
    └── No → Ask user for confirmation
```

### Examples

| User Input | Determined Type | Reason |
|------------|-----------------|--------|
| "Login returns 401 error" | bug | Contains "error" |
| "Add product favorite feature" | feature | Contains "add" |
| "Optimize database query performance" | enhancement | Contains "optimize" |
| "User reports slow page loading" | bug | Performance issue is a bug |
| "Support multiple login methods" | feature | New feature |

## Priority Assessment

### Criteria

| Priority | Criteria | Examples |
|----------|----------|----------|
| **high** | Affects core functionality, data security, user cannot use | Login failure, payment error, data loss |
| **medium** | Affects secondary features, degraded user experience | Page display issues, inaccurate search results |
| **low** | Optimization improvements, non-urgent issues | Code refactoring, performance optimization |

### Decision Flow

```
Does the issue affect user's core operations?
    ├── Yes → high
    └── No ↓
Does the issue affect user experience?
    ├── Yes → medium
    └── No → low
```

## Scope Determination

### Criteria

| Scope | Criteria |
|-------|----------|
| **backend** | Involves API, database, business logic, authentication/authorization |
| **frontend** | Involves UI components, page interactions, styles, frontend state |
| **both** | Involves both frontend and backend |
| **infra** | Involves deployment, configuration, CI/CD |

### Decision Flow

```
Does the issue involve API or database?
    ├── Yes → Does it also involve UI?
    │         ├── Yes → both
    │         └── No → backend
    └── No ↓
Does the issue involve UI or interaction?
    ├── Yes → frontend
    └── No → infra
```

## Test Strategy Decision

### Bug Fix Test Strategy

```
Bug Fix
    ↓
Does it involve API?
    ├── Yes → Write unit tests + integration tests
    └── No ↓
Does it involve UI?
    ├── Yes → Write E2E tests
    └── No → Write unit tests only
```

### New Feature Test Strategy

```
New Feature
    ↓
Impact Scope
├── backend → Unit tests + Integration tests
├── frontend → Unit tests + E2E tests
└── both → All test types
```

## CI Failure Handling Decision

### Failure Type Determination

```
CI Failure
    ↓
View failure logs
    ↓
Failure Stage
├── Build stage → Syntax errors / dependency issues
├── Unit tests → Code logic errors
├── Integration tests → API / database issues
├── E2E tests → UI / flow issues
└── Lint / Security → Code style issues
```

### Handling Strategies

| Failure Type | Handling Strategy |
|--------------|-------------------|
| Build error | Check syntax, imports, types |
| Unit test failure | Analyze assertions, fix code or tests |
| Integration test failure | Check database connections, API responses |
| E2E test failure | Check selectors, page flows |
| Lint error | Fix code style per conventions |
| Security scan failure | Check dependency vulnerabilities, update versions |

## Human Intervention Decision

### Must Intervene

| Scenario | Reason |
|----------|--------|
| 5 retries still failing | May have issues AI cannot resolve |
| Affects more than 5 files | Impact scope too large, need confirmation |
| Database migration | Data security risk |
| Security-sensitive operations | Need permission confirmation |
| Cannot locate issue | Need more information from user |

### Optional Intervention

| Scenario | Suggestion |
|----------|------------|
| New feature design | Can ask user for preferences |
| Multiple solutions | Can let user choose |
| Performance optimization | Can discuss trade-offs |

## Branch Naming Decision

### Type Mapping

| Issue Type | Branch Prefix |
|------------|---------------|
| bug | `bugfix/` |
| feature | `feature/` |
| enhancement | `enhancement/` |
| Hot fix | `hotfix/` |

### Description Generation

```
Extract keywords from issue title
    ↓
Convert to English (if needed)
    ↓
Join with hyphens
    ↓
Truncate to 30 characters
```

**Examples**:
- "Login returns 401 error" → `bugfix/issue-001-login-401-error`
- "Add product favorite feature" → `feature/issue-002-product-favorite`

## Commit Message Decision

### Type Mapping

| Issue Type | Commit Type |
|------------|-------------|
| bug | `fix` |
| feature | `feat` |
| enhancement | `refactor` or `perf` |
| Documentation update | `docs` |
| Test related | `test` |

### Scope Determination

| Impact Scope | Scope Example |
|--------------|---------------|
| Authentication related | `auth` |
| User related | `user` |
| Product related | `product` |
| Order related | `order` |
| Payment related | `payment` |
| Multiple modules | Omit or use primary module |

## Documentation Update Decision

### Update Timing

| Stage | Update Content |
|-------|----------------|
| After issue parsing | Create issue record |
| After code implementation | Update affected files list |
| After CI passes | Update workflow status |
| After PR creation | Update PR link |
| After completion | Update statistics |

### Append Location

| Document | Append Location |
|----------|-----------------|
| issue_tracking.md | End of "Issue Details" section |
| workflow_history.md | End of "Workflow Execution Records" section |
