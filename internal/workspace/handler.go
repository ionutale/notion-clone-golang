package workspace

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/ionutale/notion-clone-golang/internal/auth"
)

type Handler struct {
	svc     *Service
	authSvc *auth.Service
}

func NewHandler(svc *Service, authSvc *auth.Service) *Handler {
	return &Handler{svc: svc, authSvc: authSvc}
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
	userID := r.Context().Value(auth.CtxUserID).(string)
	workspaces, err := h.svc.List(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respond(w, http.StatusOK, workspaces)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.CtxUserID).(string)
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
	if err != nil || ws == nil {
		respondError(w, http.StatusNotFound, "workspace not found")
		return
	}
	respond(w, http.StatusOK, ws)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "workspaceId")
	userID := r.Context().Value(auth.CtxUserID).(string)
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
	userID := r.Context().Value(auth.CtxUserID).(string)
	if err := h.svc.Delete(r.Context(), id, userID); err != nil {
		respondError(w, http.StatusForbidden, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) InviteMember(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "workspaceId")
	userID := r.Context().Value(auth.CtxUserID).(string)
	var req InviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	memberID := req.UserID
	if memberID == "" && req.Email != "" {
		user, err := h.authSvc.GetUserByEmail(r.Context(), req.Email)
		if err != nil {
			respondError(w, http.StatusNotFound, "user not found")
			return
		}
		memberID = user.ID
	}

	if req.Role == "" {
		req.Role = "member"
	}

	if err := h.svc.InviteMember(r.Context(), id, memberID, req.Role, userID); err != nil {
		respondError(w, http.StatusForbidden, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListMembers(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "workspaceId")
	members, err := h.svc.ListMembers(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respond(w, http.StatusOK, members)
}

func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "workspaceId")
	memberID := chi.URLParam(r, "userId")
	userID := r.Context().Value(auth.CtxUserID).(string)
	if err := h.svc.RemoveMember(r.Context(), id, memberID, userID); err != nil {
		respondError(w, http.StatusForbidden, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
