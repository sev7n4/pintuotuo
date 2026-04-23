# 智能路由系统迭代演进规划

## 一、现状分析

### 1.1 已实现能力

| 组件 | 功能 | 状态 |
|------|------|------|
| SmartRouter | API Key 选择（价格/延迟/成功率） | ✅ 已实现 |
| UnifiedRouter | 路由模式决策（direct/litellm/proxy） | ✅ 已实现 |
| CircuitBreaker | 熔断器 | ✅ 已实现 |
| FallbackManager | 降级管理 | ✅ 已实现 |
| HealthChecker | 健康检查 | ✅ 已实现 |
| PricingService | 价格服务 | ✅ 已实现 |

### 1.2 核心问题

```
当前架构（割裂状态）：

┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  UnifiedRouter  │    │   SmartRouter   │    │   api_proxy     │
│  (路由模式)      │    │  (API Key选择)   │    │   (执行请求)     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
        │                      │                      │
        └──────────────────────┴──────────────────────┘
                      缺少联动协同机制
```

**问题清单：**
1. ❌ 三层组件割裂，没有形成联动协同机制
2. ❌ 缺少路由策略层（业务目标定义）
3. ❌ 缺少路由感知基础设施（实时状态感知）
4. ❌ 缺少统一网关层（统一 API、可观测性）
5. ❌ 缺少限流机制（令牌桶）
6. ❌ 缺少队列管理
7. ❌ 缺少请求内容分析（意图、复杂度）

---

## 二、目标架构

### 2.1 三层架构设计

```
┌─────────────────────────────────────────────────────────────────┐
│                        业务层                                    │
│  用户请求 → API Gateway → 智能路由系统                           │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                   第一层：路由策略层                              │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │ 业务目标定义                                                  ││
│  │ ├─ 请求内容分析（意图、复杂度）                                ││
│  │ ├─ 用户偏好（商户类型、等级、区域）                            ││
│  │ ├─ 成本预算                                                   ││
│  │ └─ 合规要求（区域选择、安全等级）                              ││
│  └─────────────────────────────────────────────────────────────┘│
│  ┌─────────────────────────────────────────────────────────────┐│
│  │ 策略输出                                                      ││
│  │ ├─ 性能优先 (latency_first)                                  ││
│  │ ├─ 价格优先 (price_first)                                    ││
│  │ ├─ 可靠性优先 (reliability_first)                            ││
│  │ ├─ 均衡策略 (balanced)                                       ││
│  │ ├─ 安全优先 (security_first)                                 ││
│  │ └─ 默认策略 (auto)                                           ││
│  └─────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                   第二层：路由决策层                              │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │ 输入                                                          ││
│  │ ├─ 策略层目标                                                 ││
│  │ ├─ 路由感知数据（实时状态）                                    ││
│  │ └─ 候选供应商集合                                             ││
│  └─────────────────────────────────────────────────────────────┘│
│  ┌─────────────────────────────────────────────────────────────┐│
│  │ 决策算法                                                      ││
│  │ ├─ 多目标优化算法                                             ││
│  │ ├─ 加权评分模型                                               ││
│  │ ├─ 约束满足算法                                               ││
│  │ └─ 机器学习预测（未来）                                        ││
│  └─────────────────────────────────────────────────────────────┘│
│  ┌─────────────────────────────────────────────────────────────┐│
│  │ 决策输出                                                      ││
│  │ ├─ 选中的 API Key ID                                         ││
│  │ ├─ 路由模式 (direct/litellm/proxy)                           ││
│  │ ├─ 端点 URL                                                  ││
│  │ ├─ 降级方案                                                   ││
│  │ └─ 决策原因                                                   ││
│  └─────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                   第三层：路由执行层                              │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │ 统一网关                                                      ││
│  │ ├─ 统一 API 接口                                              ││
│  │ ├─ 屏蔽厂商差异                                               ││
│  │ └─ 协议转换                                                   ││
│  └─────────────────────────────────────────────────────────────┘│
│  ┌─────────────────────────────────────────────────────────────┐│
│  │ 可靠性机制                                                    ││
│  │ ├─ 熔断器 (Circuit Breaker) ✅                               ││
│  │ ├─ 重试策略                                                   ││
│  │ ├─ 超时控制                                                   ││
│  │ ├─ 故障转移 (Fallback) ✅                                    ││
│  │ └─ 降级策略 ✅                                                ││
│  └─────────────────────────────────────────────────────────────┘│
│  ┌─────────────────────────────────────────────────────────────┐│
│  │ 流量控制                                                      ││
│  │ ├─ 令牌桶限流                                                 ││
│  │ ├─ 队列管理                                                   ││
│  │ └─ 负载均衡                                                   ││
│  └─────────────────────────────────────────────────────────────┘│
│  ┌─────────────────────────────────────────────────────────────┐│
│  │ 可观测性                                                      ││
│  │ ├─ 日志 (Logging) ✅                                         ││
│  │ ├─ 追踪 (Tracing) ✅                                         ││
│  │ ├─ 指标 (Metrics) ✅                                         ││
│  │ ├─ 成本统计                                                   ││
│  │ └─ Token 用量统计 ✅                                          ││
│  └─────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│              基础设施层：路由感知能力                             │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │ 基础信息感知                                                  ││
│  │ ├─ 区域信息 (domestic/overseas) ✅                           ││
│  │ ├─ 端点信息 ✅                                                ││
│  │ ├─ 商户类型/等级 ✅                                           ││
│  │ ├─ 厂商信息 ✅                                                ││
│  │ └─ 安全等级                                                   ││
│  └─────────────────────────────────────────────────────────────┘│
│  ┌─────────────────────────────────────────────────────────────┐│
│  │ 实时状态感知                                                  ││
│  │ ├─ 延迟 (Latency) ✅                                         ││
│  │ ├─ 错误率 (Error Rate) ✅                                     ││
│  │ ├─ 成功率 (Success Rate) ✅                                   ││
│  │ ├─ 连接池状态                                                 ││
│  │ ├─ 限流信息                                                   ││
│  │ ├─ 负载均衡状态                                               ││
│  │ └─ 实时价格 ✅                                                ││
│  └─────────────────────────────────────────────────────────────┘│
│  ┌─────────────────────────────────────────────────────────────┐│
│  │ 感知数据存储                                                  ││
│  │ ├─ Redis 缓存（实时数据）                                      ││
│  │ ├─ PostgreSQL（历史数据）✅                                   ││
│  │ └─ Prometheus（指标数据）✅                                   ││
│  └─────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
```

