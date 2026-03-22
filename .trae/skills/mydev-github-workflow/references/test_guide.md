# TDD 测试指南

> 本文档定义 TDD 核心原则、测试策略、命令和覆盖率要求。

## TDD 核心原则

```
Red-Green-Refactor 循环:
1. Red: 先写失败的测试
2. Green: 写最小代码让测试通过
3. Refactor: 在测试保护下优化代码
```

---

## 测试策略矩阵

| 影响范围 | 单元测试 | 集成测试 | E2E测试 |
|----------|----------|----------|---------|
| backend | ✅ 必须 | ✅ 必须 | ❌ 不需要 |
| frontend | ✅ 必须 | ❌ 不需要 | ✅ 必须 |
| both | ✅ 必须 | ✅ 必须 | ✅ 必须 |

---

## 覆盖率要求

| 层级 | 最低要求 |
|------|----------|
| Backend Core | ≥85% |
| Backend API | ≥80% |
| Frontend | ≥80% |

---

## 测试命令

```bash
# 后端单元测试
cd backend && go test -v -race -coverprofile=coverage.out ./...

# 后端集成测试
cd backend && go test -v -run Integration ./...

# 前端单元测试
cd frontend && npm test -- --coverage --watchAll=false

# 前端E2E测试
cd frontend && npm run test:e2e
```

---

## 测试检查清单

- [ ] 测试名称描述场景和预期结果
- [ ] 测试独立（无共享状态）
- [ ] 边界条件覆盖
- [ ] 错误路径测试
- [ ] 覆盖率达标
