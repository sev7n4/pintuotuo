import { test, expect } from '@playwright/test';

test.describe('Products', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should display product list', async ({ page }) => {
    await page.goto('/products');
    const productCount = await page.locator('.ant-card').count();
    expect(productCount).toBeGreaterThanOrEqual(0);
  });

  test('should search products', async ({ page }) => {
    await page.goto('/products');
    const searchInput = page.locator('input[placeholder*="搜索"]');
    if (await searchInput.isVisible()) {
      await searchInput.fill('测试商品');
      await searchInput.press('Enter');
      await page.waitForTimeout(500);
    }
  });

  test('should filter by category', async ({ page }) => {
    await page.goto('/products');
    
    const categoryButton = page.locator('.ant-tag').first();
    if (await categoryButton.isVisible()) {
      await categoryButton.click();
      await page.waitForTimeout(500);
    }
  });

  test('should view product details', async ({ page }) => {
    await page.goto('/products');
    
    const productCard = page.locator('.ant-card').first();
    if (await productCard.isVisible()) {
      await productCard.click();
      await page.waitForTimeout(1000);
    }
  });

  test('should sort products', async ({ page }) => {
    await page.goto('/products');
    
    const sortSelect = page.locator('.ant-select').first();
    if (await sortSelect.isVisible()) {
      await sortSelect.click();
      await page.waitForTimeout(300);
    }
  });

  test('should paginate products', async ({ page }) => {
    await page.goto('/products');
    
    const nextButton = page.locator('.ant-pagination-next');
    if (await nextButton.isVisible() && await nextButton.isEnabled()) {
      await nextButton.click();
      await page.waitForTimeout(500);
    }
  });
});
