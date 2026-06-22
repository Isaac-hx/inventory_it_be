package assets

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"inventory-it/internal/pkg"
	"net/http"
	"strconv"
)

type AssetFilter struct {
	Search   string
	Limit    int
	Category string
	Status   string
	Page     int
	OrderBy  string
}
type assetResponse struct {
	AssetId       string `json:"AssetId"`
	AssetName     string `json:"AssetName,omitempty"`
	SerialNumber  string `json:"SerialNumber,omitempty"`
	PurchasedDate string `json:"PurchasedDate,omitempty"`
	Status        string `json:"Status,omitempty"`
	BrandId       string `json:"BrandId,omitempty"`
	BrandName     string `json:"BrandName,omitempty"`
	CategoryId    string `json:"CategoryId,omitempty"`
	CategoryName  string `json:"CategoryName,omitempty"`
	Description   string `json:"Description,omitempty"`
	QuantityStock int    `json:"QuantityStock,omitempty"`
	CreatedAt     string `json:"CreatedAt,omitempty"`
	UpdatedAt     string `json:"UpdatedAt,omitempty"`
}
type assetRequest struct {
	AssetName     string `json:"asset_name"`
	SerialNumber  string `json:"serial_number"`
	PurchasedDate string `json:"purchased_date"`
	Status        string `json:"status"`
	BrandId       string `json:"brand_id"`
	CategoryId    string `json:"category_id"`
	Description   string `json:"description"`
	QuantityStock string `json:"quantity_stock"`
}
type Handler interface {
	CreateAsset(w http.ResponseWriter, r *http.Request)
	GetAllAssetsData(w http.ResponseWriter, r *http.Request)
	GetAssets(w http.ResponseWriter, r *http.Request)
	GetAssetByID(w http.ResponseWriter, r *http.Request)
	UpdateAsset(w http.ResponseWriter, r *http.Request)
	DeleteAsset(w http.ResponseWriter, r *http.Request)
	GetOverview(w http.ResponseWriter, r *http.Request)
	GetGraphicDistributionByCategory(w http.ResponseWriter, r *http.Request)
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
	convertToIntStock, _ := strconv.Atoi(assetDataReq.QuantityStock)

	if assetDataReq.AssetName == "" || assetDataReq.SerialNumber == "" || assetDataReq.PurchasedDate == "" || assetDataReq.Status == "" || assetDataReq.BrandId == "" || assetDataReq.CategoryId == "" || assetDataReq.Description == "" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "All fields are required!!", nil)
		return
	}
	if convertToIntStock < 0 {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Quantity stock can't be negative!!", nil)
	}
	purchasedDataConvert, err := pkg.ParseFromStringToDate(assetDataReq.PurchasedDate)
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
	assetData.Description = assetDataReq.Description
	assetData.QuantityStock = convertToIntStock

	err = h.usecase.CreateAsset(r.Context(), assetData)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusInternalServerError, "Failed to create asset", err)
		return
	}
	pkg.JSONResponse(w, http.StatusCreated, "Asset created successfully", assetData, nil)
}

