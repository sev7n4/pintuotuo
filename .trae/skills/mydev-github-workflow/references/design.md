# 品多多项目 - GitHub集成自动化开发工作流设计方案

## 1. 工作流概述

### 1.1 设计目标

设计一套与GitHub工作流完美结合的自动化开发流程，实现从问题/需求输入到代码合并、测试验证的完整闭环机制。

### 1.2 文档存放位置

**重要**：所有工作流相关文件存放在 `.trae/skills/mydev-github-workflow/` 目录下：

```
.trae/skills/mydev-github-workflow/
├── SKILL.md                          # Skill定义文件（第二层）
├── references/                       # 参考文档（第三层）
│   ├── design.md                     # 本文档
│   ├── decision_guide.md             # 决策指导
│   ├── error_reference.md            # 错误参考
│   ├── issue_tracking.md             # 问题跟踪
│   ├── quick_reference.md            # 快速参考
│   └── workflow_history.md           # 工作流历史
├── assets/
│   ├── templates/                    # 模板目录
│   │   ├── bug_report.md
│   │   ├── feature_request.md
│   │   ├── plan_template.md
│   │   ├── tasks_template.md
│   │   └── pr_template.md
│   ├── plans/                        # 计划目录（运行时生成）
│   │   └── {YYYY-MM-DD}_issue_{id}_plan.md
│   └── tasks/                        # 任务目录（运行时生成）
│       └── {YYYY-MM-DD}_issue_{id}_tasks.md
└── scripts/
    └── workflow_state.json           # 工作流状态缓存
```

**隔离性保证**：
- ✅ 所有文档、模板、缓存文件均在 `.trae/` 目录下
- ✅ 不修改项目原有文件结构
- ✅ 工作流状态完全独立
- ✅ 可随时删除 `.trae/` 目录而不影响项目

### 1.3 技术实现路线

#### 1.3.1 核心技术栈

| 组件         | 技术选型                  | 说明                                  |
| ---------- | --------------------- | ----------------------------------- |
| AI引擎       | Trae IDE + GLM-5      | 智能分析、代码生成、问题诊断                      |
| 代码操作       | Trae内置工具              | Read/Write/SearchReplace/DeleteFile |
| 代码搜索       | SearchCodebase + Grep | 语义搜索和正则搜索                           |
| 命令执行       | RunCommand            | Git操作、测试运行、构建                       |
| GitHub交互   | GitHub CLI (gh)       | 分支管理、PR创建、工作流监控                     |
| GitHub API | REST API              | 工作流状态查询、日志获取                        |

#### 1.3.2 实现架构

```
┌─────────────────────────────────────────────────────────────┐
│                      用户交互层                              │
│  (用户输入问题/需求描述 → AI理解并解析)                       │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                      AI智能层                                │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │
│  │ 需求分析    │ │ 代码分析    │ │ 问题诊断    │           │
│  │ 计划制定    │ │ 方案设计    │ │ 错误修复    │           │
│  └─────────────┘ └─────────────┘ └─────────────┘           │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                      工具执行层                              │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │
│  │ Read/Write  │ │ RunCommand  │ │ SearchCode  │           │
│  │ 文件操作    │ │ 命令执行    │ │ 代码搜索    │           │
│  └─────────────┘ └─────────────┘ └─────────────┘           │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                      外部系统层                              │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │
│  │ Git/GitHub  │ │ GitHub API  │ │ 测试框架    │           │
│  │ 版本控制    │ │ 工作流监控  │ │ Go/Jest     │           │
│  └─────────────┘ └─────────────┘ └─────────────┘           │
└─────────────────────────────────────────────────────────────┘
```

#### 1.3.3 关键技术实现

**1. 分支管理（使用Git命令）**

```bash
# 创建分支
git checkout -b {type}/issue-{id}-{description}

# 推送分支
git push origin {branch-name}
```

**2. GitHub工作流监控（使用GitHub CLI + API）**

```bash
# 查看工作流运行状态
gh run list --branch={branch-name}

# 获取工作流运行详情
gh run view {run-id}

# 获取工作流日志
gh run view {run-id} --log

# 查询工作流Job状态
gh api repos/{owner}/{repo}/actions/runs/{run-id}/jobs
```

**3. 测试执行（使用RunCommand）**

