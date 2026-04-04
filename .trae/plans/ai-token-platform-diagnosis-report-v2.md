# 拼托托AI Token聚合平台诊断分析报告 v2.0

> 创建时间：2026-04-04
> 版本说明：本版本为v1.x评审优化版，解决了结构混乱、内容重复、GAP编号不一致等问题

---

## 评审说明

### v1.x版本评审发现的问题

| 评审维度 | 发现问题 | 优化措施 |
|---------|---------|---------|
| 文档结构 | 章节编号混乱（第9/10/11章内容重叠） | 重新组织章节结构 |
| 内容重复 | 商户对接流程、配置清单多处重复 | 合并去重 |
| GAP编号 | 编号不连续（G1-G8后跳到GAP-10），缺少GAP-9 | 统一编号体系 |
| GAP覆盖 | 部分GAP未落入开发计划 | 建立完整映射关系 |
| 视角缺失 | 缺少三类角色的业务流程串联 | v3.0补充 |

---

## 一、执行摘要

### 1.1 项目定位

拼托托定位为AI Token聚合电商平台，核心商业模式：
- **用户侧**：提供统一API入口，用户购买套餐后调用大模型服务
- **商户侧**：大模型供应商入驻，提供API Key，成为供应商池成员
- **平台侧**：智能路由、计费结算、高可用保障

### 1.2 核心发现

| 维度 | 当前状态 | 匹配度 | 关键问题 |
|------|---------|--------|---------|
| SPU/SKU架构 | 已实现基础架构 | 70% | 缺少场景化分类、用途标签 |
| 智能路由 | 简单负载均衡 | 30% | 无健康监控、无故障切换、无成本优化 |
| 商户接入 | 基础流程已实现 | 50% | 无API Key验证、无健康检查 |
| 计费系统 | 硬编码定价 | 40% | 无动态定价、无商户成本结算 |
| 用户产品体验 | 基础电商流程 | 60% | 缺少用途导向的产品展示 |

---

## 二、当前实现深度分析

### 2.1 SPU/SKU架构分析

#### SPU与SKU的关系

```
SPU（标准产品单元）= 模型产品
├── 例：GPT-4 Turbo
│   ├── model_provider: openai
│   ├── model_name: gpt-4-turbo
│   └── 用户浏览时看到的是SPU（产品列表）
│
└── 1个SPU对应多个SKU
    │
    ▼
SKU（库存量单位）= 具体套餐
├── 例1：GPT-4 Turbo 100万Token包
│   ├── sku_type: token_pack
│   └── retail_price: 99元
├── 例2：GPT-4 Turbo 月度订阅
│   ├── sku_type: subscription
│   └── retail_price: 199元/月
└── 例3：GPT-4 Turbo 试用套餐
        ├── sku_type: trial
        └── retail_price: 0元
```

#### 与需求对比

| 需求 | 当前实现 | GAP编号 |
|------|---------|---------|
| 按用途分类（对话/PDF/图片/音视频/多模态） | 仅model_tier | GAP-03 |
| 免费套餐 | trial类型存在 | 需完善 |
| Token计费套餐 | token_pack类型存在 | ✅ |
| 周期性订阅套餐 | subscription类型存在 | GAP-08 |
| 商户极简接入 | merchant_skus表存在 | GAP-16~23 |

### 2.2 智能路由分析

#### 当前实现

```go
// 简单负载均衡：按剩余配额排序
SELECT ... FROM merchant_api_keys
WHERE provider = $1 AND status = 'active'
ORDER BY (quota_limit - quota_used) DESC
LIMIT 1
```

#### 与业界最佳实践对比

| 特性 | OpenRouter | OneAPI | 当前项目 | GAP编号 |
|------|-----------|--------|---------|---------|
| 价格优先路由 | ✅ | ✅ | ❌ | GAP-01 |
| 健康检查 | ✅ 30秒窗口 | ✅ | ❌ | GAP-06 |
| 故障自动切换 | ✅ | ✅ | ❌ | GAP-01 |
| 成本优化路由 | ✅ | ✅ | ❌ | GAP-01 |
| 延迟感知路由 | ✅ | ✅ | ❌ | GAP-01 |

### 2.3 计费系统分析

#### 当前问题

```go
// 硬编码定价，无商户成本概念
func calculateTokenCost(provider, model string, ...) float64 {
    switch provider {
    case "openai":
        inputRate = 0.01 / 1000    // 硬编码用户价格
    }
}
```

| 问题 | GAP编号 |
|------|---------|
| 定价硬编码 | GAP-02 |
| 无商户成本概念 | GAP-04 |
| 无套餐内计费 | GAP-05 |

