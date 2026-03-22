# Test Case Lifecycle Management

> This document defines the lifecycle flow and state tracking from design to archiving.

## Lifecycle Stages

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Design    │ → │    Test     │ → │   Archive   │
│   Phase     │    │   Phase     │    │   Phase     │
└─────────────┘    └─────────────┘    └─────────────┘
       ↓                  ↓                  ↓
   draft/ready      running/passed      archived
                    /failed
```

---

## Step 5: Test Case Design (TDD Red)

### 5.1 Analyze Test Scenarios

**Input Sources**:
- Get `current_fix_cases` from `workflow_state.json`
- Analyze code change impact scope

**Scenario Analysis Checklist**:
- [ ] Happy Path
- [ ] Boundary Conditions
- [ ] Error Handling
- [ ] Concurrency - if applicable

### 5.2 Design Test Cases

**Test Strategy Reference**: `references/test_guide.md`

### 5.3 Write State File

**Write to `test_cases_state.json`**:

```json
{
  "currentIssue": {
    "id": "ISSUE-001",
    "type": "bug",
    "module": "product",
    "branch": "bugfix/issue-001-product-price"
  },
  "testCases": {
    "unit": [
      {
        "id": "UT-PROD-001",
        "name": "TestValidatePrice_PositiveValue_NoError",
        "feature": "Product Price Validation",
        "status": "draft",
        "designedAt": "2026-03-22T10:00:00Z",
        "testedAt": null,
        "file": "backend/internal/service/product_test.go"
      }
    ]
  }
}
```

**Note**: Issue ID is obtained from `workflow_state.json`, not generated separately.

### 5.4 Write Failing Tests

1. Write test code according to `test_case_templates.md`
2. Run test to confirm failure
3. Update status to `ready`

---

## Step 6: Test Case Implementation (TDD Green)

### 6.1 Implement Minimal Code

- Write only minimal code to pass tests
- Avoid over-engineering

### 6.2 Run Test Verification

**Test Commands Reference**: `references/test_guide.md`

### 6.3 Update State File

**Test Passed**:
```json
{
  "status": "passed",
  "testedAt": "2026-03-22T10:05:00Z"
}
```

**Test Failed**:
```json
{
  "status": "failed",
  "testedAt": "2026-03-22T10:05:00Z",
  "errorMessage": "expected nil, got error: invalid price"
}
```

---

## Step 7: Test Case Archive (TDD Refactor)

### 7.1 Refactor Code

- [ ] Code readability
- [ ] Remove duplicate code
- [ ] Performance optimization
- [ ] Tests still pass

### 7.2 Generate Archive Document

1. Read `test_cases_state.json`
2. Group by module
3. Generate document per `test_case_templates.md` archive format

**Document Path**: `assets/test_cases/{module}/{feature}_cases.md`

### 7.3 Update State File

```json
{
  "status": "archived",
  "archivedTo": "assets/test_cases/product/price_cases.md"
}
```

---

## State Tracking

### Status Enumeration

| Status | Description | Write Timing |
|--------|-------------|--------------|
| `draft` | Designing | Step 5.3 |
| `ready` | Design complete, pending test | Step 5.4 |
| `running` | Test executing | Step 6.2 |
| `passed` | Test passed | Step 6.3 |
| `failed` | Test failed | Step 6.3 |
| `archived` | Archived | Step 7.3 |

### Write Timing

| Step | Operation | Field |
|------|-----------|-------|
| Step 5.3 | Write | `currentIssue`, `testCases.*.status=draft` |
| Step 5.4 | Update | `testCases.*.status=ready` |
| Step 6.2 | Update | `testCases.*.status=running` |
| Step 6.3 | Update | `testCases.*.status=passed/failed`, `testedAt` |
| Step 7.3 | Update | `testCases.*.status=archived`, `archivedTo` |

---

## Relationship with workflow_state.json

```
workflow_state.json          test_cases_state.json
├─ current_fix_cases    →    Used to determine test scope
├─ pr_number            ←    Statistics summary
└─ ci_status.e2e        ←    E2E case status verification
```

**Key Constraints**:
- Issue ID MUST be obtained from `workflow_state.json`
- Test case status managed independently
- E2E test case ID should correspond to `current_fix_cases`

---

## Error Handling

| Scenario | Handling |
|----------|----------|
| Test case design failure | Return to Step 4 to re-analyze requirements |
| Test implementation failure | Return to Step 6 to fix code |
| Archive failure | Retry archive, don't block workflow |
