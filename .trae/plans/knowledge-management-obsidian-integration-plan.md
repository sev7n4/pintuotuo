# 项目文档管理与知识沉淀详细方案

> 基于 PARA 方法、Obsidian CLI/MCP 集成、AI 编码工作流最佳实践的综合方案

---

## 一、现状分析

### 1.1 项目文档现状

| 目录 | 文件数 | 用途 |
|------|--------|------|
| `.trae/plans/` | ~20 | 开发计划、诊断报告 |
| `.trae/specs/` | ~40 | 规格文档、任务清单 |
| `.trae/documents/` | ~15 | 分析文档、设计文档 |
| `.trae/rules/` | 1 | AI 规则文件 |
| 根目录 | ~10 | README、开发指南等 |
| `backend/` | 3 | API 文档、内部文档 |
| `deploy/` | 2 | 部署相关文档 |

**问题识别**：
- 文档分散在多个目录，缺乏统一索引
- 计划文档与规格文档边界模糊
- 缺乏知识沉淀机制（对话内容、决策记录）
- 与 Obsidian 未建立有效连接

### 1.2 Obsidian Vault 现状

- **位置**: `/Users/4seven/Documents/Obsidian Vault`
- **笔记数量**: ~60 个
- **结构**: 扁平结构，无分类组织
- **内容类型**: 项目知识积累、技术调研、业务设计
- **已安装 CLI**: `/usr/local/bin/obsidian-cli`

**问题识别**：
- 未使用 PARA 方法组织
- 缺乏与项目代码仓库的联动
- 笔记命名不规范（含冒号等特殊字符）
- 未利用 CLI 能力进行自动化

### 1.3 已有能力盘点

| 能力 | 状态 | 位置 |
|------|------|------|
| Obsidian CLI | ✅ 已安装 | `/usr/local/bin/obsidian-cli` |
| Obsidian Vault | ✅ 已存在 | `/Users/4seven/Documents/Obsidian Vault` |
| 项目文档结构 | ✅ 已有 | `.trae/` 目录 |
| AI 规则文件 | ✅ 已有 | `.trae/rules/project_rules.md` |

---

## 二、目标与原则

### 2.1 核心目标

1. **项目文档规范化** - 建立清晰的文档分类和生命周期管理
2. **知识沉淀自动化** - AI 对话内容可一键沉淀到知识库
3. **Obsidian 深度集成** - 项目知识库与个人知识库双向联动
4. **AI 可访问** - AI 助手能读取和写入知识库

### 2.2 设计原则

- **最小侵入** - 不改变现有项目结构，增量优化
- **双向同步** - 项目文档 ↔ Obsidian 知识库
- **渐进式采用** - 可分阶段实施，每阶段独立可用
- **AI 友好** - 所有文档 AI 可读可写
- **CLI 优先** - 优先使用 CLI 实现自动化，降低插件依赖

---

## 三、技术方案详解

### 3.1 Obsidian CLI 能力详解

#### 3.1.1 已安装 CLI 命令清单

```bash
# 查看帮助
obsidian-cli --help

# 设置默认 vault
obsidian-cli set-default "Obsidian Vault"

# 列出 vault 内容
obsidian-cli list

# 创建笔记
obsidian-cli create --name "笔记名称" --content "内容"

# 搜索笔记
obsidian-cli search "关键词"
obsidian-cli search-content "内容关键词"

# 读取笔记
obsidian-cli print "笔记名称"

# 打开笔记
obsidian-cli open "笔记名称"

# 移动/重命名
obsidian-cli move "旧路径" "新路径"

# 删除笔记
obsidian-cli delete "笔记名称"

# 操作 frontmatter
obsidian-cli frontmatter get "笔记名称" "字段名"
obsidian-cli frontmatter set "笔记名称" "字段名" "值"

# 日记操作
obsidian-cli daily
```

#### 3.1.2 官方 CLI（可选升级）

Obsidian 官方 CLI 提供更强大的功能：

```bash
# 安装（如果需要）
npm install -g @obsidian/cli

# 官方 CLI 功能
obsidian daily                    # 打开今日日记
obsidian search query="关键词"     # 搜索
obsidian daily:append content="- [ ] 任务"  # 追加到日记
obsidian read                     # 读取当前文件
obsidian tasks daily              # 列出日记任务
obsidian template                 # 从模板创建
```

