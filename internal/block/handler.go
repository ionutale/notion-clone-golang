package block

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ionutale/notion-clone-golang/internal/auth"
	"github.com/ionutale/notion-clone-golang/internal/httputil"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func workspaceIDFromRequest(r *http.Request) (uuid.UUID, error) {
	id := chi.URLParam(r, "workspaceId")
	return uuid.Parse(id)
}

func userIDFromRequest(r *http.Request) (uuid.UUID, error) {
	id, ok := r.Context().Value(auth.CtxUserID).(string)
	if !ok {
		return uuid.Nil, nil
	}
	return uuid.Parse(id)
}

func (h *Handler) CreatePage(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Title == "" {
		req.Title = "Untitled"
	}
	wsID, err := workspaceIDFromRequest(r)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid workspace id")
		return
	}
	uID, err := userIDFromRequest(r)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}
	page, err := h.svc.CreatePage(r.Context(), wsID, uID, req.Title)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.JSON(w, http.StatusCreated, page)
}

func (h *Handler) GetPageTree(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid id")
		return
	}
	tree, err := h.svc.GetPageTree(r.Context(), id)
	if err != nil {
		httputil.Error(w, http.StatusNotFound, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, tree)
}

func (h *Handler) ListPages(w http.ResponseWriter, r *http.Request) {
	wsID, err := workspaceIDFromRequest(r)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid workspace id")
		return
	}
	pages, err := h.svc.ListPages(r.Context(), wsID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, pages)
}

func (h *Handler) CreateBlock(w http.ResponseWriter, r *http.Request) {
	var req CreateBlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	wsID, err := workspaceIDFromRequest(r)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid workspace id")
		return
	}
	uID, err := userIDFromRequest(r)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}
	block, err := h.svc.CreateBlock(r.Context(), wsID, uID, req)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httputil.JSON(w, http.StatusCreated, block)
}

func (h *Handler) UpdateBlock(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req UpdateBlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	block, err := h.svc.UpdateBlock(r.Context(), id, req)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, block)
}

func (h *Handler) DeleteBlock(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.svc.DeleteBlock(r.Context(), id); err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RestoreBlock(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid id")
		return
	}
	block, err := h.svc.RestoreBlock(r.Context(), id)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, block)
}

func (h *Handler) MoveBlock(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req MoveBlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	wsID, err := workspaceIDFromRequest(r)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid workspace id")
		return
	}
	block, err := h.svc.MoveBlock(r.Context(), wsID, id, req)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, block)
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		httputil.Error(w, http.StatusBadRequest, "query parameter 'q' is required")
		return
	}
	if len(query) > 200 {
		httputil.Error(w, http.StatusBadRequest, "query too long")
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}
	wsID, err := workspaceIDFromRequest(r)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid workspace id")
		return
	}

	results, err := h.svc.Search(r.Context(), wsID, query, limit, offset)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, results)
}

func (h *Handler) ListFavorites(w http.ResponseWriter, r *http.Request) {
	wsID, err := workspaceIDFromRequest(r)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid workspace id")
		return
	}
	pages, err := h.svc.ListFavorites(r.Context(), wsID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, pages)
}

func (h *Handler) ListTrash(w http.ResponseWriter, r *http.Request) {
	wsID, err := workspaceIDFromRequest(r)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid workspace id")
		return
	}
	pages, err := h.svc.ListTrash(r.Context(), wsID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, pages)
}

func (h *Handler) PermanentDelete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.svc.PermanentDelete(r.Context(), id); err != nil {
		slog.Error("permanent delete error", "error", err)
		httputil.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/pages", h.ListPages)
	r.Post("/pages", h.CreatePage)
	r.Get("/pages/{id}", h.GetPageTree)
	r.Post("/blocks", h.CreateBlock)
	r.Patch("/blocks/{id}", h.UpdateBlock)
	r.Delete("/blocks/{id}", h.DeleteBlock)
	r.Patch("/blocks/{id}/restore", h.RestoreBlock)
	r.Patch("/blocks/{id}/move", h.MoveBlock)
	r.Get("/search", h.Search)
	r.Get("/favorites", h.ListFavorites)
	r.Get("/trash", h.ListTrash)
	r.Delete("/blocks/{id}/permanent", h.PermanentDelete)
}
