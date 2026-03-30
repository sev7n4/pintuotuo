# SKU 系统重构影响分析报告

**文档版本**: V1.0  
**创建日期**: 2026-03-29  
**分析范围**: 现有产品模块与 SPU/SKU 架构的兼容性

---

## 1. 现有系统架构分析

### 1.1 当前产品模型 (Product)

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

### 1.2 关联模块

| 模块 | 关联方式 | 影响程度 |
|------|----------|----------|
| 订单 (orders) | `product_id` 外键 | 🔴 高 |
| 购物车 (cart_items) | `product_id` 外键 | 🔴 高 |
| 拼团 (groups) | `product_id` 外键 | 🔴 高 |
| 收藏 (favorites) | `product_id` 外键 | 🟡 中 |
| 浏览历史 (browse_history) | `product_id` 外键 | 🟡 中 |
| 商户统计 | 聚合查询 products | 🟡 中 |

---

## 2. SPU/SKU 架构对比

### 2.1 架构差异

```
┌─────────────────────────────────────────────────────────────────┐
│                        现有架构                                   │
├─────────────────────────────────────────────────────────────────┤
│  Product (单一表)                                                 │
│  - 包含模型信息、价格、库存                                        │
│  - 商户直接创建商品                                               │
│  - 无标准化分类                                                   │
└─────────────────────────────────────────────────────────────────┘

                              ⬇️ 重构

┌─────────────────────────────────────────────────────────────────┐
│                        目标架构                                   │
├─────────────────────────────────────────────────────────────────┤
│  SPU (标准产品单元)                                               │
│  - 模型厂商、模型名称、技术参数                                    │
│  - 算力点消耗系数                                                 │
│  - 平台统一管理                                                   │
├─────────────────────────────────────────────────────────────────┤
│  SKU (库存量单位)                                                 │
│  - 关联 SPU                                                      │
│  - 计费类型、价格、库存                                           │
│  - 支持商户定价加价                                               │
│  - 拼团配置                                                      │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2 核心变化

| 维度 | 现有系统 | 目标系统 | 变化说明 |
|------|----------|----------|----------|
| 商品层级 | 单层 Product | SPU + SKU 双层 | 需要数据迁移 |
| 商品创建 | 商户自主创建 | 平台定义 SPU，商户配置 SKU | 流程重构 |
| 分类方式 | 自由文本 category | 模型层级 (pro/lite/mini/vision) | 标准化 |
| 计费模式 | 单一价格 | 多种计费类型 | 逻辑扩展 |
| 算力点 | 无 | 统一计价单位 | 新增模块 |

---

## 3. 影响模块详细分析

### 3.1 订单模块 (🔴 高影响)

**涉及文件**:
- `backend/handlers/order_and_group.go`
- `frontend/src/services/order.ts`
- `frontend/src/pages/CheckoutPage.tsx`
- `frontend/src/pages/OrderDetailPage.tsx`

**影响点**:

1. **订单创建逻辑**
   - 现有: `product_id` → `products` 表
   - 目标: `sku_id` → `skus` 表，同时关联 `spu_id`

2. **价格计算**
   - 现有: `product.price * quantity`
   - 目标: 需要考虑 SKU 类型（Token包/订阅/并发），拼团折扣

3. **订单数据模型**
   ```go
   // 现有
   type Order struct {
       ProductID  int     `json:"product_id"`
       UnitPrice  float64 `json:"unit_price"`
       TotalPrice float64 `json:"total_price"`
   }
   
   // 目标
   type Order struct {
       SKUID      int     `json:"sku_id"`      // 改为 SKU
       SPUID      int     `json:"spu_id"`      // 新增 SPU
       ProductID  int     `json:"product_id"`  // 兼容字段
       UnitPrice  float64 `json:"unit_price"`
       TotalPrice float64 `json:"total_price"`
       SKUType    string  `json:"sku_type"`    // 新增：token_pack/subscription
   }
   ```

### 3.2 购物车模块 (🔴 高影响)

**涉及文件**:
- `backend/handlers/cart.go`
- `frontend/src/pages/CartPage.tsx`

**影响点**:

1. **购物车数据结构**
   - 现有: `cart_items.product_id` → `products`
   - 目标: `cart_items.sku_id` → `skus`

2. **价格展示**
   - 需要展示 SKU 规格信息（Token数量/订阅周期等）

### 3.3 拼团模块 (🔴 高影响)

**涉及文件**:
- `backend/handlers/order_and_group.go` (CreateGroup, JoinGroup)
- `frontend/src/pages/GroupListPage.tsx`
- `frontend/src/pages/GroupProgressPage.tsx`

**影响点**:

1. **拼团创建**
   - 现有: 基于 `product_id` 创建拼团
   - 目标: 基于 `sku_id` 创建拼团，SKU 需启用拼团功能

2. **拼团折扣**
   - 现有: 无拼团折扣逻辑
   - 目标: SKU 配置 `group_discount_rate`

3. **拼团人数限制**
   - 现有: `target_count` 用户指定
   - 目标: SKU 配置 `min_group_size` / `max_group_size`

### 3.4 商品展示模块 (🟡 中影响)

**涉及文件**:
- `backend/handlers/product.go` (ListProducts, GetProductByID, GetHotProducts)
- `frontend/src/pages/ProductListPage.tsx`
- `frontend/src/pages/ProductDetailPage.tsx`
- `frontend/src/pages/HomePage.tsx`

**影响点**:

1. **商品列表**
   - 现有: 展示 `products` 表数据
   - 目标: 展示 `skus` + `spus` 联合数据

2. **筛选维度**
   - 现有: 按 `category` 文本筛选
   - 目标: 按 `model_tier`、`sku_type`、`model_provider` 筛选

3. **商品详情**
   - 需要展示 SPU 信息（模型介绍、技术参数）
   - 需要展示 SKU 规格（Token数量、有效期等）

### 3.5 商户模块 (🟡 中影响)

**涉及文件**:
- `backend/handlers/merchant.go` (GetMerchantProducts, GetMerchantStats)
- `frontend/src/pages/merchant/MerchantProducts.tsx`

**影响点**:

1. **商品创建流程**
   - 现有: 商户直接创建商品
   - 目标: 商户选择平台 SPU，配置自己的 SKU

2. **商户统计**
   - 需要按 SKU 维度统计销量

### 3.6 收藏/浏览历史模块 (🟡 中影响)

**涉及文件**:
- `backend/handlers/favorite_history.go`
- `frontend/src/pages/FavoritesPage.tsx`
- `frontend/src/pages/HistoryPage.tsx`

**影响点**:

1. **关联关系**
   - 现有: `product_id`
   - 目标: `sku_id`（或同时关联 `spu_id`）

---

## 4. 数据迁移策略

### 4.1 迁移方案

```
Phase 1: 新建表结构（不影响现有系统）
├─ 创建 model_providers 表
├─ 创建 spus 表
├─ 创建 skus 表
├─ 创建 compute_point_accounts 表
├─ 创建 user_subscriptions 表
└─ 初始化种子数据

