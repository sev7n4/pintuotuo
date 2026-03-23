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
Step 4: 代码分析 ←─────────────────┐
   ↓                               │
Step 5-7: TDD流程 ⚠️ [约束6] ←─┐    │
   ↓                          │    │
Step 8: 本地验证 ⚠️ [约束3]    │    │
   ↓                          │    │
Step 9: Push + 创建 PR ───────┼────┤
   ↓                          │    │ PR 触发
Step 10: CI监控 ⚠️ [约束4]    │    │
   ├─ CI/CD Pipeline          │    │
   ├─ Integration Tests       │    │
   ├─ E2E Tests               │    │
   │                          │    │
   ├─ 失败 → Step 11 ─────────┼────┘
   └─ 通过                    │
   ↓                          │
Step 12: 合并 PR ⚠️ [约束5]    │
   ↓                          │
Step 13: 部署监控 (main分支)   │
   ↓                          │
Step 14: 清理输出             │
                              │
Step 11: 错误修复 ─────────────┘
   ├─ 代码错误 → 返回 Step 6
   ├─ 需求错误 → 返回 Step 4
   └─ 环境问题 → 重试当前步骤
```

**触发逻辑**：
- PR 创建触发完整 CI 链路: CI/CD → Integration → E2E
- 注意: Push 到 feature 分支不会触发 CI，只有 PR 才会触发

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

**注意**: 如果是纯测试添加任务（不涉及代码逻辑修改），可跳过 TDD 流程直接进入 Step 8。

---

## Step 8: 本地验证 ⚠️ [约束3]

```bash
# 前端
cd frontend && npm run type-check && npm test

# 后端
cd backend && go test -v ./...
```

**验证项**：
- 类型检查通过
- 单元测试通过

**状态写入**: `workflow_state.json.local_validation = passed`

---

## Step 9: Push + 创建 PR

```bash
# 提交代码
git add .
git commit -m "{type}: {description}"

# 推送到远程
git push -u origin {branch-name}

# 创建 PR
gh pr create --title "{title}" --body "{description}"
```

**触发**: PR 创建触发完整 CI 链路 (CI/CD → Integration → E2E)

**状态写入**: 
- `workflow_state.json.pr_number`
- `workflow_state.json.ci_status.cicd = running`

---

## Step 10: CI监控 (PR触发) ⚠️ [约束4]

**监控顺序**: CI/CD → Integration → E2E

| 阶段 | 触发方式 | 验证要求 | 失败处理 |
|------|----------|----------|----------|
| 10.1 CI/CD | PR | 全量通过 | Step 11 → Step 6 |
| 10.2 Integration | PR (workflow_run) | 全量通过 | Step 11 → Step 4 |
| 10.3 E2E | PR (workflow_run) | current_fix_cases 通过 | Step 11 → Step 4/6 |

**详细脚本**: `references/monitor_scripts.md`

**监控命令**:
```bash
# 获取 PR 触发的 workflow runs
gh run list --branch={branch-name} --limit 5 --json databaseId,name,status,conclusion

# 查看失败日志
gh run view {run-id} --log-failed
```

---

## Step 11: 错误修复

```bash
gh run view {run-id} --log-failed
```

### 错误类型判断

| 错误类型 | 关键词 | 返回步骤 | 说明 |
|----------|--------|----------|------|
| **代码错误** | `undefined`, `type error`, `syntax error`, `Cannot find` | Step 6 | 实现层面问题 |
| **需求理解错误** | `assertion failed`, `wrong result`, E2E 业务流程错误 | Step 4 | 需求理解偏差 |
| **环境问题** | `timeout`, `ECONNREFUSED`, `service unavailable` | 重试 | 外部依赖问题 |

### 修复流程

```
Step 11.1: 获取失败日志
Step 11.2: 分析错误类型
Step 11.3: 更新状态文件 (记录错误、重试次数)
Step 11.4: 判断是否超过重试限制
   ├─ 是 → 请求人工介入
   └─ 否 → 返回对应步骤修复
Step 11.5: 修复后重新 commit + push
```

**详细判断逻辑**: `references/error_reference.md`

---

## Step 12: 合并 PR ⚠️ [约束5]

**合并条件**：
- CI/CD Pipeline: 全量通过 ✅
- Integration Tests: 全量通过 ✅
- E2E Tests: current_fix_cases 通过 ✅

```bash
gh pr merge {pr-number} --merge --delete-branch
```

**状态写入**: `workflow_state.json.merged = true`

---

## Step 13: 部署监控

**触发**: PR 合并后自动触发 (workflow_run: E2E Tests completed)

```bash
gh run list --workflow="deploy-tencent.yml" --limit 1
```

**状态写入**: `workflow_state.json.ci_status.deploy = running`

---

## Step 14: 输出要求

```
✅ 工作流完成 | PR: #{pr-number} | 分支: {branch} → main
```
