# 拼托托AI Token聚合平台诊断分析报告

> 创建时间：2026-04-04
> 分析范围：项目代码实现、业界最佳实践对比、商业模式匹配度

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
| 商户接入 | 基础流程已实现 | 50% | 无自动配额管理、无健康检查 |
| 计费系统 | 硬编码定价 | 40% | 无动态定价、无商户成本结算 |
| 用户产品体验 | 基础电商流程 | 60% | 缺少用途导向的产品展示 |

---

## 二、当前实现深度分析

### 2.1 SPU/SKU架构分析

#### 当前实现

```
SPU（标准产品单元）
├── 模型基础信息（model_provider, model_name, model_tier）
├── 厂商对接字段（provider_model_id, provider_api_endpoint）
├── 计费配置（base_compute_points, billing_coefficient）
└── 适配器配置（input_length_ranges, billing_adapter, routing_rules）

SKU（库存量单位）
├── 关联SPU（spu_id）
├── 销售属性（sku_type, token_amount, compute_points）
├── 定价（retail_price, wholesale_price, original_price）
├── 限制配置（tpm_limit, rpm_limit, valid_days）
└── 拼团配置（group_enabled, group_discount_rate）
```

#### 与需求对比

| 需求 | 当前实现 | GAP |
|------|---------|-----|
| 按用途分类（对话/PDF/图片/音视频/多模态） | 仅model_tier（pro/lite/mini/vision） | 缺少用途标签、场景分类 |
| 免费套餐 | trial类型存在 | 实现不完整，无免费额度管理 |
| Token计费套餐 | token_pack类型存在 | 已实现 |
| 周期性订阅套餐 | subscription类型存在 | 已实现，缺少自动续费扣款 |
| 商户极简接入 | merchant_skus表存在 | 缺少一键上架、自动配置 |

### 2.2 智能路由分析

#### 当前实现（api_proxy.go）

```go
func selectAPIKeyForRequest(db *sql.DB, req APIProxyRequest, apiKey *models.MerchantAPIKey) error {
    // 1. 指定APIKeyID时直接使用
    // 2. 指定MerchantSKUID时关联查询
    // 3. 默认：按剩余配额排序取第一个
    return db.QueryRow(`
        SELECT ... FROM merchant_api_keys
        WHERE provider = $1 AND status = 'active'
        ORDER BY (quota_limit - quota_used) DESC
        LIMIT 1
    `, req.Provider).Scan(...)
}
```

#### 与业界最佳实践对比

| 特性 | OpenRouter | OneAPI | 当前项目 | GAP |
|------|-----------|--------|---------|-----|
| 价格优先路由 | ✅ 默认策略 | ✅ 支持 | ❌ | 缺失 |
| 吞吐量优先路由 | ✅ :nitro后缀 | ✅ 支持 | ❌ | 缺失 |
| 健康检查 | ✅ 30秒窗口 | ✅ 支持 | ❌ | 缺失 |
| 故障自动切换 | ✅ 自动fallback | ✅ 支持 | ❌ | 缺失 |
| 成本优化路由 | ✅ :floor后缀 | ✅ 支持 | ❌ | 缺失 |
| 延迟感知路由 | ✅ 支持 | ✅ 支持 | ❌ | 缺失 |
| 多Provider负载均衡 | ✅ 支持 | ✅ 支持 | ⚠️ 简单实现 | 需增强 |

#### OpenRouter路由策略参考

```
1. 过滤掉过去30秒内有重大故障的Provider
2. 在剩余候选中按价格排序
3. 支持动态变体：
   - :online → 优先有网络访问的版本
   - :nitro → 优先吞吐量
   - :floor → 最低成本
```

### 2.3 商户接入流程分析

#### 当前流程

```
1. 商户注册 → 提交审核资料
2. 管理员审核 → 通过/拒绝
3. 商户添加API Key → 加密存储
4. 商户选择SKU上架 → 绑定API Key
5. 用户购买 → 平台路由 → 调用商户API
```

#### 与需求对比

| 需求 | 当前实现 | GAP |
|------|---------|-----|
| 极简接入（勾选SKU+绑定APIKey） | ✅ 已实现 | - |
| 平台定义SKU，商户勾选上架 | ✅ 已实现 | - |
| 商户透明（用户看不到商户） | ⚠️ 部分实现 | 需要隐藏商户信息 |
| 实时健康监控 | ❌ 未实现 | 需要健康检查服务 |
| 故障自动切换 | ❌ 未实现 | 需要路由层支持 |
| 精准成本结算 | ⚠️ 部分实现 | 需要完善计费引擎 |

### 2.4 计费系统分析

#### 当前实现（api_proxy.go:321-364）

```go
func calculateTokenCost(provider, model string, inputTokens, outputTokens int) float64 {
    switch provider {
    case "openai":
        switch {
        case strings.Contains(model, "gpt-4-turbo"):
            inputRate = 0.01 / 1000    // 硬编码
            outputRate = 0.03 / 1000
        // ...
        }
    case "anthropic":
        // 硬编码定价...
    }
}
```

#### 问题分析

| 问题 | 影响 | 严重程度 |
|------|------|---------|
| 定价硬编码 | 无法动态调价，需重新部署 | 高 |
| 无商户成本概念 | 无法计算平台利润 | 高 |
| 无SPU定价关联 | 定价与SPU配置脱节 | 中 |
| 无套餐内计费 | 用户购买套餐后如何计费不清晰 | 高 |

### 2.5 订单履约分析

#### 当前实现（fulfillment_service.go）

```go
func (s *FulfillmentService) FulfillOrder(tx *sql.Tx, orderID int) error {
    // 根据SKU类型执行不同履约逻辑
    switch st {
    case skuTypeTokenPack:     // 充值Token
    case skuTypeComputePoints: // 充值算力点
    case skuTypeSubscription:  // 创建订阅
    case skuTypeTrial:         // 试用
    case skuTypeConcurrent:    // 并发配额
    }
}
```

#### 与需求对比

| 需求 | 当前实现 | GAP |
|------|---------|-----|
| 订单支付后自动履约 | ✅ 已实现 | - |
| Token包充值 | ✅ 已实现 | - |
| 订阅周期管理 | ✅ 已实现 | - |
| API调用扣费 | ⚠️ 简单实现 | 需关联用户套餐 |

---

## 三、业界最佳实践对标

### 3.1 OpenRouter架构参考

#### 核心设计理念

1. **统一API入口**
   - 单一端点 `/api/v1/chat/completions`
   - 完全兼容OpenAI SDK
   - 通过model字段区分不同模型

2. **智能路由策略**
   - 价格优先（默认）
   - 吞吐量优先（:nitro）
   - 成本优先（:floor）
   - 在线能力优先（:online）

3. **高可用保障**
   - 30秒健康检查窗口
   - 自动故障切换
   - 多Provider备份

4. **计费透明**
   - 实时Token计数
   - 精确到每次请求的成本
   - 缓存命中折扣

### 3.2 OneAPI路由算法参考

#### 动态权重计算

```
模型权重 = α × 成本权重 + β × 性能权重 + γ × 质量权重 + δ × 配额权重
```

- α、β、γ、δ 可根据业务需求动态调整
- 实时更新权重，确保路由决策最优

### 3.3 电商SPU/SKU最佳实践（拼多多/美团）

#### 拼团模式设计

```
1. 低价引流SKU → 吸引用户进入
2. 主力SKU → 正常利润
3. 高端SKU → 品牌形象
4. 活动专享SKU → 限时限量
```

#### SKU布局策略

| 策略 | 说明 | 当前项目应用 |
|------|------|-------------|
| 低价引流 | 极低价格吸引用户 | trial类型可承担 |
| 活动卡位 | 限量SKU触发火爆标签 | flash_sales已实现 |
| 拼团裂变 | 邀请返利 | group_enabled已实现 |
| 组合销售 | 多SKU打包 | 未实现 |

---

## 四、GAP分析与优先级

### 4.1 关键GAP清单

| ID | GAP描述 | 影响 | 优先级 | 复杂度 |
|----|---------|------|--------|--------|
| G1 | 智能路由缺失健康检查和故障切换 | 高可用无法保障 | P0 | 高 |
| G2 | 计费引擎硬编码，无动态定价 | 无法灵活定价和结算 | P0 | 中 |
| G3 | 无用途场景分类（对话/PDF/多模态） | 用户选择困难 | P1 | 低 |
| G4 | 无商户成本结算机制 | 无法精准结算 | P1 | 中 |
| G5 | 无API调用与用户套餐关联 | 计费逻辑不完整 | P1 | 高 |
| G6 | 无实时监控告警 | 运维盲区 | P2 | 中 |
| G7 | 无缓存优化机制 | 成本浪费 | P2 | 中 |
| G8 | 无自动续费扣款 | 订阅体验差 | P2 | 低 |

### 4.2 GAP依赖关系

```
G1 智能路由
├── G6 实时监控（依赖）
└── G4 商户结算（关联）

G2 计费引擎
├── G4 商户结算（依赖）
└── G5 套餐关联（关联）

G3 用途分类
└── 独立实现
```

---

## 五、详细开发计划

### Phase 1：智能路由核心能力（P0，预计2周）

#### 1.1 健康检查服务

**目标**：实时监控商户API健康状态

**实现方案**：
```go
// services/health_checker.go
type HealthChecker struct {
    checkInterval   time.Duration     // 检查间隔（30秒）
    providers       map[string]*ProviderHealth
    alertThreshold  int               // 连续失败阈值
}

type ProviderHealth struct {
    ProviderID      int
    Status          string  // healthy, degraded, unhealthy
    LastCheck       time.Time
    ConsecutiveFail int
    AvgLatency      time.Duration
    ErrorRate       float64
}
```

