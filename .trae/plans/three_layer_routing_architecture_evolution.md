# 三层路由架构演进开发计划

## 概述

本计划旨在将当前分散在 `api_proxy.go` 中的路由和转发逻辑，重构为完整的三层路由架构，实现职责清晰、可测试、可扩展的代码结构。

---

## 目标架构

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           api_proxy.go (业务编排层)                          │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌────────────────────┐ │
│  │ 用户认证     │ │ 权益校验     │ │ 计费处理     │ │ 调用路由Pipeline   │ │
│  └──────────────┘ └──────────────┘ └──────────────┘ └────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
                                       │
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                     ThreeLayerRoutingPipeline (路由编排层)                   │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────────────────────────────┐ │
│  │ 策略层       │ │ 决策层       │ │ 执行层 (新增)                         │ │
│  │ StrategyLayer│ │ DecisionLayer│ │ ExecutionLayer                       │ │
│  │              │ │              │ │                                      │ │
│  │ DefineGoal() │ │ SelectKey()  │ │ Execute() ─────────────────────────┐ │ │
│  └──────────────┘ └──────────────┘ │                                  ▼ │ │
│                                     │  ┌────────────────────────────────┐ │ │
│                                     │  │ ExecutionEngine                │ │ │
│                                     │  │ (封装HTTP转发能力)             │ │ │
│                                     │  └────────────────────────────────┘ │ │
│                                     └──────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 执行顺序优化

### 当前顺序（问题）
```
1. 用户认证 → 2. 请求解析 → 3. 权益校验(第一轮) → 4. 余额检查 → 
5. 计费预扣 → 6. 商户ID解析 → 7. 权益校验(第二轮) → 8. 智能路由 → 
9. API Key选择 → 10. 流量转发
```

### 优化后顺序
```
第一阶段：准入验证（快速失败）
  1. 用户认证
  2. 请求解析
  3. 权益校验（合并为一轮）
  4. 余额检查（快速检查 > 0）

第二阶段：路由决策（三层路由）
  5. 策略层 - 确定路由目标
  6. 决策层 - 选择最优API Key
  7. 执行层 - 准备执行参数

第三阶段：计费与执行
  8. 计费预扣（使用准确价格）
  9. 流量转发
  10. 计费结算
```

---

## 开发任务

### Phase 1: 基础设施建设（预计 2 天）

#### 1.1 创建 ExecutionEngine（执行引擎）

**文件**: `backend/services/execution_engine.go`

```go
// 核心结构
type ExecutionEngine struct {
    httpClient    *http.Client
    retryPolicy   RetryPolicy
    gatewayConfig *GatewayConfig
}

type ExecutionInput struct {
    Provider      string
    Model         string
    APIKey        string
    EndpointURL   string
    RequestFormat string
    RequestBody   []byte
    Headers       map[string]string
}

type ExecutionResult struct {
    Success      bool
    StatusCode   int
    LatencyMs    int
    ResponseBody []byte
    ErrorMessage string
    Usage        *TokenUsage
    Provider     string
    ActualModel  string
}

// 核心方法
func (e *ExecutionEngine) Execute(ctx context.Context, input *ExecutionInput) (*ExecutionResult, error)
func (e *ExecutionEngine) ExecuteWithRetry(ctx context.Context, input *ExecutionInput) (*ExecutionResult, error)
func (e *ExecutionEngine) buildHTTPRequest(ctx context.Context, input *ExecutionInput) (*http.Request, error)
func (e *ExecutionEngine) parseResponse(resp *http.Response) (*ExecutionResult, error)
```

**任务清单**:
- [ ] 创建 `execution_engine.go` 文件
- [ ] 实现 `ExecutionEngine` 结构体
- [ ] 实现 `Execute()` 方法（基础转发）
- [ ] 实现 `ExecuteWithRetry()` 方法（含重试逻辑）
- [ ] 实现 `buildHTTPRequest()` 方法（构建请求）
- [ ] 实现 `parseResponse()` 方法（解析响应）
- [ ] 添加单元测试 `execution_engine_test.go`

#### 1.2 创建 ExecutionLayer（执行层）

**文件**: `backend/services/execution_layer.go`

```go
// 核心结构
type ExecutionLayer struct {
    engine *ExecutionEngine
    db     *sql.DB
}

type ExecutionLayerInput struct {
    RoutingDecision *RoutingDecision
    RequestBody     []byte
    ProviderConfig  *ProviderRuntimeConfig
    DecryptedAPIKey string
}

type ExecutionLayerOutput struct {
    Result      *ExecutionResult
    Decision    *RoutingDecision
    DurationMs  int
}

// 核心方法
func (l *ExecutionLayer) Execute(ctx context.Context, input *ExecutionLayerInput) (*ExecutionLayerOutput, error)
func (l *ExecutionLayer) prepareExecutionInput(decision *RoutingDecision, reqBody []byte) (*ExecutionInput, error)
func (l *ExecutionLayer) recordExecutionResult(decision *RoutingDecision, result *ExecutionResult)
```

