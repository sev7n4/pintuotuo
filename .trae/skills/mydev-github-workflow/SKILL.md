---
name: "mydev-github-workflow"
description: "Automated development workflow integrated with GitHub CI/CD. Invoke when user reports a bug, requests a feature, describes a code change requirement, or provides structured issue input (type/title/description/priority/scope)."
---

# MyDev GitHub Workflow - 自动化开发工作流

## 触发条件

当用户满足以下任一条件时，**必须**立即调用此skill：

1. 用户报告Bug（如："有个bug..."、"登录失败..."）
2. 用户请求新功能（如："需要添加..."、"实现一个..."）
3. 用户描述代码改进需求（如："优化..."、"重构..."）
4. 用户提供结构化问题输入（包含类型、标题、描述等）

## 工作流概述

这是一个完整的自动化开发工作流，实现从问题输入到代码合并的闭环：

```
用户输入 → 制定计划 → 创建分支 → 代码分析 → 代码实现 → 测试编写 → 本地验证 → 提交代码 → GitHub工作流监控 → (失败则循环修复) → 生成总结 → 更新文档
```

## 文件位置约定

**重要**：所有工作流相关文件**必须**存放在 `.trae/` 目录下，确保不影响项目原有结构：

```
.trae/
├── skills/mydev-github-workflow/SKILL.md  # 本skill文件
├── documents/
│   ├── github_workflow_design.md          # 工作流设计文档
│   ├── issue_tracking.md                  # 问题跟踪文档
│   ├── workflow_history.md                # 工作流历史
│   └── templates/                         # 模板目录
│       ├── bug_report.md
│       ├── feature_request.md
│       ├── plan_template.md
│       ├── tasks_template.md
│       └── pr_template.md
├── plans/                                 # 运行时生成计划
│   └── issue_{id}_plan.md
├── tasks/                                 # 运行时生成任务
│   └── issue_{id}_tasks.md
└── cache/
    └── workflow_state.json                # 工作流状态缓存
```

## 工作流执行步骤

### 步骤1：问题解析

**目标**：解析用户输入，生成结构化问题对象

**处理逻辑**：
1. 分析用户输入内容
2. 判断问题类型：`bug` | `feature` | `enhancement`
3. 评估优先级：`high` | `medium` | `low`
4. 确定影响范围：`backend` | `frontend` | `both`
5. 生成Issue ID（格式：ISSUE-{序号}）

**输出示例**：
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

### 步骤2：计划生成

**目标**：生成计划文档和任务清单

**工具**：Write

**操作**：
1. 读取模板：`.trae/documents/templates/plan_template.md`
2. 读取模板：`.trae/documents/templates/tasks_template.md`
3. 基于问题分析填充模板
4. 生成文件：`.trae/plans/issue_{id}_plan.md`
5. 生成文件：`.trae/tasks/issue_{id}_tasks.md`

### 步骤3：分支创建

**目标**：创建符合规范的Git分支

**工具**：RunCommand (git)

**分支命名规范**：
```
bugfix/issue-{id}-{short-description}     # Bug修复
feature/issue-{id}-{short-description}    # 新功能
enhancement/issue-{id}-{short-description} # 增强
hotfix/issue-{id}-{short-description}     # 紧急修复
```

**命令序列**：
```bash
git checkout main
git pull origin main
git checkout -b {type}/issue-{id}-{description}
git push -u origin {branch-name}
```

### 步骤4：代码分析

**目标**：定位需要修改的代码位置

**工具**：SearchCodebase, Grep, Read

**分析流程**：
1. 使用SearchCodebase进行语义搜索
2. 使用Grep进行关键词/正则搜索
3. 使用Read读取相关文件
4. 分析代码依赖关系
5. 确定修改范围

**输出**：
- 受影响文件列表
- 代码修改方案
- 测试覆盖需求

### 步骤5：代码实现

**目标**：实现代码修改

**工具**：SearchReplace, Write

