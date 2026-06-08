# Auth & Workspaces Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add user authentication (signup/login/logout with JWT + refresh token rotation) and multi-workspace support.

**Architecture:** New `internal/auth/` and `internal/workspace/` Go packages. JWT access token (15min, in memory), refresh token (7d, HTTP-only cookie, bcrypt hash in DB). Frontend auth store with auto-refresh interceptor. Existing block/page routes re-scoped under `/api/v1/workspaces/:workspaceId/`.

**Tech Stack:** Go (Chi router, golang-jwt/jwt/v5, bcrypt), Svelte 5 runes, Heroicons

---

### Task 1: Migration 000003 — users, refresh_tokens, workspaces, workspace_members

**Files:**
- Create: `migrations/000003_auth.up.sql`
- Create: `migrations/000003_auth.down.sql`

- [ ] **Step 1: Create up migration**

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    name TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE workspace_members (
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'member' CHECK (role IN ('owner', 'admin', 'member')),
    joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (workspace_id, user_id)
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_workspace_members_user_id ON workspace_members(user_id);
```

- [ ] **Step 2: Create down migration**

```sql
DROP TABLE IF EXISTS workspace_members;
DROP TABLE IF EXISTS workspaces;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS users;
```

- [ ] **Step 3: Commit**

```bash
git add migrations/000003_auth.up.sql migrations/000003_auth.down.sql
git commit -m "feat: add users, refresh_tokens, workspaces, workspace_members tables"
```

### Task 2: JWT utility package

**Files:**
- Create: `internal/auth/jwt.go`

- [ ] **Step 1: Create jwt.go**

```go
package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

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
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ValidateAccessToken(tokenString, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	return claims, nil
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/auth/jwt.go
git commit -m "feat: add JWT access token generation and validation"
```

### Task 3: Auth handler, service, repository

**Files:**
- Create: `internal/auth/repository.go`
- Create: `internal/auth/service.go`
- Create: `internal/auth/handler.go`
- Create: `internal/auth/model.go`

- [ ] **Step 1: Create model.go**

```go
package auth

import "time"

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"created_at"`
}

type RefreshToken struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	TokenHash string    `json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	User         User   `json:"user"`
	AccessToken  string `json:"access_token"`
}
```

- [ ] **Step 2: Create repository.go**

```go
package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreateUser(ctx context.Context, email, password, name string) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &User{}
	err = r.pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, name) VALUES ($1, $2, $3)
		 RETURNING id, email, name, created_at`,
		email, string(hash), name,
	).Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	user := &User{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, name, created_at FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id string) (*User, error) {
	user := &User{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, name, created_at FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *Repository) CreateRefreshToken(ctx context.Context, userID string, expiresAt string) (string, error) {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	token := hex.EncodeToString(bytes)
	hash, err := bcrypt.GenerateFromPassword(bytes, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	_, err = r.pool.Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, string(hash), expiresAt,
	)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (r *Repository) ValidateRefreshToken(ctx context.Context, tokenHex string) (string, error) {
	bytes, err := hex.DecodeString(tokenHex)
	if err != nil {
		return "", err
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, token_hash FROM refresh_tokens WHERE expires_at > now()`,
	)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	for rows.Next() {
		var id, userID, hash string
		if err := rows.Scan(&id, &userID, &hash); err != nil {
			continue
		}
		if bcrypt.CompareHashAndPassword([]byte(hash), bytes) == nil {
			// Delete the used token (rotation)
			r.pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE id = $1`, id)
			return userID, nil
		}
	}
	return "", bcrypt.ErrMismatchedHashAndPassword
}

func (r *Repository) DeleteUserRefreshTokens(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE user_id = $1`, userID)
	return err
}
```

- [ ] **Step 3: Create service.go**

```go
package auth

import (
	"context"
	"errors"
	"net/http"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/ionutale/notion-clone-golang/internal/workspace"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailTaken         = errors.New("email already in use")
)

type Service struct {
	repo      *Repository
	wsSvc     *workspace.Service
	jwtSecret string
}

func NewService(repo *Repository, wsSvc *workspace.Service) *Service {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret-change-in-production"
	}
	return &Service{repo: repo, wsSvc: wsSvc, jwtSecret: secret}
}

