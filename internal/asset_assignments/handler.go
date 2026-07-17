package assetassignments

import (
	"database/sql"
	"encoding/json"
	"errors"
	"inventory-it/internal/pkg"
	"net/http"
	"strconv"
	"time"
)

type AssetAssignmentFilter struct {
	Search  string
	Limit   int
	Status  string
	Page    int
	OrderBy string
}

type assetAssignmentRequest struct {
	AssetId      string `json:"asset_id"`
	UserId       string `json:"user_id"`
	Corporation  string `json:"corporation"`
	Notes        string `json:"notes"`
	Status       string `json:"status"`
	AssignedDate string `json:"assigned_date"`
}

type assetAssignmentResponse struct {
	AssignmentId string `json:"AssignmentId"`
	Notes        string `json:"Notes,omitempty"`
	Status       string `json:"Status,omitempty"`
	AssignedDate string `json:"AssignedDate"`
	ReturnDate   string `json:"ReturnDate,omitempty"`
	Corporation  string `json:"Corporation,omitempty"`
	UserId       string `json:"UserId"`
	Username     string `json:"Username"`

	//user assgined
	AssignedToId       string `json:"AssignedToId,omitempty"`
	AssignedToEmail    string `json:"AssignedToEmail,omitempty"`
	AssignedToUsername string `json:"AssignedToUsername,omitempty"`
	AssignedToRole     string `json:"AssignedToRole,omitempty"`

	//user assignedby
	AssignedById       string `json:"AssignedById,omitempty"`
	AssignedByUsername string `json:"AssignedByUsername,omitempty"`

	//asset
	AssetId       string `json:"AssetId"`
	AssetName     string `json:"AssetName,omitempty"`
	Processor     string `json:"Processor,omitempty"`
	Ram           string `json:"Ram,omitempty"`
	Storage       string `json:"Storage,omitempty"`
	SerialNumber  string `json:"SerialNumber,omitempty"`
	PurchasedDate string `json:"PurchasedDate,omitempty"`
	QuantityStock int    `json:"QuantityStock,omitempty"`

	//brand
	BrandId   string `json:"BrandId,omitempty"`
	BrandName string `json:"BrandName,omitempty"`

	//category
	CategoryId   string `json:"CategoryId,omitempty"`
	CategoryName string `json:"CategoryName,omitempty"`

	//department
	DepartmentId   string `json:"DepartmentId,omitempty"`
	DepartmentName string `json:"DepartmentName,omitempty"`
	CreatedAt      string `json:"CreatedAt"`
	UpdatedAt      string `json:"UpdatedAt"`
}

type Handler interface {
	GetAllAssetAssignments(w http.ResponseWriter, r *http.Request)
	GetAllAssignmentsData(w http.ResponseWriter, r *http.Request)
	GetAssetAssignmentById(w http.ResponseWriter, r *http.Request)
	CreateAssetAssignment(w http.ResponseWriter, r *http.Request) // Fix typo uppercase S
	UpdateAssetAssignment(w http.ResponseWriter, r *http.Request)
	UpdateAssetAssignmentStatus(w http.ResponseWriter, r *http.Request)
	GetAssignmentsByUserId(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	usecase UsecaseAssetAssingments
}

// ✅ FIX: Menambahkan nama parameter 'u' sebelum tipe data usecase
func NewHandlerAssetAssignments(u UsecaseAssetAssingments) Handler {
	return &handler{
		usecase: u,
	}
}

func (h *handler) GetAllAssetAssignments(w http.ResponseWriter, r *http.Request) {
	var filter AssetAssignmentFilter
	query := r.URL.Query()
	filter.Search = query.Get("search")
	filter.Status = query.Get("status")

	limitStr := query.Get("limit")
	pageStr := query.Get("page")
	orderBy := query.Get("order_by")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	// ✅ FIX: Memperbaiki logical bug pada validasi status menggunakan teknik whitelist mapping
	if filter.Status != "" {
		isValidStatus := filter.Status == string(Assigned) ||
			filter.Status == string(Damaged) ||
			filter.Status == string(Returned) ||
			filter.Status == string(Lost)

		if !isValidStatus {
			pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid status value!!", "Status must be assigned, damaged, returned, or lost")
			return
		}
	}

	if orderBy == "created_at_asc" || orderBy == "created_at_desc" || orderBy == "assigned_date_asc" || orderBy == "assigned_date_desc" {
		filter.OrderBy = orderBy
	} else {
		filter.OrderBy = "created_at_desc"
	}

	filter.Page = page
	filter.Limit = limit

	assignments, meta, err := h.usecase.GetAllAssetAssignments(r.Context(), filter)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusInternalServerError, "Failed to fetch asset assignments", err.Error())
		return
	}

	var assignmentResponse []assetAssignmentResponse // Inisialisasi empty slice agar di JSON tidak null
	for _, item := range assignments {
		var assignmentItem assetAssignmentResponse
		assignmentItem.AssignmentId = item.AssignmentId
		assignmentItem.AssetId = item.AssetId
		assignmentItem.AssetName = item.Asset.AssetName // Pastikan entity domain memiliki field AssetName
		assignmentItem.UserId = item.UserId
		assignmentItem.Username = item.User.Username // ✅ FIX: Hapus duplikasi baris
		assignmentItem.AssignedById = item.AssignedById
		assignmentItem.AssignedByUsername = item.AssignedByUsername
		assignmentItem.Status = string(item.Status)
		assignmentItem.Corporation = item.Corporation
		assignmentItem.AssignedDate = pkg.ParseFromDateToString(item.AssignedDate)

		assignmentItem.Processor = item.Asset.Processor
		assignmentItem.Storage = item.Asset.Storage
		assignmentItem.SerialNumber = item.Asset.SerialNumber
		assignmentItem.Ram = item.Asset.Ram
		assignmentResponse = append(assignmentResponse, assignmentItem)
	}
	pkg.JSONResponse(w, http.StatusOK, "Success retrieve data!", assignmentResponse, meta)
}