```bash
# 后端测试
cd backend && go test -v -race -coverprofile=coverage.out ./...

# 前端测试
cd frontend && npm test -- --coverage --watchAll=false

# E2E测试
cd frontend && npm run test:e2e
```

**4. PR创建（使用GitHub CLI）**

```bash
# 创建PR
gh pr create --title "{title}" --body-file {body-file}

# 查看PR状态
gh pr view {pr-number}

# 合并PR
gh pr merge {pr-number} --squash
```

#### 1.3.4 工作流状态管理

使用JSON文件存储当前工作流状态：

```json
{
  "currentIssue": {
    "id": "ISSUE-001",
    "type": "bug",
    "title": "登录失败",
    "branch": "bugfix/issue-001-login"
  },
  "workflowState": {
    "stage": "ci_validation",
    "ciStatus": "running",
    "runId": "12345678"
  },
  "tasks": [
    {"id": "1", "name": "代码分析", "status": "completed"},
    {"id": "2", "name": "代码实现", "status": "completed"},
    {"id": "3", "name": "测试编写", "status": "completed"},
    {"id": "4", "name": "CI验证", "status": "in_progress"}
  ]
}
```

#### 1.3.5 错误处理机制

```
检测到错误
    ↓
解析错误类型
    ├── 编译错误 → 分析代码 → 修复语法/类型错误
    ├── 测试失败 → 分析测试日志 → 修复代码/测试
    ├── 工作流失败 → 获取失败日志 → 定位问题 → 修复
    └── 合并冲突 → 拉取最新代码 → 解决冲突 → 重新提交
    ↓
本地验证
    ↓
重新提交
    ↓
循环直到成功
```

### 1.4 核心流程图

```
用户输入问题/需求描述
        ↓
    自动制定计划
        ↓
    创建新分支
        ↓
  分析需求/问题定位
        ↓
    代码逻辑分析
        ↓
      完善代码
        ↓
  编写测试用例(单元/集成/E2E)
        ↓
   本地运行全量测试
        ↓
   ┌─────┴─────┐
   │ 测试通过？ │
   └─────┬─────┘
      否 ↓ 是
   返回修复  ↓
           提交代码
              ↓
        触发GitHub工作流
              ↓
       监控工作流执行
              ↓
       ┌─────┴─────┐
       │ 工作流通过？│
       └─────┬─────┘
          否 ↓ 是
       分析错误  ↓
       返回修复  生成总结
                 ↓
          更新问题跟踪文档
                 ↓
              创建PR
                 ↓
             流程结束
```

## 2. 工作流阶段详解

### 2.1 阶段一：需求/问题输入与计划制定

**输入格式**：

```
类型: [bug|feature|enhancement]
标题: 简短描述
描述: 详细描述问题或需求
优先级: [high|medium|low]
影响范围: [backend|frontend|both]
```

**输出**：

* 生成计划文档：`.trae/plans/{issue_id}_plan.md`

* 生成任务清单：`.trae/tasks/{issue_id}_tasks.md`

### 2.2 阶段二：分支管理

**分支命名规范**：

```
bugfix/issue-{id}-{short-description}    # Bug修复
feature/issue-{id}-{short-description}   # 新功能
enhancement/issue-{id}-{short-description} # 增强
hotfix/issue-{id}-{short-description}    # 紧急修复
```

**示例**：

```
bugfix/issue-123-fix-login-error
feature/issue-456-add-payment-method
```

### 2.3 阶段三：代码分析与定位

**分析步骤**：

1. 搜索相关代码文件
2. 分析代码依赖关系
3. 定位需要修改的模块
4. 评估影响范围

**输出**：

* 受影响文件列表

* 代码修改方案

* 测试覆盖需求

### 2.4 阶段四：代码实现

**实现规范**：

* 遵循项目代码规范

* 保持代码风格一致

* 添加必要的注释

* 处理边界情况和错误

### 2.5 阶段五：测试用例编写

#### 2.5.1 单元测试

**后端单元测试**：

```go
// 文件命名: {module}_test.go
// 位置: backend/{module}/{module}_test.go

func TestFunctionName(t *testing.T) {
    tests := []struct {
        name    string
        input   InputType
        want    OutputType
        wantErr bool
    }{
        // 测试用例
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 测试逻辑
        })
    }
}
```

**前端单元测试**：

