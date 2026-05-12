-- 将 model_providers.endpoints 中历史占位主机改为 compose 服务名 litellm（与 LLM_GATEWAY_LITELLM_URL 默认一致）
-- 修复 backend 容器内 DNS 无法解析 litellm-domestic / litellm-overseas 的问题

UPDATE model_providers
SET endpoints = replace(
  replace(endpoints::text, 'http://litellm-domestic:4000/v1', 'http://litellm:4000/v1'),
  'http://litellm-overseas:4000/v1',
  'http://litellm:4000/v1'
)::jsonb,
updated_at = CURRENT_TIMESTAMP
WHERE endpoints::text LIKE '%litellm-domestic:4000%'
   OR endpoints::text LIKE '%litellm-overseas:4000%';
