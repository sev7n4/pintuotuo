-- +goose Up
-- +goose StatementBegin

-- 补充缺失的路由策略
INSERT INTO routing_strategies (name, code, description, price_weight, latency_weight, reliability_weight, max_retry_count, retry_backoff_base, circuit_breaker_threshold, circuit_breaker_timeout, is_default, status) VALUES
  ('成本优先', 'cost_first', '优先选择成本最低的Provider，适合对成本敏感的场景', 0.7, 0.1, 0.2, 3, 1000, 5, 60, FALSE, 'active'),
  ('性能优先', 'performance_first', '优先选择响应最快的Provider，适合对延迟敏感的场景', 0.1, 0.5, 0.2, 3, 1000, 5, 60, FALSE, 'active'),
  ('安全优先', 'security_first', '优先选择安全等级最高的Provider，适合对数据安全要求高的场景', 0.1, 0.1, 0.2, 3, 1000, 5, 60, FALSE, 'active'),
  ('自动策略', 'auto', '根据请求特征自动选择最优策略，适合通用场景', 0.2, 0.3, 0.3, 3, 1000, 5, 60, FALSE, 'active')
ON CONFLICT (code) DO UPDATE SET
  name = EXCLUDED.name,
  description = EXCLUDED.description,
  price_weight = EXCLUDED.price_weight,
  latency_weight = EXCLUDED.latency_weight,
  reliability_weight = EXCLUDED.reliability_weight,
  status = EXCLUDED.status;

-- 更新 reliability_first 策略的权重
UPDATE routing_strategies 
SET price_weight = 0.2, latency_weight = 0.2, reliability_weight = 0.6
WHERE code = 'reliability_first';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- 删除新增的策略
DELETE FROM routing_strategies WHERE code IN ('cost_first', 'performance_first', 'security_first', 'auto');

-- 恢复 reliability_first 策略的权重
UPDATE routing_strategies 
SET price_weight = 0.4, latency_weight = 0.3, reliability_weight = 0.3
WHERE code = 'reliability_first';

-- +goose StatementEnd