```typescript
// 文件命名: {Component}.test.tsx 或 {service}.test.ts
// 位置: frontend/src/{module}/__tests__/{name}.test.ts(x)

describe('ModuleName', () => {
    it('should do something', () => {
        // 测试逻辑
    });
});
```

#### 2.5.2 集成测试

**后端集成测试**：

```go
// 文件命名: {module}_integration_test.go
// 位置: backend/{module}/{module}_integration_test.go
// 标记: 使用 // +build integration 或 -run Integration

func TestIntegration_APIEndpoint(t *testing.T) {
    // 需要真实数据库连接
    // 测试完整流程
}
```

#### 2.5.3 E2E测试

**前端E2E测试**：

```typescript
// 文件命名: {feature}.spec.ts
// 位置: frontend/e2e/{feature}.spec.ts

test.describe('Feature Name', () => {
    test('should complete user flow', async ({ page }) => {
        // E2E测试逻辑
    });
});
```

### 2.6 阶段六：本地测试验证

**测试命令**：

```bash
# 后端单元测试
cd backend && go test -v -short -race -coverprofile=coverage.out ./...

# 后端集成测试
cd backend && go test -v -run Integration ./...

# 前端单元测试
cd frontend && npm test -- --coverage --watchAll=false

# 前端E2E测试
cd frontend && npm run test:e2e

# 全量测试
make test
```

### 2.7 阶段七：代码提交与推送

**提交信息规范**：

```
<type>(<scope>): <subject>

<body>

<footer>
```

**类型**：

* `feat`: 新功能

* `fix`: Bug修复

* `docs`: 文档更新

* `style`: 代码格式

* `refactor`: 重构

* `test`: 测试

* `chore`: 构建/工具

**示例**：

```
feat(auth): add OAuth2 login support

- Add Google OAuth2 provider
- Add GitHub OAuth2 provider
- Update login page UI

Closes #123
```

### 2.8 阶段八：GitHub工作流监控

**工作流触发顺序**：

```
push/PR → ci-cd.yml → integration-tests.yml → e2e-tests.yml
```

**监控内容**：

1. 工作流状态（成功/失败）
2. 失败步骤定位
3. 错误日志分析
4. 自动修复建议

### 2.9 阶段九：错误处理与循环修复

**错误处理流程**：

```
检测到工作流失败
    ↓
分析失败原因
    ↓
定位错误代码
    ↓
修复代码/测试
    ↓
本地验证
    ↓
提交修复
    ↓
重新触发工作流
    ↓
循环直到通过
```

### 2.10 阶段十：总结与文档更新

**生成内容**：

1. 变更总结
2. 测试覆盖率报告
3. 影响范围说明
4. 后续建议

**更新文档**：

* `.trae/documents/issue_tracking.md` - 问题跟踪

* `.trae/documents/test_quality_tracking.md` - 测试质量跟踪

## 3. 文件结构设计

```
.trae/
├── documents/                    # 文档目录
│   ├── issue_tracking.md        # 问题跟踪文档
│   ├── test_quality_tracking.md # 测试质量跟踪
│   ├── workflow_history.md      # 工作流历史记录
│   └── templates/               # 模板目录
│       ├── bug_report.md        # Bug报告模板
│       ├── feature_request.md   # 功能请求模板
│       └── pr_template.md       # PR模板
├── plans/                        # 计划目录
│   └── issue_{id}_plan.md       # 各问题的计划文档
├── tasks/                        # 任务目录
│   └── issue_{id}_tasks.md      # 各问题的任务清单
└── cache/                        # 缓存目录
    └── workflow_state.json      # 工作流状态缓存
```

## 4. GitHub工作流配置

### 4.1 现有工作流分析

| 工作流                   | 触发条件                 | 主要任务                             |
| --------------------- | -------------------- | -------------------------------- |
| ci-cd.yml             | push/PR到main/develop | 后端单元测试、前端构建、前端单元测试、安全扫描、Docker构建 |
| integration-tests.yml | ci-cd完成后             | 后端集成测试                           |
| e2e-tests.yml         | integration-tests完成后 | E2E测试                            |
| deploy-tencent.yml    | 手动触发                 | 部署到腾讯云                           |

### 4.2 工作流优化建议

#### 4.2.1 添加工作流状态通知

