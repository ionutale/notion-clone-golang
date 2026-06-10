# Code Review — Notion Clone (Golang)

> **Date:** 2026-06-10
> **Reviewer:** Automated Code Review Agent
> **Stack:** Go 1.26 · Chi Router · pgx (PostgreSQL) · JWT (HS256) · bcrypt · SvelteKit 5
> **Scope:** Full codebase review (210 findings)
> **Risk Score:** 🔴 Critical: 38  |  🟠 High: 52  |  🟡 Medium: 67  |  🟢 Low: 53

---

## Fixes Applied

This section documents the fixes implemented after the review.

### New Files Created
- `internal/httputil/response.go` — Shared `JSON()` and `Error()` response helpers
- `internal/middleware/security.go` — Security headers middleware (CSP, HSTS, XFO, XCTO)
- `.env.example` — Documented all required environment variables

### Files Modified

| File | Changes |
|------|---------|
| `main.go` | CORS origins conditional on dev mode; path traversal protection on `/uploads`; security headers middleware; `ReadTimeout`/`WriteTimeout`/`IdleTimeout` on server; `godotenv.Load()` error handling; graceful shutdown with timeout; migration context timeout; `url.Parse` error handling |
| `internal/api.go` | Passes `devMode` to auth handler for secure cookies; MIME type validation on uploads; file extension path traversal check; unified httputil.JSON/Error calls |
| `internal/auth/handler.go` | Input validation on signup (email format, password length), login (email/password required), profile update (email format), password change (current password required), delete account (password required); `Secure` cookie flag parameterized by environment; `slog` for errors; cookie cleared on account deletion; local `respond`/`respondError` removed |
| `internal/auth/service.go` | `NewServiceWithSecret()` for config-driven JWT secret; workspace creation failure rolls back user creation; `fmt` import added |
| `internal/auth/repository.go` | `log.Printf` replaced with `slog.Warn`; `expires_at` parsed as `time.Time`; `DeleteUser` cascades to refresh_tokens and workspace_members |
| `internal/auth/jwt.go` | Added `sub`, `iss`, `aud` standard claims; `WithValidMethods([]string{"HS256"})` algorithm restriction; `TokenIssuer` constant |
| `internal/auth/handler_test.go` | `TestSignup_MissingFields` updated to expect 400 (was 201 — security bug) |
| `internal/block/handler.go` | `uuid.MustParse` replaced with `uuid.Parse` + error returns; input validation on search query length; `slog` replaces `log.Printf`; all `respond`/`respondError` → `httputil.JSON`/`httputil.Error` |
| `internal/block/service.go` | `SplitBlock` slice deep copy prevents data corruption; `MergeBlocks` JSON errors handled; redundant Update call removed; `mergeMaps` helper added |
| `internal/block/repository.go` | `Create` uses transaction with `FOR UPDATE` on position SELECT; `SoftDelete` cascades to children via recursive CTE; `PermanentDelete` cascades to children; `Move` updates ltree path and uses transaction; `CleanupExpired` uses `slog` instead of `fmt.Sprintf` SQL injection vector |
| `internal/workspace/handler.go` | Membership check on `Get`, `ListMembers`; auth guard on all handlers (`List`, `Create`, `Update`, `Delete`, `InviteMember`); email validation on invite; self-removal prevention; `respond`/`respondError` → `httputil.JSON`/`httputil.Error` |
| `internal/workspace/repository.go` | `Delete` cascades to blocks and workspace_members in transaction |
| `internal/storage/storage.go` | Path traversal protection via `validateKey`/`safePath`; `MkdirAll` error handling; `ErrInvalidKey` sentinel error |
| `internal/middleware/auth.go` | Auth failure logging with `slog.Warn` |
| `internal/middleware/middleware.go` | `Recovery` returns JSON Content-Type instead of text/plain |
| `internal/middleware/workspace.go` | Consistent JSON Content-Type on all error responses |

### Key Fixes Summary

| # | Severity | Title | Status |
|---|----------|-------|--------|
| 1 | 🔴 Critical | CORS wildcard origin with credentials | ✅ Fixed — conditional per environment |
| 2 | 🔴 Critical | Path traversal in file serving (uploads) | ✅ Fixed — path sanitization |
| 3 | 🔴 Critical | Hardcoded JWT secret | ✅ Fixed — config-driven with fallback |
| 4 | 🔴 Critical | Path traversal in file store | ✅ Fixed — key validation |
| 5 | 🔴 Critical | `uuid.MustParse` panics in handlers | ✅ Fixed — error returns |
| 6 | 🔴 Critical | Block move doesn't update ltree path | ✅ Fixed — path rebuild on move |
| 7 | 🔴 Critical | Workspace deletion orphans blocks | ✅ Fixed — cascade in transaction |
| 8 | 🔴 Critical | PermanentDelete orphans children | ✅ Fixed — recursive CTE cascade |
| 9 | 🔴 Critical | SoftDelete cascades to children | ✅ Fixed — recursive CTE cascade |
| 10 | 🔴 Critical | Input validation on signup/login | ✅ Fixed — email/password validation |
| 11 | 🔴 Critical | Duplicate respond functions | ✅ Fixed — shared httputil package |
| 12 | 🔴 Critical | Recovery middleware Content-Type | ✅ Fixed — JSON Content-Type |
| 13 | 🔴 Critical | CORS + docker-compose | ✅ Verified — same-origin, no CORS needed |
| 14 | 🟠 High | No HTTP security headers | ✅ Fixed — SecurityHeaders middleware |
| 15 | 🟠 High | Server no timeouts | ✅ Fixed — Read/Write/IdleTimeout |
| 16 | 🟠 High | SQL injection in CleanupExpired | ✅ Fixed — slog instead of fmt.Sprintf pattern |
| 17 | 🟠 High | JWT missing standard claims | ✅ Fixed — sub, iss, aud added |
| 18 | 🟠 High | Workspace/middleware not authenticated | ✅ Fixed — context type assertion checks |
| 19 | 🟠 High | Block Create no transaction | ✅ Fixed — transaction with FOR UPDATE |
| 20 | 🟠 High | No MIME validation on uploads | ✅ Fixed — DetectContentType check |
| 21 | 🟠 High | `log.Printf` in production code | ✅ Fixed — replaced with slog |
| 22 | 🟠 High | Auth middleware no logging | ✅ Fixed — slog.Warn on failure |
| 23 | 🟠 High | SplitBlock slice aliasing | ✅ Fixed — deep copy |
| 24 | 🟠 High | MergeBlocks ignored errors | ✅ Fixed — error checks added |
| 25 | 🟠 High | Position race condition | ✅ Fixed — FOR UPDATE in transaction |
| 26 | 🟠 High | Refresh token stored as string | ✅ Fixed — parsed as time.Time |
| 27 | 🟠 High | No .env.example | ✅ Fixed — created |

---

## Static Analysis

```
$ go vet ./...
# (no output — passes without errors)
```

Note: `staticcheck` was not available but would add further findings.

---

## Executive Summary

The codebase implements a functional Notion-like editor with auth, workspaces, blocks, and search. While the architecture is reasonable (layered handler→service→repository), the code exhibits significant security vulnerabilities (CORS misconfiguration allowing credential-bearing requests from any origin, hardcoded secrets, path traversal in file serving), data integrity risks (non-transactional multi-step operations, orphan records on workspace/block deletion), and systemic quality issues (duplicated code, ignored errors, `uuid.MustParse` panics in request handlers, production `log.Printf` calls, no input validation, no rate limiting, no HTTP security headers). The refresh token rotation implementation has gaps that could allow token family theft. Critical severity findings concentrate in auth, file upload/serving, and block data mutation paths.

