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
用户输入
    ↓
问题解析 ──→ 判断类型
    ↓           ├─ bug → bugfix/issue-{id}-{desc}
    ↓           ├─ feature → feature/issue-{id}-{desc}
    ↓           └─ enhancement → enhancement/issue-{id}-{desc}
    ↓
计划生成 ──→ 读取模板 → 生成计划和任务文档
    ↓
分支创建 ──→ git checkout -b {branch}
    ↓
代码分析 ──→ SearchCodebase → Grep → Read
    ↓           └─ 找不到? → 询问用户
    ↓
代码实现 ──→ SearchReplace / Write
    ↓
测试编写 ──→ 单元测试 + 集成测试 + E2E测试
    ↓
本地验证 ──→ 运行测试
    ↓           └─ 失败? → 修复 → 重新验证
    ↓
代码提交 ──→ git commit -m "..."
    ↓
CI监控 ──→ gh run watch
    ↓           ├─ 成功 → 文档更新 → 创建PR
    ↓           └─ 失败 → 错误修复循环 (最多5次)
    ↓
完成
```

## 核心步骤

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

**决策点**：问题复杂? → 先加载 `references/decision_guide.md`

### Step 3: 分支创建

```bash
git checkout main && git pull origin main
git checkout -b bugfix/issue-001-login-401  # 示例
git push -u origin bugfix/issue-001-login-401
```

### Step 4: 代码分析

```
SearchCodebase: "登录认证逻辑"
Grep: "func.*Login|auth.*handler"
Read: backend/handlers/auth.go
```

**决策点**：找不到代码? → 询问用户提供更多信息

### Step 5: 代码实现

- SearchReplace: 精确修改现有文件
- Write: 创建新文件
- 原则：最小化修改，保持风格一致

### Step 6: 测试编写

| 类型 | 位置 | 覆盖率 |
|------|------|--------|
| 单元测试 | `backend/{module}_test.go` | ≥85% |
| 集成测试 | `backend/{module}_integration_test.go` | 核心流程 |
| E2E测试 | `frontend/e2e/{feature}.spec.ts` | 用户场景 |

### Step 7: 本地验证

```bash
# 后端
cd backend && go test -v -race -coverprofile=coverage.out ./...

# 前端  
cd frontend && npm test -- --coverage --watchAll=false
```

**决策点**：失败? → 分析日志 → 修复 → 重新验证

### Step 8: 代码提交

```bash
git add .
git commit -m "fix(auth): resolve login 401 error

- Fix jwtSecret initialization
- Add unit tests

Closes #ISSUE-001"
git push origin bugfix/issue-001-login-401
```

### Step 9: CI监控

```bash
gh run list --branch=bugfix/issue-001-login-401 --limit=1
gh run watch {run-id}
```

**决策点**：
- 成功 → Step 11
- 失败 → Step 10

### Step 10: 错误修复循环

```
1. gh run view {run-id} --log-failed
2. 分析错误 (参考 references/error_reference.md)
3. 定位问题代码
4. 应用修复
5. 本地验证
6. 重新提交
7. 返回 Step 9
```

**终止条件**：成功 或 重试5次 → 请求人工介入

### Step 11: 文档更新

追加更新：
- `references/issue_tracking.md`
- `references/workflow_history.md`

### Step 12: 创建PR

```bash
gh pr create --title "fix(auth): resolve login 401 error" --body "..."
```

## 状态管理

**文件**：`scripts/workflow_state.json`

每个步骤完成后更新 `workflowState.stage`，详见 `references/design.md`

## 人工介入条件

| 条件 | 原因 |
|------|------|
| 重试5次仍失败 | 可能存在AI无法解决的问题 |
| 无法定位代码 | 需要用户提供更多信息 |
| 影响>5个文件 | 影响范围过大需确认 |
| 数据库迁移 | 数据安全风险 |
| 安全敏感操作 | 需要权限确认 |

## Reference Files

| 文件 | 用途 |
|------|------|
| `references/decision_guide.md` | 决策指导：类型判断、优先级评估 |
| `references/error_reference.md` | 错误参考：常见错误及处理方式 |
| `references/design.md` | 完整设计方案 |
| `references/quick_reference.md` | 命令速查表 |

## Error Handling

| 错误类型 | 处理方式 |
|----------|----------|
| 编译错误 | 分析语法，修复代码 |
| 测试失败 | 分析日志，修复代码/测试 |
| CI失败 | 获取日志，定位修复 |

详见 `references/error_reference.md`