// 🚀 LANJUTAN KODE: GET BY ID
func (h *handler) GetAssetAssignmentById(w http.ResponseWriter, r *http.Request) {
	// Ambil ID dari query params (misal: /assignments?id=xxx) atau dari router library kamu
	assignmentId := r.PathValue("assignment_id")
	if assignmentId == "" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Assignment ID is required", nil)
		return
	}

	assignment, err := h.usecase.GetAssetAssigmentById(r.Context(), assignmentId)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusInternalServerError, "Failed to get assignment detail", err.Error())
		return
	}
	var assignmentResponseData assetAssignmentResponse

	//assignments
	assignmentResponseData.AssignmentId = assignment.AssignmentId
	assignmentResponseData.Notes = assignment.Notes
	assignmentResponseData.Corporation = assignment.Corporation
	assignmentResponseData.Status = string(assignment.Status)
	assignmentResponseData.AssignedDate = pkg.ParseFromDateToString(assignment.AssignedDate)
	if assignment.ReturnDate != nil {
		assignmentResponseData.ReturnDate = pkg.ParseFromDateToString(*assignment.ReturnDate)
	} else {
		assignmentResponseData.ReturnDate = ""
	}

	//user assigned
	assignmentResponseData.AssignedToId = assignment.UserId
	assignmentResponseData.AssignedToUsername = assignment.User.Username
	assignmentResponseData.AssignedToEmail = assignment.User.Email
	assignmentResponseData.AssignedToRole = assignment.User.Role
	//user whos assign
	assignmentResponseData.AssignedById = assignment.AssignedById
	assignmentResponseData.AssignedByUsername = assignment.AssignedByUsername

	//asset
	assignmentResponseData.AssetId = assignment.AssetId
	assignmentResponseData.AssetName = assignment.Asset.AssetName
	assignmentResponseData.Processor = assignment.Asset.Processor
	assignmentResponseData.Ram = assignment.Asset.Ram
	assignmentResponseData.Storage = assignment.Asset.Storage
	assignmentResponseData.SerialNumber = assignment.Asset.SerialNumber
	assignmentResponseData.QuantityStock = assignment.Asset.QuantityStock
	assignmentResponseData.PurchasedDate = pkg.ParseFromDateToString(assignment.Asset.PurchasedDate)
	//Brand
	assignmentResponseData.BrandId = assignment.Asset.BrandId
	assignmentResponseData.BrandName = assignment.Asset.Brand.BrandName

	//category
	assignmentResponseData.CategoryId = assignment.Asset.CategoryId
	assignmentResponseData.CategoryName = assignment.Asset.CategoryName

	//Department
	assignmentResponseData.DepartmentId = assignment.User.DepartmentId
	assignmentResponseData.DepartmentName = assignment.User.DepartmentName
	assignmentResponseData.CreatedAt = pkg.ParseFromDateToString(assignment.CreatedAt)
	assignmentResponseData.UpdatedAt = pkg.ParseFromDateToString(assignment.UpdatedAt)

	pkg.JSONResponse(w, http.StatusOK, "Success retrieve assignment detail!", assignmentResponseData, nil)
}

