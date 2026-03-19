#!/bin/bash

set -e

PROJECT_DIR="/opt/pintuotuo"
LOG_FILE="$PROJECT_DIR/deploy.log"

# 初始化日志文件
echo "[$(date '+%Y-%m-%d %H:%M:%S')] 开始部署" > $LOG_FILE

echo "==========================================" | tee -a $LOG_FILE
echo "  拼脱脱项目 - 生产环境部署脚本" | tee -a $LOG_FILE
echo "==========================================" | tee -a $LOG_FILE
echo "" | tee -a $LOG_FILE

cd $PROJECT_DIR

echo "1. 检查环境配置..." | tee -a $LOG_FILE
if [ ! -f ".env" ]; then
    echo "错误: .env 文件不存在" | tee -a $LOG_FILE
    echo "请先创建 .env 文件，参考 .env.example" | tee -a $LOG_FILE
    exit 1
fi
echo "环境配置检查通过" | tee -a $LOG_FILE

echo ""
echo "2. 拉取最新代码..." | tee -a $LOG_FILE
if [ -d ".git" ]; then
    git pull | tee -a $LOG_FILE
else
    echo "警告: 非Git仓库，跳过代码拉取" | tee -a $LOG_FILE
fi

echo ""
echo "3. 停止旧容器..." | tee -a $LOG_FILE
docker-compose -f docker-compose.prod.yml down 2>&1 | tee -a $LOG_FILE || true

echo ""
echo "4. 清理旧镜像..." | tee -a $LOG_FILE
docker image prune -f 2>&1 | tee -a $LOG_FILE

echo ""
echo "5. 构建镜像..." | tee -a $LOG_FILE
docker-compose -f docker-compose.prod.yml build --no-cache 2>&1 | tee -a $LOG_FILE

echo ""
echo "6. 启动服务..." | tee -a $LOG_FILE
docker-compose -f docker-compose.prod.yml up -d 2>&1 | tee -a $LOG_FILE

echo ""
echo "7. 等待服务启动..." | tee -a $LOG_FILE
sleep 15

echo ""
echo "8. 检查服务状态..." | tee -a $LOG_FILE
docker-compose -f docker-compose.prod.yml ps 2>&1 | tee -a $LOG_FILE

echo ""
echo "9. 健康检查..." | tee -a $LOG_FILE

# 后端健康检查
echo "后端健康检查:" | tee -a $LOG_FILE
BACKEND_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/v1/health 2>&1 || echo "000")
if [[ "$BACKEND_STATUS" == "200" ]]; then
    echo "后端服务正常 (HTTP $BACKEND_STATUS)" | tee -a $LOG_FILE
else
    echo "后端服务异常 (HTTP $BACKEND_STATUS)" | tee -a $LOG_FILE
    # 不直接退出，继续检查其他服务
fi

# 前端健康检查
echo "" | tee -a $LOG_FILE
echo "前端检查:" | tee -a $LOG_FILE
FRONTEND_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:80 2>&1 || echo "000")
if [[ "$FRONTEND_STATUS" =~ ^(200|302)$ ]]; then
    echo "前端服务正常 (HTTP $FRONTEND_STATUS)" | tee -a $LOG_FILE
else
    echo "前端服务异常 (HTTP $FRONTEND_STATUS)" | tee -a $LOG_FILE
    # 不直接退出，继续完成部署
fi

echo "" | tee -a $LOG_FILE
echo "==========================================" | tee -a $LOG_FILE
echo "  部署完成！" | tee -a $LOG_FILE
echo "==========================================" | tee -a $LOG_FILE
echo "" | tee -a $LOG_FILE

# 获取服务器公网IP
SERVER_IP=$(curl -s ifconfig.me 2>/dev/null || echo "localhost")

echo "服务地址:" | tee -a $LOG_FILE
echo "  前端: http://$SERVER_IP" | tee -a $LOG_FILE
echo "  后端 API: http://$SERVER_IP:8080" | tee -a $LOG_FILE
echo "  Prometheus: http://$SERVER_IP:9090" | tee -a $LOG_FILE
echo "  Grafana: http://$SERVER_IP:3001" | tee -a $LOG_FILE
echo "" | tee -a $LOG_FILE

echo "查看日志: docker-compose -f docker-compose.prod.yml logs -f" | tee -a $LOG_FILE
echo "重启服务: docker-compose -f docker-compose.prod.yml restart" | tee -a $LOG_FILE
echo "停止服务: docker-compose -f docker-compose.prod.yml down" | tee -a $LOG_FILE
echo "" | tee -a $LOG_FILE

echo "[$(date '+%Y-%m-%d %H:%M:%S')] 部署完成" >> $LOG_FILE

# 输出部署状态供CI/CD系统使用
if [[ "$BACKEND_STATUS" == "200" && "$FRONTEND_STATUS" =~ ^(200|302)$ ]]; then
    echo "DEPLOYMENT_STATUS=SUCCESS" >> $GITHUB_ENV 2>/dev/null || true
    exit 0
else
    echo "DEPLOYMENT_STATUS=FAILURE" >> $GITHUB_ENV 2>/dev/null || true
    exit 1
fi
