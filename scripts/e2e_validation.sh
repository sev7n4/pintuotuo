#!/bin/bash

set -e

echo "=========================================="
echo "端到端业务流程验证脚本"
echo "=========================================="
echo ""

DB_HOST="localhost"
DB_PORT="5432"
DB_NAME="pintuotuo"
DB_USER="postgres"
DB_PASS="your_password"

API_BASE="http://localhost:8080"

ADMIN_EMAIL="admin@163.com"
ADMIN_PASS="111111"
MERCHANT_EMAIL="mc100@163.com"
MERCHANT_PASS="111111"
USER_EMAIL="user100@163.com"
USER_PASS="111111"

echo "=== 1. 用户购买 SKU 后余额验证 ==="
echo ""

echo "检查用户 Token 余额..."
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
SELECT 
    u.id as user_id,
    u.email,
    t.balance as token_balance,
    t.created_at,
    t.updated_at
FROM users u
LEFT JOIN tokens t ON u.id = t.user_id
WHERE u.email = '$USER_EMAIL';
"

echo ""
echo "检查用户订单履约记录..."
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
SELECT 
    o.id as order_id,
    o.user_id,
    o.sku_id,
    s.name as sku_name,
    s.token_amount,
    o.status,
    o.created_at
FROM orders o
JOIN skus s ON o.sku_id = s.id
WHERE o.user_id = (SELECT id FROM users WHERE email = '$USER_EMAIL')
ORDER BY o.created_at DESC
LIMIT 5;
"

echo ""
echo "=== 2. API 调用预扣费验证 ==="
echo ""

echo "检查预扣费记录..."
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
SELECT 
    pd.id,
    pd.user_id,
    pd.request_id,
    pd.pre_deduct_amount,
    pd.status,
    pd.created_at,
    pd.settled_at
FROM pre_deductions pd
WHERE pd.user_id = (SELECT id FROM users WHERE email = '$USER_EMAIL')
ORDER BY pd.created_at DESC
LIMIT 10;
"

echo ""
echo "=== 3. API 调用结算验证 ==="
echo ""

echo "检查 API 使用日志..."
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
SELECT 
    aul.id,
    aul.user_id,
    aul.merchant_id,
    aul.model,
    aul.input_tokens,
    aul.output_tokens,
    aul.token_usage,
    aul.cost,
    aul.created_at
FROM api_usage_logs aul
WHERE aul.user_id = (SELECT id FROM users WHERE email = '$USER_EMAIL')
ORDER BY aul.created_at DESC
LIMIT 10;
"

echo ""
echo "=== 4. 商户结算金额验证 ==="
echo ""

echo "检查商户结算记录..."
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
SELECT 
    ms.id,
    ms.merchant_id,
    m.name as merchant_name,
    ms.period_start,
    ms.period_end,
    ms.total_sales_cny,
    ms.total_tokens,
    ms.status,
    ms.created_at
FROM merchant_settlements ms
JOIN merchants m ON ms.merchant_id = m.id
WHERE ms.merchant_id = (SELECT id FROM merchants WHERE email = '$MERCHANT_EMAIL')
ORDER BY ms.created_at DESC
LIMIT 5;
"

echo ""
echo "检查商户账单明细..."
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
SELECT 
    mb.id,
    mb.merchant_id,
    mb.type,
    mb.amount,
    mb.description,
    mb.created_at
FROM merchant_bills mb
WHERE mb.merchant_id = (SELECT id FROM merchants WHERE email = '$MERCHANT_EMAIL')
ORDER BY mb.created_at DESC
LIMIT 10;
"

echo ""
echo "=== 5. 数据一致性验证 ==="
echo ""

echo "验证用户 Token 消费与日志一致性..."
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
WITH user_tokens AS (
    SELECT 
        u.id as user_id,
        u.email,
        t.balance as current_balance
    FROM users u
    JOIN tokens t ON u.id = t.user_id
    WHERE u.email = '$USER_EMAIL'
),
consumption AS (
    SELECT 
        user_id,
        SUM(token_usage) as total_consumed
    FROM api_usage_logs
    WHERE user_id = (SELECT id FROM users WHERE email = '$USER_EMAIL')
    GROUP BY user_id
)
SELECT 
    ut.user_id,
    ut.email,
    ut.current_balance,
    COALESCE(c.total_consumed, 0) as total_consumed,
    (SELECT token_amount FROM skus s 
     JOIN orders o ON s.id = o.sku_id 
     WHERE o.user_id = ut.user_id AND o.status = 'completed'
     ORDER BY o.created_at DESC LIMIT 1) as last_order_tokens
FROM user_tokens ut
LEFT JOIN consumption c ON ut.user_id = c.user_id;
"

echo ""
echo "验证商户结算与 API 使用日志一致性..."
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
SELECT 
    m.id as merchant_id,
    m.name as merchant_name,
    COALESCE(SUM(aul.token_usage), 0) as total_tokens_from_logs,
    COALESCE(SUM(aul.cost), 0) as total_cost_from_logs,
    COALESCE((SELECT SUM(total_tokens) FROM merchant_settlements WHERE merchant_id = m.id), 0) as total_tokens_from_settlements,
    COALESCE((SELECT SUM(total_sales_cny) FROM merchant_settlements WHERE merchant_id = m.id), 0) as total_sales_from_settlements
FROM merchants m
LEFT JOIN api_usage_logs aul ON m.id = aul.merchant_id
WHERE m.email = '$MERCHANT_EMAIL'
GROUP BY m.id, m.name;
"

echo ""
echo "=== 6. 预扣费配置验证 ==="
echo ""

echo "检查 SKU 级别预扣费配置..."
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
SELECT 
    id,
    name,
    token_amount,
    pre_deduct_multiplier,
    pre_deduct_max_multiplier,
    created_at,
    updated_at
FROM skus
WHERE pre_deduct_multiplier IS NOT NULL
ORDER BY updated_at DESC
LIMIT 10;
"

echo ""
echo "检查 SPU 级别预扣费配置..."
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
SELECT 
    id,
    name,
    provider_code,
    pre_deduct_multiplier,
    pre_deduct_max_multiplier,
    created_at,
    updated_at
FROM spus
WHERE pre_deduct_multiplier IS NOT NULL
ORDER BY updated_at DESC
LIMIT 10;
"

echo ""
echo "=== 7. 余额不足阻断验证 ==="
echo ""

echo "检查是否有余额不足的阻断记录..."
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
SELECT 
    u.email,
    t.balance,
    COUNT(pd.id) as pre_deduct_attempts,
    SUM(CASE WHEN pd.status = 'cancelled' THEN 1 ELSE 0 END) as cancelled_count
FROM users u
JOIN tokens t ON u.id = t.user_id
LEFT JOIN pre_deductions pd ON u.id = pd.user_id
WHERE u.email = '$USER_EMAIL'
GROUP BY u.email, t.balance;
"

echo ""
echo "=========================================="
echo "端到端验证完成"
echo "=========================================="