### 3.2 项目文档结构优化

```
pintuotuo/
├── .trae/
│   ├── rules/                    # AI 规则（保持）
│   │   └── project_rules.md
│   │
│   ├── specs/                    # 规格文档（保持）
│   │   └── [feature-name]/
│   │       ├── spec.md           # 功能规格
│   │       ├── tasks.md          # 任务清单
│   │       └── checklist.md      # 检查清单
│   │
│   ├── decisions/                # 【新增】架构决策记录 (ADR)
│   │   ├── INDEX.md              # 决策索引
│   │   └── YYYY-MM-DD-topic.md
│   │
│   ├── sessions/                 # 【新增】对话会话归档
│   │   ├── INDEX.md              # 会话索引
│   │   └── YYYY-MM-DD-topic.md
│   │
│   └── knowledge/                # 【新增】知识沉淀
│       ├── INDEX.md              # 知识索引
│       ├── patterns/             # 设计模式
│       │   └── pattern-name.md
│       ├── troubleshooting/      # 问题解决记录
│       │   └── issue-name.md
│       └── lessons-learned/      # 经验教训
│           └── lesson-name.md
│
├── docs/                         # 【新增】项目文档入口
│   ├── INDEX.md                  # 文档总索引
│   ├── architecture/             # 架构文档
│   │   └── overview.md
│   ├── api/                      # API 文档
│   │   └── endpoints.md
│   └── runbooks/                 # 运维手册
│       └── deployment.md
│
└── [现有结构保持不变]
```

### 3.3 Obsidian Vault 重构（PARA 方法）

```
Obsidian Vault/
├── 00_Inbox/                     # 快速捕获
│   └── README.md                 # 说明文件
│
├── 01_Projects/                  # 活跃项目
│   └── pintuotuo/
│       ├── README.md             # 项目概览
│       ├── active/               # 进行中的工作
│       │   └── current-sprint.md
│       ├── decisions/            # 决策记录（软链接）
│       ├── sessions/             # 会话归档（软链接）
│       └── specs/                # 规格文档（软链接）
│
├── 02_Areas/                     # 持续关注领域
│   ├── development/              # 开发实践
│   │   ├── coding-standards.md
│   │   └── testing-guide.md
│   ├── architecture/             # 架构设计
│   │   ├── system-design.md
│   │   └── api-design.md
│   ├── devops/                   # 运维部署
│   │   ├── ci-cd.md
│   │   └── monitoring.md
│   └── business/                 # 业务设计
│       ├── product-model.md
│       └── user-flow.md
│
├── 03_Resources/                 # 参考资料
│   ├── ai-tools/                 # AI 工具使用
│   │   ├── claude-code.md
│   │   ├── cursor-ai.md
│   │   └── trae-ai.md
│   ├── tech-stack/               # 技术栈文档
│   │   ├── go.md
│   │   ├── react.md
│   │   └── postgresql.md
│   ├── best-practices/           # 最佳实践
│   │   ├── git-workflow.md
│   │   └── code-review.md
│   └── research/                 # 调研报告
│       ├── competitors.md
│       └── market-analysis.md
│
├── 04_Archives/                  # 归档
│   └── 2025/                     # 按年份归档
│       └── [归档内容]
│
└── Templates/                    # 模板
    ├── session-template.md
    ├── decision-template.md
    ├── knowledge-template.md
    └── daily-note-template.md
```

### 3.4 双向同步机制设计

#### 方案选择对比

| 方案 | 优点 | 缺点 | 推荐场景 |
|------|------|------|----------|
| **软链接** | 零延迟同步、无冲突 | 仅限同机、路径敏感 | 单机开发 |
| **Git 同步** | 版本控制、跨设备 | 需手动提交 | 团队协作 |
| **脚本同步** | 灵活可控 | 需维护脚本 | 自定义需求 |
| **MCP 插件** | AI 直接操作 | 需安装插件 | AI 自动化 |

#### 推荐方案：软链接 + CLI 脚本

