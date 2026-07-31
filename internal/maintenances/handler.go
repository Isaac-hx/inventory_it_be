package maintenances

import (
	"database/sql"
	"encoding/json"
	"errors"
	"inventory-it/internal/pkg"
	"log"
	"net/http"
	"strconv"
	"time"
)

type MaintenanceRes struct {
	MaintenanceId string `json:"MaintenanceId"`
	Description   string `json:"Description"`
	Cost          int64  `json:"Cost"`
	Status        string `json:"Status"`
	AssetId       string `json:"AssetId"`
	AssignmentId  string `json:"AssignmentId"`
	AssetName     string `json:"AssetName"`
	MaintenanceAt string `json:"MaintenanceAt,omitempty"`
	CompletedAt   string `json:"CompletedAt,omitempty"`
	SerialNumber  string `json:"SerialNumber,omitempty"`
	Ram           string `json:"Ram,omitempty"`
	Storage       string `json:"Storage,omitempty"`
	Processor     string `json:"Processor,omitempty"`
	BrandId       string `json:"BrandId,omitempty"`
	BrandName     string `json:"BrandName,omitempty"`
	CategoryId    string `json:"CategoryId,omitempty"`
	CategoryName  string `json:"CategoryName,omitempty"`

	CreatedAt string `json:"CreatedAt"`
	UpdatedAt string `json:"UpdatedAt"`
}
type MaintenanceReq struct {
	Description   string `json:"description,omitempty"`
	Cost          int64  `json:"cost,omitempty"`
	Status        string `json:"status,omitempty"`
	AssetId       string `json:"asset_id,omitempty"`
	AssignmentId  string `json:"assignment_id,omitempty"`
	MaintenanceAt string `json:"maintenance_at,omitempty"`
	CompletedAt   string `json:"completed_at,omitempty"`
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
	UpdateMaintenance(w http.ResponseWriter, r *http.Request)
	UpdateStatusMaintenanceById(w http.ResponseWriter, r *http.Request)
	GetAllMaintenancesData(w http.ResponseWriter, r *http.Request)
	CreateRequest(w http.ResponseWriter, r *http.Request)
	GetAllMaintenancesByUserId(w http.ResponseWriter, r *http.Request)
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
	maintenances, meta, err := h.usecase.GetAllMaintenances(r.Context(), filter)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusInternalServerError, "Failed to get maintenances!!", err.Error())
		return
	}
	//response to client
	var responseMaintenance []MaintenanceRes
	for _, item := range maintenances {
		var maintenance MaintenanceRes
		maintenance.MaintenanceId = item.MaintenanceId
		maintenance.AssetName = item.Asset.AssetName
		maintenance.AssetId = item.Asset.AssetId
		maintenance.Description = item.Description
		maintenance.Cost = item.Cost
		maintenance.MaintenanceAt = pkg.ParseFromDateToString(item.MaintenanceAt)
		maintenance.CompletedAt = ""
		maintenance.Status = string(item.Status)

		responseMaintenance = append(responseMaintenance, maintenance)
	}

	pkg.JSONResponse(w, http.StatusOK, "Maintenances retrieved successfully!!", responseMaintenance, meta)
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

	var maintenanceResponseData MaintenanceRes
	maintenanceResponseData.MaintenanceId = maintenance.MaintenanceId
	maintenanceResponseData.Description = maintenance.Description
	maintenanceResponseData.Cost = maintenance.Cost
	maintenanceResponseData.Status = string(maintenance.Status)
	if maintenance.CompletedAt == nil {
		maintenanceResponseData.CompletedAt = "-"
	} else {
		maintenanceResponseData.CompletedAt = pkg.ParseFromDateToString(*maintenance.CompletedAt)

	}
	maintenanceResponseData.MaintenanceAt = pkg.ParseFromDateToString(maintenance.MaintenanceAt)
	maintenanceResponseData.AssetId = maintenance.Asset.AssetId
	maintenanceResponseData.AssetName = maintenance.Asset.AssetName
	maintenanceResponseData.SerialNumber = maintenance.Asset.SerialNumber

	maintenanceResponseData.CategoryId = maintenance.Category.CategoryId
	maintenanceResponseData.CategoryName = maintenance.Category.CategoryName

	maintenanceResponseData.BrandId = maintenance.Brand.BrandId
	maintenanceResponseData.BrandName = maintenance.Brand.BrandName

	maintenanceResponseData.CreatedAt = pkg.ParseFromDateToString(maintenance.CreatedAt)
	maintenanceResponseData.UpdatedAt = pkg.ParseFromDateToString(maintenance.UpdatedAt)
	pkg.JSONResponse(w, http.StatusOK, "Maintenance retrieved successfully!!", maintenanceResponseData, nil)
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
	createMaintenance.Assignment.AssignmentId = maintenanceRequest.AssignmentId
	timeParse, err := pkg.ParseFromStringToDate(maintenanceRequest.MaintenanceAt)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid type time data", nil)
		return

	}

	log.Println(createMaintenance.Assignment.AssetId)
	createMaintenance.MaintenanceAt = timeParse

	maintenanceData, err := h.usecase.CreateMaintenance(r.Context(), createMaintenance)
	if err != nil {

		pkg.ErrorResponse(w, http.StatusInternalServerError, "Failed to create maintenance!!", err.Error())
		return
	}
	pkg.JSONResponse(w, http.StatusCreated, "Maintenance created successfully!!", maintenanceData, nil)
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
	maintenanceUpdated.Asset.AssetId = maintenanceReq.AssetId

	err = h.usecase.UpdateMaintenance(r.Context(), maintenanceId, maintenanceUpdated)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			pkg.ErrorResponse(w, http.StatusNotFound, "Maintenance not found", nil)
			return
		}
		pkg.ErrorResponse(w, http.StatusInternalServerError, "Failed to update maintenance!!", err)
		return
	}
	pkg.JSONResponse(w, http.StatusOK, "Maintenance updated successfully!!", nil, nil)
}

