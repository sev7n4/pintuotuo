UPDATE merchant_api_keys mak
SET route_config = jsonb_set(
    route_config,
    '{endpoints,proxy}',
    '{}'::jsonb
),
updated_at = CURRENT_TIMESTAMP
FROM model_providers mp
WHERE mak.provider = mp.code
  AND mp.provider_region = 'overseas'
  AND mak.route_config->'endpoints'->'proxy' IS NOT NULL
  AND (
    mak.route_config->'endpoints'->'proxy'->>'gaap' LIKE '%example.com%'
    OR mak.route_config->'endpoints'->'proxy'->>'nginx_hk' LIKE '%example.com%'
  );
