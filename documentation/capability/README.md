# OpenAI 兼容非 Chat 能力矩阵（Phase 0）

本目录承载 **Phase 0** 可开工交付物：行业调研摘要、能力矩阵模板、环境快照 Runbook、最小请求样例，以及上游探测说明。

> **路径说明**：仓库根 `.gitignore` 忽略了 `/docs/`，故本套材料放在 `documentation/capability/`，以便纳入版本管理。

## 端点覆盖（Admin vs CLI）

**单一事实来源**：[endpoint-coverage-matrix.md](./endpoint-coverage-matrix.md)（各 `probe` 在 Admin / CLI 下是否真实请求上游、与图/音等「非 chat」的关系）。

## 文件索引

| 文件 | 对应 Phase | 说明 |
|------|------------|------|
| [industry-research-brief.md](./industry-research-brief.md) | 0.1 | 行业共识、引用与对平台的含义 |
| [matrix-template.md](./matrix-template.md) | 0.2 | 厂商 × 端点矩阵模板与状态枚举 |
| [runbook-env-snapshot.md](./runbook-env-snapshot.md) | 0.3 | 脱敏导出 `model_providers` / 密钥元数据的步骤 |
| [minimal-request-snippets.md](./minimal-request-snippets.md) | 0.5 | 各端点最小 `curl` / JSON 样例（便于复现探测） |
| [phase0-scope.md](./phase0-scope.md) | 0.x | **用量与范围**：默认/计费探测次数、部署流水线行为 |
| [endpoint-coverage-matrix.md](./endpoint-coverage-matrix.md) | 0.x | **端点覆盖矩阵**：Admin vs CLI、是否真实 POST、与「全量非 chat」的关系 |
| [byok-routing-ssot.md](./byok-routing-ssot.md) | — | **BYOK 路由 SSOT**：`merchant_api_keys` 与 `model_providers` 回退、api_proxy 与 ExecutionLayer |
| [risk-register-template.md](./risk-register-template.md) | 0.6 | 风险登记占位（复制到 Wiki/Jira 亦可） |

## 上游探测命令（Go，从数据库读取 BYOK）

实现路径：共享包 [`backend/capabilityprobe/`](../../backend/capabilityprobe/)（与 `cmd/capability-probe`、Admin API 共用探测逻辑）；CLI 入口为 [`backend/cmd/capability-probe/`](../../backend/cmd/capability-probe/)（`main.go`）。

- **生产镜像**：与 `server` 一并构建，路径 **`/app/capability-probe`**（见 [backend/Dockerfile](../../backend/Dockerfile)）。
- **厂商 API Key** 仅从 **`merchant_api_keys.api_key_encrypted`** 解密；**不**使用操作员本机的 `OPENAI_API_KEY` 等。
- **全量端点**：与 `services.EndpointType*` 对齐——`get_models` + 各 `post_*`（非 `openai` 的 `post_*` 仅写 `skipped` 行）。其中 **计费类**（chat、图像、语音、转写）默认 **跳过**，需显式传 **`-billable`** 才会真实调用上游（见 [phase0-scope.md](./phase0-scope.md)）。
- **进程环境**：仍需 **`DATABASE_URL`**、**`ENCRYPTION_KEY`**（与 backend 容器一致）；`route_mode=litellm` 时与现网一致可使用 **`LITELLM_MASTER_KEY`**。

### 在部署机执行（示例）

```bash
ssh -i ~/.ssh/tencent_cloud_deploy root@119.29.173.89
cd /opt/pintuotuo
# 生产 compose 容器名为 pintuotuo-backend（见 docker-compose.prod.yml）
docker exec pintuotuo-backend /app/capability-probe -out /tmp/capability-probe-latest.csv -limit 30
# 可选：打开计费类探测（慎用，见 phase0-scope.md）
# docker exec pintuotuo-backend /app/capability-probe -out /tmp/cap-billable.csv -limit 3 -billable -api-key-id 123
```

本地：`make capability-probe`（`CAPABILITY_PROBE_FLAGS='-out /tmp/c.csv -provider openai -limit 5'`）。

**GitHub Actions**：默认 **不再** 在 `deploy-tencent.yml` 末尾跑重 probe（避免阻塞部署）。需要全量/重探测时，在 Actions 中手动运行 **Capability probe (Tencent production)**（`.github/workflows/capability-probe-tencent.yml`，`workflow_dispatch`，可选 `-limit` / `-skip-embeddings`），或在部署机 `docker exec`（见上文示例）。

**Admin（单 Key）**：行为与可选范围见 [endpoint-coverage-matrix.md](./endpoint-coverage-matrix.md)。接口：`POST /api/v1/admin/byok-routing/:id/capability-probe`（可选 `skip_embeddings`、`probes`、各模型字段、**`billable`** 等），`GET /api/v1/admin/byok-routing/:id/probe-models`；响应顶层 **`rows`** 与 CSV 列一致。用于上架前矩阵证据，勿对外泄露 `note` 中的 URL 片段。

探测输出勿提交仓库；`.gitignore`：`documentation/capability/capability-probe-output*.csv`。

## 与代码的对照

- 端点路径常量：[`backend/services/execution_layer.go`](../../backend/services/execution_layer.go) 中 `endpointPathSuffixes`。
- 模型列表探测（仅 GET models）：[`backend/services/provider_probe.go`](../../backend/services/provider_probe.go)。

## 下一步（需人工）

- **0.4 / 0.5**：在预发/生产按 Runbook 导出配置后运行脚本，将 CSV 摘要回填矩阵。
- **0.6**：召开范围签字会，冻结 P1 格子；将结论链接回 Epic / 本 README。
