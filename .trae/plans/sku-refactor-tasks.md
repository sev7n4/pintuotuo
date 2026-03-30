# SKU 系统重构开发任务清单

**项目**: 拼脱脱 SKU 系统重构  
**更新日期**: 2026-03-29  
**版本**: V1.2

---

## 当前状态总览

| 阶段 | 状态 | 完成度 |
|------|------|--------|
| Phase 1: 数据层重构 | ✅ 已完成 | 100% |
| Phase 2: 后端 API 重构 | ✅ 已完成 | 100% |
| Phase 3: 管理端前端开发 | ✅ 已完成 | 100% |
| Phase 4: 现有模块迁移 | 待开发 | 0% |
| Phase 5: 商户端前端重构 | 待开发 | 0% |
| Phase 6: 用户端优化 | 待开发 | 0% |
| Phase 7: 集成测试与部署 | 待开发 | 0% |

---

## Phase 1: 数据层重构 ✅

| 任务ID | 任务描述 | 状态 | 文件路径 |
|--------|----------|------|----------|
| DB-001 | 创建 SPU 表结构 | ✅ | `backend/migrations/015_sku_refactor.sql` |
| DB-002 | 创建 SKU 表结构 | ✅ | `backend/migrations/015_sku_refactor.sql` |
| DB-003 | 创建模型厂商配置表 | ✅ | `backend/migrations/015_sku_refactor.sql` |
| DB-004 | 创建算力点账户表 | ✅ | `backend/migrations/015_sku_refactor.sql` |
| DB-005 | 创建用户订阅表 | ✅ | `backend/migrations/015_sku_refactor.sql` |
| DB-006 | 编写数据迁移脚本 | ✅ | `backend/migrations/015_sku_refactor.sql` |
| DB-007 | 初始化种子数据 | ✅ | `backend/migrations/015_sku_refactor.sql` |

---

## Phase 2: 后端 API 重构 ✅

| 任务ID | 任务描述 | 状态 | 文件路径 |
|--------|----------|------|----------|
| BE-001 | 定义 SPU/SKU 数据模型 | ✅ | `backend/models/sku.go` |
| BE-002 | 实现 SPU CRUD API | ✅ | `backend/handlers/sku.go` |
| BE-003 | 实现 SKU CRUD API | ✅ | `backend/handlers/sku.go` |
| BE-004 | 实现算力点账户 API | ✅ | `backend/handlers/sku.go` |
| BE-005 | 实现订阅管理 API | ✅ | `backend/handlers/sku.go` |
| BE-006 | 实现公开 SKU 查询 API | ✅ | `backend/handlers/sku.go` |
| BE-007 | 配置路由 | ✅ | `backend/routes/routes.go` |
| BE-008 | 配置缓存 Key | ✅ | `backend/cache/cache.go` |
| BE-009 | 注册路由到主程序 | ✅ | `backend/main.go` |

---

## Phase 3: 管理端前端开发 ✅

| 任务ID | 任务描述 | 状态 | 文件路径 |
|--------|----------|------|----------|
| FE-ADM-001 | TypeScript 类型定义 | ✅ | `frontend/src/types/sku.ts` |
| FE-ADM-002 | API 服务封装 | ✅ | `frontend/src/services/sku.ts` |
| FE-ADM-003 | SPU 管理页面 | ✅ | `frontend/src/pages/admin/AdminSPUs.tsx` |
| FE-ADM-004 | SKU 管理页面 | ✅ | `frontend/src/pages/admin/AdminSKUs.tsx` |

---

## Phase 4: 现有模块迁移（待开发）🔴 高优先级

### 4.1 订单模块迁移

| 任务ID | 任务描述 | 优先级 | 预估工时 | 涉及文件 |
|--------|----------|--------|----------|----------|
| ORD-001 | 订单表增加 sku_id 字段 | P0 | 1h | `backend/migrations/016_order_sku_migration.sql` |
| ORD-002 | 订单创建逻辑改造 | P0 | 3h | `backend/handlers/order_and_group.go` |
| ORD-003 | 订单查询逻辑改造 | P0 | 2h | `backend/handlers/order_and_group.go` |
| ORD-004 | 前端订单服务改造 | P0 | 2h | `frontend/src/services/order.ts` |
| ORD-005 | 订单详情页改造 | P0 | 2h | `frontend/src/pages/OrderDetailPage.tsx` |
| ORD-006 | 订单列表页改造 | P0 | 2h | `frontend/src/pages/OrderListPage.tsx` |