```bash
# 创建软链接，实现项目与 Obsidian 双向同步
ln -s /Users/4seven/workspace/pintuotuo/.trae/decisions \
      "/Users/4seven/Documents/Obsidian Vault/01_Projects/pintuotuo/decisions"

ln -s /Users/4seven/workspace/pintuotuo/.trae/sessions \
      "/Users/4seven/Documents/Obsidian Vault/01_Projects/pintuotuo/sessions"

ln -s /Users/4seven/workspace/pintuotuo/.trae/specs \
      "/Users/4seven/Documents/Obsidian Vault/01_Projects/pintuotuo/specs"
```

### 3.5 AI 对话沉淀自动化方案

#### 3.5.1 架构图

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        知识沉淀自动化架构                                  │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌──────────────┐     ┌──────────────┐     ┌──────────────┐            │
│  │  AI 对话窗口  │────►│  沉淀脚本    │────►│  Obsidian    │            │
│  │ (Trae/Cursor)│     │ (Shell/Python)│     │   Vault      │            │
│  └──────────────┘     └──────────────┘     └──────────────┘            │
│         │                    │                    │                     │
│         │                    │                    │                     │
│         ▼                    ▼                    ▼                     │
│  ┌──────────────┐     ┌──────────────┐     ┌──────────────┐            │
│  │  手动触发    │     │  模板渲染    │     │  自动索引    │            │
│  │  /save       │     │  Frontmatter │     │  双向链接    │            │
│  └──────────────┘     └──────────────┘     └──────────────┘            │
│                                                                         │
│  数据流向：                                                              │
│  对话内容 → 提取关键信息 → 填充模板 → CLI 写入 → 自动索引更新            │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

#### 3.5.2 沉淀脚本设计

**脚本 1：会话归档脚本** (`save-session.sh`)

```bash
#!/bin/bash
# save-session.sh - AI 会话归档到 Obsidian
# 用法: ./save-session.sh "会话标题" "项目名" "标签1,标签2"

VAULT_PATH="/Users/4seven/Documents/Obsidian Vault"
SESSIONS_PATH="$VAULT_PATH/01_Projects/pintuotuo/sessions"
DATE=$(date +%Y-%m-%d)
TIME=$(date +%H:%M:%S)
TITLE="$1"
PROJECT="${2:-pintuotuo}"
TAGS="${3}"

# 生成文件名（替换特殊字符）
FILENAME=$(echo "$TITLE" | sed 's/[^a-zA-Z0-9\u4e00-\u9fa5]/-/g')
FILEPATH="$SESSIONS_PATH/$DATE-$FILENAME.md"

# 从 stdin 读取内容
CONTENT=$(cat)

# 生成 frontmatter
FRONTMATTER="---
type: session
date: $DATE
time: $TIME
project: $PROJECT
tags: [$TAGS]
status: completed
---

# $TITLE

> 归档时间: $DATE $TIME

"

# 写入文件
echo "$FRONTMATTER$CONTENT" > "$FILEPATH"

echo "✅ 会话已归档: $FILEPATH"

# 更新索引
update_session_index "$SESSIONS_PATH"
```

**脚本 2：决策记录脚本** (`save-decision.sh`)

```bash
#!/bin/bash
# save-decision.sh - 架构决策记录
# 用法: ./save-decision.sh "决策标题" "状态(proposed|accepted|deprecated)"

VAULT_PATH="/Users/4seven/Documents/Obsidian Vault"
DECISIONS_PATH="$VAULT_PATH/01_Projects/pintuotuo/decisions"
DATE=$(date +%Y-%m-%d)
TITLE="$1"
STATUS="${2:-proposed}"

# 获取下一个 ADR 编号
NEXT_NUM=$(ls "$DECISIONS_PATH" 2>/dev/null | grep -E '^ADR-[0-9]+' | wc -l | xargs printf '%03d')
ADR_NUM="ADR-$((10#$NEXT_NUM + 1))"

FILENAME="$DATE-$TITLE.md"
FILEPATH="$DECISIONS_PATH/$FILENAME"

cat > "$FILEPATH" << EOF
---
type: adr
number: $ADR_NUM
status: $STATUS
date: $DATE
deciders: []
---

# $ADR_NUM: $TITLE

## 状态
$STATUS

## 背景
<!-- 决策背景 -->

## 决策
<!-- 具体决策内容 -->

## 理由
<!-- 为什么做这个决策 -->

## 后果
<!-- 决策带来的影响 -->

## 替代方案
<!-- 考虑过但未采用的方案 -->

## 相关
- [[]]
EOF

echo "✅ 决策记录已创建: $FILEPATH"
```