---

## 三、GAP总览与优先级

### 3.1 完整GAP清单

| 编号 | GAP描述 | 影响 | 优先级 | 落入开发计划 |
|------|---------|------|--------|-------------|
| GAP-01 | 智能路由缺失健康检查和故障切换 | 高可用无法保障 | P0 | ✅ Phase 1 |
| GAP-02 | 计费引擎硬编码，无动态定价 | 无法灵活定价和结算 | P0 | ✅ Phase 2 |
| GAP-03 | 无用途场景分类 | 用户选择困难 | P1 | ✅ Phase 3 |
| GAP-04 | 无商户成本结算机制 | 无法精准结算 | P1 | ✅ Phase 4 |
| GAP-05 | 无API调用与用户套餐关联 | 计费逻辑不完整 | P1 | ✅ Phase 2 |
| GAP-06 | 无实时监控告警 | 运维盲区 | P2 | ✅ Phase 5 |
| GAP-07 | 无缓存优化机制 | 成本浪费 | P2 | ❌ 未纳入 |
| GAP-08 | 无自动续费扣款 | 订阅体验差 | P2 | ❌ 未纳入 |
| GAP-09 | api_usage_logs表结构不完整 | 无法精准记录成本 | P0 | ✅ Phase 2 |
| GAP-10 | 商户成本定价字段缺失 | 无法计算平台利润 | P0 | ✅ Phase 2 |
| GAP-11 | SPU缺少用途场景标签 | 用户无法按用途筛选 | P1 | ✅ Phase 3 |
| GAP-12 | SKU缺少商户成本定价字段 | 无法计算平台利润 | P0 | ✅ Phase 2 |
| GAP-13 | 缺少SPU-SKU批量创建工具 | 运营效率低 | P2 | ❌ 未纳入 |
| GAP-14 | 缺少产品上架审核流程 | 无质量控制 | P1 | ❌ 未纳入 |
| GAP-15 | SPU/SKU缺少审核日志 | 无法追溯变更 | P2 | ❌ 未纳入 |
| GAP-16 | merchant_api_keys缺少关键字段 | 无法配置端点/模型/成本 | P0 | ✅ Phase 2 |
| GAP-17 | 商户上架时无法配置成本定价 | 无法计算平台利润 | P0 | ✅ Phase 2 |
| GAP-18 | 登录响应缺少商户状态信息 | 商户体验差 | P2 | ❌ 未纳入 |
| GAP-19 | 无API Key有效性验证 | 无效Key直接上架 | P0 | ✅ Phase 1 |
| GAP-20 | 无网络可达性测试 | 网络不通的Key被使用 | P0 | ✅ Phase 1 |
| GAP-21 | 无模型列表自动获取 | 商户手动配置易出错 | P1 | ✅ Phase 1 |
| GAP-22 | 无成本定价验证 | 可能出现负利润 | P0 | ✅ Phase 2 |
| GAP-23 | 无健康状态初始化 | 新Key状态未知 | P1 | ✅ Phase 1 |

### 3.2 GAP覆盖统计

| 状态 | 数量 | 占比 |
|------|------|------|
| 已纳入开发计划 | 17 | 74% |
| 未纳入开发计划 | 6 | 26% |
| **总计** | **23** | **100%** |

### 3.3 未纳入开发计划的GAP

| GAP | 描述 | 建议 |
|-----|------|------|
| GAP-07 | 无缓存优化机制 | 后续版本迭代 |
| GAP-08 | 无自动续费扣款 | 后续版本迭代 |
| GAP-13 | 缺少批量创建工具 | 运营工具单独规划 |
| GAP-14 | 缺少产品审核流程 | 流程优化单独规划 |
| GAP-15 | 缺少审核日志 | 流程优化单独规划 |
| GAP-18 | 登录响应缺少商户状态 | 快速修复项 |

---

## 四、详细开发计划

### Phase 1：智能路由核心能力（P0，预计2周）

**解决GAP**：GAP-01, GAP-06, GAP-19, GAP-20, GAP-21, GAP-23

#### 1.1 健康检查服务

```go
type HealthChecker struct {
    checkInterval   time.Duration
    providers       map[string]*ProviderHealth
}

type ProviderHealth struct {
    ProviderID      int
    Status          string  // healthy, degraded, unhealthy
    LastCheck       time.Time
    ConsecutiveFail int
    AvgLatency      time.Duration
}
```

#### 1.2 智能路由引擎

