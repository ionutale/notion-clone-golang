-- Add password_hash to existing users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash TEXT NOT NULL DEFAULT '';

-- Add owner_id to existing workspaces table
ALTER TABLE workspaces ADD COLUMN IF NOT EXISTS owner_id UUID REFERENCES users(id);

-- New table for refresh tokens
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);

-- Drop and recreate workspace_members with updated schema (CHECK constraint)
ALTER TABLE workspace_members DROP CONSTRAINT IF EXISTS workspace_members_role_check;
ALTER TABLE workspace_members ADD CONSTRAINT workspace_members_role_check CHECK (role IN ('owner', 'admin', 'member'));

CREATE INDEX IF NOT EXISTS idx_workspace_members_user_id ON workspace_members(user_id);
