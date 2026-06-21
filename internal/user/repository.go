package user

import (
	"context"
	"database/sql"
	"inventory-it/internal/pkg"
	"math"
)

// Create header method
type Repository interface {
	GetAllUsers(ctx context.Context, filter UserFilter) ([]User, error)
	GetUserById(ctx context.Context, userId string) (User, error)
	DeleteUserById(ctx context.Context, userId string) error
	UpdateUserById(ctx context.Context, userId string, user User) error
	GetTotalPageAndTotalDataUsers(context.Context, UserFilter) (pkg.PaginationMeta, error)
	GetAllDataUsers(ctx context.Context) ([]User, error)
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
	query := `
	SELECT
		u.user_id,
		u.username,
		u.email,
		u.role,
		u.department_id,
		d.department_name,
		u.created_at,
		u.updated_at
	FROM users u
	LEFT JOIN departments d
		ON u.department_id = d.department_id
	WHERE u.user_id = ?`

	var user User
	err := r.db.QueryRowContext(ctx, query, userId).Scan(
		&user.UserId,
		&user.Username,
		&user.Email,
		&user.Role,
		&user.DepartmentId,
		&user.DepartmentName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (r *repository) DeleteUserById(ctx context.Context, userId string) error {
	query := `DELETE FROM users WHERE user_id = ?`
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

func (r *repository) GetTotalPageAndTotalDataUsers(ctx context.Context, filter UserFilter) (pkg.PaginationMeta, error) {
	query := `
		SELECT COUNT(*)
		FROM users u
		LEFT JOIN departments d
			ON u.department_id = d.department_id
		WHERE 1=1
	`

	args := []any{}
	var paginationData pkg.PaginationMeta

	if filter.Search != "" {
		query += `
			AND (
				u.username LIKE ?
				OR u.email LIKE ?
				OR d.department_name LIKE ?
			)
		`

		search := "%" + filter.Search + "%"
		args = append(args, search, search, search)
	}
	var totalData int

	err := r.db.QueryRowContext(
		ctx,
		query,
		args...,
	).Scan(&totalData)

	if err != nil {
		return paginationData, err
	}

	var totalPage int

	if filter.Limit > 0 {
		totalPage = int(math.Ceil(
			float64(totalData) / float64(filter.Limit),
		))
	}

	paginationData.Page = filter.Page
	paginationData.Limit = filter.Limit
	paginationData.TotalData = totalData
	paginationData.TotalPage = totalPage

	return paginationData, nil
}

func (r *repository) GetAllDataUsers(ctx context.Context) ([]User, error) {
	// 1. Hapus 'WHERE u.user_id = ?' agar mengambil semua data
	query := `
        SELECT
            u.user_id,
            u.username,
            u.email,
            u.role,
            u.department_id,
            COALESCE(d.department_name, '') AS department_name, -- Antisipasi jika NULL akibat LEFT JOIN
            u.created_at,
            u.updated_at
        FROM users u
        LEFT JOIN departments d
            ON u.department_id = d.department_id`

	// 2. Gunakan QueryContext (bukan QueryRowContext) untuk data banyak
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 3. Siapkan slice untuk menampung semua user
	var users []User

	// 4. Looping untuk membaca setiap baris data dari database
	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.UserId,
			&user.Username,
			&user.Email,
			&user.Role,
			&user.DepartmentId,
			&user.DepartmentName,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Masukkan data user ke dalam slice/list
		users = append(users, user)
	}

	// 5. Cek apakah ada error yang terjadi selama proses looping
	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Kembalikan semua data users
	return users, nil
}
