# Notion Clone

A full-stack Notion-like note-taking application built with **Go** (Chi router, pgx/PostgreSQL) and **SvelteKit 5** (Runes, Tailwind, DaisyUI).

![Full page demo](docs/screenshots/10-full-page.png)

## Features

### Editor
- Rich text blocks: headings, paragraphs, bullet lists, numbered lists, toggles, dividers
- Inline formatting: bold (`Ctrl+B`), italic (`Ctrl+I`), underline (`Ctrl+U`), strikethrough
- Slash menu (`/`) for quick block type switching
- Page icons with emoji picker
- Drag-and-drop block reordering
- Split and merge blocks
- Cover images for pages

### Organization
- Multi-workspace support with member invitations
- Sidebar with page list and favorites
- Full-text search across all pages
- Trash with restore and permanent delete

### Authentication
- Email/password signup and login
- JWT access tokens (15 min) + refresh token rotation (7 days)
- Profile management and password change
- Account deletion

### Architecture

```
┌─────────────┐     ┌──────────────┐     ┌────────────┐
│  SvelteKit  │────▶│  Chi Router  │────▶│ PostgreSQL │
│  (SPA)      │     │  (Go binary) │     │            │
└─────────────┘     └──────────────┘     └────────────┘
                           │
                    ┌──────┴──────┐
                    │  Middleware  │
                    │  Auth / WS   │
                    │  CORS / Gzip │
                    │  Security    │
                    └─────────────┘
```

**Backend:** Go 1.26, Chi v5, pgx v5, JWT (HS256), bcrypt, PostgreSQL (ltree, full-text search, JSONB)

**Frontend:** SvelteKit 5 (Runes), TypeScript, Tailwind CSS, DaisyUI

## Screenshots

| Signup | Empty Home | New Page |
|--------|-----------|----------|
| ![Signup](docs/screenshots/01-signup.png) | ![Home](docs/screenshots/02-home-empty.png) | ![Empty page](docs/screenshots/03-empty-page.png) |

| Page Icon | Heading & Formatting | Bullet List |
|-----------|---------------------|-------------|
| ![Page icon](docs/screenshots/04-page-with-icon.png) | ![Formatting](docs/screenshots/05-heading.png) | ![List](docs/screenshots/07-bullet-list.png) |

| Numbered List | Toggle Block | Full Page |
|--------------|-------------|-----------|
| ![Numbered list](docs/screenshots/08-numbered-list.png) | ![Toggle](docs/screenshots/09-toggle-block.png) | ![Full page](docs/screenshots/10-full-page.png) |

## Quick Start

### Prerequisites

- Docker & Docker Compose
- Go 1.26+ (for local development)
- pnpm (for frontend development)

### Run with Docker Compose

```bash
# Start all services
docker compose up -d --build

# Open in browser
open http://localhost:8080
```

The application will be available at `http://localhost:8080`. Sign up for a new account to get started.

### Local Development

```bash
# Start database
docker compose up -d db

# Start frontend dev server
cd web && pnpm install && pnpm dev &

# Start Go backend
DATABASE_URL="postgres://notion:notion@localhost:5434/notion?sslmode=disable" \
DEV_MODE=true \
JWT_SECRET="dev-secret" \
go run .
```

Frontend dev server runs on `http://localhost:5173`, API on `http://localhost:8080`.

### Configuration

Environment variables are documented in [`.env.example`](.env.example):

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | — | PostgreSQL connection string |
| `PORT` | `8080` | HTTP server port |
| `DEV_MODE` | `false` | Enable Vite dev proxy + dev CORS |
| `JWT_SECRET` | `dev-secret-change-in-production` | Token signing key |
| `CORS_ORIGINS` | `http://localhost:5173 http://localhost:8080` | Allowed origins (dev only) |
| `APP_URL` | — | Production app URL (sets CORS origin) |
| `STORAGE_EMULATOR_HOST` | — | GCS emulator host (unused, file storage is local) |

## API Overview

All API endpoints are under `/api/v1/`.

### Authentication

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/auth/signup` | Create account |
| `POST` | `/auth/login` | Log in |
| `POST` | `/auth/refresh` | Refresh access token |
| `POST` | `/auth/logout` | Log out |
| `GET` | `/auth/me` | Get current user |
| `PATCH` | `/auth/me` | Update profile |
| `PATCH` | `/auth/me/password` | Change password |
| `DELETE` | `/auth/me` | Delete account |

### Workspaces

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/workspaces` | List workspaces |
| `POST` | `/workspaces` | Create workspace |
| `GET` | `/workspaces/{id}` | Get workspace |
| `PATCH` | `/workspaces/{id}` | Update workspace |
| `DELETE` | `/workspaces/{id}` | Delete workspace |
| `GET` | `/workspaces/{id}/members` | List members |
| `POST` | `/workspaces/{id}/members` | Invite member |
| `DELETE` | `/workspaces/{id}/members/{userId}` | Remove member |

### Pages & Blocks

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/workspaces/{id}/pages` | List pages (cursor pagination) |
| `POST` | `/workspaces/{id}/pages` | Create page |
| `GET` | `/workspaces/{id}/pages/{pageId}` | Get page tree |
| `POST` | `/workspaces/{id}/blocks` | Create block |
| `PATCH` | `/workspaces/{id}/blocks/{blockId}` | Update block |
| `DELETE` | `/workspaces/{id}/blocks/{blockId}` | Soft delete block |
| `PATCH` | `/workspaces/{id}/blocks/{blockId}/restore` | Restore block |
| `PATCH` | `/workspaces/{id}/blocks/{blockId}/move` | Move block |
| `DELETE` | `/workspaces/{id}/blocks/{blockId}/permanent` | Permanently delete |
| `GET` | `/workspaces/{id}/search` | Full-text search |
| `GET` | `/workspaces/{id}/favorites` | List favorites |
| `GET` | `/workspaces/{id}/trash` | List trash |

## Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl+B` | Bold |
| `Ctrl+I` | Italic |
| `Ctrl+U` | Underline |
| `Ctrl+Shift+S` | Strikethrough |
| `Tab` / `Shift+Tab` | Indent / Outdent |
| `Enter` | New block |
| `Backspace` on empty | Delete block |
| `/` | Open slash menu |
| `Alt+↑` / `Alt+↓` | Move block up/down |

## Project Structure

```
├── main.go                 # Entry point, router, server setup
├── internal/
│   ├── api.go              # Route mounting
│   ├── auth/               # Auth module (handler, service, repository, JWT)
│   ├── block/              # Blocks/pages module
│   ├── config/             # Environment configuration
│   ├── db/                 # Database connection
│   ├── httputil/           # Shared HTTP utilities
│   ├── middleware/         # Auth, workspace, security, recovery middleware
│   ├── storage/            # File storage abstraction
│   └── workspace/          # Workspace module
├── migrations/             # SQL migrations
├── web/                    # SvelteKit frontend
│   ├── src/
│   │   ├── lib/            # Components, stores, API client, types
│   │   └── routes/         # Pages (login, signup, editor, search, etc.)
│   └── tests/              # Playwright E2E tests
└── docs/
    └── screenshots/        # Demo screenshots
```

## Running Tests

```bash
# Go backend tests
go test ./...

# Frontend E2E tests (requires running app)
cd web && npx playwright test
```

## License

MIT