// 🚀 LANJUTAN KODE: CREATE ASSIGNMENT (DENGAN CONTEXT CLAIMS)
func (h *handler) CreateAssetAssignment(w http.ResponseWriter, r *http.Request) {
	var req assetAssignmentRequest

	// Decode json body
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid JSON payload", err.Error())
		return
	}

	// Validasi input minimal
	if req.AssetId == "" || req.UserId == "" || req.Notes == "" || req.AssignedDate == "" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid body request", "Invalid body request")
		return
	}

	if req.Status != "" {
		isValidStatus := req.Status == string(Assigned) ||
			req.Status == string(Damaged) ||
			req.Status == string(Returned) ||
			req.Status == string(Lost)

		if !isValidStatus {
			pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid status value!!", "Status must be assigned, damaged, returned, or lost")
			return
		}
	}
	assignedDateParsed, err := pkg.ParseFromStringToDate(req.AssignedDate)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid assigned date", nil)
	}
	var assignment AssetAssignment
	assignment.AssetId = req.AssetId
	assignment.Notes = req.Notes
	assignment.Corporation = req.Corporation
	assignment.Status = AssignmentStatus(req.Status)
	assignment.AssignedDate = assignedDateParsed
	assignment.UserId = req.UserId

	// Panggil usecase untuk mengeksekusi pembuatan data transaksi assignment
	err = h.usecase.CreateAssignment(r.Context(), assignment)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusInternalServerError, "Failed to create asset assignment", err.Error())
		return
	}

	pkg.JSONResponse(w, http.StatusCreated, "Asset assignment has been successfully recorded!", nil, nil)
}

func (h *handler) UpdateAssetAssignment(w http.ResponseWriter, r *http.Request) {
	assignmentId := r.PathValue("assignment_id")
	if assignmentId == "" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Assignment Id is not valid", nil)
		return
	}

	var assignmentReq assetAssignmentRequest
	err := json.NewDecoder(r.Body).Decode(&assignmentReq)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusBadRequest, err.Error(), err.Error())
		return

	}

	//validation if the field is empty
	if assignmentReq.Notes == "" || assignmentReq.AssetId == "" || assignmentReq.Status == "" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid body request", nil)
		return
	}

	var updatedData AssetAssignment
	updatedData.AssetId = assignmentReq.AssetId
	updatedData.Status = AssignmentStatus(assignmentReq.Status)
	updatedData.Notes = assignmentReq.Notes
	updatedData.UserId = assignmentReq.UserId
	updatedData.Corporation = assignmentReq.Corporation
	now := time.Now()
	updatedData.ReturnDate = &now
	err = h.usecase.UpdateAssignmentById(r.Context(), assignmentId, updatedData)
	if errors.Is(err, sql.ErrNoRows) {
		pkg.ErrorResponse(w, http.StatusNotFound, err.Error(), err)
		return

	}
	if err != nil {
		pkg.ErrorResponse(w, http.StatusInternalServerError, err.Error(), err)
		return
	}
	pkg.JSONResponse(w, http.StatusOK, "Asset assignment has been successfully updated!", nil, nil)

}

func (h *handler) GetAllAssignmentsData(w http.ResponseWriter, r *http.Request) {
	assignments, err := h.usecase.GetAllAssignmentsData(r.Context())
	if err != nil {
		pkg.ErrorResponse(w, http.StatusInternalServerError, err.Error(), err.Error())
		return
	}

	var assignmentsResponseData []assetAssignmentResponse
	for _, item := range assignments {
		var assignment assetAssignmentResponse
		//assignments
		assignment.AssignmentId = item.AssignmentId
		assignment.Notes = item.Notes
		assignment.Status = string(item.Status)
		assignment.AssignedDate = pkg.ParseFromDateToString(item.AssignedDate)
		if item.ReturnDate != nil {
			assignment.ReturnDate = pkg.ParseFromDateToString(*item.ReturnDate)
		} else {
			assignment.ReturnDate = ""
		}

		//user assigned
		assignment.AssignedToId = item.UserId
		assignment.AssignedToUsername = item.User.Username
		assignment.AssignedToEmail = item.User.Email
		assignment.AssignedToRole = item.User.Role
		assignment.Corporation = item.Corporation
		//user whos assign
		assignment.AssignedById = item.AssignedById
		assignment.AssignedByUsername = item.AssignedByUsername

		//asset
		assignment.AssetId = item.AssetId
		assignment.AssetName = item.Asset.AssetName
		assignment.SerialNumber = item.Asset.SerialNumber
		assignment.PurchasedDate = pkg.ParseFromDateToString(item.Asset.PurchasedDate)

		//Brand
		assignment.BrandId = item.Asset.BrandId
		assignment.BrandName = item.Asset.Brand.BrandName

		//category
		assignment.CategoryId = item.Asset.CategoryId
		assignment.CategoryName = item.Asset.Category.CategoryName

		//Department
		assignment.DepartmentId = item.User.Department.DepartmentId
		assignment.DepartmentName = item.User.Department.DepartmentName
		assignment.CreatedAt = pkg.ParseFromDateToString(item.CreatedAt)
		assignment.UpdatedAt = pkg.ParseFromDateToString(item.UpdatedAt)

		assignmentsResponseData = append(assignmentsResponseData, assignment)
	}
	pkg.JSONResponse(w, http.StatusOK, "Sucess retrieve data assignments", assignmentsResponseData, nil)
}

