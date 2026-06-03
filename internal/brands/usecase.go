package brands

import (
	"context"
	"fmt"
	"inventory-it/internal/pkg"
	"strings"
)

type Usecase interface {
	GetAllBrands(context.Context, BrandFilter) ([]Brand, pkg.PaginationMeta, error)
	GetBrandById(context.Context, string) (Brand, error)
	CreateBrand(context.Context, Brand) error
	UpdateBrand(context.Context, string, Brand) error
	DeleteBrand(context.Context, string) error
}

type usecase struct {
	repo Repository
}

func NewBrandUsecase(r Repository) Usecase {
	return &usecase{
		repo: r,
	}
}

func (u *usecase) GetAllBrands(ctx context.Context, filter BrandFilter) ([]Brand, pkg.PaginationMeta, error) {
	brands, err := u.repo.GetAllBrands(ctx, filter)
	if err != nil {
		return nil, pkg.PaginationMeta{}, nil
	}
	meta, err := u.repo.GetTotalPageAndTotalDataBrands(ctx, filter)
	if err != nil {
		return nil, pkg.PaginationMeta{}, nil

	}
	return brands, meta, nil
}

func (u *usecase) GetBrandById(ctx context.Context, brandId string) (Brand, error) {
	return u.repo.GetBrandById(ctx, brandId)
}

func (u *usecase) CreateBrand(ctx context.Context, brand Brand) error {
	brand.BrandId = fmt.Sprintf("%v-%v", "BRAND", strings.ToUpper(brand.BrandName))
	brand.BrandName = strings.ToTitle(brand.BrandName)
	return u.repo.CreateBrand(ctx, brand)
}

func (u *usecase) UpdateBrand(ctx context.Context, brandId string, brand Brand) error {
	brand.BrandName = strings.ToTitle(brand.BrandName)
	return u.repo.UpdateBrand(ctx, brandId, brand)
}

func (u *usecase) DeleteBrand(ctx context.Context, brandId string) error {
	return u.repo.DeleteBrand(ctx, brandId)
}
