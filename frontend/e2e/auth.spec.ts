import { test, expect } from '@playwright/test';

test.describe('Authentication', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should display login page', async ({ page }) => {
    await expect(page.locator('text=登录')).toBeVisible();
  });

  test('should show validation error for empty fields', async ({ page }) => {
    await page.click('button:has-text("登录")');
    await expect(page.locator('text=请输入')).toBeVisible();
  });

  test('should navigate to register page', async ({ page }) => {
    await page.click('text=注册');
    await expect(page.locator('text=注册账号')).toBeVisible();
  });

  test('should register new user', async ({ page }) => {
    const uniqueEmail = `test${Date.now()}@example.com`;
    
    await page.click('text=注册');
    await page.fill('input[placeholder="请输入用户名"]', 'testuser');
    await page.fill('input[placeholder="请输入邮箱"]', uniqueEmail);
    await page.fill('input[placeholder="请输入密码"]', 'Test123456!');
    await page.fill('input[placeholder="请确认密码"]', 'Test123456!');
    await page.click('button:has-text("注册")');
    
    await expect(page.locator('text=注册成功')).toBeVisible({ timeout: 10000 });
  });

  test('should login with valid credentials', async ({ page }) => {
    await page.fill('input[placeholder="请输入邮箱"]', 'demo@example.com');
    await page.fill('input[placeholder="请输入密码"]', 'demo123456');
    await page.click('button:has-text("登录")');
    
    await expect(page).toHaveURL(/.*home|.*products/, { timeout: 10000 });
  });

  test('should show error for invalid credentials', async ({ page }) => {
    await page.fill('input[placeholder="请输入邮箱"]', 'invalid@example.com');
    await page.fill('input[placeholder="请输入密码"]', 'wrongpassword');
    await page.click('button:has-text("登录")');
    
    await expect(page.locator('text=登录失败')).toBeVisible({ timeout: 10000 });
  });

  test('should logout successfully', async ({ page }) => {
    await page.fill('input[placeholder="请输入邮箱"]', 'demo@example.com');
    await page.fill('input[placeholder="请输入密码"]', 'demo123456');
    await page.click('button:has-text("登录")');
    
    await expect(page).toHaveURL(/.*home|.*products/, { timeout: 10000 });
    
    await page.click('[data-testid="user-menu"]');
    await page.click('text=退出登录');
    
    await expect(page.locator('text=登录')).toBeVisible();
  });
});
