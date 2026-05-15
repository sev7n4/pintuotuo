# Anthropic 兼容出站（`XX_anthropic`）最小配置检查清单与配置指导

> **适用版本**：已合并并部署  
> - PR #513：`merchant_skus.anthropic_api_key_id`、strict 权益合并、Claude 入口出站改写  
> - PR #514：`api_format=anthropic` / `*_anthropic` 的验证与健康探测（Messages API）  
>
> **读者**：平台运营（Admin）、商户运营、研发排障  
> **约定**：下文用 `XX` 表示目录**主厂商** code（如 `alibaba`、`zhipu`、`moonshot`），影子厂商 code 固定为 **`XX_anthropic`**（后缀 `_anthropic` 为产品级约定，非某一家云厂商硬编码）。

---

## 1. 架构一句话

| 角色 | 厂商 code | 用途 |
|------|-----------|------|
| 目录 / SPU / 计价 / strict 主权益 | `XX` | OpenAI 兼容或平台统一目录模型前缀 `XX/模型名` |
| Claude Code / Anthropic Messages 出站 | `XX_anthropic` | `api_format=anthropic`，`POST …/messages`，鉴权 `x-api-key` |
| 商户密钥 | 两条 `merchant_api_keys` | 主 Key：`provider=XX`；副 Key：`provider=XX_anthropic` |
| 商户 SKU | 一行 `merchant_skus` | `api_key_id`（主）+ 可选 `anthropic_api_key_id`（副，全局唯一活跃绑定） |

用户请求模型仍为 **`XX/某模型`**；系统在存在 active 的 `XX_anthropic` 时，将 Anthropic 入口流量出站改写到影子厂商，并从 SKU 副槽选 Key。

---

## 2. 接入前条件（必须满足）

- [ ] 上游已提供 **Anthropic Messages API 兼容** 端点（能 `POST {base}/messages`，接受 `x-api-key` + `anthropic-version`）。
- [ ] 已知兼容根 URL（按各云文档：**阿里云百炼 Anthropic 兼容**见下文 §3.3，勿与 OpenAI 的 `compatible-mode` 混用）。
- [ ] 已知至少一个可用于**深度验证探测**的模型 ID（该账号下真实可调用）。
- [ ] 平台已部署迁移 **099**（`merchant_skus.anthropic_api_key_id` 列存在）。
- [ ] 目录侧已有 SPU/SKU，`spus.model_provider = XX`（**不要**把 SPU 主厂商设为 `XX_anthropic`）。

---

## 3. Admin：厂商配置（`model_providers`）

### 3.1 主厂商 `XX`（通常已存在）

| 检查项 | 要求 |
|--------|------|
| `code` | `XX`（小写推荐，与目录一致） |
| `api_format` | 一般为 `openai`（与现有 OpenAI 兼容代理一致） |
| `api_base_url` | 主线路 OpenAI 兼容 base（供 `XX` 密钥与 OpenAI 路径使用） |
| `status` | `active` |

### 3.2 影子厂商 `XX_anthropic`（新建）

在 **Admin → 模型厂商配置** 新增一条：

| 字段 | 要求 | 说明 |
|------|------|------|
| `code` | **`XX_anthropic`** | 必须等于 `{主厂商}_anthropic`，与代码 `AnthropicSiblingProviderCode(XX)` 一致 |
| `name` | 可读名称 | 如「XX（Anthropic 出站）」 |
| `api_format` | **`anthropic`** | 触发 Messages 探测与 x-api-key 鉴权 |
| `api_base_url` | 上游 Anthropic 兼容根路径（**按各厂商文档**） | **不要**手工拼 `/messages`（系统按 base 是否以 `/v1` 结尾拼接 `…/v1/messages 或 …/messages`）。阿里云见 §3.3。 |
| `status` | **`active`** | 非 active 时代理不会改写出站 |
| `litellm_model_template` 等 | 若 Key 走 LiteLLM | 按该 upstream 在 LiteLLM 中的命名填写（见 §7） |

