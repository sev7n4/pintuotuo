# MVP 测试用例差距分析报告

**生成时间**: 2026-03-23
**分析范围**: 后端、前端、E2E测试

---

## 一、执行摘要

### 总体测试覆盖率

| 层级 | 已覆盖 | 未覆盖 | 覆盖率 |
|------|--------|--------|--------|
| 后端 Handlers | 8/16 文件 | 8 文件 | 50% |
| 前端页面组件 | 32/43 | 11 | 74.4% |
| E2E 测试场景 | 56/105 | 49 | 53% |

### 关键发现

1. **后端核心模块完全缺失测试**: 订单、支付、拼团、管理员模块
2. **前端商家/管理员后台无测试覆盖**: 8个页面完全未测试
3. **E2E核心用户流程缺失**: 购物车、结算、支付、Token管理流程

---

## 二、后端测试差距详情

### 2.1 高优先级缺口

| 文件 | 函数数 | 测试状态 | 业务风险 |
|------|--------|----------|----------|
| `order_and_group.go` | 10 | ❌ 0% | 订单金额、库存管理、拼团逻辑错误 |
| `payment_and_token.go` | 9 | ⚠️ 11% | 支付金额错误、退款漏洞、代币丢失 |
| `payment_v2.go` | 7 | ❌ 0% | 支付回调处理错误、状态不一致 |
| `admin.go` | 3 | ❌ 0% | 权限绕过、数据泄露 |
| `merchant.go` | 7 | ❌ 0% | 商家数据错误、结算问题 |

### 2.2 中优先级缺口

| 文件 | 函数数 | 测试状态 | 业务风险 |
|------|--------|----------|----------|
| `merchant_apikey.go` | 8 | ❌ 0% | API密钥泄露、配额绕过 |
| `notification.go` | 7 | ❌ 0% | 通知丢失、设备令牌管理错误 |

### 2.3 低优先级缺口

| 文件 | 函数数 | 测试状态 | 业务风险 |
|------|--------|----------|----------|
| `health.go` | 4 | ❌ 0% | 监控数据不准确 |

### 2.4 需要补充的测试文件

```
backend/handlers/
├── order_and_group_test.go    (新建)
├── payment_and_token_test.go  (新建)
├── payment_v2_test.go         (新建)
├── admin_test.go              (新建)
├── merchant_test.go           (新建)
├── merchant_apikey_test.go    (新建)
├── notification_test.go       (新建)
└── health_test.go             (新建)
```

---

## 三、前端测试差距详情

### 3.1 主页面组件

| 组件 | 测试状态 | 需要测试的功能 |
|------|----------|----------------|
| `CheckoutPage.tsx` | ❌ 无 | 空购物车、商品选择、支付方式、订单提交 |
| `Consumption.tsx` | ❌ 无 | 消费记录表格、日期筛选、导出CSV |
| `MyToken.tsx` | ❌ 无 | 余额显示、交易记录、API密钥管理 |

### 3.2 商家页面 (全部缺失)

| 组件 | 测试状态 | 需要测试的功能 |
|------|----------|----------------|
| `MerchantDashboard.tsx` | ❌ 无 | 统计数据、最近订单、权限验证 |
| `MerchantOrders.tsx` | ❌ 无 | 订单列表、状态筛选、导出功能 |
| `MerchantProducts.tsx` | ❌ 无 | 商品CRUD、表单验证 |
| `MerchantSettings.tsx` | ❌ 无 | 店铺信息、Logo上传 |
| `MerchantSettlements.tsx` | ❌ 无 | 结算记录、申请结算 |
| `MerchantAPIKeys.tsx` | ❌ 无 | API密钥管理、配额显示 |

### 3.3 管理员页面 (全部缺失)

| 组件 | 测试状态 | 需要测试的功能 |
|------|----------|----------------|
| `AdminDashboard.tsx` | ❌ 无 | 统计数据、最近用户/订单 |
| `AdminUsers.tsx` | ❌ 无 | 用户列表、创建管理员 |

### 3.4 已覆盖良好的模块

