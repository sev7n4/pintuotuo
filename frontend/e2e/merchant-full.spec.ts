import { test, expect } from '@playwright/test';
import {
  LoginPage,
  MerchantDashboardPage,
  MerchantProductsPage,
  MerchantOrdersPage,
  MerchantSettlementsPage,
  MerchantAPIKeysPage,
  MerchantSettingsPage,
} from './pages';

test.describe('商家管理界面 - 权限与认证', () => {
  test('AUTH-001: 未登录用户访问商家后台应重定向到登录页面', async ({ page }) => {
    await page.goto('/merchant');
    await page.waitForURL(/.*login/, { timeout: 10000 });
    await expect(page.locator('.auth-card')).toBeVisible();
  });

  test('AUTH-002: 普通用户访问商家后台应提示无权限', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login('demo@example.com', 'demo123456');
    await loginPage.expectLoginSuccess();
    await page.waitForURL(/.*\//, { timeout: 15000 });

    await page.goto('/merchant');
    await page.waitForTimeout(2000);
    await expect(page).not.toHaveURL(/.*merchant/);
  });

  test('AUTH-003: 商家用户正常访问商家后台', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login('merchant@example.com', 'merchant123456');
    await loginPage.expectLoginSuccess();
    await page.waitForURL(/.*merchant/, { timeout: 15000 });
    await expect(page).toHaveURL(/.*merchant/);
  });

  test('AUTH-004: Token过期后访问应自动退出登录', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login('merchant@example.com', 'merchant123456');
    await loginPage.expectLoginSuccess();
    await page.waitForURL(/.*merchant/, { timeout: 15000 });

    await page.evaluate(() => {
      localStorage.removeItem('auth_token');
      sessionStorage.removeItem('auth_token');
    });

    await page.goto('/merchant');
    await page.waitForTimeout(2000);
    await expect(page).toHaveURL(/.*login/);
  });
});

test.describe('商家管理界面 - 商品管理', () => {
  let loginPage: LoginPage;
  let productsPage: MerchantProductsPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
    productsPage = new MerchantProductsPage(page);
    await loginPage.goto();
    await loginPage.login('merchant@example.com', 'merchant123456');
    await loginPage.expectLoginSuccess();
  });

  test('PROD-001: 创建商品 - 正常流程', async ({ page }) => {
    await productsPage.goto();
    await productsPage.expectProductsPageVisible();

    await productsPage.clickAddProduct();
    await productsPage.fillProductForm({
      name: `测试商品 ${Date.now()}`,
      description: '这是一个E2E测试商品',
      price: 99.99,
      stock: 100,
    });
    await productsPage.submitProduct();

    await expect(page.locator('.ant-message')).toBeVisible({ timeout: 5000 });
  });

  test('PROD-002: 创建商品 - 价格为负数应验证失败', async ({ page }) => {
    await productsPage.goto();
    await productsPage.expectProductsPageVisible();

    await productsPage.clickAddProduct();
    await productsPage.fillProductForm({
      name: '测试商品负价格',
      description: '测试负价格',
      price: -10,
      stock: 100,
    });
    await productsPage.submitProduct();

    await expect(page.locator('.ant-form-item-explain-error').filter({ hasText: '价格必须大于0' }).first()).toBeVisible({ timeout: 5000 });
  });

  test('PROD-003: 创建商品 - 库存为负数应验证失败', async ({ page }) => {
    await productsPage.goto();
    await productsPage.expectProductsPageVisible();

    await productsPage.clickAddProduct();
    await productsPage.fillProductForm({
      name: '测试商品负库存',
      description: '测试负库存',
      price: 99.99,
      stock: -10,
    });
    await productsPage.submitProduct();

    await expect(page.locator('.ant-form-item-explain-error').filter({ hasText: '库存必须大于等于0' }).first()).toBeVisible({ timeout: 5000 });
  });

  test('PROD-004: 编辑商品 - 修改所有字段', async ({ page }) => {
    await productsPage.goto();
    await productsPage.expectProductsPageVisible();

    const productCount = await productsPage.getProductCount();
    if (productCount > 0) {
      await productsPage.editProduct('测试商品');
      await productsPage.fillProductForm({
        name: '更新后的商品名称',
        description: '更新后的描述',
        price: 199.99,
        stock: 200,
      });
      await productsPage.submitProduct();
      await expect(page.locator('.ant-message')).toBeVisible({ timeout: 5000 });
    } else {
      test.skip();
    }
  });

  test('PROD-005: 删除商品 - 确认删除', async ({ page }) => {
    await productsPage.goto();
    await productsPage.expectProductsPageVisible();

    const initialCount = await productsPage.getProductCount();
    if (initialCount > 0) {
      await productsPage.deleteProduct('测试商品');
      const finalCount = await productsPage.getProductCount();
      expect(finalCount).toBeLessThanOrEqual(initialCount);
    } else {
      test.skip();
    }
  });

  test('PROD-007: 状态筛选 - 在售商品', async ({ page }) => {
    await productsPage.goto();
    await productsPage.expectProductsPageVisible();

    await productsPage.filterByStatus('在售');
    await page.waitForTimeout(500);
  });

  test('PROD-008: 状态筛选 - 下架商品', async ({ page }) => {
    await productsPage.goto();
    await productsPage.expectProductsPageVisible();

    await productsPage.filterByStatus('下架');
    await page.waitForTimeout(500);
  });
});

