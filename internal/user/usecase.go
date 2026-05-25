package user

import (
	"context"
)

type Usecase interface {
	GetAllUsers(ctx context.Context, filter UserFilter) ([]User, error)
	GetUserById(ctx context.Context, userId string) (User, error)
	DeleteUserById(ctx context.Context, userId string) error
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
) ([]User, error) {
	if filter.Limit <= 0 {
		filter.Limit = 10
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}

	return u.repo.GetAllUsers(ctx, filter)
}

func (u *usecase) GetUserById(ctx context.Context, userId string) (User, error) {
	return u.repo.GetUserById(ctx, userId)
}

func (u *usecase) DeleteUserById(ctx context.Context, userId string) error {
	return u.repo.DeleteUserById(ctx, userId)
}
