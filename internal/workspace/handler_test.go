package workspace

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

	"github.com/ionutale/notion-clone-golang/internal/auth"
)

func userContext(userID string) context.Context {
	return context.WithValue(context.Background(), auth.CtxUserID, userID)
}

func newRouter(h *Handler) chi.Router {
	r := chi.NewRouter()
	h.RegisterRoutes(r)
	return r
}

func TestCreate_Success(t *testing.T) {
	mockRepo := &MockWorkspaceRepo{
		CreateFn: func(ctx context.Context, name, ownerID string) (*Workspace, error) {
			return &Workspace{ID: "ws-1", Name: name, OwnerID: ownerID, CreatedAt: time.Now()}, nil
		},
	}
	svc := NewService(mockRepo)
	h := NewHandler(svc, nil)

	body := CreateRequest{Name: "My Workspace"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/workspaces", bytes.NewReader(bodyBytes))
	req = req.WithContext(userContext("user-1"))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	var ws Workspace
	err := json.Unmarshal(rec.Body.Bytes(), &ws)
	require.NoError(t, err)
	assert.Equal(t, "My Workspace", ws.Name)
	assert.Equal(t, "user-1", ws.OwnerID)
}

func TestCreate_DefaultName(t *testing.T) {
	mockRepo := &MockWorkspaceRepo{
		CreateFn: func(ctx context.Context, name, ownerID string) (*Workspace, error) {
			return &Workspace{ID: "ws-1", Name: name, OwnerID: ownerID, CreatedAt: time.Now()}, nil
		},
	}
	svc := NewService(mockRepo)
	h := NewHandler(svc, nil)

	body := CreateRequest{}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/workspaces", bytes.NewReader(bodyBytes))
	req = req.WithContext(userContext("user-1"))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	var ws Workspace
	json.Unmarshal(rec.Body.Bytes(), &ws)
	assert.Equal(t, "My Workspace", ws.Name)
}

func TestList_Empty(t *testing.T) {
	mockRepo := &MockWorkspaceRepo{
		ListByUserFn: func(ctx context.Context, userID string) ([]Workspace, error) {
			return []Workspace{}, nil
		},
	}
	svc := NewService(mockRepo)
	h := NewHandler(svc, nil)

	req := httptest.NewRequest("GET", "/workspaces", nil)
	req = req.WithContext(userContext("user-1"))
	rec := httptest.NewRecorder()

	h.List(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var workspaces []Workspace
	err := json.Unmarshal(rec.Body.Bytes(), &workspaces)
	require.NoError(t, err)
	assert.Empty(t, workspaces)
}

func TestList_Multiple(t *testing.T) {
	now := time.Now()
	mockRepo := &MockWorkspaceRepo{
		ListByUserFn: func(ctx context.Context, userID string) ([]Workspace, error) {
			return []Workspace{
				{ID: "ws-1", Name: "First", OwnerID: userID, CreatedAt: now},
				{ID: "ws-2", Name: "Second", OwnerID: userID, CreatedAt: now.Add(-time.Hour)},
			}, nil
		},
	}
	svc := NewService(mockRepo)
	h := NewHandler(svc, nil)

	req := httptest.NewRequest("GET", "/workspaces", nil)
	req = req.WithContext(userContext("user-1"))
	rec := httptest.NewRecorder()

	h.List(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var workspaces []Workspace
	err := json.Unmarshal(rec.Body.Bytes(), &workspaces)
	require.NoError(t, err)
	assert.Len(t, workspaces, 2)
	assert.Equal(t, "First", workspaces[0].Name)
	assert.Equal(t, "Second", workspaces[1].Name)
}

func TestGet_Success(t *testing.T) {
	mockRepo := &MockWorkspaceRepo{
		GetByIDFn: func(ctx context.Context, id string) (*Workspace, error) {
			return &Workspace{ID: id, Name: "My Workspace", OwnerID: "user-1", CreatedAt: time.Now()}, nil
		},
	}
	svc := NewService(mockRepo)
	h := NewHandler(svc, nil)
	router := newRouter(h)

	req := httptest.NewRequest("GET", "/workspaces/ws-1", nil)
	req = req.WithContext(userContext("user-1"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var ws Workspace
	err := json.Unmarshal(rec.Body.Bytes(), &ws)
	require.NoError(t, err)
	assert.Equal(t, "My Workspace", ws.Name)
}

func TestGet_NotFound(t *testing.T) {
	mockRepo := &MockWorkspaceRepo{
		GetByIDFn: func(ctx context.Context, id string) (*Workspace, error) {
			return nil, nil
		},
	}
	svc := NewService(mockRepo)
	h := NewHandler(svc, nil)
	router := newRouter(h)

	req := httptest.NewRequest("GET", "/workspaces/ws-nonexistent", nil)
	req = req.WithContext(userContext("user-1"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestUpdate_Success(t *testing.T) {
	mockRepo := &MockWorkspaceRepo{
		GetByIDFn: func(ctx context.Context, id string) (*Workspace, error) {
			return &Workspace{ID: id, Name: "Old Name", OwnerID: "user-1", CreatedAt: time.Now()}, nil
		},
		UpdateFn: func(ctx context.Context, id, name string) error {
			return nil
		},
	}
	svc := NewService(mockRepo)
	h := NewHandler(svc, nil)
	router := newRouter(h)

	body := CreateRequest{Name: "New Name"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("PATCH", "/workspaces/ws-1", bytes.NewReader(bodyBytes))
	req = req.WithContext(userContext("user-1"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestUpdate_NotOwner(t *testing.T) {
	mockRepo := &MockWorkspaceRepo{
		GetByIDFn: func(ctx context.Context, id string) (*Workspace, error) {
			return &Workspace{ID: id, Name: "My Workspace", OwnerID: "other-user", CreatedAt: time.Now()}, nil
		},
	}
	svc := NewService(mockRepo)
	h := NewHandler(svc, nil)
	router := newRouter(h)

	body := CreateRequest{Name: "New Name"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("PATCH", "/workspaces/ws-1", bytes.NewReader(bodyBytes))
	req = req.WithContext(userContext("user-1"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestDelete_Success(t *testing.T) {
	mockRepo := &MockWorkspaceRepo{
		GetByIDFn: func(ctx context.Context, id string) (*Workspace, error) {
			return &Workspace{ID: id, Name: "My Workspace", OwnerID: "user-1", CreatedAt: time.Now()}, nil
		},
		DeleteFn: func(ctx context.Context, id string) error {
			return nil
		},
	}
	svc := NewService(mockRepo)
	h := NewHandler(svc, nil)
	router := newRouter(h)

	req := httptest.NewRequest("DELETE", "/workspaces/ws-1", nil)
	req = req.WithContext(userContext("user-1"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestDelete_NotOwner(t *testing.T) {
	mockRepo := &MockWorkspaceRepo{
		GetByIDFn: func(ctx context.Context, id string) (*Workspace, error) {
			return &Workspace{ID: id, Name: "My Workspace", OwnerID: "other-user", CreatedAt: time.Now()}, nil
		},
	}
	svc := NewService(mockRepo)
	h := NewHandler(svc, nil)
	router := newRouter(h)

	req := httptest.NewRequest("DELETE", "/workspaces/ws-1", nil)
	req = req.WithContext(userContext("user-1"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestInviteMember_Success(t *testing.T) {
	mockRepo := &MockWorkspaceRepo{
		GetByIDFn: func(ctx context.Context, id string) (*Workspace, error) {
			return &Workspace{ID: id, Name: "My Workspace", OwnerID: "user-1", CreatedAt: time.Now()}, nil
		},
		AddMemberFn: func(ctx context.Context, workspaceID, userID, role string) error {
			return nil
		},
	}
	svc := NewService(mockRepo)
	h := NewHandler(svc, nil)
	router := newRouter(h)

	body := InviteRequest{UserID: "user-2", Role: "member"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/workspaces/ws-1/members", bytes.NewReader(bodyBytes))
	req = req.WithContext(userContext("user-1"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestInviteMember_NotOwner(t *testing.T) {
	mockRepo := &MockWorkspaceRepo{
		GetByIDFn: func(ctx context.Context, id string) (*Workspace, error) {
			return &Workspace{ID: id, Name: "My Workspace", OwnerID: "other-user", CreatedAt: time.Now()}, nil
		},
	}
	svc := NewService(mockRepo)
	h := NewHandler(svc, nil)
	router := newRouter(h)

	body := InviteRequest{UserID: "user-2", Role: "member"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/workspaces/ws-1/members", bytes.NewReader(bodyBytes))
	req = req.WithContext(userContext("user-1"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestRemoveMember_Success(t *testing.T) {
	mockRepo := &MockWorkspaceRepo{
		GetByIDFn: func(ctx context.Context, id string) (*Workspace, error) {
			return &Workspace{ID: id, Name: "My Workspace", OwnerID: "user-1", CreatedAt: time.Now()}, nil
		},
		RemoveMemberFn: func(ctx context.Context, workspaceID, userID string) error {
			return nil
		},
	}
	svc := NewService(mockRepo)
	h := NewHandler(svc, nil)
	router := newRouter(h)

	req := httptest.NewRequest("DELETE", "/workspaces/ws-1/members/user-2", nil)
	req = req.WithContext(userContext("user-1"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestRemoveMember_NotOwner(t *testing.T) {
	mockRepo := &MockWorkspaceRepo{
		GetByIDFn: func(ctx context.Context, id string) (*Workspace, error) {
			return &Workspace{ID: id, Name: "My Workspace", OwnerID: "other-user", CreatedAt: time.Now()}, nil
		},
	}
	svc := NewService(mockRepo)
	h := NewHandler(svc, nil)
	router := newRouter(h)

	req := httptest.NewRequest("DELETE", "/workspaces/ws-1/members/user-2", nil)
	req = req.WithContext(userContext("user-1"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}
