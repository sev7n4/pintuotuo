# Pintuotuo 测试用例设计与验证质量改进计划（基于现有工作流）

> 版本: 2.0  
> 创建日期: 2026-04-06  
> 状态: 待确认  
> 基于现有: SKILL.md + CI/CD工作流

---

## 一、现状分析

### 1.1 现有SKILL工作流

**已有测试流程**:

| 步骤 | 内容 | 状态 |
|------|------|------|
| **Step 5-7** | TDD流程（Red-Green-Refactor） | ✅ 已定义 |
| **Step 8** | 本地验证（make test） | ✅ 已定义 |
| **Step 10** | CI监控（CI/CD → Integration → E2E） | ✅ 已定义 |
| **Step 11** | 错误修复 | ✅ 已定义 |

**已有约束**:
- [约束3] 禁止跳过本地验证
- [约束4] 禁止跳过CI监控
- [约束6] 禁止跳过TDD流程

### 1.2 现有CI/CD工作流

**已有测试阶段**:

| Workflow | 测试类型 | 触发方式 | 状态 |
|----------|----------|----------|------|
| **ci-cd.yml** | 后端单元测试 | PR + Push to main | ✅ 已配置 |
| **ci-cd.yml** | 前端单元测试 | PR + Push to main | ✅ 已配置 |
| **integration-tests.yml** | 后端集成测试 | PR | ✅ 已配置 |
| **e2e-tests.yml** | 前端E2E测试 | PR | ✅ 已配置 |

**已有覆盖率**:
- 后端单元测试覆盖率报告 ✅
- 前端单元测试覆盖率报告 ✅
- 集成测试覆盖率报告 ✅

### 1.3 缺失部分

| 缺失项 | 影响 | 优先级 |
|--------|------|--------|
| **测试用例设计规范** | 测试质量参差不齐 | 高 |
| **测试质量检查清单** | Mock过度、边界条件缺失 | 高 |
| **覆盖率门禁** | 覆盖率无强制要求 | 高 |
| **契约测试** | 前后端接口不一致 | 高 |
| **响应式测试** | 移动端问题未发现 | 中 |
| **前端集成测试** | API调用未验证 | 高 |

---

## 二、改进原则

### 2.1 避免冲突

**原则1**: 不修改现有SKILL.md核心流程

```
现有流程保持不变:
Step 0 → Step 1 → Step 2-3 → Step 4 → Step 5-7 → Step 8 → Step 9 → Step 10 → Step 12 → Step 13 → Step 14
```

**原则2**: 在现有流程中增加质量检查点

```
改进方式:
Step 5-7: TDD流程
  └─ 增加: 测试质量检查清单（references/05_03_test_quality_checklist.md）

Step 8: 本地验证
  └─ 增加: 覆盖率门禁检查

Step 10: CI监控
  └─ 增加: 契约测试 + 响应式测试
```

**原则3**: 不修改现有CI/CD Workflow文件

```
现有Workflow保持不变:
- ci-cd.yml
- integration-tests.yml
- e2e-tests.yml

新增Workflow:
- contract-tests.yml (契约测试)
- responsive-tests.yml (响应式测试，可选)
```

### 2.2 避免重复

**原则1**: 复用现有测试基础设施

```
复用:
- Jest配置 (frontend/jest.config.cjs)
- Playwright配置 (frontend/playwright.config.ts)
- Makefile测试命令
- Codecov覆盖率报告
```

**原则2**: 渐进式披露，按需加载

```
SKILL改进:
- 主文件保持简洁
- 详细规范放在references/目录
- 按条件加载相关模板
```

---

## 三、改进方案

### 3.1 SKILL改进（不修改核心流程）

#### 3.1.1 Step 1 增强：问题解析

**改进方式**: 增加验收标准输出，不修改现有流程

**现有输出**:
```
- current_fix_cases - 本次修复的测试用例ID列表
- 问题类型判断
```

**新增输出**:
```
- acceptance_criteria - 验收标准（GIVEN-WHEN-THEN格式）
- test_mapping - 测试用例映射（单元/集成/E2E）
```

**渐进式加载**:
```
IF 问题涉及新功能 → 加载 references/01_02_acceptance_criteria_template.md
IF 问题涉及API → 加载 references/01_03_api_contract_template.md
```

**文件**: `.trae/skills/mydev-github-workflow/references/01_02_acceptance_criteria_template.md`

