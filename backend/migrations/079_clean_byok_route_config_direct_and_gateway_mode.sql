UPDATE merchant_api_keys
SET route_config = (route_config - 'gateway_mode') ||
                   jsonb_set(
                     COALESCE(route_config, '{}'::jsonb),
                     '{endpoints}',
                     COALESCE(route_config->'endpoints', '{}'::jsonb) - 'direct'
                   ),
    updated_at = CURRENT_TIMESTAMP
WHERE route_config->'endpoints'->'direct' IS NOT NULL
   OR route_config?'gateway_mode';
