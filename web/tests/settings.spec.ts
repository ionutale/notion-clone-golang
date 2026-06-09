import { test, expect } from '@playwright/test';
import { createTestUser, loginAs } from '../test-utils/helpers';

test.describe('Settings', () => {
  test('change display name', async ({ page }) => {
    const { email, password } = await createTestUser();
    await loginAs(page, email, password);

    await page.locator('a[href="/settings"]').click();
    const nameInput = page.locator('#settings-name');
    await nameInput.waitFor({ timeout: 10000 });
    await nameInput.fill('New Name');
    await page.locator('form').first().locator('button[type="submit"]').click();
    await expect(page.getByText('Profile updated')).toBeVisible();
  });

  test('change password then login with new password', async ({ page }) => {
    const { email, password } = await createTestUser();
    await loginAs(page, email, password);

    const newPassword = 'NewPass789!';
    await page.locator('a[href="/settings"]').click();
    await page.locator('#settings-current-pw').waitFor({ timeout: 10000 });
    await page.locator('#settings-current-pw').fill(password);
    await page.locator('#settings-new-pw').fill(newPassword);
    await page.locator('#settings-confirm-pw').fill(newPassword);
    await page.locator('form').nth(1).locator('button[type="submit"]').click();
    await expect(page.getByText('Password updated')).toBeVisible();

    await page.evaluate(() => fetch('/api/v1/auth/logout', { method: 'POST' }));
    await page.goto('/login');
    await page.locator('input[type="email"]').waitFor({ timeout: 10000 });
    await page.locator('input[type="email"]').fill(email);
    await page.locator('input[type="password"]').fill(newPassword);
    await page.locator('button[type="submit"]').click();
    await expect(page).toHaveURL('/');
  });
});
