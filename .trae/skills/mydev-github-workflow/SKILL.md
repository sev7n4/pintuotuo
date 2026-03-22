---
name: "mydev-github-workflow"
description: |
  自动化开发工作流，监控完整CI链路(CI/CD→Integration→E2E)。
  触发条件：(1)报告bug ("登录失败","返回401","有个错误") (2)请求功能 ("添加功能","实现一个","新增") (3)代码改进 ("优化","重构","改进")
---

# MyDev GitHub Workflow

## ⚠️ 硬约束 (HARD CONSTRAINTS)

```
✅ Phase 1: PR验证阶段 (PR分支)
   └─ 开发 → 测试 → CI链路验证 → 通过/重试

✅ Phase 2: 合并部署阶段
   └─ 合并 → 部署

[1] 禁止忽略 current_fix_cases 失败！必须修复后才能合并！
[2] 禁止跳过状态跟踪写入！每个步骤必须更新状态文件！
[3] 禁止跳过本地验证！必须通过测试+类型检查才能提交！
[4] 禁止跳过CI监控！必须完成 CI/CD → Integration → E2E 完整链路！
[5] 禁止在CI验证通过前合并PR！必须满足合并条件！
[6] 如果涉及代码逻辑修改，禁止跳过 TDD 流程（Step 5-7）！
```

**重试限制**: 最多 5 次迭代，超过后请求人工介入

---

## 工作流链

```
Step 0: 初始化检查
   ↓
Step 1: 问题解析 ⚠️ [约束2]
   ↓
Step 2-3: 分支创建
   ↓
Step 4: 代码分析
   ↓
Step 5-7: TDD流程 ⚠️ [约束6]
   ↓
Step 8: 本地验证 ⚠️ [约束3]
   ↓
Step 9: git push ──────────────────┐
   ↓                               │ Push 触发
Step 10: CI监控 ⚠️ [约束4] ── CI/CD + Integration
   ├─ 失败 → Step 13 → Step 4/6    │
   └─ 通过                         │
   ↓                               │
Step 11: 创建 PR ──────────────────┤
   ↓                               │ PR 触发
Step 12: E2E监控 ⚠️ [约束1] ── current_fix_cases
   ├─ 失败 → Step 13 → Step 4/6    │
   └─ 通过                         │
   ↓                               │
Step 13: 错误修复 ⚠️ [约束2]        │
   ↓                               │
Step 14: 合并 PR ⚠️ [约束5]         │
   ↓                               │
Step 15: 部署监控 (main分支)        │
   ↓                               │
Step 16: 清理输出                  │
```

**触发逻辑**：
- Push 触发: CI/CD(全量) + Integration(全量)
- PR 触发: E2E(current_fix_cases)

**合并条件**：
- CI/CD Pipeline: 全量通过 ✅
- Integration Tests: 全量通过 ✅
- E2E Tests: current_fix_cases 通过 ✅ (其他用例失败可忽略)

---

## Step 0: 初始化检查

```bash
git status --porcelain
```

**检查项**：
- 工作区是否干净
- 是否在 main 分支

**状态写入**: 初始化 `scripts/workflow_state.json`

---

## Step 1: 问题解析

**输入**: 用户描述的问题/需求

**输出**: 
- `current_fix_cases` - 本次修复的测试用例ID列表
- 问题类型判断

**状态写入**: `workflow_state.json.current_fix_cases`

---

## Step 2-3: 分支创建

```bash
git checkout main
git pull origin main
git checkout -b {type}/issue-{id}
```

**分支命名**: `{type}/issue-{id}`
- type: bugfix / feature / enhancement
- id: ISSUE-XXX

---

## Step 4: 代码分析

**工具链**: SearchCodebase → Grep → Read

**分析内容**:
- 定位问题代码
- 理解上下文
- 确定修改范围

---

## Step 5-7: TDD 流程

**详细指南**: `references/test_lifecycle.md`

| 步骤 | TDD阶段 | 状态写入 |
|------|---------|----------|
| Step 5 | Red - 设计失败测试 | `test_cases_state.json` |
| Step 6 | Green - 最小实现 | `test_cases_state.json` |
| Step 7 | Refactor - 重构归档 | `test_cases_state.json` |

**模版参考**: `references/test_case_templates.md`

---

## Step 9: 代码推送

```bash
git add .
git commit -m "{type}: {description}"
git push -u origin {branch-name}
```

**触发**: Push 到远程分支触发 CI/CD + Integration

**状态写入**: `workflow_state.json.ci_status.cicd = running`

---

## Step 10: CI监控 (Push触发)

**监控顺序**: CI/CD → Integration

| 阶段 | 触发方式 | 验证要求 | 失败处理 |
|------|----------|----------|----------|
| 10.1 CI/CD | Push | 全量通过 | Step 13 → Step 6 |
| 10.2 Integration | Push | 全量通过 | Step 13 → Step 4 |

**详细脚本**: `references/monitor_scripts.md`

---

## Step 11: 创建PR

```bash
gh pr create --title "{title}" --body "{description}"
```

**触发**: PR 创建触发 E2E 测试

**状态写入**: `workflow_state.json.pr_number`

---

## Step 12: E2E监控 (PR触发)

**触发方式**: PR 创建/更新

| 阶段 | 验证要求 | 失败处理 |
|------|----------|----------|
| E2E ⚠️ [约束1] | current_fix_cases 通过 | Step 13 → Step 4/6 |

**注意**: 只监控 `current_fix_cases`，其他用例失败可忽略

**详细脚本**: `references/monitor_scripts.md`

---

## Step 13: 错误修复

```bash
gh run view {run-id} --log-failed
```

| 错误类型 | 返回步骤 |
|----------|----------|
| 代码错误（编译/Lint/单元测试） | Step 6 |
| 需求理解错误（集成/E2E设计问题） | Step 4 |
| 环境问题 | 重试当前 Step |

**详细判断逻辑**: `references/error_reference.md`

---

## Step 14: 合并条件

- CI/CD Pipeline: 全量通过 ✅
- Integration Tests: 全量通过 ✅
- E2E Tests: current_fix_cases 通过 ✅

```bash
gh pr merge {pr-number} --merge --delete-branch
```

**状态写入**: `workflow_state.json.merged = true`

---

## Step 15: 部署监控

**触发**: PR 合并后自动触发

```bash
gh run list --workflow="deploy.yml" --limit 1
```

**状态写入**: `workflow_state.json.ci_status.deploy = running`

---

## Step 16: 输出要求

```
✅ 工作流完成 | PR: #{pr-number} | 分支: {branch} → main
```
