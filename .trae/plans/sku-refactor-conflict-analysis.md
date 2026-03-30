# SKU 架构重构与现有产品模块冲突分析报告

**文档版本**: V1.0  
**创建日期**: 2026-03-29  
**分析目的**: 明确重构范围、冲突点、已完成工作

---

## 1. 架构对比总览

### 1.1 现有架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        现有产品架构                               │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────┐                                                │
│  │   Product   │ ◄───────────────────────────────────────┐     │
│  │  (单一表)   │                                         │     │
│  │             │                                         │     │
│  │ • id        │◄────┬────┬────┬────┬────┐              │     │
│  │ • name      │     │    │    │    │    │              │     │
│  │ • price     │     │    │    │    │    │              │     │
│  │ • stock     │     │    │    │    │    │              │     │
│  │ • category  │     │    │    │    │    │              │     │
│  └─────────────┘     │    │    │    │    │              │     │
│                      │    │    │    │    │              │     │
│  ┌─────────────┐ ┌───┴──┐ │    │    │    ┌─────────┐  │     │
│  │   Order     │ │ Cart │ │    │    │    │Favorite │  │     │
│  │ • product_id│ │product│ │    │    │    │product_id│  │     │
│  └─────────────┘ └──────┘ │    │    │    └─────────┘  │     │
│                           │    │    │                 │     │
│  ┌─────────────┐ ┌───────┴─┐ ┌┴────┐ ┌─────────────┐ │     │
│  │   Group     │ │BrowseHis│ │Token│ │ MerchantAPI │ │     │
│  │ • product_id│ │product_id│ │     │ │             │ │     │
│  └─────────────┘ └─────────┘ └─────┘ └─────────────┘ │     │
│                                                      │     │
└─────────────────────────────────────────────────────────────────┘
```

### 1.2 目标架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        目标 SKU 架构                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────┐       ┌─────────────┐                          │
│  │     SPU     │ 1:N   │     SKU     │ ◄────────────────────┐   │
│  │ (厂商模型)  │──────►│ (售卖套餐)  │                       │   │
│  │             │       │             │                       │   │
│  │ • 模型信息  │       │ • 价格      │◄──┬───┬───┬───┐      │   │
│  │ • 厂商对接  │       │ • 计费类型  │   │   │   │   │      │   │
│  │ • 技术参数  │       │ • 拼团配置  │   │   │   │   │      │   │
│  └─────────────┘       └─────────────┘   │   │   │   │      │   │
│                                           │   │   │   │      │   │
│  ┌─────────────┐ ┌─────────────┐ ┌───────┴─┐ │   │   │      │   │
│  │   Order     │ │    Cart     │ │Favorite │ │   │   │      │   │
│  │ • sku_id    │ │ • sku_id    │ │ • sku_id│ │   │   │      │   │
│  │ • spu_id    │ │ • spu_id    │ │         │ │   │   │      │   │
│  └─────────────┘ └─────────────┘ └─────────┘ │   │   │      │   │
│                                              │   │   │      │   │
│  ┌─────────────┐ ┌─────────────┐ ┌───────────┴─┐ │   │      │   │
│  │   Group     │ │ BrowseHist  │ │ComputePoints│ │   │      │   │
│  │ • sku_id    │ │ • sku_id    │ │ • balance   │ │   │      │   │
│  └─────────────┘ └─────────────┘ └─────────────┘ │   │      │   │
│                                                   │   │      │   │
│  ┌─────────────┐ ┌─────────────┐ ┌───────────────┘   │      │   │
│  │Subscription │ │ ProviderCost│ │ Adapter Layer     │      │   │
│  │ • sku_id    │ │ • spu_id    │ │                   │      │   │
│  └─────────────┘ └─────────────┘ └───────────────────┘      │   │
│                                                              │   │
└─────────────────────────────────────────────────────────────────┘
```

---

## 2. 数据模型冲突分析

### 2.1 核心冲突：Product vs SPU+SKU

| 维度 | 现有 Product 表 | 目标 SPU+SKU 架构 | 冲突程度 |
|------|----------------|-------------------|----------|
| **层级结构** | 单层 | 双层 (SPU+SKU) | 🔴 高 |
| **商品编码** | 无 | spu_code + sku_code | 🟡 中 |
| **分类方式** | category (自由文本) | model_tier (枚举) | 🔴 高 |
| **计费模式** | 单一价格 | 多种计费类型 | 🔴 高 |
| **厂商对接** | 无 | provider_* 字段 | 🟡 中 |
| **算力点** | 无 | compute_points | 🟡 中 |
| **拼团配置** | 无 | group_* 字段 | 🟡 中 |

### 2.2 关联表冲突