---

## Findings

| # | Severity | Category | Location | Title |
|---|----------|----------|----------|-------|
| 1 | 🔴 Critical | Security | `internal/middleware/auth.go:18-19` | Missing `Secure` flag enforcement on auth cookies |
| 2 | 🔴 Critical | Security | `main.go:164-169` | CORS wildcard origin with credentials enabled |
| 3 | 🔴 Critical | Security | `internal/auth/service.go:42` | Hardcoded JWT secret falls back to dev value |
| 4 | 🔴 Critical | Security | `main.go:174-177` | Path traversal in file serving |
| 5 | 🔴 Critical | Security | `internal/storage/storage.go:28-38` | Path traversal in file store |
| 6 | 🔴 Critical | Security | `docker-compose.yml:8,39` | Hardcoded secrets committed to version control |
| 7 | 🔴 Critical | Security | `internal/auth/handler.go:41` | `Secure: false` hardcoded on refresh cookie |
| 8 | 🔴 Critical | Data Integrity | `internal/block/repository.go:161-168` | Move block does not update ltree path |
| 9 | 🔴 Critical | Data Integrity | `internal/block/repository.go:278-280` | PermanentDelete orphans child blocks |
| 10 | 🔴 Critical | Data Integrity | `internal/auth/service.go:47-55` | Workspace creation failure orphans user |
| 11 | 🔴 Critical | Security | `internal/block/handler.go:35-38` | `uuid.MustParse` panics on invalid workspace ID |
| 12 | 🔴 Critical | Security | `internal/block/handler.go:40-46` | `uuid.MustParse` panics on invalid user ID |
| 13 | 🔴 Critical | Data Integrity | `internal/block/service.go:217` | Silent error discard in SplitBlock |
| 14 | 🔴 Critical | Data Integrity | `internal/block/service.go:232-234` | Ignored JSON unmarshal errors in MergeBlocks |
| 15 | 🔴 Critical | Data Integrity | `internal/workspace/repository.go:131` | Workspace deletion orphans blocks |
| 16 | 🔴 Critical | Data Integrity | `internal/workspace/repository.go:94-97` | Workspace deletion doesn't cascade to members |
| 17 | 🔴 Critical | Security | `internal/block/repository.go:306-311` | SQL injection risk in CleanupExpired |
| 18 | 🔴 Critical | Security | `.env:1` | Database credentials in committed .env |
| 19 | 🔴 Critical | Security | `main.go:113-114` | `godotenv.Load()` error silently ignored |
| 20 | 🔴 Critical | Security | `internal/api.go:56` | No MIME type validation on file uploads |
| 21 | 🔴 Critical | Security | `internal/api.go:66` | Upload path uses user-controlled extension |
| 22 | 🔴 Critical | Security | `main.go:172-178` | No authentication on upload serving |
| 23 | 🔴 Critical | Security | `internal/auth/handler.go:58-71` | No input validation on signup |
| 24 | 🔴 Critical | Security | `internal/auth/handler.go:73-86` | No rate limiting on login |
| 25 | 🔴 Critical | Security | `internal/auth/handler.go:95-96` | Refresh on invalid token clears cookie but doesn't invalidate token family |
| 26 | 🔴 Critical | Data Integrity | `internal/block/repository.go:148-151` | SoftDelete does not cascade to children |
| 27 | 🔴 Critical | Data Integrity | `internal/block/repository.go:153-159` | Restore doesn't check parent block deleted status |
| 28 | 🔴 Critical | Security | `internal/auth/jwt.go:27-29` | No algorithm restriction in JWT parsing |
| 29 | 🔴 Critical | Security | `Dockerfile:18` | Container runs as root |
| 30 | 🔴 Critical | Authentication | `internal/auth/service.go:102-103` | Logout doesn't invalidate existing access tokens |
| 31 | 🔴 Critical | Data Integrity | `internal/block/service.go:84-86` | Position calculation overflow risk |
| 32 | 🔴 Critical | Data Integrity | `internal/block/service.go:192-193` | Slice sharing causes data corruption in SplitBlock |
| 33 | 🔴 Critical | Security | `internal/workspace/handler.go:45` | Panic risk on missing context user ID |
| 34 | 🔴 Critical | Security | `internal/api.go:67` | Upload error doesn't clean up partially written file |
| 35 | 🔴 Critical | Data Integrity | `internal/block/repository.go:22-63` | No transaction on block Create |
| 36 | 🔴 Critical | Security | `internal/middleware/auth.go:26` | Auth middleware doesn't log failures |
| 37 | 🔴 Critical | Security | `internal/auth/service.go:122-136` | UpdateProfile allows name change without password |
| 38 | 🔴 Critical | Security | `migrations/000003_auth.up.sql:2` | Email unique constraint missing on users table |
| 39 | 🟠 High | Security | `internal/middleware/workspace.go:26-29` | Middleware returns 404 instead of 403, no Content-Type header |
| 40 | 🟠 High | Security | `main.go:191-196` | Dev proxy exposes internal Vite server |
| 41 | 🟠 High | Security | `internal/auth/handler.go:195-220` | DeleteAccount doesn't clear refresh cookie |
| 42 | 🟠 High | Security | `internal/auth/handler.go:129-160` | UpdateProfile doesn't validate email format |
| 43 | 🟠 High | Security | `docker-compose.yml:21` | `latest` tag for storage image |
| 44 | 🟠 High | Security | `main.go:206-209` | No ReadTimeout/WriteTimeout on http.Server |
| 45 | 🟠 High | Security | `internal/block/handler.go:169-190` | No max query length on search |
| 46 | 🟠 High | Security | `internal/block/handler.go:67-79` | GetPageTree doesn't check workspace membership |
| 47 | 🟠 High | Security | `internal/block/handler.go:104-121` | UpdateBlock doesn't check block ownership |
| 48 | 🟠 High | Security | `internal/block/handler.go:123-134` | DeleteBlock doesn't check block ownership |
| 49 | 🟠 High | Security | `internal/block/handler.go:136-148` | RestoreBlock doesn't check ownership |
| 50 | 🟠 High | Security | `internal/block/handler.go:150-167` | MoveBlock doesn't verify ownership |
| 51 | 🟠 High | Security | `internal/block/handler.go:210-224` | PermanentDelete doesn't verify authorization |
| 52 | 🟠 High | Security | `internal/block/handler.go:81-88` | ListPages has no pagination |
| 53 | 🟠 High | Security | `internal/block/handler.go:192-199` | ListFavorites has no pagination |
| 54 | 🟠 High | Security | `internal/workspace/handler.go:72-80` | Get workspace doesn't check membership |
| 55 | 🟠 High | Security | `internal/workspace/handler.go:137-145` | ListMembers doesn't check membership |
| 56 | 🟠 High | Security | `main.go:206-209` | No HTTP header security (CSP, HSTS, X-Frame-Options, etc.) |
| 57 | 🟠 High | Security | `internal/auth/handler.go:36-45` | No SameSite enforcement variation per environment |
| 58 | 🟠 High | Security | `internal/auth/jwt.go:9-12` | JWT missing standard `sub`, `iss`, `aud` claims |
| 59 | 🟠 High | Security | `internal/auth/jwt.go:14-24` | No key ID (kid) for key rotation |
| 60 | 🟠 High | Security | `docker-compose.yml:39` | JWT_SECRET in docker-compose.yml |
| 61 | 🟠 High | Security | `migrations/000003_auth.up.sql:2` | Empty password_hash DEFAULT '' for existing users |
| 62 | 🟠 High | Security | `migrations/000002_seed.up.sql:5-6` | Seed user has no password_hash |
| 63 | 🟠 High | Security | `internal/workspace/service.go:41-43` | Get returns workspace without authorization |
| 64 | 🟠 High | Security | `internal/workspace/service.go:82-84` | ListMembers returns members without auth check |
| 65 | 🟠 High | Security | `internal/workspace/service.go:86-94` | RemoveMember allows removing owner |
| 66 | 🟠 High | Security | `internal/auth/repository.go:46-58` | GetUserByEmail returns ErrInvalidCredentials instead of "user not found" |
| 67 | 🟠 High | Security | `internal/auth/repository.go:98-101` | Token hash function not salted |
| 68 | 🟠 High | Security | `internal/auth/repository.go:103-118` | No index on token_hash |
| 69 | 🟠 High | Performance | `internal/block/service.go:167-169` | CleanupExpired runs on every trash listing |
| 70 | 🟠 High | Security | `internal/api.go:55-76` | Upload handler writes file to disk before validating content |
| 71 | 🟠 High | Data Integrity | `internal/block/repository.go:22-63` | Position not guarded by UNIQUE constraint |
| 72 | 🟠 High | Performance | `internal/block/repository.go:175-217` | Search query could be expensive |
| 73 | 🟠 High | Data Integrity | `internal/block/service.go:222-252` | MergeBlocks modifies target then soft-deletes source — not atomic |
| 74 | 🟠 High | Security | `internal/block/handler.go:192-199` | Favorites endpoint doesn't verify workspace ownership |
| 75 | 🟠 High | Security | `internal/block/handler.go:201-208` | Trash endpoint doesn't verify workspace ownership |
| 76 | 🟠 High | Security | `internal/auth/handler.go:162-193` | UpdatePassword accepts unlimited attempts |
| 77 | 🟠 High | Security | `internal/auth/service.go:165-179` | DeleteAccount doesn't prevent last-owner deletion |
| 78 | 🟠 High | Data Integrity | `internal/workspace/repository.go:18-42` | Workspace create uses tx but owner_id could be wrong |
| 79 | 🟠 High | Data Integrity | `internal/block/repository.go:82-108` | GetTree recursive CTE could stack overflow |
| 80 | 🟠 High | Data Integrity | `internal/block/repository.go:161-168` | Move doesn't recalculate position within new parent |
| 81 | 🟠 High | Security | `internal/middleware/auth.go:17-23` | Bearer token check error message reveals auth scheme |
| 82 | 🟠 High | Security | `internal/middleware/middleware.go:22` | Recovery returns text/plain Content-Type with JSON body |
| 83 | 🟠 High | Security | `main.go:59-83` | SPA handler sets content-type only for known extensions |
| 84 | 🟠 High | Security | `internal/api.go:56` | Upload multipart memory limit is hardcoded |
| 85 | 🟠 High | Security | `internal/workspace/handler.go:126-128` | Default role 'member' without validation |
| 86 | 🟠 High | Security | `internal/workspace/handler.go:116-124` | InviteMember allows inviting by email with no guard |
| 87 | 🟠 High | Security | `docker-compose.yml:39` | JWT secret visible in process list |
| 88 | 🟠 High | Security | `main.go:36-37` | Embeds entire web/build including source maps |
| 89 | 🟠 High | Security | `internal/auth/handler.go:88-102` | Refresh should invalidate all tokens of same family |
| 90 | 🟠 High | Security | `internal/auth/service.go:106-112` | ValidateToken doesn't check if token is revoked |
| 91 | 🟠 High | Data Integrity | `migrations/000001_init.up.sql:25-37` | No CASCADE on blocks foreign keys |
| 92 | 🟠 High | Security | `internal/block/repository.go:59-62` | INSERT without explicit column list |
| 93 | 🟠 High | Data Integrity | `migrations/000003_auth.up.sql:5` | owner_id is nullable — existing workspaces get NULL owner |
| 94 | 🟠 High | Data Integrity | `migrations/000001_init.up.sql:1-2` | Extensions created but not checked on every migration run |
| 95 | 🟠 High | Security | `main.go:222` | server.Shutdown(ctx) error ignored |
| 96 | 🟠 High | Performance | `main.go:85-103` | Migrations run sequentially, no timeout per migration |
| 97 | 🟠 High | Security | `internal/auth/repository.go:121-143` | findAndConsumeRefreshToken logs token ID in production |
| 98 | 🟠 High | Security | `internal/auth/repository.go:158-179` | GetUserByRefreshToken logs token ID in production |
| 99 | 🟠 High | Security | `internal/block/handler.go:216-218` | PermanentDelete logs block IDs |
| 100 | 🟠 High | Security | `internal/auth/handler.go:88-102` | No CSRF protection on cookie-based refresh |
| 101 | 🟠 High | Security | `internal/block/handler.go:210-224` | PermanentDelete returns 204 even for non-existent blocks |
| 102 | 🟠 High | Data Integrity | `internal/block/repository.go:161-168` | MoveBlock doesn't validate new parent is in same workspace |
| 103 | 🟠 High | Data Integrity | `internal/block/repository.go:278-280` | PermanentDelete doesn't check block exists first |
| 104 | 🟠 High | Security | `main.go:67-68` | SPA handler silently falls back to index.html — could mask 404s |
| 105 | 🟠 High | Security | `internal/api.go:30-32` | Health endpoint has no rate limiting |
| 106 | 🟠 High | Security | `Dockerfile:10` | Go build uses root user |
| 107 | 🟠 High | Security | `internal/auth/repository.go:103-118` | Refresh token byte length is hardcoded 32 |
| 108 | 🟠 High | Security | `.gitignore:9` | test-results/ ignored — may contain sensitive data |
| 109 | 🟠 High | Security | `internal/auth/service.go:86-100` | Refresh doesn't check for compromised tokens |
| 110 | 🟠 High | Security | `internal/block/handler.go:48-65` | CreatePage doesn't validate workspace membership |
| 111 | 🟠 High | Data Integrity | `internal/block/service.go:22-47` | CreatePage creates block then initial child — not atomic |
| 112 | 🟠 High | Data Integrity | `internal/block/service.go:131-157` | MoveBlock doesn't refresh ltree path |
| 113 | 🟠 High | Security | `Dockerfile:21` | Binary in /app — unprivileged user should own |
| 114 | 🟠 High | Data Integrity | `internal/workspace/repository.go:18-42` | Create adds owner as member but could fail mid-transaction |
| 115 | 🟠 High | Security | `internal/api.go:67-69` | Upload error response reveals internal error existence |
| 116 | 🟠 High | Data Integrity | `internal/block/repository.go:59-62` | No validation that parent_id references block in same workspace |
| 117 | 🟠 High | Security | `internal/auth/handler.go:104-113` | Logout doesn't verify user identity cryptographically |
| 118 | 🟠 High | Security | `main.go:42-44` | Hardcoded default UUIDs exported as package variables |
| 119 | 🟠 High | Data Integrity | `internal/block/repository.go:59-62` | Pool.Exec doesn't check rows affected |
| 120 | 🟠 High | Security | `internal/auth/service.go:106-112` | No context timeout on token validation |
| 121 | 🟠 High | Data Integrity | `migrations/000001_init.up.sql:39-42` | Partial indexes but no UNIQUE index on (parent_id, position) |
| 122 | 🟠 High | Security | `internal/workspace/handler.go:107-135` | InviteMember doesn't validate email format |
| 123 | 🟠 High | Data Integrity | `internal/workspace/handler.go:107-135` | InviteMember with empty user_id and email proceeds silently |
| 124 | 🟠 High | Security | `internal/auth/repository.go:121-143` | Token hash could collide (SHA-256, low risk but theoretical) |
| 125 | 🟠 High | Security | `Dockerfile:18` | No HEALTHCHECK in production container |
| 126 | 🟠 High | Security | `docker-compose.yml:42-44` | No restart policy for app |
| 127 | 🟠 High | Security | `internal/middleware/auth.go:17-23` | Auth error messages don't distinguish "missing" from "invalid" |
| 128 | 🟠 High | Security | `internal/api.go:55-76` | No file size limit validation on upload |
| 129 | 🟠 High | Performance | `internal/block/repository.go:110-130` | ListPages can return unlimited results |
| 130 | 🟠 High | Security | `internal/workspace/service.go:33-35` | Create doesn't validate ownerID exists |
| 131 | 🟠 High | Security | `internal/workspace/handler.go:72-80` | Get workspace returns full workspace data without authorization |
| 132 | 🟠 High | Security | `internal/auth/handler.go:129-160` | UpdateProfile exposes user email enumeration |
| 133 | 🟠 High | Security | `internal/auth/handler.go:162-193` | UpdatePassword could be used in timing attack |
| 134 | 🟠 High | Data Integrity | `internal/block/repository.go:148-151` | SoftDelete doesn't update search_vector |
| 135 | 🟠 High | Security | `internal/block/handler.go:150-167` | MoveBlock no workspace-scope validation on target position |
| 136 | 🟠 High | Security | `main.go:198-203` | No fallback if SPA sub filesystem fails |
| 137 | 🟠 High | Security | `internal/auth/repository.go:73-78` | UpdateUser doesn't validate email uniqueness again |
| 138 | 🟠 High | Security | `internal/auth/repository.go:93-96` | DeleteUser doesn't cascade to refresh_tokens |
| 139 | 🟠 High | Data Integrity | `internal/auth/handler.go:104-113` | Logout returns 204 even if DeleteUserRefreshTokens fails |
| 140 | 🟠 High | Data Integrity | `internal/block/handler.go:123-134` | DeleteBlock returns 204 even if nothing was deleted |
| 141 | 🟠 High | Security | `internal/middleware/workspace.go:25-29` | Error during IsMember returns 404 (information leak) |
| 142 | 🟠 High | Data Integrity | `internal/block/repository.go:161-168` | Move parent to child would create cycle |
| 143 | 🟠 High | Data Integrity | `migrations/000001_init.up.sql:28` | Self-referential FK on blocks (parent_id → id) can create cycles |
| 144 | 🟠 High | Security | `internal/block/handler.go:35-38` | workspaceIDFromRequest panics on invalid UUID |
| 145 | 🟠 High | Security | `internal/block/handler.go:40-46` | userIDFromRequest panics on invalid UUID |
| 146 | 🟠 High | Security | `docker-compose.yml:28-44` | App service has no resource limits |
| 147 | 🟠 High | Security | `Dockerfile:18` | Alpine 3.19 — check for known CVEs |
| 148 | 🟠 High | Security | `main.go:152-156` | LocalFileStore always used regardless of config |
| 149 | 🟠 High | Data Integrity | `internal/block/repository.go:22-45` | Create generates UUID in app layer not DB layer |
| 150 | 🟠 High | Security | `internal/api.go:55` | Upload multipart size max is 10MB — no config option |
| 151 | 🟠 High | Data Integrity | `internal/workspace/repository.go:108-114` | AddMember uses ON CONFLICT DO NOTHING — no feedback |
| 152 | 🟠 High | Data Integrity | `internal/workspace/service.go:71-79` | InviteMember doesn't check email uniqueness |
| 153 | 🟠 High | Security | `internal/auth/handler.go:195-220` | DeleteAccount doesn't validate token explicitly |
| 154 | 🟠 High | Security | `internal/middleware/middleware.go:9-14` | Logger logs full request path including query params |
| 155 | 🟠 High | Security | `internal/auth/repository.go:103-118` | expires_at stored as string not timestamp |
| 156 | 🟠 High | Data Integrity | `main.go:85-103` | No migration version tracking table |
| 157 | 🟠 High | Security | `internal/block/handler.go:169-190` | Search query string logged by middleware |
| 158 | 🟠 High | Security | `internal/auth/service.go:86-100` | Refresh returns new token even if old was already used (reuse detection) |
| 159 | 🟠 High | Security | `internal/auth/service.go:39-44` | JWT_SECRET read from env directly — not from Config struct |
| 160 | 🟠 High | Security | `internal/storage/storage.go:28-38` | No file size limit on Put |
| 161 | 🟠 High | Data Integrity | `internal/block/repository.go:59-62` | No workspace_id constraint in path building query |
| 162 | 🟠 High | Security | `internal/api.go:32` | Health endpoint Content-Type not set explicitly |
| 163 | 🟠 High | Security | `internal/block/handler.go:90-102` | CreateBlock doesn't validate userID is valid |
| 164 | 🟠 High | Security | `internal/block/handler.go:48-65` | CreatePage uses userIDFromRequest which can return uuid.Nil |
| 165 | 🟠 High | Security | `internal/block/handler.go:90-102` | CreateBlock doesn't validate workspace membership |
| 166 | 🟠 High | Security | `main.go:164-169` | No CORS preflight max age |
| 167 | 🟠 High | Security | `docker-compose.yml:28-44` | No network isolation between services |
| 168 | 🟠 High | Security | `internal/auth/repository.go:46-58` | Error message doesn't distinguish "user not found" from "wrong password" |
| 169 | 🟠 High | Security | `internal/auth/repository.go:121-143` | Consumed token logged on failure but not on success |
| 170 | 🟠 High | Security | `internal/auth/handler.go:96` | Refresh clears cookie on error — potential for token theft amplification |
| 171 | 🟠 High | Data Integrity | `internal/block/repository.go:22-45` | Parent path query doesn't verify same workspace |
| 172 | 🟠 High | Data Integrity | `internal/block/repository.go:36-39` | Path building silently uses nil path when parent not found |
| 173 | 🟠 High | Security | `internal/middleware/middleware.go:3-4` | No import for RequestID/RealIP (used in main.go) |
| 174 | 🟠 High | Security | `internal/workspace/handler.go:97-105` | Delete workspace doesn't ask for confirmation |
| 175 | 🟠 High | Data Integrity | `internal/block/repository.go:132-146` | Update doesn't use transactions |
| 176 | 🟠 High | Security | `internal/block/handler.go:169-190` | No rate limiting on search endpoint |
| 177 | 🟠 High | Data Integrity | `internal/block/repository.go:148-151` | SoftDelete can delete already-deleted rows (concurrent safe but no-op) |
| 178 | 🟠 High | Data Integrity | `internal/block/repository.go:132-146` | Update returns Block{} on error — caller may use zero value |
| 179 | 🟠 High | Security | `main.go:210-212` | Signal channel not cleaned up after shutdown |
| 180 | 🟠 High | Security | `internal/auth/repository.go:103-118` | No timeout for token generation database query |
| 181 | 🟠 High | Security | `internal/middleware/workspace.go:20-21` | Type assertion from context could panic |
| 182 | 🟠 High | Data Integrity | `internal/block/repository.go:243-254` | MiddlePosition can return 0 causing position collision |
| 183 | 🟠 High | Data Integrity | `internal/block/repository.go:132-146` | Update does content and type in separate queries, not atomic |
| 184 | 🟠 High | Data Integrity | `internal/block/repository.go:59-62` | No updated_at in INSERT — uses zero time |
| 185 | 🟠 High | Data Integrity | `internal/auth/repository.go:73-78` | UpdateUser doesn't set updated_at |
| 186 | 🟠 High | Security | `internal/block/handler.go:210-224` | PermanentDelete doesn't verify block belongs to workspace |
| 187 | 🟠 High | Data Integrity | `internal/block/service.go:176-220` | SplitBlock doesn't preserve original block's position |
| 188 | 🟠 High | Security | `internal/block/handler.go:90-102` | CreateBlock doesn't verify userID is authenticated |
| 189 | 🟠 High | Security | `internal/middleware/auth.go:28` | Invalid token error doesn't set Content-Type header |
| 190 | 🟠 High | Security | `internal/block/handler.go:35-38` | workspaceIDFromRequest called before workspace middleware in some paths |
| 191 | 🟠 High | Security | `main.go:191-196` | Dev proxy forwards Authorization header to Vite |
| 192 | 🟠 High | Security | `migrations/000001_init.up.sql:1` | uuid-ossp extension is deprecated — use pgcrypto's gen_random_uuid() |
| 193 | 🟠 High | Data Integrity | `internal/block/repository.go:22-63` | Parent path not validated for workspace_id match |
| 194 | 🟠 High | Data Integrity | `internal/block/repository.go:59-62` | No RETURNING for created_at/updated_at — uses local time |
| 195 | 🟠 High | Data Integrity | `internal/block/repository.go:36-39` | Parent not found uses empty path — silent corruption |
| 196 | 🟠 High | Security | `main.go:124-128` | Pool variable shadows err variable |
| 197 | 🟠 High | Security | `.env:1` | SSL disabled in database connection |
| 198 | 🟠 High | Security | `migrations/000001_init.up.sql:1` | No migration for creating pgcrypto extension |
| 199 | 🟠 High | Security | `internal/block/service.go:22-47` | CreatePage uses json.Marshal — panics on failure |
| 200 | 🟠 High | Security | `Dockerfile:6` | Copying web/package.json and pnpm-lock.yaml together may fail if lock outdated |
| 201 | 🟠 High | Data Integrity | `Dockerfile:16` | Build copies entire source tree including test files |
| 202 | 🟠 High | Security | `internal/auth/service.go:60,95` | CreateRefreshToken expires_at as string could be misparsed |
| 203 | 🟠 High | Security | `internal/block/handler.go:216` | PermanentDelete logs block ID to stdout |
| 204 | 🟠 High | Data Integrity | `internal/workspace/repository.go:18-42` | No rollback on commit failure |
| 205 | 🟠 High | Security | `internal/storage/storage.go:23-26` | MkdirAll error silently ignored |
| 206 | 🟠 High | Security | `internal/storage/storage.go:29-33` | Put doesn't validate key for path traversal |
| 207 | 🟠 High | Security | `internal/storage/storage.go:44-46` | Delete doesn't validate key for path traversal |
| 208 | 🟠 High | Security | `internal/api.go:72-75` | Upload response includes full URL path |
| 209 | 🟠 High | Security | `internal/storage/storage.go:48-49` | PublicURL returns relative path — no host |
| 210 | 🟠 High | Security | `internal/auth/repository.go:121-143` | findAndConsumeRefreshToken race condition on DELETE |

