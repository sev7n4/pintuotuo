-- 014_add_categories.sql
-- 创建分类表

CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    level INT NOT NULL CHECK (level IN (1, 2)),
    description TEXT,
    icon VARCHAR(255),
    sort_order INT DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_categories_level ON categories(level);
CREATE INDEX IF NOT EXISTS idx_categories_active ON categories(is_active);
CREATE INDEX IF NOT EXISTS idx_categories_sort ON categories(sort_order);

-- 唯一约束
CREATE UNIQUE INDEX IF NOT EXISTS idx_categories_name_level ON categories(name, level);

-- 修改 products 表，添加分类关联字段
ALTER TABLE products 
ADD COLUMN IF NOT EXISTS model_id INT REFERENCES categories(id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS package_id INT REFERENCES categories(id) ON DELETE SET NULL;

-- 产品表索引
CREATE INDEX IF NOT EXISTS idx_products_model ON products(model_id);
CREATE INDEX IF NOT EXISTS idx_products_package ON products(package_id);

-- 插入默认一级分类：厂家/模型（level=1）
INSERT INTO categories (name, level, description, sort_order) VALUES
('GPT 系列', 1, 'OpenAI GPT 系列（GPT-3.5, GPT-4, GPT-4o 等）', 1),
('Claude 系列', 1, 'Anthropic Claude 系列（Claude 3 Opus/Sonnet/Haiku）', 2),
('Gemini 系列', 1, 'Google Gemini 系列', 3),
('文心一言', 1, '百度文心系列', 4),
('通义千问', 1, '阿里通义系列', 5),
('讯飞星火', 1, '讯飞星火系列', 6),
('智谱清言', 1, '智谱 AI GLM 系列', 7),
('月之暗面', 1, 'Moonshot Kimi 系列', 8),
('百川智能', 1, '百川大模型', 9),
('MiniMax', 1, 'MiniMax 大模型', 10),
('其他模型', 1, '其他 AI 模型', 99)
ON CONFLICT (name, level) DO NOTHING;

-- 插入默认二级分类：计费类型（level=2）
INSERT INTO categories (name, level, description, sort_order) VALUES
('免费试用', 2, '免费体验套餐', 0),
('月度基础版', 2, '月度入门级不限量套餐', 1),
('月度标准版', 2, '月度标准级不限量套餐', 2),
('月度高级版', 2, '月度高级不限量套餐', 3),
('季度基础版', 2, '季度入门级不限量套餐', 4),
('季度标准版', 2, '季度标准级不限量套餐', 5),
('季度高级版', 2, '季度高级不限量套餐', 6),
('年度基础版', 2, '年度入门级不限量套餐', 7),
('年度标准版', 2, '年度标准级不限量套餐', 8),
('年度高级版', 2, '年度高级不限量套餐', 9),
('按量付费', 2, '按 Token 计费', 10),
('企业定制', 2, '企业级定制方案', 11)
ON CONFLICT (name, level) DO NOTHING;
