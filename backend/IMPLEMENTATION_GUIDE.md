# 后端实现完成指南 - Week 2

**目标**: 完善和优化 API 处理程序，达到生产级质量

---

## 📋 检查清单

### 用户认证 (handlers/auth.go)

- [x] **注册** ✅
  - 邮箱重复检查
  - 密码哈希
  - Token 余额初始化

- [x] **登录** ✅
  - 邮箱和密码验证
  - JWT Token 生成

- [ ] **完善项目**:
  - [ ] 添加邮箱验证（可选）
  - [ ] 密码强度验证
  - [ ] Token 刷新端点
  - [ ] 密码重置功能

### 产品管理 (handlers/product.go)

- [x] **列表** ✅
  - 分页支持
  - 状态过滤

- [x] **搜索** ✅
  - ILIKE 全文搜索

- [ ] **完善项目**:
  - [ ] 添加排序功能（价格、创建时间）
  - [ ] 图片 URL 支持
  - [ ] 分类/标签支持
  - [ ] 商户评分显示

### 订单系统 (handlers/order_and_group.go)

- [x] **创建** ✅
  - 库存检查

- [x] **取消** ✅
  - 状态验证

- [ ] **完善项目**:
  - [ ] 订单超时自动取消
  - [ ] 库存扣减和恢复
  - [ ] 订单号生成优化
  - [ ] 物流追踪字段

### 分组购买 (handlers/order_and_group.go)

- [x] **创建** ✅
  - 截止日期验证

- [x] **加入** ✅
  - 自动成团
  - 人数上限检查

- [ ] **完善项目**:
  - [ ] 自动退款（成团失败）
  - [ ] 分组历史记录
  - [ ] 成团通知
  - [ ] 自动匹配建议

### 支付处理 (handlers/payment_and_token.go)

- [x] **发起** ✅
- [x] **Webhook** ✅
  - Alipay 回调
  - WeChat 回调

- [ ] **完善项目**:
  - [ ] 支付超时处理
  - [ ] 并发支付检查
  - [ ] 对账系统
  - [ ] 交易日志完整

### Token 管理 (handlers/payment_and_token.go, apikey.go)

- [x] **查询** ✅
- [x] **转账** ✅
- [x] **API 密钥** ✅

- [ ] **完善项目**:
  - [ ] Token 过期时间
  - [ ] 转账限额
  - [ ] API 配额管理
  - [ ] 使用统计

---

## 🔧 实现优化方案

### 1️⃣ 缓存层集成 (Redis)

**当前**: 直接数据库查询
**目标**: 热数据缓存

```go
// 缓存常访问的产品
const productCacheTTL = 1 * time.Hour

func GetProductByID(c *gin.Context) {
    id := c.Param("id")

    // 先查缓存
    cacheKey := fmt.Sprintf("product:%s", id)
    if cached, _ := redisClient.Get(cacheKey).Result(); cached != "" {
        // 返回缓存数据
    }

    // 查询数据库
    product := fetchFromDB(id)

    // 存入缓存
    redisClient.Set(cacheKey, product, productCacheTTL)

    c.JSON(200, product)
}
```

**缓存策略**:
- 产品列表: 5分钟
- 产品详情: 1小时
- 用户信息: 30分钟
- 分组信息: 实时

### 2️⃣ 数据库查询优化

**添加关键索引**:

```sql
-- 已存在的索引
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_products_merchant ON products(merchant_id);
CREATE INDEX idx_products_status ON products(status);
CREATE INDEX idx_orders_user ON orders(user_id);
CREATE INDEX idx_orders_created ON orders(created_at DESC);

-- 需要添加的索引
CREATE INDEX idx_groups_product ON groups(product_id);
CREATE INDEX idx_groups_status ON groups(status);
CREATE INDEX idx_payments_order ON payments(order_id);
CREATE INDEX idx_group_members_group ON group_members(group_id);
```

**查询优化**:
- 使用分页避免大数据集
- 利用索引进行排序
- 避免 N+1 查询问题

```go
// 不好: N+1 查询
for _, order := range orders {
    product := db.Query("SELECT * FROM products WHERE id = ?", order.ProductID)
}

// 好: 一次查询
products := db.Query("SELECT * FROM products WHERE id IN (...)")
productMap := makeMap(products)
for _, order := range orders {
    product := productMap[order.ProductID]
}
```

### 3️⃣ 错误处理统一

当前: 每个处理器独立处理
目标: 统一的错误响应

```go
// errors/errors.go
type AppError struct {
    Code    string      // 错误代码
    Message string      // 用户消息
    Status  int         // HTTP 状态码
    Internal error      // 内部错误
}

var (
    ErrUserNotFound = AppError{
        Code: "USER_NOT_FOUND",
        Message: "用户不存在",
        Status: 404,
    }
    ErrInsufficientStock = AppError{
        Code: "INSUFFICIENT_STOCK",
        Message: "库存不足",
        Status: 409,
    }
)

// 在处理器中使用
if product.Stock < req.Quantity {
    return c.JSON(ErrInsufficientStock.Status, ErrInsufficientStock)
}
```