```yaml
# .github/workflows/ci-cd.yml 添加
notify-status:
  name: Notify Workflow Status
  runs-on: ubuntu-latest
  needs: [backend-unit-tests, frontend-build, frontend-unit-tests, security-scan]
    if: always()
    steps:
      - name: Create status file
        run: |
          echo "workflow_status=${{ needs.backend-unit-tests.result == 'success' && needs.frontend-unit-tests.result == 'success' && 'success' || 'failure' }}" >> $GITHUB_OUTPUT
```

#### 4.2.2 添加测试失败详情报告

```yaml
# .github/workflows/ci-cd.yml 添加
test-report:
  name: Generate Test Report
  runs-on: ubuntu-latest
  needs: [backend-unit-tests, frontend-unit-tests]
  if: failure()
  steps:
    - uses: actions/checkout@v4
    - name: Generate failure report
      run: |
        echo "## Test Failure Report" >> $GITHUB_STEP_SUMMARY
        echo "Failed tests need attention" >> $GITHUB_STEP_SUMMARY
```

## 5. 问题跟踪文档模板

### 5.1 issue\_tracking.md 结构

```markdown
# 品多多项目 - 问题跟踪文档

## 当前活跃问题

| ID | 类型 | 标题 | 状态 | 分支 | 创建日期 | 更新日期 |
|----|------|------|------|------|----------|----------|
| 001 | bug | 登录失败 | 进行中 | bugfix/issue-001-login | 2026-03-21 | 2026-03-21 |

## 问题详情

### ISSUE-001: 登录失败

**类型**: bug
**状态**: 进行中
**优先级**: high
**分支**: bugfix/issue-001-login

**描述**:
用户使用正确凭据登录时返回401错误

**分析**:
- 根本原因: jwtSecret初始化时序问题
- 影响文件: backend/handlers/auth.go

**解决方案**:
将jwtSecret初始化移到init()函数

**测试覆盖**:
- [x] 单元测试
- [x] 集成测试
- [ ] E2E测试

**工作流状态**:
- CI/CD: ✅ 通过
- 集成测试: ✅ 通过
- E2E测试: 🔄 进行中

**总结**:
修复了jwtSecret初始化问题，所有测试通过。
```

## 6. 实现步骤

### 6.1 第一阶段：基础设施搭建

#### 6.1.1 创建目录结构

**工具**: RunCommand (mkdir) 或 Write (创建.gitkeep)

```bash
mkdir -p .trae/plans .trae/tasks .trae/cache .trae/documents/templates
```

**已完成**:
- ✅ `.trae/documents/` - 文档目录
- ✅ `.trae/documents/templates/` - 模板目录
- ✅ `.trae/cache/` - 缓存目录
- 🔄 `.trae/plans/` - 计划目录（运行时创建）
- 🔄 `.trae/tasks/` - 任务目录（运行时创建）

#### 6.1.2 创建模板文件

**工具**: Write

| 模板文件 | 状态 | 用途 |
|----------|------|------|
| bug_report.md | ✅ 已创建 | Bug报告模板 |
| feature_request.md | ✅ 已创建 | 功能请求模板 |
| plan_template.md | ✅ 已创建 | 计划模板 |
| tasks_template.md | ✅ 已创建 | 任务清单模板 |
| pr_template.md | ✅ 已创建 | PR模板 |

#### 6.1.3 创建跟踪文档

**工具**: Write

| 文档 | 状态 | 用途 |
|------|------|------|
| issue_tracking.md | ✅ 已创建 | 问题跟踪 |
| workflow_history.md | ✅ 已创建 | 工作流历史 |
| workflow_state.json | ✅ 已创建 | 状态缓存 |

### 6.2 第二阶段：工作流执行引擎

#### 6.2.1 问题解析模块

**输入**: 用户自然语言描述

**处理流程**:
1. AI解析用户输入，提取关键信息
2. 判断问题类型（bug/feature/enhancement）
3. 评估优先级和影响范围
4. 生成Issue ID

**输出**: 结构化的问题对象

```json
{
  "id": "ISSUE-001",
  "type": "bug",
  "title": "用户登录返回401错误",
  "description": "用户使用正确的用户名和密码登录时，系统返回401未授权错误",
  "priority": "high",
  "scope": "backend",
  "createdAt": "2026-03-21T10:00:00Z"
}
```