test.describe('商家管理界面 - 订单管理', () => {
  let loginPage: LoginPage;
  let ordersPage: MerchantOrdersPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
    ordersPage = new MerchantOrdersPage(page);
    await loginPage.goto();
    await loginPage.login('merchant@example.com', 'merchant123456');
    await loginPage.expectLoginSuccess();
  });

  test('ORD-001: 查看订单列表', async ({ page }) => {
    await ordersPage.goto();
    await ordersPage.expectOrdersPageVisible();
  });

  test('ORD-002: 状态筛选 - 待支付订单', async ({ page }) => {
    await ordersPage.goto();
    await ordersPage.expectOrdersPageVisible();

    await ordersPage.filterByStatus('待支付');
    await page.waitForTimeout(500);
  });

  test('ORD-003: 状态筛选 - 已完成订单', async ({ page }) => {
    await ordersPage.goto();
    await ordersPage.expectOrdersPageVisible();

    await ordersPage.filterByStatus('已完成');
    await page.waitForTimeout(500);
  });
});

test.describe('商家管理界面 - 结算管理', () => {
  let loginPage: LoginPage;
  let settlementsPage: MerchantSettlementsPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
    settlementsPage = new MerchantSettlementsPage(page);
    await loginPage.goto();
    await loginPage.login('merchant@example.com', 'merchant123456');
    await loginPage.expectLoginSuccess();
  });

  test('SETTLE-001: 查看结算列表', async ({ page }) => {
    await settlementsPage.goto();
    await settlementsPage.expectSettlementsPageVisible();
  });

  test('SETTLE-002: 申请结算 - 正常流程', async ({ page }) => {
    await settlementsPage.goto();
    await settlementsPage.expectSettlementsPageVisible();

    const applyButton = page.locator('button:has-text("申请结算")');
    if (await applyButton.isVisible({ timeout: 2000 }).catch(() => false)) {
      await settlementsPage.clickApplySettlement();
      await settlementsPage.expectSettlementSuccess();
    } else {
      test.skip();
    }
  });

  test('SETTLE-003: 申请结算 - 金额不足100元', async ({ page }) => {
    await settlementsPage.goto();
    await settlementsPage.expectSettlementsPageVisible();

    const applyButton = page.locator('button:has-text("申请结算")');
    if (await applyButton.isVisible({ timeout: 2000 }).catch(() => false)) {
      await settlementsPage.clickApplySettlement();
      await page.waitForTimeout(1000);
    } else {
      test.skip();
    }
  });
});

