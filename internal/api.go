package internal

import (
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ionutale/notion-clone-golang/internal/block"
	"github.com/ionutale/notion-clone-golang/internal/storage"
)

func MountAPI(r chi.Router, blockSvc *block.Service, fileStore storage.FileStore) {
	blockHandler := block.NewHandler(blockSvc)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		})

		r.Get("/pages", blockHandler.ListPages)
		r.Post("/pages", blockHandler.CreatePage)
		r.Get("/pages/{id}", blockHandler.GetPageTree)

		r.Post("/blocks", blockHandler.CreateBlock)
		r.Patch("/blocks/{id}", blockHandler.UpdateBlock)
		r.Delete("/blocks/{id}", blockHandler.DeleteBlock)
		r.Patch("/blocks/{id}/restore", blockHandler.RestoreBlock)
		r.Patch("/blocks/{id}/move", blockHandler.MoveBlock)

		r.Post("/uploads", func(w http.ResponseWriter, r *http.Request) {
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
