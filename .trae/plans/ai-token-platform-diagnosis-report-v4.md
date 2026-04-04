# 拼托托AI Token聚合平台诊断分析报告 v4.0

> 创建时间：2026-04-04
> 版本说明：本版本在v3.0基础上，逐条审视GAP在项目当前代码实际实现情况，前后端拆解分析，给出详细解决方案

---

## 版本演进说明

| 版本 | 主要内容 | 变更说明 |
|------|---------|---------|
| v1.x | 初稿 | 原始诊断内容 |
| v2.0 | 评审优化版 | 解决结构混乱、内容重复、GAP编号不一致问题 |
| v3.0 | 业务流程视角版 | 新增三类角色业务流程串联诊断 |
| v4.0 | 代码实现审视版 | 逐条验证GAP代码实现，前后端拆解，详细解决方案 |

---

## 一、代码实现现状总览

### 1.1 后端代码结构

```
backend/
├── models/           # 数据模型
│   ├── models.go     # User, Order, Merchant, MerchantAPIKey等
│   ├── sku.go        # SPU, SKU, ModelProvider等
│   └── merchant_sku.go # MerchantSKU关联表
├── handlers/         # HTTP处理器
│   ├── api_proxy.go  # API代理（智能路由）
│   ├── merchant_apikey.go # 商户API Key管理
│   ├── merchant.go   # 商户注册/审核
│   ├── sku.go        # SPU/SKU管理
│   └── order_and_group.go # 订单处理
├── billing/          # 计费引擎
│   └── engine.go     # 硬编码定价
├── services/         # 业务服务
│   └── fulfillment_service.go # 订单履约
├── scheduler/        # 定时任务
│   └── settlement.go # 结算调度（框架存在，逻辑不完整）
└── migrations/       # 数据库迁移
```

### 1.2 前端代码结构

```
frontend/src/
├── pages/
│   ├── merchant/
│   │   ├── MerchantAPIKeys.tsx    # 商户API Key管理页面
│   │   ├── MerchantSKUs.tsx       # 商户SKU管理页面
│   │   └── MerchantSettlements.tsx # 商户结算页面
│   ├── admin/
│   │   ├── AdminSPUs.tsx          # SPU管理
│   │   └── AdminSKUs.tsx          # SKU管理
│   └── ProductListPage.tsx        # 用户产品列表
├── services/
│   └── merchant.ts                # 商户API服务
└── types/
    └── index.ts                   # 类型定义
```

### 1.3 关键发现

| 模块 | 实现状态 | 关键问题 |
|------|---------|---------|
| SPU/SKU架构 | ✅ 基础完成 | 缺少用途场景分类 |
| 智能路由 | ⚠️ 简单实现 | 无健康检查、无故障切换、硬编码定价 |
| 商户API Key | ⚠️ 基础实现 | 无验证、无端点配置、无成本定价 |
| 订单履约 | ✅ 已实现 | 履约逻辑完整 |
| 商户结算 | ⚠️ 框架存在 | 无实际结算逻辑、无成本计算 |

---

## 二、GAP代码实现验证总表

### 2.1 用户角色GAP验证

| GAP编号 | 描述 | 后端代码验证 | 前端代码验证 | 实际状态 |
|---------|------|-------------|-------------|---------|
| GAP-U01 | 用户信息字段不完整 | `models.go:User`缺少偏好字段 | 无偏好设置UI | ❌ 未实现 |
| GAP-U02 | 缺少用途场景筛选 | `sku.go`无场景字段 | `ProductListPage.tsx`无场景筛选 | ❌ 未实现 |
| GAP-U03 | 价格排序不完善 | `sku.go`支持基础排序 | UI有排序但不完善 | ⚠️ 部分实现 |
| GAP-U04 | 缺少性能指标排序 | 无性能数据字段 | 无性能排序UI | ❌ 未实现 |
| GAP-U05 | 缺少产品对比功能 | N/A | 无对比组件 | ❌ 未实现 |
| GAP-U06 | 缺少智能推荐 | 无推荐算法 | 无推荐UI | ❌ 未实现 |
| GAP-U07 | 订单状态追踪不完善 | 无通知机制 | 无实时更新 | ⚠️ 部分实现 |
| GAP-U08 | 缺少自动续费功能 | `subscription_renewal.go`框架存在 | 无续费设置UI | ⚠️ 框架存在 |
| GAP-U09 | 智能路由能力不足 | `api_proxy.go`简单负载均衡 | N/A | ⚠️ 简单实现 |
| GAP-U10 | 扣费逻辑硬编码 | `api_proxy.go:calculateTokenCost`硬编码 | N/A | ❌ 硬编码 |
| GAP-U11 | 用量统计不完善 | `api_proxy.go:GetAPIUsageStats`基础统计 | `Consumption.tsx`基础UI | ⚠️ 部分实现 |
| GAP-U12 | 套餐余额查询不完善 | 仅Token余额 | `MyToken.tsx`仅显示Token | ⚠️ 部分实现 |
| GAP-U13 | 账单详情缺少成本明细 | 无成本字段 | 无成本展示 | ❌ 未实现 |
| GAP-U14 | 缺少账单导出功能 | 无导出API | 无导出按钮 | ❌ 未实现 |

### 2.2 平台运营方GAP验证

| GAP编号 | 描述 | 后端代码验证 | 前端代码验证 | 实际状态 |
|---------|------|-------------|-------------|---------|
| GAP-P01 | 定价配置硬编码 | `billing/engine.go`硬编码 | 无定价配置UI | ❌ 硬编码 |
| GAP-P02 | 缺少产品上架审核流程 | `sku.go`无审核流程 | `AdminSPUs.tsx`无审核UI | ❌ 未实现 |
| GAP-P03 | 缺少批量创建工具 | 无批量API | 无批量UI | ❌ 未实现 |
| GAP-P04 | 缺少审核日志 | 无日志表 | 无日志UI | ❌ 未实现 |
| GAP-P05 | 商户资料审核不完善 | `merchant.go`基础实现 | `AdminMerchants.tsx`基础UI | ⚠️ 部分实现 |
| GAP-P06 | 缺少API Key验证 | `merchant_apikey.go`无验证 | 无验证按钮 | ❌ 未实现 |
| GAP-P07 | 审核流程不够完善 | 简单状态更新 | 无流程可视化 | ⚠️ 部分实现 |
| GAP-P08 | 缺少审核结果通知 | 无通知机制 | 无通知UI | ❌ 未实现 |
| GAP-P09 | 缺少路由策略配置 | 无策略配置 | 无配置UI | ❌ 未实现 |
| GAP-P10 | 缺少权重设置 | 无权重字段 | 无权重设置 | ❌ 未实现 |
| GAP-P11 | 缺少健康检查 | 无健康检查服务 | 无健康状态展示 | ❌ 未实现 |
| GAP-P12 | 缺少故障切换 | 无故障切换逻辑 | N/A | ❌ 未实现 |
| GAP-P13 | 缺少监控告警 | 无告警服务 | 无告警UI | ❌ 未实现 |
| GAP-P14 | 缺少结算单生成 | `settlement.go`框架存在 | 无生成触发 | ⚠️ 框架存在 |
| GAP-P15 | 缺少商户确认流程 | 无确认逻辑 | 无确认UI | ❌ 未实现 |
| GAP-P16 | 缺少财务审核流程 | 无审核逻辑 | 无审核UI | ❌ 未实现 |
| GAP-P17 | 缺少打款记录 | 无打款表 | 无打款UI | ❌ 未实现 |
| GAP-P18 | 缺少成本计算 | 无成本字段 | 无成本展示 | ❌ 未实现 |

### 2.3 商户角色GAP验证

