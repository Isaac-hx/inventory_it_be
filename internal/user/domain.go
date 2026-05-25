package user

import (
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
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
