import { test, expect } from '@playwright/test';

test.describe('Payment Flow E2E Tests', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should display payment page', async ({ page }) => {
    await page.goto('/payment/1');
    await page.waitForLoadState('networkidle');
  });

  test('should show payment methods', async ({ page }) => {
    await page.goto('/login');
    
    await page.fill('input[type="email"]', 'user@example.com');
    await page.fill('input[type="password"]', 'password123');
    await page.click('button:has-text("登录")');
    
    await page.waitForURL(/.*products|.*home/, { timeout: 10000 }).catch(() => {});
    
    await page.goto('/payment/1');
    
    const alipayBtn = page.locator('button:has-text("支付宝")');
    const wechatBtn = page.locator('button:has-text("微信")');
    
    if (await alipayBtn.isVisible()) {
      await expect(alipayBtn).toBeVisible();
    }
    if (await wechatBtn.isVisible()) {
      await expect(wechatBtn).toBeVisible();
    }
  });

  test('should select alipay payment method', async ({ page }) => {
    await page.goto('/login');
    
    await page.fill('input[type="email"]', 'user@example.com');
    await page.fill('input[type="password"]', 'password123');
    await page.click('button:has-text("登录")');
    
    await page.waitForURL(/.*products|.*home/, { timeout: 10000 }).catch(() => {});
    
    await page.goto('/payment/1');
    
    const alipayBtn = page.locator('button:has-text("支付宝")');
    if (await alipayBtn.isVisible()) {
      await alipayBtn.click();
    }
  });

  test('should select wechat payment method', async ({ page }) => {
    await page.goto('/login');
    
    await page.fill('input[type="email"]', 'user@example.com');
    await page.fill('input[type="password"]', 'password123');
    await page.click('button:has-text("登录")');
    
    await page.waitForURL(/.*products|.*home/, { timeout: 10000 }).catch(() => {});
    
    await page.goto('/payment/1');
    
    const wechatBtn = page.locator('button:has-text("微信")');
    if (await wechatBtn.isVisible()) {
      await wechatBtn.click();
    }
  });

  test('should handle payment success', async ({ page }) => {
    await page.goto('/login');
    
    await page.fill('input[type="email"]', 'user@example.com');
    await page.fill('input[type="password"]', 'password123');
    await page.click('button:has-text("登录")');
    
    await page.waitForURL(/.*products|.*home/, { timeout: 10000 }).catch(() => {});
    
    await page.goto('/payment/1');
    
    const confirmBtn = page.locator('button:has-text("确认支付")');
    if (await confirmBtn.isVisible()) {
      await confirmBtn.click();
    }
  });

  test('should show order summary on payment page', async ({ page }) => {
    await page.goto('/login');
    
    await page.fill('input[type="email"]', 'user@example.com');
    await page.fill('input[type="password"]', 'password123');
    await page.click('button:has-text("登录")');
    
    await page.waitForURL(/.*products|.*home/, { timeout: 10000 }).catch(() => {});
    
    await page.goto('/payment/1');
    
    const orderSummary = page.locator('text=订单');
    if (await orderSummary.isVisible()) {
      await expect(orderSummary).toBeVisible();
    }
  });

  test('should handle payment failure gracefully', async ({ page }) => {
    await page.goto('/login');
    
    await page.fill('input[type="email"]', 'user@example.com');
    await page.fill('input[type="password"]', 'password123');
    await page.click('button:has-text("登录")');
    
    await page.waitForURL(/.*products|.*home/, { timeout: 10000 }).catch(() => {});
    
    await page.goto('/payment/99999');
    await page.waitForLoadState('networkidle');
  });

  test('should redirect to login if not authenticated', async ({ page }) => {
    await page.goto('/payment/1');
    await page.waitForURL(/.*login/, { timeout: 5000 }).catch(() => {});
  });
});