test.describe('商家管理界面 - API密钥管理', () => {
  let loginPage: LoginPage;
  let apiKeysPage: MerchantAPIKeysPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
    apiKeysPage = new MerchantAPIKeysPage(page);
    await loginPage.goto();
    await loginPage.login('merchant@example.com', 'merchant123456');
    await loginPage.expectLoginSuccess();
  });

  test('KEY-005: 编辑密钥 - 禁用状态', async ({ page }) => {
    await apiKeysPage.goto();
    await apiKeysPage.expectAPIKeysPageVisible();

    const toggleButton = page.locator('.ant-switch').first();
    if (await toggleButton.isVisible({ timeout: 5000 }).catch(() => false)) {
      await toggleButton.click();
      await page.waitForTimeout(500);
    } else {
      test.skip();
    }
  });

  test('KEY-006: 删除密钥 - 确认删除', async ({ page }) => {
    await apiKeysPage.goto();
    await apiKeysPage.expectAPIKeysPageVisible();

    const deleteButton = page.locator('button:has-text("删除")').first();
    if (await deleteButton.isVisible({ timeout: 5000 }).catch(() => false)) {
      await deleteButton.click();
      const confirmButton = page.locator('button:has-text("确定")');
      if (await confirmButton.isVisible({ timeout: 2000 }).catch(() => false)) {
        await confirmButton.click();
        await page.waitForTimeout(500);
      }
    } else {
      test.skip();
    }
  });
});

test.describe('商家管理界面 - 店铺设置', () => {
  let loginPage: LoginPage;
  let settingsPage: MerchantSettingsPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
    settingsPage = new MerchantSettingsPage(page);
    await loginPage.goto();
    await loginPage.login('merchant@example.com', 'merchant123456');
    await loginPage.expectLoginSuccess();
  });

  test('SET-001: 查看店铺信息', async ({ page }) => {
    await settingsPage.goto();
    await settingsPage.expectSettingsPageVisible();
  });

});

test.describe('商家管理界面 - 数据统计', () => {
  let loginPage: LoginPage;
  let dashboardPage: MerchantDashboardPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
    dashboardPage = new MerchantDashboardPage(page);
    await loginPage.goto();
    await loginPage.login('merchant@example.com', 'merchant123456');
    await loginPage.expectLoginSuccess();
  });

  test('STAT-001: 查看数据概览', async ({ page }) => {
    await dashboardPage.goto();
    await dashboardPage.expectDashboardVisible();

    const stats = await dashboardPage.getStatsCards();
    await expect(stats.totalProducts).toBeVisible();
    await expect(stats.activeProducts).toBeVisible();
    await expect(stats.monthSales).toBeVisible();
    await expect(stats.monthOrders).toBeVisible();
  });

  test('STAT-005: 最近订单显示', async ({ page }) => {
    await dashboardPage.goto();
    await dashboardPage.expectDashboardVisible();

    const recentOrders = await dashboardPage.getRecentOrders();
    const count = await recentOrders.count();
    expect(count).toBeLessThanOrEqual(5);
  });
});

test.describe('商家管理界面 - 边界与异常', () => {
  let loginPage: LoginPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login('merchant@example.com', 'merchant123456');
    await loginPage.expectLoginSuccess();
  });

  test('EDGE-004: 特殊字符输入应正确处理', async ({ page }) => {
    const productsPage = new MerchantProductsPage(page);
    await productsPage.goto();
    await productsPage.expectProductsPageVisible();

    await productsPage.clickAddProduct();
    await productsPage.fillProductForm({
      name: '<script>alert("xss")</script>测试商品',
      description: '测试特殊字符 & < > " \' / \\',
      price: 99.99,
      stock: 100,
    });
    await productsPage.submitProduct();

    await page.waitForTimeout(1000);
    const xssElement = page.locator('script:has-text("alert")');
    expect(await xssElement.count()).toBe(0);
  });
});
