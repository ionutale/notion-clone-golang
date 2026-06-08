# Authentication & Workspaces Design

## Overview

Add user authentication (signup/login/logout) and multi-workspace support to the Notion clone. Users can create and join multiple workspaces, switch between them, and all page/block data is scoped to a workspace.

## Auth Approach

- **JWT + refresh token rotation**
- Access token (15min) — stored in Svelte store (JS memory, not localStorage)
- Refresh token (7d) — bcrypt hash in DB, set as HTTP-only cookie named `refresh_token`
- On 401 response, client calls `/auth/refresh` which reads the cookie, validates, rotates tokens
- On page load, client calls `/auth/refresh` to restore the session
- Logout revokes refresh token in DB and clears the cookie

## Backend Structure

### New packages
- `internal/auth/` — handler, service, repository for users + tokens
- `internal/workspace/` — handler, service, repository for workspaces

### Auth endpoints
- `POST /api/v1/auth/signup` — creates user + default personal workspace, sets refresh cookie, returns access token + user
- `POST /api/v1/auth/login` — validates email+password, sets refresh cookie, returns access token + user
- `POST /api/v1/auth/refresh` — reads cookie, validates, rotates tokens, returns new access token
- `POST /api/v1/auth/logout` — revokes refresh token, clears cookie
- `GET /api/v1/auth/me` — returns current user info (protected)

### Workspace endpoints
- `GET /api/v1/workspaces` — list user's workspaces
- `POST /api/v1/workspaces` — create workspace
- `PATCH /api/v1/workspaces/:id` — update name
- `DELETE /api/v1/workspaces/:id` — soft delete (owner only)
- `POST /api/v1/workspaces/:id/members` — invite by email
- `DELETE /api/v1/workspaces/:id/members/:userId` — remove member

### Existing endpoints (re-scoped)
Existing page/block CRUD moves under workspace prefix: `/api/v1/workspaces/:workspaceId/pages`, `/api/v1/workspaces/:workspaceId/blocks`

### Middleware
- **Auth middleware** — extracts JWT from `Authorization: Bearer <token>`, validates, injects `user_id` into context
- **Workspace middleware** — checks that requesting user is a member of `:workspaceId`

### Database tables (migration 000003)
- `users` — id UUID, email TEXT UNIQUE, password_hash TEXT, name TEXT, created_at TIMESTAMPTZ
- `refresh_tokens` — id UUID, user_id UUID FK, token_hash TEXT, expires_at TIMESTAMPTZ, created_at TIMESTAMPTZ
- `workspaces` — id UUID, name TEXT, owner_id UUID FK, created_at TIMESTAMPTZ
- `workspace_members` — workspace_id UUID FK, user_id UUID FK, role TEXT (owner/admin/member), joined_at TIMESTAMPTZ, PRIMARY KEY (workspace_id, user_id)

### Libraries
- `github.com/golang-jwt/jwt/v5` — JWT signing/verification (HS256)
- `golang.org/x/crypto/bcrypt` — password hashing, refresh token hashing

## Frontend

### Auth store (`web/src/lib/stores/auth.svelte.ts`)
- `user` — current user object or null
- `accessToken` — current access token or null
- `login(email, password)` — POST /auth/login, store token and user
- `signup(email, password, name)` — POST /auth/signup, store token and user
- `logout()` — POST /auth/logout, clear state
- `refresh()` — POST /auth/refresh, update access token
- `check()` — called on app mount, tries to restore session

### Active workspace store (`web/src/lib/stores/workspace.svelte.ts`)
- `workspaces` — list of workspaces the user belongs to
- `activeWorkspaceId` — currently selected workspace
- `switchWorkspace(id)` — changes active workspace, reloads page tree
- `createWorkspace(name)` — creates new workspace

### API client changes
- Add `Authorization: Bearer <token>` header to all requests when token exists
- Auto-refresh: on 401 response, call `/auth/refresh`, retry original request once
- If refresh also fails, redirect to login

### Pages
- `web/src/routes/login/+page.svelte` — login form (email, password)
- `web/src/routes/signup/+page.svelte` — signup form (email, password, name)
- `web/src/routes/+layout.svelte` — wrap in auth gate, redirect to /login if unauthenticated

### Sidebar updates
- Workspace switcher dropdown at the top
- Page list scoped to active workspace
- Create workspace button

## Out of Scope
- OAuth / social login
- Email verification
- Password reset flow
- Workspace settings page
- Role-based permissions (admin vs member distinction reserved for future)
