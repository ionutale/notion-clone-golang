import { test, expect } from '@playwright/test';
import { createTestUser, loginAs } from '../test-utils/helpers';
import path from 'path';
import fs from 'fs';

const SCREENSHOT_DIR = path.resolve(import.meta.dirname, '../../docs/screenshots');

test.describe('Code Block', () => {
  test.beforeAll(() => {
    fs.mkdirSync(SCREENSHOT_DIR, { recursive: true });
  });

  test('code block renders and accepts typed code', async ({ page, request }) => {
    test.setTimeout(60000);
    const { email, password } = await createTestUser();
    await loginAs(page, email, password);

    // Create a page via API
    const loginRes = await request.post('/api/v1/auth/login', { data: { email, password } });
    const loginData = await loginRes.json();
    const token = loginData.access_token;
    const authHeaders = { Authorization: `Bearer ${token}` };

    const wsRes = await request.get('/api/v1/workspaces', { headers: authHeaders });
    const wsData = await wsRes.json();
    const wsId = wsData[0]?.id || wsData.items?.[0]?.id;

    const pageRes = await request.post(`/api/v1/workspaces/${wsId}/pages`, {
      headers: { ...authHeaders, 'Content-Type': 'application/json' },
      data: { title: 'Code Block Demo' }
    });
    const pageData = await pageRes.json();
    const pageId = pageData.id;

    // Create a text block
    await request.post(`/api/v1/workspaces/${wsId}/blocks`, {
      headers: { ...authHeaders, 'Content-Type': 'application/json' },
      data: { parent_id: pageId, type: 'text', content: { html: '<p>Code example below:</p>' } }
    });

    // Create a code block
    const blockRes = await request.post(`/api/v1/workspaces/${wsId}/blocks`, {
      headers: { ...authHeaders, 'Content-Type': 'application/json' },
      data: { parent_id: pageId, type: 'code', content: { code: '' } }
    });
    const blockData = await blockRes.json();
    console.log(`Page: ${pageId}  Code block: ${blockData.id}`);

    // Navigate to the page
    await page.goto(`/pages/${pageId}`);
    await page.waitForTimeout(2000);
    await page.waitForLoadState('networkidle');

    // Wait for the code block
    const codeBlock = page.locator('pre[contenteditable="true"]').first();
    const visible = await codeBlock.isVisible({ timeout: 10000 }).catch(() => false);
    console.log('Code block visible:', visible);

    if (!visible) {
      // Debug
      const blocks = await page.evaluate(() => {
        return Array.from(document.querySelectorAll('[data-block-id]')).map(b => {
          const id = b.getAttribute('data-block-id');
          const edits = b.querySelectorAll('[contenteditable]');
          return { id, editCount: edits.length, editTags: Array.from(edits).map(e => e.tagName) };
        });
      });
      console.log('All blocks:', JSON.stringify(blocks, null, 2));

      await page.screenshot({ path: path.join(SCREENSHOT_DIR, 'debug-code-block.png'), fullPage: true });
      throw new Error('Code block not rendered');
    }

    // Type code using evaluate to avoid contenteditable issues
    await codeBlock.click();
    await page.waitForTimeout(300);

    await page.evaluate(() => {
      const pre = document.querySelector('pre[contenteditable="true"]');
      if (pre) pre.textContent = '';
    });
    await page.waitForTimeout(200);

    const codeLines = [
      'package main', '',
      'import "fmt"', '',
      'func main() {',
      '  fmt.Println("Hello, Notion Clone!")',
      '}',
    ];
    for (let i = 0; i < codeLines.length; i++) {
      await page.keyboard.type(codeLines[i]);
      if (i < codeLines.length - 1) {
        await page.keyboard.press('Enter');
        await page.waitForTimeout(100);
      }
    }
    await page.waitForTimeout(500);

    const text = await page.evaluate(() => {
      const pre = document.querySelector('pre[contenteditable="true"]');
      return pre?.textContent ?? '';
    });
    expect(text).toContain('package main');
    expect(text).toContain('fmt.Println');
    console.log('Code verified!');

    // Screenshots
    await page.screenshot({
      path: path.join(SCREENSHOT_DIR, '13-code-block.png'),
      fullPage: true,
    });

    await codeBlock.scrollIntoViewIfNeeded();
    await page.waitForTimeout(300);
    await page.screenshot({
      path: path.join(SCREENSHOT_DIR, '14-code-block-closeup.png'),
      fullPage: false,
    });
  });
});
