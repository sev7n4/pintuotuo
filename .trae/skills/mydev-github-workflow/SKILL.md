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
Step 10: CI监控 → 监控完整工作流链
Step 11: 错误修复 → 分析日志 → 修复 → 重试(最多5次)
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

### Step 10: CI监控

**工作流链**: CI/CD → Integration Tests → E2E Tests

```bash
# 检查所有工作流状态
gh pr view {pr-number} --json statusCheckRollup
```

**失败类型决策**:

| 失败类型 | 错误特征 | 下一步 |
|----------|----------|--------|
| **编译错误** | `undefined`, `type error`, `cannot find` | Step 11 → 修复代码 |
| **测试失败** | `FAIL`, `assertion failed`, `expected` | Step 11 → 修复测试/代码 |
| **Lint错误** | `errcheck`, `no-unused-vars`, `staticcheck` | Step 11 → 修复代码风格 |
| **安全漏洞** | `CVE`, `Critical`, `High` | Step 11 → 更新依赖 |
| **环境问题** | `timeout`, `out of memory`, `permission` | 重试1次 → 失败则人工介入 |

**成功?** → Step 12

### Step 11: 错误修复

```bash
# 获取失败日志
gh run view {run-id} --log-failed

# 分析错误位置
grep -i "error\|fail\|panic" logs.txt
```

**修复流程**:
```
1. 识别错误类型 (编译/测试/Lint/安全/环境)
2. 定位错误位置 (文件:行号)
3. 分析错误原因
4. 应用修复
5. 本地验证
6. 重新提交
7. 返回 Step 10
```

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