| GAP编号 | 描述 | 后端代码验证 | 前端代码验证 | 实际状态 |
|---------|------|-------------|-------------|---------|
| GAP-M01 | 企业资料字段不完整 | `models.go:Merchant`部分字段 | 注册表单不完整 | ⚠️ 部分实现 |
| GAP-M02 | 审核流程不完善 | 简单状态更新 | 无流程可视化 | ⚠️ 部分实现 |
| GAP-M03 | 缺少端点URL配置 | `merchant_api_keys`无字段 | `MerchantAPIKeys.tsx`无输入 | ❌ 未实现 |
| GAP-M04 | 缺少一键验证 | 无验证逻辑 | 无验证按钮 | ❌ 未实现 |
| GAP-M05 | 缺少模型列表自动获取 | 无获取逻辑 | 无展示 | ❌ 未实现 |
| GAP-M06 | 缺少成本定价配置 | 无成本字段 | 无输入框 | ❌ 未实现 |
| GAP-M07 | 缺少健康状态初始化 | 无初始化逻辑 | N/A | ❌ 未实现 |
| GAP-M08 | 缺少SKU成本定价配置 | `merchant_skus`无成本字段 | 无配置UI | ❌ 未实现 |
| GAP-M09 | 缺少上架审核 | 无审核流程 | 无审核UI | ❌ 未实现 |
| GAP-M10 | 资源池状态管理不完善 | 简单状态字段 | 无状态管理UI | ⚠️ 部分实现 |
| GAP-M11 | 缺少结算列表 | `settlement.go`框架存在 | `MerchantSettlements.tsx`框架存在 | ⚠️ 框架存在 |
| GAP-M12 | 缺少结算详情 | 无详情API | 无详情UI | ❌ 未实现 |
| GAP-M13 | 缺少结算确认 | 无确认逻辑 | 无确认UI | ❌ 未实现 |
| GAP-M14 | 缺少打款记录 | 无打款表 | 无打款UI | ❌ 未实现 |
| GAP-M15 | 用量概览不完善 | 基础统计 | 基础UI | ⚠️ 部分实现 |
| GAP-M16 | 缺少详细统计 | 无详细统计 | 无详细UI | ❌ 未实现 |
| GAP-M17 | 缺少导出报表 | 无导出API | 无导出按钮 | ❌ 未实现 |
| GAP-M18 | API调用明细不完善 | 基础日志 | 基础展示 | ⚠️ 部分实现 |

---

## 三、详细开发计划

> 本章节继承v1.0版本详细开发计划，结合已确认决策进行细化

### 3.1 数据模型修改汇总

#### 3.1.1 用户偏好设置（Migration: 024_user_preferences.sql）

**目标**：支持用户个性化偏好，为智能推荐提供基础

```sql
ALTER TABLE users ADD COLUMN IF NOT EXISTS preferred_scenarios JSONB DEFAULT '[]';
ALTER TABLE users ADD COLUMN IF NOT EXISTS budget_level VARCHAR(20) DEFAULT 'standard';
```

**字段说明**：
| 字段 | 类型 | 说明 | 示例 |
|------|------|------|------|
| preferred_scenarios | JSONB | 用户偏好场景列表 | `["chat", "pdf", "code"]` |
| budget_level | VARCHAR(20) | 预算级别 | `economy`, `standard`, `premium` |

---

#### 3.1.2 SPU场景分类（Migration: 025_spu_scenarios.sql）

**目标**：实现产品按用途场景分类，提升用户选购体验

```sql
CREATE TABLE IF NOT EXISTS usage_scenarios (
  id SERIAL PRIMARY KEY,
  code VARCHAR(50) UNIQUE NOT NULL,
  name VARCHAR(100) NOT NULL,
  description TEXT,
  icon_url VARCHAR(500),
  sort_order INT DEFAULT 0,
  status VARCHAR(20) DEFAULT 'active'
);

CREATE TABLE IF NOT EXISTS spu_scenarios (
  spu_id INT NOT NULL REFERENCES spus(id) ON DELETE CASCADE,
  scenario_id INT NOT NULL REFERENCES usage_scenarios(id) ON DELETE CASCADE,
  is_primary BOOLEAN DEFAULT FALSE,
  PRIMARY KEY (spu_id, scenario_id)
);

ALTER TABLE spus ADD COLUMN IF NOT EXISTS avg_latency_ms INT;
ALTER TABLE spus ADD COLUMN IF NOT EXISTS availability_rate DECIMAL(5,2) DEFAULT 99.9;

INSERT INTO usage_scenarios (code, name, description, sort_order) VALUES
('chat', '日常对话', '适用于日常问答、聊天互动', 1),
('pdf', 'PDF处理', '适用于PDF文档解析、摘要、问答', 2),
('image', '图片处理', '适用于图像理解、OCR、图像生成', 3),
('audio', '音频处理', '适用于语音识别、语音合成', 4),
('video', '视频处理', '适用于视频理解、视频生成', 5),
('multimodal', '多模态', '支持多种输入输出格式', 6),
('code', '代码生成', '适用于代码编写、调试、解释', 7),
('reasoning', '复杂推理', '适用于数学推理、逻辑分析', 8);
```

**数据模型关系**：
```
usage_scenarios (场景表)
      │
      │ N:M
      ▼
spu_scenarios (关联表) ────── spus (产品表)
```

---

#### 3.1.3 智能路由核心（Migration: 027_smart_routing.sql）

**目标**：支持多维度智能路由、健康检查、动态定价

```sql
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS api_base_url VARCHAR(500);
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS supported_models JSONB DEFAULT '[]';
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS input_price_per_1k DECIMAL(10,6);
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS output_price_per_1k DECIMAL(10,6);
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS health_status VARCHAR(20) DEFAULT 'unknown';
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS health_checked_at TIMESTAMP;
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS avg_latency_ms INT;
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS success_rate DECIMAL(5,2) DEFAULT 100;
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS consecutive_failures INT DEFAULT 0;
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS routing_weight INT DEFAULT 100;
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS health_check_interval INT DEFAULT 300;
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS health_check_level VARCHAR(20) DEFAULT 'medium';

ALTER TABLE api_usage_logs ADD COLUMN IF NOT EXISTS merchant_id INT;
ALTER TABLE api_usage_logs ADD COLUMN IF NOT EXISTS user_cost DECIMAL(15,6);
ALTER TABLE api_usage_logs ADD COLUMN IF NOT EXISTS merchant_cost DECIMAL(15,6);
ALTER TABLE api_usage_logs ADD COLUMN IF NOT EXISTS platform_profit DECIMAL(15,6);

CREATE TABLE IF NOT EXISTS routing_strategies (
  id SERIAL PRIMARY KEY,
  name VARCHAR(50) NOT NULL,
  strategy_type VARCHAR(20) NOT NULL,
  weight_config JSONB,
  is_default BOOLEAN DEFAULT FALSE,
  status VARCHAR(20) DEFAULT 'active'
);

CREATE TABLE IF NOT EXISTS api_key_health_history (
  id SERIAL PRIMARY KEY,
  api_key_id INT NOT NULL REFERENCES merchant_api_keys(id),
  status VARCHAR(20) NOT NULL,
  latency_ms INT,
  error_message TEXT,
  checked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**字段说明**：
| 字段 | 类型 | 说明 | 决策依据 |
|------|------|------|---------|
| api_base_url | VARCHAR(500) | API端点URL | 商户可配置 |
| supported_models | JSONB | 支持的模型列表 | 自动获取或手动配置 |
| input_price_per_1k | DECIMAL(10,6) | 输入Token成本定价 | 动态定价决策 |
| output_price_per_1k | DECIMAL(10,6) | 输出Token成本定价 | 动态定价决策 |
| health_check_level | VARCHAR(20) | 健康检查级别 | 多级可配置决策 |
| health_check_interval | INT | 健康检查间隔(秒) | 多级可配置决策 |

---

#### 3.1.4 商户API Key验证（Migration: 029_merchant_verification.sql）

**目标**：支持API Key异步验证，验证失败禁止路由

```sql
CREATE TABLE IF NOT EXISTS api_key_verifications (
  id SERIAL PRIMARY KEY,
  api_key_id INT NOT NULL REFERENCES merchant_api_keys(id),
  verification_type VARCHAR(50) NOT NULL,
  status VARCHAR(20) NOT NULL,
  details JSONB,
  error_message TEXT,
  verified_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS verified_at TIMESTAMP;
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS verification_result JSONB;
```

**验证类型**：
| verification_type | 说明 | 验证内容 |
|-------------------|------|---------|
| connection | 连接验证 | API Key是否有效 |
| models | 模型列表 | 获取支持的模型 |
| pricing | 定价验证 | 获取成本定价信息 |

---

#### 3.1.5 结算系统增强（Migration: 030_settlement_enhancement.sql）

**目标**：支持商户结算、成本计算、打款记录

```sql
CREATE TABLE IF NOT EXISTS merchant_accounts (
  id SERIAL PRIMARY KEY,
  merchant_id INT UNIQUE NOT NULL REFERENCES merchants(id),
  balance DECIMAL(15,2) DEFAULT 0,
  pending_balance DECIMAL(15,2) DEFAULT 0,
  total_earned DECIMAL(15,2) DEFAULT 0,
  total_settled DECIMAL(15,2) DEFAULT 0,
  bank_account VARCHAR(50),
  bank_name VARCHAR(100),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS merchant_settlement_items (
  id SERIAL PRIMARY KEY,
  settlement_id INT NOT NULL REFERENCES merchant_settlements(id),
  api_usage_log_id INT NOT NULL REFERENCES api_usage_logs(id),
  user_cost DECIMAL(15,6),
  merchant_cost DECIMAL(15,6),
  platform_profit DECIMAL(15,6)
);

CREATE TABLE IF NOT EXISTS settlement_payments (
  id SERIAL PRIMARY KEY,
  settlement_id INT NOT NULL REFERENCES merchant_settlements(id),
  merchant_id INT NOT NULL REFERENCES merchants(id),
  amount DECIMAL(15,2) NOT NULL,
  payment_method VARCHAR(50),
  transaction_id VARCHAR(100),
  status VARCHAR(20) DEFAULT 'pending',
  paid_at TIMESTAMP
);

ALTER TABLE merchant_settlements ADD COLUMN IF NOT EXISTS merchant_confirmed BOOLEAN DEFAULT FALSE;
ALTER TABLE merchant_settlements ADD COLUMN IF NOT EXISTS finance_approved BOOLEAN DEFAULT FALSE;
```

---

#### 3.1.6 商户资料增强（Migration: 031_merchant_profile_enhancement.sql）

**目标**：完善商户资料，支持结算周期配置

```sql
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS bank_account VARCHAR(50);
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS bank_name VARCHAR(100);
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS tax_id VARCHAR(50);
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS settlement_cycle VARCHAR(20) DEFAULT 'monthly';
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS monthly_threshold DECIMAL(15,2) DEFAULT 10000.00;
```

---

#### 3.1.7 商户SKU成本定价（Migration: 032_merchant_sku_cost.sql）

**目标**：支持商户级别的SKU成本定价

```sql
ALTER TABLE merchant_skus ADD COLUMN IF NOT EXISTS cost_input_rate DECIMAL(10,6);
ALTER TABLE merchant_skus ADD COLUMN IF NOT EXISTS cost_output_rate DECIMAL(10,6);
ALTER TABLE merchant_skus ADD COLUMN IF NOT EXISTS profit_margin DECIMAL(5,2) DEFAULT 20.00;
```

---

### 3.2 后端开发计划

#### 3.2.1 健康检查服务（services/health_checker.go）【P0】

**目标**：实时监控商户API健康状态，支持多级检查频率

**核心结构**：
```go
package services

import (
    "context"
    "database/sql"
    "sync"
    "time"
)

type HealthChecker struct {
    db             *sql.DB
    checkInterval  time.Duration
    providers      map[int]*ProviderHealth
    providersMutex sync.RWMutex
    alertThreshold int
    stopCh         chan struct{}
}

type ProviderHealth struct {
    APIKeyID         int
    MerchantID       int
    Status           string        // healthy, degraded, unhealthy
    LastCheck        time.Time
    LastSuccess      time.Time
    ConsecutiveFail  int
    AvgLatency       time.Duration
    SuccessRate      float64
    HealthCheckLevel string
}

type HealthCheckLevel string

const (
    HealthCheckHigh   HealthCheckLevel = "high"    // 1分钟
    HealthCheckMedium HealthCheckLevel = "medium"  // 5分钟
    HealthCheckLow    HealthCheckLevel = "low"     // 30分钟
    HealthCheckDaily  HealthCheckLevel = "daily"   // 24小时
)

func (l HealthCheckLevel) Interval() time.Duration {
    switch l {
    case HealthCheckHigh:
        return 1 * time.Minute
    case HealthCheckMedium:
        return 5 * time.Minute
    case HealthCheckLow:
        return 30 * time.Minute
    case HealthCheckDaily:
        return 24 * time.Hour
    default:
        return 5 * time.Minute
    }
}

func NewHealthChecker(db *sql.DB) *HealthChecker {
    return &HealthChecker{
        db:             db,
        providers:      make(map[int]*ProviderHealth),
        alertThreshold: 3,
        stopCh:         make(chan struct{}),
    }
}

func (h *HealthChecker) Start() {
    go h.run()
}

func (h *HealthChecker) Stop() {
    close(h.stopCh)
}

func (h *HealthChecker) run() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            h.checkAllProviders()
        case <-h.stopCh:
            return
        }
    }
}