```go
type RoutingStrategy string
const (
    StrategyPrice      RoutingStrategy = "price"
    StrategyThroughput RoutingStrategy = "throughput"
    StrategyLatency    RoutingStrategy = "latency"
)

func (r *SmartRouter) SelectProvider(req *APIProxyRequest) (*RoutingCandidate, error) {
    // 1. 获取候选Provider
    // 2. 过滤不健康的Provider
    // 3. 计算各维度分数
    // 4. 按综合分数排序
}
```

#### 1.3 API Key验证服务

```go
type ValidationResult struct {
    Success         bool
    NetworkLatency  int
    SupportedModels []ModelInfo
    Pricing         map[string]PricingInfo
    HealthStatus    string
}

func (v *APIKeyValidator) ValidateAPIKey(provider, apiKey, apiBaseURL string) (*ValidationResult, error)
```

---

### Phase 2：计费引擎重构（P0，预计1.5周）

**解决GAP**：GAP-02, GAP-04, GAP-05, GAP-09, GAP-10, GAP-12, GAP-16, GAP-17, GAP-22

#### 2.1 数据库增强

```sql
-- merchant_api_keys表增强
ALTER TABLE merchant_api_keys ADD COLUMN api_base_url VARCHAR(500);
ALTER TABLE merchant_api_keys ADD COLUMN supported_models JSONB;
ALTER TABLE merchant_api_keys ADD COLUMN input_price_per_1k DECIMAL(10,6);
ALTER TABLE merchant_api_keys ADD COLUMN output_price_per_1k DECIMAL(10,6);
ALTER TABLE merchant_api_keys ADD COLUMN health_status VARCHAR(20);

-- merchant_skus表增强
ALTER TABLE merchant_skus ADD COLUMN cost_input_rate DECIMAL(10,6);
ALTER TABLE merchant_skus ADD COLUMN cost_output_rate DECIMAL(10,6);

-- api_usage_logs表增强
ALTER TABLE api_usage_logs ADD COLUMN merchant_id INT;
ALTER TABLE api_usage_logs ADD COLUMN user_cost DECIMAL(15,6);
ALTER TABLE api_usage_logs ADD COLUMN merchant_cost DECIMAL(15,6);
ALTER TABLE api_usage_logs ADD COLUMN platform_profit DECIMAL(15,6);
```

#### 2.2 动态定价服务

```go
type CostResult struct {
    UserCost       float64
    MerchantCost   float64
    PlatformProfit float64
}

func (s *PricingService) CalculateCost(merchantAPIKeyID int, provider, model string, 
                                        inputTokens, outputTokens int) *CostResult
```

---

### Phase 3：用途场景分类（P1，预计0.5周）

**解决GAP**：GAP-03, GAP-11

#### 3.1 场景分类表

```sql
CREATE TABLE usage_scenarios (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT
);

INSERT INTO usage_scenarios (code, name) VALUES
('chat', '日常对话'),
('pdf', 'PDF处理'),
('image', '图片处理'),
('audio', '音频处理'),
('video', '视频处理'),
('multimodal', '多模态');
```

---

### Phase 4：商户结算系统（P1，预计1周）

**解决GAP**：GAP-04

#### 4.1 结算数据模型

```sql
CREATE TABLE merchant_settlement_items (
    id SERIAL PRIMARY KEY,
    settlement_id INT,
    api_usage_log_id INT,
    user_cost DECIMAL(15,6),
    merchant_cost DECIMAL(15,6),
    platform_profit DECIMAL(15,6)
);

CREATE TABLE merchant_accounts (
    id SERIAL PRIMARY KEY,
    merchant_id INT UNIQUE,
    balance DECIMAL(15,2) DEFAULT 0,
    pending_balance DECIMAL(15,2) DEFAULT 0
);
```

---

### Phase 5：监控告警系统（P2，预计1周）

**解决GAP**：GAP-06

#### 5.1 监控指标

```go
type ProviderMetrics struct {
    AvailabilityRate float64
    AvgLatency       time.Duration
    Throughput       float64
    CostPer1kTokens  float64
}
```

---

## 五、实施路线图

### 5.1 总体时间规划

```
Week 1-2:  Phase 1 - 智能路由核心能力
Week 3-4:  Phase 2 - 计费引擎重构
Week 5:    Phase 3 - 用途场景分类
Week 6-7:  Phase 4 - 商户结算系统
Week 8:    Phase 5 - 监控告警系统
Week 9:    集成测试与优化
```

### 5.2 里程碑

