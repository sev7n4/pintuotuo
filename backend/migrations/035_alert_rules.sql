-- Alert Rules Migration
-- Version: 035
-- Date: 2026-04-04
-- Description: 创建告警规则表和告警历史表

-- ============================================================
-- Phase 1: 创建告警规则表
-- ============================================================

CREATE TABLE IF NOT EXISTS alert_rules (
  id SERIAL PRIMARY KEY,
  
  name VARCHAR(100) NOT NULL,
  code VARCHAR(100) UNIQUE NOT NULL,
  description TEXT,
  
  metric_type VARCHAR(50) NOT NULL,
  entity_type VARCHAR(50),
  
  condition_type VARCHAR(20) NOT NULL,
  threshold DECIMAL(10, 2) NOT NULL,
  duration_seconds INT DEFAULT 60,
  
  notification_channels JSONB DEFAULT '["in_app", "email"]',
  notification_template_id INT REFERENCES notification_templates(id),
  
  severity VARCHAR(20) DEFAULT 'warning',
  
  status VARCHAR(20) DEFAULT 'active',
  
  created_by INT REFERENCES users(id),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE alert_rules IS '告警规则表';
COMMENT ON COLUMN alert_rules.metric_type IS '指标类型: health_status/latency/error_rate/usage/settlement';
COMMENT ON COLUMN alert_rules.entity_type IS '实体类型: provider/api_key/merchant/system';
COMMENT ON COLUMN alert_rules.condition_type IS '条件类型: greater_than/less_than/equals/not_equals';
COMMENT ON COLUMN alert_rules.severity IS '严重程度: info/warning/critical/emergency';

CREATE INDEX IF NOT EXISTS idx_alert_rules_code ON alert_rules(code);
CREATE INDEX IF NOT EXISTS idx_alert_rules_metric ON alert_rules(metric_type);
CREATE INDEX IF NOT EXISTS idx_alert_rules_status ON alert_rules(status);

-- ============================================================
-- Phase 2: 创建告警历史表
-- ============================================================

CREATE TABLE IF NOT EXISTS alert_history (
  id SERIAL PRIMARY KEY,
  
  rule_id INT NOT NULL REFERENCES alert_rules(id) ON DELETE CASCADE,
  
  entity_type VARCHAR(50),
  entity_id INT,
  entity_name VARCHAR(200),
  
  metric_value DECIMAL(10, 2),
  threshold DECIMAL(10, 2),
  
  status VARCHAR(20) DEFAULT 'firing',
  severity VARCHAR(20),
  
  fired_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  resolved_at TIMESTAMP,
  
  notification_sent BOOLEAN DEFAULT FALSE,
  notification_sent_at TIMESTAMP,
  
  acknowledged_by INT REFERENCES users(id),
  acknowledged_at TIMESTAMP,
  
  notes TEXT
);

COMMENT ON TABLE alert_history IS '告警历史表';
COMMENT ON COLUMN alert_history.status IS '状态: firing/resolved/acknowledged';

CREATE INDEX IF NOT EXISTS idx_alert_history_rule ON alert_history(rule_id);
CREATE INDEX IF NOT EXISTS idx_alert_history_entity ON alert_history(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_alert_history_status ON alert_history(status);
CREATE INDEX IF NOT EXISTS idx_alert_history_fired ON alert_history(fired_at);

-- ============================================================
-- Phase 3: 插入默认告警规则
-- ============================================================

INSERT INTO alert_rules (name, code, description, metric_type, entity_type, condition_type, threshold, duration_seconds, severity) VALUES
  ('Provider健康状态异常', 'provider_health_unhealthy', 'Provider健康状态变为unhealthy', 'health_status', 'provider', 'equals', 0, 0, 'critical'),
  ('Provider健康状态降级', 'provider_health_degraded', 'Provider健康状态变为degraded', 'health_status', 'provider', 'equals', 1, 0, 'warning'),
  ('API响应延迟过高', 'api_latency_high', 'API响应延迟超过阈值', 'latency', 'api_key', 'greater_than', 5000, 60, 'warning'),
  ('API错误率过高', 'api_error_rate_high', 'API错误率超过阈值', 'error_rate', 'api_key', 'greater_than', 10, 300, 'critical'),
  ('商户结算金额异常', 'settlement_amount_anomaly', '商户结算金额与预期差异过大', 'settlement', 'merchant', 'greater_than', 20, 0, 'warning')
ON CONFLICT (code) DO NOTHING;

-- ============================================================
-- Phase 4: 创建告警统计视图
-- ============================================================

CREATE OR REPLACE VIEW alert_statistics AS
SELECT 
  ar.id as rule_id,
  ar.name as rule_name,
  ar.metric_type,
  ar.severity,
  COUNT(ah.id) as total_alerts,
  COUNT(CASE WHEN ah.status = 'firing' THEN 1 END) as firing_count,
  COUNT(CASE WHEN ah.status = 'resolved' THEN 1 END) as resolved_count,
  MAX(ah.fired_at) as last_fired_at
FROM alert_rules ar
LEFT JOIN alert_history ah ON ar.id = ah.rule_id
GROUP BY ar.id, ar.name, ar.metric_type, ar.severity;

COMMENT ON VIEW alert_statistics IS '告警统计视图';

-- ============================================================
-- Phase 5: 创建更新时间触发器
-- ============================================================

CREATE OR REPLACE FUNCTION update_alert_rules_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_alert_rules_updated_at
  BEFORE UPDATE ON alert_rules
  FOR EACH ROW
  EXECUTE FUNCTION update_alert_rules_updated_at();
