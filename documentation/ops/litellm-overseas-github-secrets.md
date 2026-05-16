# 海外 LiteLLM 节点 — GitHub Secrets 与首次部署规范

> 适用 Workflow：[`.github/workflows/deploy-litellm-overseas.yml`](../../.github/workflows/deploy-litellm-overseas.yml)  
> 目标主机：腾讯云新加坡（示例 IP `43.160.204.9`），部署目录 `/opt/pintuotuo-litellm`  
> 与大陆 `deploy-tencent.yml` / `deploy-litellm.yml` **解耦**，仅同步 `deploy/litellm/` SSOT 配置。

---

## 1. 失败现象与根因

合并 PR 后若 **Deploy LiteLLM Overseas** 失败，日志中常见：

| 日志特征 | 含义 |
|----------|------|
| `HOST="@"` | `TENCENT_CLOUD_OVERSEAS_USER` / `TENCENT_CLOUD_OVERSEAS_IP` 未配置 |
| `echo "" > ~/.ssh/deploy_key` | `TENCENT_CLOUD_OVERSEAS_SSH_KEY` 未配置 |
| `OVERSEAS_DIR:` 为空 | `TENCENT_CLOUD_OVERSEAS_LITELLM_DIR` 未配置 |
| `Authorization: Bearer` 后无内容 | `LITELLM_MASTER_KEY` 未配置（或未挂到 `production` 环境） |
| `ssh: usage: ssh ...` | 上述 SSH 相关 Secret 缺失导致 `ssh @` 非法调用 |

**结论**：Workflow 已触发，但 **`production` 环境下海外专用 Secrets 尚未写入**。这与大陆 `TENCENT_CLOUD_*` 是两套变量，不会自动继承。

---

## 2. Secret 配置位置（必须）

本 Workflow 声明了：

```yaml
environment: production
```

因此 Secrets 必须配置在：

**GitHub 仓库 → Settings → Environments → `production` → Environment secrets**

仅配置在 **Repository secrets** 且未在 Environment 中重复时，部分 Secret 可能对 `production` 环境不可见；**推荐全部写入 `production` Environment secrets**，与 `deploy-tencent.yml` 保持一致。

路径示例：

`https://github.com/<owner>/pintuotuo/settings/environments/production`

---

## 3. 必填 Secrets 清单

| Secret 名称 | 类型 | 说明 | 示例值（勿照抄密钥） |
|-------------|------|------|---------------------|
| `TENCENT_CLOUD_OVERSEAS_SSH_KEY` | 私钥全文 | 海外机 SSH **私钥**（PEM），对应已加入 `authorized_keys` 的公钥 | `-----BEGIN OPENSSH PRIVATE KEY-----` … |
| `TENCENT_CLOUD_OVERSEAS_USER` | 字符串 | SSH 登录用户 | `ubuntu` |
| `TENCENT_CLOUD_OVERSEAS_IP` | 字符串 | 海外机公网 IP | `43.160.204.9` |
| `TENCENT_CLOUD_OVERSEAS_LITELLM_DIR` | 字符串 | 海外 LiteLLM 部署目录（无尾部 `/`） | `/opt/pintuotuo-litellm` |
| `LITELLM_MASTER_KEY` | 字符串 | LiteLLM Master Key；须与**大陆 backend** `LITELLM_MASTER_KEY`、海外机 `.env` 中一致 | `sk-litellm-...` |

### 与大陆 Secret 对照

| 大陆（`deploy-tencent.yml`） | 海外（`deploy-litellm-overseas.yml`） |
|------------------------------|-------------------------------------|
| `TENCENT_CLOUD_SSH_KEY` | `TENCENT_CLOUD_OVERSEAS_SSH_KEY` |
| `TENCENT_CLOUD_USER` | `TENCENT_CLOUD_OVERSEAS_USER` |
| `TENCENT_CLOUD_IP` | `TENCENT_CLOUD_OVERSEAS_IP` |
| `TENCENT_CLOUD_PROJECT_DIR` | `TENCENT_CLOUD_OVERSEAS_LITELLM_DIR` |
| `LITELLM_MASTER_KEY`（若已配） | **同名复用**，值必须一致 |

---

## 4. 配置步骤（操作清单）

### 4.1 海外机准备 SSH（一次性）

在**能登录海外机的终端**执行（示例用户 `ubuntu`）：

```bash
# 本机生成专用密钥（若无）
ssh-keygen -t ed25519 -f ~/.ssh/pintuotuo_overseas_deploy -N "" -C "github-actions-litellm-overseas"

# 将公钥写入海外机（按实际 IP/用户替换）
ssh-copy-id -i ~/.ssh/pintuotuo_overseas_deploy.pub ubuntu@43.160.204.9

# 验证免密登录
ssh -i ~/.ssh/pintuotuo_overseas_deploy ubuntu@43.160.204.9 'whoami && docker --version'
```

