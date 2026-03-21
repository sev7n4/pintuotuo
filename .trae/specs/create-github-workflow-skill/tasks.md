# Tasks

## 阶段一：Skill目录结构创建（Anthropic规范）

- [x] Task 1: 创建符合Anthropic规范的目录结构
  - [x] SubTask 1.1: 创建 `.trae/skills/mydev-github-workflow/` 主目录
  - [x] SubTask 1.2: 创建 `references/` 子目录（参考文档）
  - [x] SubTask 1.3: 创建 `assets/templates/` 子目录（模板资源）
  - [x] SubTask 1.4: 创建 `scripts/` 子目录（可执行脚本）

## 阶段二：核心文件创建

- [x] Task 2: 编写 SKILL.md 核心定义文件（<500行）
  - [x] SubTask 2.1: 定义YAML frontmatter（name, description ~100词）
  - [x] SubTask 2.2: 定义触发条件（清晰场景描述）
  - [x] SubTask 2.3: 定义核心工作流程（12步骤精简版）
  - [x] SubTask 2.4: 定义错误处理机制
  - [x] SubTask 2.5: 定义质量标准和注意事项

## 阶段三：参考文档迁移（references/）

- [x] Task 3: 迁移参考文档到references/
  - [x] SubTask 3.1: 复制设计文档为 `design.md`
  - [x] SubTask 3.2: 复制问题跟踪模板 `issue_tracking.md`
  - [x] SubTask 3.3: 复制工作流历史模板 `workflow_history.md`
  - [x] SubTask 3.4: 复制快速参考指南 `quick_reference.md`

## 阶段四：模板资源迁移（assets/templates/）

- [x] Task 4: 迁移模板文件到assets/templates/
  - [x] SubTask 4.1: 复制 `plan_template.md`
  - [x] SubTask 4.2: 复制 `tasks_template.md`
  - [x] SubTask 4.3: 复制 `pr_template.md`
  - [x] SubTask 4.4: 复制 `bug_report.md`
  - [x] SubTask 4.5: 复制 `feature_request.md`

## 阶段五：脚本文件迁移（scripts/）

- [x] Task 5: 迁移状态管理文件
  - [x] SubTask 5.1: 复制 `workflow_state.json` 到scripts/

## 阶段六：清理旧结构

- [x] Task 6: 清理不符合规范的目录
  - [x] SubTask 6.1: 删除 `.trae/tryskills/` 目录
  - [x] SubTask 6.2: 删除 `.trae/documents/templates/` 目录
  - [x] SubTask 6.3: 删除 `.trae/cache/` 目录

## 阶段七：文档更新

- [ ] Task 7: 更新spec文档反映新结构
  - [ ] SubTask 7.1: 更新spec.md中的目录结构说明
  - [ ] SubTask 7.2: 更新checklist.md检查项
  - [ ] SubTask 7.3: 更新tasks.md任务状态

# Task Dependencies

- Task 2-5 依赖 Task 1（需要先创建目录）
- Task 6 依赖 Task 2-5 完成
- Task 7 依赖 Task 6 完成
