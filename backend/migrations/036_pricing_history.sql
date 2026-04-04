-- Pricing History Migration
-- Version: 036
-- Date: 2026-04-04
-- Description: 创建定价历史表和定价变更计划表

-- ============================================================
-- Phase 1: 创建定价历史表
-- ============================================================

CREATE TABLE IF NOT EXISTS pricing_history (
  id SERIAL PRIMARY KEY,
  
  entity_type VARCHAR(20) NOT NULL,
  entity_id INT NOT NULL,
  
  old_input_price DECIMAL(10, 6),
  old_output_price DECIMAL(10, 6),
  new_input_price DECIMAL(10, 6),
  new_output_price DECIMAL(10, 6),
  
  old_retail_price DECIMAL(10, 2),
  new_retail_price DECIMAL(10, 2),
  
  change_reason VARCHAR(255),
  
  changed_by INT REFERENCES users(id),
  changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  
  effective_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  
  ip_address VARCHAR(50),
  
  metadata JSONB
);

COMMENT ON TABLE pricing_history IS '定价历史表';
COMMENT ON COLUMN pricing_history.entity_type IS '实体类型: spu/sku/merchant_api_key/merchant_sku';
COMMENT ON COLUMN pricing_history.change_reason IS '变更原因';
COMMENT ON COLUMN pricing_history.effective_at IS '生效时间';

CREATE INDEX IF NOT EXISTS idx_pricing_history_entity ON pricing_history(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_pricing_history_changed ON pricing_history(changed_at);
CREATE INDEX IF NOT EXISTS idx_pricing_history_effective ON pricing_history(effective_at);

-- ============================================================
-- Phase 2: 创建定价变更计划表
-- ============================================================

CREATE TABLE IF NOT EXISTS pricing_schedules (
  id SERIAL PRIMARY KEY,
  
  entity_type VARCHAR(20) NOT NULL,
  entity_id INT NOT NULL,
  
  new_input_price DECIMAL(10, 6),
  new_output_price DECIMAL(10, 6),
  new_retail_price DECIMAL(10, 2),
  
  scheduled_at TIMESTAMP NOT NULL,
  status VARCHAR(20) DEFAULT 'pending',
  
  change_reason VARCHAR(255),
  
  created_by INT REFERENCES users(id),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  
  executed_at TIMESTAMP,
  executed_by INT REFERENCES users(id),
  
  notification_sent BOOLEAN DEFAULT FALSE
);

COMMENT ON TABLE pricing_schedules IS '定价变更计划表';
COMMENT ON COLUMN pricing_schedules.status IS '状态: pending/executed/cancelled';

CREATE INDEX IF NOT EXISTS idx_pricing_schedules_entity ON pricing_schedules(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_pricing_schedules_scheduled ON pricing_schedules(scheduled_at);
CREATE INDEX IF NOT EXISTS idx_pricing_schedules_status ON pricing_schedules(status);

-- ============================================================
-- Phase 3: 创建定价对比视图
-- ============================================================

CREATE OR REPLACE VIEW pricing_comparison AS
SELECT 
  ph.id,
  ph.entity_type,
  ph.entity_id,
  
  CASE ph.entity_type
    WHEN 'spu' THEN s.name
    WHEN 'sku' THEN sk.sku_code
    WHEN 'merchant_api_key' THEN mak.name
    ELSE 'Unknown'
  END as entity_name,
  
  ph.old_input_price,
  ph.old_output_price,
  ph.new_input_price,
  ph.new_output_price,
  
  ph.new_input_price - ph.old_input_price as input_price_diff,
  ph.new_output_price - ph.old_output_price as output_price_diff,
  
  CASE 
    WHEN ph.old_input_price > 0 THEN 
      ROUND(((ph.new_input_price - ph.old_input_price) / ph.old_input_price) * 100, 2)
    ELSE NULL
  END as input_price_change_pct,
  
  ph.change_reason,
  u.name as changed_by_name,
  ph.changed_at,
  ph.effective_at
  
FROM pricing_history ph
LEFT JOIN users u ON ph.changed_by = u.id
LEFT JOIN spus s ON ph.entity_type = 'spu' AND ph.entity_id = s.id
LEFT JOIN skus sk ON ph.entity_type = 'sku' AND ph.entity_id = sk.id
LEFT JOIN merchant_api_keys mak ON ph.entity_type = 'merchant_api_key' AND ph.entity_id = mak.id;

COMMENT ON VIEW pricing_comparison IS '定价对比视图';

-- ============================================================
-- Phase 4: 创建定价审计函数
-- ============================================================

CREATE OR REPLACE FUNCTION log_pricing_change(
  p_entity_type VARCHAR,
  p_entity_id INT,
  p_old_input_price DECIMAL,
  p_old_output_price DECIMAL,
  p_new_input_price DECIMAL,
  p_new_output_price DECIMAL,
  p_change_reason VARCHAR,
  p_changed_by INT,
  p_ip_address VARCHAR DEFAULT NULL
) RETURNS void AS $$
BEGIN
  INSERT INTO pricing_history (
    entity_type, entity_id,
    old_input_price, old_output_price,
    new_input_price, new_output_price,
    change_reason, changed_by, ip_address
  ) VALUES (
    p_entity_type, p_entity_id,
    p_old_input_price, p_old_output_price,
    p_new_input_price, p_new_output_price,
    p_change_reason, p_changed_by, p_ip_address
  );
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION log_pricing_change IS '记录定价变更审计日志';
