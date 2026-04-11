# 拼托托AI Token聚合平台开发计划 v1.0

> 创建时间：2026-04-04
> 基于诊断报告：ai-token-platform-diagnosis-report-v4.md
> 工作流规范：mydev-github-workflow

---

## 一、开发计划总览

### 1.1 项目范围

基于v4.0诊断报告，本次开发计划覆盖以下核心能力：

| 模块 | 优先级 | 预计工期 | 核心目标 |
|------|--------|---------|---------|
| 智能路由核心 | P0 | 2周 | 健康检查、故障切换、动态定价 |
| 商户API Key管理 | P0 | 1周 | 验证、成本配置、健康状态 |
| 结算系统 | P1 | 1.5周 | 月结生成、商户确认、财务审核 |
| 用户体验优化 | P1 | 1周 | 场景筛选、性能排序 |

### 1.2 开发原则

遵循 `mydev-github-workflow` 规范：

```
✅ Phase 1: PR验证阶段 (PR分支)
   └─ 开发 → 测试 → CI链路验证 → 通过/重试

✅ Phase 2: 合并部署阶段
   └─ 合并 → 部署

[1] 禁止忽略 current_fix_cases 失败
[2] 禁止跳过状态跟踪写入
[3] 禁止跳过本地验证
[4] 禁止跳过CI监控
[5] 禁止在CI验证通过前合并PR
[6] 涉及代码逻辑修改必须执行 TDD 流程
[7] 禁止直接 push 到 main 分支
```

---

## 二、Phase分解

### Phase 1: 智能路由核心能力（P0，2周）

**目标**：实现健康检查、智能路由、动态定价、故障切换

#### Phase 1.1: 数据库迁移（Week 1, Day 1-2）

| 任务ID | 任务描述 | 文件 | 状态 |
|--------|---------|------|------|
| T1.1.1 | 创建用户偏好设置迁移 | `migrations/024_user_preferences.sql` | ⬜ |
| T1.1.2 | 创建SPU场景分类迁移 | `migrations/025_spu_scenarios.sql` | ⬜ |
| T1.1.3 | 创建智能路由核心迁移 | `migrations/027_smart_routing.sql` | ⬜ |
| T1.1.4 | 创建商户API Key验证迁移 | `migrations/029_merchant_verification.sql` | ⬜ |
| T1.1.5 | 创建结算系统增强迁移 | `migrations/030_settlement_enhancement.sql` | ⬜ |
| T1.1.6 | 创建商户资料增强迁移 | `migrations/031_merchant_profile_enhancement.sql` | ⬜ |
| T1.1.7 | 创建商户SKU成本定价迁移 | `migrations/032_merchant_sku_cost.sql` | ⬜ |
| T1.1.8 | 插入场景种子数据 | `migrations/025_spu_scenarios.sql` | ⬜ |

**验收标准**：
- [ ] 所有迁移文件创建完成
- [ ] 本地数据库迁移执行成功
- [ ] 种子数据插入正确

---

#### Phase 1.2: 健康检查服务（Week 1, Day 3-5）

| 任务ID | 任务描述 | 文件 | 状态 |
|--------|---------|------|------|
| T1.2.1 | 创建HealthChecker结构体 | `services/health_checker.go` | ⬜ |
| T1.2.2 | 实现多级检查频率逻辑 | `services/health_checker.go` | ⬜ |
| T1.2.3 | 实现Provider健康检查 | `services/health_checker.go` | ⬜ |
| T1.2.4 | 实现健康状态记录 | `services/health_checker.go` | ⬜ |
| T1.2.5 | 创建健康检查API端点 | `handlers/health.go` | ⬜ |
| T1.2.6 | 编写健康检查单元测试 | `services/health_checker_test.go` | ⬜ |

**验收标准**：
- [ ] 健康检查服务启动正常
- [ ] 多级频率配置生效
- [ ] 健康状态正确记录到数据库
- [ ] 单元测试覆盖率 > 80%

---

#### Phase 1.3: 智能路由服务（Week 2, Day 1-3）

| 任务ID | 任务描述 | 文件 | 状态 |
|--------|---------|------|------|
| T1.3.1 | 创建SmartRouter结构体 | `services/smart_router.go` | ⬜ |
| T1.3.2 | 实现候选Provider获取 | `services/smart_router.go` | ⬜ |
| T1.3.3 | 实现多维度评分计算 | `services/smart_router.go` | ⬜ |
| T1.3.4 | 实现故障切换逻辑 | `services/smart_router.go` | ⬜ |
| T1.3.5 | 实现路由策略配置 | `services/smart_router.go` | ⬜ |
| T1.3.6 | 编写智能路由单元测试 | `services/smart_router_test.go` | ⬜ |