func (s *Service) Signup(ctx context.Context, req SignupRequest) (*AuthResponse, string, error) {
	user, err := s.repo.CreateUser(ctx, req.Email, req.Password, req.Name)
	if err != nil {
		return nil, "", ErrEmailTaken
	}
	// Create default personal workspace
	_, err = s.wsSvc.Create(ctx, user.Name+"'s Workspace", user.ID)
	if err != nil {
		return nil, "", err
	}
	accessToken, err := GenerateAccessToken(user.ID, s.jwtSecret)
	if err != nil {
		return nil, "", err
	}
	refreshToken, err := s.repo.CreateRefreshToken(ctx, user.ID, time.Now().Add(7*24*time.Hour).Format(time.RFC3339))
	if err != nil {
		return nil, "", err
	}
	return &AuthResponse{User: *user, AccessToken: accessToken}, refreshToken, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, string, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, "", ErrInvalidCredentials
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		return nil, "", ErrInvalidCredentials
	}
	accessToken, err := GenerateAccessToken(user.ID, s.jwtSecret)
	if err != nil {
		return nil, "", err
	}
	refreshToken, err := s.repo.CreateRefreshToken(ctx, user.ID, time.Now().Add(7*24*time.Hour).Format(time.RFC3339))
	if err != nil {
		return nil, "", err
	}
	return &AuthResponse{User: User{ID: user.ID, Email: user.Email, Name: user.Name, CreatedAt: user.CreatedAt}, AccessToken: accessToken}, refreshToken, nil
}

func (s *Service) Refresh(ctx context.Context, refreshTokenHex string) (*AuthResponse, string, error) {
	userID, err := s.repo.ValidateRefreshToken(ctx, refreshTokenHex)
	if err != nil {
		return nil, "", err
	}
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, "", err
	}
	accessToken, err := GenerateAccessToken(userID, s.jwtSecret)
	if err != nil {
		return nil, "", err
	}
	newRefresh, err := s.repo.CreateRefreshToken(ctx, userID, time.Now().Add(7*24*time.Hour).Format(time.RFC3339))
	if err != nil {
		return nil, "", err
	}
	return &AuthResponse{User: *user, AccessToken: accessToken}, newRefresh, nil
}

func (s *Service) Logout(ctx context.Context, userID string) error {
	return s.repo.DeleteUserRefreshTokens(ctx, userID)
}

func (s *Service) ValidateToken(tokenString string) (string, error) {
	claims, err := ValidateAccessToken(tokenString, s.jwtSecret)
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}

func (s *Service) GetUser(ctx context.Context, userID string) (*User, error) {
	return s.repo.GetUserByID(ctx, userID)
}
```

- [ ] **Step 4: Create handler.go**

```go
package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

const refreshCookieName = "refresh_token"

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/auth/signup", h.Signup)
	r.Post("/auth/login", h.Login)
	r.Post("/auth/refresh", h.Refresh)
	r.Post("/auth/logout", h.Logout)
	r.Get("/auth/me", h.Me)
}

func (h *Handler) setRefreshCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // true in production
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	})
}

func (h *Handler) clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		MaxAge:   -1,
	})
}

func respond(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func respondError(w http.ResponseWriter, status int, msg string) {
	respond(w, status, map[string]string{"error": msg})
}

func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	resp, refreshToken, err := h.svc.Signup(r.Context(), req)
	if err != nil {
		respondError(w, http.StatusConflict, err.Error())
		return
	}
	h.setRefreshCookie(w, refreshToken)
	respond(w, http.StatusCreated, resp)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	resp, refreshToken, err := h.svc.Login(r.Context(), req)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}
	h.setRefreshCookie(w, refreshToken)
	respond(w, http.StatusOK, resp)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(refreshCookieName)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "no refresh token")
		return
	}
	resp, newToken, err := h.svc.Refresh(r.Context(), cookie.Value)
	if err != nil {
		h.clearRefreshCookie(w)
		respondError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}
	h.setRefreshCookie(w, newToken)
	respond(w, http.StatusOK, resp)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	h.svc.Logout(r.Context(), userID)
	h.clearRefreshCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	user, err := h.svc.GetUser(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}
	respond(w, http.StatusOK, user)
}
```

- [ ] **Step 5: Update go.mod**

Run: `cd /Users/ionutale/developer/notion-clone-golang && go get github.com/golang-jwt/jwt/v5 golang.org/x/crypto`

- [ ] **Step 6: Verify build**

Run: `cd /Users/ionutale/developer/notion-clone-golang && go build ./...`
Expected: Build succeeds.

- [ ] **Step 7: Commit**

```bash
git add internal/auth/ go.mod go.sum
git commit -m "feat: add auth handler with signup, login, refresh, logout, me"
```

### Task 4: Auth middleware

**Files:**
- Create: `internal/middleware/auth.go`

- [ ] **Step 1: Create auth middleware**

```go
package middleware

