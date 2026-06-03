package user

import (
	"context"
	"inventory-it/internal/pkg"
	"strings"
)

type Usecase interface {
	GetAllUsers(ctx context.Context, filter UserFilter) ([]User, pkg.PaginationMeta, error)
	GetUserById(ctx context.Context, userId string) (User, error)
	DeleteUserById(ctx context.Context, userId string) error
	UpdateUserById(ctx context.Context, userId string, user User) error
}

type usecase struct {
	repo Repository
}

func NewUsecase(r Repository) Usecase {
	return &usecase{
		repo: r,
	}
}

func (u *usecase) GetAllUsers(
	ctx context.Context,
	filter UserFilter,
) ([]User, pkg.PaginationMeta, error) {
	users, err := u.repo.GetAllUsers(ctx, filter)
	if err != nil {
		return nil, pkg.PaginationMeta{}, err
	}
	meta, err := u.repo.GetTotalPageAndTotalDataUsers(ctx, filter)
	if err != nil {
		return nil, pkg.PaginationMeta{}, err
	}
	return users, meta, nil
}

func (u *usecase) GetUserById(ctx context.Context, userId string) (User, error) {
	return u.repo.GetUserById(ctx, userId)
}

func (u *usecase) DeleteUserById(ctx context.Context, userId string) error {
	return u.repo.DeleteUserById(ctx, userId)
}

func (u *usecase) UpdateUserById(ctx context.Context, userId string, user User) error {
	var userUpdate User

	userUpdate.Username = strings.ToLower(user.Username)
	userUpdate.Email = user.Email
	userUpdate.Role = user.Role
	userUpdate.DepartmentId = user.DepartmentId
	return u.repo.UpdateUserById(ctx, userId, userUpdate)
}