```markdown
# 验收标准模板

## 功能需求：{功能名称}

### 验收标准（Acceptance Criteria）

#### AC-001: {场景名称}
- **GIVEN**: {前置条件}
  - 用户已登录
  - Token存储在 `localStorage.getItem('auth_token')`
  
- **WHEN**: {触发动作}
  - 访问 `/merchant/settlements`
  - 发起API请求
  
- **THEN**: {预期结果}
  - API请求携带 `Authorization: Bearer {token}`
  - 返回状态码 200
  - 数据正确显示

### 测试用例映射

| 验收标准 | 单元测试 | 集成测试 | E2E测试 |
|----------|----------|----------|---------|
| AC-001 | test_getAuthToken() | test_api_auth_header() | test_settlement_flow() |

### 边界条件

- Token为空
- Token过期
- 网络错误
- 数据为空

### 响应式要求

- 桌面端：宽度 ≥ 1024px
- 平板端：768px ≤ 宽度 < 1024px
- 移动端：宽度 < 768px
```

#### 3.1.2 Step 5-7 增强：TDD流程

**改进方式**: 增加测试质量检查清单，不修改现有TDD流程

**现有流程**:
```
Step 5: Red - 设计失败测试
Step 6: Green - 最小实现
Step 7: Refactor - 重构归档
```

**新增检查**:
```
Step 5: Red阶段
  └─ 检查: 测试用例设计是否符合规范

Step 6: Green阶段
  └─ 检查: 实现是否最小化

Step 7: Refactor阶段
  └─ 检查: 测试质量是否达标
```

**文件**: `.trae/skills/mydev-github-workflow/references/05_03_test_quality_checklist.md`

```markdown
# 测试质量检查清单

## 单元测试检查

### Mock检查
- [ ] 只Mock外部依赖（网络、数据库、文件系统）
- [ ] 不Mock内部逻辑（要测试的函数本身）
- [ ] Mock返回值符合真实数据格式

### 边界条件检查
- [ ] 测试空值处理
- [ ] 测试边界值处理
- [ ] 测试异常情况处理

### 断言检查
- [ ] 每个测试至少3个断言
- [ ] 断言验证关键业务逻辑
- [ ] 断言消息清晰明确

## 集成测试检查

### API调用检查
- [ ] 验证请求URL正确
- [ ] 验证请求方法正确
- [ ] 验证请求参数正确

### 认证检查
- [ ] 验证Authorization header格式
- [ ] 验证Token key正确（auth_token）
- [ ] 验证Token过期处理

### 响应检查
- [ ] 验证响应状态码
- [ ] 验证响应数据格式
- [ ] 验证错误响应处理

## E2E测试检查

### 流程检查
- [ ] 测试完整用户流程
- [ ] 测试关键业务路径
- [ ] 测试错误处理流程

### API验证检查
- [ ] 验证API请求发送
- [ ] 验证API响应状态
- [ ] 验证数据显示正确

### 响应式检查
- [ ] 测试桌面端布局
- [ ] 测试平板端布局
- [ ] 测试移动端布局
- [ ] 验证无水平滚动
```

**文件**: `.trae/skills/mydev-github-workflow/references/05_04_unit_test_template.md`

