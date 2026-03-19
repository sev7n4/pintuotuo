#!/bin/bash

set -e

echo "=========================================="
echo "  临时部署脚本"
echo "=========================================="
echo ""

# 设置远程仓库地址
echo "1. 设置远程仓库地址..."
git remote set-url origin `https://github.com/sev7n4/pintuotuo.git`

echo ""
echo "2. 修复 git 仓库..."
# 修复 git 仓库
git fsck --full
git gc --prune=now

echo ""
echo "3. 拉取最新代码..."
# 拉取最新代码
git fetch origin
git reset --hard origin/main

echo ""
echo "4. 重新构建并启动..."
# 重新构建并启动
docker-compose -f docker-compose.prod.yml up -d --build

echo ""
echo "5. 查看后端日志..."
# 查看日志
docker-compose -f docker-compose.prod.yml logs -f backend

echo ""
echo "=========================================="
echo "  部署完成！"
echo "=========================================="
