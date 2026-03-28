# Pintuotuo Project Rules for Trae AI

> 本文件定义了 Trae AI 在本项目中必须遵循的规则和工作流。
> 参考：Anthropic Claude Code Best Practices & OpenAI Engineering Guidelines

---

## 🚨 CRITICAL: 强制工作流规则

### 绝对禁止事项（NEVER）

1. **NEVER 直接推送到 main 分支** - 所有代码变更必须通过 PR 流程
2. **NEVER 跳过版本一致性检查** - 开始任务前必须验证三环境同步
3. **NEVER 在脏分支上开始新任务** - 必须确保工作区干净
4. **NEVER 提交敏感信息** - 密码、API Key、Token 等
5. **NEVER 跳过测试直接提交** - 必须运行 `make test` 通过

### 强制执行顺序

```
1. 版本检查（三环境一致性）
   ↓
2. 分支状态检查（工作区干净）
   ↓
3. 创建功能分支（从 main）
   ↓
4. 开发/修复
   ↓
5. 本地测试（make test）
   ↓
6. 代码质量检查（make format && make lint）
   ↓
7. 提交（遵循提交规范）
   ↓
8. 推送分支
   ↓
9. 创建 PR
   ↓
10. 等待 CI 通过 + 审批
   ↓
11. Squash & Merge 到 main
```

---

## 📋 任务开始前强制检查

### Step 1: 版本一致性验证

**必须执行以下命令**：

```bash
echo "=== 本地环境 ===" && git log --oneline -3 && echo "" && echo "=== 远程仓库 (origin/main) ===" && git log origin/main --oneline -3 && echo "" && echo "=== 腾讯云服务器 ===" && ssh -i ~/.ssh/tencent_cloud_deploy root@119.29.173.89 "cd /opt/pintuotuo && git log --oneline -3"
```

**验证要求**：
- 三个环境的最新 commit hash 必须一致
- 如果不一致，必须先同步：
  ```bash
  git checkout main
  git pull origin main
  ```

### Step 2: 分支状态检查

```bash
git status
```

**必须显示**：`nothing to commit, working tree clean`

### Step 3: 创建功能分支

```bash
git checkout main
git pull origin main
git checkout -b feature/功能描述
```

---

## 🌿 分支命名规范

| 类型 | 格式 | 示例 |
|------|------|------|
| 功能 | `feature/描述` | `feature/user-profile` |
| 修复 | `fix/描述` | `fix/login-error` |
| 热修复 | `hotfix/描述` | `hotfix/payment-critical` |
| 重构 | `refactor/描述` | `refactor/auth-module` |

---

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

---

## ✅ 提交前检查清单

在执行 `git commit` 前，必须确保：

```bash
# 1. 运行测试
make test

# 2. 代码格式化
make format

# 3. Lint 检查
make lint

# 4. 确认无敏感信息
grep -r "password\|secret\|api_key" --include="*.go" --include="*.ts" --include="*.tsx"
```

---

## 🔄 PR 工作流

### PR 合并条件

- [ ] CI 通过（单元测试 + 静态分析）
- [ ] 至少 1 人批准
- [ ] 无合并冲突
- [ ] 代码审查完成

### 合并方式

使用 **Squash & Merge** 合并到 main

---

## 🚀 部署信息

### 生产环境

- 服务器 IP: `119.29.173.89`
- SSH 密钥: `~/.ssh/tencent_cloud_deploy`
- 部署路径: `/opt/pintuotuo`

### 自动部署触发

合并到 main 分支后自动触发部署

---

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

---

## 📁 项目结构

```
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

---

## 🔒 安全规则

1. **禁止提交敏感信息**
2. **使用环境变量管理配置**
3. **SSH 密钥不提交到仓库**

---

## ⚠️ 紧急修复流程

```bash
# 1. 从 main 创建热修复分支
git checkout main && git pull origin main
git checkout -b hotfix/xxx

# 2. 快速修复并测试
make test

# 3. 提交并推送
git commit -m "hotfix: 描述"
git push origin hotfix/xxx

# 4. 创建 PR 并快速审查合并
```

---

## 📚 相关文档

- [DEVELOPMENT_STANDARD.md](./DEVELOPMENT_STANDARD.md) - 完整开发规范
- [DEVELOPMENT.md](./DEVELOPMENT.md) - 开发指南
- [DEPLOYMENT.md](./DEPLOYMENT.md) - 部署指南
- [CLAUDE.md](./CLAUDE.md) - Claude Code 规则

---

**最后更新**: 2026-03-28
