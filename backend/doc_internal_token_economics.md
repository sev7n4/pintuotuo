# 用户余额与扣费口径（内部说明）

## 用户侧（零售）

- `tokens.balance`：用户可消费的额度，与 `PricingService` / 代理里 `calculateTokenCost` 产出的 **`cost` 必须使用同一单位与语义**（例如统一为平台内部计价单位，与「元/1K tokens」类定价一致）。
- 订单履约：`FulfillmentService` 入账时优先使用 **`orders` 表快照**（`sku_type`、`token_amount`、`compute_points`），在快照缺失时回退到当前 `skus` 行，避免 SKU 事后变更与下单时不一致。
- `token_transactions`：`purchase`（入账）与 `usage`（调用扣费）应对账一致。

## 成本 / 结算侧（与零售隔离）

- 商户结算、上游成本核对应以 **`api_usage_logs`**（及合同价、商户 Key 等）为准。
- **不要**用 `skus.token_amount` 或用户 `tokens.balance` 推导上游按量成本；若代码中出现此类混用，应改为 usage 链路。

## 相关代码

- 扣费：`handlers/api_proxy.go`
- 履约：`services/fulfillment_service.go`
- 结算：`services/settlement_service.go`
