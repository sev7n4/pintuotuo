# TDD Test Guide

> This document defines TDD core principles, test strategy, commands, and coverage requirements.

## TDD Core Principles

```
Red-Green-Refactor Cycle:
1. Red: Write failing test first
2. Green: Write minimal code to pass
3. Refactor: Optimize under test protection
```

---

## Test Strategy Matrix

| Impact Scope | Unit Tests | Integration Tests | E2E Tests |
|--------------|------------|-------------------|-----------|
| backend | ✅ Required | ✅ Required | ❌ Not needed |
| frontend | ✅ Required | ❌ Not needed | ✅ Required |
| both | ✅ Required | ✅ Required | ✅ Required |

---

## Coverage Requirements

| Layer | Minimum |
|-------|---------|
| Backend Core | ≥85% |
| Backend API | ≥80% |
| Frontend | ≥80% |

---

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

---

## Test Checklist

- [ ] Test name describes scenario and expected result
- [ ] Test is independent (no shared state)
- [ ] Edge cases covered
- [ ] Error paths tested
- [ ] Coverage meets requirements
