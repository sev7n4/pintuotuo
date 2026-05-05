-- Migration: 082_endpoint_extension
-- Description: 扩展端点支持，添加 endpoint_type 字段和端点计费表
-- Created: 2026-05-05

-- 1. 添加 skus 表 endpoint_type 字段
ALTER TABLE skus ADD COLUMN IF NOT EXISTS endpoint_type VARCHAR(50) DEFAULT 'chat_completions';

-- 2. 创建 endpoint_pricing 表
CREATE TABLE IF NOT EXISTS endpoint_pricing (
    id SERIAL PRIMARY KEY,
    endpoint_type VARCHAR(50) NOT NULL,
    provider_code VARCHAR(50) NOT NULL,
    unit_type VARCHAR(20) NOT NULL,
    unit_price DECIMAL(10, 6) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(endpoint_type, provider_code)
);

-- 3. 创建索引
CREATE INDEX IF NOT EXISTS idx_endpoint_pricing_endpoint_type ON endpoint_pricing(endpoint_type);
CREATE INDEX IF NOT EXISTS idx_endpoint_pricing_provider_code ON endpoint_pricing(provider_code);

-- 4. 初始化端点计费配置数据
INSERT INTO endpoint_pricing (endpoint_type, provider_code, unit_type, unit_price) VALUES
    ('embeddings', 'openai', 'token', 0.0001),
    ('images_generations', 'openai', 'image', 0.02),
    ('audio_speech', 'openai', 'character', 0.000015),
    ('audio_transcriptions', 'openai', 'second', 0.006)
ON CONFLICT (endpoint_type, provider_code) DO NOTHING;

-- 5. 更新现有 SKU 的 endpoint_type 默认值
UPDATE skus SET endpoint_type = 'chat_completions' WHERE endpoint_type IS NULL OR endpoint_type = '';

-- Down Migration
-- DROP TABLE IF EXISTS endpoint_pricing;
-- ALTER TABLE skus DROP COLUMN IF EXISTS endpoint_type;
