# P1 监控看板与埋点校验

## 目标指标

- 订单转化率：`paid+completed / total_orders`
- 支付成功率：`payments.success / payments.total`
- 订单取消率：`orders.cancelled / total_orders`
- 多明细订单占比：`order_items 按 order_id 聚合后 item_count>1 / total_orders`

## 管理端看板

- 接口：`GET /api/v1/admin/stats`
- 新增字段：
  - `order_conversion_rate`
  - `payment_success_rate`
  - `cancellation_rate`
  - `multi_item_order_ratio`
  - `pending_orders` / `paid_orders` / `cancelled_orders`

## 埋点校验步骤（上线后每日巡检）

### 1) 数据一致性（SQL）

```sql
-- 订单总量
SELECT COUNT(*) AS total_orders FROM orders;

-- 支付成功率校验
SELECT
  COUNT(*) AS total_payments,
  COUNT(*) FILTER (WHERE status='success') AS success_payments
FROM payments;

-- 多明细订单占比校验
SELECT
  COUNT(*) AS multi_item_orders
FROM (
  SELECT order_id
  FROM order_items
  GROUP BY order_id
  HAVING COUNT(*) > 1
) t;
```

### 2) 接口校验

- 管理员调用 `/api/v1/admin/stats`
- 核对看板指标与上面 SQL 结果在可接受误差范围内（实时场景允许小幅延迟）

### 3) 业务链路抽检

- 抽取 5~10 个多明细订单：
  - `orders.total_price` 应等于 `order_items.total_price` 汇总
  - `orders.status` 与对应 `payments.status` 逻辑一致
  - 订单详情页明细条目数与 DB 一致

## 告警建议

- 支付成功率低于 90%（15 分钟窗口）告警
- 取消率高于 20%（日维度）告警
- 多明细订单占比突降（环比下降 > 50%）告警