func (h *HealthChecker) checkAllProviders() {
    apiKeys, err := h.getActiveAPIKeys()
    if err != nil {
        return
    }
    
    var wg sync.WaitGroup
    for _, key := range apiKeys {
        if h.shouldCheck(key) {
            wg.Add(1)
            go func(apiKey *MerchantAPIKey) {
                defer wg.Done()
                h.checkProvider(apiKey)
            }(key)
        }
    }
    wg.Wait()
}

func (h *HealthChecker) shouldCheck(key *MerchantAPIKey) bool {
    interval := HealthCheckLevel(key.HealthCheckLevel).Interval()
    lastCheck := key.HealthCheckedAt
    return time.Since(lastCheck) >= interval
}

func (h *HealthChecker) checkProvider(apiKey *MerchantAPIKey) {
    start := time.Now()
    
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    err := h.pingProvider(ctx, apiKey)
    latency := time.Since(start)
    
    h.providersMutex.Lock()
    defer h.providersMutex.Unlock()
    
    health, exists := h.providers[apiKey.ID]
    if !exists {
        health = &ProviderHealth{
            APIKeyID: apiKey.ID,
            MerchantID: apiKey.MerchantID,
        }
        h.providers[apiKey.ID] = health
    }
    
    health.LastCheck = time.Now()
    
    if err != nil {
        health.ConsecutiveFail++
        if health.ConsecutiveFail >= h.alertThreshold {
            health.Status = "unhealthy"
        } else {
            health.Status = "degraded"
        }
        h.recordFailure(apiKey.ID, err, latency)
    } else {
        health.ConsecutiveFail = 0
        health.Status = "healthy"
        health.LastSuccess = time.Now()
        health.AvgLatency = (health.AvgLatency + latency) / 2
        h.recordSuccess(apiKey.ID, latency)
    }
    
    h.updateDB(apiKey.ID, health)
}

func (h *HealthChecker) RecordSuccess(apiKeyID int) {
    h.providersMutex.Lock()
    defer h.providersMutex.Unlock()
    
    if health, exists := h.providers[apiKeyID]; exists {
        health.ConsecutiveFail = 0
        health.Status = "healthy"
        health.LastSuccess = time.Now()
    }
}

func (h *HealthChecker) RecordFailure(apiKeyID int, err error) {
    h.providersMutex.Lock()
    defer h.providersMutex.Unlock()
    
    if health, exists := h.providers[apiKeyID]; exists {
        health.ConsecutiveFail++
        if health.ConsecutiveFail >= h.alertThreshold {
            health.Status = "unhealthy"
        } else {
            health.Status = "degraded"
        }
    }
}

func (h *HealthChecker) GetHealthyProviders() []*ProviderHealth {
    h.providersMutex.RLock()
    defer h.providersMutex.RUnlock()
    
    var healthy []*ProviderHealth
    for _, p := range h.providers {
        if p.Status == "healthy" || p.Status == "degraded" {
            healthy = append(healthy, p)
        }
    }
    return healthy
}
```

**API端点**：
```
GET  /api/v1/admin/api-keys/health           - 获取所有API Key健康状态
POST /api/v1/admin/api-keys/:id/health-check - 手动触发健康检查
```

---

#### 3.2.2 智能路由服务（services/smart_router.go）【P0】

**目标**：基于多维度权重的智能路由，支持平台统一配置策略

**核心结构**：
```go
package services

import (
    "database/sql"
    "sort"
    "time"
)

type SmartRouter struct {
    db             *sql.DB
    healthChecker  *HealthChecker
    pricingService *PricingService
    strategy       RoutingStrategy
}

type RoutingStrategy string

const (
    StrategyPrice     RoutingStrategy = "price"      // 价格优先
    StrategyLatency   RoutingStrategy = "latency"    // 延迟优先
    StrategyBalanced  RoutingStrategy = "balanced"   // 均衡策略
    StrategyCost      RoutingStrategy = "cost"       // 成本优先
)

type RoutingCandidate struct {
    APIKeyID      int
    MerchantID    int
    Provider      string
    
    HealthScore   float64  // 健康分数 0-1
    PriceScore    float64  // 价格分数
    LatencyScore  float64  // 延迟分数
    QuotaScore    float64  // 配额分数
    FinalScore    float64  // 综合分数
    
    InputPrice    float64
    OutputPrice   float64
    AvgLatency    time.Duration
    QuotaRemain   float64
    Weight        int
}

func NewSmartRouter(db *sql.DB, hc *HealthChecker, ps *PricingService) *SmartRouter {
    return &SmartRouter{
        db:             db,
        healthChecker:  hc,
        pricingService: ps,
        strategy:       StrategyBalanced, // 默认均衡策略
    }
}

