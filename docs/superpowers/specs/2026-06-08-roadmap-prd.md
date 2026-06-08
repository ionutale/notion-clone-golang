# Product Requirements Document: Notion Clone — v2 Feature Set

## Problem Statement

The Notion clone has a solid core (block editing, auth, workspaces, page icons, block drag-and-drop) but lacks several essential features that make Notion a useful daily driver: discoverable search, page recovery, bookmarking favorites, workspace collaboration, and user customization.

## Solution

Ship eight features that transform the app from a basic editor into a usable personal knowledge base:

1. **Sidebar page reordering** — drag pages to rearrange them in the sidebar
2. **Full-text search** — find any block content across all pages instantly
3. **Trash** — recover recently deleted pages, with auto-expiry
4. **Page cover images** — add a banner image to pages
5. **Favorites** — pin important pages to the top of the sidebar
6. **User settings** — manage name, email, password, and account deletion
7. **Workspace invite** — add members by email
8. **Backend tests** — test coverage for auth and workspace handlers

## User Stories

1. As a user, I want to drag pages in the sidebar to reorder them, so that I can organize my workspace.
2. As a user, I want to search across all page titles and block content, so that I can quickly find information even in large workspaces.
3. As a user, I want to see block-level search results with context snippets, so that I can jump directly to the relevant content.
4. As a user, I want to view my recently deleted pages, so that I can recover something I accidentally deleted.
5. As a user, I want trashed items to auto-expire after a configurable period, so that storage stays clean without manual effort.
6. As a user, I want to add a cover image to my pages, so that they feel more personalized and visually distinct.
7. As a user, I want to favorite important pages, so that they appear at the top of my sidebar for quick access.
8. As a user, I want to update my name, email, and password, so that my account stays current.
9. As a user, I want to delete my account, so that I can remove my data when I no longer need it.
10. As a user, I want to invite other users to my workspace by email, so that we can collaborate.
11. As a developer, I want tests for auth and workspace handlers, so that I can refactor with confidence.

## Implementation Decisions

### Sidebar Page Reordering
- Native HTML5 Drag & Drop API (consistent with existing block DnD)
- Each sidebar list item gets `draggable="true"`, a 6-dot grip handle on hover, 500ms long-press for touch
- Uses existing `PATCH /blocks/{id}/move` with `{ parent_id: null, position: N }`
- Drag disabled when search filter is active (filtered order ≠ real order)
- Optimistic local reorder, async persistence

### Full-Text Search
- PostgreSQL `tsvector` + GIN index for full-text search across block content
- New migration adds a generated `search_vector tsvector` column
- New endpoint: `GET /workspaces/{workspaceId}/search?q=...` returns block-level results with context snippets
- Frontend: `/search?q=...` full-page route (not palette/overlay)
- Results show page title + matching block content excerpt (HTML tags stripped for display)
- Existing sidebar filter remains but links to `/search?q=...` for full results

### Trash
- New route `/trash` listing soft-deleted pages (`deleted_at IS NOT NULL`)
- Restore via existing `PATCH /blocks/{id}/restore`
- Permanent delete via new `DELETE /blocks/{id}/permanent`
- Auto-expire: configurable threshold (default 30 days), checked on list load
- Toast undo from inline deletion still works — trash view supplements it

### Page Cover Image
- Store `cover` (URL), `cover_type` ("image"), and `cover_color` as fallback in `block.content` JSONB
- No backend changes — same `PATCH /blocks/{id}` pattern as icons
- Editor: fixed-height 200px banner at the top, before the page icon
- Click to change: popover with upload + URL input (reuse pattern from `IconPopover`)
- Cover color fallback for pages without an image cover (solid color from a palette)

### Favorites
- Store `favorited: true` in page `block.content` JSONB — no migration
- Sidebar: "Favorites" section at the top of the page list, featuring favorited pages
- Toggle star icon on hover in sidebar or at the top of the Editor
- Backend: `GET /workspaces/{workspaceId}/favorites` endpoint filtering `(content->>'favorited')::boolean = true`

### User Settings
- Main settings page at `/settings` with sections: Profile, Password, Danger Zone
- Profile: change name + email (with current password confirmation)
- Password: change password (current + new + confirm)
- Danger Zone: delete account (with confirmation flow — type email to confirm)
- Backend: `PATCH /auth/me`, `PATCH /auth/me/password`, `DELETE /auth/me`

### Workspace Invite
- Workspace settings section: email input + "Invite" button
- Uses existing backend `InviteMember`
- On invite, the invited user is immediately added as a workspace member
- No email notification, no approval flow

### Backend Tests
- `testify` suite with mock repository interfaces
- Auth handler tests: signup (success, duplicate email, missing fields), login (success, wrong password, nonexistent email), refresh (valid token, expired token), logout, me (authenticated, unauthenticated)
- Workspace handler tests: create, list, rename, invite member (success, duplicate, nonexistent user)
- No database dependency — mock the repository layer

## Testing Decisions
- Tests should verify external behavior, not implementation details
- Use testify for assertions + mock interfaces for repository layer
- Handler tests follow the pattern: create request → call handler → assert response code + body
- Test files live in same package (`internal/auth/`, `internal/workspace/`) using `_test.go` suffix
- Prior art: no existing tests in codebase (first test suite)

## Out of Scope
- Drag to reorder nested pages (only root-level pages in sidebar)
- Emoji search or categories in icon picker
- Email notifications for invites
- Two-factor authentication
- Workspace-level roles beyond member/owner
- Bulk delete from trash
- Search within specific page (Ctrl+F in-page)
- Page templates

## Further Notes
- All features use existing backend patterns (JSONB in block.content for page metadata)
- Only Trash and Search require database migrations
- Frontend features are Svelte 5 runes throughout
