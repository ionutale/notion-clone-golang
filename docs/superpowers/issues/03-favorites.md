# Issue 3: Favorites (pinning pages)

**Status:** pending
**Dependencies:** None — can start immediately
**Estimate:** Small

## What to build

Users can favorite/pin pages. Favorited pages appear in a "Favorites" section at the top of the sidebar. Toggle via star icon on hover in sidebar or at top of the Editor.

Storage: `favorited: true` in page `block.content` JSONB — no migration.

Backend: new `GET /workspaces/{workspaceId}/favorites` endpoint filtering `(content->>'favorited')::boolean = true`.

## Acceptance Criteria

- [ ] Backend: `GET /workspaces/{workspaceId}/favorites` returns list of favorited pages
- [ ] Sidebar: "Favorites" section at the top, above search, showing favorited pages
- [ ] Star icon on hover in sidebar toggles favorite — optimistic UI update
- [ ] Star icon in Editor (near the page icon area) toggles favorite
- [ ] Toggle calls `PATCH /blocks/{id}` with `content: { ..., favorited: true/false }`
- [ ] Build passes: `go build ./...` + `pnpm build`
