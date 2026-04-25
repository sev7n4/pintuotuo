# 价格优先路由策略算法优化实施计划

> 版本: v1.0
> 创建日期: 2026-04-25
> 状态: 待审批

---

## 一、背景与目标

### 1.1 背景

当前路由系统存在以下问题：

| 问题 | 当前实现 | 影响 |
|------|----------|------|
| 价格计算不严谨 | `(input_rate + output_rate) / 2` | 丢失输入/输出价格差异，无法反映真实成本 |
| Token 预估缺失 | 无 | 无法计算真实成本 |
| 策略选择 BUG | 硬编码 `balanced` | 忽略数据库默认策略配置 |
| 评分维度不足 | 3 维 | 缺少安全、负载均衡维度 |

### 1.2 目标

1. **修复策略选择 BUG**：从数据库读取默认策略
2. **实现真实成本计算**：基于 Token 预估计算实际成本
3. **完善五维评分体系**：价格、延迟、可靠性、安全、负载均衡
4. **建立数据驱动机制**：基于历史统计优化决策

---

## 二、数据模型设计

### 2.1 新增表：model_token_statistics

```sql
CREATE TABLE model_token_statistics (
    id SERIAL PRIMARY KEY,
    model_name VARCHAR(100) NOT NULL UNIQUE,
    
    -- Token 统计（基于历史请求）
    avg_input_tokens NUMERIC(10,2) DEFAULT 0,
    avg_output_tokens NUMERIC(10,2) DEFAULT 0,
    p50_input_tokens INT DEFAULT 0,
    p50_output_tokens INT DEFAULT 0,
    p90_input_tokens INT DEFAULT 0,
    p90_output_tokens INT DEFAULT 0,
    
    -- 比例统计
    input_output_ratio NUMERIC(5,2) DEFAULT 1.0,
    
    -- 样本统计
    total_requests INT DEFAULT 0,
    sample_start_date DATE,
    sample_end_date DATE,
    
    -- 元数据
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 索引
CREATE INDEX idx_model_token_stats_model ON model_token_statistics(model_name);
CREATE INDEX idx_model_token_stats_updated ON model_token_statistics(updated_at);

-- 注释
COMMENT ON TABLE model_token_statistics IS '模型维度的 Token 统计，用于预估请求成本';
```

### 2.2 现有表字段说明

```
merchant_api_keys (已存在)
├── cost_input_rate     -- 输入 token 单价
├── cost_output_rate    -- 输出 token 单价
├── avg_latency_ms      -- 平均延迟
├── success_rate        -- 成功率
├── health_status       -- 健康状态
└── security_level      -- 安全级别

routing_strategies (已存在，已添加字段)
├── price_weight        -- 价格权重
├── latency_weight      -- 延迟权重
├── reliability_weight  -- 可靠性权重
├── security_weight     -- 安全权重 (新增)
└── load_balance_weight -- 负载均衡权重 (新增)
```

---

## 三、核心算法设计

### 3.1 成本预估公式

```
预估成本 = 预估输入tokens × 输入单价 + 预估输出tokens × 输出单价
```

### 3.2 Token 预估优先级

```
1. 请求参数预估
   └── max_tokens 作为输出上限

2. 历史统计预估
   └── model_token_statistics.avg_input_tokens
   └── model_token_statistics.avg_output_tokens

3. 实时估算 (Fallback)
   └── 输入: len(messages) / 4 (字符转token)
   └── 输出: 默认 500 tokens
```

### 3.3 五维评分算法

| 维度 | 计算公式 | 说明 |
|------|----------|------|
| 价格评分 | `1 - (成本 - min) / (max - min)` | 成本越低分数越高 |
| 延迟评分 | `1 - (延迟 - min) / (max - min)` | 延迟越低分数越高 |
| 可靠性评分 | `成功率 × 健康权重` | 健康状态加权 |
| 安全评分 | `安全级别分 × 合规检查` | 不合规直接淘汰 |
| 负载评分 | `1 - (使用量 - min) / (max - min)` | 使用越少分数越高 |

### 3.4 综合评分公式

```
总分 = 价格评分 × 价格权重 
     + 延迟评分 × 延迟权重
     + 可靠性评分 × 可靠性权重
     + 安全评分 × 安全权重
     + 负载评分 × 负载权重
```

