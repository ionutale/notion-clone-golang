DROP TABLE IF EXISTS refresh_tokens;
ALTER TABLE users DROP COLUMN IF EXISTS password_hash;
ALTER TABLE workspaces DROP COLUMN IF EXISTS owner_id;
ALTER TABLE workspace_members DROP CONSTRAINT IF EXISTS workspace_members_role_check;
DROP INDEX IF EXISTS idx_workspace_members_user_id;
