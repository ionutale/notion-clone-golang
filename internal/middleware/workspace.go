package middleware

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/ionutale/notion-clone-golang/internal/auth"
)

type WorkspaceValidator interface {
	IsMember(ctx context.Context, workspaceID, userID string) (bool, error)
}

func WorkspaceMiddleware(validator WorkspaceValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			workspaceID := chi.URLParam(r, "workspaceId")
			userID, ok := r.Context().Value(auth.CtxUserID).(string)
			if !ok {
				http.Error(w, `{"error":"not authenticated"}`, http.StatusUnauthorized)
				return
			}
			ok, err := validator.IsMember(r.Context(), workspaceID, userID)
			if err != nil || !ok {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"error":"workspace not found"}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
