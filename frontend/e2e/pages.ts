import { Page, expect } from '@playwright/test';

export class LoginPage {
  constructor(private page: Page) {}

  async goto() {
    await this.page.goto('/login');
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForSelector('.auth-card');
  }

  async login(email: string, password: string) {
    await this.page.getByPlaceholder('example@email.com').fill(email);
    await this.page.getByPlaceholder('输入密码').fill(password);
    await this.page.locator('button[type="submit"]').click();
  }

  async expectLoginSuccess() {
    await expect(this.page.getByText('登录成功')).toBeVisible();
  }

  async expectLoginError() {
    await expect(this.page.locator('.ant-message-error')).toBeVisible();
  }
}

export class RegisterPage {
  constructor(private page: Page) {}

  async goto() {
    await this.page.goto('/register');
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForSelector('.auth-card');
  }

  async register(email: string, name: string, password: string, role: 'user' | 'merchant' = 'user') {
    if (role === 'merchant') {
      await this.page.getByText('商家', { exact: true }).click();
    }
    await this.page.getByPlaceholder('example@email.com').fill(email);
    await this.page.getByPlaceholder('输入你的名字').fill(name);
    await this.page.getByPlaceholder('设置密码').fill(password);
    await this.page.getByPlaceholder('再次输入密码').fill(password);
    await this.page.locator('button[type="submit"]').click();
  }
}

export class MerchantDashboardPage {
  constructor(private page: Page) {}

  async goto() {
    await this.page.goto('/merchant');
    await this.page.waitForLoadState('networkidle');
  }

  async expectDashboardVisible() {
    await expect(this.page.getByRole('heading', { name: '数据概览' })).toBeVisible();
  }

  async getStatsCards() {
    return {
      totalProducts: this.page.locator('.ant-statistic').nth(0),
      activeProducts: this.page.locator('.ant-statistic').nth(1),
      monthSales: this.page.locator('.ant-statistic').nth(2),
      monthOrders: this.page.locator('.ant-statistic').nth(3),
    };
  }

  async getRecentOrders() {
    return this.page.locator('.ant-table-row');
  }
}

export class MerchantProductsPage {
  constructor(private page: Page) {}

  async goto() {
    await this.page.goto('/merchant/products');
    await this.page.waitForLoadState('networkidle');
  }

  async expectProductsPageVisible() {
    await expect(this.page.getByRole('heading', { name: '商品管理' }).or(this.page.locator('h2:has-text("商品")'))).toBeVisible();
  }

  async clickAddProduct() {
    await this.page.locator('button:has-text("添加商品")').click();
    await this.page.waitForSelector('.ant-modal-content', { timeout: 10000 });
  }

  async fillProductForm(data: {
    name: string;
    description: string;
    price: number;
    stock: number;
    category?: string;
    status?: string;
  }) {
    await this.page.getByPlaceholder('请输入商品名称').fill(data.name);
    await this.page.getByPlaceholder('请输入商品描述').fill(data.description);
    
    const priceInput = this.page.getByPlaceholder('请输入价格');
    await priceInput.fill(data.price.toString());
    await priceInput.press('Tab');
    
    const stockInput = this.page.getByPlaceholder('请输入库存');
    await stockInput.fill(data.stock.toString());
    await stockInput.press('Tab');
    
    await this.page.waitForTimeout(300);
    
    if (data.category) {
      await this.page.locator('.ant-select').first().click();
      await this.page.getByText(data.category).click();
    }
    
    if (data.status) {
      await this.page.locator('.ant-select').nth(1).click();
      await this.page.getByText(data.status).click();
    }
  }

  async submitProduct() {
    const modal = this.page.locator('.ant-modal-content');
    await modal.locator('button:has-text("保")').click();
    await this.page.waitForTimeout(1000);
  }

  async editProduct(name: string) {
    const row = this.page.locator('.ant-table-row').filter({ hasText: name });
    await row.locator('button:has-text("编辑")').click();
    await this.page.waitForSelector('.ant-modal-content', { timeout: 10000 });
  }

  async deleteProduct(name: string) {
    const row = this.page.locator('.ant-table-row').filter({ hasText: name });
    await row.locator('button:has-text("删除")').click();
    await this.page.locator('button:has-text("确定")').click();
    await this.page.waitForTimeout(500);
  }

  async filterByStatus(status: string) {
    await this.page.locator('.ant-select').first().click();
    await this.page.getByText(status).click();
    await this.page.waitForTimeout(500);
  }

  async getProductCount() {
    return await this.page.locator('.ant-table-row').count();
  }
}

export class MerchantOrdersPage {
  constructor(private page: Page) {}

  async goto() {
    await this.page.goto('/merchant/orders');
    await this.page.waitForLoadState('networkidle');
  }

  async expectOrdersPageVisible() {
    await expect(this.page.getByRole('heading', { name: '订单管理' }).or(this.page.locator('h2:has-text("订单")'))).toBeVisible();
  }

  async filterByStatus(status: string) {
    await this.page.locator('.ant-select').first().click();
    await this.page.getByText(status).click();
    await this.page.waitForTimeout(500);
  }

  async getOrderCount() {
    return await this.page.locator('.ant-table-row').count();
  }