**脚本 3：知识沉淀脚本** (`save-knowledge.sh`)

```bash
#!/bin/bash
# save-knowledge.sh - 知识沉淀
# 用法: ./save-knowledge.sh "标题" "分类(pattern|troubleshooting|lesson)"

VAULT_PATH="/Users/4seven/Documents/Obsidian Vault"
KNOWLEDGE_PATH="$VAULT_PATH/03_Resources"
CATEGORY="$2"
TITLE="$1"
DATE=$(date +%Y-%m-%d)

case "$CATEGORY" in
  pattern)
    DIR="$KNOWLEDGE_PATH/patterns"
    ;;
  troubleshooting)
    DIR="$KNOWLEDGE_PATH/troubleshooting"
    ;;
  lesson)
    DIR="$KNOWLEDGE_PATH/lessons-learned"
    ;;
  *)
    DIR="$KNOWLEDGE_PATH"
    ;;
esac

FILENAME="$TITLE.md"
FILEPATH="$DIR/$FILENAME"

cat > "$FILEPATH" << EOF
---
type: knowledge
category: $CATEGORY
tags: []
date: $DATE
---

# $TITLE

## 问题描述
<!-- 遇到什么问题/场景 -->

## 解决方案
<!-- 如何解决 -->

## 代码示例
\`\`\`go
// 示例代码
\`\`\`

## 注意事项
<!-- 需要注意的点 -->

## 相关链接
- [[]]
EOF

echo "✅ 知识已沉淀: $FILEPATH"
```

#### 3.5.3 使用 Obsidian CLI 的自动化方案

```bash
# 设置默认 vault
obsidian-cli set-default "Obsidian Vault"

# 创建会话笔记
obsidian-cli create --name "01_Projects/pintuotuo/sessions/2026-04-27-routing-optimization" \
  --content "$(cat << 'EOF'
---
type: session
date: 2026-04-27
project: pintuotuo
topics: [routing, optimization]
---

# 智能路由系统优化

## 背景
...
EOF
)"

# 搜索相关笔记
obsidian-cli search-content "路由"

# 读取笔记内容
obsidian-cli print "APIkey 路由业务逻辑实现"
```

### 3.6 AI 集成方案

#### 3.6.1 方案 A：CLI 集成（推荐起步）

AI 助手可以通过 `RunCommand` 工具直接调用 Obsidian CLI：

```bash
# AI 创建笔记
obsidian-cli create --name "路径/笔记名" --content "内容"

# AI 搜索知识
obsidian-cli search-content "关键词"

# AI 读取笔记
obsidian-cli print "笔记名"
```

**优点**：
- 无需安装额外插件
- AI 可直接调用
- 命令简单可控

#### 3.6.2 方案 B：MCP 插件集成（进阶）

