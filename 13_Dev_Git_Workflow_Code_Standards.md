# Git Workflow & Code Standards Guide

**Document ID**: 13_Dev_Git_Workflow_Code_Standards
**Version**: 1.0
**Last Updated**: 2026-03-14
**Owner**: Technical Lead

---

## 📖 Git Workflow (Git Flow)

### Branch Strategy

```
main (Production)
  ↑
release/1.0 (Pre-release)
  ↑
develop (Integration)
  ↑
feature/*, bugfix/* (Development)
```

### Branch Naming Convention

```
format: [type]/[category]/[description]

Types:
  feature/   - New feature
  bugfix/    - Bug fix
  hotfix/    - Urgent production fix
  release/   - Release preparation
  refactor/  - Code refactoring

Examples:
  ✅ feature/user-authentication
  ✅ feature/payment-integration
  ✅ bugfix/api-token-validation
  ✅ refactor/database-queries
  ✅ hotfix/payment-critical-bug
```

### Step-by-Step Workflow

#### 1. Start New Feature/Fix

```bash
# Update develop branch
git checkout develop
git pull origin develop

# Create feature branch
git checkout -b feature/user-authentication

# Push to remote (for backup)
git push origin feature/user-authentication
```

#### 2. Daily Development

```bash
# Make changes to files
# Edit code in your IDE

# Check what changed
git status

# Stage changes
git add .
# Or stage specific files
git add src/components/LoginForm.tsx

# Commit changes
git commit -m "feat(auth): implement login form component"
# See Commit Message Convention below

# Push to remote
git push origin feature/user-authentication
```

#### 3. Before Creating Pull Request

```bash
# Fetch latest changes
git fetch origin

# Rebase on latest develop (to get any updates)
git rebase origin/develop

# If conflicts occur, resolve them
# Then: git rebase --continue

# Verify build passes locally
npm run build    # Frontend
go test ./...    # Backend

# Run linter
npm run lint     # Frontend
golangci-lint run ./...  # Backend
```

#### 4. Create Pull Request

```bash
# Push rebased code
git push origin feature/user-authentication --force-with-lease
# (Use force-with-lease, NOT force!)

# Go to GitHub/GitLab and create PR:
# - Base: develop
# - Compare: feature/user-authentication
# - Title: "feat: implement user authentication"
# - Description: [See PR Template below]
```

#### 5. Code Review & Merge

```bash
# Wait for reviews (minimum 1 approval)
# Address feedback and push updates:

# Make changes
git add .
git commit -m "refactor(auth): improve error handling"
git push origin feature/user-authentication

# Once approved, merge via GitHub/GitLab UI
# Delete branch after merge
```

---

## 📝 Commit Message Convention

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Type (Required)

```
feat:     A new feature
fix:      A bug fix
docs:     Documentation changes
style:    Code style changes (formatting, semicolons, etc.)
refactor: Code refactoring
perf:     Performance improvements
test:     Test additions/changes
ci:       CI/CD changes
chore:    Dependency updates, tooling
```

### Scope (Optional)

```
Affected component/module
Examples: auth, payment, database, api-gateway, frontend
```

### Subject (Required)

```
Rules:
  • Use imperative mood ("add" not "added" or "adds")
  • Don't capitalize first letter
  • No period at the end
  • Limit to 50 characters
```

### Body (Optional but Recommended)

```
Explain what and why, not how
Separate from subject with blank line
Wrap at 72 characters
Can include:
  • Motivation for the change
  • Contrast with previous behavior
  • Related issues
```

### Footer (Optional)

```
Reference issues:
Fixes #123
Closes #456
Related to #789

Breaking changes:
BREAKING CHANGE: description of what broke and migration path
```

### Examples

**Good**:
```
feat(auth): implement JWT token refresh mechanism

Add automatic token refresh when access token expires.
This prevents users from being logged out unexpectedly.

Fixes #42
```

**Good**:
```
fix(payment): handle failed payment status codes

The payment API returns 402 for insufficient funds,
which wasn't being handled properly. Now display
appropriate error message to user.
```

**Good**:
```
refactor(database): optimize user query performance

Changed from N+1 query to single JOIN query.
Performance improved 10x for large user lists.
```

**Bad**:
```
Updated stuff
Fixed things
Work in progress
WIP
```

---

## 🔍 Pull Request Process

### PR Title Format
```
[Type] Brief description

Examples:
  [Feature] User Authentication System
  [Bugfix] Payment Processing Error
  [Refactor] Database Query Optimization
```

