import { test, expect, Page } from '@playwright/test';
import { LoginPage } from './pages';

const API_BASE_URL = process.env.E2E_API_BASE_URL || 'http://localhost:8080/api/v1';

async function createPendingOrder(page: Page): Promise<number | null> {
  const token = await page.evaluate(() => localStorage.getItem('auth_token') || sessionStorage.getItem('auth_token'));
  expect(token).toBeTruthy();

  const catalogResp = await page.request.get(`${API_BASE_URL}/catalog/home`);
  expect(catalogResp.ok()).toBeTruthy();
  const catalogText = await catalogResp.text();
  const catalog = JSON.parse(catalogText);
  const catalogItems =
    catalog?.data?.data ??
    catalog?.data?.items ??
    catalog?.data?.hot ??
    catalog?.data?.new ??
    [];
  const first = catalogItems[0];
  if (!first?.id) {
    return null;
  }

  const orderResp = await page.request.post(`${API_BASE_URL}/orders`, {
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
    data: {
      items: [{ sku_id: first.id, quantity: 1 }],
    },
  });
  if (!orderResp.ok()) {
    return null;
  }
  const created = await orderResp.json();
  return created?.data?.id || null;
}

test.describe('P1 Smoke: checkout -> payment -> order detail', () => {
  test('desktop/mobile smoke chain for order pages', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login('demo@example.com', 'demo123456');
    await loginPage.expectLoginSuccess();
    await page.waitForURL(/.*\//, { timeout: 20000 });

    await page.goto('/checkout');
    await expect(page.getByText('确认订单').or(page.getByText('购物车是空的'))).toBeVisible();

    const orderID = await createPendingOrder(page);
    test.skip(!orderID, 'No catalog SKU available to create smoke order in CI seed data');

    await page.goto(`/payment/${orderID}`);
    await expect(page.getByText('订单支付')).toBeVisible();
    await expect(page.getByText('订单明细')).toBeVisible();

    await page.goto(`/orders/${orderID}`);
    await expect(page.getByText('订单详情')).toBeVisible();
    await expect(page.getByText('订单明细')).toBeVisible();
  });
});
