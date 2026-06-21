package auth

import (
	"encoding/json"
	"inventory-it/internal/pkg"
	"net/http"
)

type userResponse struct {
	User  User   `json:"user"`
	Token string `json:"token"`
}
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
	if !pkg.IsValidEmail(user.Email) {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid email format!!", nil)
		return
	}

	if user.Role != "admin_it" && user.Role != "user" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid role", "Role must be admin_it or user")
		return
	}

	//assign to domain
	userDomain := User{
		Username:      user.Username,
		Email:         user.Email,
		Password:      user.Password,
		Role:          user.Role,
		Department_id: user.Department_id,
	}

	userData, err := h.usecase.Register(
		r.Context(),
		userDomain,
	)
	if err != nil {
		if err.Error() == "Email already registered" || err.Error() == "Username already registered!" {
			pkg.ErrorResponse(w, http.StatusConflict, err.Error(), nil)
			return
		}
		pkg.ErrorResponse(w, http.StatusInternalServerError, err.Error(), err.Error())
		return
	}
	var userResponse User
	userResponse.UserId = userData.UserId
	userResponse.Username = userData.Username
	userResponse.Email = userData.Email
	userResponse.Department_id = userData.Department_id
	userResponse.Role = userData.Role

	pkg.JSONResponse(w, http.StatusCreated, "User created", userResponse, nil)
}

func (h *handler) Login(w http.ResponseWriter, r *http.Request) {
	var user struct {
		UsernameOrEmail string `json:"username_or_email"`
		Password        string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusBadRequest, err.Error(), err.Error())
		return
	}

	if user.UsernameOrEmail == "" || user.Password == "" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Missing required field!!", nil)
		return
	}

	if len(user.UsernameOrEmail) < 5 {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Username must be at least 5 characters!!", nil)
		return
	}

	if len(user.Password) < 8 {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Password must be at least 8 characters!!", nil)
		return
	}

	userRegistered, token, err := h.usecase.Login(r.Context(), user.UsernameOrEmail, user.Password)
	if err != nil {
		if err.Error() == "invalid username or password" {
			pkg.ErrorResponse(w, http.StatusUnauthorized, err.Error(), nil)
			return
		}
		pkg.ErrorResponse(w, http.StatusInternalServerError, "Internal server error", err.Error())
		return
	}

	var userResp userResponse
	userResp.User.UserId = userRegistered.UserId
	userResp.User.Email = userRegistered.Email
	userResp.User.Department_id = userRegistered.Department_id
	userResp.User.Role = userRegistered.Role
	userResp.User.Username = userRegistered.Username
	userResp.Token = token
	pkg.JSONResponse(w, http.StatusOK, "Login successful", userResp, nil)
}