#### 6.2.2 计划生成模块

**工具**: Write

**流程**:
1. 基于问题类型选择模板
2. AI分析问题，生成解决方案
3. 分解任务步骤
4. 生成计划文档 `.trae/plans/issue_{id}_plan.md`
5. 生成任务清单 `.trae/tasks/issue_{id}_tasks.md`

#### 6.2.3 分支管理模块

**工具**: RunCommand (git)

**命令序列**:
```bash
# 1. 确保在主分支
git checkout main

# 2. 拉取最新代码
git pull origin main

# 3. 创建新分支
git checkout -b {type}/issue-{id}-{description}

# 4. 推送分支
git push -u origin {branch-name}
```

#### 6.2.4 代码分析模块

**工具**: SearchCodebase, Grep, Read

**流程**:
1. 使用SearchCodebase进行语义搜索
2. 使用Grep进行关键词搜索
3. 使用Read读取相关文件
4. AI分析代码逻辑和依赖关系
5. 定位需要修改的代码位置

**示例搜索**:
```
SearchCodebase: "登录认证逻辑"
Grep: "func.*Login|auth.*handler"
Read: backend/handlers/auth.go
```

#### 6.2.5 代码修改模块

**工具**: SearchReplace, Write

**流程**:
1. 分析需要修改的代码
2. 生成修改方案
3. 使用SearchReplace进行精确修改
4. 或使用Write创建新文件

**修改原则**:
- 最小化修改范围
- 保持代码风格一致
- 添加必要的错误处理
- 遵循项目规范

#### 6.2.6 测试生成模块

**工具**: Write

**单元测试生成**:
```go
// 后端单元测试
func TestFunctionName(t *testing.T) {
    // AI生成测试用例
}
```

**集成测试生成**:
```go
// 后端集成测试
func TestIntegration_APIFlow(t *testing.T) {
    // AI生成集成测试场景
}
```

**E2E测试生成**:
```typescript
// 前端E2E测试
test('should complete user flow', async ({ page }) => {
    // AI生成E2E测试步骤
});
```

#### 6.2.7 本地测试模块

**工具**: RunCommand

**测试命令序列**:
```bash
# 1. 后端单元测试
cd backend && go test -v -short -race -coverprofile=coverage.out ./...

# 2. 后端集成测试（如需要）
cd backend && go test -v -run Integration ./...

# 3. 前端单元测试
cd frontend && npm test -- --coverage --watchAll=false

# 4. 前端E2E测试
cd frontend && npm run test:e2e

# 5. 代码风格检查
cd backend && golangci-lint run
cd frontend && npm run lint
```

**结果解析**:
- 解析测试输出
- 提取失败信息
- 计算覆盖率
- 决定是否继续

#### 6.2.8 代码提交模块

**工具**: RunCommand (git)

**提交流程**:
```bash
# 1. 查看变更
git status
git diff

# 2. 添加文件
git add .

# 3. 提交（AI生成提交信息）
git commit -m "fix(auth): resolve jwtSecret initialization timing issue

- Move jwtSecret initialization to init() function
- Add unit tests for auth handler
- Add integration tests for login flow

Closes #ISSUE-001"

# 4. 推送
git push origin {branch-name}
```

#### 6.2.9 GitHub工作流监控模块

**工具**: RunCommand (gh), GitHub API

**监控流程**:
```bash
# 1. 获取工作流运行列表
gh run list --branch={branch-name} --limit=1

# 2. 获取运行ID
RUN_ID=$(gh run list --branch={branch-name} --json databaseId -q '.[0].databaseId')

# 3. 查看运行状态
gh run view $RUN_ID

# 4. 获取Job列表
gh api repos/{owner}/{repo}/actions/runs/$RUN_ID/jobs

# 5. 获取失败日志
gh run view $RUN_ID --log-failed

# 6. 等待完成
gh run watch $RUN_ID
```

**状态轮询**:
```
while true; do
    STATUS=$(gh run view $RUN_ID --json status -q '.status')
    if [ "$STATUS" = "completed" ]; then
        CONCLUSION=$(gh run view $RUN_ID --json conclusion -q '.conclusion')
        break
    fi
    sleep 30
done
```

#### 6.2.10 错误修复循环模块

