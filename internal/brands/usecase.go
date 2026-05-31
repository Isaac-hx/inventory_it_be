package brands

import (
	"context"
	"fmt"
	"strings"
)

type Usecase interface {
	GetAllBrands(context.Context, BrandFilter) ([]Brands, error)
	GetBrandById(context.Context, string) (Brands, error)
	CreateBrand(context.Context, Brands) error
	UpdateBrand(context.Context, string, Brands) error
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

func (u *usecase) GetAllBrands(ctx context.Context, brandFilter BrandFilter) ([]Brands, error) {
	return u.repo.GetAllBrands(ctx, brandFilter)
}

func (u *usecase) GetBrandById(ctx context.Context, brandId string) (Brands, error) {
	return u.repo.GetBrandById(ctx, brandId)
}

func (u *usecase) CreateBrand(ctx context.Context, brand Brands) error {
	brand.BrandId = fmt.Sprintf("%v-%v", "BRAND", brand.BrandName)
	brand.BrandName = strings.ToTitle(brand.BrandName)
	return u.repo.CreateBrand(ctx, brand)
}

func (u *usecase) UpdateBrand(ctx context.Context, brandId string, brand Brands) error {
	brand.BrandName = strings.ToTitle(brand.BrandName)
	return u.repo.UpdateBrand(ctx, brandId, brand)
}

func (u *usecase) DeleteBrand(ctx context.Context, brandId string) error {
	return u.repo.DeleteBrand(ctx, brandId)
}
