# SKU 管理问题修复计划

**日期**: 2026-03-30
**状态**: 待确认

---

## 1. 问题列表

| # | 问题描述 | 类型 |
|---|---------|------|
| 1 | 用户端主页没有SKU展示，仅显示了产品 | bug |
| 2 | 主页点击SKU套餐后跳转首页 | bug |
| 3 | 点击"超级拼团"返回500错误 | bug |
| 4 | 点击订单返回500错误 | bug |
| 5 | 商户端SKU管理重复按钮 | bug |
| 6 | 商户端获取可用SKU列表失败 | bug |
| 7 | 用户分析页面报错500 | bug |

---

## 2. 问题分析

### 问题1: 用户端主页没有SKU展示

**根因**: 主页组件未调用 SKU API 或未正确渲染 SKU 数据

**修复方案**: 
- 检查主页组件的 SKU 数据获取逻辑
- 确保 public SKU API 正常工作

### 问题2: 主页点击SKU套餐后跳转首页

**根因**: SKU 详情页路由或导航逻辑错误

**修复方案**:
- 检查路由配置
- 修复导航逻辑

### 问题3: 点击"超级拼团"返回500错误

**根因**: 拼团 API 或数据库查询错误

**修复方案**:
- 检查拼团相关 API
- 修复后端错误

### 问题4: 点击订单返回500错误

**根因**: 订单 API 错误

**修复方案**:
- 检查订单相关 API
- 修复后端错误

### 问题5: 商户端SKU管理重复按钮

**根因**: 前端组件渲染逻辑问题

**修复方案**:
- 检查 MerchantSKUs 组件
- 移除重复按钮

### 问题6: 商户端获取可用SKU列表失败

**根因**: API 调用失败或后端错误

**修复方案**:
- 检查 merchant SKU API
- 修复后端错误

### 问题7: 用户分析页面报错500

**根因**: 分析 API 错误

**修复方案**:
- 检查分析相关 API
- 修复后端错误

---

## 3. 修复范围

### 后端修复

| 文件 | 修改内容 |
|------|----------|
| `backend/handlers/*.go` | 修复各 API 500 错误 |
| `backend/routes/routes.go` | 检查路由配置 |

### 前端修复

| 文件 | 修改内容 |
|------|----------|
| `frontend/src/pages/Home.tsx` | 添加 SKU 展示 |
| `frontend/src/pages/merchant/MerchantSKUs.tsx` | 移除重复按钮 |
| `frontend/src/services/*.ts` | 修复 API 调用 |

---

## 4. 实施步骤

### Step 1: 创建修复分支

```bash
git checkout main
git pull origin main
git checkout -b fix/sku-management-issues
```

### Step 2: 代码分析

- 定位每个问题的具体代码位置
- 分析错误原因

### Step 3: 修复实现

- 修复后端 API 错误
- 修复前端组件问题

### Step 4: 本地验证

```bash
# 后端
cd backend && go build ./... && go test ./...

# 前端
cd frontend && npm run type-check
```

### Step 5: 创建 PR 并监控 CI

```bash
git add .
git commit -m "fix: resolve multiple SKU management issues"
git push -u origin fix/sku-management-issues
gh pr create --title "fix: resolve multiple SKU management issues" --body "..."
```

### Step 6: 合并部署

- CI 通过后合并 PR
- 监控部署状态

---

**请确认以上计划。**
