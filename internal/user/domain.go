package user

import (
	"inventory-it/internal/departments"
	"time"
)

type User struct {
	UserId         string
	Username       string
	Password       string
	Email          string
	DepartmentId   string
	Role           string
	DepartmentName string
	Department     departments.Department
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
