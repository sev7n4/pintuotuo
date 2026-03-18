import { test, expect } from '@playwright/test';

test.describe('Smoke Test', () => {
  test('should login and browse products', async ({ page }) => {
    // Mock login API
    await page.route('**/api/v1/users/login', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            user: { id: '1', email: 'test@example.com', name: 'Test User' },
            token: 'mock-token'
          }
        }),
      });
    });

    // Mock products API
    await page.route('**/api/v1/products*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: [
            { id: '1', name: 'Test Product 1', price: 100, originalPrice: 200, groupPrice: 80, stock: 10, imageUrl: '' },
            { id: '2', name: 'Test Product 2', price: 200, originalPrice: 300, groupPrice: 150, stock: 5, imageUrl: '' },
          ]
        }),
      });
    });

    // Go to login page
    await page.goto('/LoginPage');

    // Fill login form
    await page.fill('input[name="email"]', 'test@example.com');
    await page.fill('input[name="password"]', 'password');
    await page.click('button[type="submit"]');

    // Should redirect to products page
    await expect(page).toHaveURL(/\/products/);

    // Should see products
    await expect(page.getByText('Test Product 1')).toBeVisible();
    await expect(page.getByText('Test Product 2')).toBeVisible();
  });
});
