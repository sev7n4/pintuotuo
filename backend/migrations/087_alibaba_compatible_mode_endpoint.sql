UPDATE model_providers
SET endpoints = jsonb_set(
  jsonb_set(
    endpoints,
    '{direct,domestic}',
    '"https://dashscope.aliyuncs.com/compatible-mode/v1"'
  ),
  '{direct,overseas}',
  '"https://dashscope.aliyuncs.com/compatible-mode/v1"'
)
WHERE code = 'alibaba'
  AND endpoints->'direct'->>'domestic' = 'https://dashscope.aliyuncs.com/api/v1';
