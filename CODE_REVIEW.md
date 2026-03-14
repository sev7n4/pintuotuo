# 代码审查清单 - Pintuotuo 项目

## 🔍 代码审查标准

本文档定义了 Pintuotuo 项目的代码审查标准和最佳实践。所有 Pull Request 必须满足这些条件才能被合并。

---

## 📋 通用检查清单

### 代码质量
- [ ] 代码遵循项目编码标准（见 CLAUDE.md）
- [ ] 变量名称有意义且易读
- [ ] 函数长度合理（<50 行）
- [ ] 没有重复代码（DRY 原则）
- [ ] 错误处理完整
- [ ] 没有硬编码值（使用配置/常量）

### 文档
- [ ] 公共函数有 JSDoc/GoDoc 注释
- [ ] 复杂逻辑有内联注释
- [ ] README 已更新（如果添加新功能）
- [ ] API 文档已更新

### 测试
- [ ] 新功能有单元测试
- [ ] 测试通过且覆盖率 >70%
- [ ] Edge case 已测试
- [ ] 没有 TODO 注释的测试

### 安全性
- [ ] 没有硬编码的密钥或密码
- [ ] 输入验证完整
- [ ] SQL 查询使用参数化（防止注入）
- [ ] 敏感操作有日志记录
- [ ] 没有 console.log 在生产代码中

---

## 🐹 Go 代码审查

### 格式和样式
- [ ] 代码通过 `go fmt`
- [ ] 通过 `go vet` 检查
- [ ] 通过 `golangci-lint` 检查
- [ ] 行长度 ≤ 100 字符

### 性能
- [ ] 没有不必要的 goroutine 创建
- [ ] 合理使用指针
- [ ] 没有内存泄漏（连接、goroutine）
- [ ] 数据库查询优化

### 特定于处理程序
- [ ] 所有端点返回适当的 HTTP 状态码
- [ ] 错误响应一致
- [ ] 请求验证完整
- [ ] 日志记录有用且非冗余

**示例：不好 vs 好**

```go
// 不好：没有验证，没有错误处理
func CreateProduct(c *gin.Context) {
    var req struct {
        Name  string
        Price float64
    }
    c.BindJSON(&req)
    // ... 直接使用，可能会 panic
}

// 好：完整的验证和错误处理
func CreateProduct(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }

    var req struct {
        Name  string  `json:"name" binding:"required,min=3"`
        Price float64 `json:"price" binding:"required,gt=0"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // ... 处理逻辑
}
```

---

## 🎨 TypeScript/React 代码审查

### 格式和样式
- [ ] 代码通过 `npm run lint`
- [ ] 代码通过 `npm run format`
- [ ] 通过 TypeScript 严格模式
- [ ] 行长度 ≤ 100 字符

### React 最佳实践
- [ ] 使用函数组件而不是类组件
- [ ] 正确使用 hooks（依赖数组完整）
- [ ] 避免在 render 中创建新对象
- [ ] Key props 有意义（不是 index）
- [ ] 组件 prop 类型已定义

### 状态管理
- [ ] Zustand stores 清晰组织
- [ ] 避免不必要的全局状态
- [ ] State 更新是原子的（immutable）

### 性能
- [ ] 没有不必要的 re-render
- [ ] 使用 React.memo 处理重 UI
- [ ] 异步操作有 loading/error 状态

**示例：不好 vs 好**

```typescript
// 不好：没有类型，直接修改状态
const MyComponent = () => {
    const users = useAuthStore().users
    users.push(newUser)  // 不好！直接修改
    return <div>{users.length}</div>
}

// 好：类型完整，immutable 更新
const MyComponent = () => {
    const users = useAuthStore((state) => state.users)
    const addUser = useAuthStore((state) => state.addUser)

    return (
        <div>
            <button onClick={() => addUser(newUser)}>Add</button>
            <div>{users.length}</div>
        </div>
    )
}
```

---

## 🗄️ 数据库审查

### 迁移
- [ ] 迁移是可逆的（有 UP 和 DOWN）
- [ ] 没有删除列的迁移（应该是 deprecated）
- [ ] 大表的迁移有性能考量
- [ ] 索引针对常见查询优化

### 查询
- [ ] 使用参数化查询（防止注入）
- [ ] N+1 查询问题已解决
- [ ] 有适当的索引
- [ ] 没有在循环中的数据库查询

---

## 📊 API 设计审查

### REST 约定
- [ ] 使用正确的 HTTP 动词（GET, POST, PUT, DELETE）
- [ ] 适当的 HTTP 状态码
  - 200: 成功
  - 201: 创建
  - 400: 请求无效
  - 401: 未授权
  - 403: 禁止
  - 404: 未找到
  - 500: 服务器错误

### 响应格式
- [ ] 响应结构一致
- [ ] 错误响应有清晰的消息
- [ ] 分页一致

**示例响应格式：**
```json
// 成功
{
  "data": { ... },
  "status": "success"
}

// 错误
{
  "error": "Descriptive error message",
  "code": "ERROR_CODE"
}

// 列表
{
  "data": [...],
  "total": 100,
  "page": 1,
  "per_page": 20
}
```

---

## 🔐 安全审查

### 认证和授权
- [ ] 所有受保护的端点验证令牌
- [ ] 用户只能访问自己的数据
- [ ] 管理操作有权限检查
- [ ] 密码已加密（不是明文）

### 输入验证
- [ ] 所有用户输入都经过验证
- [ ] SQL 查询参数化
- [ ] 文件上传有大小和类型限制
- [ ] XSS 防护（在 React 中自动）

### 数据保护
- [ ] 敏感数据不在日志中
- [ ] 使用 HTTPS（生产）
- [ ] CORS 配置限制性
- [ ] 密钥轮换策略

---

## 📝 审查反馈模板

### 如果发现问题

```markdown
## 问题：[简短标题]

**位置：** `file.go:123`

**问题：** [详细说明问题]

**建议：** [提供解决方案]

**参考：** [如果适用，链接到标准或示例]
```

### 如果赞同更改

```markdown
✅ 看起来很好！

可选的细微评论：
- 这部分可以简化...
- 考虑添加...
```

---

## ✅ 审查检查清单（评审者）

- [ ] 代码审查工具通过
- [ ] 所有测试通过
- [ ] 覆盖率满足要求
- [ ] 没有安全问题
- [ ] 文档清晰完整
- [ ] API 更改向后兼容
- [ ] 性能可接受
- [ ] 不违反既定的设计决策

---

## 🚀 提交 PR 前的检查

1. **本地验证**
   ```bash
   make lint
   make test
   make format
   ```

2. **最后检查**
   - [ ] 分支从 `develop` 创建
   - [ ] 提交消息遵循约定
   - [ ] No WIP 或 TODO 注释
   - [ ] 没有调试代码或 console.log

3. **PR 描述**
   - [ ] 清晰的标题
   - [ ] 链接相关 issue
   - [ ] 描述更改的动机
   - [ ] 列出关键更改

---

## 📚 参考文档

- [CLAUDE.md](../CLAUDE.md) - 编码标准
- [DEVELOPMENT.md](../DEVELOPMENT.md) - 开发指南
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Airbnb JavaScript Style Guide](https://github.com/airbnb/javascript)
- [React Best Practices](https://react.dev/learn)

---

**最后更新**：2026-03-14
**维护者**：Engineering Team
