//This file contain registered routes for the application

package pkg

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

func PublicRoute(
	mux *http.ServeMux,
	method string,
	path string,
	handler http.Handler,
) {
	route := fmt.Sprintf("%s %s", method, path)

	log.Printf("[PUBLIC]    %s", route)

	mux.Handle(route, handler)
}

func ProtectedRoute(
	mux *http.ServeMux,
	method string,
	path string,
	roles []string,
	handler http.Handler,
	jwtMiddleware func(http.Handler) http.Handler,
	rbacMiddleware func(http.Handler) http.Handler,
) {

	route := fmt.Sprintf("%s %s", method, path)

	log.Printf(
		"[PROTECTED] %s | roles: [%s]",
		route,
		strings.Join(roles, ", "),
	)

	mux.Handle(
		route,
		jwtMiddleware(
			rbacMiddleware(handler),
		),
	)
}
