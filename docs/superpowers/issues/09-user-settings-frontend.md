# Issue 9: User settings frontend

**Status:** pending
**Dependencies:** Issue 8 (User Settings Backend)
**Estimate:** Small

## What to build

A `/settings` route with profile management: change name, email, password, and delete account.

## Acceptance Criteria

- [ ] `/settings` route with three sections:
  - **Profile**: name input, email input, Save button (shows current password dialog if email changed)
  - **Password**: current password, new password, confirm new password, Change button
  - **Danger Zone**: Delete Account button → confirmation modal (type email to confirm) → deletes
- [ ] All forms validate: email format, password min 8 chars, passwords match
- [ ] Success feedback (toast or inline message) on save
- [ ] Error feedback for API errors
- [ ] Loading/disabled state on submit buttons while saving
- [ ] Delete account redirects to `/login` after success
- [ ] Build passes: `pnpm build`
