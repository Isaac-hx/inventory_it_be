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
)

type Asset struct {
	AssetId       string
	AssetName     string
	SerialNumber  string
	Description   string
	PurchasedDate time.Time
	Status        AssetStatus
	BrandId       string
	Brand         brands.Brand
	CategoryId    string
	Category      categories.Categories
	CategoryName  string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
