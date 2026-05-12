# OpenAI 兼容非 Chat 能力矩阵（Phase 0）

本目录承载 **Phase 0** 可开工交付物：行业调研摘要、能力矩阵模板、环境快照 Runbook、最小请求样例，以及上游探测说明。

> **路径说明**：仓库根 `.gitignore` 忽略了 `/docs/`，故本套材料放在 `documentation/capability/`，以便纳入版本管理。

## 文件索引

| 文件 | 对应 Phase | 说明 |
|------|------------|------|
| [industry-research-brief.md](./industry-research-brief.md) | 0.1 | 行业共识、引用与对平台的含义 |
| [matrix-template.md](./matrix-template.md) | 0.2 | 厂商 × 端点矩阵模板与状态枚举 |
| [runbook-env-snapshot.md](./runbook-env-snapshot.md) | 0.3 | 脱敏导出 `model_providers` / 密钥元数据的步骤 |
| [minimal-request-snippets.md](./minimal-request-snippets.md) | 0.5 | 各端点最小 `curl` / JSON 样例（便于复现探测） |
| [risk-register-template.md](./risk-register-template.md) | 0.6 | 风险登记占位（复制到 Wiki/Jira 亦可） |

## 上游探测命令（Go，从数据库读取 BYOK）

实现路径：[backend/cmd/capability-probe/main.go](../../backend/cmd/capability-probe/main.go)

- **厂商 API Key 仅从表 `merchant_api_keys.api_key_encrypted` 解密得到**，不在命令行或操作员本机环境变量中配置 `OPENAI_API_KEY` 等上游密钥。
- 与线上健康检查一致：对每条密钥执行 **`GET …/models`**（经 `HealthChecker.FullVerification`）；对 `model_providers.api_format = openai` 的密钥额外尝试 **`POST …/embeddings`**（极小请求体）。
- **仍需**与后端相同的 **`DATABASE_URL`**、**`ENCRYPTION_KEY`**（由部署环境注入进程，与 `docker compose` / systemd 一致），否则无法连库与解密。
- 若商户密钥 `route_mode=litellm`，访问网关时的 Bearer 与现有 [`health_checker`](../../backend/services/health_checker.go) 一致：优先使用进程环境变量 **`LITELLM_MASTER_KEY`**（属于部署侧基础设施，不是把 BYOK 明文放到本机 env）。

### 在部署机执行（示例）

```bash
ssh -i ~/.ssh/tencent_cloud_deploy root@119.29.173.89
cd /opt/pintuotuo/backend
# 与线上 backend 容器相同环境变量（DATABASE_URL、ENCRYPTION_KEY 等）
docker compose exec -T backend sh -lc 'cd /app && ./bin/capability-probe -out /tmp/capability-probe-output.csv'
```

若镜像内尚未编译该子命令，可在挂载源码的树中：

```bash
docker compose exec -T backend sh -lc 'cd /app/backend && go run ./cmd/capability-probe -out /tmp/capability-probe-output.csv'
```

本地 Makefile：`make capability-probe`（可选 `CAPABILITY_PROBE_FLAGS='-out /tmp/cap.csv -provider openai -limit 5'`）。

探测输出 CSV 建议写到 `/tmp` 等受控目录；仓库已 `.gitignore`：`documentation/capability/capability-probe-output*.csv`。

## 与代码的对照

- 端点路径常量：[`backend/services/execution_layer.go`](../../backend/services/execution_layer.go) 中 `endpointPathSuffixes`。
- 模型列表探测（仅 GET models）：[`backend/services/provider_probe.go`](../../backend/services/provider_probe.go)。

## 下一步（需人工）

- **0.4 / 0.5**：在预发/生产按 Runbook 导出配置后运行脚本，将 CSV 摘要回填矩阵。
- **0.6**：召开范围签字会，冻结 P1 格子；将结论链接回 Epic / 本 README。
