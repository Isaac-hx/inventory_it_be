package maintenances

import (
	assetassignments "inventory-it/internal/asset_assignments"
	"inventory-it/internal/assets"
	"inventory-it/internal/brands"
	"inventory-it/internal/categories"
	"inventory-it/internal/user"
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
	MaintenanceAt time.Time
	CompletedAt   *time.Time
	Assignment    assetassignments.AssetAssignment
	Category      categories.Categories
	Brand         brands.Brand
	Asset         assets.Asset
	User          user.User // Tambahkan struct User untuk CreateRequest & GetAllMaintenancesByUserId
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
