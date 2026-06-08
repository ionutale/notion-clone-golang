# Issue 1: Backend tests for auth + workspace handlers

**Status:** pending
**Dependencies:** None — can start immediately
**Estimate:** Medium

## What to build

Add test coverage for auth (signup, login, refresh, logout, me) and workspace (create, list, rename, invite member) handlers using mock repository interfaces.

## Acceptance Criteria

- [ ] `testify` is added as a dependency
- [ ] Auth handler tests cover: signup success, duplicate email, missing fields; login success, wrong password, nonexistent email; refresh with valid/expired token; logout; me with auth header and without
- [ ] Workspace handler tests cover: create success, list, rename, invite member success, invite duplicate, invite nonexistent user
- [ ] All tests pass with `go test ./internal/auth/ ./internal/workspace/ -v`
- [ ] No database dependency — repository layer is mocked
- [ ] Tests follow pattern: create http request → call handler → assert response code + body
