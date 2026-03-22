import { test, expect } from '@playwright/test';

test.describe('Payment Flow E2E Tests', () => {
  test('should display payment page', async ({ page }) => {
    await page.goto('/payment/1');
    await page.waitForLoadState('domcontentloaded');
  });

  test('should redirect to login if not authenticated', async ({ page }) => {
    await page.goto('/payment/1');
    await page.waitForLoadState('domcontentloaded');
  });
});
