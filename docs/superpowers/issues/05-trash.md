# Issue 5: Trash / recently deleted

**Status:** pending
**Dependencies:** None — can start immediately
**Estimate:** Medium

## What to build

A `/trash` route listing soft-deleted pages with restore and permanent delete actions. Auto-expire: trashed pages older than a configurable threshold (default 30 days) are permanently deleted on list load.

Existing toast undo for inline deletion still works — trash view supplements it.

## Acceptance Criteria

**Backend:**
- [ ] New migration: add `auto_expire_days` to `users` table (nullable, default 30), or store in user preferences JSON
- [ ] New endpoint: `GET /workspaces/{workspaceId}/trash` — returns soft-deleted pages (`deleted_at IS NOT NULL`, `type='page'`)
- [ ] On list, permanently delete pages where `deleted_at < now() - interval 'N days'`
- [ ] New endpoint: `DELETE /blocks/{id}/permanent` — hard-deletes a block from the database
- [ ] Existing `PATCH /blocks/{id}/restore` preserved

**Frontend:**
- [ ] `/trash` route showing list of deleted pages with deleted_at timestamp
- [ ] Each item has: title, deleted date, Restore button, Delete Forever button
- [ ] Restore calls `blockStore.restoreBlock(id)` and removes from list
- [ ] Delete Forever calls new permanent delete endpoint and removes from list
- [ ] Empty state: "No deleted pages"
- [ ] Build passes: `go build ./...` + `pnpm build`