**实现原则**：
- 最小化修改范围
- 保持代码风格一致
- 添加必要的错误处理
- 遵循项目现有规范

**操作**：
- 修改现有文件：使用SearchReplace进行精确替换
- 创建新文件：使用Write创建

### 步骤6：测试编写

**目标**：编写完整的测试用例

**工具**：Write

**测试类型**：

1. **单元测试**
   - 后端：`backend/{module}/{module}_test.go`
   - 前端：`frontend/src/{module}/__tests__/{name}.test.ts(x)`

2. **集成测试**
   - 后端：`backend/{module}/{module}_integration_test.go`
   - 函数命名：`TestIntegration_*`

3. **E2E测试**
   - 前端：`frontend/e2e/{feature}.spec.ts`

**覆盖率要求**：
- 后端单元测试覆盖率 ≥ 85%
- 前端单元测试覆盖率 ≥ 80%

### 步骤7：本地验证

**目标**：确保本地测试通过

**工具**：RunCommand

**测试命令**：

```bash
# 后端单元测试
cd backend && go test -v -short -race -coverprofile=coverage.out ./...

# 后端集成测试（如需要）
cd backend && go test -v -run Integration ./...

# 前端单元测试
cd frontend && npm test -- --coverage --watchAll=false

# 前端E2E测试
cd frontend && npm run test:e2e

# 代码风格检查
cd backend && golangci-lint run --timeout=5m
cd frontend && npm run lint
```

**验证标准**：
- 所有测试通过
- 覆盖率达标
- 无Lint错误

### 步骤8：代码提交

**目标**：提交代码并推送到远程

**工具**：RunCommand (git)

**提交信息规范**：
```
<type>(<scope>): <subject>

<body>

<footer>
```

**类型**：`feat` | `fix` | `docs` | `style` | `refactor` | `test` | `chore`

**命令序列**：
```bash
git status
git diff
git add .
git commit -m "fix(auth): resolve jwtSecret initialization timing issue

- Move jwtSecret initialization to init() function
- Add unit tests for auth handler

Closes #ISSUE-001"
git push origin {branch-name}
```

### 步骤9：GitHub工作流监控

**目标**：监控CI/CD工作流执行状态

**工具**：RunCommand (gh)

**工作流触发顺序**：
```
push → ci-cd.yml → integration-tests.yml → e2e-tests.yml
```

**监控命令**：
```bash
# 获取最新工作流运行
gh run list --branch={branch-name} --limit=1

# 获取运行ID
RUN_ID=$(gh run list --branch={branch-name} --json databaseId -q '.[0].databaseId')

# 查看运行状态
gh run view $RUN_ID

# 等待完成
gh run watch $RUN_ID

# 获取失败日志
gh run view $RUN_ID --log-failed
```

**状态判断**：
- `success`：工作流成功，进入步骤11
- `failure`：工作流失败，进入步骤10

### 步骤10：错误修复循环

**目标**：分析失败原因并修复

**触发条件**：GitHub工作流失败

**处理流程**：
```
1. 获取失败日志
   gh run view $RUN_ID --log-failed

2. 分析失败原因
   - 编译错误 → 修复语法/类型问题
   - 测试失败 → 分析测试日志，修复代码或测试
   - Lint错误 → 修复代码风格
   - 安全漏洞 → 修复安全问题

3. 定位问题代码
   使用 SearchCodebase/Grep/Read

4. 应用修复
   使用 SearchReplace/Write

5. 本地验证
   重新执行步骤7

6. 提交修复
   git add . && git commit -m "fix: resolve CI failure" && git push

7. 重新监控
   返回步骤9
```

**循环终止条件**：
- 工作流成功
- 或达到最大重试次数（建议5次）

### 步骤11：生成总结与更新文档

**目标**：生成解决总结，更新跟踪文档

**工具**：Read, SearchReplace

**更新内容**：

1. **更新问题跟踪文档** `.trae/documents/issue_tracking.md`
   - 添加到"已解决问题"列表
   - 填写问题详情和解决方案
   - 记录工作流状态