### 4️⃣ 日志和监控

当前: 基础日志中间件
目标: 结构化日志和指标

```go
// 添加结构化日志
type RequestLog struct {
    Timestamp    time.Time
    Method       string
    Path         string
    Status       int
    Duration     time.Duration
    UserID       int
    RequestID    string
    Error        string `json:",omitempty"`
}

// Prometheus 指标
var (
    requestCounter = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
        },
        []string{"method", "endpoint", "status"},
    )

    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
        },
        []string{"method", "endpoint"},
    )
)
```

### 5️⃣ 事务和数据一致性

**当前问题**: 支付成功但订单未更新
**解决方案**: 使用事务

```go
func HandlePaymentCallback(c *gin.Context) {
    var req PaymentCallback

    tx := db.BeginTx(context.Background(), nil)
    defer tx.Rollback()

    // 更新支付状态
    tx.Exec("UPDATE payments SET status = ? WHERE id = ?", "success", req.PaymentID)

    // 更新订单状态
    tx.Exec("UPDATE orders SET status = ? WHERE id = (...)", "paid")

    // 扣减库存
    tx.Exec("UPDATE products SET stock = stock - ? WHERE id = (...)", quantity)

    // 更新分组
    tx.Exec("UPDATE groups SET current_count = current_count + 1 WHERE id = (...)")

    // 提交事务
    if err := tx.Commit(); err != nil {
        return c.JSON(500, gin.H{"error": "Transaction failed"})
    }

    c.JSON(200, gin.H{"message": "Success"})
}
```

---

## 🧪 测试完成度目标

### 单元测试

目标: **>80% 代码覆盖率**

```bash
# 运行测试并生成覆盖率报告
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

**必须测试的场景**:

1. **认证**:
   - [x] 正常注册
   - [x] 邮箱重复
   - [x] 密码过短
   - [x] 正常登录
   - [ ] 错误密码
   - [ ] 邮箱不存在

2. **产品**:
   - [x] 列表查询
   - [x] 搜索功能
   - [ ] 创建验证
   - [ ] 权限检查（只有商户可创建）

3. **订单**:
   - [x] 创建订单
   - [x] 库存检查
   - [ ] 取消订单
   - [ ] 状态验证

4. **分组**:
   - [x] 创建分组
   - [x] 加入分组
   - [ ] 自动成团
   - [ ] 过期检查

### 集成测试

```go
// 测试完整流程
func TestCompleteOrderFlow(t *testing.T) {
    // 1. 用户注册
    user := registerUser(...)

    // 2. 获取产品
    product := getProduct(...)

    // 3. 创建订单
    order := createOrder(user.ID, product.ID, 1)

    // 4. 发起支付
    payment := initiatePayment(order.ID, "alipay")

    // 5. 处理回调
    handlePaymentCallback(payment.ID, "success")

    // 6. 验证最终状态
    assert.Equal(t, order.Status, "paid")
}
```

---

## 📊 性能目标

| 指标 | 目标 | 当前 |
|------|------|------|
| API 响应时间 | <200ms | 待测试 |
| 数据库查询 | <50ms | 待优化 |
| 缓存命中率 | >70% | 待部署 |
| 错误率 | <0.1% | 待监测 |
| 并发连接 | >1000 | 待测试 |

**性能测试命令**:

```bash
# 使用 Apache Bench
ab -n 10000 -c 100 http://localhost:8080/api/v1/products

# 使用 wrk
wrk -t12 -c400 -d30s http://localhost:8080/api/v1/products
```

---

## 📝 Code Review 检查清单

在推送前检查:

- [ ] 所有错误都被处理
- [ ] SQL 查询使用参数化（防注入）
- [ ] 没有硬编码的密钥
- [ ] 适当的日志记录
- [ ] 函数长度 <50 行
- [ ] 有单元测试覆盖
- [ ] 通过 `golangci-lint`
- [ ] 代码格式化 `go fmt`

---

## 🚀 Week 2 里程碑

**Monday-Tuesday**:
- [ ] 完成所有缺失的单元测试
- [ ] 集成 Redis 缓存
- [ ] 优化数据库查询

**Wednesday-Thursday**:
- [ ] 前后端集成测试
- [ ] 修复发现的 bug
- [ ] 性能测试和优化

**Friday**:
- [ ] 代码审查完成
- [ ] 文档更新
- [ ] 演示和验收

---

## 🔗 相关文件

- `API_DOCUMENTATION.md` - API 端点详细文档
- `handlers/*.go` - 所有处理器实现
- `CLAUDE.md` - 编码标准
- `CODE_REVIEW.md` - 代码审查标准

---

**编写者**: Engineering Team
**日期**: 2026-03-14
**状态**: 实现进行中 🚀
