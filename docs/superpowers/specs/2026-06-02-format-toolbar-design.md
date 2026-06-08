# Rich Text Formatting Toolbar

## Overview

Add a persistent formatting toolbar between the page title and the block list in the editor, providing one-click access to inline formatting and block-level actions.

## Position

- Sits between the page title (`<h1>`) and the block list in `Editor.svelte`
- Sticky within the editor column (scrolls with content, not fixed to viewport)
- Always visible — no focus/selection requirement to show/hide

## Visual Style

- Dark-themed pill bar matching the mockup (`bg-base-200/90` with subtle shadow)
- Buttons use icons or short labels, with hover and active states
- Active button highlighted when cursor is inside formatted text

## Buttons

### Row 1 — Inline Formatting
- **Bold** — toggles `<b>` via `document.execCommand('bold')`
- **Italic** — toggles `<i>` via `document.execCommand('italic')`
- **Underline** — toggles `<u>` via `document.execCommand('underline')`
- **Strikethrough** — toggles `<s>` via `document.execCommand('strikeThrough')`
- **Inline Code** — wraps selection in `<code>` via custom range manipulation

### Row 2 — Block / Insert Actions
- **Heading** — dropdown with H1/H2/H3 options, transforms current block type
- **Link** — prompts for URL, wraps selection in `<a href="...">` via `document.execCommand('createLink')`
- **Clear Formatting** — removes all formatting via `document.execCommand('removeFormat')`

## Active State Detection

- Listen for `selectionchange` on the document
- Call `document.queryCommandState('bold')`, `.queryCommandState('italic')`, etc.
- Update an `$state` map in the toolbar component
- Buttons render with active/inactive CSS class based on this map

## Integration

- Single new component: `web/src/lib/components/FormatToolbar.svelte`
- Inserted into `Editor.svelte` between the page title and the blocks `{#each}`
- Operates on whichever contenteditable block has focus (works universally)
- No changes needed to individual block components or the block store

## Files Modified
- `web/src/lib/components/Editor.svelte` — add FormatToolbar between title and blocks
- `web/src/lib/components/FormatToolbar.svelte` — new component

## Out of Scope
- Floating toolbar on text selection (overridden by sticky choice)
- Color picker / highlight (could be future addition)
- Font family / size controls
