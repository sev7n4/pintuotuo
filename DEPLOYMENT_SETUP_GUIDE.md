# 腾讯云自动化部署配置指南

## 第一步：在GitHub仓库中配置Secrets

### 1.1 访问GitHub仓库设置
1. 打开您的GitHub仓库：https://github.com/sev7n4/pintuotuo
2. 点击「Settings」标签
3. 在左侧菜单中找到「Secrets and variables」→「Actions」
4. 点击「New repository secret」按钮

### 1.2 添加必需的Secrets

#### Secret 1: TENCENT_CLOUD_SSH_KEY
- **名称**: `TENCENT_CLOUD_SSH_KEY`
- **值**: 您的腾讯云服务器SSH私钥内容
- **获取方法**:
  ```bash
  # 在本地执行以下命令查看私钥内容
  cat ~/.ssh/id_rsa
  # 或者如果您使用的是腾讯云的密钥对
  cat ~/.ssh/tencent_cloud_key.pem
  ```
- **注意**: 私钥内容应该包含 `-----BEGIN OPENSSH PRIVATE KEY-----` 和 `-----END OPENSSH PRIVATE KEY-----`

#### Secret 2: TENCENT_CLOUD_IP
- **名称**: `TENCENT_CLOUD_IP`
- **值**: 您的腾讯云服务器公网IP地址
- **示例**: `111.222.333.444`
- **获取方法**: 
  - 在腾讯云控制台查看云服务器详情
  - 或者在服务器上执行 `curl ifconfig.me`

#### Secret 3: TENCENT_CLOUD_USER
- **名称**: `TENCENT_CLOUD_USER`
- **值**: 服务器登录用户名
- **示例**: `ubuntu` 或 `root` 或 `centos`

#### Secret 4: TENCENT_CLOUD_PROJECT_DIR
- **名称**: `TENCENT_CLOUD_PROJECT_DIR`
- **值**: 项目在服务器上的目录路径
- **示例**: `/opt/pintuotuo` 或 `/home/ubuntu/pintuotuo`

### 1.3 添加可选的邮件通知Secrets（可选）

#### Secret 5: SMTP_USERNAME
- **名称**: `SMTP_USERNAME`
- **值**: 邮件发送用户名（通常是邮箱地址）
- **示例**: `your-email@gmail.com`

#### Secret 6: SMTP_PASSWORD
- **名称**: `SMTP_PASSWORD`
- **值**: 邮件发送密码或应用专用密码
- **注意**: 对于Gmail，需要使用应用专用密码

#### Secret 7: DEPLOYMENT_NOTIFICATION_EMAIL
- **名称**: `DEPLOYMENT_NOTIFICATION_EMAIL`
- **值**: 接收部署通知的邮箱地址
- **示例**: `team@example.com`

## 第二步：在腾讯云服务器上准备环境

### 2.1 登录腾讯云服务器
```bash
ssh ubuntu@您的服务器IP
```

### 2.2 安装必要软件（如果尚未安装）
```bash
# 更新系统
sudo apt update && sudo apt upgrade -y

# 安装Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# 安装Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# 安装Git
sudo apt install git -y

# 重新登录以应用Docker权限
exit
# 重新SSH登录
```

### 2.3 克隆项目代码
```bash
# 创建项目目录
sudo mkdir -p /opt/pintuotuo
sudo chown $USER:$USER /opt/pintuotuo
cd /opt/pintuotuo

# 克隆代码仓库
git clone https://github.com/sev7n4/pintuotuo.git .

# 设置Git凭据缓存（避免每次输入密码）
git config credential.helper store
```

### 2.4 配置环境变量
```bash
# 复制环境变量模板
cp .env.example .env

# 编辑环境变量文件
nano .env
```

在 `.env` 文件中填写以下必要配置：
```env
# 数据库配置
DB_PASSWORD=your_secure_password

# JWT密钥（生成方法：openssl rand -base64 32）
JWT_SECRET=your_jwt_secret_here

# 加密密钥（生成方法：openssl rand -base64 32）
ENCRYPTION_KEY=your_encryption_key_here

# CORS配置
CORS_ALLOWED_ORIGINS=http://your-domain.com,http://localhost:3000

# 前端API地址
VITE_API_URL=http://your-server-ip:8080/api/v1

# Grafana配置（可选）
GRAFANA_ADMIN_USER=admin
GRAFANA_ADMIN_PASSWORD=your_grafana_password
```

### 2.5 配置SSH密钥认证（重要）
```bash
# 在服务器上生成SSH密钥对（如果还没有）
ssh-keygen -t rsa -b 4096 -C "github-actions@tencent-cloud"

# 将公钥添加到authorized_keys
cat ~/.ssh/id_rsa.pub >> ~/.ssh/authorized_keys
chmod 600 ~/.ssh/authorized_keys

# 测试SSH连接（应该不需要密码）
ssh localhost
exit
```

