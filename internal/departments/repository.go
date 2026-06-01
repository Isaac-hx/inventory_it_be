package departments

import (
	"context"
	"database/sql"
)

type Repository interface {
	CreateDepartment(ctx context.Context, department Department) error
	UpdateDepartmentNameById(ctx context.Context, departmentId string, departmentUpdated Department) error
	DeleteDepartmentById(ctx context.Context, departmentId string) error
	GetDepartmentById(ctx context.Context, departmentId string) (Department, error)
	GetAllDepartments(ctx context.Context, filter DepartmentFilter) ([]Department, error)
	GetDepartmentByName(ctx context.Context, departmentName string) (Department, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateDepartment(ctx context.Context, department Department) error {
	// Implementation for creating a department in the database
	_, err := r.db.ExecContext(ctx, `INSERT INTO departments (department_id, department_name) VALUES (?, ?)`, department.DepartmentId, department.DepartmentName)
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) UpdateDepartmentNameById(ctx context.Context, departmentId string, departmentUpdated Department) error {
	// Implementation for updating a department name by its ID in the database
	result, err := r.db.ExecContext(ctx, `UPDATE departments SET department_name = ? WHERE department_id = ?`, departmentUpdated.DepartmentName, departmentId)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *repository) DeleteDepartmentById(ctx context.Context, departmentId string) error {
	// Implementation for deleting a department by its ID from the database
	result, err := r.db.ExecContext(ctx, `DELETE FROM departments WHERE department_id = ?`, departmentId)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *repository) GetDepartmentById(ctx context.Context, departmentId string) (Department, error) {
	// Implementation for retrieving a department by its ID from the database
	row := r.db.QueryRowContext(ctx, `SELECT department_id, department_name, created_at, updated_at FROM departments WHERE department_id = ?`, departmentId)
	department := Department{}
	err := row.Scan(&department.DepartmentId, &department.DepartmentName, &department.CreatedAt, &department.UpdatedAt)
	if err != nil {
		return Department{}, err
	}
	return department, nil
}

func (r *repository) GetAllDepartments(ctx context.Context, filter DepartmentFilter) ([]Department, error) {

	query := `
		SELECT 
			department_id, 
			department_name, 
			created_at, 
			updated_at
		FROM departments
		WHERE 1=1
	`

	args := []any{}

	if filter.Search != "" {
		query += ` AND department_name LIKE ? `
		search := "%" + filter.Search + "%"
		args = append(args, search)
	}

	switch filter.OrderBy {
	case "department_name_asc":
		query += ` ORDER BY department_name ASC `
	case "department_name_desc":
		query += ` ORDER BY department_name DESC `
	case "created_at_asc":
		query += ` ORDER BY created_at ASC `
	default:
		query += ` ORDER BY created_at DESC `
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}

	if filter.Limit <= 0 {
		filter.Limit = 10
	}

	offset := (filter.Page - 1) * filter.Limit

	query += ` LIMIT ? OFFSET ? `
	args = append(args, filter.Limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	departmentList := []Department{}

	for rows.Next() {
		var department Department

		err := rows.Scan(
			&department.DepartmentId,
			&department.DepartmentName,
			&department.CreatedAt,
			&department.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		departmentList = append(departmentList, department)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return departmentList, nil
}

func (r *repository) GetDepartmentByName(ctx context.Context, departmentName string) (Department, error) {
	row := r.db.QueryRowContext(ctx, `SELECT department_id, department_name, created_at, updated_at FROM departments WHERE department_name = ?`, departmentName)
	department := Department{}
	err := row.Scan(&department.DepartmentId, &department.DepartmentName, &department.CreatedAt, &department.UpdatedAt)
	if err != nil {
		return Department{}, err
	}
	return department, nil
}
