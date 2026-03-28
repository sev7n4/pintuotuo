# Pintuotuo Project Rules for Trae AI

> 本文件定义了 Trae AI 在本项目中必须遵循的规则。
> 工作流执行请使用 `mydev-github-workflow` skill

***

## 🚨 绝对禁止事项（NEVER）

1. **NEVER 直接推送到 main 分支** - 所有代码变更必须通过 PR 流程
2. **NEVER 跳过版本一致性检查** - 开始任务前必须验证三环境同步
3. **NEVER 在脏分支上开始新任务** - 必须确保工作区干净
4. **NEVER 提交敏感信息** - 密码、API Key、Token 等
5. **NEVER 跳过测试直接提交** - 必须运行 `make test` 通过

***

## 🔄 开发工作流

⚠️ 开发任务请调用 `mydev-github-workflow` skill 执行完整工作流

**工作流包含**：
- 版本一致性检查（三环境同步）
- 分支创建与管理
- TDD 开发流程
- CI/CD 监控
- 部署验证

***

## 🌿 分支命名规范

| 类型 | 格式 | 示例 |
|------|------|------|
| 功能 | `feature/issue-{id}` | `feature/issue-123` |
| 修复 | `fix/issue-{id}` | `fix/issue-456` |
| 热修复 | `hotfix/issue-{id}` | `hotfix/issue-789` |
| 重构 | `refactor/issue-{id}` | `refactor/issue-101` |

> **说明**: `{id}` 为 GitHub Issue 编号或任务标识

***

## 📝 提交信息规范

格式：`<type>(<scope>): <subject>`

| 类型 | 说明 |
|------|------|
| `feat` | 新功能 |
| `fix` | Bug 修复 |
| `docs` | 文档更新 |
| `refactor` | 代码重构 |
| `test` | 测试相关 |
| `chore` | 构建/工具 |
| `perf` | 性能优化 |
| `hotfix` | 紧急修复 |

**示例**: `feat(auth): add OAuth2 login support`

***

## 🔒 安全规则

1. **禁止提交敏感信息**
2. **使用环境变量管理配置**
3. **SSH 密钥不提交到仓库**

***

## 🚀 部署信息

### 生产环境

- 服务器 IP: `119.29.173.89`
- SSH 密钥: `~/.ssh/tencent_cloud_deploy`
- 部署路径: `/opt/pintuotuo`

### 自动部署触发

合并到 main 分支后自动触发部署

***

## 🛠️ 常用命令

### 开发

```bash
make docker-up      # 启动 PostgreSQL、Redis
make dev-backend    # 启动后端服务
make dev-frontend   # 启动前端服务
```

### 测试

```bash
make test           # 运行所有测试
make test-backend   # 后端测试
make test-frontend  # 前端测试
```

### 代码质量

```bash
make format         # 格式化代码
make lint           # Lint 检查
```

***

## 📁 项目结构

```text
pintuotuo/
├── backend/           # Go 后端 (Gin)
│   ├── handlers/      # HTTP 处理器
│   ├── models/        # 数据模型
│   ├── db/            # 数据库操作
│   └── migrations/    # 数据库迁移
├── frontend/          # React 前端 (Vite)
│   ├── src/pages/     # 页面组件
│   ├── src/services/  # API 服务
│   └── src/stores/    # 状态管理
├── .github/workflows/ # CI/CD 配置
└── Makefile           # 开发命令
```

***

## 📚 相关文档

- [DEVELOPMENT_STANDARD.md](./DEVELOPMENT_STANDARD.md) - 完整开发规范
- [DEVELOPMENT.md](./DEVELOPMENT.md) - 开发指南
- [DEPLOYMENT.md](./DEPLOYMENT.md) - 部署指南
- [CLAUDE.md](./CLAUDE.md) - Claude Code 规则

***

**最后更新**: 2026-03-29