**流程**:
```
检测失败
    ↓
获取失败日志
    ↓
AI分析错误原因
    ↓
定位问题代码
    ↓
生成修复方案
    ↓
应用修复
    ↓
本地验证
    ↓
提交修复
    ↓
重新监控工作流
    ↓
循环直到成功
```

**错误类型处理**:

| 错误类型 | 处理方式 |
|----------|----------|
| 编译错误 | 分析语法错误，修复代码 |
| 测试失败 | 分析测试日志，修复代码或测试 |
| Lint错误 | 按规范修复代码风格 |
| 安全漏洞 | 修复安全问题 |
| 工作流配置错误 | 修复YAML配置 |

#### 6.2.11 文档更新模块

**工具**: Read, SearchReplace

**更新流程**:
1. 读取现有跟踪文档
2. 更新问题状态
3. 添加工作流结果
4. 生成解决总结
5. 更新统计数据

### 6.3 第三阶段：GitHub工作流优化

#### 6.3.1 添加工作流状态输出

**文件**: `.github/workflows/ci-cd.yml`

**添加内容**:
```yaml
outputs:
  workflow_status:
    description: 'Overall workflow status'
    value: ${{ jobs.notify.outputs.status }}
```

#### 6.3.2 添加失败详情报告

**文件**: `.github/workflows/ci-cd.yml`

**添加Job**:
```yaml
failure-report:
  name: Generate Failure Report
  runs-on: ubuntu-latest
  needs: [backend-unit-tests, frontend-unit-tests]
  if: failure()
  steps:
    - name: Create failure summary
      run: |
        echo "## Failure Report" >> $GITHUB_STEP_SUMMARY
        echo "One or more jobs failed. Please check the logs." >> $GITHUB_STEP_SUMMARY
```

### 6.4 第四阶段：集成测试

#### 6.4.1 端到端流程测试

**测试场景**:
1. Bug修复流程测试
2. 新功能开发流程测试
3. 错误恢复流程测试
4. 工作流失败处理测试

#### 6.4.2 边界条件测试

**测试项**:
- 合并冲突处理
- 多次失败重试
- 超时处理
- 网络错误处理

## 7. 使用示例

### 7.1 Bug修复流程示例

**用户输入**：

```
类型: bug
标题: 用户登录返回401错误
描述: 用户使用正确的用户名和密码登录时，系统返回401未授权错误
优先级: high
影响范围: backend
```

**执行流程**：

1. 创建计划文档 `.trae/plans/issue_001_plan.md`
2. 创建分支 `bugfix/issue-001-login-401`
3. 分析代码，定位问题在 `backend/handlers/auth.go`
4. 修复jwtSecret初始化问题
5. 编写单元测试 `backend/handlers/auth_test.go`
6. 编写集成测试 `backend/handlers/auth_integration_test.go`
7. 本地运行测试通过
8. 提交代码
9. 推送分支，触发CI/CD
10. 监控工作流执行
11. 所有测试通过
12. 生成总结，更新问题跟踪文档

### 7.2 新功能开发流程示例

**用户输入**：

```
类型: feature
标题: 添加商品收藏功能
描述: 用户可以收藏喜欢的商品，在个人中心查看收藏列表
优先级: medium
影响范围: both
```

**执行流程**：

1. 创建计划文档
2. 创建分支 `feature/issue-002-favorite-products`
3. 设计数据库表结构
4. 实现后端API
5. 实现前端页面
6. 编写后端单元测试和集成测试
7. 编写前端单元测试
8. 编写E2E测试
9. 本地运行全量测试
10. 提交代码，触发工作流
11. 监控并修复问题
12. 生成总结，更新文档

## 8. 质量保证

### 8.1 代码质量检查

* [ ] 代码风格检查通过

* [ ] 静态分析通过

* [ ] 安全扫描通过

* [ ] 无编译警告

### 8.2 测试质量检查

* [ ] 单元测试覆盖率达标（后端≥85%，前端≥80%）

* [ ] 集成测试覆盖核心流程

* [ ] E2E测试覆盖用户场景

* [ ] 所有测试用例通过

### 8.3 文档质量检查

* [ ] 代码注释完整

* [ ] API文档更新

* [ ] 变更日志更新

* [ ] 问题跟踪文档更新

## 9. 异常处理

### 9.1 测试失败处理

