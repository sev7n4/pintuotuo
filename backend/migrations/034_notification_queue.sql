-- Notification Queue Migration
-- Version: 034
-- Date: 2026-04-04
-- Description: 创建通知队列表，支持多渠道通知

-- ============================================================
-- Phase 1: 创建通知队列表
-- ============================================================

CREATE TABLE IF NOT EXISTS notification_queue (
  id SERIAL PRIMARY KEY,
  
  user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  
  notification_type VARCHAR(50) NOT NULL,
  title VARCHAR(200) NOT NULL,
  content TEXT NOT NULL,
  
  channel VARCHAR(20) NOT NULL DEFAULT 'in_app',
  priority VARCHAR(20) DEFAULT 'normal',
  
  status VARCHAR(20) DEFAULT 'pending',
  retry_count INT DEFAULT 0,
  max_retries INT DEFAULT 3,
  
  scheduled_at TIMESTAMP,
  sent_at TIMESTAMP,
  
  error_message TEXT,
  
  metadata JSONB,
  
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE notification_queue IS '通知队列表';
COMMENT ON COLUMN notification_queue.notification_type IS '通知类型: audit/order/alert/settlement/system';
COMMENT ON COLUMN notification_queue.channel IS '通知渠道: in_app/email/sms/dingtalk';
COMMENT ON COLUMN notification_queue.priority IS '优先级: low/normal/high/urgent';
COMMENT ON COLUMN notification_queue.status IS '状态: pending/sent/failed/cancelled';

-- ============================================================
-- Phase 2: 创建索引
-- ============================================================

CREATE INDEX IF NOT EXISTS idx_notification_queue_user ON notification_queue(user_id);
CREATE INDEX IF NOT EXISTS idx_notification_queue_status ON notification_queue(status);
CREATE INDEX IF NOT EXISTS idx_notification_queue_channel ON notification_queue(channel);
CREATE INDEX IF NOT EXISTS idx_notification_queue_scheduled ON notification_queue(scheduled_at);
CREATE INDEX IF NOT EXISTS idx_notification_queue_created ON notification_queue(created_at);

-- ============================================================
-- Phase 3: 创建用户通知状态表
-- ============================================================

CREATE TABLE IF NOT EXISTS user_notification_status (
  id SERIAL PRIMARY KEY,
  user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  
  unread_count INT DEFAULT 0,
  last_read_at TIMESTAMP,
  
  email_enabled BOOLEAN DEFAULT TRUE,
  sms_enabled BOOLEAN DEFAULT FALSE,
  dingtalk_enabled BOOLEAN DEFAULT FALSE,
  
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  
  UNIQUE(user_id)
);

COMMENT ON TABLE user_notification_status IS '用户通知状态表';

CREATE INDEX IF NOT EXISTS idx_notification_status_user ON user_notification_status(user_id);

-- ============================================================
-- Phase 4: 创建通知模板表
-- ============================================================

CREATE TABLE IF NOT EXISTS notification_templates (
  id SERIAL PRIMARY KEY,
  
  code VARCHAR(100) UNIQUE NOT NULL,
  name VARCHAR(200) NOT NULL,
  
  notification_type VARCHAR(50) NOT NULL,
  
  title_template VARCHAR(200),
  content_template TEXT NOT NULL,
  
  variables JSONB,
  
  status VARCHAR(20) DEFAULT 'active',
  
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE notification_templates IS '通知模板表';
COMMENT ON COLUMN notification_templates.variables IS '模板变量列表，如 ["user_name", "order_id"]';

CREATE INDEX IF NOT EXISTS idx_notification_templates_code ON notification_templates(code);
CREATE INDEX IF NOT EXISTS idx_notification_templates_type ON notification_templates(notification_type);

-- ============================================================
-- Phase 5: 插入默认通知模板
-- ============================================================

INSERT INTO notification_templates (code, name, notification_type, title_template, content_template, variables) VALUES
  ('audit_approved', '审核通过通知', 'audit', '您的申请已通过审核', '尊敬的{user_name}，您的{entity_type}申请已通过审核。', '["user_name", "entity_type"]'),
  ('audit_rejected', '审核拒绝通知', 'audit', '您的申请被拒绝', '尊敬的{user_name}，您的{entity_type}申请被拒绝。原因：{reason}', '["user_name", "entity_type", "reason"]'),
  ('order_completed', '订单完成通知', 'order', '订单已完成', '尊敬的{user_name}，您的订单{order_id}已完成。', '["user_name", "order_id"]'),
  ('settlement_created', '结算单生成通知', 'settlement', '新的结算单已生成', '尊敬的{user_name}，您有新的结算单待确认，金额：{amount}元。', '["user_name", "amount"]'),
  ('provider_health_alert', 'Provider健康告警', 'alert', 'Provider健康状态异常', 'Provider {provider_name} 健康状态变更为{status}，请及时处理。', '["provider_name", "status"]')
ON CONFLICT (code) DO NOTHING;

-- ============================================================
-- Phase 6: 创建更新时间触发器
-- ============================================================

CREATE OR REPLACE FUNCTION update_user_notification_status_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_user_notification_status_updated_at
  BEFORE UPDATE ON user_notification_status
  FOR EACH ROW
  EXECUTE FUNCTION update_user_notification_status_updated_at();

CREATE OR REPLACE FUNCTION update_notification_templates_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_notification_templates_updated_at
  BEFORE UPDATE ON notification_templates
  FOR EACH ROW
  EXECUTE FUNCTION update_notification_templates_updated_at();
