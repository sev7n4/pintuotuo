ALTER TABLE endpoint_pricing DROP CONSTRAINT IF EXISTS endpoint_pricing_endpoint_type_provider_code_key;
ALTER TABLE endpoint_pricing ADD CONSTRAINT endpoint_pricing_endpoint_type_provider_code_unit_type_key UNIQUE (endpoint_type, provider_code, unit_type);

INSERT INTO endpoint_pricing (endpoint_type, provider_code, unit_type, unit_price) VALUES
    ('responses', 'openai', 'request', 0.01),
    ('responses', 'openai', 'image', 20.0)
ON CONFLICT (endpoint_type, provider_code, unit_type) DO NOTHING;