1. 分析失败原因
2. 定位问题代码
3. 修复并重新测试
4. 记录问题和解决方案

### 9.2 工作流失败处理

1. 获取失败日志
2. 分析失败步骤
3. 本地复现问题
4. 修复并重新提交

### 9.3 合并冲突处理

1. 拉取最新主分支
2. 解决冲突
3. 重新测试
4. 提交解决结果

## 10. 持续改进

### 10.1 定期回顾

* 每周回顾工作流效率

* 分析常见失败原因

* 优化测试用例

* 更新工作流配置

### 10.2 指标跟踪

* 平均问题解决时间

* 测试通过率

* 代码覆盖率趋势

* 工作流成功率

## 11. 附录

### 11.1 常用命令

```bash
# 创建新分支
git checkout -b bugfix/issue-{id}-{description}

# 运行后端测试
cd backend && go test -v -race -coverprofile=coverage.out ./...

# 运行前端测试
cd frontend && npm test -- --coverage

# 运行E2E测试
cd frontend && npm run test:e2e

# 提交代码
git add .
git commit -m "type(scope): subject"

# 推送分支
git push origin <branch-name>
```

### 11.2 相关文档

* [测试质量跟踪文档](./test_quality_tracking.md)

* [集成测试用例设计](./integration_test_case_design.md)

* [测试质量提升计划](./test_quality_improvement_plan.md)

---

## 12. 实施状态总结

### 12.1 已完成工作

| 阶段 | 任务 | 状态 | 备注 |
|------|------|------|------|
| 基础设施 | 目录结构创建 | ✅ 完成 | .trae/plans, .trae/tasks, .trae/cache, .trae/documents, .trae/skills |
| 基础设施 | 模板文件创建 | ✅ 完成 | 5个模板文件 |
| 基础设施 | 跟踪文档创建 | ✅ 完成 | issue_tracking.md, workflow_history.md |
| 基础设施 | 状态缓存创建 | ✅ 完成 | workflow_state.json |
| 基础设施 | 快速参考指南 | ✅ 完成 | quick_reference.md |
| 设计文档 | 工作流设计方案 | ✅ 完成 | 本文档 |
| **Skill** | **mydev-github-workflow** | ✅ 完成 | **核心执行引擎** |

### 12.2 目录结构现状

```
.trae/
├── skills/
│   └── mydev-github-workflow/
│       └── SKILL.md              ✅ 已创建（核心Skill）
├── cache/
│   └── workflow_state.json       ✅ 已创建
├── documents/
│   ├── templates/
│   │   ├── bug_report.md         ✅ 已创建
│   │   ├── feature_request.md    ✅ 已创建
│   │   ├── plan_template.md      ✅ 已创建
│   │   ├── pr_template.md        ✅ 已创建
│   │   └── tasks_template.md     ✅ 已创建
│   ├── github_workflow_design.md ✅ 已创建（本文档）
│   ├── integration_test_case_design.md ✅ 已存在
│   ├── issue_tracking.md         ✅ 已创建
│   ├── quick_reference.md        ✅ 已创建
│   ├── test_quality_improvement_plan.md ✅ 已存在
│   ├── test_quality_tracking.md  ✅ 已存在
│   └── workflow_history.md       ✅ 已创建
├── plans/                        ✅ 目录已创建（运行时生成文件）
└── tasks/                        ✅ 目录已创建（运行时生成文件）
```

### 12.3 隔离性验证

| 检查项 | 状态 | 说明 |
|--------|------|------|
| 文件位置 | ✅ 通过 | 所有文件均在 `.trae/` 目录下 |
| 项目结构 | ✅ 通过 | 未修改项目原有文件 |
| Git忽略 | ⚠️ 建议 | 建议在 `.gitignore` 中添加 `.trae/cache/` |

### 12.4 使用方式

用户只需按照以下格式输入问题或需求，**mydev-github-workflow** skill将自动触发并执行完整流程：

```
类型: [bug|feature|enhancement]
标题: 简短描述
描述: 详细描述问题或需求
优先级: [high|medium|low]
影响范围: [backend|frontend|both]
```

或者直接用自然语言描述：
```
登录功能有问题，用户使用正确的账号密码登录时返回401错误
```

系统将自动执行完整的开发和测试验证流程，直到GitHub工作流全部通过，并生成总结报告。

