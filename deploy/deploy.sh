#!/bin/bash

set -e

echo "=========================================="
echo "  拼脱脱生产环境部署脚本"
echo "=========================================="
echo ""

NAMESPACE="pintuotuo"
DEPLOY_DIR="$(cd "$(dirname "$0")" && pwd)"

check_command() {
    if ! command -v $1 &> /dev/null; then
        echo "❌ 错误: $1 未安装"
        exit 1
    fi
    echo "✅ $1 已安装"
}

check_kubectl() {
    if ! kubectl cluster-info &> /dev/null; then
        echo "❌ 错误: 无法连接到 Kubernetes 集群"
        echo "请确保 kubectl 已正确配置"
        exit 1
    fi
    echo "✅ Kubernetes 集群连接正常"
}

echo "1. 检查必需工具..."
check_command kubectl
check_command docker
check_kubectl
echo ""

echo "2. 检查命名空间..."
if kubectl get namespace $NAMESPACE &> /dev/null; then
    echo "✅ 命名空间 $NAMESPACE 已存在"
else
    echo "创建命名空间 $NAMESPACE..."
    kubectl create namespace $NAMESPACE
    echo "✅ 命名空间已创建"
fi
echo ""

echo "3. 检查 Secrets..."
if kubectl get secret pintuotuo-secrets -n $NAMESPACE &> /dev/null; then
    echo "⚠️  Secrets 已存在，跳过创建"
    echo "   如需更新，请手动执行: kubectl delete secret pintuotuo-secrets -n $NAMESPACE"
else
    echo "❌ Secrets 未配置!"
    echo ""
    echo "请先创建 deploy/k8s/secrets.yml 文件，内容如下:"
    echo ""
    cat << 'EOF'
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
EOF
    echo ""
    echo "创建完成后重新运行此脚本"
    exit 1
fi
echo ""

echo "4. 检查 Ingress Controller..."
if kubectl get namespace ingress-nginx &> /dev/null; then
    echo "✅ NGINX Ingress Controller 已安装"
else
    echo "安装 NGINX Ingress Controller..."
    kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.2/deploy/static/provider/cloud/deploy.yaml
    echo "等待 Ingress Controller 就绪..."
    kubectl wait --namespace ingress-nginx \
        --for=condition=ready pod \
        --selector=app.kubernetes.io/component=controller \
        --timeout=120s || true
    echo "✅ NGINX Ingress Controller 安装完成"
fi
echo ""

echo "5. 检查 Cert-Manager..."
if kubectl get namespace cert-manager &> /dev/null; then
    echo "✅ Cert-Manager 已安装"
else
    echo "安装 Cert-Manager..."
    kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.1/cert-manager.yaml
    echo "等待 Cert-Manager 就绪..."
    sleep 30
    kubectl wait --namespace cert-manager \
        --for=condition=ready pod \
        --selector=app.kubernetes.io/component=controller \
        --timeout=120s || true
    
    echo "创建 ClusterIssuer..."
    kubectl apply -f - <<EOF
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@pintuotuo.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
EOF
    echo "✅ Cert-Manager 安装完成"
fi
echo ""

echo "6. 部署应用..."
kubectl apply -f $DEPLOY_DIR/k8s/app.yml
echo "✅ 应用配置已应用"
echo ""

echo "7. 等待部署就绪..."
echo "等待 Backend..."
kubectl rollout status deployment/backend -n $NAMESPACE --timeout=300s
echo "等待 Frontend..."
kubectl rollout status deployment/frontend -n $NAMESPACE --timeout=300s
echo ""

echo "8. 获取部署信息..."
echo ""
echo "=========================================="
echo "  部署状态"
echo "=========================================="
kubectl get pods -n $NAMESPACE
echo ""
echo "=========================================="
echo "  服务信息"
echo "=========================================="
kubectl get svc -n $NAMESPACE
echo ""
echo "=========================================="
echo "  Ingress 信息"
echo "=========================================="
kubectl get ingress -n $NAMESPACE
echo ""

INGRESS_IP=$(kubectl get svc -n ingress-nginx ingress-nginx-controller -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "pending")
if [ "$INGRESS_IP" = "pending" ] || [ -z "$INGRESS_IP" ]; then
    INRESS_IP=$(kubectl get svc -n ingress-nginx ingress-nginx-controller -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null || echo "pending")
fi

echo "=========================================="
echo "  DNS 配置"
echo "=========================================="
echo "请将以下域名指向: $INGRESS_IP"
echo "  - api.pintuotuo.com"
echo "  - www.pintuotuo.com"
echo ""

echo "=========================================="
echo "  健康检查"
echo "=========================================="
echo "部署完成后，请验证:"
echo "  - https://api.pintuotuo.com/api/v1/health"
echo "  - https://www.pintuotuo.com"
echo ""

echo "=========================================="
echo "  部署完成!"
echo "=========================================="
