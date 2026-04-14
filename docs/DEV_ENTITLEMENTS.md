# 权益白名单与价目快照（开发说明）

## 环境变量


| 变量                        | 值         | 说明                                                                  |
| ------------------------- | --------- | ------------------------------------------------------------------- |
| `ENTITLEMENT_ENFORCEMENT` | `off`（默认） | 不强制权益白名单；`api_proxy` 扣费保持与改造前兼容的回退行为                                |
| `ENTITLEMENT_ENFORCEMENT` | `strict`  | 预扣前校验权益；仅使用命中权益的 `pricing_version_id` 快照计价，**禁止**无权益时按外网 live 价静默扣费 |


测试可使用 `services.SetEntitlementEnforcementForTest` 覆盖环境变量（见 `entitlement_test.go`）。

## 数据表与字段

- `**orders`**：`pricing_version_id` — 已履约订单的价目锚点（下单/履约时写入）。
- `**user_subscriptions**`：
  - `pricing_version_id` — 订阅权益的价目锚点；履约与自动续费成功时更新。
  - `entitlement_anchor_at` — 权益锚点时间，与订单 `fulfilled_at` 一起用于「多条命中时取最近」排序。
- `**pricing_versions` / `pricing_version_spu_rates**`：按 `spu_id` 存快照输入/输出单价。

迁移：`backend/migrations/050_user_subscription_entitlement_pricing.sql`（含回填说明）。

## 代码入口

- 权益解析：`backend/services/entitlement.go` — `ResolveChosenPricingVersion`。
- 快照计价：`backend/services/pricing_version.go` — `CalculateCostFromPricingVersion`（`provider` + `provider_model_id` / `model_name` 匹配）。
- 代理网关：`backend/handlers/api_proxy.go` — strict 下 PreDeduct 前校验；`calculateTokenCost` 传入选中的 version ID。
- 续费写锚点：`backend/services/subscription_renewal.go`。
- 履约写订阅：`backend/services/fulfillment_service.go`。

## 本地造数验证（简要）

1. 准备含目标 SPU 的 SKU；确保 `spus.model_provider`、`provider_model_id` 或 `model_name` 与代理请求一致。
2. 插入已支付且已履约订单：`orders.status` 已支付、`fulfilled_at` 非空、`pricing_version_id` 指向含该 SPU 费率的版本；`pricing_version_spu_rates` 中有对应 `spu_id` 行。
3. 或插入有效订阅：`user_subscriptions.status = active`、`end_date` 未过期，且 `pricing_version_id` 已填。
4. 设置 `ENTITLEMENT_ENFORCEMENT=strict`，调用代理相同 `provider`/`model`，应 200 且费用来自快照；去掉权益应 403 `ENTITLEMENT_DENIED`。

更多经济字段说明见 [backend/doc_internal_token_economics.md](../backend/doc_internal_token_economics.md)。