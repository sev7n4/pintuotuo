---
name: "mydev-github-workflow"
description: "Automated development workflow integrated with GitHub CI/CD. Invoke when user reports a bug, requests a feature, describes a code change requirement, or provides structured issue input with type/title/description/priority/scope fields."
---

# MyDev GitHub Workflow

## 触发条件

当用户满足以下任一条件时**必须**触发：

1. 报告Bug："有个bug"、"登录失败"、"返回401错误"
2. 请求新功能："需要添加"、"实现一个"、"新增功能"
3. 代码改进："优化"、"重构"、"改进"
4. 结构化输入：包含类型、标题、描述等字段

## 资源引用指南

**按需加载顺序**：

| 时机 | 加载资源 | 用途 |
|------|----------|------|
| 问题解析后 | `assets/templates/plan_template.md` | 生成计划文档 |
| 问题解析后 | `assets/templates/tasks_template.md` | 生成任务清单 |
| 需要详细设计参考 | `references/design.md` | 查看完整设计方案 |
| 需要命令参考 | `references/quick_reference.md` | 查看测试命令等 |
| 需要决策帮助 | `references/decision_guide.md` | 查看决策指导 |
| 遇到常见错误 | `references/error_reference.md` | 查看错误处理 |

## 核心工作流程

```
问题输入 → 解析 → 计划 → 分支 → 分析 → 实现 → 测试 → 验证 → 提交 → CI监控 → (失败循环) → 总结
```

### 步骤1：问题解析

**目标**：解析用户输入，生成结构化对象

**执行**：
1. 分析用户输入内容
2. 判断问题类型：
   - 包含"bug"、"错误"、"失败" → `bug`
   - 包含"添加"、"实现"、"新增" → `feature`
   - 包含"优化"、"重构"、"改进" → `enhancement`
3. 评估优先级：
   - 影响核心功能、安全 → `high`
   - 影响次要功能 → `medium`
   - 优化改进 → `low`
4. 确定影响范围：
   - 涉及API、数据库 → `backend`
   - 涉及UI、交互 → `frontend`
   - 两者都有 → `both`
5. 生成Issue ID：`ISSUE-{三位序号}`

**状态更新**：更新 `scripts/workflow_state.json`：
```json
{
  "currentIssue": { "id": "ISSUE-001", "type": "bug", ... },
  "workflowState": { "stage": "parsing", ... }
}
```

### 步骤2：计划生成

**目标**：生成计划文档和任务清单

**执行**：
1. 读取模板：
   - `assets/templates/plan_template.md`
   - `assets/templates/tasks_template.md`
2. 填充模板内容
3. 生成文件：
   - `assets/plans/{YYYY-MM-DD}_issue_{id}_plan.md`
   - `assets/tasks/{YYYY-MM-DD}_issue_{id}_tasks.md`

**决策点**：
- 如果问题复杂（涉及多个模块），先加载 `references/design.md` 参考
- 如果不确定如何填写，参考 `references/decision_guide.md`

### 步骤3：分支创建

**目标**：创建符合规范的Git分支

**执行**：
```bash
git checkout main && git pull origin main
git checkout -b {type}/issue-{id}-{desc}
git push -u origin {branch}
```

**分支命名**：
| 类型 | 格式 |
|------|------|
| bug | `bugfix/issue-{id}-{desc}` |
| feature | `feature/issue-{id}-{desc}` |
| enhancement | `enhancement/issue-{id}-{desc}` |

**状态更新**：`workflowState.stage = "branching"`

### 步骤4：代码分析

**目标**：定位需要修改的代码位置

**执行**：
1. 使用 SearchCodebase 进行语义搜索
2. 使用 Grep 进行关键词搜索
3. 使用 Read 读取相关文件
4. 分析代码依赖关系

**决策点**：
- 如果找不到相关代码，询问用户更多信息
- 如果影响范围过大，评估是否需要拆分任务

**状态更新**：`workflowState.stage = "analyzing"`

### 步骤5：代码实现

**目标**：实现代码修改

**执行**：
- SearchReplace: 精确修改现有文件
- Write: 创建新文件

**原则**：
- 最小化修改范围
- 保持代码风格一致
- 添加必要的错误处理

**状态更新**：`workflowState.stage = "implementing"`

### 步骤6：测试编写

**目标**：编写完整的测试用例