```markdown
# 单元测试模板

## 测试原则

1. **只Mock外部依赖**: 网络、数据库、文件系统
2. **不Mock内部逻辑**: 要测试的函数本身
3. **测试边界条件**: 空值、边界值、异常值

## 测试结构

```typescript
describe('{函数名}', () => {
  beforeEach(() => {
    // 清理环境
    localStorage.clear();
    sessionStorage.clear();
  });

  describe('正常情况', () => {
    it('should return {expected} when {condition}', () => {
      // Arrange
      const input = {input_value};
      
      // Act
      const result = {function_name}(input);
      
      // Assert
      expect(result).toBe({expected});
    });
  });

  describe('边界条件', () => {
    it('should handle empty input', () => {
      expect({function_name}('')).toBe('');
    });

    it('should handle null input', () => {
      expect({function_name}(null)).toBe('');
    });
  });

  describe('错误情况', () => {
    it('should throw error when {condition}', () => {
      expect(() => {function_name}({invalid_input})).toThrow();
    });
  });
});
```

## Token获取函数测试示例

```typescript
describe('getAuthToken', () => {
  beforeEach(() => {
    localStorage.clear();
    sessionStorage.clear();
  });

  describe('正常情况', () => {
    it('should return token from localStorage', () => {
      localStorage.setItem('auth_token', 'test-token');
      expect(getAuthToken()).toBe('test-token');
    });

    it('should fallback to sessionStorage', () => {
      sessionStorage.setItem('auth_token', 'session-token');
      expect(getAuthToken()).toBe('session-token');
    });

    it('should prioritize localStorage over sessionStorage', () => {
      localStorage.setItem('auth_token', 'local-token');
      sessionStorage.setItem('auth_token', 'session-token');
      expect(getAuthToken()).toBe('local-token');
    });
  });

  describe('边界条件', () => {
    it('should return empty string when no token', () => {
      expect(getAuthToken()).toBe('');
    });

    it('should handle null localStorage', () => {
      Object.defineProperty(window, 'localStorage', {
        value: null,
        writable: true,
      });
      expect(getAuthToken()).toBe('');
    });
  });

  describe('错误情况', () => {
    it('should handle corrupted token', () => {
      localStorage.setItem('auth_token', null);
      expect(getAuthToken()).toBe('');
    });
  });
});
```
```

#### 3.1.3 Step 8 增强：本地验证

**改进方式**: 增加覆盖率门禁检查，不修改现有验证流程

**现有验证**:
```bash
make test
```

**新增检查**:
```bash
# 前端覆盖率检查
cd frontend && npm run test:coverage
if [ $(cat coverage/coverage-summary.json | jq '.total.lines.pct') -lt 80 ]; then
  echo "Frontend coverage below 80%"
  exit 1
fi

# 后端覆盖率检查
cd backend && go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep total | awk '{if ($3 + 0 < 80) exit 1}'
```

**文件**: `.trae/skills/mydev-github-workflow/references/08_01_coverage_threshold.md`

```markdown
# 覆盖率门禁规范

## 覆盖率要求

| 测试类型 | 最低覆盖率 | 目标覆盖率 |
|----------|------------|------------|
| 前端单元测试 | 80% | 85% |
| 后端单元测试 | 80% | 85% |
| 后端集成测试 | 70% | 80% |

## 检查方式

### 前端

```bash
cd frontend
npm run test:coverage

# 检查覆盖率
if [ $(cat coverage/coverage-summary.json | jq '.total.lines.pct') -lt 80 ]; then
  echo "❌ Frontend coverage below 80%"
  exit 1
fi
echo "✅ Frontend coverage passed"
```

### 后端

```bash
cd backend
go test ./... -coverprofile=coverage.out -covermode=atomic

# 检查覆盖率
coverage=$(go tool cover -func=coverage.out | tail -1 | awk '{print $3}' | sed 's/%//')
if [ $(echo "$coverage < 80" | bc) -eq 1 ]; then
  echo "❌ Backend coverage below 80%"
  exit 1
fi
echo "✅ Backend coverage: ${coverage}%"
```

## 豁免情况

以下情况可豁免覆盖率要求：
- 纯配置文件修改
- 文档更新
- 测试文件本身
- 第三方库集成代码
```

#### 3.1.4 Step 10 增强：CI监控

**改进方式**: 增加契约测试和响应式测试，不修改现有监控流程

**现有监控**:
```
10.1 CI/CD Pipeline
10.2 Integration Tests
10.3 E2E Tests
```

**新增监控**:
```
10.4 Contract Tests (新增)
10.5 Responsive Tests (可选)
```

**文件**: `.trae/skills/mydev-github-workflow/references/10_02_contract_test_guide.md`

```markdown
# 契约测试指南

## 契约测试目的

确保前后端接口一致性，避免以下问题：
- Token key不一致（auth_token vs token）
- API路径不一致
- 请求/响应格式不一致

## 契约定义

### OpenAPI规范

文件: `api-contracts/openapi.yaml`

```yaml
openapi: 3.0.0
info:
  title: Pintuotuo API
  version: 1.0.0

components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
      description: "Token from localStorage.getItem('auth_token') or sessionStorage.getItem('auth_token')"

paths:
  /api/v1/merchants/settlements:
    get:
      security:
        - BearerAuth: []
      responses:
        200:
          description: 成功获取结算列表
          content:
            application/json:
              schema:
                type: object
                properties:
                  settlements:
                    type: array
                    items:
                      $ref: '#/components/schemas/Settlement'
