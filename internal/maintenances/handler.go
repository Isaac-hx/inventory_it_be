package maintenances

import (
	"database/sql"
	"encoding/json"
	"errors"
	"inventory-it/internal/pkg"
	"net/http"
	"strconv"
	"time"
)

type MaintenanceReq struct {
	Description string `json:"description"`
	Cost        int64  `json:"cost"`
	Status      string `json:"status"`
	AssetId     string `json:"asset_id"`
}

type MaintenanceFilter struct {
	Search  string
	Status  string
	Limit   int
	Page    int
	OrderBy string
}

type Handler interface {
	GetAllMaintenances(w http.ResponseWriter, r *http.Request)
	GetMaintenanceById(w http.ResponseWriter, r *http.Request)
	CreateMaintenance(w http.ResponseWriter, r *http.Request)
	UpdateStatusMaintenance(w http.ResponseWriter, r *http.Request)
	UpdateMaintenance(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	usecase Usecase
}

func NewMaintenanceHandler(usecase Usecase) Handler {
	return &handler{
		usecase: usecase,
	}
}

func (h *handler) GetAllMaintenances(w http.ResponseWriter, r *http.Request) {
	var filter MaintenanceFilter

	// Get query parameters
	query := r.URL.Query()
	filter.Search = query.Get("search")
	filter.Status = query.Get("status")

	// Parse limit and page query parameters
	limitStr := query.Get("limit")
	pageStr := query.Get("page")
	orderBy := query.Get("order_by")

	//handling validation
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10 // Default limit
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1 // Default page
	}
	if filter.Status != "" && filter.Status != string(Pending) && filter.Status != string(InProgress) && filter.Status != string(Completed) {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid status value!!", nil)
		return
	}
	// Get order_by query parameter
	if orderBy == "created_at_asc" || orderBy == "created_at_desc" || orderBy == "cost_asc" || orderBy == "cost_desc" {
		filter.OrderBy = orderBy
	} else {
		filter.OrderBy = "created_at_desc" // Default order by
	}

	filter.Page = page
	filter.Limit = limit

	// Call usecase to get maintenances
	maintenances, err := h.usecase.GetAllMaintenances(r.Context(), filter)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusInternalServerError, "Failed to get maintenances!!", err)
		return
	}
	pkg.JSONResponse(w, http.StatusOK, "Maintenances retrieved successfully!!", maintenances)
}

func (h *handler) GetMaintenanceById(w http.ResponseWriter, r *http.Request) {
	maintenanceId := r.PathValue("maintenance_id")
	if maintenanceId == "" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Maintenance ID is required", nil)
		return
	}
	maintenance, err := h.usecase.GetMaintenanceById(r.Context(), maintenanceId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			pkg.ErrorResponse(w, http.StatusNotFound, "Maintenance not found", nil)
			return
		}
		pkg.ErrorResponse(w, http.StatusInternalServerError, "Failed to get maintenance!!", err)
		return
	}
	pkg.JSONResponse(w, http.StatusOK, "Maintenance retrieved successfully!!", maintenance)
}

func (h *handler) CreateMaintenance(w http.ResponseWriter, r *http.Request) {
	var maintenanceRequest MaintenanceReq
	err := json.NewDecoder(r.Body).Decode(&maintenanceRequest)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid request body!!", err)
		return
	}
	if maintenanceRequest.Description == "" || maintenanceRequest.AssetId == "" || maintenanceRequest.Status == "" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Description, Asset ID, and Status are required!!", nil)
		return
	}
	if maintenanceRequest.Status != string(Pending) && maintenanceRequest.Status != string(InProgress) && maintenanceRequest.Status != string(Completed) {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid status value!!", nil)
		return
	}

	var createMaintenance Maintenance
	createMaintenance.Description = maintenanceRequest.Description
	createMaintenance.Cost = maintenanceRequest.Cost
	createMaintenance.Status = MaintenanceStatus(maintenanceRequest.Status)
	createMaintenance.AssetId = maintenanceRequest.AssetId
	createMaintenance.MaintenanceAt = time.Now()

	maintenanceData, err := h.usecase.CreateMaintenance(r.Context(), createMaintenance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			pkg.ErrorResponse(w, http.StatusNotFound, "Asset not found", nil)
			return
		}
		pkg.ErrorResponse(w, http.StatusInternalServerError, "Failed to create maintenance!!", err)
		return
	}
	pkg.JSONResponse(w, http.StatusCreated, "Maintenance created successfully!!", maintenanceData)
}

func (h *handler) UpdateMaintenance(w http.ResponseWriter, r *http.Request) {
	// This function can be implemented to update maintenance details such as description and cost.
	// It would involve parsing the request body for the new details, validating them, and then calling the usecase to perform the update.
	var maintenanceReq MaintenanceReq
	err := json.NewDecoder(r.Body).Decode(&maintenanceReq)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid request body!!", err)
		return
	}
	maintenanceId := r.PathValue("maintenance_id")
	if maintenanceId == "" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Maintenance ID is required", nil)
		return
	}

	if maintenanceReq.Description == "" || maintenanceReq.AssetId == "" || maintenanceReq.Status == "" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Description, Asset ID, and Status are required!!", nil)
		return
	}
	if maintenanceReq.Status != string(Pending) && maintenanceReq.Status != string(InProgress) && maintenanceReq.Status != string(Completed) {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid status value!!", nil)
		return
	}
	if maintenanceReq.Cost < 0 {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Cost cannot be negative!!", nil)
		return
	}

	var maintenanceUpdated Maintenance
	maintenanceUpdated.MaintenanceId = maintenanceId
	maintenanceUpdated.Description = maintenanceReq.Description
	maintenanceUpdated.Cost = maintenanceReq.Cost
	maintenanceUpdated.Status = MaintenanceStatus(maintenanceReq.Status)
	maintenanceUpdated.AssetId = maintenanceReq.AssetId

	err = h.usecase.UpdateMaintenance(r.Context(), maintenanceUpdated)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			pkg.ErrorResponse(w, http.StatusNotFound, "Maintenance not found", nil)
			return
		}
		pkg.ErrorResponse(w, http.StatusInternalServerError, "Failed to update maintenance!!", err)
		return
	}
	pkg.JSONResponse(w, http.StatusOK, "Maintenance updated successfully!!", nil)
}

func (h *handler) UpdateStatusMaintenance(w http.ResponseWriter, r *http.Request) {
	var statusUpdate struct {
		Status string `json:"status"`
	}
	maintenanceId := r.PathValue("maintenance_id")
	if maintenanceId == "" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Maintenance ID is required", nil)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&statusUpdate)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid request body!!", err)
		return
	}
	if statusUpdate.Status != string(Pending) && statusUpdate.Status != string(InProgress) && statusUpdate.Status != string(Completed) && statusUpdate.Status != string(Cancelled) {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid status value!!", nil)
		return
	}

	var maintenanceUpdated Maintenance
	maintenanceUpdated.MaintenanceId = maintenanceId
	maintenanceUpdated.Status = MaintenanceStatus(statusUpdate.Status)
	if statusUpdate.Status == string(Completed) {
		maintenanceUpdated.CompletedAt = time.Now()
	}

	err = h.usecase.UpdateStatusMaintenance(r.Context(), maintenanceUpdated)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			pkg.ErrorResponse(w, http.StatusNotFound, "Maintenance not found", nil)
			return
		}
		pkg.ErrorResponse(w, http.StatusInternalServerError, "Failed to update maintenance status!!", err)
		return
	}
	pkg.JSONResponse(w, http.StatusOK, "Maintenance status updated successfully!!", nil)
}
