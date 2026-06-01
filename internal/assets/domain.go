package assets

import "time"

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
	PurchasedDate time.Time
	Status        AssetStatus
	BrandId       string
	CategoryId    string
	BrandName     string
	CategoryName  string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
