package departments

import (
	"context"
	"fmt"
	"inventory-it/internal/pkg"
	"strings"
)

type Usecase interface {
	CreateDepartment(ctx context.Context, departmentName string) (Department, error)
	UpdateDepartmentNameById(ctx context.Context, departmentId string, departmentUpdateName string) error
	DeleteDepartmentById(ctx context.Context, departmentId string) error
	GetDepartmentById(ctx context.Context, departmentId string) (Department, error)
	GetAllDepartments(ctx context.Context, filter DepartmentFilter) ([]Department, pkg.PaginationMeta, error)
}

type usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return &usecase{repo: repo}
}

func (u *usecase) CreateDepartment(ctx context.Context, departmentName string) (Department, error) {
	var departmentData Department

	//create department object
	departmentData.DepartmentId = fmt.Sprintf(`%v-%v`, "DEPT", strings.ToUpper(departmentName))

	//edit department name to uppercase
	departmentData.DepartmentName = strings.ToUpper(departmentName)
	return departmentData, u.repo.CreateDepartment(ctx, departmentData)
}

func (u *usecase) UpdateDepartmentNameById(ctx context.Context, departmentId string, departmentUpdateName string) error {
	var departementUpdated Department

	//edit updatre department to uppercase
	departementUpdated.DepartmentName = strings.ToUpper(departmentUpdateName)

	//upcase department ID
	departmentId = strings.ToUpper(departmentId)
	return u.repo.UpdateDepartmentNameById(ctx, departmentId, departementUpdated)
}

// masih belum handle kalau department id tidak ditemukan, masih return success meskipun id tidak ditemukan, harusnya kalau id tidak ditemukan return error not found

func (u *usecase) DeleteDepartmentById(ctx context.Context, departmentId string) error {
	departmentId = strings.ToUpper(departmentId)
	_, err := u.repo.GetDepartmentById(ctx, departmentId)
	if err != nil {
		return err
	}
	return u.repo.DeleteDepartmentById(ctx, departmentId)
}

func (u *usecase) GetDepartmentById(ctx context.Context, departmentId string) (Department, error) {
	departmentId = strings.ToUpper(departmentId)
	return u.repo.GetDepartmentById(ctx, departmentId)
}

func (u *usecase) GetAllDepartments(ctx context.Context, filter DepartmentFilter) ([]Department, pkg.PaginationMeta, error) {
	// 1. Set default value di level Usecase agar konsisten untuk semua panggilan repo
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}

	// 2. Ambil data departments
	departments, err := u.repo.GetAllDepartments(ctx, filter)
	if err != nil {
		return nil, pkg.PaginationMeta{}, err
	}

	// 3. Ambil metadata pagination
	meta, err := u.repo.GetTotalPageAndTotalDataDepartments(ctx, filter)
	if err != nil {
		return nil, pkg.PaginationMeta{}, err
	}

	return departments, meta, nil
}
