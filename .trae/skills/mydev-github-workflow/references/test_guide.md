# Test Guide (TDD)

## TDD Core Principles

```
Red-Green-Refactor Cycle:
1. Red: Write failing test first
2. Green: Write minimal code to pass
3. Refactor: Optimize under test protection
```

## Test Strategy Matrix

| Impact Scope | Unit Tests | Integration Tests | E2E Tests |
|--------------|------------|-------------------|-----------|
| backend | ✅ Required | ✅ Required | ❌ Not needed |
| frontend | ✅ Required | ❌ Not needed | ✅ Required |
| both | ✅ Required | ✅ Required | ✅ Required |

## Test Naming Convention

```
Unit Tests:        Test{Function}_{Scenario}_{ExpectedResult}
Integration Tests: Test{Feature}Integration_{Scenario}
E2E Tests:         {Feature} - {User Action} - {Expected Result}
```

## Coverage Requirements

| Layer | Minimum |
|-------|---------|
| Backend Core | ≥85% |
| Backend API | ≥80% |
| Frontend | ≥80% |

## Test Templates

### Unit Test (Go)

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

### Table-Driven Test (Go)

```go
func Test{Function}_TableDriven(t *testing.T) {
    tests := []struct {
        name    string
        input   Input
        want    Output
        wantErr bool
    }{
        {"valid", Input{Field: "value"}, Output{Result: "expected"}, false},
        {"invalid", Input{Field: ""}, Output{}, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := Function(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.want, result)
            }
        })
    }
}
```

### E2E Test (Playwright)

```typescript
test.describe('{Feature}', () => {
  test.beforeEach(async ({ page }) => {
    // Login if needed
    await page.goto('/login')
    await page.fill('[name="email"]', 'test@example.com')
    await page.fill('[name="password"]', 'password')
    await page.click('button[type="submit"]')
  })

  test('should {action} successfully', async ({ page }) => {
    await page.goto('/target')
    await page.click('button')
    await expect(page.locator('.result')).toBeVisible()
  })
})
```

### File Upload Test

```typescript
test('should upload file', async ({ page }) => {
  await page.goto('/profile')
  const fileInput = page.locator('input[type="file"]')
  await fileInput.setInputFiles('test/fixtures/file.jpg')
  await expect(page.locator('.success')).toBeVisible()
})
```

## Test Commands

```bash
# Backend unit tests
cd backend && go test -v -race -coverprofile=coverage.out ./...

# Backend integration tests
cd backend && go test -v -run Integration ./...

# Frontend unit tests
cd frontend && npm test -- --coverage --watchAll=false

# Frontend E2E tests
cd frontend && npm run test:e2e
```

## Test Checklist

- [ ] Test name describes scenario and expected result
- [ ] Test is independent (no shared state)
- [ ] Edge cases covered
- [ ] Error paths tested
- [ ] Coverage meets requirements