func (r *SmartRouter) SetStrategy(strategy RoutingStrategy) {
    r.strategy = strategy
}

func (r *SmartRouter) SelectProvider(model string) (*RoutingCandidate, error) {
    candidates, err := r.getCandidates(model)
    if err != nil {
        return nil, err
    }
    
    candidates = r.filterUnhealthy(candidates)
    
    if len(candidates) == 0 {
        return nil, ErrNoHealthyProvider
    }
    
    for i := range candidates {
        r.calculateScores(&candidates[i])
    }
    
    sort.Slice(candidates, func(i, j int) bool {
        return candidates[i].FinalScore > candidates[j].FinalScore
    })
    
    return &candidates[0], nil
}

func (r *SmartRouter) getCandidates(model string) ([]RoutingCandidate, error) {
    rows, err := r.db.Query(`
        SELECT 
            mak.id, mak.merchant_id, mak.provider,
            mak.input_price_per_1k, mak.output_price_per_1k,
            mak.avg_latency_ms, mak.quota_limit - mak.quota_used as quota_remain,
            mak.routing_weight, mak.health_status
        FROM merchant_api_keys mak
        WHERE mak.status = 'active'
          AND mak.health_status IN ('healthy', 'degraded')
          AND mak.verified_at IS NOT NULL
          AND (mak.supported_models = '[]'::jsonb OR $1 = ANY(SELECT jsonb_array_elements_text(mak.supported_models)))
        ORDER BY mak.routing_weight DESC
    `, model)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var candidates []RoutingCandidate
    for rows.Next() {
        var c RoutingCandidate
        var latencyMs int
        var healthStatus string
        err := rows.Scan(
            &c.APIKeyID, &c.MerchantID, &c.Provider,
            &c.InputPrice, &c.OutputPrice,
            &latencyMs, &c.QuotaRemain,
            &c.Weight, &healthStatus,
        )
        if err != nil {
            continue
        }
        c.AvgLatency = time.Duration(latencyMs) * time.Millisecond
        candidates = append(candidates, c)
    }
    
    return candidates, nil
}

func (r *SmartRouter) filterUnhealthy(candidates []RoutingCandidate) []RoutingCandidate {
    var healthy []RoutingCandidate
    for _, c := range candidates {
        health := r.healthChecker.GetHealth(c.APIKeyID)
        if health != nil && health.Status != "unhealthy" {
            healthy = append(healthy, c)
        }
    }
    return healthy
}

func (r *SmartRouter) calculateScores(c *RoutingCandidate) {
    c.HealthScore = r.calcHealthScore(c)
    c.PriceScore = r.calcPriceScore(c)
    c.LatencyScore = r.calcLatencyScore(c)
    c.QuotaScore = r.calcQuotaScore(c)
    
    switch r.strategy {
    case StrategyPrice:
        c.FinalScore = 0.5*c.PriceScore + 0.2*c.HealthScore + 0.2*c.LatencyScore + 0.1*c.QuotaScore
    case StrategyLatency:
        c.FinalScore = 0.5*c.LatencyScore + 0.2*c.HealthScore + 0.2*c.PriceScore + 0.1*c.QuotaScore
    case StrategyCost:
        c.FinalScore = 0.5*c.PriceScore + 0.3*c.HealthScore + 0.1*c.LatencyScore + 0.1*c.QuotaScore
    default: // balanced
        c.FinalScore = 0.3*c.HealthScore + 0.3*c.PriceScore + 0.2*c.LatencyScore + 0.2*c.QuotaScore
    }
}

func (r *SmartRouter) calcHealthScore(c *RoutingCandidate) float64 {
    health := r.healthChecker.GetHealth(c.APIKeyID)
    if health == nil {
        return 0.5
    }
    
    score := 1.0
    if health.Status == "degraded" {
        score = 0.7
    }
    score -= float64(health.ConsecutiveFail) * 0.1
    if score < 0 {
        score = 0
    }
    return score
}

func (r *SmartRouter) calcPriceScore(c *RoutingCandidate) float64 {
    return 1.0 / (1.0 + c.InputPrice + c.OutputPrice)
}

func (r *SmartRouter) calcLatencyScore(c *RoutingCandidate) float64 {
    return 1.0 / (1.0 + float64(c.AvgLatency.Milliseconds())/1000.0)
}

func (r *SmartRouter) calcQuotaScore(c *RoutingCandidate) float64 {
    return c.QuotaRemain / (c.QuotaRemain + 1000)
}

func (r *SmartRouter) ExecuteWithFallback(req *APIProxyRequest, maxRetries int) (*APIResponse, error) {
    candidates, err := r.getCandidates(req.Model)
    if err != nil {
        return nil, err
    }
    
    candidates = r.filterUnhealthy(candidates)
    
    if len(candidates) == 0 {
        return nil, ErrNoHealthyProvider
    }
    
    var lastErr error
    for i := 0; i < min(maxRetries, len(candidates)); i++ {
        candidate := &candidates[i]
        
        resp, err := r.executeRequest(req, candidate)
        if err == nil {
            r.healthChecker.RecordSuccess(candidate.APIKeyID)
            return resp, nil
        }
        
        r.healthChecker.RecordFailure(candidate.APIKeyID, err)
        lastErr = err
    }
    
    return nil, fmt.Errorf("all providers failed: %w", lastErr)
}
```

---

#### 3.2.3 动态定价服务（services/pricing_service.go）【P0】

**目标**：从数据库读取定价，支持动态更新和缓存

**核心结构**：
```go
package services

import (
    "database/sql"
    "sync"
    "time"
)

type PricingService struct {
    db         *sql.DB
    cache      map[int]*PricingData
    cacheMutex sync.RWMutex
    cacheTTL   time.Duration
}

type PricingData struct {
    SKUID          int
    SPUID          int
    MerchantID     *int
    
    RetailPrice    float64
    InputRate      float64
    OutputRate     float64
    
    CostInputRate  float64
    CostOutputRate float64
    
    BillingType    string
    UpdatedAt      time.Time
}

type CostResult struct {
    UserCost       float64
    MerchantCost   float64
    PlatformProfit float64
    InputTokens    int
    OutputTokens   int
}

func NewPricingService(db *sql.DB) *PricingService {
    return &PricingService{
        db:       db,
        cache:    make(map[int]*PricingData),
        cacheTTL: 5 * time.Minute,
    }
}

func (s *PricingService) GetPricing(skuID int) (*PricingData, error) {
    s.cacheMutex.RLock()
    if cached, ok := s.cache[skuID]; ok {
        if time.Since(cached.UpdatedAt) < s.cacheTTL {
            s.cacheMutex.RUnlock()
            return cached, nil
        }
    }
    s.cacheMutex.RUnlock()
    
    return s.loadFromDB(skuID)
}

func (s *PricingService) loadFromDB(skuID int) (*PricingData, error) {
    var pricing PricingData
    var updatedAt time.Time
    
    err := s.db.QueryRow(`
        SELECT 
            s.id as sku_id, s.spu_id, s.retail_price,
            spu.input_price_per_1k, spu.output_price_per_1k,
            s.cost_input_rate, s.cost_output_rate,
            s.sku_type, GREATEST(s.updated_at, spu.updated_at) as updated_at
        FROM skus s
        JOIN spus spu ON s.spu_id = spu.id
        WHERE s.id = $1 AND s.status = 'active'
    `, skuID).Scan(
        &pricing.SKUID, &pricing.SPUID, &pricing.RetailPrice,
        &pricing.InputRate, &pricing.OutputRate,
        &pricing.CostInputRate, &pricing.CostOutputRate,
        &pricing.BillingType, &updatedAt,
    )
    if err != nil {
        return nil, err
    }
    
    pricing.UpdatedAt = time.Now()
    
    s.cacheMutex.Lock()
    s.cache[skuID] = &pricing
    s.cacheMutex.Unlock()
    
    return &pricing, nil
}

func (s *PricingService) GetMerchantPricing(apiKeyID int) (*PricingData, error) {
    var pricing PricingData
    
    err := s.db.QueryRow(`
        SELECT 
            id, merchant_id, input_price_per_1k, output_price_per_1k
        FROM merchant_api_keys
        WHERE id = $1 AND status = 'active'
    `, apiKeyID).Scan(
        &pricing.SKUID, &pricing.MerchantID,
        &pricing.CostInputRate, &pricing.CostOutputRate,
    )
    if err != nil {
        return nil, err
    }
    
    return &pricing, nil
}

