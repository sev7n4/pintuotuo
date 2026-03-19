#!/bin/bash

set -e

PROJECT_DIR="/opt/pintuotuo"

echo "=========================================="
echo "  拼脱脱项目 - 生产环境部署脚本"
echo "=========================================="
echo ""

cd $PROJECT_DIR

echo "1. 检查环境配置..."
if [ ! -f ".env" ]; then
    echo "错误: .env 文件不存在"
    echo "请先创建 .env 文件，参考 .env.example"
    exit 1
fi

echo ""
echo "2. 拉取最新代码..."
if [ -d ".git" ]; then
    git pull
fi

echo ""
echo "3. 停止旧容器..."
docker-compose -f docker-compose.prod.yml down || true

echo ""
echo "4. 清理旧镜像..."
docker image prune -f

echo ""
echo "5. 构建镜像..."
docker-compose -f docker-compose.prod.yml build --no-cache

echo ""
echo "6. 启动服务..."
docker-compose -f docker-compose.prod.yml up -d

echo ""
echo "7. 等待服务启动..."
sleep 10

echo ""
echo "8. 检查服务状态..."
docker-compose -f docker-compose.prod.yml ps

echo ""
echo "9. 健康检查..."
echo "后端健康检查:"
curl -s http://localhost:8080/api/v1/health || echo "后端未就绪"

echo ""
echo "前端检查:"
curl -s -o /dev/null -w "HTTP状态码: %{http_code}\n" http://localhost:80 || echo "前端未就绪"

echo ""
echo "=========================================="
echo "  部署完成！"
echo "=========================================="
echo ""
echo "服务地址:"
echo "  前端: http://$(curl -s ifconfig.me)"
echo "  后端 API: http://$(curl -s ifconfig.me):8080"
echo "  Prometheus: http://$(curl -s ifconfig.me):9090"
echo "  Grafana: http://$(curl -s ifconfig.me):3001"
echo ""
echo "查看日志: docker-compose -f docker-compose.prod.yml logs -f"
echo "重启服务: docker-compose -f docker-compose.prod.yml restart"
echo "停止服务: docker-compose -f docker-compose.prod.yml down"
echo ""
