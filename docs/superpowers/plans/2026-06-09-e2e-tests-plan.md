# E2E Tests Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a full Playwright e2e test suite covering auth, pages, blocks, formatting, drag-and-drop, search, trash, favorites, and settings.

**Architecture:** Playwright runs on the host against the full stack (db + storage + Go app) spun up via `docker compose up -d`. Each test creates a fresh user via API for isolation. Tests are sequential (1 worker) to avoid DB conflicts.

**Tech Stack:** Playwright, TypeScript, Go backend on :8080

---

### Task 1: Install Playwright, create config, helpers, .gitignore

**Files:**
- Modify: `web/package.json:7-14`
- Create: `web/playwright.config.ts`
- Create: `web/test-utils/helpers.ts`
- Modify: `web/.gitignore:1-23`

- [ ] **Step 1: Install Playwright**

```bash
pnpm add -D @playwright/test
npx playwright install chromium
```

- [ ] **Step 2: Update package.json with e2e scripts**

Edit `web/package.json` to add scripts after the `"format"` line:

```json
		"test:e2e": "playwright test",
		"test:e2e:ui": "playwright test --ui"
```

- [ ] **Step 3: Create playwright.config.ts**

```typescript
import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './tests',
  fullyParallel: false,
  retries: 1,
  workers: 1,
  reporter: 'html',
  use: {
    baseURL: 'http://localhost:8080',
    trace: 'on-first-retry',
    video: 'on-first-retry',
  },
});
```

- [ ] **Step 4: Create test-utils/helpers.ts**

```typescript
import { request as pwRequest, type Page } from '@playwright/test';

const BASE_URL = 'http://localhost:8080';

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
```

- [ ] **Step 5: Update .gitignore**

Add to `web/.gitignore`:

```
# Playwright
/test-results
/playwright-report
```

---

### Task 2: Auth tests

**Files:**
- Create: `web/tests/auth.spec.ts`

- [ ] **Write auth.spec.ts**

```typescript
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
    await expect(page.getByText('E2E User')).toBeVisible();
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
```

---

### Task 3: Pages tests

**Files:**
- Create: `web/tests/pages.spec.ts`

- [ ] **Write pages.spec.ts**

```typescript
import { test, expect } from '@playwright/test';
import { createTestUser, loginAs } from '../test-utils/helpers';

test.describe('Pages', () => {
  test.beforeEach(async ({ page }) => {
    const { email, password } = await createTestUser();
    await loginAs(page, email, password);
  });

  test('create page via sidebar button opens editor', async ({ page }) => {
    // On root page with no pages, show "Create your first page"
    await page.locator('button', { hasText: 'Create your first page' }).click();
    await expect(page).toHaveURL(/\/pages\//);
    // Editor loads — page title should be visible
    await expect(page.locator('h1')).toBeVisible();
  });

  test('create page via sidebar "New Page" button', async ({ page }) => {
    // First create one page so the sidebar shows buttons
    await page.locator('button', { hasText: 'Create your first page' }).click();
    await page.waitForURL(/\/pages\//);

    // Now the sidebar "New Page" button should be visible
    await page.locator('button', { hasText: 'New Page' }).click();
    await expect(page).toHaveURL(/\/pages\//);
    // Two page items should exist in sidebar
    const items = page.locator('[data-page-id]');
    await expect(items).toHaveCount(2);
  });

  test('clicking a page in sidebar opens it', async ({ page }) => {
    await page.locator('button', { hasText: 'Create your first page' }).click();
    await page.waitForURL(/\/pages\//);
    const firstUrl = page.url();

    // Create a second page
    await page.locator('button', { hasText: 'New Page' }).click();
    await page.waitForURL(/\/pages\//);
    const secondUrl = page.url();
    expect(firstUrl).not.toBe(secondUrl);

    // Click first page in sidebar
    await page.locator('[data-page-id]').first().click();
    await expect(page).toHaveURL(firstUrl);
  });

  test('delete page via sidebar context button', async ({ page }) => {
    await page.locator('button', { hasText: 'Create your first page' }).click();
    await page.waitForURL(/\/pages\//);
    const pageUrl = page.url();

    // Hover the sidebar page item to reveal delete button
    const sidebarItem = page.locator('[data-page-id]').first();
    await sidebarItem.hover();
    // Click the delete button (last button in the row)
    await sidebarItem.locator('button[title="Delete page"]').click();

    // Confirm dialog
    await expect(page.locator('text=Delete this page?')).toBeVisible();
    // Accept the dialog — Playwright auto-dismisses dialogs by default
    page.once('dialog', d => d.accept());
    // Wait for redirect to root
    await expect(page).toHaveURL('/');
  });
});
```

