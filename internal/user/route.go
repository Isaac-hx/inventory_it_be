//This file contain registered user routes

package user

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
func (r *Routes) RegisterRoutes() {
	pkg.ProtectedRoute(
		r.mux,
		"GET",
		"/users",
		[]string{"superuser"},
		http.HandlerFunc(r.handler.GetAllUsers),
		middleware.JWTMiddleware(r.jwtConfig),
		middleware.RBACMiddleware("superuser"),
	)
	pkg.ProtectedRoute(
		r.mux,
		"GET",
		"/users/{user_id}",
		[]string{"superuser"},
		http.HandlerFunc(r.handler.GetUserById),
		middleware.JWTMiddleware(r.jwtConfig),
		middleware.RBACMiddleware("superuser"),
	)

	pkg.ProtectedRoute(
		r.mux,
		"DELETE",
		"/users/{user_id}",
		[]string{"superuser"},
		http.HandlerFunc(r.handler.DeleteUserById),
		middleware.JWTMiddleware(r.jwtConfig),
		middleware.RBACMiddleware("superuser"),
	)

}
