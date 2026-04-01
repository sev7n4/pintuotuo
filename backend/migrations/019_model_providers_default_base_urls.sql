-- 为种子数据中的 model_providers 补齐默认 api_base_url。
-- 代理层会按 api_format 拼接路径：openai 系为 {base}/chat/completions，anthropic 为 {base}/messages。
-- 部署后请确保至少有一条商户 merchant_api_keys 对应 provider（如 zhipu）且密钥有效，否则仍会报无可用 Key。

UPDATE model_providers SET
  api_base_url = 'https://api.openai.com/v1',
  updated_at = CURRENT_TIMESTAMP
WHERE code = 'openai';

UPDATE model_providers SET
  api_base_url = 'https://api.anthropic.com/v1',
  updated_at = CURRENT_TIMESTAMP
WHERE code = 'anthropic';

UPDATE model_providers SET
  api_base_url = 'https://api.deepseek.com',
  updated_at = CURRENT_TIMESTAMP
WHERE code = 'deepseek';

UPDATE model_providers SET
  api_base_url = 'https://open.bigmodel.cn/api/paas/v4',
  updated_at = CURRENT_TIMESTAMP
WHERE code = 'zhipu';
