---
name: "mydev-github-workflow"
description: |
Automated development workflow for bug fixes, features, and code changes with GitHub CI/CD integration. 
Use when user: (1) reports a bug ("登录失败", "返回401", "有个错误"), (2) requests a feature ("添加功能", "实现一个", "新增"), (3) describes improvements ("优化", "重构", "改进"), or (4) provides structured issue input (type/title/description fields).
---

# MyDev GitHub Workflow

## Overview

自动化开发工作流，实现从问题输入到代码合并的完整闭环。集成GitHub CI/CD，自动执行：问题解析 → 计划生成 → 分支创建 → 代码实现 → 测试编写 → 本地验证 → 提交推送 → CI监控 → 文档更新。

## Quick Start

```
用户：登录功能返回401错误

→ 自动执行完整流程，直到CI通过并创建PR
```

```
用户：
类型: feature
标题: 添加商品收藏功能
描述: 用户可以收藏喜欢的商品
优先级: medium
影响范围: both

→ 自动执行开发流程
```

## When to Use

| 触发场景 | 用户可能说的话 |
|----------|----------------|
| Bug修复 | "有个bug"、"登录失败"、"返回401错误"、"出问题了" |
| 新功能 | "添加功能"、"实现一个"、"新增"、"开发" |
| 代码改进 | "优化"、"重构"、"改进"、"提升" |
| 结构化输入 | 提供type/title/description字段 |

## Workflow Decision Tree

```
Step 0: 初始化检查
    ↓
    ├─ Git有未提交更改? → 询问用户: 暂存/提交/忽略
    └─ 有未完成任务? → 询问用户: 继续/放弃
    ↓
Step 1: 问题解析 ──→ 判断类型
    ↓                   ├─ bug → bugfix/issue-{id}
    ↓                   ├─ feature → feature/issue-{id}
    ↓                   └─ enhancement → enhancement/issue-{id}
    ↓
Step 2: 计划生成 ──→ 读取模板 → 生成计划和任务文档
    ↓
Step 3: 分支创建 ──→ git checkout -b {branch}
    ↓                   └─ 推送失败? → git pull --rebase → 重试
    ↓
Step 4: 代码分析 ──→ SearchCodebase → Grep → Read
    ↓                   └─ 找不到? → 询问用户
    ↓
Step 5: 测试设计 ──→ TDD: 先写失败测试 (Red)
    ↓                   ├─ Bug: 写复现测试
    ↓                   └─ Feature: 写验收测试
    ↓
Step 6: 最小实现 ──→ 写最少代码通过测试 (Green)
    ↓
Step 7: 重构优化 ──→ 测试保护下优化代码 (Refactor)
    ↓
Step 8: 本地验证 ──→ 运行测试 + 覆盖率检查
    ↓                   └─ 失败? → 修复 → 重新验证
    ↓
Step 9: 代码提交 ──→ git commit -m "..."
    ↓                   └─ 推送失败? → git pull --rebase → 重试
    ↓
Step 10: CI监控 ──→ gh run watch
    ↓                   ├─ 成功 → Step 12
    ↓                   └─ 失败 → Step 11 (最多5次)
    ↓
Step 11: 错误修复 ──→ 分析日志 → 修复 → 重新提交
    ↓
Step 12: 文档更新 ──→ 更新跟踪文档
    ↓
Step 13: 创建PR ──→ gh pr create
    ↓
Step 14: 清理 ──→ 更新状态 → 输出PR链接
    ↓
完成
```

## 核心步骤

### Step 0: 初始化检查

**目的**: 确保环境干净，避免中途失败

**检查项**:
```bash
# 检查Git状态
git status --porcelain
# 有输出? → 询问用户处理方式

# 检查未完成任务
# 读取 workflow_state.json
# stage != "idle" && stage != "completed"? → 询问用户
```

**决策点**: 有问题? → 询问用户，不自动处理

### Step 1: 问题解析

**判断类型**：
- 含"bug"、"错误"、"失败"、"异常" → `bug`
- 含"添加"、"实现"、"新增"、"开发" → `feature`  
- 含"优化"、"重构"、"改进"、"提升" → `enhancement`

**评估优先级**：
- 影响核心功能/安全 → `high`
- 影响次要功能 → `medium`
- 优化改进 → `low`

**确定范围**：
- 涉及API/数据库 → `backend`
- 涉及UI/交互 → `frontend`
- 两者都有 → `both`

**生成Issue ID**：读取 `scripts/workflow_state.json` 获取下一个序号

### Step 2: 计划生成

