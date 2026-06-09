import { test, expect } from '@playwright/test';
import { createTestUser, loginAs } from '../test-utils/helpers';

test.describe('Blocks', () => {
  test.beforeEach(async ({ page }) => {
    const { email, password } = await createTestUser();
    await loginAs(page, email, password);
    await page.locator('button', { hasText: 'Create your first page' }).click();
    await page.locator('[contenteditable="true"]').first().waitFor({ timeout: 15000 });
  });

  test('new page has an empty text block', async ({ page }) => {
    const block = page.locator('[data-block-id]').first();
    await expect(block).toBeVisible();
    const editor = block.locator('[contenteditable="true"]');
    await expect(editor).toHaveAttribute('role', 'textbox');
  });

  test('typing in a block and Enter creates a new block', async ({ page }) => {
    const editor = page.locator('[contenteditable="true"]').first();
    await editor.fill('');
    await editor.type('Hello world');
    await editor.press('Enter');
    const blocks = page.locator('[data-block-id]');
    await expect(blocks).toHaveCount(2);
  });

  test('slash menu transforms block to heading', async ({ page }) => {
    const editor = page.locator('[contenteditable="true"]').first();
    await editor.fill('');
    await editor.type('/');
    const slashMenu = page.locator('[role="listbox"]');
    await expect(slashMenu).toBeVisible();
    await page.keyboard.type('heading 1');
    await page.keyboard.press('Enter');
    // After transform, the block editor should still be present but now a heading
    const blockEditor = page.locator('[role="textbox"]').first();
    await expect(blockEditor).toBeVisible();
  });

  test('slash menu transforms to bullet list', async ({ page }) => {
    const editor = page.locator('[contenteditable="true"]').first();
    await editor.fill('');
    await editor.type('/');
    await page.keyboard.type('bullet');
    await page.keyboard.press('Enter');
    const bullet = page.locator('ul li').first();
    await expect(bullet).toBeVisible();
  });

  test('slash menu transforms to divider', async ({ page }) => {
    const editor = page.locator('[contenteditable="true"]').first();
    await editor.fill('');
    await editor.type('/');
    await page.keyboard.type('divider');
    await page.keyboard.press('Enter');
    const divider = page.locator('[data-block-id]').first();
    await expect(divider).toBeVisible();
  });

  test('Enter on text block creates new blocks with correct content', async ({ page }) => {
    // Memorize initial state
    const initialBlocks = page.locator('[data-block-id]');
    const initialCount = await initialBlocks.count();
    const initialText = await page.locator('[role="textbox"]').first().textContent();

    // Type "test-text" via innerHTML
    await page.evaluate(() => {
      const ce = document.querySelector('[role="textbox"]') as HTMLElement;
      if (ce) ce.innerHTML = 'test-text';
    });
    await page.waitForTimeout(200);

    // Press Enter to create new block
    await page.locator('[role="textbox"]').first().press('Enter');
    await page.waitForTimeout(500);

    // Type "new-demo-text" in the new block
    await page.evaluate(() => {
      const editors = document.querySelectorAll('[role="textbox"]');
      const second = editors[1] as HTMLElement;
      if (second) second.innerHTML = 'new-demo-text';
    });
    await page.waitForTimeout(200);

    // Press Enter again
    await page.locator('[role="textbox"]').nth(1).press('Enter');
    await page.waitForTimeout(500);

    // Type "some text" in the third block
    await page.evaluate(() => {
      const editors = document.querySelectorAll('[role="textbox"]');
      const third = editors[2] as HTMLElement;
      if (third) third.innerHTML = 'some text';
    });
    await page.waitForTimeout(200);

    // Verify: block count = initial + 2
    const finalCount = await initialBlocks.count();
    expect(finalCount).toBe(initialCount + 2);

    // Verify first block text preserved
    const texts = await page.locator('[role="textbox"]').allTextContents();
    expect(texts[0].trim()).toBe('test-text');
    expect(texts[1].trim()).toBe('new-demo-text');
    expect(texts[2].trim()).toBe('some text');
  });

  test('Enter in numbered list creates new empty item and moves focus', async ({ page }) => {
    // Transform first block to numbered list via slash menu
    const editor = page.locator('[contenteditable="true"]').first();
    await editor.fill('');
    await editor.focus();
    await page.keyboard.press('/');
    await page.locator('[role="listbox"]').waitFor({ timeout: 10000 });
    await page.locator('input[placeholder="Filter..."]').fill('numbered');
    await page.keyboard.press('Enter');
    await page.waitForTimeout(500);

    const listEditors = page.locator('.list-item [contenteditable="true"]');
    await expect(listEditors).toHaveCount(1);

    // Set text content via innerHTML
    await page.evaluate(() => {
      const ce = document.querySelector('.list-item [contenteditable="true"]') as HTMLElement;
      if (ce) ce.innerHTML = 'Line one';
    });
    await page.waitForTimeout(200);

    // Press Enter to create new block below
    await listEditors.first().press('Enter');
    await page.waitForTimeout(500);

    // Should have 2 blocks
    await expect(listEditors).toHaveCount(2);

    // First editor should still have original text
    await expect(listEditors.nth(0)).toHaveText('Line one');
    // Second editor should be empty (NOT duplicated)
    await expect(listEditors.nth(1)).toHaveText('');

    // Focus should be on the second editor
    const activeRole = await page.evaluate(() => {
      return document.activeElement?.getAttribute('role') ?? null;
    });
    expect(activeRole).toBe('textbox');

    // The active editor should be the second one (has empty content)
    const activeContent = await page.evaluate(() => {
      return (document.activeElement as HTMLElement)?.textContent ?? '';
    });
    expect(activeContent.trim()).toBe('');

    // Type in the second editor
    await page.evaluate(() => {
      const editors = document.querySelectorAll('.list-item [contenteditable="true"]');
      const second = editors[1] as HTMLElement;
      if (second) second.innerHTML = 'Line two';
    });
    await page.waitForTimeout(200);
    await listEditors.nth(1).press('Enter');
    await page.waitForTimeout(500);

    // Should have 3 blocks with correct content
    await expect(page.locator('.list-item [contenteditable="true"]')).toHaveCount(3);

    // Verify no text was duplicated from earlier blocks
    const texts = await page.locator('.list-item [contenteditable="true"]').allTextContents();
    expect(texts[0].trim()).toBe('Line one');
    expect(texts[1].trim()).toBe('Line two');
    expect(texts[2].trim()).toBe('');
  });

  test('Enter in bullet list does not duplicate text', async ({ page }) => {
    // Transform first block to bullet list via slash menu
    const editor = page.locator('[contenteditable="true"]').first();
    await editor.fill('');
    await editor.focus();
    await page.keyboard.press('/');
    await page.locator('[role="listbox"]').waitFor({ timeout: 10000 });
    await page.locator('input[placeholder="Filter..."]').fill('bullet');
    await page.keyboard.press('Enter');
    await page.waitForTimeout(500);

    // Should have 1 block (the transformed bullet list)
    const bulletEditors = page.locator('.list-item [contenteditable="true"]');
    await expect(bulletEditors).toHaveCount(1);

    // Set text content via innerHTML to avoid fill() + async save() race
    await page.evaluate(() => {
      const ce = document.querySelector('.list-item [contenteditable="true"]') as HTMLElement;
      if (ce) ce.innerHTML = 'Line one';
    });
    await page.waitForTimeout(200);

    // Press Enter to create new block below
    const editorEl = page.locator('.list-item [contenteditable="true"]').first();
    await editorEl.press('Enter');
    await page.waitForTimeout(500);

    // Should now have 2 blocks (both bullet list items)
    const allEditors = page.locator('.list-item [contenteditable="true"]');
    await expect(allEditors).toHaveCount(2);

    // First editor should still have original text
    await expect(allEditors.nth(0)).toHaveText('Line one');
    // Second editor should be empty
    await expect(allEditors.nth(1)).toHaveText('');

    // Type in second editor via innerHTML
    await page.evaluate(() => {
      const editors = document.querySelectorAll('.list-item [contenteditable="true"]');
      const second = editors[1] as HTMLElement;
      if (second) second.innerHTML = 'Line two';
    });
    await page.waitForTimeout(200);
    await page.locator('.list-item [contenteditable="true"]').nth(1).press('Enter');
    await page.waitForTimeout(500);

    // Should have 3 blocks with correct content
    await expect(page.locator('.list-item [contenteditable="true"]')).toHaveCount(3);
    const texts = await page.locator('.list-item [contenteditable="true"]').allTextContents();
    expect(texts[0].trim()).toBe('Line one');
    expect(texts[1].trim()).toBe('Line two');
    expect(texts[2].trim()).toBe('');
  });
});
