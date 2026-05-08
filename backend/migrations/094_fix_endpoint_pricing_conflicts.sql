-- Migration: 094_fix_endpoint_pricing_conflicts
-- Description: Fix images_generations unit_price conflict (0.02 -> 20.0) and add missing endpoint pricing records
-- Created: 2026-05-08

UPDATE endpoint_pricing
SET unit_price = 20.0, unit_type = 'image'
WHERE endpoint_type = 'images_generations' AND provider_code = 'openai' AND unit_type = 'image';

INSERT INTO endpoint_pricing (endpoint_type, provider_code, unit_type, unit_price) VALUES
    ('images_variations', 'openai', 'image', 20.0),
    ('images_edits', 'openai', 'image', 20.0),
    ('audio_translations', 'openai', 'second', 0.006),
    ('moderations', 'openai', 'request', 0.0),
    ('responses', 'openai', 'token', 0.001)
ON CONFLICT (endpoint_type, provider_code, unit_type) DO NOTHING;

-- Down Migration
-- UPDATE endpoint_pricing SET unit_price = 0.02 WHERE endpoint_type = 'images_generations' AND provider_code = 'openai' AND unit_type = 'image';
-- DELETE FROM endpoint_pricing WHERE endpoint_type IN ('images_variations', 'images_edits', 'audio_translations', 'moderations') AND provider_code = 'openai';
