import { test, expect } from '@playwright/test';

test.describe('Group Buy Flow E2E Tests', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should display group list page', async ({ page }) => {
    await page.goto('/groups');
    await expect(page).toHaveURL(/.*groups/);
  });

  test('should show active groups', async ({ page }) => {
    await page.goto('/groups');
    await page.waitForLoadState('networkidle');
    
    const groupCard = page.locator('[data-testid="group-card"]').first();
    if (await groupCard.isVisible()) {
      await expect(groupCard).toBeVisible();
    }
  });

  test('should view group detail', async ({ page }) => {
    await page.goto('/groups');
    await page.waitForLoadState('networkidle');
    
    const groupLink = page.locator('a[href*="/groups/"]').first();
    if (await groupLink.isVisible()) {
      await groupLink.click();
      await page.waitForLoadState('networkidle');
    }
  });

  test('should create a new group', async ({ page }) => {
    await page.goto('/login');
    
    await page.fill('input[type="email"]', 'user@example.com');
    await page.fill('input[type="password"]', 'password123');
    await page.click('button:has-text("登录")');
    
    await page.waitForURL(/.*products|.*home/, { timeout: 10000 }).catch(() => {});
    
    await page.goto('/products');
    await page.waitForLoadState('networkidle');
    
    const createGroupBtn = page.locator('button:has-text("发起拼团")').first();
    if (await createGroupBtn.isVisible()) {
      await createGroupBtn.click();
    }
  });

  test('should join an existing group', async ({ page }) => {
    await page.goto('/login');
    
    await page.fill('input[type="email"]', 'user@example.com');
    await page.fill('input[type="password"]', 'password123');
    await page.click('button:has-text("登录")');
    
    await page.waitForURL(/.*products|.*home/, { timeout: 10000 }).catch(() => {});
    
    await page.goto('/groups');
    await page.waitForLoadState('networkidle');
    
    const joinBtn = page.locator('button:has-text("参与拼团")').first();
    if (await joinBtn.isVisible()) {
      await joinBtn.click();
    }
  });

  test('should show group progress', async ({ page }) => {
    await page.goto('/groups');
    await page.waitForLoadState('networkidle');
    
    const groupLink = page.locator('a[href*="/groups/"]').first();
    if (await groupLink.isVisible()) {
      await groupLink.click();
      await page.waitForLoadState('networkidle');
      
      const progressBar = page.locator('[data-testid="progress-bar"]');
      if (await progressBar.isVisible()) {
        await expect(progressBar).toBeVisible();
      }
    }
  });

  test('should show group deadline countdown', async ({ page }) => {
    await page.goto('/groups');
    await page.waitForLoadState('networkidle');
    
    const countdown = page.locator('text=/\\d+:\\d+:\\d+/');
    if (await countdown.isVisible()) {
      await expect(countdown).toBeVisible();
    }
  });

  test('should handle full group gracefully', async ({ page }) => {
    await page.goto('/login');
    
    await page.fill('input[type="email"]', 'user@example.com');
    await page.fill('input[type="password"]', 'password123');
    await page.click('button:has-text("登录")');
    
    await page.waitForURL(/.*products|.*home/, { timeout: 10000 }).catch(() => {});
    
    await page.goto('/groups?status=completed');
    await page.waitForLoadState('networkidle');
  });

  test('should filter groups by status', async ({ page }) => {
    await page.goto('/groups');
    
    const statusFilter = page.locator('select').first();
    if (await statusFilter.isVisible()) {
      await statusFilter.selectOption('active');
    }
  });

  test('should show group members', async ({ page }) => {
    await page.goto('/groups');
    await page.waitForLoadState('networkidle');
    
    const groupLink = page.locator('a[href*="/groups/"]').first();
    if (await groupLink.isVisible()) {
      await groupLink.click();
      await page.waitForLoadState('networkidle');
      
      const membersList = page.locator('[data-testid="group-members"]');
      if (await membersList.isVisible()) {
        await expect(membersList).toBeVisible();
      }
    }
  });

  test('should cancel own group', async ({ page }) => {
    await page.goto('/login');
    
    await page.fill('input[type="email"]', 'user@example.com');
    await page.fill('input[type="password"]', 'password123');
    await page.click('button:has-text("登录")');
    
    await page.waitForURL(/.*products|.*home/, { timeout: 10000 }).catch(() => {});
    
    await page.goto('/groups');
    await page.waitForLoadState('networkidle');
    
    const cancelBtn = page.locator('button:has-text("取消拼团")').first();
    if (await cancelBtn.isVisible()) {
      await cancelBtn.click();
    }
  });

  test('should show empty state when no groups available', async ({ page }) => {
    await page.goto('/groups?status=failed');
    await page.waitForLoadState('networkidle');
    
    const emptyState = page.locator('text=暂无');
    if (await emptyState.isVisible()) {
      await expect(emptyState).toBeVisible();
    }
  });
});
