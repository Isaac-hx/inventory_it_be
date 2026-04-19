package middleware

import (
	"context"
	"inventory-it/internal/auth"
	"net/http"
	"strings"
)

type contextKey string

const Claimskey contextKey = "claims"

func JWTAuth(jwtConfig *auth.JwtConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				http.Error(w, "missing token", http.StatusUnauthorized)
				return
			}

			claims, err := jwtConfig.ParseToken(strings.TrimPrefix(header, "Bearer "))
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), Claimskey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