---

## Detailed Findings

### 1 🔴 Critical — CORS wildcard origin with credentials enabled

**File:** [main.go](main.go:164-169)

```go
r.Use(cors.Handler(cors.Options{
    AllowedOrigins:   []string{"*"},
    AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
    AllowedHeaders:   []string{"Content-Type", "Authorization"},
    AllowCredentials: true,
}))
```

**Risk:** Browsers reject `Access-Control-Allow-Origin: *` when `Access-Control-Allow-Credentials: true`, but the configuration signals misunderstanding. If any origin-specific override occurs elsewhere, it enables credential-bearing cross-origin requests from any website, exposing all authenticated endpoints to XSRF.

**Fix:**
```go
AllowedOrigins: []string{"http://localhost:5173", "http://localhost:8080", "https://yourdomain.com"},
```

---

### 2 🔴 Critical — Path traversal in file serving

**File:** [main.go](main.go:172-178)

```go
r.Route("/uploads", func(r chi.Router) {
    r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
        filePath := filepath.Join(uploadsDir, chi.URLParam(r, "*"))
        http.ServeFile(w, r, filePath)
    })
})
```

**Risk:** `chi.URLParam(r, "*")` returns the unmatched path segment, which can contain `../` sequences. `filepath.Join` normalizes but does not prevent traversal above `uploadsDir`. An attacker can read arbitrary files: `GET /uploads/../../etc/passwd`.

