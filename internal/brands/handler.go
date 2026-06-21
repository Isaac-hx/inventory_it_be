package brands

import (
	"database/sql"
	"encoding/json"
	"errors"
	"inventory-it/internal/pkg"
	"net/http"
	"strconv"
)

type BrandResponse struct {
	BrandId   string `json:"BrandId,omitempty"`
	BrandName string `json:"BrandName,omitempty"`
	CreatedAt string `json:"CreatedAt,omitempty"`
	UpdatedAt string `json:"UpdatedAt,omitempty"`
}

type BrandReq struct {
	BrandName string `json:"brand_name"`
}
type BrandFilter struct {
	Search  string
	Limit   int
	Page    int
	OrderBy string
}
type Handler interface {
	GetAllBrand(w http.ResponseWriter, r *http.Request)
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

func (h *handler) GetAllBrand(w http.ResponseWriter, r *http.Request) {
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

	brands, meta, err := h.usecase.GetAllBrands(r.Context(), brandFilter)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusInternalServerError, "Failed to get Brand", err.Error())
		return
	}
	var brandResponseData []BrandResponse

	for _, item := range brands {
		var brand BrandResponse
		brand.BrandId = item.BrandId
		brand.BrandName = item.BrandName
		brand.CreatedAt = pkg.ParseFromDateToString(item.CreatedAt)
		brand.UpdatedAt = pkg.ParseFromDateToString(item.UpdatedAt)

		brandResponseData = append(brandResponseData, brand)

	}
	pkg.JSONResponse(w, http.StatusOK, "Success get all Brand", brandResponseData, meta)
}

func (h *handler) GetBrandById(w http.ResponseWriter, r *http.Request) {
	//get brand id from url params
	brandId := r.PathValue("brand_id")

	brand, err := h.usecase.GetBrandById(r.Context(), brandId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			pkg.ErrorResponse(w, http.StatusNotFound, "Brand not found", nil)
			return
		}
		pkg.ErrorResponse(w, http.StatusInternalServerError, "Failed to get brand", err.Error())
		return
	}
	var responseBrand BrandResponse

	responseBrand.BrandId = brand.BrandId
	responseBrand.BrandName = brand.BrandName
	responseBrand.CreatedAt = pkg.ParseFromDateToString(brand.CreatedAt)
	responseBrand.UpdatedAt = pkg.ParseFromDateToString(brand.UpdatedAt)

	pkg.JSONResponse(w, http.StatusOK, "Success get brand by id", responseBrand, nil)
}

func (h *handler) CreateBrand(w http.ResponseWriter, r *http.Request) {
	//parse request body to struct brand
	var brand BrandReq
	if brand.BrandName == "anjing" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid request cant use toxic word", errors.New("Invalid request cant use toxic word"))
		return
	}
	err := json.NewDecoder(r.Body).Decode(&brand)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if brand.BrandName == "" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Brand name is required", nil)
		return
	}

	var createBrand Brand
	createBrand.BrandName = brand.BrandName

	err = h.usecase.CreateBrand(r.Context(), createBrand)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusInternalServerError, "Failed to create brand", err.Error())
		return
	}

	pkg.JSONResponse(w, http.StatusOK, "Success create brand", nil, nil)
}

func (h *handler) UpdateBrand(w http.ResponseWriter, r *http.Request) {
	//get brand id from url params
	brandId := r.PathValue("brand_id")

	//parse request body to struct brand
	var brand BrandReq
	err := json.NewDecoder(r.Body).Decode(&brand)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}
	if brand.BrandName == "" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Brand name is required", nil)
		return
	}

	var updateBrand Brand
	updateBrand.BrandName = brand.BrandName

	err = h.usecase.UpdateBrand(r.Context(), brandId, updateBrand)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			pkg.ErrorResponse(w, http.StatusNotFound, "Brand not found", nil)
			return
		}
		pkg.ErrorResponse(w, http.StatusInternalServerError, "Failed to update brand", err.Error())
		return
	}

	pkg.JSONResponse(w, http.StatusOK, "Success update brand", nil, nil)
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

	pkg.JSONResponse(w, http.StatusOK, "Success delete brand", nil, nil)
}