func (s *PricingService) CalculateCost(skuID int, inputTokens, outputTokens int) (*CostResult, error) {
    pricing, err := s.GetPricing(skuID)
    if err != nil {
        return nil, err
    }
    
    userCost := pricing.InputRate*float64(inputTokens)/1000 +
        pricing.OutputRate*float64(outputTokens)/1000
    
    merchantCost := pricing.CostInputRate*float64(inputTokens)/1000 +
        pricing.CostOutputRate*float64(outputTokens)/1000
    
    return &CostResult{
        UserCost:       userCost,
        MerchantCost:   merchantCost,
        PlatformProfit: userCost - merchantCost,
        InputTokens:    inputTokens,
        OutputTokens:   outputTokens,
    }, nil
}

func (s *PricingService) InvalidateCache(skuID int) {
    s.cacheMutex.Lock()
    delete(s.cache, skuID)
    s.cacheMutex.Unlock()
}

func (s *PricingService) RefreshCache() {
    s.cacheMutex.Lock()
    s.cache = make(map[int]*PricingData)
    s.cacheMutex.Unlock()
}
```

---

#### 3.2.4 API Key验证服务（services/api_key_validator.go）【P0】

**目标**：异步验证商户API Key，验证失败禁止路由

**核心结构**：
```go
package services

import (
    "context"
    "database/sql"
    "encoding/json"
    "time"
)

type APIKeyValidator struct {
    db *sql.DB
}

type VerificationResult struct {
    APIKeyID       int
    Type           string
    Status         string
    Details        map[string]interface{}
    ErrorMessage   string
    SupportedModels []string
}

func NewAPIKeyValidator(db *sql.DB) *APIKeyValidator {
    return &APIKeyValidator{db: db}
}

func (v *APIKeyValidator) ValidateAsync(apiKeyID int) {
    go func() {
        result := v.validateAPIKey(apiKeyID)
        v.updateVerificationStatus(apiKeyID, result)
    }()
}

func (v *APIKeyValidator) validateAPIKey(apiKeyID int) *VerificationResult {
    result := &VerificationResult{
        APIKeyID: apiKeyID,
        Status:   "pending",
        Details:  make(map[string]interface{}),
    }
    
    var apiKey, apiSecret, apiBaseURL, provider string
    err := v.db.QueryRow(`
        SELECT api_key_encrypted, api_secret_encrypted, api_base_url, provider
        FROM merchant_api_keys WHERE id = $1
    `, apiKeyID).Scan(&apiKey, &apiSecret, &apiBaseURL, &provider)
    if err != nil {
        result.Status = "failed"
        result.ErrorMessage = err.Error()
        return result
    }
    
    decryptedKey, err := utils.Decrypt(apiKey)
    if err != nil {
        result.Status = "failed"
        result.ErrorMessage = "Failed to decrypt API key"
        return result
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    result.Type = "connection"
    if err := v.testConnection(ctx, provider, decryptedKey, apiBaseURL); err != nil {
        result.Status = "failed"
        result.ErrorMessage = err.Error()
        return result
    }
    
    result.Type = "models"
    models, err := v.fetchSupportedModels(ctx, provider, decryptedKey, apiBaseURL)
    if err != nil {
        result.Status = "partial"
        result.ErrorMessage = err.Error()
    } else {
        result.SupportedModels = models
        result.Details["supported_models"] = models
    }
    
    result.Type = "pricing"
    pricing, err := v.fetchPricing(ctx, provider, decryptedKey, apiBaseURL)
    if err == nil {
        result.Details["pricing"] = pricing
    }
    
    if result.Status == "pending" {
        result.Status = "success"
    }
    
    return result
}

func (v *APIKeyValidator) testConnection(ctx context.Context, provider, apiKey, baseURL string) error {
    client := NewProviderClient(provider, apiKey, baseURL)
    return client.Ping(ctx)
}

func (v *APIKeyValidator) fetchSupportedModels(ctx context.Context, provider, apiKey, baseURL string) ([]string, error) {
    client := NewProviderClient(provider, apiKey, baseURL)
    return client.ListModels(ctx)
}

func (v *APIKeyValidator) fetchPricing(ctx context.Context, provider, apiKey, baseURL string) (map[string]interface{}, error) {
    client := NewProviderClient(provider, apiKey, baseURL)
    return client.GetPricing(ctx)
}

func (v *APIKeyValidator) updateVerificationStatus(apiKeyID int, result *VerificationResult) {
    detailsJSON, _ := json.Marshal(result.Details)
    
    v.db.Exec(`
        INSERT INTO api_key_verifications (api_key_id, verification_type, status, details, error_message, verified_at)
        VALUES ($1, $2, $3, $4, $5, NOW())
    `, apiKeyID, result.Type, result.Status, detailsJSON, result.ErrorMessage)
    
    if result.Status == "success" || result.Status == "partial" {
        v.db.Exec(`
            UPDATE merchant_api_keys 
            SET verified_at = NOW(), 
                verification_result = $1,
                supported_models = $2
            WHERE id = $3
        `, detailsJSON, result.SupportedModels, apiKeyID)
    }
}
```

---

#### 3.2.5 结算服务（services/settlement_service.go）【P1】

**目标**：自动生成结算单，支持商户确认和财务审核

**核心结构**：
```go
package services

import (
    "database/sql"
    "time"
)

type SettlementService struct {
    db *sql.DB
}

type SettlementPeriod struct {
    MerchantID  int
    PeriodStart time.Time
    PeriodEnd   time.Time
    Cycle       string
}

func NewSettlementService(db *sql.DB) *SettlementService {
    return &SettlementService{db: db}
}

func (s *SettlementService) GenerateMonthlySettlements() error {
    now := time.Now()
    periodStart := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, time.UTC)
    periodEnd := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).Add(-time.Second)
    
    merchants, err := s.getMerchantsForSettlement(periodStart, periodEnd)
    if err != nil {
        return err
    }
    
    for _, merchant := range merchants {
        if err := s.GenerateMerchantSettlement(merchant, periodStart, periodEnd); err != nil {
            continue
        }
    }
    
    return nil
}

func (s *SettlementService) GenerateMerchantSettlement(merchantID int, periodStart, periodEnd time.Time) error {
    tx, err := s.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    var settlementID int
    err = tx.QueryRow(`
        INSERT INTO merchant_settlements (merchant_id, period_start, period_end, status)
        VALUES ($1, $2, $3, 'pending')
        RETURNING id
    `, merchantID, periodStart, periodEnd).Scan(&settlementID)
    if err != nil {
        return err
    }
    
    rows, err := tx.Query(`
        SELECT id, user_cost, merchant_cost, platform_profit
        FROM api_usage_logs
        WHERE merchant_id = $1
          AND created_at >= $2
          AND created_at < $3
          AND settlement_id IS NULL
    `, merchantID, periodStart, periodEnd)
    if err != nil {
        return err
    }
    defer rows.Close()
    
    var totalMerchantCost, totalPlatformProfit float64
    
    for rows.Next() {
        var logID int
        var userCost, merchantCost, platformProfit float64
        if err := rows.Scan(&logID, &userCost, &merchantCost, &platformProfit); err != nil {
            continue
        }
        
        _, err := tx.Exec(`
            INSERT INTO merchant_settlement_items (settlement_id, api_usage_log_id, user_cost, merchant_cost, platform_profit)
            VALUES ($1, $2, $3, $4, $5)
        `, settlementID, logID, userCost, merchantCost, platformProfit)
        if err != nil {
            continue
        }
        
        totalMerchantCost += merchantCost
        totalPlatformProfit += platformProfit
    }
    
    _, err = tx.Exec(`
        UPDATE merchant_settlements 
        SET total_merchant_cost = $1, platform_profit = $2
        WHERE id = $3
    `, totalMerchantCost, totalPlatformProfit, settlementID)
    if err != nil {
        return err
    }
    
    _, err = tx.Exec(`
        UPDATE api_usage_logs 
        SET settlement_id = $1
        WHERE merchant_id = $2 AND created_at >= $3 AND created_at < $4
    `, settlementID, merchantID, periodStart, periodEnd)
    if err != nil {
        return err
    }
    
    return tx.Commit()
}

func (s *SettlementService) ConfirmSettlement(settlementID, merchantID int) error {
    result, err := s.db.Exec(`
        UPDATE merchant_settlements 
        SET merchant_confirmed = TRUE, confirmed_at = NOW()
        WHERE id = $1 AND merchant_id = $2 AND status = 'pending'
    `, settlementID, merchantID)
    if err != nil {
        return err
    }
    
    rows, _ := result.RowsAffected()
    if rows == 0 {
        return ErrSettlementNotFound
    }
    
    return nil
}