**验收标准**：
- [ ] 路由决策延迟 < 10ms
- [ ] 故障切换正常工作
- [ ] 支持价格/延迟/均衡三种策略
- [ ] 单元测试覆盖率 > 80%

---

#### Phase 1.4: 动态定价服务（Week 2, Day 4-5）

| 任务ID | 任务描述 | 文件 | 状态 |
|--------|---------|------|------|
| T1.4.1 | 创建PricingService结构体 | `services/pricing_service.go` | ⬜ |
| T1.4.2 | 实现数据库定价读取 | `services/pricing_service.go` | ⬜ |
| T1.4.3 | 实现定价缓存机制 | `services/pricing_service.go` | ⬜ |
| T1.4.4 | 实现成本计算逻辑 | `services/pricing_service.go` | ⬜ |
| T1.4.5 | 重构api_proxy.go计费逻辑 | `handlers/api_proxy.go` | ⬜ |
| T1.4.6 | 编写定价服务单元测试 | `services/pricing_service_test.go` | ⬜ |

**验收标准**：
- [ ] 定价数据从数据库读取
- [ ] 缓存命中率 > 95%
- [ ] 成本计算准确
- [ ] 单元测试覆盖率 > 80%

---

### Phase 2: 商户API Key管理（P0，1周）

**目标**：实现API Key验证、成本配置、健康状态展示

#### Phase 2.1: API Key验证服务（Week 3, Day 1-3）

| 任务ID | 任务描述 | 文件 | 状态 |
|--------|---------|------|------|
| T2.1.1 | 创建APIKeyValidator结构体 | `services/api_key_validator.go` | ⬜ |
| T2.1.2 | 实现连接验证 | `services/api_key_validator.go` | ⬜ |
| T2.1.3 | 实现模型列表获取 | `services/api_key_validator.go` | ⬜ |
| T2.1.4 | 实现异步验证流程 | `services/api_key_validator.go` | ⬜ |
| T2.1.5 | 创建验证API端点 | `handlers/merchant_apikey.go` | ⬜ |
| T2.1.6 | 编写验证服务单元测试 | `services/api_key_validator_test.go` | ⬜ |

**验收标准**：
- [ ] 异步验证正常工作
- [ ] 验证结果正确存储
- [ ] 验证失败禁止路由
- [ ] 单元测试覆盖率 > 80%

---

#### Phase 2.2: 商户API Key前端（Week 3, Day 4-5）

| 任务ID | 任务描述 | 文件 | 状态 |
|--------|---------|------|------|
| T2.2.1 | 更新MerchantAPIKey类型定义 | `types/index.ts` | ⬜ |
| T2.2.2 | 添加API端点URL输入 | `MerchantAPIKeys.tsx` | ⬜ |
| T2.2.3 | 添加成本定价配置表单 | `MerchantAPIKeys.tsx` | ⬜ |
| T2.2.4 | 添加健康状态展示 | `MerchantAPIKeys.tsx` | ⬜ |
| T2.2.5 | 添加一键验证按钮 | `MerchantAPIKeys.tsx` | ⬜ |
| T2.2.6 | 添加健康检查频率选择 | `MerchantAPIKeys.tsx` | ⬜ |

**验收标准**：
- [ ] 表单提交正常
- [ ] 验证按钮触发异步验证
- [ ] 健康状态正确展示
- [ ] TypeScript类型检查通过

---

### Phase 3: 结算系统（P1，1.5周）

**目标**：实现月结生成、商户确认、财务审核

#### Phase 3.1: 结算服务（Week 4, Day 1-3）

| 任务ID | 任务描述 | 文件 | 状态 |
|--------|---------|------|------|
| T3.1.1 | 创建SettlementService结构体 | `services/settlement_service.go` | ⬜ |
| T3.1.2 | 实现月结生成逻辑 | `services/settlement_service.go` | ⬜ |
| T3.1.3 | 实现商户确认逻辑 | `services/settlement_service.go` | ⬜ |
| T3.1.4 | 实现财务审核逻辑 | `services/settlement_service.go` | ⬜ |
| T3.1.5 | 创建结算API端点 | `handlers/settlement.go` | ⬜ |
| T3.1.6 | 编写结算服务单元测试 | `services/settlement_service_test.go` | ⬜ |

**验收标准**：
- [ ] 月结单正确生成
- [ ] 商户确认流程正常
- [ ] 财务审核流程正常
- [ ] 单元测试覆盖率 > 80%

---

#### Phase 3.2: 商户结算前端（Week 4, Day 4-5）

