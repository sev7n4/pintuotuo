-- 扩展 verification_result 枚举值，区分轻量验证和深度验证状态

-- 更新 verification_result 字段注释
COMMENT ON COLUMN merchant_api_keys.verification_result IS '验证结果: pending(待验证), verified(深度验证通过), suspend(余额不足), unreachable(连接失败), invalid(认证失败), failed(其他错误)';

-- 添加索引优化候选筛选查询
CREATE INDEX IF NOT EXISTS idx_merchant_api_keys_verification_result 
ON merchant_api_keys(verification_result);

-- 对于已有的 verified 记录，如果 pricing_verified 为 true，则保持 verified
-- 如果 pricing_verified 为 false 或 NULL，则设置为 pending（需要重新深度验证）
UPDATE merchant_api_keys mak
SET verification_result = 'pending'
FROM merchant_api_keys
WHERE mak.verification_result = 'verified'
  AND mak.id NOT IN (
    SELECT DISTINCT api_key_id 
    FROM api_key_verifications 
    WHERE status = 'success' 
      AND pricing_verified = true
  );