安装 [obsidian-mcp-plugin](https://github.com/aaronsb/obsidian-mcp-plugin) 后，AI 获得：

| 工具 | 功能 |
|------|------|
| `save_knowledge_note` | 创建知识笔记 |
| `save_code_snippet` | 保存代码片段 |
| `save_thread_summary` | 归档对话 |
| `search_notes` | 搜索笔记 |
| `list_notes` | 列出笔记 |
| `read_note` | 读取笔记 |
| `update_note` | 更新笔记 |

**配置步骤**：

1. 在 Obsidian 中安装社区插件 "Obsidian MCP"
2. 启用插件并配置 MCP Server
3. 在 Trae/Cursor 中配置 MCP 连接
4. 重启 AI 工具，验证连接

### 3.7 现有笔记迁移方案

#### 3.7.1 迁移映射表

| 现有笔记 | 目标位置 | 分类 |
|----------|----------|------|
| `APIkey 路由业务逻辑实现.md` | `02_Areas/architecture/` | 架构设计 |
| `SKILL设计哲学.md` | `03_Resources/ai-tools/` | AI 工具 |
| `产品形态调研.md` | `03_Resources/research/` | 调研报告 |
| `开发规范：.md` | `02_Areas/development/` | 开发实践 |
| `分支开发业界最佳实践：.md` | `03_Resources/best-practices/` | 最佳实践 |
| `关于商业模式设计.md` | `02_Areas/business/` | 业务设计 |
| `商户API秘钥管理的核心业务逻辑设计.md` | `02_Areas/business/` | 业务设计 |
| `Prometheus&grafana接入指标分析.md` | `02_Areas/devops/` | 运维部署 |
| `磁盘清理：.md` | `04_Archives/2025/` | 归档 |

#### 3.7.2 迁移脚本

```bash
#!/bin/bash
# migrate-notes.sh - 迁移现有笔记到 PARA 结构

VAULT="/Users/4seven/Documents/Obsidian Vault"

# 创建目录结构
mkdir -p "$VAULT"/{00_Inbox,01_Projects/pintuotuo,02_Areas/{development,architecture,devops,business},03_Resources/{ai-tools,tech-stack,best-practices,research},04_Archives/2025,Templates}

# 迁移函数
migrate() {
  local src="$1"
  local dst="$2"
  if [ -f "$VAULT/$src" ]; then
    mv "$VAULT/$src" "$VAULT/$dst/"
    echo "✅ 迁移: $src → $dst"
  fi
}

# 执行迁移
migrate "APIkey 路由业务逻辑实现.md" "02_Areas/architecture"
migrate "SKILL设计哲学.md" "03_Resources/ai-tools"
migrate "产品形态调研.md" "03_Resources/research"
migrate "开发规范：.md" "02_Areas/development"
migrate "分支开发业界最佳实践：.md" "03_Resources/best-practices"
migrate "关于商业模式设计.md" "02_Areas/business"
migrate "Prometheus&grafana接入指标分析.md" "02_Areas/devops"

echo "迁移完成！"
```

---

## 四、详细实施计划

### Phase 1：基础整理（预计 2-3 小时）

#### 任务清单

| 编号 | 任务 | 命令/操作 | 预计时间 |
|------|------|-----------|----------|
| 1.1 | 创建 Obsidian 目录结构 | `mkdir -p ...` | 10 分钟 |
| 1.2 | 创建项目文档目录 | `mkdir -p .trae/{decisions,sessions,knowledge}` | 5 分钟 |
| 1.3 | 迁移现有笔记 | 执行迁移脚本 | 30 分钟 |
| 1.4 | 创建软链接 | `ln -s ...` | 10 分钟 |
| 1.5 | 创建模板文件 | 手动创建 | 30 分钟 |
| 1.6 | 创建索引文件 | 手动创建 | 30 分钟 |
| 1.7 | 设置默认 vault | `obsidian-cli set-default` | 5 分钟 |

#### 详细操作步骤

**步骤 1.1：创建 Obsidian 目录结构**

```bash
VAULT="/Users/4seven/Documents/Obsidian Vault"

mkdir -p "$VAULT/00_Inbox"
mkdir -p "$VAULT/01_Projects/pintuotuo"/{active,decisions,sessions,specs}
mkdir -p "$VAULT/02_Areas"/{development,architecture,devops,business}
mkdir -p "$VAULT/03_Resources"/{ai-tools,tech-stack,best-practices,research}
mkdir -p "$VAULT/04_Archives/2025"
mkdir -p "$VAULT/Templates"
```

**步骤 1.2：创建项目文档目录**

```bash
cd /Users/4seven/workspace/pintuotuo
mkdir -p .trae/decisions
mkdir -p .trae/sessions
mkdir -p .trae/knowledge/{patterns,troubleshooting,lessons-learned}
mkdir -p docs/{architecture,api,runbooks}
```

**步骤 1.3：创建软链接**

```bash
VAULT="/Users/4seven/Documents/Obsidian Vault"
PROJECT="/Users/4seven/workspace/pintuotuo"

# 删除可能存在的空目录
rmdir "$VAULT/01_Projects/pintuotuo/decisions" 2>/dev/null
rmdir "$VAULT/01_Projects/pintuotuo/sessions" 2>/dev/null
rmdir "$VAULT/01_Projects/pintuotuo/specs" 2>/dev/null

# 创建软链接
ln -s "$PROJECT/.trae/decisions" "$VAULT/01_Projects/pintuotuo/decisions"
ln -s "$PROJECT/.trae/sessions" "$VAULT/01_Projects/pintuotuo/sessions"
ln -s "$PROJECT/.trae/specs" "$VAULT/01_Projects/pintuotuo/specs"
```

**步骤 1.4：设置默认 vault**

```bash
obsidian-cli set-default "Obsidian Vault"
```

### Phase 2：对话沉淀机制（预计 1-2 小时）

#### 任务清单

| 编号 | 任务 | 说明 | 预计时间 |
|------|------|------|----------|
| 2.1 | 创建沉淀脚本 | `save-session.sh` 等 | 30 分钟 |
| 2.2 | 定义归档规范 | 文档化流程 | 20 分钟 |
| 2.3 | 创建 ADR 模板 | 决策记录模板 | 15 分钟 |
| 2.4 | 测试沉淀流程 | 端到端测试 | 30 分钟 |

#### 详细操作步骤

**步骤 2.1：创建脚本目录和脚本**

```bash
mkdir -p /Users/4seven/workspace/pintuotuo/scripts/knowledge

# 创建三个脚本文件（内容见 3.5.2 节）
# - save-session.sh
# - save-decision.sh
# - save-knowledge.sh

chmod +x /Users/4seven/workspace/pintuotuo/scripts/knowledge/*.sh
```

### Phase 3：MCP 集成（可选，预计 1-2 小时）

#### 任务清单

| 编号 | 任务 | 说明 | 预计时间 |
|------|------|------|----------|
| 3.1 | 安装 Obsidian MCP 插件 | 在 Obsidian 社区插件中搜索安装 | 15 分钟 |
| 3.2 | 配置 MCP Server | 按插件文档配置 | 20 分钟 |
| 3.3 | 配置 AI 工具连接 | 在 Trae/Cursor 中添加 MCP | 15 分钟 |
| 3.4 | 测试 AI 操作 | 验证创建、搜索、读取 | 30 分钟 |

### Phase 4：持续优化（长期）

| 任务 | 频率 | 说明 |
|------|------|------|
| 清理 Inbox | 每周 | 整理临时笔记到正确位置 |
| 更新索引 | 每次新增 | 维护文档索引 |
| 归档旧内容 | 每月 | 移动不活跃内容到 Archives |
| 知识图谱维护 | 每月 | 检查和更新笔记链接 |

---

## 五、模板文件详解

### 5.1 会话归档模板

**文件位置**: `Templates/session-template.md`

```markdown
---
type: session
date: {{date}}
time: {{time}}
project: {{project}}
topics: [{{topics}}]
status: completed
related: []
---

# {{title}}

> 归档时间: {{date}} {{time}}

## 背景
<!-- 会话发起的背景和目标 -->


## 讨论过程

### 问题分析
<!-- 遇到的问题 -->


### 解决方案
<!-- 讨论出的方案 -->


## 关键决策
<!-- 做出的决策及理由 -->


## 产出物
<!-- 代码、文档等产出 -->


## 待办事项
- [ ] 

## 相关链接
- [[]]
```

### 5.2 架构决策记录模板 (ADR)

**文件位置**: `Templates/decision-template.md`

```markdown
---
type: adr
number: ADR-XXX
status: proposed | accepted | deprecated
date: {{date}}
deciders: []
---

# ADR-XXX: {{title}}

## 状态
{{status}}

## 背景
<!-- 决策背景 -->


## 决策
<!-- 具体决策内容 -->


## 理由
<!-- 为什么做这个决策 -->


## 后果
<!-- 决策带来的影响 -->


## 替代方案
<!-- 考虑过但未采用的方案 -->


## 相关
- [[]]
```

### 5.3 知识沉淀模板

**文件位置**: `Templates/knowledge-template.md`

```markdown
---
type: knowledge
category: pattern | troubleshooting | lesson
tags: []
date: {{date}}
---

# {{title}}

## 问题描述
<!-- 遇到什么问题/场景 -->


## 解决方案
<!-- 如何解决 -->


## 代码示例
```{{language}}
// 示例代码
```

## 注意事项
<!-- 需要注意的点 -->


## 相关链接
- [[]]
```

### 5.4 日记模板

**文件位置**: `Templates/daily-note-template.md`

```markdown
---
date: {{date}}
type: daily
---

# {{date}} 日记

## 今日重点
- 

## 工作记录
- 

## 学习笔记
- 

## 明日计划
- 

## 相关会话
- [[]]
```

---

## 六、索引文件设计

### 6.1 文档总索引

**文件位置**: `docs/INDEX.md`

```markdown
# 拼脱脱项目文档索引

> 最后更新: {{date}}

## 快速导航

| 分类 | 说明 | 链接 |
|------|------|------|
| 架构文档 | 系统架构设计 | [architecture/](./architecture/) |
| API 文档 | 接口文档 | [api/](./api/) |
| 运维手册 | 部署运维指南 | [runbooks/](./runbooks/) |

## 开发指南

- [开发规范](../DEVELOPMENT_STANDARD.md)
- [开发指南](../DEVELOPMENT.md)
- [部署指南](../DEPLOYMENT.md)

## AI 规则

- [项目规则](../.trae/rules/project_rules.md)

## 规格文档

| 功能 | 状态 | 链接 |
|------|------|------|
| ... | ... | [specs/xxx](../.trae/specs/) |

## 决策记录

| 日期 | 决策 | 链接 |
|------|------|------|
| ... | ... | [decisions/xxx](../.trae/decisions/) |

## 会话归档

| 日期 | 主题 | 链接 |
|------|------|------|
| ... | ... | [sessions/xxx](../.trae/sessions/) |
```

### 6.2 会话索引

**文件位置**: `.trae/sessions/INDEX.md`

```markdown
# 会话归档索引

> 最后更新: {{date}}

## 归档统计

- 总会话数: X
- 本月新增: X

## 会话列表

| 日期 | 主题 | 项目 | 状态 |
|------|------|------|------|
| 2026-04-27 | 智能路由优化 | pintuotuo | ✅ |
| ... | ... | ... | ... |

## 按标签分类

### routing
- [[2026-04-27-routing-optimization]]

### billing
- ...
```

---

## 七、工具与插件推荐

### 7.1 Obsidian 插件

| 插件 | 用途 | 优先级 | 安装方式 |
|------|------|--------|----------|
| **Templater** | 模板自动化 | ⭐⭐⭐ 高 | 社区插件 |
| **Dataview** | 笔记查询统计 | ⭐⭐⭐ 高 | 社区插件 |
| **Obsidian Git** | 版本控制 | ⭐⭐ 中 | 社区插件 |
| **Obsidian MCP** | AI 集成 | ⭐⭐ 中 | 社区插件 |
| **Calendar** | 日历视图 | ⭐ 低 | 核心插件 |
| **Graph View** | 知识图谱 | ⭐ 低 | 核心插件 |

### 7.2 CLI 工具

| 工具 | 用途 | 安装状态 |
|------|------|----------|
| `obsidian-cli` | 命令行操作 | ✅ 已安装 |
| `ai-conversation-exporter` | AI 对话导出 | 可选安装 |

### 7.3 脚本工具

| 脚本 | 用途 | 位置 |
|------|------|------|
| `save-session.sh` | 会话归档 | `scripts/knowledge/` |
| `save-decision.sh` | 决策记录 | `scripts/knowledge/` |
| `save-knowledge.sh` | 知识沉淀 | `scripts/knowledge/` |
| `migrate-notes.sh` | 笔记迁移 | `scripts/knowledge/` |

---

## 八、工作流程图

### 8.1 会话沉淀流程

```
┌─────────────────────────────────────────────────────────────────┐
│                     会话沉淀工作流                                │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   AI 对话窗口                                                    │
│   ┌──────────────────────────────────────────┐                  │
│   │ 用户: 帮我优化路由系统...                  │                  │
│   │ AI: 好的，让我分析一下...                 │                  │
│   │ ...                                      │                  │
│   │ [对话结束，有价值内容]                    │                  │
│   └──────────────────────────────────────────┘                  │
│                      │                                          │
│                      ▼                                          │
│   触发沉淀命令                                                   │
│   ┌──────────────────────────────────────────┐                  │
│   │ /save-session "路由优化" pintuotuo routing │                  │
│   │ 或                                        │                  │
│   │ ./save-session.sh "路由优化" pintuotuo    │                  │
│   └──────────────────────────────────────────┘                  │
│                      │                                          │
│                      ▼                                          │
│   脚本处理                                                       │
│   ┌──────────────────────────────────────────┐                  │
│   │ 1. 提取对话关键信息                        │                  │
│   │ 2. 填充模板                               │                  │
│   │ 3. 生成 frontmatter                       │                  │
│   │ 4. 写入文件                               │                  │
│   │ 5. 更新索引                               │                  │
│   └──────────────────────────────────────────┘                  │
│                      │                                          │
│                      ▼                                          │
│   Obsidian Vault                                                │
│   ┌──────────────────────────────────────────┐                  │
│   │ 01_Projects/pintuotuo/sessions/           │                  │
│   │ └── 2026-04-27-路由优化.md                │                  │
│   └──────────────────────────────────────────┘                  │
│                      │                                          │
│                      ▼                                          │
│   自动同步（软链接）                                              │
│   ┌──────────────────────────────────────────┐                  │
│   │ 项目 .trae/sessions/ 同步更新             │                  │
│   └──────────────────────────────────────────┘                  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 8.2 日常知识管理流程

```
┌─────────────────────────────────────────────────────────────────┐
│                     日常知识管理流程                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   每日                                                           │
│   ┌─────────────────────────────────────────────────────────┐   │
│   │ • 记录工作内容到日记                                      │   │
│   │ • 有价值的对话及时沉淀                                    │   │
│   │ • 新学知识记录到 Resources                                │   │
│   └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│   每周                                                           │
│   ┌─────────────────────────────────────────────────────────┐   │
│   │ • 清理 Inbox，分类到正确位置                              │   │
│   │ • 回顾本周会话，补充遗漏的沉淀                            │   │
│   │ • 更新项目进度文档                                        │   │
│   └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│   每月                                                           │
│   ┌─────────────────────────────────────────────────────────┐   │
│   │ • 归档不活跃内容到 Archives                              │   │
│   │ • 检查知识图谱，补充缺失链接                              │   │
│   │ • 回顾决策记录，更新状态                                  │   │
│   └─────────────────────────────────────────────────────────┘   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 九、成功指标

| 指标 | 目标 | 衡量方式 | 检查频率 |
|------|------|----------|----------|
| 文档可发现性 | 90% 文档有索引 | 索引覆盖率 | 每周 |
| 知识沉淀率 | 重要会话 100% 归档 | 归档比例 | 每周 |
| AI 可访问性 | AI 能读写知识库 | CLI/MCP 功能测试 | 每次使用 |
| 知识复用率 | 减少 50% 重复解释 | 对话上下文长度 | 每月 |
| Inbox 清空率 | 每周清空 | Inbox 文件数 | 每周 |

---

## 十、风险与缓解

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| 学习成本高 | 中 | 中 | 提供详细模板和脚本，渐进式采用 |
| 工具依赖 | 低 | 低 | 保留手动流程作为备选 |
| 软链接失效 | 中 | 低 | 使用绝对路径，定期检查 |
| 信息过载 | 中 | 中 | 定期归档，保持 Inbox 清空 |
| 笔记命名冲突 | 低 | 低 | 使用日期前缀，规范命名 |

---

## 十一、下一步行动

### 立即可执行

1. **确认方案** - 用户确认整体方案
2. **执行 Phase 1** - 创建目录结构，迁移笔记
3. **创建模板** - 建立标准化模板
4. **创建脚本** - 编写沉淀脚本
5. **试点运行** - 选择一个会话进行完整流程测试

### 后续规划

1. **Phase 2** - 完善沉淀机制，建立习惯
2. **Phase 3** - 评估是否需要 MCP 插件
3. **Phase 4** - 持续优化，建立知识管理文化

---

**方案版本**: v2.0
**创建日期**: 2026-04-27
**更新日期**: 2026-04-27
**参考资源**:
- [Obsidian CLI](https://obsidian.md/cli)
- [Obsidian MCP Plugin](https://github.com/aaronsb/obsidian-mcp-plugin)
- [PARA Method](https://fortelabs.com/blog/para-method/)
- [AI Conversation Exporter](https://pypi.org/project/ai-conversation-exporter/)
