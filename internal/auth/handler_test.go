package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func newRouter(h *Handler) chi.Router {
	r := chi.NewRouter()
	h.RegisterRoutes(r)
	return r
}

func TestSignup_Success(t *testing.T) {
	mockRepo := &MockUserRepo{
		CreateUserFn: func(ctx context.Context, email, password, name string) (*User, error) {
			return &User{ID: "user-1", Email: email, Name: name, CreatedAt: time.Now()}, nil
		},
		CreateRefreshTokenFn: func(ctx context.Context, userID, expiresAt string) (string, error) {
			return "refresh-token-1", nil
		},
	}
	mockWSCreator := &MockWorkspaceCreator{
		CreateFn: func(ctx context.Context, name, ownerID string) (interface{}, error) {
			return map[string]interface{}{"id": "ws-1", "name": name}, nil
		},
	}

	svc := NewService(mockRepo, mockWSCreator)
	h := NewHandler(svc)
	router := newRouter(h)

	body := SignupRequest{Email: "test@test.com", Password: "password123", Name: "Test User"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/auth/signup", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	var resp AuthResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "test@test.com", resp.User.Email)
	assert.Equal(t, "Test User", resp.User.Name)
	assert.NotEmpty(t, resp.AccessToken)

	cookies := rec.Result().Cookies()
	var hasRefresh bool
	for _, c := range cookies {
		if c.Name == refreshCookieName {
			hasRefresh = true
			assert.Equal(t, "refresh-token-1", c.Value)
		}
	}
	assert.True(t, hasRefresh, "refresh cookie should be set")
}

func TestSignup_DuplicateEmail(t *testing.T) {
	mockRepo := &MockUserRepo{
		CreateUserFn: func(ctx context.Context, email, password, name string) (*User, error) {
			return nil, ErrEmailTaken
		},
	}
	mockWSCreator := &MockWorkspaceCreator{}

	svc := NewService(mockRepo, mockWSCreator)
	h := NewHandler(svc)
	router := newRouter(h)

	body := SignupRequest{Email: "taken@test.com", Password: "password123", Name: "Test"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/auth/signup", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
	var errResp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &errResp)
	assert.Contains(t, errResp["error"], "email already in use")
}

func TestSignup_MissingFields(t *testing.T) {
	mockRepo := &MockUserRepo{}
	mockWSCreator := &MockWorkspaceCreator{}
	svc := NewService(mockRepo, mockWSCreator)
	h := NewHandler(svc)
	router := newRouter(h)

	body := []byte(`{"email": "test@test.com"}`)
	req := httptest.NewRequest("POST", "/auth/signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestSignup_InvalidJSON(t *testing.T) {
	mockRepo := &MockUserRepo{}
	mockWSCreator := &MockWorkspaceCreator{}
	svc := NewService(mockRepo, mockWSCreator)
	h := NewHandler(svc)
	router := newRouter(h)

	req := httptest.NewRequest("POST", "/auth/signup", bytes.NewReader([]byte(`{invalid`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestLogin_Success(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	mockRepo := &MockUserRepo{
		GetUserByEmailFn: func(ctx context.Context, email string) (*User, error) {
			return &User{
				ID:           "user-1",
				Email:        email,
				Name:         "Test User",
				PasswordHash: string(hash),
				CreatedAt:    time.Now(),
			}, nil
		},
		CreateRefreshTokenFn: func(ctx context.Context, userID, expiresAt string) (string, error) {
			return "refresh-token-login", nil
		},
	}
	mockWSCreator := &MockWorkspaceCreator{}
	svc := NewService(mockRepo, mockWSCreator)
	h := NewHandler(svc)
	router := newRouter(h)

	body := LoginRequest{Email: "test@test.com", Password: "password123"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp AuthResponse
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "test@test.com", resp.User.Email)
	assert.NotEmpty(t, resp.AccessToken)

	cookies := rec.Result().Cookies()
	var hasRefresh bool
	for _, c := range cookies {
		if c.Name == refreshCookieName {
			hasRefresh = true
			assert.Equal(t, "refresh-token-login", c.Value)
		}
	}
	assert.True(t, hasRefresh, "refresh cookie should be set")
}

func TestLogin_WrongPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	require.NoError(t, err)

	mockRepo := &MockUserRepo{
		GetUserByEmailFn: func(ctx context.Context, email string) (*User, error) {
			return &User{
				ID:           "user-1",
				Email:        email,
				Name:         "Test User",
				PasswordHash: string(hash),
			}, nil
		},
	}
	mockWSCreator := &MockWorkspaceCreator{}
	svc := NewService(mockRepo, mockWSCreator)
	h := NewHandler(svc)
	router := newRouter(h)

	body := LoginRequest{Email: "test@test.com", Password: "wrong-password"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestLogin_NonExistentEmail(t *testing.T) {
	mockRepo := &MockUserRepo{
		GetUserByEmailFn: func(ctx context.Context, email string) (*User, error) {
			return nil, ErrInvalidCredentials
		},
	}
	mockWSCreator := &MockWorkspaceCreator{}
	svc := NewService(mockRepo, mockWSCreator)
	h := NewHandler(svc)
	router := newRouter(h)

	body := LoginRequest{Email: "nonexistent@test.com", Password: "password123"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	var errResp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &errResp)
	assert.Contains(t, errResp["error"], "invalid email or password")
}

func TestRefresh_Success(t *testing.T) {
	mockRepo := &MockUserRepo{
		GetUserByRefreshTokenFn: func(ctx context.Context, tokenHex string) (*User, error) {
			return &User{ID: "user-1", Email: "test@test.com", Name: "Test", CreatedAt: time.Now()}, nil
		},
		CreateRefreshTokenFn: func(ctx context.Context, userID, expiresAt string) (string, error) {
			return "new-refresh-token", nil
		},
	}
	mockWSCreator := &MockWorkspaceCreator{}
	svc := NewService(mockRepo, mockWSCreator)
	h := NewHandler(svc)
	router := newRouter(h)

	req := httptest.NewRequest("POST", "/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: refreshCookieName, Value: "valid-refresh-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp AuthResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "test@test.com", resp.User.Email)
	assert.NotEmpty(t, resp.AccessToken)
}

func TestRefresh_MissingCookie(t *testing.T) {
	mockRepo := &MockUserRepo{}
	mockWSCreator := &MockWorkspaceCreator{}
	svc := NewService(mockRepo, mockWSCreator)
	h := NewHandler(svc)
	router := newRouter(h)

	req := httptest.NewRequest("POST", "/auth/refresh", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	var errResp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &errResp)
	assert.Contains(t, errResp["error"], "no refresh token")
}

func TestRefresh_InvalidToken(t *testing.T) {
	mockRepo := &MockUserRepo{
		GetUserByRefreshTokenFn: func(ctx context.Context, tokenHex string) (*User, error) {
			return nil, ErrInvalidCredentials
		},
	}
	mockWSCreator := &MockWorkspaceCreator{}
	svc := NewService(mockRepo, mockWSCreator)
	h := NewHandler(svc)
	router := newRouter(h)

	req := httptest.NewRequest("POST", "/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: refreshCookieName, Value: "expired-or-invalid-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	var errResp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &errResp)
	assert.Contains(t, errResp["error"], "invalid refresh token")
}

func TestLogout_Success(t *testing.T) {
	mockRepo := &MockUserRepo{
		DeleteUserRefreshTokensFn: func(ctx context.Context, userID string) error {
			return nil
		},
	}
	mockWSCreator := &MockWorkspaceCreator{}
	svc := NewService(mockRepo, mockWSCreator)
	h := NewHandler(svc)

	req := httptest.NewRequest("POST", "/auth/logout", nil)
	ctx := context.WithValue(req.Context(), CtxUserID, "user-1")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	h.Logout(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)

	cookies := rec.Result().Cookies()
	var hasClearCookie bool
	for _, c := range cookies {
		if c.Name == refreshCookieName && c.MaxAge < 0 {
			hasClearCookie = true
		}
	}
	assert.True(t, hasClearCookie, "refresh cookie should be cleared")
}

func TestLogout_Unauthenticated(t *testing.T) {
	mockRepo := &MockUserRepo{}
	mockWSCreator := &MockWorkspaceCreator{}
	svc := NewService(mockRepo, mockWSCreator)
	h := NewHandler(svc)

	req := httptest.NewRequest("POST", "/auth/logout", nil)
	rec := httptest.NewRecorder()

	h.Logout(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestMe_Authenticated(t *testing.T) {
	mockRepo := &MockUserRepo{
		GetUserByIDFn: func(ctx context.Context, id string) (*User, error) {
			return &User{ID: id, Email: "test@test.com", Name: "Test User", CreatedAt: time.Now()}, nil
		},
	}
	mockWSCreator := &MockWorkspaceCreator{}
	svc := NewService(mockRepo, mockWSCreator)
	h := NewHandler(svc)

	req := httptest.NewRequest("GET", "/auth/me", nil)
	ctx := context.WithValue(req.Context(), CtxUserID, "user-1")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	h.Me(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var user User
	err := json.Unmarshal(rec.Body.Bytes(), &user)
	require.NoError(t, err)
	assert.Equal(t, "test@test.com", user.Email)
	assert.Equal(t, "Test User", user.Name)
}

func TestMe_Unauthenticated(t *testing.T) {
	mockRepo := &MockUserRepo{}
	mockWSCreator := &MockWorkspaceCreator{}
	svc := NewService(mockRepo, mockWSCreator)
	h := NewHandler(svc)

	req := httptest.NewRequest("GET", "/auth/me", nil)
	rec := httptest.NewRecorder()

	h.Me(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
