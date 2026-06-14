package internal

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ionutale/notion-clone-golang/internal/auth"
	"github.com/ionutale/notion-clone-golang/internal/block"
	"github.com/ionutale/notion-clone-golang/internal/httputil"
	"github.com/ionutale/notion-clone-golang/internal/middleware"
	"github.com/ionutale/notion-clone-golang/internal/storage"
	ws "github.com/ionutale/notion-clone-golang/internal/workspace"
)

var allowedUploadMIMETypes = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"image/gif":       true,
	"image/webp":      true,
	"image/svg+xml":   true,
	"application/pdf": true,
}

func MountAPI(
	r chi.Router,
	blockSvc *block.Service,
	fileStore storage.FileStore,
	authSvc *auth.Service,
	wsSvc *ws.Service,
	devMode bool,
) {
	var authH *auth.Handler
	if devMode {
		authH = auth.NewHandlerDev(authSvc)
	} else {
		authH = auth.NewHandler(authSvc)
	}
	wsH := ws.NewHandler(wsSvc, authSvc)
	blockH := block.NewHandler(blockSvc)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			httputil.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
		})

		fmt.Printf("Mounting API routes (dev mode: %v)\n", devMode)

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
			if err := r.ParseMultipartForm(10 << 20); err != nil {
				httputil.Error(w, http.StatusBadRequest, "invalid multipart form")
				return
			}
			file, header, err := r.FormFile("file")
			if err != nil {
				httputil.Error(w, http.StatusBadRequest, "missing file")
				return
			}
			defer file.Close()

			// Validate MIME type
			buf := make([]byte, 512)
			n, _ := file.Read(buf)
			file.Seek(0, 0)
			contentType := http.DetectContentType(buf[:n])
			if !allowedUploadMIMETypes[contentType] && !strings.HasPrefix(contentType, "image/") {
				httputil.Error(w, http.StatusBadRequest, "unsupported file type")
				return
			}

			ext := filepath.Ext(header.Filename)
			if ext != "" && strings.ContainsAny(ext, "/\\") {
				httputil.Error(w, http.StatusBadRequest, "invalid file extension")
				return
			}
			key := uuid.New().String() + ext

			if err := fileStore.Put(r.Context(), key, file); err != nil {
				httputil.Error(w, http.StatusInternalServerError, "upload failed")
				return
			}

			url := fileStore.PublicURL(key)
			httputil.JSON(w, http.StatusCreated, map[string]string{"url": url})
		})
	})
}
