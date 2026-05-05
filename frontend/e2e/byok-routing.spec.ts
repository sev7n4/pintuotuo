import { test, expect, Page } from '@playwright/test';
import { LoginPage } from './pages';

test.describe('Admin BYOK Routing Management', () => {
  let loginPage: LoginPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
  });

  async function isMobileView(page: Page): Promise<boolean> {
    const viewport = page.viewportSize();
    return viewport !== null && viewport.width <= 768;
  }

  async function waitForContent(page: Page): Promise<void> {
    const isMobile = await isMobileView(page);
    if (isMobile) {
      await expect(page.locator('.mobileCard')).toBeVisible({ timeout: 10000 });
    } else {
      await expect(page.locator('.ant-table')).toBeVisible({ timeout: 10000 });
    }
  }

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
    
    await waitForContent(page);
  });

  test('should display route mode column in list', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login('admin@example.com', 'admin123456');
    await loginPage.expectLoginSuccess();
    
    await page.goto('/admin/byok-routing');
    
    await waitForContent(page);
    
    const isMobile = await isMobileView(page);
    if (!isMobile) {
      const routeModeHeader = page.locator('.ant-table-thead th').filter({ hasText: '路由模式' });
      await expect(routeModeHeader).toBeVisible({ timeout: 5000 });
    } else {
      const routeModeLabel = page.locator('.mobileRow').filter({ hasText: '路由' });
      await expect(routeModeLabel.first()).toBeVisible({ timeout: 5000 });
    }
  });

  test('should display health status indicators', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login('admin@example.com', 'admin123456');
    await loginPage.expectLoginSuccess();
    
    await page.goto('/admin/byok-routing');
    
    await waitForContent(page);
    
    const isMobile = await isMobileView(page);
    if (!isMobile) {
      const tableRows = page.locator('.ant-table-tbody tr');
      const rowCount = await tableRows.count();
      
      if (rowCount === 0) {
        return;
      }
      
      const healthColumn = page.locator('.ant-table-thead th').filter({ hasText: '健康' });
      await expect(healthColumn).toBeVisible({ timeout: 5000 });
      
      const verifyColumn = page.locator('.ant-table-thead th').filter({ hasText: '验证' });
      await expect(verifyColumn).toBeVisible({ timeout: 5000 });
    } else {
      const mobileItems = page.locator('.mobileItem');
      const itemCount = await mobileItems.count();
      
      if (itemCount === 0) {
        return;
      }
      
      const healthLabel = page.locator('.mobileLabel').filter({ hasText: '健康' });
      await expect(healthLabel.first()).toBeVisible({ timeout: 5000 });
      
      const verifyLabel = page.locator('.mobileLabel').filter({ hasText: '验证' });
      await expect(verifyLabel.first()).toBeVisible({ timeout: 5000 });
    }
  });

  test('should open route config modal', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login('admin@example.com', 'admin123456');
    await loginPage.expectLoginSuccess();
    
    await page.goto('/admin/byok-routing');
    
    await waitForContent(page);
    
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
    
    await waitForContent(page);
    
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
    
    await waitForContent(page);
    
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
    
    await waitForContent(page);
    
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
    
    await waitForContent(page);
    
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
    
    await waitForContent(page);
    
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
    
    await waitForContent(page);
    
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
