# 开发指南 - Pintuotuo 项目

## 快速开始（5分钟）

### 1. 环境准备

**系统要求：**
- Go 1.21+
- Node.js v18+
- Docker & Docker Compose
- Make（可选，但推荐）

**Go 模块缓存（若在 Cursor/沙箱中 `go test` 报大量「no such file or directory」）：**  
项目根目录的 `make test` / `make test-backend` 已通过 `scripts/ensure-go-modcache.sh` 将 `GOMODCACHE` 指到稳定目录（默认 `$HOME/go/pkg/mod`）。若仍异常，可执行 `go clean -modcache` 后重试，或在 shell 中手动 `export GOMODCACHE="$HOME/go/pkg/mod"`。

### 2. 初始化项目

```bash
# 自动初始化（推荐）
bash scripts/setup.sh

# 或手动步骤
docker-compose up -d          # 启动数据库和Redis
cd backend && go mod download # 安装Go依赖
cd frontend && npm install    # 安装Node依赖
```

### 3. 启动应用

```bash
# 方式1：使用Makefile（推荐）
make dev-backend    # 在终端1启动后端（8080端口）
make dev-frontend   # 在终端2启动前端（5173端口）

# 方式2：手动启动
cd backend && go run main.go
cd frontend && npm run dev
```

### 4. 访问应用

- **前端应用**：http://localhost:5173
- **后端API**：http://localhost:8080/api/v1
- **健康检查**：http://localhost:8080/health

## 常用命令

```bash
# 数据库迁移
make migrate                  # 运行所有迁移
make db-shell                 # 连接到数据库
make db-reset                 # 重置数据库（警告：删除所有数据）

# 测试和质量
make test                     # 运行所有测试
make test-backend             # 仅后端测试
make test-frontend            # 仅前端测试
make lint                      # 代码检查
make format                    # 代码格式化

# 构建
make build                     # 构建前后端
make build-backend             # 仅构建后端
make build-frontend            # 仅构建前端

# Docker管理
make docker-up                 # 启动容器
make docker-down               # 停止容器
make docker-logs               # 查看日志

# 清理
make clean                     # 清理build artifacts
```

## 项目结构

```
pintuotuo/
├── backend/                  # Go后端应用
│   ├── main.go              # 服务器入口点
│   ├── config/              # 配置管理
│   ├── models/              # 数据模型
│   ├── handlers/            # HTTP处理程序
│   ├── middleware/          # 中间件
│   ├── routes/              # 路由定义
│   ├── migrations/          # 数据库迁移
│   └── cmd/                 # CLI工具
│
├── frontend/                # React前端应用
│   ├── src/
│   │   ├── components/      # 可复用组件
│   │   ├── pages/           # 页面组件
│   │   ├── services/        # API服务
│   │   ├── stores/          # Zustand状态管理
│   │   ├── hooks/           # 自定义hooks
│   │   ├── types/           # TypeScript类型
│   │   └── utils/           # 工具函数
│   └── public/              # 静态资源
│
├── scripts/                 # 脚本和工具
├── docker-compose.yml       # Docker容器配置
└── Makefile                 # 开发命令
```

## 环境变量

创建 `.env.development` 文件（基于 `.env.example`）：

```bash
cp .env.example .env.development
# 编辑.env.development根据需要
```

常用变量：
- `PORT` - 后端服务器端口（默认：8080）
- `DATABASE_URL` - PostgreSQL连接字符串
- `JWT_SECRET` - JWT签名密钥
- `GIN_MODE` - Gin框架模式（debug/release）

## 数据库迁移

### 运行迁移

```bash
make migrate
```

### 创建新迁移

```bash
# 自动创建新迁移文件
make migrate-create name=add_user_columns
```

## 认证和授权

### 获取认证令牌

1. **注册用户：**
```bash
curl -X POST http://localhost:8080/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "name": "User Name",
    "password": "password123"
  }'
```

2. **登录获取令牌：**
```bash
curl -X POST http://localhost:8080/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

3. **使用令牌调用API：**
```bash
curl -X GET http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer <your-token>"
```

## API文档

### 主要端点

**用户管理：**
- `POST /users/register` - 用户注册
- `POST /users/login` - 用户登录
- `GET /users/me` - 获取当前用户
- `PUT /users/me` - 更新用户信息

**产品：**
- `GET /products` - 获取产品列表
- `GET /products/:id` - 获取产品详情
- `GET /products/search` - 搜索产品

**订单：**
- `POST /orders` - 创建订单
- `GET /orders` - 获取订单列表
- `GET /orders/:id` - 获取订单详情
- `PUT /orders/:id/cancel` - 取消订单

**分组购买：**
- `POST /groups` - 创建分组
- `GET /groups` - 获取分组列表
- `POST /groups/:id/join` - 加入分组

**支付：**
- `POST /payments` - 发起支付
- `GET /payments/:id` - 获取支付状态

**Token：**
- `GET /tokens/balance` - 获取余额
- `POST /tokens/transfer` - 转账

## 测试

### 运行测试

```bash
# 所有测试
make test

# 后端测试（含覆盖率）
make test-backend