**数据库设计**：
```sql
CREATE TABLE provider_health_status (
    id SERIAL PRIMARY KEY,
    provider_id INT REFERENCES model_providers(id),
    merchant_api_key_id INT REFERENCES merchant_api_keys(id),
    status VARCHAR(20) NOT NULL,  -- healthy, degraded, unhealthy
    last_check_at TIMESTAMP,
    last_success_at TIMESTAMP,
    consecutive_failures INT DEFAULT 0,
    avg_latency_ms INT,
    error_rate DECIMAL(5,4),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**API端点**：
- `GET /api/v1/admin/providers/health` - 获取所有Provider健康状态
- `POST /api/v1/admin/providers/:id/health-check` - 手动触发健康检查

#### 1.2 智能路由引擎

**目标**：基于多维度权重的智能路由

**实现方案**：
```go
// services/smart_router.go
type SmartRouter struct {
    healthChecker *HealthChecker
    pricingService *PricingService
    strategy      RoutingStrategy
}

type RoutingStrategy string
const (
    StrategyPrice      RoutingStrategy = "price"       // 价格优先
    StrategyThroughput RoutingStrategy = "throughput"  // 吞吐量优先
    StrategyLatency    RoutingStrategy = "latency"     // 延迟优先
    StrategyCost       RoutingStrategy = "cost"        // 成本优先
)

type RoutingCandidate struct {
    ProviderID    int
    APIKeyID      int
    MerchantID    int
    HealthScore   float64  // 健康分数 0-1
    PriceScore    float64  // 价格分数
    LatencyScore  float64  // 延迟分数
    QuotaScore    float64  // 配额分数
    FinalScore    float64  // 综合分数
}

func (r *SmartRouter) SelectProvider(req *APIProxyRequest) (*RoutingCandidate, error) {
    // 1. 获取所有候选Provider
    candidates := r.getCandidates(req.Provider)
    
    // 2. 过滤不健康的Provider
    candidates = r.filterUnhealthy(candidates)
    
    // 3. 计算各维度分数
    for i := range candidates {
        candidates[i].HealthScore = r.calcHealthScore(candidates[i])
        candidates[i].PriceScore = r.calcPriceScore(candidates[i])
        candidates[i].LatencyScore = r.calcLatencyScore(candidates[i])
        candidates[i].QuotaScore = r.calcQuotaScore(candidates[i])
        
        // 综合分数计算
        candidates[i].FinalScore = 
            0.3 * candidates[i].HealthScore +
            0.3 * candidates[i].PriceScore +
            0.2 * candidates[i].LatencyScore +
            0.2 * candidates[i].QuotaScore
    }
    
    // 4. 按综合分数排序，选择最优
    sort.Slice(candidates, func(i, j int) bool {
        return candidates[i].FinalScore > candidates[j].FinalScore
    })
    
    return candidates[0], nil
}
```

#### 1.3 故障自动切换

**目标**：Provider故障时自动切换到备用

**实现方案**：
```go
func (r *SmartRouter) ExecuteWithFallback(req *APIProxyRequest, maxRetries int) (*APIResponse, error) {
    candidates := r.SelectProviders(req.Provider) // 获取排序列表
    
    var lastErr error
    for i := 0; i < min(maxRetries, len(candidates)); i++ {
        candidate := candidates[i]
        
        resp, err := r.executeRequest(req, candidate)
        if err == nil {
            // 成功，更新健康状态
            r.healthChecker.RecordSuccess(candidate.APIKeyID)
            return resp, nil
        }
        
        // 失败，记录并尝试下一个
        r.healthChecker.RecordFailure(candidate.APIKeyID, err)
        lastErr = err
    }
    
    return nil, fmt.Errorf("all providers failed: %w", lastErr)
}
```

---

### Phase 2：计费引擎重构（P0，预计1.5周）

#### 2.1 动态定价服务

**目标**：从数据库读取定价，支持动态更新

**数据库设计**：
```sql
-- SPU定价表
ALTER TABLE spus ADD COLUMN IF NOT EXISTS input_price_per_1k DECIMAL(10,6);
ALTER TABLE spus ADD COLUMN IF NOT EXISTS output_price_per_1k DECIMAL(10,6);
ALTER TABLE spus ADD COLUMN IF NOT EXISTS pricing_updated_at TIMESTAMP;

-- SKU成本定价（商户侧）
ALTER TABLE skus ADD COLUMN IF NOT EXISTS cost_input_rate DECIMAL(10,6);
ALTER TABLE skus ADD COLUMN IF NOT EXISTS cost_output_rate DECIMAL(10,6);
```

**服务实现**：
```go
// services/pricing_service.go
type PricingService struct {
    db         *sql.DB
    cache      map[int]*PricingData
    cacheMutex sync.RWMutex
    cacheTTL   time.Duration
}

type PricingData struct {
    SKUID            int
    SPUID            int
    MerchantID       *int
    
    // 用户侧定价
    RetailPrice      float64
    InputRate        float64  // 元/1K tokens
    OutputRate       float64
    
    // 商户侧成本
    CostInputRate    float64
    CostOutputRate   float64
    
    // 计费模式
    BillingType      string   // token_pack, subscription, payg
}