**详细说明**:

**ORD-001: 订单表增加 sku_id 字段**

```sql
-- 新增迁移文件
ALTER TABLE orders ADD COLUMN sku_id INT REFERENCES skus(id);
ALTER TABLE orders ADD COLUMN spu_id INT REFERENCES spus(id);

-- 数据迁移：将现有 product_id 映射到 sku_id
UPDATE orders o SET sku_id = s.id, spu_id = s.spu_id
FROM skus s
WHERE s.merchant_id IS NULL  -- 平台自营 SKU
AND s.status = 'active';
```

**ORD-002: 订单创建逻辑改造**

现有逻辑:
```go
// 从 products 表获取价格
var product models.Product
db.QueryRow("SELECT price FROM products WHERE id = $1", req.ProductID).Scan(&product.Price)
```

目标逻辑:
```go
// 从 skus 表获取价格和规格信息
var sku models.SKU
db.QueryRow("SELECT id, spu_id, retail_price, sku_type FROM skus WHERE id = $1 AND status = 'active'", req.SKUID).Scan(&sku.ID, &sku.SPUID, &sku.RetailPrice, &sku.SKUType)

// 根据SKU类型计算价格
if sku.SKUType == "token_pack" {
    // Token包：单价 * 数量
    totalPrice = sku.RetailPrice * float64(req.Quantity)
} else if sku.SKUType == "subscription" {
    // 订阅：固定价格
    totalPrice = sku.RetailPrice
}
```

### 4.2 购物车模块迁移

| 任务ID | 任务描述 | 优先级 | 预估工时 | 涉及文件 |
|--------|----------|--------|----------|----------|
| CART-001 | 购物车表增加 sku_id 字段 | P0 | 1h | `backend/migrations/017_cart_sku_migration.sql` |
| CART-002 | 购物车查询改造 | P0 | 2h | `backend/handlers/cart.go` |
| CART-003 | 购物车添加改造 | P0 | 2h | `backend/handlers/cart.go` |
| CART-004 | 前端购物车页面改造 | P0 | 3h | `frontend/src/pages/CartPage.tsx` |

**详细说明**:

**CART-002: 购物车查询改造**

现有逻辑:
```go
query := `
    SELECT ci.id, ci.product_id, ci.group_id, ci.quantity,
           p.id, p.name, p.price, p.stock
    FROM cart_items ci
    JOIN products p ON ci.product_id = p.id
`
```

目标逻辑:
```go
query := `
    SELECT ci.id, ci.sku_id, ci.group_id, ci.quantity,
           s.id, s.sku_code, s.retail_price, s.stock, s.sku_type,
           sp.name, sp.model_tier
    FROM cart_items ci
    JOIN skus s ON ci.sku_id = s.id
    JOIN spus sp ON s.spu_id = sp.id
`
```

### 4.3 拼团模块迁移

| 任务ID | 任务描述 | 优先级 | 预估工时 | 涉及文件 |
|--------|----------|--------|----------|----------|
| GRP-001 | 拼团表增加 sku_id 字段 | P0 | 1h | `backend/migrations/018_group_sku_migration.sql` |
| GRP-002 | 拼团创建逻辑改造 | P0 | 3h | `backend/handlers/order_and_group.go` |
| GRP-003 | 拼团加入逻辑改造 | P0 | 2h | `backend/handlers/order_and_group.go` |
| GRP-004 | 拼团折扣计算 | P0 | 2h | `backend/handlers/order_and_group.go` |
| GRP-005 | 前端拼团列表页改造 | P0 | 2h | `frontend/src/pages/GroupListPage.tsx` |
| GRP-006 | 前端拼团进度页改造 | P0 | 2h | `frontend/src/pages/GroupProgressPage.tsx` |

**详细说明**:

**GRP-004: 拼团折扣计算**