func (h *handler) GetAllMaintenancesData(w http.ResponseWriter, r *http.Request) {
	data, err := h.usecase.GetAllMaintenancesData(r.Context())
	if err != nil {
		pkg.ErrorResponse(w, http.StatusInternalServerError, err.Error(), err.Error())
		return
	}

	// 1. Inisialisasi dengan make agar aman jika data kosong (mengembalikan [] bukan null)
	maintenanceResponseData := make([]MaintenanceRes, 0)

	// 2. Perbaikan sintaks looping yang benar: for _, item := range data
	for _, item := range data {
		var maintenance MaintenanceRes

		// Pemetaan data utama Maintenance
		maintenance.MaintenanceId = item.MaintenanceId
		maintenance.Description = item.Description
		maintenance.Cost = item.Cost
		maintenance.Status = string(item.Status)
		maintenance.AssetId = item.Asset.AssetId

		// 3. Pemetaan & Pemformatan Waktu/Tanggal menjadi string yang rapi
		maintenance.MaintenanceAt = item.MaintenanceAt.Format("02 January 2006")
		// Pastikan tipe field di struct MaintenanceRes adalah string
		if item.CompletedAt != nil {
			maintenance.CompletedAt = pkg.ParseFromDateToString(*item.CompletedAt)
		} else {
			maintenance.CompletedAt = "" // Lempar string kosong jika nil
		}
		maintenance.CreatedAt = pkg.ParseFromDateToString(item.CreatedAt)
		maintenance.UpdatedAt = pkg.ParseFromDateToString(item.UpdatedAt)

		// 4. Pemetaan data dari Relasi (Nested Struct dari hasil JOIN)
		maintenance.AssetName = item.Asset.AssetName
		maintenance.SerialNumber = item.Asset.SerialNumber
		maintenance.BrandId = item.Asset.Brand.BrandId
		maintenance.BrandName = item.Asset.Brand.BrandName
		maintenance.CategoryId = item.Asset.Category.CategoryId
		maintenance.CategoryName = item.Asset.Category.CategoryName

		// 5. KUNCI UTAMA: Masukkan objek yang sudah di-mapping ke dalam slice
		maintenanceResponseData = append(maintenanceResponseData, maintenance)
	}

	// 6. Kirim respon JSON sukses ke client
	pkg.JSONResponse(w, http.StatusOK, "Success retrieve data maintenance", maintenanceResponseData, nil)
}

