---
name: "mydev-github-workflow"
description: "Automated development workflow integrated with GitHub CI/CD. Invoke when user reports a bug, requests a feature, describes a code change requirement, or provides structured issue input with type/title/description/priority/scope fields."
---

# MyDev GitHub Workflow

## 触发条件

当用户满足以下任一条件时自动触发：

1. 报告Bug："有个bug"、"登录失败"、"返回401错误"
2. 请求新功能："需要添加"、"实现一个"、"新增功能"
3. 代码改进："优化"、"重构"、"改进"
4. 结构化输入：包含类型、标题、描述等字段

## 核心工作流程

```
问题输入 → 解析 → 计划 → 分支 → 分析 → 实现 → 测试 → 验证 → 提交 → CI监控 → (失败循环) → 总结
```

### 步骤1：问题解析

解析用户输入，生成结构化对象：
- id: ISSUE-{序号}
- type: bug | feature | enhancement
- title, description, priority, scope

### 步骤2：计划生成

使用模板生成：
- `assets/plans/issue_{id}_plan.md`
- `assets/tasks/issue_{id}_tasks.md`

### 步骤3：分支创建

命名规范：
- bugfix/issue-{id}-{desc}
- feature/issue-{id}-{desc}

```bash
git checkout main && git pull
git checkout -b {type}/issue-{id}-{desc}
git push -u origin {branch}
```

### 步骤4：代码分析

- SearchCodebase: 语义搜索
- Grep: 关键词搜索
- Read: 读取文件

### 步骤5：代码实现

- SearchReplace: 精确修改
- Write: 创建新文件
- 遵循项目规范，最小化修改

### 步骤6：测试编写

| 类型 | 位置 | 覆盖率要求 |
|------|------|-----------|
| 单元测试 | backend/{module}_test.go | ≥85% |
| 集成测试 | backend/{module}_integration_test.go | 核心流程 |
| E2E测试 | frontend/e2e/{feature}.spec.ts | 用户场景 |

### 步骤7：本地验证

```bash
cd backend && go test -v -race -coverprofile=coverage.out ./...
cd frontend && npm test -- --coverage --watchAll=false
cd frontend && npm run test:e2e
```

### 步骤8：代码提交

格式：`<type>(<scope>): <subject>`

```bash
git add . && git commit -m "fix(auth): resolve login 401 error

- Fix jwtSecret initialization
- Add unit tests

Closes #ISSUE-001"
git push
```

### 步骤9：CI监控

```bash
gh run list --branch={branch} --limit=1
gh run watch {run-id}
gh run view {run-id} --log-failed  # 失败时
```

### 步骤10：错误修复循环

失败时：
1. 获取失败日志
2. 分析错误原因
3. 定位问题代码
4. 应用修复
5. 本地验证
6. 重新提交
7. 返回步骤9

最大重试：5次

### 步骤11：文档更新

更新跟踪文档：
- references/issue_tracking.md
- references/workflow_history.md

### 步骤12：创建PR

```bash
gh pr create --title "{title}" --body-file {body}
```

## 错误处理

| 错误类型 | 处理方式 |
|----------|----------|
| 编译错误 | 分析语法，修复代码 |
| 测试失败 | 分析日志，修复代码/测试 |
| Lint错误 | 按规范修复代码风格 |
| CI失败 | 获取日志，定位修复 |

## 文件位置

所有文件在 `.trae/skills/mydev-github-workflow/` 目录下：

```
├── SKILL.md              # 本文件
├── references/           # 参考文档（按需加载）
│   ├── design.md
│   ├── issue_tracking.md
│   ├── workflow_history.md
│   └── quick_reference.md
├── assets/
│   ├── templates/        # 模板资源
│   │   ├── plan_template.md
│   │   ├── tasks_template.md
│   │   ├── pr_template.md
│   │   ├── bug_report.md
│   │   └── feature_request.md
│   ├── plans/            # 🆕 运行时生成
│   │   └── issue_{id}_plan.md
│   └── tasks/            # 🆕 运行时生成
│       └── issue_{id}_tasks.md
└── scripts/              # 状态管理
    └── workflow_state.json
```

## 使用示例

**Bug修复**：
```
用户：登录功能返回401错误
→ 自动执行完整流程直到CI通过
```

**新功能**：
```
用户：添加商品收藏功能
→ 自动执行开发流程，包括数据库设计、API实现、前端开发、测试
```
