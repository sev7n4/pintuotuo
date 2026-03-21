# GitHub集成自动化开发工作流 Skill 规格

## Why

当前项目开发流程缺乏自动化，每次处理Bug或新功能需要手动执行多个步骤（创建分支、编写代码、测试、提交、监控CI等），效率低下且容易遗漏步骤。需要一套自动化的工作流系统，实现从问题输入到代码合并的完整闭环，提高开发效率和代码质量。

## What Changes

- 创建 `mydev-github-workflow` Skill，封装完整的自动化开发流程
- 创建配套的文档模板和跟踪系统
- 实现与现有GitHub CI/CD工作流的完美集成
- **BREAKING**: 无破坏性变更，所有文件隔离在 `.trae/` 目录下

## Impact

- Affected specs: 无现有spec受影响（新增功能）
- Affected code: 不影响项目原有代码，仅在 `.trae/` 目录下新增文件

## ADDED Requirements

### Requirement: Skill核心功能

系统SHALL提供 `mydev-github-workflow` Skill，当用户输入问题或需求描述时自动触发并执行完整的开发流程。

#### Scenario: Bug修复流程

- **WHEN** 用户输入Bug描述（如："登录功能返回401错误"）
- **THEN** 系统自动执行：问题解析 → 计划制定 → 分支创建 → 代码分析 → 代码修复 → 测试编写 → 本地验证 → 代码提交 → CI监控 → 文档更新

#### Scenario: 新功能开发流程

- **WHEN** 用户输入功能需求（如："添加商品收藏功能"）
- **THEN** 系统自动执行完整开发流程，包括数据库设计、API实现、前端开发、测试编写等

### Requirement: 问题解析模块

系统SHALL能够解析用户的自然语言输入，提取关键信息并生成结构化的问题对象。

#### Scenario: 结构化输入解析

- **WHEN** 用户提供结构化输入（类型、标题、描述、优先级、影响范围）
- **THEN** 系统生成包含 id、type、title、description、priority、scope、createdAt 的JSON对象

#### Scenario: 自然语言输入解析

- **WHEN** 用户用自然语言描述问题（如："登录失败返回401"）
- **THEN** 系统AI智能分析并提取问题类型、优先级、影响范围等信息

### Requirement: 分支管理模块

系统SHALL按照规范创建和管理Git分支。

#### Scenario: Bug修复分支创建

- **WHEN** 问题类型为bug
- **THEN** 创建格式为 `bugfix/issue-{id}-{short-description}` 的分支

#### Scenario: 新功能分支创建

- **WHEN** 问题类型为feature
- **THEN** 创建格式为 `feature/issue-{id}-{short-description}` 的分支

### Requirement: 代码分析与修改模块

系统SHALL能够分析项目代码并定位需要修改的位置。

#### Scenario: 代码定位

- **WHEN** 系统需要修改代码
- **THEN** 使用 SearchCodebase（语义搜索）、Grep（关键词搜索）、Read（文件读取）工具定位代码

#### Scenario: 代码修改

- **WHEN** 定位到需要修改的代码
- **THEN** 使用 SearchReplace 进行精确修改，或使用 Write 创建新文件

### Requirement: 测试生成模块

系统SHALL为每个修改生成完整的测试用例。

#### Scenario: 后端单元测试生成

- **WHEN** 修改后端代码
- **THEN** 在 `backend/{module}/{module}_test.go` 生成单元测试，覆盖率目标 ≥85%

#### Scenario: 前端单元测试生成

- **WHEN** 修改前端代码
- **THEN** 在 `frontend/src/{module}/__tests__/{name}.test.ts(x)` 生成单元测试，覆盖率目标 ≥80%

#### Scenario: 集成测试生成

- **WHEN** 修改涉及API或数据库操作
- **THEN** 生成集成测试用例，函数命名以 `TestIntegration_` 开头

#### Scenario: E2E测试生成

- **WHEN** 修改涉及用户界面流程
- **THEN** 在 `frontend/e2e/{feature}.spec.ts` 生成E2E测试

### Requirement: 本地验证模块

系统SHALL在提交代码前执行本地测试验证。

#### Scenario: 后端测试验证

- **WHEN** 后端代码修改完成
- **THEN** 执行 `go test -v -short -race -coverprofile=coverage.out ./...`

#### Scenario: 前端测试验证

- **WHEN** 前端代码修改完成
- **THEN** 执行 `npm test -- --coverage --watchAll=false`

#### Scenario: 测试失败处理

- **WHEN** 本地测试失败
- **THEN** 分析失败原因，修复代码或测试，重新验证

### Requirement: 代码提交模块

系统SHALL按照规范提交代码。

#### Scenario: 提交信息格式

- **WHEN** 提交代码
- **THEN** 使用格式 `<type>(<scope>): <subject>`，类型包括 feat、fix、docs、style、refactor、test、chore

#### Scenario: 提交内容关联

- **WHEN** 提交代码
- **THEN** 在提交信息footer中包含 `Closes #ISSUE-{id}`

### Requirement: GitHub工作流监控模块

系统SHALL监控GitHub CI/CD工作流执行状态。

#### Scenario: 工作流触发

- **WHEN** 代码推送到远程分支
- **THEN** 自动触发 ci-cd.yml → integration-tests.yml → e2e-tests.yml 工作流链

#### Scenario: 工作流状态查询

- **WHEN** 需要检查工作流状态
- **THEN** 使用 `gh run list` 和 `gh run view` 命令获取状态

#### Scenario: 工作流失败处理

- **WHEN** 工作流失败
- **THEN** 使用 `gh run view {run-id} --log-failed` 获取失败日志，分析并修复

### Requirement: 错误修复循环模块

系统SHALL在工作流失败时自动进入修复循环。

#### Scenario: 编译错误修复

- **WHEN** 检测到编译错误
- **THEN** 分析语法/类型错误，修复代码，重新提交

#### Scenario: 测试失败修复

- **WHEN** 检测到测试失败
- **THEN** 分析测试日志，修复代码或测试用例，重新验证

#### Scenario: 最大重试限制

- **WHEN** 修复循环达到最大次数（5次）
- **THEN** 生成详细报告，请求人工介入

### Requirement: 文档更新模块

系统SHALL在工作流完成后更新跟踪文档。

#### Scenario: 问题跟踪更新

- **WHEN** 工作流成功完成
- **THEN** 更新 `.trae/documents/issue_tracking.md`，添加问题详情和解决方案

#### Scenario: 工作流历史记录

- **WHEN** 工作流执行完成
- **THEN** 更新 `.trae/documents/workflow_history.md`，记录执行时间和结果

### Requirement: 文件隔离性

系统SHALL确保所有工作流相关文件存放在 `.trae/` 目录下，不影响项目原有结构。

#### Scenario: 目录结构隔离

- **WHEN** 创建任何工作流文件
- **THEN** 文件路径必须以 `.trae/` 开头

#### Scenario: 可删除性

- **WHEN** 用户删除 `.trae/` 目录
- **THEN** 项目原有代码和配置不受影响

## MODIFIED Requirements

无修改的现有需求。

## REMOVED Requirements

无删除的现有需求。
