# 部署和运维指南 - Pintuotuo 项目

## 目录

1. [部署流程](#部署流程)
2. [环境配置](#环境配置)
3. [监控和日志](#监控和日志)
4. [备份和恢复](#备份和恢复)
5. [常见问题](#常见问题)

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

**最后更新**：2026-03-14
**维护者**：DevOps Team
**下次审查**：2026-04-14
