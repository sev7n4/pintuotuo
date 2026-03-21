# 错误参考文档

> 本文档提供常见错误类型及处理方式参考。

## 编译错误

### Go编译错误

| 错误信息 | 原因 | 解决方案 |
|----------|------|----------|
| `undefined: xxx` | 变量/函数未定义 | 检查拼写、导入包 |
| `cannot use xxx as type yyy` | 类型不匹配 | 检查类型转换 |
| `imported and not used` | 导入未使用 | 删除未使用的导入 |
| `declared but not used` | 变量声明未使用 | 使用或删除变量 |

### TypeScript编译错误

| 错误信息 | 原因 | 解决方案 |
|----------|------|----------|
| `Cannot find name 'xxx'` | 变量未定义 | 检查拼写、导入 |
| `Type 'xxx' is not assignable to type 'yyy'` | 类型不匹配 | 检查类型定义 |
| `Property 'xxx' does not exist on type 'yyy'` | 属性不存在 | 检查类型定义或扩展 |

## 测试错误

### Go测试失败

| 错误模式 | 原因 | 解决方案 |
|----------|------|----------|
| `panic: runtime error` | 空指针/数组越界 | 添加nil检查/边界检查 |
| `expected: x, got: y` | 断言失败 | 检查业务逻辑 |
| `timeout` | 测试超时 | 优化性能或增加超时时间 |
| `race detected` | 竞态条件 | 使用mutex或channel同步 |

### Jest测试失败

| 错误模式 | 原因 | 解决方案 |
|----------|------|----------|
| `Cannot read property 'xxx' of undefined` | 访问undefined属性 | 添加空值检查 |
| `expect(received).toBe(expected)` | 断言失败 | 检查预期值 |
| `Timeout - Async callback was not invoked` | 异步超时 | 检查async/await使用 |

### E2E测试失败

| 错误模式 | 原因 | 解决方案 |
|----------|------|----------|
| `Timeout waiting for selector` | 元素未找到 | 检查选择器、等待时间 |
| `Element is not attached` | 元素已从DOM移除 | 重新获取元素 |
| `net::ERR_CONNECTION_REFUSED` | 服务未启动 | 确保服务运行 |
| `Multiple elements found` | 选择器匹配多个 | 使用更精确的选择器 |

## Lint错误

### golangci-lint

| 错误代码 | 说明 | 解决方案 |
|----------|------|----------|
| `errcheck` | 未检查错误返回值 | 添加错误处理 |
| `govet` | 静态分析问题 | 按提示修复 |
| `ineffassign` | 无效赋值 | 使用或删除变量 |
| `staticcheck` | 静态检查问题 | 按提示修复 |

### ESLint

| 错误代码 | 说明 | 解决方案 |
|----------|------|----------|
| `no-unused-vars` | 变量未使用 | 使用或删除变量 |
| `no-explicit-any` | 使用了any类型 | 定义具体类型 |
| `react-hooks/exhaustive-deps` | Hook依赖缺失 | 添加依赖项 |

## CI/CD错误

### GitHub Actions失败

| 错误场景 | 原因 | 解决方案 |
|----------|------|----------|
| `Permission denied` | 权限不足 | 检查workflow权限配置 |
| `Out of memory` | 内存不足 | 优化内存使用 |
| `Timeout` | 步骤超时 | 增加超时时间或优化 |
| `Service unavailable` | 服务不可用 | 检查服务状态 |

### Docker构建失败

| 错误场景 | 原因 | 解决方案 |
|----------|------|----------|
| `COPY failed` | 文件不存在 | 检查文件路径 |
| `npm install failed` | 依赖安装失败 | 检查package.json |
| `Build timeout` | 构建超时 | 优化Dockerfile |

## 安全扫描错误

### Trivy扫描

| 错误类型 | 说明 | 解决方案 |
|----------|------|----------|
| CVE漏洞 | 依赖存在漏洞 | 更新依赖版本 |
| 敏感信息泄露 | 检测到密钥等 | 使用环境变量 |

### npm audit

| 错误类型 | 说明 | 解决方案 |
|----------|------|----------|
| `Critical` | 严重漏洞 | 立即更新 |
| `High` | 高危漏洞 | 尽快更新 |
| `Moderate` | 中危漏洞 | 计划更新 |

## 常见问题排查

### 问题：测试在本地通过但CI失败

**排查步骤**：
1. 检查环境变量差异
2. 检查依赖版本
3. 检查时区/时间相关测试
4. 检查并发/竞态条件

### 问题：E2E测试间歇性失败

**排查步骤**：
1. 增加等待时间
2. 使用更可靠的选择器
3. 检查网络请求
4. 检查动画是否完成

### 问题：集成测试数据库连接失败

**排查步骤**：
1. 检查数据库服务状态
2. 检查连接字符串
3. 检查网络连通性
4. 检查认证信息

## 错误日志分析

### 关键信息提取

从错误日志中提取：
1. 错误类型
2. 错误位置（文件、行号）
3. 错误堆栈
4. 相关变量值

### 日志搜索命令

```bash
# 查找错误
grep -i "error\|fail\|panic" logs.txt

# 查找特定错误
grep "undefined:" logs.txt

# 查找上下文
grep -A 5 -B 5 "error" logs.txt
```

## 错误修复优先级

| 优先级 | 错误类型 |
|--------|----------|
| P0 | 编译错误、安全漏洞 |
| P1 | 测试失败、CI阻塞 |
| P2 | Lint警告、代码质量 |
| P3 | 文档问题、优化建议 |