---

### Task 4: Block editing tests

**Files:**
- Create: `web/tests/blocks.spec.ts`

- [ ] **Write blocks.spec.ts**

```typescript
import { test, expect } from '@playwright/test';
import { createTestUser, loginAs } from '../test-utils/helpers';

test.describe('Blocks', () => {
  test.beforeEach(async ({ page }) => {
    const { email, password } = await createTestUser();
    await loginAs(page, email, password);
    await page.locator('button', { hasText: 'Create your first page' }).click();
    await page.waitForURL(/\/pages\//);
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
    // Should now have two blocks
    const blocks = page.locator('[data-block-id]');
    await expect(blocks).toHaveCount(2);
  });

  test('slash menu transforms block to heading', async ({ page }) => {
    const editor = page.locator('[contenteditable="true"]').first();
    await editor.fill('');
    // Type / to open slash menu
    await editor.type('/');
    // Slash menu should appear
    const slashMenu = page.locator('[role="listbox"]');
    await expect(slashMenu).toBeVisible();
    // Type "heading 1" to filter
    await page.keyboard.type('heading 1');
    await page.keyboard.press('Enter');
    // Block should now be a heading (h1)
    const heading = page.locator('h1').first();
    await expect(heading).toBeVisible();
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
});
```

---

### Task 5: Formatting toolbar tests

**Files:**
- Create: `web/tests/formatting.spec.ts`

- [ ] **Write formatting.spec.ts**

```typescript
import { test, expect } from '@playwright/test';
import { createTestUser, loginAs } from '../test-utils/helpers';

test.describe('Formatting', () => {
  test.beforeEach(async ({ page }) => {
    const { email, password } = await createTestUser();
    await loginAs(page, email, password);
    await page.locator('button', { hasText: 'Create your first page' }).click();
    await page.waitForURL(/\/pages\//);
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
    // Select "Hello"
    await editor.evaluate(el => {
      const range = document.createRange();
      const textNode = el.firstChild!;
      range.setStart(textNode, 0);
      range.setEnd(textNode, 5);
      document.getSelection()?.removeAllRanges();
      document.getSelection()?.addRange(range);
    });
    await page.keyboard.press('Meta+b');
    // Check that bold was applied
    const boldInner = await editor.innerHTML();
    expect(boldInner).toContain('<strong>');
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
    // Now clear formatting
    await page.locator('button[title="Clear Formatting"]').click();
    const afterClear = await editor.innerHTML();
    expect(afterClear).not.toContain('<strong>');
  });
});
```

---

### Task 6: Drag-and-drop sidebar reorder tests

**Files:**
- Create: `web/tests/drag-and-drop.spec.ts`

- [ ] **Write drag-and-drop.spec.ts**

