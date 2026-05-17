package departments

import (
	"context"
	"fmt"
)

type Usecase interface {
	CreateDepartment(ctx context.Context, departmentName string) error
	UpdateDepartmentNameById(ctx context.Context, departmentId string, departmentUpdateName string) error
	DeleteDepartmentById(ctx context.Context, departmentId string) error
	GetDepartmentById(ctx context.Context, departmentId string) (*Department, error)
	GetAllDepartments(ctx context.Context) ([]*Department, error)
}

type usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return &usecase{repo: repo}
}

func (u *usecase) CreateDepartment(ctx context.Context, departmentName string) error {
	var departmentDomain Department

	//Check if department name already exists
	_, err := u.repo.GetDepartmentByName(ctx, departmentName)
	if err == nil {
		return fmt.Errorf("Department with name %s already exists", departmentName)
	}

	//create department object
	departmentDomain.DepartmentId = fmt.Sprintf(`%v-%v`, "dept", departmentName)
	departmentDomain.DepartmentName = departmentName
	return u.repo.CreateDepartment(ctx, &departmentDomain)
}

func (u *usecase) UpdateDepartmentNameById(ctx context.Context, departmentId string, departmentUpdateName string) error {
	var departementUpdated Department
	departementUpdated.DepartmentName = departmentUpdateName
	return u.repo.UpdateDepartmentNameById(ctx, departmentId, &departementUpdated)
}

func (u *usecase) DeleteDepartmentById(ctx context.Context, departmentId string) error {
	return u.repo.DeleteDepartmentById(ctx, departmentId)
}

func (u *usecase) GetDepartmentById(ctx context.Context, departmentId string) (*Department, error) {
	return u.repo.GetDepartmentById(ctx, departmentId)
}

func (u *usecase) GetAllDepartments(ctx context.Context) ([]*Department, error) {
	return u.repo.GetAllDepartments(ctx)
}
