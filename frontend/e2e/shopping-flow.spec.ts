import { test, expect } from '@playwright/test';

test.describe('Shopping Flow E2E Tests', () => {
  test('should display product list on home page', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('domcontentloaded');
  });

  test('should navigate to product list', async ({ page }) => {
    await page.goto('/products');
    await expect(page).toHaveURL(/.*products/, { timeout: 5000 });
  });

  test('should view cart page', async ({ page }) => {
    await page.goto('/cart');
    await expect(page).toHaveURL(/.*cart/, { timeout: 5000 });
  });

  test('should view checkout page', async ({ page }) => {
    await page.goto('/checkout');
    await page.waitForLoadState('domcontentloaded');
  });
});