```typescript
import { test, expect } from '@playwright/test';
import { createTestUser, loginAs } from '../test-utils/helpers';

test.describe('Drag-and-drop', () => {
  test.beforeEach(async ({ page }) => {
    const { email, password } = await createTestUser();
    await loginAs(page, email, password);
  });

  test('sidebar pages can be reordered by drag and drop', async ({ page }) => {
    // Create two pages
    await page.locator('button', { hasText: 'Create your first page' }).click();
    await page.waitForURL(/\/pages\//);
    await page.locator('button', { hasText: 'New Page' }).click();
    await page.waitForURL(/\/pages\//);
    await page.locator('button', { hasText: 'New Page' }).click();
    await page.waitForURL(/\/pages\//);

    // Get sidebar page items
    const items = page.locator('[data-page-id]');
    await expect(items).toHaveCount(3);

    // Get initial order
    const firstTitle = await items.first().innerText();
    const lastTitle = await items.last().innerText();

    // Hover the last item to show drag handle, then drag it to the top
    const source = items.last();
    await source.hover();
    const target = items.first();

    // Use dragTo with the drag-handle
    const dragHandle = source.locator('.drag-handle').first();
    await dragHandle.dragTo(target, { targetPosition: { x: 0, y: 0 } });

    // Verify order changed
    const newFirstTitle = await items.first().innerText();
    expect(newFirstTitle).toBe(lastTitle);
    expect(newFirstTitle).not.toBe(firstTitle);
  });
});
```

---

### Task 7: Search + Favorites tests

**Files:**
- Create: `web/tests/search.spec.ts`
- Create: `web/tests/favorites.spec.ts`

- [ ] **Write search.spec.ts**

```typescript
import { test, expect } from '@playwright/test';
import { createTestUser, loginAs } from '../test-utils/helpers';

test.describe('Search', () => {
  test.beforeEach(async ({ page }) => {
    const { email, password } = await createTestUser();
    await loginAs(page, email, password);
  });

  test('search finds a page by title', async ({ page }) => {
    // Create a page with a distinctive title
    await page.locator('button', { hasText: 'Create your first page' }).click();
    await page.waitForURL(/\/pages\//);

    // Go to search page directly
    await page.goto('/search');
    await expect(page.locator('input[type="search"]')).toBeVisible();

    // Type search query
    await page.locator('input[type="search"]').fill('Untitled');
    // Wait for results (debounced 300ms)
    await page.waitForTimeout(500);

    // Should show results
    await expect(page.getByText('Untitled')).toBeVisible();
  });
});
```

- [ ] **Write favorites.spec.ts**

```typescript
import { test, expect } from '@playwright/test';
import { createTestUser, loginAs } from '../test-utils/helpers';

test.describe('Favorites', () => {
  test.beforeEach(async ({ page }) => {
    const { email, password } = await createTestUser();
    await loginAs(page, email, password);
  });

  test('favoriting a page shows it in favorites section', async ({ page }) => {
    await page.locator('button', { hasText: 'Create your first page' }).click();
    await page.waitForURL(/\/pages\//);

    // Click the star button (title "Add to favorites") in the editor header
    await page.locator('button[title="Add to favorites"]').click();

    // Favorites section should appear in sidebar
    await expect(page.getByText('Favorites')).toBeVisible();
  });

  test('unfavoriting removes from favorites', async ({ page }) => {
    await page.locator('button', { hasText: 'Create your first page' }).click();
    await page.waitForURL(/\/pages\//);

    // Add to favorites
    await page.locator('button[title="Add to favorites"]').click();
    await expect(page.getByText('Favorites')).toBeVisible();

    // Remove from favorites
    await page.locator('button[title="Remove from favorites"]').click();
    // Wait a moment for state update
    await page.waitForTimeout(300);
    // Favorites section should disappear
    await expect(page.getByText('Favorites')).not.toBeVisible();
  });
});
```

---

### Task 8: Trash tests

**Files:**
- Create: `web/tests/trash.spec.ts`

- [ ] **Write trash.spec.ts**

