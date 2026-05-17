-- OpenRouter LiteLLM 网关映射（与 openrouter/* model_list 及 BYOK 探测对齐）
UPDATE model_providers
SET litellm_model_template = 'openrouter/{model_id}',
    litellm_gateway_api_key_env = 'OPENROUTER_API_KEY',
    litellm_gateway_api_base = NULL,
    api_base_url = COALESCE(NULLIF(TRIM(api_base_url), ''), 'https://openrouter.ai/api/v1'),
    updated_at = CURRENT_TIMESTAMP
WHERE code = 'openrouter';

INSERT INTO model_providers (
    code, name, api_format, billing_type, status, sort_order,
    api_base_url, litellm_model_template, litellm_gateway_api_key_env
)
SELECT
    'openrouter', 'OpenRouter', 'openai', 'flat', 'active', 50,
    'https://openrouter.ai/api/v1', 'openrouter/{model_id}', 'OPENROUTER_API_KEY'
WHERE NOT EXISTS (SELECT 1 FROM model_providers WHERE code = 'openrouter');