| 里程碑 | 时间 | 交付物 |
|--------|------|--------|
| M1 | Week 2 | 智能路由上线，支持健康检查和故障切换 |
| M2 | Week 4 | 动态计费上线，支持商户成本结算 |
| M3 | Week 5 | 场景分类上线，用户体验优化 |
| M4 | Week 7 | 商户结算上线，完整商业闭环 |
| M5 | Week 8 | 监控告警上线，运维能力完善 |

---

## 六、风险评估

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| 路由切换延迟 | 用户体验下降 | 预热连接池、异步健康检查 |
| 定价数据不一致 | 计费错误 | 事务保证、定期对账 |
| 商户API不稳定 | 服务不可用 | 多商户备份、熔断机制 |
| 并发结算压力 | 系统性能下降 | 批量处理、异步队列 |

---

## 七、验收标准

### 功能验收

- [ ] 智能路由支持价格/吞吐量/延迟三种策略
- [ ] Provider故障30秒内自动切换
- [ ] 定价数据从数据库读取，支持热更新
- [ ] 用户购买套餐后API调用正确扣费
- [ ] 商户结算单准确反映实际服务量
- [ ] 场景分类正确展示，用户可按用途筛选
- [ ] API Key验证通过后才可上架

### 性能验收

- [ ] 路由决策延迟 < 10ms
- [ ] 健康检查不影响正常请求
- [ ] 计费服务缓存命中率 > 95%
- [ ] 结算批处理支持万级订单

---

## 八、商户对接完整流程

### 8.1 商户生命周期

```
阶段1: 用户注册/登录
    ↓
阶段2: 商户入驻申请（提交企业资料）
    ↓
阶段3: 平台审核（管理员审批）
    ↓
阶段4: 商户配置
    ├── 添加API Key
    ├── 一键验证（网络可达性、Key有效性、模型列表、成本定价）
    └── 配置成本定价
    ↓
阶段5: SKU上架（选择平台SKU + 绑定API Key）
    ↓
阶段6: 进入供应商资源池（可被智能路由选择）
```

### 8.2 商户配置内容清单

| 配置项 | 存储位置 | 当前状态 | 配置方 |
|--------|---------|---------|--------|
| 商户信息 | merchants表 | ✅ 已实现 | 商户 |
| API Key | merchant_api_keys.api_key_encrypted | ✅ 已实现 | 商户注入 |
| API端点 | merchant_api_keys.api_base_url | ❌ 缺失 | 商户配置 |
| 支持模型 | merchant_api_keys.supported_models | ❌ 缺失 | 自动获取 |
| 输入Token成本 | merchant_api_keys.input_price_per_1k | ❌ 缺失 | 商户配置 |
| 输出Token成本 | merchant_api_keys.output_price_per_1k | ❌ 缺失 | 商户配置 |
| SKU关联 | merchant_skus表 | ✅ 已实现 | 商户选择 |
| SKU成本定价 | merchant_skus.cost_input/output_rate | ❌ 缺失 | 商户配置 |

### 8.3 极简对接三步走

```
第1步: 输入API Key
├── 选择Provider
├── 输入API Key
└── 点击"一键验证"

第2步: 确认自动填充的信息
├── 验证通过
├── 自动获取支持的模型列表
├── 自动获取官方定价
├── 计算利润空间
└── 商户确认上架

第3步: 完成
└── 进入供应商资源池
```

---

## 九、用户下单与API调用流程

### 9.1 核心理念

**用户下单-平台履约，商户透明**

### 9.2 完整流程

```
阶段0: 平台产品上架（运营端）
├── 创建SPU（模型产品）
└── 创建SKU（具体套餐）
    ↓
阶段1: 用户浏览和下单
├── 浏览SPU列表（产品列表）
├── 选择SKU（具体套餐）
└── 创建订单（无商户信息）
    ↓
阶段2: 平台履约
└── 充值Token/算力点/订阅 到用户账户（不涉及商户）
    ↓
阶段3: 用户使用权益（调用API）
├── 智能路由选择商户API Key
└── 调用商户API
    ↓
阶段4: 记录与结算
├── api_usage_logs记录实际服务商户
├── 计算用户成本、商户成本、平台利润
└── 商户结算
```

### 9.3 关键验证

| 环节 | 当前实现 | 状态 |
|------|---------|------|
| 订单不记录商户 | ✅ orders表无merchant_id | 正确 |
| 履约不涉及商户 | ✅ 只充值用户账户 | 正确 |
| API调用记录商户 | ⚠️ 只有key_id | 需增强 |
| 商户成本计算 | ❌ 未实现 | GAP-04 |

---

**报告版本**: v2.0
**最后更新**: 2026-04-04
**评审优化**: 解决结构混乱、内容重复、GAP编号不一致问题
