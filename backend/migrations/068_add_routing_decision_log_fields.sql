-- 添加路由决策日志表的完整三层路由字段

-- 添加策略层原因字段
ALTER TABLE routing_decision_logs ADD COLUMN IF NOT EXISTS strategy_layer_reason TEXT;

-- 添加路由模式字段
ALTER TABLE routing_decision_logs ADD COLUMN IF NOT EXISTS routing_mode VARCHAR(50);

-- 添加执行层详细字段
ALTER TABLE routing_decision_logs ADD COLUMN IF NOT EXISTS execution_success BOOLEAN DEFAULT false;
ALTER TABLE routing_decision_logs ADD COLUMN IF NOT EXISTS execution_status_code INTEGER;
ALTER TABLE routing_decision_logs ADD COLUMN IF NOT EXISTS execution_latency_ms INTEGER;
ALTER TABLE routing_decision_logs ADD COLUMN IF NOT EXISTS execution_error_message TEXT;

-- 添加索引以支持按路由模式查询
CREATE INDEX IF NOT EXISTS idx_routing_decision_logs_routing_mode ON routing_decision_logs(routing_mode);
CREATE INDEX IF NOT EXISTS idx_routing_decision_logs_execution_success ON routing_decision_logs(execution_success);

-- 添加注释
COMMENT ON COLUMN routing_decision_logs.strategy_layer_reason IS '策略选择原因，如"Stream request requires low latency"';
COMMENT ON COLUMN routing_decision_logs.routing_mode IS '路由模式：direct, litellm, proxy';
COMMENT ON COLUMN routing_decision_logs.execution_success IS '执行是否成功';
COMMENT ON COLUMN routing_decision_logs.execution_status_code IS 'HTTP状态码';
COMMENT ON COLUMN routing_decision_logs.execution_latency_ms IS '执行耗时（毫秒）';
COMMENT ON COLUMN routing_decision_logs.execution_error_message IS '执行错误信息';
