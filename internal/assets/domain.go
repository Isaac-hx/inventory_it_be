package assets

import (
	"inventory-it/internal/brands"
	"inventory-it/internal/categories"
	"time"
)

type AssetStatus string

const (
	Available   AssetStatus = "available"
	Maintenance AssetStatus = "maintenance"
	Retired     AssetStatus = "retired"
	Assigned    AssetStatus = "assigned"
)

type Asset struct {
	AssetId       string
	AssetName     string
	SerialNumber  string
	Description   string
	PurchasedDate time.Time
	QuantityStock int
	Status        AssetStatus
	BrandId       string
	Brand         brands.Brand
	CategoryId    string
	Category      categories.Categories
	CategoryName  string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type ChartDataAssetByCategories struct {
	CategoryName string `json:"category_name,omitempty"`
	TotalAsset   string `json:"total_asset,omitempty"`
}
