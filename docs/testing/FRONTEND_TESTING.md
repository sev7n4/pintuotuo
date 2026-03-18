# 前端测试与构建规范

## 测试文件
- 位置：frontend/src 下，与组件或模块就近
- 命名：*.test.tsx / *.test.ts
- 配置与环境：
  - Jest 配置：[jest.config.cjs](file:///Users/4seven/pintuotuo/frontend/jest.config.cjs)
  - 初始化：[setup-tests.ts](file:///Users/4seven/pintuotuo/frontend/src/setup-tests.ts)

## 路径别名与类型
- tsconfig 路径别名参考：
  - "@/types" → src/types
- 导入示例：
```ts
import type { User, APIResponse, PaginatedResponse } from '@/types'
```
- 不要直接从声明文件路径导入类型（避免 TS6137）：
  - 错误：import type { User } from '@types/index'
  - 正确：import type { User } from '@/types'

## APIResponse 与分页结构
- APIResponse<T> 为响应体，实体在 response.data
- 分页：
```ts
const page: PaginatedResponse<Product> = response.data
const items = page.data
const total = page.total
```

## 常见问题与修复建议
- TS6137：从具体声明文件路径导入类型，改为从别名入口导入
- 将 APIResponse<T> 直接当作实体：改用 response.data
- AntD 组件 props 不匹配：以官方类型签名为准，移除不存在的属性
- 重复键：确保对象 key 唯一（例如 utils 映射）
- 构建失败不阻塞后端测试：在 CI 中可将前端测试/构建独立作业

## 运行命令
- 单元测试：
```
cd frontend
npm ci
CI=true npm test -- --watchAll=false
```
- 构建验证：
```
npm run build
```

