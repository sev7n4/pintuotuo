# 部署和运维指南 - Pintuotuo 项目

## 目录

1. [部署流程](#部署流程)
2. [环境配置](#环境配置)
3. [海外大模型代理部署（Phase 1）](#海外大模型代理部署phase-1)
4. [监控和日志](#监控和日志)
5. [备份和恢复](#备份和恢复)
6. [常见问题](#常见问题)

---

## 部署流程

### 开发环境（开发者本地）

```bash
# 1. 初始化环境
bash scripts/setup.sh

# 2. 启动服务
make docker-up
make dev

# 3. 运行迁移
make migrate
```

### 测试环境（CI/CD自动化）

GitHub Actions 会在以下情况自动部署：
- 推送到 `develop` 分支
- 创建 Pull Request
- 合并到 `main` 分支

**自动检查流程：**
1. 运行后端测试和代码检查
2. 运行前端测试和 TypeScript 检查
3. 构建 Docker 镜像
4. 运行安全扫描
5. 生成覆盖率报告

### 生产环境（手动或 CI/CD）

#### 前置条件

- Kubernetes 集群（或其他容器编排）
- PostgreSQL 数据库（管理托管）
- Redis 缓存（管理托管或自托管）
- Kafka 消息队列
- Docker Registry（存储镜像）

#### 部署步骤

```bash
# 1. 构建镜像
docker build -t pintuotuo/backend:v1.0.0 ./backend
docker build -t pintuotuo/frontend:v1.0.0 ./frontend

# 2. 推送到 Registry
docker push pintuotuo/backend:v1.0.0
docker push pintuotuo/frontend:v1.0.0

# 3. 更新 Kubernetes 部署
kubectl apply -f k8s/deployment.yaml

# 4. 验证部署
kubectl get pods -l app=pintuotuo
kubectl logs deployment/pintuotuo-backend

# 5. 运行数据库迁移
kubectl exec -it deployment/pintuotuo-backend -- /app/pintuotuo-backend migrate

# 6. 检查服务状态
kubectl port-forward svc/pintuotuo-backend-service 8080:8080
curl http://localhost:8080/health
```

---

## 环境配置

### 环境变量管理

**开发环境：**
```bash
cp .env.example .env.development
# 编辑 .env.development
```

**生产环境：**
```bash
# 创建 Kubernetes Secret
kubectl create secret generic pintuotuo-secrets \
  --from-file=.env.production

# 或使用密钥管理服务（如 AWS Secrets Manager）
```

### 必需的环境变量

| 变量 | 环境 | 说明 | 示例 |
|------|------|------|------|
| `PORT` | 全部 | 服务器端口 | 8080 |
| `DATABASE_URL` | 全部 | PostgreSQL 连接 | postgresql://user:pass@host/db |
| `JWT_SECRET` | 全部 | JWT 签名密钥 | your-secret-key |
| `GIN_MODE` | 全部 | Gin 模式 | debug/release |
| `REDIS_URL` | 全部 | Redis 连接 | redis://localhost:6379 |
| `KAFKA_BROKERS` | 全部 | Kafka 代理 | localhost:9092 |

### BYOK路由模式环境变量

BYOK (Bring Your Own Key) 路由模式需要以下环境变量配置：

| 变量 | 环境 | 说明 | 示例 |
|------|------|------|------|
| `LLM_GATEWAY_LITELLM_URL` | 生产 | LiteLLM网关地址 | http://litellm:4000 |
| `LITELLM_MASTER_KEY` | 生产 | LiteLLM认证密钥 | sk-litellm-master-key |

**配置示例**：
```bash
# .env.production
LLM_GATEWAY_LITELLM_URL=http://litellm:4000
LITELLM_MASTER_KEY=sk-litellm-master-key
```

**Docker Compose配置**：
```yaml
# docker-compose.prod.yml
services:
  backend:
    environment:
      - LLM_GATEWAY_LITELLM_URL=http://litellm:4000
      - LITELLM_MASTER_KEY=sk-litellm-master-key
```

**Kubernetes配置**：
```yaml
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: byok-config
data:
  LLM_GATEWAY_LITELLM_URL: "http://litellm:4000"
---
# k8s/secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: byok-secrets
type: Opaque
stringData:
  LITELLM_MASTER_KEY: "sk-litellm-master-key"
```

**配置优先级**：
1. `route_config.endpoints.{mode}.{region}` - 最高优先级
2. `route_config.base_url` 或 `route_config.endpoint_url`
3. 环境变量配置
4. `model_providers.api_base_url` - 最低优先级

### PgBouncer 与连接池（可选）

后端使用 `database/sql` + `lib/pq`，`DATABASE_URL` 指向 **PostgreSQL 直连**（或 **PgBouncer**）均可。应用侧已通过 `SetMaxOpenConns` / `SetMaxIdleConns` 限制连接数。

若前置 **PgBouncer** 以减轻数据库连接数、提升伸缩性，请注意：

- **推荐 `pool_mode=session`**：与 `database/sql` 的常见用法（同一连接上多语句、预编译语句）兼容性最好。
- **`transaction` 池化** 下，部分客户端对 prepared statement 较敏感；若出现异常协议类错误，需查阅 PgBouncer 与驱动文档（例如是否需禁用服务端预编译或改用兼容模式），**不要**仅通过加大连接数掩盖同一事务内「结果集未关闭又发新语句」这类应用层错误——此类问题应在业务 SQL 代码中修复。

### 腾讯云 Docker Compose：邮箱魔法链接（Mock，不发信）

适用于 `docker-compose.prod.yml` 部署：在服务器项目目录（如 `/opt/pintuotuo`）的 **`.env`** 中增加或修改为：

```bash
AUTH_MAGIC_LINK=true
EMAIL_MAGIC_MOCK=true
```

**重要：**

- Mock 模式下**不要**同时配齐 `SMTP_HOST`、`SMTP_PORT`、`SMTP_FROM`（及发信账号），否则后端会优先走真实发信，接口响应里**不会出现** `debug_link`。
- 验证链接必须是你**在浏览器里能打开的公网地址**。二选一配置：
  - **推荐**：`PUBLIC_API_BASE_URL=https://你的域名`（与对外 API 一致，无尾斜杠）。
  - 或只设 `FRONTEND_URL=https://你的站点`，并保证 Nginx 将 `https://你的站点/api/v1` 反代到后端（与线上一致）。

写入 `.env` 后重建后端容器使环境变量生效：

```bash
cd /opt/pintuotuo   # 以实际路径为准
docker-compose -f docker-compose.prod.yml up -d --force-recreate backend
```

**如何自测（Mock）**

1. **能力开关**：`curl -sS 'https://你的域名/api/v1/users/auth/capabilities'`，应见 `"email_magic": true`。
2. **请求发链**：`POST /api/v1/users/email/magic/send`，JSON `{"email":"你的邮箱@example.com"}`，应 **HTTP 200** 且 JSON 含 **`debug_link`**（同时服务端日志会有 `[EMAIL_MAGIC_MOCK]`）。
3. **完成登录**：用浏览器打开 `debug_link` 中的完整 URL（一次性，约 15 分钟内），应重定向到前端并带上 token，完成登录。
4. **前端**：登录页「发送邮箱魔法链接」可点；开发环境下前端可能对 `debug_link` 再弹一条 `message.info`，生产环境请以接口或日志为准。

切换到真实邮件时：关闭 `EMAIL_MAGIC_MOCK`（或设为 `false`），配置完整 `SMTP_*`，并去掉与 Mock 冲突的依赖。

---

## 海外大模型代理部署（Phase 1）

国内服务器无法直连海外大模型 Provider API（OpenAI、Anthropic、Google、OpenRouter 等），需要通过 HTTPS_PROXY 出站代理解决网络限制。本节指导在腾讯云 CVM 上部署 Mihomo（Clash Meta 内核）作为宿主机级 HTTP 代理。

### 环境信息

| 项目 | 值 |
|------|------|
| 服务器 IP | `119.29.173.89` |
| 部署路径 | `/opt/pintuotuo` |
| SSH 密钥 | `~/.ssh/tencent_cloud_deploy` |
| 架构 | Linux x86_64 (腾讯云 CVM) |

### 第一步：SSH 登录服务器

```bash
ssh -i ~/.ssh/tencent_cloud_deploy root@119.29.173.89
```

> 如果不是 root 用户，将后续命令中的 `sudo` 保留；如果是 root，`sudo` 可省略。

### 第二步：安装 Mihomo（Clash Meta 内核）

**方式 A：使用项目自带脚本**

```bash
cd /opt/pintuotuo
bash deploy/scripts/setup-clash.sh
```

脚本会自动完成：下载 mihomo 二进制、安装到 `/usr/local/bin/mihomo`、创建配置目录 `/etc/mihomo`、复制配置文件、创建 systemd 服务并启动。

**⚠️ 脚本启动前必须先编辑配置文件**（见第三步），建议分步执行：

```bash
cd /opt/pintuotuo

# 1. 下载 mihomo
curl -L -o /tmp/mihomo.gz \
  "https://github.com/MetaCubeX/mihomo/releases/download/v1.19.0/mihomo-linux-amd64-v1.19.0.gz"
gunzip /tmp/mihomo.gz
chmod +x /tmp/mihomo
sudo cp /tmp/mihomo /usr/local/bin/mihomo

# 2. 创建配置目录
sudo mkdir -p /etc/mihomo

# 3. 验证安装
mihomo -v
```

> **如果 GitHub 下载超时**（大陆服务器常见），可以在本地下载后 scp 上传：
> ```bash
> # 本地机器执行
> curl -L -o /tmp/mihomo.gz \
>   "https://github.com/MetaCubeX/mihomo/releases/download/v1.19.0/mihomo-linux-amd64-v1.19.0.gz"
> scp -i ~/.ssh/tencent_cloud_deploy /tmp/mihomo.gz root@119.29.173.89:/tmp/
>
> # 然后在服务器上执行
> gunzip /tmp/mihomo.gz && chmod +x /tmp/mihomo && sudo cp /tmp/mihomo /usr/local/bin/
> ```

### 第三步：编辑 Clash 配置文件

```bash
sudo vi /etc/mihomo/config.yaml
```

需要修改 `proxies` 部分，将 `<REPLACE_WITH_YOUR_SERVER>` 和 `<REPLACE_WITH_YOUR_PASSWORD>` 替换为实际代理节点信息。

**方式 A：手动配置单个节点**

```yaml
proxies:
  - name: "my-proxy"
    type: trojan            # 根据你的节点类型修改
    server: your.server.com # ← 替换为实际服务器地址
    port: 443               # ← 替换为实际端口
    password: your-password # ← 替换为实际密码
    udp: true
    sni: your.server.com    # ← 通常与 server 相同
    skip-cert-verify: false # 生产环境建议 false
```

支持的代理类型：`trojan`、`vmess`、`vless`、`ss`（Shadowsocks）、`hysteria2`。

**方式 B：使用订阅链接（推荐）**

如果有 Clash 订阅 URL，可用 `proxy-providers` 替代手动配置：

```yaml
mixed-port: 7890
allow-lan: false
bind-address: 127.0.0.1
mode: rule
log-level: info
ipv6: false

proxy-providers:
  my-provider:
    type: http
    url: "https://your-subscription-url/clash"  # ← 替换为你的订阅链接
    interval: 3600
    path: ./providers/my-provider.yaml
    health-check:
      enable: true
      url: https://www.gstatic.com/generate_204
      interval: 300

proxy-groups:
  - name: "overseas-api"
    type: select
    use:
      - my-provider
    proxies:
      - DIRECT

rules:
  - DOMAIN-SUFFIX,openai.com,overseas-api
  - DOMAIN-SUFFIX,anthropic.com,overseas-api
  - DOMAIN-SUFFIX,googleapis.com,overseas-api
  - DOMAIN-SUFFIX,openrouter.ai,overseas-api
  - DOMAIN-SUFFIX,github.com,overseas-api
  - IP-CIDR,127.0.0.0/8,DIRECT
  - IP-CIDR,10.0.0.0/8,DIRECT
  - IP-CIDR,172.16.0.0/12,DIRECT
  - IP-CIDR,192.168.0.0/16,DIRECT
  - MATCH,DIRECT
```

**关键配置说明**：

| 配置项 | 值 | 说明 |
|--------|------|------|
| `mixed-port` | `7890` | HTTP+SOCKS5 混合代理端口 |
| `allow-lan` | `false` | 仅监听本机，不暴露到外网 |
| `bind-address` | `127.0.0.1` | 绑定本机回环地址 |
| `rules` | 见上方 | 海外 API 域名走代理，其余直连 |

### 第四步：创建 systemd 服务并启动

```bash
sudo tee /etc/systemd/system/mihomo.service > /dev/null <<'EOF'
[Unit]
Description=Mihomo (Clash Meta)
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/mihomo -d /etc/mihomo
Restart=on-failure
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable mihomo
sudo systemctl start mihomo
sudo systemctl status mihomo
```

预期输出：
```
● mihomo.service - Mihomo (Clash Meta)
     Loaded: loaded (/etc/systemd/system/mihomo.service; enabled)
     Active: active (running)
```

如果启动失败，查看日志：
```bash
sudo journalctl -u mihomo -n 50 --no-pager
```

常见失败原因：配置文件 YAML 语法错误、代理节点信息不正确、端口 7890 被占用。

### 第五步：验证 Clash 代理可用

```bash
# 1. 检查 mihomo 进程
ps aux | grep mihomo

# 2. 检查端口监听
ss -tlnp | grep 7890
# 预期: LISTEN 127.0.0.1:7890

# 3. 测试代理访问 OpenAI
curl -x http://127.0.0.1:7890 \
  -s -o /dev/null -w "HTTP %{http_code} (%{time_total}s)\n" \
  --connect-timeout 10 --max-time 15 \
  https://api.openai.com/v1/models
# 预期: HTTP 401 (1.5s)  ← 401 说明网络通了，只是没带 API Key

# 4. 批量测试所有海外 Provider
cd /opt/pintuotuo
bash deploy/scripts/test-connectivity.sh
```

预期输出：
```
=== Connectivity Test ===
Proxy: http://127.0.0.1:7890

--- Via Proxy ---
OpenAI       proxy    AUTH (401) (1500ms)
Anthropic    proxy    AUTH (401) (800ms)
Google       proxy    AUTH (401) (600ms)
OpenRouter   proxy    OK (200) (500ms)

--- Direct (expected to timeout for overseas) ---
OpenAI       direct   TIMEOUT (10000ms)
Anthropic    direct   TIMEOUT (10000ms)
...
```

> `AUTH (401)` = 网络通了，只是没带 API Key，这是正确的预期结果。
> `OK (200)` = OpenRouter 不需要 Key 也能列出模型。
> `TIMEOUT` = 直连不通，符合国内服务器预期。

### 第六步：配置 Docker Compose 代理注入

Clash 验证通过后，让 Backend 容器通过代理出站：

```bash
cd /opt/pintuotuo

# 1. 创建 proxy override 文件
cp docker-compose.proxy.override.yml.example docker-compose.proxy.override.yml

# 2. 确认内容（默认即可，无需修改）
cat docker-compose.proxy.override.yml
```

默认内容：
```yaml
services:
  backend:
    extra_hosts:
      - "host.docker.internal:host-gateway"
    environment:
      - HTTPS_PROXY=${HTTPS_PROXY:-http://host.docker.internal:7890}
      - HTTP_PROXY=${HTTP_PROXY:-}
      - NO_PROXY=localhost,127.0.0.1,postgres,redis,backend,frontend,pintuotuo-postgres,pintuotuo-redis
```

关键说明：
- `host.docker.internal:host-gateway` — 让容器能解析宿主机 IP
- `HTTPS_PROXY=http://host.docker.internal:7890` — 容器通过宿主机 Clash 代理出站
- `NO_PROXY` — 内部服务直连，不走代理

### 第七步：重启 Backend 容器（带代理）

```bash
cd /opt/pintuotuo

docker compose -f docker-compose.prod.yml \
  -f docker-compose.prod.images.yml \
  -f docker-compose.proxy.override.yml \
  up -d --force-recreate backend
```

> **注意**：后续每次部署（CI/CD 自动部署）也需要包含 proxy override，见第八步。

### 第八步：更新 CI/CD 部署 Workflow

当前 `.github/workflows/deploy-tencent.yml` 的启动命令是：
```bash
IMAGE_TAG=${{ github.sha }} docker-compose -f docker-compose.prod.yml -f docker-compose.prod.images.yml up -d
```

需要加上 proxy override：
```bash
IMAGE_TAG=${{ github.sha }} docker-compose -f docker-compose.prod.yml -f docker-compose.prod.images.yml -f docker-compose.proxy.override.yml up -d
```

修改 `.github/workflows/deploy-tencent.yml` 中对应的 `docker-compose up` 命令，添加 `-f docker-compose.proxy.override.yml`。

### 第九步：端到端验证

Backend 重启后，验证海外模型 API 代理访问：

```bash
# 1. 检查 backend 容器内的环境变量
docker exec pintuotuo-backend env | grep -i proxy
# 预期输出:
# HTTPS_PROXY=http://host.docker.internal:7890
# NO_PROXY=localhost,127.0.0.1,postgres,redis,...

# 2. 从容器内测试代理连通性
docker exec pintuotuo-backend wget -q -O - \
  --timeout=10 \
  "https://api.openai.com/v1/models" 2>&1 | head -5
# 预期: 返回 401 JSON 或模型列表

# 3. 通过 Backend API 测试（需要有效的海外 Provider API Key）
# 先在管理后台配置 OpenAI Provider 和 BYOK Key，然后：
curl -X POST http://127.0.0.1:8080/api/v1/proxy/chat \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_USER_TOKEN" \
  -d '{
    "provider": "openai",
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

### 操作清单

| # | 步骤 | 命令/操作 | 验证方式 |
|---|------|----------|----------|
| 1 | SSH 登录 | `ssh -i ~/.ssh/tencent_cloud_deploy root@119.29.173.89` | 登录成功 |
| 2 | 安装 mihomo | 手动下载或 `setup-clash.sh` | `mihomo -v` |
| 3 | 编辑配置 | `sudo vi /etc/mihomo/config.yaml` | 填入实际代理节点 |
| 4 | 启动服务 | `sudo systemctl start mihomo` | `systemctl status mihomo` |
| 5 | 验证代理 | `curl -x http://127.0.0.1:7890 https://api.openai.com/v1/models` | HTTP 401 |
| 6 | 创建 override | `cp docker-compose.proxy.override.yml.example docker-compose.proxy.override.yml` | 文件存在 |
| 7 | 重启 backend | `docker compose -f ... -f proxy.override.yml up -d --force-recreate backend` | 容器运行 |
| 8 | 更新 CI/CD | 修改 `deploy-tencent.yml` 添加 `-f docker-compose.proxy.override.yml` | PR 合并 |
| 9 | E2E 验证 | 通过 Backend API 请求海外模型 | 成功返回 |

### 常见问题排查

| 问题 | 排查命令 | 解决方案 |
|------|----------|----------|
| mihomo 启动失败 | `journalctl -u mihomo -n 50` | 检查 YAML 语法和节点配置 |
| 代理超时 | `curl -x http://127.0.0.1:7890 https://www.google.com` | 检查代理节点是否可用 |
| 容器无法访问宿主机 | `docker exec backend ping host.docker.internal` | 检查 `extra_hosts` 配置 |
| Backend 不走代理 | `docker exec backend env \| grep PROXY` | 检查 override 文件是否加载 |
| 内部服务走代理了 | 检查 `NO_PROXY` 环境变量 | 确保 postgres/redis 在 NO_PROXY 中 |
| CI/CD 部署后代理丢失 | 检查 workflow 中的 docker-compose 命令 | 添加 `-f docker-compose.proxy.override.yml` |

---

## 监控和日志

### 日志收集

**后端日志：**
```bash
# 本地查看
docker-compose logs -f pintuotuo_backend

# Kubernetes
kubectl logs -f deployment/pintuotuo-backend
```

**前端日志：**
浏览器开发工具 → Console 标签页

### 指标监控

**推荐工具：**
- **Prometheus** - 指标收集
- **Grafana** - 可视化仪表板
- **Datadog** - 完整的监控解决方案
- **New Relic** - APM 监控

**关键指标：**
- API 响应时间
- 错误率
- 数据库连接数
- 内存使用率
- CPU 使用率
- 请求吞吐量

### 健康检查端点

```bash
# 后端健康检查
curl http://localhost:8080/health

# 预期响应
{
  "status": "healthy",
  "message": "Pintuotuo Backend Server is running",
  "timestamp": "2026-03-14T23:00:00Z"
}
```

### 内部经济对账（Token 用量）

零售与 `api_proxy` 按量扣费相关的抽样对账、排障步骤见：[`backend/doc_internal_economics_runbook.md`](backend/doc_internal_economics_runbook.md)。

---

## 备份和恢复

### 数据库备份

**自动备份（推荐）：**
```bash
# 使用 pg_dump 定期备份
0 2 * * * pg_dump -U pintuotuo pintuotuo_db > /backups/pintuotuo_$(date +\%Y\%m\%d).sql
```

**手动备份：**
```bash
# 备份
docker-compose exec postgres pg_dump -U pintuotuo pintuotuo_db > backup.sql

# 恢复
docker-compose exec -T postgres psql -U pintuotuo pintuotuo_db < backup.sql
```

### 数据库恢复

```bash
# 1. 停止应用
docker-compose down

# 2. 恢复数据
docker-compose up -d postgres
docker-compose exec -T postgres psql -U pintuotuo pintuotuo_db < backup.sql

# 3. 重启应用
docker-compose up -d
```

### Redis 备份

```bash
# 备份
docker-compose exec redis redis-cli SAVE

# 复制 dump.rdb
docker cp pintuotuo_redis:/data/dump.rdb ./redis_backup.rdb

# 恢复
docker cp redis_backup.rdb pintuotuo_redis:/data/dump.rdb
docker-compose restart redis
```

---

## 常见问题

### Q: 部署后应用无法连接数据库

**解决方案：**
1. 检查数据库凭证
2. 验证网络连接
3. 检查防火墙规则
4. 查看应用日志

```bash
kubectl logs deployment/pintuotuo-backend
```

### Q: 内存使用率持续增长

**可能原因：**
- 连接池泄漏
- 缓存无限增长
- Goroutine 泄漏

**诊断：**
```bash
# 查看内存使用
kubectl top pod <pod-name>

# 获取堆转储
curl http://localhost:8080/debug/pprof/heap > heap.dump
```

### Q: API 响应时间缓慢

**优化步骤：**
1. 检查数据库查询（添加索引）
2. 启用 Redis 缓存
3. 增加服务副本数
4. 分析瓶颈（使用 pprof）

```bash
# CPU 分析
curl http://localhost:8080/debug/pprof/profile > cpu.prof
go tool pprof cpu.prof
```

### Q: 如何升级到新版本

**蓝绿部署：**
```bash
# 1. 部署新版本（作为独立 Deployment）
kubectl apply -f k8s/deployment-v2.yaml

# 2. 运行迁移
kubectl exec -it deployment/pintuotuo-backend-v2 -- /app/pintuotuo-backend migrate

# 3. 切换流量（更新 Service selector）
kubectl patch service pintuotuo-backend-service -p '{"spec":{"selector":{"version":"v2"}}}'

# 4. 监控新版本
kubectl logs -f deployment/pintuotuo-backend-v2

# 5. 如需回滚
kubectl patch service pintuotuo-backend-service -p '{"spec":{"selector":{"version":"v1"}}}'
```

### Q: 如何处理数据库迁移失败

**回滚策略：**
```bash
# 1. 检查迁移状态
kubectl exec deployment/pintuotuo-backend -- /app/pintuotuo-backend status

# 2. 恢复备份
docker-compose exec -T postgres psql -U pintuotuo pintuotuo_db < backup_before_migration.sql

# 3. 修复迁移脚本

# 4. 重新应用
make migrate
```

---

## 性能优化建议

### 数据库优化

```sql
-- 添加常用查询的索引
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_products_merchant ON products(merchant_id);
CREATE INDEX idx_orders_user ON orders(user_id);
CREATE INDEX idx_orders_created ON orders(created_at DESC);

-- 启用自动分析
ALTER TABLE orders SET (autovacuum_vacuum_scale_factor = 0.01);
ALTER TABLE products SET (autovacuum_vacuum_scale_factor = 0.01);
```

### 缓存策略

```go
// 使用 Redis 缓存热数据
cache.Set("products:list", products, 1*time.Hour)
```

### 数据库连接池配置

```go
DB.SetMaxOpenConns(50)    // 根据负载调整
DB.SetMaxIdleConns(10)    // 保持连接池大小
DB.SetConnMaxLifetime(10 * time.Minute)
```

---

## 安全最佳实践

1. **定期更新依赖**
   ```bash
   cd backend && go get -u ./...
   cd frontend && npm audit fix
   ```

2. **启用 HTTPS**
   - 使用自签名证书或 Let's Encrypt
   - 配置 nginx 使用 SSL

3. **密钥轮换**
   - 定期更新 JWT_SECRET
   - 轮换数据库密码

4. **备份验证**
   - 定期测试备份恢复
   - 验证备份完整性

5. **访问控制**
   - 使用 RBAC 限制 Kubernetes 访问
   - 启用审计日志

---

**最后更新**：2026-05-04
**维护者**：DevOps Team
**下次审查**：2026-06-04