### 2.6 初始部署测试
```bash
# 在服务器上手动执行一次部署，确保环境正确
cd /opt/pintuotuo
docker-compose -f docker-compose.prod.yml up -d --build

# 查看服务状态
docker-compose -f docker-compose.prod.yml ps

# 查看日志
docker-compose -f docker-compose.prod.yml logs -f
```

## 第三步：测试自动化部署流程

### 3.1 提交代码触发自动部署
在本地开发机器上：
```bash
# 确保在项目目录
cd /path/to/pintuotuo

# 创建一个测试提交
echo "# Test deployment" >> README.md
git add README.md
git commit -m "Test automated deployment to Tencent Cloud"
git push origin main
```

### 3.2 查看GitHub Actions执行情况
1. 访问您的GitHub仓库
2. 点击「Actions」标签
3. 查看最新的工作流运行
4. 点击具体的工作流查看详细日志

### 3.3 手动触发部署（可选）
1. 在GitHub仓库的「Actions」页面
2. 选择「Deploy to Tencent Cloud」工作流
3. 点击「Run workflow」
4. 选择要部署的分支
5. 点击绿色的「Run workflow」按钮

## 第四步：验证部署结果

### 4.1 检查服务状态
在腾讯云服务器上执行：
```bash
# 查看Docker容器状态
docker-compose -f docker-compose.prod.yml ps

# 检查后端健康状态
curl http://localhost:8080/api/v1/health

# 检查前端服务
curl -I http://localhost:80
```

### 4.2 查看部署日志
```bash
# 查看部署日志文件
cat /opt/pintuotuo/deploy.log

# 查看Docker服务日志
docker-compose -f docker-compose.prod.yml logs -f backend
docker-compose -f docker-compose.prod.yml logs -f frontend
```

### 4.3 访问服务
- **前端**: http://您的服务器IP
- **后端API**: http://您的服务器IP:8080/api/v1/health
- **Prometheus**: http://您的服务器IP:9090
- **Grafana**: http://您的服务器IP:3001

## 故障排查

### 问题1: SSH连接失败
**症状**: GitHub Actions显示SSH连接超时或拒绝连接

**解决方案**:
```bash
# 在服务器上检查SSH服务状态
sudo systemctl status sshd

# 确保SSH端口开放
sudo ufw allow 22

# 检查SSH配置
sudo nano /etc/ssh/sshd_config
# 确保 PasswordAuthentication no 和 PubkeyAuthentication yes
```

### 问题2: Docker权限问题
**症状**: permission denied while trying to connect to the Docker daemon

**解决方案**:
```bash
# 将用户添加到docker组
sudo usermod -aG docker $USER

# 重新登录或执行
newgrp docker
```

### 问题3: Git拉取失败
**症状**: Git pull或fetch失败

**解决方案**:
```bash
# 检查Git配置
git config --list

# 重新设置远程地址
git remote set-url origin https://github.com/sev7n4/pintuotuo.git

# 手动拉取测试
git pull origin main
```

### 问题4: 环境变量缺失
**症状**: 服务启动失败，提示环境变量未定义

**解决方案**:
```bash
# 检查.env文件是否存在
ls -la /opt/pintuotuo/.env

# 检查文件内容
cat /opt/pintuotuo/.env

# 确保所有必要的环境变量都已填写
```

### 问题5: 端口被占用
**症状**: Docker容器无法启动，提示端口已被使用

**解决方案**:
```bash
# 查看端口占用情况
sudo netstat -tulpn | grep :80
sudo netstat -tulpn | grep :8080

# 停止占用端口的服务
sudo systemctl stop nginx  # 如果有nginx在运行
```

## 安全建议

1. **定期更新系统和软件**
   ```bash
   sudo apt update && sudo apt upgrade -y
   ```

2. **配置防火墙**
   ```bash
   sudo ufw allow 22/tcp    # SSH
   sudo ufw allow 80/tcp    # HTTP
   sudo ufw allow 8080/tcp  # Backend API
   sudo ufw enable
   ```

3. **定期备份数据**
   ```bash
   # 备份PostgreSQL数据
   docker exec pintuotuo-postgres pg_dump -U pintuotuo pintuotuo_db > backup_$(date +%Y%m%d).sql
   ```

4. **监控服务状态**
   - 使用Prometheus和Grafana监控服务健康状态
   - 设置告警规则

## 下一步

完成以上配置后，您的自动化部署流程就已经准备就绪。每次您向main或master分支推送代码时，GitHub Actions会自动：

1. 拉取最新代码
2. 连接到腾讯云服务器
3. 在服务器上拉取最新代码
4. 重新构建并启动Docker容器
5. 验证服务健康状态
6. 发送部署通知（如果配置了邮件）

您可以在GitHub Actions页面查看每次部署的详细日志和状态。
