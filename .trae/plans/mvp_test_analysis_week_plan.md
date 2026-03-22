# 一周攻关任务计划：MVP测试用例分析与实现

**创建时间**: 2026-03-23
**执行周期**: 1周
**目标**: 对MVP测试用例进行全面分析，识别未实现功能和问题，通过工作流逐个实现，形成闭环

---

## 一、任务背景

### 1.1 项目现状
- **项目名称**: 拼脱脱 (Pintuotuo) - AI Token二级市场平台
- **当前阶段**: MVP开发阶段
- **CI/CD状态**: 完整链路已建立 (CI/CD → Integration → E2E → Deploy)
- **工作流**: 使用 `mydev-github-workflow` skill 进行自动化开发

### 1.2 测试覆盖率现状

| 层级 | 文件数 | 覆盖状态 | 主要缺口 |
|-----|-------|---------|---------|
| 后端单元测试 | 27 | ⚠️ 部分 | 订单、支付、拼团、管理员模块缺失 |
| 后端集成测试 | 5 | ⚠️ 部分 | 核心业务流程集成测试不足 |
| 前端单元测试 | 37 | ✅ 良好 | 管理员后台未测试 |
| E2E测试 | 5 | ⚠️ 部分 | 完整购物流程、支付流程缺失 |

---

## 二、任务分解（按天）

### Day 1 (周一): 需求分析与测试用例梳理

#### 上午：MVP功能清单核对
- [ ] 对照 `01_PRD_Complete_Product_Specification.md` 核对已实现功能
- [ ] 对照 `04_API_Specification.md` 核对已实现API
- [ ] 对照 `03_Data_Model_Design.md` 核对数据模型

#### 下午：测试用例差距分析
- [ ] 整理现有测试用例清单
- [ ] 识别未覆盖的业务场景
- [ ] 输出：`mvp_test_gap_analysis.md`

### Day 2 (周二): 后端核心模块分析

#### 上午：订单模块分析
- [ ] 分析 `handlers/order_and_group.go` 代码逻辑
- [ ] 对照API文档验证实现完整性
- [ ] 识别缺失的测试用例
- [ ] 输出：订单模块问题清单

#### 下午：支付模块分析
- [ ] 分析 `handlers/payment_and_token.go` 和 `handlers/payment_v2.go`
- [ ] 验证支付流程完整性
- [ ] 识别缺失的测试用例
- [ ] 输出：支付模块问题清单

### Day 3 (周三): 拼团与管理员模块分析

#### 上午：拼团模块分析
- [ ] 分析拼团创建、加入、完成逻辑
- [ ] 验证拼团状态机完整性
- [ ] 识别缺失的测试用例
- [ ] 输出：拼团模块问题清单

#### 下午：管理员模块分析
- [ ] 分析 `handlers/admin.go` 功能
- [ ] 验证权限控制逻辑
- [ ] 识别缺失的测试用例
- [ ] 输出：管理员模块问题清单

### Day 4 (周四): E2E流程分析与问题单创建

#### 上午：核心用户流程分析
- [ ] 分析完整购物流程覆盖情况
- [ ] 分析支付流程覆盖情况
- [ ] 分析拼团参与流程覆盖情况
- [ ] 输出：E2E流程缺口清单

#### 下午：创建GitHub Issues
- [ ] 为每个缺失功能创建 Feature Issue
- [ ] 为每个问题创建 Bug Issue
- [ ] 标记优先级和模块标签
- [ ] 输出：Issue清单

### Day 5 (周五): 工作流执行与跟踪

#### 上午：启动高优先级任务
- [ ] 选择最高优先级的Issue
- [ ] 使用 `mydev-github-workflow` 启动开发
- [ ] 执行TDD流程

#### 下午：验证与记录
- [ ] 完成本周启动的任务
- [ ] 更新状态跟踪文件
- [ ] 输出：周报与下周计划

---

## 三、问题单/需求单模板

### 3.1 功能需求单 (Feature Request)

```markdown
## 功能描述
[描述需要实现的功能]

## 业务场景
[描述功能对应的业务场景]

## 验收标准
- [ ] 标准1
- [ ] 标准2

## 测试用例
- 单元测试: [测试用例描述]
- 集成测试: [测试用例描述]
- E2E测试: [测试用例描述]

## 优先级
- [ ] P0 - 阻塞发布
- [ ] P1 - 本周必须完成
- [ ] P2 - 本迭代完成
- [ ] P3 - 可延后

## 模块标签
- [ ] backend-order
- [ ] backend-payment
- [ ] backend-group
- [ ] backend-admin
- [ ] frontend-page
- [ ] e2e-flow
```