# 前端测试
make test-frontend
```

### 测试覆盖率目标

- **后端**：>80% 代码覆盖率
- **前端**：>70% 覆盖率

### BYOK路由模式测试

BYOK路由模式功能包含完整的测试覆盖：

**后端单元测试**：
- `backend/services/api_key_validator_route_mode_test.go` - 路由模式解析测试
- `backend/services/api_key_validator_error_mapping_test.go` - 错误映射测试
- `backend/models/models_byok_test.go` - 模型方法测试

**后端集成测试**：
- `backend/handlers/admin_byok_routing_test.go` - API端点测试

**前端E2E测试**：
- `frontend/e2e/byok-routing.spec.ts` - 端到端流程测试

运行BYOK相关测试：
```bash
# 后端路由模式测试
cd backend && go test ./services -run "TestResolveEndpointByRouteMode" -v

# 后端集成测试
cd backend && go test ./handlers -run "TestBYOK" -v

# 前端E2E测试
cd frontend && npx playwright test byok-routing.spec.ts
```

## BYOK路由模式开发指南

### 架构概述

BYOK路由模式支持四种路由方式：Direct、LiteLLM、Proxy、Auto。核心实现位于：

- **模型层**：`backend/models/models.go` - MerchantAPIKey模型
- **服务层**：`backend/services/api_key_validator.go` - 验证和路由逻辑
- **健康检查**：`backend/services/health_checker.go` - 健康检查逻辑
- **API层**：`backend/handlers/admin_byok_routing.go` - 管理接口
- **前端**：`frontend/src/pages/admin/AdminByokRouting.tsx` - 管理界面

### 路由模式实现

#### 1. Direct模式

直接访问上游API，性能最优。

**实现文件**：
- `backend/services/api_key_validator.go` - `resolveDirectEndpoint()`
- `backend/services/health_checker.go` - `resolveDirectEndpoint()`

**Endpoint优先级**：
1. `route_config.endpoint_url`
2. `route_config.endpoints.direct.{region}`
3. `model_providers.api_base_url`

**开发要点**：
- 支持region选择（domestic/overseas）
- 默认使用overseas region
- 错误处理使用`MapProviderError`

#### 2. LiteLLM模式

通过LiteLLM网关访问，支持负载均衡。

**实现文件**：
- `backend/services/api_key_validator.go` - `resolveLitellmEndpoint()`
- `backend/services/health_checker.go` - `resolveLitellmEndpoint()`

**Endpoint优先级**：
1. `route_config.endpoints.litellm.{region}`
2. `route_config.base_url`
3. 环境变量 `LLM_GATEWAY_LITELLM_URL`

**认证Token处理**：
- 使用环境变量 `LITELLM_MASTER_KEY`
- 如果没有master key，使用原始API key

**开发要点**：
- 默认使用domestic region
- 需要配置LiteLLM环境变量
- 支持master key认证

#### 3. Proxy模式

通过代理服务器访问，适用于网络限制场景。

**实现文件**：
- `backend/services/api_key_validator.go` - `resolveProxyEndpoint()`
- `backend/services/health_checker.go` - `resolveProxyEndpoint()`

**Endpoint优先级**：
1. `route_config.endpoints.proxy.{type}` (优先gaap)
2. `route_config.proxy_url`

**开发要点**：
- 支持多种代理类型（gaap、cdn等）
- 优先使用gaap类型
- 需要配置代理服务器地址

#### 4. Auto模式

根据系统配置自动选择最佳路由。

**实现文件**：
- `backend/services/api_key_validator.go` - `resolveAutoEndpoint()`
- `backend/services/health_checker.go` - `resolveAutoEndpoint()`

**降级策略**：
- Direct → LiteLLM → Proxy
- 按顺序尝试，直到找到可用的endpoint

**开发要点**：
- 实现endpoint可用性检查
- 支持自定义降级顺序
- 记录降级决策日志

### 错误处理规范

所有路由模式必须使用统一的错误处理机制：

**错误映射函数**：`MapProviderError()` in `backend/services/provider_error_mapper.go`

**错误分类**：
- `AUTH_INVALID_KEY` - API密钥无效
- `AUTH_PERMISSION_DENIED` - 权限被拒绝
- `QUOTA_INSUFFICIENT` - 配额不足
- `RATE_LIMITED` - 速率限制
- `MODEL_NOT_FOUND` - 模型不存在
- `SERVICE_UNAVAILABLE` - 服务不可用
- `NETWORK_TIMEOUT` - 网络超时
- `UNKNOWN` - 未知错误

**重试机制**：
- 使用`ProviderErrorInfo.Retryable`字段判断是否重试
- 最大重试次数：3次
- 重试间隔：指数退避（1s, 2s, 4s）

### 数据库迁移

BYOK路由模式相关的数据库迁移：

**迁移文件**：
- `backend/migrations/074_byok_routing_ssot.sql` - 基础字段
- `backend/migrations/075_byok_route_mode_support.sql` - 验证表扩展

**关键字段**：
- `merchant_api_keys.route_mode` - 路由模式
- `merchant_api_keys.route_config` - 路由配置（JSONB）
- `api_key_verifications.route_mode` - 验证时使用的路由模式
- `api_key_verifications.endpoint_used` - 验证时使用的endpoint
- `api_key_verifications.error_category` - 错误分类

### 环境变量配置

**必需的环境变量**：
```bash
# LiteLLM配置
LLM_GATEWAY_LITELLM_URL=http://litellm:4000
LITELLM_MASTER_KEY=sk-litellm-master-key
```

**配置位置**：
- `.env.example` - 示例配置
- `docker-compose.yml` - Docker环境配置
- `docker-compose.prod.yml` - 生产环境配置

### 前端开发

**关键组件**：
- `frontend/src/pages/admin/AdminByokRouting.tsx` - BYOK路由管理页面
- `frontend/src/services/adminByokRouting.ts` - API服务层
- `frontend/src/utils/byokRouteMode.ts` - 辅助函数

**功能实现**：
1. 路由模式显示和编辑
2. 验证结果展示（轻验、深验）
3. 健康状态指示器（红黄绿灯）
4. 立即探测功能
5. 错误信息展示

**状态指示器**：
- 🟢 绿色：健康/验证成功
- 🟡 黄色：警告/部分功能受限
- 🔴 红色：不健康/验证失败
- ⚪ 灰色：未知/未验证

### 调试技巧

**后端调试**：
```bash
# 查看验证日志
tail -f logs/verification.log | grep "route_mode"

