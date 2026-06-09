import { test, expect } from '@playwright/test';
import { createTestUser, loginAs } from '../test-utils/helpers';

test.describe('Auth', () => {
  test('signup creates account and redirects to app', async ({ page }) => {
    const email = `e2e-${Date.now()}@test.com`;
    await page.goto('/signup');
    await page.locator('input[placeholder="Name"]').fill('E2E User');
    await page.locator('input[type="email"]').fill(email);
    await page.locator('input[placeholder="Password"]').fill('Test123!');
    await page.locator('button[type="submit"]').click();
    await expect(page).toHaveURL('/');
    await expect(page.getByRole('button', { name: 'Log out' })).toBeVisible();
  });

  test('login redirects to app', async ({ page }) => {
    const { email, password } = await createTestUser();
    await page.goto('/login');
    await page.locator('input[type="email"]').fill(email);
    await page.locator('input[type="password"]').fill(password);
    await page.locator('button[type="submit"]').click();
    await expect(page).toHaveURL('/');
  });

  test('logout redirects to login', async ({ page }) => {
    const { email, password } = await createTestUser();
    await loginAs(page, email, password);
    await page.getByRole('button', { name: 'Log out' }).click();
    await expect(page).toHaveURL('/login');
  });

  test('protected route redirects to login when unauthenticated', async ({ page }) => {
    await page.goto('/pages/some-id');
    await expect(page).toHaveURL('/login');
  });
});
