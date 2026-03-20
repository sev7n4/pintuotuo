import { Page, expect } from '@playwright/test';

export class LoginPage {
  constructor(private page: Page) {}

  async goto() {
    await this.page.goto('/login');
    await this.page.waitForSelector('text=拼脱脱 - 登录');
  }

  async login(email: string, password: string) {
    await this.page.getByPlaceholder('example@email.com').fill(email);
    await this.page.getByPlaceholder('输入密码').fill(password);
    const submitButton = this.page.getByRole('button', { name: '登录' });
    await submitButton.waitFor({ state: 'visible' });
    await submitButton.click();
  }

  async expectLoginSuccess() {
    await expect(this.page.getByText('登录成功')).toBeVisible();
  }

  async expectLoginError() {
    await expect(this.page.locator('.ant-message-error, text=登录失败')).toBeVisible();
  }
}

export class RegisterPage {
  constructor(private page: Page) {}

  async goto() {
    await this.page.goto('/register');
    await this.page.waitForSelector('text=拼脱脱 - 注册');
  }

  async register(email: string, name: string, password: string, role: 'user' | 'merchant' = 'user') {
    if (role === 'merchant') {
      await this.page.getByText('商家', { exact: true }).click();
    }
    await this.page.getByPlaceholder('example@email.com').fill(email);
    await this.page.getByPlaceholder('输入你的名字').fill(name);
    await this.page.getByPlaceholder('设置密码').fill(password);
    await this.page.getByPlaceholder('再次输入密码').fill(password);
    const submitButton = this.page.getByRole('button', { name: '创建账户' });
    await submitButton.waitFor({ state: 'visible' });
    await submitButton.click();
  }
}
