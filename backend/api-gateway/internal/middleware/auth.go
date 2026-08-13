package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/IlyushaChic/financial-platform/backend/api-gateway/internal/clients"
)

type contextKey string

const UserIDKey contextKey = "userID"

// AuthMiddleware проверяет JWT и добавляет userID в контекст
func AuthMiddleware(authClient *clients.AuthClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Пропускаем playground и healthcheck
			if r.URL.Path == "/" || r.URL.Path == "/health" {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing authorization header", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
				return
			}
			token := parts[1]

			// Валидируем токен через Auth Service
			resp, err := authClient.ValidateToken(r.Context(), token)
			if err != nil || !resp.Valid {
				http.Error(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}

			// Кладём userID в контекст
			ctx := context.WithValue(r.Context(), UserIDKey, resp.UserId)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