func (h *handler) UpdateStatusMaintenanceById(w http.ResponseWriter, r *http.Request) {
	maintainanceId := r.PathValue("maintenance_id")
	if maintainanceId == "" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Maintenance ID is required", nil)
		return
	}
	var maintenanceReq MaintenanceReq
	err := json.NewDecoder(r.Body).Decode(&maintenanceReq)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid request body!!", err)
		return
	}

	if maintenanceReq.Status != string(Pending) && maintenanceReq.Status != string(InProgress) && maintenanceReq.Status != string(Completed) && maintenanceReq.Status != string(Cancelled) {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid status value!!", nil)
		return
	}
	var completedAt *time.Time

	// 2. Logika Pengecekan Status
	if maintenanceReq.Status == string(Completed) {
		// Jika statusnya Completed, wajib parsing string ke time.Time
		parsedDate, err := pkg.ParseFromStringToDate(maintenanceReq.CompletedAt)
		if err != nil {
			pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid completed at format!!", err.Error())
			return
		}
		// Ambil alamat memorinya (pointer) agar tidak nil
		completedAt = &parsedDate
	} else {
		// Jika status BUKAN Completed (misal: under_maintenance / cancelled),
		// kita set nil agar di database jadi NULL
		completedAt = nil
	}

	err = h.usecase.UpdateStatusMaintenance(r.Context(), maintainanceId, maintenanceReq.Status, completedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			pkg.ErrorResponse(w, http.StatusNotFound, "Maintenance ID not found!!", err.Error())
			return
		}
		pkg.ErrorResponse(w, http.StatusInternalServerError, err.Error(), err.Error())
		return
	}

	pkg.JSONResponse(w, http.StatusOK, "Succesfully updated status", nil, nil)
}

func (h *handler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	var maintenanceReq MaintenanceReq
	err := json.NewDecoder(r.Body).Decode(&maintenanceReq)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusBadRequest, err.Error(), err.Error())
		return
	}

	if maintenanceReq.AssetId == "" && maintenanceReq.Description == "" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid asset ID and description", nil)
		return
	}
	var createMaintenance Maintenance
	createMaintenance.Asset.AssetId = maintenanceReq.AssetId
	createMaintenance.Description = maintenanceReq.Description
	err = h.usecase.CreateRequest(r.Context(), createMaintenance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			pkg.ErrorResponse(w, http.StatusNotFound, "ID not found!", nil)
			return
		}
		pkg.ErrorResponse(w, http.StatusInternalServerError, err.Error(), err.Error())
		return
	}
	pkg.JSONResponse(w, http.StatusCreated, "Success create request maintenance!!", nil, nil)
}

func (h *handler) GetAllMaintenancesByUserId(w http.ResponseWriter, r *http.Request) {
	maintenances, err := h.usecase.GetAllMaintenancesByUserId(r.Context())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			pkg.ErrorResponse(w, http.StatusNotFound, "User Id is not found", nil)
			return
		}
		pkg.ErrorResponse(w, http.StatusInternalServerError, err.Error(), err.Error())
		return
	}
	var maintenancesResponseData []MaintenanceRes
	for _, item := range maintenances {
		var maintenance MaintenanceRes
		maintenance.MaintenanceId = item.MaintenanceId
		maintenance.Status = string(item.Status)
		maintenance.Description = item.Description
		maintenance.Cost = item.Cost
		maintenance.MaintenanceAt = pkg.ParseFromDateToString(item.MaintenanceAt)
		if item.CompletedAt == nil {
			maintenance.CompletedAt = "-"
		} else {
			maintenance.CompletedAt = pkg.ParseFromDateToString(*item.CompletedAt)

		}
		maintenance.CreatedAt = pkg.ParseFromDateToString(item.CreatedAt)
		maintenance.UpdatedAt = pkg.ParseFromDateToString(item.UpdatedAt)

		maintenance.AssetName = item.Asset.AssetName
		maintenance.SerialNumber = item.Asset.SerialNumber
		maintenance.Processor = item.Asset.Processor
		maintenance.Ram = item.Asset.Ram
		maintenance.Storage = item.Asset.Storage

		maintenance.CategoryName = item.Asset.Category.CategoryName

		maintenance.BrandName = item.Asset.Brand.BrandName

		maintenancesResponseData = append(maintenancesResponseData, maintenance)
	}

	pkg.JSONResponse(w, http.StatusOK, "Success retrieve maintenance by user ID", maintenancesResponseData, nil)
}
