package categories

import (
	"context"
	"fmt"
	"inventory-it/internal/pkg"
	"strings"
)

type Usecase interface {
	GetAllCategories(context.Context, CategoryFilter) ([]Categories, pkg.PaginationMeta, error)
	GetCategoryById(context.Context, string) (Categories, error)
	CreateCategory(context.Context, Categories) error
	UpdateCategory(context.Context, string, Categories) error
	DeleteCategory(context.Context, string) error
}

type usecase struct {
	repo Repository
}

func NewCategoryUsecase(r Repository) Usecase {
	return &usecase{
		repo: r,
	}
}

func (u *usecase) GetAllCategories(ctx context.Context, filter CategoryFilter) ([]Categories, pkg.PaginationMeta, error) {
	categories, err := u.repo.GetAllCategories(ctx, filter)
	if err != nil {
		return nil, pkg.PaginationMeta{}, err
	}
	meta, err := u.repo.GetTotalPageAndTotalDataCategories(ctx, filter)
	if err != nil {
		return nil, pkg.PaginationMeta{}, err
	}
	return categories, meta, nil
}

func (u *usecase) GetCategoryById(ctx context.Context, categoryId string) (Categories, error) {
	return u.repo.GetCategoryById(ctx, categoryId)
}

func (u *usecase) CreateCategory(ctx context.Context, category Categories) error {
	category.CategoryId = fmt.Sprintf("%v-%v", "CAT", strings.ToUpper(category.CategoryName))
	category.CategoryName = strings.ToTitle(category.CategoryName)

	return u.repo.CreateCategory(ctx, category)
}

func (u *usecase) UpdateCategory(ctx context.Context, categoryId string, category Categories) error {
	category.CategoryName = strings.ToTitle(category.CategoryName)
	return u.repo.UpdateCategory(ctx, categoryId, category)
}

func (u *usecase) DeleteCategory(ctx context.Context, categoryId string) error {
	return u.repo.DeleteCategory(ctx, categoryId)
}
