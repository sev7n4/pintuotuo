# Error Reference

> This document provides reference for common error types and handling methods.

## Build Errors

### Go Build Errors

| Error Message | Cause | Solution |
|---------------|-------|----------|
| `undefined: xxx` | Variable/function not defined | Check spelling, import packages |
| `cannot use xxx as type yyy` | Type mismatch | Check type conversion |
| `imported and not used` | Import not used | Remove unused imports |
| `declared but not used` | Variable declared but not used | Use or remove variable |

### TypeScript Build Errors

| Error Message | Cause | Solution |
|---------------|-------|----------|
| `Cannot find name 'xxx'` | Variable not defined | Check spelling, imports |
| `Type 'xxx' is not assignable to type 'yyy'` | Type mismatch | Check type definitions |
| `Property 'xxx' does not exist on type 'yyy'` | Property doesn't exist | Check type definition or extend |

## Test Errors

### Go Test Failures

| Error Pattern | Cause | Solution |
|---------------|-------|----------|
| `panic: runtime error` | Nil pointer / array out of bounds | Add nil checks / boundary checks |
| `expected: x, got: y` | Assertion failed | Check business logic |
| `timeout` | Test timeout | Optimize performance or increase timeout |
| `race detected` | Race condition | Use mutex or channel for synchronization |

### Jest Test Failures

| Error Pattern | Cause | Solution |
|---------------|-------|----------|
| `Cannot read property 'xxx' of undefined` | Accessing undefined property | Add null checks |
| `expect(received).toBe(expected)` | Assertion failed | Check expected values |
| `Timeout - Async callback was not invoked` | Async timeout | Check async/await usage |

### E2E Test Failures

| Error Pattern | Cause | Solution |
|---------------|-------|----------|
| `Timeout waiting for selector` | Element not found | Check selector, wait time |
| `Element is not attached` | Element removed from DOM | Re-get element |
| `net::ERR_CONNECTION_REFUSED` | Service not started | Ensure service is running |
| `Multiple elements found` | Selector matches multiple | Use more precise selector |

## Lint Errors

### golangci-lint

| Error Code | Description | Solution |
|------------|-------------|----------|
| `errcheck` | Error return value not checked | Add error handling |
| `govet` | Static analysis issue | Fix per suggestion |
| `ineffassign` | Ineffective assignment | Use or remove variable |
| `staticcheck` | Static check issue | Fix per suggestion |

### ESLint

| Error Code | Description | Solution |
|------------|-------------|----------|
| `no-unused-vars` | Variable not used | Use or remove variable |
| `no-explicit-any` | Using any type | Define specific type |
| `react-hooks/exhaustive-deps` | Hook dependency missing | Add dependencies |

## CI/CD Errors

### GitHub Actions Failures

| Error Scenario | Cause | Solution |
|----------------|-------|----------|
| `Permission denied` | Insufficient permissions | Check workflow permission config |
| `Out of memory` | Out of memory | Optimize memory usage |
| `Timeout` | Step timeout | Increase timeout or optimize |
| `Service unavailable` | Service unavailable | Check service status |

### Docker Build Failures

| Error Scenario | Cause | Solution |
|----------------|-------|----------|
| `COPY failed` | File doesn't exist | Check file path |
| `npm install failed` | Dependency install failed | Check package.json |
| `Build timeout` | Build timeout | Optimize Dockerfile |

## Security Scan Errors

### Trivy Scan

| Error Type | Description | Solution |
|------------|-------------|----------|
| CVE vulnerability | Dependency has vulnerability | Update dependency version |
| Sensitive info leak | Detected secrets | Use environment variables |

### npm audit

| Error Type | Description | Solution |
|------------|-------------|----------|
| `Critical` | Critical vulnerability | Update immediately |
| `High` | High vulnerability | Update soon |
| `Moderate` | Moderate vulnerability | Schedule update |

## Common Issue Troubleshooting

### Issue: Tests pass locally but fail in CI

**Troubleshooting Steps**:
1. Check environment variable differences
2. Check dependency versions
3. Check timezone/time-related tests
4. Check concurrency/race conditions

### Issue: E2E tests fail intermittently

**Troubleshooting Steps**:
1. Increase wait time
2. Use more reliable selectors
3. Check network requests
4. Check if animations complete

### Issue: Integration test database connection fails

**Troubleshooting Steps**:
1. Check database service status
2. Check connection string
3. Check network connectivity
4. Check authentication info

## Error Log Analysis

### Key Information Extraction

Extract from error logs:
1. Error type
2. Error location (file, line number)
3. Error stack trace
4. Related variable values

### Log Search Commands

```bash
# Find errors
grep -i "error\|fail\|panic" logs.txt

# Find specific error
grep "undefined:" logs.txt

# Find context
grep -A 5 -B 5 "error" logs.txt
```

## Error Fix Priority

| Priority | Error Type |
|----------|------------|
| P0 | Build errors, security vulnerabilities |
| P1 | Test failures, CI blocking |
| P2 | Lint warnings, code quality |
| P3 | Documentation issues, optimization suggestions |

---

## Error Type Judgment Logic (Step 11)

> For determining which step to return to

### Judgment Flow

```
Get failed logs
    ↓
Analyze error type
    ├─ Code error → Step 6 (minimal implementation)
    ├─ Requirement misunderstanding → Step 4 (code analysis)
    └─ Environment issue → Retry current Step
```

### Error Type Judgment Table

| Error Type | Keywords | Return Step | Description |
|------------|----------|-------------|-------------|
| **Code error** | `undefined`, `type error`, `syntax error`, `Cannot find`, `expected`, `got` | Step 6 | Implementation issue |
| **Requirement misunderstanding** | `assertion failed`, `wrong result`, `logic error`, E2E business flow error | Step 4 | Requirement understanding deviation |
| **Environment issue** | `timeout`, `ECONNREFUSED`, `ENOTFOUND`, `service unavailable` | Retry | External dependency issue |

### Detailed Judgment Logic

```bash
# Get failed logs
LOGS=$(gh run view {run-id} --log-failed)

# Determine error type
if echo "$LOGS" | grep -qE "undefined|type error|syntax error|Cannot find"; then
  # Code error → Step 6 (minimal implementation)
  echo "CODE_ERROR: Return to Step 6"
  
elif echo "$LOGS" | grep -qE "assertion failed|wrong result|logic error|expected.*got"; then
  # Requirement misunderstanding → Step 4 (code analysis)
  echo "REQUIREMENT_ERROR: Return to Step 4"
  
elif echo "$LOGS" | grep -qE "timeout|ECONNREFUSED|ENOTFOUND|service unavailable"; then
  # Environment issue → Retry current Step
  echo "ENVIRONMENT_ERROR: Retry"
  
else
  # Unknown error → Request human intervention
  echo "UNKNOWN_ERROR: Request human intervention"
fi
```

### Error Handling by Stage

| Stage | Error Type | Return Step |
|-------|------------|-------------|
| CI/CD Pipeline | Code error | Step 6 |
| Integration Tests | Requirement misunderstanding | Step 4 |
| E2E Tests (current_fix_cases) | Code/Requirement error | Step 4/6 |

### Retry Limit

- Max 5 iterations
- Request human intervention if exceeded
