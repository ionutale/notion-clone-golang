ALTER TABLE blocks ADD COLUMN IF NOT EXISTS search_vector tsvector
  GENERATED ALWAYS AS (
    to_tsvector('english', coalesce(content->>'title', '')) ||
    to_tsvector('english', coalesce(regexp_replace(coalesce(content->>'html', ''), '<[^>]+>', '', 'g'), ''))
  ) STORED;

CREATE INDEX IF NOT EXISTS idx_blocks_search ON blocks USING GIN (search_vector);