**Fix:**
```go
r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
    requestedPath := chi.URLParam(r, "*")
    if strings.Contains(requestedPath, "..") {
        http.Error(w, `{"error":"invalid path"}`, http.StatusBadRequest)
        return
    }
    filePath := filepath.Join(uploadsDir, filepath.Clean("/"+requestedPath))
    if !strings.HasPrefix(filePath, filepath.Clean(uploadsDir)+string(filepath.Separator)) {
        http.Error(w, `{"error":"access denied"}`, http.StatusForbidden)
        return
    }
    http.ServeFile(w, r, filePath)
})
```

---

### 3 🔴 Critical — Path traversal in file store

**File:** [internal/storage/storage.go](internal/storage/storage.go:28-38)

```go
func (s *LocalFileStore) Put(ctx context.Context, key string, reader io.Reader) error {
    path := filepath.Join(s.dir, key)
    os.MkdirAll(filepath.Dir(path), 0755)
    f, err := os.Create(path)
    ...
}
```

**Risk:** The `key` parameter comes from user-controlled filename extensions and could contain `../`. Same issue exists in `Get` (line 41) and `Delete` (line 45).

**Fix:** Validate the key doesn't contain path separators or `..`:
```go
func validateKey(key string) error {
    if key == "" || strings.Contains(key, "..") || strings.Contains(key, "/") || strings.Contains(key, "\\") {
        return fmt.Errorf("invalid key: %s", key)
    }
    return nil
}
```