Phase 2: 数据迁移（保留 products 表）
├─ 从 products 提取 SPU 数据
│   └─ 根据 name/description 推断模型信息
├─ 从 products 迁移 SKU 数据
│   └─ 每个 product 对应一个 sku
├─ 更新订单关联
│   └─ orders.product_id → orders.sku_id
└─ 更新其他关联表

Phase 3: 兼容层（过渡期）
├─ 创建 products_v2 视图（兼容旧 API）
├─ 后端 API 双写
└─ 前端渐进式迁移

Phase 4: 完全切换
├─ 移除兼容层
├─ 删除旧 products 表
└─ 更新所有外键约束
```

### 4.2 数据映射规则

**products → spus 映射**:
| products 字段 | spus 字段 | 转换规则 |
|---------------|-----------|----------|
| name | name | 直接映射 |
| name | spu_code | 生成唯一编码 |
| name | model_provider | 根据关键词推断 |
| name | model_tier | 默认 'lite' |
| description | description | 直接映射 |

**products → skus 映射**:
| products 字段 | skus 字段 | 转换规则 |
|---------------|-----------|----------|
| id | - | 不迁移 |
| merchant_id | merchant_id | 直接映射 |
| price | retail_price | 直接映射 |
| original_price | original_price | 直接映射 |
| stock | stock | 直接映射 |
| sold_count | sales_count | 直接映射 |
| - | sku_type | 默认 'token_pack' |
| - | sku_code | 生成唯一编码 |

---

## 5. API 兼容性

### 5.1 需要保持兼容的 API

| API | 现有行为 | 兼容方案 |
|-----|----------|----------|
| `GET /products` | 返回 products 列表 | 返回 skus + spus 联合数据 |
| `GET /products/:id` | 返回单个 product | 返回 sku + spu 联合数据 |
| `POST /orders` | 接收 product_id | 同时支持 product_id 和 sku_id |
| `GET /orders` | 返回订单含 product_id | 同时返回 sku_id 和 spu_id |

### 5.2 新增 API

| API | 说明 |
|-----|------|
| `GET /skus` | 公开 SKU 列表 |
| `GET /skus/:id` | SKU 详情 |
| `GET /spus` | SPU 列表（管理端） |
| `POST /admin/spus` | 创建 SPU |
| `POST /admin/skus` | 创建 SKU |
| `GET /compute-points/balance` | 算力点余额 |
| `GET /subscriptions` | 用户订阅列表 |

---

## 6. 前端页面改造

### 6.1 需要改造的页面

| 页面 | 改造内容 | 优先级 |
|------|----------|--------|
| ProductListPage | 展示 SKU 列表，增加筛选维度 | P0 |
| ProductDetailPage | 展示 SPU + SKU 信息 | P0 |
| CartPage | 支持 SKU 规格展示 | P0 |
| CheckoutPage | 支持 SKU 类型判断 | P0 |
| MerchantProducts | 改为选择 SPU 创建 SKU | P1 |
| HomePage | 热门/新品改为 SKU | P1 |
| OrderDetailPage | 展示 SKU 信息 | P1 |

### 6.2 新增页面

| 页面 | 说明 | 优先级 |
|------|------|--------|
| AdminSPUs | SPU 管理 | P0 |
| AdminSKUs | SKU 管理 | P0 |
| ComputePoints | 算力点账户 | P1 |
| Subscriptions | 订阅管理 | P1 |

---

## 7. 风险评估

### 7.1 高风险项

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 订单数据迁移失败 | 业务中断 | 先在测试环境验证，保留回滚脚本 |
| API 不兼容 | 前端报错 | 使用兼容层，渐进式迁移 |
| 数据丢失 | 不可恢复 | 备份数据库，双写机制 |

### 7.2 中风险项

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 性能下降 | 用户体验差 | 添加索引，优化查询 |
| 前端样式问题 | 视觉不一致 | 使用现有 UI 组件库 |
| 商户操作习惯改变 | 投诉 | 提供操作指南，培训 |

---

## 8. 实施建议

### 8.1 推荐实施顺序

```
1. 数据层准备
   └─ 创建新表，初始化种子数据

