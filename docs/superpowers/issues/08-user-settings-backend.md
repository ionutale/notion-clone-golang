# Issue 8: User settings backend

**Status:** pending
**Dependencies:** None — can start immediately
**Estimate:** Small

## What to build

Three new auth handler endpoints for updating user profile and deleting account.

## Acceptance Criteria

- [ ] `PATCH /auth/me` — updates name and/or email. Requires current password for email change. Returns updated user.
  ```go
  type UpdateProfileRequest struct {
    Name            string `json:"name,omitempty"`
    Email           string `json:"email,omitempty"`
    CurrentPassword string `json:"current_password,omitempty"` // required if email changes
  }
  ```
- [ ] `PATCH /auth/me/password` — changes password. Requires current password + new password + confirm.
  ```go
  type UpdatePasswordRequest struct {
    CurrentPassword string `json:"current_password"`
    NewPassword     string `json:"new_password"`
  }
  ```
- [ ] `DELETE /auth/me` — deletes the authenticated user's account. Requires password confirmation.
  ```go
  type DeleteAccountRequest struct {
    Password string `json:"password"`
  }
  ```
- [ ] Deleting account also: removes user as workspace member from all workspaces, or transfers ownership if last owner
- [ ] Validation: email format, password strength (min 8 chars), current password match
- [ ] Error responses: proper HTTP status codes (400, 401, 404, 409)
- [ ] Build passes: `go build ./...`
