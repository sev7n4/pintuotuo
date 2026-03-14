# 拼脱脱 API接口清单

**文档用途**：为后端开发和系统集成提供完整的API端点参考

**API基础信息**：
- 基础URL: `https://api.pintuotuo.com`
- API版本: `v1`
- 认证方式: Bearer Token (JWT)
- 请求格式: JSON
- 响应格式: JSON

---

## 目录

1. [C端用户API](#c端用户api)
2. [C端订单API](#c端订单api)
3. [C端拼团API](#c端拼团api)
4. [C端Token管理API](#c端token管理api)
5. [B端商家API](#b端商家api)
6. [B端数据API](#b端数据api)
7. [平台通用API](#平台通用api)
8. [Webhook回调接口](#webhook回调接口)

---

## C端用户API

### 1. 用户注册

```
POST /v1/users/register

请求体:
{
  "email": "user@example.com",
  "phone": "13800138000",
  "password": "hashedPassword",
  "nickname": "张三"
}

响应 (201 Created):
{
  "code": 200,
  "message": "注册成功",
  "data": {
    "user_id": 12345,
    "email": "user@example.com",
    "nickname": "张三",
    "created_at": "2026-03-14T10:00:00Z",
    "token": "eyJhbGc..."
  }
}

错误响应 (400 Bad Request):
{
  "code": 400,
  "message": "邮箱已存在或格式错误",
  "errors": {
    "email": "email_exists"
  }
}
```

### 2. 用户登录

```
POST /v1/users/login

请求体:
{
  "email_or_phone": "user@example.com",
  "password": "hashedPassword"
}

响应 (200 OK):
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "user_id": 12345,
    "email": "user@example.com",
    "nickname": "张三",
    "token": "eyJhbGc...",
    "expires_in": 86400,
    "token_type": "Bearer"
  }
}

错误响应 (401 Unauthorized):
{
  "code": 401,
  "message": "邮箱或密码错误"
}
```

### 3. 获取用户信息

```
GET /v1/users/me

请求头:
Authorization: Bearer {token}

响应 (200 OK):
{
  "code": 200,
  "data": {
    "user_id": 12345,
    "email": "user@example.com",
    "phone": "13800138000",
    "nickname": "张三",
    "avatar_url": "https://...",
    "user_level": "active",
    "credit_score": 78,
    "balance": 2500000,
    "total_consumption": 5000,
    "purchase_count": 8,
    "referral_count": 3,
    "referral_earning": 15,
    "created_at": "2026-01-01T00:00:00Z"
  }
}
```

### 4. 修改用户信息

```
PUT /v1/users/me

请求体:
{
  "nickname": "新昵称",
  "avatar_url": "https://..."
}

响应 (200 OK):
{
  "code": 200,
  "message": "更新成功",
  "data": {
    "user_id": 12345,
    "nickname": "新昵称"
  }
}
```

### 5. 修改密码

```
PUT /v1/users/password

请求体:
{
  "old_password": "oldHash",
  "new_password": "newHash"
}

响应 (200 OK):
{
  "code": 200,
  "message": "密码修改成功"
}
```

---

## C端订单API

### 1. 创建订单

```
POST /v1/orders

请求体:
{
  "product_id": 101,
  "order_type": "solo",
  "quantity": 1,
  "coupon_code": "NEWUSER20" (可选)
}

或拼团订单:
{
  "product_id": 101,
  "order_type": "group",
  "group_rule_id": 5,
  "group_id": 999 (可选,如果加入现有拼团)
}

响应 (201 Created):
{
  "code": 200,
  "message": "订单创建成功",
  "data": {
    "order_id": 50001,
    "order_no": "PIN202603141001",
    "user_id": 12345,
    "product_id": 101,
    "order_type": "group",
    "original_price": 100,
    "actual_price": 60,
    "total_amount": 60,
    "status": "pending_payment",
    "group_id": 999,
    "created_at": "2026-03-14T10:30:00Z"
  }
}
```

### 2. 获取订单列表

```
GET /v1/orders?status=all&page=1&limit=20

查询参数:
- status: all, pending_payment, paid, grouping, completed, cancelled
- page: 页码(默认1)
- limit: 每页数量(默认20, 最大100)

响应 (200 OK):
{
  "code": 200,
  "data": {
    "total": 50,
    "page": 1,
    "limit": 20,
    "items": [
      {
        "order_id": 50001,
        "order_no": "PIN202603141001",
        "product_name": "Kimi K2.5 编码大师包",
        "order_type": "group",
        "status": "grouping",
        "original_price": 100,
        "actual_price": 60,
        "total_amount": 60,
        "group_id": 999,
        "created_at": "2026-03-14T10:30:00Z",
        "group_info": {
          "current_count": 1,
          "target_count": 2,
          "deadline": "2026-03-14T12:30:00Z"
        }
      }
    ]
  }
}
```

### 3. 获取订单详情

```
GET /v1/orders/{order_id}

响应 (200 OK):
{
  "code": 200,
  "data": {
    "order_id": 50001,
    "order_no": "PIN202603141001",
    "user_id": 12345,
    "product_id": 101,
    "product_name": "Kimi K2.5 编码大师包",
    "order_type": "group",
    "status": "grouping",
    "original_price": 100,
    "actual_price": 60,
    "subsidy_amount": 0,
    "total_amount": 60,
    "token_amount": 1000000,
    "created_at": "2026-03-14T10:30:00Z",
    "paid_at": null,
    "completed_at": null,
    "group_id": 999,
    "group_info": {
      "group_no": "GRP202603141001",
      "current_count": 1,
      "target_count": 2,
      "members": [
        {
          "user_id": 12345,
          "nickname": "张三",
          "avatar_url": "https://...",
          "role": "initiator",
          "joined_at": "2026-03-14T10:30:00Z"
        }
      ],
      "deadline": "2026-03-14T12:30:00Z"
    }
  }
}
```

### 4. 取消订单

```
DELETE /v1/orders/{order_id}

响应 (200 OK):
{
  "code": 200,
  "message": "订单已取消,退款处理中",
  "data": {
    "order_id": 50001,
    "status": "cancelled",
    "refund_amount": 60
  }
}
```

---

## C端拼团API

### 1. 创建拼团

```
POST /v1/groups

请求体:
{
  "product_id": 101,
  "group_rule_id": 5
}

响应 (201 Created):
{
  "code": 200,
  "message": "拼团创建成功",
  "data": {
    "group_id": 999,
    "group_no": "GRP202603141001",
    "product_id": 101,
    "product_name": "Kimi K2.5 编码大师包",
    "initiator_id": 12345,
    "target_count": 2,
    "current_count": 1,
    "share_code": "ZHANGSAN123",
    "share_url": "https://pintuotuo.com/group/999?code=ZHANGSAN123",
    "status": "recruiting",
    "deadline": "2026-03-14T12:30:00Z",
    "created_at": "2026-03-14T10:30:00Z"
  }
}
```

### 2. 获取拼团详情

```
GET /v1/groups/{group_id}

响应 (200 OK):
{
  "code": 200,
  "data": {
    "group_id": 999,
    "group_no": "GRP202603141001",
    "product_id": 101,
    "product_name": "Kimi K2.5 编码大师包",
    "target_count": 2,
    "current_count": 1,
    "members": [
      {
        "user_id": 12345,
        "nickname": "张三",
        "avatar_url": "https://...",
        "role": "initiator",
        "status": "paid"
      }
    ],
    "status": "recruiting",
    "deadline": "2026-03-14T12:30:00Z",
    "remaining_slots": 1,
    "remaining_minutes": 118
  }
}
```

### 3. 加入拼团

```
POST /v1/groups/{group_id}/join

请求体:
{
  "referral_code": "ZHANGSAN123" (可选)
}

响应 (200 OK):
{
  "code": 200,
  "message": "加入拼团成功,请完成支付",
  "data": {
    "group_id": 999,
    "user_id": 54321,
    "status": "joined",
    "current_count": 2,
    "target_count": 2,
    "auto_checkout": true
  }
}
```

### 4. 获取拼团进度

```
GET /v1/groups/{group_id}/progress

响应 (200 OK):
{
  "code": 200,
  "data": {
    "group_id": 999,
    "current_count": 1,
    "target_count": 2,
    "remaining_slots": 1,
    "deadline": "2026-03-14T12:30:00Z",
    "remaining_minutes": 118,
    "remaining_seconds": 7080,
    "status": "recruiting",
    "auto_group_possible": false,
    "auto_subsidy_rate": 50
  }
}
```

### 5. 取消拼团

```
DELETE /v1/groups/{group_id}

响应 (200 OK):
{
  "code": 200,
  "message": "拼团已取消",
  "data": {
    "group_id": 999,
    "status": "cancelled",
    "refund_reason": "user_cancelled"
  }
}
```

### 6. 获取推荐拼团

```
GET /v1/groups/recommended?product_id=101

响应 (200 OK):
{
  "code": 200,
  "data": {
    "groups": [
      {
        "group_id": 999,
        "product_name": "Kimi K2.5 编码大师包",
        "current_count": 1,
        "target_count": 2,
        "remaining_slots": 1,
        "deadline": "2026-03-14T12:30:00Z",
        "initiator_nickname": "李四"
      }
    ]
  }
}
```

---

## C端Token管理API

### 1. 获取API Key列表

```
GET /v1/tokens/keys

响应 (200 OK):
{
  "code": 200,
  "data": {
    "total_balance": 2500000,
    "monthly_usage": 125000,
    "keys": [
      {
        "key_id": "key_001",
        "key_masked": "sk_****XXXX",
        "model_name": "GLM-5",
        "balance": 1000000,
        "source_product": "Kimi K2.5 编码大师包",
        "created_at": "2026-02-01T00:00:00Z",
        "expires_at": "2027-02-01T00:00:00Z",
        "status": "active"
      }
    ]
  }
}
```

### 2. 复制API Key

```
POST /v1/tokens/keys/{key_id}/copy

请求体:
{
  "verify_password": "passwordHash"
}

响应 (200 OK):
{
  "code": 200,
  "message": "密钥已复制到剪贴板",
  "data": {
    "key": "sk_full_key_here_base64_encoded"
  }
}
```

### 3. 禁用API Key

```
PUT /v1/tokens/keys/{key_id}/disable

响应 (200 OK):
{
  "code": 200,
  "message": "API Key已禁用",
  "data": {
    "key_id": "key_001",
    "status": "disabled"
  }
}
```

### 4. 删除API Key

```
DELETE /v1/tokens/keys/{key_id}

响应 (200 OK):
{
  "code": 200,
  "message": "API Key已删除"
}
```

### 5. 获取消费明细

```
GET /v1/tokens/consumption?start_date=2026-03-01&end_date=2026-03-14&model=GLM-5&limit=100

查询参数:
- start_date: 开始日期(YYYY-MM-DD)
- end_date: 结束日期
- model: 模型名称(可选)
- limit: 返回记录数(默认100)

响应 (200 OK):
{
  "code": 200,
  "data": {
    "total": 520,
    "total_cost": 52.5,
    "items": [
      {
        "date": "2026-03-14",
        "model": "GLM-5",
        "api_calls": 125,
        "input_tokens": 100000,
        "output_tokens": 25000,
        "total_tokens": 125000,
        "unit_price": 0.0001,
        "cost": 12.5
      }
    ]
  }
}
```

### 6. 获取账单

```
GET /v1/tokens/bills/{year}/{month}

响应 (200 OK):
{
  "code": 200,
  "data": {
    "year": 2026,
    "month": 3,
    "total_consumption": 3750000,
    "total_cost": 375,
    "summary": {
      "GLM-5": { "tokens": 2000000, "cost": 200 },
      "K2.5": { "tokens": 1500000, "cost": 150 },
      "Claude": { "tokens": 250000, "cost": 25 }
    }
  }
}
```

---

## B端商家API

### 1. 商家注册/入驻

```
POST /v1/merchants/register

请求体:
{
  "merchant_name": "智谱AI",
  "legal_person_name": "张三",
  "business_license_number": "123456789",
  "contact_person": "李四",
  "contact_phone": "13800138000",
  "contact_email": "contact@zhipu.ai",
  "bank_account_name": "张三",
  "bank_account_number": "6225xxx",
  "bank_name": "工商银行"
}

响应 (201 Created):
{
  "code": 200,
  "message": "申请已提交,请等待审核",
  "data": {
    "merchant_id": 1001,
    "status": "pending_review",
    "created_at": "2026-03-14T10:00:00Z"
  }
}
```

### 2. 获取商家信息

```
GET /v1/merchants/me

响应 (200 OK):
{
  "code": 200,
  "data": {
    "merchant_id": 1001,
    "merchant_name": "智谱AI",
    "status": "approved",
    "commission_rate": 30,
    "total_sales": 256980,
    "total_orders": 2560,
    "total_earning": 77094,
    "monthly_quota": 10000000,
    "current_month_usage": 3250000
  }
}
```

### 3. 创建SKU

```
POST /v1/merchants/products

请求体:
{
  "product_name": "Kimi K2.5 编码大师包",
  "model_name": "K2.5",
  "model_version": "v1.0",
  "token_amount": 1000000,
  "token_price_per_unit": 0.0001,
  "retail_price": 100,
  "wholesale_price": 80,
  "context_window": 128,
  "supported_functions": ["编码", "文本生成"],
  "valid_days": 365,
  "daily_inventory_limit": 1000,
  "description": "...",
  "thumbnail_url": "https://..."
}

响应 (201 Created):
{
  "code": 200,
  "message": "商品创建成功",
  "data": {
    "product_id": 101,
    "sku_code": "K25-1M-001",
    "status": "draft"
  }
}
```

### 4. 获取商品列表

```
GET /v1/merchants/products?status=active&page=1&limit=20

响应 (200 OK):
{
  "code": 200,
  "data": {
    "total": 5,
    "items": [
      {
        "product_id": 101,
        "product_name": "Kimi K2.5 编码大师包",
        "model_name": "K2.5",
        "retail_price": 100,
        "status": "active",
        "total_sales": 1200,
        "total_sales_amount": 72000,
        "average_rating": 4.8,
        "created_at": "2026-02-01T00:00:00Z"
      }
    ]
  }
}
```

### 5. 编辑SKU

```
PUT /v1/merchants/products/{product_id}

请求体:
{
  "product_name": "新名称",
  "retail_price": 110,
  "daily_inventory_limit": 2000
}

响应 (200 OK):
{
  "code": 200,
  "message": "商品已更新",
  "data": {
    "product_id": 101
  }
}
```

### 6. 创建拼团规则

```
POST /v1/merchants/products/{product_id}/group-rules

请求体:
{
  "target_count": 5,
  "price_per_person": 50,
  "discount_rate": 50,
  "max_duration_minutes": 240
}

响应 (201 Created):
{
  "code": 200,
  "data": {
    "rule_id": 5,
    "product_id": 101
  }
}
```

### 7. 上传API Key

```
POST /v1/merchants/api-keys

请求体:
{
  "key": "sk_full_api_key_here",
  "model_name": "K2.5",
  "monthly_quota": 10000000,
  "remark": "主生产密钥"
}

响应 (201 Created):
{
  "code": 200,
  "message": "API Key已保存",
  "data": {
    "key_id": "merchant_key_001",
    "status": "active"
  }
}
```

### 8. 获取API Key列表

```
GET /v1/merchants/api-keys

响应 (200 OK):
{
  "code": 200,
  "data": {
    "keys": [
      {
        "key_id": "merchant_key_001",
        "key_masked": "sk_****XXXX",
        "model_name": "K2.5",
        "monthly_quota": 10000000,
        "used_quota": 3250000,
        "status": "active"
      }
    ]
  }
}
```

### 9. 编辑API Key配额

```
PUT /v1/merchants/api-keys/{key_id}

请求体:
{
  "monthly_quota": 15000000
}

响应 (200 OK):
{
  "code": 200,
  "message": "配额已更新"
}
```

---

## B端数据API

### 1. 获取销售数据看板

```
GET /v1/merchants/analytics/sales?start_date=2026-03-01&end_date=2026-03-14

响应 (200 OK):
{
  "code": 200,
  "data": {
    "summary": {
      "total_sales": 256980,
      "total_orders": 2560,
      "group_success_rate": 92.5,
      "average_order_value": 100.38
    },
    "daily_sales": [
      {
        "date": "2026-03-14",
        "sales_amount": 8530,
        "order_count": 85,
        "group_count": 78,
        "group_success_rate": 91.8
      }
    ]
  }
}
```

### 2. 获取订单详情列表

```
GET /v1/merchants/orders?page=1&limit=50

响应 (200 OK):
{
  "code": 200,
  "data": {
    "total": 2560,
    "items": [
      {
        "order_id": 50001,
        "order_no": "PIN202603141001",
        "user_nickname": "张三",
        "product_name": "Kimi K2.5 编码大师包",
        "order_type": "group",
        "actual_price": 60,
        "status": "grouping",
        "created_at": "2026-03-14T10:30:00Z"
      }
    ]
  }
}
```

### 3. 获取分润明细

```
GET /v1/merchants/analytics/revenue?month=202603

响应 (200 OK):
{
  "code": 200,
  "data": {
    "month": "2026-03",
    "total_sales": 256980,
    "commission_rate": 30,
    "platform_commission": 77094,
    "api_cost": 102792,
    "merchant_earnings": 77094,
    "gross_margin_rate": 30,
    "details": [
      {
        "date": "2026-03-14",
        "sales": 8530,
        "commission": 2559,
        "cost": 3412,
        "earning": 2559
      }
    ]
  }
}
```

### 4. 获取用户分析

```
GET /v1/merchants/analytics/users

响应 (200 OK):
{
  "code": 200,
  "data": {
    "new_users": 256,
    "new_user_rate": 35,
    "returning_users": 476,
    "returning_rate": 65,
    "average_purchase_frequency": 3.4,
    "average_order_value": 100.38,
    "repeat_purchase_rate": 65,
    "top_models": [
      {
        "model_name": "K2.5",
        "user_count": 450,
        "rate": 45
      }
    ]
  }
}
```

---

## 平台通用API

### 1. 首页推荐

```
GET /v1/home/feed?page=1&limit=20

响应 (200 OK):
{
  "code": 200,
  "data": {
    "banners": [
      {
        "id": 1,
        "image_url": "https://...",
        "target_url": "https://pintuotuo.com/...",
        "title": "百亿补贴"
      }
    ],
    "products": [
      {
        "product_id": 101,
        "product_name": "Kimi K2.5 编码大师包",
        "retail_price": 100,
        "group_rules": [
          {
            "rule_id": 5,
            "target_count": 2,
            "price_per_person": 60
          }
        ],
        "sales_count": 10000,
        "average_rating": 4.8,
        "thumbnail_url": "https://..."
      }
    ]
  }
}
```

### 2. 搜索商品

```
GET /v1/search?q=编码&category=model&page=1&limit=20

响应 (200 OK):
{
  "code": 200,
  "data": {
    "total": 150,
    "items": [
      {
        "product_id": 101,
        "product_name": "Kimi K2.5 编码大师包",
        "highlight": "Kimi K2.5 <em>编码</em>大师包",
        "retail_price": 100
      }
    ]
  }
}
```

### 3. 获取分类列表

```
GET /v1/categories

响应 (200 OK):
{
  "code": 200,
  "data": {
    "categories": [
      {
        "id": "large-model",
        "name": "大模型",
        "icon": "https://...",
        "subcategories": [
          {
            "id": "coding",
            "name": "编码"
          }
        ]
      }
    ]
  }
}
```

### 4. 获取分类商品

```
GET /v1/categories/{category_id}/products?page=1&limit=20&sort=sales

响应 (200 OK):
{
  "code": 200,
  "data": {
    "category_name": "编码",
    "total": 45,
    "items": [...]
  }
}
```

### 5. 获取商品详情

```
GET /v1/products/{product_id}

响应 (200 OK):
{
  "code": 200,
  "data": {
    "product_id": 101,
    "product_name": "Kimi K2.5 编码大师包",
    "merchant_name": "智谱AI",
    "model_name": "K2.5",
    "token_amount": 1000000,
    "retail_price": 100,
    "group_rules": [
      {
        "rule_id": 5,
        "target_count": 2,
        "price_per_person": 60,
        "discount_rate": 40
      }
    ],
    "sales_count": 10000,
    "average_rating": 4.8,
    "description": "...",
    "images": [...]
  }
}
```

### 6. 获取用户邀请信息

```
GET /v1/users/me/referral

响应 (200 OK):
{
  "code": 200,
  "data": {
    "referral_code": "ZHANGSAN123",
    "referral_link": "https://pintuotuo.com/?uid=zhangsan",
    "total_invited": 8,
    "total_earned": 40,
    "pending_earned": 10,
    "referrals": [
      {
        "user_id": 54321,
        "user_nickname": "李四",
        "invited_at": "2026-03-10T00:00:00Z",
        "status": "completed",
        "reward": 5
      }
    ]
  }
}
```

### 7. 分享邀请链接

```
POST /v1/referrals/share

请求体:
{
  "channel": "wechat",
  "product_id": 101 (可选)
}

响应 (200 OK):
{
  "code": 200,
  "data": {
    "share_url": "https://pintuotuo.com/?uid=zhangsan&code=xyz123",
    "share_text": "朋友推荐你来拼脱脱购买Token,享受50%折扣!",
    "qr_code": "data:image/png;base64,..."
  }
}
```

---

## Webhook回调接口

### 1. 拼团成功回调

```
POST {merchant_callback_url}/webhook/group-success

请求体:
{
  "event": "group_success",
  "timestamp": 1710404400,
  "data": {
    "group_id": 999,
    "product_id": 101,
    "members": [
      {
        "user_id": 12345,
        "order_id": 50001
      },
      {
        "user_id": 54321,
        "order_id": 50002
      }
    ],
    "total_amount": 120,
    "token_amount": 2000000
  }
}

响应 (200 OK):
{
  "code": 200,
  "message": "已收到"
}
```

### 2. 订单支付回调

```
POST {platform_callback_url}/webhook/payment-completed

支付渠道(支付宝/微信)调用此接口

请求体:
{
  "event": "payment_completed",
  "timestamp": 1710404400,
  "data": {
    "payment_no": "PAY20260314001",
    "order_id": 50001,
    "amount": 60,
    "payment_method": "alipay",
    "external_transaction_id": "2026031400000001"
  }
}
```

### 3. API消耗回调

```
POST {merchant_callback_url}/webhook/token-consumed

用户调用API并消耗Token时,平台通知商家

请求体:
{
  "event": "token_consumed",
  "timestamp": 1710404400,
  "data": {
    "user_id": 12345,
    "order_id": 50001,
    "tokens_used": 125000,
    "cost": 12.5,
    "api_calls": 125
  }
}
```

---

## 错误响应统一格式

```json
{
  "code": 400,
  "message": "错误描述信息",
  "errors": {
    "field_name": "error_code"
  }
}
```

**常见错误码**:

| 错误码 | HTTP状态 | 描述 |
|--------|----------|------|
| 200 | 200 | 成功 |
| 201 | 201 | 创建成功 |
| 400 | 400 | 请求参数错误 |
| 401 | 401 | 认证失败 |
| 403 | 403 | 权限不足 |
| 404 | 404 | 资源不存在 |
| 422 | 422 | 业务逻辑错误 |
| 500 | 500 | 服务器错误 |

---

**END OF DOCUMENT**
