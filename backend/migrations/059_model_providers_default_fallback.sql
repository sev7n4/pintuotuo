-- Endpoint resolution fallback provider (__default__)
-- Version: 059
-- 目标：
-- 1) 确保存在一条可用的兜底 provider（code=__default__）
-- 2) 在不覆盖已有人工配置的前提下补齐关键字段（status/api_base_url/api_format）

INSERT INTO model_providers (
  code, name, api_base_url, api_format, billing_type, status, sort_order, compat_prefixes
)
SELECT
  '__default__',
  '系统兜底提供商',
  COALESCE(
    (SELECT NULLIF(TRIM(api_base_url), '') FROM model_providers WHERE code = 'openai' AND status = 'active' ORDER BY updated_at DESC, id DESC LIMIT 1),
    'https://api.openai.com/v1'
  ),
  'openai',
  'flat',
  'inactive',
  9999,
  '{}'::text[]
WHERE NOT EXISTS (
  SELECT 1 FROM model_providers WHERE code = '__default__'
);

UPDATE model_providers
SET
  name = COALESCE(NULLIF(TRIM(name), ''), '系统兜底提供商'),
  status = COALESCE(NULLIF(TRIM(status), ''), 'inactive'),
  api_format = COALESCE(NULLIF(TRIM(api_format), ''), 'openai'),
  billing_type = COALESCE(NULLIF(TRIM(billing_type), ''), 'flat'),
  sort_order = COALESCE(sort_order, 9999),
  api_base_url = COALESCE(
    NULLIF(TRIM(api_base_url), ''),
    (SELECT NULLIF(TRIM(api_base_url), '') FROM model_providers WHERE code = 'openai' AND status = 'active' ORDER BY updated_at DESC, id DESC LIMIT 1),
    'https://api.openai.com/v1'
  ),
  updated_at = CURRENT_TIMESTAMP
WHERE code = '__default__';
