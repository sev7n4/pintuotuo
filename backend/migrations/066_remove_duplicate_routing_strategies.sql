-- 删除重复的路由策略：cost_first 和 auto
-- cost_first 与 price_first 语义重复
-- auto 与 balanced 语义重复

DELETE FROM routing_strategies WHERE code IN ('cost_first', 'auto');