func (s *SettlementService) ApproveSettlement(settlementID, adminID int) error {
    tx, err := s.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    var merchantID int
    var totalMerchantCost float64
    err = tx.QueryRow(`
        SELECT merchant_id, total_merchant_cost 
        FROM merchant_settlements 
        WHERE id = $1 AND merchant_confirmed = TRUE AND status = 'pending'
    `, settlementID).Scan(&merchantID, &totalMerchantCost)
    if err != nil {
        return ErrSettlementNotFound
    }
    
    _, err = tx.Exec(`
        UPDATE merchant_settlements 
        SET finance_approved = TRUE, approved_at = NOW(), approved_by = $1, status = 'approved'
        WHERE id = $2
    `, adminID, settlementID)
    if err != nil {
        return err
    }
    
    _, err = tx.Exec(`
        UPDATE merchant_accounts 
        SET pending_balance = pending_balance - $1,
            balance = balance + $1,
            total_settled = total_settled + $1
        WHERE merchant_id = $2
    `, totalMerchantCost, merchantID)
    if err != nil {
        return err
    }
    
    return tx.Commit()
}
```

---

#### 3.2.6 API代理重构（handlers/api_proxy.go）【P0】

**目标**：集成智能路由、动态定价、健康检查

**核心修改**：
```go
func ProxyAPIRequest(c *gin.Context) {
    userID := c.GetInt("user_id")
    var req APIProxyRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    smartRouter := services.GetSmartRouter()
    pricingService := services.GetPricingService()
    billingEngine := services.GetBillingEngine()
    
    candidate, err := smartRouter.SelectProvider(req.Model)
    if err != nil {
        c.JSON(503, gin.H{"error": "No healthy provider available"})
        return
    }
    
    start := time.Now()
    resp, err := executeAPIRequest(c.Request.Context(), req, candidate)
    latency := time.Since(start)
    
    if err != nil {
        smartRouter.healthChecker.RecordFailure(candidate.APIKeyID, err)
        resp, err = smartRouter.ExecuteWithFallback(&req, 3)
        if err != nil {
            c.JSON(502, gin.H{"error": err.Error()})
            return
        }
    }
    
    inputTokens, outputTokens := extractTokenUsage(resp)
    
    costResult, err := pricingService.CalculateCostFromMerchant(
        candidate.APIKeyID, inputTokens, outputTokens,
    )
    if err != nil {
        costResult = &services.CostResult{
            UserCost:       0,
            MerchantCost:   0,
            PlatformProfit: 0,
        }
    }
    
    if err := billingEngine.DeductFromUserAsset(userID, costResult.UserCost, int64(inputTokens+outputTokens)); err != nil {
        c.JSON(402, gin.H{"error": "Insufficient balance"})
        return
    }
    
    tx, _ := db.Begin()
    tx.Exec(`
        INSERT INTO api_usage_logs (
            user_id, merchant_id, key_id, request_id, 
            provider, model, status_code, latency_ms,
            input_tokens, output_tokens,
            user_cost, merchant_cost, platform_profit
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
    `, userID, candidate.MerchantID, candidate.APIKeyID, generateRequestID(),
        candidate.Provider, req.Model, resp.StatusCode, latency.Milliseconds(),
        inputTokens, outputTokens,
        costResult.UserCost, costResult.MerchantCost, costResult.PlatformProfit,
    )
    
    tx.Exec(`
        UPDATE merchant_api_keys 
        SET quota_used = quota_used + $1,
            avg_latency_ms = (avg_latency_ms + $2) / 2,
            health_checked_at = NOW()
        WHERE id = $3
    `, costResult.MerchantCost, latency.Milliseconds(), candidate.APIKeyID)
    
    tx.Commit()
    
    smartRouter.healthChecker.RecordSuccess(candidate.APIKeyID)
    
    c.Data(resp.StatusCode, "application/json", resp.Body)
}
```

---

### 3.3 前端开发计划

#### 3.3.1 商户API Key管理页面（pages/merchant/MerchantAPIKeys.tsx）【P0】

**目标**：支持API Key验证、成本配置、健康状态展示

**新增功能**：
- API端点URL输入
- 一键验证按钮
- 成本定价配置
- 健康状态展示
- 支持模型列表展示

**核心组件**：
```typescript
interface MerchantAPIKey {
  id: number;
  name: string;
  provider: string;
  api_base_url: string;
  supported_models: string[];
  input_price_per_1k: number;
  output_price_per_1k: number;
  health_status: 'healthy' | 'degraded' | 'unhealthy' | 'unknown';
  health_check_level: 'high' | 'medium' | 'low' | 'daily';
  verified_at: string | null;
  verification_result: {
    supported_models: string[];
    pricing: Record<string, any>;
  } | null;
  quota_limit: number;
  quota_used: number;
  avg_latency_ms: number;
  success_rate: number;
}

const MerchantAPIKeys: React.FC = () => {
  const [apiKeys, setAPIKeys] = useState<MerchantAPIKey[]>([]);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingKey, setEditingKey] = useState<MerchantAPIKey | null>(null);
  const [verifying, setVerifying] = useState<number | null>(null);
  const [form] = Form.useForm();

  const handleVerify = async (id: number) => {
    setVerifying(id);
    try {
      await api.post(`/merchant/api-keys/${id}/verify`);
      message.success('验证请求已提交，请稍后刷新查看结果');
    } catch (error) {
      message.error('验证失败');
    } finally {
      setVerifying(null);
    }
  };

  const handleSubmit = async (values: any) => {
    const data = {
      ...values,
      input_price_per_1k: parseFloat(values.input_price_per_1k) || 0,
      output_price_per_1k: parseFloat(values.output_price_per_1k) || 0,
    };
    
    if (editingKey) {
      await api.put(`/merchant/api-keys/${editingKey.id}`, data);
    } else {
      await api.post('/merchant/api-keys', data);
    }
    
    setModalVisible(false);
    form.resetFields();
    loadAPIKeys();
  };

  const columns = [
    { title: '名称', dataIndex: 'name', key: 'name' },
    { title: '提供商', dataIndex: 'provider', key: 'provider' },
    { 
      title: '端点URL', 
      dataIndex: 'api_base_url', 
      key: 'api_base_url',
      render: (url: string) => url || '默认',
    },
    { 
      title: '健康状态', 
      dataIndex: 'health_status', 
      key: 'health_status',
      render: (status: string) => {
        const colors = {
          healthy: 'green',
          degraded: 'orange',
          unhealthy: 'red',
          unknown: 'default',
        };
        return <Tag color={colors[status]}>{status}</Tag>;
      },
    },
    { 
      title: '验证状态', 
      key: 'verified',
      render: (_: any, record: MerchantAPIKey) => (
        record.verified_at ? (
          <Tag color="green">已验证</Tag>
        ) : (
          <Tag color="default">未验证</Tag>
        )
      ),
    },
    { 
      title: '成本定价', 
      key: 'pricing',
      render: (_: any, record: MerchantAPIKey) => (
        <span>
          输入: ¥{record.input_price_per_1k}/1K | 
          输出: ¥{record.output_price_per_1k}/1K
        </span>
      ),
    },
    { 
      title: '配额', 
      key: 'quota',
      render: (_: any, record: MerchantAPIKey) => (
        <Progress 
          percent={(record.quota_used / record.quota_limit) * 100} 
          size="small"
        />
      ),
    },
    {
      title: '操作',
      key: 'actions',
      render: (_: any, record: MerchantAPIKey) => (
        <Space>
          <Button 
            type="link" 
            onClick={() => handleVerify(record.id)}
            loading={verifying === record.id}
          >
            验证
          </Button>
          <Button type="link" onClick={() => handleEdit(record)}>编辑</Button>
          <Button type="link" danger onClick={() => handleDelete(record.id)}>删除</Button>
        </Space>
      ),
    },
  ];

  return (
    <div className={styles.apiKeys}>
      <Card 
        title="API Key管理" 
        extra={
          <Button type="primary" onClick={() => setModalVisible(true)}>
            添加API Key
          </Button>
        }
      >
        <Table columns={columns} dataSource={apiKeys} rowKey="id" />
      </Card>
      
      <Modal
        title={editingKey ? '编辑API Key' : '添加API Key'}
        open={modalVisible}
        onCancel={() => { setModalVisible(false); form.resetFields(); }}
        onOk={() => form.submit()}
        width={600}
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Form.Item name="name" label="名称" rules={[{ required: true }]}>
            <Input placeholder="如：OpenAI主账号" />
          </Form.Item>
          
          <Form.Item name="provider" label="提供商" rules={[{ required: true }]}>
            <Select>
              <Select.Option value="openai">OpenAI</Select.Option>
              <Select.Option value="anthropic">Anthropic</Select.Option>
              <Select.Option value="google">Google</Select.Option>
              <Select.Option value="azure">Azure OpenAI</Select.Option>
            </Select>
          </Form.Item>
          
          <Form.Item name="api_key" label="API Key" rules={[{ required: !editingKey }]}>
            <Input.Password placeholder="sk-..." />
          </Form.Item>
          
          <Form.Item name="api_base_url" label="API端点URL（可选）">
            <Input placeholder="https://api.openai.com/v1" />
          </Form.Item>
          
          <Form.Item name="quota_limit" label="配额限制（元）">
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
          
          <Divider>成本定价配置</Divider>
          
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="input_price_per_1k" label="输入Token成本（元/1K）">
                <InputNumber min={0} step={0.0001} precision={6} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="output_price_per_1k" label="输出Token成本（元/1K）">
                <InputNumber min={0} step={0.0001} precision={6} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>
          
          <Form.Item name="health_check_level" label="健康检查频率" initialValue="medium">
            <Select>
              <Select.Option value="high">高频（每1分钟）</Select.Option>
              <Select.Option value="medium">中频（每5分钟）</Select.Option>
              <Select.Option value="low">低频（每30分钟）</Select.Option>
              <Select.Option value="daily">日级（每24小时）</Select.Option>
            </Select>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};