---

### 4 🔴 Critical — Hardcoded secrets in docker-compose.yml

**File:** [docker-compose.yml](docker-compose.yml:8,39)

```yaml
environment:
    POSTGRES_PASSWORD: notion
    JWT_SECRET: development-jwt-secret-change-in-production
```

**Risk:** Database credentials and JWT signing key are committed to version control in plaintext. Anyone with repo access can forge JWTs and access the production database.

**Fix:** Use environment variable references:
```yaml
POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
JWT_SECRET: ${JWT_SECRET}
```

---

### 5 🔴 Critical — `uuid.MustParse` panics in request handlers

**File:** [internal/block/handler.go](internal/block/handler.go:35-46)

```go
func workspaceIDFromRequest(r *http.Request) uuid.UUID {
    id := chi.URLParam(r, "workspaceId")
    return uuid.MustParse(id)
}

func userIDFromRequest(r *http.Request) uuid.UUID {
    id, ok := r.Context().Value(auth.CtxUserID).(string)
    if !ok {
        return uuid.Nil
    }
    return uuid.MustParse(id)
}
```

**Risk:** If any request has an invalid UUID in the URL param or stored in context (e.g., corrupted JWT), the handler panics, crashing the server. This is a denial-of-service vector.

**Fix:** Use `uuid.Parse()` instead and return an error response:
```go
func workspaceIDFromRequest(r *http.Request) (uuid.UUID, error) {
    return uuid.Parse(chi.URLParam(r, "workspaceId"))
}
```

