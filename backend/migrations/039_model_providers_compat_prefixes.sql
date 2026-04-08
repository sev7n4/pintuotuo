-- OpenAI 兼容入口：按无前缀 model 名推断 provider（数据驱动，避免在代码里写死厂商列表）。
-- 「provider/model」语法仍优先于前缀表。

ALTER TABLE model_providers ADD COLUMN IF NOT EXISTS compat_prefixes TEXT[] NOT NULL DEFAULT '{}';

COMMENT ON COLUMN model_providers.compat_prefixes IS '无前缀模型名匹配：小写比较 strings.HasPrefix；多前缀取最长匹配';

UPDATE model_providers SET compat_prefixes = ARRAY['deepseek']::text[], updated_at = CURRENT_TIMESTAMP WHERE code = 'deepseek';
UPDATE model_providers SET compat_prefixes = ARRAY['claude']::text[], updated_at = CURRENT_TIMESTAMP WHERE code = 'anthropic';
UPDATE model_providers SET compat_prefixes = ARRAY['gemini']::text[], updated_at = CURRENT_TIMESTAMP WHERE code = 'google';
UPDATE model_providers SET compat_prefixes = ARRAY['glm-', 'chatglm', 'cog-']::text[], updated_at = CURRENT_TIMESTAMP WHERE code = 'zhipu';