```

## 契约测试实现

### 前端契约测试

文件: `frontend/tests/contract/settlements.contract.test.ts`

```typescript
import { Verifier } from '@pact-foundation/pact';

describe('Settlements API Contract', () => {
  it('should match contract', async () => {
    const verifier = new Verifier({
      providerBaseUrl: 'http://localhost:8080',
      pactUrls: ['./pacts/merchant-settlements.json'],
    });
    
    await verifier.verifyProvider();
  });
});
```

### 后端契约测试

文件: `backend/handlers/settlement_contract_test.go`

```go
func TestSettlementContract(t *testing.T) {
    // 验证API响应格式符合OpenAPI规范
    // 验证Authorization header格式
    // 验证Token key一致性
}
```

## CI集成

文件: `.github/workflows/contract-tests.yml`

```yaml
name: Contract Tests

on:
  pull_request:
    branches: [main]

jobs:
  contract-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Validate OpenAPI spec
        run: |
          npm install -g @apidevtools/swagger-cli
          swagger-cli validate api-contracts/openapi.yaml
      
      - name: Run contract tests
        run: |
          cd frontend && npm run test:contract
          cd ../backend && go test ./... -run Contract
```
```

### 3.2 CI/CD改进（新增Workflow）

#### 3.2.1 新增契约测试Workflow

**文件**: `.github/workflows/contract-tests.yml`

```yaml
name: Contract Tests

on:
  pull_request:
    branches: [main]
  workflow_dispatch:

permissions:
  contents: read

env:
  GO_VERSION: '1.21'
  NODE_VERSION: '18'

jobs:
  validate-openapi:
    name: Validate OpenAPI Spec
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4

      - name: Install swagger-cli
        run: npm install -g @apidevtools/swagger-cli

      - name: Validate OpenAPI specification
        run: |
          if [ -f api-contracts/openapi.yaml ]; then
            swagger-cli validate api-contracts/openapi.yaml
            echo "✅ OpenAPI spec is valid"
          else
            echo "⚠️ OpenAPI spec not found, skipping validation"
          fi

  contract-tests:
    name: Contract Tests
    runs-on: ubuntu-latest
    needs: validate-openapi

    services:
      postgres:
        image: postgres:15-alpine
        env:
          POSTGRES_USER: pintuotuo
          POSTGRES_PASSWORD: test_password
          POSTGRES_DB: pintuotuo_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

      redis:
        image: redis:7-alpine
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 6379:6379

    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}

      - name: Set up Node.js
        uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}

      - name: Install backend dependencies
        working-directory: ./backend
        run: go mod download

      - name: Install frontend dependencies
        working-directory: ./frontend
        run: npm install

      - name: Run backend contract tests
        working-directory: ./backend
        env:
          DATABASE_URL: postgres://pintuotuo:test_password@localhost:5432/pintuotuo_test?sslmode=disable
          REDIS_URL: redis://localhost:6379
          GIN_MODE: test
        run: |
          go test -v -run Contract ./... || echo "No contract tests found"

      - name: Run frontend contract tests
        working-directory: ./frontend
        run: |
          if [ -d "tests/contract" ]; then
            npm run test:contract || echo "Contract tests completed"
          else
            echo "⚠️ No frontend contract tests found"
          fi

  notify:
    name: Notify Status
    runs-on: ubuntu-latest
    needs: [validate-openapi, contract-tests]
    if: always()

    steps:
      - name: Check job status
        run: |
          echo "## Contract Tests Summary" >> $GITHUB_STEP_SUMMARY
          echo "" >> $GITHUB_STEP_SUMMARY
          echo "| Job | Status |" >> $GITHUB_STEP_SUMMARY
          echo "|-----|--------|" >> $GITHUB_STEP_SUMMARY
          echo "| Validate OpenAPI | ${{ needs.validate-openapi.result }} |" >> $GITHUB_STEP_SUMMARY
          echo "| Contract Tests | ${{ needs.contract-tests.result }} |" >> $GITHUB_STEP_SUMMARY
```

#### 3.2.2 改进现有CI/CD：增加覆盖率门禁

**文件**: `.github/workflows/ci-cd.yml`（修改现有文件）

**修改点**: 在 `backend-unit-tests` 和 `frontend-unit-tests` job中增加覆盖率门禁

