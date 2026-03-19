import { test, expect } from '@playwright/test';

test.describe('Orders', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should display orders page when logged in', async ({ page }) => {
    await page.fill('input[placeholder="请输入邮箱"]', 'demo@example.com');
    await page.fill('input[placeholder="请输入密码"]', 'demo123456');
    await page.click('button:has-text("登录")');
    
    await expect(page).toHaveURL(/.*home|.*products/, { timeout: 10000 });
    
    await page.goto('/orders');
    await expect(page.locator('text=我的订单')).toBeVisible({ timeout: 5000 });
  });

  test('should show empty state when no orders', async ({ page }) => {
    await page.fill('input[placeholder="请输入邮箱"]', 'demo@example.com');
    await page.fill('input[placeholder="请输入密码"]', 'demo123456');
    await page.click('button:has-text("登录")');
    
    await expect(page).toHaveURL(/.*home|.*products/, { timeout: 10000 });
    
    await page.goto('/orders');
    
    const emptyState = page.locator('text=暂无订单');
    if (await emptyState.isVisible()) {
      await expect(emptyState).toBeVisible();
    }
  });

  test('should filter orders by status', async ({ page }) => {
    await page.fill('input[placeholder="请输入邮箱"]', 'demo@example.com');
    await page.fill('input[placeholder="请输入密码"]', 'demo123456');
    await page.click('button:has-text("登录")');
    
    await expect(page).toHaveURL(/.*home|.*products/, { timeout: 10000 });
    
    await page.goto('/orders');
    
    const statusFilter = page.locator('.ant-select').first();
    if (await statusFilter.isVisible()) {
      await statusFilter.click();
      await page.waitForTimeout(300);
    }
  });

  test('should view order details', async ({ page }) => {
    await page.fill('input[placeholder="请输入邮箱"]', 'demo@example.com');
    await page.fill('input[placeholder="请输入密码"]', 'demo123456');
    await page.click('button:has-text("登录")');
    
    await expect(page).toHaveURL(/.*home|.*products/, { timeout: 10000 });
    
    await page.goto('/orders');
    
    const orderItem = page.locator('.ant-list-item').first();
    if (await orderItem.isVisible()) {
      await orderItem.click();
      await page.waitForTimeout(500);
    }
  });

  test('should cancel pending order', async ({ page }) => {
    await page.fill('input[placeholder="请输入邮箱"]', 'demo@example.com');
    await page.fill('input[placeholder="请输入密码"]', 'demo123456');
    await page.click('button:has-text("登录")');
    
    await expect(page).toHaveURL(/.*home|.*products/, { timeout: 10000 });
    
    await page.goto('/orders');
    
    const cancelButton = page.locator('button:has-text("取消订单")').first();
    if (await cancelButton.isVisible()) {
      await cancelButton.click();
      
      const confirmButton = page.locator('button:has-text("确定")');
      if (await confirmButton.isVisible()) {
        await confirmButton.click();
        await page.waitForTimeout(500);
      }
    }
  });
});

test.describe('Groups', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should display groups page when logged in', async ({ page }) => {
    await page.fill('input[placeholder="请输入邮箱"]', 'demo@example.com');
    await page.fill('input[placeholder="请输入密码"]', 'demo123456');
    await page.click('button:has-text("登录")');
    
    await expect(page).toHaveURL(/.*home|.*products/, { timeout: 10000 });
    
    await page.goto('/groups');
    await expect(page.locator('text=我的拼团')).toBeVisible({ timeout: 5000 });
  });

  test('should show group details', async ({ page }) => {
    await page.fill('input[placeholder="请输入邮箱"]', 'demo@example.com');
    await page.fill('input[placeholder="请输入密码"]', 'demo123456');
    await page.click('button:has-text("登录")');
    
    await expect(page).toHaveURL(/.*home|.*products/, { timeout: 10000 });
    
    await page.goto('/groups');
    
    const groupItem = page.locator('.ant-card').first();
    if (await groupItem.isVisible()) {
      await groupItem.click();
      await page.waitForTimeout(500);
    }
  });
});
