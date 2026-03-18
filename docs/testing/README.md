# 测试总览与规范

## 目录结构
- 后端单元测试：与源码就近存放，遵循 *_test.go
  - 例如：
    - services/order：[service_test.go](file:///Users/4seven/pintuotuo/backend/services/order/service_test.go)
    - services/group：[service_test.go](file:///Users/4seven/pintuotuo/backend/services/group/service_test.go)
    - handlers：[handlers_test.go](file:///Users/4seven/pintuotuo/backend/handlers/handlers_test.go)
- 后端集成测试：集中在专用目录
  - [backend/tests/integration](file:///Users/4seven/pintuotuo/backend/tests/integration)
    - [workflow_test.go](file:///Users/4seven/pintuotuo/backend/tests/integration/workflow_test.go)
    - [consistency_test.go](file:///Users/4seven/pintuotuo/backend/tests/integration/consistency_test.go)
    - [helpers.go](file:///Users/4seven/pintuotuo/backend/tests/integration/helpers.go)
- 前端测试：位于前端源码 src 下，使用 *.test.tsx / *.test.ts
  - 配置：[jest.config.cjs](file:///Users/4seven/pintuotuo/frontend/jest.config.cjs)
  - 初始化：[setup-tests.ts](file:///Users/4seven/pintuotuo/frontend/src/setup-tests.ts)
- 数据库初始化脚本：
  - [full_schema.sql](file:///Users/4seven/pintuotuo/scripts/db/full_schema.sql)
- 本地沙箱测试与 CI 等效运行脚本：
  - [run_local_sandbox_tests.sh](file:///Users/4seven/pintuotuo/scripts/testing/run_local_sandbox_tests.sh)
  - [run_ci_equivalent.sh](file:///Users/4seven/pintuotuo/scripts/testing/run_ci_equivalent.sh)

## 环境变量与约定
- TEST_MODE=true：启用测试数据的幂等清理与预置
- 数据库/缓存：
  - DATABASE_URL=postgresql://pintuotuo:dev_password_123@localhost:5433/pintuotuo_db?sslmode=disable
  - REDIS_URL=redis://localhost:6380
- 其他：
  - JWT_SECRET=pintuotuo-secret-key-dev
  - GIN_MODE=release

## 本地运行（沙箱）
1) 启动测试容器
```
docker-compose -f docker-compose.test.yml up -d
```
2) 后端单元 + 集成测试
```
cd backend
TEST_MODE=true \
DATABASE_URL=postgresql://pintuotuo:dev_password_123@localhost:5433/pintuotuo_db?sslmode=disable \
REDIS_URL=redis://localhost:6380 \
JWT_SECRET=pintuotuo-secret-key-dev \
GIN_MODE=release \
go test -v -count=1 -p 1 ./...
go test -v -count=1 -p 1 ./tests/integration
```
3) 前端测试（如有）
```
cd frontend
npm ci
CI=true npm test -- --watchAll=false
```
4) 清理
```
docker-compose -f docker-compose.test.yml down
```

## CI 等效运行（本地模拟）
- 统一使用 -count=1 禁用测试缓存、-p 1 串行
- 通过 scripts/testing/run_ci_equivalent.sh 复用上述变量与步骤

## 常见问题
- 端口冲突：释放 8080/3000 或修改监听端口
- 测试并发导致的竞态：保持 -p 1；确保 [database.go](file:///Users/4seven/pintuotuo/backend/config/database.go) 的 TruncateAndSeed 使用 sync.Once
- 前端 TypeScript 类型问题：遵循路径别名与 APIResponse<T> 的 data 结构，参考 FRONTEND_TESTING.md

