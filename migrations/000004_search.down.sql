DROP INDEX IF EXISTS idx_blocks_search;
ALTER TABLE blocks DROP COLUMN IF EXISTS search_vector;