### PR Template (Description)

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] New feature
- [ ] Bug fix
- [ ] Breaking change
- [ ] Documentation update

## Related Issue
Fixes #(issue number)

## Testing
Describe testing performed:
- [ ] Unit tests added/updated
- [ ] Integration tests passed
- [ ] Manual testing completed

## Screenshots (if applicable)
For UI changes, include screenshots

## Checklist
- [ ] Code follows style guidelines
- [ ] Comments added for complex logic
- [ ] Documentation updated
- [ ] Tests added/updated
- [ ] No new warnings generated
- [ ] Tested locally

## Performance Impact
Any performance considerations?

## Breaking Changes
Any breaking changes? If yes, describe migration path
```

### PR Review Checklist (Reviewer)

- [ ] Code is clear and understandable
- [ ] Tests adequately cover changes
- [ ] No performance regressions
- [ ] Security implications reviewed
- [ ] Documentation is accurate
- [ ] Follows team standards
- [ ] No hardcoded values (credentials, etc.)

### Merge Requirements

- [x] At least 1 approval (2 for main branch)
- [x] All CI/CD checks passing
- [x] No merge conflicts
- [x] Squash commits on merge (keep history clean)

---

## 💻 Code Style Standards

### General Rules

```
• 2-space indentation (no tabs)
• Max line length: 100 characters
• One variable declaration per line
• Meaningful variable/function names
• No console.log in production code
• Remove debug code before commit
```

### JavaScript/TypeScript

```typescript
// ✅ GOOD
const getUserById = async (id: string): Promise<User> => {
  const user = await database.users.findById(id);
  if (!user) {
    throw new UserNotFoundError(`User ${id} not found`);
  }
  return user;
};

// ❌ BAD
const getUserById = async (id) => {
  let u = await db.users.findById(id);
  if (u == null) {
    throw new Error('not found');
  }
  return u;
};

// Style rules
• Use const by default, let if needed, never var
• Use arrow functions for callbacks
• Use destructuring for object properties
• Use template literals for string interpolation
• Use async/await (not .then().catch())
• Use TypeScript types for function parameters
```

### React Components

```typescript
// ✅ GOOD
interface LoginFormProps {
  onSubmit: (credentials: Credentials) => Promise<void>;
  isLoading?: boolean;
}

const LoginForm: React.FC<LoginFormProps> = ({
  onSubmit,
  isLoading = false,
}) => {
  const [email, setEmail] = useState('');
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await onSubmit({ email, password: '' });
    } catch (err) {
      setError(err.message);
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      {/* Form content */}
    </form>
  );
};

export default LoginForm;

// Component rules
• Use functional components with hooks
• Define prop interfaces at top
• Use meaningful component names
• One component per file
• Keep components focused and small
• Extract complex logic to custom hooks
```

### Go Code

```go
// ✅ GOOD
package auth

import (
  "context"
  "errors"
)

// UserService handles user authentication operations
type UserService struct {
  db Database
}

// Authenticate validates user credentials
func (s *UserService) Authenticate(
  ctx context.Context,
  email string,
  password string,
) (*User, error) {
  if email == "" || password == "" {
    return nil, errors.New("email and password required")
  }

  user, err := s.db.GetUserByEmail(ctx, email)
  if err != nil {
    return nil, err
  }

  if !s.validatePassword(password, user.PasswordHash) {
    return nil, errors.New("invalid credentials")
  }

  return user, nil
}

// Style rules
• Use PascalCase for exported names
• Use camelCase for unexported names
• Write comments for exported functions
• Error returns should be last
• Use table-driven tests
• Keep functions small (< 20 lines)
• Interface segregation principle
```

### SQL Queries

```sql
-- ✅ GOOD
SELECT
  u.id,
  u.email,
  u.name,
  COUNT(o.id) as order_count
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
WHERE u.created_at > NOW() - INTERVAL '30 days'
GROUP BY u.id
ORDER BY order_count DESC;

-- Style rules
• Use uppercase for keywords (SELECT, FROM, WHERE)
• Use lowercase for table/column names
• Proper indentation
• Use meaningful aliases
• Avoid SELECT * (explicitly list columns)
• Use JOINs instead of subqueries when possible
• Add indexes for frequently searched columns
```

---

## 🧪 Testing Standards

### Frontend (Jest + React Testing Library)

```typescript
// ✅ GOOD
describe('LoginForm', () => {
  it('should display error message on login failure', async () => {
    const mockOnSubmit = jest.fn().mockRejectedValue(
      new Error('Invalid credentials')
    );

    render(<LoginForm onSubmit={mockOnSubmit} />);

    const emailInput = screen.getByLabelText(/email/i);
    const submitButton = screen.getByRole('button', { name: /login/i });

    await userEvent.type(emailInput, 'test@example.com');
    await userEvent.click(submitButton);

    const errorMessage = await screen.findByText(/invalid credentials/i);
    expect(errorMessage).toBeInTheDocument();
  });
});

