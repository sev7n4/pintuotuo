# Code Quality Guide

## Quality Principles

### Core Values

1. **Readability First** - Code is read more than written
2. **Simplicity** - No unnecessary complexity
3. **Consistency** - Follow project conventions
4. **Testability** - Design for easy testing

### SOLID Principles

| Principle | Description | Example |
|-----------|-------------|---------|
| **S**ingle Responsibility | One reason to change | One handler per route |
| **O**pen/Closed | Open for extension, closed for modification | Use interfaces |
| **L**iskov Substitution | Subtypes must be substitutable | Proper inheritance |
| **I**nterface Segregation | Many specific interfaces | Small, focused interfaces |
| **D**ependency Inversion | Depend on abstractions | Inject dependencies |

## Code Style

### Go Backend

```go
// Good: Clear naming, error handling, documentation
func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    if err := req.Validate(); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    user := &User{
        Email:     req.Email,
        Name:      req.Name,
        CreatedAt: time.Now(),
    }
    
    if err := s.repo.Create(ctx, user); err != nil {
        return nil, fmt.Errorf("create user: %w", err)
    }
    
    return user, nil
}
```

### TypeScript Frontend

```typescript
// Good: Clear types, proper error handling
interface CreateUserRequest {
  email: string;
  name: string;
}

async function createUser(req: CreateUserRequest): Promise<User> {
  if (!req.email || !req.name) {
    throw new ValidationError('Email and name are required');
  }
  
  const response = await api.post<User>('/users', req);
  return response.data;
}
```

## Naming Conventions

### General Rules

| Element | Convention | Example |
|---------|------------|---------|
| Variables | camelCase / snake_case | `userName`, `user_name` |
| Constants | UPPER_SNAKE_CASE | `MAX_RETRY_COUNT` |
| Functions | verb + noun | `getUserById`, `calculateTotal` |
| Classes/Types | PascalCase | `UserService`, `CreateUserRequest` |
| Interfaces | PascalCase + prefix/suffix | `UserRepository`, `IUserService` |
| Files | snake_case / kebab-case | `user_service.go`, `user-handler.ts` |

### Meaningful Names

```go
// Bad
func process(d []byte) error

// Good
func ProcessPaymentRequest(data []byte) error

// Bad
var x, y int

// Good
var retryCount, maxRetries int
```

## Error Handling

### Go Error Handling

```go
// Good: Wrap errors with context
func (s *Service) DoSomething(id int) error {
    user, err := s.repo.FindByID(id)
    if err != nil {
        return fmt.Errorf("find user %d: %w", id, err)
    }
    
    if err := s.process(user); err != nil {
        return fmt.Errorf("process user %d: %w", id, err)
    }
    
    return nil
}

// Good: Custom error types for specific errors
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error: %s - %s", e.Field, e.Message)
}
```

### TypeScript Error Handling

```typescript
// Good: Typed errors with context
class AppError extends Error {
  constructor(
    message: string,
    public readonly code: string,
    public readonly statusCode: number = 500
  ) {
    super(message);
    this.name = 'AppError';
  }
}

async function fetchUser(id: string): Promise<User> {
  try {
    const response = await api.get<User>(`/users/${id}`);
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.status === 404) {
      throw new AppError(`User ${id} not found`, 'USER_NOT_FOUND', 404);
    }
    throw new AppError(`Failed to fetch user: ${error}`, 'FETCH_ERROR', 500);
  }
}
```

## Code Review Checklist

### Before Committing

- [ ] Code compiles without warnings
- [ ] All tests pass
- [ ] Coverage meets requirements (≥85% backend, ≥80% frontend)
- [ ] No hardcoded secrets or credentials
- [ ] Error messages are clear and actionable
- [ ] Complex logic has comments explaining "why"
- [ ] Functions are not too long (<50 lines)
- [ ] Files are not too large (<500 lines)

### Review Criteria

| Category | Check |
|----------|-------|
| **Correctness** | Does it do what it's supposed to? |
| **Security** | Any vulnerabilities? SQL injection, XSS, etc.? |
| **Performance** | Any obvious inefficiencies? |
| **Maintainability** | Easy to understand and modify? |
| **Testability** | Properly tested? |

## Static Analysis Tools