---

## 四、实施任务清单

### 阶段 P0：紧急修复（预计 0.5 天）

| 任务ID | 任务描述 | 涉及文件 | 优先级 |
|--------|----------|----------|--------|
| P0-1 | 修复策略选择 BUG，从数据库读取默认策略 | `routing_strategy_engine.go` | 🔴 高 |
| P0-2 | 添加单元测试验证策略选择逻辑 | `routing_strategy_test.go` | 🔴 高 |

### 阶段 P1：核心功能（预计 2 天）

| 任务ID | 任务描述 | 涉及文件 | 优先级 |
|--------|----------|----------|--------|
| P1-1 | 创建 `model_token_statistics` 表迁移文件 | `migrations/072_*.sql` | 🔴 高 |
| P1-2 | 实现 Token 预估服务 | `services/token_estimation.go` | 🔴 高 |
| P1-3 | 实现成本预估算法 | `services/cost_estimation.go` | 🔴 高 |
| P1-4 | 重构价格评分算法 | `services/smart_router.go` | 🔴 高 |
| P1-5 | 更新路由决策日志记录 | `services/unified_routing_engine.go` | 🔴 高 |
| P1-6 | 添加单元测试 | `services/*_test.go` | 🔴 高 |

### 阶段 P2：完善功能（预计 1.5 天）

| 任务ID | 任务描述 | 涉及文件 | 优先级 |
|--------|----------|----------|--------|
| P2-1 | 实现安全评分算法 | `services/smart_router.go` | 🟡 中 |
| P2-2 | 实现负载均衡评分算法 | `services/smart_router.go` | 🟡 中 |
| P2-3 | 创建 Token 统计定时任务 | `services/token_stats_worker.go` | 🟡 中 |
| P2-4 | 更新 Admin 策略配置页面 | `frontend/src/pages/` | 🟡 中 |
| P2-5 | 添加集成测试 | `tests/` | 🟡 中 |

### 阶段 P3：优化增强（预计 1 天）

| 任务ID | 任务描述 | 涉及文件 | 优先级 |
|--------|----------|----------|--------|
| P3-1 | 实现自动策略动态权重 | `services/routing_strategy_engine.go` | 🟢 低 |
| P3-2 | 添加路由决策监控指标 | `services/metrics.go` | 🟢 低 |
| P3-3 | 优化 Admin 策略配置 UI | `frontend/src/pages/` | 🟢 低 |

---

## 五、详细任务说明

### P0-1：修复策略选择 BUG

**问题**：`RoutingStrategyEngine.determineStrategy` 使用硬编码的 `e.defaultStrategy`，未从数据库读取默认策略。

**修复方案**：

```go
func (e *RoutingStrategyEngine) determineStrategy(reqCtx *RequestContext) StrategyGoal {
    // 1. 用户指定策略优先
    if reqCtx.UserPreferences != nil {
        if strategy, ok := reqCtx.UserPreferences["strategy"].(string); ok {
            switch StrategyGoal(strategy) {
            case GoalPerformanceFirst, GoalPriceFirst, GoalReliabilityFirst,
                GoalBalanced, GoalSecurityFirst:
                return StrategyGoal(strategy)
            }
        }
    }
    
    // 2. 请求特征推断
    if reqCtx.RequestAnalysis != nil {
        analysis := reqCtx.RequestAnalysis
        if analysis.Complexity == ComplexityComplex {
            return GoalReliabilityFirst
        }
        if analysis.EstimatedTokens > 8000 {
            return GoalPriceFirst
        }
        if analysis.Stream {
            return GoalPerformanceFirst
        }
    }
    
    // 3. 从数据库读取默认策略
    defaultStrategy := e.getDefaultStrategyFromDB()
    if defaultStrategy != "" {
        return defaultStrategy
    }
    
    // 4. 最终 fallback
    return GoalBalanced
}

func (e *RoutingStrategyEngine) getDefaultStrategyFromDB() StrategyGoal {
    var code string
    err := e.db.QueryRow(
        "SELECT code FROM routing_strategies WHERE is_default = true AND status = 'active' LIMIT 1",
    ).Scan(&code)
    if err != nil {
        return ""
    }
    return StrategyGoal(code)
}
```

