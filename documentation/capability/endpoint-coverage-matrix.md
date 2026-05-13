# 端点覆盖矩阵：Admin 能力探测 vs CLI vs 共享实现

**目的**：作为单一事实来源（SSOT），说明「OpenAI 兼容矩阵里有哪些探测项、谁在什么入口会真实打上游、默认是否计费」，避免把 **Admin 弹窗里可勾选的端点** 误解为 **全量非 chat**。

**实现关系**（必读）：

- 探测逻辑在 Go 包 [`backend/capabilityprobe/`](../../backend/capabilityprobe/)，**Admin HTTP API** 与 **`cmd/capability-probe` CLI** 共用同一套代码路径；Admin **不会** `exec` CLI 二进制。
- Admin 单 Key：`POST /api/v1/admin/byok-routing/:id/capability-probe`（可选 `probes` / 模型字段等）；模型列表辅助：`GET /api/v1/admin/byok-routing/:id/probe-models`（与 `FullVerification` 同源 `GET …/models`）。
- CLI：[`backend/cmd/capability-probe/`](../../backend/cmd/capability-probe/)，适合批量、长耗时、可选 `-billable` 的全矩阵证据。

---

## 1. 总表（`api_format = openai`）

下列为 `runOpenAIFormatProbes` / `ProbeScannedKey` 相关行为在 **默认 Admin（`Billable=false`）** 与 **CLI 默认（无 `-billable`）** 下的对照。`api_format ≠ openai` 时，各 `post_*` 仅写 `skipped` 行，不发上游 POST（见 [phase0-scope.md](./phase0-scope.md)）。

| CSV / `rows` 中 `probe` 前缀 | 上游行为概要 | Admin：是否**真实**请求 | Admin：UI 是否可勾选 | CLI 无 `-billable` | CLI 带 `-billable` | 备注 |
|------------------------------|----------------|-------------------------|----------------------|-------------------|---------------------|------|
| `get_models` | `GET …/models` | **是**（每次探测） | 否（固定执行） | **是** | **是** | 与 `HealthChecker.FullVerification` 一致 |
| `post_embeddings` | JSON POST | **是**（若未跳过且选中） | **是** | **是** | **是** | Admin 可 `skip_embeddings` 或从 `probes` 排除 |
| `post_moderations` | JSON POST | **是**（若选中） | **是** | **是** | **是** | |
| `post_responses` | JSON POST | **是**（若选中） | **是** | **是** | **是** | 模型字段见 Admin / `ProbeFlags` |
| `post_chat_completions` | JSON POST | **否**（写 skipped） | 否 | **否** | **是** | Admin 固定非计费；note 提示需 `-billable` |
| `post_images_generations` | JSON POST | **否**（skipped） | 否 | **否** | **是** | 计费类 |
| `post_images_variations` | multipart POST | **否**（skipped） | 否 | **否** | **是** | |
| `post_images_edits` | multipart POST | **否**（skipped） | 否 | **否** | **是** | |
| `post_audio_transcriptions` | multipart POST | **否**（skipped） | 否 | **否** | **是** | |
| `post_audio_translations` | multipart POST | **否**（skipped） | 否 | **否** | **是** | |
| `post_audio_speech` | JSON POST | **否**（skipped） | 否 | **否** | **是** | |

**小结**：

- **「非 chat」在 OpenAI 文档语义上**包含 embeddings、moderations、responses、images、audio 等；**本表**按「探测实现里实际出现的 `probe` 行」枚举。
- **Admin 当前产品范围**：仅对 **embeddings / moderations / responses** 提供勾选；其余非 chat（图、音）在 Admin 路径下**不会**发起真实 POST，结果表中为 `skipped`（与 `billable_disabled` 等 note），**全量覆盖需 CLI 加 `-billable`**（慎用，见 [phase0-scope.md](./phase0-scope.md) §2～3）。

---

## 2. 与 Phase B/C 的衔接（规划占位）

| 阶段 | 建议动作 |
|------|----------|
| **Phase B** | 产品/风控确认：是否允许在 Admin 内显式开启「计费类」探测；若否，CLI / 手工工作流保持为全矩阵主入口。 |
| **Phase C（轻量）** | Admin「Phase0 能力探测」弹窗已提供 **复制 CLI 命令**（`docker exec … capability-probe -api-key-id …`）；Runbook 仍可补充环境差异说明。 |
| **Phase C（重量）** | 若批准：Admin 增加 `billable` + 图/音勾选、异步任务、审计日志；与本表同步更新「Admin 是否真实请求」列。 |

---

## 3. 相关链接

- 用量与计费默认值：[phase0-scope.md](./phase0-scope.md)
- 总索引与命令示例：[README.md](./README.md)
- 路由与 ExecutionLayer：[byok-routing-ssot.md](./byok-routing-ssot.md)
