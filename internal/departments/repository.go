package departments

import (
	"context"
	"database/sql"
)

type Repository interface {
	CreateDepartment(ctx context.Context, department *Department) error
	UpdateDepartmentNameById(ctx context.Context, departmentId string, departmentUpdated *Department) error
	DeleteDepartmentById(ctx context.Context, departmentId string) error
	GetDepartmentById(ctx context.Context, departmentId string) (*Department, error)
	GetAllDepartments(ctx context.Context) ([]*Department, error)
	GetDepartmentByName(ctx context.Context, departmentName string) (*Department, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateDepartment(ctx context.Context, department *Department) error {
	// Implementation for creating a department in the database
	_, err := r.db.ExecContext(ctx, `INSERT INTO departments (department_id, department_name) VALUES (?, ?)`, department.DepartmentId, department.DepartmentName)
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) UpdateDepartmentNameById(ctx context.Context, departmentId string, departmentUpdated *Department) error {
	// Implementation for updating a department name by its ID in the database
	_, err := r.db.ExecContext(ctx, `UPDATE departments SET department_name = ? WHERE department_id = ?`, departmentUpdated.DepartmentName, departmentId)
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) DeleteDepartmentById(ctx context.Context, departmentId string) error {
	// Implementation for deleting a department by its ID from the database
	_, err := r.db.ExecContext(ctx, `DELETE FROM departments WHERE department_id = ?`, departmentId)
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) GetDepartmentById(ctx context.Context, departmentId string) (*Department, error) {
	// Implementation for retrieving a department by its ID from the database
	row := r.db.QueryRowContext(ctx, `SELECT department_id, department_name, created_at, updated_at FROM departments WHERE department_id = ?`, departmentId)
	department := &Department{}
	err := row.Scan(&department.DepartmentId, &department.DepartmentName, &department.CreatedAt, &department.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return department, nil
}

func (r *repository) GetAllDepartments(ctx context.Context) ([]*Department, error) {
	// Implementation for retrieving all departments from the database
	rows, err := r.db.QueryContext(ctx, `SELECT department_id, department_name, created_at, updated_at FROM departments`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var departments []*Department
	for rows.Next() {
		department := &Department{}
		err := rows.Scan(&department.DepartmentId, &department.DepartmentName, &department.CreatedAt, &department.UpdatedAt)
		if err != nil {
			return nil, err
		}
		departments = append(departments, department)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return departments, nil
}

func (r *repository) GetDepartmentByName(ctx context.Context, departmentName string) (*Department, error) {
	row := r.db.QueryRowContext(ctx, `SELECT department_id, department_name, created_at, updated_at FROM departments WHERE department_name = ?`, departmentName)
	department := &Department{}
	err := row.Scan(&department.DepartmentId, &department.DepartmentName, &department.CreatedAt, &department.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return department, nil
}