### 3.3 阿里云百炼（`alibaba_anthropic`）基址

与 **OpenAI 兼容** `…/compatible-mode/v1` **不是同一路径**。Anthropic Messages 兼容基址应为：

`https://dashscope.aliyuncs.com/apps/anthropic`

（国际站等见阿里云「Anthropic API 兼容」文档中的区域域名。）

**严禁**使用 **`compatible-model`**（如 `…/compatible-model/v1`）：该路径对上游**无效**，会得到 **HTTP 404**，轻量/深度验证与立即探测均失败；界面可能将 **404** 归入「模型不存在」类提示。

**常见错误**

- `code` 写成 `anthropic_XX` 或 `XX-anthropic` → SKU 校验与代理改写会失败。
- `api_format` 仍为 `openai` → 验证会走 `GET /models`，极易失败。
- 阿里云误填 **`compatible-model`**，或与主线路混用 **`compatible-mode` 去拼 Messages** → **404**；应使用 **`/apps/anthropic`**。

---

## 4. 目录与商品（SPU/SKU）

| 检查项 | 要求 |
|--------|------|
| SPU `model_provider` | **`XX`**（主厂商） |
| 平台 SKU / 商户上架 SKU | 关联上述 SPU；商户 `merchant_skus` 绑主 Key |
| 模型名 | 对外仍为 `XX/模型名`；无需在目录增加 `XX_anthropic/…` 作为售卖模型 |

---

## 5. 商户：API Key

每个需要 Claude 出站的**商户 SKU**，建议准备 **2 条** active 密钥：

| 顺序 | `merchant_api_keys.provider` | 用途 |
|------|------------------------------|------|
| 1 | `XX` | 主 Key，绑 `merchant_skus.api_key_id` |
| 2 | `XX_anthropic` | 副 Key，绑 `merchant_skus.anthropic_api_key_id` |

**注意**

- 副 Key 与主 Key **不是**同一条记录的两个字段；须**各建一行**。
- 同一 `anthropic_api_key_id` 不能绑到多条 **active** `merchant_skus`（库表唯一索引）。
- 路由（`route_mode` / `route_config` / `region`）若与主 Key 不同，验证与探测会按**各 Key 自己的 SSOT** 解析；建议与主 Key 对齐，除非刻意分流。

---

## 6. 验证与健康（商户端与 Admin BYOK 共用后端）

以下操作在**商户端**与 **Admin BYOK** 走同一套 `APIKeyValidator` / `HealthChecker`，仅 `verification_type` 标签不同（`merchant_*` / `admin_*`）。

### 6.1 推荐顺序（副 Key `XX_anthropic`）

**与主 Key（`XX`）的差异（重要）**

| 项 | `alibaba`（OpenAI 兼容） | `alibaba_anthropic`（Anthropic Messages） |
|----|--------------------------|----------------------------------------|
| 基址 | `…/compatible-mode/v1` | `…/apps/anthropic`（**不是** compatible-mode） |
| 拉模型列表 | 上游 `GET …/models`，可返回上百个 id | 上游**无** `GET /models`（会 404） |
| 平台探测 | 直接解析上游列表 | 先 `POST …/messages` 验连通；模型列表在部署 **PR #514+ 增强** 后，会再用**同一密钥**对主线路 `GET …/compatible-mode/v1/models` 填充下拉（与 pkey 列表一致） |
| 深度验证探测模型 | 任选帐号内 chat 模型 | 须为 **Anthropic 线路文档支持的 model**（如 `qwen-plus`）；目录里的 `glm-5.1` 等未必能在 Messages 线路上调用 |

1. **深度验证**（必选，若需 `verified` + strict 放行）  
   - 在 UI 选择或填写 **探测模型**（`probe_model`）：须为 **Anthropic 兼容线路上真实可调用的** model id（阿里云见官方「支持的模型」表，常为 qwen 系列）。  
   - 勿默认选仅在 OpenAI `compatible-mode` 列表里出现、但 Anthropic 线不支持的 id。
