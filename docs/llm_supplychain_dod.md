# ExecutionLayer 统一出站验收口径

> 更新时间: 2026-04-27
> 版本: v2.0

## 1. 目标

完成 ExecutionLayer 统一出站入口的灰度发布验收，确保三层路由架构（策略层 → 决策层 → 执行层）稳定运行。

## 2. 验收对象

- ExecutionLayer：统一出站入口
- UnifiedRouter：配置驱动路由决策
- RouteCache：路由决策缓存
- 路由模式：direct / litellm / proxy

## 3. 评分维度与权重

| 维度 | 权重 | 说明 |
| --- | --- | --- |
| 兼容性 | 25% | OpenAI API 兼容程度、错误语义一致性 |
| 稳定性 | 30% | 成功率、限流恢复能力、故障隔离能力 |
| 性能 | 25% | P95/P99 延迟、缓存命中率 |
| 运维复杂度 | 20% | 配置复杂度、排障路径、灰度开关 |

总分 = 各维度得分 * 权重，满分 100。

## 4. 统一验收指标（DoD）

### 4.1 联通性 DoD

- [ ] ExecutionLayer 能完成 `/v1/chat/completions` 代表模型调用
- [ ] 三种路由模式 (direct/litellm/proxy) 均可正常工作
- [ ] 降级链路可正常触发：配置驱动 → 环境变量驱动 → Direct

### 4.2 稳定性 DoD

- [ ] 注入 429/5xx/timeout 后，系统能在 5 分钟内恢复到基线成功率
- [ ] 熔断触发后可自动半开探测并恢复
- [ ] 请求 trace 可看到路由决策、执行结果、重试次数

### 4.3 可观测性 DoD

- [ ] Prometheus 可采集路由决策计数、执行延迟、错误类型
- [ ] Grafana 具备 ExecutionLayer 监控看板
- [ ] Alertmanager 告警链路可用

### 4.4 质量 DoD

- [ ] E2E 测试覆盖配置驱动路由场景
- [ ] E2E 测试覆盖降级机制验证
- [ ] 性能基准测试通过

## 5. 发布门槛

连续 7 天满足：
- 成功率 >= 99.0%
- P95 延迟 <= 3s
- E2E 测试通过率 100%

存在可执行回滚路径：
```bash
USE_EXECUTION_LAYER=false
docker-compose -f docker-compose.prod.yml restart backend
```

## 6. 建议执行节奏

- 第 1 周：ExecutionLayer 联通 + 稳定性底座验证
- 第 2 周：监控告警 + 灰度切换 + 全量发布
