# 用户余额与扣费口径（内部说明）

**最后更新**：2026-04-10

GLM Coding Plan 档位参考价（若本地有副本）：`docs/glm_coding_plan_internal_economics.csv`（仓库 `docs/` 已 gitignore，与平台内部账本单位无直接等价关系）。

---

## 一、已拍板目标架构

### 1. 单一账本（single ledger）

- 用户通过零售订单获得的**可消费额度**应落在**同一条主账本**上；OpenAI 兼容代理路径只消费这一条。
- 现状存在 `tokens` 与 `compute_point_accounts` 双轨时，以本文件为原则**收敛为一条**（含历史迁移或废弃策略），避免「充 A 扣 B」。

### 2. 订单快照与入账 1:1

- 履约入账以 **`orders` 表（及订单行）上的快照字段**为准：`sku_type`、`token_amount`、`compute_points` 等；快照缺失时再回退当前 `skus` 行。
- **规则**：订单里写的是多少（例如 `1_000_000`），履约后**余额就加多少**，**不在入账时**再乘「元/K token」。
- 目标：与未来将引入的 **`pricing_version_id`** 一致——**购买时点**的权益数量与价目绑定，避免事后改 SKU 导致买用不一致。

### 3. 换算层（按次扣费）

- 账本主单位：**内部可加减单位**（与订单快照口径一致），**不以人民币为主余额**。
- 代理每次调用：`input/output tokens` → 根据**该调用所解析出的价目版本**中的 **元 / 1K tokens**（或等价）计算 **本次应扣内部单位** → 从**同一账本**扣减。
- 人民币收入、成本、GMV：在**报表层**由内部单位或独立流水折算，与原子扣费逻辑分离。

### 4. 零售 vs 成本结算

- **用户侧**：`tokens`（收敛后唯一主账本）与 `PricingService` / `calculateTokenCost` 产出的扣减量须**同一语义**（内部单位）。
- **成本 / 商户结算**：仍以 **`api_usage_logs`**、合同价、商户 Key 等为准；**不要**用 `skus.token_amount` 或用户余额反推上游按量成本。

---

## 二、现状与目标差距（实现时需覆盖）

| 现状 | 目标 |
|------|------|
| ~~`compute_points` 曾入账 `compute_point_accounts`~~（已改为 `tokens`，046 合并历史） | 统一入账到**单一账本** |
| ~~`token_pack` SKU 曾必填 `compute_points`~~（已改为可选/展示；履约只认 `token_amount`） | 配置与履约一致 |
| `api_proxy` 扣 `tokens.balance`；未消费算力点账户 | 扣费路径只认**收敛后的主账本**；换算用**版本化单价** |
| 定价多为实时读库 / 缓存 | 扣费解析优先：**订单或权益 → `pricing_version_id` → 单价** |

---

## 三、相关代码（落点）

- 扣费：`handlers/api_proxy.go`（`calculateTokenCost`：优先 **最近履约订单** 的 `pricing_version_id` → `pricing_version_spu_rates`；无绑定或快照无该模型时回退 **live `spus`**；再扣 `tokens`）
- 履约：`services/fulfillment_service.go`（`fulfillTokenPack`、`fulfillComputePoints`）
- 定价：`services/pricing_service.go`
- 结算：`services/settlement_service.go`（商户侧，与用户余额隔离）
- 价目版本表：`migrations/045_pricing_versions.sql`（`pricing_versions`、`pricing_version_spu_rates`；`orders.pricing_version_id` 可空）
- 下单绑定：`handlers/order_and_group.go` 在创建订单时写入 **baseline** 价目版本（`services.BaselinePricingVersionID`）；库中无 baseline 时保持 `NULL`（兼容未跑迁移的环境）
- **IE-2 零售单一账本**：`fulfillComputePoints` 入账 `tokens` + `token_transactions`；`migrations/046_merge_compute_points_to_tokens.sql` 合并历史 `compute_point_accounts`；算力点余额/流水 API（`GetComputePointBalance` / `GetComputePointTransactions`）与 Token 同源。

---

## 四、迁移与风险（提醒）

- 合并账本前需：**数据迁移脚本、对账规则、灰度或只读期**。
- 价目版本上线后需：**老订单无 version 时的默认策略**（显式规则，避免静默按新价扣费）。
