# 品多多项目 - GitHub集成自动化开发工作流设计方案

## 1. 工作流概述

### 1.1 设计目标

设计一套与GitHub工作流完美结合的自动化开发流程，实现从问题/需求输入到代码合并、测试验证的完整闭环机制。

### 1.2 核心流程图

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
- 生成计划文档：`.trae/plans/{issue_id}_plan.md`
- 生成任务清单：`.trae/tasks/{issue_id}_tasks.md`

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
- 受影响文件列表
- 代码修改方案
- 测试覆盖需求

### 2.4 阶段四：代码实现

**实现规范**：
- 遵循项目代码规范
- 保持代码风格一致
- 添加必要的注释
- 处理边界情况和错误

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
- `feat`: 新功能
- `fix`: Bug修复
- `docs`: 文档更新
- `style`: 代码格式
- `refactor`: 重构
- `test`: 测试
- `chore`: 构建/工具

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
- `.trae/documents/issue_tracking.md` - 问题跟踪
- `.trae/documents/test_quality_tracking.md` - 测试质量跟踪

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

| 工作流 | 触发条件 | 主要任务 |
|--------|----------|----------|
| ci-cd.yml | push/PR到main/develop | 后端单元测试、前端构建、前端单元测试、安全扫描、Docker构建 |
| integration-tests.yml | ci-cd完成后 | 后端集成测试 |
| e2e-tests.yml | integration-tests完成后 | E2E测试 |
| deploy-tencent.yml | 手动触发 | 部署到腾讯云 |

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

### 5.1 issue_tracking.md 结构

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

1. **创建目录结构**
   - 创建 `.trae/plans/` 目录
   - 创建 `.trae/tasks/` 目录
   - 创建 `.trae/cache/` 目录
   - 创建 `.trae/documents/templates/` 目录

2. **创建模板文件**
   - Bug报告模板
   - 功能请求模板
   - PR模板

3. **创建问题跟踪文档**
   - 初始化 `issue_tracking.md`
   - 初始化 `workflow_history.md`

### 6.2 第二阶段：工作流配置优化

1. **优化现有GitHub工作流**
   - 添加状态输出
   - 添加失败详情报告
   - 添加工作流ID便于追踪

2. **创建工作流监控配置**
   - 定义监控指标
   - 定义失败处理策略

### 6.3 第三阶段：工作流执行规范

1. **定义标准操作流程**
   - 问题输入规范
   - 分支创建规范
   - 提交信息规范

2. **定义测试验证标准**
   - 单元测试覆盖率要求
   - 集成测试场景要求
   - E2E测试流程要求

### 6.4 第四阶段：文档更新机制

1. **自动更新规则**
   - 工作流完成后更新状态
   - 测试通过后更新覆盖率
   - 问题解决后更新总结

2. **文档模板**
   - 问题解决报告模板
   - 测试报告模板
   - 变更日志模板

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

- [ ] 代码风格检查通过
- [ ] 静态分析通过
- [ ] 安全扫描通过
- [ ] 无编译警告

### 8.2 测试质量检查

- [ ] 单元测试覆盖率达标（后端≥85%，前端≥80%）
- [ ] 集成测试覆盖核心流程
- [ ] E2E测试覆盖用户场景
- [ ] 所有测试用例通过

### 8.3 文档质量检查

- [ ] 代码注释完整
- [ ] API文档更新
- [ ] 变更日志更新
- [ ] 问题跟踪文档更新

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

- 每周回顾工作流效率
- 分析常见失败原因
- 优化测试用例
- 更新工作流配置

### 10.2 指标跟踪

- 平均问题解决时间
- 测试通过率
- 代码覆盖率趋势
- 工作流成功率

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

- [测试质量跟踪文档](./test_quality_tracking.md)
- [集成测试用例设计](./integration_test_case_design.md)
- [测试质量提升计划](./test_quality_improvement_plan.md)
