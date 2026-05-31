package assets

import "time"

type Asset struct {
	AssetId       string
	AssetName     string
	SerialNumber  string
	PurchasedDate time.Time
	Status        string
	BrandId       string
	CategoryId    string
	BrandName     string
	CategoryName  string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
