-- +goose Up
-- +goose StatementBegin

-- 添加 security_weight 和 load_balance_weight 字段到 routing_strategies 表
ALTER TABLE routing_strategies ADD COLUMN IF NOT EXISTS security_weight NUMERIC(5, 4) DEFAULT 0.1;
ALTER TABLE routing_strategies ADD COLUMN IF NOT EXISTS load_balance_weight NUMERIC(5, 4) DEFAULT 0.1;

-- 更新现有策略的权重值，确保总和为 1.0
UPDATE routing_strategies SET 
    security_weight = 0.1,
    load_balance_weight = 0.1
WHERE code = 'balanced' AND security_weight IS NULL;

UPDATE routing_strategies SET 
    security_weight = 0.05,
    load_balance_weight = 0.1
WHERE code = 'cost_first' AND security_weight IS NULL;

UPDATE routing_strategies SET 
    security_weight = 0.1,
    load_balance_weight = 0.1
WHERE code = 'performance_first' AND security_weight IS NULL;

UPDATE routing_strategies SET 
    security_weight = 0.5,
    load_balance_weight = 0.1
WHERE code = 'security_first' AND security_weight IS NULL;

UPDATE routing_strategies SET 
    security_weight = 0.1,
    load_balance_weight = 0.1
WHERE code = 'reliability_first' AND security_weight IS NULL;

UPDATE routing_strategies SET 
    security_weight = 0.1,
    load_balance_weight = 0.1
WHERE code = 'auto' AND security_weight IS NULL;

-- 添加执行层输入字段到 routing_decision_logs 表
ALTER TABLE routing_decision_logs ADD COLUMN IF NOT EXISTS execution_layer_input JSONB;

-- 添加索引
CREATE INDEX IF NOT EXISTS idx_routing_strategies_security_weight ON routing_strategies(security_weight);
CREATE INDEX IF NOT EXISTS idx_routing_strategies_load_balance_weight ON routing_strategies(load_balance_weight);

-- 添加注释
COMMENT ON COLUMN routing_strategies.security_weight IS '安全权重 (0-1)';
COMMENT ON COLUMN routing_strategies.load_balance_weight IS '负载均衡权重 (0-1)';
COMMENT ON COLUMN routing_decision_logs.execution_layer_input IS '执行层输入信息：gateway_mode, endpoint_url, auth_method, resolved_model, request_format';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- 删除执行层输入字段
ALTER TABLE routing_decision_logs DROP COLUMN IF EXISTS execution_layer_input;

-- 删除 security_weight 和 load_balance_weight 字段
ALTER TABLE routing_strategies DROP COLUMN IF EXISTS security_weight;
ALTER TABLE routing_strategies DROP COLUMN IF EXISTS load_balance_weight;

-- 删除索引
DROP INDEX IF EXISTS idx_routing_strategies_security_weight;
DROP INDEX IF EXISTS idx_routing_strategies_load_balance_weight;

-- +goose StatementEnd
