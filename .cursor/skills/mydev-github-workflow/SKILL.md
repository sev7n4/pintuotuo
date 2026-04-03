---
name: mydev-github-workflow
description: >-
  Automates GitHub development workflow with full CI monitoring (CI/CD → Integration → E2E),
  branch lifecycle, TDD steps, state file updates, PR merge conditions, and deploy checks.
  Use when the user reports bugs (e.g. login failures, 401), requests features or changes,
  asks for refactors or optimizations, or wants to run or continue this end-to-end PR/CI workflow.
---

# MyDev GitHub Workflow

本 skill 位于项目根目录 `.cursor/skills/mydev-github-workflow/`。文中相对路径 `references/`、`scripts/`、`assets/` 均相对于该目录。

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
[7] 禁止直接 push / 直接合并代码到 main！变更只能从功能分支经 PR 审核后合入（与仓库 `.cursorrules` 一致）！
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

**命名规范**: `references/00_00_naming_convention.md`

```bash
git status --porcelain
```

**检查项**：
- 工作区是否干净
- 是否在 main 分支

**状态写入**: 初始化 `scripts/00_01_workflow_state.json`

**详细字段**: `references/00_01_state_fields.md`

---

## Step 1: 问题解析

**输入**: 用户描述的问题/需求

**输出**: 
- `current_fix_cases` - 本次修复的测试用例ID列表
- 问题类型判断

**状态写入**: `scripts/00_01_workflow_state.json.current_fix_cases`

**详细规则**: `references/01_01_issue_parsing.md`

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

**详细规范**: `references/03_01_branch_naming.md`

---

## Step 4: 代码分析

**工具链**: SearchCodebase → Grep → Read

**分析内容**:
- 定位问题代码
- 理解上下文
- 确定修改范围

**分析输出**:
- 问题定位: `{文件路径}:{行号}`
- 影响范围: `{影响的功能列表}`
- 修改范围: `{需要修改的文件列表}`

**详细指南**: `references/04_01_code_analysis.md`

---

## Step 5-7: TDD 流程

| 步骤 | TDD阶段 | 状态写入 |
|------|---------|----------|
| Step 5 | Red - 设计失败测试 | `scripts/05_01_test_cases_state.json` |
| Step 6 | Green - 最小实现 | `scripts/05_01_test_cases_state.json` |
| Step 7 | Refactor - 重构归档 | `scripts/05_01_test_cases_state.json` |

**Step 6 Green 阶段定义**:
- 编写使测试通过的最小代码
- 不做优化或重构
- 只关注让测试从红变绿

**详细指南**: `references/05_01_test_lifecycle.md`

**模版参考**: `references/05_02_test_case_templates.md`

**跳过条件** (满足任一即可跳过 TDD):
- 纯测试添加: 只新增测试文件，不修改业务代码
- 文档更新: 只修改 README、注释、文档
- 配置变更: 只修改配置文件，不涉及代码逻辑

**必须执行 TDD**:
- 修改业务逻辑代码
- 修改 API 接口
- 修改数据模型

---

## Step 8: 本地验证 ⚠️ [约束3]

**本仓库（pintuotuo）**：若存在根目录 `Makefile`，优先使用统一入口：

```bash
make test
```

其他仓库或需分项执行时：

```bash
# 前端
cd frontend && npm run type-check && npm test

# 后端
cd backend && go test -v ./...
```

**验证项**：
- 类型检查通过
- 单元测试通过

**状态写入**: `scripts/00_01_workflow_state.json.local_validation = passed`

---

## Step 9: Push + 创建 PR

```bash
# 提交代码 (格式见 references/09_01_pr_template.md)
git add .
git commit -m "{type}: {description}"

# 推送到远程
git push -u origin {branch-name}

# 创建 PR（目标分支须为 main，勿在本地直推 main）
gh pr create --base main --title "{title}" --body "{description}"
```

**触发**: PR 创建触发完整 CI 链路 (CI/CD → Integration → E2E)

**PR 模板**: `references/09_01_pr_template.md`

**状态写入**: 
- `scripts/00_01_workflow_state.json.pr_number`
- `scripts/00_01_workflow_state.json.ci_status.cicd = running`

---

## Step 10: CI监控 (PR触发) ⚠️ [约束4]

**监控顺序**: CI/CD → Integration → E2E

| 阶段 | 触发方式 | 验证要求 | 失败处理 |
|------|----------|----------|----------|
| 10.1 CI/CD | PR | 全量通过 | Step 11 → Step 6 |
| 10.2 Integration | PR (workflow_run) | 全量通过 | Step 11 → Step 4 |
| 10.3 E2E | PR (workflow_run) | current_fix_cases 通过 | Step 11 → 判断标准见 11_01_error_reference.md |

**详细脚本**: `references/10_01_monitor_scripts.md`

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
Step 11.3: 更新状态文件
Step 11.4: 判断是否超过重试限制
   ├─ 是 → 请求人工介入
   └─ 否 → 返回对应步骤修复
Step 11.5: 修复后重新 commit + push
```

**状态更新字段**:
- `retry_count`: 重试次数 +1
- `error_history`: 追加 `{step, error_type, message, timestamp}`
- `current_fix_cases`: 保留当前失败用例

**重试策略**:
- 第1-2次: 快速修复 + 重试 (间隔 30s)
- 第3-4次: 深入分析原因 (间隔 5min)
- 第5次: 请求人工介入

**详细判断逻辑**: `references/11_01_error_reference.md`

---

## Step 12: 合并 PR ⚠️ [约束5]

**合并条件**：
- CI/CD Pipeline: 全量通过 ✅
- Integration Tests: 全量通过 ✅
- E2E Tests: current_fix_cases 通过 ✅

```bash
# 与仓库规范一致：GitHub 上使用 Squash & Merge（单条主线历史）
gh pr merge {pr-number} --squash --delete-branch
```

**说明**: 勿使用 `--merge`（会产生 merge commit）；本仓库要求 **Squash & Merge** 合入 `main`（见根目录 `.cursorrules` / `CLAUDE.md`）。

**状态写入**: `scripts/00_01_workflow_state.json.merged = true`

---

## Step 13: 部署监控

**触发**: PR 合并后自动触发 (workflow_run: E2E Tests completed)

```bash
gh run list --workflow="deploy-tencent.yml" --limit 1
```

**详细指南**: `references/13_01_deploy_guide.md`

**状态写入**: `scripts/00_01_workflow_state.json.ci_status.deploy = running`

---

## Step 14: 清理输出

**清理范围**:
- 删除本地 feature 分支
- 清理临时文件 (node_modules, build, .cache)

**保留内容**:
- 工作流状态文件 (`scripts/00_01_workflow_state.json`)
- 测试用例归档 (`assets/test_cases/`)
- PR 记录 (GitHub)

```bash
git checkout main
git branch -d {branch-name}
```

**输出要求**:
```
✅ 工作流完成 | PR: #{pr-number} | 分支: {branch} → main
```
