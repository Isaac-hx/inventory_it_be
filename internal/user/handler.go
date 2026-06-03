// This file handles domain user input validation and sanitazation
package user

import (
	"database/sql"
	"encoding/json"
	"errors"
	"inventory-it/internal/pkg"
	"net/http"
	"strconv"
)

type userRequest struct {
	Username      string `json:"username"`
	Email         string `json:"email"`
	Role          string `json:"role"`
	Department_id string `json:"department_id"`
}
type UserFilter struct {
	Role    string
	Search  string
	Limit   int
	Page    int
	OrderBy string
}
type Handler interface {
	GetAllUsers(w http.ResponseWriter, r *http.Request)
	GetUserById(w http.ResponseWriter, r *http.Request)
	DeleteUserById(w http.ResponseWriter, r *http.Request)
	UpdateUserById(w http.ResponseWriter, r *http.Request)
}

// Object Handler wiring to usecase
type handler struct {
	usecase Usecase
}

// Constructor object handler
func NewHandler(u Usecase) Handler {
	return &handler{
		usecase: u,
	}
}

// Method get all users
func (h *handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	var userFilter UserFilter
	//define context

	//get query params
	query := r.URL.Query()
	role := query.Get("role")
	search := query.Get("search")
	limitStr := query.Get("limit")
	pageStr := query.Get("page")
	orderBy := query.Get("order_by")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	if orderBy != "asc" && orderBy != "desc" {
		orderBy = "asc"
	}
	if role != "superuser" && role != "admin_it" && role != "user" && role != "" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid role", "Role must be admin_it or user")
		return
	}

	//structuring object userFilter
	userFilter.Role = role
	userFilter.Search = search
	userFilter.Limit = limit
	userFilter.Page = page
	userFilter.OrderBy = orderBy

	//call usecase get all users
	users, meta, err := h.usecase.GetAllUsers(r.Context(), userFilter)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusInternalServerError, err.Error(), err.Error())
		return
	}

	pkg.JSONResponse(w, http.StatusOK, "Success get all users", users, meta)
}

func (h *handler) GetUserById(w http.ResponseWriter, r *http.Request) {
	//get user id from query params
	userId := r.PathValue("user_id")
	if userId == "" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "User ID is required", nil)
		return
	}

	//call usecase get user by id
	user, err := h.usecase.GetUserById(r.Context(), userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			pkg.ErrorResponse(w, http.StatusNotFound, "User not found", nil)
			return
		}
		pkg.ErrorResponse(w, http.StatusInternalServerError, err.Error(), err.Error())
		return
	}

	pkg.JSONResponse(w, http.StatusOK, "Success get user by id", user, nil)
}

func (h *handler) DeleteUserById(w http.ResponseWriter, r *http.Request) {
	//get user id from query params
	userId := r.PathValue("user_id")
	if userId == "" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "User ID is required", nil)
		return
	}

	//call usecase delete user by id
	err := h.usecase.DeleteUserById(r.Context(), userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			pkg.ErrorResponse(w, http.StatusNotFound, "User not found", nil)
			return
		}

		pkg.ErrorResponse(w, http.StatusInternalServerError, err.Error(), err.Error())
		return
	}

	pkg.JSONResponse(w, http.StatusOK, "Success delete user by id", nil, nil)
}

func (h *handler) UpdateUserById(w http.ResponseWriter, r *http.Request) {
	//get user id from query params
	var userUpdateRequest userRequest

	userId := r.PathValue("user_id")
	if userId == "" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "User ID is required", nil)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&userUpdateRequest)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusBadRequest, err.Error(), err.Error())
		return
	}

	if userUpdateRequest.Username == "" || userUpdateRequest.Email == "" || userUpdateRequest.Role == "" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Missing required field!!", nil)
		return
	}

	if len(userUpdateRequest.Username) < 8 {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Username must be at least 8 characters!!", nil)
		return
	}

	if !pkg.IsValidEmail(userUpdateRequest.Email) {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid email format!!", nil)
		return
	}
	if userUpdateRequest.Role != "superuser" && userUpdateRequest.Role != "admin_it" && userUpdateRequest.Role != "user" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid role", "Role must be admin_it or user")
		return
	}
	var updatedUser User
	updatedUser.Username = userUpdateRequest.Username
	updatedUser.Email = userUpdateRequest.Email
	updatedUser.Role = userUpdateRequest.Role
	updatedUser.DepartmentId = userUpdateRequest.Department_id
	err = h.usecase.UpdateUserById(r.Context(), userId, updatedUser)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			pkg.ErrorResponse(w, http.StatusNotFound, "User not found", nil)
			return
		}
		pkg.ErrorResponse(w, http.StatusInternalServerError, err.Error(), err.Error())
		return
	}

	pkg.JSONResponse(w, http.StatusOK, "Success update user by id", nil, nil)
}