// Test rules
• Test user behavior, not implementation
• Use data-testid sparingly (prefer semantic queries)
• Mock external dependencies
• Test one thing per test
• Use descriptive test names
• Aim for > 80% coverage
```

### Backend (Go Testing)

```go
// ✅ GOOD
func TestAuthenticateUser(t *testing.T) {
  tests := []struct {
    name      string
    email     string
    password  string
    wantErr   bool
    wantError string
  }{
    {
      name:     "valid credentials",
      email:    "test@example.com",
      password: "correct_password",
      wantErr:  false,
    },
    {
      name:      "invalid credentials",
      email:     "test@example.com",
      password:  "wrong_password",
      wantErr:   true,
      wantError: "invalid credentials",
    },
  }

  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      user, err := service.Authenticate(
        context.Background(),
        tt.email,
        tt.password,
      )

      if tt.wantErr && err == nil {
        t.Errorf("expected error, got nil")
      }
      if !tt.wantErr && err != nil {
        t.Errorf("unexpected error: %v", err)
      }
    })
  }
}

// Test rules
• Use table-driven tests
• Test both happy path and error cases
• Use meaningful test names
• Use subtests (t.Run())
• Aim for > 80% coverage
• Test concurrent scenarios
```

---

## 📋 Code Review Checklist

**Before submitting PR:**

- [ ] Code compiles without errors
- [ ] All tests pass locally
- [ ] Linter passes (no warnings)
- [ ] Code follows style guide
- [ ] Added tests for new functionality
- [ ] Updated documentation
- [ ] Commit messages are clear
- [ ] No debugging code left
- [ ] No hardcoded values
- [ ] No security issues

**Reviewer should verify:**

- [ ] Code is clear and maintainable
- [ ] Tests adequately cover changes
- [ ] No performance regressions
- [ ] Follows architectural patterns
- [ ] Error handling is appropriate
- [ ] Logging is appropriate
- [ ] Security considerations addressed
- [ ] Database migrations (if applicable)

---

## 🚀 Pre-Commit Hooks (Optional)

```bash
# Install husky (optional)
npm install husky --save-dev
npx husky install

# Add pre-commit hook
npx husky add .husky/pre-commit "npm run lint && npm run test:unit"

# Auto-formats code before commit
```

---

## 🔄 Handling Merge Conflicts

```bash
# If conflicts occur during rebase
git status  # See which files have conflicts

# Edit conflicted files
# Remove markers: <<<<<<<, =======, >>>>>>>
# Keep desired changes

# Mark as resolved
git add <file>

# Continue rebase
git rebase --continue

# Or abort if needed
git rebase --abort
```

---

## ✅ Daily Git Commands

```bash
# Start day
git checkout develop
git pull origin develop

# During day (multiple times)
git status
git add .
git commit -m "feat(module): description"
git push origin feature/branch-name

# Before PR
git fetch origin
git rebase origin/develop
npm run build && npm run lint

# After code review
git add .
git commit -m "fix(module): address review feedback"
git push origin feature/branch-name
```

---

## 📞 Common Issues & Solutions

### Issue: "Your branch has diverged"
```bash
git fetch origin
git rebase origin/develop
# (resolve conflicts if any)
git push origin feature/branch-name --force-with-lease
```

### Issue: "Failed to push branch"
```bash
# Pull latest
git pull origin feature/branch-name --rebase

# Push again
git push origin feature/branch-name
```

### Issue: "Accidentally committed to wrong branch"
```bash
# Undo last commit (keep changes)
git reset --soft HEAD~1

# Create correct branch
git checkout -b feature/correct-name

# Commit again
git commit -m "message"
```

---

## 🎯 Summary

| Aspect | Standard |
|--------|----------|
| **Branch** | feature/descriptive-name |
| **Commits** | feat(scope): description |
| **PR Review** | 1+ approval, all checks pass |
| **Code Style** | Prettier, ESLint, golangci-lint |
| **Testing** | >80% coverage |
| **Line Length** | 100 characters max |
| **Indentation** | 2 spaces |
| **Comments** | For complex logic only |

---

**Version**: 1.0
**Last Updated**: 2026-03-14
**Status**: Active & In Use
