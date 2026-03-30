# 拼脱脱 SKU 系统重构详细设计文档

**文档版本**: V1.2  
**更新日期**: 2026-03-29  
**文档用途**: 指导 SKU 系统重构开发

---

## 目录

1. [概述](#1-概述)
2. [现有系统影响分析](#2-现有系统影响分析)
3. [核心架构设计](#3-核心架构设计)
4. [数据模型设计](#4-数据模型设计)
5. [API 接口设计](#5-api-接口设计)
6. [管理运营端设计](#6-管理运营端设计)
7. [商户端设计](#7-商户端设计)
8. [用户端设计](#8-用户端设计)
9. [数据迁移策略](#9-数据迁移策略)
10. [开发任务清单](#10-开发任务清单)

---

## 1. 概述

### 1.1 背景

根据 2026 年行业动态，云计算行业已告别"只降不涨"的历史。平台核心价值在于对冲供应链波动和降低 B 端商户接入复杂度，通过"拼团"模式聚合 C 端长尾流量。

### 1.2 设计目标

1. **统一计量与映射层**: 适配不同上游厂商的计费模式
2. **SPU + SKU 架构**: 实现标准化商品管理
3. **多维度计费**: 支持按量付费、包月订阅、混合计费等模式
4. **算力点体系**: 统一平台内部计价单位

### 1.3 现有系统差距分析

| 维度 | 现有系统 | 目标系统 | 状态 |
|------|----------|----------|------|
| 商品模型 | 简单 Product 表 | SPU + SKU 双层架构 | ✅ 数据层已实现 |
| 模型分类 | 无 | 四象限分类（旗舰/标准/轻量/多模态） | ✅ 已实现 |
| 计费方式 | 单一价格 | 多种计费模式 | ✅ 已实现 |
| 算力点 | 无 | 统一计价单位 | ✅ 已实现 |
| 管理端 | 空页面 | 完整 SKU 管理功能 | ✅ 已实现 |
| 商户端 | 基础商品管理 | SKU 配置与上架 | 待增强 |
| 订单模块 | 关联 product_id | 关联 sku_id | 待迁移 |
| 购物车模块 | 关联 product_id | 关联 sku_id | 待迁移 |
| 拼团模块 | 关联 product_id | 关联 sku_id | 待迁移 |

---

## 2. 现有系统影响分析

### 2.1 现有产品模型

**数据表**: `products`

```sql
CREATE TABLE products (
  id SERIAL PRIMARY KEY,
  merchant_id INT,
  name VARCHAR(255),
  description TEXT,
  price DECIMAL(10, 2),
  original_price DECIMAL(10, 2),
  stock INT,
  sold_count INT DEFAULT 0,
  category VARCHAR(100),
  status VARCHAR(50) DEFAULT 'active',
  created_at TIMESTAMP,
  updated_at TIMESTAMP
);
```

**特点**:
- 单层商品结构，无 SPU/SKU 分层
- 商户关联 (merchant_id)，支持商户自营商品
- 简单的价格和库存管理
- 分类字段为自由文本

### 2.2 受影响模块清单

| 模块 | 涉及文件 | 影响程度 | 改造内容 |
|------|----------|----------|----------|
| 订单 | `order_and_group.go`, `order.ts`, `CheckoutPage.tsx` | 🔴 高 | product_id → sku_id，价格计算逻辑 |
| 购物车 | `cart.go`, `CartPage.tsx` | 🔴 高 | product_id → sku_id，SKU 规格展示 |
| 拼团 | `order_and_group.go`, `GroupListPage.tsx` | 🔴 高 | product_id → sku_id，拼团折扣逻辑 |
| 商品展示 | `product.go`, `ProductListPage.tsx` | 🟡 中 | 展示 SKU + SPU 联合数据 |
| 商户商品 | `merchant.go`, `MerchantProducts.tsx` | 🟡 中 | 改为选择 SPU 创建 SKU |
| 收藏/历史 | `favorite_history.go` | 🟡 中 | product_id → sku_id |

### 2.3 核心变更对比

| 维度 | 现有系统 | 目标系统 | 变化说明 |
|------|----------|----------|----------|
| 商品层级 | 单层 Product | SPU + SKU 双层 | 需要数据迁移 |
| 商品创建 | 商户自主创建 | 平台定义 SPU，商户配置 SKU | 流程重构 |
| 分类方式 | 自由文本 category | 模型层级 (pro/lite/mini/vision) | 标准化 |
| 计费模式 | 单一价格 | 多种计费类型 | 逻辑扩展 |
| 算力点 | 无 | 统一计价单位 | 新增模块 |

### 2.4 详细影响分析

详见: [sku-refactor-impact-analysis.md](./sku-refactor-impact-analysis.md)

---

## 3. 核心架构设计

### 2.1 SPU + SKU 架构

```
┌─────────────────────────────────────────────────────────────┐
│                         SPU (标准产品单元)                    │
│  例: DeepSeek-V3.2 模型服务                                   │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │   SKU 1     │  │   SKU 2     │  │   SKU 3     │  ...     │
│  │ 100万Token包│  │ 包月无限量版 │  │ 专属并发版   │          │
│  └─────────────┘  └─────────────┘  └─────────────┘          │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 模型分类体系

| 层级 | 代码 | 说明 | 适用场景 | 定价策略 |
|------|------|------|----------|----------|
| 旗舰版 | `pro` | doubao-seed-2.0-pro, ERNIE 5.0 | 复杂推理/代码生成 | 高单价/按量付费 |
| 标准版 | `lite` | DeepSeek-V3, GLM-4 | 日常对话/内容创作 | 主力走量/适合拼团 |
| 轻量版 | `mini` | doubao-seed-1.6-flash, ERNIE Speed | 实时交互/分类 | 极低价/引流 |
| 多模态版 | `vision` | Nano Banana, doubao-vision | 图像生成/视觉理解 | 按张/Token混合 |

### 2.3 SKU 类型

| 类型 | 代码 | 说明 |
|------|------|------|
| Token 包 | `token_pack` | 固定 Token 数量，有效期限制 |
| 订阅套餐 | `subscription` | 月/季/年订阅，可选无限量 |
| 并发套餐 | `concurrent` | 专属并发资源，TPM/RPM 限制 |
| 试用套餐 | `trial` | 免费试用，限制时长和用量 |

### 2.4 算力点体系

**基础定义**:
- 1 算力点 = 1K Token（标准上下文）

**汇率机制**:
| 模型类型 | 消耗系数 | 说明 |
|----------|----------|------|
| DeepSeek-V3 | 1.0 | 基准系数 |
| GLM-4 | 1.2 | 略高成本 |
| ERNIE 5.0 | 3.0 | 高端模型 |
| doubao-seed-2.0-pro | 2.5 | 旗舰版 |
| 多模态模型 | 5.0 | 特殊计费 |

### 2.5 SPU 适配器设计（新增）

根据需求文档 **2.1 适配器模型设计**，SPU 需要支持不同上游厂商的计费模式适配：

#### 输入长度区间配置 (input_length_ranges)

用于支持火山引擎的分段计价模式：

```json
[
  {"min_tokens": 0, "max_tokens": 32000, "label": "32K", "surcharge": 0},
  {"min_tokens": 32000, "max_tokens": 128000, "label": "128K", "surcharge": 20},
  {"min_tokens": 128000, "max_tokens": 256000, "label": "256K", "surcharge": 50}
]
```

#### 计费适配器配置 (billing_adapter)

支持三种计费类型：

| 类型 | 说明 | 适用厂商 |
|------|------|----------|
| `flat` | 统一计费 | DeepSeek, OpenAI |
| `segment` | 分段计费 | 火山引擎, 百度千帆 |
| `tiered` | 阶梯计费 | 阿里云, 腾讯云 |

```json
{
  "type": "segment",
  "segment_config": [
    {"input_range": "32K", "multiplier": 1.0},
    {"input_range": "128K", "multiplier": 1.2},
    {"input_range": "256K", "multiplier": 1.5}
  ],
  "cache_enabled": true,
  "cache_discount_rate": 50
}
```

#### 智能路由规则 (routing_rules)

根据用户套餐自动路由到对应价格区间：

```json
{
  "auto_route": true,
  "default_range": "32K",
  "range_mapping": {
    "lite": "32K",
    "pro": "128K"
  }
}
```

#### 批量推理配置 (batch_inference)

利用厂商批量推理低价通道（百度千帆批量推理价格仅为在线推理的40%）：

```json
{
  "enabled": true,
  "discount_rate": 60,
  "async_only": true
}
```

---

## 3. 数据模型设计

### 3.1 模型厂商配置表 (model_providers)

```sql
CREATE TABLE model_providers (
  id SERIAL PRIMARY KEY,
  code VARCHAR(50) UNIQUE NOT NULL,
  name VARCHAR(100) NOT NULL,
  
  api_base_url VARCHAR(255),
  api_format VARCHAR(50) DEFAULT 'openai',
  billing_type VARCHAR(50),
  segment_config JSONB,
  
  cache_enabled BOOLEAN DEFAULT FALSE,
  cache_discount_rate DECIMAL(5, 2),
  
  status VARCHAR(50) DEFAULT 'active',
  sort_order INT DEFAULT 0,
  
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 3.2 SPU 表 (spus)

```sql
CREATE TABLE spus (
  id SERIAL PRIMARY KEY,
  spu_code VARCHAR(100) UNIQUE NOT NULL,
  name VARCHAR(255) NOT NULL,
  
  model_provider VARCHAR(100) NOT NULL,
  model_name VARCHAR(100) NOT NULL,
  model_version VARCHAR(50),
  model_tier VARCHAR(50) NOT NULL CHECK (model_tier IN ('pro', 'lite', 'mini', 'vision')),
  
  context_window INT,
  max_output_tokens INT,
  supported_functions JSONB,
  
  base_compute_points DECIMAL(10, 4) NOT NULL DEFAULT 1.0,
  
  description TEXT,
  features JSONB,
  thumbnail_url VARCHAR(500),
  
  -- 适配器配置（新增）
  input_length_ranges JSONB,    -- 输入长度区间配置
  billing_adapter JSONB,        -- 计费适配器配置
  routing_rules JSONB,          -- 智能路由规则
  batch_inference JSONB,        -- 批量推理配置
  
  status VARCHAR(50) NOT NULL DEFAULT 'active',
  sort_order INT DEFAULT 0,
  
  total_sales_count BIGINT DEFAULT 0,
  average_rating DECIMAL(3, 2),
  
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 字段注释
COMMENT ON COLUMN spus.input_length_ranges IS '输入长度区间配置，支持分段计价';
COMMENT ON COLUMN spus.billing_adapter IS '计费适配器配置，支持不同厂商计费模式';
COMMENT ON COLUMN spus.routing_rules IS '智能路由规则，根据用户套餐自动路由';
COMMENT ON COLUMN spus.batch_inference IS '批量推理配置，利用低价通道';
```

### 3.3 SKU 表 (skus)

```sql
CREATE TABLE skus (
  id SERIAL PRIMARY KEY,
  spu_id INT NOT NULL,
  sku_code VARCHAR(100) UNIQUE NOT NULL,
  merchant_id INT,
  
  sku_type VARCHAR(50) NOT NULL CHECK (sku_type IN ('token_pack', 'subscription', 'concurrent', 'trial')),
  
  token_amount BIGINT,
  compute_points DECIMAL(15, 2),
  
  subscription_period VARCHAR(50) CHECK (subscription_period IN ('monthly', 'quarterly', 'yearly')),
  is_unlimited BOOLEAN DEFAULT FALSE,
  fair_use_limit BIGINT,
  
  tpm_limit INT,
  rpm_limit INT,
  concurrent_requests INT,
  
  valid_days INT DEFAULT 365,
  
  retail_price DECIMAL(10, 2) NOT NULL,
  wholesale_price DECIMAL(10, 2),
  original_price DECIMAL(10, 2),
  
  stock INT DEFAULT -1,
  daily_limit INT,
  
  group_enabled BOOLEAN DEFAULT TRUE,
  min_group_size INT DEFAULT 2,
  max_group_size INT DEFAULT 10,
  group_discount_rate DECIMAL(5, 2),
  
  is_trial BOOLEAN DEFAULT FALSE,
  trial_duration_days INT,
  
  status VARCHAR(50) NOT NULL DEFAULT 'active',
  is_promoted BOOLEAN DEFAULT FALSE,
  
  sales_count BIGINT DEFAULT 0,
  
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  
  FOREIGN KEY (spu_id) REFERENCES spus(id) ON DELETE CASCADE,
  FOREIGN KEY (merchant_id) REFERENCES users(id) ON DELETE SET NULL
);
```

### 3.4 算力点账户表 (compute_point_accounts)

```sql
CREATE TABLE compute_point_accounts (
  id SERIAL PRIMARY KEY,
  user_id INT NOT NULL UNIQUE,
  
  balance DECIMAL(15, 2) NOT NULL DEFAULT 0,
  total_earned DECIMAL(15, 2) DEFAULT 0,
  total_used DECIMAL(15, 2) DEFAULT 0,
  total_expired DECIMAL(15, 2) DEFAULT 0,
  
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

### 3.5 算力点交易记录表 (compute_point_transactions)

```sql
CREATE TABLE compute_point_transactions (
  id SERIAL PRIMARY KEY,
  user_id INT NOT NULL,
  
  type VARCHAR(50) NOT NULL CHECK (type IN ('purchase', 'reward', 'usage', 'refund', 'expire', 'group_bonus')),
  amount DECIMAL(15, 2) NOT NULL,
  balance_after DECIMAL(15, 2) NOT NULL,
  
  order_id INT,
  sku_id INT,
  
  description VARCHAR(500),
  metadata JSONB,
  
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE SET NULL,
  FOREIGN KEY (sku_id) REFERENCES skus(id) ON DELETE SET NULL
);
```

### 3.6 用户订阅表 (user_subscriptions)

```sql
CREATE TABLE user_subscriptions (
  id SERIAL PRIMARY KEY,
  user_id INT NOT NULL,
  sku_id INT NOT NULL,
  
  start_date DATE NOT NULL,
  end_date DATE NOT NULL,
  
  used_tokens BIGINT DEFAULT 0,
  used_compute_points DECIMAL(15, 2) DEFAULT 0,
  
  status VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'expired', 'cancelled')),
  auto_renew BOOLEAN DEFAULT FALSE,
  
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY (sku_id) REFERENCES skus(id) ON DELETE CASCADE
);
```

---

## 4. API 接口设计

### 4.1 管理端 API

#### SPU 管理
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/admin/spus` | 获取 SPU 列表 |
| GET | `/api/v1/admin/spus/:id` | 获取 SPU 详情 |
| POST | `/api/v1/admin/spus` | 创建 SPU |
| PUT | `/api/v1/admin/spus/:id` | 更新 SPU |
| DELETE | `/api/v1/admin/spus/:id` | 删除 SPU |

#### SKU 管理
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/admin/skus` | 获取 SKU 列表 |
| GET | `/api/v1/admin/skus/:id` | 获取 SKU 详情 |
| POST | `/api/v1/admin/skus` | 创建 SKU |
| PUT | `/api/v1/admin/skus/:id` | 更新 SKU |
| DELETE | `/api/v1/admin/skus/:id` | 删除 SKU |

#### 其他
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/admin/model-providers` | 获取模型厂商列表 |

### 4.2 公开 API

#### SKU 查询
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/skus` | 获取公开 SKU 列表 |
| GET | `/api/v1/skus/:id` | 获取公开 SKU 详情 |

#### 算力点
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/compute-points/balance` | 获取算力点余额 |
| GET | `/api/v1/compute-points/transactions` | 获取算力点交易记录 |

#### 订阅
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/subscriptions` | 获取用户订阅列表 |

---

## 5. 管理运营端设计

### 5.1 SPU 管理页面

**路径**: `/admin/spus`

**功能**:
- SPU 列表展示（支持按厂商、层级筛选）
- 新增/编辑 SPU
- 配置模型技术参数
- 设置算力点消耗系数

### 5.2 SKU 管理页面

**路径**: `/admin/skus`

**功能**:
- SKU 列表展示（支持按 SPU、类型筛选）
- 新增/编辑 SKU
- 配置计费方式
- 设置拼团规则

---

## 6. 商户端设计

### 6.1 商品上架页面（待重构）

**路径**: `/merchant/products`

**新增功能**:
- 选择平台 SPU 创建商品
- 配置商户专属 SKU
- 设置商户定价（在平台批发价基础上加价）
- 配置拼团参数

---

## 7. 用户端设计

### 7.1 商品展示优化

**Token 可视化**:
```
剩余 Token: 500,000
├─ 可生成约 250 篇小红书文案
├─ 可进行约 100 次长文档翻译
└─ 可进行约 50 次代码生成任务
```

---

## 8. 开发任务清单

### 8.1 已完成任务

| 任务 | 状态 | 说明 |
|------|------|------|
| 数据库迁移脚本 | ✅ | `015_sku_refactor.sql` |
| 后端数据模型 | ✅ | `backend/models/sku.go` |
| 后端 API 处理器 | ✅ | `backend/handlers/sku.go` |
| 后端路由配置 | ✅ | `backend/routes/routes.go` |
| 缓存 Key 配置 | ✅ | `backend/cache/cache.go` |
| 前端类型定义 | ✅ | `frontend/src/types/sku.ts` |
| 前端 API 服务 | ✅ | `frontend/src/services/sku.ts` |
| 管理端 SPU 页面 | ✅ | `frontend/src/pages/admin/AdminSPUs.tsx` |
| 管理端 SKU 页面 | ✅ | `frontend/src/pages/admin/AdminSKUs.tsx` |

### 8.2 待完成任务

| 任务 | 优先级 | 预估工时 |
|------|--------|----------|
| 商户端商品上架页面重构 | P0 | 6h |
| 用户端商品列表页优化 | P1 | 4h |
| 用户端商品详情页优化 | P1 | 4h |
| Token 可视化组件 | P1 | 3h |
| 单元测试编写 | P0 | 8h |
| 集成测试编写 | P1 | 4h |
| E2E 测试编写 | P1 | 4h |

---

## 9. 文件清单

### 9.1 后端文件

| 文件路径 | 说明 |
|----------|------|
| `backend/migrations/015_sku_refactor.sql` | 数据库迁移脚本 |
| `backend/models/sku.go` | SPU/SKU 数据模型 |
| `backend/handlers/sku.go` | SPU/SKU API 处理器 |
| `backend/routes/routes.go` | 路由配置（已更新） |
| `backend/cache/cache.go` | 缓存 Key 配置（已更新） |
| `backend/main.go` | 主程序（已更新） |

### 9.2 前端文件

| 文件路径 | 说明 |
|----------|------|
| `frontend/src/types/sku.ts` | TypeScript 类型定义 |
| `frontend/src/services/sku.ts` | API 服务 |
| `frontend/src/pages/admin/AdminSPUs.tsx` | SPU 管理页面 |
| `frontend/src/pages/admin/AdminSKUs.tsx` | SKU 管理页面 |

---

**文档结束**
