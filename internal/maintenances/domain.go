package maintenances

import (
	"inventory-it/internal/assets"
	"inventory-it/internal/brands"
	"inventory-it/internal/categories"
	"time"
)

type MaintenanceStatus string

const (
	Completed  MaintenanceStatus = "completed"
	Pending    MaintenanceStatus = "pending"
	InProgress MaintenanceStatus = "in_progress"
	Cancelled  MaintenanceStatus = "cancelled"
)

type Maintenance struct {
	MaintenanceId string
	Description   string
	Cost          int64
	Status        MaintenanceStatus
	AssetId       string
	MaintenanceAt time.Time
	CompletedAt   *time.Time
	Asset         assets.Asset
	Brand         brands.Brand
	Category      categories.Categories
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
