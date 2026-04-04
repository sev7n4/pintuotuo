-- Smart Routing Migration
-- Version: 027
-- Date: 2026-04-04
-- Description: 创建智能路由相关表，支持健康检查和路由策略

-- ============================================================
-- Phase 1: 添加健康检查配置字段到 merchant_api_keys
-- ============================================================

ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS health_check_interval INT DEFAULT 300;
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS health_check_level VARCHAR(20) DEFAULT 'medium';
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS endpoint_url VARCHAR(500);
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS health_status VARCHAR(20) DEFAULT 'unknown';
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS last_health_check_at TIMESTAMP;
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS consecutive_failures INT DEFAULT 0;

COMMENT ON COLUMN merchant_api_keys.health_check_interval IS '健康检查间隔（秒），默认300秒';
COMMENT ON COLUMN merchant_api_keys.health_check_level IS '健康检查级别: high(1min)/medium(5min)/low(30min)/daily(24h)';
COMMENT ON COLUMN merchant_api_keys.endpoint_url IS '自定义API端点URL';
COMMENT ON COLUMN merchant_api_keys.health_status IS '健康状态: healthy/degraded/unhealthy/unknown';
COMMENT ON COLUMN merchant_api_keys.last_health_check_at IS '最后健康检查时间';
COMMENT ON COLUMN merchant_api_keys.consecutive_failures IS '连续失败次数，用于熔断器';

-- ============================================================
-- Phase 2: 创建健康检查历史表
-- ============================================================

CREATE TABLE IF NOT EXISTS api_key_health_history (
  id SERIAL PRIMARY KEY,
  api_key_id INT NOT NULL REFERENCES merchant_api_keys(id) ON DELETE CASCADE,
  
  check_type VARCHAR(20) NOT NULL,
  status VARCHAR(20) NOT NULL,
  latency_ms INT,
  error_message TEXT,
  
  models_available JSONB,
  pricing_info JSONB,
  
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE api_key_health_history IS 'API Key健康检查历史记录';
COMMENT ON COLUMN api_key_health_history.check_type IS '检查类型: lightweight/full/passive';
COMMENT ON COLUMN api_key_health_history.status IS '检查状态: healthy/degraded/unhealthy';
COMMENT ON COLUMN api_key_health_history.latency_ms IS '响应延迟（毫秒）';
COMMENT ON COLUMN api_key_health_history.error_message IS '错误信息';
COMMENT ON COLUMN api_key_health_history.models_available IS '可用模型列表';
COMMENT ON COLUMN api_key_health_history.pricing_info IS '定价信息';

CREATE INDEX IF NOT EXISTS idx_health_history_api_key ON api_key_health_history(api_key_id);
CREATE INDEX IF NOT EXISTS idx_health_history_status ON api_key_health_history(status);
CREATE INDEX IF NOT EXISTS idx_health_history_created ON api_key_health_history(created_at);

-- ============================================================
-- Phase 3: 创建路由策略配置表
-- ============================================================

CREATE TABLE IF NOT EXISTS routing_strategies (
  id SERIAL PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  code VARCHAR(50) UNIQUE NOT NULL,
  description TEXT,
  
  price_weight DECIMAL(5, 2) DEFAULT 0.4,
  latency_weight DECIMAL(5, 2) DEFAULT 0.3,
  reliability_weight DECIMAL(5, 2) DEFAULT 0.3,
  
  max_retry_count INT DEFAULT 3,
  retry_backoff_base INT DEFAULT 1000,
  
  circuit_breaker_threshold INT DEFAULT 5,
  circuit_breaker_timeout INT DEFAULT 60,
  
  is_default BOOLEAN DEFAULT FALSE,
  status VARCHAR(20) DEFAULT 'active',
  
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE routing_strategies IS '路由策略配置表';
COMMENT ON COLUMN routing_strategies.price_weight IS '价格权重（0-1）';
COMMENT ON COLUMN routing_strategies.latency_weight IS '延迟权重（0-1）';
COMMENT ON COLUMN routing_strategies.reliability_weight IS '可靠性权重（0-1）';
COMMENT ON COLUMN routing_strategies.max_retry_count IS '最大重试次数';
COMMENT ON COLUMN routing_strategies.retry_backoff_base IS '重试退避基数（毫秒）';
COMMENT ON COLUMN routing_strategies.circuit_breaker_threshold IS '熔断器阈值（连续失败次数）';
COMMENT ON COLUMN routing_strategies.circuit_breaker_timeout IS '熔断器超时时间（秒）';

CREATE INDEX IF NOT EXISTS idx_routing_strategies_code ON routing_strategies(code);
CREATE INDEX IF NOT EXISTS idx_routing_strategies_status ON routing_strategies(status);

-- ============================================================
-- Phase 4: 插入默认路由策略
-- ============================================================

INSERT INTO routing_strategies (name, code, description, price_weight, latency_weight, reliability_weight, is_default) VALUES
  ('价格优先', 'price_first', '优先选择价格最低的Provider', 0.6, 0.2, 0.2, FALSE),
  ('延迟优先', 'latency_first', '优先选择延迟最低的Provider', 0.2, 0.6, 0.2, FALSE),
  ('均衡策略', 'balanced', '均衡考虑价格、延迟和可靠性', 0.33, 0.34, 0.33, TRUE),
  ('可靠性优先', 'reliability_first', '优先选择最可靠的Provider', 0.2, 0.2, 0.6, FALSE)
ON CONFLICT (code) DO NOTHING;

-- ============================================================
-- Phase 5: 创建路由决策日志表
-- ============================================================

CREATE TABLE IF NOT EXISTS routing_decisions (
  id SERIAL PRIMARY KEY,
  request_id VARCHAR(100),
  user_id INT REFERENCES users(id) ON DELETE SET NULL,
  
  model_requested VARCHAR(100) NOT NULL,
  strategy_used VARCHAR(50) NOT NULL,
  
  candidates JSONB,
  selected_provider INT,
  selected_api_key_id INT REFERENCES merchant_api_keys(id) ON DELETE SET NULL,
  
  decision_latency_ms INT,
  was_retry BOOLEAN DEFAULT FALSE,
  retry_count INT DEFAULT 0,
  
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE routing_decisions IS '路由决策日志表';
COMMENT ON COLUMN routing_decisions.candidates IS '候选Provider列表及评分';
COMMENT ON COLUMN routing_decisions.selected_provider IS '选中的Provider ID';
COMMENT ON COLUMN routing_decisions.was_retry IS '是否为重试请求';

CREATE INDEX IF NOT EXISTS idx_routing_decisions_user ON routing_decisions(user_id);
CREATE INDEX IF NOT EXISTS idx_routing_decisions_model ON routing_decisions(model_requested);
CREATE INDEX IF NOT EXISTS idx_routing_decisions_created ON routing_decisions(created_at);

-- ============================================================
-- Phase 6: 创建更新时间触发器
-- ============================================================

CREATE OR REPLACE FUNCTION update_routing_strategies_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_routing_strategies_updated_at
  BEFORE UPDATE ON routing_strategies
  FOR EACH ROW
  EXECUTE FUNCTION update_routing_strategies_updated_at();
