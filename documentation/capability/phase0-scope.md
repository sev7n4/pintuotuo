# Phase 0 范围与「用量」定义（能力探测）

**版本**：1.0  
**日期**：2026-05-12  

本文约定：**谁在什么环境、跑什么探测、产生多少上游调用、默认是否计费**，避免 Phase 0 与后续 Phase 1～3 开发范围纠缠不清。

## 1. Phase 0 目标（只做能力矩阵证据，不做产品改造）

| 属于 Phase 0 | 不属于 Phase 0 |
|--------------|------------------|
| 产出/更新「厂商 × 端点」证据（CSV、矩阵单元格、Runbook） | 修改 `proxyAPIRequestCore` / 权益 / 计费生产逻辑 |
| 在**部署镜像**内提供 `capability-probe` 可执行文件 | 对 C 端用户默认开启新的非 chat 能力 |
| 运维按需（手动工作流或 `docker exec`）跑默认（无 `-billable`）探测 | 全量 `-billable` 探测作为 CI 必过门禁（默认否） |

## 2. `capability-probe` 探测与「用量」

对每条符合条件的 **`merchant_api_keys`**（与 `model_providers` join，活跃且已验证等，见程序内 SQL）：

### 2.1 每条密钥固定会发生的调用

| 探测名 `probe` | 上游 HTTP | 默认是否计费 | 说明 |
|----------------|-------------|----------------|------|
| `get_models` | `GET …/models`（与 `HealthChecker.FullVerification` 完全一致） | 通常否 | 所有 `api_format` 均执行 |

### 2.2 `api_format = openai` 时，每条密钥额外调用

| `probe` | 条件 | 默认是否计费 | 说明 |
|---------|------|----------------|------|
| `post_chat_completions` | 仅当 `-billable` | **是**（token） | 极小 `max_tokens=1` |
| `post_embeddings` | 除非 `-skip-embeddings` | **是**（按 token/次） | 极小 `input` |
| `post_moderations` | 总是 | 视厂商而定，通常低 | 极小 `input` |
| `post_images_generations` | 仅 `-billable` | **是**（按张） | `256x256`、`n=1` |
| `post_images_variations` | 仅 `-billable` | **是** | multipart，内置 1×1 PNG，`dall-e-2` |
| `post_images_edits` | 仅 `-billable` | **是** | multipart，`dall-e-2` |
| `post_audio_transcriptions` | 仅 `-billable` | **是** | 内置约 200ms 静音 WAV，`whisper-1` |
| `post_audio_translations` | 仅 `-billable` | **是** | 同上 |
| `post_audio_speech` | 仅 `-billable` | **是** | 极短文本 `tts-1` |
| `post_responses` | 总是 | 视厂商而定 | `max_output_tokens=1`，可能 4xx 仍记为有效证据 |

### 2.3 `api_format ≠ openai` 时

对上述 `post_*` 各写一行 **`ok=skipped`**，`note=api_format_not_openai_…`；**不**发上游 POST（用量为 0）。

## 3. 推荐默认（控制成本）

| 场景 | 建议命令片段 | 预期每条密钥上游 POST 次数（openai） |
|------|----------------|--------------------------------------|
| **手动工作流 / 服务器巡检** | 不加 `-billable` | 约 **3**（embeddings + moderations + responses；若 `-skip-embeddings` 则 2） |
| **一次性全量计费探测** | `-billable` 且限制 `-limit`、或 `-api-key-id` 单条 | 最多 **10** 次 POST + 1 次 GET models |

**网关 `litellm`**：与 `health_checker` 相同，若存在 `LITELLM_MASTER_KEY`，对网关请求使用该 Bearer；**直连**仍用解密后的商户 Key。

## 4. 与 CI / 部署工作流的关系

- **`.github/workflows/deploy-tencent.yml`**：默认部署**不再**自动跑重 `capability-probe`（避免长时间阻塞合并部署）。
- **`.github/workflows/capability-probe-tencent.yml`**：`workflow_dispatch`，与原先 deploy 内逻辑等价（SSH 到腾讯云、`docker exec`、日志中 `tail` CSV 片段）；用于发版后或巡检时按需触发。
- **全量计费探测**不在 CI 中默认开启；由运维在服务器上手工执行并保存 CSV。

## 5. Phase 0 完成判据（DoD 补充）

- [ ] 生产镜像中存在 `/app/capability-probe` 且 `docker exec … test -x` 通过。
- [ ] 至少一次 **capability-probe-tencent 工作流**、**deploy 之外的手工**或手工归档的 CSV 中，对**每个活跃 BYOK provider** 有 `get_models` 行；对承诺的 OpenAI 系 provider 有 `post_*` 行（含 `skipped` 或 HTTP 码）。
- [ ] 矩阵模板中 P1 格子与 CSV 证据已互链（路径或 PR 附件）。