### P1-2：Token 预估服务

**新建文件**：`services/token_estimation.go`

```go
package services

type TokenEstimation struct {
    EstimatedInputTokens  float64
    EstimatedOutputTokens float64
    Source                string // "request", "statistics", "fallback"
}

type TokenEstimationService struct {
    db *sql.DB
}

func NewTokenEstimationService() *TokenEstimationService {
    return &TokenEstimationService{db: config.GetDB()}
}

func (s *TokenEstimationService) EstimateTokens(req *RoutingRequest) *TokenEstimation {
    // 优先级 1: 请求参数
    if req.MaxTokens > 0 {
        inputTokens := s.estimateInputFromMessages(req.Messages)
        return &TokenEstimation{
            EstimatedInputTokens:  float64(inputTokens),
            EstimatedOutputTokens: float64(req.MaxTokens),
            Source:                "request",
        }
    }
    
    // 优先级 2: 历史统计
    stats, err := s.getModelStatistics(req.Model)
    if err == nil && stats.TotalRequests > 0 {
        return &TokenEstimation{
            EstimatedInputTokens:  stats.AvgInputTokens,
            EstimatedOutputTokens: stats.AvgOutputTokens,
            Source:                "statistics",
        }
    }
    
    // 优先级 3: Fallback
    return s.getFallbackEstimation(req)
}

func (s *TokenEstimationService) estimateInputFromMessages(messages []Message) int {
    totalChars := 0
    for _, msg := range messages {
        totalChars += len(msg.Content)
    }
    return totalChars / 4 // 粗略估算：4字符≈1token
}

func (s *TokenEstimationService) getModelStatistics(model string) (*ModelTokenStats, error) {
    // 从 model_token_statistics 表查询
}

func (s *TokenEstimationService) getFallbackEstimation(req *RoutingRequest) *TokenEstimation {
    return &TokenEstimation{
        EstimatedInputTokens:  500,
        EstimatedOutputTokens: 500,
        Source:                "fallback",
    }
}
```

### P1-3：成本预估算法

**新建文件**：`services/cost_estimation.go`

```go
package services

type CostEstimation struct {
    APIKeyID       int
    MerchantID     int
    EstimatedCost  float64
    InputCost      float64
    OutputCost     float64
    InputPrice     float64
    OutputPrice    float64
}

type CostEstimationService struct {
    tokenService *TokenEstimationService
}

func (s *CostEstimationService) CalculateEstimatedCost(
    candidate *RoutingCandidate,
    estimation *TokenEstimation,
) *CostEstimation {
    inputCost := estimation.EstimatedInputTokens * candidate.InputPrice
    outputCost := estimation.EstimatedOutputTokens * candidate.OutputPrice
    
    return &CostEstimation{
        APIKeyID:      candidate.APIKeyID,
        MerchantID:    candidate.MerchantID,
        EstimatedCost: inputCost + outputCost,
        InputCost:     inputCost,
        OutputCost:    outputCost,
        InputPrice:    candidate.InputPrice,
        OutputPrice:   candidate.OutputPrice,
    }
}

func (s *CostEstimationService) CalculatePriceScores(
    candidates []RoutingCandidate,
    estimation *TokenEstimation,
) map[int]float64 {
    costs := make(map[int]float64)
    
    for _, c := range candidates {
        cost := s.CalculateEstimatedCost(&c, estimation)
        costs[c.APIKeyID] = cost.EstimatedCost
    }
    
    minCost, maxCost := minMax(costs)
    scores := make(map[int]float64)
    
    for id, cost := range costs {
        if maxCost == minCost {
            scores[id] = 1.0
        } else {
            scores[id] = 1.0 - (cost - minCost) / (maxCost - minCost)
        }
    }
    
    return scores
}
```

### P2-3：Token 统计定时任务

**新建文件**：`services/token_stats_worker.go`