# 测试特定路由模式
curl -X POST http://localhost:8080/api/v1/admin/byok-routing/1/light-verify \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json"
```

**前端调试**：
```bash
# 查看网络请求
# 浏览器开发者工具 -> Network -> Filter: byok

# 查看组件状态
# React DevTools -> AdminByokRouting组件
```

**数据库查询**：
```sql
-- 查看BYOK路由配置
SELECT id, merchant_id, provider, route_mode, route_config 
FROM merchant_api_keys 
WHERE route_mode IS NOT NULL;

-- 查看验证结果
SELECT id, api_key_id, route_mode, endpoint_used, error_category, created_at
FROM api_key_verifications
ORDER BY created_at DESC
LIMIT 10;
```

## 代码标准

参考 [CLAUDE.md](../CLAUDE.md) 了解详细的代码标准和约定。

### 快速检查

```bash
# 格式化代码
make format

# 运行linter
make lint
```

## 构建和部署

### 构建

```bash
# 构建所有
make build

# 构建Docker镜像（待实现）
docker build -t pintuotuo-backend:latest ./backend
docker build -t pintuotuo-frontend:latest ./frontend
```

### 部署

详见 [05_Technical_Architecture_and_Tech_Stack.md](../05_Technical_Architecture_and_Tech_Stack.md)

## 常见问题

### 问题：PostgreSQL连接失败

**解决方案：**
```bash
# 检查容器状态
docker-compose ps

# 查看日志
docker-compose logs postgres

# 重启PostgreSQL
docker-compose restart postgres
```

### 问题：端口已被占用

**解决方案：**
```bash
# 修改 docker-compose.yml 中的端口映射
# 或
# 杀死占用该端口的进程
lsof -i :8080  # 查找占用8080的进程
kill -9 <PID>   # 杀死进程
```

### 问题：Go依赖问题

**解决方案：**
```bash
cd backend
go clean -modcache
go mod download
```

## 性能监控

### 本地开发

```bash
# 后端性能分析
cd backend && go test -bench=. -benchmem

# 前端构建分析
cd frontend && npm run build -- --analyze
```

## 资源

- [docs/DEV_ENTITLEMENTS.md](docs/DEV_ENTITLEMENTS.md) - 权益白名单开关、表字段与本地造数
- [docs/PRODUCT_ENTITLEMENTS.md](docs/PRODUCT_ENTITLEMENTS.md) - 产品侧：严格白名单语义与用户可见错误
- [backend/doc_internal_token_economics.md](backend/doc_internal_token_economics.md) - 内部经济（余额 / 扣费 / 价目）；任务清单见 [.trae/plans/development-plan-v1.md](.trae/plans/development-plan-v1.md) **附录 A**
- [CLAUDE.md](../CLAUDE.md) - 项目开发指南
- [Gin框架文档](https://gin-gonic.com/)
- [React文档](https://react.dev/)
- [Zustand文档](https://github.com/pmndrs/zustand)
- [PostgreSQL文档](https://www.postgresql.org/docs/)

## 获取帮助

```bash
# 查看所有Makefile命令
make help

# 查看后端健康状态
curl http://localhost:8080/health

# 查看docker容器日志
docker-compose logs <service-name>
```

## 提交代码

参考 [13_Dev_Git_Workflow_Code_Standards.md](../13_Dev_Git_Workflow_Code_Standards.md) 了解Git工作流和提交规范。

```bash
# 创建新分支
git checkout -b feature/your-feature

# 提交代码
git add .
git commit -m "feat: Your feature description"

# 推送
git push origin feature/your-feature
```

---

**最后更新**：2026-04-10
**状态**：开发中 🚀