import (
	"context"
	"net/http"
	"strings"
)

type TokenValidator interface {
	ValidateToken(tokenString string) (string, error)
}

func AuthMiddleware(validator TokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
				return
			}
			token := strings.TrimPrefix(auth, "Bearer ")
			userID, err := validator.ValidateToken(token)
			if err != nil {
				http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), "user_id", userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/middleware/auth.go
git commit -m "feat: add JWT auth middleware"
```

### Task 5: Workspace handler, service, repository

**Files:**
- Create: `internal/workspace/model.go`
- Create: `internal/workspace/repository.go`
- Create: `internal/workspace/service.go`
- Create: `internal/workspace/handler.go`

- [ ] **Step 1: Create model.go**

```go
package workspace

import "time"

type Workspace struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	OwnerID   string    `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Member struct {
	UserID   string `json:"user_id"`
	Role     string `json:"role"`
	JoinedAt string `json:"joined_at"`
}

type CreateRequest struct {
	Name string `json:"name"`
}

type InviteRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}
```

- [ ] **Step 2: Create repository.go**

```go
package workspace

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, name, ownerID string) (*Workspace, error) {
	w := &Workspace{}
	err := r.pool.QueryRow(ctx,
		`INSERT INTO workspaces (name, owner_id) VALUES ($1, $2)
		 RETURNING id, name, owner_id, created_at`,
		name, ownerID,
	).Scan(&w.ID, &w.Name, &w.OwnerID, &w.CreatedAt)
	if err != nil {
		return nil, err
	}
	_, err = r.pool.Exec(ctx,
		`INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1, $2, 'owner')`,
		w.ID, ownerID,
	)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func (r *Repository) ListByUser(ctx context.Context, userID string) ([]Workspace, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT w.id, w.name, w.owner_id, w.created_at
		 FROM workspaces w
		 JOIN workspace_members wm ON w.id = wm.workspace_id
		 WHERE wm.user_id = $1
		 ORDER BY w.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var workspaces []Workspace
	for rows.Next() {
		var w Workspace
		if err := rows.Scan(&w.ID, &w.Name, &w.OwnerID, &w.CreatedAt); err != nil {
			return nil, err
		}
		workspaces = append(workspaces, w)
	}
	return workspaces, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*Workspace, error) {
	w := &Workspace{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, owner_id, created_at FROM workspaces WHERE id = $1`,
		id,
	).Scan(&w.ID, &w.Name, &w.OwnerID, &w.CreatedAt)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func (r *Repository) Update(ctx context.Context, id, name string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE workspaces SET name = $1 WHERE id = $2`,
		name, id,
	)
	return err
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	// cascade deletes workspace_members
	_, err := r.pool.Exec(ctx, `DELETE FROM workspaces WHERE id = $1`, id)
	return err
}

func (r *Repository) IsMember(ctx context.Context, workspaceID, userID string) (bool, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM workspace_members WHERE workspace_id = $1 AND user_id = $2`,
		workspaceID, userID,
	).Scan(&count)
	return count > 0, err
}

func (r *Repository) AddMember(ctx context.Context, workspaceID, userID, role string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1, $2, $3)
		 ON CONFLICT DO NOTHING`,
		workspaceID, userID, role,
	)
	return err
}

func (r *Repository) RemoveMember(ctx context.Context, workspaceID, userID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM workspace_members WHERE workspace_id = $1 AND user_id = $2`,
		workspaceID, userID,
	)
	return err
}
```

- [ ] **Step 3: Create service.go**

```go
package workspace

import (
	"context"
	"errors"
)

var (
	ErrNotFound = errors.New("workspace not found")
	ErrNotOwner = errors.New("only the owner can perform this action")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, name, ownerID string) (*Workspace, error) {
	return s.repo.Create(ctx, name, ownerID)
}