| 关联表 | 现有关联字段 | 目标关联字段 | 冲突程度 | 影响范围 |
|--------|-------------|-------------|----------|----------|
| **orders** | product_id | sku_id + spu_id | 🔴 高 | 订单创建/查询/统计 |
| **cart_items** | product_id | sku_id + spu_id | 🔴 高 | 购物车增删改查 |
| **groups** | product_id | sku_id + spu_id | 🔴 高 | 拼团创建/加入 |
| **favorites** | product_id | sku_id | 🟡 中 | 收藏列表 |
| **browse_history** | product_id | sku_id | 🟡 中 | 浏览历史 |

---

## 3. API 冲突分析

### 3.1 现有 API 清单

| API 路径 | 功能 | 关联字段 | 冲突程度 |
|----------|------|----------|----------|
| `GET /products` | 商品列表 | products 表 | 🔴 高 |
| `GET /products/:id` | 商品详情 | products 表 | 🔴 高 |
| `GET /products/search` | 商品搜索 | products 表 | 🔴 高 |
| `POST /products` | 创建商品 | products 表 | 🔴 高 |
| `PUT /products/:id` | 更新商品 | products 表 | 🔴 高 |
| `DELETE /products/:id` | 删除商品 | products 表 | 🔴 高 |
| `GET /products/hot` | 热门商品 | products 表 | 🔴 高 |
| `GET /products/new` | 新品商品 | products 表 | 🔴 高 |
| `GET /home` | 首页数据 | products 表 | 🔴 高 |
| `POST /orders` | 创建订单 | product_id | 🔴 高 |
| `GET /orders` | 订单列表 | product_id | 🟡 中 |
| `POST /groups` | 创建拼团 | product_id | 🔴 高 |
| `POST /groups/:id/join` | 加入拼团 | product_id | 🔴 高 |

### 3.2 API 兼容性要求

| API | 兼容策略 | 实现方式 |
|-----|----------|----------|
| `GET /products` | 必须兼容 | 返回 SKU+SPU 联合数据，字段映射 |
| `GET /products/:id` | 必须兼容 | SKU ID 作为 product_id 返回 |
| `POST /orders` | 必须兼容 | 同时支持 product_id 和 sku_id |
| `POST /groups` | 必须兼容 | 同时支持 product_id 和 sku_id |

---

## 4. 代码模块冲突分析

### 4.1 后端模块冲突

| 模块 | 文件路径 | 冲突点 | 改造内容 |
|------|----------|--------|----------|
| **Product Handler** | `handlers/product.go` | 🔴 高 | 所有查询需改为 SKU+SPU 联合查询 |
| **Order Handler** | `handlers/order_and_group.go` | 🔴 高 | 订单创建逻辑需支持 SKU |
| **Cart Handler** | `handlers/cart.go` | 🔴 高 | 购物车需支持 SKU |
| **Group Handler** | `handlers/order_and_group.go` | 🔴 高 | 拼团需支持 SKU 配置 |
| **Favorite Handler** | `handlers/favorite_history.go` | 🟡 中 | 收藏需支持 SKU |
| **Merchant Handler** | `handlers/merchant.go` | 🟡 中 | 商户商品管理需重构 |
| **Models** | `models/models.go` | 🔴 高 | Product 结构需重新定义 |

### 4.2 前端模块冲突

| 模块 | 文件路径 | 冲突点 | 改造内容 |
|------|----------|--------|----------|
| **ProductListPage** | `pages/ProductListPage.tsx` | 🔴 高 | 展示 SKU 列表，增加筛选维度 |
| **ProductDetailPage** | `pages/ProductDetailPage.tsx` | 🔴 高 | 展示 SPU+SKU 信息 |
| **CartPage** | `pages/CartPage.tsx` | 🔴 高 | 支持 SKU 规格展示 |
| **CheckoutPage** | `pages/CheckoutPage.tsx` | 🔴 高 | 支持 SKU 类型判断 |
| **GroupListPage** | `pages/GroupListPage.tsx` | 🔴 高 | 展示 SKU 拼团信息 |
| **MerchantProducts** | `pages/merchant/MerchantProducts.tsx` | 🔴 高 | 改为选择 SPU 创建 SKU |
| **HomePage** | `pages/HomePage.tsx` | 🟡 中 | 热门/新品改为 SKU |
| **OrderDetailPage** | `pages/OrderDetailPage.tsx` | 🟡 中 | 展示 SKU 信息 |

---

## 5. 已完成工作清单

### 5.1 数据层 ✅