将 **`~/.ssh/pintuotuo_overseas_deploy` 私钥完整内容**（含首尾行）粘贴到 GitHub Secret `TENCENT_CLOUD_OVERSEAS_SSH_KEY`。

> 若海外机此前仅密码登录，必须先完成上述公钥部署，否则 CI 无法 SSH。

### 4.2 海外机准备部署目录与 `.env`（一次性）

Workflow **会同步**仓库内配置文件，**不会**创建或覆盖 `.env`。

```bash
ssh ubuntu@43.160.204.9

sudo mkdir -p /opt/pintuotuo-litellm
sudo chown "$USER:$USER" /opt/pintuotuo-litellm
cd /opt/pintuotuo-litellm

# 创建 .env（值与大陆生产 LITELLM_MASTER_KEY 一致）
cat > .env <<'EOF'
LITELLM_MASTER_KEY=<与大陆 backend 相同的 master key>
EOF
chmod 600 .env
```

### 4.3 在 GitHub 写入 Environment secrets

1. 打开 **Settings → Environments → production**。
2. 点击 **Add secret**，按 [第 3 节](#3-必填-secrets-清单) 逐项添加。
3. `LITELLM_MASTER_KEY`：从大陆生产 `/opt/pintuotuo` 的 `.env` 或现有 `production` Secret 复制，**勿使用占位符**。

### 4.4 确认大陆 backend 指向海外网关

大陆生产 `docker-compose` / `.env` 应包含（示例）：

```bash
LLM_GATEWAY_LITELLM_URL_OVERSEAS=http://43.160.204.9:4000
```

修改后需走 `deploy-tencent` 重启 backend（非本 Workflow 职责）。

### 4.5 手动重跑部署

1. **Actions** → **Deploy LiteLLM Overseas** → **Run workflow**
2. `branch`: `main`
3. `litellm_image_tag`: `v1.83.3-stable`（与 `docker-compose.overseas.yml` 默认一致）

或在配置 Secrets 后，对失败 Run 点击 **Re-run all jobs**。

---

## 5. 验收标准

部署 Job 成功时应满足：

1. 日志无 `HOST="@"`、无 `ssh: usage`。
2. 海外机存在容器 `pintuotuo-litellm-overseas` 且 `docker ps` 为 Up。
3. 日志：`LiteLLM /health HTTP status: 200`。
4. 大陆机探测（可选）：

```bash
curl -s -o /dev/null -w '%{http_code}\n' \
  -H "Authorization: Bearer <LITELLM_MASTER_KEY>" \
  http://43.160.204.9:4000/health
# 期望 200
```

---

## 6. 触发时机

| 触发方式 | 说明 |
|----------|------|
| `push` → `main` 且变更 `deploy/litellm/**` 或本 workflow 文件 | 自动部署海外配置 |
| `workflow_dispatch` | 手动指定分支与镜像 tag |

**不会**随 `deploy-tencent.yml` 自动执行；合并 backend 代码不会单独更新海外 LiteLLM，除非触及上述路径。

---

## 7. 安全与运维约定

1. **禁止**将 SSH 私钥、`LITELLM_MASTER_KEY`、厂商 API Key 提交到 Git。
2. 海外 `.env` 仅保存在海外机，由运维维护；CI 只读校验其存在。
3. 轮换 `LITELLM_MASTER_KEY` 时须**同时**更新：大陆 backend `.env`、`production` Secret、海外 `/opt/pintuotuo-litellm/.env`，再重跑两侧部署。
4. `TENCENT_CLOUD_OVERSEAS_SSH_KEY` 建议仅授予 `ubuntu` + docker 权限，勿使用 root 密钥除非必要。

---

## 8. 常见问题

### Q: 大陆 deploy-tencent 成功，海外仍失败？

A: 两套 Secret 独立。检查 `production` Environment 中 **OVERSEAS_** 前缀四项是否齐全。

### Q: SSH 仍失败，Secrets 已填？

A: 检查私钥是否完整复制、公钥是否在 `~/.ssh/authorized_keys`、安全组是否放行 22、用户是否为 `ubuntu`。

### Q: 健康检查非 200？

A: 查看 Job 日志中 `docker logs pintuotuo-litellm-overseas`；常见原因：`.env` 中 `LITELLM_MASTER_KEY` 与 GitHub Secret 不一致、镜像拉取失败、4000 端口被占用。

### Q: 能否复用大陆 `TENCENT_CLOUD_SSH_KEY`？

A: 仅当**同一私钥**已授权登录海外机时可临时复用，但仍建议独立密钥与独立 Secret 名，便于权限隔离与轮换。

---

## 9. 相关文档

- [`deploy/litellm/README.md`](../../deploy/litellm/README.md) — LiteLLM SSOT 与命令
- [`documentation/capability/byok-routing-ssot.md`](../capability/byok-routing-ssot.md) — BYOK 路由与模型 ID SSOT
- [`DEPLOYMENT.md`](../../DEPLOYMENT.md) — 整体部署指南

**文档版本**：2026-05-16（PR #520 合并后）
