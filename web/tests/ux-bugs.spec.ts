import { test, expect } from '@playwright/test';
import { createTestUser, loginAs } from '../test-utils/helpers';

test.describe('UX Bug Fixes', () => {
  test.beforeEach(async ({ page }) => {
    const { email, password } = await createTestUser();
    await loginAs(page, email, password);
  });

  // Bug 1: Page title should be editable
  test('page title should be editable', async ({ page }) => {
    await page.locator('button', { hasText: 'Create your first page' }).click();
    await page.locator('[contenteditable="true"]').first().waitFor({ timeout: 15000 });

    const title = page.locator('input[aria-label="Page title"]');
    await expect(title).toBeVisible();
    await title.fill('My Custom Title');
    await page.locator('[contenteditable="true"]').first().focus();
    await page.waitForTimeout(500);
    await expect(title).toHaveValue('My Custom Title');
  });

  // Bug 2: Add cover should open CoverPopover
  test('add cover opens popover', async ({ page }) => {
    await page.locator('button', { hasText: 'Create your first page' }).click();
    await page.locator('[contenteditable="true"]').first().waitFor({ timeout: 15000 });

    await page.evaluate(() => {
      const btn = Array.from(document.querySelectorAll('button')).find(b => b.textContent?.trim() === 'Add cover');
      if (btn) btn.dispatchEvent(new MouseEvent('click', { bubbles: true, cancelable: true }));
    });
    await page.waitForTimeout(1000);

    await expect(page.locator('[role="dialog"]')).toBeVisible({ timeout: 5000 });
  });

  // Bug 3: No native dialogs - custom modals
  test('settings page should not use native prompt() for email change', async ({ page }) => {
    await page.locator('button', { hasText: 'Create your first page' }).click();
    await page.locator('[contenteditable="true"]').first().waitFor({ timeout: 15000 });

    await page.locator('a[href="/settings"]').click();
    await page.waitForURL('/settings');

    await page.locator('#settings-email').fill('newemail@test.com');

    let dialogCount = 0;
    page.on('dialog', () => dialogCount++);

    await page.locator('button[type="submit"]').first().click();
    await page.waitForTimeout(500);

    expect(dialogCount).toBe(0);
  });

  // Bug 4: Enter on empty block should not create new block
  test('pressing Enter on empty block does not create new block', async ({ page }) => {
    await page.locator('button', { hasText: 'Create your first page' }).click();
    await page.locator('[contenteditable="true"]').first().waitFor({ timeout: 15000 });

    const editor = page.locator('[contenteditable="true"]').first();
    await editor.focus();
    await editor.fill('');

    const initialCount = await page.locator('[data-block-id]').count();

    await editor.press('Enter');
    await page.waitForTimeout(500);

    const newCount = await page.locator('[data-block-id]').count();
    expect(newCount).toBe(initialCount);
  });

  // Bug 5: Slash menu should open on /
  test('slash menu opens on / key', async ({ page }) => {
    await page.locator('button', { hasText: 'Create your first page' }).click();
    await page.locator('[contenteditable="true"]').first().waitFor({ timeout: 15000 });

    const editor = page.locator('[contenteditable="true"]').first();
    await editor.fill('');
    await editor.focus();
    await page.keyboard.press('/');

    await expect(page.locator('[role="listbox"]')).toBeVisible({ timeout: 10000 });
  });

  // Bug 6: Pasting image should create image block
  test('pasting image HTML creates image block', async ({ page }) => {
    await page.locator('button', { hasText: 'Create your first page' }).click();
    await page.locator('[contenteditable="true"]').first().waitFor({ timeout: 15000 });

    const editor = page.locator('[contenteditable="true"]').first();
    await editor.focus();
    await editor.fill('');

    await page.evaluate(() => {
      const el = document.activeElement;
      if (!el) return;
      const dt = new DataTransfer();
      dt.setData('text/html', '<img src="https://placehold.co/600x400/png" />');
      const event = new ClipboardEvent('paste', {
        clipboardData: dt,
        bubbles: true,
        cancelable: true,
      });
      el.dispatchEvent(event);
    });
    await page.waitForTimeout(1000);

    const imageBlocks = page.locator('.image-block');
    await expect(imageBlocks).toHaveCount(1);
  });

  // Bug 7: Settings has back button
  test('settings page has a back button', async ({ page }) => {
    await page.locator('button', { hasText: 'Create your first page' }).click();
    await page.locator('[contenteditable="true"]').first().waitFor({ timeout: 15000 });

    await page.locator('a[href="/settings"]').click();
    await page.waitForURL('/settings');

    await expect(page.locator('button:has-text("Back")').first()).toBeVisible();
  });
});