| 任务 | 文件 | 状态 |
|------|------|------|
| SPU 表结构 | `migrations/015_sku_refactor.sql` | ✅ 已完成 |
| SKU 表结构 | `migrations/015_sku_refactor.sql` | ✅ 已完成 |
| model_providers 表 | `migrations/015_sku_refactor.sql` | ✅ 已完成 |
| compute_point_accounts 表 | `migrations/015_sku_refactor.sql` | ✅ 已完成 |
| compute_point_transactions 表 | `migrations/015_sku_refactor.sql` | ✅ 已完成 |
| user_subscriptions 表 | `migrations/015_sku_refactor.sql` | ✅ 已完成 |
| 种子数据 | `migrations/015_sku_refactor.sql` | ✅ 已完成 |

### 5.2 后端 API ✅

| 任务 | 文件 | 状态 |
|------|------|------|
| SPU/SKU 数据模型 | `models/sku.go` | ✅ 已完成 |
| SPU CRUD API | `handlers/sku.go` | ✅ 已完成 |
| SKU CRUD API | `handlers/sku.go` | ✅ 已完成 |
| 算力点账户 API | `handlers/sku.go` | ✅ 已完成 |
| 订阅管理 API | `handlers/sku.go` | ✅ 已完成 |
| 公开 SKU 查询 API | `handlers/sku.go` | ✅ 已完成 |
| 路由配置 | `routes/routes.go` | ✅ 已完成 |
| 缓存 Key 配置 | `cache/cache.go` | ✅ 已完成 |

### 5.3 前端页面 ✅

| 任务 | 文件 | 状态 |
|------|------|------|
| TypeScript 类型定义 | `types/sku.ts` | ✅ 已完成 |
| API 服务封装 | `services/sku.ts` | ✅ 已完成 |
| SPU 管理页面 | `pages/admin/AdminSPUs.tsx` | ✅ 已完成 |
| SKU 管理页面 | `pages/admin/AdminSKUs.tsx` | ✅ 已完成 |

### 5.4 设计文档 ✅

| 文档 | 文件路径 | 状态 |
|------|----------|------|
| 完整架构设计 | `.trae/plans/sku-refactor-complete-architecture.md` | ✅ 已完成 |
| 影响分析报告 | `.trae/plans/sku-refactor-impact-analysis.md` | ✅ 已完成 |
| 详细设计文档 | `.trae/plans/sku-refactor-design.md` | ✅ 已完成 |
| 开发任务清单 | `.trae/plans/sku-refactor-tasks.md` | ✅ 已完成 |

---

## 6. 待完成工作清单

### 6.1 数据层迁移 🔴 高优先级

| 任务 | 涉及表 | 状态 |
|------|--------|------|
| orders 表增加 sku_id | orders | ❌ 待开发 |
| cart_items 表增加 sku_id | cart_items | ❌ 待开发 |
| groups 表增加 sku_id | groups | ❌ 待开发 |
| favorites 表增加 sku_id | favorites | ❌ 待开发 |
| browse_history 表增加 sku_id | browse_history | ❌ 待开发 |
| 数据迁移脚本 | products → spus/skus | ❌ 待开发 |
| 兼容视图创建 | products_v2 | ❌ 待开发 |

### 6.2 后端模块改造 🔴 高优先级

| 任务 | 涉及文件 | 状态 |
|------|----------|------|
| 订单创建逻辑改造 | `handlers/order_and_group.go` | ❌ 待开发 |
| 订单查询逻辑改造 | `handlers/order_and_group.go` | ❌ 待开发 |
| 购物车查询改造 | `handlers/cart.go` | ❌ 待开发 |
| 购物车添加改造 | `handlers/cart.go` | ❌ 待开发 |
| 拼团创建逻辑改造 | `handlers/order_and_group.go` | ❌ 待开发 |
| 拼团加入逻辑改造 | `handlers/order_and_group.go` | ❌ 待开发 |
| 拼团折扣计算 | `handlers/order_and_group.go` | ❌ 待开发 |
| 商品列表 API 兼容层 | `handlers/product.go` | ❌ 待开发 |
| 商品详情 API 兼容层 | `handlers/product.go` | ❌ 待开发 |
| 收藏 API 改造 | `handlers/favorite_history.go` | ❌ 待开发 |
| 浏览历史 API 改造 | `handlers/favorite_history.go` | ❌ 待开发 |

### 6.3 前端页面改造 🔴 高优先级