func (s *Service) List(ctx context.Context, userID string) ([]Workspace, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *Service) Get(ctx context.Context, id string) (*Workspace, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id, name, userID string) error {
	ok, err := s.repo.IsMember(ctx, id, userID)
	if err != nil || !ok {
		return ErrNotFound
	}
	return s.repo.Update(ctx, id, name)
}

func (s *Service) Delete(ctx context.Context, id, userID string) error {
	w, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrNotFound
	}
	if w.OwnerID != userID {
		return ErrNotOwner
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) IsMember(ctx context.Context, workspaceID, userID string) (bool, error) {
	return s.repo.IsMember(ctx, workspaceID, userID)
}

func (s *Service) InviteMember(ctx context.Context, workspaceID, email, role, requesterID string) error {
	w, err := s.repo.GetByID(ctx, workspaceID)
	if err != nil {
		return ErrNotFound
	}
	if w.OwnerID != requesterID {
		return ErrNotOwner
	}
	// For now, accept any email — in production, look up user by email
	return nil
}

func (s *Service) RemoveMember(ctx context.Context, workspaceID, memberID, requesterID string) error {
	w, err := s.repo.GetByID(ctx, workspaceID)
	if err != nil {
		return ErrNotFound
	}
	if w.OwnerID != requesterID {
		return ErrNotOwner
	}
	return s.repo.RemoveMember(ctx, workspaceID, memberID)
}
```

- [ ] **Step 4: Create handler.go**

```go
package workspace

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func respond(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func respondError(w http.ResponseWriter, status int, msg string) {
	respond(w, status, map[string]string{"error": msg})
}

func userIDFromContext(r *http.Request) string {
	return r.Context().Value("user_id").(string)
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/workspaces", h.List)
	r.Post("/workspaces", h.Create)
	r.Get("/workspaces/{workspaceId}", h.Get)
	r.Patch("/workspaces/{workspaceId}", h.Update)
	r.Delete("/workspaces/{workspaceId}", h.Delete)
	r.Post("/workspaces/{workspaceId}/members", h.InviteMember)
	r.Delete("/workspaces/{workspaceId}/members/{userId}", h.RemoveMember)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r)
	workspaces, err := h.svc.List(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respond(w, http.StatusOK, workspaces)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r)
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		req.Name = "My Workspace"
	}
	ws, err := h.svc.Create(r.Context(), req.Name, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respond(w, http.StatusCreated, ws)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "workspaceId")
	ws, err := h.svc.Get(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "workspace not found")
		return
	}
	respond(w, http.StatusOK, ws)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "workspaceId")
	userID := userIDFromContext(r)
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.Update(r.Context(), id, req.Name, userID); err != nil {
		respondError(w, http.StatusForbidden, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "workspaceId")
	userID := userIDFromContext(r)
	if err := h.svc.Delete(r.Context(), id, userID); err != nil {
		respondError(w, http.StatusForbidden, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) InviteMember(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "workspaceId")
	userID := userIDFromContext(r)
	var req InviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.InviteMember(r.Context(), id, req.Email, req.Role, userID); err != nil {
		respondError(w, http.StatusForbidden, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "workspaceId")
	memberID := chi.URLParam(r, "userId")
	userID := userIDFromContext(r)
	if err := h.svc.RemoveMember(r.Context(), id, memberID, userID); err != nil {
		respondError(w, http.StatusForbidden, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
```

- [ ] **Step 5: Verify build**

Run: `cd /Users/ionutale/developer/notion-clone-golang && go build ./...`
Expected: Build succeeds.

- [ ] **Step 6: Commit**

```bash
git add internal/workspace/
git commit -m "feat: add workspace CRUD and member management"
```

### Task 6: Re-scope existing block/page routes under workspace prefix + workspace middleware

**Files:**
- Modify: `internal/api.go` — add auth/workspace routes, re-scope block routes
- Modify: `main.go` — pass new dependencies to MountAPI
- Create: `internal/middleware/workspace.go`

- [ ] **Step 1: Create workspace middleware**

```go
package middleware

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type WorkspaceValidator interface {
	IsMember(ctx context.Context, workspaceID, userID string) (bool, error)
}

func WorkspaceMiddleware(validator WorkspaceValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			workspaceID := chi.URLParam(r, "workspaceId")
			userID := r.Context().Value("user_id").(string)
			ok, err := validator.IsMember(r.Context(), workspaceID, userID)
			if err != nil || !ok {
				http.Error(w, `{"error":"workspace not found"}`, http.StatusNotFound)
				return
			}
			ctx := context.WithValue(r.Context(), "workspace_id", workspaceID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
```

- [ ] **Step 2: Update internal/api.go — re-scope routes**

Update the `MountAPI` signature and body to accept auth + workspace services and register all routes:

```go
package internal

import (
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ionutale/notion-clone-golang/internal/auth"
	"github.com/ionutale/notion-clone-golang/internal/block"
	"github.com/ionutale/notion-clone-golang/internal/middleware"
	"github.com/ionutale/notion-clone-golang/internal/storage"
	ws "github.com/ionutale/notion-clone-golang/internal/workspace"
)

func MountAPI(
	r chi.Router,
	blockSvc *block.Service,
	fileStore storage.FileStore,
	authSvc *auth.Service,
	wsSvc *ws.Service,
) {
	authH := auth.NewHandler(authSvc)
	wsH := ws.NewHandler(wsSvc)
	blockH := block.NewHandler(blockSvc)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		})

		// Public auth routes
		authH.RegisterRoutes(r)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(authSvc))

			// Workspace routes (list, create, etc.)
			wsH.RegisterRoutes(r)

			// Workspace-scoped block/page CRUD
			r.Route("/workspaces/{workspaceId}", func(r chi.Router) {
				r.Use(middleware.WorkspaceMiddleware(wsSvc))
				blockH.RegisterRoutes(r)
			})
		})

		// Upload (protected)
		r.With(middleware.AuthMiddleware(authSvc)).Post("/uploads", func(w http.ResponseWriter, r *http.Request) {
			r.ParseMultipartForm(10 << 20)
			file, header, err := r.FormFile("file")
			if err != nil {
				http.Error(w, `{"error":"missing file"}`, http.StatusBadRequest)
				return
			}
			defer file.Close()

			ext := path.Ext(header.Filename)
			key := uuid.New().String() + ext

			if err := fileStore.Put(r.Context(), key, file); err != nil {
				http.Error(w, `{"error":"upload failed"}`, http.StatusInternalServerError)
				return
			}

			url := fileStore.PublicURL(key)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{"url": url})
		})
	})
}
```

Note: add `"path"` to the import list.

- [ ] **Step 3: Update block handler — add RegisterRoutes method**

In `internal/block/handler.go`, add a `RegisterRoutes` method:

```go
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/pages", h.ListPages)
	r.Post("/pages", h.CreatePage)
	r.Get("/pages/{id}", h.GetPageTree)
	r.Post("/blocks", h.CreateBlock)
	r.Patch("/blocks/{id}", h.UpdateBlock)
	r.Delete("/blocks/{id}", h.DeleteBlock)
	r.Patch("/blocks/{id}/restore", h.RestoreBlock)
	r.Patch("/blocks/{id}/move", h.MoveBlock)
}
```

- [ ] **Step 4: Update main.go — construct new services, pass to MountAPI**

In `main.go`, construct auth and workspace services and pass them to `MountAPI`. Remove `internal` import and replace with direct imports.

```go
import (
	// ... existing imports plus:
	"github.com/ionutale/notion-clone-golang/internal/auth"
	"github.com/ionutale/notion-clone-golang/internal/workspace"
)