**执行**：
| 类型 | 位置 | 覆盖率要求 |
|------|------|-----------|
| 单元测试 | `backend/{module}_test.go` | ≥85% |
| 集成测试 | `backend/{module}_integration_test.go` | 核心流程 |
| E2E测试 | `frontend/e2e/{feature}.spec.ts` | 用户场景 |

**决策点**：
- Bug修复：至少编写一个回归测试
- 新功能：编写完整测试套件

**状态更新**：`workflowState.stage = "testing"`

### 步骤7：本地验证

**目标**：确保本地测试通过

**执行**：
```bash
# 后端
cd backend && go test -v -race -coverprofile=coverage.out ./...

# 前端
cd frontend && npm test -- --coverage --watchAll=false
cd frontend && npm run test:e2e
```

**决策点**：
- 测试失败 → 分析日志，修复代码或测试，重新验证
- 覆盖率不达标 → 补充测试用例

**状态更新**：`workflowState.stage = "verifying"`

### 步骤8：代码提交

**目标**：提交代码并推送

**执行**：
```bash
git add .
git commit -m "<type>(<scope>): <subject>

<body>

Closes #ISSUE-{id}"
git push origin {branch}
```

**提交类型**：`feat` | `fix` | `docs` | `style` | `refactor` | `test` | `chore`

**状态更新**：`workflowState.stage = "committing"`

### 步骤9：CI监控

**目标**：监控GitHub工作流执行状态

**执行**：
```bash
gh run list --branch={branch} --limit=1
gh run watch {run-id}
```

**工作流链**：CI/CD → Integration Tests → E2E Tests

**决策点**：
- 成功 → 进入步骤11
- 失败 → 进入步骤10

**状态更新**：`workflowState.stage = "ci_monitoring"`

### 步骤10：错误修复循环

**目标**：分析失败原因并修复

**执行**：
1. 获取失败日志：`gh run view {run-id} --log-failed`
2. 分析错误原因（参考 `references/error_reference.md`）
3. 定位问题代码
4. 应用修复
5. 本地验证
6. 重新提交
7. 返回步骤9

**终止条件**：
- 工作流成功
- 达到最大重试次数（5次）→ 请求人工介入

**状态更新**：`workflowState.stage = "fixing"`

### 步骤11：文档更新

**目标**：更新跟踪文档

**执行**：
1. 追加更新 `references/issue_tracking.md`
2. 追加更新 `references/workflow_history.md`

**状态更新**：`workflowState.stage = "documenting"`

### 步骤12：创建PR

**目标**：创建Pull Request

**执行**：
```bash
gh pr create --title "{title}" --body-file {body}
```

**状态更新**：`workflowState.stage = "completed"`

## 状态管理

**文件**：`scripts/workflow_state.json`

**更新时机**：每个步骤完成后更新

**字段说明**：
```json
{
  "currentIssue": {        // 当前处理的问题
    "id": "ISSUE-001",
    "type": "bug",
    "title": "...",
    "branch": "bugfix/issue-001-..."
  },
  "workflowState": {       // 工作流状态
    "stage": "analyzing",  // 当前阶段
    "startedAt": "2026-03-21T10:00:00Z",
    "retryCount": 0        // 重试次数
  },
  "statistics": { ... }    // 统计信息
}
```

## 错误处理

| 错误类型 | 处理方式 | 参考 |
|----------|----------|------|
| 编译错误 | 分析语法，修复代码 | `error_reference.md` |
| 测试失败 | 分析日志，修复代码/测试 | `error_reference.md` |
| Lint错误 | 按规范修复代码风格 | `quick_reference.md` |
| CI失败 | 获取日志，定位修复 | `error_reference.md` |

## 人工介入条件

以下情况需要请求用户确认：
1. 重试次数达到5次仍失败
2. 无法定位问题代码
3. 影响范围评估过大（>5个文件）
4. 涉及数据库迁移
5. 涉及安全敏感操作

## 文件结构

```
├── SKILL.md              # 本文件（第二层）
├── references/           # 参考文档（第三层，按需加载）
│   ├── design.md
│   ├── decision_guide.md
│   ├── error_reference.md
│   ├── issue_tracking.md
│   ├── quick_reference.md
│   └── workflow_history.md
├── assets/
│   ├── templates/        # 模板资源
│   ├── plans/            # 运行时生成
│   └── tasks/            # 运行时生成
└── scripts/
    └── workflow_state.json
```
