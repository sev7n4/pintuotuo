-- Migration: 092_moderations_endpoint_pricing
-- Description: 审核端点计费配置（按请求计费，免费）
-- Created: 2026-05-08

INSERT INTO endpoint_pricing (endpoint_type, provider_code, unit_type, unit_price) VALUES
    ('moderations', 'openai', 'request', 0.0)
ON CONFLICT (endpoint_type, provider_code, unit_type) DO NOTHING;

-- Down Migration
-- DELETE FROM endpoint_pricing WHERE endpoint_type = 'moderations' AND provider_code = 'openai' AND unit_type = 'request';
