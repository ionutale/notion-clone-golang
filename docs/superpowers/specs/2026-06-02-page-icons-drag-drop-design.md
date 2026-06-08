# Page Icons & Drag & Drop

## Overview

Two independent features: emoji/image page icons, and block drag-and-drop reordering.

---

## Page Icons

### Storage

Store in `block.content` JSONB (no schema migration):

```json
{
  "title": "My Page",
  "icon": "🚀",
  "icon_type": "emoji"
}
```

`icon_type` is `"emoji"`, `"image"`, or `null` (no icon).

### Backend

No new handler or endpoint. Icon is set via existing `PATCH /api/v1/blocks/{id}` with `content: { title, icon, icon_type }`. Loaded and returned as part of existing block queries.

### Editor Icon Area

Above the page `<h1>` title in `Editor.svelte`:

- If `icon` is set, render a 48x48px element (emoji as text, image as `<img>`)
- Clicking the icon area opens a popover
- Popover has two tabs: **Emoji** and **Image**
- Popover has a "Remove" button to clear the icon
- Click outside closes popover

### Emoji Picker

Compact grid of ~42 common emojis, rendered as a 6-column grid. Categories excluded — single flat set. Selected emoji sets `content.icon` and `content.icon_type = "emoji"`.

### Image Upload Tab

Two options in the Image tab:
1. **Upload** — hidden `<input type="file" accept="image/*">` triggering `api.uploadFile()`
2. **URL** — text input for pasting an image URL

On success, sets `content.icon` to the URL and `content.icon_type = "image"`.

### Sidebar Icon

In `Sidebar.svelte`, render a small icon before each page title:
- Emoji: 16px text
- Image: 24x24px `<img>` with object-fit cover and rounded

---

## Drag & Drop Reordering

### Backend

No changes needed. Uses existing `PATCH /api/v1/workspaces/{workspaceId}/blocks/{id}/move` with `{ parent_id, position }`.

### Frontend — Drag Handle

The existing `BlockDragHandle.svelte` component is enhanced:

- Visible on hover (desktop)
- Grip icon (6 dots / ≡)
- `draggable="true"` set on the block wrapper
- Touch: long-press (500ms) enters drag mode, then finger drag moves the block
- `dragstart`: sets `dataTransfer.effectAllowed = 'move'`, serializes block data
- `dragover`: shows drop indicator line between blocks
- `drop`: calls `blockStore.moveBlock(id, newParentId, newPosition)`
- `dragend`: cleans up visual indicators

### Drop Zones

Each block acts as a drop target. The drop indicator is a horizontal line between blocks. Blocks always drop as siblings — nesting via drag is out of scope.

### Touch Support

- Long-press (500ms) on the drag handle activates drag mode
- Haptic feedback via `navigator.vibrate(10)` if available
- The block follows the finger using `touchmove`
- Drop on `touchend`
- Visual ghost overlay while dragging

### Visual Feedback

- Dragged block has reduced opacity + slight rotation
- Drop zone shows a blue accent line at the insertion point
- Smooth CSS transitions on position change after drop

---

## Files Changed

### Page Icons
- `web/src/lib/components/Editor.svelte` — icon display + click-to-open popover
- `web/src/lib/components/IconPopover.svelte` (new) — emoji grid / image upload / remove
- `web/src/lib/components/Sidebar.svelte` — small icon before page title
- `web/src/lib/components/blocks/PageBlock.svelte` — icon in nested page blocks

### Drag & Drop
- `web/src/lib/components/BlockDragHandle.svelte` — enhanced drag logic
- `web/src/lib/components/BlockRenderer.svelte` — drag event handlers, drop zones
- `web/src/lib/stores/blocks.svelte.ts` — moveBlock method (may already exist)

---

## Out of Scope

- Folder / nested page reordering
- Multi-select + bulk move
- Drag across columns (no columns exist)
- Emoji search or categories