```
读取: assets/templates/plan_template.md
      assets/templates/tasks_template.md
生成: assets/plans/{YYYY-MM-DD}_issue_{id}_plan.md
      assets/tasks/{YYYY-MM-DD}_issue_{id}_tasks.md
```

### Step 3: 分支创建

```bash
git checkout main && git pull origin main
git fetch origin main
git log HEAD..origin/main --oneline  # 检查新提交
# 有新提交? → git rebase origin/main
git checkout -b bugfix/issue-001-login-401
git push -u origin bugfix/issue-001-login-401
```

**失败处理**: 推送失败? → `git pull --rebase origin main` → 重试推送

### Step 4: 代码分析

```
SearchCodebase: "登录认证逻辑"
Grep: "func.*Login|auth.*handler"
Read: backend/handlers/auth.go
```

**决策点**：找不到代码? → 询问用户提供更多信息

### Step 5: 测试设计 (TDD - Red)

**原则**: 先写失败测试，明确预期行为

**测试策略决策树**:
```
功能影响范围
    │
    ├─ backend (纯后端API)
    │   └─ 单元测试 + 集成测试
    │
    ├─ frontend (纯前端UI)
    │   └─ 单元测试 + E2E测试
    │
    └─ both (全栈功能)
        └─ 单元测试 + 集成测试 + E2E测试
```

**Bug修复流程**:
```
1. 分析Bug现象
2. 写单元测试复现Bug (测试应失败)
3. 如果涉及API交互 → 写集成测试
4. 如果涉及用户流程 → 写E2E测试
5. 确认所有测试失败 = Bug被捕获
```

**新功能流程**:
```
1. 定义验收标准
2. 根据影响范围确定测试类型:
   - both → E2E测试 + 集成测试 + 单元测试
   - backend → 集成测试 + 单元测试
   - frontend → E2E测试 + 单元测试
3. 按顺序编写测试:
   a. E2E测试 (验收测试，定义用户视角的行为)
   b. 集成测试 (API契约测试)
   c. 单元测试 (函数行为定义)
4. 所有测试应失败 (功能未实现)
```

**测试文件位置**:
| 测试类型 | 后端位置 | 前端位置 |
|----------|----------|----------|
| 单元测试 | `backend/handlers/*_test.go` | `frontend/src/**/*.test.ts` |
| 集成测试 | `backend/integration/*_test.go` | - |
| E2E测试 | - | `frontend/e2e/*.spec.ts` |

**测试命名规范**:
```
单元测试: Test{FunctionName}_{Scenario}_{ExpectedResult}
集成测试: Test{Feature}Integration_{Scenario}
E2E测试: {Feature} - {User Action} - {Expected Result}
```

**E2E测试模板**: 参考 `references/test_guide.md`

**详细测试指南**: `references/test_guide.md`

### Step 6: 最小实现 (TDD - Green)

**原则**: 写最少代码使测试通过

```
1. 实现最小功能满足测试
2. 不添加测试未覆盖的功能
3. 代码可以丑陋，但必须通过测试
```

**实现方式**:
- SearchReplace: 精确修改现有文件
- Write: 创建新文件
- 保持风格一致

### Step 7: 重构优化 (TDD - Refactor)

**原则**: 在测试保护下优化代码

```
1. 确保所有测试通过
2. 重构代码结构
3. 运行测试确认未破坏功能
4. 重复直到满意
```

**重构检查清单**:
- [ ] 消除重复代码
- [ ] 提取函数/方法
- [ ] 改善命名
- [ ] 简化条件逻辑

**参考**: `references/code_quality_guide.md`

### Step 8: 本地验证

**验证步骤**（按顺序执行）：

```bash
# 1. 后端单元测试
cd backend && go test -v -race -coverprofile=coverage.out ./...

# 2. 后端集成测试（如果有）
cd backend && go test -v -run Integration -race ./...

# 3. 前端单元测试
cd frontend && npm test -- --coverage --watchAll=false

# 4. 前端E2E测试（如果功能涉及UI）
cd frontend && npm run test:e2e
```

**验证检查清单**：
- [ ] 后端单元测试通过
- [ ] 后端集成测试通过（如有）
- [ ] 前端单元测试通过
- [ ] 前端E2E测试通过（如有）
- [ ] 覆盖率达标

**覆盖率要求**:
| 层级 | 最低覆盖率 |
|------|-----------|
| Backend Core | ≥85% |
| Backend API | ≥80% |
| Frontend | ≥80% |

**E2E测试前置条件**：
- 后端服务运行在 localhost:8080
- 前端服务运行在 localhost:5173
- 测试数据库已初始化

