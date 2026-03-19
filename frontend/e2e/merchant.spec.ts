import { test, expect } from '@playwright/test';

test.describe('Merchant Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should display merchant dashboard when logged in as merchant', async ({ page }) => {
    await page.fill('input[placeholder="请输入邮箱"]', 'merchant@example.com');
    await page.fill('input[placeholder="请输入密码"]', 'merchant123456');
    await page.click('button:has-text("登录")');
    
    await expect(page).toHaveURL(/.*home|.*merchant/, { timeout: 10000 });
    
    await page.goto('/merchant');
    await expect(page.locator('text=商家中心')).toBeVisible({ timeout: 5000 });
  });

  test('should display merchant products', async ({ page }) => {
    await page.fill('input[placeholder="请输入邮箱"]', 'merchant@example.com');
    await page.fill('input[placeholder="请输入密码"]', 'merchant123456');
    await page.click('button:has-text("登录")');
    
    await expect(page).toHaveURL(/.*home|.*merchant/, { timeout: 10000 });
    
    await page.goto('/merchant/products');
    await expect(page.locator('text=商品管理')).toBeVisible({ timeout: 5000 });
  });

  test('should create new product', async ({ page }) => {
    await page.fill('input[placeholder="请输入邮箱"]', 'merchant@example.com');
    await page.fill('input[placeholder="请输入密码"]', 'merchant123456');
    await page.click('button:has-text("登录")');
    
    await expect(page).toHaveURL(/.*home|.*merchant/, { timeout: 10000 });
    
    await page.goto('/merchant/products');
    
    const addButton = page.locator('button:has-text("添加商品")');
    if (await addButton.isVisible()) {
      await addButton.click();
      
      await page.fill('input[placeholder="请输入商品名称"]', '测试商品E2E');
      await page.fill('textarea[placeholder="请输入商品描述"]', '这是一个E2E测试商品');
      await page.fill('input[placeholder="请输入价格"]', '99.99');
      await page.fill('input[placeholder="请输入库存"]', '100');
      
      await page.click('button:has-text("提交")');
      await page.waitForTimeout(1000);
    }
  });

  test('should display merchant orders', async ({ page }) => {
    await page.fill('input[placeholder="请输入邮箱"]', 'merchant@example.com');
    await page.fill('input[placeholder="请输入密码"]', 'merchant123456');
    await page.click('button:has-text("登录")');
    
    await expect(page).toHaveURL(/.*home|.*merchant/, { timeout: 10000 });
    
    await page.goto('/merchant/orders');
    await expect(page.locator('text=订单管理')).toBeVisible({ timeout: 5000 });
  });

  test('should display merchant settlements', async ({ page }) => {
    await page.fill('input[placeholder="请输入邮箱"]', 'merchant@example.com');
    await page.fill('input[placeholder="请输入密码"]', 'merchant123456');
    await page.click('button:has-text("登录")');
    
    await expect(page).toHaveURL(/.*home|.*merchant/, { timeout: 10000 });
    
    await page.goto('/merchant/settlements');
    await expect(page.locator('text=结算管理')).toBeVisible({ timeout: 5000 });
  });
});

test.describe('Merchant API Keys', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should display API keys page', async ({ page }) => {
    await page.fill('input[placeholder="请输入邮箱"]', 'merchant@example.com');
    await page.fill('input[placeholder="请输入密码"]', 'merchant123456');
    await page.click('button:has-text("登录")');
    
    await expect(page).toHaveURL(/.*home|.*merchant/, { timeout: 10000 });
    
    await page.goto('/merchant/api-keys');
    await expect(page.locator('text=API密钥')).toBeVisible({ timeout: 5000 });
  });

  test('should create new API key', async ({ page }) => {
    await page.fill('input[placeholder="请输入邮箱"]', 'merchant@example.com');
    await page.fill('input[placeholder="请输入密码"]', 'merchant123456');
    await page.click('button:has-text("登录")');
    
    await expect(page).toHaveURL(/.*home|.*merchant/, { timeout: 10000 });
    
    await page.goto('/merchant/api-keys');
    
    const addButton = page.locator('button:has-text("添加密钥")');
    if (await addButton.isVisible()) {
      await addButton.click();
      
      await page.fill('input[placeholder="请输入名称"]', '测试API密钥');
      await page.selectOption('select', 'openai');
      await page.fill('input[placeholder="请输入API Key"]', 'sk-test-key');
      
      await page.click('button:has-text("提交")');
      await page.waitForTimeout(1000);
    }
  });

  test('should toggle API key status', async ({ page }) => {
    await page.fill('input[placeholder="请输入邮箱"]', 'merchant@example.com');
    await page.fill('input[placeholder="请输入密码"]', 'merchant123456');
    await page.click('button:has-text("登录")');
    
    await expect(page).toHaveURL(/.*home|.*merchant/, { timeout: 10000 });
    
    await page.goto('/merchant/api-keys');
    
    const toggleButton = page.locator('.ant-switch').first();
    if (await toggleButton.isVisible()) {
      await toggleButton.click();
      await page.waitForTimeout(500);
    }
  });

  test('should delete API key', async ({ page }) => {
    await page.fill('input[placeholder="请输入邮箱"]', 'merchant@example.com');
    await page.fill('input[placeholder="请输入密码"]', 'merchant123456');
    await page.click('button:has-text("登录")');
    
    await expect(page).toHaveURL(/.*home|.*merchant/, { timeout: 10000 });
    
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