| 类别 | 覆盖率 | 说明 |
|------|--------|------|
| Services | 100% | 9/9 文件有测试 |
| Stores | 100% | 10/10 文件有测试 |
| Hooks | 100% | 2/2 文件有测试 |

---

## 四、E2E测试差距详情

### 4.1 已覆盖的模块

| 模块 | 测试数 | 覆盖率 |
|------|--------|--------|
| 认证/授权 | 17 | 100% |
| 商品浏览 | 6 | 100% |
| 商家后台 | 28 | 100% |
| 订单管理 | 5 | 100% |

### 4.2 缺失的核心流程

#### 完整购物流程 (0% 覆盖)
- 添加商品到购物车
- 购物车修改数量/删除
- 结算页面选择支付方式
- 提交订单
- 支付确认

#### 拼团流程 (22% 覆盖)
- ❌ 创建拼团
- ❌ 加入拼团
- ❌ 拼团成功（达到目标）
- ❌ 拼团失败（超时）

#### Token管理 (0% 覆盖)
- ❌ 查看余额
- ❌ 交易记录
- ❌ API密钥管理
- ❌ 消费记录

#### 推荐系统 (0% 覆盖)
- ❌ 邀请码复制/分享
- ❌ 绑定邀请码
- ❌ 返利查看
- ❌ 提现申请

#### 管理后台 (20% 覆盖)
- ❌ 用户管理
- ❌ 创建管理员
- ❌ 数据统计

### 4.3 需要新增的E2E测试文件

```
frontend/e2e/
├── cart.spec.ts           (新建 - 购物车)
├── checkout.spec.ts       (新建 - 结算)
├── payment-flow.spec.ts   (新建 - 支付流程)
├── group-buy-flow.spec.ts (新建 - 拼团流程)
├── token.spec.ts          (新建 - Token管理)
├── referral.spec.ts       (新建 - 推荐系统)
└── admin.spec.ts          (新建 - 管理后台)
```

---

## 五、优先级排序

### P0 - 最高优先级 (阻塞核心业务)

1. **后端支付模块测试** - 资金安全
2. **后端订单模块测试** - 核心业务逻辑
3. **E2E完整购物流程** - 端到端验证

### P1 - 高优先级 (本周必须完成)

4. **后端拼团模块测试** - 核心功能
5. **前端商家后台测试** - B端核心
6. **E2E支付流程** - 支付验证

### P2 - 中优先级 (本迭代完成)

7. **后端管理员模块测试** - 权限控制
8. **前端管理员后台测试** - 管理功能
9. **E2E Token管理** - 用户功能

### P3 - 低优先级 (可延后)

10. **后端通知模块测试**
11. **后端健康检查测试**
12. **E2E推荐系统**

---

## 六、测试补充计划

### Week 1 Day 1-2: 后端核心模块

```bash
# 创建测试文件
touch backend/handlers/order_and_group_test.go
touch backend/handlers/payment_and_token_test.go
touch backend/handlers/payment_v2_test.go
```

### Week 1 Day 3: 前端商家后台

```bash
# 创建测试文件
touch frontend/src/pages/merchant/__tests__/MerchantDashboard.test.tsx
touch frontend/src/pages/merchant/__tests__/MerchantProducts.test.tsx
touch frontend/src/pages/merchant/__tests__/MerchantOrders.test.tsx
```

### Week 1 Day 4-5: E2E核心流程

```bash
# 创建E2E测试文件
touch frontend/e2e/shopping-flow.spec.ts
touch frontend/e2e/payment-flow.spec.ts
touch frontend/e2e/group-buy-flow.spec.ts
```

---

## 七、预期成果

| 指标 | 当前值 | 目标值 |
|------|--------|--------|
| 后端测试覆盖率 | ~40% | >= 70% |
| 前端测试覆盖率 | ~74% | >= 85% |
| E2E场景覆盖率 | ~53% | >= 80% |
| 高优先级Issue解决 | 0 | 100% |

---

**报告生成者**: AI Assistant
**下一步**: 创建 GitHub Issues 并启动工作流