func (h *handler) UpdateAssetAssignmentStatus(w http.ResponseWriter, r *http.Request) {
	assignmentId := r.PathValue("assignment_id")
	if assignmentId == "" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Assignment ID is required", nil)
		return
	}

	var updateAssetAssignment struct {
		Status string `json:"status,omitempty"`
	}
	err := json.NewDecoder(r.Body).Decode(&updateAssetAssignment)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusBadRequest, err.Error(), err.Error())
		return
	}
	if updateAssetAssignment.Status != string(Returned) {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Invalid status request!", nil)
		return
	}

	err = h.usecase.UpdateAssignmentStatus(r.Context(), assignmentId, Returned)
	if errors.Is(err, sql.ErrNoRows) {
		pkg.ErrorResponse(w, http.StatusNotFound, err.Error(), err)
		return

	}
	if err != nil {
		pkg.ErrorResponse(w, http.StatusInternalServerError, err.Error(), err)
		return
	}
	pkg.JSONResponse(w, http.StatusOK, "Asset assignment status has been successfully updated!", nil, nil)
}

func (h *handler) GetAssignmentsByUserId(w http.ResponseWriter, r *http.Request) {
	// 1. Tangkap parameter status dari query URL (?status=...)
	var statusAssignment AssignmentStatus
	status := r.URL.Query().Get("status")
	switch status {
	case "assigned":
		statusAssignment = Assigned
	case "returned":
		statusAssignment = Returned
	case "damaged":
		statusAssignment = Damaged
	case "lost":
		statusAssignment = Lost
	default:
		statusAssignment = Assigned
	}
	// 2. Oper parameter status ke layer usecase (pastikan fungsi di usecase sudah menerima parameter ini)
	assignments, err := h.usecase.GetAssetAssignmentByUserId(r.Context(), statusAssignment)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusInternalServerError, err.Error(), err.Error())
		return
	}

	// Buat slice kosong bawaan agar jika data kosong, JSON di frontend menghasilkan [] bukan null
	assignmentsResponseData := []assetAssignmentResponse{}

	for _, item := range assignments {
		var assignment assetAssignmentResponse

		// assignments
		assignment.AssignmentId = item.AssignmentId
		assignment.Notes = item.Notes
		assignment.Status = string(item.Status)
		assignment.AssignedDate = pkg.ParseFromDateToString(item.AssignedDate)
		if item.ReturnDate != nil {
			assignment.ReturnDate = pkg.ParseFromDateToString(*item.ReturnDate)
		} else {
			assignment.ReturnDate = ""
		}

		// user assigned
		assignment.AssignedToId = item.UserId
		assignment.AssignedToUsername = item.User.Username
		assignment.AssignedToEmail = item.User.Email
		assignment.AssignedToRole = item.User.Role
		assignment.Corporation = item.Corporation

		// user who assigned
		assignment.AssignedById = item.AssignedById
		assignment.AssignedByUsername = item.AssignedByUsername

		// asset
		assignment.AssetId = item.AssetId
		assignment.AssetName = item.Asset.AssetName
		assignment.SerialNumber = item.Asset.SerialNumber
		assignment.PurchasedDate = pkg.ParseFromDateToString(item.Asset.PurchasedDate)

		// Brand
		assignment.BrandId = item.Asset.BrandId
		assignment.BrandName = item.Asset.Brand.BrandName

		// category
		assignment.CategoryId = item.Asset.CategoryId
		assignment.CategoryName = item.Asset.Category.CategoryName

		// Department
		assignment.DepartmentId = item.User.Department.DepartmentId
		assignment.DepartmentName = item.User.Department.DepartmentName
		assignment.CreatedAt = pkg.ParseFromDateToString(item.CreatedAt)
		assignment.UpdatedAt = pkg.ParseFromDateToString(item.UpdatedAt)

		assignmentsResponseData = append(assignmentsResponseData, assignment)
	}

	pkg.JSONResponse(w, http.StatusOK, "Success retrieve data assignments by user id!", assignmentsResponseData, nil)
}
