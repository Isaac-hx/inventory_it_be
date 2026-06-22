// This file contain registered asset routes

package assets

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
		"/assets",
		[]string{"superuser", "admin_it"},
		http.HandlerFunc(r.handler.CreateAsset),
		middleware.JWTMiddleware(r.jwtConfig),
		middleware.RBACMiddleware("superuser", "admin_it"),
	)

	pkg.ProtectedRoute(
		r.mux,
		"GET",
		"/assets",
		[]string{"superuser", "admin_it"},
		http.HandlerFunc(r.handler.GetAssets),
		middleware.JWTMiddleware(r.jwtConfig),
		middleware.RBACMiddleware("superuser", "admin_it"),
	)

	pkg.ProtectedRoute(
		r.mux,
		"GET",
		"/all-assets",
		[]string{"superuser", "admin_it"},
		http.HandlerFunc(r.handler.GetAllAssetsData),
		middleware.JWTMiddleware(r.jwtConfig),
		middleware.RBACMiddleware("superuser", "admin_it"),
	)

	pkg.ProtectedRoute(
		r.mux,
		"GET",
		"/assets/{asset_id}",
		[]string{"superuser", "admin_it"},
		http.HandlerFunc(r.handler.GetAssetByID),
		middleware.JWTMiddleware(r.jwtConfig),
		middleware.RBACMiddleware("superuser", "admin_it"),
	)

	pkg.ProtectedRoute(
		r.mux,
		"PUT",
		"/assets/{asset_id}",
		[]string{"superuser", "admin_it"},
		http.HandlerFunc(r.handler.UpdateAsset),
		middleware.JWTMiddleware(r.jwtConfig),
		middleware.RBACMiddleware("superuser", "admin_it"),
	)

	pkg.ProtectedRoute(
		r.mux,
		"DELETE",
		"/assets/{asset_id}",
		[]string{"superuser", "admin_it"},
		http.HandlerFunc(r.handler.DeleteAsset),
		middleware.JWTMiddleware(r.jwtConfig),
		middleware.RBACMiddleware("superuser", "admin_it"),
	)

	pkg.PublicRoute(
		r.mux,
		"GET",
		"/assets/analytics",
		http.HandlerFunc(r.handler.GetOverview),
	)

	pkg.PublicRoute(
		r.mux,
		"GET",
		"/assets/analytics/category-distribution",
		http.HandlerFunc(r.handler.GetGraphicDistributionByCategory),
	)
}
