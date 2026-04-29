-- BYOK路由模式支持迁移
-- 扩展api_key_verifications表，添加路由模式相关字段

-- 扩展api_key_verifications表
ALTER TABLE api_key_verifications 
  ADD COLUMN IF NOT EXISTS route_mode VARCHAR(20),
  ADD COLUMN IF NOT EXISTS endpoint_used VARCHAR(512),
  ADD COLUMN IF NOT EXISTS error_category VARCHAR(64);

-- 添加注释
COMMENT ON COLUMN api_key_verifications.route_mode IS '验证时使用的路由模式: direct/litellm/proxy/auto';
COMMENT ON COLUMN api_key_verifications.endpoint_used IS '验证时实际使用的endpoint地址';
COMMENT ON COLUMN api_key_verifications.error_category IS '错误分类（参考provider_error_mapper.go）';

-- 添加索引
CREATE INDEX IF NOT EXISTS idx_api_key_verifications_route_mode ON api_key_verifications(route_mode);
CREATE INDEX IF NOT EXISTS idx_api_key_verifications_error_category ON api_key_verifications(error_category);
