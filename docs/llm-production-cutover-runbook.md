# LLM 网关主方案收敛与灰度切换 Runbook

## 0. 双轨网关部署与联调（跑通 checklist）

### 0.1 部署行为

- 生产部署采用「CI 预构建镜像 + 服务器拉取启动」：使用 **`docker-compose -f docker-compose.prod.yml -f docker-compose.prod.images.yml --profile llm-gateway`**，会拉起 **LiteLLM**（`pintuotuo-litellm`）与 **OneAPI**（`pintuotuo-oneapi`）。部署流水线在启动后会检查这两个容器处于 **running**。
- 服务器项目目录下的 **`.env`** 需与 `docker-compose.prod.yml` 对齐（说明见仓库根目录 `.env.example`），至少关注：
  - **`LLM_GATEWAY_ACTIVE`**：`none` | `litellm` | `oneapi`（切网关前可保持 `none`，仅先验证容器与冒烟脚本）。
  - **`LLM_GATEWAY_LITELLM_URL` / `LLM_GATEWAY_ONEAPI_URL`**：Compose 默认 `http://litellm:4000`、`http://oneapi:3000` 即可。
  - **路径 A（平台 Key → 网关）**：在 `.env` 中配置 **`LITELLM_MASTER_KEY`** 和/或 **`ONEAPI_ACCESS_TOKEN`**（已注入 **backend** 容器，对应 `api_proxy` 的 `resolveGatewayAuthToken`）。
  - **路径 B（BYOK）**：不配上述平台 Key 时，对 **OpenAI 格式** 出站请求使用 **商户库内解密的 Key** 作为 `Authorization: Bearer`。
  - LiteLLM / OneAPI **服务容器**自身的环境变量（如 `OPENAI_API_KEY`、`ONEAPI_SQL_DSN` 等）按各镜像要求填写，否则网关进程虽起但上游调用会失败。

### 0.2 主机上冒烟（容器 + 可选 Master Key）

在服务器项目根目录执行：

```bash
chmod +x scripts/verify_llm_gateway_smoke.sh
set -a && source .env && set +a
./scripts/verify_llm_gateway_smoke.sh
```

若已设置 `LITELLM_MASTER_KEY`，脚本会请求本机 `http://127.0.0.1:4000/v1/models` 做最小验证。

### 0.3 业务侧 BYOK 联调

- 使用 **`POST /api/v1/proxy/chat`** 或 **`POST /api/v1/openai/v1/chat/completions`**（鉴权以现有 `AuthMiddleware` / `APIKeyOrJWTAuthMiddleware` 为准），在 **清空或未设置** 后端环境变量 `LITELLM_MASTER_KEY`（及 OneAPI 对应 token）时，确认出站走 **商户 Key**。

## 1. 选型收敛规则

在连续 7 天观测窗口内对 LiteLLM 与 OneAPI 进行评分，按 `docs/llm_supplychain_dod.md` 权重计算总分。

- 总分更高者作为主方案。
- 另一路保留为灾备通道，不下线配置。

## 2. 发布前检查

- 双轨均可完成代表模型调用。
- 告警链路可触发且通知可达。
- Promptfoo 最近一次回归通过率 >= 80%。
- 熔断/重试指标已接入并可查询。

## 3. 灰度步骤

1. 5% 灰度（24h）
   - 设置 `SMART_ROUTING_GRAY_PERCENT=5`
   - 设置 `LLM_GATEWAY_ACTIVE=<winner>`
2. 25% 灰度（24h）
3. 50% 灰度（24h）
4. 100% 切换（24h）

每个阶段都需检查：
- 成功率、5xx、429、P95 延迟
- 关键业务接口错误率
- token 成本是否异常上升

## 4. 回滚策略

满足任一条件立即回滚：
- 5xx 错误率 > 5% 持续 10 分钟
- P95 延迟 > 3 秒持续 10 分钟
- 关键链路可用性 < 99%

回滚动作：

1. `LLM_GATEWAY_ACTIVE=none`（退回原有直连）
2. `SMART_ROUTING_GRAY_PERCENT=0`
3. 保留故障时段 trace 与告警记录，进入复盘

## 5. 复盘输出

- 触发时间线
- 根因分类（上游限流/网络/配置/代码）
- 修复项与防复发动作
- 是否恢复灰度及下一次窗口
