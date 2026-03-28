-- ============================================
-- E2E Test Data Seed Script
-- This script creates test accounts for E2E testing
-- ============================================

-- Password hash algorithm: SHA256(password + JWT_SECRET)
-- Default JWT_SECRET: "pintuotuo-secret-key-dev"

-- Regular user: demo@example.com / demo123456
INSERT INTO users (email, name, password_hash, role, status)
VALUES (
    'demo@example.com',
    'Demo User',
    'd5e40f73f3eb863d24cff64aa15877de65a52fcea56d29db3585b90988a51311',
    'user',
    'active'
) ON CONFLICT (email) DO UPDATE SET
    password_hash = EXCLUDED.password_hash,
    role = EXCLUDED.role,
    status = EXCLUDED.status;

-- Merchant user: merchant@example.com / merchant123456
INSERT INTO users (email, name, password_hash, role, status)
VALUES (
    'merchant@example.com',
    'Test Merchant',
    'f43af1330d72420b14e89d803373cc3f2db79b3d293aec9257f227b10588e5c9',
    'merchant',
    'active'
) ON CONFLICT (email) DO UPDATE SET
    password_hash = EXCLUDED.password_hash,
    role = EXCLUDED.role,
    status = EXCLUDED.status;

-- Admin user: admin@example.com / admin123456
INSERT INTO users (email, name, password_hash, role, status)
VALUES (
    'admin@example.com',
    'Test Admin',
    'e1e80b6b77cc437cdfd1d183a7831b0a6b7c35c4b4d6a1c1d4b6932de9e75cc6',
    'admin',
    'active'
) ON CONFLICT (email) DO UPDATE SET
    password_hash = EXCLUDED.password_hash,
    role = EXCLUDED.role,
    status = EXCLUDED.status;

-- Create token records for test users
INSERT INTO tokens (user_id, balance, total_used, total_earned)
SELECT id, 100.00, 0, 100.00 FROM users WHERE email = 'demo@example.com'
ON CONFLICT (user_id) DO UPDATE SET balance = 100.00;

INSERT INTO tokens (user_id, balance, total_used, total_earned)
SELECT id, 500.00, 0, 500.00 FROM users WHERE email = 'merchant@example.com'
ON CONFLICT (user_id) DO UPDATE SET balance = 500.00;

INSERT INTO tokens (user_id, balance, total_used, total_earned)
SELECT id, 1000.00, 0, 1000.00 FROM users WHERE email = 'admin@example.com'
ON CONFLICT (user_id) DO UPDATE SET balance = 1000.00;

-- Verify test data
SELECT id, email, name, role, status FROM users 
WHERE email IN ('demo@example.com', 'merchant@example.com', 'admin@example.com');

-- Insert default categories if not exists (using simple INSERT with ON CONFLICT)
-- 一级分类：厂家/模型（level=1）
INSERT INTO categories (name, level, description, sort_order) VALUES
('GPT 系列', 1, 'OpenAI GPT 系列（GPT-3.5, GPT-4, GPT-4o 等）', 1),
('Claude 系列', 1, 'Anthropic Claude 系列（Claude 3 Opus/Sonnet/Haiku）', 2),
('Gemini 系列', 1, 'Google Gemini 系列', 3),
('其他模型', 1, '其他 AI 模型', 99)
ON CONFLICT (name, level) DO NOTHING;

-- 二级分类：计费类型（level=2）
INSERT INTO categories (name, level, description, sort_order) VALUES
('免费试用', 2, '免费体验套餐', 0),
('月度基础版', 2, '月度入门级不限量套餐', 1),
('月度标准版', 2, '月度标准级不限量套餐', 2),
('月度高级版', 2, '月度高级不限量套餐', 3),
('按量付费', 2, '按 Token 计费', 10),
('企业定制', 2, '企业级定制方案', 11)
ON CONFLICT (name, level) DO NOTHING;

-- Verify categories
SELECT id, name, level, description FROM categories ORDER BY level, sort_order;
