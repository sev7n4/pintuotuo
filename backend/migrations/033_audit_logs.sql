-- Audit Logs Migration
-- Version: 033
-- Date: 2026-04-04
-- Description: 创建审核日志表，记录所有关键操作

-- ============================================================
-- Phase 1: 创建审核日志表
-- ============================================================

CREATE TABLE IF NOT EXISTS audit_logs (
  id SERIAL PRIMARY KEY,
  
  entity_type VARCHAR(50) NOT NULL,
  entity_id INT NOT NULL,
  
  action VARCHAR(50) NOT NULL,
  
  old_value JSONB,
  new_value JSONB,
  
  operator_id INT REFERENCES users(id) ON DELETE SET NULL,
  operator_type VARCHAR(20) DEFAULT 'user',
  
  ip_address VARCHAR(50),
  user_agent TEXT,
  
  metadata JSONB,
  
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE audit_logs IS '审核日志表，记录所有关键操作';
COMMENT ON COLUMN audit_logs.entity_type IS '实体类型: spu/sku/merchant/settlement/pricing';
COMMENT ON COLUMN audit_logs.entity_id IS '实体ID';
COMMENT ON COLUMN audit_logs.action IS '操作类型: create/update/delete/approve/reject';
COMMENT ON COLUMN audit_logs.operator_type IS '操作人类型: user/admin/system';

-- ============================================================
-- Phase 2: 创建索引
-- ============================================================

CREATE INDEX IF NOT EXISTS idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_operator ON audit_logs(operator_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created ON audit_logs(created_at);

-- ============================================================
-- Phase 3: 创建审核日志视图
-- ============================================================

CREATE OR REPLACE VIEW audit_log_details AS
SELECT 
  al.id,
  al.entity_type,
  al.entity_id,
  al.action,
  al.old_value,
  al.new_value,
  al.operator_id,
  al.operator_type,
  u.name as operator_name,
  u.email as operator_email,
  al.ip_address,
  al.user_agent,
  al.metadata,
  al.created_at
FROM audit_logs al
LEFT JOIN users u ON al.operator_id = u.id;

COMMENT ON VIEW audit_log_details IS '审核日志详情视图，包含操作人信息';

-- ============================================================
-- Phase 4: 创建分区表（按月分区，保留12个月）
-- ============================================================

-- 注意：PostgreSQL 分区表需要单独创建，这里仅作为参考
-- 实际生产环境建议使用 pg_partman 扩展自动管理分区

-- ============================================================
-- Phase 5: 创建清理函数（保留12个月数据）
-- ============================================================

CREATE OR REPLACE FUNCTION cleanup_old_audit_logs()
RETURNS void AS $$
BEGIN
  DELETE FROM audit_logs 
  WHERE created_at < CURRENT_TIMESTAMP - INTERVAL '12 months';
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION cleanup_old_audit_logs IS '清理12个月前的审核日志';