// In main(), change from:
//     internal.MountAPI(r, blockSvc, fileStore)
// To:
var wsSvc *workspace.Service
var authSvc *auth.Service
if pool != nil {
	wsRepo := workspace.NewRepository(pool)
	wsSvc = workspace.NewService(wsRepo)

	authRepo := auth.NewRepository(pool)
	authSvc = auth.NewService(authRepo, wsSvc)
}
// Then:
internal.MountAPI(r, blockSvc, fileStore, authSvc, wsSvc)
```

- [ ] **Step 5: Verify build**

Run: `cd /Users/ionutale/developer/notion-clone-golang && go build ./...`
Expected: Build succeeds.

- [ ] **Step 6: Commit**

```bash
git add internal/api.go internal/block/handler.go internal/middleware/workspace.go main.go
git commit -m "refactor: re-scope block/page routes under workspace prefix, add auth + workspace middleware"
```

### Task 7: Frontend auth store

**Files:**
- Create: `web/src/lib/stores/auth.svelte.ts`

- [ ] **Step 1: Create auth store**

```ts
import { api } from '$lib/api';
import type { User } from '$lib/types';

class AuthStore {
  user = $state<User | null>(null);
  accessToken = $state<string | null>(null);
  loading = $state(true);

  async login(email: string, password: string) {
    const res = await api.request('POST', '/api/v1/auth/login', { email, password });
    this.user = res.user;
    this.accessToken = res.access_token;
  }

