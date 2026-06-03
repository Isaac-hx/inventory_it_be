package assetassignments

import (
	"inventory-it/internal/assets"
	"inventory-it/internal/auth"
	"time"
)

type AssignmentStatus string

const (
	Assigned AssignmentStatus = "assigned"
	Returned AssignmentStatus = "returned"
	Damaged  AssignmentStatus = "damaged"
	Lost     AssignmentStatus = "lost"
)

type AssetAssignment struct {
	AssignmentId   string
	AssetId        string
	Asset          assets.Asset
	UserId         string //user who is use to the asset
	User           auth.User
	AssignedBy     string //user who assigned the asset
	AssignedByUser string
	Status         AssignmentStatus
	Notes          string
	AssignedDate   string
	ReturnDate     *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
