# Issue 6: Full-text search backend

**Status:** pending
**Dependencies:** None — can start immediately
**Estimate:** Medium

## What to build

PostgreSQL full-text search backend for searching across all block content. New migration, new endpoint, block-level results with rank ordering.

## Acceptance Criteria

- [ ] New migration `000004_search.up.sql`:
  ```sql
  ALTER TABLE blocks ADD COLUMN search_vector tsvector
    GENERATED ALWAYS AS (
      to_tsvector('english', coalesce(content->>'title', '')) ||
      to_tsvector('english', coalesce(regexp_replace(coalesce(content->>'html', ''), '<[^>]+>', '', 'g'), ''))
    ) STORED;
  CREATE INDEX idx_blocks_search ON blocks USING GIN (search_vector);
  ```
- [ ] `000004_search.down.sql` drops the column and index
- [ ] New repository method `Search(ctx, workspaceID, query string, limit, offset)` returning:
  ```go
  type SearchResult struct {
    BlockID   uuid.UUID `json:"block_id"`
    PageID    uuid.UUID `json:"page_id"`
    PageTitle string    `json:"page_title"`
    BlockType string    `json:"block_type"`
    Excerpt   string    `json:"excerpt"`
    Rank      float64   `json:"rank"`
  }
  ```
- [ ] SQL query uses `plainto_tsquery('english', $3)` for simple query parsing, joins to parent page via recursive CTE or ltree path, limits to `workspace_id`, filters out `deleted_at IS NOT NULL`
- [ ] Excerpt extracts first ~150 chars of matching block content, HTML tags stripped
- [ ] New endpoint: `GET /workspaces/{workspaceId}/search?q=...&limit=20&offset=0`
- [ ] New endpoint: `GET /favorites` — returns favorited pages (needed by Issue 3)
- [ ] Migrations run with existing migration runner
- [ ] Build passes: `go build ./...`
