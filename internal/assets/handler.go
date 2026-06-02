package assets

import (
	"database/sql"
	"encoding/json"
	"errors"
	"inventory-it/internal/pkg"
	"net/http"
	"strconv"
)

type AssetFilter struct {
	Search  string
	Limit   int
	Page    int
	OrderBy string
}
type assetRequest struct {
	AssetName     string `json:"asset_name"`
	SerialNumber  string `json:"serial_number"`
	PurchasedDate string `json:"purchased_date"`
	Status        string `json:"status"`
	BrandId       string `json:"brand_id"`
	CategoryId    string `json:"category_id"`
}
type Handler interface {
	CreateAsset(w http.ResponseWriter, r *http.Request)
	GetAssets(w http.ResponseWriter, r *http.Request)
	GetAssetByID(w http.ResponseWriter, r *http.Request)
	UpdateAsset(w http.ResponseWriter, r *http.Request)
	DeleteAsset(w http.ResponseWriter, r *http.Request)
}
type handler struct {
	usecase Usecase
}

func NewAssetHandler(u Usecase) Handler {
	return &handler{
		usecase: u,
	}
}

func (h *handler) CreateAsset(w http.ResponseWriter, r *http.Request) {
	// Implementation for creating an asset
	var assetDataReq assetRequest

	err := json.NewDecoder(r.Body).Decode(&assetDataReq)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusBadRequest, err.Error(), err.Error())
		return
	}
	if assetDataReq.AssetName == "" || assetDataReq.SerialNumber == "" || assetDataReq.PurchasedDate == "" || assetDataReq.Status == "" || assetDataReq.BrandId == "" || assetDataReq.CategoryId == "" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "All fields are required!!", nil)
		return
	}
	purchasedDataConvert, err := pkg.ParseToDate(assetDataReq.PurchasedDate)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid purchased date format", err)
		return
	}
	if assetDataReq.Status != "available" && assetDataReq.Status != "assigned" && assetDataReq.Status != "maintenance" && assetDataReq.Status != "retired" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid status value", nil)
		return
	}
	var assetData Asset
	assetData.AssetName = assetDataReq.AssetName
	assetData.SerialNumber = assetDataReq.SerialNumber
	assetData.PurchasedDate = purchasedDataConvert
	assetData.Status = AssetStatus(assetDataReq.Status)
	assetData.BrandId = assetDataReq.BrandId
	assetData.CategoryId = assetDataReq.CategoryId

	err = h.usecase.CreateAsset(r.Context(), assetData)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusInternalServerError, "Failed to create asset", err)
		return
	}
	pkg.JSONResponse(w, http.StatusCreated, "Asset created successfully", assetData)
}

func (h *handler) GetAssets(w http.ResponseWriter, r *http.Request) {
	// Implementation for getting all assets
	var filter AssetFilter

	query := r.URL.Query()
	filter.Search = query.Get("search")
	filter.OrderBy = query.Get("order_by")

	limit := query.Get("limit")
	page := query.Get("page")

	if limit != "" {
		limitInt, err := strconv.Atoi(limit)
		if err != nil {
			pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid limit value", err)
			return
		}
		filter.Limit = limitInt
	} else {
		filter.Limit = 10 // Default limit
	}

	if page != "" {
		pageInt, err := strconv.Atoi(page)
		if err != nil {
			pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid page value", err)
			return
		}
		filter.Page = pageInt
	} else {
		filter.Page = 1 // Default page
	}

	assets, err := h.usecase.GetAllAssets(r.Context(), filter)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusInternalServerError, "Failed to get assets", err)
		return
	}
	pkg.JSONResponse(w, http.StatusOK, "Assets retrieved successfully", assets)
}

func (h *handler) GetAssetByID(w http.ResponseWriter, r *http.Request) {
	// Implementation for getting an asset by ID
	assetId := r.PathValue("asset_id")
	if assetId == "" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Asset ID is required", nil)
		return
	}
	asset, err := h.usecase.GetAssetById(r.Context(), assetId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			pkg.ErrorResponse(w, http.StatusNotFound, "Asset not found", nil)
			return
		}
		pkg.ErrorResponse(w, http.StatusInternalServerError, "Failed to get asset!!", err)
		return
	}
	pkg.JSONResponse(w, http.StatusOK, "Asset retrieved successfully!!", asset)
}

func (h *handler) UpdateAsset(w http.ResponseWriter, r *http.Request) {
	// Implementation for updating an asset
	assetId := r.PathValue("asset_id")
	if assetId == "" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Asset ID is required", nil)
		return
	}
	var assetDataReq assetRequest

	err := json.NewDecoder(r.Body).Decode(&assetDataReq)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusBadRequest, err.Error(), err.Error())
		return
	}
	if assetDataReq.AssetName == "" || assetDataReq.SerialNumber == "" || assetDataReq.PurchasedDate == "" || assetDataReq.Status == "" || assetDataReq.BrandId == "" || assetDataReq.CategoryId == "" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "All fields are required!!", nil)
		return
	}

	if assetDataReq.Status != "available" && assetDataReq.Status != "assigned" && assetDataReq.Status != "maintenance" && assetDataReq.Status != "retired" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid status value", nil)
		return
	}
	purchasedDataConvert, err := pkg.ParseToDate(assetDataReq.PurchasedDate)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid purchased date format", err)
		return
	}
	var assetData Asset
	assetData.AssetName = assetDataReq.AssetName
	assetData.SerialNumber = assetDataReq.SerialNumber
	assetData.PurchasedDate = purchasedDataConvert
	assetData.Status = AssetStatus(assetDataReq.Status)
	assetData.BrandId = assetDataReq.BrandId
	assetData.CategoryId = assetDataReq.CategoryId

	err = h.usecase.UpdateAssetById(r.Context(), assetId, assetData)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			pkg.ErrorResponse(w, http.StatusNotFound, "Asset not found", nil)
			return
		}
		pkg.ErrorResponse(w, http.StatusInternalServerError, "Failed to update asset!!", err)
		return
	}
	pkg.JSONResponse(w, http.StatusOK, "Asset updated successfully!!", assetData)
}

func (h *handler) DeleteAsset(w http.ResponseWriter, r *http.Request) {
	// Implementation for deleting an asset
	assetId := r.PathValue("asset_id")
	if assetId == "" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Asset ID is required", nil)
		return
	}
	err := h.usecase.DeleteAssetById(r.Context(), assetId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			pkg.ErrorResponse(w, http.StatusNotFound, "Asset not found", nil)
			return
		}
		pkg.ErrorResponse(w, http.StatusInternalServerError, "Failed to delete asset!!", err)
		return
	}
	pkg.JSONResponse(w, http.StatusOK, "Asset deleted successfully!!", nil)
}