| 任务ID | 任务描述 | 文件 | 状态 |
|--------|---------|------|------|
| T3.2.1 | 更新Settlement类型定义 | `types/index.ts` | ⬜ |
| T3.2.2 | 实现结算列表展示 | `MerchantSettlements.tsx` | ⬜ |
| T3.2.3 | 实现结算详情弹窗 | `MerchantSettlements.tsx` | ⬜ |
| T3.2.4 | 实现确认按钮功能 | `MerchantSettlements.tsx` | ⬜ |

**验收标准**：
- [ ] 结算列表正确展示
- [ ] 确认按钮正常工作
- [ ] TypeScript类型检查通过

---

#### Phase 3.3: 管理员结算管理前端（Week 5, Day 1-2）

| 任务ID | 任务描述 | 文件 | 状态 |
|--------|---------|------|------|
| T3.3.1 | 创建AdminSettlements页面 | `AdminSettlements.tsx` | ⬜ |
| T3.3.2 | 实现审核按钮功能 | `AdminSettlements.tsx` | ⬜ |
| T3.3.3 | 实现打款标记功能 | `AdminSettlements.tsx` | ⬜ |

**验收标准**：
- [ ] 审核流程正常
- [ ] 打款标记正常
- [ ] TypeScript类型检查通过

---

### Phase 4: 用户体验优化（P1，1周）

**目标**：实现场景筛选、性能排序

#### Phase 4.1: 场景分类后端（Week 5, Day 3-4）

| 任务ID | 任务描述 | 文件 | 状态 |
|--------|---------|------|------|
| T4.1.1 | 创建场景列表API | `handlers/catalog.go` | ⬜ |
| T4.1.2 | 创建场景产品筛选API | `handlers/catalog.go` | ⬜ |
| T4.1.3 | 更新SPU列表查询逻辑 | `handlers/sku.go` | ⬜ |

**验收标准**：
- [ ] 场景列表API正常
- [ ] 场景筛选API正常
- [ ] 性能排序API正常

---

#### Phase 4.2: 场景分类前端（Week 5, Day 5）

| 任务ID | 任务描述 | 文件 | 状态 |
|--------|---------|------|------|
| T4.2.1 | 创建场景导航组件 | `components/ScenarioFilter.tsx` | ⬜ |
| T4.2.2 | 更新产品列表页面 | `ProductListPage.tsx` | ⬜ |
| T4.2.3 | 添加性能排序功能 | `ProductListPage.tsx` | ⬜ |

**验收标准**：
- [ ] 场景导航正常
- [ ] 筛选功能正常
- [ ] 排序功能正常

---

## 三、任务执行流程

每个任务执行时遵循 `mydev-github-workflow` 规范：

```
Step 0: 初始化检查
   ↓
Step 1: 问题解析
   ↓
Step 2-3: 分支创建
   ↓
Step 4: 代码分析
   ↓
Step 5-7: TDD流程
   ↓
Step 8: 本地验证
   ↓
Step 9: Push + 创建 PR
   ↓
Step 10: CI监控
   ↓
Step 11: 错误修复（如需要）
   ↓
Step 12: 合并 PR
   ↓
Step 13: 部署监控
   ↓
Step 14: 清理输出
```

---

## 四、验收标准总表

### 4.1 功能验收

| 模块 | 验收项 | 预期结果 |
|------|--------|---------|
| 智能路由 | 路由策略 | 支持价格/延迟/均衡三种策略 |
| 智能路由 | 故障切换 | Provider故障30秒内自动切换 |
| 动态定价 | 定价来源 | 从数据库读取，支持热更新 |
| API Key验证 | 验证流程 | 异步验证，失败禁止路由 |
| 结算系统 | 月结生成 | 每月1日自动生成上月结算单 |
| 结算系统 | 确认流程 | 商户确认 → 财务审核 → 打款 |
| 场景分类 | 筛选功能 | 用户可按场景筛选产品 |

### 4.2 性能验收

| 指标 | 目标值 |
|------|--------|
| 路由决策延迟 | < 10ms |
| 健康检查影响 | 不影响正常请求 |
| 定价缓存命中率 | > 95% |
| 结算批处理 | 支持万级订单 |

### 4.3 质量验收

| 指标 | 目标值 |
|------|--------|
| 后端单元测试覆盖率 | > 80% |
| 前端TypeScript检查 | 0 errors |
| CI通过率 | 100% |

---

## 五、风险与依赖

### 5.1 技术风险

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| 路由切换延迟 | 用户体验下降 | 预热连接池、异步健康检查 |
| 定价数据不一致 | 计费错误 | 事务保证、定期对账 |
| 商户API不稳定 | 服务不可用 | 多商户备份、熔断机制 |
| 并发结算压力 | 系统性能下降 | 批量处理、异步队列 |