func (h *handler) GetAssets(w http.ResponseWriter, r *http.Request) {
	// Implementation for getting all assets
	var filter AssetFilter

	query := r.URL.Query()
	filter.Search = query.Get("search")
	filter.OrderBy = query.Get("order_by")
	filter.Status = query.Get("status")
	filter.Category = query.Get("category")
	limit := query.Get("limit")
	page := query.Get("page")

	if filter.Status != "" {
		isValidStatus := filter.Status == string(Available) ||
			filter.Status == string(Assigned) ||
			filter.Status == string(Retired) ||
			filter.Status == string(Maintenance)

		if !isValidStatus {
			pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid status value!!", "Status must be available, assigned, retired, or maintenance")
			return
		}
	}

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

	assets, meta, err := h.usecase.GetAllAssets(r.Context(), filter)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusInternalServerError, "Failed to get assets", err)
		return
	}

	var assetResponseData []assetResponse
	for _, item := range assets {
		var asset assetResponse
		asset.AssetId = item.AssetId
		asset.AssetName = item.AssetName
		asset.Description = item.Description
		asset.SerialNumber = item.SerialNumber
		asset.Status = string(item.Status)
		asset.PurchasedDate = pkg.ParseFromDateToString(item.PurchasedDate)
		asset.BrandId = item.Brand.BrandId
		asset.BrandName = item.Brand.BrandName
		asset.CategoryId = item.Category.CategoryId
		asset.CategoryName = item.Category.CategoryName
		asset.QuantityStock = item.QuantityStock
		asset.CreatedAt = pkg.ParseFromDateToString(item.CreatedAt)
		asset.UpdatedAt = pkg.ParseFromDateToString(item.UpdatedAt)

		assetResponseData = append(assetResponseData, asset)
	}

	pkg.JSONResponse(w, http.StatusOK, "Assets retrieved successfully", assetResponseData, meta)
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

	//parsing to response asset
	var assetResponseData assetResponse
	assetResponseData.AssetId = asset.AssetId
	assetResponseData.AssetName = asset.AssetName
	assetResponseData.Description = asset.Description
	assetResponseData.SerialNumber = asset.SerialNumber
	assetResponseData.QuantityStock = asset.QuantityStock
	assetResponseData.Status = string(asset.Status)
	assetResponseData.PurchasedDate = pkg.ParseFromDateToString(asset.PurchasedDate)
	assetResponseData.CreatedAt = pkg.ParseFromDateToString(asset.CreatedAt)
	assetResponseData.UpdatedAt = pkg.ParseFromDateToString(asset.UpdatedAt)
	assetResponseData.BrandId = asset.Brand.BrandId
	assetResponseData.BrandName = asset.Brand.BrandName
	assetResponseData.CategoryId = asset.Category.CategoryId
	assetResponseData.CategoryName = asset.Category.CategoryName
	fmt.Println("ini berjalan")
	pkg.JSONResponse(w, http.StatusOK, "Asset retrieved successfully!!", assetResponseData, nil)
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

	convertToIntStock, _ := strconv.Atoi(assetDataReq.QuantityStock)
	if assetDataReq.AssetName == "" || assetDataReq.SerialNumber == "" || assetDataReq.PurchasedDate == "" || assetDataReq.Status == "" || assetDataReq.BrandId == "" || assetDataReq.CategoryId == "" || assetDataReq.Description == "" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "All fields are required!!", nil)
		return
	}

	if convertToIntStock < 0 {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Quantity stock can't be negative!!", nil)
	}
	if assetDataReq.Status != "available" && assetDataReq.Status != "assigned" && assetDataReq.Status != "maintenance" && assetDataReq.Status != "retired" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid status value", nil)
		return
	}
	purchasedDataConvert, err := pkg.ParseFromStringToDate(assetDataReq.PurchasedDate)
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
	assetData.Description = assetDataReq.Description
	assetData.QuantityStock = convertToIntStock

	err = h.usecase.UpdateAssetById(r.Context(), assetId, assetData)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			pkg.ErrorResponse(w, http.StatusNotFound, "Asset not found", nil)
			return
		}
		pkg.ErrorResponse(w, http.StatusInternalServerError, "Failed to update asset!!", err)
		return
	}
	pkg.JSONResponse(w, http.StatusOK, "Asset updated successfully!!", assetData, nil)
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
	pkg.JSONResponse(w, http.StatusOK, "Asset deleted successfully!!", nil, nil)
}

func (h *handler) GetAllAssetsData(w http.ResponseWriter, r *http.Request) {
	assets, err := h.usecase.GetAllAssetsData(r.Context())
	if err != nil {
		pkg.ErrorResponse(w, http.StatusInternalServerError, err.Error(), err.Error())
		return
	}

	var assetResponseData []assetResponse
	for _, item := range assets {
		var asset assetResponse
		asset.AssetId = item.AssetId
		asset.AssetName = item.AssetName
		asset.Description = item.Description
		asset.SerialNumber = item.SerialNumber
		asset.Status = string(item.Status)
		asset.PurchasedDate = pkg.ParseFromDateToString(item.PurchasedDate)
		asset.BrandId = item.Brand.BrandId
		asset.BrandName = item.Brand.BrandName
		asset.CategoryId = item.Category.CategoryId
		asset.CategoryName = item.Category.CategoryName
		asset.QuantityStock = item.QuantityStock
		asset.CreatedAt = pkg.ParseFromDateToString(item.CreatedAt)
		asset.UpdatedAt = pkg.ParseFromDateToString(item.UpdatedAt)

		assetResponseData = append(assetResponseData, asset)
	}

	pkg.JSONResponse(w, http.StatusOK, "Success retrieve data assets", assetResponseData, nil)
}

func (h *handler) GetOverview(w http.ResponseWriter, r *http.Request) {

	overviewData, err := h.usecase.GetOverviewData(r.Context())
	if err != nil {
		pkg.ErrorResponse(w, http.StatusInternalServerError, err.Error(), err.Error())
		return
	}

	pkg.JSONResponse(w, http.StatusOK, "Success retrieve data overview aset", overviewData, nil)
}

func (h *handler) GetGraphicDistributionByCategory(w http.ResponseWriter, r *http.Request) {
	infoGraphic, err := h.usecase.GetCountGroupCategoryAssets(r.Context())
	if err != nil {
		pkg.ErrorResponse(w, http.StatusInternalServerError, err.Error(), err.Error())
		return
	}
	pkg.JSONResponse(w, http.StatusOK, "Success retrieve data graphic asset", infoGraphic, nil)
}
