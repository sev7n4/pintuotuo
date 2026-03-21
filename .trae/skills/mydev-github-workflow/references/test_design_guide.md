# Test Design Guide (TDD)

## TDD Core Principles

### Red-Green-Refactor Cycle

```
┌─────────────────────────────────────────────────────────┐
│                     TDD Cycle                           │
│                                                         │
│    ┌─────────┐      ┌─────────┐      ┌──────────┐     │
│    │  RED    │ ───→ │  GREEN  │ ───→ │ REFACTOR │     │
│    │ 写失败测试│      │ 最小实现  │      │   优化   │     │
│    └─────────┘      └─────────┘      └──────────┘     │
│         ↑                                    │         │
│         └────────────────────────────────────┘         │
└─────────────────────────────────────────────────────────┘
```

**Execution Order**:
1. **Red**: Write a failing test first (clarify requirements)
2. **Green**: Write minimal code to pass the test
3. **Refactor**: Optimize code while keeping tests green

### Three Laws of TDD

1. **Write no production code** unless it makes a failing test pass
2. **Write only enough test** to demonstrate a failure
3. **Write only enough production code** to pass the test

## Test Design Methods

### AAA Pattern

```go
func TestUserLogin(t *testing.T) {
    // Arrange - Setup test data and dependencies
    user := &User{Email: "test@example.com", Password: "hashed"}
    mockDB := NewMockDB()
    mockDB.On("FindByEmail", "test@example.com").Return(user, nil)
    auth := NewAuthService(mockDB)
    
    // Act - Execute the behavior being tested
    token, err := auth.Login("test@example.com", "password")
    
    // Assert - Verify the outcome
    assert.NoError(t, err)
    assert.NotEmpty(t, token)
    mockDB.AssertExpectations(t)
}
```

### Boundary Value Analysis

```go
func TestDiscountCalculation(t *testing.T) {
    tests := []struct {
        name     string
        amount   float64
        expected float64
    }{
        {"Below threshold", 99.99, 0},           // Below 100
        {"At threshold", 100.00, 10.00},         // Exactly 100
        {"Above threshold", 100.01, 10.001},     // Just above
        {"Large amount", 10000.00, 1000.00},     // Large value
        {"Zero", 0, 0},                          // Edge case
        {"Negative", -100, 0},                   // Invalid input
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := CalculateDiscount(tt.amount)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### Table-Driven Tests

```go
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {"Valid email", "user@example.com", false},
        {"Valid with subdomain", "user@mail.example.com", false},
        {"Invalid no @", "userexample.com", true},
        {"Invalid no domain", "user@", true},
        {"Empty string", "", true},
        {"Special chars", "user+test@example.com", false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.email)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

## Test Categories

### Unit Tests

**Purpose**: Test single function/method in isolation

**Naming Convention**:
```
Test{FunctionName}_{Scenario}_{ExpectedResult}

Examples:
- TestLogin_ValidCredentials_ReturnsToken
- TestLogin_InvalidPassword_ReturnsError
- TestLogin_EmptyEmail_ReturnsValidationError
```

**Coverage Requirement**: Backend ≥85%, Frontend ≥80%

### Integration Tests

**Purpose**: Test interaction between components

**Focus Areas**:
- Database operations
- API endpoints
- External service calls

**Naming Convention**:
```
Test{Feature}Integration_{Scenario}

Examples:
- TestUserAuthIntegration_CompleteFlow
- TestOrderIntegration_CreateAndProcess
```

### E2E Tests

**Purpose**: Test complete user workflows

**Focus Areas**:
- Critical user journeys
- Cross-component interactions
- Real browser/device testing

**Naming Convention**:
```
E2E: {Feature} - {User Action} - {Expected Result}

Examples:
- E2E: Login - User enters valid credentials - Redirects to dashboard
- E2E: Checkout - User completes purchase - Order created
```

## Test Priority Matrix

| Priority | Type | When to Write |
|----------|------|---------------|
| P0 | Unit Tests | Always, for all new/changed code |
| P1 | Integration Tests | For API/DB interactions |
| P2 | E2E Tests | For critical user flows only |

## Bug Fix Test Process

