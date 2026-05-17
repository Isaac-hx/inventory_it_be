package auth

import "time"

type User struct {
	UserId        string
	Username      string
	Password      string
	Email         string
	Department_id string
	Role          string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
