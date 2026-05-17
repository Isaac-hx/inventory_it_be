package auth

import (
	"encoding/json"
	"inventory-it/internal/pkg"
	"net/http"
)

type userRequest struct {
	Username      string `json:"username"`
	Email         string `json:"email"`
	Password      string `json:"password"`
	Role          string `json:"role"`
	Department_id string `json:"department_id"`
}

type Handler interface {
	Register(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	usecase Usecase
}

func NewHandler(u Usecase) Handler {
	return &handler{
		usecase: u,
	}
}

func (h *handler) Register(w http.ResponseWriter, r *http.Request) {
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

	err = h.usecase.Register(
		r.Context(),
		user.Username,
		user.Email,
		user.Password,
		user.Role,
		user.Department_id,
	)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusInternalServerError, err.Error(), err.Error())
		return
	}

	pkg.JSONResponse(w, http.StatusCreated, "User created", nil)
}

func (h *handler) Login(w http.ResponseWriter, r *http.Request) {
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
		pkg.ErrorResponse(w, http.StatusUnauthorized, "Invalid username or password", nil)
		return
	}

	pkg.JSONResponse(w, http.StatusOK, "Login successful", map[string]string{
		"token": token,
	})
}