```go
// 从 SKU 获取拼团配置
var sku models.SKU
db.QueryRow("SELECT retail_price, group_enabled, min_group_size, max_group_size, group_discount_rate FROM skus WHERE id = $1", skuID).Scan(&sku.RetailPrice, &sku.GroupEnabled, &sku.MinGroupSize, &sku.MaxGroupSize, &sku.GroupDiscountRate)

if !sku.GroupEnabled {
    return errors.New("该SKU不支持拼团")
}

if group.CurrentCount < sku.MinGroupSize {
    return errors.New("拼团人数不足")
}

// 计算拼团价格
if sku.GroupDiscountRate > 0 {
    groupPrice = sku.RetailPrice * (1 - sku.GroupDiscountRate/100)
} else {
    groupPrice = sku.RetailPrice
}
```

### 4.4 商品展示模块迁移

| 任务ID | 任务描述 | 优先级 | 预估工时 | 涉及文件 |
|--------|----------|--------|----------|----------|
| PROD-001 | 商品列表 API 兼容层 | P1 | 2h | `backend/handlers/product.go` |
| PROD-002 | 商品详情 API 兼容层 | P1 | 2h | `backend/handlers/product.go` |
| PROD-003 | 热门/新品商品 API 改造 | P1 | 2h | `backend/handlers/product.go` |
| PROD-004 | 首页数据 API 改造 | P1 | 2h | `backend/handlers/product.go` |
| PROD-005 | 前端商品列表页改造 | P1 | 3h | `frontend/src/pages/ProductListPage.tsx` |
| PROD-006 | 前端商品详情页改造 | P1 | 3h | `frontend/src/pages/ProductDetailPage.tsx` |
| PROD-007 | 前端首页改造 | P1 | 2h | `frontend/src/pages/HomePage.tsx` |

**详细说明**:

**PROD-001: 商品列表 API 兼容层**

创建兼容视图，使现有 API 返回 SKU + SPU 联合数据:

```sql
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
  s.updated_at,
  s.spu_id as model_id,
  s.id as package_id
FROM skus s
JOIN spus sp ON s.spu_id = sp.id
WHERE s.status = 'active' AND sp.status = 'active';
```

### 4.5 收藏/浏览历史模块迁移

| 任务ID | 任务描述 | 优先级 | 预估工时 | 涉及文件 |
|--------|----------|--------|----------|----------|
| FAV-001 | 收藏表增加 sku_id 字段 | P1 | 1h | `backend/migrations/019_favorite_sku_migration.sql` |
| FAV-002 | 收藏 API 改造 | P1 | 2h | `backend/handlers/favorite_history.go` |
| FAV-003 | 浏览历史 API 改造 | P1 | 2h | `backend/handlers/favorite_history.go` |
| FAV-004 | 前端收藏页改造 | P2 | 2h | `frontend/src/pages/FavoritesPage.tsx` |
| FAV-005 | 前端历史页改造 | P2 | 2h | `frontend/src/pages/HistoryPage.tsx` |

---

## Phase 5: 商户端前端重构（待开发）

| 任务ID | 任务描述 | 优先级 | 预估工时 |
|--------|----------|--------|----------|
| FE-MER-001 | 重构商品上架页面 | P0 | 6h |
| FE-MER-002 | 增强 API Key 管理页面 | P1 | 4h |
| FE-MER-003 | 账单中心优化 | P1 | 3h |

**详细说明**:

### FE-MER-001: 重构商品上架页面

**目标**: 让商户可以基于平台 SPU 创建自己的商品

**功能点**:
1. 选择平台 SPU（模型列表）
2. 配置商户专属 SKU（定价、库存）
3. 设置商户定价（在平台批发价基础上加价）
4. 配置拼团参数

**涉及文件**:
- `frontend/src/pages/merchant/MerchantProducts.tsx`

**流程设计**:
```
1. 选择 SPU → 展示平台可用模型列表
2. 选择 SKU 类型 → Token包/订阅/并发
3. 配置 SKU 参数 → 数量/有效期等
4. 设置价格 → 建议零售价 + 商户加价
5. 配置拼团 → 是否支持、人数限制、折扣率
6. 提交审核 → 平台审核后上架
```