**决策点**：失败? → 分析日志 → 修复 → 重新验证

### Step 9: 代码提交

```bash
git add .
git commit -m "fix(auth): resolve login 401 error

- Fix jwtSecret initialization
- Add unit tests

Closes #ISSUE-001"
git push origin bugfix/issue-001-login-401
```

**失败处理**: 推送失败? → `git pull --rebase origin {branch}` → 解决冲突 → 重试推送

### Step 10: CI监控

**工作流触发链**：
```
PR创建/Push
    ↓
┌─────────────────┐
│ CI/CD Pipeline  │ ← 第一个触发
└────────┬────────┘
         │ workflow_run: completed (success)
         ↓
┌─────────────────┐
│ Integration     │ ← 自动触发
│ Tests           │
└────────┬────────┘
         │ workflow_run: completed (success)
         ↓
┌─────────────────┐
│ E2E Tests       │ ← 自动触发
└─────────────────┘
```

**监控步骤**：

```bash
# 1. 监控 CI/CD Pipeline
gh run list --workflow="CI/CD Pipeline" --branch=bugfix/issue-001-login-401 --limit=1
gh run watch {run-id}

# 2. 等待 Integration Tests 触发并监控
sleep 10
gh run list --workflow="Integration Tests" --limit=1
gh run watch {integration-run-id}

# 3. 等待 E2E Tests 触发并监控
sleep 10
gh run list --workflow="E2E Tests" --limit=1
gh run watch {e2e-run-id}
```

**决策点**：
- 所有工作流成功 → Step 12
- 任一工作流失败 → Step 11

**完整检查命令**：
```bash
# 检查所有工作流状态
gh pr view {pr-number} --json statusCheckRollup
```

### Step 11: 错误修复循环

```
1. gh run view {run-id} --log-failed
2. 分析错误 (参考 references/error_reference.md)
3. 定位问题代码
4. 应用修复
5. 本地验证
6. 重新提交
7. 返回 Step 10
```

**终止条件**：成功 或 重试5次 → 请求人工介入

### Step 12: 文档更新

追加更新：
- `references/issue_tracking.md`
- `references/workflow_history.md`

### Step 13: 创建PR

```bash
gh pr create --title "fix(auth): resolve login 401 error" --body "..."
```

### Step 14: 清理

```
1. 更新 workflow_state.json: stage="completed"
2. 更新统计信息
3. 输出PR链接给用户
```

## 失败处理原则

| 失败类型 | 处理方式 |
|----------|----------|
| Git操作失败 | 自动重试1次，失败后询问用户 |
| 测试失败 | 分析日志，修复后重新验证 |
| CI失败 | 最多重试5次，超过后请求人工介入 |
| 无法定位代码 | 询问用户提供更多信息 |

**核心原则**: 遇到无法自动解决的问题时，立即询问用户，不猜测

## 人工介入条件

| 条件 | 原因 |
|------|------|
| 重试5次仍失败 | 可能存在AI无法解决的问题 |
| 无法定位代码 | 需要用户提供更多信息 |
| 影响>5个文件 | 影响范围过大需确认 |
| 数据库迁移 | 数据安全风险 |
| 安全敏感操作 | 需要权限确认 |
| Git冲突 | 需要用户决定保留哪个版本 |

## Reference Files

### 参考文档 (references/)

| 文件 | 用途 |
|------|------|
| `decision_guide.md` | 决策指导：类型判断、优先级评估、测试策略 |
| `error_reference.md` | 错误参考：常见错误及处理方式 |
| `design.md` | 完整设计方案和技术架构 |
| `quick_reference.md` | 命令速查表、分支规范、提交规范 |
| `state_fields.md` | 状态字段说明：workflow_state.json各字段含义 |
| `issue_tracking.md` | 问题跟踪文档（运行时更新） |
| `workflow_history.md` | 工作流历史记录（运行时更新） |
| `test_guide.md` | TDD测试指南：Red-Green-Refactor、测试模板、覆盖率要求 |

### 模板文件 (assets/templates/)

| 文件 | 用途 |
|------|------|
| `plan_template.md` | 计划文档模板 |
| `tasks_template.md` | 任务清单模板 |
| `pr_template.md` | PR描述模板 |
| `bug_report.md` | Bug报告模板 |
| `feature_request.md` | 功能请求模板 |

### 状态管理 (scripts/)

| 文件 | 用途 |
|------|------|
| `workflow_state.json` | 工作流状态缓存，记录当前进度和统计 |

## Error Handling

详见 `references/error_reference.md`
