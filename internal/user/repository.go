package user

import (
	"context"
	"database/sql"
)

// Create header method
type Repository interface {
	GetAllUsers(ctx context.Context, filter UserFilter) ([]User, error)
	GetUserById(ctx context.Context, userId string) (User, error)
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
		SELECT user_id, username, email, role,  created_at,department_id, updated_at
		FROM users WHERE 1=1
	`

	args := []any{}

	if filter.Search != "" {
		query += ` AND (username LIKE ? OR email LIKE ?) `
		search := "%" + filter.Search + "%"
		args = append(args, search, search)
	}

	if filter.Role != "" {
		query += ` AND role = ? `
		args = append(args, filter.Role)
	}

	switch filter.OrderBy {
	case "username_asc":
		query += ` ORDER BY username ASC `
	case "username_desc":
		query += ` ORDER BY username DESC `
	case "created_at_asc":
		query += ` ORDER BY created_at ASC `
	default:
		query += ` ORDER BY created_at DESC `
	}

	offset := (filter.Page - 1) * filter.Limit

	query += ` LIMIT ? OFFSET ? `
	args = append(args, filter.Limit, offset)

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
	query := `SELECT user_id, username, email, role,  created_at,department_id, updated_at FROM users WHERE user_id = ?`
	var user User
	err := r.db.QueryRowContext(ctx, query, userId).Scan(
		&user.UserId,
		&user.Username,
		&user.Email,
		&user.Role,
		&user.CreatedAt,
		&user.DepartmentId,
		&user.UpdatedAt,
	)
	if err != nil {
		return User{}, err
	}
	return user, nil
}
