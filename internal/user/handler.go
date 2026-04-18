// This file handles domain user input validation and sanitazation
package user

import (
	"encoding/json"
	"inventory-it/internal/pkg"
	"net/http"
)

// User request via json scheme
type userRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// Object Handler wiring to usecase
type Handler struct {
	usecase Usecase
}

// Constructor object handler
func NewHandler(u Usecase) *Handler {
	return &Handler{usecase: u}
}

// Method register is used for registration user
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var user userRequest

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusBadRequest, err.Error(), err.Error())
		return
	}
	if user.Username == "" || user.Password == "" || user.Email == "" || user.Role == "" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Missing required field!!", nil)
		return
	}
	if len(user.Username) < 5 {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Username must be at least 5 characters!!", nil)
		return
	}
	if len(user.Password) < 8 {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Password must be at least 8 characters!!", nil)

		return
	}
	err = h.usecase.Register(r.Context(), user.Username, user.Email, user.Password, user.Role)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusInternalServerError, err.Error(), err.Error())

		return
	}

	pkg.JSONResponse(w, 200, "User created", nil)

}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var user userRequest

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusBadRequest, err.Error(), err.Error())
		return
	}
	if user.Username == "" || user.Password == "" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Missing required field!!", nil)
		return
	}
	if len(user.Username) < 5 {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Username must be at least 5 characters!!", nil)
		return
	}
	if len(user.Password) < 8 {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Password must be at least 8 characters!!", nil)

		return
	}
	token, err := h.usecase.Login(r.Context(), user.Username, user.Password)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusInternalServerError, err.Error(), err.Error())

		return
	}

	pkg.JSONResponse(w, 200, "Login successful", map[string]string{"token": token})

}
