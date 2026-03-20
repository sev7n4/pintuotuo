import { test, expect } from '@playwright/test';

test.describe('Authentication', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should display login page', async ({ page }) => {
    await page.goto('/login');
    await expect(page.locator('text=拼脱脱 - 登录')).toBeVisible();
  });

  test('should show validation error for empty fields', async ({ page }) => {
    await page.goto('/login');
    await page.click('button:has-text("登录")');
    await expect(page.locator('text=请输入邮箱')).toBeVisible();
  });

  test('should navigate to register page from login', async ({ page }) => {
    await page.goto('/login');
    await page.click('text=创建新账户');
    await expect(page.locator('text=拼脱脱 - 注册')).toBeVisible();
  });

  test('should navigate to login page from register', async ({ page }) => {
    await page.goto('/register');
    await page.click('text=立即登录');
    await expect(page.locator('text=拼脱脱 - 登录')).toBeVisible();
  });

  test('should register new user and redirect to products', async ({ page }) => {
    const uniqueEmail = `test${Date.now()}@example.com`;
    
    await page.goto('/register');
    await page.fill('input[placeholder="example@email.com"]', uniqueEmail);
    await page.fill('input[placeholder="输入你的名字"]', 'testuser');
    await page.fill('input[placeholder="设置密码"]', 'Test123456!');
    await page.fill('input[placeholder="再次输入密码"]', 'Test123456!');
    await page.click('button:has-text("创建账户")');
    
    await expect(page.locator('text=注册成功')).toBeVisible({ timeout: 10000 });
    await expect(page).toHaveURL(/.*products/, { timeout: 10000 });
  });

  test('should show error for duplicate email registration', async ({ page }) => {
    await page.goto('/register');
    await page.fill('input[placeholder="example@email.com"]', 'demo@example.com');
    await page.fill('input[placeholder="输入你的名字"]', 'testuser');
    await page.fill('input[placeholder="设置密码"]', 'Test123456!');
    await page.fill('input[placeholder="再次输入密码"]', 'Test123456!');
    await page.click('button:has-text("创建账户")');
    
    await expect(page.locator('text=注册失败')).toBeVisible({ timeout: 10000 });
  });

  test('should show error for password mismatch', async ({ page }) => {
    const uniqueEmail = `test${Date.now()}@example.com`;
    
    await page.goto('/register');
    await page.fill('input[placeholder="example@email.com"]', uniqueEmail);
    await page.fill('input[placeholder="输入你的名字"]', 'testuser');
    await page.fill('input[placeholder="设置密码"]', 'Test123456!');
    await page.fill('input[placeholder="再次输入密码"]', 'DifferentPassword!');
    await page.click('button:has-text("创建账户")');
    
    await expect(page.locator('text=两次输入的密码不一致')).toBeVisible();
  });

  test('should login with valid credentials and redirect', async ({ page }) => {
    await page.goto('/login');
    await page.fill('input[placeholder="example@email.com"]', 'demo@example.com');
    await page.fill('input[placeholder="输入密码"]', 'demo123456');
    await page.click('button:has-text("登录")');
    
    await expect(page.locator('text=登录成功')).toBeVisible({ timeout: 10000 });
    await expect(page).toHaveURL(/.*products/, { timeout: 10000 });
  });

  test('should show error for invalid credentials', async ({ page }) => {
    await page.goto('/login');
    await page.fill('input[placeholder="example@email.com"]', 'invalid@example.com');
    await page.fill('input[placeholder="输入密码"]', 'wrongpassword');
    await page.click('button:has-text("登录")');
    
    await expect(page.locator('.ant-message-error, text=登录失败')).toBeVisible({ timeout: 10000 });
  });

  test('should show error for invalid email format', async ({ page }) => {
    await page.goto('/login');
    await page.fill('input[placeholder="example@email.com"]', 'invalid-email');
    await page.fill('input[placeholder="输入密码"]', 'somepassword');
    await page.click('button:has-text("登录")');
    
    await expect(page.locator('text=邮箱格式不正确')).toBeVisible();
  });

  test('should logout successfully', async ({ page }) => {
    await page.goto('/login');
    await page.fill('input[placeholder="example@email.com"]', 'demo@example.com');
    await page.fill('input[placeholder="输入密码"]', 'demo123456');
    await page.click('button:has-text("登录")');
    
    await expect(page).toHaveURL(/.*products/, { timeout: 10000 });
    
    await page.click('[data-testid="user-dropdown"]');
    await page.click('text=退出登录');
    
    await expect(page.locator('text=登录')).toBeVisible({ timeout: 5000 });
  });

  test('should display user info after login', async ({ page }) => {
    await page.goto('/login');
    await page.fill('input[placeholder="example@email.com"]', 'demo@example.com');
    await page.fill('input[placeholder="输入密码"]', 'demo123456');
    await page.click('button:has-text("登录")');
    
    await expect(page).toHaveURL(/.*products/, { timeout: 10000 });
    await expect(page.locator('[data-testid="user-dropdown"]')).toBeVisible();
  });

  test('should persist login state after page refresh', async ({ page }) => {
    await page.goto('/login');
    await page.fill('input[placeholder="example@email.com"]', 'demo@example.com');
    await page.fill('input[placeholder="输入密码"]', 'demo123456');
    await page.click('button:has-text("登录")');
    
    await expect(page).toHaveURL(/.*products/, { timeout: 10000 });
    
    await page.reload();
    
    await expect(page.locator('[data-testid="user-dropdown"]')).toBeVisible({ timeout: 5000 });
  });
});

