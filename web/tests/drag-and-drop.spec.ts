import { test, expect } from '@playwright/test';
import { createTestUser, loginAs } from '../test-utils/helpers';

test.describe('Drag-and-drop', () => {
  test.beforeEach(async ({ page }) => {
    const { email, password } = await createTestUser();
    await loginAs(page, email, password);
  });

  test('sidebar pages can be reordered by drag and drop', async ({ page }) => {
    await page.locator('button', { hasText: 'Create your first page' }).click();
    await page.locator('[contenteditable="true"]').first().waitFor({ timeout: 15000 });
    await page.locator('button', { hasText: 'New Page' }).click();
    await page.locator('[contenteditable="true"]').first().waitFor({ timeout: 15000 });
    await page.locator('button', { hasText: 'New Page' }).click();
    await page.locator('[contenteditable="true"]').first().waitFor({ timeout: 15000 });

    const items = page.locator('[data-page-id]');
    await expect(items).toHaveCount(3);

    const firstId = await items.first().getAttribute('data-page-id');
    const lastId = await items.last().getAttribute('data-page-id');

    const source = items.last();
    await source.hover();
    const target = items.first();
    const dragHandle = source.locator('.drag-handle').first();
    await dragHandle.dragTo(target, { targetPosition: { x: 0, y: 0 } });

    const newFirstId = await items.first().getAttribute('data-page-id');
    expect(newFirstId).toBe(lastId);
    expect(newFirstId).not.toBe(firstId);
  });
});
