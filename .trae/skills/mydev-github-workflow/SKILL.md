---
name: "mydev-github-workflow"
description: |
Automated development workflow for bug fixes, features, and code changes with GitHub CI/CD integration. 
Use when user: (1) reports a bug ("登录失败", "返回401", "有个错误"), (2) requests a feature ("添加功能", "实现一个", "新增"), (3) describes improvements ("优化", "重构", "改进"), or (4) provides structured issue input (type/title/description fields).
---

# MyDev GitHub Workflow

## Overview

自动化开发工作流：问题解析 → 计划生成 → 分支创建 → TDD开发 → 测试验证 → CI监控 → 创建PR。

## Quick Start

```
用户：登录功能返回401错误
→ 自动执行完整流程，直到CI通过并创建PR
```

## Workflow Decision Tree

```
Step 0: 初始化检查 → Git状态/未完成任务
Step 1: 问题解析 → 类型/优先级/范围/Issue ID
Step 2: 计划生成 → 读取模板 → 生成计划文档
Step 3: 分支创建 → git checkout -b {type}/issue-{id}
Step 4: 代码分析 → SearchCodebase → Grep → Read
Step 5: 测试设计 → TDD Red: 写失败测试
Step 6: 最小实现 → TDD Green: 写最少代码
Step 7: 重构优化 → TDD Refactor: 优化代码
Step 8: 本地验证 → 运行测试 + 覆盖率检查
Step 9: 代码提交 → git commit -m "..."
Step 10: 实时监控 → 每30秒分析日志 → 发现失败立即取消
   ├─ 失败 → Step 11 (分析错误 → 返回Step 6)
   └─ 成功 → Step 12
Step 11: 错误修复 → 分析日志 → 返回Step 6继续最小实现
Step 12: 文档更新 → 更新跟踪文档
Step 13: 创建PR → gh pr create
Step 14: 清理 → 输出PR链接
```

## 核心步骤

### Step 0: 初始化检查

```bash
# 检查当前分支是否干净
git status --porcelain  # 有输出? → 询问用户: 暂存/提交/忽略
```

### Step 1: 问题解析

| 判断 | 关键词 | 类型 |
|------|--------|------|
| Bug | "错误"、"失败"、"异常" | bugfix |
| 功能 | "添加"、"实现"、"新增" | feature |
| 改进 | "优化"、"重构"、"改进" | enhancement |

**范围判断**: API/DB → backend | UI → frontend | 两者 → both

### Step 2-3: 计划与分支

```bash
# 1. 切换到main并拉取最新
git checkout main && git pull origin main

# 2. 再次验证main分支干净
git status --porcelain  # 有输出? → 警告用户，不继续

# 3. 创建新分支
git checkout -b bugfix/issue-001-{description}
```

### Step 4: 代码分析

找不到代码? → 询问用户

### Step 5: 测试设计 (TDD - Red)

**测试策略**: backend → 单元+集成 | frontend → 单元+E2E | both → 全部

**详细指南**: `references/test_guide.md`

### Step 6-7: 实现与重构 (TDD - Green/Refactor)

```
Red: 写失败测试 → Green: 最小实现 → Refactor: 优化
```

### Step 8: 本地验证

```bash
cd backend && go test -v -race ./...
cd frontend && npm test -- --coverage
```

**覆盖率**: Backend ≥85%, Frontend ≥80%

**失败?** → 分析日志 → 修复 → 重新验证

### Step 9: 代码提交

```bash
git commit -m "fix(auth): resolve login 401 error

- Fix jwtSecret initialization
- Add unit tests

Closes #ISSUE-001"
```

### Step 10: CI实时监控

**工作流链**: CI/CD → Integration → E2E (顺序触发)

**核心原则**: 实时监控，快速失败，及时止损

```bash
# 实时获取工作流日志
gh run view {run-id} --log 2>&1 | grep -E "FAIL|Error|error|✘|✗"

# 取消正在运行的工作流
gh run cancel {run-id}
```

