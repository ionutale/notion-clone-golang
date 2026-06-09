import { test, expect } from '@playwright/test';
import { createTestUser, loginAs } from '../test-utils/helpers';

test.describe('Formatting', () => {
  test.beforeEach(async ({ page }) => {
    const { email, password } = await createTestUser();
    await loginAs(page, email, password);
    await page.locator('button', { hasText: 'Create your first page' }).click();
    await page.locator('[contenteditable="true"]').first().waitFor({ timeout: 15000 });
  });

  test('toolbar is visible on a page', async ({ page }) => {
    const toolbar = page.locator('[role="toolbar"]');
    await expect(toolbar).toBeVisible();
    await expect(toolbar.locator('button[title="Bold (Cmd+B)"]')).toBeVisible();
    await expect(toolbar.locator('button[title="Italic (Cmd+I)"]')).toBeVisible();
    await expect(toolbar.locator('button[title="Underline (Cmd+U)"]')).toBeVisible();
  });

  test('keyboard shortcut Cmd+B bolds selected text', async ({ page }) => {
    const editor = page.locator('[contenteditable="true"]').first();
    await editor.fill('Hello World');
    await editor.evaluate(el => {
      const range = document.createRange();
      const textNode = el.firstChild!;
      range.setStart(textNode, 0);
      range.setEnd(textNode, 5);
      document.getSelection()?.removeAllRanges();
      document.getSelection()?.addRange(range);
    });
    await page.keyboard.press('Meta+b');
    const boldInner = await editor.innerHTML();
    expect(boldInner).toMatch(/<(strong|b)>/);
  });

  test('toolbar clear formatting button works', async ({ page }) => {
    const editor = page.locator('[contenteditable="true"]').first();
    await editor.fill('Hello World');
    await editor.evaluate(el => {
      const range = document.createRange();
      const textNode = el.firstChild!;
      range.setStart(textNode, 0);
      range.setEnd(textNode, 5);
      document.getSelection()?.removeAllRanges();
      document.getSelection()?.addRange(range);
    });
    await page.keyboard.press('Meta+b');
    await page.locator('button[title="Clear Formatting"]').click();
    const afterClear = await editor.innerHTML();
    expect(afterClear).not.toMatch(/<(strong|b)>/);
  });
});
