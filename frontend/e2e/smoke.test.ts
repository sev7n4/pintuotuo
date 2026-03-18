import { test, expect } from '@playwright/test';

test.describe('Smoke Test', () => {
  test('should login and browse products', async ({ page }) => {
    // Mock login API
    await page.route('**/api/v1/users/login', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          code: 200,
          message: 'success',
          data: {
            user: { id: 1, email: 'test@example.com', name: 'Test User' },
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
          code: 200,
          message: 'success',
          data: {
            total: 2,
            page: 1,
            per_page: 20,
            data: [
              { id: 1, name: 'Test Product 1', price: 100, originalPrice: 200, groupPrice: 80, stock: 10, imageUrl: '', status: 'active' },
              { id: 2, name: 'Test Product 2', price: 200, originalPrice: 300, groupPrice: 150, stock: 5, imageUrl: '', status: 'active' },
            ]
          }
        }),
      });
    });

    // Go to login page
    await page.goto('/login');

    // Fill login form
    // Ant Design inputs are best targeted by their labels or placeholders
    await page.getByLabel('邮箱').fill('test@example.com');
    await page.getByLabel('密码').fill('password');
    await page.locator('button[type="submit"]').click();

    // Wait for the products API response to ensure the page is loaded
    await page.waitForResponse('**/api/v1/products*');

    // Should redirect to products page
    await expect(page).toHaveURL(/\/products/);

    // Should see products
    await expect(page.getByText('Test Product 1')).toBeVisible();
    await expect(page.getByText('Test Product 2')).toBeVisible();
  });
});
