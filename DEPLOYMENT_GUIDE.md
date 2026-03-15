# Pintuotuo Backend - Deployment Guide

**Project**: B2B2C AI Token Secondary Market Platform
**Status**: Production Ready (Week 7 Complete)
**Last Updated**: 2026-03-15

---

## 📋 Table of Contents

1. [Pre-Deployment Checklist](#pre-deployment-checklist)
2. [Environment Setup](#environment-setup)
3. [Database Initialization](#database-initialization)
4. [Docker Deployment](#docker-deployment)
5. [Kubernetes Deployment](#kubernetes-deployment)
6. [Configuration Management](#configuration-management)
7. [Monitoring & Logging](#monitoring--logging)
8. [Troubleshooting](#troubleshooting)

---

## Pre-Deployment Checklist

### Code Quality ✅
- [x] All 23 integration tests passing (100%)
- [x] Backend compiles without warnings
- [x] Code follows project standards
- [x] Git history clean with meaningful commits
- [x] No hardcoded secrets in code

### Infrastructure ✅
- [x] PostgreSQL 15 configured
- [x] Redis 7 available for caching
- [x] Docker environment prepared
- [x] Network infrastructure ready

### Services ✅
- [x] 7 service layers implemented
- [x] 40+ HTTP endpoints
- [x] ACID transactions verified
- [x] Race conditions tested
- [x] Performance benchmarked

### Documentation ✅
- [x] API endpoints documented
- [x] Database schema documented
- [x] Service architecture clear
- [x] Deployment procedures defined

---

## Environment Setup

### Local Development

```bash
# Clone repository
git clone https://github.com/pintuotuo/pintuotuo.git
cd pintuotuo

# Create environment file
cp .env.example .env.development

# Configure environment
export DATABASE_URL="postgresql://pintuotuo:dev_password_123@localhost:5432/pintuotuo_db?sslmode=disable"
export REDIS_URL="redis://localhost:6379"
export JWT_SECRET="your-secret-key-change-in-production"
export GIN_MODE=debug
export PORT=8080
```

### Staging Environment

```bash
# Create staging configuration
cat > .env.staging << 'CONFIG'
# Database
DATABASE_URL="postgresql://pintuotuo_staging:secure_password@postgres-staging:5432/pintuotuo_staging"

# Cache
REDIS_URL="redis://redis-staging:6379"

# JWT
JWT_SECRET="staging-jwt-secret-key"
JWT_EXPIRE_HOURS=24

# Application
GIN_MODE=release
PORT=8080
APP_ENV=staging
APP_LOG_LEVEL=info
CONFIG
```

### Production Environment

```bash
# Create production configuration
cat > .env.production << 'CONFIG'
# Database (managed service recommended)
DATABASE_URL="postgresql://user:password@rds-instance.region.rds.amazonaws.com:5432/pintuotuo"

# Cache (managed service recommended)
REDIS_URL="redis://elasticache-endpoint:6379"

# JWT
JWT_SECRET="production-jwt-secret-key-use-secrets-manager"
JWT_EXPIRE_HOURS=24

# Application
GIN_MODE=release
PORT=8080
APP_ENV=production
APP_LOG_LEVEL=warn

# Security
CORS_ALLOWED_ORIGINS="https://pintuotuo.com,https://app.pintuotuo.com"
CSRF_PROTECTION=true
SSL_REDIRECT=true
CONFIG
```

---

## Database Initialization

### Schema Creation

```bash
# Connect to PostgreSQL
psql -h localhost -U pintuotuo -d pintuotuo_db

# Initialize schema
\i scripts/db/full_schema.sql

# Verify tables
\dt
```

### Initial Data Setup

```sql
-- Seed merchant data
INSERT INTO products (name, price, merchant_id, stock) VALUES 
  ('Standard Plan', 99.99, 1, 1000),
  ('Pro Plan', 199.99, 1, 500),
  ('Enterprise Plan', 499.99, 1, 100);

-- Verify
SELECT COUNT(*) FROM products;
```

### Backup Strategy

```bash
# Daily backup
pg_dump -h localhost -U pintuotuo pintuotuo_db > backup_$(date +%Y%m%d).sql

# Restore from backup
psql -h localhost -U pintuotuo pintuotuo_db < backup_20260315.sql

# Automated backup (cron job)
0 2 * * * /usr/local/bin/backup-db.sh
```

---

## Docker Deployment

### Build Image

```bash
# Build production image
docker build -t pintuotuo-backend:1.0.0 -f backend/Dockerfile .

# Tag for registry
docker tag pintuotuo-backend:1.0.0 docker-registry.company.com/pintuotuo-backend:1.0.0

# Push to registry
docker push docker-registry.company.com/pintuotuo-backend:1.0.0
```

### Run Container

```bash
# Simple run
docker run -d \
  --name pintuotuo-backend \
  -p 8080:8080 \
  -e DATABASE_URL="postgresql://user:pass@postgres:5432/db" \
  -e REDIS_URL="redis://redis:6379" \
  -e JWT_SECRET="secret" \
  pintuotuo-backend:1.0.0

# Check logs
docker logs -f pintuotuo-backend

# Check health
curl http://localhost:8080/health
```

### Docker Compose for Full Stack

```bash
# Start all services
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f backend

# Stop services
docker-compose down
```

---

## Kubernetes Deployment

### Prerequisites

```bash
# Install kubectl
brew install kubectl

# Install helm (recommended)
brew install helm

# Connect to cluster
kubectl config use-context production-cluster
```

### Create ConfigMap and Secrets

```yaml
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: pintuotuo-config
  namespace: production
data:
  GIN_MODE: "release"
  APP_LOG_LEVEL: "info"
  CORS_ALLOWED_ORIGINS: "https://pintuotuo.com"

---
apiVersion: v1
kind: Secret
metadata:
  name: pintuotuo-secrets
  namespace: production
type: Opaque
stringData:
  DATABASE_URL: "postgresql://user:pass@postgres-service:5432/pintuotuo"
  REDIS_URL: "redis://redis-service:6379"
  JWT_SECRET: "production-secret-key"
```

### Kubernetes Deployment Manifest

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pintuotuo-backend
  namespace: production
spec:
  replicas: 3
  selector:
    matchLabels:
      app: pintuotuo-backend
  template:
    metadata:
      labels:
        app: pintuotuo-backend
    spec:
      containers:
      - name: backend
        image: docker-registry.company.com/pintuotuo-backend:1.0.0
        ports:
        - containerPort: 8080
        envFrom:
        - configMapRef:
            name: pintuotuo-config
        - secretRef:
            name: pintuotuo-secrets
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5

---
apiVersion: v1
kind: Service
metadata:
  name: pintuotuo-backend-service
  namespace: production
spec:
  type: LoadBalancer
  selector:
    app: pintuotuo-backend
  ports:
  - port: 80
    targetPort: 8080
```

### Deploy to Kubernetes

```bash
# Apply manifests
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/secrets.yaml
kubectl apply -f k8s/deployment.yaml

# Verify deployment
kubectl get pods -n production
kubectl logs -n production -l app=pintuotuo-backend -f

# Check service
kubectl get svc -n production

# Update deployment (rolling update)
kubectl set image deployment/pintuotuo-backend \
  backend=docker-registry.company.com/pintuotuo-backend:1.0.1 \
  -n production

# Rollback if needed
kubectl rollout undo deployment/pintuotuo-backend -n production
```

---

## Configuration Management

### Environment Variables

| Variable | Default | Description | Production |
|----------|---------|-------------|-----------|
| DATABASE_URL | localhost:5432 | PostgreSQL connection | RDS endpoint |
| REDIS_URL | localhost:6379 | Redis connection | ElastiCache endpoint |
| JWT_SECRET | dev-secret | JWT signing key | Secrets Manager |
| JWT_EXPIRE_HOURS | 24 | Token expiration | 24 hours |
| GIN_MODE | debug | Gin framework mode | release |
| PORT | 8080 | Server port | 8080 |
| APP_ENV | development | Environment name | production |
| APP_LOG_LEVEL | debug | Logging level | warn |

### Secrets Management

```bash
# Using AWS Secrets Manager
aws secretsmanager create-secret \
  --name pintuotuo/production \
  --secret-string file://secrets.json

# Retrieve secret
aws secretsmanager get-secret-value \
  --secret-id pintuotuo/production

# Using HashiCorp Vault
vault kv put secret/pintuotuo \
  jwt_secret="secret" \
  db_password="password"
```

---

## Monitoring & Logging

### Health Check

```bash
# Manual health check
curl -s http://localhost:8080/health | jq .

# Expected response
{
  "status": "healthy",
  "message": "Pintuotuo Backend Server is running",
  "timestamp": "2026-03-15T19:00:00Z"
}
```

### Logging

```bash
# View application logs
docker logs pintuotuo-backend

# Filter by level
docker logs pintuotuo-backend | grep ERROR

# View logs in Kubernetes
kubectl logs -n production deployment/pintuotuo-backend -f

# Log aggregation (recommended)
# - ELK Stack (Elasticsearch, Logstash, Kibana)
# - CloudWatch (AWS)
# - Stackdriver (Google Cloud)
```

### Metrics

```bash
# Performance metrics (to be added)
# - Request latency
# - Database query performance
# - Cache hit rates
# - Error rates
# - Concurrent connections

# Recommended tools:
# - Prometheus for metrics collection
# - Grafana for visualization
# - DataDog for monitoring
```

---

## Troubleshooting

### Connection Issues

**Problem**: Cannot connect to database
```bash
# Check PostgreSQL is running
pg_isready -h localhost -p 5432

# Check connection string
echo $DATABASE_URL

# Test connection manually
psql $DATABASE_URL -c "SELECT 1"
```

**Problem**: Redis connection failed
```bash
# Check Redis is running
redis-cli ping

# Check Redis URL
echo $REDIS_URL

# Test connection
redis-cli -u $REDIS_URL ping
```

### Application Issues

**Problem**: Server won't start
```bash
# Check logs
docker logs pintuotuo-backend

# Verify all environment variables
env | grep PINTUOTUO

# Check port availability
lsof -i :8080
```

**Problem**: Slow requests
```bash
# Check database connection pool
# - MaxOpenConns: 25
# - MaxIdleConns: 5

# Monitor slow queries
# Add query logging to PostgreSQL

# Check Redis cache hit rates
redis-cli info stats
```

### Deployment Issues

**Problem**: Kubernetes pod not starting
```bash
# Check pod status
kubectl describe pod <pod-name> -n production

# Check logs
kubectl logs <pod-name> -n production

# Check resource limits
kubectl top node
kubectl top pod

# Check events
kubectl get events -n production --sort-by='.lastTimestamp'
```

---

## Performance Optimization

### Database Optimization

```sql
-- Analyze query performance
EXPLAIN ANALYZE SELECT * FROM orders WHERE user_id = 1;

-- Create indexes
CREATE INDEX idx_orders_user_id ON orders(user_id);

-- Vacuum and analyze
VACUUM ANALYZE;
```

### Caching Strategy

```go
// Configure cache TTL
- Token balance: 5 minutes
- User profile: 10 minutes
- Product catalog: 15 minutes
- Payment list: 2 minutes

// Cache invalidation on write
- Update cache on PUT/POST/DELETE
- Use pattern-based invalidation
```

### Concurrency

```bash
# Test with Apache Bench
ab -n 10000 -c 100 http://localhost:8080/health

# Test with wrk
wrk -t4 -c100 -d30s http://localhost:8080/health

# Test with hey
hey -n 10000 -c 100 http://localhost:8080/health
```

---

## Maintenance

### Regular Tasks

- Daily: Monitor logs and metrics
- Weekly: Backup database
- Monthly: Review performance metrics
- Quarterly: Security audit
- Annually: Load testing and capacity planning

### Updates

```bash
# Create new release
git tag -a v1.0.0 -m "Release 1.0.0"
git push origin v1.0.0

# Build new image
docker build -t pintuotuo-backend:1.0.0 .

# Deploy to staging
docker-compose -f docker-compose.staging.yml up -d

# Run integration tests
go test -v ./tests/integration

# Deploy to production
kubectl rolling-update pintuotuo-backend \
  --image=pintuotuo-backend:1.0.0
```

---

## Support & Documentation

- **Architecture**: See `05_Technical_Architecture_and_Tech_Stack.md`
- **API Spec**: See `04_API_Specification.md`
- **Code Standards**: See `13_Dev_Git_Workflow_Code_Standards.md`
- **Setup Guide**: See `12_Dev_Setup_Environment_Configuration.md`

---

**Status**: Ready for Production Deployment ✅
**Last Tested**: 2026-03-15 (All 23 integration tests passing)
**Next Review**: 2026-03-22

