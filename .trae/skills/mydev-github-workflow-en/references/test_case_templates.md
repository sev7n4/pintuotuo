# Test Case Templates

> This document defines test case ID conventions, three types of test templates, and archive document format.

## Test Case ID Convention

| Type | Format | Example |
|------|--------|---------|
| Unit Test | `UT-{MODULE}-{SEQ}` | UT-PROD-001 |
| Integration Test | `IT-{MODULE}-{SEQ}` | IT-PROD-001 |
| E2E Test | `E2E-{MODULE}-{SEQ}` | E2E-PROD-001 |

**MODULE Abbreviations**:
- PROD: Product
- USER: User
- ORD: Order
- AUTH: Auth
- PAY: Payment

---

## Unit Test Template

### Test Function Naming

```
Test{Function}_{Scenario}_{ExpectedResult}
```

### Code Template

```go
func Test{Function}_{Scenario}_{ExpectedResult}(t *testing.T) {
    // Arrange
    input := &Input{Field: "value"}
    
    // Act
    result, err := Function(input)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

---

## Integration Test Template

### Test Function Naming

```
Test{Feature}Integration_{Scenario}
```

### Code Template

```go
func Test{Feature}Integration_{Scenario}(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    defer db.Close()
    
    server := setupTestServer(t, db)
    defer server.Close()
    
    // Arrange
    token := getTestToken(t, server)
    payload := `{"name": "test", "price": 100}`
    
    // Act
    req := httptest.NewRequest("POST", "/api/products", strings.NewReader(payload))
    req.Header.Set("Authorization", "Bearer "+token)
    rec := httptest.NewRecorder()
    server.ServeHTTP(rec, req)
    
    // Assert
    assert.Equal(t, http.StatusCreated, rec.Code)
}
```

---

## E2E Test Template

### Test Case Naming

```
{Feature} - {User Action} - {Expected Result}
```

### Code Template (Playwright)

```typescript
test.describe('{Feature}', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login')
    await page.fill('[name="email"]', 'test@example.com')
    await page.fill('[name="password"]', 'password')
    await page.click('button[type="submit"]')
    await page.waitForURL('**/dashboard')
  })

  test('{Feature} - {Action} - {Result}', async ({ page }) => {
    await page.goto('/target')
    await page.click('button.new-item')
    await page.fill('[name="name"]', 'Test Item')
    await page.click('button[type="submit"]')
    
    await expect(page.locator('.success-message')).toBeVisible()
  })
})
```

---

## Archive Document Format

**File Path**: `assets/test_cases/{module}/{feature}_cases.md`

```markdown
# {Module} - {Feature} Test Cases Document

> Generated: {timestamp}
> Related Issue: {issue_id}
> Branch: {branch_name}

## Test Case Statistics

| Type | Total | Passed | Failed | Pending |
|------|-------|--------|--------|---------|
| Unit Tests | {count} | {count} | {count} | {count} |
| Integration Tests | {count} | {count} | {count} | {count} |
| E2E Tests | {count} | {count} | {count} | {count} |

---

## Unit Test Cases

### UT-{MODULE}-{SEQ}: {Test Name}

| Field | Value |
|-------|-------|
| Case ID | UT-PROD-001 |
| Module | Product |
| Test Function | ValidatePrice |
| Status | ✅ passed |

**Test Scenario**:
- Given: {precondition}
- When: {action}
- Then: {expected_result}

**Code Location**: `{file_path}`

---

## Integration Test Cases

### IT-{MODULE}-{SEQ}: {Test Name}
...

---

## E2E Test Cases

### E2E-{MODULE}-{SEQ}: {Test Name}

| Field | Value |
|-------|-------|
| Case ID | E2E-PROD-001 |
| Module | Product Management |
| User Story | {user_story} |
| Status | ✅ passed |

**Test Steps**:
1. {step_1}
2. {step_2}

**Checkpoints**:
- [x] {checkpoint_1}
- [x] {checkpoint_2}

**Code Location**: `{file_path}`
```
