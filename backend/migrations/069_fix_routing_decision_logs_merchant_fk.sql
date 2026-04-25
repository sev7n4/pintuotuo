-- 修改 routing_decision_logs 表，允许 merchant_id 为 NULL
-- 因为用户可能没有自己的 merchant 记录，但仍可使用其他 merchant 的 SKU

ALTER TABLE routing_decision_logs ALTER COLUMN merchant_id DROP NOT NULL;

-- 删除现有的外键约束
ALTER TABLE routing_decision_logs DROP CONSTRAINT IF EXISTS routing_decision_logs_merchant_id_fkey;

-- 添加新的外键约束，允许 NULL
ALTER TABLE routing_decision_logs 
ADD CONSTRAINT routing_decision_logs_merchant_id_fkey 
FOREIGN KEY (merchant_id) REFERENCES merchants(id) ON DELETE SET NULL;
