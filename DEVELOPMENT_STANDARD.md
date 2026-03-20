# 开发规范文档 - Pintuotuo 项目

**文档版本**: 1.0
**最后更新**: 2026-03-20
**状态**: 正式执行

---

## 目录

1. [流程概览](#1-流程概览)
2. [开发阶段](#2-开发阶段)
3. [本地测试阶段](#3-本地测试阶段)
4. [代码提交阶段](#4-代码提交阶段)
5. [CI/CD 自动化阶段](#5-cicd-自动化阶段)
6. [合并与部署阶段](#6-合并与部署阶段)
7. [常用命令速查](#7-常用命令速查)
8. [分支管理规范](#8-分支管理规范)
9. [提交信息规范](#9-提交信息规范)
10. [代码审查规范](#10-代码审查规范)
11. [紧急修复流程](#11-紧急修复流程)
12. [常见问题处理](#12-常见问题处理)

---

## 1. 流程概览

### 完整开发流程图

```
┌──────────────────────────────────────────────────────────────────────┐
│                        开发到部署完整流程                              │
├──────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐           │
│  │ 创建分支 │───▶│ 本地开发 │───▶│ 本地测试 │───▶│ 提交代码 │           │
│  └─────────┘    └─────────┘    └─────────┘    └────┬────┘           │
│                                                     │                │
│                                                     ▼                │
│  ┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐           │
│  │ 自动部署 │◀───│ 合并main │◀───│ PR审查  │◀───│ CI测试  │           │
│  └─────────┘    └─────────┘    └─────────┘    └─────────┘           │
│       │                                                              │
│       ▼                                                              │
│  ┌─────────────────────┐                                            │
│  │ 腾讯云生产服务器      │                                            │
│  │ IP: 119.29.173.89   │                                            │
│  └─────────────────────┘                                            │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

### 阶段检查清单

| 阶段 | 必须完成 | 验证命令 |
|------|----------|----------|
| 开发 | 功能实现完成 | 手动验证 |
| 测试 | 所有测试通过 | `make test` |
| 提交 | 代码格式化、Lint通过 | `make format && make lint` |
| CI | GitHub Actions 通过 | 查看 Actions 页面 |
| 部署 | 服务正常运行 | 访问生产环境验证 |

---

## 2. 开发阶段

### 2.1 开发新功能

```bash
# 步骤 1: 确保在最新的 main 分支上
git checkout main
git pull origin main

# 步骤 2: 创建功能分支
# 命名规范: feature/功能描述
git checkout -b feature/referral-enhancement

# 步骤 3: 启动本地开发环境
make docker-up          # 启动 PostgreSQL、Redis 等服务

# 步骤 4: 启动应用服务（需要两个终端）
# 终端 1:
make dev-backend

# 终端 2:
make dev-frontend
```

### 2.2 修复 Bug

```bash
# 步骤 1: 创建修复分支
# 命名规范: fix/bug描述
git checkout -b fix/login-validation-error

# 步骤 2: 启动开发环境（同上）
make docker-up
make dev-backend    # 终端 1
make dev-frontend   # 终端 2
```

### 2.3 开发环境访问地址

| 服务 | 地址 | 说明 |
|------|------|------|
| 前端应用 | http://localhost:5173 | Vite 开发服务器 |
| 后端 API | http://localhost:8080/api/v1 | Gin 服务器 |
| 健康检查 | http://localhost:8080/health | 服务状态 |
| PostgreSQL | localhost:5432 | 数据库 |
| Redis | localhost:6379 | 缓存服务 |

---

## 3. 本地测试阶段

### 3.1 运行测试

```bash
# 运行所有测试
make test

# 仅运行后端测试（含覆盖率报告）
make test-backend

# 仅运行前端测试
make test-frontend
```

### 3.2 代码质量检查

```bash
# 代码格式化
make format

# 代码 Lint 检查
make lint
```

### 3.3 测试覆盖率要求

| 模块 | 最低覆盖率 | 目标覆盖率 |
|------|------------|------------|
| 后端核心业务 | 80% | 90% |
| 后端工具函数 | 70% | 85% |
| 前端组件 | 70% | 80% |
| 前端 Hooks | 80% | 90% |

### 3.4 测试文件位置

```
backend/
├── handlers/*_test.go      # 处理器测试
├── services/*_test.go      # 服务层测试
├── utils/*_test.go         # 工具函数测试

frontend/src/
├── __tests__/              # 集成测试
├── components/__tests__/   # 组件测试
├── hooks/__tests__/        # Hooks 测试
├── services/__tests__/     # API 服务测试
├── stores/__tests__/       # 状态管理测试
```

---

## 4. 代码提交阶段

### 4.1 提交前检查清单

- [ ] 所有测试通过 (`make test`)
- [ ] 代码已格式化 (`make format`)
- [ ] Lint 检查通过 (`make lint`)
- [ ] 无 console.log/debugger 等调试代码
- [ ] 无敏感信息（密码、密钥等）

### 4.2 提交流程

```bash
# 步骤 1: 查看修改
git status
git diff

# 步骤 2: 添加修改
git add .                    # 添加所有修改
# 或
git add <specific-files>     # 添加指定文件

# 步骤 3: 提交（遵循提交信息规范）
git commit -m "feat: 添加推荐奖励功能增强"

# 步骤 4: 推送到远程
git push origin feature/referral-enhancement
```

### 4.3 提交信息规范

**格式**: `<type>(<scope>): <subject>`

**类型说明**:

| 类型 | 说明 | 示例 |
|------|------|------|
| `feat` | 新功能 | `feat: 添加用户头像上传功能` |
| `fix` | Bug 修复 | `fix: 修复订单金额计算错误` |
| `docs` | 文档更新 | `docs: 更新 API 文档` |
| `refactor` | 代码重构 | `refactor: 优化支付流程` |
| `test` | 测试相关 | `test: 添加订单服务单元测试` |
| `chore` | 构建/工具 | `chore: 更新依赖版本` |
| `style` | 代码风格 | `style: 格式化代码` |
| `perf` | 性能优化 | `perf: 优化数据库查询` |

**示例**:

```bash
# 新功能
git commit -m "feat(auth): 添加 JWT token 自动刷新机制"

# Bug 修复
git commit -m "fix(payment): 处理支付失败状态码"

# 重构
git commit -m "refactor(database): 优化用户查询性能"
```

---

## 5. CI/CD 自动化阶段

### 5.1 CI 流水线说明

推送代码后，GitHub Actions 自动运行以下检查：

```
┌─────────────────────────────────────────────────────────┐
│                    CI/CD Pipeline                       │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  ┌─────────────────────┐                               │
│  │ Backend Unit Tests  │  go test + coverage           │
│  └─────────────────────┘                               │
│                                                         │
│  ┌─────────────────────┐                               │
│  │ Frontend Build      │  npm run build                │
│  └─────────────────────┘                               │
│                                                         │
│  ┌─────────────────────┐                               │
│  │ Frontend Unit Tests │  npm test                     │
│  └─────────────────────┘                               │
│                                                         │
│  ┌─────────────────────┐                               │
│  │ Security Scan       │  Trivy + npm audit + gosec    │
│  └─────────────────────┘                               │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

### 5.2 查看 CI 结果

1. 打开 GitHub 仓库: https://github.com/sev7n4/pintuotuo
2. 点击 **Actions** 标签
3. 查看最新的工作流运行状态
4. 确保所有检查通过（绿色 ✓）

### 5.3 CI 失败处理

```bash
# 如果 CI 失败，查看错误信息
# 在 GitHub Actions 页面查看详细日志

# 常见失败原因及解决:

# 1. 测试失败
make test              # 本地运行测试，修复失败的测试

# 2. Lint 失败
make lint              # 本地运行 lint，修复问题

# 3. 构建失败
make build             # 本地构建，修复编译错误
```

---

## 6. 合并与部署阶段

### 6.1 创建 Pull Request

```bash
# 方式一: 通过 GitHub Web 界面（推荐）
# 1. 访问 https://github.com/sev7n4/pintuotuo
# 2. 点击 "Compare & pull request"
# 3. 填写 PR 描述
# 4. 等待 CI 通过和代码审查
```

**PR 描述模板**:

```markdown
## 变更描述
简要描述本次变更内容

## 变更类型
- [ ] 新功能
- [ ] Bug 修复
- [ ] 重构
- [ ] 文档更新

## 相关 Issue
Fixes #(issue number)

## 测试说明
- [ ] 已添加/更新单元测试
- [ ] 本地测试通过
- [ ] 手动测试完成

## 检查清单
- [ ] 代码符合规范
- [ ] 已添加必要注释
- [ ] 已更新相关文档
```

### 6.2 代码审查要求

| 要求 | 说明 |
|------|------|
| CI 状态 | 所有检查必须通过 |
| 审查人数 | 至少 1 人批准 |
| 冲突状态 | 无合并冲突 |

### 6.3 合并到主分支

```bash
# 通过 GitHub PR 界面合并（推荐）
# 或命令行合并:

git checkout main
git pull origin main
git merge feature/referral-enhancement
git push origin main
```

### 6.4 自动部署流程

合并到 `main` 分支后，自动触发部署：

```
┌─────────────────────────────────────────────────────────┐
│              Deploy to Tencent Cloud                    │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  1. SSH 连接到腾讯云服务器                               │
│     ↓                                                   │
│  2. git fetch origin                                    │
│     ↓                                                   │
│  3. git reset --hard origin/main                        │
│     ↓                                                   │
│  4. docker-compose -f docker-compose.prod.yml up -d     │
│     ↓                                                   │
│  5. 验证部署状态 (HTTP 200/302)                          │
│     ↓                                                   │
│  6. 发送邮件通知                                         │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

### 6.5 部署后验证

```bash
# 访问生产环境
# 前端: http://119.29.173.89

# 检查服务状态
curl http://119.29.173.89/health

# 查看部署日志
# GitHub Actions → 对应的 workflow run → 查看详细日志
```

---

## 7. 常用命令速查

### 7.1 开发命令

| 命令 | 说明 |
|------|------|
| `make dev` | 启动前后端开发服务 |
| `make dev-backend` | 启动后端服务 |
| `make dev-frontend` | 启动前端服务 |
| `make docker-up` | 启动 Docker 容器 |
| `make docker-down` | 停止 Docker 容器 |
| `make docker-logs` | 查看 Docker 日志 |

### 7.2 测试命令

| 命令 | 说明 |
|------|------|
| `make test` | 运行所有测试 |
| `make test-backend` | 运行后端测试 |
| `make test-frontend` | 运行前端测试 |
| `make lint` | 代码检查 |
| `make format` | 代码格式化 |

### 7.3 数据库命令

| 命令 | 说明 |
|------|------|
| `make migrate` | 运行数据库迁移 |
| `make db-shell` | 进入数据库 Shell |
| `make db-reset` | 重置数据库（危险操作） |

### 7.4 构建命令

| 命令 | 说明 |
|------|------|
| `make build` | 构建前后端 |
| `make build-backend` | 构建后端 |
| `make build-frontend` | 构建前端 |

### 7.5 Git 命令

| 命令 | 说明 |
|------|------|
| `git checkout -b feature/xxx` | 创建功能分支 |
| `git checkout -b fix/xxx` | 创建修复分支 |
| `git status` | 查看状态 |
| `git add .` | 添加所有修改 |
| `git commit -m "message"` | 提交 |
| `git push origin <branch>` | 推送分支 |
| `git pull origin main` | 拉取最新代码 |

---

## 8. 分支管理规范

### 8.1 分支类型

| 分支类型 | 命名格式 | 说明 | 示例 |
|----------|----------|------|------|
| 主分支 | `main` | 生产环境代码 | - |
| 功能分支 | `feature/*` | 新功能开发 | `feature/user-profile` |
| 修复分支 | `fix/*` | Bug 修复 | `fix/login-error` |
| 热修复分支 | `hotfix/*` | 紧急生产修复 | `hotfix/payment-critical` |
| 重构分支 | `refactor/*` | 代码重构 | `refactor/auth-module` |

### 8.2 分支生命周期

```
创建分支 → 开发 → 测试 → PR → 审查 → 合并 → 删除分支
```

### 8.3 分支命名规范

```bash
# ✅ 正确示例
feature/user-authentication
feature/payment-integration
fix/api-token-validation
hotfix/payment-critical-bug
refactor/database-queries

# ❌ 错误示例
my-feature
test
fix-bug
20240320-update
```

---

## 9. 提交信息规范

### 9.1 格式要求

```
<type>(<scope>): <subject>

<body>

<footer>
```

### 9.2 各部分说明

| 部分 | 必填 | 说明 |
|------|------|------|
| type | 是 | 提交类型 |
| scope | 否 | 影响范围 |
| subject | 是 | 简短描述（50字符内） |
| body | 否 | 详细描述 |
| footer | 否 | 关联 Issue |

### 9.3 完整示例

```bash
git commit -m "feat(auth): 添加 JWT token 自动刷新机制

- 添加 token 过期检测
- 实现自动刷新逻辑
- 更新相关测试用例

Fixes #42"
```

---

## 10. 代码审查规范

### 10.1 审查者检查清单

- [ ] 代码逻辑正确
- [ ] 测试覆盖充分
- [ ] 无性能问题
- [ ] 无安全风险
- [ ] 代码风格一致
- [ ] 注释清晰
- [ ] 文档已更新

### 10.2 审查反馈规范

```markdown
# 建议格式
**问题**: 描述问题
**位置**: 文件名:行号
**建议**: 改进建议

# 示例
**问题**: 缺少错误处理
**位置**: handlers/auth.go:45
**建议**: 添加 err != nil 的判断和处理
```

---

## 11. 紧急修复流程

### 11.1 热修复流程

```bash
# 步骤 1: 从 main 创建热修复分支
git checkout main
git pull origin main
git checkout -b hotfix/critical-payment-error

# 步骤 2: 快速修复并测试
# ... 修改代码 ...
make test

# 步骤 3: 提交并推送
git add .
git commit -m "hotfix: 修复支付关键错误"
git push origin hotfix/critical-payment-error

# 步骤 4: 创建 PR 并快速审查合并
# 步骤 5: 验证生产环境
```

### 11.2 紧急回滚

```bash
# 如果部署出现问题，需要回滚

# 方式一: 通过 GitHub Actions 重新部署之前的版本
# 在 Actions 页面选择之前成功的 workflow，点击 "Re-run"

# 方式二: 服务器手动回滚
ssh root@119.29.173.89
cd /opt/pintuotuo
git reset --hard HEAD~1
docker-compose -f docker-compose.prod.yml up -d --build
```

---

## 12. 常见问题处理

### 12.1 Git 相关问题

**问题: 分支落后于远程**

```bash
git fetch origin
git rebase origin/main
```

**问题: 合并冲突**

```bash
# 查看冲突文件
git status

# 手动解决冲突后
git add <resolved-files>
git rebase --continue
```

**问题: 提交到错误分支**

```bash
# 撤销最后一次提交（保留修改）
git reset --soft HEAD~1

# 切换到正确分支
git checkout -b feature/correct-branch

# 重新提交
git commit -m "feat: your message"
```

### 12.2 测试相关问题

**问题: 测试失败**

```bash
# 查看详细错误
make test-backend
# 或
cd backend && go test -v ./...

# 更新快照（前端）
cd frontend && npm test -- -u
```

### 12.3 部署相关问题

**问题: 部署失败**

```bash
# SSH 到服务器查看日志
ssh root@119.29.173.89
cd /opt/pintuotuo
docker-compose -f docker-compose.prod.yml logs -f
```

**问题: 服务无法访问**

```bash
# 检查容器状态
docker-compose -f docker-compose.prod.yml ps

# 重启服务
docker-compose -f docker-compose.prod.yml restart
```

---

## 附录

### A. 项目目录结构

```
pintuotuo/
├── .github/workflows/     # GitHub Actions 配置
│   ├── ci-cd.yml         # CI 流水线
│   └── deploy-tencent.yml # 部署配置
├── backend/              # Go 后端
├── frontend/             # React 前端
├── deploy/               # 部署配置
├── scripts/              # 工具脚本
├── Makefile              # 开发命令
├── docker-compose.yml    # 开发环境
└── docker-compose.prod.yml # 生产环境
```

### B. 相关文档

- [DEVELOPMENT.md](./DEVELOPMENT.md) - 开发指南
- [13_Dev_Git_Workflow_Code_Standards.md](./13_Dev_Git_Workflow_Code_Standards.md) - Git 工作流详细规范
- [DEPLOYMENT.md](./DEPLOYMENT.md) - 部署指南

### C. 联系方式

- GitHub: https://github.com/sev7n4/pintuotuo
- 生产环境: http://119.29.173.89

---

**文档维护**: 本文档应随项目发展持续更新

**执行要求**: 所有开发人员必须严格遵循本规范执行
