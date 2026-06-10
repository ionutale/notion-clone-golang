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
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/ionutale/notion-clone-golang/internal"
	"github.com/ionutale/notion-clone-golang/internal/auth"
	"github.com/ionutale/notion-clone-golang/internal/block"
	"github.com/ionutale/notion-clone-golang/internal/config"
	"github.com/ionutale/notion-clone-golang/internal/db"
	"github.com/ionutale/notion-clone-golang/internal/middleware"
	"github.com/ionutale/notion-clone-golang/internal/storage"
	ws "github.com/ionutale/notion-clone-golang/internal/workspace"
)

//go:embed all:web/build
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

		if strings.HasPrefix(path, "_app/immutable/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
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

type wsCreatorAdapter struct {
	svc *ws.Service
}

func (a *wsCreatorAdapter) Create(ctx context.Context, name, ownerID string) (interface{}, error) {
	return a.svc.Create(ctx, name, ownerID)
}

func main() {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		slog.Warn("error loading .env file", "error", err)
	}

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	migrateCtx, migrateCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer migrateCancel()

	var pool *pgxpool.Pool
	pool, err = db.Connect(migrateCtx, cfg.DatabaseURL)
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
	var authSvc *auth.Service
	var wsSvc *ws.Service
	if pool != nil {
		blockSvc = block.NewService(pool)
		blockSvc.StartCleanupLoop(context.Background(), 1*time.Hour, 30)
		wsRepo := ws.NewRepository(pool)
		wsSvc = ws.NewService(wsRepo)
		authRepo := auth.NewRepository(pool)
		authSvc = auth.NewService(authRepo, &wsCreatorAdapter{svc: wsSvc})
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
	r.Use(chimw.Compress(5))
	r.Use(middleware.Logger)
	r.Use(middleware.Recovery)

	var corsOrigins []string
	if cfg.DevMode {
		corsOrigins = []string{"http://localhost:5173", "http://localhost:8080"}
		if envOrigins := os.Getenv("CORS_ORIGINS"); envOrigins != "" {
			corsOrigins = strings.Fields(envOrigins)
		}
	} else if origin := os.Getenv("APP_URL"); origin != "" {
		corsOrigins = []string{origin}
	}
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   corsOrigins,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	r.Use(middleware.SecurityHeaders(cfg.DevMode))

	// Serve uploaded files
	uploadsDir := "./data/uploads"
	r.Route("/uploads", func(r chi.Router) {
		r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			requestedPath := chi.URLParam(r, "*")
			if strings.Contains(requestedPath, "..") {
				http.Error(w, `{"error":"invalid path"}`, http.StatusBadRequest)
				return
			}
			cleanPath := filepath.Clean("/" + requestedPath)
			filePath := filepath.Join(uploadsDir, cleanPath)
			if !strings.HasPrefix(filePath, filepath.Clean(uploadsDir)+string(filepath.Separator)) &&
				filePath != filepath.Clean(uploadsDir) {
				http.Error(w, `{"error":"access denied"}`, http.StatusForbidden)
				return
			}
			http.ServeFile(w, r, filePath)
		})
	})

	if blockSvc != nil {
		internal.MountAPI(r, blockSvc, fileStore, authSvc, wsSvc, cfg.DevMode)
	} else {
		r.Route("/api/v1", func(r chi.Router) {
			r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"status":"ok","database":"disconnected"}`))
			})
		})
	}

	if cfg.DevMode {
		target, err := url.Parse("http://localhost:5173")
		if err != nil {
			slog.Error("failed to parse dev proxy target", "error", err)
			os.Exit(1)
		}
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
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
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
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", "error", err)
	}
}
