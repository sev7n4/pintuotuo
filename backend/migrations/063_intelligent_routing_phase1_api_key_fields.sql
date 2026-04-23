-- 智能路由系统 Phase 1 - API Key 路由感知字段扩展
-- Version: 063
-- 目标：
-- 1) 扩展 merchant_api_keys 表，添加区域和安全等级字段
-- 2) 为智能路由决策层提供必要的路由感知数据

-- ============================================================================
-- 1. 扩展 merchant_api_keys 表
-- ============================================================================

-- 添加 region 字段（API Key 区域）
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS region VARCHAR(20) DEFAULT 'domestic';

-- 添加 security_level 字段（安全等级）
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS security_level VARCHAR(20) DEFAULT 'standard';

-- 确保 route_preference 字段存在（已在 062 迁移中添加，这里做兼容性检查）
-- 如果不存在则添加
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public'
    AND table_name = 'merchant_api_keys'
    AND column_name = 'route_preference'
  ) THEN
    ALTER TABLE merchant_api_keys ADD COLUMN route_preference JSONB DEFAULT '{}'::jsonb;
  END IF;
END $$;

-- ============================================================================
-- 2. 添加字段注释
-- ============================================================================

COMMENT ON COLUMN merchant_api_keys.region IS 'API Key 区域：domestic(国内), overseas(海外)';
COMMENT ON COLUMN merchant_api_keys.security_level IS '安全等级：standard(标准), high(高安全)';
COMMENT ON COLUMN merchant_api_keys.route_preference IS '路由偏好配置（JSONB）：包含策略偏好、自定义端点等配置';

-- ============================================================================
-- 3. 创建索引
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_merchant_api_keys_region ON merchant_api_keys(region);
CREATE INDEX IF NOT EXISTS idx_merchant_api_keys_security_level ON merchant_api_keys(security_level);

-- route_preference 已在 062 迁移中创建索引，这里做兼容性检查
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_indexes
    WHERE schemaname = 'public'
    AND tablename = 'merchant_api_keys'
    AND indexname = 'idx_merchant_api_keys_route_preference'
  ) THEN
    CREATE INDEX idx_merchant_api_keys_route_preference ON merchant_api_keys USING GIN(route_preference);
  END IF;
END $$;

-- ============================================================================
-- 4. 数据迁移 - 从 merchant_region 迁移到 region
-- ============================================================================

-- 如果存在 merchant_region 字段，将其值迁移到 region 字段
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public'
    AND table_name = 'merchant_api_keys'
    AND column_name = 'merchant_region'
  ) THEN
    UPDATE merchant_api_keys
    SET region = merchant_region
    WHERE region IS NULL OR region = 'domestic';
  END IF;
END $$;

-- ============================================================================
-- 5. 初始化默认值
-- ============================================================================

-- 将所有现有 API Key 设置为国内区域和标准安全等级
UPDATE merchant_api_keys
SET region = 'domestic',
    security_level = 'standard'
WHERE region IS NULL OR security_level IS NULL;
