# 拼脱脱 数据模型设计

**文档用途**：为数据库设计和后端开发提供完整的数据模型参考

---

## 目录

1. [核心概念与关系](#核心概念与关系)
2. [详细表结构设计](#详细表结构设计)
3. [数据关系图](#数据关系图)
4. [索引与性能优化](#索引与性能优化)

---

## 核心概念与关系

### 实体关系概览

```
用户 (Users)
  ├─ 一对多 → 订单 (Orders)
  ├─ 一对多 → 拼团参与 (GroupMembers)
  ├─ 一对多 → 消费记录 (Consumption)
  └─ 一对多 → 邀请关系 (Referrals)

商家 (Merchants)
  ├─ 一对多 → 商品 (Products/SKUs)
  ├─ 一对多 → API Key (ApiKeys)
  └─ 一对多 → 销售数据 (SalesData)

商品 (Products)
  ├─ 一对多 → 订单 (Orders)
  ├─ 一对多 → 拼团 (Groups)
  ├─ 一对多 → 拼团规则 (GroupRules)
  └─ 多对多 → 模型 (Models)

订单 (Orders)
  ├─ 多对一 → 用户 (Users)
  ├─ 多对一 → 商品 (Products)
  ├─ 多对一 → 拼团 (Groups) [可选]
  └─ 一对多 → 支付记录 (Payments)

拼团 (Groups)
  ├─ 一对多 → 拼团成员 (GroupMembers)
  ├─ 一对多 → 拼团规则 (GroupRules)
  └─ 多对一 → 商品 (Products)

API调用 (ApiCalls)
  ├─ 多对一 → 用户 (Users)
  ├─ 多对一 → API Key (ApiKeys)
  └─ 一对一 → 消费记录 (Consumption)
```

---

## 详细表结构设计

### 1. Users 表（用户表）

**用途**：存储C端用户基本信息和账户数据

```sql
CREATE TABLE users (
  -- 基本字段
  id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '用户ID',
  username VARCHAR(100) UNIQUE NOT NULL COMMENT '用户名',
  email VARCHAR(255) UNIQUE NOT NULL COMMENT '邮箱',
  phone VARCHAR(20) UNIQUE COMMENT '手机号',
  password_hash VARCHAR(255) NOT NULL COMMENT '密码哈希值(bcrypt加密)',

  -- 用户信息
  nickname VARCHAR(100) COMMENT '昵称',
  avatar_url VARCHAR(500) COMMENT '头像URL',
  real_name VARCHAR(100) COMMENT '真实姓名',
  id_card_number VARCHAR(50) COMMENT '身份证号(可选)',

  -- 账户数据
  user_level VARCHAR(50) NOT NULL DEFAULT 'new' COMMENT '用户等级: new/active/loyal/inactive',
  credit_score INT NOT NULL DEFAULT 60 COMMENT '信用评分(60-100)',
  balance DECIMAL(18, 2) NOT NULL DEFAULT 0 COMMENT 'Token余额(已转换为人民币价值)',

  -- 统计数据
  total_consumption DECIMAL(18, 2) NOT NULL DEFAULT 0 COMMENT '总消费额',
  purchase_count INT NOT NULL DEFAULT 0 COMMENT '购买次数',
  referral_count INT NOT NULL DEFAULT 0 COMMENT '成功邀请人数',
  referral_earning DECIMAL(18, 2) NOT NULL DEFAULT 0 COMMENT '邀请返利总额',

  -- 用户标签 (JSON格式存储)
  user_tags JSON COMMENT '{"usage_models": ["GLM-5", "K2.5"], "time_slots": ["night"], "consumption_level": "medium"}',

  -- 状态管理
  is_active TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否激活(1=激活, 0=禁用)',
  is_banned TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否被禁用(1=禁用, 0=正常)',
  ban_reason VARCHAR(500) COMMENT '禁用原因(如作弊、违规等)',

  -- 时间戳
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  last_login_at TIMESTAMP COMMENT '最后登录时间',

  -- 索引
  INDEX idx_email (email),
  INDEX idx_phone (phone),
  INDEX idx_user_level (user_level),
  INDEX idx_credit_score (credit_score),
  INDEX idx_created_at (created_at)
);
```

### 2. Merchants 表（商家表）

**用途**：存储B端商家信息

```sql
CREATE TABLE merchants (
  -- 基本字段
  id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '商家ID',
  merchant_name VARCHAR(255) NOT NULL UNIQUE COMMENT '商家名称',
  merchant_code VARCHAR(50) UNIQUE NOT NULL COMMENT '商家代码',

  -- 认证信息
  business_license_number VARCHAR(100) UNIQUE COMMENT '营业执照号',
  legal_person_name VARCHAR(100) COMMENT '法人名称',
  legal_person_id_card VARCHAR(50) COMMENT '法人身份证号',
  business_scope VARCHAR(500) COMMENT '经营范围',

  -- 联系信息
  contact_person VARCHAR(100) NOT NULL COMMENT '联系人',
  contact_phone VARCHAR(20) NOT NULL COMMENT '联系电话',
  contact_email VARCHAR(255) NOT NULL COMMENT '联系邮箱',
  address VARCHAR(500) COMMENT '公司地址',

  -- 结算信息
  bank_account_name VARCHAR(100) NOT NULL COMMENT '开户名(与营业执照法人一致)',
  bank_account_number VARCHAR(50) NOT NULL COMMENT '银行账号',
  bank_name VARCHAR(100) NOT NULL COMMENT '银行名称',
  bank_branch VARCHAR(100) COMMENT '支行名称',
  tax_id VARCHAR(100) COMMENT '税号',

  -- 平台数据
  commission_rate DECIMAL(5, 2) NOT NULL DEFAULT 30 COMMENT '平台佣金比例(%)',
  monthly_quota BIGINT NOT NULL COMMENT '月度Token配额',
  current_month_usage BIGINT NOT NULL DEFAULT 0 COMMENT '当月已使用Token',

  -- 统计数据
  total_sales DECIMAL(18, 2) NOT NULL DEFAULT 0 COMMENT '总销售额',
  total_orders INT NOT NULL DEFAULT 0 COMMENT '总订单数',
  total_earning DECIMAL(18, 2) NOT NULL DEFAULT 0 COMMENT '总收入(扣除佣金)',

  -- 店铺信息
  shop_icon_url VARCHAR(500) COMMENT '店铺图标',
  shop_description VARCHAR(1000) COMMENT '店铺描述',
  shop_notice VARCHAR(500) COMMENT '店铺公告',

  -- 状态管理
  status VARCHAR(50) NOT NULL DEFAULT 'pending_review' COMMENT '审核状态: pending_review/approved/rejected/suspended',
  audit_status VARCHAR(50) NOT NULL DEFAULT 'not_audited' COMMENT '定期审核状态',
  is_active TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否激活',

  -- 时间戳
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '申请时间',
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  approved_at TIMESTAMP COMMENT '审核通过时间',

  -- 索引
  INDEX idx_merchant_name (merchant_name),
  INDEX idx_merchant_code (merchant_code),
  INDEX idx_status (status),
  INDEX idx_created_at (created_at)
);
```

### 3. Products 表（SKU商品表）

**用途**：存储商家上架的商品SKU

```sql
CREATE TABLE products (
  -- 基本字段
  id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '商品ID',
  sku_code VARCHAR(100) UNIQUE NOT NULL COMMENT 'SKU代码(如: GLM-5-100W)',
  merchant_id BIGINT NOT NULL COMMENT '商家ID',

  -- 商品信息
  product_name VARCHAR(255) NOT NULL COMMENT '商品名称',
  model_name VARCHAR(100) NOT NULL COMMENT '模型名称(如GLM-5)',
  model_version VARCHAR(50) COMMENT '模型版本(如v1.0)',

  -- Token信息
  token_amount BIGINT NOT NULL COMMENT 'Token数量(单位: 个)',
  token_price_per_unit DECIMAL(10, 8) NOT NULL COMMENT '单个Token的价格(元)',

  -- 定价信息
  retail_price DECIMAL(10, 2) NOT NULL COMMENT '零售价(单独购买价格)',
  wholesale_price DECIMAL(10, 2) NOT NULL COMMENT 'B端成本价(用于计算最低拼团价)',

  -- 商品属性
  context_window INT COMMENT '上下文窗口(K)',
  supported_functions JSON COMMENT '["编码", "文本生成", "数据分析"]',
  max_concurrent_requests INT COMMENT '最大并发请求数(可选)',
  rate_limit INT COMMENT '限流(requests/min)',

  -- 有效期
  valid_days INT NOT NULL DEFAULT 365 COMMENT '有效期(天)',
  valid_start_date DATE COMMENT '生效开始日期',
  valid_end_date DATE COMMENT '生效结束日期',

  -- 库存管理
  daily_inventory_limit INT COMMENT '每日库存上限',
  inventory_warning_threshold INT COMMENT '库存预警阈值',

  -- 描述和详情
  description TEXT COMMENT '商品描述',
  thumbnail_url VARCHAR(500) COMMENT '商品缩略图',
  detailed_images JSON COMMENT '详细图片JSON数组',
  faq JSON COMMENT 'FAQ内容',

  -- 统计数据
  total_sales_count INT NOT NULL DEFAULT 0 COMMENT '总销量',
  total_sales_amount DECIMAL(18, 2) NOT NULL DEFAULT 0 COMMENT '总销售额',
  average_rating DECIMAL(3, 2) COMMENT '平均评分(1-5)',
  review_count INT NOT NULL DEFAULT 0 COMMENT '评价数',

  -- 状态管理
  status VARCHAR(50) NOT NULL DEFAULT 'draft' COMMENT '状态: draft/active/inactive/suspended',
  is_promoted TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否参与平台推广',

  -- 时间戳
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

  -- 索引
  INDEX idx_merchant_id (merchant_id),
  INDEX idx_model_name (model_name),
  INDEX idx_status (status),
  INDEX idx_created_at (created_at),
  FOREIGN KEY (merchant_id) REFERENCES merchants(id)
);
```

### 4. GroupRules 表（拼团规则表）

**用途**：存储每个商品的拼团规则

```sql
CREATE TABLE group_rules (
  -- 基本字段
  id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '规则ID',
  product_id BIGINT NOT NULL COMMENT '商品ID',
  rule_code VARCHAR(100) UNIQUE NOT NULL COMMENT '规则代码(如: 2-person-team)',

  -- 拼团参数
  target_count INT NOT NULL COMMENT '目标人数(如2、5)',
  price_per_person DECIMAL(10, 2) NOT NULL COMMENT '每人价格',
  discount_rate DECIMAL(5, 2) NOT NULL COMMENT '折扣率(%)',
  max_duration_minutes INT NOT NULL DEFAULT 120 COMMENT '最长拼团时间(分钟)',

  -- 成团条件
  min_members_for_auto_group INT NOT NULL DEFAULT 2 COMMENT '自动成团最少人数',
  auto_group_subsidy_rate DECIMAL(5, 2) NOT NULL DEFAULT 50 COMMENT '自动成团补贴比例(%)',

  -- 状态
  is_active TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否激活',

  -- 时间戳
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

  -- 索引
  INDEX idx_product_id (product_id),
  UNIQUE KEY unique_product_rule (product_id, target_count),
  FOREIGN KEY (product_id) REFERENCES products(id)
);
```

### 5. Orders 表（订单表）

**用途**：存储所有用户订单信息

```sql
CREATE TABLE orders (
  -- 基本字段
  id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '订单ID',
  order_no VARCHAR(50) UNIQUE NOT NULL COMMENT '订单号(如PIN202603141001)',
  user_id BIGINT NOT NULL COMMENT '用户ID',
  merchant_id BIGINT NOT NULL COMMENT '商家ID',
  product_id BIGINT NOT NULL COMMENT '商品ID',

  -- 拼团信息
  group_id BIGINT COMMENT '拼团ID(单独购买为NULL)',
  order_type VARCHAR(50) NOT NULL COMMENT '订单类型: solo(单独购买)/group(拼团)',

  -- 金额信息
  original_price DECIMAL(10, 2) NOT NULL COMMENT '原始价格(零售价)',
  actual_price DECIMAL(10, 2) NOT NULL COMMENT '实际支付价格',
  subsidy_amount DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '补贴金额',
  referral_reward DECIMAL(10, 2) NOT NULL DEFAULT 0 COMMENT '邀请返利',
  total_amount DECIMAL(10, 2) NOT NULL COMMENT '总金额',

  -- Token信息
  token_amount BIGINT NOT NULL COMMENT '获得的Token数量',

  -- 状态流转
  status VARCHAR(50) NOT NULL DEFAULT 'pending_payment' COMMENT '订单状态: pending_payment/paid/grouping/completed/cancelled/refunded',

  -- 时间戳
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  paid_at TIMESTAMP COMMENT '支付时间',
  completed_at TIMESTAMP COMMENT '完成时间',
  cancelled_at TIMESTAMP COMMENT '取消时间',

  -- 备注
  remark VARCHAR(500) COMMENT '备注',

  -- 索引
  INDEX idx_user_id (user_id),
  INDEX idx_merchant_id (merchant_id),
  INDEX idx_product_id (product_id),
  INDEX idx_group_id (group_id),
  INDEX idx_status (status),
  INDEX idx_created_at (created_at),
  INDEX idx_order_no (order_no),
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (merchant_id) REFERENCES merchants(id),
  FOREIGN KEY (product_id) REFERENCES products(id)
);
```

### 6. Groups 表（拼团表）

**用途**：存储拼团活动的基本信息

```sql
CREATE TABLE groups (
  -- 基本字段
  id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '拼团ID',
  group_no VARCHAR(50) UNIQUE NOT NULL COMMENT '拼团号',
  product_id BIGINT NOT NULL COMMENT '商品ID',
  group_rule_id BIGINT NOT NULL COMMENT '拼团规则ID',

  -- 发起人
  initiator_id BIGINT NOT NULL COMMENT '发起人用户ID',

  -- 拼团人数
  target_count INT NOT NULL COMMENT '目标成团人数',
  current_count INT NOT NULL DEFAULT 1 COMMENT '当前已拼人数',

  -- 时间信息
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  deadline TIMESTAMP NOT NULL COMMENT '拼团截止时间',
  completed_at TIMESTAMP COMMENT '成团/完成时间',

  -- 状态
  status VARCHAR(50) NOT NULL DEFAULT 'recruiting' COMMENT '状态: recruiting(招募中)/success(成功)/failed(失败)',
  failure_reason VARCHAR(500) COMMENT '失败原因(如: 时间到期人数不足)',

  -- 拼团链接
  share_code VARCHAR(50) UNIQUE COMMENT '分享码',

  -- 索引
  INDEX idx_product_id (product_id),
  INDEX idx_initiator_id (initiator_id),
  INDEX idx_status (status),
  INDEX idx_deadline (deadline),
  INDEX idx_created_at (created_at),
  FOREIGN KEY (product_id) REFERENCES products(id),
  FOREIGN KEY (group_rule_id) REFERENCES group_rules(id),
  FOREIGN KEY (initiator_id) REFERENCES users(id)
);
```

### 7. GroupMembers 表（拼团成员表）

**用途**：存储每个拼团的成员列表

```sql
CREATE TABLE group_members (
  -- 基本字段
  id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '成员记录ID',
  group_id BIGINT NOT NULL COMMENT '拼团ID',
  user_id BIGINT NOT NULL COMMENT '用户ID',
  order_id BIGINT NOT NULL COMMENT '对应的订单ID',

  -- 角色
  role VARCHAR(50) NOT NULL COMMENT '角色: initiator(发起人)/member(普通成员)',

  -- 邀请追踪
  invited_by_user_id BIGINT COMMENT '邀请人ID(NULL表示自主参加)',
  referral_code VARCHAR(50) COMMENT '邀请码',

  -- 状态
  status VARCHAR(50) NOT NULL DEFAULT 'joined' COMMENT '状态: joined(已加入)/paid(已支付)',

  -- 时间戳
  joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '加入时间',

  -- 索引
  INDEX idx_group_id (group_id),
  INDEX idx_user_id (user_id),
  UNIQUE KEY unique_group_member (group_id, user_id),
  FOREIGN KEY (group_id) REFERENCES groups(id),
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (order_id) REFERENCES orders(id)
);
```

### 8. ApiKeys 表（API密钥表）

**用途**：存储商家托管到平台的API密钥

```sql
CREATE TABLE api_keys (
  -- 基本字段
  id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT 'API Key ID',
  merchant_id BIGINT NOT NULL COMMENT '商家ID',

  -- 密钥信息
  key_hash VARCHAR(255) NOT NULL UNIQUE COMMENT '密钥哈希值(加密存储)',
  key_type VARCHAR(50) NOT NULL COMMENT '密钥类型: bearer/api-key',
  model_name VARCHAR(100) NOT NULL COMMENT '对应模型名',

  -- 配额管理
  monthly_quota BIGINT NOT NULL COMMENT '月度Token配额',
  used_quota BIGINT NOT NULL DEFAULT 0 COMMENT '已使用配额',
  quota_reset_day INT NOT NULL DEFAULT 1 COMMENT '配额重置日期(1-28)',

  -- 状态
  status VARCHAR(50) NOT NULL DEFAULT 'active' COMMENT '状态: active/disabled/deleted',

  -- 备注
  remark VARCHAR(500) COMMENT '备注',

  -- 时间戳
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

  -- 索引
  INDEX idx_merchant_id (merchant_id),
  INDEX idx_model_name (model_name),
  INDEX idx_status (status),
  FOREIGN KEY (merchant_id) REFERENCES merchants(id)
);
```

### 9. Consumption 表（消费记录表）

**用途**：记录用户调用API的详细消费数据

```sql
CREATE TABLE consumption (
  -- 基本字段
  id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '消费ID',
  user_id BIGINT NOT NULL COMMENT '用户ID',
  api_key_id BIGINT NOT NULL COMMENT 'API Key ID',

  -- 消费详情
  model_name VARCHAR(100) NOT NULL COMMENT '模型名',
  method VARCHAR(50) NOT NULL COMMENT 'API方法(如chat/completions)',

  -- Token统计
  input_tokens BIGINT NOT NULL COMMENT '输入Token数',
  output_tokens BIGINT NOT NULL COMMENT '输出Token数',
  total_tokens BIGINT NOT NULL COMMENT '总Token数',

  -- 费用计算
  unit_price DECIMAL(10, 8) NOT NULL COMMENT '单价(元/Token)',
  cost DECIMAL(10, 4) NOT NULL COMMENT '费用(元)',

  -- 请求信息
  request_id VARCHAR(100) COMMENT '外部请求ID(用于追踪)',

  -- 状态
  status VARCHAR(50) NOT NULL DEFAULT 'completed' COMMENT '状态: pending/completed/failed',

  -- 时间戳
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',

  -- 索引
  INDEX idx_user_id (user_id),
  INDEX idx_api_key_id (api_key_id),
  INDEX idx_model_name (model_name),
  INDEX idx_created_at (created_at),
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (api_key_id) REFERENCES api_keys(id)
);
```

### 10. Referrals 表（邀请关系表）

**用途**：追踪用户之间的邀请关系和返利

```sql
CREATE TABLE referrals (
  -- 基本字段
  id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '邀请记录ID',
  referrer_user_id BIGINT NOT NULL COMMENT '邀请人ID',
  referred_user_id BIGINT NOT NULL COMMENT '被邀请人ID',

  -- 邀请信息
  referral_code VARCHAR(50) NOT NULL COMMENT '邀请码',
  referral_link VARCHAR(500) COMMENT '邀请链接',
  referred_order_id BIGINT COMMENT '触发返利的订单ID',

  -- 返利信息
  referral_reward DECIMAL(10, 2) NOT NULL COMMENT '返利金额',
  reward_status VARCHAR(50) NOT NULL DEFAULT 'pending' COMMENT '返利状态: pending/earned/paid',

  -- 时间戳
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '邀请时间',
  referred_order_created_at TIMESTAMP COMMENT '被邀请人首单时间',
  reward_paid_at TIMESTAMP COMMENT '返利支付时间',

  -- 索引
  INDEX idx_referrer_user_id (referrer_user_id),
  INDEX idx_referred_user_id (referred_user_id),
  INDEX idx_reward_status (reward_status),
  FOREIGN KEY (referrer_user_id) REFERENCES users(id),
  FOREIGN KEY (referred_user_id) REFERENCES users(id)
);
```

### 11. Payments 表（支付记录表）

**用途**：记录所有支付交易信息

```sql
CREATE TABLE payments (
  -- 基本字段
  id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '支付ID',
  payment_no VARCHAR(50) UNIQUE NOT NULL COMMENT '支付单号',
  order_id BIGINT NOT NULL COMMENT '订单ID',
  user_id BIGINT NOT NULL COMMENT '用户ID',

  -- 支付信息
  payment_method VARCHAR(50) NOT NULL COMMENT '支付方式: alipay/wechat/bankcard',
  payment_channel VARCHAR(100) COMMENT '支付渠道代码',

  -- 金额信息
  amount DECIMAL(10, 2) NOT NULL COMMENT '支付金额',

  -- 状态
  status VARCHAR(50) NOT NULL DEFAULT 'pending' COMMENT '支付状态: pending/success/failed/refunded',

  -- 外部参考
  external_transaction_id VARCHAR(100) COMMENT '支付渠道的交易ID',

  -- 时间戳
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  paid_at TIMESTAMP COMMENT '支付完成时间',

  -- 索引
  INDEX idx_order_id (order_id),
  INDEX idx_user_id (user_id),
  INDEX idx_status (status),
  FOREIGN KEY (order_id) REFERENCES orders(id),
  FOREIGN KEY (user_id) REFERENCES users(id)
);
```

### 12. UserTags 表（用户标签表，用于个性化推荐）

**用途**：存储用户的多维度标签，支持个性化推荐

```sql
CREATE TABLE user_tags (
  -- 基本字段
  id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '标签ID',
  user_id BIGINT NOT NULL COMMENT '用户ID',

  -- 标签维度
  model_preference JSON COMMENT '模型偏好: {"GLM-5": 0.7, "K2.5": 0.5}',
  consumption_level VARCHAR(50) COMMENT '消费级别: low/medium/high',
  activity_type VARCHAR(50) COMMENT '活跃类型: sharing/lurking/purchasing',
  time_slot VARCHAR(50) COMMENT '活跃时段: morning/afternoon/night',

  -- 时间戳
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

  -- 索引
  INDEX idx_user_id (user_id),
  FOREIGN KEY (user_id) REFERENCES users(id)
);
```

---

## 数据关系图

```
┌─────────────┐
│   Users     │◄────────────────────┐
│  (C端用户)  │                     │
└──┬──────────┘                     │
   │                                │
   ├─────►┌──────────────┐         │
   │      │ Orders       │         │
   │      │ (订单)       │         │
   │      └──┬───────────┘         │
   │         │                     │
   │         ├──┐                  │
   │         │  └──►┌─────────────┐│
   │         │      │ GroupMembers││
   │         │      │ (拼团成员)  ││
   │         │      └─────────────┘│
   │         │                     │
   │         └──►┌──────────────┐  │
   │            │ Payments     │  │
   │            │ (支付记录)   │  │
   │            └──────────────┘  │
   │                              │
   ├─────►┌──────────────────┐    │
   │      │ Consumption      │    │
   │      │ (消费记录)       │    │
   │      └──────────────────┘    │
   │                              │
   ├─────────────────────────────┘
   │      (邀请人)
   │
   └─────►┌──────────────┐
          │ Referrals    │
          │ (邀请关系)   │
          └──────────────┘
                 │
                 └──►被邀请人→Users

┌──────────────┐
│ Merchants    │◄──────────────────────┐
│ (B端商家)    │                       │
└──┬───────────┘                       │
   │                                   │
   ├─────►┌──────────────┐             │
   │      │ Products     │             │
   │      │ (商品SKU)    │             │
   │      └──┬───────────┘             │
   │         │                        │
   │         ├─►┌──────────────┐      │
   │         │  │ GroupRules   │      │
   │         │  │ (拼团规则)   │      │
   │         │  └──────────────┘      │
   │         │                        │
   │         ├─►┌──────────────┐      │
   │         │  │ Orders       │──────┘
   │         │  │ (订单)       │
   │         │  └──────────────┘
   │         │
   │         └─►┌──────────────┐
   │            │ Groups       │
   │            │ (拼团)       │
   │            └──────────────┘
   │
   └─────►┌──────────────┐
          │ ApiKeys      │
          │ (API密钥)    │
          └──┬───────────┘
             │
             └─►┌──────────────┐
                │ Consumption  │
                │ (消费记录)   │
                └──────────────┘
```

---

## 索引与性能优化

### 核心索引策略

```
1. 用户表索引
   - PRIMARY KEY: id
   - UNIQUE: email, phone, username
   - 常规: user_level, credit_score, created_at
   - 用途: 快速查询用户信息、按等级分组

2. 订单表索引
   - PRIMARY KEY: id
   - UNIQUE: order_no
   - 常规: user_id, product_id, status, created_at
   - 用途: 快速查询用户订单、订单状态统计

3. 拼团表索引
   - PRIMARY KEY: id
   - UNIQUE: group_no, share_code
   - 常规: product_id, status, deadline, created_at
   - 用途: 快速查询拼团进度、检查成团条件

4. 消费记录表索引
   - PRIMARY KEY: id
   - 常规: user_id, api_key_id, model_name, created_at
   - 用途: 快速查询用户消费、按模型统计

5. API密钥表索引
   - PRIMARY KEY: id
   - UNIQUE: key_hash
   - 常规: merchant_id, status, model_name
   - 用途: 快速查询密钥、按状态筛选
```

### 性能优化建议

**分表策略**：
- 订单表：按日期分表（orders_202603, orders_202604等）
- 消费记录：按日期分表（consumption_202603, consumption_202604等）
- 用途：减少单表数据量，提高查询速度

**缓存策略**：
- Redis缓存用户余额 (user:{user_id}:balance)
- Redis缓存拼团进度 (group:{group_id}:progress)
- Redis缓存热销商品 (product:{product_id}:hotness)

**查询优化**：
- 使用联接查询时，确保关联字段都有索引
- 避免在WHERE子句中使用函数
- 使用EXPLAIN分析查询执行计划

---

**END OF DOCUMENT**
