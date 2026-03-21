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

### Phase 2: Code Implementation
- [ ] Modify backend code (if needed)
- [ ] Modify frontend code (if needed)
- [ ] Database migration (if needed)

### Phase 3: Test Writing
- [ ] Write unit tests
- [ ] Write integration tests
- [ ] Write E2E tests

### Phase 4: Local Verification
- [ ] Run unit tests
- [ ] Run integration tests
- [ ] Run E2E tests
- [ ] Check code style

### Phase 5: Code Commit
- [ ] Commit code
- [ ] Push branch
- [ ] Create PR

### Phase 6: CI Verification
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
| Code Implementation | X minutes |
| Test Writing | X minutes |
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

### Phase 2: Code Implementation
- [x] Modify backend code

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
| Code Implementation | 10 minutes |
| Test Writing | 15 minutes |
| Local Verification | 5 minutes |
| CI Verification | 10 minutes |
| **Total** | **45 minutes** |
```
