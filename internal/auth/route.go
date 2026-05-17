//This file contain registered user routes

package auth

import (
	"inventory-it/internal/middleware"
	"inventory-it/internal/pkg"
	"net/http"
)

type Routes struct {
	mux       *http.ServeMux
	handler   Handler
	jwtConfig *pkg.JwtConfig
}

// Constructor routes object
func NewRoutes(h Handler, mux *http.ServeMux, jwtConfig *pkg.JwtConfig) *Routes {
	return &Routes{handler: h, mux: mux, jwtConfig: jwtConfig}
}

// Method register routes
func (r *Routes) RegisterRoutes() {
	pkg.PublicRoute(r.mux,
		"POST",
		"/login",
		http.HandlerFunc(r.handler.Login))
	pkg.ProtectedRoute(
		r.mux, "POST",
		"/register",
		[]string{"superuser"},
		http.HandlerFunc(r.handler.Register),
		middleware.JWTMiddleware(r.jwtConfig),
		middleware.RBACMiddleware("superuser"))

}
