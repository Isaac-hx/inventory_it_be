//This file is contain domain user entity

package user

import "time"

// Type object for users
type User struct {
	UserId    string
	Username  string
	Password  string
	Email     string
	Role      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
