package workspace

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/ionutale/notion-clone-golang/internal/auth"
	"github.com/ionutale/notion-clone-golang/internal/httputil"
)

type Handler struct {
	svc     *Service
	authSvc *auth.Service
}

func NewHandler(svc *Service, authSvc *auth.Service) *Handler {
	return &Handler{svc: svc, authSvc: authSvc}
}



func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/workspaces", h.List)
	r.Post("/workspaces", h.Create)
	r.Get("/workspaces/{workspaceId}", h.Get)
	r.Patch("/workspaces/{workspaceId}", h.Update)
	r.Delete("/workspaces/{workspaceId}", h.Delete)
	r.Get("/workspaces/{workspaceId}/members", h.ListMembers)
	r.Post("/workspaces/{workspaceId}/members", h.InviteMember)
	r.Delete("/workspaces/{workspaceId}/members/{userId}", h.RemoveMember)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.CtxUserID).(string)
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	workspaces, err := h.svc.List(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, workspaces)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.CtxUserID).(string)
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		req.Name = "My Workspace"
	}
	ws, err := h.svc.Create(r.Context(), req.Name, userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.JSON(w, http.StatusCreated, ws)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "workspaceId")
	userID, ok := r.Context().Value(auth.CtxUserID).(string)
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	member, err := h.svc.IsMember(r.Context(), id, userID)
	if err != nil || !member {
		httputil.Error(w, http.StatusNotFound, "workspace not found")
		return
	}
	ws, err := h.svc.Get(r.Context(), id)
	if err != nil || ws == nil {
		httputil.Error(w, http.StatusNotFound, "workspace not found")
		return
	}
	httputil.JSON(w, http.StatusOK, ws)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "workspaceId")
	userID, ok := r.Context().Value(auth.CtxUserID).(string)
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.Update(r.Context(), id, req.Name, userID); err != nil {
		httputil.Error(w, http.StatusForbidden, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "workspaceId")
	userID, ok := r.Context().Value(auth.CtxUserID).(string)
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	if err := h.svc.Delete(r.Context(), id, userID); err != nil {
		httputil.Error(w, http.StatusForbidden, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) InviteMember(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "workspaceId")
	userID, ok := r.Context().Value(auth.CtxUserID).(string)
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	var req InviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	memberID := req.UserID
	if memberID == "" && req.Email != "" {
		if !strings.Contains(req.Email, "@") {
			httputil.Error(w, http.StatusBadRequest, "valid email is required")
			return
		}
		if h.authSvc == nil {
			httputil.Error(w, http.StatusInternalServerError, "auth service unavailable")
			return
		}
		user, err := h.authSvc.GetUserByEmail(r.Context(), req.Email)
		if err != nil {
			httputil.Error(w, http.StatusNotFound, "user not found")
			return
		}
		memberID = user.ID
	}
	if memberID == "" {
		httputil.Error(w, http.StatusBadRequest, "user_id or email is required")
		return
	}

	if req.Role == "" {
		req.Role = "member"
	}

	if err := h.svc.InviteMember(r.Context(), id, memberID, req.Role, userID); err != nil {
		httputil.Error(w, http.StatusForbidden, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListMembers(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "workspaceId")
	userID, ok := r.Context().Value(auth.CtxUserID).(string)
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	member, err := h.svc.IsMember(r.Context(), id, userID)
	if err != nil || !member {
		httputil.Error(w, http.StatusNotFound, "workspace not found")
		return
	}
	members, err := h.svc.ListMembers(r.Context(), id)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, members)
}

func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "workspaceId")
	memberID := chi.URLParam(r, "userId")
	userID, ok := r.Context().Value(auth.CtxUserID).(string)
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	if memberID == userID {
		httputil.Error(w, http.StatusBadRequest, "cannot remove yourself")
		return
	}
	if err := h.svc.RemoveMember(r.Context(), id, memberID, userID); err != nil {
		httputil.Error(w, http.StatusForbidden, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
