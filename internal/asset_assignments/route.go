// This file contain registered asset routes

package assetassignments

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
	return &Routes{
		handler:   h,
		mux:       mux,
		jwtConfig: jwtConfig,
	}
}

// Method register routes
func (r *Routes) RegisterRoutes() {
	pkg.ProtectedRoute(
		r.mux,
		"POST",
		"/asset-assignments",
		[]string{"superuser", "admin_it"},
		http.HandlerFunc(r.handler.CreateAssetAssignment),
		middleware.JWTMiddleware(r.jwtConfig),
		middleware.RBACMiddleware("superuser", "admin_it"),
	)

	pkg.ProtectedRoute(
		r.mux,
		"GET",
		"/asset-assignments",
		[]string{"superuser", "admin_it"},
		http.HandlerFunc(r.handler.GetAllAssetAssignments),
		middleware.JWTMiddleware(r.jwtConfig),
		middleware.RBACMiddleware("superuser", "admin_it"),
	)
	pkg.ProtectedRoute(
		r.mux,
		"GET",
		"/all-asset-assignments",
		[]string{"superuser", "admin_it"},
		http.HandlerFunc(r.handler.GetAllAssignmentsData),
		middleware.JWTMiddleware(r.jwtConfig),
		middleware.RBACMiddleware("superuser", "admin_it"),
	)

	pkg.ProtectedRoute(
		r.mux,
		"GET",
		"/asset-assignments/{assignment_id}",
		[]string{"superuser", "admin_it"},
		http.HandlerFunc(r.handler.GetAssetAssignmentById),
		middleware.JWTMiddleware(r.jwtConfig),
		middleware.RBACMiddleware("superuser", "admin_it"),
	)
	pkg.ProtectedRoute(
		r.mux,
		"PUT",
		"/asset-assignments/{assignment_id}",
		[]string{"superuser", "admin_it"},
		http.HandlerFunc(r.handler.UpdateAssetAssignment),
		middleware.JWTMiddleware(r.jwtConfig),
		middleware.RBACMiddleware("superuser", "admin_it"),
	)
	pkg.ProtectedRoute(
		r.mux,
		"PUT",
		"/asset-assignments/status/{assignment_id}",
		[]string{"superuser", "admin_it"},
		http.HandlerFunc(r.handler.UpdateAssetAssignmentStatus),
		middleware.JWTMiddleware(r.jwtConfig),
		middleware.RBACMiddleware("superuser", "admin_it"),
	)
}