### 3.2 问题单 (Bug Report)

```markdown
## 问题描述
[描述发现的问题]

## 复现步骤
1. 步骤1
2. 步骤2
3. 步骤3

## 预期行为
[描述预期行为]

## 实际行为
[描述实际行为]

## 影响范围
- [ ] 阻塞核心功能
- [ ] 影响用户体验
- [ ] 仅影响边缘场景

## 测试覆盖
- [ ] 需要补充单元测试
- [ ] 需要补充集成测试
- [ ] 需要补充E2E测试
```

---

## 四、状态跟踪机制

### 4.1 状态文件结构

```
.trae/skills/mydev-github-workflow/scripts/
├── workflow_state.json      # 工作流状态
├── test_cases_state.json    # 测试用例状态
└── mvp_progress.json        # MVP进度跟踪
```

### 4.2 MVP进度跟踪文件格式

```json
{
  "last_updated": "2026-03-23T10:00:00Z",
  "summary": {
    "total_issues": 0,
    "open_issues": 0,
    "in_progress": 0,
    "completed": 0
  },
  "modules": {
    "backend-order": {
      "status": "analyzing",
      "issues": [],
      "test_coverage": {
        "unit": 0,
        "integration": 0,
        "e2e": 0
      }
    },
    "backend-payment": {
      "status": "pending",
      "issues": [],
      "test_coverage": {
        "unit": 0,
        "integration": 0,
        "e2e": 0
      }
    }
  },
  "daily_progress": [
    {
      "date": "2026-03-23",
      "tasks_completed": [],
      "issues_created": [],
      "issues_resolved": [],
      "notes": ""
    }
  ]
}
```

---

## 五、工作流执行指南

### 5.1 触发条件
使用 `mydev-github-workflow` skill 的触发条件：
1. 报告bug: "登录失败"、"返回401"、"有个错误"
2. 请求功能: "添加功能"、"实现一个"、"新增"
3. 代码改进: "优化"、"重构"、"改进"

### 5.2 执行流程
```
Step 0: 初始化检查
   ↓
Step 1: 问题解析
   ↓
Step 2-3: 分支创建
   ↓
Step 4: 代码分析
   ↓
Step 5-7: TDD流程 (Red → Green → Refactor)
   ↓
Step 8: 本地验证
   ↓
Step 9-12: CI监控 + PR + E2E
   ↓
Step 14-15: 合并 + 部署
   ↓
Step 16: 清理输出
```

### 5.3 硬约束
1. 禁止忽略 current_fix_cases 失败
2. 禁止跳过状态跟踪写入
3. 禁止跳过本地验证
4. 禁止跳过CI监控
5. 禁止在CI验证通过前合并PR
6. 涉及代码逻辑修改必须走TDD流程

---

## 六、预期产出

### 6.1 文档产出
1. `mvp_test_gap_analysis.md` - 测试用例差距分析报告
2. `mvp_module_analysis.md` - 模块分析报告
3. `mvp_weekly_report.md` - 周报

### 6.2 代码产出
1. 补充的单元测试文件
2. 补充的集成测试文件
3. 补充的E2E测试文件
4. 修复的代码问题

### 6.3 Issue产出
1. 功能需求Issues (Feature)
2. 问题修复Issues (Bug)
3. 技术债务Issues (Tech Debt)

---

## 七、成功指标

| 指标 | 目标值 | 验证方式 |
|-----|-------|---------|
| 测试覆盖率提升 | +15% | Codecov报告 |
| 高优先级Issue解决率 | 80% | GitHub Issue统计 |
| E2E核心流程覆盖 | 3个核心流程 | Playwright报告 |
| CI通过率 | 100% | GitHub Actions |

---

## 八、风险与应对

| 风险 | 影响 | 应对措施 |
|-----|-----|---------|
| 发现大量未实现功能 | 延期 | 优先处理P0/P1问题 |
| CI环境问题 | 阻塞 | 提前验证CI配置 |
| 测试数据准备困难 | 延期 | 使用seed脚本准备数据 |
| 依赖外部服务 | 阻塞 | 使用Mock服务 |

---

**创建者**: AI Assistant
**审核者**: 待定
**状态**: 待确认