  async signup(email: string, password: string, name: string) {
    const res = await api.request('POST', '/api/v1/auth/signup', { email, password, name });
    this.user = res.user;
    this.accessToken = res.access_token;
  }

  async logout() {
    await api.request('POST', '/api/v1/auth/logout');
    this.user = null;
    this.accessToken = null;
  }

  async refresh() {
    try {
      const res = await api.request('POST', '/api/v1/auth/refresh');
      this.accessToken = res.access_token;
      this.user = res.user;
    } catch {
      this.user = null;
      this.accessToken = null;
    }
  }

  async check() {
    this.loading = true;
    await this.refresh();
    this.loading = false;
  }
}

export const authStore = new AuthStore();
```

- [ ] **Step 2: Add User type to types.ts**

```ts
export interface User {
  id: string;
  email: string;
  name: string;
  created_at: string;
}
```

- [ ] **Step 3: Commit**

```bash
git add web/src/lib/stores/auth.svelte.ts web/src/lib/types.ts
git commit -m "feat: add frontend auth store with login/signup/logout/refresh"
```

### Task 8: Frontend workspace store

**Files:**
- Create: `web/src/lib/stores/workspaces.svelte.ts`

- [ ] **Step 1: Create workspace store**

```ts
import { api } from '$lib/api';
import { blockStore } from './blocks.svelte';

interface Workspace {
  id: string;
  name: string;
  owner_id: string;
  created_at: string;
}

class WorkspaceStore {
  workspaces = $state<Workspace[]>([]);
  activeWorkspaceId = $state<string | null>(null);

  get activeWorkspace() {
    return this.workspaces.find(w => w.id === this.activeWorkspaceId) ?? null;
  }

  async load() {
    const ws = await api.request('GET', '/api/v1/workspaces');
    this.workspaces = ws;
    if (ws.length > 0 && !this.activeWorkspaceId) {
      this.activeWorkspaceId = ws[0].id;
    }
  }

  async create(name: string) {
    const ws = await api.request('POST', '/api/v1/workspaces', { name });
    this.workspaces = [...this.workspaces, ws];
    this.activeWorkspaceId = ws.id;
  }

  async switchWorkspace(id: string) {
    this.activeWorkspaceId = id;
    // Reload pages for this workspace — will need updated API
  }
}

export const workspaceStore = new WorkspaceStore();
```

- [ ] **Step 2: Commit**

```bash
git add web/src/lib/stores/workspaces.svelte.ts
git commit -m "feat: add frontend workspace store with list/create/switch"
```

### Task 9: Frontend API client — Bearer token + auto-refresh

**Files:**
- Modify: `web/src/lib/api.ts`

- [ ] **Step 1: Update api.ts**

Add authorization header injection and auto-refresh on 401:

```ts
import { authStore } from '$lib/stores/auth.svelte';

