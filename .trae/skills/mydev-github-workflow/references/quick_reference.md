# GitHub集成工作流快速参考指南

## 快速开始

### 1. 提交问题/需求

使用以下格式描述问题或需求：

```
类型: [bug|feature|enhancement]
标题: 简短描述
描述: 详细描述问题或需求
优先级: [high|medium|low]
影响范围: [backend|frontend|both]
```

### 2. 工作流自动执行

提交后，系统将自动执行以下流程：

1. **制定计划** → 生成计划文档和任务清单
2. **创建分支** → 按规范创建新分支
3. **代码分析** → 定位问题代码
4. **代码实现** → 修改和完善代码
5. **测试编写** → 编写单元/集成/E2E测试
6. **本地验证** → 运行全量测试
7. **代码提交** → 提交并推送代码
8. **CI验证** → 监控GitHub工作流
9. **问题修复** → 如有失败，循环修复
10. **生成总结** → 更新问题跟踪文档

## 分支命名规范

| 类型 | 格式 | 示例 |
|------|------|------|
| Bug修复 | `bugfix/issue-{id}-{desc}` | `bugfix/issue-123-fix-login` |
| 新功能 | `feature/issue-{id}-{desc}` | `feature/issue-456-add-payment` |
| 增强 | `enhancement/issue-{id}-{desc}` | `enhancement/issue-789-optimize-db` |
| 紧急修复 | `hotfix/issue-{id}-{desc}` | `hotfix/issue-100-patch-security` |

## 提交信息规范

```
<type>(<scope>): <subject>

<body>

<footer>
```

**类型**：`feat` | `fix` | `docs` | `style` | `refactor` | `test` | `chore`

**示例**：
```
feat(auth): add OAuth2 login support

- Add Google OAuth2 provider
- Add GitHub OAuth2 provider

Closes #123
```

## 测试命令速查

### 后端测试

```bash
# 单元测试
cd backend && go test -v -short -race -coverprofile=coverage.out ./...

# 集成测试
cd backend && go test -v -run Integration ./...

# 查看覆盖率
cd backend && go tool cover -html=coverage.out
```

### 前端测试

```bash
# 单元测试
cd frontend && npm test -- --coverage --watchAll=false

# E2E测试
cd frontend && npm run test:e2e

# E2E测试报告
cd frontend && npm run test:e2e:report
```

### 全量测试

```bash
make test
```

## GitHub工作流触发顺序

```
push/PR
    ↓
ci-cd.yml (单元测试 + 构建 + 安全扫描)
    ↓
integration-tests.yml (集成测试)
    ↓
e2e-tests.yml (E2E测试)
```

## 文档位置

| 文档 | 路径 |
|------|------|
| 工作流设计 | `references/design.md` |
| 决策指导 | `references/decision_guide.md` |
| 错误参考 | `references/error_reference.md` |
| 问题跟踪 | `references/issue_tracking.md` |
| 工作流历史 | `references/workflow_history.md` |
| 计划模板 | `assets/templates/plan_template.md` |
| 任务模板 | `assets/templates/tasks_template.md` |
| PR模板 | `assets/templates/pr_template.md` |

## 常见问题

### Q: 测试失败怎么办？

1. 查看失败日志
2. 分析失败原因
3. 本地复现问题
4. 修复代码或测试
5. 重新提交

### Q: 工作流失败怎么办？

1. 检查GitHub Actions日志
2. 定位失败步骤
3. 分析错误信息
4. 修复并重新推送

### Q: 如何查看测试覆盖率？

```bash
# 后端
cd backend && go tool cover -func=coverage.out

# 前端
cd frontend && npm test -- --coverage
```

## 质量标准

| 指标 | 后端 | 前端 |
|------|------|------|
| 单元测试覆盖率 | ≥85% | ≥80% |
| 集成测试 | 核心流程覆盖 | - |
| E2E测试 | - | 主要用户流程 |
| 代码风格 | golangci-lint | ESLint |
| 安全扫描 | Trivy | npm audit |
