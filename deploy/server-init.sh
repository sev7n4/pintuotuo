#!/bin/bash

set -e

echo "=========================================="
echo "  拼脱脱项目 - 服务器初始化脚本"
echo "=========================================="
echo ""

echo "1. 更新系统..."
yum update -y || apt update -y

echo ""
echo "2. 安装必要工具..."
yum install -y git curl wget vim || apt install -y git curl wget vim

echo ""
echo "3. 安装 Docker..."
if ! command -v docker &> /dev/null; then
    curl -fsSL https://get.docker.com | sh
    systemctl start docker
    systemctl enable docker
    echo "Docker 安装完成"
else
    echo "Docker 已安装"
fi

echo ""
echo "4. 安装 Docker Compose..."
if ! command -v docker-compose &> /dev/null; then
    curl -L "https://github.com/docker/compose/releases/download/v2.23.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose
    ln -sf /usr/local/bin/docker-compose /usr/bin/docker-compose
    echo "Docker Compose 安装完成"
else
    echo "Docker Compose 已安装"
fi

echo ""
echo "5. 配置 Docker 镜像加速（国内服务器）..."
mkdir -p /etc/docker
cat > /etc/docker/daemon.json << 'EOF'
{
  "registry-mirrors": [
    "https://registry.cn-hangzhou.aliyuncs.com",
    "https://mirror.ccs.tencentyun.com"
  ],
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "100m",
    "max-file": "3"
  }
}
EOF
systemctl daemon-reload
systemctl restart docker

echo ""
echo "6. 配置防火墙..."
if command -v firewall-cmd &> /dev/null; then
    systemctl start firewalld
    systemctl enable firewalld
    firewall-cmd --permanent --add-port=22/tcp
    firewall-cmd --permanent --add-port=80/tcp
    firewall-cmd --permanent --add-port=443/tcp
    firewall-cmd --permanent --add-port=8080/tcp
    firewall-cmd --reload
    echo "防火墙配置完成"
elif command -v ufw &> /dev/null; then
    ufw allow 22/tcp
    ufw allow 80/tcp
    ufw allow 443/tcp
    ufw allow 8080/tcp
    ufw --force enable
    echo "防火墙配置完成"
else
    echo "请手动配置防火墙开放端口: 22, 80, 443, 8080"
fi

echo ""
echo "7. 创建项目目录..."
mkdir -p /opt/pintuotuo
cd /opt/pintuotuo

echo ""
echo "=========================================="
echo "  初始化完成！"
echo "=========================================="
echo ""
echo "Docker 版本: $(docker --version)"
echo "Docker Compose 版本: $(docker-compose --version)"
echo ""
echo "下一步："
echo "1. 上传项目代码到 /opt/pintuotuo"
echo "2. 配置环境变量"
echo "3. 运行部署脚本"
echo ""
