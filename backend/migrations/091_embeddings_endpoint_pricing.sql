-- Migration: 091_embeddings_endpoint_pricing
-- Description: 嵌入端点计费配置
-- Created: 2026-05-08

INSERT INTO endpoint_pricing (endpoint_type, provider_code, unit_type, unit_price) VALUES
    ('embeddings', 'openai', 'token', 0.0001)
ON CONFLICT (endpoint_type, provider_code, unit_type) DO NOTHING;

-- Down Migration
-- DELETE FROM endpoint_pricing WHERE endpoint_type = 'embeddings' AND provider_code = 'openai' AND unit_type = 'token';
