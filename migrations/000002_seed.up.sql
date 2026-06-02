INSERT INTO workspaces (id, name) VALUES
  ('00000000-0000-0000-0000-000000000001', 'My Workspace')
ON CONFLICT (id) DO NOTHING;

INSERT INTO users (id, email, name) VALUES
  ('00000000-0000-0000-0000-000000000002', 'dev@notion-clone.local', 'Dev User')
ON CONFLICT (id) DO NOTHING;

INSERT INTO workspace_members (workspace_id, user_id, role) VALUES
  ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000002', 'owner')
ON CONFLICT (workspace_id, user_id) DO NOTHING;
