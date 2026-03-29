# 商户端 SKU 管理功能开发计划

**日期**: 2026-03-29
**状态**: 已确认 ✅
**模式**: 方案1 - 纯分销模式

---

## 1. 业务模式确认

### 1.1 选择方案1：纯分销模式

```
┌─────────────────────────────────────────────────────────────────┐
│                    模式A + 方案1：纯分销模式                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Admin 端 (平台运营)                                              │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ 1. 创建 SPU (DeepSeek-V3, GLM-4, ERNIE 等)               │   │
│  │ 2. 创建 SKU (100K Token包、包月套餐等)                     │   │
│  │ 3. 设置平台统一价格                                        │   │
│  │ 4. 设置分润比例 (如 10%)                                   │   │
│  └─────────────────────────────────────────────────────────┘   │
│                              │                                   │
│                              ▼                                   │
│  商户端 (分销商)                                                  │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ 1. 上传 API Key (已有功能)                                 │   │
│  │    - 选择厂商 (DeepSeek/OpenAI/百度等)                     │   │
│  │    - 输入 API Key + Secret                                │   │
│  │    - 设置配额限制                                          │   │
│  │                                                          │   │
│  │ 2. 选择要销售的 SKU (新功能)                               │   │
│  │    - 浏览平台 SKU 列表                                     │   │
│  │    - 勾选要销售的 SKU                                      │   │
│  │    - 关联 API Key                                         │   │
│  │    - 点击上架                                             │   │
│  │                                                          │   │
│  │ 3. 管理订单和结算 (已有功能)                                │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 1.2 核心特点

| 特点 | 说明 |
|------|------|
| **平台统一定价** | 所有商户销售价格一致 |
| **固定分润** | 商户获得平台设定比例分润 |
| **简单易用** | 商户只需勾选+配置API |
| **价格管控** | 平台完全控制价格体系 |

---

## 2. 数据模型设计

### 2.1 新增 merchant_skus 关联表

```sql
CREATE TABLE merchant_skus (
  id SERIAL PRIMARY KEY,
  merchant_id INT NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
  sku_id INT NOT NULL REFERENCES skus(id) ON DELETE CASCADE,
  
  -- API Key 关联
  api_key_id INT REFERENCES merchant_api_keys(id) ON DELETE SET NULL,
  
  -- 状态
  status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
  
  -- 销量统计
  sales_count BIGINT DEFAULT 0,
  total_sales_amount DECIMAL(15, 2) DEFAULT 0,
  
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  
  UNIQUE(merchant_id, sku_id)
);

CREATE INDEX idx_merchant_skus_merchant ON merchant_skus(merchant_id);
CREATE INDEX idx_merchant_skus_sku ON merchant_skus(sku_id);
CREATE INDEX idx_merchant_skus_status ON merchant_skus(status);
```

### 2.2 SKU 表调整

保持 `skus.merchant_id = NULL`，表示平台自营 SKU。

---

## 3. API 设计

### 3.1 商户端 API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/merchants/skus` | 获取商户已选 SKU 列表 |
| GET | `/api/v1/merchants/skus/available` | 获取平台可用 SKU 列表 |
| POST | `/api/v1/merchants/skus` | 商户选择 SKU 上架 |
| PUT | `/api/v1/merchants/skus/:id` | 更新商户 SKU（状态、API Key） |
| DELETE | `/api/v1/merchants/skus/:id` | 商户取消 SKU |

### 3.2 API 详情

#### POST /api/v1/merchants/skus - 选择 SKU 上架

**请求**:
```json
{
  "sku_id": 1,
  "api_key_id": 5
}
```

**响应**:
```json
{
  "data": {
    "id": 1,
    "merchant_id": 10,
    "sku_id": 1,
    "sku_code": "DEEPSEEK-V3-100K",
    "spu_name": "DeepSeek-V3",
    "retail_price": 9.90,
    "api_key_id": 5,
    "api_key_name": "我的DeepSeek Key",
    "status": "active",
    "sales_count": 0
  }
}
```

---

## 4. 前端设计

### 4.1 商户 SKU 管理页面

**路径**: `/merchant/skus`

**页面结构**:

```
┌─────────────────────────────────────────────────────────────┐
│  商品管理                                    [+ 选择商品上架] │
├─────────────────────────────────────────────────────────────┤
│  状态筛选: [全部] [在售] [已下架]                             │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────┐   │
│  │ SKU列表表格                                          │   │
│  │ - SKU编码 | SPU名称 | 类型 | 平台价格 | 销量 | 状态   │   │
│  │ - 关联API Key | 操作(下架/编辑)                       │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### 4.2 选择商品弹窗

```
┌─────────────────────────────────────────────────────────────┐
│  选择要上架的商品                                    [× 关闭] │
├─────────────────────────────────────────────────────────────┤
│  搜索: [________________]  厂商: [全部 ▼]  类型: [全部 ▼]    │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────┐   │
│  │ □ DeepSeek-V3 - 100K Token包 - ¥9.90                │   │
│  │ □ DeepSeek-V3 - 500K Token包 - ¥39.90               │   │
│  │ □ DeepSeek-V3 - 1M Token包 - ¥69.90                 │   │
│  │ □ GLM-4 - 100K Token包 - ¥12.90                     │   │
│  └─────────────────────────────────────────────────────┘   │
├─────────────────────────────────────────────────────────────┤
│  选择API Key: [我的DeepSeek Key ▼]                          │
│                                                             │
│                              [取消]  [确认上架]              │
└─────────────────────────────────────────────────────────────┘
```

---

## 5. 实施步骤

### Step 1: 修复现有 Bug

**文件**: `backend/handlers/sku.go`

**问题**: CreateSKU 函数尝试从 SPU 查询 merchant_id，但 SPU 表无此字段

**修复**: 将 merchant_id 设为 NULL（平台自营）

### Step 2: 创建数据库迁移

**文件**: `backend/migrations/017_merchant_skus.sql`

创建 `merchant_skus` 关联表

### Step 3: 后端 API 开发

**文件**: `backend/handlers/merchant_sku.go`

实现商户 SKU 管理 API

### Step 4: 前端开发

**文件**:
- `frontend/src/services/merchantSku.ts`
- `frontend/src/pages/merchant/MerchantSKUs.tsx`
- `frontend/src/types/merchantSku.ts`

### Step 5: 测试验证

- 后端单元测试
- 前端类型检查
- 本地功能验证

### Step 6: 部署

- 创建 PR
- CI 通过
- 合并到 main

---

## 6. 文件清单

### 新增文件

| 文件路径 | 说明 |
|----------|------|
| `backend/migrations/017_merchant_skus.sql` | 数据库迁移 |
| `backend/handlers/merchant_sku.go` | 商户 SKU API |
| `backend/models/merchant_sku.go` | 商户 SKU 模型 |
| `frontend/src/services/merchantSku.ts` | 前端服务 |
| `frontend/src/types/merchantSku.ts` | 类型定义 |
| `frontend/src/pages/merchant/MerchantSKUs.tsx` | 管理页面 |

### 修改文件

| 文件路径 | 修改内容 |
|----------|----------|
| `backend/handlers/sku.go` | 修复 CreateSKU Bug |
| `backend/routes/routes.go` | 添加商户 SKU 路由 |

---

## 7. 验收标准

- [ ] 商户可浏览平台 SKU 列表
- [ ] 商户可选择 SKU 并关联 API Key 上架
- [ ] 商户可查看已上架 SKU 列表
- [ ] 商户可下架 SKU
- [ ] 用户购买时正确关联商户
- [ ] 商户获得正确分润

---

**计划已确认，开始实施。**