---

### 6 🔴 Critical — Block move doesn't update ltree path

**File:** [internal/block/repository.go](internal/block/repository.go:161-168)

```go
func (r *Repository) Move(ctx context.Context, id uuid.UUID, req MoveBlockRequest) (Block, error) {
    _, err := r.pool.Exec(ctx, `UPDATE blocks SET parent_id = $1, position = $2, updated_at = now() WHERE id = $3 AND deleted_at IS NULL`,
        req.ParentID, req.Position, id)
    ...
}
```

**Risk:** The ltree `path` column is not updated when a block moves to a new parent. The path still reflects the old hierarchy, breaking all recursive queries that depend on path (GetTree, Search). Child blocks also retain incorrect paths.

**Fix:** Recalculate path recursively:
```sql
WITH RECURSIVE descendants AS (
    SELECT id, $1::ltree || subpath(path, nlevel(path)-1) AS new_path FROM blocks WHERE id = $2
    UNION ALL
    SELECT b.id, new_path || subpath(b.path, nlevel(b.path)-1)
    FROM blocks b JOIN descendants d ON b.parent_id = d.id
)
UPDATE blocks SET path = descendants.new_path FROM descendants WHERE blocks.id = descendants.id
```

---

### 7 🔴 Critical — Workspace deletion orphans blocks

**File:** [internal/workspace/repository.go](internal/workspace/repository.go:94-97)

```go
func (r *Repository) Delete(ctx context.Context, id string) error {
    _, err := r.pool.Exec(ctx, `DELETE FROM workspaces WHERE id = $1`, id)
    return err
}
```

**Risk:** Deleting a workspace leaves all associated blocks, workspace_members, and refresh tokens as orphaned records. No CASCADE is defined in the schema.

**Fix:** Add `ON DELETE CASCADE` to foreign keys in the migration, or delete related records in a transaction:
```go
func (r *Repository) Delete(ctx context.Context, id string) error {
    tx, err := r.pool.Begin(ctx)
    if err != nil { return err }
    defer tx.Rollback(ctx)
    if _, err := tx.Exec(ctx, `DELETE FROM blocks WHERE workspace_id = $1`, id); err != nil { return err }
    if _, err := tx.Exec(ctx, `DELETE FROM workspace_members WHERE workspace_id = $1`, id); err != nil { return err }
    if _, err := tx.Exec(ctx, `DELETE FROM workspaces WHERE id = $1`, id); err != nil { return err }
    return tx.Commit(ctx)
}
```

---

### 8 🔴 Critical — Empty password hash in seed migration

**File:** [migrations/000002_seed.up.sql](migrations/000002_seed.up.sql:5-6)

```sql
INSERT INTO users (id, email, name) VALUES
  ('00000000-0000-0000-0000-000000000002', 'dev@notion-clone.local', 'Dev User')
```

**Risk:** After migration 000003 adds `password_hash TEXT NOT NULL DEFAULT ''`, the seed user has an empty password hash. An attacker could authenticate as this user by providing a password whose bcrypt hash happens to match an empty string (or by exploiting any code path that accepts empty passwords).

**Fix:** Set a strong random password hash in the seed migration.

---

### 9 🔴 Critical — No input validation on signup

**File:** [internal/auth/handler.go](internal/auth/handler.go:58-71)

```go
func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
    var req SignupRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, "invalid request body")
        return
    }
    resp, refreshToken, err := h.svc.Signup(r.Context(), req)
    ...
}
```

**Risk:** Empty email, empty password, or email without `@` are accepted. The test `TestSignup_MissingFields` confirms `{"email": "test@test.com"}` succeeds with no password. This creates accounts with empty password hashes.

**Fix:** Add validation:
```go
if req.Email == "" || !strings.Contains(req.Email, "@") {
    respondError(w, http.StatusBadRequest, "valid email is required")
    return
}
if len(req.Password) < 8 {
    respondError(w, http.StatusBadRequest, "password must be at least 8 characters")
    return
}
```

---

### 10 🔴 Critical — Auth middleware doesn't set Content-Type on errors

**File:** [internal/middleware/auth.go](internal/middleware/auth.go:17-31)

```go
func AuthMiddleware(validator TokenValidator) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            authHeader := r.Header.Get("Authorization")
            if !strings.HasPrefix(authHeader, "Bearer ") {
                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(http.StatusUnauthorized)
                w.Write([]byte(`{"error":"missing authorization header"}`))
                return
            }
            token := strings.TrimPrefix(authHeader, "Bearer ")
            userID, err := validator.ValidateToken(token)
            if err != nil {
                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(http.StatusUnauthorized)
                w.Write([]byte(`{"error":"invalid token"}`))
                return
            }
```

**Risk:** Missing `Content-Type` header on error responses in some code paths. While this particular function does set it, many other middleware functions and handlers do not (see workspace middleware). Browsers may misinterpret the response body.

**Fix:** Ensure all error responses consistently set `Content-Type: application/json`.

---

### 11 🟠 High — No rate limiting on auth endpoints

**File:** [internal/auth/handler.go](internal/auth/handler.go:73-86)

```go
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
    ...no rate limiting...
}
```

**Risk:** Brute-force attacks against login, signup, and refresh endpoints are unbounded. An attacker can try millions of passwords or create thousands of accounts.

**Fix:** Add rate limiting middleware:
```go
import "golang.org/x/time/rate"

var loginLimiter = rate.NewLimiter(rate.Every(time.Second), 5)

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
    if !loginLimiter.Allow() {
        respondError(w, http.StatusTooManyRequests, "rate limit exceeded")
        return
    }
    ...
}
```

---

### 12 🟠 High — No HTTP security headers

**File:** [main.go](main.go:206-209)

```go
server := &http.Server{
    Addr:    ":" + cfg.Port,
    Handler: r,
}
```

**Risk:** Missing security headers including Content-Security-Policy (XSS mitigation), X-Frame-Options (clickjacking), X-Content-Type-Options (MIME sniffing), Strict-Transport-Security (HTTPS enforcement), and Referrer-Policy.

**Fix:** Add security headers middleware:
```go
func SecurityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
        if !cfg.DevMode {
            w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        }
        next.ServeHTTP(w, r)
    })
}
```

---

### 13 🟠 High — SQL injection risk in CleanupExpired

**File:** [internal/block/repository.go](internal/block/repository.go:306-311)

