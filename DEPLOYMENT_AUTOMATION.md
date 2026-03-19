# 自动化部署配置指南

## 概述
本指南将帮助您配置GitHub Actions自动化部署到腾讯云服务器的完整流程。

## 腾讯云服务器准备

### 1. 服务器初始化
1. 登录腾讯云控制台，创建一台云服务器（推荐使用Ubuntu 20.04 LTS）
2. 确保服务器已安装以下软件：
   - Git
   - Docker
   - Docker Compose

### 2. 项目目录准备
在服务器上创建项目目录：
```bash
mkdir -p /opt/pintuotuo
cd /opt/pintuotuo
git clone <your-github-repository-url> .
```

### 3. 环境配置
1. 复制环境变量文件：
   ```bash
   cp .env.example .env
   ```
2. 编辑 `.env` 文件，填写必要的环境变量。

## GitHub Secrets配置

在GitHub仓库的「Settings」→「Secrets and variables」→「Actions」中添加以下Secrets：

| Secret名称 | 描述 | 示例值 |
|-----------|------|--------|
| `TENCENT_CLOUD_SSH_KEY` | 服务器SSH私钥 | `-----BEGIN OPENSSH PRIVATE KEY-----
...
-----END OPENSSH PRIVATE KEY-----` |
| `TENCENT_CLOUD_IP` | 服务器公网IP地址 | `111.222.333.444` |
| `TENCENT_CLOUD_USER` | 服务器登录用户名 | `ubuntu` |
| `TENCENT_CLOUD_PROJECT_DIR` | 项目目录路径 | `/opt/pintuotuo` |
| `SMTP_SERVER` | 邮件服务器地址（可选） | `smtp.gmail.com` |
| `SMTP_PORT` | 邮件服务器端口（可选） | `587` |
| `SMTP_USERNAME` | 邮件发送用户名 | `your-email@gmail.com` |
| `SMTP_PASSWORD` | 邮件发送密码（应用专用密码） | `app-specific-password` |
| `DEPLOYMENT_NOTIFICATION_EMAIL` | 部署通知接收邮箱 | `team@example.com` |

## 自动化部署流程

### 触发方式
1. **自动触发**：当代码推送到 `main` 或 `master` 分支时自动执行
2. **手动触发**：在GitHub Actions界面手动触发，可选择部署分支

### 部署流程
1. GitHub Actions拉取最新代码
2. 建立与腾讯云服务器的SSH连接
3. 在服务器上拉取最新代码
4. 执行部署脚本 `deploy/deploy-prod.sh`
5. 验证部署状态
6. 生成部署报告

## 手动部署（备用方案）

如果需要手动部署，可直接登录服务器执行：
```bash
cd /opt/pintuotuo
git pull
./deploy/deploy-prod.sh
```

## 故障排查

### 常见问题
1. **SSH连接失败**：检查SSH密钥是否正确，服务器防火墙是否开放22端口
2. **部署脚本执行失败**：检查服务器环境配置和依赖安装
3. **服务启动失败**：查看Docker容器日志 `docker-compose -f docker-compose.prod.yml logs`

### 日志查看
- 服务器部署日志：`/opt/pintuotuo/deploy.log`
- GitHub Actions运行日志：在GitHub仓库的「Actions」标签页查看

## 安全建议
1. 使用密钥认证，禁用密码登录
2. 定期更新服务器系统和依赖
3. 限制服务器的入站流量，只开放必要的端口
4. 定期备份数据和配置
