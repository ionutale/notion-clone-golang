package middleware

import (
	"context"
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
				http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
				return
			}
			token := strings.TrimPrefix(authHeader, "Bearer ")
			userID, err := validator.ValidateToken(token)
			if err != nil {
				http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), auth.CtxUserID, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