**任务清单**:
- [ ] 创建 `execution_layer.go` 文件
- [ ] 实现 `ExecutionLayer` 结构体
- [ ] 实现 `Execute()` 方法
- [ ] 实现 `prepareExecutionInput()` 方法
- [ ] 实现 `recordExecutionResult()` 方法
- [ ] 添加单元测试 `execution_layer_test.go`

---

### Phase 2: Pipeline 重构（预计 2 天）

#### 2.1 重构 ThreeLayerRoutingPipeline

**文件**: `backend/services/three_layer_pipeline.go`

**修改内容**:
1. 新增 `executeExecutionLayer()` 方法
2. 修改 `Execute()` 方法，增加第三层执行
3. 移除 `RecordExecutionInput()` 和 `RecordExecutionResultExtended()` 方法（逻辑移入执行层）

```go
// 修改后的 Execute 方法
func (p *ThreeLayerRoutingPipeline) Execute(ctx context.Context, req *RoutingRequest) (*RoutingDecision, error) {
    // 第一层：策略层
    strategyOutput, err := p.executeStrategyLayer(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("strategy layer failed: %w", err)
    }
    
    // 第二层：决策层
    decisionOutput, err := p.executeDecisionLayer(ctx, req, strategyOutput)
    if err != nil {
        return nil, fmt.Errorf("decision layer failed: %w", err)
    }
    
    // 第三层：执行层（新增）
    if req.ExecuteImmediately {
        execOutput, err := p.executeExecutionLayer(ctx, req, decisionOutput)
        if err != nil {
            return nil, fmt.Errorf("execution layer failed: %w", err)
        }
        return execOutput.Decision, nil
    }
    
    return decisionOutput, nil
}
```

**任务清单**:
- [ ] 新增 `executeExecutionLayer()` 方法
- [ ] 修改 `Execute()` 方法签名和逻辑
- [ ] 更新 `RoutingRequest` 结构体，增加 `ExecuteImmediately` 字段
- [ ] 更新 `RoutingDecision` 结构体，增加执行层相关字段
- [ ] 更新单元测试

#### 2.2 更新数据库迁移

**文件**: `backend/migrations/069_add_execution_layer_fields.sql`

```sql
-- 为 routing_decision_logs 表增加执行层字段（如果不存在）
ALTER TABLE routing_decision_logs 
ADD COLUMN IF NOT EXISTS execution_retry_count INT DEFAULT 0;

ALTER TABLE routing_decision_logs 
ADD COLUMN IF NOT EXISTS execution_fallback_used BOOLEAN DEFAULT FALSE;

ALTER TABLE routing_decision_logs 
ADD COLUMN IF NOT EXISTS execution_provider_override VARCHAR(50);
```

**任务清单**:
- [ ] 创建迁移文件
- [ ] 验证迁移执行

---

### Phase 3: api_proxy.go 重构（预计 3 天）

#### 3.1 创建业务编排层

**文件**: `backend/handlers/api_proxy.go`（重构）

**重构目标**:
- 代码量从 1833 行精简至 ~600 行
- 只保留业务编排逻辑
- 调用三层路由 Pipeline 完成路由和转发

**新的代码结构**:

```go
func ProxyAPIRequest(c *gin.Context) {
    // ========== 第一阶段：准入验证 ==========
    
    // 1. 用户认证
    userID, err := authenticateUser(c)
    if err != nil {
        return
    }
    
    // 2. 请求解析
    req, err := parseRequest(c)
    if err != nil {
        return
    }
    
    // 3. 权益校验（合并为一轮）
    entCtx, err := validateEntitlement(userID, req.Provider, req.Model)
    if err != nil {
        respondWithError(c, ErrEntitlementDenied)
        return
    }
    
    // 4. 余额检查（快速检查）
    if !hasMinimumBalance(userID) {
        respondWithError(c, ErrInsufficientBalance)
        return
    }
    
    // ========== 第二阶段：路由决策 ==========
    
    // 5-7. 三层路由Pipeline
    pipeline := services.NewThreeLayerRoutingPipeline()
    routingReq := &services.RoutingRequest{
        UserID:             userID,
        Provider:           req.Provider,
        Model:              req.Model,
        RequestBody:        req.Body,
        EntitlementContext: entCtx,
        ExecuteImmediately: true,  // 立即执行
    }
    
    routingResult, err := pipeline.Execute(c.Request.Context(), routingReq)
    if err != nil {
        respondWithError(c, err)
        return
    }
    
    // ========== 第三阶段：计费与执行 ==========
    
    // 8. 计费预扣（使用准确价格）
    estimatedCost := calculateEstimatedCost(routingResult, req)
    if err := billingEngine.PreDeduct(userID, estimatedCost); err != nil {
        respondWithError(c, ErrInsufficientBalance)
        return
    }
    
    // 9. 流量转发（已在Pipeline中执行）
    // routingResult 已包含执行结果
    
    // 10. 计费结算
    actualCost := calculateActualCost(routingResult)
    billingEngine.Settle(userID, estimatedCost, actualCost)
    
    // 返回响应
    c.Data(routingResult.ExecutionStatusCode, "application/json", routingResult.ExecutionResponseBody)
}
```

