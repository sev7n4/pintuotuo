-- Migration: 090_images_audio_endpoint_pricing
-- Description: 图像+音频端点计费配置
-- Created: 2026-05-08

INSERT INTO endpoint_pricing (endpoint_type, provider_code, unit_type, unit_price) VALUES
    ('images_generations', 'openai', 'image', 20.0),
    ('images_variations', 'openai', 'image', 20.0),
    ('images_edits', 'openai', 'image', 20.0),
    ('audio_speech', 'openai', 'character', 0.000015),
    ('audio_transcriptions', 'openai', 'second', 0.006),
    ('audio_translations', 'openai', 'second', 0.006)
ON CONFLICT (endpoint_type, provider_code, unit_type) DO NOTHING;

-- Down Migration
-- DELETE FROM endpoint_pricing WHERE endpoint_type IN ('images_generations', 'images_variations', 'images_edits', 'audio_speech', 'audio_transcriptions', 'audio_translations') AND provider_code = 'openai';
