# 端点覆盖矩阵：Admin 能力探测 vs CLI vs 共享实现

**目的**：作为单一事实来源（SSOT），说明「OpenAI 兼容矩阵里有哪些探测项、谁在什么入口会真实打上游、默认是否计费」，避免把 **Admin 弹窗里可勾选的端点** 误解为 **全量非 chat**。

**实现关系**（必读）：

- 探测逻辑在 Go 包 [`backend/capabilityprobe/`](../../backend/capabilityprobe/)，**Admin HTTP API** 与 **`cmd/capability-probe` CLI** 共用同一套代码路径；Admin **不会** `exec` CLI 二进制。
- Admin 单 Key：`POST /api/v1/admin/byok-routing/:id/capability-probe`（可选 `skip_embeddings`、`probes`、各模型字段、**`billable`** 等）；模型列表辅助：`GET /api/v1/admin/byok-routing/:id/probe-models`（与 `FullVerification` 同源 `GET …/models`）。
- CLI：[`backend/cmd/capability-probe/`](../../backend/cmd/capability-probe/)，适合批量、长耗时、可选 `-billable` 的全矩阵证据。

---

## 1. 总表（`api_format = openai`）

下列为 `runOpenAIFormatProbes` / `ProbeScannedKey` 相关行为在 **Admin 默认（`billable=false`）**、**Admin `billable=true`** 与 **CLI** 下的对照。`api_format ≠ openai` 时，各 `post_*` 仅写 `skipped` 行，不发上游 POST（见 [phase0-scope.md](./phase0-scope.md)）。

| CSV / `rows` 中 `probe` 前缀 | 上游行为概要 | Admin 默认 `billable=false` | Admin `billable=true` | CLI 无 `-billable` | CLI 带 `-billable` | 备注 |
|------------------------------|----------------|------------------------------|------------------------|-------------------|---------------------|------|
| `get_models` | `GET …/models` | **是** | **是** | **是** | **是** | 与 `HealthChecker.FullVerification` 一致 |
| `post_embeddings` | JSON POST | **是**（若未跳过且选中） | **是**（同上） | **是** | **是** | Admin 可 `skip_embeddings` 或从 `probes` 排除 |
| `post_moderations` | JSON POST | **是**（若选中） | **是** | **是** | **是** | |
| `post_responses` | JSON POST | **是**（若选中） | **是** | **是** | **是** | |
| `post_chat_completions` | JSON POST | **否**（skipped） | **是**（极小 `max_tokens`） | **否** | **是** | Admin 计费类需显式 `billable` + 前端二次确认 |
| `post_images_generations` | JSON POST | **否** | **是** | **否** | **是** | 极小尺寸 / `n=1` |
| `post_images_variations` | multipart POST | **否** | **是** | **否** | **是** | |
| `post_images_edits` | multipart POST | **否** | **是** | **否** | **是** | |
| `post_audio_transcriptions` | multipart POST | **否** | **是** | **否** | **是** | 内置短静音 WAV |
| `post_audio_translations` | multipart POST | **否** | **是** | **否** | **是** | |
| `post_audio_speech` | JSON POST | **否** | **是** | **否** | **是** | |

**Admin UI**：对 **embeddings / moderations / responses** 提供单选/多选；**「计费类探测」** 开关对应请求体 **`billable: true`**（仅管理员；服务端写审计日志）。其余计费端点无单独勾选，与 CLI `-billable` 一致「一次跑齐」。

**小结**：

- **「非 chat」在 OpenAI 文档语义上**包含 embeddings、moderations、responses、images、audio 等；**本表**按「探测实现里实际出现的 `probe` 行」枚举。
- **默认 Admin** 仍控制非 chat 三项；**开启计费类** 后与 **CLI `-billable`** 在同一路径上对 chat/图/音发起**轻量**真实 POST（仍见 [phase0-scope.md](./phase0-scope.md) §2～3 用量说明）。

---

## 2. 与 Phase B/C 的衔接（规划占位）

| 阶段 | 建议动作 |
|------|----------|
| **Phase B** | **已定**：允许在 Admin 内通过 **`billable: true`** 显式开启计费类轻量探测（与深度验证类似的产品决策）；服务端记录 `admin_byok_routing` 日志。 |
| **Phase C（轻量）** | Admin「Phase0 能力探测」弹窗已提供 **复制 CLI 命令**（`docker exec … capability-probe -api-key-id …`）；Runbook 仍可补充环境差异说明。 |
| **Phase C（重量）** | 可选：计费类 **按端点子集** 勾选、**异步任务**、独立审计表；若落地须再更新本表「Admin billable」列。 |

---

## 3. 相关链接

- 用量与计费默认值：[phase0-scope.md](./phase0-scope.md)
- 总索引与命令示例：[README.md](./README.md)
- 路由与 ExecutionLayer：[byok-routing-ssot.md](./byok-routing-ssot.md)
