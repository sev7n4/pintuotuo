# 云服务器部署指南

## 目录

1. [服务器准备](#服务器准备)
2. [初始化服务器](#初始化服务器)
3. [上传项目代码](#上传项目代码)
4. [配置环境变量](#配置环境变量)
5. [部署应用](#部署应用)
6. [验证部署](#验证部署)
7. [配置域名和SSL](#配置域名和ssl)
8. [常用命令](#常用命令)

---

## 服务器准备

### 服务器信息

请记录您的服务器信息：

| 信息项 | 值 |
|--------|-----|
| 公网 IP | ________________ |
| SSH 端口 | 22 (默认) |
| 用户名 | root (默认) |
| 密码 | ________________ |

### 云服务商安全组配置

在云服务商控制台开放以下端口：

| 端口 | 用途 | 必须 |
|------|------|------|
| 22 | SSH | ✅ |
| 80 | HTTP (前端) | ✅ |
| 443 | HTTPS | ✅ |
| 8080 | 后端 API | ✅ |
| 9090 | Prometheus | 可选 |
| 3001 | Grafana | 可选 |

---

## 初始化服务器

### 步骤 1: SSH 连接服务器

```bash
# 本地终端执行
ssh root@<您的服务器IP>
```

### 步骤 2: 下载并运行初始化脚本

```bash
# 下载初始化脚本
curl -o server-init.sh https://raw.githubusercontent.com/your-repo/pintuotuo/main/deploy/server-init.sh

# 添加执行权限
chmod +x server-init.sh

# 运行初始化脚本
./server-init.sh
```

**或手动执行：**

```bash
# 更新系统
yum update -y  # CentOS
# 或
apt update -y && apt upgrade -y  # Ubuntu

# 安装 Docker
curl -fsSL https://get.docker.com | sh
systemctl start docker
systemctl enable docker

# 安装 Docker Compose
curl -L "https://github.com/docker/compose/releases/download/v2.23.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# 配置 Docker 镜像加速
mkdir -p /etc/docker
cat > /etc/docker/daemon.json << 'EOF'
{
  "registry-mirrors": [
    "https://registry.cn-hangzhou.aliyuncs.com",
    "https://mirror.ccs.tencentyun.com"
  ]
}
EOF
systemctl daemon-reload
systemctl restart docker

# 创建项目目录
mkdir -p /opt/pintuotuo
```

---

## 上传项目代码

### 方法 1: 使用 Git（推荐）

```bash
cd /opt/pintuotuo
git clone https://github.com/your-username/pintuotuo.git .
```

### 方法 2: 使用 SCP 上传

在**本地电脑**执行：

```bash
# 打包项目
cd /Users/4seven/workspace/pintuotuo
tar -czvf pintuotuo.tar.gz --exclude='node_modules' --exclude='.git' --exclude='coverage.out' .

# 上传到服务器
scp pintuotuo.tar.gz root@<服务器IP>:/opt/pintuotuo/
```

在**服务器**执行：

```bash
cd /opt/pintuotuo
tar -xzvf pintuotuo.tar.gz
rm pintuotuo.tar.gz
```

### 方法 3: 使用 rsync（推荐大文件）

```bash
# 本地执行
rsync -avz --exclude 'node_modules' --exclude '.git' \
  /Users/4seven/workspace/pintuotuo/ root@<服务器IP>:/opt/pintuotuo/
```

---

## 配置环境变量

### 创建 .env 文件

```bash
cd /opt/pintuotuo
cat > .env << 'EOF'
# 数据库密码
DB_PASSWORD=your_secure_password_here

# JWT 密钥 (至少32个字符)
JWT_SECRET=your_jwt_secret_key_at_least_32_characters_long

# 加密密钥 (32个字符)
ENCRYPTION_KEY=your_32_character_encryption_key

# 前端 API 地址
VITE_API_URL=http://<您的服务器IP>:8080/api/v1

# CORS 允许的源
CORS_ALLOWED_ORIGINS=http://<您的服务器IP>,http://<您的域名>

# Grafana 管理员密码
GRAFANA_ADMIN_USER=admin
GRAFANA_ADMIN_PASSWORD=your_grafana_password

# 支付配置 (可选)
ALIPAY_APP_ID=
ALIPAY_PRIVATE_KEY=
WECHAT_APP_ID=
WECHAT_API_KEY=
EOF
```

### 生成安全密钥

```bash
# 生成随机密码
DB_PASSWORD=$(openssl rand -base64 16)
JWT_SECRET=$(openssl rand -base64 32)
ENCRYPTION_KEY=$(openssl rand -base64 24 | cut -c1-32)

# 更新 .env 文件
sed -i "s/your_secure_password_here/$DB_PASSWORD/" .env
sed -i "s/your_jwt_secret_key_at_least_32_characters_long/$JWT_SECRET/" .env
sed -i "s/your_32_character_encryption_key/$ENCRYPTION_KEY/" .env
```

---

## 部署应用

### 一键部署

```bash
cd /opt/pintuotuo
chmod +x deploy/deploy-prod.sh
./deploy/deploy-prod.sh
```

### 或手动部署

```bash
cd /opt/pintuotuo

# 构建并启动
docker-compose -f docker-compose.prod.yml up -d --build

# 查看状态
docker-compose -f docker-compose.prod.yml ps

# 查看日志
docker-compose -f docker-compose.prod.yml logs -f
```

---

## 验证部署

### 检查服务状态

```bash
# 查看容器状态
docker ps

# 检查后端健康
curl http://localhost:8080/api/v1/health

# 检查前端
curl http://localhost:80
```

### 访问测试

| 服务 | 地址 |
|------|------|
| 前端 | http://\<服务器IP\> |
| 后端 API | http://\<服务器IP\>:8080/api/v1/health |
| Prometheus | http://\<服务器IP\>:9090 |
| Grafana | http://\<服务器IP\>:3001 |

---

## 配置域名和SSL

### 步骤 1: 域名解析

在域名服务商添加 A 记录：

| 记录类型 | 主机记录 | 记录值 |
|----------|----------|--------|
| A | @ | \<服务器IP\> |
| A | www | \<服务器IP\> |
| A | api | \<服务器IP\> |

### 步骤 2: 安装 Certbot

```bash
# CentOS
yum install -y certbot

# Ubuntu
apt install -y certbot
```

### 步骤 3: 申请 SSL 证书

```bash
# 停止前端容器（释放80端口）
docker stop pintuotuo-frontend

# 申请证书
certbot certonly --standalone -d yourdomain.com -d www.yourdomain.com

# 启动前端容器
docker start pintuotuo-frontend
```

### 步骤 4: 配置 Nginx SSL

创建 SSL 配置：

```bash
cat > /opt/pintuotuo/nginx/ssl.conf << 'EOF'
server {
    listen 443 ssl;
    server_name yourdomain.com www.yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;

    location / {
        proxy_pass http://pintuotuo-frontend:80;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}

server {
    listen 443 ssl;
    server_name api.yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;

    location / {
        proxy_pass http://pintuotuo-backend:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}

server {
    listen 80;
    server_name yourdomain.com www.yourdomain.com api.yourdomain.com;
    return 301 https://$host$request_uri;
}
EOF
```

---

## 常用命令

### Docker Compose 命令

```bash
# 启动所有服务
docker-compose -f docker-compose.prod.yml up -d

# 停止所有服务
docker-compose -f docker-compose.prod.yml down

# 重启服务
docker-compose -f docker-compose.prod.yml restart

# 查看日志
docker-compose -f docker-compose.prod.yml logs -f

# 查看特定服务日志
docker-compose -f docker-compose.prod.yml logs -f backend

# 重新构建
docker-compose -f docker-compose.prod.yml up -d --build

# 进入容器
docker exec -it pintuotuo-backend /bin/sh
```

### 数据库操作

```bash
# 进入 PostgreSQL
docker exec -it pintuotuo-postgres psql -U pintuotuo -d pintuotuo_db

# 备份数据库
docker exec pintuotuo-postgres pg_dump -U pintuotuo pintuotuo_db > backup.sql

# 恢复数据库
cat backup.sql | docker exec -i pintuotuo-postgres psql -U pintuotuo pintuotuo_db
```

### Redis 操作

```bash
# 进入 Redis
docker exec -it pintuotuo-redis redis-cli

# 清空缓存
docker exec pintuotuo-redis redis-cli FLUSHALL
```

---

## 故障排查

### 容器无法启动

```bash
# 查看容器日志
docker logs pintuotuo-backend
docker logs pintuotuo-frontend

# 检查资源
docker stats
```

### 数据库连接失败

```bash
# 检查 PostgreSQL 状态
docker exec pintuotuo-postgres pg_isready

# 检查连接
docker exec pintuotuo-postgres psql -U pintuotuo -d pintuotuo_db -c "SELECT 1"
```

### 端口被占用

```bash
# 查看端口占用
netstat -tlnp | grep :80
netstat -tlnp | grep :8080

# 杀掉占用进程
kill -9 <PID>
```

---

**文档版本**: 1.0  
**最后更新**: 2026-03-19
