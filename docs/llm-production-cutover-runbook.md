# ExecutionLayer 统一出站入口灰度发布 Runbook

> 更新时间: 2026-04-27
> 版本: v2.0

## 概述

ExecutionLayer 是三层路由架构的执行层，作为统一出站入口，负责：
- 统一 HTTP 请求执行
- 重试与熔断机制
- 路由模式切换 (direct/litellm/proxy)
- 降级链路管理

## 0. 灰度开关配置

### 0.1 环境变量

| 环境变量 | 默认值 | 说明 |
|---------|--------|------|
| `USE_EXECUTION_LAYER` | `false` | 启用 ExecutionLayer 统一出站入口 |
| `USE_CONFIG_DRIVEN_ROUTING` | `false` | 启用配置驱动路由（从数据库读取路由策略） |

### 0.2 配置方式

```bash
# .env 文件

# 灰度开启 ExecutionLayer 统一出站
USE_EXECUTION_LAYER=true

# 灰度开启配置驱动路由（可选）
USE_CONFIG_DRIVEN_ROUTING=true
```

### 0.3 路由模式说明

| 模式 | 说明 | 适用场景 |
|------|------|---------|
| `direct` | 直连 Provider API | 海外用户、低延迟需求 |
| `litellm` | 通过 LiteLLM 网关 | 国内用户访问海外 Provider |
| `proxy` | 通过代理服务器 | 网关故障降级 |

### 0.4 降级链路

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           降级链路                                           │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  1. 配置驱动路由失败 → 降级到环境变量驱动                                      │
│     ┌─────────────────────────────────────────────────────────────────┐     │
│     │ resolveRouteDecision() 失败 → 使用 LLM_GATEWAY_ACTIVE           │     │
│     └─────────────────────────────────────────────────────────────────┘     │
│                                                                              │
│  2. LiteLLM 故障 → 降级到 Proxy                                              │
│     ┌─────────────────────────────────────────────────────────────────┐     │
│     │ RouteDecision.FallbackMode = "proxy"                            │     │
│     └─────────────────────────────────────────────────────────────────┘     │
│                                                                              │
│  3. 环境变量缺失 → 降级到 Direct                                              │
│     ┌─────────────────────────────────────────────────────────────────┐     │
│     │ LLM_GATEWAY_ACTIVE 未设置 → 直连 Provider                       │     │
│     └─────────────────────────────────────────────────────────────────┘     │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

## 1. 发布前检查

- [ ] 单元测试通过：`go test ./services/... -v`
- [ ] E2E 测试通过：`go test ./services/... -run TestE2E -v`
- [ ] 编译检查通过：`go build ./...`
- [ ] 告警链路可触发且通知可达
- [ ] Promptfoo 最近一次回归通过率 >= 80%

## 2. 灰度步骤

### 2.1 阶段一：开启 ExecutionLayer (5% 流量)

```bash
# .env 配置
USE_EXECUTION_LAYER=true
```

观察指标 (24h)：
- 成功率变化
- 5xx 错误率
- P95 延迟
- Token 成本

### 2.2 阶段二：开启配置驱动路由 (可选)

```bash
# .env 配置
USE_EXECUTION_LAYER=true
USE_CONFIG_DRIVEN_ROUTING=true
```

观察指标 (24h)：
- 路由决策正确性
- 降级触发次数
- 各路由模式分布

### 2.3 阶段三：全量切换

确认以下指标正常后全量切换：
- 5xx 错误率 < 1%
- P95 延迟 < 2s
- 无关键业务接口错误

## 3. 回滚策略

满足任一条件立即回滚：
- 5xx 错误率 > 5% 持续 10 分钟
- P95 延迟 > 3 秒持续 10 分钟
- 关键链路可用性 < 99%

**回滚动作**：

```bash
# 立即关闭 ExecutionLayer
USE_EXECUTION_LAYER=false

# 重启服务
docker-compose -f docker-compose.prod.yml restart backend
```

## 4. 监控指标

### 4.1 Prometheus 指标

| 指标名称 | 说明 |
|---------|------|
| `route_decision_counter` | 路由决策计数 |
| `execution_layer_latency` | ExecutionLayer 延迟 |
| `execution_layer_requests_total` | ExecutionLayer 请求总数 |

### 4.2 关键日志

```
# 路由决策日志
logger.LogInfo(c.Request.Context(), "api_proxy", "Route decision", map[string]interface{}{
    "mode": decision.Mode,
    "endpoint": decision.Endpoint,
    "fallback_mode": decision.FallbackMode,
})

# ExecutionLayer 执行日志
logger.LogInfo(c.Request.Context(), "execution_layer", "Request executed", map[string]interface{}{
    "provider": result.Provider,
    "status_code": result.StatusCode,
    "latency_ms": result.LatencyMs,
})
```

## 5. 复盘输出

- 触发时间线
- 根因分类（上游限流/网络/配置/代码）
- 修复项与防复发动作
- 是否恢复灰度及下一次窗口

## 6. 架构参考

详细架构说明见：[docs/llm-request-flow.md](llm-request-flow.md)

关键组件：

| 组件 | 文件 | 职责 |
|------|------|------|
| ExecutionLayer | `services/execution_layer.go` | 统一出站入口 |
| ExecutionEngine | `services/execution_engine.go` | HTTP 请求执行 |
| UnifiedRouter | `services/unified_router.go` | 配置驱动路由决策 |
| RouteCache | `services/route_cache.go` | 路由决策缓存 |
