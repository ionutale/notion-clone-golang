package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/ionutale/notion-clone-golang/internal/auth"
)

type TokenValidator interface {
	ValidateToken(tokenString string) (string, error)
}

func AuthMiddleware(validator TokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"missing authorization header"}`))
				return
			}
			token := strings.TrimPrefix(authHeader, "Bearer ")
			userID, err := validator.ValidateToken(token)
			if err != nil {
				slog.Warn("auth failure", "error", err, "path", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"invalid token"}`))
				return
			}
			ctx := context.WithValue(r.Context(), auth.CtxUserID, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
