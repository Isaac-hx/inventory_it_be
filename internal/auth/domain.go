package auth

import (
	"inventory-it/internal/departments"
	"time"
)

type User struct {
	UserId        string
	Username      string
	Password      string
	Email         string
	Department_id string
	Department    departments.Department
	Role          string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
