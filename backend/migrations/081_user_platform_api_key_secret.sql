-- C 端平台 api_keys：加密存储完整密钥以便用户随时「查看/复制」，列表仅返回预览串（非机密）
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS key_encrypted TEXT;
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS key_preview VARCHAR(48);

COMMENT ON COLUMN api_keys.key_encrypted IS '完整 ptd_ 密钥的 AES-GCM 密文（base64），与 ENCRYPTION_KEY 对应；历史 NULL 表示不可揭示';
COMMENT ON COLUMN api_keys.key_preview IS '列表展示用预览，如 ptd_abcd…wxyz（不含完整密钥）';
