# 生产环境部署指南

## 目录

1. [部署前准备](#部署前准备)
2. [基础设施要求](#基础设施要求)
3. [部署步骤](#部署步骤)
4. [部署后验证](#部署后验证)
5. [常见问题](#常见问题)

---

## 部署前准备

### 1. 必需的账户和资源

| 资源 | 说明 | 状态 |
|------|------|------|
| Kubernetes 集群 | 推荐 GKE、EKS 或 AKS | ⬜ |
| 容器镜像仓库 | Docker Hub 或私有仓库 | ⬜ |
| 域名 | pintuotuo.com | ⬜ |
| SSL 证书 | Let's Encrypt (自动) | ⬜ |
| PostgreSQL 数据库 | 托管服务推荐 | ⬜ |
| Redis 缓存 | 托管服务推荐 | ⬜ |

### 2. 本地工具要求

```bash
# 检查工具是否安装
kubectl version --client
docker --version
helm version  # 可选
```

---

## 基础设施要求

### Kubernetes 集群规格

| 组件 | 最小配置 | 推荐配置 |
|------|----------|----------|
| 节点数 | 2 | 3+ |
| CPU | 2 vCPU | 4+ vCPU |
| 内存 | 4 GB | 8+ GB |
| 存储 | 20 GB | 50+ GB |

### 数据库规格

| 资源 | 最小配置 | 推荐配置 |
|------|----------|----------|
| PostgreSQL | 1 vCPU, 2GB | 2+ vCPU, 4+ GB |
| Redis | 1 vCPU, 1GB | 2+ vCPU, 2+ GB |

---

## 部署步骤

### 步骤 1: 配置 Kubernetes 集群访问

```bash
# 配置 kubectl 访问集群
# GKE 示例
gcloud container clusters get-credentials <cluster-name> --region <region>

# EKS 示例
aws eks update-kubeconfig --name <cluster-name> --region <region>

# 验证连接
kubectl get nodes
```

### 步骤 2: 创建命名空间

```bash
kubectl apply -f - <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: pintuotuo
EOF
```

### 步骤 3: 配置 Secrets

创建 `deploy/k8s/secrets.yml`:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: pintuotuo-secrets
  namespace: pintuotuo
type: Opaque
stringData:
  database-url: "postgresql://<user>:<password>@<host>:5432/pintuotuo_db?sslmode=require"
  redis-url: "redis://:<password>@<host>:6379"
  jwt-secret: "<your-32-char-jwt-secret>"
  encryption-key: "<your-32-char-encryption-key>"
  alipay-app-id: "<alipay-app-id>"
  alipay-private-key: "<base64-encoded-private-key>"
  wechat-app-id: "<wechat-app-id>"
  wechat-api-key: "<wechat-api-key>"
  smtp-password: "<smtp-password>"
```

应用 Secrets:

```bash
kubectl apply -f deploy/k8s/secrets.yml
```

### 步骤 4: 配置 ConfigMap

```bash
kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: pintuotuo-config
  namespace: pintuotuo
data:
  PORT: "8080"
  GIN_MODE: "release"
  DB_MAX_OPEN_CONNS: "50"
  DB_MAX_IDLE_CONNS: "10"
  DB_CONN_MAX_LIFETIME: "30m"
  CORS_ALLOWED_ORIGINS: "https://pintuotuo.com,https://www.pintuotuo.com"
EOF
```

### 步骤 5: 安装 Ingress Controller

```bash
# 安装 NGINX Ingress Controller
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.2/deploy/static/provider/cloud/deploy.yaml

# 等待就绪
kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=90s
```

### 步骤 6: 安装 Cert-Manager (SSL)

```bash
# 安装 Cert-Manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.1/cert-manager.yaml

# 创建 ClusterIssuer
kubectl apply -f - <<EOF
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: your-email@example.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
EOF
```

### 步骤 7: 配置域名 DNS

将域名指向 Ingress Controller 的外部 IP:

```bash
# 获取 Ingress IP
kubectl get svc -n ingress-nginx

# 配置 DNS 记录:
# api.pintuotuo.com -> <ingress-ip>
# www.pintuotuo.com -> <ingress-ip>
```

### 步骤 8: 部署应用

```bash
# 部署应用
kubectl apply -f deploy/k8s/app.yml

# 检查部署状态
kubectl get pods -n pintuotuo
kubectl get services -n pintuotuo
kubectl get ingress -n pintuotuo
```

### 步骤 9: 运行数据库迁移

```bash
# 进入后端 Pod 执行迁移
kubectl exec -it -n pintuotuo deployment/backend -- /bin/sh

# 在 Pod 内执行
for f in /app/migrations/*.sql; do
  psql "$DATABASE_URL" -f "$f"
done
```

### 步骤 10: 验证部署

```bash
# 检查 Pod 状态
kubectl get pods -n pintuotuo

# 检查服务状态
kubectl get svc -n pintuotuo

# 检查 Ingress
kubectl get ingress -n pintuotuo

# 检查日志
kubectl logs -f -n pintuotuo deployment/backend
```

---

## 部署后验证

### 健康检查

```bash
# 后端健康检查
curl https://api.pintuotuo.com/api/v1/health

# 前端访问
curl https://www.pintuotuo.com
```

### 功能验证

1. **用户注册/登录**: https://www.pintuotuo.com/register
2. **API 文档**: https://api.pintuotuo.com/swagger/index.html
3. **健康检查**: https://api.pintuotuo.com/api/v1/health

---

## 常见问题

### Q: Pod 一直处于 Pending 状态

```bash
# 检查事件
kubectl describe pod <pod-name> -n pintuotuo

# 常见原因:
# - 资源不足
# - PVC 未绑定
# - 节点选择器不匹配
```

### Q: Ingress 无法访问

```bash
# 检查 Ingress 状态
kubectl describe ingress -n pintuotuo

# 检查 Cert-Manager 日志
kubectl logs -n cert-manager deployment/cert-manager
```

### Q: 数据库连接失败

```bash
# 检查 Secret 配置
kubectl get secret -n pintuotuo pintuotuo-secrets -o yaml

# 测试连接
kubectl run psql-test --rm -it --image=postgres:15 -- \
  psql "postgresql://<user>:<password>@<host>:5432/pintuotuo_db"
```

---

## 回滚操作

```bash
# 查看部署历史
kubectl rollout history deployment/backend -n pintuotuo

# 回滚到上一版本
kubectl rollout undo deployment/backend -n pintuotuo

# 回滚到指定版本
kubectl rollout undo deployment/backend -n pintuotuo --to-revision=2
```

---

**文档版本**: 1.0  
**最后更新**: 2026-03-19
