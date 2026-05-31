package brands

import (
	"database/sql"
	"encoding/json"
	"errors"
	"inventory-it/internal/pkg"
	"net/http"
	"strconv"
)

type BrandFilter struct {
	Search  string
	Limit   int
	Page    int
	OrderBy string
}
type Handler interface {
	GetAllBrands(w http.ResponseWriter, r *http.Request)
	GetBrandById(w http.ResponseWriter, r *http.Request)
	CreateBrand(w http.ResponseWriter, r *http.Request)
	UpdateBrand(w http.ResponseWriter, r *http.Request)
	DeleteBrand(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	usecase Usecase
}

func NewHandler(u Usecase) Handler {
	return &handler{
		usecase: u,
	}
}

func (h *handler) GetAllBrands(w http.ResponseWriter, r *http.Request) {
	var brandFilter BrandFilter

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
	brandFilter.Search = search
	brandFilter.Limit = limit
	brandFilter.Page = page
	brandFilter.OrderBy = orderBy

	brands, err := h.usecase.GetAllBrands(r.Context(), brandFilter)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusInternalServerError, "Failed to get brands", err.Error())
		return
	}

	pkg.JSONResponse(w, http.StatusOK, "Success get all brands", brands)
}

func (h *handler) GetBrandById(w http.ResponseWriter, r *http.Request) {
	//get brand id from url params
	brandId := r.PathValue("brand_id")

	brand, err := h.usecase.GetBrandById(r.Context(), brandId)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusInternalServerError, "Failed to get brand", err.Error())
		return
	}

	pkg.JSONResponse(w, http.StatusOK, "Success get brand by id", brand)
}

func (h *handler) CreateBrand(w http.ResponseWriter, r *http.Request) {
	//parse request body to struct brand
	var brand Brands
	err := json.NewDecoder(r.Body).Decode(&brand)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	err = h.usecase.CreateBrand(r.Context(), brand)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusInternalServerError, "Failed to create brand", err.Error())
		return
	}

	pkg.JSONResponse(w, http.StatusOK, "Success create brand", nil)
}

func (h *handler) UpdateBrand(w http.ResponseWriter, r *http.Request) {
	//get brand id from url params
	brandId := r.PathValue("brand_id")

	//parse request body to struct brand
	var brand Brands
	err := json.NewDecoder(r.Body).Decode(&brand)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	err = h.usecase.UpdateBrand(r.Context(), brandId, brand)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			pkg.ErrorResponse(w, http.StatusNotFound, "Brand not found", nil)
			return
		}
		pkg.ErrorResponse(w, http.StatusInternalServerError, "Failed to update brand", err.Error())
		return
	}

	pkg.JSONResponse(w, http.StatusOK, "Success update brand", nil)
}

func (h *handler) DeleteBrand(w http.ResponseWriter, r *http.Request) {
	//get brand id from url params
	brandId := r.PathValue("brand_id")
	err := h.usecase.DeleteBrand(r.Context(), brandId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			pkg.ErrorResponse(w, http.StatusNotFound, "Brand not found", nil)
			return
		}
		pkg.ErrorResponse(w, http.StatusInternalServerError, "Failed to delete brand", err.Error())
		return
	}

	pkg.JSONResponse(w, http.StatusOK, "Success delete brand", nil)
}
