package categories

import (
	"database/sql"
	"encoding/json"
	"errors"
	"inventory-it/internal/pkg"
	"net/http"
	"strconv"
)

type CategoriesReq struct {
	CategoryName string `json:"category_name"`
}
type CategoryFilter struct {
	Search  string
	Limit   int
	Page    int
	OrderBy string
}
type Handler interface {
	GetAllCategories(w http.ResponseWriter, r *http.Request)
	GetCategoryById(w http.ResponseWriter, r *http.Request)
	CreateCategory(w http.ResponseWriter, r *http.Request)
	UpdateCategory(w http.ResponseWriter, r *http.Request)
	DeleteCategory(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	usecase Usecase
}

func NewCategoryHandler(u Usecase) Handler {
	return &handler{
		usecase: u,
	}
}

func (h *handler) GetAllCategories(w http.ResponseWriter, r *http.Request) {
	var categoryFilter *CategoryFilter

	//get query params
	query := r.URL.Query()
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
	categoryFilter = &CategoryFilter{}
	categoryFilter.Search = search
	categoryFilter.Limit = limit
	categoryFilter.Page = page
	categoryFilter.OrderBy = orderBy
	categories, err := h.usecase.GetAllCategories(r.Context(), categoryFilter)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusInternalServerError, "Failed to get categories", err.Error())
		return
	}
	pkg.JSONResponse(w, http.StatusOK, "Success get all categories", categories)
}

func (h *handler) GetCategoryById(w http.ResponseWriter, r *http.Request) {
	categoryId := r.PathValue("category_id")
	category, err := h.usecase.GetCategoryById(r.Context(), categoryId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			pkg.ErrorResponse(w, http.StatusNotFound, "Category not found", nil)
			return
		}
		pkg.ErrorResponse(w, http.StatusInternalServerError, "Failed to get category", err.Error())
		return
	}
	pkg.JSONResponse(w, http.StatusOK, "Success get category", category)
}
func (h *handler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var category CategoriesReq
	err := json.NewDecoder(r.Body).Decode(&category)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}
	if category.CategoryName == "" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Category name is required", nil)
		return
	}
	var createCategory Categories
	createCategory.CategoryName = category.CategoryName
	err = h.usecase.CreateCategory(r.Context(), createCategory)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusInternalServerError, "Failed to create category", err.Error())
		return
	}
	pkg.JSONResponse(w, http.StatusOK, "Success create category", nil)
}

func (h *handler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	categoryId := r.PathValue("category_id")
	var category CategoriesReq
	err := json.NewDecoder(r.Body).Decode(&category)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}
	if category.CategoryName == "" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Category name is required", nil)
		return
	}
	var updateCategory Categories
	updateCategory.CategoryName = category.CategoryName
	err = h.usecase.UpdateCategory(r.Context(), categoryId, updateCategory)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			pkg.ErrorResponse(w, http.StatusNotFound, "Category not found", nil)
			return
		}
		pkg.ErrorResponse(w, http.StatusInternalServerError, "Failed to update category", err.Error())
		return
	}
	pkg.JSONResponse(w, http.StatusOK, "Success update category", nil)
}

func (h *handler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	categoryId := r.PathValue("category_id")
	err := h.usecase.DeleteCategory(r.Context(), categoryId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			pkg.ErrorResponse(w, http.StatusNotFound, "Category not found", nil)
			return
		}
		pkg.ErrorResponse(w, http.StatusInternalServerError, "Failed to delete category", err.Error())
		return
	}
	pkg.JSONResponse(w, http.StatusOK, "Success delete category", nil)
}