2. **立即探测**（可选，刷新 `health_status`）  
   - 走 `LightweightPing` 或 `FullVerification`（由 Key 的 `health_check_level` 决定）。
3. 确认库表状态（见 §8 SQL）。

### 6.2 strict 放行相关字段（副 Key）

| 字段 | 期望 |
|------|------|
| `verification_result` | `verified`（通常需**深度验证**成功） |
| `health_status` | `healthy` 或 `degraded` |
| `status` | `active` |

轻量验证失败时，对非 deep 类型可能仍记为「成功」但**不会**把 Key 标为 `verified`；以深度验证结果为准。

### 6.3 主 Key `XX`

主 Key 仍按 OpenAI 兼容路径验证（`GET /models` 等）；与副 Key 独立。两条 Key 都建议达到上述健康状态后再对外。

---

## 7. LiteLLM 路由（可选）

仅当 `merchant_api_keys.route_mode = litellm` 时需要额外关注：

| 检查项 | 说明 |
|--------|------|
| `LLM_GATEWAY_LITELLM_URL` / `LITELLM_MASTER_KEY` | 运行环境已配置 |
| `XX_anthropic` 的 `litellm_model_template` | Admin 厂商表填写正确模板 |
| 验证探测 | 走 LiteLLM 时，连接探测会对上游 direct base 做解析；与 [BYOK 路由 SSOT](./byok-routing-ssot.md) 一致 |

若仅 **direct** 出站，可跳过本节。

---

## 8. 商户 SKU：绑定副 Key

在 **商户 SKU 管理**（上架/编辑）：

| 检查项 | 要求 |
|--------|------|
| SPU 主厂商 | `XX` |
| 主 `api_key_id` | `provider = XX` 的 Key |
| `anthropic_api_key_id` | `provider = XX_anthropic` 的 Key（可留空则 Claude 出站不合并副槽） |
| 副 Key 商户 | 与主 Key **同一 `merchant_id`** |
| 副 Key 验证 | 已 `verified`（strict 下） |

保存后，strict 白名单会合并主槽与副槽；出站 `XX_anthropic` 时自动选副 Key（PR #513 方案 A：按出站 provider 过滤）。

---

## 9. 端到端验收（Claude / Anthropic 入口）

- [ ] 使用平台 **Anthropic 兼容入口**（如 `/api/v1/anthropic/v1/messages`）发起请求。
- [ ] 请求体 `model` 为目录模型（如 `XX/glm-4.7`），计费/权益仍按 **`XX`**。
- [ ] 上游实际收到的是 **`XX_anthropic` base + /messages**，且使用副 Key 鉴权。
- [ ] 日志/账单中 `provider` 展示与计价策略符合预期（目录主厂商 `XX`）。

---

## 10. 最小配置检查清单（可打印）

复制下表，将 `XX` 替换为实际主厂商 code。

### Admin / 平台

- [ ] `model_providers.code = XX`，`status = active`
- [ ] `model_providers.code = XX_anthropic`，`api_format = anthropic`，`api_base_url` 正确，`status = active`
- [ ] SPU `model_provider = XX`（非 `XX_anthropic`）
- [ ] 迁移 099 已在目标环境执行

### 商户密钥

- [ ] 已创建 `provider = XX` 的主 Key，深度验证通过，`verified` + 健康正常
- [ ] 已创建 `provider = XX_anthropic` 的副 Key，**深度验证**时指定有效 `probe_model`，`verified` + 健康正常
- [ ] 副 Key 未绑定到其他 active `merchant_skus`

### 商户 SKU

- [ ] `merchant_skus.api_key_id` → 主 Key
- [ ] `merchant_skus.anthropic_api_key_id` → 副 Key（需要 Claude 出站时必填）
- [ ] 同一 SKU 行主/副 Key 同属一商户

### 运行时

- [ ] Claude 入口试调用成功
- [ ] strict 模式下无 403（权益/验证/健康）

---

