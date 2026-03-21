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

### Test Type Selection Matrix

| Impact Scope | Unit Tests | Integration Tests | E2E Tests |
|--------------|------------|-------------------|-----------|
| backend | ✅ Required | ✅ Required | ❌ Not needed |
| frontend | ✅ Required | ❌ Not needed | ✅ Required |
| both | ✅ Required | ✅ Required | ✅ Required |

### TDD Process Decision

```
Issue Type
    ↓
├── Bug Fix
│   └── Red: Write reproduction test → Green: Minimal fix → Refactor: Optimize
│
└── New Feature
    └── Red: Write acceptance test → Green: Minimal implementation → Refactor: Optimize
```

### Bug Fix Test Strategy

```
Bug Fix
    ↓
Does it involve API?
    ├── Yes → Write unit tests + integration tests
    └── No ↓
Does it involve UI?
    ├── Yes → Write E2E tests + unit tests
    └── No → Write unit tests only
```

**Bug Fix TDD Process**:
```
1. Analyze bug symptoms, understand expected behavior
2. Write unit test to reproduce bug (test should fail)
3. If involves API interaction → Write integration test
4. If involves user flow → Write E2E test
5. Confirm all tests fail = bug is correctly captured
6. Write minimal code to fix bug
7. Confirm all tests pass
8. Add boundary test cases
9. Refactor and optimize (optional)
```

### New Feature Test Strategy

```
New Feature
    ↓
Impact Scope
├── backend → Unit tests + Integration tests
├── frontend → Unit tests + E2E tests
└── both → Unit tests + Integration tests + E2E tests
```

**New Feature TDD Process**:
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
5. Write minimal code to pass tests
6. Refactor and optimize code
7. Repeat to add more test cases
```

### Test Case Design Principles

**Unit Tests**:
- Test single function/method behavior
- Use mocks to isolate external dependencies
- Cover normal paths and boundary conditions

**Integration Tests**:
- Test interactions between components
- Use real database/services
- Verify API contracts

**E2E Tests**:
- Test complete user flows
- Simulate real user operations
- Verify business value

### Test Type Selection

| Scenario | Recommended Test Type | Reason |
|----------|----------------------|--------|
| Pure logic functions | Unit tests | Fast, precise |
| API endpoints | Unit tests + Integration tests | Verify interface contracts |
| Database operations | Integration tests | Need real DB environment |
| User flows | E2E tests | Verify complete experience |
| UI components | Unit tests | Isolate component behavior |

### Test Naming Decision

```
Test function naming format:
Test{FunctionName}_{Scenario}_{ExpectedResult}

Scenario descriptions:
- ValidInput / InvalidInput
- EmptyInput / NilInput
- BoundaryCondition
- ConcurrentAccess
- ErrorCondition

Examples:
- TestLogin_ValidCredentials_ReturnsToken
- TestLogin_InvalidPassword_ReturnsError
- TestLogin_EmptyEmail_ReturnsValidationError
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