### Go Tools

```bash
# Linting
golangci-lint run ./...

# Format
go fmt ./...
goimports -w .

# Security
gosec ./...

# Complexity
gocyclo -over 15 .
```

### TypeScript Tools

```bash
# Linting
npm run lint

# Type checking
npm run typecheck

# Format
npm run format

# Security audit
npm audit
```

## Refactoring Guidelines

### When to Refactor

- After tests are green (TDD cycle)
- When adding new features
- When fixing bugs (if safe)
- When code smells are detected

### Common Code Smells

| Smell | Solution |
|-------|----------|
| Long method | Extract smaller methods |
| Large class | Split responsibilities |
| Duplicate code | Extract common logic |
| Long parameter list | Use parameter object |
| Feature envy | Move method to proper class |
| Primitive obsession | Use value objects |

### Safe Refactoring Steps

```
1. Ensure tests exist and pass
2. Make small, incremental changes
3. Run tests after each change
4. Commit frequently
5. Keep refactoring separate from feature changes
```

## Documentation Standards

### Code Comments

```go
// CreateUser creates a new user with the given request.
// It validates the input, hashes the password, and stores the user.
// Returns ErrDuplicateEmail if email already exists.
func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    // Implementation...
}
```

### Package Documentation

```go
// Package user provides user management functionality.
//
// The UserService handles user CRUD operations, authentication,
// and profile management. It depends on UserRepository for data access.
//
// Example usage:
//
//     service := user.NewService(repo)
//     user, err := service.CreateUser(ctx, &CreateUserRequest{...})
package user
```

### API Documentation

```go
// Create user
// @Summary Create a new user
// @Description Create a new user with the provided information
// @Tags users
// @Accept json
// @Produce json
// @Param request body CreateUserRequest true "User creation request"
// @Success 201 {object} User
// @Failure 400 {object} ErrorResponse
// @Router /users [post]
func (h *Handler) CreateUser(c *gin.Context) {
    // Implementation...
}
```

## Security Best Practices

### Input Validation

```go
// Good: Validate all inputs
func (req *CreateUserRequest) Validate() error {
    if req.Email == "" {
        return &ValidationError{Field: "email", Message: "required"}
    }
    if !isValidEmail(req.Email) {
        return &ValidationError{Field: "email", Message: "invalid format"}
    }
    if len(req.Name) > 100 {
        return &ValidationError{Field: "name", Message: "too long"}
    }
    return nil
}
```

### SQL Injection Prevention

```go
// Bad: SQL injection vulnerable
query := fmt.Sprintf("SELECT * FROM users WHERE id = %s", userID)

// Good: Use parameterized queries
err := db.QueryRow("SELECT * FROM users WHERE id = $1", userID).Scan(&user)
```

### XSS Prevention (Frontend)

```typescript
// Bad: Direct HTML insertion
element.innerHTML = userInput;

// Good: Use textContent or sanitization
element.textContent = userInput;
// Or use a sanitization library
element.innerHTML = DOMPurify.sanitize(userInput);
```

## Performance Guidelines

### Database Queries

```go
// Bad: N+1 queries
for _, user := range users {
    orders, _ := getOrders(user.ID)
    user.Orders = orders
}

// Good: Batch query
userIDs := make([]int, len(users))
for i, u := range users {
    userIDs[i] = u.ID
}
orders, _ := getOrdersByUserIDs(userIDs)
```

### Frontend Optimization

```typescript
// Good: Memoization for expensive computations
const expensiveValue = useMemo(() => {
  return computeExpensiveValue(data);
}, [data]);

// Good: Debounce for frequent events
const debouncedSearch = useMemo(
  () => debounce((query: string) => searchAPI(query), 300),
  []
);
```

## Quick Reference

### Quality Metrics

| Metric | Target |
|--------|--------|
| Test Coverage | Backend ≥85%, Frontend ≥80% |
| Cyclomatic Complexity | ≤15 per function |
| Function Length | ≤50 lines |
| File Length | ≤500 lines |
| Linting Errors | 0 |

### Commit Message Format

```
<type>(<scope>): <subject>

<body>

<footer>

Types: feat, fix, docs, style, refactor, test, chore
Example: feat(auth): add JWT refresh token support
```
