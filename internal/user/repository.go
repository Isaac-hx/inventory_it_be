package user

import (
	"context"
	"database/sql"
	"errors"
)

type UserFilter struct {
	Role    string
	Search  string
	Limit   int
	Page    int
	OrderBy string
}

// Create header method
type Repository interface {
	Create(ctx context.Context, u *User) error
	FindByEmail(ctx context.Context, email string) (User, bool, error)
	FindByUsername(ctx context.Context, username string) (User, bool, error)
	GetAllUsers(ctx context.Context, userFilter UserFilter) ([]User, error)
}

// Abstraction for databse object
type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, u *User) error {
	query := `INSERT INTO users(user_id,username,password,email,role) VALUES(?,?,?,?,?)`
	_, err := r.db.ExecContext(ctx, query, u.UserId, u.Username, u.Password, u.Email, u.Role)
	if err != nil {
		return err
	}

	return nil

}

func (r *repository) FindByEmail(ctx context.Context, email string) (User, bool, error) {
	query := `SELECT username,email,role,created_at,updated_at FROM users WHERE email = ?`
	var user User
	err := r.db.QueryRowContext(ctx, query, email).Scan(&user.Username, &user.Email, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, false, nil
		}
		return User{}, false, err
	}
	return user, true, nil
}

func (r *repository) FindByUsername(ctx context.Context, username string) (User, bool, error) {
	query := `SELECT username,password,email,role,created_at,updated_at FROM users WHERE username = ?`
	var user User
	err := r.db.QueryRowContext(ctx, query, username).Scan(&user.Username, &user.Password, &user.Email, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, false, nil
		}
		return User{}, false, err
	}
	return user, true, nil
}
func (r *repository) GetAllUsers(ctx context.Context, userFilter UserFilter) ([]User, error) {
	query := `
		SELECT user_id, username, email, role,  created_at, updated_at
		FROM users
	`

	args := []any{}

	if userFilter.Search != "" {
		query += ` AND (username LIKE ? OR email LIKE ?) `
		search := "%" + userFilter.Search + "%"
		args = append(args, search, search)
	}

	if userFilter.Role != "" {
		query += ` AND role = ? `
		args = append(args, userFilter.Role)
	}

	switch userFilter.OrderBy {
	case "username_asc":
		query += ` ORDER BY username ASC `
	case "username_desc":
		query += ` ORDER BY username DESC `
	case "created_at_asc":
		query += ` ORDER BY created_at ASC `
	default:
		query += ` ORDER BY created_at DESC `
	}

	if userFilter.Limit <= 0 {
		userFilter.Limit = 10
	}

	if userFilter.Page <= 0 {
		userFilter.Page = 1
	}

	offset := (userFilter.Page - 1) * userFilter.Limit

	query += ` LIMIT ? OFFSET ? `
	args = append(args, userFilter.Limit, offset)

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