---

## 三、迭代演进路线图

### Phase 1: 基础设施层建设（2周）

**目标：** 构建路由感知能力

#### 1.1 数据模型扩展

```sql
-- API Key 扩展字段
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS region VARCHAR(20) DEFAULT 'domestic';
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS security_level VARCHAR(20) DEFAULT 'standard';
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS route_preference JSONB DEFAULT '{}'::jsonb;

-- 实时状态表
CREATE TABLE IF NOT EXISTS api_key_realtime_status (
    api_key_id INT PRIMARY KEY,
    latency_p50 INT,
    latency_p95 INT,
    latency_p99 INT,
    error_rate DECIMAL(5,4),
    success_rate DECIMAL(5,4),
    connection_pool_size INT,
    connection_pool_active INT,
    rate_limit_remaining INT,
    rate_limit_reset_at TIMESTAMP,
    load_balance_weight DECIMAL(3,2),
    last_request_at TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 路由决策日志表
CREATE TABLE IF NOT EXISTS routing_decision_logs (
    id SERIAL PRIMARY KEY,
    request_id VARCHAR(64),
    merchant_id INT,
    api_key_id INT,
    strategy_layer_goal VARCHAR(50),
    decision_layer_output JSONB,
    execution_layer_result JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 1.2 路由感知服务

**文件：** `backend/services/route_awareness.go`

```go
type RouteAwarenessService struct {
    db          *sql.DB
    redis       *redis.Client
    prometheus  *metrics.Metrics
}

// 感知数据结构
type ProviderRealtimeStatus struct {
    APIKeyID            int
    Region              string
    Endpoint            string
    SecurityLevel       string
    
    // 实时指标
    LatencyP50          int
    LatencyP95          int
    LatencyP99          int
    ErrorRate           float64
    SuccessRate         float64
    
    // 连接池状态
    ConnectionPoolSize  int
    ConnectionPoolActive int
    
    // 限流信息
    RateLimitRemaining  int
    RateLimitResetAt    time.Time
    
    // 负载均衡
    LoadBalanceWeight   float64
}
```

#### 1.3 任务清单

| 任务 | 优先级 | 预计时间 |
|------|--------|----------|
| 扩展 merchant_api_keys 表字段 | P0 | 0.5天 |
| 创建 api_key_realtime_status 表 | P0 | 0.5天 |
| 创建 routing_decision_logs 表 | P0 | 0.5天 |
| 实现 RouteAwarenessService | P0 | 2天 |
| 实现实时状态采集器 | P0 | 2天 |
| Redis 缓存集成 | P1 | 1天 |
| Prometheus 指标扩展 | P1 | 1天 |

---

### Phase 2: 路由策略层建设（2周）

**目标：** 实现业务目标定义和策略输出

#### 2.1 请求内容分析器

**文件：** `backend/services/request_analyzer.go`

```go
type RequestAnalyzer struct {
    // 请求内容分析
}

