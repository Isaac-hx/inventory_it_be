package user

import (
	"context"
	"database/sql"
	"errors"
)

// Create header method
type Repository interface {
	Create(ctx context.Context, u *User) error
	FindByEmail(ctx context.Context, email string) (User, bool, error)
	FindByUsername(ctx context.Context, username string) (User, bool, error)
}

// Abstraction for databse object
type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db}
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