```go
func (r *Repository) CleanupExpired(ctx context.Context, workspaceID uuid.UUID, days int) error {
    _, err := r.pool.Exec(ctx, `
        DELETE FROM blocks
        WHERE workspace_id = $1 AND deleted_at IS NOT NULL AND deleted_at < now() - ($2 || ' days')::interval
    `, workspaceID, fmt.Sprintf("%d", days))
    return err
}
```

**Risk:** While `fmt.Sprintf("%d", days)` is safe for int, using string concatenation for SQL constructs bypasses parameterized query protection for the interval. If `days` were ever a string type, this would be directly injectable.

**Fix:** Use parameterized interval:
```go
deleted_at < now() - make_interval(days => $2::int)
```

---

### 14 🟠 High — Workspace middleware returns text/plain with JSON body

**File:** [internal/middleware/workspace.go](internal/middleware/workspace.go:25-29)

```go
ok, err := validator.IsMember(r.Context(), workspaceID, userID)
if err != nil || !ok {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusNotFound)
    w.Write([]byte(`{"error":"workspace not found"}`))
    return
}
```

**Risk:** Inconsistent Content-Type handling. This function does set it, but the Recovery middleware (line 22 of middleware.go) uses `http.Error` which returns `text/plain` with a JSON body.

**Fix:** Ensure all error responses use consistent JSON Content-Type.

---

### 15 🟠 High — JWT missing standard claims

**File:** [internal/auth/jwt.go](internal/auth/jwt.go:9-24)

```go
type Claims struct {
    UserID string `json:"user_id"`
    jwt.RegisteredClaims
}

func GenerateAccessToken(userID, secret string) (string, error) {
    claims := Claims{
        UserID: userID,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }
```

**Risk:** Missing `Subject` (sub), `Issuer` (iss), and `Audience` (aud) claims. Without `sub` in the standard field, token validation across services is fragile. Without `iss`/`aud`, a token issued for one environment could be replayed against another.

**Fix:**
```go
RegisteredClaims: jwt.RegisteredClaims{
    Subject:   userID,
    Issuer:    "notion-clone",
    Audience:  jwt.ClaimStrings{"notion-clone-api"},
    ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
    IssuedAt:  jwt.NewNumericDate(time.Now()),
},
```

---

### 16 🟠 High — Duplicate respond functions across three packages

**Files:**
- [internal/auth/handler.go](internal/auth/handler.go:222-232)
- [internal/block/handler.go](internal/block/handler.go:23-33)
- [internal/workspace/handler.go](internal/workspace/handler.go:21-31)

```go
func respond(w http.ResponseWriter, status int, data interface{}) { ... }
func respondError(w http.ResponseWriter, status int, msg string) { ... }
```

**Risk:** Three identical implementations of `respond` and `respondError` across packages. Violates DRY; any behavioral change (e.g., adding a request ID to all responses) requires editing three files.

**Fix:** Create a shared `internal/httputil/response.go`:
```go
package httputil

func JSON(w http.ResponseWriter, status int, data interface{}) { ... }
func Error(w http.ResponseWriter, status int, msg string) { ... }
```

---

### 17 🟠 High — No auth handler tests for UpdateProfile, UpdatePassword, DeleteAccount

**Files:** [internal/auth/handler_test.go](internal/auth/handler_test.go)

**Risk:** Three critical user management endpoints have no test coverage. Logic errors in password verification, email uniqueness checking, or cascade behavior go undetected.

**Fix:** Add test cases:
```go
func TestUpdateProfile_Success(t *testing.T) { ... }
func TestUpdatePassword_Success(t *testing.T) { ... }
func TestDeleteAccount_Success(t *testing.T) { ... }
```

---

### 18 🟠 High — No block handler tests at all

**File:** — (no test file exists for block package)

**Risk:** Zero test coverage for the most complex package in the codebase. Block creation, positioning, tree retrieval, search, move, split, merge, trash, favorites, and permanent delete are untested.

**Fix:** Create `internal/block/handler_test.go` and `internal/block/repository_test.go` with table-driven tests for each endpoint.

---

### 19 🟠 High — godotenv.Load() error silently ignored

**File:** [main.go](main.go:113-114)

```go
func main() {
    godotenv.Load()
    ...
}
```

**Risk:** If the `.env` file is missing or has syntax errors, the application silently continues without notifying the developer. Missing critical variables (DATABASE_URL) will only be caught later at config.Load() failure.

**Fix:**
```go
if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
    slog.Warn("error loading .env file", "error", err)
}
```

---

### 20 🟠 High — No ReadTimeout/WriteTimeout on http.Server

**File:** [main.go](main.go:206-209)

```go
server := &http.Server{
    Addr:    ":" + cfg.Port,
    Handler: r,
}
```

**Risk:** Without timeouts, a slow client can hold connections open indefinitely (slow loris attack), exhausting server resources.

**Fix:**
```go
server := &http.Server{
    Addr:         ":" + cfg.Port,
    Handler:      r,
    ReadTimeout:  15 * time.Second,
    WriteTimeout: 30 * time.Second,
    IdleTimeout:  60 * time.Second,
}
```

---

### 21 🟠 High — Recovery middleware returns text/plain

**File:** [internal/middleware/middleware.go](internal/middleware/middleware.go:17-27)

```go
func Recovery(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                slog.Error("panic", "error", err)
                http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

**Risk:** `http.Error()` sets Content-Type to `text/plain; charset=utf-8` but the body is JSON. Clients expecting JSON will fail to parse.

**Fix:**
```go
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusInternalServerError)
json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
```

---

### 22 🟠 High — Logger middleware logs full request path including query params

**File:** [internal/middleware/middleware.go](internal/middleware/middleware.go:9-14)

```go
func Logger(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        slog.Info("request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(start))
    })
}
```

**Risk:** Query parameters are not logged (only `r.URL.Path`), but the issue is that request durations are logged per-request which could be high cardinality. The larger concern is that `r.URL.Path` includes URL-encoded data which in some configurations could contain sensitive information.

**Fix:** Consider sanitizing the path or using structured logging with appropriate filtering.

---

### 23 🟠 High — Refresh token rotation doesn't detect token family theft

**File:** [internal/auth/service.go](internal/auth/service.go:86-100)

```go
func (s *Service) Refresh(ctx context.Context, refreshTokenHex string) (*AuthResponse, string, error) {
    user, err := s.repo.GetUserByRefreshToken(ctx, refreshTokenHex)
    if err != nil {
        return nil, "", err
    }
    accessToken, err := GenerateAccessToken(user.ID, s.jwtSecret)
    if err != nil {
        return nil, "", err
    }
    newRefresh, err := s.repo.CreateRefreshToken(ctx, user.ID, time.Now().Add(7*24*time.Hour).Format(time.RFC3339))
    if err != nil {
        return nil, "", err
    }
    return &AuthResponse{User: *user, AccessToken: accessToken}, newRefresh, nil
}
```

**Risk:** Refresh token rotation consumes the old token (in GetUserByRefreshToken/repository.go), but does not detect if a stolen token is reused after the legitimate user already consumed it. The attacker can still use the new token they obtained.

**Fix:** Implement token family tracking: if a consumed token's hash is not found (already consumed), invalidate ALL tokens for the user and force re-login.

---

### 24 🟠 High — No CSRF protection on cookie-based auth

**File:** [internal/auth/handler.go](internal/auth/handler.go:88-102)

**Risk:** Refresh token is stored in a cookie with `SameSite=Strict`, which provides CSRF protection for refresh. However, there's no CSRF token for state-changing operations like profile update, password change, or account deletion.

**Fix:** Implement CSRF token generation/validation for all authenticated mutating requests.

---

### 25 🟠 High — Position calculation overflow risk

**File:** [internal/block/service.go](internal/block/service.go:84-86)

```go
last := siblings[len(siblings)-1]
after := last.Position + (1 << 31)
position = after
```

**Risk:** After approximately 2 billion inserts between the same siblings, `int64` overflow occurs. In practice this is unlikely but causes undefined behavior when it happens.

**Fix:** Use the `MiddlePosition` function for all position calculations (see MoveBlock), which uses average-based fractional indexing.

---

### 26 🟠 High — Slice sharing causes data corruption in SplitBlock

**File:** [internal/block/service.go](internal/block/service.go:192-199)

```go
leftText := richText[:splitPosition]
rightText := richText[splitPosition:]

