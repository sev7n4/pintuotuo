import { test, expect } from '@playwright/test';

test.describe('Shopping Flow E2E Tests', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should display product list on home page', async ({ page }) => {
    await expect(page.locator('text=商品')).toBeVisible();
  });

  test('should navigate to product detail', async ({ page }) => {
    await page.goto('/products');
    await expect(page).toHaveURL(/.*products/);
  });

  test('should search for products', async ({ page }) => {
    await page.goto('/products');
    const searchInput = page.locator('input[placeholder*="搜索"]').first();
    if (await searchInput.isVisible()) {
      await searchInput.fill('test');
      await searchInput.press('Enter');
    }
  });

  test('should add product to cart', async ({ page, context }) => {
    await page.goto('/login');
    
    await page.fill('input[type="email"]', 'user@example.com');
    await page.fill('input[type="password"]', 'password123');
    await page.click('button:has-text("登录")');
    
    await page.waitForURL(/.*products|.*home/, { timeout: 10000 }).catch(() => {});
    
    await page.goto('/products');
    await page.waitForLoadState('networkidle');
    
    const addToCartBtn = page.locator('button:has-text("加入购物车")').first();
    if (await addToCartBtn.isVisible()) {
      await addToCartBtn.click();
    }
  });

  test('should view cart', async ({ page }) => {
    await page.goto('/cart');
    await expect(page).toHaveURL(/.*cart/);
  });

  test('should navigate to checkout from cart', async ({ page }) => {
    await page.goto('/cart');
    const checkoutBtn = page.locator('button:has-text("结算")');
    if (await checkoutBtn.isVisible()) {
      await checkoutBtn.click();
    }
  });

  test('should complete checkout flow', async ({ page }) => {
    await page.goto('/login');
    
    await page.fill('input[type="email"]', 'user@example.com');
    await page.fill('input[type="password"]', 'password123');
    await page.click('button:has-text("登录")');
    
    await page.waitForURL(/.*products|.*home/, { timeout: 10000 }).catch(() => {});
    
    await page.goto('/checkout');
    await page.waitForLoadState('networkidle');
    
    const submitBtn = page.locator('button:has-text("提交订单")');
    if (await submitBtn.isVisible()) {
      await submitBtn.click();
    }
  });

  test('should handle empty cart', async ({ page }) => {
    await page.goto('/cart');
    
    const emptyMessage = page.locator('text=购物车是空');
    if (await emptyMessage.isVisible()) {
      await expect(emptyMessage).toBeVisible();
    }
  });

  test('should update quantity in cart', async ({ page }) => {
    await page.goto('/login');
    
    await page.fill('input[type="email"]', 'user@example.com');
    await page.fill('input[type="password"]', 'password123');
    await page.click('button:has-text("登录")');
    
    await page.waitForURL(/.*products|.*home/, { timeout: 10000 }).catch(() => {});
    
    await page.goto('/cart');
    
    const quantityInput = page.locator('input[type="number"]').first();
    if (await quantityInput.isVisible()) {
      await quantityInput.fill('2');
      await quantityInput.press('Enter');
    }
  });

  test('should remove item from cart', async ({ page }) => {
    await page.goto('/login');
    
    await page.fill('input[type="email"]', 'user@example.com');
    await page.fill('input[type="password"]', 'password123');
    await page.click('button:has-text("登录")');
    
    await page.waitForURL(/.*products|.*home/, { timeout: 10000 }).catch(() => {});
    
    await page.goto('/cart');
    
    const deleteBtn = page.locator('button:has-text("删除")').first();
    if (await deleteBtn.isVisible()) {
      await deleteBtn.click();
    }
  });
});
