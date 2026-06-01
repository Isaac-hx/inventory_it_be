package maintenances

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
	pkg.ProtectedRoute(
		r.mux,
		"GET",
		"/maintenances",
		[]string{"superuser", "admin_it"},
		http.HandlerFunc(r.handler.GetAllMaintenances),
		middleware.JWTMiddleware(r.jwtConfig),
		middleware.RBACMiddleware("superuser", "admin_it"),
	)
	pkg.ProtectedRoute(
		r.mux,
		"GET",
		"/maintenances/{maintenance_id}",
		[]string{"superuser", "admin_it"},
		http.HandlerFunc(r.handler.GetMaintenanceById),
		middleware.JWTMiddleware(r.jwtConfig),
		middleware.RBACMiddleware("superuser", "admin_it"),
	)

	pkg.ProtectedRoute(
		r.mux,
		"PUT",
		"/maintenances/{maintenance_id}",
		[]string{"superuser", "admin_it"},
		http.HandlerFunc(r.handler.UpdateMaintenance),
		middleware.JWTMiddleware(r.jwtConfig),
		middleware.RBACMiddleware("superuser", "admin_it"),
	)

	pkg.ProtectedRoute(
		r.mux,
		"PUT",
		"/maintenances-status/{maintenance_id}",
		[]string{"superuser", "admin_it"},
		http.HandlerFunc(r.handler.UpdateStatusMaintenance),
		middleware.JWTMiddleware(r.jwtConfig),
		middleware.RBACMiddleware("superuser", "admin_it"),
	)
	pkg.ProtectedRoute(
		r.mux,
		"POST",
		"/maintenances",
		[]string{"superuser", "admin_it"},
		http.HandlerFunc(r.handler.CreateMaintenance),
		middleware.JWTMiddleware(r.jwtConfig),
		middleware.RBACMiddleware("superuser", "admin_it"),
	)

}