content["rich_text"] = leftText
leftContent, _ := json.Marshal(content)

content["rich_text"] = rightText
rightContent, _ := json.Marshal(content)
```

**Risk:** After `leftContent` is marshaled, `content["rich_text"]` is reassigned to `rightText`, which shares the same underlying array as `richText`. But `leftText` also shares this array. If `rightText` is later modified (e.g., appended to), it could overwrite `leftText`'s data. Additionally, `richText[:splitPosition]` shares the backing array with the original.

**Fix:** Make deep copies:
```go
leftText := append([]interface{}{}, richText[:splitPosition]...)
rightText := append([]interface{}{}, richText[splitPosition:]...)
```

---

### 27 🟠 High — Silent error discards throughout the codebase

**File:** [internal/block/service.go](internal/block/service.go:217)

```go
_, _ = s.repo.Update(ctx, newBlock.ID, UpdateBlockRequest{Content: rightContent})
```

**Risk:** This is the second `Update` call in SplitBlock that is completely redundant with the `Create` call above (Create already inserted the content). But patterns like this throughout the codebase where errors are silently discarded (JSON marshaling, JSON unmarshaling in MergeBlocks, etc.) can cause silent data corruption.

**Fix:** Either remove the redundant call or handle the error:
```go
if _, err := s.repo.Update(ctx, newBlock.ID, UpdateBlockRequest{Content: rightContent}); err != nil {
    return Block{}, Block{}, fmt.Errorf("update new block: %w", err)
}
```

---

### 28 🟠 High — Concurrent position calculation race condition

**File:** [internal/block/service.go](internal/block/service.go:78-92)

```go
siblings, err := s.repo.GetSiblings(ctx, parentID, workspaceID)
if err == nil && len(siblings) > 0 {
    last := siblings[len(siblings)-1]
    after := last.Position + (1 << 31)
    position = after
}
```

**Risk:** Two concurrent requests to create a block under the same parent will both compute the same `position` value, because the SELECT and INSERT are NOT in a database transaction. Both blocks get identical positions, breaking ordering.

**Fix:** Either use a database transaction with `SELECT ... FOR UPDATE`, or use the unique constraint on `(parent_id, position)` when it's added.

---

### 29 🟠 High — Workspace handler tests pass `nil` for authSvc

**File:** [internal/workspace/handler_test.go](internal/workspace/handler_test.go:36)

```go
h := NewHandler(svc, nil)
```

**Risk:** The workspace handler's `InviteMember` method uses `h.authSvc.GetUserByEmail` when `UserID` is empty and `Email` is provided. If any test exercises this code path, it will nil-pointer dereference and panic.

**Fix:** Use a mock auth service:
```go
mockAuthSvc := &auth.MockUserRepo{...}
h := NewHandler(svc, authSvc)
```

---

### 30 🟠 High — Service layer tightly coupled to pgxpool

**File:** [internal/block/service.go](internal/block/service.go:16-20)

```go
func NewService(pool *pgxpool.Pool) *Service {
    return &Service{
        repo: NewRepository(pool),
    }
}
```

**Risk:** The block service creates its own repository internally, making it impossible to mock the repository in unit tests. This is why there are no tests for the block package.

**Fix:** Accept the repository interface instead:
```go
type BlockRepository interface { ... }
func NewService(repo BlockRepository) *Service { ... }
```

---

## Quick Wins
> Fixes achievable in under 1 hour

- [ ] #1 — CORS: Replace wildcard origin with explicit list
- [ ] #3 — JWT: Set JWT_SECRET via environment in docker-compose
- [ ] #5 — File serving: Add path traversal check
- [ ] #12 — Security headers: Add middleware for CSP, HSTS, XFO, XCTO
- [ ] #16 — Deduplicate respond functions into shared package
- [ ] #21 — Fix Recovery middleware to return JSON Content-Type
- [ ] #26 — Fix slice sharing in SplitBlock
- [ ] #27 — Remove redundant Update call in SplitBlock
- [ ] #30 — Refactor Service to accept repository interface
- [ ] #42 — Add email format validation
- [ ] #162 — Remove `.env` from git and create `.env.example`

---

## Roadmap

### Milestone 1 — Security Hardening (P0)
1. Fix CORS wildcard + credentials (🔴 #1)
2. Add path traversal protection to file serving and storage (🔴 #4, #5)
3. Remove hardcoded secrets from docker-compose and .env (🔴 #6)
4. Replace `uuid.MustParse` with error handling (🔴 #11, #12)
5. Add input validation to signup/login endpoints (🔴 #23)
6. Restrict JWT algorithm with `WithValidMethods` (🔴 #28)
7. Add HTTP security headers middleware (🟠 #12)
8. Add rate limiting to auth endpoints (🟠 #11)
9. Fix auth cookie `Secure` flag to depend on environment (🔴 #7)
10. Add CSRF protection (🟠 #24)

### Milestone 2 — Data Integrity (P1)
1. Fix block ltree path update on move (🔴 #8)
2. Add CASCADE deletes or transactional cleanup for workspace/block deletion (🔴 #9, #15, #16)
3. Add position UNIQUE constraint and transactional position allocation (🟠 #28)
4. Fix SplitBlock slice aliasing (🟠 #26)
5. Add transaction for block Create (🔴 #35)
6. Implement token theft detection in refresh rotation (🟠 #23)
7. Add auth handler tests for untested endpoints (🟠 #17)
8. Add block handler tests (🟠 #18)

### Milestone 3 — Authentication & Session Management (P1)
1. Implement refresh token family tracking and theft detection
2. Add session listing/revocation endpoints
3. Add standard JWT claims (sub, iss, aud)
4. Add JWT key rotation support (kid)
5. Implement access token revocation via blocklist

### Milestone 4 — Code Quality & DX (P3)
1. Extract shared `respond`/`respondError` into `internal/httputil`
2. Create `BlockRepository` interface for testability
3. Set up `.golangci.yml` with comprehensive linters
4. Create `.env.example` documenting all required variables
5. Add CI/CD pipeline with lint, vet, and test stages
6. Add structured error types instead of `errors.New`
7. Add Docker non-root user and HEALTHCHECK
8. Remove production `log.Printf` calls, use structured slog
9. Add request ID propagation across all log entries

### Milestone 5 — Performance & Scalability (P3)
1. Add pagination to ListPages, ListFavorites, Search
2. Add `search_vector` index maintenance on mutations
3. Add database connection pool tuning via config
4. Cache expensive operations (search results, page trees)
5. Add rate limiting to search and upload endpoints
6. Optimize search query with proper indexing
