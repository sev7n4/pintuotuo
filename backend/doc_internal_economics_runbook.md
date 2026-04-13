# 内部经济运维 Runbook（对账与排障）

**适用**：零售单一账本、`api_proxy` 按量扣费、`pricing_version` 快照价目（IE-1～IE-6）。

**相关设计说明**：[`doc_internal_token_economics.md`](./doc_internal_token_economics.md)

---

## 1. 单位口径（先看再查数）

- 用户侧余额：`tokens.balance`，与 **我的 Token** 页展示一致，单位为平台内部 **Token**（与订单快照入账同量纲）。
- 每次代理成功调用：`api_usage_logs.cost` = 本次扣减量；同时写入 `token_transactions`（`type = 'usage'`，`amount` 为**负数**，绝对值等于 `cost`）。
- **报表里的 `total_cost`、账单明细里的 `cost` 均不是人民币 GMV**；人民币在订单/支付流水里体现。

---

## 2. 快速对账：单用户「用量日志 vs 流水」

在只读事务或从库执行（将 `:user_id` 换成实际用户 ID）：

```sql
SELECT
  COALESCE((SELECT SUM(cost) FROM api_usage_logs WHERE user_id = :user_id), 0)
    AS sum_api_usage_cost,
  COALESCE((SELECT SUM(-amount) FROM token_transactions WHERE user_id = :user_id AND type = 'usage'), 0)
    AS sum_usage_token_tx;
```

**期望**：两行数值在浮点误差内一致（通常 6 位小数内无差）。若长期偏离，排查：

1. 是否存在**仅写日志未扣余额**或**仅扣余额未写日志**的历史缺陷（应用版本、事务回滚）。
2. 是否有**手工改表**或离线脚本只动了一张表。
3. `token_transactions` 中 `type = 'usage'` 是否被误删、合并脚本是否只迁一张表。

---

## 3. 快速对账：余额 vs 流水累计（粗检）

入账类（`purchase`/`recharge` 等）与消耗类（`usage`）应满足「期初 + 入账 − |消耗| ≈ 当前余额」（需结合业务允许的舍入与历史迁移）。更精确做法以财务/审计脚本为准；此处仅作运维粗检。

```sql
SELECT user_id, balance, total_used, total_earned
FROM tokens
WHERE user_id = :user_id;
```

将 `total_used` 与 `SUM(cost)`（`api_usage_logs`）对照，应一致或可追溯迁移说明。

---

## 4. 价目版本（IE-4）排障

若怀疑「扣费未按购买时价目」：

1. 查该用户**最近一笔已支付且已履约**订单是否带 `pricing_version_id`：  
   `SELECT id, pricing_version_id, fulfilled_at FROM orders WHERE user_id = ? AND status = 'paid' ORDER BY fulfilled_at DESC LIMIT 5;`
2. 无版本时，`calculateTokenCost` 回退 **live `spus`**，属预期（见 `api_proxy` 日志 `pricing_source`）。

---

## 5. 发布与回滚注意

- 价目表（`spus` / `pricing_version_spu_rates`）变更会影响**新产生的扣费**；已履约订单绑定的 `pricing_version_id` 不变。
- 迁移（如 046 合并算力点）后应执行一次 **第 2 节** 抽样对账。

---

## 6. 代码辅助

后端提供 `services.ReconcileUserUsage` / `services.UsageReconcileOK`（见 `services/usage_reconcile.go`），对比 **用量日志计费 Token 量**（`COALESCE(token_usage, input+output)`）与 **usage 流水扣减（Token）**；可在集成环境或小流量抽样调用，与 **第 2 节** SQL 等价（勿再与 `cost` 元混比）。

---

## 7. 管理端 API 与 CLI 任务

**HTTP（需管理员 JWT）**

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/admin/reconciliation/ledger` | 全库计费 Token（`api_usage_logs`）与 `token_transactions` usage 扣减对比，响应含 `unit: tokens` |
| GET | `/api/v1/admin/reconciliation/ledger/drift?page=&page_size=` | 按用户列出差值（可能较慢） |
| GET | `/api/v1/admin/reconciliation/ledger/drift/export` | 差异用户 CSV（UTF-8 BOM，最多约 5 万行） |
| POST | `/api/v1/admin/reconciliation/ledger/check` | 与 GET ledger 相同计算并写审计日志，供定时任务触发 |
| GET | `/api/v1/admin/reconciliation/gmv?start_date=&end_date=` | 订单 GMV（CNY，`paid`/`completed`），日期可选 |

**CLI（cron / 部署后巡检）**

```bash
cd backend && DATABASE_URL="postgresql://..." go run ./cmd/reconcile
```

退出码：`0` 表示全库用量两侧合计一致；`1` 表示不一致或连接失败。可与 `Makefile` 中 `reconcile-check` 目标配合使用。
