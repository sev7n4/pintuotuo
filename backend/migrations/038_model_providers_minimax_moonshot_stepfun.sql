-- 扩展模型厂商：MiniMax、月之暗面（Kimi / Moonshot API）、阶跃星辰
-- OpenAI 兼容代理使用 {api_base_url}/chat/completions；部署后可于 Admin 修正 base URL。

INSERT INTO model_providers (code, name, api_format, billing_type, api_base_url, status, sort_order) VALUES
('minimax', 'MiniMax', 'openai', 'flat', 'https://api.minimax.chat/v1', 'active', 8),
('moonshot', '月之暗面 Kimi', 'openai', 'flat', 'https://api.moonshot.cn/v1', 'active', 9),
('stepfun', '阶跃星辰', 'openai', 'flat', 'https://api.stepfun.com/v1', 'active', 10)
ON CONFLICT (code) DO NOTHING;
