# Rules 与 Skill 精简方案计划

## 📊 背景

当前 `project_rules.md` 和 `mydev-github-workflow` skill 存在内容重叠和冲突，需要精简以实现职责分离。

## 🎯 目标

- **Rules** = 宪法（定义边界和禁止事项）
- **Skills** = 操作手册（定义如何执行）

## 📋 对比发现

### 重叠内容（需要精简）

| 内容 | Rules | Skill | 处理方式 |
|------|-------|-------|----------|
| 工作流步骤 | 11步概要 | 15步详细 | Rules 移除，保留 Skill |
| 分支状态检查 | Step 2 | Step 0 | Rules 移除，保留 Skill |
| 本地测试 | Step 5 | Step 8 | Rules 移除，保留 Skill |
| PR 流程 | Step 7-11 | Step 9-12 | Rules 移除，保留 Skill |

### 冲突内容（需要统一）

| 冲突项 | Rules | Skill | 统一方案 |
|--------|-------|-------|----------|
| 分支类型 | `fix` | `bugfix` | → `fix` |
| 分支类型 | `refactor` | `enhancement` | → `refactor` |
| 分支格式 | `feature/描述` | `{type}/issue-{id}` | → `{type}/issue-{id}` |
| 提交格式 | `<type>(<scope>): <subject>` | `{type}: {description}` | → `<type>(<scope>): <subject>` |

### Rules 独有内容（保留）

- ✅ 禁止事项（NEVER）
- ✅ 版本一致性检查（三环境同步）
- ✅ 安全规则
- ✅ 部署信息（服务器IP、SSH密钥）
- ✅ 常用命令
- ✅ 项目结构
- ✅ 紧急修复流程

### Skill 独有内容（保留）

- ✅ TDD 流程
- ✅ CI 监控逻辑
- ✅ 错误修复逻辑
- ✅ 状态跟踪机制
- ✅ 重试策略

## 🔧 执行步骤

### Step 1: 更新 Rules (`project_rules.md`)

**移除内容**：
- 强制执行顺序（11步流程图）
- 任务开始前强制检查（详细命令）
- 提交前检查清单（详细命令）
- PR 工作流（合并条件）

**保留内容**：
- 禁止事项（NEVER）- 5条
- 分支命名规范 - 统一格式
- 提交信息规范 - 统一格式
- 安全规则 - 3条
- 部署信息 - 服务器配置
- 常用命令 - make 命令
- 项目结构 - 目录说明
- 紧急修复流程 - 简化版
- 相关文档链接

**新增内容**：
- 工作流执行指引 → 指向 `mydev-github-workflow` skill

### Step 2: 更新 Skill (`SKILL.md`)

**修改内容**：
- Step 2-3 分支命名：`bugfix` → `fix`，`enhancement` → `refactor`
- Step 9 提交格式：`{type}: {description}` → `<type>(<scope>): <subject>`

**保留内容**：
- 其他所有内容保持不变

## 📄 精简后的 Rules 结构

```markdown
# Pintuotuo Project Rules for Trae AI

## 🚨 绝对禁止事项（NEVER）
1. NEVER 直接推送到 main 分支
2. NEVER 跳过版本一致性检查（三环境同步）
3. NEVER 在脏分支上开始新任务
4. NEVER 提交敏感信息
5. NEVER 跳过测试直接提交

## 🔄 开发工作流
⚠️ 开发任务请调用 `mydev-github-workflow` skill 执行完整 CI 链路

## 🌿 分支命名规范
| 类型 | 格式 | 示例 |
|------|------|------|
| 功能 | feature/issue-{id} | feature/issue-123 |
| 修复 | fix/issue-{id} | fix/issue-456 |
| 热修复 | hotfix/issue-{id} | hotfix/issue-789 |
| 重构 | refactor/issue-{id} | refactor/issue-101 |

## 📝 提交信息规范
格式: <type>(<scope>): <subject>

## 🔒 安全规则
1. 禁止提交敏感信息
2. 使用环境变量管理配置
3. SSH 密钥不提交到仓库

## 🚀 部署信息
- 服务器 IP: 119.29.173.89
- SSH 密钥: ~/.ssh/tencent_cloud_deploy
- 部署路径: /opt/pintuotuo

## 🛠️ 常用命令
(保留现有内容)

## 📁 项目结构
(保留现有内容)

## ⚠️ 紧急修复流程
(简化版，保留关键步骤)

## 📚 相关文档
(保留现有内容)
```

## ✅ 预期结果

| 指标 | 精简前 | 精简后 |
|------|--------|--------|
| Rules 行数 | ~240 行 | ~120 行 |
| 重叠内容 | 4 处 | 0 处 |
| 冲突内容 | 4 处 | 0 处 |
| 职责分离 | 模糊 | 清晰 |

## 📌 注意事项

1. 精简后需要验证 Rules 和 Skill 的引用关系
2. 确保紧急修复流程仍然可用
3. 保持文档链接的有效性