2. **更新工作流历史** `.trae/documents/workflow_history.md`
   - 记录执行时间
   - 记录工作流结果
   - 记录失败次数（如有）

3. **更新状态缓存** `.trae/cache/workflow_state.json`
   - 清空当前工作流状态
   - 更新统计数据

**总结模板**：
```markdown
## ISSUE-{id} 解决总结

**问题**：{问题描述}
**类型**：{bug/feature/enhancement}
**解决方案**：{解决方案描述}

**变更文件**：
- backend/path/to/file.go
- frontend/src/path/to/file.tsx

**测试覆盖**：
- 单元测试：{数量}个用例，覆盖率{X}%
- 集成测试：{数量}个用例
- E2E测试：{数量}个用例

**工作流状态**：
- CI/CD：✅ 通过
- 集成测试：✅ 通过
- E2E测试：✅ 通过

**耗时**：{X}分钟
**重试次数**：{X}次
```

### 步骤12：创建PR（可选）

**目标**：创建Pull Request

**工具**：RunCommand (gh)

**命令**：
```bash
gh pr create --title "{type}({scope}): {subject}" --body-file .trae/pr_body.md
```

**PR描述模板**：`.trae/documents/templates/pr_template.md`

## 错误处理

### 常见错误类型

| 错误类型 | 处理方式 |
|----------|----------|
| 编译错误 | 分析编译输出，修复语法/类型错误 |
| 测试失败 | 分析测试日志，修复代码或测试用例 |
| Lint错误 | 按规范修复代码风格 |
| 安全漏洞 | 修复安全问题或添加例外 |
| 合并冲突 | 拉取最新代码，解决冲突，重新提交 |
| 工作流超时 | 检查是否有死循环或性能问题 |

### 重试策略

- 最大重试次数：5次
- 每次重试前必须分析失败原因
- 超过最大次数后，生成详细报告供人工介入

## 质量标准

### 代码质量
- [ ] 代码风格检查通过
- [ ] 静态分析通过
- [ ] 安全扫描通过
- [ ] 无编译警告

### 测试质量
- [ ] 后端单元测试覆盖率 ≥ 85%
- [ ] 前端单元测试覆盖率 ≥ 80%
- [ ] 集成测试覆盖核心流程
- [ ] E2E测试覆盖用户场景
- [ ] 所有测试用例通过

### 文档质量
- [ ] 代码注释完整
- [ ] 问题跟踪文档更新
- [ ] 工作流历史更新

## 使用示例

### 示例1：Bug修复

**用户输入**：
```
登录功能有问题，用户使用正确的账号密码登录时返回401错误
```

**执行流程**：
1. 解析为Bug类型，优先级high
2. 生成计划 `.trae/plans/issue_001_plan.md`
3. 创建分支 `bugfix/issue-001-login-401`
4. 分析定位到 `backend/handlers/auth.go`
5. 修复jwtSecret初始化问题
6. 编写单元测试和集成测试
7. 本地测试通过
8. 提交代码
9. 监控CI/CD → 通过
10. 更新跟踪文档

### 示例2：新功能开发

**用户输入**：
```
类型: feature
标题: 添加商品收藏功能
描述: 用户可以收藏喜欢的商品，在个人中心查看收藏列表
优先级: medium
影响范围: both
```

**执行流程**：
1. 解析为Feature类型
2. 生成计划
3. 创建分支 `feature/issue-002-favorite-products`
4. 设计数据库表结构
5. 实现后端API
6. 实现前端页面
7. 编写完整测试
8. 本地验证
9. 提交并监控工作流
10. 循环修复直到通过
11. 更新文档

## 注意事项

1. **文件隔离**：所有工作流文件必须在 `.trae/` 目录下
2. **分支规范**：严格遵守分支命名规范
3. **提交规范**：严格遵守提交信息格式
4. **测试优先**：本地测试通过后才能提交
5. **循环修复**：工作流失败必须分析原因后修复
6. **文档更新**：完成后必须更新跟踪文档
