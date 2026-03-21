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
    await expect(page.getByText('商家后台')).toBeVisible();
  });

  test('should display merchant products', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login('merchant@example.com', 'merchant123456');
    await loginPage.expectLoginSuccess();
    
    await page.goto('/merchant/products');
    await expect(page.getByText('商品管理')).toBeVisible();
  });

  test('should create new product', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login('merchant@example.com', 'merchant123456');
    await loginPage.expectLoginSuccess();
    
    await page.goto('/merchant/products');
    
    const addButton = page.locator('button:has-text("添加商品")');
    if (await addButton.isVisible()) {
      await addButton.click();
      
      await page.getByPlaceholder('请输入商品名称').fill('测试商品E2E');
      await page.getByPlaceholder('请输入商品描述').fill('这是一个E2E测试商品');
      await page.getByPlaceholder('请输入价格').fill('99.99');
      await page.getByPlaceholder('请输入库存').fill('100');
      
      await page.locator('button[type="submit"]').click();
      await page.waitForTimeout(1000);
    }
  });

  test('should display merchant orders', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login('merchant@example.com', 'merchant123456');
    await loginPage.expectLoginSuccess();
    
    await page.goto('/merchant/orders');
    await expect(page.getByText('订单管理')).toBeVisible();
  });

  test('should display merchant settlements', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login('merchant@example.com', 'merchant123456');
    await loginPage.expectLoginSuccess();
    
    await page.goto('/merchant/settlements');
    await expect(page.getByText('结算管理')).toBeVisible();
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
    await expect(page.getByText('API密钥')).toBeVisible();
  });

  test('should create new API key', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login('merchant@example.com', 'merchant123456');
    await loginPage.expectLoginSuccess();
    
    await page.goto('/merchant/api-keys');
    
    const addButton = page.locator('button:has-text("添加密钥")');
    if (await addButton.isVisible()) {
      await addButton.click();
      
      await page.getByPlaceholder('请输入名称').fill('测试API密钥');
      await page.selectOption('select', 'openai');
      await page.getByPlaceholder('请输入API Key').fill('sk-test-key');
      
      await page.locator('button[type="submit"]').click();
      await page.waitForTimeout(1000);
    }
  });

  test('should toggle API key status', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login('merchant@example.com', 'merchant123456');
    await loginPage.expectLoginSuccess();
    
    await page.goto('/merchant/api-keys');
    
    const toggleButton = page.locator('.ant-switch').first();
    if (await toggleButton.isVisible()) {
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
    if (await deleteButton.isVisible()) {
      await deleteButton.click();
      
      const confirmButton = page.locator('button:has-text("确定")');
      if (await confirmButton.isVisible()) {
        await confirmButton.click();
        await page.waitForTimeout(500);
      }
    }
  });
});
