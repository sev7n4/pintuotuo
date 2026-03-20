import { test, expect, Page } from '@playwright/test';

class LoginPage {
  constructor(private page: Page) {}

  async goto() {
    await this.page.goto('/login');
  }

  async login(email: string, password: string) {
    await this.page.fill('input[placeholder="example@email.com"]', email);
    await this.page.fill('input[placeholder="输入密码"]', password);
    await this.page.click('button:has-text("登录")');
  }

  async expectLoginSuccess() {
    await expect(this.page.locator('text=登录成功')).toBeVisible();
  }

  async expectLoginError() {
    await expect(this.page.locator('.ant-message-error, text=登录失败')).toBeVisible();
  }
}

class RegisterPage {
  constructor(private page: Page) {}

  async goto() {
    await this.page.goto('/register');
  }

  async register(email: string, name: string, password: string, role: 'user' | 'merchant' = 'user') {
    if (role === 'merchant') {
      await this.page.click('text=商家');
    }
    await this.page.fill('input[placeholder="example@email.com"]', email);
    await this.page.fill('input[placeholder="输入你的名字"]', name);
    await this.page.fill('input[placeholder="设置密码"]', password);
    await this.page.fill('input[placeholder="再次输入密码"]', password);
    await this.page.click('button:has-text("创建账户")');
  }
}

test.describe('Authentication', () => {
  test('should display login page', async ({ page }) => {
    await page.goto('/login');
    await expect(page.locator('text=拼脱脱 - 登录')).toBeVisible();
  });

  test('should show validation error for empty fields', async ({ page }) => {
    await page.goto('/login');
    await page.click('button:has-text("登录")');
    await expect(page.locator('text=请输入邮箱')).toBeVisible();
  });

  test('should navigate between login and register', async ({ page }) => {
    await page.goto('/login');
    await page.click('text=创建新账户');
    await expect(page.locator('text=拼脱脱 - 注册')).toBeVisible();
    
    await page.click('text=立即登录');
    await expect(page.locator('text=拼脱脱 - 登录')).toBeVisible();
  });

  test('should show error for invalid email format', async ({ page }) => {
    await page.goto('/login');
    await page.fill('input[placeholder="example@email.com"]', 'invalid-email');
    await page.fill('input[placeholder="输入密码"]', 'somepassword');
    await page.click('button:has-text("登录")');
    await expect(page.locator('text=邮箱格式不正确')).toBeVisible();
  });
});

test.describe('Login Flow', () => {
  let loginPage: LoginPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
    await loginPage.goto();
  });

  test('should login as regular user and redirect to products', async ({ page }) => {
    await loginPage.login('demo@example.com', 'demo123456');
    await loginPage.expectLoginSuccess();
    await expect(page).toHaveURL(/.*products/);
  });

  test('should login as merchant and redirect to merchant dashboard', async ({ page }) => {
    await loginPage.login('merchant@example.com', 'merchant123456');
    await loginPage.expectLoginSuccess();
    await expect(page).toHaveURL(/.*merchant/);
  });

  test('should login as admin and redirect to admin dashboard', async ({ page }) => {
    await loginPage.login('admin@example.com', 'admin123456');
    await loginPage.expectLoginSuccess();
    await expect(page).toHaveURL(/.*admin/);
  });

  test('should show error for invalid credentials', async ({ page }) => {
    await loginPage.login('invalid@example.com', 'wrongpassword');
    await loginPage.expectLoginError();
  });

  test('should logout successfully', async ({ page }) => {
    await loginPage.login('demo@example.com', 'demo123456');
    await expect(page).toHaveURL(/.*products/);
    
    await page.click('[data-testid="user-dropdown"]');
    await page.click('text=退出登录');
    await expect(page.locator('text=登录')).toBeVisible();
  });

  test('should persist login state after page refresh', async ({ page }) => {
    await loginPage.login('demo@example.com', 'demo123456');
    await expect(page).toHaveURL(/.*products/);
    
    await page.reload();
    await expect(page.locator('[data-testid="user-dropdown"]')).toBeVisible();
  });
});

test.describe('Registration', () => {
  let registerPage: RegisterPage;

  test.beforeEach(async ({ page }) => {
    registerPage = new RegisterPage(page);
    await registerPage.goto();
  });

  test('should have role selection on register page', async ({ page }) => {
    await expect(page.locator('text=用户')).toBeVisible();
    await expect(page.locator('text=商家')).toBeVisible();
  });

  test('should register new user and redirect to products', async ({ page }) => {
    const uniqueEmail = `test${Date.now()}@example.com`;
    await registerPage.register(uniqueEmail, 'testuser', 'Test123456!');
    
    await expect(page.locator('text=注册成功')).toBeVisible();
    await expect(page).toHaveURL(/.*products/);
  });

  test('should register as merchant and redirect to merchant dashboard', async ({ page }) => {
    const uniqueEmail = `merchant${Date.now()}@example.com`;
    await registerPage.register(uniqueEmail, 'merchant_user', 'Test123456!', 'merchant');
    
    await expect(page.locator('text=注册成功')).toBeVisible();
    await expect(page).toHaveURL(/.*merchant/);
  });

  test('should show error for duplicate email', async ({ page }) => {
    await registerPage.register('demo@example.com', 'testuser', 'Test123456!');
    await expect(page.locator('text=注册失败')).toBeVisible();
  });

  test('should show error for password mismatch', async ({ page }) => {
    const uniqueEmail = `test${Date.now()}@example.com`;
    
    await page.fill('input[placeholder="example@email.com"]', uniqueEmail);
    await page.fill('input[placeholder="输入你的名字"]', 'testuser');
    await page.fill('input[placeholder="设置密码"]', 'Test123456!');
    await page.fill('input[placeholder="再次输入密码"]', 'DifferentPassword!');
    await page.click('button:has-text("创建账户")');
    
    await expect(page.locator('text=两次输入的密码不一致')).toBeVisible();
  });
});

test.describe('Access Control', () => {
  test('should deny regular user access to merchant dashboard', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login('demo@example.com', 'demo123456');
    
    await page.goto('/merchant/dashboard');
    await expect(page).not.toHaveURL(/.*merchant/);
  });

  test('should deny regular user access to admin dashboard', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login('demo@example.com', 'demo123456');
    
    await page.goto('/admin');
    await expect(page).not.toHaveURL(/.*admin/);
  });

  test('should deny merchant access to admin dashboard', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login('merchant@example.com', 'merchant123456');
    
    await page.goto('/admin');
    await expect(page).not.toHaveURL(/.*admin/);
  });

  test('should display admin dashboard for admin user', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login('admin@example.com', 'admin123456');
    
    await page.goto('/admin');
    await expect(page.locator('text=运营管理')).toBeVisible();
  });
});
