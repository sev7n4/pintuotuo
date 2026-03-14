# 拼脱脱 (Pintuotuo) - Claude Code Configuration Guide

**Project**: Pintuotuo - B2B2C AI Token Secondary Market Platform
**Status**: MVP Development (Week 1-8, 2026-03-17 to 2026-05-11)
**Last Updated**: 2026-03-14
**Version**: 1.0

---

## 📋 Quick Start (5 Minutes)

```bash
# Clone and enter project
git clone https://github.com/pintuotuo/pintuotuo.git
cd pintuotuo

# Start all services with Docker
docker-compose up -d

# Install dependencies
cd frontend && npm install && cd ../backend && go mod download && cd ..

# Verify setup
npm start              # Frontend: http://localhost:3000
go run main.go         # Backend: http://localhost:8080
# Mock API: http://localhost:3001
```

---

## 📖 Table of Contents

1. [Project Overview](#project-overview)
2. [Development Environment](#development-environment)
3. [Git Workflow & Branch Strategy](#git-workflow--branch-strategy)
4. [Code Standards](#code-standards)
5. [Architecture & Technical Stack](#architecture--technical-stack)
6. [Testing Standards](#testing-standards)
7. [Deployment & Release](#deployment--release)
8. [Claude Code Guidance](#claude-code-guidance)
9. [Quick Command Reference](#quick-command-reference)
10. [Troubleshooting](#troubleshooting)

---

## Project Overview

### What is Pintuotuo?

Pintuotuo is a B2B2C AI Token secondary market platform that enables:
- **B端 (Businesses)**: Monetize unused computing resources through API tokenization
- **C端 (Consumers)**: Purchase AI tokens at discounted rates through group buying (拼团)
- **Platform**: Earn commission through transaction facilitation

### Core Features (MVP)

- User authentication with JWT
- Product/Token catalog management (B端)
- Group buying (拼团) with smart auto-completion
- Payment integration (Alipay/WeChat)
- Token consumption tracking and balance management
- Real-time API key provisioning

### Team Structure

- **Backend**: 4 engineers (Go + Node.js)
- **Frontend**: 3 engineers (React + TypeScript)
- **QA/DevOps**: 2.5 engineers
- **Product/Design**: 6+ team members

### Key Links

- **Architecture**: `05_Technical_Architecture_and_Tech_Stack.md`
- **API Spec**: `04_API_Specification.md`
- **Code Standards**: `13_Dev_Git_Workflow_Code_Standards.md`
- **Setup Guide**: `12_Dev_Setup_Environment_Configuration.md`

---

## Development Environment

### Required Software

| Tool | Version | Purpose |
|------|---------|---------|
| **Git** | Latest | Version control |
| **Docker** | 20.10+ | Container runtime |
| **Docker Compose** | 1.29+ | Multi-container orchestration |
| **Node.js** | 18+ | Frontend development |
| **npm** | 8+ | Package management |
| **Go** | 1.21+ | Backend development |

### IDE & Tools

**Recommended**: VS Code with extensions:
- Go (golang.go)
- ES7+ React/Redux (dsznajder.es7-react-js-snippets)
- Prettier - Code formatter
- ESLint
- Thunder Client or Postman (API testing)

### Environment Variables

Create `.env.development` in project root:

```bash
# Backend
BACKEND_PORT=8080
DATABASE_URL=postgresql://pintuotuo:dev_password_123@localhost:5432/pintuotuo_db
REDIS_URL=redis://localhost:6379
KAFKA_BROKERS=localhost:9092

# Frontend
VITE_API_URL=http://localhost:8000
VITE_MOCK_API_URL=http://localhost:3001

# JWT
JWT_SECRET=dev_secret_key_change_in_production
JWT_EXPIRE_HOURS=24

# App
APP_ENV=development
APP_LOG_LEVEL=debug
```

### Docker Services

```bash
# Start all services
docker-compose up -d

# Services running on:
# PostgreSQL: localhost:5432 (pintuotuo/dev_password_123)
# Redis: localhost:6379
# Kafka: localhost:9092
# Mock API: localhost:3001
```

---

## Git Workflow & Branch Strategy

### Branch Strategy (Git Flow)

```
main (production)
  ↑
develop (integration)
  ↑
feature/* (development)
bugfix/* (bug fixes)
hotfix/* (production hotfixes)
```

### Branch Naming Convention

```
format: [type]/[category]/[description]

Types:
  feature/   - New feature
  bugfix/    - Bug fix
  hotfix/    - Urgent production fix
  refactor/  - Code refactoring
  release/   - Release preparation

Examples:
  feature/user-authentication
  feature/payment-integration
  bugfix/api-token-validation
  refactor/database-queries
```

### Standard Workflow

```bash
# 1. Create feature branch from develop
git checkout develop
git pull origin develop
git checkout -b feature/my-feature

# 2. Daily development
git add <specific-files>
git commit -m "feat(scope): description"
git push origin feature/my-feature

# 3. Before PR: rebase on latest develop
git fetch origin
git rebase origin/develop
npm run build && npm run lint

# 4. Create PR on GitHub/GitLab
# Base: develop
# Title: [Type] Brief description
# Description: Use template (see PR Template below)

# 5. After approval, merge via GitHub UI
# Use "Squash and merge" to keep history clean
```

### Commit Message Format

```
<type>(<scope>): <subject>

<body>

<footer>

Types: feat, fix, docs, style, refactor, perf, test, ci, chore
Scope: auth, payment, product, order, group, api, database, etc.
Subject: Use imperative mood, 50 chars max, no period

Example:
feat(auth): implement JWT token refresh mechanism

Add automatic token refresh when access token expires.
This prevents users from being logged out unexpectedly.

Fixes #42
```

### PR Template

```markdown
## Description
[Brief description of changes]

## Type of Change
- [ ] New feature
- [ ] Bug fix
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests passed
- [ ] Manual testing completed

## Checklist
- [ ] Code follows style guidelines
- [ ] Comments added for complex logic
- [ ] Documentation updated
- [ ] Tests added/updated
- [ ] No new warnings generated
- [ ] Tested locally
```

### Merge Requirements

- ✅ At least 1 approval (2 for main branch)
- ✅ All CI/CD checks passing
- ✅ No merge conflicts
- ✅ Squash commits on merge

---

## Code Standards

### General Rules (All Languages)

```
• 2-space indentation (no tabs)
• Max line length: 100 characters
• Meaningful variable/function names
• No console.log in production code
• Remove debug code before commit
• One variable declaration per line
```

### JavaScript/TypeScript

```typescript
// ✅ GOOD: Type annotations, proper error handling
const getUserById = async (id: string): Promise<User> => {
  const user = await database.users.findById(id);
  if (!user) {
    throw new UserNotFoundError(`User ${id} not found`);
  }
  return user;
};

// ❌ BAD: Missing types, poor naming
const getUserById = async (id) => {
  let u = await db.users.findById(id);
  if (u == null) throw new Error('not found');
  return u;
};

Style Rules:
• Use const by default, let if needed, never var
• Use arrow functions for callbacks
• Use destructuring for objects
• Use template literals for string interpolation
• Use async/await (not .then().catch())
• Always type function parameters and returns
```

### React Components

```typescript
// ✅ GOOD: Functional component with hooks
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
      setError(err instanceof Error ? err.message : 'Unknown error');
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      {error && <ErrorAlert message={error} />}
      {/* Form content */}
    </form>
  );
};

export default LoginForm;

Rules:
• Use functional components with hooks
• Define prop interfaces at top
• One component per file
• Keep components focused and small (< 200 lines)
• Extract complex logic to custom hooks
• Use semantic HTML
```

### Go Code

```go
// ✅ GOOD: Clear structure, good error handling
package auth

import (
  "context"
  "errors"
)

// UserService handles authentication operations
type UserService struct {
  db Database
}

// Authenticate validates user credentials
func (s *UserService) Authenticate(
  ctx context.Context,
  email, password string,
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

Rules:
• Use PascalCase for exported names (public)
• Use camelCase for unexported names (private)
• Write comments for all exported functions
• Error returns should be last
• Use table-driven tests
• Keep functions small (< 20 lines target)
• Interface segregation principle
```

### SQL Queries

```sql
-- ✅ GOOD: Clear, efficient
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

Rules:
• Use uppercase for keywords (SELECT, FROM, WHERE)
• Use lowercase for table/column names
• Proper indentation and formatting
• Use meaningful aliases
• Avoid SELECT * (list columns explicitly)
• Use JOINs instead of subqueries
• Create indexes for frequently searched columns
```

### File Organization

```
frontend/
  src/
    components/        # Reusable UI components
      LoginForm/
        LoginForm.tsx
        LoginForm.test.tsx
        LoginForm.module.css
    pages/             # Page components
    hooks/             # Custom React hooks
    stores/            # Zustand state management
    services/          # API clients
    types/             # TypeScript interfaces
    utils/             # Utility functions

backend/
  internal/            # Private packages
    auth/              # Auth domain
      handler.go       # HTTP handlers
      service.go       # Business logic
      repository.go    # Database access
    handlers.go        # Main HTTP routes
  pkg/                 # Public packages
  main.go             # Entry point
  go.mod, go.sum      # Dependencies
```

---

## Architecture & Technical Stack

### Technology Stack

| Layer | Technology | Purpose |
|-------|-----------|---------|
| **Frontend** | React 18 + TypeScript | Web UI |
| **State Mgmt** | Zustand | Client-side state |
| **HTTP Client** | Axios | API requests |
| **Backend** | Go 1.21 | High-performance API services |
| **Framework** | Gin | Lightweight web framework |
| **Database** | PostgreSQL 15 | Primary data store |
| **Cache** | Redis 7 | Session, cache, real-time data |
| **Messaging** | Kafka 7.5 | Async task processing |
| **Container** | Docker + K8s | Deployment |

### Core Services Architecture

```
Client Apps (Web, iOS, Android)
    ↓
API Gateway (Kong/Nginx) - Auth, Rate Limiting, Routing
    ↓
┌─────────────────────────────────────────┐
│ Microservices                           │
├─────────────────────────────────────────┤
│ User Service      - Auth, user mgmt     │
│ Product Service   - Product catalog     │
│ Order Service     - Order management    │
│ Group Service     - Group buying logic  │
│ Token Service     - Token balance       │
│ Payment Service   - Payment processing  │
└─────────────────────────────────────────┘
    ↓
┌─────────────────────────────────────────┐
│ Data & Cache Layer                      │
├─────────────────────────────────────────┤
│ PostgreSQL (Master + Read Replicas)     │
│ Redis (Cache + Session)                 │
│ Kafka (Event Streaming)                 │
└─────────────────────────────────────────┘
```

### Key Design Decisions

1. **Microservices**: Each domain has independent service for scalability
2. **Event-Driven**: Async processing via Kafka for non-critical tasks
3. **Caching Strategy**: Multi-layer (browser → CDN → Redis → DB)
4. **Database**: PostgreSQL with read replicas for performance
5. **API Gateway**: Centralized authentication, rate limiting, routing

### API Design Principles

```
RESTful endpoints:
  GET    /api/v1/products               - List
  GET    /api/v1/products/{id}          - Detail
  POST   /api/v1/products               - Create
  PUT    /api/v1/products/{id}          - Update
  DELETE /api/v1/products/{id}          - Delete

Authentication:
  Authorization: Bearer <jwt_token>

Error Responses:
  {
    "error": "error_code",
    "message": "Human readable message",
    "details": { "field": "error details" }
  }

Pagination:
  ?page=1&limit=20&sort=created_at&order=desc
```

---

## Testing Standards

### Frontend Testing (Jest + React Testing Library)

```typescript
// ✅ GOOD: Test user behavior, not implementation
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

Best Practices:
• Test user interactions, not implementation details
• Use semantic queries (getByRole, getByLabelText)
• Mock external dependencies
• One assertion per test (or related assertions)
• Use descriptive test names
• Aim for > 80% coverage
```

### Backend Testing (Go Testing)

```go
// ✅ GOOD: Table-driven tests
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
      user, err := service.Authenticate(context.Background(), tt.email, tt.password)

      if tt.wantErr && err == nil {
        t.Errorf("expected error, got nil")
      }
      if !tt.wantErr && err != nil {
        t.Errorf("unexpected error: %v", err)
      }
    })
  }
}

Best Practices:
• Use table-driven tests
• Test both happy and error cases
• Use meaningful test names
• Use subtests (t.Run())
• Aim for > 80% coverage
• Test concurrent scenarios
```

### Test Running

```bash
# Frontend
npm test                  # Run tests
npm run test:watch       # Watch mode
npm run test:coverage    # Coverage report

# Backend
go test ./...            # Run all tests
go test -v ./...         # Verbose output
go test -cover ./...     # Coverage report
go test -race ./...      # Check race conditions
```

---

## Deployment & Release

### Environments

| Environment | Purpose | URL | Deploy Frequency |
|-------------|---------|-----|------------------|
| **Development** | Local dev | localhost:3000 | Every commit |
| **Testing** | QA testing | test.pintuotuo.com | Daily |
| **Staging** | Pre-production | staging.pintuotuo.com | Weekly |
| **Production** | Live users | pintuotuo.com | Bi-weekly |

### CI/CD Pipeline

```
1. Code Push → GitHub/GitLab
   ↓
2. Run Tests
   ├─ Unit tests
   ├─ Integration tests
   └─ Linting
   ↓
3. Build Artifacts
   ├─ Frontend: npm run build
   ├─ Backend: go build
   └─ Docker: Build images
   ↓
4. Push to Registry
   ├─ Docker Hub or
   └─ Private Registry
   ↓
5. Deploy to K8s
   ├─ Rolling update
   ├─ Health checks
   └─ Rollback on failure
```

### Deployment Steps

```bash
# 1. Ensure all tests pass
npm run test      # Frontend
go test ./...     # Backend

# 2. Build production artifacts
npm run build     # Frontend
go build          # Backend

# 3. Create Docker image
docker build -t pintuotuo:v1.0.0 .

# 4. Push to registry
docker push pintuotuo:v1.0.0

# 5. Deploy to K8s
kubectl apply -f deployment.yaml

# 6. Verify deployment
kubectl get pods
kubectl logs -f <pod-name>
```

### Version Bumping

Use semantic versioning (MAJOR.MINOR.PATCH):
- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes

```bash
# Tag release in Git
git tag v1.0.0
git push origin v1.0.0
```

---

## Claude Code Guidance

### When to Use Certain Approaches

#### Creating New Files
- **Do create** new feature files (components, services, handlers)
- **Don't create** utility files for single-use functions
- **Do ask user** before creating large new files or restructuring
- **Check first**: Is there an existing pattern/file to extend?

#### Code Generation
- **Always add**: Error handling, type annotations, basic validation
- **Never skip**: Testing, documentation for public APIs
- **Ask user**: If unsure about architectural decisions
- **Reference**: Existing code patterns in the project

#### Error Handling
```typescript
// ✅ DO THIS: Specific error handling
try {
  const user = await fetchUser(id);
  return user;
} catch (error) {
  if (error instanceof NotFoundError) {
    throw new NotFoundError(`User ${id} not found`);
  }
  throw error;
}

// ❌ DON'T DO THIS: Generic error handling
try {
  // ...
} catch (e) {
  console.log('error');
}
```

#### Comments & Documentation
- ✅ Add comments for complex business logic
- ✅ Add JSDoc for public/exported functions
- ✅ Explain the "why", not the "what"
- ❌ Don't comment obvious code
- ❌ Don't add comments for every line

```typescript
// ✅ GOOD: Explains the why
/**
 * Validates user's password strength.
 * Uses NIST guidelines: min 8 chars, no common patterns.
 * @param password - User input password
 * @returns true if password meets requirements
 */
const isValidPassword = (password: string): boolean => {
  // Implementation...
};

// ❌ BAD: States the obvious
// Check if password has more than 5 characters
const isValidPassword = (password: string): boolean => {
  return password.length > 5;
};
```

#### Security Best Practices
- ✅ Always validate user input
- ✅ Use prepared statements for SQL queries (GORM, Sequelize)
- ✅ Hash passwords (bcrypt, not SHA)
- ✅ Validate JWT tokens on protected endpoints
- ✅ Use HTTPS in production
- ❌ Never hardcode secrets in code
- ❌ Never trust client-side validation alone
- ❌ Never log sensitive data

#### Performance Considerations
- ✅ Use caching for frequently accessed data
- ✅ Batch database queries when possible
- ✅ Lazy load components in React
- ✅ Use pagination for large lists
- ❌ Don't fetch all data upfront
- ❌ Don't make synchronous blocking calls

### Common Workflows with Claude Code

**Feature Implementation**:
```bash
1. Read existing similar code to understand patterns
2. Create or update files following project conventions
3. Add tests before or alongside implementation
4. Ensure code passes linter and tests
5. Ask for user feedback if architectural decisions needed
```

**Bug Fixing**:
```bash
1. Locate the problematic code
2. Add test case that reproduces the bug
3. Fix the issue
4. Verify test passes and related tests still pass
5. Create PR with clear description
```

**Code Review**:
```bash
1. Check style consistency with project standards
2. Verify error handling is appropriate
3. Ensure tests are adequate
4. Look for security issues
5. Provide constructive feedback
```

---

## Quick Command Reference

### Git Commands

```bash
# Setup
git clone <repo>
git config user.name "Your Name"
git config user.email "your@email.com"

# Daily workflow
git status                              # Check status
git pull origin develop                 # Get latest
git checkout -b feature/my-feature      # Create branch
git add .                               # Stage changes
git commit -m "feat(scope): description"
git push origin feature/my-feature      # Push

# Before PR
git fetch origin
git rebase origin/develop               # Rebase on develop
git push origin feature/my-feature --force-with-lease

# After conflicts
git status                              # See conflicts
# Edit files to resolve
git add <resolved-files>
git rebase --continue
```

### Docker Commands

```bash
# Start/stop
docker-compose up -d                    # Start all services
docker-compose down                     # Stop all services
docker-compose restart <service>        # Restart service

# Debugging
docker-compose ps                       # List containers
docker-compose logs -f <service>        # View logs
docker exec -it <container> bash        # Shell into container

# Database
docker exec pintuotuo_postgres psql -U pintuotuo -d pintuotuo_db
```

### Frontend Commands

```bash
# Development
npm install                             # Install deps
npm start                               # Dev server (http://localhost:3000)
npm run dev                             # Vite dev server
npm run build                           # Production build
npm run preview                         # Preview build

# Quality
npm run lint                            # Check code style
npm run format                          # Auto-format code
npm test                                # Run tests
npm run test:coverage                   # Test coverage report
npm run analyze                         # Bundle analysis
```

### Backend Commands

```bash
# Setup
go mod download                         # Download deps
go mod tidy                             # Clean up deps
go mod verify                           # Verify modules

# Development
go run main.go                          # Run server
go build                                # Build binary

# Quality
go fmt ./...                            # Format code
go vet ./...                            # Static analysis
go test ./...                           # Run tests
go test -cover ./...                    # Coverage report
golangci-lint run ./...                 # Comprehensive linting
```

### Database Commands

```bash
# Connect to database
psql -h localhost -U pintuotuo -d pintuotuo_db

# Common queries
\dt                                     # List tables
\d <table_name>                         # Show table structure
SELECT COUNT(*) FROM users;             # Count rows
\q                                      # Exit

# Backup/restore
pg_dump -h localhost -U pintuotuo pintuotuo_db > backup.sql
psql -h localhost -U pintuotuo pintuotuo_db < backup.sql
```

---

## Troubleshooting

### Docker Issues

**"Docker daemon is not running"**
```bash
# macOS/Windows: Start Docker Desktop
# Linux: sudo systemctl start docker
```

**"Port already in use"**
```bash
# Find and kill process using port (e.g., 5432)
lsof -i :5432
kill -9 <PID>

# Or change port in docker-compose.yml
```

**"Cannot connect to database"**
```bash
# Verify container is running
docker ps

# Check logs
docker-compose logs postgres

# Restart service
docker-compose restart postgres
```

### Node/npm Issues

**"npm ERR! ERESOLVE unable to resolve dependency tree"**
```bash
npm cache clean --force
npm install --legacy-peer-deps
```

**"Module not found"**
```bash
rm -rf node_modules package-lock.json
npm install
```

### Go Issues

**"go: command not found"**
```bash
# Add Go to PATH
export PATH=$PATH:/usr/local/go/bin
# Add to ~/.bashrc or ~/.zshrc for permanent
```

**"go mod tidy fails"**
```bash
go clean -modcache
go mod tidy
```

### Build Issues

**"TypeScript compilation errors"**
```bash
# Check syntax
npx tsc --noEmit

# Clear cache
rm -rf node_modules/.vite
npm install
```

**"Go build failures"**
```bash
go clean -cache
go build -v ./...
```

---

## Key File Locations

| File/Directory | Purpose |
|---|---|
| `/frontend` | React + TypeScript web application |
| `/backend` | Go API services |
| `/services` | Additional microservices |
| `/scripts` | Helper scripts (db, docker, etc.) |
| `.env.development` | Development environment config |
| `docker-compose.yml` | Docker service definitions |
| `13_Dev_Git_Workflow_Code_Standards.md` | Complete git/code standards |
| `05_Technical_Architecture_and_Tech_Stack.md` | Architecture details |
| `12_Dev_Setup_Environment_Configuration.md` | Environment setup guide |
| `04_API_Specification.md` | API endpoint documentation |

---

## Getting Help

**Setup Issues?**
- Check "Troubleshooting" section above
- Review `12_Dev_Setup_Environment_Configuration.md`
- Ask in #development Slack channel

**Code Questions?**
- Review `13_Dev_Git_Workflow_Code_Standards.md` for standards
- Check `05_Technical_Architecture_and_Tech_Stack.md` for architecture decisions
- Ask in #engineering Slack or tag @tech-lead

**API Questions?**
- Check `04_API_Specification.md`
- Review existing service implementations for patterns
- Ask in #api-design Slack channel

---

## Summary Checklist

Before committing code, verify:

- [ ] Code follows 2-space indentation and 100-char line limit
- [ ] TypeScript types are present, no `any` types
- [ ] Tests are passing (`npm test` / `go test ./...`)
- [ ] Linter is passing (`npm run lint` / `golangci-lint`)
- [ ] No console.log or debug code
- [ ] Error handling is appropriate
- [ ] Comments added for complex logic
- [ ] Commit message follows convention: `type(scope): message`
- [ ] Branch rebased on latest develop
- [ ] PR description includes what and why

---

**Document Version**: 1.0
**Last Updated**: 2026-03-14
**Status**: Active & In Use
**Maintained By**: CTO / Technical Lead