```typescript
import { test, expect } from '@playwright/test';
import { createTestUser, loginAs } from '../test-utils/helpers';

test.describe('Trash', () => {
  test.beforeEach(async ({ page }) => {
    const { email, password } = await createTestUser();
    await loginAs(page, email, password);
  });

  test('deleted page appears in trash', async ({ page }) => {
    await page.locator('button', { hasText: 'Create your first page' }).click();
    await page.waitForURL(/\/pages\//);

    // Delete the page via sidebar
    const sidebarItem = page.locator('[data-page-id]').first();
    await sidebarItem.hover();
    page.once('dialog', d => d.accept());
    await sidebarItem.locator('button[title="Delete page"]').click();

    // Go to trash
    await page.goto('/trash');
    await expect(page.getByText('Untitled')).toBeVisible();
  });

  test('restore brings page back from trash', async ({ page }) => {
    await page.locator('button', { hasText: 'Create your first page' }).click();
    await page.waitForURL(/\/pages\//);
    const pageUrl = page.url();

    // Delete
    const sidebarItem = page.locator('[data-page-id]').first();
    await sidebarItem.hover();
    page.once('dialog', d => d.accept());
    await sidebarItem.locator('button[title="Delete page"]').click();

    // Go to trash and restore
    await page.goto('/trash');
    // Hover the trash item to reveal Restore button
    await page.locator('text=Untitled').hover();
    await page.locator('button', { hasText: 'Restore' }).click();
    await page.waitForTimeout(300);

    // Page should be back — navigate to it
    await page.goto(pageUrl);
    const editor = page.locator('[contenteditable="true"]').first();
    await expect(editor).toBeVisible();
  });

  test('permanent delete removes page forever', async ({ page }) => {
    await page.locator('button', { hasText: 'Create your first page' }).click();
    await page.waitForURL(/\/pages\//);

    // Delete
    const sidebarItem = page.locator('[data-page-id]').first();
    await sidebarItem.hover();
    page.once('dialog', d => d.accept());
    await sidebarItem.locator('button[title="Delete page"]').click();

    // Go to trash and permanently delete
    await page.goto('/trash');
    page.once('dialog', d => d.accept());
    await page.locator('text=Untitled').hover();
    await page.locator('button', { hasText: 'Delete forever' }).click();
    await page.waitForTimeout(300);

    // Trash should be empty
    await expect(page.getByText('Trash is empty')).toBeVisible();
  });
});
```

---

### Task 9: Settings tests

**Files:**
- Create: `web/tests/settings.spec.ts`

- [ ] **Write settings.spec.ts**

```typescript
import { test, expect } from '@playwright/test';
import { createTestUser, loginAs } from '../test-utils/helpers';

test.describe('Settings', () => {
  test('change display name', async ({ page }) => {
    const { email, password } = await createTestUser();
    await loginAs(page, email, password);

    await page.goto('/settings');
    const nameInput = page.locator('#settings-name');
    await nameInput.fill('New Name');
    await page.locator('form').first().locator('button[type="submit"]').click();
    await expect(page.getByText('Profile updated')).toBeVisible();

    // Sidebar should reflect new name
    await page.getByText('New Name').first().waitFor();
  });

  test('change password then login with new password', async ({ page }) => {
    const { email, password } = await createTestUser();
    await loginAs(page, email, password);

    const newPassword = 'NewPass789!';
    await page.goto('/settings');
    await page.locator('#settings-current-pw').fill(password);
    await page.locator('#settings-new-pw').fill(newPassword);
    await page.locator('#settings-confirm-pw').fill(newPassword);
    await page.locator('form').nth(1).locator('button[type="submit"]').click();
    await expect(page.getByText('Password updated')).toBeVisible();

    // Logout and login with new password
    await page.getByRole('button', { name: 'Log out' }).click();
    await page.waitForURL('/login');
    await page.locator('input[type="email"]').fill(email);
    await page.locator('input[type="password"]').fill(newPassword);
    await page.locator('button[type="submit"]').click();
    await expect(page).toHaveURL('/');
  });
});
```

---

### Task 10: Final run and verify

- [ ] **Step 1: Build the frontend**

```bash
cd web && pnpm build
```

- [ ] **Step 2: Start the stack**

```bash
docker compose up -d --build
```

Wait for the app to be healthy at `http://localhost:8080/api/v1/health` (status `ok`).

- [ ] **Step 3: Run the e2e tests**

```bash
cd web && pnpm test:e2e
```

Expected: all tests pass.

- [ ] **Step 4: Commit**

```bash
git add web/playwright.config.ts web/test-utils/ web/tests/ web/package.json web/.gitignore
git commit -m "test: add Playwright e2e test suite"
```
