import { test, expect } from '@playwright/test';
import { LoginPage } from './pages';

test.describe('Developer center', () => {
  test('redirects unauthenticated user to login with redirect param', async ({ page }) => {
    await page.goto('/developer/quickstart');
    await expect(page).toHaveURL(/\/login\?redirect=/);
  });

  test('shows developer sider after login as demo user', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login('demo@example.com', 'demo123456');
    await loginPage.expectLoginSuccess();
    await page.goto('/developer/quickstart');
    await expect(page.getByTestId('developer-sider')).toBeVisible();
    await expect(page.getByRole('heading', { name: '快速开始' })).toBeVisible();
  });

  test('redirect query returns to developer center after login', async ({ page }) => {
    await page.goto('/developer/models');
    await expect(page).toHaveURL(/\/login\?redirect=/);
    const loginPage = new LoginPage(page);
    await loginPage.login('demo@example.com', 'demo123456');
    await loginPage.expectLoginSuccess();
    await page.waitForURL(/\/developer\/models/, { timeout: 15000 });
    await expect(page.getByRole('heading', { name: '模型与权益' })).toBeVisible();
  });
});