type RequestAnalysis struct {
    Intent          string  // 意图：chat/completion/embedding
    Complexity      string  // 复杂度：low/medium/high
    TokenEstimate   int     // 预估 Token 数
    Priority        string  // 优先级：low/normal/high
    SecurityRequire string  // 安全要求：standard/enhanced
}
```

#### 2.2 策略引擎

**文件：** `backend/services/routing_strategy_engine.go`

```go
type RoutingStrategyEngine struct {
    db *sql.DB
}

type StrategyGoal struct {
    // 业务目标
    UserPreference   string  // 用户偏好
    CostBudget       float64 // 成本预算
    ComplianceRegion string  // 合规区域
    
    // 策略输出
    Strategy         RoutingStrategy
    Constraints      []Constraint
    Weights          StrategyWeights
}

// 策略类型
const (
    StrategyPerformance  = "performance_first"   // 性能优先
    StrategyPrice        = "price_first"         // 价格优先
    StrategyReliability  = "reliability_first"   // 可靠性优先
    StrategyBalanced     = "balanced"            // 均衡策略
    StrategySecurity     = "security_first"      // 安全优先
    StrategyAuto         = "auto"                // 默认策略
)
```

#### 2.3 任务清单

| 任务 | 优先级 | 预计时间 |
|------|--------|----------|
| 实现 RequestAnalyzer | P0 | 2天 |
| 实现 RoutingStrategyEngine | P0 | 3天 |
| 扩展策略类型（新增 reliability_first, security_first） | P0 | 1天 |
| 实现成本预算控制 | P1 | 2天 |
| 实现合规区域约束 | P1 | 1天 |
| 策略配置管理 API | P1 | 1天 |

---

### Phase 3: 路由决策层重构（2周）

**目标：** 整合三层联动协同机制

#### 3.1 统一路由决策引擎

**文件：** `backend/services/unified_routing_engine.go`

```go
type UnifiedRoutingEngine struct {
    strategyEngine   *RoutingStrategyEngine
    awarenessService *RouteAwarenessService
    smartRouter      *SmartRouter
    
    // 三层联动
}

type RoutingDecision struct {
    // 第一层：策略层输出
    StrategyGoal     *StrategyGoal
    
    // 第二层：决策层输出
    SelectedAPIKeyID int
    RouteMode        string
    Endpoint         string
    FallbackPlan     *FallbackPlan
    DecisionReason   string
    
    // 第三层：执行层配置
    ExecutionConfig  *ExecutionConfig
}

func (e *UnifiedRoutingEngine) Decide(
    ctx context.Context,
    request *RequestAnalysis,
    merchant *MerchantConfig,
    provider *ProviderConfig,
) (*RoutingDecision, error) {
    // 1. 策略层：定义目标
    goal := e.strategyEngine.DefineGoal(request, merchant)
    
    // 2. 获取感知数据
    status := e.awarenessService.GetRealtimeStatus(provider.Code)
    
    // 3. 决策层：做出选择
    candidates := e.smartRouter.GetCandidates(ctx, request.Model, provider.Code)
    selected := e.applyDecisionAlgorithm(candidates, goal, status)
    
    // 4. 生成执行配置
    execConfig := e.generateExecutionConfig(selected, goal)
    
    return &RoutingDecision{
        StrategyGoal:     goal,
        SelectedAPIKeyID: selected.APIKeyID,
        RouteMode:        selected.Mode,
        Endpoint:         selected.Endpoint,
        FallbackPlan:     selected.Fallback,
        ExecutionConfig:  execConfig,
    }, nil
}
```

#### 3.2 多目标优化算法

**文件：** `backend/services/decision_algorithm.go`

```go
type DecisionAlgorithm struct{}