class ApiClient {
  private async request<T>(method: string, path: string, body?: any): Promise<T> {
    const opts: RequestInit = { method };
    if (body !== undefined) {
      opts.headers = { 'Content-Type': 'application/json' };
      opts.body = JSON.stringify(body);
    }
    
    // Add auth header if token available
    if (authStore.accessToken) {
      opts.headers = { ...opts.headers as any, 'Authorization': `Bearer ${authStore.accessToken}` };
    }
    
    const res = await fetch(path, opts);
    
    // Auto-refresh on 401
    if (res.status === 401 && authStore.accessToken) {
      await authStore.refresh();
      if (authStore.accessToken) {
        opts.headers = { ...opts.headers as any, 'Authorization': `Bearer ${authStore.accessToken}` };
        const retryRes = await fetch(path, opts);
        if (!retryRes.ok) throw new ApiError(retryRes.status, await retryRes.text());
        if (retryRes.status === 204) return undefined as T;
        return retryRes.json();
      }
    }
    
    if (!res.ok) throw new ApiError(res.status, await res.text());
    if (res.status === 204) return undefined as T;
    return res.json();
  }
}
```

Also prepend `/api/v1` in the BASE_URL or in the request method so that API paths like `/auth/login` resolve to `/api/v1/auth/login`. Update the method signature to accept the full path:

```ts
  private async request<T>(method: string, path: string, body?: any): Promise<T> {
    const opts: RequestInit = { method };
    if (body !== undefined) {
      opts.headers = { 'Content-Type': 'application/json' };
      opts.body = JSON.stringify(body);
    }
    
    // Add auth header if token available
    if (authStore.accessToken) {
      opts.headers = { ...opts.headers as any, 'Authorization': `Bearer ${authStore.accessToken}` };
    }
    
    const res = await fetch(`/api/v1${path}`, opts);
    
    // Auto-refresh on 401
    if (res.status === 401 && authStore.accessToken) {
      await authStore.refresh();
      if (authStore.accessToken) {
        opts.headers = { ...opts.headers as any, 'Authorization': `Bearer ${authStore.accessToken}` };
        const retryRes = await fetch(`/api/v1${path}`, opts);
        if (!retryRes.ok) throw new ApiError(retryRes.status, await retryRes.text());
        if (retryRes.status === 204) return undefined as T;
        return retryRes.json();
      }
    }
    
    if (!res.ok) throw new ApiError(res.status, await res.text());
    if (res.status === 204) return undefined as T;
    return res.json();
  }
```

Also update all existing API client method paths to remove `/api/v1` prefix since the request method now adds it. For example:
```ts
  createPage(title = 'Untitled'): Promise<Block> {
    return this.request('POST', '/pages', { title });
  }
```

- [ ] **Step 2: Commit**

```bash
git add web/src/lib/api.ts
git commit -m "feat: add Bearer token header and auto-refresh interceptor to API client"
```

### Task 10: Login and signup pages

**Files:**
- Create: `web/src/routes/login/+page.svelte`
- Create: `web/src/routes/signup/+page.svelte`

- [ ] **Step 1: Create login page**

```svelte
<script lang="ts">
  import { authStore } from '$lib/stores/auth.svelte';
  import { goto } from '$app/navigation';

  let email = $state('');
  let password = $state('');
  let error = $state('');
  let loading = $state(false);

  async function handleSubmit(e: Event) {
    e.preventDefault();
    error = '';
    loading = true;
    try {
      await authStore.login(email, password);
      goto('/');
    } catch (err: any) {
      error = err.message ?? 'Login failed';
    } finally {
      loading = false;
    }
  }
</script>

