-- IE-2: 零售单一账本 — 将历史 compute_point_accounts 并入 tokens
-- Version: 046
-- Date: 2026-04-10
-- 幂等：合并后清零算力账户，避免每次 migrate 重复叠加

INSERT INTO tokens (user_id, balance, total_used, total_earned)
SELECT user_id, balance, total_used, total_earned
FROM compute_point_accounts
WHERE balance <> 0 OR total_used <> 0 OR total_earned <> 0
ON CONFLICT (user_id) DO UPDATE SET
  balance = tokens.balance + EXCLUDED.balance,
  total_used = tokens.total_used + EXCLUDED.total_used,
  total_earned = tokens.total_earned + EXCLUDED.total_earned,
  updated_at = CURRENT_TIMESTAMP;

UPDATE compute_point_accounts
SET balance = 0, total_used = 0, total_earned = 0, updated_at = CURRENT_TIMESTAMP;

COMMENT ON TABLE compute_point_accounts IS '已弃用：零售余额已合并至 tokens（046）；行保留为零占位，仅作历史兼容';
