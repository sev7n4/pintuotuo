UPDATE model_providers
SET api_base_url = 'https://generativelanguage.googleapis.com/v1beta/openai',
    updated_at = CURRENT_TIMESTAMP
WHERE code = 'google'
  AND COALESCE(api_base_url, '') IN (
      'https://generativelanguage.googleapis.com/v1',
      'https://generativelanguage.googleapis.com/v1beta'
  );