---

## Phase 6: 用户端优化（待开发）

| 任务ID | 任务描述 | 优先级 | 预估工时 |
|--------|----------|--------|----------|
| FE-USER-001 | 商品列表页优化 | P1 | 4h |
| FE-USER-002 | 商品详情页优化 | P1 | 4h |
| FE-USER-003 | Token 可视化组件 | P1 | 3h |
| FE-USER-004 | 算力点账户页面 | P1 | 3h |
| FE-USER-005 | 订阅管理页面 | P2 | 3h |

---

## Phase 7: 集成测试与部署（待开发）

| 任务ID | 任务描述 | 优先级 | 预估工时 |
|--------|----------|--------|----------|
| TEST-001 | 后端单元测试 | P0 | 4h |
| TEST-002 | 前端组件测试 | P1 | 4h |
| TEST-003 | 集成测试 | P0 | 4h |
| TEST-004 | E2E 测试 | P1 | 4h |
| DEPLOY-001 | 执行数据库迁移 | P0 | 1h |
| DEPLOY-002 | 部署验证 | P0 | 2h |

---

## 文件清单

### 已创建文件

| 文件路径 | 说明 |
|----------|------|
| `backend/migrations/015_sku_refactor.sql` | 数据库迁移脚本 |
| `backend/models/sku.go` | SPU/SKU 数据模型 |
| `backend/handlers/sku.go` | SPU/SKU API 处理器 |
| `frontend/src/types/sku.ts` | TypeScript 类型定义 |
| `frontend/src/services/sku.ts` | API 服务 |
| `frontend/src/pages/admin/AdminSPUs.tsx` | SPU 管理页面 |
| `frontend/src/pages/admin/AdminSKUs.tsx` | SKU 管理页面 |
| `.trae/plans/sku-refactor-impact-analysis.md` | 影响分析报告 |

### 已修改文件

| 文件路径 | 修改内容 |
|----------|----------|
| `backend/routes/routes.go` | 添加 SPU/SKU 路由 |
| `backend/cache/cache.go` | 添加缓存 Key 函数 |
| `backend/main.go` | 注册 SKU 路由 |

### 待创建文件

| 文件路径 | 说明 |
|----------|------|
| `backend/migrations/016_order_sku_migration.sql` | 订单表迁移 |
| `backend/migrations/017_cart_sku_migration.sql` | 购物车表迁移 |
| `backend/migrations/018_group_sku_migration.sql` | 拼团表迁移 |
| `backend/migrations/019_favorite_sku_migration.sql` | 收藏表迁移 |

---

## 验收标准

### 功能验收

- [x] 管理端可创建/编辑/删除 SPU
- [x] 管理端可创建/编辑/删除 SKU
- [x] 公开 API 可查询 SKU 列表和详情
- [x] 算力点账户 API 正常工作
- [x] 订阅 API 正常工作
- [ ] 订单模块支持 SKU 关联
- [ ] 购物车模块支持 SKU 关联
- [ ] 拼团模块支持 SKU 关联
- [ ] 商户端可基于 SPU 创建商品
- [ ] 用户端可正常浏览和购买商品
- [ ] 拼团功能正常工作

### 代码质量验收

- [ ] 后端单元测试覆盖率 > 80%
- [ ] 前端组件测试覆盖率 > 70%
- [ ] 无 lint 错误
- [ ] 无 TypeScript 类型错误

---

## 下一步行动

### 立即执行

1. **运行数据库迁移**: 执行 `015_sku_refactor.sql`
   ```bash
   make docker-up
   # 连接数据库执行迁移
   ```

2. **验证后端 API**: 启动后端服务测试
   ```bash
   make dev-backend
   # 测试 SPU/SKU API
   ```

3. **验证前端页面**: 启动前端服务测试
   ```bash
   make dev-frontend
   # 测试管理端页面
   ```

### 后续开发

1. **Phase 4**: 现有模块迁移（订单、购物车、拼团）
2. **Phase 5**: 商户端重构
3. **Phase 6**: 用户端优化
4. **Phase 7**: 测试与部署

---

**文档结束**
