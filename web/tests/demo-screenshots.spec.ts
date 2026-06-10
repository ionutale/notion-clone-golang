import { test, expect } from '@playwright/test';
import path from 'path';
import fs from 'fs';

const SCREENSHOT_DIR = path.resolve(import.meta.dirname, '../../docs/screenshots');

test.describe('Demo Screenshots', () => {
  test.beforeAll(() => {
    fs.mkdirSync(SCREENSHOT_DIR, { recursive: true });
  });

  test('capture full demo workflow', async ({ page }) => {
    const email = `demo-${Date.now()}@demo.com`;
    const password = 'DemoPassword123!';

    // --- Screenshot 1: Signup page ---
    await page.goto('/signup');
    await page.waitForLoadState('networkidle');
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, '01-signup.png'), fullPage: true });

    // --- Sign up ---
    await page.locator('input[placeholder="Name"]').fill('Demo User');
    await page.locator('input[type="email"]').fill(email);
    await page.locator('input[placeholder="Password"]').fill(password);
    await page.locator('button[type="submit"]').click();
    await page.waitForURL('/');
    await page.waitForLoadState('networkidle');

    // --- Screenshot 2: Empty home / sidebar ---
    await page.waitForTimeout(1000);
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, '02-home-empty.png'), fullPage: true });

    // --- Create a new page ---
    await page.locator('button:has-text("New page")').first().click();
    await page.waitForTimeout(1500);
    await page.waitForLoadState('networkidle');

    // --- Screenshot 3: Fresh empty page ---
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, '03-empty-page.png'), fullPage: true });

    // --- Set page title and emoji icon ---
    const titleField = page.locator('[contenteditable]').first();
    await titleField.click();
    await titleField.fill('');

    // Type the title using keyboard
    await page.keyboard.type('Welcome to Notion Clone');
    await page.waitForTimeout(500);

    // --- Add emoji icon via the icon popover button ---
    // Click the "Add icon" button that appears on hover
    const addIconBtn = page.locator('button:has-text("Add icon"), button[title*="icon"], button:has-text("Icon")').first();
    if (await addIconBtn.isVisible().catch(() => false)) {
      await addIconBtn.click();
      await page.waitForTimeout(500);
      // Click on a rocket emoji or similar
      const emoji = page.locator('button:has-text("🚀"), text=🚀').first();
      if (await emoji.isVisible().catch(() => false)) {
        await emoji.click();
        await page.waitForTimeout(500);
      }
    }

    await page.screenshot({ path: path.join(SCREENSHOT_DIR, '04-page-with-icon.png'), fullPage: true });

    // --- Add a heading 1 block ---
    // Press Enter to create a new block after the title
    await page.keyboard.press('Enter');
    await page.waitForTimeout(300);

    // Type /heading1 or use the slash menu
    await page.keyboard.type('/heading 1');
    await page.waitForTimeout(300);
    // Press Enter to select heading 1 from slash menu
    await page.keyboard.press('Enter');
    await page.waitForTimeout(300);
    await page.keyboard.type('Getting Started with Notion Clone');
    await page.keyboard.press('Enter');
    await page.waitForTimeout(300);

    // --- Screenshot 5: Page with heading ---
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, '05-heading.png'), fullPage: true });

    // --- Add regular text with formatting ---
    await page.keyboard.type('This is a demonstration page showing all the features of the Notion Clone app. ');
    await page.waitForTimeout(200);

    // Add bold text
    await page.keyboard.press('Control+b');
    await page.keyboard.type('Bold text');
    await page.keyboard.press('Control+b');
    await page.keyboard.type(' and ');
    await page.keyboard.press('Control+i');
    await page.keyboard.type('italic text');
    await page.keyboard.press('Control+i');
    await page.keyboard.type(' can be created with keyboard shortcuts.');
    await page.keyboard.press('Enter');
    await page.waitForTimeout(300);

    // --- Add a divider ---
    await page.keyboard.type('/divider');
    await page.waitForTimeout(300);
    await page.keyboard.press('Enter');
    await page.waitForTimeout(300);
    await page.keyboard.press('Enter');
    await page.waitForTimeout(300);

    // --- Add heading 2 ---
    await page.keyboard.type('/heading 2');
    await page.waitForTimeout(300);
    await page.keyboard.press('Enter');
    await page.waitForTimeout(300);
    await page.keyboard.type('Features Overview');
    await page.keyboard.press('Enter');
    await page.waitForTimeout(300);

    // --- Screenshot 6: Features overview ---
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, '06-formatting.png'), fullPage: true });

    // --- Add bullet list items ---
    await page.keyboard.type('/bullet list');
    await page.waitForTimeout(300);
    await page.keyboard.press('Enter');
    await page.waitForTimeout(300);
    await page.keyboard.type('Rich text editing with bold, italic, underline, and strikethrough');
    await page.keyboard.press('Enter');
    await page.waitForTimeout(200);
    await page.keyboard.type('Multiple heading levels (H1, H2, H3)');
    await page.keyboard.press('Enter');
    await page.waitForTimeout(200);
    await page.keyboard.type('Bullet lists and numbered lists');
    await page.keyboard.press('Enter');
    await page.waitForTimeout(200);
    await page.keyboard.type('Toggle lists for collapsible content');
    await page.keyboard.press('Enter');
    await page.waitForTimeout(200);
    await page.keyboard.type('Image uploads and embedding');
    await page.keyboard.press('Enter');
    await page.waitForTimeout(200);
    await page.keyboard.type('Drag and drop to reorder blocks');
    await page.keyboard.press('Enter');
    await page.waitForTimeout(200);
    await page.keyboard.type('Full-text search across all pages');
    await page.keyboard.press('Enter');
    await page.waitForTimeout(200);
    await page.keyboard.type('Page icons and cover images');
    await page.keyboard.press('Enter');
    await page.waitForTimeout(200);
    await page.keyboard.type('Favorites for quick access');
    await page.keyboard.press('Enter');
    await page.waitForTimeout(200);
    await page.keyboard.type('Trash with restore and permanent delete');
    await page.keyboard.press('Enter');
    await page.waitForTimeout(300);

    // --- Screenshot 7: Bullet list ---
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, '07-bullet-list.png'), fullPage: true });

    // --- Add heading 3 ---
    await page.keyboard.type('/heading 3');
    await page.waitForTimeout(300);
    await page.keyboard.press('Enter');
    await page.waitForTimeout(300);
    await page.keyboard.type('Numbered List Example');
    await page.keyboard.press('Enter');
    await page.waitForTimeout(300);

    // --- Add numbered list ---
    await page.keyboard.type('/numbered list');
    await page.waitForTimeout(300);
    await page.keyboard.press('Enter');
    await page.waitForTimeout(300);
    await page.keyboard.type('First step: Create an account');
    await page.keyboard.press('Enter');
    await page.waitForTimeout(200);
    await page.keyboard.type('Second step: Create a workspace');
    await page.keyboard.press('Enter');
    await page.waitForTimeout(200);
    await page.keyboard.type('Third step: Start writing');
    await page.keyboard.press('Enter');
    await page.waitForTimeout(300);

    // --- Screenshot 8: Numbered list ---
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, '08-numbered-list.png'), fullPage: true });

    // --- Add toggle block ---
    await page.keyboard.type('/toggle');
    await page.waitForTimeout(300);
    await page.keyboard.press('Enter');
    await page.waitForTimeout(300);
    await page.keyboard.type('Click to expand — Advanced Features');
    await page.keyboard.press('Enter');
    await page.waitForTimeout(300);
    // Content inside the toggle - need to tab in or click inside
    // Try clicking on the toggle content area
    await page.keyboard.press('Tab');
    await page.waitForTimeout(200);
    await page.keyboard.type('Workspace management with member invitations');
    await page.keyboard.press('Enter');
    await page.waitForTimeout(200);
    await page.keyboard.type('Page icons and cover images for customization');
    await page.keyboard.press('Enter');
    await page.waitForTimeout(200);
    await page.keyboard.type('Keyboard shortcuts for efficient editing');
    await page.keyboard.press('Enter');
    await page.waitForTimeout(300);

    // --- Screenshot 9: Toggle block ---
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, '09-toggle-block.png'), fullPage: true });

    // --- Add another heading ---
    await page.keyboard.type('/heading 2');
    await page.waitForTimeout(300);
    await page.keyboard.press('Enter');
    await page.waitForTimeout(300);
    await page.keyboard.type('Search and Discovery');
    await page.keyboard.press('Enter');
    await page.waitForTimeout(300);

    // --- Add regular paragraph with emoji ---
    await page.keyboard.type('The search functionality allows you to quickly find any content across all your pages. ');
    await page.keyboard.type('Simply click the search icon in the sidebar or press Ctrl+K to open search. ');
    await page.keyboard.type('Results include page titles and content excerpts with relevance ranking. ');
    await page.keyboard.press('Enter');
    await page.waitForTimeout(300);

    // --- Add another heading ---
    await page.keyboard.type('/heading 2');
    await page.waitForTimeout(300);
    await page.keyboard.press('Enter');
    await page.waitForTimeout(300);
    await page.keyboard.type('Drag and Drop');
    await page.keyboard.press('Enter');
    await page.waitForTimeout(300);

    await page.keyboard.type('Blocks can be reordered using drag and drop. Hover over any block to reveal ');
    await page.keyboard.type('the drag handle on the left side, then click and drag to move blocks up or down. ');
    await page.keyboard.type('You can also nest blocks by dragging them slightly to the right.');
    await page.keyboard.press('Enter');
    await page.waitForTimeout(300);

    // --- Final screenshot: Full page ---
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, '10-full-page.png'), fullPage: true });

    // --- Screenshot of sidebar with pages list ---
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, '11-sidebar-and-pages.png'), fullPage: true });

    // --- Take a viewport-only screenshot for the editor close-up ---
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, '12-editor-viewport.png'), fullPage: false });
  });
});
