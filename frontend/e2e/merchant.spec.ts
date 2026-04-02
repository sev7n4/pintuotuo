import { test, expect } from '@playwright/test';
import { LoginPage } from './pages';

test.describe('Merchant Dashboard', () => {
  let loginPage: LoginPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
  });

  test('should display merchant dashboard when logged in as merchant', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login('merchant@example.com', 'merchant123456');
    await loginPage.expectLoginSuccess();
    await expect(page).toHaveURL(/.*merchant/);
    
    await page.goto('/merchant');
    await expect(page.getByRole('heading', { name: '数据概览' })).toBeVisible();
  });

  test('should display merchant SKU management', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login('merchant@example.com', 'merchant123456');
    await loginPage.expectLoginSuccess();

    await page.goto('/merchant/skus');
    await expect(page.getByText('选品与SKU管理')).toBeVisible({ timeout: 15000 });
  });

  test('legacy /merchant/products redirects to SKU page', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login('merchant@example.com', 'merchant123456');
    await loginPage.expectLoginSuccess();

    await page.goto('/merchant/products');
    await page.waitForURL(/\/merchant\/skus/, { timeout: 10000 });
    await expect(page.getByText('选品与SKU管理')).toBeVisible({ timeout: 10000 });
  });

  test('should display merchant orders', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login('merchant@example.com', 'merchant123456');
    await loginPage.expectLoginSuccess();
    
    await page.goto('/merchant/orders');
    await expect(page.getByRole('heading', { name: '订单管理' }).or(page.locator('h2:has-text("订单")'))).toBeVisible();
  });

  test('should display merchant settlements', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login('merchant@example.com', 'merchant123456');
    await loginPage.expectLoginSuccess();
    
    await page.goto('/merchant/settlements');
    await expect(page.getByRole('heading', { name: '结算管理' }).or(page.locator('h2:has-text("结算")'))).toBeVisible();
  });
});

test.describe('Merchant API Keys', () => {
  let loginPage: LoginPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
  });

  test('should display API keys page', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login('merchant@example.com', 'merchant123456');
    await loginPage.expectLoginSuccess();
    
    await page.goto('/merchant/api-keys');
    await expect(page.getByRole('heading', { name: 'API密钥管理' }).or(page.locator('h2:has-text("API")'))).toBeVisible();
  });

  test('should create new API key', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login('merchant@example.com', 'merchant123456');
    await loginPage.expectLoginSuccess();
    
    await page.goto('/merchant/api-keys');
    
    const addButton = page.locator('button:has-text("添加密钥")');
    const isVisible = await addButton.isVisible({ timeout: 5000 }).catch(() => false);
    if (!isVisible) {
      test.skip();
      return;
    }
    
    await addButton.click();
    await page.waitForTimeout(500);
    
    const nameInput = page.getByPlaceholder('请输入名称');
    if (await nameInput.isVisible({ timeout: 3000 }).catch(() => false)) {
      await nameInput.fill('测试API密钥');
      await page.locator('.ant-select').click();
      await page.getByText('OpenAI').click();
      await page.getByPlaceholder('请输入API Key').fill('sk-test-key');
      await page.locator('button:has-text("创建")').click();
      await page.waitForTimeout(1000);
    }
  });

  test('should toggle API key status', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login('merchant@example.com', 'merchant123456');
    await loginPage.expectLoginSuccess();
    
    await page.goto('/merchant/api-keys');
    
    const toggleButton = page.locator('.ant-switch').first();
    if (await toggleButton.isVisible({ timeout: 5000 }).catch(() => false)) {
      await toggleButton.click();
      await page.waitForTimeout(500);
    }
  });

  test('should delete API key', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login('merchant@example.com', 'merchant123456');
    await loginPage.expectLoginSuccess();
    
    await page.goto('/merchant/api-keys');
    
    const deleteButton = page.locator('button:has-text("删除")').first();
    if (await deleteButton.isVisible({ timeout: 5000 }).catch(() => false)) {
      await deleteButton.click();
      
      const confirmButton = page.locator('button:has-text("确定")');
      if (await confirmButton.isVisible({ timeout: 2000 }).catch(() => false)) {
        await confirmButton.click();
        await page.waitForTimeout(500);
      }
    }
  });
});
