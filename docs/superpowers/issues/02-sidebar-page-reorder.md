# Issue 2: Sidebar page reordering (drag & drop)

**Status:** pending
**Dependencies:** None — can start immediately
**Estimate:** Small

## What to build

Drag-and-drop reordering of pages in the sidebar. Each page list item gets a grip handle on hover, `draggable="true"`, drop zones between items, and optimistic reorder + persistence via `PATCH /blocks/{id}/move`.

Touch: 500ms long-press activates drag mode, same as block DnD in the editor.

## Acceptance Criteria

- [ ] Each page list item shows a 6-dot grip handle on hover (same SVG as `BlockDragHandle`)
- [ ] Desktop: `draggable="true"`, drag visual (opacity-40), drop indicator line (blue accent), drop calls `blockStore.moveBlock(id, null, newPos)`
- [ ] Touch: 500ms long-press activates drag, finger-drag shows drop indicator, `touchend` persists new position
- [ ] Search mode disables drag (filtered order ≠ real order)
- [ ] Edge cases: drag to first position, drag to last position, drag cancellation via `dragend`
- [ ] Local `pages` array optimistically reordered before API response
- [ ] Build passes: `pnpm build`