2. 后端 API 扩展
   ├─ 新增 SPU/SKU 管理 API
   ├─ 新增公开 SKU 查询 API
   └─ 保持现有 API 兼容

3. 管理端开发
   ├─ SPU 管理页面
   └─ SKU 管理页面

4. 用户端改造
   ├─ 商品列表页
   ├─ 商品详情页
   └─ 购物流程

5. 商户端改造
   └─ 商品上架页面

6. 数据迁移
   ├─ 迁移现有数据
   └─ 切换流量

7. 清理工作
   └─ 移除兼容层
```

### 8.2 关键决策点

1. **是否保留 products 表？**
   - 建议: 保留作为兼容视图，最终删除

2. **商户是否可以创建 SPU？**
   - 建议: 仅平台管理员可创建 SPU，商户基于 SPU 创建 SKU

3. **现有订单如何处理？**
   - 建议: 迁移时关联到新 SKU，保留原 product_id 作为兼容字段

---

## 9. 总结

### 9.1 核心变更

1. **数据模型**: Product 单表 → SPU + SKU 双表
2. **业务流程**: 商户创建商品 → 商户配置 SKU
3. **计费模式**: 单一价格 → 多种计费类型
4. **新增模块**: 算力点体系、订阅管理

### 9.2 影响范围

- **后端**: 6 个核心模块需要改造
- **前端**: 10+ 页面需要改造
- **数据库**: 6 张新表，数据迁移

### 9.3 预估工期

| 阶段 | 工期 |
|------|------|
| 数据层准备 | 2 天 |
| 后端 API 扩展 | 5 天 |
| 管理端开发 | 3 天 |
| 用户端改造 | 4 天 |
| 商户端改造 | 3 天 |
| 数据迁移与测试 | 3 天 |
| **总计** | **20 天** |

---

**文档结束**