```

---

#### 3.3.2 产品列表页面增强（pages/ProductListPage.tsx）【P1】

**目标**：支持场景筛选、性能排序

**新增功能**：
- 场景分类导航
- 性能指标展示
- 智能推荐

**核心组件**：
```typescript
const ProductListPage: React.FC = () => {
  const [scenarios, setScenarios] = useState<Scenario[]>([]);
  const [selectedScenario, setSelectedScenario] = useState<string | null>(null);
  const [products, setProducts] = useState<Product[]>([]);
  const [sortBy, setSortBy] = useState<'price' | 'performance' | 'popularity'>('popularity');

  useEffect(() => {
    loadScenarios();
    loadProducts();
  }, [selectedScenario, sortBy]);

  const loadScenarios = async () => {
    const response = await api.get('/catalog/scenarios');
    setScenarios(response.data);
  };

  const loadProducts = async () => {
    const params = new URLSearchParams();
    if (selectedScenario) params.append('scenario', selectedScenario);
    params.append('sort', sortBy);
    
    const response = await api.get(`/catalog/products?${params}`);
    setProducts(response.data);
  };

  return (
    <div className={styles.productList}>
      <div className={styles.scenarioNav}>
        <div 
          className={selectedScenario === null ? styles.active : ''}
          onClick={() => setSelectedScenario(null)}
        >
          全部
        </div>
        {scenarios.map(scenario => (
          <div
            key={scenario.code}
            className={selectedScenario === scenario.code ? styles.active : ''}
            onClick={() => setSelectedScenario(scenario.code)}
          >
            <img src={scenario.icon_url} alt={scenario.name} />
            <span>{scenario.name}</span>
          </div>
        ))}
      </div>
      
      <div className={styles.toolbar}>
        <Radio.Group value={sortBy} onChange={e => setSortBy(e.target.value)}>
          <Radio.Button value="popularity">热门优先</Radio.Button>
          <Radio.Button value="price">价格优先</Radio.Button>
          <Radio.Button value="performance">性能优先</Radio.Button>
        </Radio.Group>
      </div>
      
      <div className={styles.products}>
        {products.map(product => (
          <ProductCard 
            key={product.id} 
            product={product}
            showPerformance={sortBy === 'performance'}
          />
        ))}
      </div>
    </div>
  );
};
```

---

#### 3.3.3 商户结算页面（pages/merchant/MerchantSettlements.tsx）【P1】

**目标**：展示结算列表、详情、确认流程

**核心组件**：
```typescript
const MerchantSettlements: React.FC = () => {
  const [settlements, setSettlements] = useState<Settlement[]>([]);
  const [detailVisible, setDetailVisible] = useState(false);
  const [selectedSettlement, setSelectedSettlement] = useState<Settlement | null>(null);

  const handleConfirm = async (id: number) => {
    await api.post(`/merchant/settlements/${id}/confirm`);
    message.success('已确认结算单');
    loadSettlements();
  };

  const columns = [
    { title: '结算周期', dataIndex: 'period', key: 'period' },
    { 
      title: '商户收入', 
      dataIndex: 'total_merchant_cost', 
      key: 'total_merchant_cost',
      render: (v: number) => `¥${v.toFixed(2)}`,
    },
    { 
      title: '平台利润', 
      dataIndex: 'platform_profit', 
      key: 'platform_profit',
      render: (v: number) => `¥${v.toFixed(2)}`,
    },
    { 
      title: '状态', 
      dataIndex: 'status', 
      key: 'status',
      render: (status: string, record: Settlement) => {
        if (status === 'pending' && !record.merchant_confirmed) {
          return <Tag color="orange">待确认</Tag>;
        } else if (status === 'pending' && record.merchant_confirmed) {
          return <Tag color="blue">待审核</Tag>;
        } else if (status === 'approved') {
          return <Tag color="green">已审核</Tag>;
        } else if (status === 'paid') {
          return <Tag color="green">已打款</Tag>;
        }
        return <Tag>{status}</Tag>;
      },
    },
    {
      title: '操作',
      key: 'actions',
      render: (_: any, record: Settlement) => (
        <Space>
          <Button type="link" onClick={() => { setSelectedSettlement(record); setDetailVisible(true); }}>
            详情
          </Button>
          {record.status === 'pending' && !record.merchant_confirmed && (
            <Button type="primary" size="small" onClick={() => handleConfirm(record.id)}>
              确认
            </Button>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div>
      <Card title="结算记录">
        <Table columns={columns} dataSource={settlements} rowKey="id" />
      </Card>
      
      <Modal
        title="结算详情"
        open={detailVisible}
        onCancel={() => setDetailVisible(false)}
        footer={null}
        width={800}
      >
        {selectedSettlement && <SettlementDetail settlement={selectedSettlement} />}
      </Modal>
    </div>
  );
};
```

---

#### 3.3.4 管理员结算管理页面（pages/admin/AdminSettlements.tsx）【P1】

**目标**：财务审核、打款记录

**核心组件**：
```typescript
const AdminSettlements: React.FC = () => {
  const [settlements, setSettlements] = useState<Settlement[]>([]);

  const handleApprove = async (id: number) => {
    await api.post(`/admin/settlements/${id}/approve`);
    message.success('审核通过');
    loadSettlements();
  };

  const handleMarkPaid = async (id: number, transactionId: string) => {
    await api.post(`/admin/settlements/${id}/mark-paid`, { transaction_id: transactionId });
    message.success('已标记为打款');
    loadSettlements();
  };

  const columns = [
    { title: '商户', dataIndex: 'merchant_name', key: 'merchant_name' },
    { title: '结算周期', dataIndex: 'period', key: 'period' },
    { title: '商户收入', dataIndex: 'total_merchant_cost', key: 'total_merchant_cost' },
    { title: '平台利润', dataIndex: 'platform_profit', key: 'platform_profit' },
    { 
      title: '状态', 
      key: 'status',
      render: (_: any, record: Settlement) => {
        if (!record.merchant_confirmed) return <Tag color="default">待商户确认</Tag>;
        if (!record.finance_approved) return <Tag color="orange">待财务审核</Tag>;
        if (record.status === 'approved') return <Tag color="blue">待打款</Tag>;
        return <Tag color="green">已完成</Tag>;
      },
    },
    {
      title: '操作',
      key: 'actions',
      render: (_: any, record: Settlement) => (
        <Space>
          {record.merchant_confirmed && !record.finance_approved && (
            <Button type="primary" size="small" onClick={() => handleApprove(record.id)}>
              审核
            </Button>
          )}
          {record.status === 'approved' && (
            <Button type="primary" size="small" onClick={() => showPaidModal(record.id)}>
              标记打款
            </Button>
          )}
          <Button type="link" onClick={() => showDetail(record)}>详情</Button>
        </Space>
      ),
    },
  ];

  return (
    <Card title="结算管理">
      <Table columns={columns} dataSource={settlements} rowKey="id" />
    </Card>
  );
};
```

## 四、已确认决策

> 以下决策已与用户讨论确认，作为后续开发的指导原则

### 4.1 智能路由策略选择 ✅ 已确认

**决策结果**：**平台统一配置**

**详细说明**：
- 平台运营方设置全局路由策略
- 商户无法自定义（初期版本）
- 后续版本迭代支持商户自定义路由偏好

**实现要点**：
```sql
-- 路由策略配置表
CREATE TABLE routing_strategies (
  id SERIAL PRIMARY KEY,
  name VARCHAR(50) NOT NULL,
  strategy_type VARCHAR(20) NOT NULL,  -- price_first, latency_first, balanced
  weight_config JSONB,
  is_default BOOLEAN DEFAULT FALSE,
  status VARCHAR(20) DEFAULT 'active'
);
```

---

### 4.2 成本定价模式 ✅ 已确认

**决策结果**：**动态定价（从数据库获取）**

**详细说明**：
- 定价数据存储在数据库中，支持实时更新
- SPU层面设置基准定价
- SKU层面可设置差异化定价
- 商户层面可设置成本定价

**实现要点**：
```sql
-- SPU定价字段
ALTER TABLE spus ADD COLUMN input_price_per_1k DECIMAL(10,6);
ALTER TABLE spus ADD COLUMN output_price_per_1k DECIMAL(10,6);

-- SKU成本定价
ALTER TABLE skus ADD COLUMN cost_input_rate DECIMAL(10,6);
ALTER TABLE skus ADD COLUMN cost_output_rate DECIMAL(10,6);

-- 商户API Key成本定价
ALTER TABLE merchant_api_keys ADD COLUMN input_price_per_1k DECIMAL(10,6);
ALTER TABLE merchant_api_keys ADD COLUMN output_price_per_1k DECIMAL(10,6);
```

```go
// 动态定价服务
type PricingService struct {
    db         *sql.DB
    cache      map[int]*PricingData
    cacheMutex sync.RWMutex
    cacheTTL   time.Duration  // 缓存5分钟
}

func (s *PricingService) GetPricing(skuID int) (*PricingData, error) {
    // 先查缓存
    if cached, ok := s.cache[skuID]; ok {
        return cached, nil
    }
    // 从数据库读取
    return s.loadFromDB(skuID)
}
```

---

### 4.3 结算周期 ✅ 已确认

**决策结果**：**月结（大商户可申请周结）**

**详细说明**：
- 默认每月1日生成上月结算单
- 大商户（月结算额超过阈值）可申请周结
- 结算单生成后需商户确认、财务审核

**实现要点**：
```go
// 结算周期配置
type SettlementCycle string

const (
    SettlementMonthly SettlementCycle = "monthly"  // 月结
    SettlementWeekly  SettlementCycle = "weekly"   // 周结
)

// 商户结算配置
ALTER TABLE merchants ADD COLUMN settlement_cycle VARCHAR(20) DEFAULT 'monthly';
ALTER TABLE merchants ADD COLUMN monthly_threshold DECIMAL(15,2) DEFAULT 10000.00;
```

---

### 4.4 API Key验证时机 ✅ 已确认

**决策结果**：**保存后异步验证**

**详细说明**：
- 商户添加API Key时先保存到数据库
- 后台异步执行验证
- 验证失败标记状态，禁止路由使用
- 验证成功后才能参与路由

**实现要点**：
```sql
-- API Key验证状态
ALTER TABLE merchant_api_keys ADD COLUMN verified_at TIMESTAMP;
ALTER TABLE merchant_api_keys ADD COLUMN verification_result JSONB;

-- 验证记录表
CREATE TABLE api_key_verifications (
  id SERIAL PRIMARY KEY,
  api_key_id INT REFERENCES merchant_api_keys(id),
  verification_type VARCHAR(50),  -- connection, models, pricing
  status VARCHAR(20),             -- pending, success, failed
  details JSONB,
  error_message TEXT,
  verified_at TIMESTAMP
);
```

```go
// 异步验证流程
func (s *APIKeyValidator) ValidateAsync(apiKeyID int) {
    go func() {
        result := s.validateAPIKey(apiKeyID)
        s.updateVerificationStatus(apiKeyID, result)
    }()
}
```

---

### 4.5 健康检查频率 ✅ 已确认

**决策结果**：**多级可配置（支持分钟级/日级）**

**详细说明**：
- 支持商户级别配置健康检查频率
- 参考OpenAI最佳实践，健康检查会消耗API配额
- 默认采用中频检查（每5分钟）

**频率级别**：

| 级别 | 频率 | 适用场景 | API配额消耗 |
|------|------|---------|------------|
| `high` | 每1分钟 | 关键商户、高流量 | 高 |
| `medium` | 每5分钟 | 默认配置 | 中 |
| `low` | 每30分钟 | 低流量商户 | 低 |
| `daily` | 每24小时 | 仅检查可用性 | 极低 |

**实现要点**：
```sql
-- 商户API Key健康检查配置
ALTER TABLE merchant_api_keys ADD COLUMN health_check_interval INT DEFAULT 300;
ALTER TABLE merchant_api_keys ADD COLUMN health_check_level VARCHAR(20) DEFAULT 'medium';

-- 健康检查历史记录
CREATE TABLE api_key_health_history (
  id SERIAL PRIMARY KEY,
  api_key_id INT REFERENCES merchant_api_keys(id),
  status VARCHAR(20),
  latency_ms INT,
  error_message TEXT,
  checked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

```go
// 健康检查级别
type HealthCheckLevel string

const (
    HealthCheckHigh   HealthCheckLevel = "high"    // 1分钟
    HealthCheckMedium HealthCheckLevel = "medium"  // 5分钟
    HealthCheckLow    HealthCheckLevel = "low"     // 30分钟
    HealthCheckDaily  HealthCheckLevel = "daily"   // 24小时
)

func (l HealthCheckLevel) Interval() time.Duration {
    switch l {
    case HealthCheckHigh:
        return 1 * time.Minute
    case HealthCheckMedium:
        return 5 * time.Minute
    case HealthCheckLow:
        return 30 * time.Minute
    case HealthCheckDaily:
        return 24 * time.Hour
    default:
        return 5 * time.Minute
    }
}
```

---

### 4.6 决策汇总表

| # | 问题 | 决策结果 | 备注 |
|---|------|---------|------|
| 1 | 智能路由策略选择 | 平台统一配置 | 后续支持商户自定义 |
| 2 | 成本定价模式 | 动态定价（数据库获取） | 支持缓存加速 |
| 3 | 结算周期 | 月结 | 大商户可申请周结 |
| 4 | API Key验证时机 | 保存后异步验证 | 验证失败禁止路由 |
| 5 | 健康检查频率 | 多级可配置 | 支持分钟级/日级 |

---

## 五、实施优先级

### Phase 1：核心能力（P0，2-3周）

| 任务 | 后端 | 前端 | 数据库 |
|------|------|------|--------|
| 智能路由核心 | ✅ | N/A | ✅ |
| 健康检查服务 | ✅ | N/A | ✅ |
| API Key验证 | ✅ | ✅ | ✅ |
| 动态定价 | ✅ | ✅ | ✅ |
| 商户成本配置 | ✅ | ✅ | ✅ |

### Phase 2：结算系统（P1，1-2周）

| 任务 | 后端 | 前端 | 数据库 |
|------|------|------|--------|
| 结算单生成 | ✅ | N/A | ✅ |
| 商户确认流程 | ✅ | ✅ | ✅ |
| 财务审核流程 | ✅ | ✅ | ✅ |
| 打款记录 | ✅ | ✅ | ✅ |

### Phase 3：用户体验（P1，1周）

| 任务 | 后端 | 前端 | 数据库 |
|------|------|------|--------|
| 场景分类 | ✅ | ✅ | ✅ |
| 产品筛选增强 | ✅ | ✅ | N/A |
| 用量统计增强 | ✅ | ✅ | N/A |

### Phase 4：运营能力（P2，1周）

| 任务 | 后端 | 前端 | 数据库 |
|------|------|------|--------|
| 审核流程完善 | ✅ | ✅ | ✅ |
| 监控告警 | ✅ | ✅ | ✅ |
| 商户统计报表 | ✅ | ✅ | N/A |

---

## 六、验收标准

### 6.1 功能验收

- [ ] 智能路由支持价格/延迟/均衡三种策略
- [ ] API Key故障30秒内自动切换
- [ ] 商户可配置成本定价
- [ ] 结算单自动生成并支持商户确认
- [ ] 用户可按场景筛选产品

### 6.2 性能验收

- [ ] 路由决策延迟 < 10ms
- [ ] 健康检查不影响正常请求
- [ ] 结算批处理支持万级订单

---

**报告版本**: v4.0
**最后更新**: 2026-04-04
**核心价值**: 基于代码实际实现情况，给出详细解决方案，所有决策已确认
**决策状态**: ✅ 全部确认（5/5）
