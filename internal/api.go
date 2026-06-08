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
	wsH := ws.NewHandler(wsSvc, authSvc)
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

			// Auth routes (logout, me, settings)
			authH.RegisterProtectedRoutes(r)

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

			ext := filepath.Ext(header.Filename)
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