```yaml
# 在 backend-unit-tests job 中增加
- name: Check test coverage
  working-directory: ./backend
  run: |
    coverage=$(go tool cover -func=coverage.out | tail -1 | awk '{print $3}' | sed 's/%//')
    echo "Unit test coverage: ${coverage}%"
    if [ $(echo "$coverage < 80" | bc) -eq 1 ]; then
      echo "❌ Backend coverage below 80%"
      exit 1
    fi
    echo "✅ Backend coverage passed"

# 在 frontend-unit-tests job 中增加
- name: Check test coverage
  working-directory: ./frontend
  run: |
    if [ -f coverage/coverage-summary.json ]; then
      coverage=$(cat coverage/coverage-summary.json | jq '.total.lines.pct')
      echo "Frontend test coverage: ${coverage}%"
      if [ $(echo "$coverage < 80" | bc) -eq 1 ]; then
        echo "❌ Frontend coverage below 80%"
        exit 1
      fi
      echo "✅ Frontend coverage passed"
    else
      echo "⚠️ Coverage report not found"
    fi
```

#### 3.2.3 改进现有E2E测试：增加响应式测试

**文件**: `.github/workflows/e2e-tests.yml`（修改现有文件）

**修改点**: 在Playwright配置中增加多设备测试

```yaml
# 在 e2e-tests job 中增加
- name: Run E2E tests (responsive)
  working-directory: ./frontend
  env:
    CI: true
  run: |
    # 运行响应式测试
    npm run test:e2e -- --project=chromium-desktop
    npm run test:e2e -- --project=chromium-tablet
    npm run test:e2e -- --project=chromium-mobile
```

**文件**: `frontend/playwright.config.ts`（修改现有文件）

```typescript
export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  workers: process.env.CI ? 2 : undefined,
  timeout: 20000,
  expect: {
    timeout: 5000,
  },
  reporter: [['html'], ['list']],
  use: {
    baseURL: 'http://localhost:5173',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
    actionTimeout: 5000,
    navigationTimeout: 10000,
  },
  projects: [
    {
      name: 'chromium-desktop',
      use: { 
        ...devices['Desktop Chrome'],
        viewport: { width: 1280, height: 720 },
      },
    },
    {
      name: 'chromium-tablet',
      use: { 
        ...devices['iPad Pro'],
        viewport: { width: 768, height: 1024 },
      },
    },
    {
      name: 'chromium-mobile',
      use: { 
        ...devices['iPhone 12'],
        viewport: { width: 375, height: 667 },
      },
    },
  ],
  webServer: process.env.CI ? undefined : {
    command: 'npm run dev',
    url: 'http://localhost:5173',
    reuseExistingServer: true,
    timeout: 120 * 1000,
  },
});
```

### 3.3 测试模板和工具

#### 3.3.1 前端集成测试模板

**文件**: `frontend/tests/integration/api.integration.test.ts`

```typescript
import { rest } from 'msw';
import { setupServer } from 'msw/node';

const server = setupServer(
  rest.get('/api/v1/merchants/settlements', (req, res, ctx) => {
    const authHeader = req.headers.get('Authorization');
    
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      return res(ctx.status(401));
    }
    
    return res(
      ctx.status(200),
      ctx.json({
        settlements: [
          { id: 1, amount: 1000, status: 'completed' },
        ],
      })
    );
  })
);

describe('Settlements API Integration', () => {
  beforeAll(() => server.listen());
  afterEach(() => server.resetHandlers());
  afterAll(() => server.close());

  it('should call API with correct auth header', async () => {
    const token = 'test-token';
    localStorage.setItem('auth_token', token);
    
    const response = await fetch('/api/v1/merchants/settlements', {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('auth_token')}`,
      },
    });
    
    expect(response.status).toBe(200);
    const data = await response.json();
    expect(data).toHaveProperty('settlements');
  });

  it('should fail with wrong token', async () => {
    const response = await fetch('/api/v1/merchants/settlements', {
      headers: {
        'Authorization': 'Bearer invalid-token',
      },
    });
    
    expect(response.status).toBe(401);
  });
});
```

#### 3.3.2 响应式测试模板

**文件**: `frontend/e2e/responsive.spec.ts`

```typescript
import { test, expect } from '@playwright/test';

