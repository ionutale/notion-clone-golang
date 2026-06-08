# Issue 10: Workspace invite UI

**Status:** pending
**Dependencies:** None — can start immediately (backend exists)
**Estimate:** Small

## What to build

A workspace settings UI for inviting members by email. The backend `POST /workspaces/{workspaceId}/members` already exists. The frontend needs a settings panel with an email input and invite button, plus a list of current members.

## Acceptance Criteria

- [ ] `GET /workspaces/{workspaceId}/members` endpoint or check if it exists; if not, add it
- [ ] Workspace settings panel accessible from sidebar (gear icon next to workspace name or in workspace switcher dropdown)
- [ ] Invite section: email input + "Invite" button
- [ ] On success: member appears in member list, input clears
- [ ] On error (user not found, already a member): show error inline
- [ ] Member list shows: name, email, role (owner/member)
- [ ] Only workspace owner can invite members
- [ ] Build passes: `go build ./...` + `pnpm build`
