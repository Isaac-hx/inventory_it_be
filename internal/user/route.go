//This file contain registered user routes

package user

import (
	"log"
	"net/http"
)

type Routes struct {
	handler Handler
}

// Constructor routes object
func NewRoutes(h Handler) *Routes {
	return &Routes{handler: h}
}

// Method register routes
func (r *Routes) RegisterRoutes(mux *http.ServeMux) {
	r.register("POST /users/register", mux, r.handler.Register)
	r.register("POST  /users/login", mux, r.handler.Login)
}
func (r *Routes) register(pattern string, mux *http.ServeMux, handler http.HandlerFunc) {
	log.Println("ENDPOINT users :", pattern)
	mux.HandleFunc(pattern, handler)
}
