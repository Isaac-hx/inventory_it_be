package categories

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
	pkg.PublicRoute(
		r.mux,
		"GET",
		"/categories",
		http.HandlerFunc(r.handler.GetAllCategories),
	)

	pkg.PublicRoute(
		r.mux,
		"GET",
		"/categories/{category_id}",
		http.HandlerFunc(r.handler.GetCategoryById),
	)
	pkg.ProtectedRoute(
		r.mux,
		"POST",
		"/categories",
		[]string{"superuser", "admin_it"},
		http.HandlerFunc(r.handler.CreateCategory),
		middleware.JWTMiddleware(r.jwtConfig),
		middleware.RBACMiddleware("superuser", "admin_it"),
	)

	pkg.ProtectedRoute(
		r.mux,
		"DELETE",
		"/categories/{category_id}",
		[]string{"superuser", "admin_it"},
		http.HandlerFunc(r.handler.DeleteCategory),
		middleware.JWTMiddleware(r.jwtConfig),
		middleware.RBACMiddleware("superuser", "admin_it"),
	)
	pkg.ProtectedRoute(
		r.mux,
		"PUT",
		"/categories/{category_id}",
		[]string{"superuser", "admin_it"},
		http.HandlerFunc(r.handler.UpdateCategory),
		middleware.JWTMiddleware(r.jwtConfig),
		middleware.RBACMiddleware("superuser", "admin_it"),
	)
}
