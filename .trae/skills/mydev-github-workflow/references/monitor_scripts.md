# CI 链路监控脚本

## 触发逻辑

```
Push 触发: CI/CD + Integration
PR 触发:   E2E
```

---

## 通用监控模板

```bash
# 实时监控日志（每30秒轮询）
while true; do
  LOGS=$(gh run view {run-id} --log 2>&1)
  
  # 捕捉错误日志
  ERRORS=$(echo "$LOGS" | grep -E "FAIL|Error|error|✘|panic")
  
  if [ -n "$ERRORS" ]; then
    # 发现错误 → 立即停止工作流
    gh run cancel {run-id}
    # 进入 Step 13 分析错误类型
    break
  fi
  
  # 检查工作流是否完成
  STATUS=$(gh run view {run-id} --json status -q '.status')
  [ "$STATUS" = "completed" ] && break
  
  sleep 30
done
```

---

## Step 10: CI监控 (Push触发)

**触发条件**: `git push` 到远程分支

### 10.1 CI/CD Pipeline 监控

**验证要求**: 全量通过

```bash
# 获取最新的 workflow run
RUN_ID=$(gh run list --workflow="ci-cd.yml" --limit 1 --json databaseId -q '.[0].databaseId')

# 使用通用监控模板
# 失败 → Step 13 → Step 6 (代码错误)
# 成功 → 进入 10.2 Integration 监控
```

### 10.2 Integration Tests 监控

**验证要求**: 全量通过

```bash
# 获取最新的 workflow run
RUN_ID=$(gh run list --workflow="integration.yml" --limit 1 --json databaseId -q '.[0].databaseId')

# 使用通用监控模板
# 失败 → Step 13 → Step 4 (需求理解错误)
# 成功 → Step 11 (创建PR)
```

---

## Step 12: E2E监控 (PR触发)

**触发条件**: PR 创建/更新

**验证要求**: current_fix_cases 通过（其他用例失败可忽略）

```bash
# 从状态文件读取 current_fix_cases（动态获取）
CURRENT_CASES=$(cat .trae/skills/mydev-github-workflow/scripts/workflow_state.json | jq -r '.current_fix_cases | join("|")')

# 获取 PR 触发的 E2E workflow run
PR_NUMBER=$(cat .trae/skills/mydev-github-workflow/scripts/workflow_state.json | jq -r '.pr_number')
RUN_ID=$(gh run list --workflow="e2e.yml" --branch="$(git branch --show-current)" --limit 1 --json databaseId -q '.[0].databaseId')

# 实时监控日志（每30秒轮询）
while true; do
  LOGS=$(gh run view $RUN_ID --log 2>&1)
  
  # 仅捕捉 current_fix_cases 相关的错误日志（动态匹配）
  CURRENT_ERRORS=$(echo "$LOGS" | grep -E "$CURRENT_CASES" | grep -E "FAIL|✘|Error")
  
  if [ -n "$CURRENT_ERRORS" ]; then
    # current_fix_cases 有错误 → 立即停止工作流
    gh run cancel $RUN_ID
    # 进入 Step 13 分析错误类型
    break
  fi
  
  # 其他用例错误 → 忽略，继续执行
  
  # 检查工作流是否完成
  STATUS=$(gh run view $RUN_ID --json status -q '.status')
  [ "$STATUS" = "completed" ] && break
  
  sleep 30
done

# current_fix_cases 失败 → Step 13 → Step 4/6
# current_fix_cases 通过 → Step 14 (允许合并)
```

---

## 错误日志分析

```bash
# 获取失败日志
gh run view {run-id} --log-failed

# 常见错误模式
# ├─ 编译错误: "undefined", "type error", "syntax error"
# ├─ 测试失败: "FAIL", "expected", "got"
# └─ 环境问题: "timeout", "connection refused", "ECONNREFUSED"
```

---

## 状态写入时机

| 步骤 | 写入内容 | 说明 |
|------|----------|------|
| Step 9 | `ci_status.cicd = running` | Push 触发 CI/CD |
| Step 10.1 | `ci_status.cicd = passed/failed` | CI/CD 结果 |
| Step 10.2 | `ci_status.integration = passed/failed` | Integration 结果 |
| Step 11 | `pr_number` | PR 编号 |
| Step 12 | `ci_status.e2e = passed/failed` | E2E 结果 |
