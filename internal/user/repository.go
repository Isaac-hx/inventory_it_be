package user

import (
	"context"
	"database/sql"
)

// Create header method
type Repository interface {
	GetAllUsers(ctx context.Context, filter UserFilter) ([]User, error)
	GetUserById(ctx context.Context, userId string) (User, error)
	DeleteUserById(ctx context.Context, userId string) error
	UpdateUserById(ctx context.Context, userId string, user User) error
}

// Abstraction for databse object
type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetAllUsers(ctx context.Context, filter UserFilter) ([]User, error) {
	query := `
		SELECT 
		users.user_id, users.username, users.email, users.role,  users.created_at,users.department_id, users.updated_at,
		departments.department_name
		FROM users
		LEFT JOIN departments  ON users.department_id = departments.department_id
		WHERE 1=1
	`

	args := []any{}

	if filter.Search != "" {
		query += ` AND (users.username LIKE ? OR users.email LIKE ?) `
		search := "%" + filter.Search + "%"
		args = append(args, search, search)
	}

	if filter.Role != "" {
		query += ` AND users.role = ? `
		args = append(args, filter.Role)
	}

	switch filter.OrderBy {
	case "username_asc":
		query += ` ORDER BY users.username ASC `
	case "username_desc":
		query += ` ORDER BY users.username DESC `
	case "created_at_asc":
		query += ` ORDER BY users.created_at ASC `
	default:
		query += ` ORDER BY users.created_at DESC `
	}

	offset := (filter.Page - 1) * filter.Limit

	query += ` LIMIT ? OFFSET ? `
	args = append(args, filter.Limit, offset)

	//query join dengan department

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	userList := []User{}

	for rows.Next() {
		var user User

		err := rows.Scan(
			&user.UserId,
			&user.Username,
			&user.Email,
			&user.Role,
			&user.CreatedAt,
			&user.DepartmentId,
			&user.UpdatedAt,
			&user.DepartmentName,
		)
		if err != nil {
			return nil, err
		}

		userList = append(userList, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return userList, nil
}

func (r *repository) GetUserById(ctx context.Context, userId string) (User, error) {
	query := `SELECT user_id, username, email, role,  created_at,department_id, updated_at ,departments.department_name FROM users LEFT JOIN departments ON users.department_id = departments.department_id WHERE user_id = ?`
	var user User
	err := r.db.QueryRowContext(ctx, query, userId).Scan(
		&user.UserId,
		&user.Username,
		&user.Email,
		&user.Role,
		&user.CreatedAt,
		&user.DepartmentId,
		&user.UpdatedAt,
		&user.DepartmentName,
	)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (r *repository) DeleteUserById(ctx context.Context, userId string) error {
	query := `DELETE FROM user WHERE user_id = ?`
	result, err := r.db.ExecContext(ctx, query, userId)
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
func (r *repository) UpdateUserById(ctx context.Context, userId string, user User) error {
	query := `UPDATE users SET username = ?, email = ?, role = ?, department_id = ? WHERE user_id = ?`
	result, err := r.db.ExecContext(ctx, query, user.Username, user.Email, user.Role, user.DepartmentId, userId)
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
