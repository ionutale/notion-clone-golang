package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/ionutale/notion-clone-golang/internal"
	"github.com/ionutale/notion-clone-golang/internal/block"
	"github.com/ionutale/notion-clone-golang/internal/config"
	"github.com/ionutale/notion-clone-golang/internal/db"
	"github.com/ionutale/notion-clone-golang/internal/middleware"
	"github.com/ionutale/notion-clone-golang/internal/storage"
)

//go:embed web/build
var spaFS embed.FS

//go:embed migrations/*.sql
var migrationFS embed.FS

var (
	DefaultWorkspaceID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	DefaultUserID      = uuid.MustParse("00000000-0000-0000-0000-000000000002")
)

func spaHandler(fsys fs.FS) http.HandlerFunc {
	mimeTypes := map[string]string{
		".js":   "application/javascript",
		".css":  "text/css",
		".html": "text/html",
		".json": "application/json",
		".svg":  "image/svg+xml",
		".png":  "image/png",
		".ico":  "image/x-icon",
		".webp": "image/webp",
	}

	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			data, err = fs.ReadFile(fsys, "index.html")
			if err != nil {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
		}

		for ext, mime := range mimeTypes {
			if strings.HasSuffix(path, ext) {
				w.Header().Set("Content-Type", mime)
				break
			}
		}

		w.Write(data)
	}
}

func runMigrations(pool *pgxpool.Pool) error {
	entries, err := fs.Glob(migrationFS, "migrations/*.up.sql")
	if err != nil {
		return fmt.Errorf("list migrations: %w", err)
	}
	sort.Strings(entries)

	for _, entry := range entries {
		slog.Info("running migration", "file", filepath.Base(entry))
		sql, err := migrationFS.ReadFile(entry)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", entry, err)
		}
		if _, err := pool.Exec(context.Background(), string(sql)); err != nil {
			return fmt.Errorf("run migration %s: %w", entry, err)
		}
	}
	return nil
}

func main() {
	godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()

	var pool *pgxpool.Pool
	pool, err = db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Warn("database not available, starting without db", "error", err)
	} else {
		defer pool.Close()
		slog.Info("connected to database")

		if err := runMigrations(pool); err != nil {
			slog.Error("migration failed", "error", err)
			os.Exit(1)
		}
		slog.Info("migrations complete")
	}

	var blockSvc *block.Service
	if pool != nil {
		blockSvc = block.NewService(pool, DefaultWorkspaceID, DefaultUserID)
	}

	var fileStore storage.FileStore
	if cfg.StorageEmulatorHost != "" {
		fileStore = storage.NewLocalFileStore("./data/uploads", "/uploads")
		slog.Info("using local file store", "dir", "./data/uploads")
	} else {
		fileStore = storage.NewLocalFileStore("./data/uploads", "/uploads")
	}

	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recovery)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	if blockSvc != nil {
		internal.MountAPI(r, blockSvc, fileStore)
	} else {
		r.Route("/api/v1", func(r chi.Router) {
			r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"status":"ok","database":"disconnected"}`))
			})
		})
	}

	if cfg.DevMode {
		target, _ := url.Parse("http://localhost:5173")
		proxy := httputil.NewSingleHostReverseProxy(target)
		r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			proxy.ServeHTTP(w, r)
		})
	} else {
		spaSub, err := fs.Sub(spaFS, "web/build")
		if err != nil {
			slog.Error("failed to get spa sub filesystem", "error", err)
			os.Exit(1)
		}
		r.Get("/*", spaHandler(spaSub))
	}

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("server starting", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-quit
	slog.Info("shutting down...")
	server.Shutdown(ctx)
}
