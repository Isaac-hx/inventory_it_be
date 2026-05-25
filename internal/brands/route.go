package brands

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
		"/brands",
		http.HandlerFunc(r.handler.GetAllBrands),
	)

	pkg.PublicRoute(
		r.mux,
		"GET",
		"/brands/{brand_id}",
		http.HandlerFunc(r.handler.GetBrandById),
	)
	pkg.ProtectedRoute(
		r.mux,
		"POST",
		"/brands",
		[]string{"superuser", "admin_it"},
		http.HandlerFunc(r.handler.CreateBrand),
		middleware.JWTMiddleware(r.jwtConfig),
		middleware.RBACMiddleware("superuser", "admin_it"),
	)

	pkg.ProtectedRoute(
		r.mux,
		"DELETE",
		"/brands/{brand_id}",
		[]string{"superuser", "admin_it"},
		http.HandlerFunc(r.handler.DeleteBrand),
		middleware.JWTMiddleware(r.jwtConfig),
		middleware.RBACMiddleware("superuser", "admin_it"),
	)
	pkg.ProtectedRoute(
		r.mux,
		"PUT",
		"/brands/{brand_id}",
		[]string{"superuser", "admin_it"},
		http.HandlerFunc(r.handler.UpdateBrand),
		middleware.JWTMiddleware(r.jwtConfig),
		middleware.RBACMiddleware("superuser", "admin_it"),
	)
}
