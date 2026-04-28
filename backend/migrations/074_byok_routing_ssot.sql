-- BYOK 路由配置 SSOT 改进
-- 将路由配置下沉到 merchant_api_keys 表

-- 1. 新增字段
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS byok_type VARCHAR(20) DEFAULT 'official';
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS route_mode VARCHAR(20) DEFAULT 'auto';
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS fallback_endpoint_url VARCHAR(500);
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS route_config JSONB DEFAULT '{}'::jsonb;

-- 2. 添加注释
COMMENT ON COLUMN merchant_api_keys.byok_type IS 'BYOK类别: official(官方), reseller(代理商), self_hosted(自建商)';
COMMENT ON COLUMN merchant_api_keys.route_mode IS '路由出站模式: auto(自动), direct(直连), litellm, proxy(代理)';
COMMENT ON COLUMN merchant_api_keys.fallback_endpoint_url IS '备用端点URL，主端点不可用时使用';
COMMENT ON COLUMN merchant_api_keys.route_config IS '完整路由配置（JSONB）: gateway_mode, endpoints, timeout, retry等';

-- 3. 创建索引
CREATE INDEX IF NOT EXISTS idx_merchant_api_keys_byok_type ON merchant_api_keys(byok_type);
CREATE INDEX IF NOT EXISTS idx_merchant_api_keys_route_mode ON merchant_api_keys(route_mode);
CREATE INDEX IF NOT EXISTS idx_merchant_api_keys_route_config ON merchant_api_keys USING GIN(route_config);

-- 4. 数据迁移：从 model_providers 复制默认配置到 merchant_api_keys
UPDATE merchant_api_keys mak
SET route_config = jsonb_build_object(
    'gateway_mode', COALESCE(
        mp.route_strategy->>'default_mode',
        'auto'
    ),
    'endpoints', COALESCE(mp.endpoints, '{}'::jsonb)
)
FROM model_providers mp
WHERE mak.provider = mp.code
  AND (mak.route_config IS NULL OR mak.route_config = '{}'::jsonb);