func (a *DecisionAlgorithm) Optimize(
    candidates []RoutingCandidate,
    goal *StrategyGoal,
    status []*ProviderRealtimeStatus,
) *RoutingCandidate {
    // 多目标优化：
    // 1. 价格优化
    // 2. 延迟优化
    // 3. 可靠性优化
    // 4. 安全约束满足
    // 5. 区域约束满足
}
```

#### 3.3 任务清单

| 任务 | 优先级 | 预计时间 |
|------|--------|----------|
| 创建 UnifiedRoutingEngine | P0 | 3天 |
| 实现三层联动机制 | P0 | 2天 |
| 实现多目标优化算法 | P0 | 3天 |
| 重构 api_proxy.go 集成新引擎 | P0 | 2天 |
| 决策日志记录 | P1 | 1天 |
| 决策可视化 API | P1 | 1天 |

---

### Phase 4: 路由执行层增强（2周）

**目标：** 完善统一网关和可靠性机制

#### 4.1 统一网关

**文件：** `backend/services/unified_gateway.go`

#### 4.2 令牌桶限流

**文件：** `backend/services/rate_limiter.go`

#### 4.3 队列管理

**文件：** `backend/services/request_queue.go`

#### 4.4 任务清单

| 任务 | 优先级 | 预计时间 |
|------|--------|----------|
| 实现 UnifiedGateway | P0 | 3天 |
| 实现令牌桶限流 | P0 | 2天 |
| 实现请求队列管理 | P1 | 2天 |
| 实现重试策略增强 | P1 | 1天 |
| 实现超时控制增强 | P1 | 1天 |
| 可观测性增强 | P1 | 1天 |

---

### Phase 5: 管理界面与监控（1周）

**目标：** 完善管理界面和监控告警

#### 5.1 管理界面扩展

- 路由策略配置页面
- 实时状态监控页面
- 决策日志查询页面
- API Key 路由配置页面

#### 5.2 监控告警

- 路由决策延迟监控
- 端点健康状态告警
- 成本超预算告警
- 熔断器状态告警

#### 5.3 任务清单

| 任务 | 优先级 | 预计时间 |
|------|--------|----------|
| 路由策略配置页面 | P1 | 2天 |
| 实时状态监控页面 | P1 | 2天 |
| 决策日志查询页面 | P2 | 1天 |
| API Key 路由配置页面 | P1 | 1天 |
| 监控告警配置 | P1 | 1天 |

---

## 四、实施计划

### 4.1 时间规划

| 阶段 | 内容 | 时间 | 依赖 |
|------|------|------|------|
| Phase 1 | 基础设施层建设 | 2周 | 无 |
| Phase 2 | 路由策略层建设 | 2周 | Phase 1 |
| Phase 3 | 路由决策层重构 | 2周 | Phase 1, 2 |
| Phase 4 | 路由执行层增强 | 2周 | Phase 3 |
| Phase 5 | 管理界面与监控 | 1周 | Phase 4 |

**总计：9周**

### 4.2 里程碑

| 里程碑 | 时间 | 交付物 |
|--------|------|--------|
| M1 | 第2周末 | 路由感知能力上线 |
| M2 | 第4周末 | 策略层能力上线 |
| M3 | 第6周末 | 三层联动机制上线 |
| M4 | 第8周末 | 统一网关上线 |
| M5 | 第9周末 | 完整系统上线 |

---

## 五、风险评估

### 5.1 技术风险

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 三层联动复杂度高 | 高 | 分阶段实施，充分测试 |
| 实时状态采集延迟 | 中 | Redis 缓存 + 异步更新 |
| 多目标优化算法性能 | 中 | 预计算 + 缓存 |
| 限流器精度 | 低 | 使用成熟算法实现 |

### 5.2 业务风险

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 策略配置错误 | 高 | 配置校验 + 灰度发布 |
| 路由决策延迟增加 | 中 | 性能优化 + 监控告警 |
| 成本预算超支 | 中 | 实时监控 + 自动熔断 |

---

## 六、验收标准

### 6.1 功能验收

- [ ] 路由感知能力：实时获取供应商状态
- [ ] 策略层：支持 6 种策略类型
- [ ] 决策层：三层联动协同工作
- [ ] 执行层：统一网关 + 可靠性机制
- [ ] 限流：令牌桶限流生效
- [ ] 队列：请求队列管理生效
- [ ] 监控：完整可观测性

### 6.2 性能验收

- [ ] 路由决策延迟 < 10ms (P99)
- [ ] 实时状态更新延迟 < 1s
- [ ] 限流精度误差 < 5%
- [ ] 队列吞吐量 > 1000 QPS

### 6.3 可靠性验收

- [ ] 熔断器正确触发
- [ ] 降级机制正确执行
- [ ] 重试策略正确执行
- [ ] 故障转移正确执行

---

**文档版本：** v1.0  
**创建时间：** 2026-04-22  
**作者：** Trae AI
