import { test, expect, Page } from '@playwright/test';
import { LoginPage } from './pages';

async function createPendingOrder(page: Page): Promise<number> {
  const token = await page.evaluate(() => localStorage.getItem('auth_token') || sessionStorage.getItem('auth_token'));
  expect(token).toBeTruthy();

  const catalogResp = await page.request.get('/catalog', {
    params: { page: 1, per_page: 1 },
  });
  expect(catalogResp.ok()).toBeTruthy();
  const contentType = (catalogResp.headers()['content-type'] || '').toLowerCase();
  expect(contentType.includes('application/json')).toBeTruthy();
  const catalog = await catalogResp.json();
  const first = catalog?.data?.data?.[0];
  expect(first?.id).toBeTruthy();

  const orderResp = await page.request.post('/orders', {
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
    data: {
      items: [{ sku_id: first.id, quantity: 1 }],
    },
  });
  expect(orderResp.ok()).toBeTruthy();
  const created = await orderResp.json();
  return created?.data?.id;
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
    expect(orderID).toBeTruthy();

    await page.goto(`/payment/${orderID}`);
    await expect(page.getByText('订单支付')).toBeVisible();
    await expect(page.getByText('订单明细')).toBeVisible();

    await page.goto(`/orders/${orderID}`);
    await expect(page.getByText('订单详情')).toBeVisible();
    await expect(page.getByText('订单明细')).toBeVisible();
  });
});
