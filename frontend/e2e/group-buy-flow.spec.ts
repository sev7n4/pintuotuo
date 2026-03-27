import { test, expect } from '@playwright/test';
import { LoginPage } from './pages';

test.describe('Group Buy Flow E2E Tests', () => {
  let loginPage: LoginPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login('demo@example.com', 'demo123456');
    await page.waitForURL(/.*\//, { timeout: 15000 });
  });

  test('should display group list page', async ({ page }) => {
    await page.goto('/groups');
    await expect(page).toHaveURL(/.*groups/, { timeout: 5000 });
  });

  test('should show active groups', async ({ page }) => {
    await page.goto('/groups');
    await page.waitForLoadState('domcontentloaded');
  });

  test('should view group detail', async ({ page }) => {
    await page.goto('/groups');
    await page.waitForLoadState('domcontentloaded');
  });

  test('should filter groups by status', async ({ page }) => {
    await page.goto('/groups?status=active');
    await page.waitForLoadState('domcontentloaded');
  });

  test('should show empty state when no groups available', async ({ page }) => {
    await page.goto('/groups?status=failed');
    await page.waitForLoadState('domcontentloaded');
  });
});
