-- 添加路由决策日志的成本和商户字段
ALTER TABLE routing_decision_logs ADD COLUMN IF NOT EXISTS selected_merchant_id INTEGER;
ALTER TABLE routing_decision_logs ADD COLUMN IF NOT EXISTS input_token_cost NUMERIC(20,10);
ALTER TABLE routing_decision_logs ADD COLUMN IF NOT EXISTS output_token_cost NUMERIC(20,10);

-- 添加外键约束
ALTER TABLE routing_decision_logs 
ADD CONSTRAINT fk_selected_merchant_id 
FOREIGN KEY (selected_merchant_id) REFERENCES merchants(id) ON DELETE SET NULL;

-- 添加索引
CREATE INDEX IF NOT EXISTS idx_routing_decision_logs_selected_merchant_id ON routing_decision_logs(selected_merchant_id);
