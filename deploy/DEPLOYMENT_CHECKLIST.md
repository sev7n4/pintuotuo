# 生产环境部署检查清单

## 部署前检查 (Pre-Deployment)

### 基础设施

- [ ] Kubernetes 集群已创建并可访问
- [ ] kubectl 已配置正确的 context
- [ ] 集群节点状态正常 (`kubectl get nodes`)
- [ ] 集群资源充足 (CPU/内存/存储)

### 数据库

- [ ] PostgreSQL 数据库已创建
- [ ] 数据库连接信息已获取
- [ ] 数据库用户权限正确
- [ ] SSL 连接已配置 (生产环境必须)

### 缓存

- [ ] Redis 实例已创建
- [ ] Redis 连接信息已获取
- [ ] Redis 密码已设置

### 域名和证书

- [ ] 域名已购买
- [ ] DNS 管理权限已获取
- [ ] 域名已备案 (中国大陆)

### 容器镜像

- [ ] Docker Hub 账户已准备
- [ ] 镜像仓库访问权限已配置
- [ ] 后端镜像已推送 (`pintuotuo/backend:latest`)
- [ ] 前端镜像已推送 (`pintuotuo/frontend:latest`)

---

## Secrets 配置检查

### 必需的 Secrets

- [ ] `database-url` - PostgreSQL 连接字符串
- [ ] `redis-url` - Redis 连接字符串
- [ ] `jwt-secret` - JWT 签名密钥 (32+ 字符)
- [ ] `encryption-key` - 数据加密密钥 (32 字符)

### 支付配置 (可选)

- [ ] `alipay-app-id` - 支付宝应用 ID
- [ ] `alipay-private-key` - 支付宝私钥
- [ ] `wechat-app-id` - 微信应用 ID
- [ ] `wechat-api-key` - 微信 API 密钥

### 邮件配置 (可选)

- [ ] `smtp-host` - SMTP 服务器
- [ ] `smtp-password` - SMTP 密码

---

## 部署步骤检查

### 步骤 1: 命名空间

```bash
kubectl create namespace pintuotuo
```
- [ ] 命名空间已创建

### 步骤 2: Secrets

```bash
kubectl apply -f deploy/k8s/secrets.yml
```
- [ ] Secrets 已创建
- [ ] 验证: `kubectl get secrets -n pintuotuo`

### 步骤 3: ConfigMap

```bash
kubectl apply -f - <<EOF
...
EOF
```
- [ ] ConfigMap 已创建
- [ ] 验证: `kubectl get configmap -n pintuotuo`

### 步骤 4: Ingress Controller

```bash
kubectl apply -f https://...ingress-nginx...
```
- [ ] NGINX Ingress Controller 已安装
- [ ] 验证: `kubectl get pods -n ingress-nginx`

### 步骤 5: Cert-Manager

```bash
kubectl apply -f https://...cert-manager...
```
- [ ] Cert-Manager 已安装
- [ ] ClusterIssuer 已创建
- [ ] 验证: `kubectl get pods -n cert-manager`

### 步骤 6: DNS 配置

- [ ] A 记录: `api.pintuotuo.com` -> Ingress IP
- [ ] A 记录: `www.pintuotuo.com` -> Ingress IP
- [ ] DNS 解析已生效

### 步骤 7: 应用部署

```bash
kubectl apply -f deploy/k8s/app.yml
```
- [ ] Deployment 已创建
- [ ] Service 已创建
- [ ] Ingress 已创建
- [ ] HPA 已创建

### 步骤 8: 数据库迁移

- [ ] 迁移脚本已执行
- [ ] 数据库表已创建
- [ ] 索引已创建

---

## 部署后验证

### Pod 状态

```bash
kubectl get pods -n pintuotuo
```
- [ ] 所有 Pod 状态为 Running
- [ ] Pod 重启次数为 0
- [ ] Ready 状态为 1/1 或更高

### Service 状态

```bash
kubectl get svc -n pintuotuo
```
- [ ] backend service 已创建
- [ ] frontend service 已创建

### Ingress 状态

```bash
kubectl get ingress -n pintuotuo
```
- [ ] Ingress 已创建
- [ ] ADDRESS 已分配

### 健康检查

```bash
curl https://api.pintuotuo.com/api/v1/health
```
- [ ] 后端健康检查返回 200
- [ ] 前端页面可访问

### SSL 证书

```bash
kubectl get certificate -n pintuotuo
```
- [ ] 证书状态为 Ready
- [ ] HTTPS 访问正常

---

## 功能验证

### 用户功能

- [ ] 用户注册功能正常
- [ ] 用户登录功能正常
- [ ] Token 刷新功能正常

### 商品功能

- [ ] 商品列表可访问
- [ ] 商品详情可查看
- [ ] 商品搜索功能正常

### 订单功能

- [ ] 订单创建功能正常
- [ ] 订单查询功能正常
- [ ] 订单取消功能正常

### 支付功能

- [ ] 支付发起功能正常
- [ ] 支付回调处理正常

---

## 监控配置

### Prometheus

- [ ] Prometheus 已部署
- [ ] 指标采集正常
- [ ] 告警规则已配置

### Grafana

- [ ] Grafana 已部署
- [ ] 数据源已配置
- [ ] 仪表板已导入

### 日志

- [ ] 日志采集正常
- [ ] 日志可查询

---

## 安全检查

### 网络安全

- [ ] HTTPS 强制跳转已配置
- [ ] CORS 配置正确
- [ ] Rate Limiting 已启用

### 访问控制

- [ ] RBAC 已配置
- [ ] ServiceAccount 已创建
- [ ] 网络策略已配置 (可选)

### 密钥管理

- [ ] 所有密钥已通过 Secrets 管理
- [ ] 密钥未硬编码在代码中
- [ ] 密钥轮换计划已制定

---

## 备份和恢复

### 数据库备份

- [ ] 自动备份已配置
- [ ] 备份保留策略已设置
- [ ] 恢复流程已测试

### 配置备份

- [ ] Kubernetes 配置已备份
- [ ] Secrets 已安全存储

---

## 文档

- [ ] 部署文档已更新
- [ ] 运维手册已创建
- [ ] 应急预案已制定
- [ ] 联系人列表已更新

---

**检查人**: ________________  
**检查日期**: ________________  
**签字确认**: ________________
