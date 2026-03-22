# 测试用例生命周期管理

> 本文档定义测试用例从设计到归档的生命周期流程和状态跟踪。

## 生命周期阶段

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   设计阶段   │ → │   测试阶段   │ → │   归档阶段   │
│  (Design)   │    │   (Test)    │    │  (Archive)  │
└─────────────┘    └─────────────┘    └─────────────┘
       ↓                  ↓                  ↓
   draft/ready      running/passed      archived
                    /failed
```

---

## Step 5: 测试用例设计 (TDD Red)

### 5.1 分析测试场景

**输入来源**：
- 从 `workflow_state.json` 获取 `current_fix_cases`
- 分析代码变更影响范围

**场景分析清单**：
- [ ] 正常路径 (Happy Path)
- [ ] 边界条件 (Boundary)
- [ ] 异常处理 (Error Handling)
- [ ] 并发场景 (Concurrency) - 如适用

### 5.2 设计测试用例

**测试策略参考**: `references/test_guide.md`

### 5.3 写入状态文件

**写入 `test_cases_state.json`**：

```json
{
  "currentIssue": {
    "id": "ISSUE-001",
    "type": "bug",
    "module": "product",
    "branch": "bugfix/issue-001-product-price"
  },
  "testCases": {
    "unit": [
      {
        "id": "UT-PROD-001",
        "name": "TestValidatePrice_PositiveValue_NoError",
        "feature": "Product Price Validation",
        "status": "draft",
        "designedAt": "2026-03-22T10:00:00Z",
        "testedAt": null,
        "file": "backend/internal/service/product_test.go"
      }
    ]
  }
}
```

**注意**：Issue ID 从 `workflow_state.json` 获取，不单独生成。

### 5.4 编写失败测试

1. 按 `test_case_templates.md` 编写测试代码
2. 运行测试确认失败
3. 更新状态为 `ready`

---

## Step 6: 测试用例实现 (TDD Green)

### 6.1 实现最小代码

- 只写让测试通过的最小代码
- 不做过度设计

### 6.2 运行测试验证

**测试命令参考**: `references/test_guide.md`

### 6.3 更新状态文件

**测试通过**：
```json
{
  "status": "passed",
  "testedAt": "2026-03-22T10:05:00Z"
}
```

**测试失败**：
```json
{
  "status": "failed",
  "testedAt": "2026-03-22T10:05:00Z",
  "errorMessage": "expected nil, got error: invalid price"
}
```

---

## Step 7: 测试用例归档 (TDD Refactor)

### 7.1 重构优化代码

- [ ] 代码可读性
- [ ] 重复代码消除
- [ ] 性能优化
- [ ] 测试仍然通过

### 7.2 生成归档文档

1. 读取 `test_cases_state.json`
2. 按 module 分组
3. 按 `test_case_templates.md` 归档格式生成文档

**文档路径**: `assets/test_cases/{module}/{feature}_cases.md`

### 7.3 更新状态文件

```json
{
  "status": "archived",
  "archivedTo": "assets/test_cases/product/price_cases.md"
}
```

---

## 状态跟踪

### 状态枚举

| 状态 | 说明 | 写入时机 |
|------|------|----------|
| `draft` | 设计中 | Step 5.3 |
| `ready` | 设计完成，待测试 | Step 5.4 |
| `running` | 测试执行中 | Step 6.2 |
| `passed` | 测试通过 | Step 6.3 |
| `failed` | 测试失败 | Step 6.3 |
| `archived` | 已归档 | Step 7.3 |

### 写入时机

| 步骤 | 操作 | 字段 |
|------|------|------|
| Step 5.3 | 写入 | `currentIssue`, `testCases.*.status=draft` |
| Step 5.4 | 更新 | `testCases.*.status=ready` |
| Step 6.2 | 更新 | `testCases.*.status=running` |
| Step 6.3 | 更新 | `testCases.*.status=passed/failed`, `testedAt` |
| Step 7.3 | 更新 | `testCases.*.status=archived`, `archivedTo` |

---

## 与 workflow_state.json 的关系

```
workflow_state.json          test_cases_state.json
├─ current_fix_cases    →    用于确定测试范围
├─ pr_number            ←    统计信息汇总
└─ ci_status.e2e        ←    E2E用例状态验证
```

**关键约束**：
- Issue ID 必须从 `workflow_state.json` 获取
- 测试用例状态独立管理
- E2E测试用例ID需与 `current_fix_cases` 对应

---

## 错误处理

| 场景 | 处理方式 |
|------|----------|
| 测试用例设计失败 | 返回 Step 4 重新分析需求 |
| 测试实现失败 | 返回 Step 6 修复代码 |
| 归档失败 | 重试归档，不阻塞工作流 |
