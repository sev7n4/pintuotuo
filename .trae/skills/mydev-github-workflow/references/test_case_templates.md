# 测试用例模版

> 本文档定义测试用例ID规范、三类测试模版和归档文档格式。

## 用例ID命名规范

| 类型 | 格式 | 示例 |
|------|------|------|
| 单元测试 | `UT-{MODULE}-{SEQ}` | UT-PROD-001 |
| 集成测试 | `IT-{MODULE}-{SEQ}` | IT-PROD-001 |
| E2E测试 | `E2E-{MODULE}-{SEQ}` | E2E-PROD-001 |

**MODULE缩写参考**：
- PROD: Product
- USER: User
- ORD: Order
- AUTH: Auth
- PAY: Payment

---

## 单元测试模版

### 测试函数命名

```
Test{Function}_{Scenario}_{ExpectedResult}
```

### 代码模版

```go
func Test{Function}_{Scenario}_{ExpectedResult}(t *testing.T) {
    // Arrange
    input := &Input{Field: "value"}
    
    // Act
    result, err := Function(input)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

---

## 集成测试模版

### 测试函数命名

```
Test{Feature}Integration_{Scenario}
```

### 代码模版

```go
func Test{Feature}Integration_{Scenario}(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    defer db.Close()
    
    server := setupTestServer(t, db)
    defer server.Close()
    
    // Arrange
    token := getTestToken(t, server)
    payload := `{"name": "test", "price": 100}`
    
    // Act
    req := httptest.NewRequest("POST", "/api/products", strings.NewReader(payload))
    req.Header.Set("Authorization", "Bearer "+token)
    rec := httptest.NewRecorder()
    server.ServeHTTP(rec, req)
    
    // Assert
    assert.Equal(t, http.StatusCreated, rec.Code)
}
```

---

## E2E测试模版

### 测试用例命名

```
{Feature} - {User Action} - {Expected Result}
```

### 代码模版 (Playwright)

```typescript
test.describe('{Feature}', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login')
    await page.fill('[name="email"]', 'test@example.com')
    await page.fill('[name="password"]', 'password')
    await page.click('button[type="submit"]')
    await page.waitForURL('**/dashboard')
  })

  test('{Feature} - {Action} - {Result}', async ({ page }) => {
    await page.goto('/target')
    await page.click('button.new-item')
    await page.fill('[name="name"]', 'Test Item')
    await page.click('button[type="submit"]')
    
    await expect(page.locator('.success-message')).toBeVisible()
  })
})
```

---

## 归档文档格式

**文件路径**: `assets/test_cases/{module}/{feature}_cases.md`

```markdown
# {Module} - {Feature} 测试用例文档

> 生成时间: {timestamp}
> 关联Issue: {issue_id}
> 分支: {branch_name}

## 测试用例统计

| 类型 | 总数 | 通过 | 失败 | 待定 |
|------|------|------|------|------|
| 单元测试 | {count} | {count} | {count} | {count} |
| 集成测试 | {count} | {count} | {count} | {count} |
| E2E测试 | {count} | {count} | {count} | {count} |

---

## 单元测试用例

### UT-{MODULE}-{SEQ}: {Test Name}

| 字段 | 值 |
|------|-----|
| 用例ID | UT-PROD-001 |
| 所属模块 | Product |
| 测试函数 | ValidatePrice |
| 状态 | ✅ passed |

**测试场景**：
- Given: {前置条件}
- When: {执行动作}
- Then: {预期结果}

**代码位置**：`{file_path}`

---

## 集成测试用例

### IT-{MODULE}-{SEQ}: {Test Name}
...

---

## E2E测试用例

### E2E-{MODULE}-{SEQ}: {Test Name}

| 字段 | 值 |
|------|-----|
| 用例ID | E2E-PROD-001 |
| 所属模块 | Product Management |
| 用户故事 | {user_story} |
| 状态 | ✅ passed |

**测试步骤**：
1. {step_1}
2. {step_2}

**验证点**：
- [x] {checkpoint_1}
- [x] {checkpoint_2}

**代码位置**：`{file_path}`
```
