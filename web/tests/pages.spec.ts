import { test, expect } from '@playwright/test';
import { createTestUser, loginAs } from '../test-utils/helpers';

test.describe('Pages', () => {
  test.beforeEach(async ({ page }) => {
    const { email, password } = await createTestUser();
    await loginAs(page, email, password);
  });

  test('create page via sidebar button opens editor', async ({ page }) => {
    await page.locator('button', { hasText: 'Create your first page' }).click();
    await page.locator('[contenteditable="true"]').first().waitFor({ timeout: 15000 });
    await expect(page.locator('input[aria-label="Page title"]')).toBeVisible();
  });

  test('create page via sidebar "New Page" button', async ({ page }) => {
    await page.locator('button', { hasText: 'Create your first page' }).click();
    await page.locator('[contenteditable="true"]').first().waitFor({ timeout: 15000 });
    const firstUrl = page.url();

    await page.locator('button', { hasText: 'New Page' }).click();
    await page.waitForFunction(url => window.location.href !== url, firstUrl, { timeout: 15000 });
    const items = page.locator('[data-page-id]');
    await expect(items).toHaveCount(2);
  });

  test('clicking a page in sidebar opens it', async ({ page }) => {
    await page.locator('button', { hasText: 'Create your first page' }).click();
    await page.locator('[contenteditable="true"]').first().waitFor({ timeout: 15000 });
    const firstUrl = page.url();

    await page.locator('button', { hasText: 'New Page' }).click();
    await page.waitForFunction(url => window.location.href !== url, firstUrl, { timeout: 15000 });
    const secondUrl = page.url();
    expect(firstUrl).not.toBe(secondUrl);

    await page.locator('[data-page-id]').first().click();
    await expect(page).toHaveURL(firstUrl);
  });

  test('delete page via sidebar context button', async ({ page }) => {
    await page.locator('button', { hasText: 'Create your first page' }).click();
    await page.locator('[contenteditable="true"]').first().waitFor({ timeout: 15000 });

    const sidebarItem = page.locator('[data-page-id]').first();
    await sidebarItem.hover();
    await sidebarItem.locator('button[title="Delete page"]').click();
    // Confirm in custom dialog
    await page.locator('[role="dialog"] button:has-text("Delete")').click();

    await expect(page).toHaveURL('/');
  });
});
