UPDATE model_providers
SET api_base_url = 'https://generativelanguage.googleapis.com/v1beta/openai/v1',
    updated_at = CURRENT_TIMESTAMP
WHERE code = 'google'
  AND api_base_url = 'https://generativelanguage.googleapis.com/v1beta/openai';