func (s *PricingService) CalculateCost(skuID int, inputTokens, outputTokens int) (*CostResult, error) {
    pricing := s.GetPricing(skuID)
    
    userCost := pricing.InputRate * float64(inputTokens) / 1000 +
                pricing.OutputRate * float64(outputTokens) / 1000
    
    platformCost := pricing.CostInputRate * float64(inputTokens) / 1000 +
                    pricing.CostOutputRate * float64(outputTokens) / 1000
    
    return &CostResult{
        UserCost:     userCost,
        PlatformCost: platformCost,
        PlatformProfit: userCost - platformCost,
    }, nil
}
```

#### 2.2 用户套餐关联计费

**目标**：API调用时关联用户购买的套餐

**数据模型**：
```sql
-- 用户资产表
CREATE TABLE user_assets (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    sku_id INT NOT NULL REFERENCES skus(id),
    order_id INT REFERENCES orders(id),
    
    -- Token包
    remaining_tokens BIGINT,
    
    -- 算力点
    remaining_compute_points DECIMAL(15,2),
    
    -- 订阅
    subscription_end_date DATE,
    
    -- 状态
    status VARCHAR(20) NOT NULL,  -- active, exhausted, expired
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**计费逻辑**：
```go
func (s *BillingEngine) DeductFromUserAsset(userID int, cost float64, tokens int64) error {
    // 1. 查询用户有效资产（按过期时间排序）
    assets := s.getUserAssets(userID)
    
    // 2. 按优先级扣费
    for _, asset := range assets {
        if asset.RemainingTokens >= tokens {
            // 从Token包扣除
            s.deductTokens(asset.ID, tokens)
            return nil
        } else if asset.RemainingComputePoints >= cost {
            // 从算力点扣除
            s.deductComputePoints(asset.ID, cost)
            return nil
        }
    }
    
    // 3. 无有效资产，从余额扣除
    return s.deductBalance(userID, cost)
}
```

---

### Phase 3：用途场景分类（P1，预计0.5周）

#### 3.1 SPU场景标签

**数据库设计**：
```sql
-- 场景分类表
CREATE TABLE usage_scenarios (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,  -- chat, pdf, image, audio, video, multimodal
    name VARCHAR(100) NOT NULL,
    description TEXT,
    icon_url VARCHAR(500),
    sort_order INT DEFAULT 0,
    status VARCHAR(20) DEFAULT 'active'
);

-- SPU-场景关联表
CREATE TABLE spu_scenarios (
    spu_id INT REFERENCES spus(id) ON DELETE CASCADE,
    scenario_id INT REFERENCES usage_scenarios(id) ON DELETE CASCADE,
    is_primary BOOLEAN DEFAULT FALSE,  -- 主要场景
    PRIMARY KEY (spu_id, scenario_id)
);
```

**种子数据**：
```sql
INSERT INTO usage_scenarios (code, name, description, sort_order) VALUES
('chat', '日常对话', '适用于日常问答、聊天互动', 1),
('pdf', 'PDF处理', '适用于PDF文档解析、摘要、问答', 2),
('image', '图片处理', '适用于图像理解、OCR、图像生成', 3),
('audio', '音频处理', '适用于语音识别、语音合成', 4),
('video', '视频处理', '适用于视频理解、视频生成', 5),
('multimodal', '多模态', '支持多种输入输出格式', 6),
('code', '代码生成', '适用于代码编写、调试、解释', 7),
('reasoning', '复杂推理', '适用于数学推理、逻辑分析', 8);
```

#### 3.2 前端场景导航

**API设计**：
```
GET /api/v1/catalog/scenarios - 获取所有场景分类
GET /api/v1/catalog/scenarios/:code/products - 获取场景下的产品
```

---

### Phase 4：商户结算系统（P1，预计1周）

#### 4.1 结算数据模型

```sql
-- 商户结算明细表
CREATE TABLE merchant_settlement_items (
    id SERIAL PRIMARY KEY,
    settlement_id INT REFERENCES merchant_settlements(id),
    api_usage_log_id INT REFERENCES api_usage_logs(id),
    
    -- 调用信息
    request_id VARCHAR(100),
    model VARCHAR(100),
    input_tokens INT,
    output_tokens INT,
    
    -- 费用信息
    user_cost DECIMAL(15,6),       -- 用户支付
    merchant_cost DECIMAL(15,6),   -- 商户成本
    platform_profit DECIMAL(15,6), -- 平台利润
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 商户账户表
CREATE TABLE merchant_accounts (
    id SERIAL PRIMARY KEY,
    merchant_id INT UNIQUE REFERENCES merchants(id),
    balance DECIMAL(15,2) DEFAULT 0,        -- 可提现余额
    pending_balance DECIMAL(15,2) DEFAULT 0, -- 待结算金额
    total_earned DECIMAL(15,2) DEFAULT 0,   -- 累计收入
    total_settled DECIMAL(15,2) DEFAULT 0,  -- 累计已结算
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 4.2 结算服务

```go
// services/settlement_service.go
type SettlementService struct {
    db *sql.DB
}

func (s *SettlementService) GenerateMerchantSettlement(merchantID int, periodStart, periodEnd time.Time) error {
    // 1. 汇总期间API调用
    items := s.aggregateAPIUsage(merchantID, periodStart, periodEnd)
    
    // 2. 计算各项费用
    var totalMerchantCost, totalPlatformProfit float64
    for _, item := range items {
        totalMerchantCost += item.MerchantCost
        totalPlatformProfit += item.PlatformProfit
    }
    
    // 3. 创建结算单
    settlement := &MerchantSettlement{
        MerchantID:       merchantID,
        PeriodStart:      periodStart,
        PeriodEnd:        periodEnd,
        TotalMerchantCost: totalMerchantCost,
        PlatformProfit:   totalPlatformProfit,
        Status:           "pending",
    }
    
    return s.saveSettlement(settlement, items)
}
```

---

### Phase 5：监控告警系统（P2，预计1周）

#### 5.1 监控指标

```go
// metrics/provider_metrics.go
type ProviderMetrics struct {
    // 可用性指标
    AvailabilityRate   float64  // 可用率
    MTTR              time.Duration  // 平均恢复时间
    MTBF              time.Duration  // 平均故障间隔
    
    // 性能指标
    AvgLatency        time.Duration
    P99Latency        time.Duration
    Throughput        float64  // QPS
    
    // 成本指标
    CostPer1kTokens   float64
    DailyCost         float64
}
```

#### 5.2 告警规则

```yaml
# config/alert_rules.yaml
rules:
  - name: provider_unhealthy
    condition: consecutive_failures >= 3
    severity: critical
    actions:
      - type: email
        recipients: [ops@example.com]
      - type: webhook
        url: https://hooks.slack.com/xxx
        
  - name: high_latency
    condition: avg_latency_ms > 5000
    severity: warning
    actions:
      - type: email
        
  - name: cost_anomaly
    condition: daily_cost > baseline * 1.5
    severity: warning
```

---

## 六、实施路线图

### 6.1 总体时间规划

```
Week 1-2:  Phase 1 - 智能路由核心能力
Week 3-4:  Phase 2 - 计费引擎重构
Week 5:    Phase 3 - 用途场景分类
Week 6-7:  Phase 4 - 商户结算系统
Week 8:    Phase 5 - 监控告警系统
Week 9:    集成测试与优化
```

### 6.2 里程碑

| 里程碑 | 时间 | 交付物 |
|--------|------|--------|
| M1 | Week 2 | 智能路由上线，支持健康检查和故障切换 |
| M2 | Week 4 | 动态计费上线，支持商户成本结算 |
| M3 | Week 5 | 场景分类上线，用户体验优化 |
| M4 | Week 7 | 商户结算上线，完整商业闭环 |
| M5 | Week 8 | 监控告警上线，运维能力完善 |

---

## 七、风险评估

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| 路由切换延迟 | 用户体验下降 | 预热连接池、异步健康检查 |
| 定价数据不一致 | 计费错误 | 事务保证、定期对账 |
| 商户API不稳定 | 服务不可用 | 多商户备份、熔断机制 |
| 并发结算压力 | 系统性能下降 | 批量处理、异步队列 |

---

## 八、验收标准

### 功能验收

- [ ] 智能路由支持价格/吞吐量/延迟三种策略
- [ ] Provider故障30秒内自动切换
- [ ] 定价数据从数据库读取，支持热更新
- [ ] 用户购买套餐后API调用正确扣费
- [ ] 商户结算单准确反映实际服务量
- [ ] 场景分类正确展示，用户可按用途筛选

### 性能验收

- [ ] 路由决策延迟 < 10ms
- [ ] 健康检查不影响正常请求
- [ ] 计费服务缓存命中率 > 95%
- [ ] 结算批处理支持万级订单

---

## 九、用户下单与API调用流程分析

### 9.1 正确的商业模式理解

**核心理念：用户下单-平台履约，商户透明**

#### SPU与SKU的关系澄清

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         SPU与SKU的关系                                       │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  SPU（标准产品单元）= 模型产品                                               │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ 例：GPT-4 Turbo                                                     │   │
│  │ - model_provider: openai                                            │   │
│  │ - model_name: gpt-4-turbo                                           │   │
│  │ - model_tier: pro                                                   │   │
│  │ - 用户浏览时看到的是SPU（产品列表）                                    │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                              │                                              │
│                              │ 1个SPU对应多个SKU                             │
│                              ▼                                              │
│  SKU（库存量单位）= 具体套餐                                                 │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ 例1：GPT-4 Turbo 100万Token包                                        │   │
│  │ - sku_type: token_pack                                              │   │
│  │ - token_amount: 1000000                                             │   │
│  │ - retail_price: 99元                                                │   │
│  │                                                                      │   │
│  │ 例2：GPT-4 Turbo 月度订阅                                            │   │
│  │ - sku_type: subscription                                            │   │
│  │ - subscription_period: monthly                                       │   │
│  │ - retail_price: 199元/月                                            │   │
│  │                                                                      │   │
│  │ 例3：GPT-4 Turbo 试用套餐                                            │   │
│  │ - sku_type: trial                                                   │   │
│  │ - trial_duration_days: 7                                            │   │
│  │ - retail_price: 0元                                                 │   │
│  │                                                                      │   │
│  │ - 用户下单时选择的是SKU（具体套餐）                                    │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

#### 完整业务流程

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        用户下单-平台履约模式                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  阶段0: 平台产品上架（运营端）                                               │
│  ┌─────────────┐    ┌─────────────┐                                        │
│  │ 创建SPU     │───>│ 创建SKU     │                                        │
│  │ (模型产品)  │    │ (具体套餐)  │                                        │
│  └─────────────┘    └─────────────┘                                        │
│         │                  │                                                │
│         │                  │                                                │
│         ▼                  ▼                                                │
│  ┌─────────────────────────────────────────────────────────────┐           │
│  │ SPU: GPT-4 Turbo, Claude-3, 文心一言...                     │           │
│  │ SKU: Token包、订阅、试用、并发...                            │           │
│  └─────────────────────────────────────────────────────────────┘           │
│                              │                                              │
│                              ▼                                              │
│  阶段1: 用户浏览和下单                                                       │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐                     │
│  │ 浏览SPU列表 │───>│ 选择SKU     │───>│ 创建订单    │                     │
│  │ (产品列表)  │    │ (具体套餐)  │    │ (无商户信息)│                     │
│  └─────────────┘    └─────────────┘    └─────────────┘                     │
│                              │                                              │
│                              ▼                                              │
│  阶段2: 平台履约                                                            │
│  ┌─────────────────────────────────────────────────────────────┐           │
│  │ 充值Token/算力点/订阅 到用户账户（不涉及商户）                 │           │
│  └─────────────────────────────────────────────────────────────┘           │
│                              │                                              │
│                              ▼                                              │
│  阶段3: 用户使用权益（调用API）                                              │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐                     │
│  │ 用户调用API │───>│ 智能路由    │───>│ 选择商户    │                     │
│  │             │    │             │    │ API Key     │                     │
│  └─────────────┘    └─────────────┘    └─────────────┘                     │
│                              │                                              │
│                              ▼                                              │
│  阶段4: 记录与结算                                                          │
│  ┌─────────────────────────────────────────────────────────────┐           │
│  │ api_usage_logs记录:                                         │           │
│  │ - user_id (用户)                                            │           │
│  │ - merchant_id (实际服务商户)                                 │           │
│  │ - user_cost (用户成本)                                       │           │
│  │ - merchant_cost (商户成本)                                   │           │
│  │ - platform_profit (平台利润 = user_cost - merchant_cost)    │           │
│  └─────────────────────────────────────────────────────────────┘           │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 9.2 平台产品上架流程分析

#### 当前实现

```go
// handlers/sku.go:190-231 - 管理员创建SPU
func CreateSPU(c *gin.Context) {
    // 只有管理员可以创建SPU
    if !ensureAdmin(c) { return }
    
    var req models.SPUCreateRequest
    // 创建模型产品
    err := db.QueryRow(
        `INSERT INTO spus (spu_code, name, model_provider, model_name, model_version, 
                          model_tier, context_window, base_compute_points, description, 
                          thumbnail_url, status, sort_order) 
         VALUES (...) RETURNING ...`,
    ).Scan(...)
}

// handlers/sku.go:649-733 - 管理员创建SKU
func CreateSKU(c *gin.Context) {
    // 只有管理员可以创建SKU
    if !ensureAdmin(c) { return }
    
    var req models.SKUCreateRequest
    // 创建具体套餐
    err = db.QueryRow(
        `INSERT INTO skus (spu_id, sku_code, sku_type, token_amount, compute_points, 
                          subscription_period, retail_price, ...) 
         VALUES (...) RETURNING ...`,
    ).Scan(...)
}
```

#### 验证结果

| 环节 | 当前实现 | 是否正确 |
|------|---------|---------|
| SPU由平台运营创建 | ✅ CreateSPU只有管理员可调用 | 正确 |
| SKU由平台运营创建 | ✅ CreateSKU只有管理员可调用 | 正确 |
| 用户浏览SPU列表 | ✅ ListPublicSKUs返回SKU+SPU信息 | 正确 |
| 用户选择SKU下单 | ✅ CreateOrder传入sku_id | 正确 |
| 订单记录sku_id和spu_id | ✅ orders表包含两个字段 | 正确 |

#### 平台产品上架流程GAP

| GAP | 问题描述 | 影响 |
|-----|---------|------|
| GAP-12 | SPU缺少用途场景标签 | 用户无法按用途（对话/PDF/图片）筛选产品 |
| GAP-13 | SKU缺少商户成本定价 | 无法计算平台利润 |
| GAP-14 | 缺少SPU-SKU批量创建工具 | 运营效率低 |
| GAP-15 | 缺少产品上架审核流程 | 无质量控制 |

#### SPU/SKU创建与审核流程详细分析

**当前实现**：

```go
// handlers/sku.go:190-231 - SPU创建
func CreateSPU(c *gin.Context) {
    // 1. 只有管理员可调用
    if !ensureAdmin(c) { return }
    
    // 2. 直接创建，状态默认为active
    if req.Status == "" {
        req.Status = "active"  // 直接激活，无审核
    }
    
    // 3. 插入数据库
    db.QueryRow(`INSERT INTO spus (...) VALUES (...) RETURNING ...`)
}

// handlers/sku.go:649-733 - SKU创建
func CreateSKU(c *gin.Context) {
    // 1. 只有管理员可调用
    if !ensureAdmin(c) { return }
    
    // 2. 直接创建，状态默认为active
    if req.Status == "" {
        req.Status = "active"  // 直接激活，无审核
    }
    
    // 3. 插入数据库
    db.QueryRow(`INSERT INTO skus (...) VALUES (...) RETURNING ...`)
}
```

**SPU/SKU数据模型**：

```go
// models/sku.go - SPU模型
type SPU struct {
    ID               int       `json:"id"`
    SPUCode          string    `json:"spu_code"`
    Name             string    `json:"name"`
    ModelProvider    string    `json:"model_provider"`
    ModelName        string    `json:"model_name"`
    ModelTier        string    `json:"model_tier"`
    BaseComputePoints float64  `json:"base_compute_points"`
    Status           string    `json:"status"`  // active, inactive
    // 缺失: reviewed_at, review_note, created_by
}

// models/sku.go - SKU模型
type SKU struct {
    ID            int       `json:"id"`
    SPUID         int       `json:"spu_id"`
    SKUCode       string    `json:"sku_code"`
    SKUType       string    `json:"sku_type"`
    RetailPrice   float64   `json:"retail_price"`
    Status        string    `json:"status"`  // active, inactive
    // 缺失: reviewed_at, review_note, created_by
}
```

**问题分析**：

| 问题 | 当前状态 | 理想状态 | GAP |
|------|---------|---------|-----|
| 创建权限 | ✅ 仅管理员 | 仅管理员 | 无 |
| 创建后状态 | 直接active | draft → review → active | **GAP-19** |
| 审核流程 | ❌ 无 | 多级审核 | **GAP-19** |
| 审核记录 | ❌ 无 | 审核日志 | **GAP-20** |
| 创建人记录 | ❌ 无 | created_by字段 | **GAP-21** |
| 修改历史 | ❌ 无 | 版本历史表 | **GAP-22** |

**GAP-19: 缺少SPU/SKU审核流程**

当前流程：
```
管理员创建 → 直接上架（active）
```

理想流程：
```
运营创建 → draft状态
    ↓
提交审核 → pending_review状态
    ↓
管理员审核 → active/rejected状态
    ↓
上架销售
```

**需要增强的数据结构**：

```sql
-- SPU表增强
ALTER TABLE spus ADD COLUMN IF NOT EXISTS created_by INT REFERENCES users(id);
ALTER TABLE spus ADD COLUMN IF NOT EXISTS reviewed_by INT REFERENCES users(id);
ALTER TABLE spus ADD COLUMN IF NOT EXISTS reviewed_at TIMESTAMP;
ALTER TABLE spus ADD COLUMN IF NOT EXISTS review_note TEXT;
ALTER TABLE spus ADD COLUMN IF NOT EXISTS version INT DEFAULT 1;

COMMENT ON COLUMN spus.created_by IS '创建人ID';
COMMENT ON COLUMN spus.reviewed_by IS '审核人ID';
COMMENT ON COLUMN spus.reviewed_at IS '审核时间';
COMMENT ON COLUMN spus.review_note IS '审核备注';

-- SKU表增强
ALTER TABLE skus ADD COLUMN IF NOT EXISTS created_by INT REFERENCES users(id);
ALTER TABLE skus ADD COLUMN IF NOT EXISTS reviewed_by INT REFERENCES users(id);
ALTER TABLE skus ADD COLUMN IF NOT EXISTS reviewed_at TIMESTAMP;
ALTER TABLE skus ADD COLUMN IF NOT EXISTS review_note TEXT;
ALTER TABLE skus ADD COLUMN IF NOT EXISTS version INT DEFAULT 1;

-- SPU审核日志表
CREATE TABLE IF NOT EXISTS spu_audit_logs (
    id SERIAL PRIMARY KEY,
    spu_id INT NOT NULL REFERENCES spus(id),
    action VARCHAR(50) NOT NULL,  -- create, update, review, status_change
    old_status VARCHAR(20),
    new_status VARCHAR(20),
    operator_id INT REFERENCES users(id),
    note TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- SKU审核日志表
CREATE TABLE IF NOT EXISTS sku_audit_logs (
    id SERIAL PRIMARY KEY,
    sku_id INT NOT NULL REFERENCES skus(id),
    action VARCHAR(50) NOT NULL,
    old_status VARCHAR(20),
    new_status VARCHAR(20),
    operator_id INT REFERENCES users(id),
    note TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**GAP-20: 缺少审核日志**

需要记录：
- 创建记录（谁创建、何时创建）
- 修改记录（修改了什么字段）
- 审核记录（谁审核、审核结果、审核意见）
- 上下架记录（状态变更历史）

**GAP-21: 缺少创建人记录**

当前问题：
- 无法追溯谁创建了SPU/SKU
- 无法区分运营创建和管理员审核

**GAP-22: 缺少版本历史**

当前问题：
- 修改价格后无法追溯历史价格
- 修改配置后无法回滚
- 无法审计变更历史

**建议的审核流程API**：

```
POST /api/v1/admin/spus                    - 创建SPU（draft状态）
POST /api/v1/admin/spus/:id/submit-review  - 提交审核
POST /api/v1/admin/spus/:id/approve        - 审核通过
POST /api/v1/admin/spus/:id/reject         - 审核拒绝
POST /api/v1/admin/spus/:id/activate       - 上架
POST /api/v1/admin/spus/:id/deactivate     - 下架

POST /api/v1/admin/skus                    - 创建SKU（draft状态）
POST /api/v1/admin/skus/:id/submit-review  - 提交审核
POST /api/v1/admin/skus/:id/approve        - 审核通过
POST /api/v1/admin/skus/:id/reject         - 审核拒绝
POST /api/v1/admin/skus/:id/activate       - 上架
POST /api/v1/admin/skus/:id/deactivate     - 下架
```

**SPU/SKU状态流转**：

```
┌─────────────────────────────────────────────────────────────────┐
│                    SPU/SKU状态流转                               │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌────────┐    创建    ┌────────┐    提交    ┌──────────────┐  │
│  │  draft │ ─────────> │ pending│ ─────────> │ under_review │  │
│  └────────┘            │ review │            └──────────────┘  │
│                        └────────┘                    │         │
│                             │                        │         │
│                             │ 撤回                   │ 审核     │
│                             ▼                        ▼         │
│                        ┌────────┐            ┌──────────────┐  │
│                        │  draft │            │    approved  │  │
│                        └────────┘            └──────────────┘  │
│                                                    │           │
│                                                    │ 上架      │
│                                                    ▼           │
│                                              ┌──────────────┐  │
│                                              │    active    │  │
│                                              └──────────────┘  │
│                                                    │           │
│                              ┌─────────────────────┼───────┐   │
│                              │ 下架                │ 拒绝  │   │
│                              ▼                     ▼       │   │
│                        ┌──────────┐          ┌──────────┐  │   │
│                        │ inactive │          │ rejected │  │   │
│                        └──────────┘          └──────────┘  │   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

#### SPU/SKU创建审核流程对比

| 环节 | 当前实现 | 理想实现 | GAP |
|------|---------|---------|-----|
| 创建权限 | ✅ 仅管理员 | 运营可创建 | 可优化 |
| 创建后状态 | 直接active | draft | **GAP-19** |
| 审核流程 | ❌ 无 | 多级审核 | **GAP-19** |
| 审核日志 | ❌ 无 | 完整记录 | **GAP-20** |
| 创建人记录 | ❌ 无 | created_by | **GAP-21** |
| 版本历史 | ❌ 无 | 版本表 | **GAP-22** |
| 批量操作 | ❌ 无 | 批量创建/审核 | **GAP-14** |

### 9.3 当前实现验证

#### 订单流程（正确实现）

```go
// handlers/order_and_group.go:185-189
err = tx.QueryRow(
    `INSERT INTO orders (user_id, product_id, sku_id, spu_id, group_id, quantity, unit_price, total_price, status) 
     VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) 
     RETURNING id, user_id, product_id, sku_id, spu_id, ...`,
    userID, pid, skuID, spuID, groupID, req.Quantity, unitPrice, totalPrice, orderStatusPending,
).Scan(...)
```

**验证结果**：✅ 订单记录不包含merchant_id，符合设计

#### 履约流程（正确实现）

```go
// services/fulfillment_service.go:116-137
func (s *FulfillmentService) fulfillTokenPack(tx *sql.Tx, userID, skuID, orderID, qty int, ...) error {
    // 只给用户充值Token，不涉及商户
    _, err := tx.Exec(`
        INSERT INTO tokens (user_id, balance, total_used, total_earned)
        VALUES ($1, $2, 0, $2)
        ON CONFLICT (user_id) DO UPDATE SET
            balance = tokens.balance + EXCLUDED.balance,
            total_earned = tokens.total_earned + EXCLUDED.balance,
            updated_at = NOW()`,
        userID, add,
    )
}
```

**验证结果**：✅ 履约只充值用户账户，不涉及商户

#### API调用流程（部分正确）

```go
// handlers/api_proxy.go:302-304
_, err = tx.Exec(
    "INSERT INTO api_usage_logs (user_id, key_id, request_id, provider, model, method, path, status_code, latency_ms, input_tokens, output_tokens, cost) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)",
    userIDInt, apiKey.ID, requestID, req.Provider, req.Model, "POST", requestPath, resp.StatusCode, latency, inputTokens, outputTokens, cost,
)
```

**验证结果**：⚠️ 部分正确
- ✅ 记录了key_id（可追溯到商户）
- ❌ 没有记录merchant_id（需要JOIN才能获取）
- ❌ cost字段语义不清（用户成本？商户成本？）
- ❌ 缺少user_cost、merchant_cost、platform_profit分离

### 9.4 关键GAP分析

#### GAP-10: api_usage_logs表结构不完整

**当前结构**：
```sql
CREATE TABLE api_usage_logs (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL,
  key_id INTEGER NOT NULL,        -- 商户API Key ID
  provider VARCHAR(50) NOT NULL,
  model VARCHAR(100) NOT NULL,
  input_tokens INTEGER NOT NULL DEFAULT 0,
  output_tokens INTEGER NOT NULL DEFAULT 0,
  cost DECIMAL(15, 6) NOT NULL DEFAULT 0,  -- 语义不清！
  ...
);
```

**需要增强为**：
```sql
ALTER TABLE api_usage_logs ADD COLUMN IF NOT EXISTS merchant_id INT REFERENCES merchants(id);
ALTER TABLE api_usage_logs ADD COLUMN IF NOT EXISTS user_cost DECIMAL(15,6);      -- 用户成本
ALTER TABLE api_usage_logs ADD COLUMN IF NOT EXISTS merchant_cost DECIMAL(15,6);  -- 商户成本
ALTER TABLE api_usage_logs ADD COLUMN IF NOT EXISTS platform_profit DECIMAL(15,6); -- 平台利润

-- 重命名原cost字段或保持兼容
COMMENT ON COLUMN api_usage_logs.user_cost IS '用户侧成本（元）';
COMMENT ON COLUMN api_usage_logs.merchant_cost IS '商户侧成本（元）';
COMMENT ON COLUMN api_usage_logs.platform_profit IS '平台利润 = user_cost - merchant_cost';
```

#### GAP-11: 商户成本定价缺失

**当前实现**：
```go
// handlers/api_proxy.go:321-364
func calculateTokenCost(provider, model string, inputTokens, outputTokens int) float64 {
    // 硬编码用户侧定价，无商户成本概念
    switch provider {
    case "openai":
        inputRate = 0.01 / 1000    // 用户价格
        outputRate = 0.03 / 1000
    // ...
    }
}
```

**需要实现**：
```go
type CostResult struct {
    UserCost       float64  // 用户成本
    MerchantCost   float64  // 商户成本
    PlatformProfit float64  // 平台利润
}

func (s *PricingService) CalculateCost(merchantAPIKeyID int, provider, model string, inputTokens, outputTokens int) *CostResult {
    // 1. 获取商户定价（从merchant_skus或merchant_api_keys）
    merchantPricing := s.getMerchantPricing(merchantAPIKeyID, provider, model)
    
    // 2. 获取用户定价（从spus或skus）
    userPricing := s.getUserPricing(provider, model)
    
    // 3. 计算成本
    userCost := userPricing.InputRate * float64(inputTokens)/1000 + 
                userPricing.OutputRate * float64(outputTokens)/1000
    
    merchantCost := merchantPricing.InputRate * float64(inputTokens)/1000 + 
                    merchantPricing.OutputRate * float64(outputTokens)/1000
    
    return &CostResult{
        UserCost:       userCost,
        MerchantCost:   merchantCost,
        PlatformProfit: userCost - merchantCost,
    }
}
```

### 9.5 完整业务流程对比

| 环节 | 正确实现 | 当前实现 | GAP |
|------|---------|---------|-----|
| 用户下单 | 订单不记录商户 | ✅ 正确 | 无 |
| 平台履约 | 只充值用户账户 | ✅ 正确 | 无 |
| API调用路由 | 智能选择商户 | ⚠️ 简单实现 | 需增强 |
| API使用记录 | 记录merchant_id, user_cost, merchant_cost | ❌ 只有key_id和cost | **GAP-10** |
| 成本计算 | 分离用户成本和商户成本 | ❌ 只有用户成本 | **GAP-11** |
| 商户结算 | 基于merchant_cost结算 | ❌ 未实现 | **GAP-4** |

---

## 十、商户全流程GAP分析

### 10.1 商户生命周期流程

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           商户完整生命周期                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  阶段1: 用户注册/登录                                                        │
│  ┌─────────────┐    ┌─────────────┐                                        │
│  │ 用户注册    │───>│ 用户登录    │                                        │
│  │ (普通用户)  │    │ 获取Token   │                                        │
│  └─────────────┘    └─────────────┘                                        │
│         │                                                                   │
│         ▼                                                                   │
│  阶段2: 商户入驻申请                                                         │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐                     │
│  │ 申请入驻    │───>│ 提交资料    │───>│ 状态变更    │                     │
│  │ RegisterMer │    │ SubmitDocs  │    │ pending →   │                     │
│  │ chant       │    │             │    │ reviewing   │                     │
│  └─────────────┘    └─────────────┘    └─────────────┘                     │
│         │                  │                                                │
│         │                  │ 提交内容:                                       │
│         │                  │ - company_name (公司名称)                       │
│         │                  │ - business_license_url (营业执照)               │
│         │                  │ - id_card_front/back (身份证)                   │
│         │                  │ - contact_name/phone/email                     │
│         │                  │                                                │
│         ▼                                                                   │
│  阶段3: 平台审核                                                             │
│  ┌─────────────┐    ┌─────────────┐                                        │
│  │ 管理员审核  │───>│ 审核结果    │                                        │
│  │ Approve/    │    │ active /    │                                        │
│  │ Reject      │    │ rejected    │                                        │
│  └─────────────┘    └─────────────┘                                        │
│         │                                                                   │
│         ▼                                                                   │
│  阶段4: 商户配置（上架准备）                                                  │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐                     │
│  │ 添加API Key │───>│ 配置端点    │───>│ 设置定价    │                     │
│  │ (商户注入)  │    │ (缺失!)     │    │ (缺失!)     │                     │
│  └─────────────┘    └─────────────┘    └─────────────┘                     │
│         │                                                                   │
│         ▼                                                                   │
│  阶段5: SKU上架                                                              │
│  ┌─────────────┐    ┌─────────────┐                                        │
│  │ 选择平台SKU │───>│ 绑定API Key │                                        │
│  │ GetAvailable│    │ CreateMerch │                                        │
│  │ SKUs        │    │ antSKU      │                                        │
│  └─────────────┘    └─────────────┘                                        │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 10.2 当前实现验证

#### 商户注册/入驻流程

```go
// handlers/merchant.go:30-104 - 商户注册
func RegisterMerchant(c *gin.Context) {
    // 1. 验证用户登录状态
    userID, exists := c.Get("user_id")
    
    // 2. 检查是否已注册商户
    err := db.QueryRow("SELECT id FROM merchants WHERE user_id = $1", userIDInt).Scan(&existingID)
    
    // 3. 创建商户记录（状态: pending）
    err = db.QueryRow(
        `INSERT INTO merchants (user_id, company_name, business_license, contact_name, 
                               contact_phone, contact_email, address, description, status)
         VALUES (...) RETURNING ...`,
    ).Scan(...)
    
    // 4. 更新用户角色
    db.Exec("UPDATE users SET is_merchant = true, merchant_id = $1, role = 'merchant' WHERE id = $2", ...)
}

// handlers/merchant.go:641-750 - 提交审核资料
func SubmitMerchantDocuments(c *gin.Context) {
    // 提交营业执照、身份证等
    // 状态变更为 'reviewing'
    db.QueryRow(`UPDATE merchants SET ... status = 'reviewing' WHERE user_id = $1`, ...)
}
```

**验证结果**：✅ 商户入驻流程已实现

#### 管理员审核流程

```go
// handlers/admin.go:280-356 - 审核通过
func ApproveMerchant(c *gin.Context) {
    // 状态变更为 'active'
    db.QueryRow(`UPDATE merchants SET status = 'active', reviewed_at = NOW() WHERE id = $1`, ...)
    
    // 记录审核日志
    insertMerchantAuditLog(db, merchant.ID, adminID, "approve", ...)
}

// handlers/admin.go:358-xxx - 审核拒绝
func RejectMerchant(c *gin.Context) {
    // 状态变更为 'rejected'
    db.QueryRow(`UPDATE merchants SET status = 'rejected', review_note = $1 WHERE id = $2`, ...)
}
```

**验证结果**：✅ 审核流程已实现

#### 商户添加API Key流程

```go
// handlers/merchant_apikey.go:20-111 - 创建API Key
func CreateMerchantAPIKey(c *gin.Context) {
    var req struct {
        Name       string  `json:"name" binding:"required"`
        Provider   string  `json:"provider" binding:"required"`  // 只有provider名称
        APIKey     string  `json:"api_key" binding:"required"`
        APISecret  string  `json:"api_secret"`
        QuotaLimit float64 `json:"quota_limit"`
    }
    
    // 加密存储API Key
    apiKeyEncrypted, err := utils.Encrypt(req.APIKey)
    
    // 插入数据库
    db.QueryRow(
        `INSERT INTO merchant_api_keys (merchant_id, name, provider, api_key_encrypted, 
                                        api_secret_encrypted, quota_limit, quota_used, status) 
         VALUES ($1, $2, $3, $4, $5, $6, 0, 'active') ...`,
    )
}
```

**验证结果**：⚠️ 部分实现，**缺失关键字段**

#### 商户SKU上架流程

```go
// handlers/merchant_sku.go:248-403 - 商户选择SKU上架
func CreateMerchantSKU(c *gin.Context) {
    var req models.MerchantSKUCreateRequest  // { sku_id, api_key_id }
    
    // 验证SKU存在
    db.QueryRow("SELECT EXISTS(SELECT 1 FROM skus WHERE id = $1 AND status = 'active')", req.SKUID)
    
    // 验证API Key归属
    db.QueryRow("SELECT EXISTS(SELECT 1 FROM merchant_api_keys WHERE id = $1 AND merchant_id = $2)", ...)
    
    // 创建关联
    db.QueryRow(
        `INSERT INTO merchant_skus (merchant_id, sku_id, api_key_id, status) 
         VALUES ($1, $2, $3, 'active') ...`,
    )
}
```

**验证结果**：✅ 基本流程已实现，**缺失成本定价配置**

### 10.3 关键GAP分析

#### GAP-16: merchant_api_keys表缺失关键字段

**当前表结构**：
```sql
CREATE TABLE merchant_api_keys (
    id SERIAL PRIMARY KEY,
    merchant_id INT,
    name VARCHAR,
    provider VARCHAR,           -- ✅ 有
    api_key_encrypted VARCHAR,  -- ✅ 有
    api_secret_encrypted VARCHAR,
    quota_limit DECIMAL,
    quota_used DECIMAL,
    status VARCHAR
);
```

**缺失的关键字段**：

| 字段 | 用途 | 重要性 |
|------|------|--------|
| `api_base_url` | 自定义API端点（私有部署） | 高 |
| `supported_models` | 支持的模型列表 | 高 |
| `model_mapping` | 模型名称映射 | 中 |
| `input_price_per_1k` | 商户输入Token成本单价 | 高 |
| `output_price_per_1k` | 商户输出Token成本单价 | 高 |
| `health_status` | 健康状态 | 高 |
| `last_health_check` | 最后健康检查时间 | 高 |
| `consecutive_failures` | 连续失败次数 | 中 |

**需要增强为**：
```sql
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS api_base_url VARCHAR(500);
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS supported_models JSONB DEFAULT '[]';
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS model_mapping JSONB DEFAULT '{}';
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS input_price_per_1k DECIMAL(10,6);
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS output_price_per_1k DECIMAL(10,6);
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS health_status VARCHAR(20) DEFAULT 'unknown';
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS last_health_check TIMESTAMP;
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS consecutive_failures INT DEFAULT 0;

COMMENT ON COLUMN merchant_api_keys.api_base_url IS '自定义API端点，如https://api.my-company.com/v1';
COMMENT ON COLUMN merchant_api_keys.supported_models IS '支持的模型列表，如["gpt-4", "gpt-3.5-turbo"]';
COMMENT ON COLUMN merchant_api_keys.model_mapping IS '模型名称映射，如{"my-gpt4": "gpt-4"}';
COMMENT ON COLUMN merchant_api_keys.input_price_per_1k IS '商户成本-输入Token单价(元/1K)';
COMMENT ON COLUMN merchant_api_keys.output_price_per_1k IS '商户成本-输出Token单价(元/1K)';
COMMENT ON COLUMN merchant_api_keys.health_status IS '健康状态: healthy, degraded, unhealthy, unknown';
```

#### GAP-17: 商户上架时无法配置成本定价

**当前CreateMerchantSKU请求**：
```go
type MerchantSKUCreateRequest struct {
    SKUID    int  `json:"sku_id"`
    APIKeyID *int `json:"api_key_id"`
    // 缺失: cost_input_rate, cost_output_rate, priority
}
```

**需要增强为**：
```go
type MerchantSKUCreateRequest struct {
    SKUID           int      `json:"sku_id" binding:"required"`
    APIKeyID        *int     `json:"api_key_id"`
    
    // 新增：商户成本定价
    CostInputRate   *float64 `json:"cost_input_rate"`   // 商户输入Token成本
    CostOutputRate  *float64 `json:"cost_output_rate"`  // 商户输出Token成本
    
    // 新增：路由配置
    Priority        *int     `json:"priority"`          // 路由优先级
    Weight          *int     `json:"weight"`            // 负载均衡权重
}
```

#### GAP-18: 商户登录流程复用用户登录

**当前实现**：商户和普通用户共用登录接口

**问题**：无独立的商户后台登录入口

**建议**：保持现有设计（用户-商户关联），但需在登录后返回商户状态信息

```go
// 登录响应应包含商户信息
type LoginResponse struct {
    Token     string `json:"token"`
    UserID    int    `json:"user_id"`
    Role      string `json:"role"`
    IsMerchant bool   `json:"is_merchant"`
    MerchantID *int   `json:"merchant_id,omitempty"`
    MerchantStatus string `json:"merchant_status,omitempty"` // pending/active/rejected
}
```

### 10.4 商户上架流程对比

| 环节 | 理想实现 | 当前实现 | GAP |
|------|---------|---------|-----|
| 用户注册/登录 | 独立商户后台 | ✅ 复用用户登录 | 可接受 |
| 商户入驻申请 | 提交企业资料 | ✅ 已实现 | 无 |
| 平台审核 | 管理员审批 | ✅ 已实现 | 无 |
| 添加API Key | 配置端点+成本+模型 | ⚠️ 只有provider+key | **GAP-16** |
| SKU上架 | 绑定API Key+成本定价 | ⚠️ 只有绑定 | **GAP-17** |
| 健康监控 | 自动检测API状态 | ❌ 未实现 | **GAP-6** |

### 10.5 商户配置内容清单

| 配置项 | 存储位置 | 当前状态 | 配置方 |
|--------|---------|---------|--------|
| 商户信息 | `merchants`表 | ✅ 已实现 | 商户 |
| API Key | `merchant_api_keys.api_key_encrypted` | ✅ 已实现 | 商户注入 |
| API端点 | `merchant_api_keys.api_base_url` | ❌ 缺失 | 商户配置 |
| 支持模型 | `merchant_api_keys.supported_models` | ❌ 缺失 | 商户配置 |
| 模型映射 | `merchant_api_keys.model_mapping` | ❌ 缺失 | 商户配置 |
| 输入Token成本 | `merchant_api_keys.input_price_per_1k` | ❌ 缺失 | 商户配置 |
| 输出Token成本 | `merchant_api_keys.output_price_per_1k` | ❌ 缺失 | 商户配置 |
| SKU关联 | `merchant_skus`表 | ✅ 已实现 | 商户选择 |
| SKU成本定价 | `merchant_skus.cost_input/output_rate` | ❌ 缺失 | 商户配置 |
| 路由优先级 | `merchant_skus.priority` | ❌ 缺失 | 商户配置 |

### 10.6 商户对接验证流程GAP分析

#### 当前实现问题

**CreateMerchantAPIKey当前流程**：
```go
// handlers/merchant_apikey.go:20-111
func CreateMerchantAPIKey(c *gin.Context) {
    var req struct {
        Name       string  `json:"name" binding:"required"`
        Provider   string  `json:"provider" binding:"required"`
        APIKey     string  `json:"api_key" binding:"required"`
        APISecret  string  `json:"api_secret"`
        QuotaLimit float64 `json:"quota_limit"`
    }
    
    // 1. 验证商户身份
    // 2. 加密存储API Key
    // 3. 直接插入数据库，状态设为active
    // ❌ 没有任何验证！
    
    db.QueryRow(
        `INSERT INTO merchant_api_keys (..., status) VALUES (..., 'active') ...`,
    )
}
```

**关键问题**：
- ❌ 不验证API Key有效性
- ❌ 不验证网络可达性
- ❌ 不获取支持的模型列表
- ❌ 不验证成本定价是否合理
- ❌ 不检查健康状态
- ❌ 直接设为active状态

#### 理想的商户对接验证流程

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                     商户对接验证流程（极简设计）                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  步骤1: 商户输入API Key                                                      │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ 输入: provider, api_key, api_secret(可选), api_base_url(可选)        │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                              │                                              │
│                              ▼                                              │
│  步骤2: 平台自动验证（一键验证）                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ 2.1 网络可达性测试                                                   │   │
│  │     - 检测API端点是否可访问                                          │   │
│  │     - 测试响应时间                                                   │   │
│  │                                                                      │   │
│  │ 2.2 API Key有效性验证                                                │   │
│  │     - 调用Provider的models接口获取可用模型列表                         │   │
│  │     - 验证API Key权限                                                │   │
│  │                                                                      │   │
│  │ 2.3 模型能力探测                                                     │   │
│  │     - 获取支持的模型列表                                             │   │
│  │     - 探测模型能力（对话/图片/音频/视频/多模态）                        │   │
│  │                                                                      │   │
│  │ 2.4 成本定价验证                                                     │   │
│  │     - 获取Provider官方定价                                           │   │
│  │     - 对比平台销售价格                                               │   │
│  │     - 计算利润空间                                                   │   │
│  │                                                                      │   │
│  │ 2.5 健康状态初始化                                                   │   │
│  │     - 设置初始健康状态                                               │   │
│  │     - 记录首次验证结果                                               │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                              │                                              │
│                              ▼                                              │
│  步骤3: 验证结果展示                                                         │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ ✅ 验证通过 → 自动填充支持的模型列表、建议定价                          │   │
│  │ ❌ 验证失败 → 显示具体错误原因，引导商户修正                            │   │
│  │                                                                      │   │
│  │ 验证结果包含:                                                        │   │
│  │ - 网络延迟: xxx ms                                                   │   │
│  │ - 支持模型: [gpt-4, gpt-3.5-turbo, ...]                              │   │
│  │ - 官方定价: input $0.01/1K, output $0.03/1K                          │   │
│  │ - 利润空间: 建议售价 ≥ $0.015/1K (input)                              │   │
│  │ - 健康状态: healthy                                                  │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                              │                                              │
│                              ▼                                              │
│  步骤4: 商户确认配置                                                         │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ 商户确认:                                                            │   │
│  │ - 选择要上架的模型（从自动获取的列表中勾选）                            │   │
│  │ - 设置成本定价（可使用建议定价或自定义）                               │   │
│  │ - 设置配额限制                                                       │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                              │                                              │
│                              ▼                                              │
│  步骤5: 进入供应商资源池                                                     │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ 状态: verified → 可被智能路由选择                                     │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

#### 以字节跳动火山引擎为例的对接验证

**字节跳动火山引擎大模型产品线**：

| 产品类型 | 模型名称 | API端点 | 验证方式 |
|---------|---------|---------|---------|
| 自然语言 | Doubao-Pro | `https://ark.cn-beijing.volces.com/api/v3/chat/completions` | 调用chat接口 |
| 自然语言 | Doubao-Lite | 同上 | 调用chat接口 |
| 多模态 | Doubao-Vision | 同上 | 发送图片+文本 |
| 音频 | Doubao-Audio | `https://openspeech.bytedance.com/api/v1/...` | 音频合成/识别 |
| 视频 | 视频生成模型 | 特定端点 | 视频生成任务 |

**自动验证流程示例**：

```go
// services/api_key_validator.go
type APIKeyValidator struct {
    db *sql.DB
}

type ValidationResult struct {
    Success         bool                    `json:"success"`
    NetworkLatency  int                     `json:"network_latency_ms"`
    SupportedModels []ModelInfo             `json:"supported_models"`
    Pricing         map[string]PricingInfo  `json:"pricing"`
    HealthStatus    string                  `json:"health_status"`
    Errors          []string                `json:"errors,omitempty"`
}

type ModelInfo struct {
    ModelID      string   `json:"model_id"`
    ModelName    string   `json:"model_name"`
    Capabilities []string `json:"capabilities"` // chat, vision, audio, video
}

type PricingInfo struct {
    InputPricePer1K  float64 `json:"input_price_per_1k"`
    OutputPricePer1K float64 `json:"output_price_per_1k"`
    Currency         string  `json:"currency"`
}

func (v *APIKeyValidator) ValidateAPIKey(provider, apiKey, apiSecret, apiBaseURL string) (*ValidationResult, error) {
    result := &ValidationResult{}
    
    // 1. 确定API端点
    baseURL := v.getBaseURL(provider, apiBaseURL)
    
    // 2. 网络可达性测试
    latency, err := v.testNetworkConnectivity(baseURL)
    if err != nil {
        result.Errors = append(result.Errors, fmt.Sprintf("网络不可达: %v", err))
        result.HealthStatus = "unreachable"
        return result, nil
    }
    result.NetworkLatency = latency
    
    // 3. API Key有效性验证
    models, err := v.fetchAvailableModels(provider, baseURL, apiKey)
    if err != nil {
        result.Errors = append(result.Errors, fmt.Sprintf("API Key无效: %v", err))
        result.HealthStatus = "invalid_key"
        return result, nil
    }
    result.SupportedModels = models
    
    // 4. 获取官方定价
    pricing, err := v.fetchOfficialPricing(provider)
    if err == nil {
        result.Pricing = pricing
    }
    
    // 5. 设置健康状态
    result.HealthStatus = "healthy"
    result.Success = true
    
    return result, nil
}

func (v *APIKeyValidator) testNetworkConnectivity(baseURL string) (int, error) {
    start := time.Now()
    resp, err := http.Get(baseURL + "/models")
    if err != nil {
        return 0, err
    }
    defer resp.Body.Close()
    
    return int(time.Since(start).Milliseconds()), nil
}

func (v *APIKeyValidator) fetchAvailableModels(provider, baseURL, apiKey string) ([]ModelInfo, error) {
    switch provider {
    case "volcengine", "bytedance":
        return v.fetchVolcengineModels(baseURL, apiKey)
    case "openai":
        return v.fetchOpenAIModels(baseURL, apiKey)
    case "anthropic":
        return v.fetchAnthropicModels(baseURL, apiKey)
    default:
        return v.fetchOpenAICompatibleModels(baseURL, apiKey)
    }
}

func (v *APIKeyValidator) fetchVolcengineModels(baseURL, apiKey string) ([]ModelInfo, error) {
    // 火山引擎模型列表API
    req, _ := http.NewRequest("GET", baseURL+"/models", nil)
    req.Header.Set("Authorization", "Bearer "+apiKey)
    
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var modelsResp struct {
        Data []struct {
            ID     string `json:"id"`
            Object string `json:"object"`
            OwnedBy string `json:"owned_by"`
        } `json:"data"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
        return nil, err
    }
    
    var models []ModelInfo
    for _, m := range modelsResp.Data {
        capabilities := v.detectModelCapabilities(m.ID)
        models = append(models, ModelInfo{
            ModelID:      m.ID,
            ModelName:    m.ID,
            Capabilities: capabilities,
        })
    }
    
    return models, nil
}

func (v *APIKeyValidator) detectModelCapabilities(modelID string) []string {
    var capabilities []string
    
    // 基于模型名称推断能力
    lowerModelID := strings.ToLower(modelID)
    
    if strings.Contains(lowerModelID, "vision") || 
       strings.Contains(lowerModelID, "multimodal") {
        capabilities = append(capabilities, "vision", "multimodal")
    }
    if strings.Contains(lowerModelID, "audio") || 
       strings.Contains(lowerModelID, "tts") || 
       strings.Contains(lowerModelID, "asr") {
        capabilities = append(capabilities, "audio")
    }
    if strings.Contains(lowerModelID, "video") {
        capabilities = append(capabilities, "video")
    }
    
    // 默认都有对话能力
    capabilities = append(capabilities, "chat")
    
    return capabilities
}
```

#### 验证API设计

```
POST /api/v1/merchants/api-keys/validate
请求:
{
    "provider": "volcengine",
    "api_key": "xxx",
    "api_secret": "xxx",        // 可选
    "api_base_url": "https://ark.cn-beijing.volces.com/api/v3"  // 可选，自定义端点
}

响应:
{
    "success": true,
    "network_latency_ms": 45,
    "supported_models": [
        {
            "model_id": "doubao-pro-32k",
            "model_name": "Doubao Pro 32K",
            "capabilities": ["chat", "context_32k"]
        },
        {
            "model_id": "doubao-vision",
            "model_name": "Doubao Vision",
            "capabilities": ["chat", "vision", "multimodal"]
        }
    ],
    "pricing": {
        "doubao-pro-32k": {
            "input_price_per_1k": 0.0008,
            "output_price_per_1k": 0.002,
            "currency": "CNY"
        }
    },
    "health_status": "healthy",
    "suggestions": {
        "recommended_models": ["doubao-pro-32k", "doubao-vision"],
        "min_selling_price": {
            "input_price_per_1k": 0.001,  // 建议最低售价（成本+利润）
            "output_price_per_1k": 0.0025
        }
    }
}
```

#### 成本定价验证逻辑

```go
func (v *APIKeyValidator) validatePricing(merchantCost, platformPrice PricingInfo) (*PricingValidation, error) {
    result := &PricingValidation{}
    
    // 检查商户成本是否低于平台售价
    if merchantCost.InputPricePer1K >= platformPrice.InputPricePer1K {
        result.InputValid = false
        result.InputMessage = fmt.Sprintf(
            "输入Token成本(%.6f)高于平台售价(%.6f)，无法盈利",
            merchantCost.InputPricePer1K, platformPrice.InputPricePer1K,
        )
    } else {
        result.InputValid = true
        result.InputProfitMargin = (platformPrice.InputPricePer1K - merchantCost.InputPricePer1K) / platformPrice.InputPricePer1K * 100
    }
    
    if merchantCost.OutputPricePer1K >= platformPrice.OutputPricePer1K {
        result.OutputValid = false
        result.OutputMessage = fmt.Sprintf(
            "输出Token成本(%.6f)高于平台售价(%.6f)，无法盈利",
            merchantCost.OutputPricePer1K, platformPrice.OutputPricePer1K,
        )
    } else {
        result.OutputValid = true
        result.OutputProfitMargin = (platformPrice.OutputPricePer1K - merchantCost.OutputPricePer1K) / platformPrice.OutputPricePer1K * 100
    }
    
    result.IsValid = result.InputValid && result.OutputValid
    
    return result, nil
}
```

#### GAP清单

| GAP | 问题描述 | 影响 | 优先级 |
|-----|---------|------|--------|
| **GAP-23** | 无API Key有效性验证 | 无效Key直接上架，用户调用失败 | P0 |
| **GAP-24** | 无网络可达性测试 | 网络不通的Key被使用 | P0 |
| **GAP-25** | 无模型列表自动获取 | 商户手动配置易出错 | P1 |
| **GAP-26** | 无成本定价验证 | 可能出现负利润 | P0 |
| **GAP-27** | 无自动化对接工具 | 商户接入效率低 | P1 |
| **GAP-28** | 无健康状态初始化 | 新Key状态未知 | P1 |

#### 极简对接设计方案

**目标**：商户只需3步完成对接

```
┌─────────────────────────────────────────────────────────────────┐
│                    极简对接三步走                                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  第1步: 输入API Key                                             │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ [选择Provider: 火山引擎 ▼]                               │   │
│  │ [API Key: sk-xxxxxxxxxxxxx        ]                     │   │
│  │ [API Secret: (可选)               ]                     │   │
│  │ [自定义端点: (可选)               ]                     │   │
│  │                                                          │   │
│  │                    [一键验证]                            │   │
│  └─────────────────────────────────────────────────────────┘   │
│                              │                                  │
│                              ▼                                  │
│  第2步: 确认自动填充的信息                                       │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ ✅ 验证通过！                                            │   │
│  │                                                          │   │
│  │ 支持的模型:                                              │   │
│  │ ☑ doubao-pro-32k   (对话, 32K上下文)                    │   │
│  │ ☑ doubao-vision    (多模态, 图片理解)                   │   │
│  │ ☐ doubao-audio     (音频处理)                           │   │
│  │                                                          │   │
│  │ 成本定价 (自动获取，可修改):                              │   │
│  │ 输入: ¥0.0008/1K tokens                                 │   │
│  │ 输出: ¥0.0020/1K tokens                                 │   │
│  │                                                          │   │
│  │ 配额限制: [1000000] tokens                               │   │
│  │                                                          │   │
│  │                    [确认上架]                            │   │
│  └─────────────────────────────────────────────────────────┘   │
│                              │                                  │
│                              ▼                                  │
│  第3步: 完成                                                    │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ 🎉 对接成功！                                            │   │
│  │                                                          │   │
│  │ 您的API已加入供应商资源池，可被智能路由选择。              │   │
│  │                                                          │   │
│  │ 健康状态: ● 健康                                         │   │
│  │ 网络延迟: 45ms                                           │   │
│  │ 可用模型: 2个                                            │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

#### 自动化对接工具建议

**平台应提供的自动化工具**：

| 工具 | 功能 | 价值 |
|------|------|------|
| API Key验证器 | 一键验证Key有效性 | 避免无效Key上架 |
| 模型探测工具 | 自动获取支持的模型 | 减少手动配置 |
| 定价计算器 | 自动获取官方定价，计算利润空间 | 防止负利润 |
| 健康监控服务 | 持续监控API状态 | 高可用保障 |
| 一键测试工具 | 发送测试请求验证完整流程 | 端到端验证 |

**数据库增强**：

```sql
-- merchant_api_keys表增加验证相关字段
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS verified_at TIMESTAMP;
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS verified_by VARCHAR(50);  -- system/manual
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS verification_result JSONB;
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS last_test_at TIMESTAMP;
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS last_test_result JSONB;

COMMENT ON COLUMN merchant_api_keys.verified_at IS '验证通过时间';
COMMENT ON COLUMN merchant_api_keys.verified_by IS '验证方式: system(自动)/manual(手动)';
COMMENT ON COLUMN merchant_api_keys.verification_result IS '验证结果详情: {network_latency, models, pricing, ...}';
COMMENT ON COLUMN merchant_api_keys.last_test_at IS '最后一次测试时间';
COMMENT ON COLUMN merchant_api_keys.last_test_result IS '最后一次测试结果';
```

---

## 十一、商户对接实现深度分析

### 11.1 当前实现架构

#### 数据模型关系

```
┌─────────────────┐     ┌──────────────────────┐     ┌─────────────────┐
│    merchants    │     │   merchant_api_keys  │     │      spus       │
│─────────────────│     │──────────────────────│     │─────────────────│
│ id              │────<│ merchant_id          │     │ id              │
│ user_id         │     │ id                   │     │ spu_code        │
│ company_name    │     │ name                 │     │ model_provider  │
│ status          │     │ provider ◄───────────│─────│ model_name      │
│ ...             │     │ api_key_encrypted    │     │ ...             │
└─────────────────┘     │ quota_limit          │     └─────────────────┘
                        │ status               │              │
                        └──────────────────────┘              │
                                  │                           │
                                  │                           │
                                  ▼                           ▼
                        ┌──────────────────────┐     ┌─────────────────┐
                        │    merchant_skus     │     │      skus       │
                        │──────────────────────│     │─────────────────│
                        │ merchant_id          │────<│ id              │
                        │ sku_id               │     │ spu_id          │
                        │ api_key_id           │     │ sku_code        │
                        │ status               │     │ retail_price    │
                        └──────────────────────┘     │ ...             │
                                                     └─────────────────┘
```

#### 对接流程

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           商户对接当前实现                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  1. 商户注册                                                                │
│     POST /api/v1/merchants/register                                        │
│     { company_name, business_license, contact_info... }                    │
│                              ↓                                              │
│  2. 管理员审核                                                              │
│     POST /api/v1/admin/merchants/:id/approve                               │
│                              ↓                                              │
│  3. 商户添加API Key（商户自己注入）                                          │
│     POST /api/v1/merchants/api-keys                                        │
│     { name, provider, api_key, quota_limit }                               │
│                              ↓                                              │
│  4. 商户选择SKU上架（SKU层面对接）                                           │
│     POST /api/v1/merchants/skus                                            │
│     { sku_id, api_key_id }                                                 │
│                              ↓                                              │
│  5. 用户调用API                                                             │
│     POST /api/v1/proxy/chat                                                │
│     { provider, model, messages, merchant_sku_id }                         │
│                              ↓                                              │
│  6. 平台路由选择                                                            │
│     selectAPIKeyForRequest() → 按 quota_limit - quota_used DESC            │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 11.2 关键问题分析

#### 问题1：对接层次不清晰

| 层次 | 当前实现 | 问题 |
|------|---------|------|
| SPU层 | ❌ 无商户对接 | 商户无法直接对接某个模型（SPU） |
| SKU层 | ✅ merchant_skus表 | 商户只能选择平台定义的套餐上架 |
| API Key层 | ⚠️ 仅provider字段 | 缺少endpoint、模型映射配置 |

**当前设计的问题**：
- 商户无法自定义模型端点（如私有部署的模型）
- 商户无法配置模型映射（如自己的模型名称与平台模型的对应关系）
- 商户API Key只关联provider，无法精确到具体模型

#### 问题2：API Key配置不完整

**当前merchant_api_keys表字段**：
```sql
CREATE TABLE merchant_api_keys (
    id SERIAL PRIMARY KEY,
    merchant_id INT,
    name VARCHAR,
    provider VARCHAR,           -- 仅存储provider名称，如"openai"
    api_key_encrypted VARCHAR,  -- API Key加密存储
    api_secret_encrypted VARCHAR,
    quota_limit DECIMAL,
    quota_used DECIMAL,
    status VARCHAR
);
```

**缺失的关键字段**：
- `api_base_url` - 自定义API端点
- `supported_models` - 支持的模型列表
- `model_mapping` - 模型名称映射
- `health_status` - 健康状态

#### 问题3：商户与SPU的关系缺失

**当前设计**：
- 商户通过`merchant_skus`表关联SKU
- 没有商户与SPU的直接关联

**需求场景**：
- 商户可能只想提供某个模型（SPU）的服务
- 商户可能有多个API Key对应不同的模型
- 需要在SPU层面标识哪些商户可以提供服务

### 11.3 优化方案

#### 方案A：增强SKU层对接（推荐）

**优点**：改动小，兼容现有设计
**实现**：增强`merchant_api_keys`表和`merchant_skus`表

```sql
-- 1. 增强merchant_api_keys表
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS api_base_url VARCHAR(500);
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS supported_models JSONB;
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS model_mapping JSONB;
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS health_status VARCHAR(20) DEFAULT 'unknown';
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS last_health_check TIMESTAMP;

COMMENT ON COLUMN merchant_api_keys.api_base_url IS '自定义API端点，如https://api.my-company.com/v1';
COMMENT ON COLUMN merchant_api_keys.supported_models IS '支持的模型列表，如["gpt-4", "gpt-3.5-turbo"]';
COMMENT ON COLUMN merchant_api_keys.model_mapping IS '模型名称映射，如{"my-gpt4": "gpt-4"}';
COMMENT ON COLUMN merchant_api_keys.health_status IS '健康状态: healthy, degraded, unhealthy, unknown';

-- 2. 增强merchant_skus表
ALTER TABLE merchant_skus ADD COLUMN IF NOT EXISTS cost_input_rate DECIMAL(10,6);
ALTER TABLE merchant_skus ADD COLUMN IF NOT EXISTS cost_output_rate DECIMAL(10,6);
ALTER TABLE merchant_skus ADD COLUMN IF NOT EXISTS priority INT DEFAULT 0;

COMMENT ON COLUMN merchant_skus.cost_input_rate IS '商户成本输入Token单价(元/1K)';
COMMENT ON COLUMN merchant_skus.cost_output_rate IS '商户成本输出Token单价(元/1K)';
COMMENT ON COLUMN merchant_skus.priority IS '路由优先级，数值越高优先级越高';
```

#### 方案B：增加SPU层对接

**优点**：更灵活，支持商户直接对接模型
**实现**：新增`merchant_spu_providers`表

```sql
-- 新增商户SPU对接表
CREATE TABLE merchant_spu_providers (
    id SERIAL PRIMARY KEY,
    merchant_id INT NOT NULL REFERENCES merchants(id),
    spu_id INT NOT NULL REFERENCES spus(id),
    api_key_id INT REFERENCES merchant_api_keys(id),
    
    -- 成本定价
    cost_input_rate DECIMAL(10,6),
    cost_output_rate DECIMAL(10,6),
    
    -- 路由配置
    priority INT DEFAULT 0,
    weight INT DEFAULT 100,
    
    -- 状态
    status VARCHAR(20) DEFAULT 'active',
    health_status VARCHAR(20) DEFAULT 'unknown',
    
    -- 统计
    total_requests BIGINT DEFAULT 0,
    total_tokens BIGINT DEFAULT 0,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(merchant_id, spu_id)
);
```

### 11.4 推荐实施方案

**采用方案A（增强SKU层对接）**，理由：

1. **兼容性好**：不破坏现有数据结构
2. **改动较小**：只需增强字段，不需新建表
3. **满足需求**：可以支持自定义端点、模型映射、成本定价

#### 实施步骤

```
Phase 1: 数据库增强（1天）
├── 1.1 merchant_api_keys表增加字段
├── 1.2 merchant_skus表增加字段
└── 1.3 编写迁移脚本

Phase 2: API增强（2天）
├── 2.1 商户API Key管理API增强
│   ├── 支持自定义端点配置
│   ├── 支持模型映射配置
│   └── 支持健康状态查询
├── 2.2 商户SKU管理API增强
│   ├── 支持成本定价配置
│   └── 支持优先级配置
└── 2.3 前端页面更新

Phase 3: 路由增强（2天）
├── 3.1 智能路由支持自定义端点
├── 3.2 智能路由支持模型映射
└── 3.3 健康检查集成

Phase 4: 测试验证（1天）
├── 4.1 单元测试
├── 4.2 集成测试
└── 4.3 端到端测试
```

### 10.5 商户对接职责划分

| 职责 | 平台运营端 | 商户端 |
|------|-----------|--------|
| 定义SPU（模型） | ✅ 负责 | ❌ 无权限 |
| 定义SKU（套餐） | ✅ 负责 | ❌ 无权限 |
| 商户入驻审核 | ✅ 负责 | ❌ 无权限 |
| 添加API Key | ❌ 无需 | ✅ 商户自己注入 |
| 配置API端点 | ❌ 无需 | ✅ 商户自己配置 |
| 选择SKU上架 | ❌ 无需 | ✅ 商户自己选择 |
| 绑定API Key到SKU | ❌ 无需 | ✅ 商户自己绑定 |
| 配置成本定价 | ⚠️ 可代配置 | ✅ 商户自己配置 |
| 健康监控 | ✅ 平台负责 | ✅ 可查看状态 |
| 结算计算 | ✅ 平台负责 | ❌ 无权限 |

### 11.6 对接内容清单

| 对接内容 | 存储位置 | 配置方 | 说明 |
|---------|---------|--------|------|
| 商户信息 | `merchants`表 | 商户注册 | 公司信息、联系方式 |
| 大模型Provider | `merchant_api_keys.provider` | 商户配置 | 如openai, anthropic |
| API端点 | `merchant_api_keys.api_base_url`（需新增） | 商户配置 | 自定义API地址 |
| API Key | `merchant_api_keys.api_key_encrypted` | 商户注入 | 加密存储 |
| 模型映射 | `merchant_api_keys.model_mapping`（需新增） | 商户配置 | 模型名称对应关系 |
| 成本定价 | `merchant_skus.cost_input_rate/output_rate`（需新增） | 商户配置 | 商户成本单价 |
| SKU关联 | `merchant_skus`表 | 商户选择 | 选择上架的套餐 |

---

**报告版本**: v1.6
**最后更新**: 2026-04-04
