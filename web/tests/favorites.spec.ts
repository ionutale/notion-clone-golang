import { test, expect } from '@playwright/test';
import { createTestUser, loginAs } from '../test-utils/helpers';

test.describe('Favorites', () => {
  test.beforeEach(async ({ page }) => {
    const { email, password } = await createTestUser();
    await loginAs(page, email, password);
  });

  test('favoriting a page shows it in favorites section', async ({ page }) => {
    await page.locator('button', { hasText: 'Create your first page' }).click();
    await page.locator('[contenteditable="true"]').first().waitFor({ timeout: 15000 });

    await page.locator('button[title="Add to favorites"]').click();
    await expect(page.getByText('Favorites')).toBeVisible();
  });

  test('unfavoriting removes from favorites', async ({ page }) => {
    await page.locator('button', { hasText: 'Create your first page' }).click();
    await page.locator('[contenteditable="true"]').first().waitFor({ timeout: 15000 });

    await page.locator('button[title="Add to favorites"]').click();
    await expect(page.getByText('Favorites')).toBeVisible();

    await page.locator('button[title="Remove from favorites"]').click();
    await page.waitForTimeout(300);
    await expect(page.getByText('Favorites')).not.toBeVisible();
  });
});
