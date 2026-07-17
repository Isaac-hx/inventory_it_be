package assetassignments

import (
	"inventory-it/internal/assets"
	"inventory-it/internal/user"

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
	AssignmentId       string
	AssetId            string
	Corporation        string
	Asset              assets.Asset
	UserId             string //user who is use to the asset
	User               user.User
	AssignedById       string //user who assigned the asset
	AssignedByUsername string
	Status             AssignmentStatus
	Notes              string
	AssignedDate       time.Time
	ReturnDate         *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
