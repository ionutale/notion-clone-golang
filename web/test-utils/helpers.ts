import { request as pwRequest, type Page } from '@playwright/test';

const BASE_URL = 'http://127.0.0.1:8080';

export async function createTestUser(): Promise<{ email: string; password: string }> {
  const ctx = await pwRequest.newContext({ baseURL: BASE_URL });
  const email = `e2e-${Date.now()}@test.com`;
  const password = 'Test123!';
  const res = await ctx.post('/api/v1/auth/signup', {
    data: { email, password, name: 'E2E User' },
  });
  await ctx.dispose();
  if (!res.ok()) throw new Error(`Signup failed: ${await res.text()}`);
  return { email, password };
}

export async function loginAs(page: Page, email: string, password: string): Promise<void> {
  await page.goto('/login');
  await page.locator('input[type="email"]').fill(email);
  await page.locator('input[type="password"]').fill(password);
  await page.locator('button[type="submit"]').click();
  await page.waitForURL('/');
}
