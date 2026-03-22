# 工作流状态字段说明

> 本文档说明 `scripts/workflow_state.json` 中各字段的含义和更新时机。

## 字段说明

### 顶层字段

| 字段 | 类型 | 说明 | 更新时机 |
|------|------|------|----------|
| `version` | string | 状态文件版本 | 仅升级时更新 |
| `lastUpdated` | string | 最后更新时间（ISO 8601） | 每次更新时 |
| `currentIssue` | object/null | 当前处理的问题 | 开始/结束时 |
| `activeBranch` | string/null | 当前工作分支 | 分支创建/删除时 |

### currentIssue 对象

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | string | Issue ID，如 "ISSUE-001" |
| `type` | string | 问题类型：bug/feature/enhancement |
| `title` | string | 问题标题 |
| `branch` | string | 分支名称 |

### workflowState 对象

| 字段 | 类型 | 说明 | 可能值 |
|------|------|------|--------|
| `stage` | string | 当前阶段 | 见下方阶段列表 |
| `issueId` | string/null | 关联的Issue ID | ISSUE-XXX |
| `branchName` | string/null | 当前分支名 | bugfix/feature/... |
| `planFile` | string/null | 计划文件路径 | assets/plans/... |
| `tasksFile` | string/null | 任务文件路径 | assets/tasks/... |
| `startedAt` | string/null | 开始时间 | ISO 8601 |
| `lastCheckpoint` | string/null | 最后检查点时间 | ISO 8601 |
| `retryCount` | number | 重试次数 | 0-5 |

### stage 可能值

| 值 | 说明 | 对应步骤 |
|----|------|----------|
| `idle` | 空闲状态 | 无任务时 |
| `parsing` | 问题解析中 | Step 1 |
| `planning` | 计划生成中 | Step 2 |
| `branching` | 分支创建中 | Step 3 |
| `analyzing` | 代码分析中 | Step 4 |
| `implementing` | 代码实现中 | Step 5-6 |
| `verifying` | 本地验证中 | Step 7-8 |
| `committing` | 代码提交中 | Step 8 |
| `ci_monitoring` | CI监控中 | Step 9 |
| `fixing` | 错误修复中 | Step 10 |
| `documenting` | 文档更新中 | Step 11 |
| `pr_created` | PR已创建 | Step 12 |
| `completed` | 已完成 | Step 13 |

### statistics 对象

| 字段 | 类型 | 说明 |
|------|------|------|
| `totalIssues` | number | 总问题数 |
| `resolvedIssues` | number | 已解决问题数 |
| `totalWorkflows` | number | 工作流运行次数 |
| `successfulWorkflows` | number | 成功的工作流次数 |
| `failedWorkflows` | number | 失败的工作流次数 |

## 更新示例

### 开始新任务

```json
{
  "currentIssue": {
    "id": "ISSUE-001",
    "type": "bug",
    "title": "登录返回401错误",
    "branch": "bugfix/issue-001-login-401"
  },
  "activeBranch": "bugfix/issue-001-login-401",
  "workflowState": {
    "stage": "parsing",
    "issueId": "ISSUE-001",
    "branchName": null,
    "planFile": null,
    "tasksFile": null,
    "startedAt": "2026-03-21T10:00:00Z",
    "lastCheckpoint": "2026-03-21T10:00:00Z",
    "retryCount": 0
  }
}
```

### 分支创建后

```json
{
  "workflowState": {
    "stage": "branching",
    "branchName": "bugfix/issue-001-login-401",
    "planFile": "assets/plans/2026-03-21_issue_001_plan.md",
    "tasksFile": "assets/tasks/2026-03-21_issue_001_tasks.md"
  }
}
```

### CI失败后

```json
{
  "workflowState": {
    "stage": "fixing",
    "retryCount": 1
  }
}
```

### 任务完成后

```json
{
  "currentIssue": null,
  "activeBranch": null,
  "workflowState": {
    "stage": "completed",
    "issueId": null,
    "branchName": null,
    "planFile": null,
    "tasksFile": null,
    "startedAt": null,
    "lastCheckpoint": null,
    "retryCount": 0
  },
  "statistics": {
    "totalIssues": 1,
    "resolvedIssues": 1,
    "totalWorkflows": 1,
    "successfulWorkflows": 1,
    "failedWorkflows": 0
  }
}
```

---

## 简化版状态跟踪

> 用于 SKILL.md 核心流程的简化状态跟踪

### 状态文件结构

```json
{
  "current_step": 0,
  "current_fix_cases": ["PROD-002", "SET-002"],
  "retry_count": 0,
  "pr_number": null,
  "merged": false,
  "ci_status": {
    "cicd": "pending",
    "integration": "pending",
    "e2e": "pending"
  }
}
```

### 写入时机

| 步骤 | 写入内容 | 说明 |
|------|----------|------|
| Step 0 | 初始化状态文件 | 重置所有字段 |
| Step 1 | `current_fix_cases` | 本次修复的测试用例 |
| Step 2-3 | `current_step` | 当前步骤 |
| Step 10.1 | `ci_status.cicd` | CI/CD 状态 |
| Step 10.2 | `ci_status.integration` | Integration 状态 |
| Step 10.3 | `ci_status.e2e` | E2E 状态 |
| Step 11 | `retry_count` | 重试次数 |
| Step 12 | `pr_number` | PR 编号 |
| Step 13 | `merged: true` | 合并状态 |

---

## 测试用例状态跟踪

> 用于 `scripts/test_cases_state.json` 的字段说明

### 状态文件结构

```json
{
  "version": "1.0",
  "lastUpdated": "2026-03-22T10:00:00Z",
  "currentIssue": {
    "id": "ISSUE-001",
    "type": "bug",
    "module": "product",
    "branch": "bugfix/issue-001-product-price"
  },
  "testCases": {
    "unit": [],
    "integration": [],
    "e2e": []
  },
  "statistics": {
    "total": 0,
    "passed": 0,
    "failed": 0,
    "pending": 0
  }
}
```

### currentIssue 字段

| 字段 | 类型 | 说明 | 来源 |
|------|------|------|------|
| `id` | string/null | Issue ID | 从 workflow_state.json 获取 |
| `type` | string/null | 问题类型 | bug/feature/enhancement |
| `module` | string/null | 所属模块 | 代码分析推断 |
| `branch` | string/null | 分支名称 | 从 workflow_state.json 获取 |

### testCase 对象

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | string | 用例ID，如 UT-PROD-001 |
| `name` | string | 测试名称 |
| `feature` | string | 所属功能 |
| `status` | string | 状态：draft/ready/running/passed/failed/archived |
| `designedAt` | string | 设计时间 |
| `testedAt` | string/null | 测试时间 |
| `file` | string | 测试文件路径 |
| `archivedTo` | string/null | 归档文档路径（仅E2E） |

### 写入时机

| 步骤 | 写入内容 | 说明 |
|------|----------|------|
| Step 5.1 | `currentIssue` | 从 workflow_state.json 同步 |
| Step 5.3 | `testCases.*.status=draft` | 测试用例设计 |
| Step 5.4 | `testCases.*.status=ready` | 失败测试编写完成 |
| Step 6.2 | `testCases.*.status=running` | 测试执行中 |
| Step 6.3 | `testCases.*.status=passed/failed` | 测试结果 |
| Step 7.4 | `testCases.*.status=archived` | 归档完成 |

### 与 workflow_state.json 的关系

```
workflow_state.json          test_cases_state.json
├─ current_fix_cases    →    用于确定测试范围
├─ pr_number            ←    统计信息汇总
└─ ci_status.e2e        ←    E2E用例状态验证
```
