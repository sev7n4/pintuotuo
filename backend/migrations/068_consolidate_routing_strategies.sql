-- 整合路由策略配置
-- 1. 删除重复策略：performance_first（与latency_first重复）、security_first（权重配置错误且无特殊安全逻辑）
-- 2. 修复 reliability_first 权重配置（reliability_weight 应为最高）
-- 3. 删除 merchant_api_keys.route_preference 列（未使用）

-- 1. 删除重复策略
DELETE FROM routing_strategies WHERE code IN ('performance_first', 'security_first');

-- 2. 修复 reliability_first 权重配置
UPDATE routing_strategies 
SET price_weight = 0.20, 
    latency_weight = 0.20, 
    reliability_weight = 0.60,
    description = '优先选择最可靠的Provider，可靠性权重最高，适合稳定性要求高的场景'
WHERE code = 'reliability_first';

-- 3. 删除未使用的 route_preference 列
ALTER TABLE merchant_api_keys DROP COLUMN IF EXISTS route_preference;
