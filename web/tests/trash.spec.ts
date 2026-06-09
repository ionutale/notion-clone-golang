import { test, expect } from '@playwright/test';
import { createTestUser, loginAs } from '../test-utils/helpers';

test.describe('Trash', () => {
  test.beforeEach(async ({ page }) => {
    const { email, password } = await createTestUser();
    await loginAs(page, email, password);
  });

  test('deleted page appears in trash', async ({ page }) => {
    await page.locator('button', { hasText: 'Create your first page' }).click();
    await page.locator('[contenteditable="true"]').first().waitFor({ timeout: 15000 });

    const sidebarItem = page.locator('[data-page-id]').first();
    await sidebarItem.hover();
    await sidebarItem.locator('button[title="Delete page"]').click();
    // Click confirm in custom dialog
    await page.locator('[role="dialog"] button:has-text("Delete")').click();

    await page.locator('a[href="/trash"]').click();
    await expect(page.getByText('Untitled')).toBeVisible({ timeout: 15000 });
  });

  test('restore brings page back from trash', async ({ page }) => {
    await page.locator('button', { hasText: 'Create your first page' }).click();
    await page.locator('[contenteditable="true"]').first().waitFor({ timeout: 15000 });
    const pageUrl = await page.evaluate(() => window.location.href);

    const sidebarItem = page.locator('[data-page-id]').first();
    await sidebarItem.hover();
    await sidebarItem.locator('button[title="Delete page"]').click();
    // Click confirm in custom dialog
    await page.locator('[role="dialog"] button:has-text("Delete")').click();

    await page.locator('a[href="/trash"]').click();
    await page.locator('text=Untitled').first().waitFor({ timeout: 15000 });
    await page.locator('button', { hasText: 'Restore' }).waitFor({ state: 'attached', timeout: 10000 });

    await page.evaluate(() => {
      const restoreBtn = Array.from(document.querySelectorAll('button')).find(b => b.textContent?.trim() === 'Restore');
      if (restoreBtn) restoreBtn.click();
    });
    await expect(page.getByText('Trash is empty')).toBeVisible({ timeout: 15000 });

    await page.goto(pageUrl);
    const editor = page.locator('[contenteditable="true"]').first();
    await expect(editor).toBeVisible({ timeout: 15000 });
  });

  test('permanent delete removes page forever', async ({ page }) => {
    await page.locator('button', { hasText: 'Create your first page' }).click();
    await page.locator('[contenteditable="true"]').first().waitFor({ timeout: 15000 });

    const sidebarItem = page.locator('[data-page-id]').first();
    await sidebarItem.hover();
    await sidebarItem.locator('button[title="Delete page"]').click();
    // Click confirm in custom dialog
    await page.locator('[role="dialog"] button:has-text("Delete")').click();

    await page.locator('a[href="/trash"]').click();
    await page.locator('text=Untitled').first().waitFor({ timeout: 15000 });
    await page.locator('button', { hasText: 'Delete forever' }).waitFor({ state: 'attached', timeout: 10000 });

    // Click the Delete forever button, then confirm in custom dialog
    await page.locator('button:has-text("Delete forever")').last().click();
    await page.waitForTimeout(300);
    // Confirm in the custom dialog
    await page.locator('[role="dialog"] button:has-text("Delete forever")').click();

    await page.waitForTimeout(1000);
    await page.goto('/trash');
    await expect(page.getByText('Trash is empty')).toBeVisible({ timeout: 15000 });
  });
});
