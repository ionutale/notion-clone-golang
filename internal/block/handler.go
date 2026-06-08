package block

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ionutale/notion-clone-golang/internal/auth"
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

func (h *Handler) CreatePage(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Title == "" {
		req.Title = "Untitled"
	}
	page, err := h.svc.CreatePage(r.Context(), workspaceIDFromRequest(r), userIDFromRequest(r), req.Title)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respond(w, http.StatusCreated, page)
}

func (h *Handler) GetPageTree(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}
	tree, err := h.svc.GetPageTree(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	respond(w, http.StatusOK, tree)
}

func (h *Handler) ListPages(w http.ResponseWriter, r *http.Request) {
	pages, err := h.svc.ListPages(r.Context(), workspaceIDFromRequest(r))
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respond(w, http.StatusOK, pages)
}

func (h *Handler) CreateBlock(w http.ResponseWriter, r *http.Request) {
	var req CreateBlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	block, err := h.svc.CreateBlock(r.Context(), workspaceIDFromRequest(r), userIDFromRequest(r), req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respond(w, http.StatusCreated, block)
}

func (h *Handler) UpdateBlock(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req UpdateBlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	block, err := h.svc.UpdateBlock(r.Context(), id, req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respond(w, http.StatusOK, block)
}

func (h *Handler) DeleteBlock(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.svc.DeleteBlock(r.Context(), id); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RestoreBlock(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}
	block, err := h.svc.RestoreBlock(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respond(w, http.StatusOK, block)
}

func (h *Handler) MoveBlock(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req MoveBlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	block, err := h.svc.MoveBlock(r.Context(), workspaceIDFromRequest(r), id, req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respond(w, http.StatusOK, block)
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		respondError(w, http.StatusBadRequest, "query parameter 'q' is required")
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

	results, err := h.svc.Search(r.Context(), workspaceIDFromRequest(r), query, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respond(w, http.StatusOK, results)
}

func (h *Handler) ListFavorites(w http.ResponseWriter, r *http.Request) {
	pages, err := h.svc.ListFavorites(r.Context(), workspaceIDFromRequest(r))
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respond(w, http.StatusOK, pages)
}

func (h *Handler) ListTrash(w http.ResponseWriter, r *http.Request) {
	pages, err := h.svc.ListTrash(r.Context(), workspaceIDFromRequest(r))
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respond(w, http.StatusOK, pages)
}

func (h *Handler) PermanentDelete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.svc.PermanentDelete(r.Context(), id); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
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
