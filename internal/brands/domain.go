package brands

import "time"

type Brand struct {
	BrandId   string
	BrandName string
	CreatedAt time.Time
	UpdatedAt time.Time
}