**实时监控流程**:
```
1. 启动 CI/CD Pipeline 监控
   ├─ 每30秒获取最新日志
   ├─ 检测到编译/Lint错误 → 取消工作流 → Step 11
   ├─ 检测到单元测试失败 → 取消工作流 → Step 11
   └─ CI/CD成功 → 启动 Integration 监控

2. 启动 Integration Tests 监控
   ├─ 每30秒获取最新日志
   ├─ 检测到API错误 → 取消工作流 → Step 11
   ├─ 检测到数据错误 → 取消工作流 → Step 11
   └─ Integration成功 → 启动 E2E 监控

3. 启动 E2E Tests 监控
   ├─ 每30秒获取最新日志
   ├─ 重点监控当前测试用例 (PROD-002, PROD-003等)
   ├─ 检测到当前用例失败 → 取消工作流 → Step 11
   ├─ 检测到其他用例失败 → 记录，继续监控
   └─ E2E成功 → Step 12
```

**实时日志分析命令**:
```bash
# 获取当前运行中的工作流ID
RUN_ID=$(gh run list --workflow="E2E Tests" --status=in_progress --json databaseId -q '.[0].databaseId')

# 实时获取日志并分析
gh run view $RUN_ID --log 2>&1 | grep -E "PROD-002|PROD-003|✘|FAIL"

# 检查特定测试用例状态
gh run view $RUN_ID --log 2>&1 | grep -A5 "PROD-002"
```

**错误类型决策**:

| 工作流 | 错误类型 | 错误特征 | 下一步 |
|--------|----------|----------|--------|
| CI/CD | 编译错误 | `undefined`, `type error` | 取消 → 修复代码 |
| CI/CD | Lint错误 | `errcheck`, `staticcheck` | 取消 → 修复风格 |
| CI/CD | 单元测试 | `FAIL`, `assertion` | 取消 → 修复测试 |
| Integration | API错误 | `connection refused`, `500` | 取消 → 修复API/DB |
| Integration | 数据错误 | `duplicate`, `foreign key` | 取消 → 修复数据逻辑 |
| E2E | 当前用例失败 | `PROD-xxx ✘` | 取消 → Step 11 |
| E2E | 其他用例失败 | `SETTLE-xxx ✘` | 记录，继续监控 |
| 任意 | 环境问题 | `permission`, `OOM` | 重试1次 → 人工介入 |

**成功?** → Step 12

### Step 11: 错误修复与快速迭代

```bash
# 获取失败日志
gh run view {run-id} --log-failed

# 分析错误位置
grep -i "error\|fail\|panic" logs.txt
```

**快速迭代流程**:
```
1. 取消当前工作流 (如果还在运行)
   gh run cancel {run-id}

2. 分析错误类型
   ├─ 代码错误 → Step 6 (最小实现)
   ├─ 测试错误 → Step 5 (测试设计) 或 Step 6 (最小实现)
   └─ 环境问题 → 重试1次

3. 应用修复 → 本地验证 → 重新提交

4. 返回 Step 10 (实时监控)
```

**错误类型与返回点**:

| 错误类型 | 返回步骤 | 原因 |
|----------|----------|------|
| 编译错误 | Step 6 | 代码语法/类型问题 |
| Lint错误 | Step 6 | 代码风格问题 |
| 单元测试失败 | Step 6 | 实现不满足测试 |
| 集成测试失败 | Step 6 | API/数据逻辑问题 |
| E2E当前用例失败 | Step 6 | 前端实现/选择器问题 |
| 测试设计问题 | Step 5 | 测试用例本身有问题 |

**重试限制**: 最多5次 → 超过后请求人工介入

### Step 12-14: 收尾

```
更新文档 → 创建PR → 输出PR链接
```

## 失败处理

| 场景 | 处理 |
|------|------|
| Git失败 | 重试1次 → 询问用户 |
| 测试失败 | 分析日志 → 修复 |
| CI失败 | 重试5次 → 人工介入 |
| 找不到代码 | 询问用户 |

**核心原则**: 无法自动解决 → 立即询问用户

## Reference Files

详见 `references/` 目录：
- `decision_guide.md` - 决策指导
- `test_guide.md` - 测试指南
- `error_reference.md` - 错误参考
- `quick_reference.md` - 命令速查
- `design.md` - 完整设计
