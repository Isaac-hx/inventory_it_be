// This file handles domain user input validation and sanitazation
package user

import (
	"inventory-it/internal/pkg"
	"net/http"
	"strconv"
)

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
	if err != nil {
		limit = 10
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 1
	}

	if orderBy != "asc" && orderBy != "desc" {
		orderBy = "asc"
	}
	if role != "superuser" && role != "admin_it" && role != "user" && role != "" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid role", "Role must be admin_it or user")
		return
	}
	if limit < 1 {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid limit", "Limit must be greater than 0")
		return
	}

	//structuring object userFilter
	userFilter.Role = role
	userFilter.Search = search
	userFilter.Limit = limit
	userFilter.Page = page
	userFilter.OrderBy = orderBy

	//call usecase get all users
	users, err := h.usecase.GetAllUsers(r.Context(), userFilter)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusInternalServerError, err.Error(), err.Error())
		return
	}

	pkg.JSONResponse(w, http.StatusOK, "Success get all users", users)
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
		pkg.ErrorResponse(w, http.StatusInternalServerError, err.Error(), err.Error())
		return
	}

	pkg.JSONResponse(w, http.StatusOK, "Success get user by id", user)
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
		pkg.ErrorResponse(w, http.StatusInternalServerError, err.Error(), err.Error())
		return
	}

	pkg.JSONResponse(w, http.StatusOK, "Success delete user by id", nil)
}
