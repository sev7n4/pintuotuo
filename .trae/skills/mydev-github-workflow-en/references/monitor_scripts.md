# CI Chain Monitor Scripts

## Trigger Logic

```
Push Trigger: CI/CD + Integration
PR Trigger:   E2E
```

---

## Common Monitor Template

```bash
# Real-time log monitoring (poll every 30 seconds)
while true; do
  LOGS=$(gh run view {run-id} --log 2>&1)
  
  # Capture error logs
  ERRORS=$(echo "$LOGS" | grep -E "FAIL|Error|error|✘|panic")
  
  if [ -n "$ERRORS" ]; then
    # Error found → immediately stop workflow
    gh run cancel {run-id}
    # Go to Step 13 to analyze error type
    break
  fi
  
  # Check if workflow is completed
  STATUS=$(gh run view {run-id} --json status -q '.status')
  [ "$STATUS" = "completed" ] && break
  
  sleep 30
done
```

---

## Step 10: CI Monitor (Push Trigger)

**Trigger Condition**: `git push` to remote branch

### 10.1 CI/CD Pipeline Monitor

**Verification**: Full pass

```bash
# Get latest workflow run
RUN_ID=$(gh run list --workflow="ci-cd.yml" --limit 1 --json databaseId -q '.[0].databaseId')

# Use common monitor template
# Failed → Step 13 → Step 6 (code error)
# Passed → Proceed to 10.2 Integration Monitor
```

### 10.2 Integration Tests Monitor

**Verification**: Full pass

```bash
# Get latest workflow run
RUN_ID=$(gh run list --workflow="integration.yml" --limit 1 --json databaseId -q '.[0].databaseId')

# Use common monitor template
# Failed → Step 13 → Step 4 (requirement misunderstanding)
# Passed → Step 11 (Create PR)
```

---

## Step 12: E2E Monitor (PR Trigger)

**Trigger Condition**: PR creation/update

**Verification**: current_fix_cases pass (other cases can fail)

```bash
# Read current_fix_cases from state file (dynamic)
CURRENT_CASES=$(cat .trae/skills/mydev-github-workflow/scripts/workflow_state.json | jq -r '.current_fix_cases | join("|")')

# Get E2E workflow run triggered by PR
PR_NUMBER=$(cat .trae/skills/mydev-github-workflow/scripts/workflow_state.json | jq -r '.pr_number')
RUN_ID=$(gh run list --workflow="e2e.yml" --branch="$(git branch --show-current)" --limit 1 --json databaseId -q '.[0].databaseId')

# Real-time log monitoring (poll every 30 seconds)
while true; do
  LOGS=$(gh run view $RUN_ID --log 2>&1)
  
  # Only capture current_fix_cases related error logs (dynamic match)
  CURRENT_ERRORS=$(echo "$LOGS" | grep -E "$CURRENT_CASES" | grep -E "FAIL|✘|Error")
  
  if [ -n "$CURRENT_ERRORS" ]; then
    # current_fix_cases has error → immediately stop workflow
    gh run cancel $RUN_ID
    # Go to Step 13 to analyze error type
    break
  fi
  
  # Other case errors → ignore, continue execution
  
  # Check if workflow is completed
  STATUS=$(gh run view $RUN_ID --json status -q '.status')
  [ "$STATUS" = "completed" ] && break
  
  sleep 30
done

# current_fix_cases failed → Step 13 → Step 4/6
# current_fix_cases passed → Step 14 (allow merge)
```

---

## Error Log Analysis

```bash
# Get failed logs
gh run view {run-id} --log-failed

# Common error patterns
# ├─ Compile errors: "undefined", "type error", "syntax error"
# ├─ Test failures: "FAIL", "expected", "got"
# └─ Environment issues: "timeout", "connection refused", "ECONNREFUSED"
```

---

## State Write Timing

| Step | Write Content | Description |
|------|---------------|-------------|
| Step 9 | `ci_status.cicd = running` | Push triggers CI/CD |
| Step 10.1 | `ci_status.cicd = passed/failed` | CI/CD result |
| Step 10.2 | `ci_status.integration = passed/failed` | Integration result |
| Step 11 | `pr_number` | PR number |
| Step 12 | `ci_status.e2e = passed/failed` | E2E result |
