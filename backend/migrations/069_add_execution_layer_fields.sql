-- 069_add_execution_layer_fields.sql
-- 为 routing_decision_logs 表增加执行层字段

ALTER TABLE routing_decision_logs 
ADD COLUMN IF NOT EXISTS execution_layer_input JSONB;

ALTER TABLE routing_decision_logs 
ADD COLUMN IF NOT EXISTS execution_layer_result JSONB;

ALTER TABLE routing_decision_logs 
ADD COLUMN IF NOT EXISTS execution_retry_count INT DEFAULT 0;

ALTER TABLE routing_decision_logs 
ADD COLUMN IF NOT EXISTS execution_fallback_used BOOLEAN DEFAULT FALSE;

ALTER TABLE routing_decision_logs 
ADD COLUMN IF NOT EXISTS execution_provider_override VARCHAR(50);

-- 添加索引
CREATE INDEX IF NOT EXISTS idx_routing_decision_logs_execution_success 
ON routing_decision_logs(execution_success);

CREATE INDEX IF NOT EXISTS idx_routing_decision_logs_execution_status_code 
ON routing_decision_logs(execution_status_code);
