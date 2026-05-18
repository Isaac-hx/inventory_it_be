package departments

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
		"/departments",
		[]string{"superuser"},
		http.HandlerFunc(r.handler.GetAllDepartments),
		middleware.JWTMiddleware(r.jwtConfig),
		middleware.RBACMiddleware("superuser"),
	)
	pkg.ProtectedRoute(
		r.mux,
		"GET",
		"/departments/{department_id}",
		[]string{"superuser"},
		http.HandlerFunc(r.handler.GetDepartmentById),
		middleware.JWTMiddleware(r.jwtConfig),
		middleware.RBACMiddleware("superuser"),
	)
	pkg.ProtectedRoute(
		r.mux,
		"DELETE",
		"/departments/{department_id}",
		[]string{"superuser"},
		http.HandlerFunc(r.handler.DeleteDepartmentById),
		middleware.JWTMiddleware(r.jwtConfig),
		middleware.RBACMiddleware("superuser"),
	)

	pkg.ProtectedRoute(
		r.mux,
		"PUT",
		"/departments/{department_id}",
		[]string{"superuser"},
		http.HandlerFunc(r.handler.UpdateDepartmentNameById),
		middleware.JWTMiddleware(r.jwtConfig),
		middleware.RBACMiddleware("superuser"),
	)
	pkg.ProtectedRoute(
		r.mux,
		"POST",
		"/departments",
		[]string{"superuser"},
		http.HandlerFunc(r.handler.CreateDepartment),
		middleware.JWTMiddleware(r.jwtConfig),
		middleware.RBACMiddleware("superuser"),
	)
}