| 任务 | 涉及文件 | 状态 |
|------|----------|------|
| 商品列表页改造 | `pages/ProductListPage.tsx` | ❌ 待开发 |
| 商品详情页改造 | `pages/ProductDetailPage.tsx` | ❌ 待开发 |
| 购物车页面改造 | `pages/CartPage.tsx` | ❌ 待开发 |
| 结算页面改造 | `pages/CheckoutPage.tsx` | ❌ 待开发 |
| 拼团列表页改造 | `pages/GroupListPage.tsx` | ❌ 待开发 |
| 拼团进度页改造 | `pages/GroupProgressPage.tsx` | ❌ 待开发 |
| 商户商品页重构 | `pages/merchant/MerchantProducts.tsx` | ❌ 待开发 |
| 首页数据改造 | `pages/HomePage.tsx` | ❌ 待开发 |
| 订单详情页改造 | `pages/OrderDetailPage.tsx` | ❌ 待开发 |
| 收藏页改造 | `pages/FavoritesPage.tsx` | ❌ 待开发 |
| 历史页改造 | `pages/HistoryPage.tsx` | ❌ 待开发 |

### 6.4 新增功能模块 🟡 中优先级

| 任务 | 说明 | 状态 |
|------|------|------|
| 适配器管理层 | 厂商协议转换 | ❌ 待开发 |
| 计量引擎 | Token 计量与计费 | ❌ 待开发 |
| 缓存优化层 | 请求/结果缓存 | ❌ 待开发 |
| 成本计算引擎 | 动态成本计算 | ❌ 待开发 |
| 限流管理 | 多层限流 | ❌ 待开发 |
| 对账系统 | 厂商账单对账 | ❌ 待开发 |

---

## 7. 迁移策略

### 7.1 分阶段迁移

```
Phase 1: 并行运行（当前阶段）
├─ 新表已创建（spus, skus 等）
├─ 新 API 已实现（SPU/SKU 管理）
├─ 旧系统继续运行（products 表）
└─ 新旧系统独立运行

Phase 2: 数据迁移
├─ 迁移 products 数据到 spus/skus
├─ 创建兼容视图 products_v2
├─ 更新关联表（orders, cart_items 等）
└─ 验证数据一致性

Phase 3: API 切换
├─ 旧 API 返回 SKU 数据
├─ 前端渐进式迁移
├─ 双写机制（过渡期）
└─ 监控告警

Phase 4: 清理
├─ 移除兼容层
├─ 删除旧 products 表
└─ 更新所有外键约束
```

### 7.2 兼容层设计

```sql
-- 创建兼容视图，使旧 API 返回 SKU 数据
CREATE OR REPLACE VIEW products_v2 AS
SELECT 
  s.id,
  s.sku_code as sku_code,
  s.merchant_id,
  sp.name as name,
  sp.name || ' - ' || 
    CASE s.sku_type 
      WHEN 'token_pack' THEN s.token_amount::text || ' Tokens'
      WHEN 'subscription' THEN COALESCE(s.subscription_period, 'monthly')
      ELSE s.sku_type
    END as description,
  s.retail_price as price,
  s.original_price,
  CASE WHEN s.stock = -1 THEN 999999 ELSE s.stock END as stock,
  s.sales_count as sold_count,
  sp.model_tier as category,
  s.status,
  s.created_at,
  s.updated_at
FROM skus s
JOIN spus sp ON s.spu_id = sp.id
WHERE s.status = 'active' AND sp.status = 'active';
```

---

## 8. 风险评估

### 8.1 高风险项

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 订单数据迁移失败 | 业务中断 | 先在测试环境验证，保留回滚脚本 |
| API 不兼容 | 前端报错 | 使用兼容层，渐进式迁移 |
| 数据丢失 | 不可恢复 | 备份数据库，双写机制 |

### 8.2 中风险项

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 性能下降 | 用户体验差 | 添加索引，优化查询 |
| 前端样式问题 | 视觉不一致 | 使用现有 UI 组件库 |
| 商户操作习惯改变 | 投诉 | 提供操作指南，培训 |

---

## 9. 总结

### 9.1 完成度统计

| 类别 | 已完成 | 待开发 | 完成率 |
|------|--------|--------|--------|
| 数据层 | 6 | 6 | 50% |
| 后端 API | 8 | 11 | 42% |
| 前端页面 | 4 | 11 | 27% |
| 设计文档 | 4 | 0 | 100% |
| **总计** | **22** | **28** | **44%** |

### 9.2 核心冲突总结

| 冲突类型 | 数量 | 优先级 |
|----------|------|--------|
| 🔴 高冲突 | 15 | P0 |
| 🟡 中冲突 | 8 | P1 |
| 🟢 低冲突 | 5 | P2 |

### 9.3 下一步行动

1. **立即执行**: 创建数据库迁移脚本（orders/cart/groups 表增加 sku_id）
2. **后续开发**: 改造订单、购物车、拼团模块
3. **前端迁移**: 渐进式迁移前端页面
4. **测试验证**: 端到端测试确保兼容性

---

**文档结束**
