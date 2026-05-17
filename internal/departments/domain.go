package departments

import "time"

type Department struct {
	DepartmentId   string
	DepartmentName string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
