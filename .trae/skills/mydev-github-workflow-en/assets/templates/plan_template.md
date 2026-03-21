# Issue Resolution Plan Template

> **Usage**: AI fills this template based on issue analysis results to generate a plan document.

## Basic Information

- **Issue ID**: ISSUE-{three-digit sequence} (e.g., ISSUE-001)
- **Type**: bug | feature | enhancement
- **Priority**: high | medium | low
- **Scope**: backend | frontend | both | infra
- **Created**: {YYYY-MM-DD HH:mm}

## Issue Description

```
Extract issue description from user input, maintain original meaning
```

## Analysis Results

### Root Cause
```
Analyze the root cause of the issue (bug) or requirement background (feature)
```

### Impact Scope
- **Affected Files**: List file paths that need modification
- **Affected Modules**: List involved module names
- **Affected Features**: List affected feature points

### Technical Solution
```
Describe technical solution, including:
- Modification approach
- Key code changes
- Dependency changes
```

## Implementation Plan

### Phase 1: Code Analysis
- [ ] Search related code
- [ ] Analyze code dependencies
- [ ] Locate issue code
- [ ] Assess impact scope

### Phase 2: Test Design (TDD - Red)
- [ ] Design test cases
- [ ] Write failing tests
- [ ] Confirm test failure (bug captured / feature not implemented)

**Test Case Design**:
| Test Name | Scenario | Expected Result |
|-----------|----------|-----------------|
| Test{Function}_{Scenario} | Input/Condition | Expected output |

### Phase 3: Minimal Implementation (TDD - Green)
- [ ] Implement minimal code
- [ ] Confirm tests pass
- [ ] Do not add features not covered by tests

### Phase 4: Refactor (TDD - Refactor)
- [ ] Eliminate duplicate code
- [ ] Improve code structure
- [ ] Confirm tests still pass

### Phase 5: Local Verification
- [ ] Run unit tests
- [ ] Run integration tests
- [ ] Run E2E tests
- [ ] Check code coverage (Backend ≥85%, Frontend ≥80%)
- [ ] Check code style

### Phase 6: Code Commit
- [ ] Commit code
- [ ] Push branch
- [ ] Create PR

### Phase 7: CI Verification
- [ ] CI/CD passed
- [ ] Integration tests passed
- [ ] E2E tests passed

## Expected Results

```
Describe expected results, such as:
- Correct behavior after bug fix
- Usage effect of new feature
```

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| List potential risks | High/Medium/Low | High/Medium/Low | Specific mitigation measures |

## Time Estimation

| Phase | Estimated Time |
|-------|----------------|
| Code Analysis | X minutes |
| Test Design | X minutes |
| Minimal Implementation | X minutes |
| Refactor | X minutes |
| Local Verification | X minutes |
| CI Verification | X minutes |
| **Total** | **X minutes** |

## Notes

```
Other information to note
```

---

## Example Output

```markdown
# Issue Resolution Plan

## Basic Information

- **Issue ID**: ISSUE-001
- **Type**: bug
- **Priority**: high
- **Scope**: backend
- **Created**: 2026-03-21 10:00

## Issue Description

When users log in with correct username and password, the system returns 401 unauthorized error.

## Analysis Results

### Root Cause
The jwtSecret variable is called before the init() function executes, resulting in an empty value.

### Impact Scope
- **Affected Files**: backend/handlers/auth.go
- **Affected Modules**: auth
- **Affected Features**: User login

### Technical Solution
Move jwtSecret initialization to the init() function to ensure it completes before ParseToken is called.

## Implementation Plan

### Phase 1: Code Analysis
- [x] Search related code
- [x] Analyze code dependencies
- [x] Locate issue code
- [x] Assess impact scope

### Phase 2: Test Design (TDD - Red)
- [x] Design test cases
- [x] Write failing tests
- [x] Confirm test failure

**Test Case Design**:
| Test Name | Scenario | Expected Result |
|-----------|----------|-----------------|
| TestLogin_ValidCredentials_ReturnsToken | Valid email/password | Returns valid token |
| TestLogin_InvalidPassword_ReturnsError | Invalid password | Returns 401 error |
| TestLogin_EmptyEmail_ReturnsError | Empty email | Returns validation error |

### Phase 3: Minimal Implementation (TDD - Green)
- [x] Implement minimal code
- [x] Confirm tests pass

### Phase 4: Refactor (TDD - Refactor)
- [x] Code optimized
- [x] Tests still pass

### Phase 5: Local Verification
- [x] Run unit tests
- [x] Coverage 92%

...

## Expected Results

When users log in with correct credentials, return 200 success response with valid token.

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Other modules depend on jwtSecret initialization order | Low | High | Check all places using jwtSecret |

## Time Estimation

| Phase | Estimated Time |
|-------|----------------|
| Code Analysis | 5 minutes |
| Test Design | 10 minutes |
| Minimal Implementation | 10 minutes |
| Refactor | 5 minutes |
| Local Verification | 5 minutes |
| CI Verification | 10 minutes |
| **Total** | **45 minutes** |
```
