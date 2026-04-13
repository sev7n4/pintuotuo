import { test, expect } from '@playwright/test';
import { LoginPage, RegisterPage } from './pages';

test.beforeEach(async ({ page }) => {
  page.on('console', msg => {
    console.log('BROWSER:', msg.type(), msg.text());
  });
  page.on('pageerror', error => {
    console.error('PAGE ERROR:', error.message);
  });
});

test.describe('Authentication', () => {
  test('should display login page', async ({ page }) => {
    await page.goto('/login');
    await expect(page.getByText('拼脱脱 - 登录')).toBeVisible();
  });

  test('should show validation error for empty fields', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await page.locator('button[type="submit"]').click();
    await expect(page.getByText('请输入邮箱')).toBeVisible();
  });

  test('should navigate between login and register', async ({ page }) => {
    await page.goto('/login');
    await page.waitForLoadState('networkidle');
    await page.getByText('创建新账户').click();
    await expect(page.getByText('拼脱脱 - 注册')).toBeVisible();
    
    await page.getByText('立即登录').click();
    await expect(page.getByText('拼脱脱 - 登录')).toBeVisible();
  });

  test('should show error for invalid email format', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await page.getByPlaceholder('example@email.com').fill('invalid-email');
    await page.getByPlaceholder('输入密码').fill('somepassword');
    await page.locator('button[type="submit"]').click();
    await expect(page.getByText('邮箱格式不正确')).toBeVisible();
  });
});

test.describe('Login Flow', () => {
  let loginPage: LoginPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
    await loginPage.goto();
  });

  test('should login as regular user and redirect to home', async ({ page }) => {
    await loginPage.login('demo@example.com', 'demo123456');
    await loginPage.expectLoginSuccess();
    await page.waitForURL(/.*\//, { timeout: 15000 });
    await expect(page).toHaveURL(/.*\//);
  });

  test('should login as merchant and redirect to merchant dashboard', async ({ page }) => {
    await loginPage.login('merchant@example.com', 'merchant123456');
    await loginPage.expectLoginSuccess();
    await page.waitForURL(/.*merchant/, { timeout: 15000 });
    await expect(page).toHaveURL(/.*merchant/);
  });

  test('should login as admin and redirect to admin dashboard', async ({ page }) => {
    await loginPage.login('admin@example.com', 'admin123456');
    await loginPage.expectLoginSuccess();
    await page.waitForURL(/.*admin/, { timeout: 15000 });
    await expect(page).toHaveURL(/.*admin/);
  });

  test('should show error for invalid credentials', async ({ page }) => {
    await loginPage.login('invalid@example.com', 'wrongpassword');
    await page.waitForTimeout(2000);
    await expect(page).toHaveURL(/.*login/);
    await expect(page.locator('.auth-card')).toBeVisible();
  });

  test('should logout successfully', async ({ page }) => {
    await loginPage.login('demo@example.com', 'demo123456');
    await page.waitForURL(/.*\//, { timeout: 15000 });
    
    await page.locator('[data-testid="user-dropdown"]').click();
    await page.getByText('退出登录').click();
    
    await page.waitForURL(/.*login/, { timeout: 10000 });
    await expect(page.locator('.auth-card')).toBeVisible();
  });

  test('should persist login state after page refresh', async ({ page }) => {
    await loginPage.login('demo@example.com', 'demo123456');
    await page.waitForURL(/.*\//, { timeout: 15000 });

    // 刷新后 isAuthenticated 先有 token，user 需等 /users/me 拉取后才渲染下拉（见 Layout.tsx）
    const userMe = page.waitForResponse(
      (resp) => resp.url().includes('/users/me') && resp.ok(),
      { timeout: 20000 }
    );
    await page.reload();
    await userMe;
    await expect(page.locator('[data-testid="user-dropdown"]')).toBeVisible({ timeout: 10000 });
  });
});

test.describe('Registration', () => {
  let registerPage: RegisterPage;

  test.beforeEach(async ({ page }) => {
    registerPage = new RegisterPage(page);
    await registerPage.goto();
  });

  test('should have buyer and merchant registration tabs', async ({ page }) => {
    await page.waitForLoadState('networkidle');
    await expect(page.getByRole('tab', { name: /买家注册/i })).toBeVisible({ timeout: 10000 });
    await expect(page.getByRole('tab', { name: /商户入驻/i })).toBeVisible({ timeout: 10000 });
  });

  test('should register new user and redirect to home', async ({ page }) => {
    const uniqueEmail = `test${Date.now()}@example.com`;
    await registerPage.register(uniqueEmail, 'testuser', 'Test123456!');

    await page.waitForURL(/.*\//, { timeout: 15000 });
    await expect(page).toHaveURL(/.*\//);
  });

  test('should register as merchant and redirect to merchant dashboard', async ({ page }) => {
    const uniqueEmail = `merchant${Date.now()}@example.com`;
    await registerPage.register(uniqueEmail, 'merchant_user', 'Test123456!', 'merchant');
    
    await page.waitForURL(/.*merchant/, { timeout: 15000 });
    await expect(page).toHaveURL(/.*merchant/);
  });

  test('should show error for duplicate email', async ({ page }) => {
    await registerPage.register('demo@example.com', 'testuser', 'Test123456!');
    await page.waitForTimeout(2000);
    const errorMessage = page.locator('.ant-message');
    await expect(errorMessage.first()).toBeVisible({ timeout: 10000 });
  });

  test('should show error for password mismatch', async ({ page }) => {
    const uniqueEmail = `test${Date.now()}@example.com`;
    
    await page.getByPlaceholder('example@email.com').fill(uniqueEmail);
    await page.getByPlaceholder('输入你的名字').fill('testuser');
    await page.getByPlaceholder('设置密码').fill('Test123456!');
    await page.getByPlaceholder('再次输入密码').fill('DifferentPassword!');
    await page.locator('button[type="submit"]').click();
    
    await expect(page.getByText('两次输入的密码不一致')).toBeVisible({ timeout: 10000 });
  });
});

test.describe('Access Control', () => {
  test('should deny regular user access to merchant dashboard', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login('demo@example.com', 'demo123456');
    await page.waitForURL(/.*\//, { timeout: 15000 });
    
    await page.goto('/merchant');
    await page.waitForTimeout(2000);
    await expect(page).not.toHaveURL(/.*merchant/);
  });

  test('should deny regular user access to admin dashboard', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login('demo@example.com', 'demo123456');
    await page.waitForURL(/.*\//, { timeout: 15000 });
    
    await page.goto('/admin');
    await page.waitForTimeout(2000);
    await expect(page).not.toHaveURL(/.*admin/);
  });

  test('should deny merchant access to admin dashboard', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login('merchant@example.com', 'merchant123456');
    await page.waitForURL(/.*merchant/, { timeout: 15000 });
    
    await page.goto('/admin');
    await page.waitForTimeout(2000);
    await expect(page).not.toHaveURL(/.*admin/);
  });

  test('should display admin dashboard for admin user', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login('admin@example.com', 'admin123456');
    await page.waitForURL(/.*admin/, { timeout: 15000 });
    
    await expect(page.getByText('运营管理')).toBeVisible();
  });
});
