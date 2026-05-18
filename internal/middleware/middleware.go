package middleware

import (
	"context"
	"inventory-it/internal/pkg"
	"net/http"
	"strings"
)

type contextKey string

const Claimskey contextKey = "claims"

func JWTMiddleware(jwtConfig *pkg.JwtConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				pkg.ErrorResponse(w, http.StatusUnauthorized, "missing token", nil)
				return
			}

			claims, err := jwtConfig.ParseToken(strings.TrimPrefix(header, "Bearer "))
			if err != nil {
				pkg.ErrorResponse(w, http.StatusUnauthorized, "invalid token", nil)
				return
			}

			ctx := context.WithValue(r.Context(), Claimskey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RBACMiddleware(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userContext := r.Context().Value(Claimskey)
			claims, ok := userContext.(*pkg.Claims)
			if !ok {
				pkg.ErrorResponse(w, http.StatusUnauthorized, "Invalid user data!", nil)
				return

			}
			for _, role := range allowedRoles {
				if claims.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}
			pkg.ErrorResponse(w, http.StatusUnauthorized, "Forbidden: access denied", nil)

		})
	}
}
