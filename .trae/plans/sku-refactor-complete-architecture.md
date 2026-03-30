# 拼脱脱大模型聚合平台完整架构设计文档

**文档版本**: V2.0  
**创建日期**: 2026-03-29  
**文档用途**: 完整架构设计与开发指导

---

## 目录

1. [核心设计理念](#1-核心设计理念)
2. [整体架构设计](#2-整体架构设计)
3. [统一接入层设计](#3-统一接入层设计)
4. [SPU/SKU 与厂商映射实现](#4-spusku-与厂商映射实现)
5. [厂商计费模式适配策略](#5-厂商计费模式适配策略)
6. [完整调用链路设计](#6-完整调用链路设计)
7. [用量上报与计费解耦](#7-用量上报与计费解耦)
8. [缓存策略优化成本](#8-缓存策略优化成本)
9. [厂商接入流程设计](#9-厂商接入流程设计)
10. [适配器开发规范](#10-适配器开发规范)
11. [关键难点与解决方案](#11-关键难点与解决方案)
12. [SPU/SKU 数据模型设计](#12-spusku-数据模型设计)
13. [种子数据设计](#13-种子数据设计)

---

## 1. 核心设计理念

### 1.1 平台价值定位

在 2026 年"Token 即石油"的时代，平台核心价值：

| 价值维度 | 具体实现 |
|----------|----------|
| B端厂商零改造 | 保持现有API和计费模式不变 |
| C端用户统一体验 | 一个API Key访问所有模型，统一账单 |
| 未来扩展性 | 新增厂商只需开发适配器，SPU/SKU配置即可上线 |

### 1.2 核心数据模型关系

```
┌─────────────────────────────────────────────────────────────────┐
│                        核心实体关系                               │
├─────────────────────────────────────────────────────────────────┤
│  SPU = 厂商模型（技术实体）                                        │
│  SKU = 售卖套餐（商业实体）                                        │
│  适配器 = 连接桥梁（协议转换）                                     │
│  计量引擎 = 成本核算（利润管理）                                   │
└─────────────────────────────────────────────────────────────────┘
```

### 1.3 设计目标

- **对冲供应链波动**: 通过预购Token池应对上游涨价
- **降低B端接入复杂度**: 标准化API分销渠道
- **聚合C端长尾流量**: 通过"拼团"模式获取规模效应

---

## 2. 整体架构设计

### 2.1 系统架构图

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              C端用户层                                   │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐    │
│  │   Web前端   │  │   App客户端  │  │   开发者SDK  │  │   第三方应用 │    │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘    │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                            平台层（统一接入层）                           │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │                      统一API网关 (Gateway)                         │  │
│  │  • 认证鉴权  • 限流熔断  • 负载均衡  • 日志追踪  • 协议转换        │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│                                    │                                    │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐    │
│  │ SPU/SKU引擎 │  │ 计量计费引擎 │  │ 适配器管理层 │  │ 缓存优化层  │    │
│  │ • 商品管理  │  │ • Token计量 │  │ • 协议转换  │  │ • 请求缓存  │    │
│  │ • 权益判断  │  │ • 成本核算  │  │ • 认证适配  │  │ • 结果缓存  │    │
│  │ • 配额管理  │  │ • 利润计算  │  │ • 路由选择  │  │ • 智能预热  │    │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘    │
│                                    │                                    │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │                      数据持久层 (Data Layer)                       │  │
│  │  • PostgreSQL (SPU/SKU/订单/用户)  • Redis (缓存/会话/限流)       │  │
│  └──────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                            B端厂商层                                     │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐    │
│  │ 阿里云/通义 │  │ 百度千帆/文心│  │ 火山引擎/豆包│  │ DeepSeek等  │    │
│  │  千问 API   │  │  一言 API   │  │   API       │  │   API       │    │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘    │
└─────────────────────────────────────────────────────────────────────────┘
```

### 2.2 核心模块职责

| 模块 | 职责 | 关键功能 |
|------|------|----------|
| 统一API网关 | 流量入口 | 认证鉴权、限流熔断、负载均衡、日志追踪 |
| SPU/SKU引擎 | 商品管理 | 商品CRUD、权益判断、配额管理 |
| 计量计费引擎 | 成本核算 | Token计量、成本核算、利润计算、对账 |
| 适配器管理层 | 协议转换 | 协议转换、认证适配、路由选择 |
| 缓存优化层 | 成本优化 | 请求缓存、结果缓存、智能预热 |

---

## 3. 统一接入层设计

### 3.1 设计原则

借鉴 OpenRouter、Grab AI Gateway 等行业标杆实践，构建"统一计量与映射层"：

1. **平台主动适配**: 不强制厂商改变API规范
2. **协议透明转换**: 统一格式 ↔ 厂商格式
3. **计费模式映射**: 统一计费 ↔ 厂商计费

### 3.2 统一请求格式

```json
{
  "model": "deepseek-v3",
  "messages": [
    {"role": "user", "content": "Hello"}
  ],
  "temperature": 0.7,
  "max_tokens": 2048,
  "stream": false,
  "user_metadata": {
    "sku_id": "sku_dsv3_1m_onetime",
    "user_id": "user_123",
    "request_id": "req_abc123"
  }
}
```

### 3.3 统一响应格式

```json
{
  "id": "chatcmpl-abc123",
  "object": "chat.completion",
  "created": 1712345678,
  "model": "deepseek-v3",
  "choices": [{
    "index": 0,
    "message": {
      "role": "assistant",
      "content": "Hello! How can I help you?"
    },
    "finish_reason": "stop"
  }],
  "usage": {
    "prompt_tokens": 10,
    "completion_tokens": 20,
    "total_tokens": 30,
    "compute_points": 30.0
  },
  "provider_info": {
    "provider": "deepseek",
    "model_id": "deepseek-chat",
    "latency_ms": 150,
    "cache_hit": false
  }
}
```

### 3.4 API网关核心功能

```go
type APIGateway struct {
    authService    *AuthService
    rateLimiter    *RateLimiter
    router         *Router
    meter          *MeterEngine
    adapterManager *AdapterManager
}

func (g *APIGateway) HandleRequest(ctx *gin.Context) {
    // 1. 认证鉴权
    user, err := g.authService.Authenticate(ctx)
    if err != nil {
        ctx.JSON(401, gin.H{"error": "unauthorized"})
        return
    }
    
    // 2. 限流检查
    if !g.rateLimiter.Allow(user.ID) {
        ctx.JSON(429, gin.H{"error": "rate limit exceeded"})
        return
    }
    
    // 3. 权益判断
    sku, quota, err := g.router.ResolveSKU(user.ID, ctx.Param("model"))
    if err != nil || quota <= 0 {
        ctx.JSON(403, gin.H{"error": "insufficient quota"})
        return
    }
    
    // 4. 路由到适配器
    adapter := g.adapterManager.GetAdapter(sku.Provider)
    response, err := adapter.Forward(ctx.Request, user, sku)
    
    // 5. 异步计量
    go g.meter.Record(user.ID, sku.ID, response.Usage)
    
    ctx.JSON(200, response)
}
```

---

## 4. SPU/SKU 与厂商映射实现

### 4.1 映射关系表

| 平台层 | 映射对象 | 映射字段/逻辑 | 示例 |
|--------|----------|---------------|------|
| SPU | 厂商模型ID | `provider_model_id` 直接存储 | `ERNIE-4.0-8K` |
| SKU | 厂商计费模式 | `billing_strategy` 存储策略 | `TOKEN_BASED` / `HYBRID` |
| 价格 | 厂商Token单价 | 基础成本价 + 平台加成 | ¥0.008/1K → ¥0.012/1K |
| 配额 | 厂商TPM限制 | 聚合多厂商/密钥 | 500 QPM → 2000 QPM |

### 4.2 SPU 厂商对接字段

```sql
-- SPU 表厂商对接字段
ALTER TABLE spus ADD COLUMN provider_model_id VARCHAR(128);
ALTER TABLE spus ADD COLUMN provider_api_endpoint VARCHAR(512);
ALTER TABLE spus ADD COLUMN provider_auth_type VARCHAR(32) DEFAULT 'API_KEY';
ALTER TABLE spus ADD COLUMN provider_billing_type VARCHAR(32);
ALTER TABLE spus ADD COLUMN provider_input_rate DECIMAL(10,6);
ALTER TABLE spus ADD COLUMN provider_output_rate DECIMAL(10,6);
ALTER TABLE spus ADD COLUMN billing_coefficient DECIMAL(5,2) DEFAULT 1.0;
```

### 4.3 映射配置示例

```json
{
  "spu_id": "spu_deepseek_v3",
  "name": "DeepSeek-V3 模型服务",
  "provider": "deepseek",
  "provider_model_id": "deepseek-chat",
  "provider_api_endpoint": "https://api.deepseek.com/v1/chat/completions",
  "provider_auth_type": "API_KEY",
  "provider_billing_type": "FLAT",
  "provider_input_rate": 0.001,
  "provider_output_rate": 0.002,
  "billing_coefficient": 1.0,
  "input_length_ranges": [
    {"min_tokens": 0, "max_tokens": 64000, "label": "64K", "surcharge": 0}
  ],
  "billing_adapter": {
    "type": "flat",
    "cache_enabled": false
  },
  "routing_rules": {
    "auto_route": true,
    "default_range": "64K"
  },
  "batch_inference": {
    "enabled": true,
    "discount_rate": 40,
    "async_only": true
  }
}
```

---

## 5. 厂商计费模式适配策略

### 5.1 厂商计费特征分析

| 厂商 | 计费特征 | 平台适配策略 | 实现方式 |
|------|----------|--------------|----------|
| 阿里云/通义千问 | 输入输出分开计价 | 分别统计input/output tokens | 从响应头提取字段 |
| 百度千帆 | 模型版本不同计价 | SPU级别区分版本 | 不同版本对应不同SPU |
| 火山引擎/豆包 | 分段计价（输入长度区间） | 请求前计算长度，动态路由 | 适配器中实现长度判断 |
| DeepSeek | 统一单价，缓存优惠 | 平台侧缓存命中判断 | 维护请求指纹 |
| 华为云 | 用户自行配置鉴权 | 密钥托管，统一调用 | 加密存储API Key |

### 5.2 计费类型定义

```go
type BillingType string

const (
    BillingTypeFlat     BillingType = "flat"     // 统一计费
    BillingTypeSegment  BillingType = "segment"  // 分段计费
    BillingTypeTiered   BillingType = "tiered"   // 阶梯计费
    BillingTypeHybrid   BillingType = "hybrid"   // 混合计费
)

type BillingAdapter struct {
    Type           BillingType    `json:"type"`
    SegmentConfig  []SegmentRule  `json:"segment_config,omitempty"`
    CacheEnabled   bool           `json:"cache_enabled"`
    CacheDiscount  float64        `json:"cache_discount_rate,omitempty"`
}

type SegmentRule struct {
    InputRange string  `json:"input_range"`
    Multiplier float64 `json:"multiplier"`
}
```

### 5.3 成本计算引擎

```go
type CostCalculator struct {
    db *sql.DB
}

func (c *CostCalculator) Calculate(
    ctx context.Context,
    spuID int,
    inputTokens int,
    outputTokens int,
    inputLength int,
) (providerCost float64, platformRevenue float64, err error) {
    var spu SPU
    err = c.db.QueryRowContext(ctx,
        "SELECT provider_input_rate, provider_output_rate, billing_coefficient, billing_adapter FROM spus WHERE id = $1",
        spuID,
    ).Scan(&spu.ProviderInputRate, &spu.ProviderOutputRate, &spu.BillingCoefficient, &spu.BillingAdapter)
    if err != nil {
        return 0, 0, err
    }
    
    var multiplier float64 = 1.0
    
    if spu.BillingAdapter.Type == "segment" {
        multiplier = c.getSegmentMultiplier(spu.BillingAdapter.SegmentConfig, inputLength)
    }
    
    providerCost = (float64(inputTokens) * spu.ProviderInputRate +
                   float64(outputTokens) * spu.ProviderOutputRate) * multiplier
    
    platformRevenue = providerCost * (spu.BillingCoefficient - 1.0)
    
    return providerCost, platformRevenue, nil
}

func (c *CostCalculator) getSegmentMultiplier(rules []SegmentRule, inputLength int) float64 {
    for _, rule := range rules {
        if strings.Contains(rule.InputRange, "K") {
            if c.matchRange(rule.InputRange, inputLength) {
                return rule.Multiplier
            }
        }
    }
    return 1.0
}
```

---

## 6. 完整调用链路设计

### 6.1 调用时序图

```
┌────────┐     ┌────────┐     ┌────────┐     ┌────────┐     ┌────────┐     ┌────────┐
│ C端用户 │     │API网关 │     │SKU引擎 │     │计量引擎│     │ 适配器 │     │B端厂商 │
└───┬────┘     └───┬────┘     └───┬────┘     └───┬────┘     └───┬────┘     └───┬────┘
    │              │              │              │              │              │
    │ 1.API请求    │              │              │              │              │
    │─────────────>│              │              │              │              │
    │              │              │              │              │              │
    │              │ 2.验证权益   │              │              │              │
    │              │─────────────>│              │              │              │
    │              │              │              │              │              │
    │              │ 3.返回配额   │              │              │              │
    │              │<─────────────│              │              │              │
    │              │              │              │              │              │
    │              │ 4.转发请求   │              │              │              │
    │              │──────────────────────────────────────────>│              │
    │              │              │              │              │              │
    │              │              │              │              │ 5.调用厂商API│
    │              │              │              │              │─────────────>│
    │              │              │              │              │              │
    │              │              │              │              │ 6.返回响应   │
    │              │              │              │              │<─────────────│
    │              │              │              │              │              │
    │              │              │              │ 7.异步上报   │              │
    │              │              │              │<─────────────│              │
    │              │              │              │              │              │
    │              │ 8.转换响应   │              │              │              │
    │              │<──────────────────────────────────────────│              │
    │              │              │              │              │              │
    │ 9.返回结果   │              │              │              │              │
    │<─────────────│              │              │              │              │
    │              │              │              │              │              │
```

### 6.2 详细流程说明

| 步骤 | 操作 | 说明 |
|------|------|------|
| 1 | API请求 | C端用户携带API Key发起请求 |
| 2 | 验证权益 | 查询用户购买的SKU，判断配额/余额 |
| 3 | 返回配额 | 返回用户可用配额或余额不足错误 |
| 4 | 转发请求 | 将统一格式请求转发给适配器 |
| 5 | 调用厂商API | 适配器转换协议后调用厂商API |
| 6 | 返回响应 | 厂商返回结果和Token用量 |
| 7 | 异步上报 | 异步上报用量到计量引擎 |
| 8 | 转换响应 | 将厂商格式转换为统一格式 |
| 9 | 返回结果 | 返回最终结果给C端用户 |

---

## 7. 用量上报与计费解耦

### 7.1 异步计量设计

参考 Stripe AI SDK 的 meter 设计，用量上报采用异步非阻塞模式：

```go
type MeterMiddleware struct {
    meterEngine *MeterEngine
    queue       chan *UsageReport
}

type UsageReport struct {
    UserID        int64
    SKUID        int64
    OrderID      int64
    InputTokens  int64
    OutputTokens int64
    ProviderCost float64
    RequestID    string
    Timestamp    time.Time
}

func (m *MeterMiddleware) AfterRequest(
    ctx context.Context,
    request *Request,
    response *Response,
    providerResponse *ProviderResponse,
) error {
    usage := extractTokenUsage(providerResponse)
    
    report := &UsageReport{
        UserID:        request.UserID,
        SKUID:        request.SKUID,
        InputTokens:  usage.InputTokens,
        OutputTokens: usage.OutputTokens,
        ProviderCost: usage.Cost,
        RequestID:    request.RequestID,
        Timestamp:    time.Now(),
    }
    
    select {
    case m.queue <- report:
    default:
        go func() {
            m.queue <- report
        }()
    }
    
    return nil
}

func (m *MeterMiddleware) StartWorker() {
    for report := range m.queue {
        m.processReport(report)
    }
}

func (m *MeterMiddleware) processReport(report *UsageReport) {
    ctx := context.Background()
    
    tx, _ := m.meterEngine.db.BeginTx(ctx, nil)
    defer tx.Rollback()
    
    // 1. 扣减用户余额
    _, _ = tx.ExecContext(ctx,
        "UPDATE compute_point_accounts SET balance = balance - $1 WHERE user_id = $2",
        report.ComputePoints, report.UserID,
    )
    
    // 2. 记录交易
    _, _ = tx.ExecContext(ctx,
        `INSERT INTO compute_point_transactions (user_id, type, amount, balance_after, sku_id, description) 
         SELECT $1, 'usage', $2, balance - $2, $3, $4 FROM compute_point_accounts WHERE user_id = $1`,
        report.UserID, report.ComputePoints, report.SKUID, "API调用消耗",
    )
    
    // 3. 记录厂商成本
    _, _ = tx.ExecContext(ctx,
        `INSERT INTO provider_costs (provider, spu_id, input_tokens, output_tokens, cost, request_id, created_at) 
         VALUES ($1, $2, $3, $4, $5, $6, $7)`,
        report.Provider, report.SPUID, report.InputTokens, report.OutputTokens, report.ProviderCost, report.RequestID, report.Timestamp,
    )
    
    tx.Commit()
}
```

### 7.2 计量数据流

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   API调用   │────>│  用量提取   │────>│  异步队列   │────>│  计量引擎   │
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
                                                                   │
                                                                   ▼
                        ┌─────────────────────────────────────────────────┐
                        │                   数据持久化                     │
                        ├─────────────────────────────────────────────────┤
                        │ • compute_point_transactions (用户消费记录)      │
                        │ • provider_costs (厂商成本记录)                  │
                        │ • usage_statistics (用量统计)                    │
                        └─────────────────────────────────────────────────┘
```

---

## 8. 缓存策略优化成本

### 8.1 缓存架构

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           缓存优化层                                     │
├─────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                     │
│  │ 请求指纹缓存 │  │ 结果缓存    │  │ 智能预热    │                     │
│  │ • SHA256    │  │ • Redis     │  │ • 热点预测  │                     │
│  │ • TTL: 24h  │  │ • TTL: 1h   │  │ • 定时预热  │                     │
│  └─────────────┘  └─────────────┘  └─────────────┘                     │
└─────────────────────────────────────────────────────────────────────────┘
```

### 8.2 缓存命中策略

```go
type CacheStrategy struct {
    redis       *redis.Client
    cacheConfig CacheConfig
}

type CacheConfig struct {
    Enabled          bool    `json:"enabled"`
    TTLSeconds       int     `json:"ttl_seconds"`
    DiscountRate     float64 `json:"cache_discount_rate"`
    MaxCacheSize     int64   `json:"max_cache_size"`
}

func (c *CacheStrategy) GetOrForward(
    ctx context.Context,
    request *Request,
    forwardFunc func() (*Response, error),
) (*Response, error) {
    if !c.cacheConfig.Enabled {
        return forwardFunc()
    }
    
    cacheKey := c.generateCacheKey(request)
    
    cached, err := c.redis.Get(ctx, cacheKey).Result()
    if err == nil {
        var response Response
        json.Unmarshal([]byte(cached), &response)
        response.CacheHit = true
        
        go c.recordCacheHit(request.UserID, request.SKUID, response.Usage.ComputePoints)
        
        return &response, nil
    }
    
    response, err := forwardFunc()
    if err != nil {
        return nil, err
    }
    
    if response.Usage.TotalTokens > 0 {
        responseJSON, _ := json.Marshal(response)
        c.redis.Set(ctx, cacheKey, responseJSON, time.Duration(c.cacheConfig.TTLSeconds)*time.Second)
    }
    
    return response, nil
}

func (c *CacheStrategy) generateCacheKey(request *Request) string {
    content := fmt.Sprintf("%s:%s:%v", request.Model, request.Messages, request.Temperature)
    hash := sha256.Sum256([]byte(content))
    return fmt.Sprintf("cache:%x", hash)
}
```

### 8.3 缓存利润分配

| 场景 | 成本 | 收入 | 利润分配 |
|------|------|------|----------|
| 缓存命中 | 0 | 用户正常扣费 | 100%平台利润 |
| 缓存未命中 | 厂商成本 | 用户正常扣费 | 差价利润 |
| 厂商缓存命中 | 厂商优惠价 | 用户正常扣费 | 部分返还用户 |

---

## 9. 厂商接入流程设计

### 9.1 标准接入流程

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        厂商接入标准流程                                   │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  Phase 1: 调研                                                          │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │ • 分析厂商API规范、计费模式、认证方式                             │   │
│  │ • 产出: 厂商对接文档                                              │   │
│  │ • 负责人: 产品/技术                                               │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                              │                                          │
│                              ▼                                          │
│  Phase 2: 配置SPU                                                       │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │ • 在平台录入模型信息、厂商参数                                    │   │
│  │ • 产出: SPU数据记录                                               │   │
│  │ • 负责人: 运营                                                    │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                              │                                          │
│                              ▼                                          │
│  Phase 3: 配置SKU                                                       │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │ • 设计套餐规格、定价策略                                          │   │
│  │ • 产出: SKU数据记录                                               │   │
│  │ • 负责人: 运营/产品                                               │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                              │                                          │
│                              ▼                                          │
│  Phase 4: 适配器开发                                                    │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │ • 实现协议转换、认证适配                                          │   │
│  │ • 产出: 适配器代码                                                │   │
│  │ • 负责人: 开发                                                    │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                              │                                          │
│                              ▼                                          │
│  Phase 5: 测试验证                                                      │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │ • 端到端调用测试、计费验证                                        │   │
│  │ • 产出: 测试报告                                                  │   │
│  │ • 负责人: QA                                                      │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                              │                                          │
│                              ▼                                          │
│  Phase 6: 上线发布                                                      │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │ • 开放SKU售卖、监控告警                                           │   │
│  │ • 产出: 上线记录                                                  │   │
│  │ • 负责人: 运维                                                    │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### 9.2 接入检查清单

| 检查项 | 说明 | 状态 |
|--------|------|------|
| API文档 | 厂商API文档已阅读并理解 | ☐ |
| 认证方式 | API Key / AK-SK / OAuth 已确认 | ☐ |
| 计费模式 | 计费规则已分析并记录 | ☐ |
| SPU配置 | 模型信息已录入平台 | ☐ |
| SKU配置 | 套餐规格已配置 | ☐ |
| 适配器 | 协议转换已实现 | ☐ |
| 单元测试 | 适配器测试通过 | ☐ |
| 集成测试 | 端到端测试通过 | ☐ |
| 计费验证 | 成本计算正确 | ☐ |
| 监控配置 | 告警规则已配置 | ☐ |

---

## 10. 适配器开发规范

### 10.1 适配器接口定义

```go
type ProviderAdapter interface {
    TransformRequest(ctx context.Context, unifiedRequest *UnifiedRequest, userSKU *SKU) (*ProviderRequest, error)
    
    TransformResponse(ctx context.Context, providerResponse *ProviderResponse) (*UnifiedResponse, error)
    
    ExtractUsage(ctx context.Context, providerResponse *ProviderResponse) (*TokenUsage, error)
    
    CalculateCost(ctx context.Context, usage *TokenUsage, spuConfig *SPU) (float64, error)
    
    GetAuthHeaders(ctx context.Context, credentials *ProviderCredentials) (map[string]string, error)
    
    GetProviderName() string
    
    HealthCheck(ctx context.Context) error
}

type UnifiedRequest struct {
    Model       string          `json:"model"`
    Messages    []Message       `json:"messages"`
    Temperature float64         `json:"temperature,omitempty"`
    MaxTokens   int             `json:"max_tokens,omitempty"`
    Stream      bool            `json:"stream,omitempty"`
    Metadata    *RequestMetadata `json:"metadata,omitempty"`
}

type UnifiedResponse struct {
    ID          string         `json:"id"`
    Object      string         `json:"object"`
    Created     int64          `json:"created"`
    Model       string         `json:"model"`
    Choices     []Choice       `json:"choices"`
    Usage       *TokenUsage    `json:"usage"`
    ProviderInfo *ProviderInfo `json:"provider_info,omitempty"`
}

type TokenUsage struct {
    PromptTokens     int64   `json:"prompt_tokens"`
    CompletionTokens int64   `json:"completion_tokens"`
    TotalTokens      int64   `json:"total_tokens"`
    ComputePoints    float64 `json:"compute_points"`
}
```

### 10.2 适配器实现示例

```go
type DeepSeekAdapter struct {
    baseURL string
    apiKey  string
}

func NewDeepSeekAdapter(apiKey string) *DeepSeekAdapter {
    return &DeepSeekAdapter{
        baseURL: "https://api.deepseek.com/v1",
        apiKey:  apiKey,
    }
}

func (a *DeepSeekAdapter) GetProviderName() string {
    return "deepseek"
}

func (a *DeepSeekAdapter) TransformRequest(ctx context.Context, unified *UnifiedRequest, sku *SKU) (*ProviderRequest, error) {
    return &ProviderRequest{
        URL:    a.baseURL + "/chat/completions",
        Method: "POST",
        Headers: map[string]string{
            "Content-Type":  "application/json",
            "Authorization": "Bearer " + a.apiKey,
        },
        Body: map[string]interface{}{
            "model":       unified.Model,
            "messages":    unified.Messages,
            "temperature": unified.Temperature,
            "max_tokens":  unified.MaxTokens,
            "stream":      unified.Stream,
        },
    }, nil
}

func (a *DeepSeekAdapter) TransformResponse(ctx context.Context, provider *ProviderResponse) (*UnifiedResponse, error) {
    var resp struct {
        ID      string `json:"id"`
        Object  string `json:"object"`
        Created int64  `json:"created"`
        Model   string `json:"model"`
        Choices []struct {
            Index        int `json:"index"`
            Message      Message `json:"message"`
            FinishReason string `json:"finish_reason"`
        } `json:"choices"`
        Usage struct {
            PromptTokens     int64 `json:"prompt_tokens"`
            CompletionTokens int64 `json:"completion_tokens"`
            TotalTokens      int64 `json:"total_tokens"`
        } `json:"usage"`
    }
    
    if err := json.Unmarshal(provider.Body, &resp); err != nil {
        return nil, err
    }
    
    choices := make([]Choice, len(resp.Choices))
    for i, c := range resp.Choices {
        choices[i] = Choice{
            Index:        c.Index,
            Message:      c.Message,
            FinishReason: c.FinishReason,
        }
    }
    
    return &UnifiedResponse{
        ID:      resp.ID,
        Object:  resp.Object,
        Created: resp.Created,
        Model:   resp.Model,
        Choices: choices,
        Usage: &TokenUsage{
            PromptTokens:     resp.Usage.PromptTokens,
            CompletionTokens: resp.Usage.CompletionTokens,
            TotalTokens:      resp.Usage.TotalTokens,
        },
        ProviderInfo: &ProviderInfo{
            Provider:   "deepseek",
            ModelID:    resp.Model,
            LatencyMs:  provider.LatencyMs,
            CacheHit:   false,
        },
    }, nil
}

func (a *DeepSeekAdapter) ExtractUsage(ctx context.Context, provider *ProviderResponse) (*TokenUsage, error) {
    resp, err := a.TransformResponse(ctx, provider)
    if err != nil {
        return nil, err
    }
    return resp.Usage, nil
}

func (a *DeepSeekAdapter) CalculateCost(ctx context.Context, usage *TokenUsage, spu *SPU) (float64, error) {
    inputCost := float64(usage.PromptTokens) * spu.ProviderInputRate
    outputCost := float64(usage.CompletionTokens) * spu.ProviderOutputRate
    return inputCost + outputCost, nil
}

func (a *DeepSeekAdapter) GetAuthHeaders(ctx context.Context, cred *ProviderCredentials) (map[string]string, error) {
    return map[string]string{
        "Authorization": "Bearer " + cred.APIKey,
    }, nil
}

func (a *DeepSeekAdapter) HealthCheck(ctx context.Context) error {
    return nil
}
```

### 10.3 适配器注册机制

```go
type AdapterManager struct {
    adapters map[string]ProviderAdapter
    db       *sql.DB
}

func NewAdapterManager(db *sql.DB) *AdapterManager {
    return &AdapterManager{
        adapters: make(map[string]ProviderAdapter),
        db:       db,
    }
}

func (m *AdapterManager) Register(provider string, adapter ProviderAdapter) {
    m.adapters[provider] = adapter
}

func (m *AdapterManager) GetAdapter(provider string) ProviderAdapter {
    return m.adapters[provider]
}

func (m *AdapterManager) Forward(ctx context.Context, request *UnifiedRequest, user *User, sku *SKU) (*UnifiedResponse, error) {
    adapter, ok := m.adapters[sku.Provider]
    if !ok {
        return nil, fmt.Errorf("adapter not found for provider: %s", sku.Provider)
    }
    
    providerReq, err := adapter.TransformRequest(ctx, request, sku)
    if err != nil {
        return nil, err
    }
    
    providerResp, err := m.doRequest(ctx, providerReq)
    if err != nil {
        return nil, err
    }
    
    return adapter.TransformResponse(ctx, providerResp)
}
```

---

## 11. 关键难点与解决方案

### 11.1 成本系数的动态计算

**问题**: 不同厂商、不同模型的Token单价差异巨大，平台需要精确计算成本和利润。

**解决方案**:

```go
type DynamicCostEngine struct {
    db            *sql.DB
    cache         *redis.Client
    coefficientCache map[int64]float64
    mu            sync.RWMutex
}

func (e *DynamicCostEngine) CalculateDynamicCost(
    ctx context.Context,
    spuID int64,
    inputTokens int64,
    outputTokens int64,
    inputLength int,
) (providerCost float64, userCharge float64, platformRevenue float64, err error) {
    e.mu.RLock()
    coefficient, ok := e.coefficientCache[spuID]
    e.mu.RUnlock()
    
    if !ok {
        var spu SPU
        err = e.db.QueryRowContext(ctx,
            "SELECT provider_input_rate, provider_output_rate, billing_coefficient, billing_adapter, input_length_ranges FROM spus WHERE id = $1",
            spuID,
        ).Scan(&spu.ProviderInputRate, &spu.ProviderOutputRate, &spu.BillingCoefficient, &spu.BillingAdapter, &spu.InputLengthRanges)
        if err != nil {
            return 0, 0, 0, err
        }
        
        coefficient = spu.BillingCoefficient
        e.mu.Lock()
        e.coefficientCache[spuID] = coefficient
        e.mu.Unlock()
    }
    
    var multiplier float64 = 1.0
    
    providerCost = float64(inputTokens)*spu.ProviderInputRate + float64(outputTokens)*spu.ProviderOutputRate
    providerCost *= multiplier
    
    userCharge = providerCost * coefficient
    platformRevenue = userCharge - providerCost
    
    return providerCost, userCharge, platformRevenue, nil
}

func (e *DynamicCostEngine) UpdateCoefficient(spuID int64, newCoefficient float64) {
    e.mu.Lock()
    defer e.mu.Unlock()
    e.coefficientCache[spuID] = newCoefficient
}
```

### 11.2 限流与配额管理

**问题**: 多个C端用户共享B端厂商的API配额，单个用户可能耗尽所有配额。

**解决方案**:

```go
type MultiLayerRateLimiter struct {
    redis       *redis.Client
    userLimit   int
    skuLimit    int
    providerLimit int
}

func (l *MultiLayerRateLimiter) Allow(ctx context.Context, userID, skuID int64, provider string) bool {
    now := time.Now().Unix()
    window := 60
    
    userKey := fmt.Sprintf("ratelimit:user:%d:%d", userID, now/window)
    skuKey := fmt.Sprintf("ratelimit:sku:%d:%d", skuID, now/window)
    providerKey := fmt.Sprintf("ratelimit:provider:%s:%d", provider, now/window)
    
    pipe := l.redis.Pipeline()
    userCount := pipe.Incr(ctx, userKey)
    skuCount := pipe.Incr(ctx, skuKey)
    providerCount := pipe.Incr(ctx, providerKey)
    pipe.Expire(ctx, userKey, time.Duration(window)*time.Second)
    pipe.Expire(ctx, skuKey, time.Duration(window)*time.Second)
    pipe.Expire(ctx, providerKey, time.Duration(window)*time.Second)
    _, err := pipe.Exec(ctx)
    if err != nil {
        return false
    }
    
    if userCount.Val() > int64(l.userLimit) {
        return false
    }
    if skuCount.Val() > int64(l.skuLimit) {
        return false
    }
    if providerCount.Val() > int64(l.providerLimit) {
        return false
    }
    
    return true
}

type FairQueue struct {
    queues map[string]*PriorityQueue
    mu     sync.RWMutex
}

func (q *FairQueue) Enqueue(userID string, request *Request, weight int) {
    q.mu.Lock()
    defer q.mu.Unlock()
    
    if _, ok := q.queues[userID]; !ok {
        q.queues[userID] = NewPriorityQueue()
    }
    
    q.queues[userID].Push(request, weight)
}

func (q *FairQueue) Dequeue() *Request {
    q.mu.Lock()
    defer q.mu.Unlock()
    
    var selected *Request
    var selectedUserID string
    minSize := math.MaxInt32
    
    for userID, queue := range q.queues {
        if queue.Size() > 0 && queue.Size() < minSize {
            minSize = queue.Size()
            selectedUserID = userID
        }
    }
    
    if selectedUserID != "" {
        selected = q.queues[selectedUserID].Pop()
    }
    
    return selected
}
```

### 11.3 厂商账单对账

**问题**: 平台需与多家厂商分别对账，确保账单准确。

**解决方案**:

```go
type ReconciliationEngine struct {
    db *sql.DB
}

type DailyReconciliation struct {
    Provider       string
    Date           time.Time
    PlatformCost   float64
    ProviderBill   float64
    Difference     float64
    Status         string
}

func (e *ReconciliationEngine) GenerateDailyReport(ctx context.Context, provider string, date time.Time) (*DailyReconciliation, error) {
    var platformCost float64
    err := e.db.QueryRowContext(ctx,
        `SELECT COALESCE(SUM(cost), 0) FROM provider_costs 
         WHERE provider = $1 AND DATE(created_at) = $2`,
        provider, date,
    ).Scan(&platformCost)
    if err != nil {
        return nil, err
    }
    
    providerBill, err := e.fetchProviderBill(ctx, provider, date)
    if err != nil {
        providerBill = 0
    }
    
    difference := platformCost - providerBill
    status := "matched"
    if math.Abs(difference) > platformCost*0.01 {
        status = "mismatch"
    }
    
    report := &DailyReconciliation{
        Provider:     provider,
        Date:         date,
        PlatformCost: platformCost,
        ProviderBill: providerBill,
        Difference:   difference,
        Status:       status,
    }
    
    _, err = e.db.ExecContext(ctx,
        `INSERT INTO reconciliations (provider, date, platform_cost, provider_bill, difference, status) 
         VALUES ($1, $2, $3, $4, $5, $6)`,
        report.Provider, report.Date, report.PlatformCost, report.ProviderBill, report.Difference, report.Status,
    )
    
    return report, err
}

func (e *ReconciliationEngine) fetchProviderBill(ctx context.Context, provider string, date time.Time) (float64, error) {
    return 0, nil
}

func (e *ReconciliationEngine) CheckAndAlert(ctx context.Context) error {
    rows, err := e.db.QueryContext(ctx,
        `SELECT provider, date, difference FROM reconciliations 
         WHERE status = 'mismatch' AND date >= CURRENT_DATE - INTERVAL '7 days'`,
    )
    if err != nil {
        return err
    }
    defer rows.Close()
    
    for rows.Next() {
        var provider string
        var date time.Time
        var difference float64
        rows.Scan(&provider, &date, &difference)
        
        fmt.Printf("ALERT: Provider %s has reconciliation difference of %.2f on %s\n", provider, difference, date)
    }
    
    return nil
}
```

---

## 12. SPU/SKU 数据模型设计

### 12.1 核心关系图

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        核心数据模型关系                                   │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────┐       1:N        ┌─────────────┐                      │
│  │     SPU     │─────────────────>│     SKU     │                      │
│  │ (厂商模型)  │                  │ (售卖套餐)  │                      │
│  └─────────────┘                  └─────────────┘                      │
│         │                                │                              │
│         │ N:1                            │ 1:N                          │
│         ▼                                ▼                              │
│  ┌─────────────┐                  ┌─────────────┐                      │
│  │   Provider  │                  │   Orders    │                      │
│  │  (厂商配置) │                  │   (订单)    │                      │
│  └─────────────┘                  └─────────────┘                      │
│                                          │                              │
│                                          │ 1:N                          │
│                                          ▼                              │
│                                   ┌─────────────┐                      │
│                                   │ Transactions│                      │
│                                   │  (交易记录) │                      │
│                                   └─────────────┘                      │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### 12.2 完整表结构

详见: [backend/migrations/015_sku_refactor.sql](file:///Users/4seven/workspace/pintuotuo/backend/migrations/015_sku_refactor.sql)

### 12.3 关键字段说明

**SPU 表关键字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| `provider_model_id` | VARCHAR(128) | 厂商侧模型标识 |
| `provider_api_endpoint` | VARCHAR(512) | 厂商API地址 |
| `provider_billing_type` | VARCHAR(32) | 厂商计费类型 |
| `provider_input_rate` | DECIMAL(10,6) | 厂商输入Token单价 |
| `provider_output_rate` | DECIMAL(10,6) | 厂商输出Token单价 |
| `billing_coefficient` | DECIMAL(5,2) | 平台成本系数 |
| `input_length_ranges` | JSONB | 输入长度区间配置 |
| `billing_adapter` | JSONB | 计费适配器配置 |
| `routing_rules` | JSONB | 智能路由规则 |
| `batch_inference` | JSONB | 批量推理配置 |

**SKU 表关键字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| `spu_id` | INT | 关联SPU |
| `sku_type` | VARCHAR(50) | SKU类型 |
| `token_amount` | BIGINT | Token数量 |
| `compute_points` | DECIMAL(15,2) | 算力点数量 |
| `subscription_period` | VARCHAR(50) | 订阅周期 |
| `retail_price` | DECIMAL(10,2) | 零售价 |
| `wholesale_price` | DECIMAL(10,2) | 批发价 |
| `group_enabled` | BOOLEAN | 是否支持拼团 |
| `group_discount_rate` | DECIMAL(5,2) | 拼团折扣率 |

---

## 13. 种子数据设计

### 13.1 厂商配置

| 厂商代码 | 厂商名称 | API格式 | 计费类型 |
|----------|----------|---------|----------|
| `openai` | OpenAI | openai | flat |
| `anthropic` | Anthropic | anthropic | flat |
| `deepseek` | DeepSeek | openai | flat |
| `zhipu` | 智谱AI | openai | flat |
| `baidu` | 百度千帆 | baidu | segment |
| `bytedance` | 字节跳动/火山引擎 | openai | segment |
| `alibaba` | 阿里云 | openai | tiered |

### 13.2 SPU 配置示例

| SPU代码 | 名称 | 厂商 | 层级 | 算力点系数 |
|---------|------|------|------|------------|
| DEEPSEEK-V3 | DeepSeek V3 模型服务 | deepseek | lite | 1.0 |
| DEEPSEEK-V3-PRO | DeepSeek V3 Pro | deepseek | pro | 2.5 |
| GLM-4 | GLM-4 模型服务 | zhipu | lite | 1.2 |
| ERNIE-4 | 文心一言 ERNIE 4.0 | baidu | pro | 3.0 |
| DOUBAO-PRO | 豆包 Pro 模型服务 | bytedance | pro | 2.5 |

### 13.3 SKU 配置示例

| SKU代码 | SPU | 类型 | Token数量 | 零售价 | 拼团折扣 |
|---------|-----|------|-----------|--------|----------|
| DEEPSEEK-V3-100K | DEEPSEEK-V3 | token_pack | 100,000 | ¥9.90 | 20% |
| DEEPSEEK-V3-500K | DEEPSEEK-V3 | token_pack | 500,000 | ¥39.90 | 25% |
| DEEPSEEK-V3-1M | DEEPSEEK-V3 | token_pack | 1,000,000 | ¥69.90 | 30% |
| DEEPSEEK-V3-MONTHLY | DEEPSEEK-V3 | subscription | - | ¥99.00 | 15% |
| DEEPSEEK-V3-TRIAL | DEEPSEEK-V3 | trial | 10,000 | ¥0 | - |

---

## 附录

### A. 技术栈

| 层级 | 技术选型 |
|------|----------|
| 后端框架 | Go + Gin |
| 数据库 | PostgreSQL |
| 缓存 | Redis |
| 消息队列 | 内置异步队列 |
| 监控 | Prometheus + Grafana |

### B. 性能指标

| 指标 | 目标值 |
|------|--------|
| API响应时间 | < 100ms (不含厂商调用) |
| 缓存命中率 | > 30% |
| 计量延迟 | < 1s (异步) |
| 对账准确率 | > 99.9% |

### C. 安全措施

| 措施 | 说明 |
|------|------|
| API Key 加密存储 | AES-256 加密 |
| 传输加密 | HTTPS/TLS 1.3 |
| 访问控制 | RBAC 权限模型 |
| 审计日志 | 全链路追踪 |

---

**文档结束**

通过以上设计，平台实现了：

✅ **B端厂商零改造**: 保持现有API和计费模式不变  
✅ **C端用户统一体验**: 一个API Key访问所有模型，统一账单  
✅ **未来扩展性**: 新增厂商只需开发适配器，SPU/SKU配置即可上线