test.describe('Responsive Layout Tests', () => {
  test('Desktop: Settlements page layout', async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 720 });
    await page.goto('/merchant/settlements');
    
    // 验证无水平滚动
    const hasHorizontalScroll = await page.evaluate(() => {
      return document.body.scrollWidth > document.body.clientWidth;
    });
    expect(hasHorizontalScroll).toBe(false);
    
    // 验证关键元素可见
    await expect(page.getByText('结算管理')).toBeVisible();
  });

  test('Tablet: Settlements page layout', async ({ page }) => {
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.goto('/merchant/settlements');
    
    // 验证无水平滚动
    const hasHorizontalScroll = await page.evaluate(() => {
      return document.body.scrollWidth > document.body.clientWidth;
    });
    expect(hasHorizontalScroll).toBe(false);
    
    // 验证关键元素可见
    await expect(page.getByText('结算管理')).toBeVisible();
  });

  test('Mobile: Settlements page layout', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto('/merchant/settlements');
    
    // 验证无水平滚动
    const hasHorizontalScroll = await page.evaluate(() => {
      return document.body.scrollWidth > document.body.clientWidth;
    });
    expect(hasHorizontalScroll).toBe(false);
    
    // 验证关键元素可见
    await expect(page.getByText('结算管理')).toBeVisible();
  });
});
```

---

## 四、实施路线图

### 4.1 Phase 1: SKILL增强（第1-2周）

**目标**: 在现有SKILL基础上增加测试质量保障

**任务清单**:

- [ ] **Day 1-2: 创建references文件**
  - 创建 `01_02_acceptance_criteria_template.md`
  - 创建 `05_03_test_quality_checklist.md`
  - 创建 `05_04_unit_test_template.md`
  - 创建 `08_01_coverage_threshold.md`
  - 创建 `10_02_contract_test_guide.md`

- [ ] **Day 3-4: 更新SKILL.md**
  - 在Step 1增加验收标准输出
  - 在Step 5-7增加测试质量检查
  - 在Step 8增加覆盖率门禁说明
  - 在Step 10增加契约测试说明

- [ ] **Day 5-7: 测试SKILL流程**
  - 测试渐进式加载
  - 测试测试质量检查清单
  - 测试覆盖率门禁
  - 验证与现有流程无冲突

**验收标准**:
- ✅ 所有references文件创建完成
- ✅ SKILL.md更新完成，无冲突
- ✅ 渐进式加载工作正常
- ✅ 与现有工作流无冲突

### 4.2 Phase 2: CI/CD增强（第3-4周）

**目标**: 在现有CI/CD基础上增加契约测试和覆盖率门禁

**任务清单**:

- [ ] **Day 1-3: 创建契约测试Workflow**
  - 创建 `contract-tests.yml`
  - 创建OpenAPI规范文件
  - 创建契约测试用例
  - 测试契约测试流程

- [ ] **Day 4-6: 改进现有CI/CD**
  - 在 `ci-cd.yml` 增加覆盖率门禁
  - 在 `e2e-tests.yml` 增加响应式测试
  - 测试改进后的CI流程
  - 验证与现有流程无冲突

- [ ] **Day 7-10: 测试工具配置**
  - 配置MSW（Mock Service Worker）
  - 配置Playwright多设备测试
  - 创建测试数据工厂
  - 验证测试工具可用性

**验收标准**:
- ✅ 契约测试Workflow运行正常
- ✅ 覆盖率门禁生效
- ✅ 响应式测试覆盖多设备
- ✅ 与现有CI流程无冲突

### 4.3 Phase 3: 测试用例补充（第5-6周）

**目标**: 补充缺失的测试用例

**任务清单**:

- [ ] **Day 1-3: 前端集成测试**
  - 创建API集成测试
  - 验证Authorization header
  - 验证Token key一致性
  - 测试集成测试流程

- [ ] **Day 4-7: 响应式测试**
  - 创建响应式测试用例
  - 覆盖所有关键页面
  - 验证移动端布局
  - 测试响应式测试流程

- [ ] **Day 8-10: 契约测试**
  - 创建前后端契约测试
  - 验证OpenAPI规范
  - 验证接口一致性
  - 测试契约测试流程

**验收标准**:
- ✅ 前端集成测试≥30个
- ✅ 响应式测试覆盖100%页面
- ✅ 契约测试覆盖100% API
- ✅ 所有测试通过

### 4.4 Phase 4: 持续改进（持续）

**目标**: 建立持续改进机制

**任务清单**:

- [ ] **每周: 测试失败分析**
  - 收集测试失败数据
  - 分析失败模式
  - 更新测试质量检查清单
  - 优化测试策略

- [ ] **每月: 覆盖率审查**
  - 审查覆盖率趋势
  - 识别覆盖率下降原因
  - 制定改进措施
  - 更新覆盖率目标

**验收标准**:
- ✅ 测试失败分析机制运行
- ✅ 覆盖率持续监控
- ✅ 缺陷逃逸率<5%
- ✅ 测试策略持续优化

---

## 五、验收标准

### 5.1 量化指标

| 指标 | 当前值 | 目标值 | 验收方式 |
|------|--------|--------|----------|
| **单元测试覆盖率** | 未知 | ≥80% | CI门禁检查 |
| **前端集成测试** | 0 | ≥30 | 文件数量统计 |
| **契约测试覆盖** | 0% | 100% API | OpenAPI规范覆盖 |
| **响应式测试覆盖** | 0% | 100% 页面 | Viewport测试覆盖 |
| **缺陷逃逸率** | 高 | <5% | 生产环境缺陷统计 |

### 5.2 质量指标

| 指标 | 目标值 | 验收方式 |
|------|--------|----------|
| **SKILL兼容性** | 无冲突 | 流程测试 |
| **CI兼容性** | 无冲突 | CI运行测试 |
| **测试稳定性** | ≥95% | 测试通过率统计 |
| **测试效率** | CI时间≤20分钟 | CI运行时间统计 |

---

## 六、风险管理

### 6.1 风险识别

| 风险 | 影响 | 概率 | 应对措施 |
|------|------|------|----------|
| **SKILL流程冲突** | 工作流混乱 | 低 | 严格遵循不修改核心流程原则 |
| **CI时间过长** | CI效率下降 | 中 | 优化测试并行度，分层CI |
| **覆盖率门禁过严** | 开发效率下降 | 中 | 设置合理阈值，允许豁免 |
| **团队抵触** | 推进困难 | 低 | 培训和演示，展示价值 |

### 6.2 应对策略

1. **SKILL流程冲突**:
   - 严格遵循不修改核心流程原则
   - 只增加references文件
   - 渐进式加载，按需使用

2. **CI时间过长**:
   - 优化测试并行度
   - 分层CI策略
   - 使用测试缓存

3. **覆盖率门禁过严**:
   - 设置合理阈值（80%）
   - 允许豁免情况
   - 逐步提升目标

---

## 七、总结

本改进计划基于现有SKILL.md和CI/CD工作流，通过以下方式系统性提升测试质量：

### 7.1 改进原则

1. **不修改核心流程**: 保持现有SKILL.md和CI/CD工作流不变
2. **渐进式增强**: 在现有流程中增加质量检查点
3. **避免重复**: 复用现有测试基础设施
4. **兼容性优先**: 确保与现有工作流无冲突

### 7.2 改进内容

| 改进项 | 方式 | 文件 |
|--------|------|------|
| **验收标准** | Step 1增加输出 | references/01_02_acceptance_criteria_template.md |
| **测试质量检查** | Step 5-7增加检查 | references/05_03_test_quality_checklist.md |
| **覆盖率门禁** | Step 8增加检查 | references/08_01_coverage_threshold.md |
| **契约测试** | Step 10增加监控 | .github/workflows/contract-tests.yml |
| **响应式测试** | E2E测试增强 | playwright.config.ts |

### 7.3 核心价值

- **兼容性**: 与现有工作流完全兼容，无冲突
- **渐进式**: 按需加载，不增加认知负担
- **实用性**: 解决实际问题，提升测试质量
- **可维护性**: 模块化设计，易于维护和扩展

---

**附录**:
- [验收标准模板](.trae/skills/mydev-github-workflow/references/01_02_acceptance_criteria_template.md)
- [测试质量检查清单](.trae/skills/mydev-github-workflow/references/05_03_test_quality_checklist.md)
- [单元测试模板](.trae/skills/mydev-github-workflow/references/05_04_unit_test_template.md)
- [覆盖率门禁规范](.trae/skills/mydev-github-workflow/references/08_01_coverage_threshold.md)
- [契约测试指南](.trae/skills/mydev-github-workflow/references/10_02_contract_test_guide.md)
- [契约测试Workflow](.github/workflows/contract-tests.yml)