## 11. 排障速查

| 现象 | 优先检查 |
|------|----------|
| 验证/探测一直失败 | `XX_anthropic` 的 `api_format` 是否为 `anthropic`；`api_base_url` 是否可达；深度验证是否指定 `probe_model` |
| 验证成功但 strict 仍 403 | `verification_result` 是否为 `verified`；`health_status` 是否 `healthy`/`degraded`；SKU 是否绑副 Key |
| SKU 保存报 provider 不匹配 | 副 Key 的 `provider` 是否恰好为 `XX_anthropic`（与 SPU 主厂商推导一致） |
| Claude 请求仍走主 Key / 403 | DB 是否存在 **active** 的 `XX_anthropic`；`anthropic_api_key_id` 是否已绑；出站是否命中 Anthropic 入口 |
| 副 Key 绑不上第二条 SKU | 设计如此：一个 anthropic Key 全局仅允许一条 active `merchant_skus` |
| Admin「拉取模型列表」为空 | 非 openai 格式时，Admin probe-models 走 `FullVerification`；anthropic 格式返回预置/探测结果，可改用手动 `probe_model` + 深度验证 |

### 11.1 建议 SQL（生产只读排查）

将 `:xx`、`:merchant_id`、`:key_id` 替换为实际值。

```sql
-- 厂商行
SELECT code, api_format, api_base_url, status
FROM model_providers
WHERE lower(code) IN (:xx, :xx || '_anthropic');

-- 商户密钥状态
SELECT id, provider, verification_result, health_status, status, route_mode
FROM merchant_api_keys
WHERE merchant_id = :merchant_id
  AND lower(provider) IN (:xx, :xx || '_anthropic');

-- SKU 绑定
SELECT ms.id, ms.status, ms.api_key_id, ms.anthropic_api_key_id,
       sp.model_provider
FROM merchant_skus ms
JOIN skus s ON s.id = ms.sku_id
JOIN spus sp ON sp.id = s.spu_id
WHERE ms.merchant_id = :merchant_id AND ms.status = 'active';
```

---

## 12. 示例：`XX = alibaba`（参考，非模板）

| 项 | 值 |
|----|-----|
| 主厂商 | `alibaba`，OpenAI 兼容 base（如 DashScope compatible-mode） |
| 影子厂商 | `alibaba_anthropic`，`api_format=anthropic`，base `https://dashscope.aliyuncs.com/apps/anthropic`（勿用 `compatible-model`） |
| 探测模型 | 深度验证可填账号内可用模型（如 `qwen-plus` 或实际 glm 型号） |
| 商户 | 5 条主 Key + 5 条 `alibaba_anthropic` 副 Key，各 SKU 各绑一对 |

---

## 13. 与代码的对应关系（供研发）

| 能力 | 入口 |
|------|------|
| 影子 code 规则 | `services.AnthropicSiblingProviderCode`、后缀 `AnthropicSiblingProviderSuffix` |
| Anthropic 探测 | `services.ProbeProviderConnectivity`、`ProviderUsesAnthropicHTTP` |
| 验证 | `services.APIKeyValidator.performVerificationWithRouteMode` |
| 健康 / 立即探测 | `services.HealthChecker`、`HealthScheduler.TriggerImmediateCheck` |
| SKU 校验 | `handlers/merchant_sku.go`（副 Key provider 须为 `{model_provider}_anthropic`） |
| 代理改写 | `handlers/anthropic_compat.go`、`api_proxy` strict 合并副槽 |

---

## 14. 相关文档

- [BYOK 路由 SSOT](./byok-routing-ssot.md)
- [DEVELOPMENT.md](../../DEVELOPMENT.md) / [DEPLOYMENT.md](../../DEPLOYMENT.md)
- 开发者 Anthropic 入口说明：`frontend/public/docs/developer/` 下 IDE 文档

---

**文档维护**：平台约定或迁移变更时，请同步更新本节与 Admin 模型厂商页上的 `_anthropic` 提示文案。