**任务清单**:
- [ ] 提取 `authenticateUser()` 函数
- [ ] 提取 `parseRequest()` 函数
- [ ] 合并权益校验逻辑为 `validateEntitlement()` 函数
- [ ] 提取 `hasMinimumBalance()` 函数
- [ ] 重构 `ProxyAPIRequest()` 主函数
- [ ] 移除冗余的路由和转发代码
- [ ] 保留计费相关逻辑
- [ ] 更新单元测试

#### 3.2 提取辅助函数

**文件**: `backend/handlers/api_proxy_helpers.go`（新建）

```go
// 用户认证
func authenticateUser(c *gin.Context) (int, error)

// 请求解析
func parseRequest(c *gin.Context) (*APIProxyRequest, error)

// 权益校验（合并）
func validateEntitlement(db *sql.DB, userID int, provider, model string) (*services.EntitlementRoutingContext, error)

// 余额检查
func hasMinimumBalance(db *sql.DB, userID int) bool

// 计费预估
func calculateEstimatedCost(decision *services.RoutingDecision, req *APIProxyRequest) int

// 计费结算
func calculateActualCost(decision *services.RoutingDecision) int
```

**任务清单**:
- [ ] 创建 `api_proxy_helpers.go` 文件
- [ ] 实现各辅助函数
- [ ] 添加单元测试

---

### Phase 4: 流式响应支持（预计 1 天）

#### 4.1 更新 ExecutionEngine 支持流式

**文件**: `backend/services/execution_engine.go`

```go
type StreamExecutionInput struct {
    ExecutionInput
    StreamWriter io.Writer
}

func (e *ExecutionEngine) ExecuteStream(ctx context.Context, input *StreamExecutionInput) error
```

**任务清单**:
- [ ] 实现 `ExecuteStream()` 方法
- [ ] 处理 SSE 流式响应
- [ ] 更新 `api_proxy_stream.go`

---

### Phase 5: 测试与验证（预计 2 天）

#### 5.1 单元测试

**任务清单**:
- [ ] `execution_engine_test.go` - 执行引擎测试
- [ ] `execution_layer_test.go` - 执行层测试
- [ ] `three_layer_pipeline_test.go` - Pipeline 集成测试
- [ ] `api_proxy_test.go` - 业务编排层测试

#### 5.2 集成测试

**任务清单**:
- [ ] 端到端路由测试
- [ ] 流量转发测试
- [ ] 计费流程测试
- [ ] 错误处理测试

#### 5.3 性能测试

**任务清单**:
- [ ] 压力测试
- [ ] 延迟对比（重构前后）
- [ ] 内存使用对比

---

### Phase 6: 文档与部署（预计 1 天）

#### 6.1 文档更新

**任务清单**:
- [ ] 更新架构文档
- [ ] 更新 API 文档
- [ ] 更新开发指南

#### 6.2 部署验证

**任务清单**:
- [ ] 创建 PR
- [ ] CI/CD 验证
- [ ] 生产环境验证
- [ ] 监控告警验证

---

## 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 重构期间影响线上服务 | 高 | 分阶段发布，保持向后兼容 |
| 流式响应兼容性 | 中 | 充分测试流式场景 |
| 计费逻辑变更 | 高 | 保持计费逻辑不变，只调整顺序 |
| 性能下降 | 中 | 性能测试对比，优化关键路径 |

---

## 回滚方案

如果重构后出现问题，可以通过以下方式回滚：

1. **Feature Flag**: 通过环境变量控制是否使用新架构
2. **代码回滚**: Git revert 到重构前的版本
3. **数据库回滚**: 迁移文件使用 `IF NOT EXISTS`，回滚不影响现有数据

---

## 时间估算

| Phase | 任务 | 预计时间 |
|-------|------|----------|
| Phase 1 | 基础设施建设 | 2 天 |
| Phase 2 | Pipeline 重构 | 2 天 |
| Phase 3 | api_proxy.go 重构 | 3 天 |
| Phase 4 | 流式响应支持 | 1 天 |
| Phase 5 | 测试与验证 | 2 天 |
| Phase 6 | 文档与部署 | 1 天 |
| **总计** | | **11 天** |

---

## 验收标准

1. **功能验收**
   - [ ] 三层路由完整执行（策略层 → 决策层 → 执行层）
   - [ ] 所有现有 API 功能正常
   - [ ] 流式响应正常
   - [ ] 计费逻辑正确

2. **代码质量**
   - [ ] 单元测试覆盖率 > 80%
   - [ ] 无 lint 错误
   - [ ] 代码量精简至目标值

3. **性能验收**
   - [ ] 延迟无明显增加（< 5%）
   - [ ] 内存使用无明显增加（< 10%）

4. **日志验收**
   - [ ] 三层路由日志完整记录
   - [ ] 执行层输入输出日志清晰
   - [ ] 错误日志便于排查

---

## 下一步行动

确认计划后，按以下顺序开始执行：

1. 创建功能分支 `refactor/three-layer-routing-architecture`
2. 开始 Phase 1: 创建 ExecutionEngine
3. 逐步推进各阶段任务
4. 每个 Phase 完成后进行阶段性验证
