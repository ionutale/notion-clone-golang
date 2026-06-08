# Issue 4: Page cover images

**Status:** pending
**Dependencies:** None — can start immediately
**Estimate:** Small

## What to build

Users can add a banner cover image to pages. Cover appears as a fixed 200px banner at the top of the Editor, before the page icon. Click to change/remove — popover with image upload + URL input (reuse pattern from `IconPopover`).

Storage: `cover`, `cover_type`, `cover_color` in `block.content` JSONB — no backend changes. Uses existing `PATCH /blocks/{id}`.

## Acceptance Criteria

- [ ] Editor: 200px fixed-height cover banner at the top, before the page icon
- [ ] Cover supports image (upload or URL) and solid color fallback
- [ ] Click to change/remove opens a popover with Upload + URL + color picker + Remove
- [ ] Cover stored via `blockStore.updateBlock(id, { content: { ..., cover, cover_type, cover_color } })`
- [ ] No cover = no banner shown (compact layout preserved)
- [ ] Image covers use `object-cover`; color covers use inline background-color
- [ ] Build passes: `pnpm build`
