import { test, expect } from '@playwright/test';
import { createTestUser, loginAs } from '../test-utils/helpers';

test.describe('Search', () => {
  test.beforeEach(async ({ page }) => {
    const { email, password } = await createTestUser();
    await loginAs(page, email, password);
  });

  test('search finds content across pages', async ({ page }) => {
    await page.locator('button', { hasText: 'Create your first page' }).click();
    await page.locator('[contenteditable="true"]').first().waitFor({ timeout: 15000 });

    await page.keyboard.press('Meta+k');
    await expect(page.locator('input[type="search"]')).toBeVisible({ timeout: 15000 });

    await page.locator('input[type="search"]').fill('Untitled');
    await expect(page.getByText('Untitled')).toBeVisible({ timeout: 15000 });
  });
});
