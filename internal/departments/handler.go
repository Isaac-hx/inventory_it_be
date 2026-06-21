package departments

import (
	"database/sql"
	"encoding/json"
	"errors"
	"inventory-it/internal/pkg"
	"net/http"
	"strconv"
)

type DepartmentFilter struct {
	Search  string
	Limit   int
	Page    int
	OrderBy string
}
type DepartmentResponse struct {
	DepartmentId   string `json:"DepartmentId,omitempty"`
	DepartmentName string `json:"DepartmentName,omitempty"`
	CreatedAt      string `json:"CreatedAt,omitempty"`
	UpdatedAt      string `json:"UpdatedAt,omitempty"`
}
type DepartmentRequest struct {
	DepartmentName string `json:"department_name"`
}
type Handler interface {
	CreateDepartment(w http.ResponseWriter, r *http.Request)
	UpdateDepartmentNameById(w http.ResponseWriter, r *http.Request)
	DeleteDepartmentById(w http.ResponseWriter, r *http.Request)
	GetDepartmentById(w http.ResponseWriter, r *http.Request)
	GetAllDepartments(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	usecase Usecase
}

func NewHandler(u Usecase) Handler {
	return &handler{
		usecase: u,
	}
}

func (h *handler) CreateDepartment(w http.ResponseWriter, r *http.Request) {
	// Implementation for creating a department
	var departmentReq DepartmentRequest

	err := json.NewDecoder(r.Body).Decode(&departmentReq)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusBadRequest, err.Error(), err)
		return
	}
	if departmentReq.DepartmentName == "" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Department name is required!!", nil)
		return
	}
	createdDepartment, err := h.usecase.CreateDepartment(r.Context(), departmentReq.DepartmentName)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusInternalServerError, err.Error(), err)
		return
	}
	pkg.JSONResponse(w, http.StatusCreated, "Department created successfully!!", createdDepartment, nil)
}

func (h *handler) UpdateDepartmentNameById(w http.ResponseWriter, r *http.Request) {
	var departmentReq DepartmentRequest

	departmentId := r.PathValue("department_id")
	if departmentId == "" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Department ID is required!!", nil)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&departmentReq)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusBadRequest, err.Error(), err)
		return
	}
	if departmentReq.DepartmentName == "" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Department name is required!!", nil)
		return
	}

	err = h.usecase.UpdateDepartmentNameById(r.Context(), departmentId, departmentReq.DepartmentName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			pkg.ErrorResponse(w, http.StatusNotFound, "Department not found", nil)
			return
		}
		pkg.ErrorResponse(w, http.StatusInternalServerError, "Failed to update department name!!", err)
		return
	}

	pkg.JSONResponse(w, http.StatusOK, "Department updated sucessfully", nil, nil)
	// Implementation for updating a department name by its ID
}

func (h *handler) DeleteDepartmentById(w http.ResponseWriter, r *http.Request) {
	// Implementation for deleting a department by its ID
	departmentId := r.PathValue("department_id")

	if departmentId == "" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Department ID is required!!", nil)
		return
	}
	err := h.usecase.DeleteDepartmentById(r.Context(), departmentId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			pkg.ErrorResponse(w, http.StatusNotFound, "Department not found", nil)
			return
		}
		pkg.ErrorResponse(w, http.StatusInternalServerError, "Failed to delete department!!", err)
		return
	}
	pkg.JSONResponse(w, http.StatusOK, "Department deleted successfully!!", nil, nil)
}

func (h *handler) GetDepartmentById(w http.ResponseWriter, r *http.Request) {
	departmentId := r.PathValue("department_id")
	if departmentId == "" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Department ID is required", nil)
		return
	}
	department, err := h.usecase.GetDepartmentById(r.Context(), departmentId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			pkg.ErrorResponse(w, http.StatusNotFound, "Department not found", nil)
			return
		}
		pkg.ErrorResponse(w, http.StatusInternalServerError, "Failed to get department!!", err)
		return
	}
	var departmentResponseData DepartmentResponse
	departmentResponseData.DepartmentId = department.DepartmentId
	departmentResponseData.DepartmentName = department.DepartmentName
	departmentResponseData.CreatedAt = pkg.ParseFromDateToString(department.CreatedAt)
	departmentResponseData.UpdatedAt = pkg.ParseFromDateToString(department.UpdatedAt)
	pkg.JSONResponse(w, http.StatusOK, "Department retrieved successfully!!", departmentResponseData, nil)
}

func (h *handler) GetAllDepartments(w http.ResponseWriter, r *http.Request) {
	var departmentFilter DepartmentFilter
	//define context

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
	if limit <= 0 {
		departmentFilter.Limit = 10
	}

	if page <= 0 {
		departmentFilter.Page = 1
	}
	//structuring object departmentFilter
	departmentFilter.Search = search
	departmentFilter.OrderBy = orderBy

	// Implementation for retrieving all departments
	departments, meta, err := h.usecase.GetAllDepartments(r.Context(), departmentFilter)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusInternalServerError, "Failed to get departments!!", err)
		return
	}

	var departementsResponseData []DepartmentResponse
	for _, item := range departments {
		var department DepartmentResponse
		department.DepartmentId = item.DepartmentId
		department.DepartmentName = item.DepartmentName
		department.CreatedAt = pkg.ParseFromDateToString(item.CreatedAt)
		department.UpdatedAt = pkg.ParseFromDateToString(item.UpdatedAt)
		departementsResponseData = append(departementsResponseData, department)
	}

	pkg.JSONResponse(w, http.StatusOK, "Departments retrieved successfully!!", departementsResponseData, meta)
}
