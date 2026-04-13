# Pintuotuo Backend API Documentation

> 拼脱脱 B2B2C AI Token 二级市场 REST API 文档

**Base URL**: `http://localhost:8080/api/v1`
**版本**: v1.0.0
**最后更新**: 2026-03-14

---

## 📋 目录

1. [认证](#认证)
2. [用户管理](#用户管理)
3. [产品管理](#产品管理)
4. [订单系统](#订单系统)
5. [分组购买](#分组购买)
6. [支付处理](#支付处理)
7. [Token管理](#token管理)
8. [错误处理](#错误处理)

---

## 🔐 认证

所有受保护的端点都需要在请求头中包含 JWT Token：

```
Authorization: Bearer <your_token_here>
```

### 获取 Token

使用用户注册或登录后获得的 Token。

---

## 👤 用户管理

### 注册用户

```
POST /users/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123",
  "role": "user"
}
```

`name` 可选；省略时使用邮箱 `@` 前的本地部分作为展示名。`role` 可选，默认 `user`，商户注册为 `merchant`。

**响应** (201):
```json
{
  "user": {
    "id": 1,
    "email": "user@example.com",
    "name": "User Name",
    "role": "user",
    "created_at": "2026-03-14T10:00:00Z",
    "updated_at": "2026-03-14T10:00:00Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

### 用户登录

```
POST /users/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

**响应** (200):
```json
{
  "user": {
    "id": 1,
    "email": "user@example.com",
    "name": "User Name",
    "role": "user",
    "created_at": "2026-03-14T10:00:00Z",
    "updated_at": "2026-03-14T10:00:00Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

### 获取当前用户

```
GET /users/me
Authorization: Bearer <token>
```

**响应** (200):
```json
{
  "id": 1,
  "email": "user@example.com",
  "name": "User Name",
  "role": "user",
  "created_at": "2026-03-14T10:00:00Z",
  "updated_at": "2026-03-14T10:00:00Z"
}
```

### 更新当前用户

```
PUT /users/me
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "New Name"
}
```

**响应** (200): 更新后的用户对象

### 获取用户信息

```
GET /users/:id
```

**响应** (200): 用户对象

### 更新用户信息（管理员）

```
PUT /users/:id
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "name": "New Name",
  "role": "merchant"
}
```

**响应** (200): 更新后的用户对象

---

## 📦 产品管理

### 获取产品列表

```
GET /products?page=1&per_page=20&status=active
```

**参数**:
- `page` (可选): 页码，默认 1
- `per_page` (可选): 每页数量，默认 20
- `status` (可选): 产品状态 (active/inactive/archived)

**响应** (200):
```json
{
  "total": 100,
  "page": 1,
  "per_page": 20,
  "data": [
    {
      "id": 1,
      "merchant_id": 1,
      "name": "产品名称",
      "description": "产品描述",
      "price": 99.99,
      "stock": 100,
      "status": "active",
      "created_at": "2026-03-14T10:00:00Z",
      "updated_at": "2026-03-14T10:00:00Z"
    }
  ]
}
```

### 搜索产品

```
GET /products/search?q=keyword&page=1&per_page=20
```

**参数**:
- `q` (必需): 搜索关键词
- `page` (可选): 页码
- `per_page` (可选): 每页数量

**响应** (200): 与列表相同格式

### 获取产品详情

```
GET /products/:id
```

**响应** (200): 单个产品对象

### 创建产品（商户）

```
POST /products/merchants
Authorization: Bearer <merchant_token>
Content-Type: application/json

{
  "name": "产品名称",
  "description": "产品描述",
  "price": 99.99,
  "original_price": 129.99,
  "stock": 100
}
```

**响应** (201): 创建的产品对象

### 更新产品（商户）

```
PUT /products/merchants/:id
Authorization: Bearer <merchant_token>
Content-Type: application/json

{
  "name": "新名称",
  "price": 89.99,
  "stock": 150,
  "status": "active"
}
```

**响应** (200): 更新后的产品对象

### 删除产品（商户）

```
DELETE /products/merchants/:id
Authorization: Bearer <merchant_token>
```

**响应** (200):
```json
{
  "message": "Product deleted successfully"
}
```

---

## 📋 订单系统

### 创建订单

```
POST /orders
Authorization: Bearer <token>
Content-Type: application/json

{
  "product_id": 1,
  "group_id": 5,
  "quantity": 2
}
```

**响应** (201): 订单对象

### 获取订单列表

```
GET /orders?page=1&per_page=20
Authorization: Bearer <token>
```

**响应** (200):
```json
{
  "total": 10,
  "page": 1,
  "per_page": 20,
  "data": [
    {
      "id": 1,
      "user_id": 1,
      "product_id": 1,
      "group_id": null,
      "quantity": 2,
      "total_price": 199.98,
      "status": "pending",
      "created_at": "2026-03-14T10:00:00Z",
      "updated_at": "2026-03-14T10:00:00Z"
    }
  ]
}
```

### 获取订单详情

```
GET /orders/:id
Authorization: Bearer <token>
```

**响应** (200): 订单对象

### 取消订单

```
PUT /orders/:id/cancel
Authorization: Bearer <token>
```

**响应** (200): 更新后的订单对象

---

## 🤝 分组购买

### 创建分组

```
POST /groups
Authorization: Bearer <token>
Content-Type: application/json

{
  "product_id": 1,
  "target_count": 5,
  "deadline": "2026-03-21T23:59:59Z"
}
```

**响应** (201): 分组对象

### 获取分组列表

```
GET /groups?page=1&per_page=20&status=active
```

**参数**:
- `status` (可选): active/completed/failed

**响应** (200):
```json
{
  "total": 10,
  "page": 1,
  "per_page": 20,
  "data": [
    {
      "id": 1,
      "product_id": 1,
      "creator_id": 1,
      "target_count": 5,
      "current_count": 3,
      "status": "active",
      "deadline": "2026-03-21T23:59:59Z",
      "created_at": "2026-03-14T10:00:00Z",
      "updated_at": "2026-03-14T10:00:00Z"
    }
  ]
}
```

### 获取分组详情

```
GET /groups/:id
```

**响应** (200): 分组对象

### 加入分组

```
POST /groups/:id/join
Authorization: Bearer <token>
```

**响应** (200): 更新后的分组对象

### 取消分组（创建者）

```
DELETE /groups/:id
Authorization: Bearer <creator_token>
```

**响应** (200):
```json
{
  "message": "Group cancelled successfully"
}
```

### 获取分组进度

```
GET /groups/:id/progress
```

**响应** (200): 分组对象

---

## 💳 支付处理

### 发起支付

```
POST /payments
Authorization: Bearer <token>
Content-Type: application/json

{
  "order_id": 1,
  "method": "alipay"
}
```

**method 选项**:
- `alipay`: 支付宝
- `wechat`: 微信

**响应** (201):
```json
{
  "id": 1,
  "order_id": 1,
  "amount": 199.98,
  "method": "alipay",
  "status": "pending",
  "created_at": "2026-03-14T10:00:00Z",
  "updated_at": "2026-03-14T10:00:00Z"
}
```

### 获取支付详情

```
GET /payments/:id
Authorization: Bearer <token>
```

**响应** (200): 支付对象

### 退款

```
POST /payments/:id/refund
Authorization: Bearer <token>
```

**响应** (200): 更新后的支付对象

### 支付宝回调

```
POST /payments/webhooks/alipay
Content-Type: application/json

{
  "payment_id": 1,
  "status": "success",
  "amount": 199.98,
  "transaction_id": "alipay_tx_xxx"
}
```

**响应** (200):
```json
{
  "message": "Callback processed"
}
```

### 微信回调

```
POST /payments/webhooks/wechat
Content-Type: application/json

{
  "payment_id": 1,
  "status": "success",
  "amount": 199.98,
  "transaction_id": "wechat_tx_xxx"
}
```

**响应** (200):
```json
{
  "message": "Callback processed"
}
```

---

## 💰 Token 管理

### 获取 Token 余额

```
GET /tokens/balance
Authorization: Bearer <token>
```

**响应** (200):
```json
{
  "id": 1,
  "user_id": 1,
  "balance": 5000.00,
  "created_at": "2026-03-14T10:00:00Z",
  "updated_at": "2026-03-14T10:00:00Z"
}
```

### 获取 Token 消费记录

```
GET /tokens/consumption
Authorization: Bearer <token>
```

**响应** (200):
```json
[
  {
    "id": 1,
    "type": "use",
    "amount": -100.00,
    "reason": "Purchase for order",
    "created_at": "2026-03-14T10:00:00Z"
  }
]
```

### 转账 Token

```
POST /tokens/transfer
Authorization: Bearer <token>
Content-Type: application/json

{
  "recipient_id": 2,
  "amount": 100.00
}
```

**响应** (200):
```json
{
  "message": "Transfer successful"
}
```

### 列出 API 密钥

```
GET /tokens/keys
Authorization: Bearer <token>
```

**响应** (200):
```json
[
  {
    "id": 1,
    "user_id": 1,
    "name": "My API Key",
    "status": "active",
    "last_used_at": "2026-03-14T10:00:00Z",
    "created_at": "2026-03-14T10:00:00Z",
    "updated_at": "2026-03-14T10:00:00Z"
  }
]
```

### 创建 API 密钥

```
POST /tokens/keys
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "My API Key"
}
```

**响应** (201):
```json
{
  "id": 1,
  "key": "ptd_xxxxxxxxxxxxxxxxxxxxx",
  "name": "My API Key",
  "status": "active"
}
```

### 更新 API 密钥

```
PUT /tokens/keys/:id
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Updated Name",
  "status": "inactive"
}
```

**响应** (200): 更新后的 API 密钥对象

### 删除 API 密钥

```
DELETE /tokens/keys/:id
Authorization: Bearer <token>
```

**响应** (200):
```json
{
  "message": "API key deleted successfully"
}
```

---

## ⚠️ 错误处理

### 错误响应格式

```json
{
  "error": "Error message",
  "code": "ERROR_CODE"
}
```

### 常见 HTTP 状态码

| 状态码 | 说明 |
|--------|------|
| 200 | OK - 请求成功 |
| 201 | Created - 资源创建成功 |
| 400 | Bad Request - 请求参数错误 |
| 401 | Unauthorized - 未授权/无效 Token |
| 403 | Forbidden - 禁止访问 |
| 404 | Not Found - 资源不存在 |
| 409 | Conflict - 冲突（如库存不足） |
| 500 | Internal Server Error - 服务器错误 |

### 示例错误

**401 Unauthorized**:
```json
{
  "error": "unauthorized"
}
```

**400 Bad Request**:
```json
{
  "error": "invalid email address"
}
```

**404 Not Found**:
```json
{
  "error": "Product not found"
}
```

---

## 🧪 测试 API

### 使用 curl

```bash
# 注册
curl -X POST http://localhost:8080/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "name": "Test User",
    "password": "password123"
  }'

# 登录
curl -X POST http://localhost:8080/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'

# 获取当前用户（需要 Token）
curl -X GET http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"

# 获取产品列表
curl -X GET "http://localhost:8080/api/v1/products?page=1&per_page=20"
```

### 使用 Postman

1. 导入以下环境变量：
   - `base_url`: `http://localhost:8080/api/v1`
   - `token`: 登录后获得的 JWT Token

2. 创建请求集合并逐一测试

---

## 📚 参考

- [REST API 最佳实践](https://restfulapi.net/)
- [JWT 认证](https://jwt.io/)
- [HTTP 状态码](https://httpwg.org/specs/rfc7231.html#status.codes)

---

**版本历史**:
- v1.0.0 (2026-03-14): 初始版本

**维护者**: Engineering Team
**最后更新**: 2026-03-14
