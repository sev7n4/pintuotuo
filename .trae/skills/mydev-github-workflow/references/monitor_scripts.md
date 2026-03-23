# CI 链路监控脚本

## 触发逻辑

```
PR 创建触发完整 CI 链路: CI/CD → Integration → E2E
注意: Push 到 feature 分支不会触发 CI，只有 PR 才会触发
```

---

## Step 9: Push + 创建 PR

```bash
# 提交代码
git add .
git commit -m "{type}: {description}"

# 推送到远程
git push -u origin {branch-name}

# 创建 PR (触发 CI 链路)
gh pr create --title "{title}" --body "{description}"
```

---

## Step 10: CI监控 (PR触发)

**触发条件**: PR 创建

### 通用监控模板

```bash
# 获取分支名称
BRANCH=$(git branch --show-current)

# 获取 PR 触发的 workflow runs
gh run list --branch=$BRANCH --limit 5 --json databaseId,name,status,conclusion

# 实时监控日志（每30秒轮询）
while true; do
  LOGS=$(gh run view {run-id} --log 2>&1)
  
  # 捕捉错误日志
  ERRORS=$(echo "$LOGS" | grep -E "FAIL|Error|error|✘|panic")
  
  if [ -n "$ERRORS" ]; then
    # 发现错误 → 进入 Step 11 分析错误类型
    break
  fi
  
  # 检查工作流是否完成
  STATUS=$(gh run view {run-id} --json status -q '.status')
  [ "$STATUS" = "completed" ] && break
  
  sleep 30
done
```

### 10.1 CI/CD Pipeline 监控

**验证要求**: 全量通过

```bash
# 获取 CI/CD workflow run
RUN_ID=$(gh run list --workflow="ci-cd.yml" --branch=$BRANCH --limit 1 --json databaseId -q '.[0].databaseId')

# 监控状态
gh run watch $RUN_ID

# 失败 → Step 11 → Step 6 (代码错误)
# 成功 → 进入 10.2 Integration 监控
```

### 10.2 Integration Tests 监控

**验证要求**: 全量通过

**触发方式**: workflow_run (CI/CD 完成后自动触发)

```bash
# 获取 Integration workflow run
RUN_ID=$(gh run list --workflow="integration-tests.yml" --branch=$BRANCH --limit 1 --json databaseId -q '.[0].databaseId')

# 监控状态
gh run watch $RUN_ID

# 失败 → Step 11 → Step 4 (需求理解错误)
# 成功 → 进入 10.3 E2E 监控
```

### 10.3 E2E Tests 监控

**验证要求**: current_fix_cases 通过（其他用例失败可忽略）

**触发方式**: workflow_run (Integration 完成后自动触发)

```bash
# 从状态文件读取 current_fix_cases（动态获取）
CURRENT_CASES=$(cat .trae/skills/mydev-github-workflow/scripts/workflow_state.json | jq -r '.current_fix_cases | join("|")')

# 获取 E2E workflow run
RUN_ID=$(gh run list --workflow="e2e-tests.yml" --branch=$BRANCH --limit 1 --json databaseId -q '.[0].databaseId')

# 实时监控日志（每30秒轮询）
while true; do
  LOGS=$(gh run view $RUN_ID --log 2>&1)
  
  # 仅捕捉 current_fix_cases 相关的错误日志（动态匹配）
  CURRENT_ERRORS=$(echo "$LOGS" | grep -E "$CURRENT_CASES" | grep -E "FAIL|✘|Error")
  
  if [ -n "$CURRENT_ERRORS" ]; then
    # current_fix_cases 有错误 → 进入 Step 11
    break
  fi
  
  # 其他用例错误 → 忽略，继续执行
  
  # 检查工作流是否完成
  STATUS=$(gh run view $RUN_ID --json status -q '.status')
  [ "$STATUS" = "completed" ] && break
  
  sleep 30
done

# current_fix_cases 失败 → Step 11 → Step 4/6
# current_fix_cases 通过 → Step 12 (允许合并)
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
| Step 9 | `pr_number`, `ci_status.cicd = running` | PR 创建，CI 开始 |
| Step 10.1 | `ci_status.cicd = passed/failed` | CI/CD 结果 |
| Step 10.2 | `ci_status.integration = passed/failed` | Integration 结果 |
| Step 10.3 | `ci_status.e2e = passed/failed` | E2E 结果 |
| Step 11 | `error_history`, `retry_count` | 错误记录 |
| Step 12 | `merged = true` | PR 合并 |

---

## 快速命令参考

```bash
# 查看所有 workflow runs
gh run list --limit 10

# 查看特定分支的 workflow runs
gh run list --branch={branch-name}

# 查看 workflow 运行详情
gh run view {run-id}

# 查看失败日志
gh run view {run-id} --log-failed

# 取消 workflow
gh run cancel {run-id}

# 重新运行 workflow
gh run rerun {run-id}

# 实时监控 workflow
gh run watch {run-id}
```
