# E2E Tests Design

## Infrastructure

- **Stack:** PostgreSQL + fake-gcs-server + Go app via `docker compose up -d`
- **Playwright:** installed as dev dependency in `web/`, runs on host
- **Config:** `web/playwright.config.ts` — baseURL `http://localhost:8080`, Chromium only, 1 retry, video on failure

## File Structure

```
web/
  playwright.config.ts
  tests/
    auth.spec.ts
    pages.spec.ts
    blocks.spec.ts
    formatting.spec.ts
    drag-and-drop.spec.ts
    search.spec.ts
    trash.spec.ts
    favorites.spec.ts
    settings.spec.ts
  test-utils/
    helpers.ts
```

## Package Scripts

- `test:e2e` — `playwright test`
- `test:e2e:ui` — `playwright test --ui`

## Test Coverage

### auth.spec.ts
- signup → redirect to app + authenticated
- login → redirect to app
- logout → redirect to `/login`
- visit `/pages` while logged out → redirect to `/login`

### pages.spec.ts
- create page via sidebar "+" → appears in tree
- click page in sidebar → editor loads
- rename title in editor
- delete page → moves to trash

### blocks.spec.ts
- new page has default empty block
- type text in a block
- Enter creates a new block
- `/` slash menu transforms block type: heading, bullet list, numbered list, to-do, divider, code, quote
- Tab indent / Shift+Tab outdent a block

### formatting.spec.ts
- toolbar: Bold, Italic, Underline, Strikethrough, Code, Link
- keyboard shortcuts: Cmd+B, Cmd+I, Cmd+U, Cmd+Shift+S
- clear formatting button
- link toolbar → enter URL

### drag-and-drop.spec.ts
- create 3+ text blocks with distinct content
- drag block by handle → reorder
- drop indicator line appears during drag

### search.spec.ts
- create pages with distinct titles
- navigate to `/search`
- type query → result appears matching page

### trash.spec.ts
- delete page → visible in `/trash`
- restore → back in sidebar
- permanent delete → 404 on access

### favorites.spec.ts
- star a page → appears in "Favorites" section in sidebar
- unstar → removed from favorites

### settings.spec.ts
- change display name
- change password → logout → login with new password succeeds

## Test Fixtures & Patterns

- `test-utils/helpers.ts`: `createTestUser()` (signup via API POST, returns `{email, password}`), `loginAs(page, email, password)` (fill login form, wait for redirect)
- `beforeEach`: create fresh user + login — every test starts authenticated with a clean user
- `afterAll`: best-effort cleanup via delete-account endpoint (not critical; trash auto-expires in 30 days)
- Database migrations auto-apply on Go backend startup
- Tests are independent — no shared state between specs

## Runbook

```bash
# Terminal 1: start the stack
docker compose up -d

# Terminal 2: run tests
cd web && pnpm test:e2e

# Or with UI mode
cd web && pnpm test:e2e:ui
```
