-- SKU System Refactor Migration
-- Version: 015
-- Date: 2026-03-29
-- Description: 重构 SKU 系统，引入 SPU + SKU 架构

-- ============================================================
-- Phase 1: 创建新表结构
-- ============================================================

-- 1. 模型厂商配置表
CREATE TABLE IF NOT EXISTS model_providers (
  id SERIAL PRIMARY KEY,
  code VARCHAR(50) UNIQUE NOT NULL,
  name VARCHAR(100) NOT NULL,
  
  api_base_url VARCHAR(255),
  api_format VARCHAR(50) DEFAULT 'openai',
  
  billing_type VARCHAR(50),
  segment_config JSONB,
  
  cache_enabled BOOLEAN DEFAULT FALSE,
  cache_discount_rate DECIMAL(5, 2),
  
  status VARCHAR(50) DEFAULT 'active',
  sort_order INT DEFAULT 0,
  
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_model_providers_code ON model_providers(code);
CREATE INDEX IF NOT EXISTS idx_model_providers_status ON model_providers(status);

-- 2. SPU 表（标准产品单元）
CREATE TABLE IF NOT EXISTS spus (
  id SERIAL PRIMARY KEY,
  spu_code VARCHAR(100) UNIQUE NOT NULL,
  name VARCHAR(255) NOT NULL,
  
  -- 厂商对接字段
  model_provider VARCHAR(100) NOT NULL,
  provider_model_id VARCHAR(128),
  provider_api_endpoint VARCHAR(512),
  provider_auth_type VARCHAR(32) DEFAULT 'API_KEY',
  provider_billing_type VARCHAR(32),
  provider_input_rate DECIMAL(10,6),
  provider_output_rate DECIMAL(10,6),
  
  -- 模型信息
  model_name VARCHAR(100) NOT NULL,
  model_version VARCHAR(50),
  model_tier VARCHAR(50) NOT NULL CHECK (model_tier IN ('pro', 'lite', 'mini', 'vision')),
  
  -- 技术参数
  context_window INT,
  max_output_tokens INT,
  supported_functions JSONB,
  
  -- 算力点配置
  base_compute_points DECIMAL(10, 4) NOT NULL DEFAULT 1.0,
  billing_coefficient DECIMAL(5,2) DEFAULT 1.0,
  
  -- 描述信息
  description TEXT,
  features JSONB,
  thumbnail_url VARCHAR(500),
  
  -- 适配器配置
  input_length_ranges JSONB,
  billing_adapter JSONB,
  routing_rules JSONB,
  batch_inference JSONB,
  
  -- 状态
  status VARCHAR(50) NOT NULL DEFAULT 'active',
  sort_order INT DEFAULT 0,
  
  -- 统计
  total_sales_count BIGINT DEFAULT 0,
  average_rating DECIMAL(3, 2),
  
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON COLUMN spus.provider_model_id IS '厂商侧模型标识，如 ernie-4.0-8k';
COMMENT ON COLUMN spus.provider_api_endpoint IS '厂商API地址';
COMMENT ON COLUMN spus.provider_auth_type IS '认证方式: API_KEY/AK_SK/OAUTH';
COMMENT ON COLUMN spus.provider_billing_type IS '厂商计费类型: INPUT_OUTPUT/MIXED/FLAT';
COMMENT ON COLUMN spus.provider_input_rate IS '厂商输入Token单价(元/1K)';
COMMENT ON COLUMN spus.provider_output_rate IS '厂商输出Token单价(元/1K)';
COMMENT ON COLUMN spus.billing_coefficient IS '平台成本系数，用于利润计算';
COMMENT ON COLUMN spus.input_length_ranges IS '输入长度区间配置，如: [{"min_tokens":0,"max_tokens":32000,"label":"32K","surcharge":0}]';
COMMENT ON COLUMN spus.billing_adapter IS '计费适配器配置，如: {"type":"segment","cache_enabled":true,"cache_discount_rate":50}';
COMMENT ON COLUMN spus.routing_rules IS '智能路由规则，如: {"auto_route":true,"default_range":"32K"}';
COMMENT ON COLUMN spus.batch_inference IS '批量推理配置，如: {"enabled":true,"discount_rate":60,"async_only":true}';

CREATE INDEX IF NOT EXISTS idx_spus_code ON spus(spu_code);
CREATE INDEX IF NOT EXISTS idx_spus_provider ON spus(model_provider);
CREATE INDEX IF NOT EXISTS idx_spus_tier ON spus(model_tier);
CREATE INDEX IF NOT EXISTS idx_spus_status ON spus(status);

-- 3. SKU 表（库存量单位）
CREATE TABLE IF NOT EXISTS skus (
  id SERIAL PRIMARY KEY,
  spu_id INT NOT NULL,
  sku_code VARCHAR(100) UNIQUE NOT NULL,
  merchant_id INT,
  
  sku_type VARCHAR(50) NOT NULL CHECK (sku_type IN ('token_pack', 'subscription', 'concurrent', 'trial')),
  
  token_amount BIGINT,
  compute_points DECIMAL(15, 2),
  
  subscription_period VARCHAR(50) CHECK (subscription_period IN ('monthly', 'quarterly', 'yearly')),
  is_unlimited BOOLEAN DEFAULT FALSE,
  fair_use_limit BIGINT,
  
  tpm_limit INT,
  rpm_limit INT,
  concurrent_requests INT,
  
  valid_days INT DEFAULT 365,
  
  retail_price DECIMAL(10, 2) NOT NULL,
  wholesale_price DECIMAL(10, 2),
  original_price DECIMAL(10, 2),
  
  stock INT DEFAULT -1,
  daily_limit INT,
  
  group_enabled BOOLEAN DEFAULT TRUE,
  min_group_size INT DEFAULT 2,
  max_group_size INT DEFAULT 10,
  group_discount_rate DECIMAL(5, 2),
  
  is_trial BOOLEAN DEFAULT FALSE,
  trial_duration_days INT,
  
  status VARCHAR(50) NOT NULL DEFAULT 'active',
  is_promoted BOOLEAN DEFAULT FALSE,
  
  sales_count BIGINT DEFAULT 0,
  
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  
  FOREIGN KEY (spu_id) REFERENCES spus(id) ON DELETE CASCADE,
  FOREIGN KEY (merchant_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_skus_spu ON skus(spu_id);
CREATE INDEX IF NOT EXISTS idx_skus_merchant ON skus(merchant_id);
CREATE INDEX IF NOT EXISTS idx_skus_type ON skus(sku_type);
CREATE INDEX IF NOT EXISTS idx_skus_status ON skus(status);
CREATE INDEX IF NOT EXISTS idx_skus_code ON skus(sku_code);

-- 4. 算力点账户表
CREATE TABLE IF NOT EXISTS compute_point_accounts (
  id SERIAL PRIMARY KEY,
  user_id INT NOT NULL UNIQUE,
  
  balance DECIMAL(15, 2) NOT NULL DEFAULT 0,
  total_earned DECIMAL(15, 2) DEFAULT 0,
  total_used DECIMAL(15, 2) DEFAULT 0,
  total_expired DECIMAL(15, 2) DEFAULT 0,
  
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_cp_accounts_user ON compute_point_accounts(user_id);

-- 5. 算力点交易记录表
CREATE TABLE IF NOT EXISTS compute_point_transactions (
  id SERIAL PRIMARY KEY,
  user_id INT NOT NULL,
  
  type VARCHAR(50) NOT NULL CHECK (type IN ('purchase', 'reward', 'usage', 'refund', 'expire', 'group_bonus')),
  amount DECIMAL(15, 2) NOT NULL,
  balance_after DECIMAL(15, 2) NOT NULL,
  
  order_id INT,
  sku_id INT,
  
  description VARCHAR(500),
  metadata JSONB,
  
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE SET NULL,
  FOREIGN KEY (sku_id) REFERENCES skus(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_cp_trans_user ON compute_point_transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_cp_trans_type ON compute_point_transactions(type);
CREATE INDEX IF NOT EXISTS idx_cp_trans_created ON compute_point_transactions(created_at);

-- 6. 用户订阅表
CREATE TABLE IF NOT EXISTS user_subscriptions (
  id SERIAL PRIMARY KEY,
  user_id INT NOT NULL,
  sku_id INT NOT NULL,
  
  start_date DATE NOT NULL,
  end_date DATE NOT NULL,
  
  used_tokens BIGINT DEFAULT 0,
  used_compute_points DECIMAL(15, 2) DEFAULT 0,
  
  status VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'expired', 'cancelled')),
  auto_renew BOOLEAN DEFAULT FALSE,
  
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY (sku_id) REFERENCES skus(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_user ON user_subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON user_subscriptions(status);
CREATE INDEX IF NOT EXISTS idx_subscriptions_end ON user_subscriptions(end_date);

-- ============================================================
-- Phase 2: 初始化种子数据
-- ============================================================

-- 插入模型厂商
INSERT INTO model_providers (code, name, api_format, billing_type, status, sort_order) VALUES
('openai', 'OpenAI', 'openai', 'flat', 'active', 1),
('anthropic', 'Anthropic', 'anthropic', 'flat', 'active', 2),
('deepseek', 'DeepSeek', 'openai', 'flat', 'active', 3),
('zhipu', '智谱AI', 'openai', 'flat', 'active', 4),
('baidu', '百度千帆', 'baidu', 'segment', 'active', 5),
('bytedance', '字节跳动/火山引擎', 'openai', 'segment', 'active', 6),
('alibaba', '阿里云', 'openai', 'tiered', 'active', 7)
ON CONFLICT (code) DO NOTHING;

-- 插入示例 SPU（含适配器配置）
INSERT INTO spus (spu_code, name, model_provider, model_name, model_version, model_tier, context_window, base_compute_points, description, input_length_ranges, billing_adapter, routing_rules, batch_inference, status, sort_order) VALUES
('DEEPSEEK-V3', 'DeepSeek V3 模型服务', 'deepseek', 'deepseek-chat', 'v3', 'lite', 64, 1.0, 
 'DeepSeek V3 是一款高性能对话模型，适用于日常对话、内容创作等场景',
 '[{"min_tokens":0,"max_tokens":32000,"label":"32K","surcharge":0},{"min_tokens":32000,"max_tokens":64000,"label":"64K","surcharge":10}]'::jsonb,
 '{"type":"flat","cache_enabled":false}'::jsonb,
 '{"auto_route":true,"default_range":"32K"}'::jsonb,
 '{"enabled":true,"discount_rate":40,"async_only":true}'::jsonb,
 'active', 1),
('DEEPSEEK-V3-PRO', 'DeepSeek V3 Pro 模型服务', 'deepseek', 'deepseek-reasoner', 'v3', 'pro', 64, 2.5, 
 'DeepSeek V3 Pro 适用于复杂推理、代码生成等高级任务',
 '[{"min_tokens":0,"max_tokens":32000,"label":"32K","surcharge":0},{"min_tokens":32000,"max_tokens":64000,"label":"64K","surcharge":15}]'::jsonb,
 '{"type":"flat","cache_enabled":false}'::jsonb,
 '{"auto_route":true,"default_range":"32K"}'::jsonb,
 '{"enabled":true,"discount_rate":30,"async_only":true}'::jsonb,
 'active', 2),
('GLM-4', 'GLM-4 模型服务', 'zhipu', 'glm-4', 'v4', 'lite', 128, 1.2, 
 '智谱 GLM-4 是一款强大的通用对话模型',
 '[{"min_tokens":0,"max_tokens":64000,"label":"64K","surcharge":0},{"min_tokens":64000,"max_tokens":128000,"label":"128K","surcharge":20}]'::jsonb,
 '{"type":"flat","cache_enabled":true,"cache_discount_rate":50}'::jsonb,
 '{"auto_route":true,"default_range":"64K"}'::jsonb,
 '{"enabled":true,"discount_rate":35,"async_only":true}'::jsonb,
 'active', 3),
('GLM-4-PLUS', 'GLM-4 Plus 模型服务', 'zhipu', 'glm-4-plus', 'v4', 'pro', 128, 2.0, 
 '智谱 GLM-4 Plus 增强版，支持更复杂的任务',
 '[{"min_tokens":0,"max_tokens":64000,"label":"64K","surcharge":0},{"min_tokens":64000,"max_tokens":128000,"label":"128K","surcharge":25}]'::jsonb,
 '{"type":"flat","cache_enabled":true,"cache_discount_rate":50}'::jsonb,
 '{"auto_route":true,"default_range":"64K"}'::jsonb,
 '{"enabled":true,"discount_rate":30,"async_only":true}'::jsonb,
 'active', 4),
('ERNIE-4', '文心一言 ERNIE 4.0', 'baidu', 'ERNIE-4.0-8K', 'v4', 'pro', 8, 3.0, 
 '百度文心一言旗舰版，强大的中文理解能力',
 '[{"min_tokens":0,"max_tokens":8000,"label":"8K","surcharge":0}]'::jsonb,
 '{"type":"segment","segment_config":[{"input_range":"8K","multiplier":1.0}],"cache_enabled":false}'::jsonb,
 '{"auto_route":false,"default_range":"8K"}'::jsonb,
 '{"enabled":true,"discount_rate":60,"async_only":true}'::jsonb,
 'active', 5),
('ERNIE-SPEED', '文心一言 Speed', 'baidu', 'ERNIE-Speed', 'v1', 'mini', 8, 0.5, 
 '百度文心一言轻量版，快速响应',
 '[{"min_tokens":0,"max_tokens":8000,"label":"8K","surcharge":0}]'::jsonb,
 '{"type":"flat","cache_enabled":false}'::jsonb,
 '{"auto_route":false,"default_range":"8K"}'::jsonb,
 '{"enabled":false,"discount_rate":0,"async_only":false}'::jsonb,
 'active', 6),
('DOUBAO-PRO', '豆包 Pro 模型服务', 'bytedance', 'doubao-seed-2.0-pro', 'v2', 'pro', 256, 2.5, 
 '字节豆包旗舰版，支持超长上下文',
 '[{"min_tokens":0,"max_tokens":32000,"label":"32K","surcharge":0},{"min_tokens":32000,"max_tokens":128000,"label":"128K","surcharge":20},{"min_tokens":128000,"max_tokens":256000,"label":"256K","surcharge":50}]'::jsonb,
 '{"type":"segment","segment_config":[{"input_range":"32K","multiplier":1.0},{"input_range":"128K","multiplier":1.2},{"input_range":"256K","multiplier":1.5}],"cache_enabled":true,"cache_discount_rate":50}'::jsonb,
 '{"auto_route":true,"default_range":"32K","range_mapping":{"lite":"32K","pro":"128K"}}'::jsonb,
 '{"enabled":true,"discount_rate":40,"async_only":true}'::jsonb,
 'active', 7),
('DOUBAO-LITE', '豆包 Lite 模型服务', 'bytedance', 'doubao-seed-1.6-flash', 'v1.6', 'mini', 32, 0.8, 
 '字节豆包轻量版，实时响应',
 '[{"min_tokens":0,"max_tokens":32000,"label":"32K","surcharge":0}]'::jsonb,
 '{"type":"flat","cache_enabled":true,"cache_discount_rate":30}'::jsonb,
 '{"auto_route":false,"default_range":"32K"}'::jsonb,
 '{"enabled":false,"discount_rate":0,"async_only":false}'::jsonb,
 'active', 8),
('QWEN-MAX', '通义千问 Max', 'alibaba', 'qwen-max', 'v1', 'pro', 32, 2.0, 
 '阿里通义千问旗舰版',
 '[{"min_tokens":0,"max_tokens":32000,"label":"32K","surcharge":0}]'::jsonb,
 '{"type":"tiered","cache_enabled":false}'::jsonb,
 '{"auto_route":true,"default_range":"32K"}'::jsonb,
 '{"enabled":true,"discount_rate":35,"async_only":true}'::jsonb,
 'active', 9),
('QWEN-TURBO', '通义千问 Turbo', 'alibaba', 'qwen-turbo', 'v1', 'lite', 8, 0.8, 
 '阿里通义千问标准版',
 '[{"min_tokens":0,"max_tokens":8000,"label":"8K","surcharge":0}]'::jsonb,
 '{"type":"flat","cache_enabled":false}'::jsonb,
 '{"auto_route":false,"default_range":"8K"}'::jsonb,
 '{"enabled":true,"discount_rate":40,"async_only":true}'::jsonb,
 'active', 10)
ON CONFLICT (spu_code) DO NOTHING;

-- 插入示例 SKU（平台自营）
INSERT INTO skus (spu_id, sku_code, merchant_id, sku_type, token_amount, compute_points, retail_price, original_price, valid_days, status, group_enabled, group_discount_rate) VALUES
-- DeepSeek V3 Token 包
((SELECT id FROM spus WHERE spu_code = 'DEEPSEEK-V3'), 'DEEPSEEK-V3-100K', NULL, 'token_pack', 100000, 100, 9.90, 19.90, 365, 'active', TRUE, 20),
((SELECT id FROM spus WHERE spu_code = 'DEEPSEEK-V3'), 'DEEPSEEK-V3-500K', NULL, 'token_pack', 500000, 500, 39.90, 79.90, 365, 'active', TRUE, 25),
((SELECT id FROM spus WHERE spu_code = 'DEEPSEEK-V3'), 'DEEPSEEK-V3-1M', NULL, 'token_pack', 1000000, 1000, 69.90, 139.90, 365, 'active', TRUE, 30),

-- GLM-4 Token 包
((SELECT id FROM spus WHERE spu_code = 'GLM-4'), 'GLM-4-100K', NULL, 'token_pack', 100000, 120, 11.90, 23.90, 365, 'active', TRUE, 20),
((SELECT id FROM spus WHERE spu_code = 'GLM-4'), 'GLM-4-500K', NULL, 'token_pack', 500000, 600, 49.90, 99.90, 365, 'active', TRUE, 25),
((SELECT id FROM spus WHERE spu_code = 'GLM-4'), 'GLM-4-1M', NULL, 'token_pack', 1000000, 1200, 89.90, 179.90, 365, 'active', TRUE, 30),

-- 订阅型 SKU
((SELECT id FROM spus WHERE spu_code = 'DEEPSEEK-V3'), 'DEEPSEEK-V3-MONTHLY', NULL, 'subscription', NULL, 5000, 99.00, 199.00, 30, 'active', TRUE, 15),
((SELECT id FROM spus WHERE spu_code = 'GLM-4'), 'GLM-4-MONTHLY', NULL, 'subscription', NULL, 6000, 119.00, 239.00, 30, 'active', TRUE, 15),

-- 试用 SKU
((SELECT id FROM spus WHERE spu_code = 'DEEPSEEK-V3'), 'DEEPSEEK-V3-TRIAL', NULL, 'trial', 10000, 10, 0, NULL, 7, 'active', FALSE, NULL)
ON CONFLICT (sku_code) DO NOTHING;

-- ============================================================
-- Phase 3: 为现有用户创建算力点账户
-- ============================================================

INSERT INTO compute_point_accounts (user_id, balance, total_earned)
SELECT id, 0, 0 FROM users
WHERE id NOT IN (SELECT user_id FROM compute_point_accounts)
ON CONFLICT (user_id) DO NOTHING;

-- ============================================================
-- Phase 4: 创建兼容视图（过渡期使用）
-- ============================================================

CREATE OR REPLACE VIEW products_v2 AS
SELECT 
  s.id,
  s.sku_code as sku_code,
  s.merchant_id,
  p.name as name,
  sp.name || ' - ' || 
    CASE s.sku_type 
      WHEN 'token_pack' THEN s.token_amount::text || ' Tokens'
      WHEN 'subscription' THEN COALESCE(s.subscription_period, 'monthly')
      WHEN 'concurrent' THEN s.concurrent_requests::text || ' 并发'
      ELSE s.sku_type
    END as description,
  s.retail_price as price,
  s.original_price,
  CASE WHEN s.stock = -1 THEN 999999 ELSE s.stock END as stock,
  s.sales_count as sold_count,
  sp.model_tier as category,
  s.status,
  s.created_at,
  s.updated_at,
  s.spu_id as model_id,
  s.id as package_id
FROM skus s
JOIN spus sp ON s.spu_id = sp.id
LEFT JOIN products p ON p.id = s.id
WHERE s.status = 'active';

-- ============================================================
-- Phase 5: 更新触发器
-- ============================================================

CREATE OR REPLACE FUNCTION update_model_providers_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_model_providers_updated_at BEFORE UPDATE ON model_providers
  FOR EACH ROW EXECUTE FUNCTION update_model_providers_updated_at();

CREATE TRIGGER update_spus_updated_at BEFORE UPDATE ON spus
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_skus_updated_at BEFORE UPDATE ON skus
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_cp_accounts_updated_at BEFORE UPDATE ON compute_point_accounts
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_subscriptions_updated_at BEFORE UPDATE ON user_subscriptions
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- Phase 6: 性能优化索引
-- ============================================================

CREATE INDEX IF NOT EXISTS idx_skus_promoted ON skus(is_promoted) WHERE is_promoted = TRUE;
CREATE INDEX IF NOT EXISTS idx_skus_group_enabled ON skus(group_enabled) WHERE group_enabled = TRUE;
CREATE INDEX IF NOT EXISTS idx_spus_tier_status ON spus(model_tier, status);

-- ============================================================
-- Phase 7: 添加注释
-- ============================================================

COMMENT ON TABLE model_providers IS '模型厂商配置表';
COMMENT ON TABLE spus IS 'SPU标准产品单元表';
COMMENT ON TABLE skus IS 'SKU库存量单位表';
COMMENT ON TABLE compute_point_accounts IS '算力点账户表';
COMMENT ON TABLE compute_point_transactions IS '算力点交易记录表';
COMMENT ON TABLE user_subscriptions IS '用户订阅表';