test.describe('Merchant Registration', () => {
  test('should have role selection on register page', async ({ page }) => {
    await page.goto('/register');
    await expect(page.locator('text=用户')).toBeVisible();
    await expect(page.locator('text=商家')).toBeVisible();
  });

  test('should register as merchant and redirect to merchant dashboard', async ({ page }) => {
    const uniqueEmail = `merchant${Date.now()}@example.com`;
    
    await page.goto('/register');
    await page.click('text=商家');
    await page.fill('input[placeholder="example@email.com"]', uniqueEmail);
    await page.fill('input[placeholder="输入你的名字"]', 'merchant_user');
    await page.fill('input[placeholder="设置密码"]', 'Test123456!');
    await page.fill('input[placeholder="再次输入密码"]', 'Test123456!');
    await page.click('button:has-text("创建账户")');
    
    await expect(page.locator('text=注册成功')).toBeVisible({ timeout: 10000 });
    await expect(page).toHaveURL(/.*merchant/, { timeout: 10000 });
  });
});

test.describe('Login Redirect Tests', () => {
  test('should redirect merchant user to merchant dashboard after login', async ({ page }) => {
    await page.goto('/login');
    await page.fill('input[placeholder="example@email.com"]', 'merchant@example.com');
    await page.fill('input[placeholder="输入密码"]', 'merchant123456');
    await page.click('button:has-text("登录")');
    
    await expect(page.locator('text=登录成功')).toBeVisible({ timeout: 10000 });
    await expect(page).toHaveURL(/.*merchant/, { timeout: 10000 });
  });

  test('should redirect regular user to products page after login', async ({ page }) => {
    await page.goto('/login');
    await page.fill('input[placeholder="example@email.com"]', 'demo@example.com');
    await page.fill('input[placeholder="输入密码"]', 'demo123456');
    await page.click('button:has-text("登录")');
    
    await expect(page.locator('text=登录成功')).toBeVisible({ timeout: 10000 });
    await expect(page).toHaveURL(/.*products/, { timeout: 10000 });
  });

  test('should redirect regular user to products page on second login', async ({ page }) => {
    await page.goto('/login');
    await page.fill('input[placeholder="example@email.com"]', 'demo@example.com');
    await page.fill('input[placeholder="输入密码"]', 'demo123456');
    await page.click('button:has-text("登录")');
    
    await expect(page.locator('text=登录成功')).toBeVisible({ timeout: 10000 });
    await expect(page).toHaveURL(/.*products/, { timeout: 10000 });
    
    await page.click('[data-testid="user-dropdown"]');
    await page.click('text=退出登录');
    
    await expect(page.locator('text=登录')).toBeVisible({ timeout: 5000 });
    
    await page.fill('input[placeholder="example@email.com"]', 'demo@example.com');
    await page.fill('input[placeholder="输入密码"]', 'demo123456');
    await page.click('button:has-text("登录")');
    
    await expect(page.locator('text=登录成功')).toBeVisible({ timeout: 10000 });
    await expect(page).toHaveURL(/.*products/, { timeout: 10000 });
  });

  test('should deny access to merchant dashboard for regular user', async ({ page }) => {
    await page.goto('/login');
    await page.fill('input[placeholder="example@email.com"]', 'demo@example.com');
    await page.fill('input[placeholder="输入密码"]', 'demo123456');
    await page.click('button:has-text("登录")');
    
    await expect(page.locator('text=登录成功')).toBeVisible({ timeout: 10000 });
    
    await page.goto('/merchant/dashboard');
    
    await expect(page).not.toHaveURL(/.*merchant/, { timeout: 5000 });
  });
});