### 5.2 依赖关系

```
Phase 1 (智能路由)
├── Phase 2 (商户API Key) - 依赖健康检查服务
└── Phase 3 (结算系统) - 依赖动态定价服务

Phase 4 (用户体验) - 独立实现
```

---

## 六、里程碑

| 里程碑 | 时间 | 交付物 |
|--------|------|--------|
| M1 | Week 2 | 智能路由上线，支持健康检查和故障切换 |
| M2 | Week 3 | 商户API Key管理上线，支持验证和成本配置 |
| M3 | Week 5 | 结算系统上线，完整商业闭环 |
| M4 | Week 5 | 场景分类上线，用户体验优化 |

---

**文档版本**: v1.1
**最后更新**: 2026-04-10
**状态**: ⏳ 待审核

---

## 附录 A：内部经济与单一账本（2026-04-10 补充）

> 技术口径与原则索引：[backend/doc_internal_token_economics.md](../../backend/doc_internal_token_economics.md)（`docs/` 目录在仓库中 gitignore，本地可保留副本）

本附录补齐总计划中**未单独展开**的「零售余额 / 扣费 / 价目版本」链路，与 Phase 1.4 动态定价、api_proxy 重构**强相关但范围不同**：Phase 1.4 侧重定价服务与代理接入 DB；本附录侧重**账本语义统一**与**订单快照 + pricing_version** 全链路。

### A.1 目标原则（已拍板）

| 原则 | 说明 |
|------|------|
| 单一账本 | 用户可消费余额只对应一条主账本；代理扣费与履约入账对齐同一内部单位 |
| 订单快照入账 1:1 | 订单（快照）写多少，履约加多少；入账阶段不乘「元/K token」 |
| 换算层扣费 | 按次用价目版本中的元/1K（或等价）将 usage 换为应扣内部单位，再扣同一账本 |
| 价目与下单时点一致 | 订单或权益绑定 `pricing_version_id`，避免 live 改价导致买用不一致 |

### A.2 建议任务清单（可与 Phase 1.4 并行或前置评审）

| 任务ID | 任务描述 | 依赖 | 建议落点 |
|--------|----------|------|----------|
| IE-1 | 评审：`pricing_versions`（或等价）表结构与订单/权益外键 | 无 | `migrations/`、设计文档 |
| IE-2 | 订单履约：快照字段 1:1 入账单一账本（收敛 `compute_points` SKU 与 `token_pack`） | IE-1 草案 | `services/fulfillment_service.go` |
| IE-2a | 新建订单绑定 `pricing_version_id`（当前为 **baseline**） | IE-1 已合入 | `handlers/order_and_group.go`、`services/pricing_version.go` |
| IE-3 | 数据迁移：历史 `compute_point_accounts` → 主账本策略与对账 | IE-2 | `migrations/`、一次性脚本 |
| IE-4 | `api_proxy`：解析调用 → 权益/订单 → `pricing_version` → 扣减内部单位 | IE-1、IE-2 | `handlers/api_proxy.go`、`services/pricing_service.go` |
| IE-5 | ~~清理：`token_pack` 上无效必填 `compute_points`~~（已处理：后端校验 + 管理端表单） | IE-2 | `handlers/sku.go`、`AdminSKUs.tsx` |
| IE-6 | 验收：入账/扣费/报表单位一致性测试与对账用例 | IE-2–IE-4 | 集成测试、运维 Runbook |

### A.3 启动条件检查

- [ ] IE-1 数据模型评审通过（含老订单无 `pricing_version` 的默认策略）
- [ ] IE-2 / IE-4 技术方案与 Phase 1.4 边界对齐（避免重复或遗漏）
- [ ] IE-3 迁移方案与回滚路径书面确认

未勾选前，仍可启动 Phase 1.4 中「定价从 DB 读取、缓存」等子任务；**全量上线「单一账本 + 版本扣费」**建议以上条件满足后再合入生产。

### A.4 进展速记

- **IE-1**：已合入 `main`（迁移 045）。
- **IE-2a**：下单写入 `pricing_version_id`（无 baseline 行时保持 `NULL`，兼容未迁移库）。
- **IE-2 / IE-3**：履约与算力点 API 已用 `tokens`；046 一次性合并历史 `compute_point_accounts`（IE-3 与迁移合并交付）。
- **IE-4**：`api_proxy` 已按 **最近履约订单** `pricing_version_id` 解析 `pricing_version_spu_rates`；无版本或快照缺模型时回退 `PricingService`（live SPU）。
- **IE-5**：`token_pack` 创建校验与管理端表单中 **`compute_points` 已非必填**；履约仍以 `token_amount` 为准。
