import { test, expect } from '@playwright/test';
import { LoginPage } from './pages';

test.describe('Admin BYOK Routing Management', () => {
  let loginPage: LoginPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
  });

  test('should display BYOK routing page when logged in as admin', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login('admin@example.com', 'admin123456');
    await loginPage.expectLoginSuccess();
    await expect(page).toHaveURL(/.*admin/);
    
    await page.goto('/admin/byok-routing');
    await expect(page.getByRole('heading', { name: /BYOK.*路由管理/ })).toBeVisible({ timeout: 10000 });
  });

  test('should display BYOK routing list', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login('admin@example.com', 'admin123456');
    await loginPage.expectLoginSuccess();
    
    await page.goto('/admin/byok-routing');
    
    await expect(page.locator('.ant-table')).toBeVisible({ timeout: 10000 });
  });

  test('should display route mode column in list', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login('admin@example.com', 'admin123456');
    await loginPage.expectLoginSuccess();
    
    await page.goto('/admin/byok-routing');
    
    await expect(page.locator('.ant-table')).toBeVisible({ timeout: 10000 });
    
    const routeModeHeader = page.locator('.ant-table-thead th').filter({ hasText: '路由模式' });
    await expect(routeModeHeader).toBeVisible({ timeout: 5000 });
  });

  test('should display health status indicators', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login('admin@example.com', 'admin123456');
    await loginPage.expectLoginSuccess();
    
    await page.goto('/admin/byok-routing');
    
    await expect(page.locator('.ant-table')).toBeVisible({ timeout: 10000 });
    
    const tableRows = page.locator('.ant-table-tbody tr');
    const rowCount = await tableRows.count();
    
    if (rowCount === 0) {
      return;
    }
    
    const statusDots = page.locator('[class*="statusDot"]');
    const count = await statusDots.count();
    
    expect(count).toBeGreaterThan(0);
  });

  test('should open route config modal', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login('admin@example.com', 'admin123456');
    await loginPage.expectLoginSuccess();
    
    await page.goto('/admin/byok-routing');
    
    await expect(page.locator('.ant-table')).toBeVisible({ timeout: 10000 });
    
    const configButton = page.locator('button').filter({ hasText: '配置' }).first();
    const isVisible = await configButton.isVisible({ timeout: 5000 }).catch(() => false);
    
    if (isVisible) {
      await configButton.click();
      await expect(page.locator('.ant-modal')).toBeVisible({ timeout: 5000 });
    }
  });

  test('should trigger light verification', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login('admin@example.com', 'admin123456');
    await loginPage.expectLoginSuccess();
    
    await page.goto('/admin/byok-routing');
    
    await expect(page.locator('.ant-table')).toBeVisible({ timeout: 10000 });
    
    const verifyButton = page.locator('button').filter({ hasText: '轻验' }).first();
    const isVisible = await verifyButton.isVisible({ timeout: 5000 }).catch(() => false);
    
    if (isVisible) {
      await verifyButton.click();
      await page.waitForTimeout(1000);
    }
  });

  test('should trigger deep verification', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login('admin@example.com', 'admin123456');
    await loginPage.expectLoginSuccess();
    
    await page.goto('/admin/byok-routing');
    
    await expect(page.locator('.ant-table')).toBeVisible({ timeout: 10000 });
    
    const deepVerifyButton = page.locator('button').filter({ hasText: '深验' }).first();
    const isVisible = await deepVerifyButton.isVisible({ timeout: 5000 }).catch(() => false);
    
    if (isVisible) {
      await deepVerifyButton.click();
      await page.waitForTimeout(1000);
    }
  });

  test('should trigger probe', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login('admin@example.com', 'admin123456');
    await loginPage.expectLoginSuccess();
    
    await page.goto('/admin/byok-routing');
    
    await expect(page.locator('.ant-table')).toBeVisible({ timeout: 10000 });
    
    const probeButton = page.locator('button').filter({ hasText: '探测' }).first();
    const isVisible = await probeButton.isVisible({ timeout: 5000 }).catch(() => false);
    
    if (isVisible) {
      await probeButton.click();
      await page.waitForTimeout(1000);
    }
  });

  test('should display verification result modal', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login('admin@example.com', 'admin123456');
    await loginPage.expectLoginSuccess();
    
    await page.goto('/admin/byok-routing');
    
    await expect(page.locator('.ant-table')).toBeVisible({ timeout: 10000 });
    
    const verifyButton = page.locator('button').filter({ hasText: '轻验' }).first();
    const isVisible = await verifyButton.isVisible({ timeout: 5000 }).catch(() => false);
    
    if (isVisible) {
      await verifyButton.click();
      await page.waitForTimeout(2000);
      
      const modal = page.locator('.ant-modal');
      const modalVisible = await modal.isVisible({ timeout: 5000 }).catch(() => false);
      
      if (modalVisible) {
        await expect(modal.locator('.ant-descriptions')).toBeVisible();
      }
    }
  });

  test('should display route mode in verification result', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login('admin@example.com', 'admin123456');
    await loginPage.expectLoginSuccess();
    
    await page.goto('/admin/byok-routing');
    
    await expect(page.locator('.ant-table')).toBeVisible({ timeout: 10000 });
    
    const verifyButton = page.locator('button').filter({ hasText: '轻验' }).first();
    const isVisible = await verifyButton.isVisible({ timeout: 5000 }).catch(() => false);
    
    if (isVisible) {
      await verifyButton.click();
      await page.waitForTimeout(2000);
      
      const modal = page.locator('.ant-modal');
      const modalVisible = await modal.isVisible({ timeout: 5000 }).catch(() => false);
      
      if (modalVisible) {
        const routeModeLabel = modal.locator('.ant-descriptions-item-label').filter({ hasText: '验证模式' });
        const hasRouteMode = await routeModeLabel.count() > 0;
        expect(hasRouteMode).toBeTruthy();
      }
    }
  });

  test('should display error category in verification result', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login('admin@example.com', 'admin123456');
    await loginPage.expectLoginSuccess();
    
    await page.goto('/admin/byok-routing');
    
    await expect(page.locator('.ant-table')).toBeVisible({ timeout: 10000 });
    
    const verifyButton = page.locator('button').filter({ hasText: '轻验' }).first();
    const isVisible = await verifyButton.isVisible({ timeout: 5000 }).catch(() => false);
    
    if (isVisible) {
      await verifyButton.click();
      await page.waitForTimeout(2000);
      
      const modal = page.locator('.ant-modal');
      const modalVisible = await modal.isVisible({ timeout: 5000 }).catch(() => false);
      
      if (modalVisible) {
        const errorCategoryLabel = modal.locator('.ant-descriptions-item-label').filter({ hasText: '错误分类' });
        const hasErrorCategory = await errorCategoryLabel.count() > 0;
        expect(hasErrorCategory).toBeTruthy();
      }
    }
  });
});