```
1. Reproduce Bug
   └─ Write test that fails with current code
   
2. Verify Test Fails
   └─ Confirm test captures the bug
   
3. Fix Code
   └─ Write minimal code to pass test
   
4. Verify Test Passes
   └─ Confirm fix works
   
5. Add Regression Tests
   └─ Add edge case tests
```

## Feature Development Test Process

```
1. Write Acceptance Test (E2E)
   └─ Define expected behavior from user perspective
   
2. Write Integration Tests
   └─ Define API/DB contracts
   
3. Write Unit Tests
   └─ Define function behavior in detail
   
4. Implement Code
   └─ Write minimal code to pass all tests
   
5. Refactor
   └─ Optimize while keeping tests green
```

## Test Doubles

### Mock vs Stub vs Fake

| Type | Purpose | Example |
|------|---------|---------|
| Mock | Verify interactions | Assert method was called with args |
| Stub | Provide canned answers | Return fixed test data |
| Fake | Working implementation | In-memory database |

### Mock Example (Go)

```go
type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) FindByID(id int) (*User, error) {
    args := m.Called(id)
    return args.Get(0).(*User), args.Error(1)
}

func TestGetUser(t *testing.T) {
    mockRepo := new(MockUserRepository)
    mockRepo.On("FindByID", 1).Return(&User{ID: 1, Name: "Test"}, nil)
    
    service := NewUserService(mockRepo)
    user, err := service.GetUser(1)
    
    assert.NoError(t, err)
    assert.Equal(t, "Test", user.Name)
    mockRepo.AssertExpectations(t)
}
```

## Test Anti-Patterns

### Avoid

1. **Testing implementation details** - Test behavior, not implementation
2. **Over-mocking** - Use real dependencies when practical
3. **Flaky tests** - Tests should be deterministic
4. **Large test setups** - Keep tests focused and isolated
5. **Testing private methods** - Test through public interface

### Good Practices

1. **One assertion per test** - Or logically related assertions
2. **Descriptive test names** - Should read like documentation
3. **Independent tests** - No shared state between tests
4. **Fast tests** - Unit tests should run in milliseconds
5. **Clear failure messages** - Help diagnose issues quickly

## Coverage Guidelines

### Minimum Requirements

| Layer | Coverage | Rationale |
|-------|----------|-----------|
| Backend Core | ≥85% | Business logic criticality |
| Backend API | ≥80% | API contract verification |
| Frontend Components | ≥80% | UI behavior verification |
| Frontend Utils | ≥90% | Pure functions, easy to test |

### Coverage Commands

```bash
# Go backend
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Frontend (Jest)
npm test -- --coverage --watchAll=false

# View coverage report
open coverage/lcov-report/index.html
```

## Test File Organization

```
backend/
├── handlers/
│   ├── auth.go
│   └── auth_test.go          # Unit tests
├── services/
│   ├── user.go
│   └── user_test.go
├── integration/
│   └── auth_integration_test.go  # Integration tests
└── e2e/
    └── user_flow_test.go         # E2E tests

frontend/
├── src/
│   ├── components/
│   │   ├── Button.tsx
│   │   └── Button.test.tsx
│   └── utils/
│       ├── format.ts
│       └── format.test.ts
└── e2e/
    └── login.spec.ts
```

## Quick Reference

### Test Checklist

- [ ] Test name describes scenario and expected result
- [ ] Test is independent (no shared state)
- [ ] Test follows AAA pattern
- [ ] Edge cases are covered
- [ ] Error paths are tested
- [ ] Mock expectations are verified
- [ ] Test runs fast (<100ms for unit tests)

### Common Assertions

```go
// Equality
assert.Equal(t, expected, actual)
assert.NotEqual(t, unexpected, actual)

// Boolean
assert.True(t, condition)
assert.False(t, condition)

// Nil/Empty
assert.Nil(t, err)
assert.NotNil(t, result)
assert.Empty(t, slice)
assert.NotEmpty(t, result)

// Collections
assert.Contains(t, slice, element)
assert.Len(t, slice, 3)

// Errors
assert.Error(t, err)
assert.NoError(t, err)
assert.EqualError(t, err, "expected error message")

// Type
assert.IsType(t, &User{}, result)
```