```go
package services

type TokenStatsWorker struct {
    db *sql.DB
}

func NewTokenStatsWorker() *TokenStatsWorker {
    return &TokenStatsWorker{db: config.GetDB()}
}

func (w *TokenStatsWorker) Run() {
    // 每日凌晨执行
    ticker := time.NewTicker(24 * time.Hour)
    for range ticker.C {
        w.updateStatistics()
    }
}

func (w *TokenStatsWorker) updateStatistics() error {
    query := `
        INSERT INTO model_token_statistics (
            model_name, avg_input_tokens, avg_output_tokens,
            p50_input_tokens, p50_output_tokens,
            p90_input_tokens, p90_output_tokens,
            input_output_ratio, total_requests,
            sample_start_date, sample_end_date, updated_at
        )
        SELECT 
            model,
            AVG(input_tokens),
            AVG(output_tokens),
            PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY input_tokens),
            PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY output_tokens),
            PERCENTILE_CONT(0.9) WITHIN GROUP (ORDER BY input_tokens),
            PERCENTILE_CONT(0.9) WITHIN GROUP (ORDER BY output_tokens),
            AVG(input_tokens::FLOAT / NULLIF(output_tokens, 0)),
            COUNT(*),
            NOW() - INTERVAL '7 days',
            NOW(),
            NOW()
        FROM usage_logs
        WHERE created_at > NOW() - INTERVAL '7 days'
        GROUP BY model
        ON CONFLICT (model_name) DO UPDATE SET
            avg_input_tokens = EXCLUDED.avg_input_tokens,
            avg_output_tokens = EXCLUDED.avg_output_tokens,
            p50_input_tokens = EXCLUDED.p50_input_tokens,
            p50_output_tokens = EXCLUDED.p50_output_tokens,
            p90_input_tokens = EXCLUDED.p90_input_tokens,
            p90_output_tokens = EXCLUDED.p90_output_tokens,
            input_output_ratio = EXCLUDED.input_output_ratio,
            total_requests = EXCLUDED.total_requests,
            sample_start_date = EXCLUDED.sample_start_date,
            sample_end_date = EXCLUDED.sample_end_date,
            updated_at = NOW()
    `
    _, err := w.db.Exec(query)
    return err
}
```

---

## 六、测试计划

### 6.1 单元测试

| 测试项 | 测试内容 |
|--------|----------|
| 策略选择测试 | 验证从数据库读取默认策略 |
| Token 预估测试 | 验证三级预估优先级 |
| 成本计算测试 | 验证成本预估公式 |
| 评分算法测试 | 验证五维评分归一化 |

### 6.2 集成测试

| 测试项 | 测试内容 |
|--------|----------|
| 端到端路由测试 | 验证三层路由完整流程 |
| 多商户价格比较 | 验证选择最低成本商户 |
| 策略权重测试 | 验证不同策略的评分结果 |

### 6.3 验收标准

| 标准 | 验证方法 |
|------|----------|
| 策略选择正确 | Admin 设置价格优先，路由日志显示 `price_first` |
| 成本计算准确 | 多商户时选择成本最低的 API Key |
| 日志记录完整 | 三层路由输入输出都有记录 |
| 测试覆盖率 | 单元测试覆盖率 > 80% |

---

## 七、风险评估

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| Token 预估不准 | 成本计算偏差 | 使用历史数据统计 + 多级 fallback |
| 数据库迁移失败 | 功能不可用 | 幂等迁移脚本 + 回滚方案 |
| 性能下降 | 响应变慢 | 缓存策略权重 + 异步统计更新 |

---

## 八、时间规划

| 阶段 | 预计时间 | 里程碑 |
|------|----------|--------|
| P0 紧急修复 | 0.5 天 | 策略选择 BUG 修复 |
| P1 核心功能 | 2 天 | 成本预估算法上线 |
| P2 完善功能 | 1.5 天 | 五维评分完整实现 |
| P3 优化增强 | 1 天 | 动态权重 + 监控 |
| **总计** | **5 天** | |

---

## 九、审批签字

| 角色 | 姓名 | 签字 | 日期 |
|------|------|------|------|
| 技术负责人 | | | |
| 产品负责人 | | | |
| 测试负责人 | | | |

---

**文档版本历史**

| 版本 | 日期 | 修改内容 | 作者 |
|------|------|----------|------|
| v1.0 | 2026-04-25 | 初始版本 | Trae AI |
