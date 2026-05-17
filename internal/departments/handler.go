package departments

import (
	"encoding/json"
	"inventory-it/internal/pkg"
	"net/http"
)

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
	if departmentReq.DepartmentName == "" {
		pkg.ErrorResponse(w, http.StatusBadRequest, "Department name is required!!", nil)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&departmentReq)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	err = h.usecase.CreateDepartment(r.Context(), departmentReq.DepartmentName)
	if err != nil {
		pkg.ErrorResponse(w, http.StatusInternalServerError, err.Error(), err)
		return
	}
	pkg.JSONResponse(w, http.StatusCreated, "Department created successfully!!", nil)
}

func (h *handler) UpdateDepartmentNameById(w http.ResponseWriter, r *http.Request) {
	// Implementation for updating a department name by its ID
}

func (h *handler) DeleteDepartmentById(w http.ResponseWriter, r *http.Request) {
	// Implementation for deleting a department by its ID
}

func (h *handler) GetDepartmentById(w http.ResponseWriter, r *http.Request) {
	// Implementation for retrieving a department by its ID
}

func (h *handler) GetAllDepartments(w http.ResponseWriter, r *http.Request) {
	// Implementation for retrieving all departments
}
