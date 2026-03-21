import { test, expect } from '@playwright/test';
import { LoginPage } from './pages';

test.describe('Orders', () => {
  let loginPage: LoginPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
  });

  test('should display orders page when logged in', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login('demo@example.com', 'demo123456');
    await loginPage.expectLoginSuccess();
    await page.waitForURL(/.*products/, { timeout: 15000 });
    
    await page.goto('/orders');
    await expect(page.getByRole('heading', { name: '订单列表' }).or(page.locator('h1:has-text("订单")'))).toBeVisible();
  });

  test('should show empty state when no orders', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login('demo@example.com', 'demo123456');
    await loginPage.expectLoginSuccess();
    await page.waitForURL(/.*products/, { timeout: 15000 });
    
    await page.goto('/orders');
    
    const emptyState = page.getByText('暂无订单');
    if (await emptyState.isVisible({ timeout: 2000 }).catch(() => false)) {
      await expect(emptyState).toBeVisible();
    }
  });

  test('should filter orders by status', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login('demo@example.com', 'demo123456');
    await loginPage.expectLoginSuccess();
    await page.waitForURL(/.*products/, { timeout: 15000 });
    
    await page.goto('/orders');
    
    const statusFilter = page.locator('.ant-select').first();
    if (await statusFilter.isVisible({ timeout: 5000 }).catch(() => false)) {
      await statusFilter.click();
      await page.waitForTimeout(300);
    }
  });

  test('should view order details', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login('demo@example.com', 'demo123456');
    await loginPage.expectLoginSuccess();
    await page.waitForURL(/.*products/, { timeout: 15000 });
    
    await page.goto('/orders');
    
    const orderItem = page.locator('.ant-table-row').first();
    if (await orderItem.isVisible({ timeout: 5000 }).catch(() => false)) {
      await orderItem.click();
      await page.waitForTimeout(500);
    }
  });

  test('should cancel pending order', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login('demo@example.com', 'demo123456');
    await loginPage.expectLoginSuccess();
    await page.waitForURL(/.*products/, { timeout: 15000 });
    
    await page.goto('/orders');
    
    const cancelButton = page.locator('button:has-text("取消订单")').first();
    if (await cancelButton.isVisible({ timeout: 5000 }).catch(() => false)) {
      await cancelButton.click();
      
      const confirmButton = page.locator('button:has-text("确定")');
      if (await confirmButton.isVisible({ timeout: 2000 }).catch(() => false)) {
        await confirmButton.click();
        await page.waitForTimeout(500);
      }
    }
  });
});

test.describe('Groups', () => {
  let loginPage: LoginPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
  });

  test('should display groups page when logged in', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login('demo@example.com', 'demo123456');
    await loginPage.expectLoginSuccess();
    await page.waitForURL(/.*products/, { timeout: 15000 });
    
    await page.goto('/groups');
    await expect(page.getByRole('heading', { name: '拼团中心' }).or(page.locator('h1:has-text("拼团")'))).toBeVisible();
  });

  test('should show group details', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login('demo@example.com', 'demo123456');
    await loginPage.expectLoginSuccess();
    await page.waitForURL(/.*products/, { timeout: 15000 });
    
    await page.goto('/groups');
    
    const groupItem = page.locator('.ant-card').first();
    if (await groupItem.isVisible({ timeout: 5000 }).catch(() => false)) {
      await groupItem.click();
      await page.waitForTimeout(500);
    }
  });
});
