# Issue 7: Full-text search frontend

**Status:** pending
**Dependencies:** Issue 6 (Search Backend)
**Estimate:** Medium

## What to build

A `/search?q=...` full-page route that displays block-level search results from the backend search endpoint. Includes the search input at the top and a list of results below with page title + excerpt snippets.

## Acceptance Criteria

- [ ] `/search?q=...` route renders a search page
- [ ] Search input at the top (auto-focused, pre-filled from URL param)
- [ ] On input change, debounced (300ms) API call to search endpoint
- [ ] Results show: page title (linked to `/pages/{pageId}`), block type indicator, excerpt snippet (150 chars with HTML stripped)
- [ ] Empty state: "No results for '{query}'" or "Search across all pages" when no query
- [ ] Loading state: spinner
- [ ] Error state: error message with retry
- [ ] Sidebar search input: on submit (Enter), navigates to `/search?q=...` instead of client-side filtering
- [ ] Sidebar quick filter (client-side by title) still works as before for instant filtering
- [ ] Build passes: `pnpm build`
