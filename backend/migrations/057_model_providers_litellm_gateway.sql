-- LiteLLM 网关映射字段化：model_providers 为 litellm-catalog-sync 的主 SSOT（模板 + 密钥环境变量 + 可选 api_base）。
-- 可选 JSON 覆盖仍见 deploy/litellm/provider_gateway_map.json（-map 合并，文件优先）。

ALTER TABLE model_providers ADD COLUMN IF NOT EXISTS litellm_model_template TEXT;
ALTER TABLE model_providers ADD COLUMN IF NOT EXISTS litellm_gateway_api_key_env VARCHAR(128);
ALTER TABLE model_providers ADD COLUMN IF NOT EXISTS litellm_gateway_api_base VARCHAR(512);

COMMENT ON COLUMN model_providers.litellm_model_template IS 'LiteLLM litellm_params.model 模板，可含 {model_id}';
COMMENT ON COLUMN model_providers.litellm_gateway_api_key_env IS 'LiteLLM 容器内 API Key 环境变量名（如 OPENAI_API_KEY）';
COMMENT ON COLUMN model_providers.litellm_gateway_api_base IS '可选：OpenAI 兼容 api_base（如阶跃独立 base）';

-- 与 deploy/litellm/provider_gateway_map.json 等价（历史备份）
UPDATE model_providers SET litellm_model_template = 'openai/{model_id}', litellm_gateway_api_key_env = 'OPENAI_API_KEY', litellm_gateway_api_base = NULL, updated_at = CURRENT_TIMESTAMP WHERE code = 'openai';
UPDATE model_providers SET litellm_model_template = 'anthropic/{model_id}', litellm_gateway_api_key_env = 'ANTHROPIC_API_KEY', litellm_gateway_api_base = NULL, updated_at = CURRENT_TIMESTAMP WHERE code = 'anthropic';
UPDATE model_providers SET litellm_model_template = 'deepseek/{model_id}', litellm_gateway_api_key_env = 'DEEPSEEK_API_KEY', litellm_gateway_api_base = NULL, updated_at = CURRENT_TIMESTAMP WHERE code = 'deepseek';
UPDATE model_providers SET litellm_model_template = 'zai/{model_id}', litellm_gateway_api_key_env = 'ZAI_API_KEY', litellm_gateway_api_base = NULL, updated_at = CURRENT_TIMESTAMP WHERE code = 'zhipu';
UPDATE model_providers SET litellm_model_template = 'dashscope/{model_id}', litellm_gateway_api_key_env = 'DASHSCOPE_API_KEY', litellm_gateway_api_base = NULL, updated_at = CURRENT_TIMESTAMP WHERE code = 'alibaba';
UPDATE model_providers SET litellm_model_template = 'moonshot/{model_id}', litellm_gateway_api_key_env = 'MOONSHOT_API_KEY', litellm_gateway_api_base = NULL, updated_at = CURRENT_TIMESTAMP WHERE code = 'moonshot';
UPDATE model_providers SET litellm_model_template = 'minimax/{model_id}', litellm_gateway_api_key_env = 'MINIMAX_API_KEY', litellm_gateway_api_base = NULL, updated_at = CURRENT_TIMESTAMP WHERE code = 'minimax';
UPDATE model_providers SET litellm_model_template = 'openai/{model_id}', litellm_gateway_api_key_env = 'STEPFUN_API_KEY', litellm_gateway_api_base = 'https://api.stepfun.com/v1', updated_at = CURRENT_TIMESTAMP WHERE code = 'stepfun';

-- 种子库中可能尚无 google 行；039 曾 UPDATE google 的 compat_prefixes，此处补 INSERT 以便映射完整
INSERT INTO model_providers (code, name, api_format, billing_type, status, sort_order, compat_prefixes, litellm_model_template, litellm_gateway_api_key_env, litellm_gateway_api_base)
SELECT 'google', 'Google Gemini', 'openai', 'flat', 'active', 11, ARRAY['gemini']::text[], 'gemini/{model_id}', 'GOOGLE_API_KEY', NULL
WHERE NOT EXISTS (SELECT 1 FROM model_providers WHERE code = 'google');

UPDATE model_providers SET litellm_model_template = 'gemini/{model_id}', litellm_gateway_api_key_env = 'GOOGLE_API_KEY', litellm_gateway_api_base = NULL, updated_at = CURRENT_TIMESTAMP WHERE code = 'google';