<div class="min-h-screen flex items-center justify-center bg-base-200">
  <div class="card w-full max-w-sm bg-base-100 shadow-xl">
    <div class="card-body">
      <h2 class="card-title text-2xl mb-2">Log in</h2>
      <form onsubmit={handleSubmit}>
        <input bind:value={email} type="email" placeholder="Email" class="input input-bordered w-full mb-3" required />
        <input bind:value={password} type="password" placeholder="Password" class="input input-bordered w-full mb-3" required />
        {#if error}
          <div class="alert alert-error text-sm py-2 mb-3">{error}</div>
        {/if}
        <button type="submit" class="btn btn-primary w-full" disabled={loading}>
          {loading ? 'Logging in...' : 'Log in'}
        </button>
      </form>
      <p class="text-sm text-center mt-4 text-base-content/60">
        Don't have an account? <a href="/signup" class="link link-primary">Sign up</a>
      </p>
    </div>
  </div>
</div>
```

- [ ] **Step 2: Create signup page**

```svelte
<script lang="ts">
  import { authStore } from '$lib/stores/auth.svelte';
  import { goto } from '$app/navigation';

  let email = $state('');
  let password = $state('');
  let name = $state('');
  let error = $state('');
  let loading = $state(false);

  async function handleSubmit(e: Event) {
    e.preventDefault();
    error = '';
    loading = true;
    try {
      await authStore.signup(email, password, name);
      goto('/');
    } catch (err: any) {
      error = err.message ?? 'Signup failed';
    } finally {
      loading = false;
    }
  }
</script>

<div class="min-h-screen flex items-center justify-center bg-base-200">
  <div class="card w-full max-w-sm bg-base-100 shadow-xl">
    <div class="card-body">
      <h2 class="card-title text-2xl mb-2">Sign up</h2>
      <form onsubmit={handleSubmit}>
        <input bind:value={name} type="text" placeholder="Name" class="input input-bordered w-full mb-3" required />
        <input bind:value={email} type="email" placeholder="Email" class="input input-bordered w-full mb-3" required />
        <input bind:value={password} type="password" placeholder="Password" class="input input-bordered w-full mb-3" required />
        {#if error}
          <div class="alert alert-error text-sm py-2 mb-3">{error}</div>
        {/if}
        <button type="submit" class="btn btn-primary w-full" disabled={loading}>
          {loading ? 'Creating account...' : 'Sign up'}
        </button>
      </form>
      <p class="text-sm text-center mt-4 text-base-content/60">
        Already have an account? <a href="/login" class="link link-primary">Log in</a>
      </p>
    </div>
  </div>
</div>
```

- [ ] **Step 3: Commit**

```bash
git add web/src/routes/login/ web/src/routes/signup/
git commit -m "feat: add login and signup pages"
```

### Task 11: Auth gate in layout

**Files:**
- Modify: `web/src/routes/+layout.svelte`

- [ ] **Step 1: Add auth gate**

In the layout's script, add:

```ts
  import { authStore } from '$lib/stores/auth.svelte';
  import { workspaceStore } from '$lib/stores/workspaces.svelte';
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';

  const publicPaths = ['/login', '/signup'];

  onMount(async () => {
    await authStore.check();
    if (!authStore.user && !publicPaths.includes($page.url.pathname)) {
      goto('/login');
    }
    if (authStore.user) {
      await workspaceStore.load();
    }
  });
```

Wrap the existing content with an auth guard:

```svelte
{#if authStore.loading}
  <div class="flex justify-center items-center min-h-screen">
    <span class="loading loading-spinner loading-lg text-primary"></span>
  </div>
{:else if !authStore.user && !publicPaths.includes($page.url.pathname)}
  <!-- Will redirect via onMount -->
{:else}
  {children}
{/if}
```

- [ ] **Step 2: Commit**

```bash
git add web/src/routes/+layout.svelte
git commit -m "feat: add auth gate to layout, redirect unauthenticated users"
```

### Task 12: Sidebar workspace switcher

**Files:**
- Modify: `web/src/lib/components/Sidebar.svelte`

- [ ] **Step 1: Add workspace switcher to sidebar**

In the Sidebar, add a workspace dropdown at the top:

```svelte
<script lang="ts">
  import { workspaceStore } from '$lib/stores/workspaces.svelte';
  import { authStore } from '$lib/stores/auth.svelte';
  import { goto } from '$app/navigation';

  let dropdownOpen = $state(false);
</script>

<div class="p-3">
  <!-- Workspace Switcher -->
  <div class="relative mb-4">
    <button
      onclick={() => dropdownOpen = !dropdownOpen}
      class="w-full flex items-center justify-between px-3 py-2 bg-base-200 rounded-lg hover:bg-base-300 transition-colors text-sm font-medium"
    >
      <span>{workspaceStore.activeWorkspace?.name ?? 'Select workspace'}</span>
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
      </svg>
    </button>
    {#if dropdownOpen}
      <div class="absolute top-full left-0 right-0 mt-1 bg-base-100 border border-base-300 rounded-lg shadow-xl z-50 py-1">
        {#each workspaceStore.workspaces as ws}
          <button
            onclick={() => { workspaceStore.switchWorkspace(ws.id); dropdownOpen = false; }}
            class:bg-base-200={ws.id === workspaceStore.activeWorkspaceId}
            class="w-full text-left px-3 py-2 text-sm hover:bg-base-200"
          >
            {ws.name}
          </button>
        {/each}
        <hr class="border-base-200 my-1">
        <button
          onclick={/* create workspace */ ; dropdownOpen = false}
          class="w-full text-left px-3 py-2 text-sm text-primary hover:bg-base-200"
        >
          + New workspace
        </button>
      </div>
    {/if}
  </div>

  <!-- Logout -->
  <button
    onclick={async () => { await authStore.logout(); goto('/login'); }}
    class="w-full text-left px-3 py-2 text-sm text-base-content/50 hover:text-error transition-colors"
  >
    Log out
  </button>
</div>
```

- [ ] **Step 2: Commit**

```bash
git add web/src/lib/components/Sidebar.svelte
git commit -m "feat: add workspace switcher dropdown to sidebar"
```
