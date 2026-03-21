# Quick Reference Guide

## Quick Start

### 1. Submit Issue/Request

Use the following format to describe issues or requirements:

```
Type: [bug|feature|enhancement]
Title: Brief description
Description: Detailed description of the issue or requirement
Priority: [high|medium|low]
Scope: [backend|frontend|both]
```

### 2. Workflow Auto Execution

After submission, the system will automatically execute the following process:

1. **Plan Generation** → Generate plan document and task list
2. **Branch Creation** → Create new branch per conventions
3. **Code Analysis** → Locate issue code
4. **Code Implementation** → Modify and improve code
5. **Test Writing** → Write unit/integration/E2E tests
6. **Local Verification** → Run full tests
7. **Code Commit** → Commit and push code
8. **CI Verification** → Monitor GitHub workflow
9. **Issue Fix** → Loop fix if failed
10. **Generate Summary** → Update issue tracking document

## Branch Naming Conventions

| Type | Format | Example |
|------|--------|---------|
| Bug Fix | `bugfix/issue-{id}-{desc}` | `bugfix/issue-123-fix-login` |
| New Feature | `feature/issue-{id}-{desc}` | `feature/issue-456-add-payment` |
| Enhancement | `enhancement/issue-{id}-{desc}` | `enhancement/issue-789-optimize-db` |
| Hot Fix | `hotfix/issue-{id}-{desc}` | `hotfix/issue-100-patch-security` |

## Commit Message Conventions

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types**: `feat` | `fix` | `docs` | `style` | `refactor` | `test` | `chore`

**Example**:
```
feat(auth): add OAuth2 login support

- Add Google OAuth2 provider
- Add GitHub OAuth2 provider

Closes #123
```

## Test Commands Cheat Sheet

### Backend Tests

```bash
# Unit tests
cd backend && go test -v -short -race -coverprofile=coverage.out ./...

# Integration tests
cd backend && go test -v -run Integration ./...

# View coverage
cd backend && go tool cover -html=coverage.out
```

### Frontend Tests

```bash
# Unit tests
cd frontend && npm test -- --coverage --watchAll=false

# E2E tests
cd frontend && npm run test:e2e

# E2E test report
cd frontend && npm run test:e2e:report
```

### Full Tests

```bash
make test
```

## GitHub Workflow Trigger Sequence

```
push/PR
    ↓
ci-cd.yml (unit tests + build + security scan)
    ↓
integration-tests.yml (integration tests)
    ↓
e2e-tests.yml (E2E tests)
```

## Document Locations

All files are relative to `.trae/skills/mydev-github-workflow-en/` directory:

| Document | Path |
|----------|------|
| Workflow Design | `references/design.md` |
| Decision Guide | `references/decision_guide.md` |
| Error Reference | `references/error_reference.md` |
| Issue Tracking | `references/issue_tracking.md` |
| Workflow History | `references/workflow_history.md` |
| Plan Template | `assets/templates/plan_template.md` |
| Task Template | `assets/templates/tasks_template.md` |
| PR Template | `assets/templates/pr_template.md` |
| State Management | `scripts/workflow_state.json` |

## Common Issues

### Q: What to do if tests fail?

1. View failure logs
2. Analyze failure reason
3. Reproduce issue locally
4. Fix code or tests
5. Re-submit

### Q: What to do if workflow fails?

1. Check GitHub Actions logs
2. Locate failed step
3. Analyze error message
4. Fix and re-push

### Q: How to view test coverage?

```bash
# Backend
cd backend && go tool cover -func=coverage.out

# Frontend
cd frontend && npm test -- --coverage
```

## Quality Standards

| Metric | Backend | Frontend |
|--------|---------|----------|
| Unit Test Coverage | ≥85% | ≥80% |
| Integration Tests | Core flow coverage | - |
| E2E Tests | - | Main user flows |
| Code Style | golangci-lint | ESLint |
| Security Scan | Trivy | npm audit |