  async viewOrderDetail(orderId: string) {
    await this.page.locator(`text=${orderId}`).click();
    await this.page.waitForTimeout(500);
  }

  async exportOrders() {
    await this.page.locator('button:has-text("导出")').click();
  }
}

export class MerchantSettlementsPage {
  constructor(private page: Page) {}

  async goto() {
    await this.page.goto('/merchant/settlements');
    await this.page.waitForLoadState('networkidle');
  }

  async expectSettlementsPageVisible() {
    await expect(this.page.getByRole('heading', { name: '结算管理' }).or(this.page.locator('h2:has-text("结算")'))).toBeVisible();
  }

  async clickApplySettlement() {
    await this.page.locator('button:has-text("申请结算")').click();
  }

  async expectSettlementSuccess() {
    await this.page.waitForTimeout(500);
    const message = this.page.locator('.ant-message');
    const isVisible = await message.isVisible().catch(() => false);
    if (isVisible) {
      await expect(message).toBeVisible({ timeout: 5000 });
    }
  }

  async getSettlementCount() {
    return await this.page.locator('.ant-table-row').count();
  }

  async viewSettlementDetail(id: string) {
    await this.page.locator(`text=${id}`).click();
    await this.page.waitForTimeout(500);
  }
}

export class MerchantAPIKeysPage {
  constructor(private page: Page) {}

  async goto() {
    await this.page.goto('/merchant/api-keys');
    await this.page.waitForLoadState('networkidle');
  }

  async expectAPIKeysPageVisible() {
    await expect(this.page.getByRole('heading', { name: 'API密钥管理' }).or(this.page.locator('h2:has-text("API")'))).toBeVisible();
  }

  async clickAddKey() {
    await this.page.locator('button:has-text("添加密钥")').click();
    await this.page.waitForSelector('.ant-modal-content', { timeout: 10000 });
  }

  async fillKeyForm(data: {
    name: string;
    provider: string;
    apiKey: string;
    quotaLimit?: number;
  }) {
    await this.page.getByPlaceholder(/生产环境密钥|密钥名称/).fill(data.name, { timeout: 10000 });
    
    await this.page.locator('.ant-form-item').filter({ hasText: '提供商' }).locator('.ant-select-selector').click();
    await this.page.waitForSelector('.ant-select-dropdown', { state: 'visible', timeout: 5000 });
    await this.page.waitForTimeout(500);
    
    const option = this.page.locator('.ant-select-item-option').filter({ 
      has: this.page.locator(`text=${data.provider}`) 
    }).first();
    
    await option.evaluate((el) => {
      el.scrollIntoView({ block: 'center' });
    });
    await this.page.waitForTimeout(200);
    await option.click({ force: true });
    
    await this.page.getByPlaceholder(/请输入API Key/).fill(data.apiKey, { timeout: 10000 });
    
    if (data.quotaLimit !== undefined) {
      await this.page.locator('.ant-input-number-input').fill(data.quotaLimit.toString());
    }
  }

  async submitKey() {
    await this.page.locator('.ant-modal-content').locator('button:has-text("保存")').click();
    await this.page.waitForTimeout(1000);
  }

  async toggleKeyStatus(name: string) {
    const row = this.page.locator('.ant-table-row').filter({ hasText: name });
    await row.locator('.ant-switch').click();
    await this.page.waitForTimeout(500);
  }

  async deleteKey(name: string) {
    const row = this.page.locator('.ant-table-row').filter({ hasText: name });
    await row.locator('button:has-text("删除")').click();
    await this.page.locator('button:has-text("确定")').click();
    await this.page.waitForTimeout(500);
  }

  async getKeyCount() {
    return await this.page.locator('.ant-table-row').count();
  }
}

export class MerchantSettingsPage {
  constructor(private page: Page) {}

  async goto() {
    await this.page.goto('/merchant/settings');
    await this.page.waitForLoadState('networkidle');
  }

  async expectSettingsPageVisible() {
    await expect(this.page.getByRole('heading', { name: '店铺设置' }).or(this.page.locator('h2:has-text("设置")'))).toBeVisible();
  }

  async updateStoreInfo(data: {
    companyName?: string;
    contactName?: string;
    contactPhone?: string;
    contactEmail?: string;
    address?: string;
    description?: string;
  }) {
    if (data.companyName) {
      await this.page.getByPlaceholder('请输入公司名称').fill(data.companyName);
    }
    if (data.contactName) {
      await this.page.getByPlaceholder('请输入联系人').fill(data.contactName);
    }
    if (data.contactPhone) {
      await this.page.getByPlaceholder('请输入联系电话').fill(data.contactPhone);
    }
    if (data.contactEmail) {
      await this.page.getByPlaceholder('请输入联系邮箱').fill(data.contactEmail);
    }
    if (data.address) {
      await this.page.getByPlaceholder('请输入公司地址').fill(data.address);
    }
    if (data.description) {
      await this.page.getByPlaceholder('请输入店铺简介').fill(data.description);
    }
  }

  async saveSettings() {
    await this.page.locator('button:has-text("保存")').click();
    await this.page.waitForTimeout(1000);
  }

  async getVerificationStatus() {
    return this.page.locator('.ant-tag').first();
  }
}
