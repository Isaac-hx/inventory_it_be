package assets

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

type Usecase interface {
	CreateAsset(ctx context.Context, asset Asset) error
	GetAllAssets(ctx context.Context, filter AssetFilter) ([]Asset, error)
	GetAssetById(ctx context.Context, assetId string) (Asset, error)
	UpdateAssetById(ctx context.Context, assetId string, asset Asset) error
	DeleteAssetById(ctx context.Context, assetId string) error
}

type usecase struct {
	repo Repository
}

func NewAssetUsecase(r Repository) Usecase {
	return &usecase{
		repo: r,
	}
}

func (u *usecase) GetAllAssets(ctx context.Context, filter AssetFilter) ([]Asset, error) {
	return u.repo.GetAllAssets(ctx, filter)
}

func (u *usecase) GetAssetById(ctx context.Context, assetId string) (Asset, error) {
	return u.repo.GetAssetById(ctx, assetId)
}

func (u *usecase) CreateAsset(ctx context.Context, asset Asset) error {
	asset.AssetId = uuid.New().String()
	asset.AssetName = strings.ToTitle(asset.AssetName)
	asset.SerialNumber = strings.ToUpper(asset.SerialNumber)
	return u.repo.CreateAsset(ctx, asset)
}

func (u *usecase) UpdateAssetById(ctx context.Context, assetId string, asset Asset) error {
	asset.AssetName = strings.ToTitle(asset.AssetName)
	asset.AssetName = strings.ToTitle(asset.AssetName)
	asset.SerialNumber = strings.ToUpper(asset.SerialNumber)
	return u.repo.UpdateAssetById(ctx, assetId, asset)
}

func (u *usecase) DeleteAssetById(ctx context.Context, assetId string) error {
	return u.repo.DeleteAssetById(ctx, assetId)
}
