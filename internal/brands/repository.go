package brands

import (
	"context"
	"database/sql"
)

type Repository interface {
	GetAllBrands(context.Context, BrandFilter) ([]Brands, error)
	GetBrandById(ctx context.Context, brandId string) (Brands, error)
	CreateBrand(ctx context.Context, brand Brands) error
	UpdateBrand(ctx context.Context, brandId string, brand Brands) error
	DeleteBrand(ctx context.Context, brandId string) error
}

type repository struct {
	db *sql.DB
}

func NewBrandRepository(db *sql.DB) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) GetAllBrands(ctx context.Context, brandFilter BrandFilter) ([]Brands, error) {
	query := `SELECT brand_id, brand_name, created_at, updated_at FROM brands WHERE 1=1`
	args := []any{}

	if brandFilter.Search != "" {
		query += ` AND brand_name LIKE ? `
		search := "%" + brandFilter.Search + "%"
		args = append(args, search)
	}

	switch brandFilter.OrderBy {
	case "asc":
		query += ` ORDER BY brand_name ASC `
	case "desc":
		query += ` ORDER BY brand_name DESC `
	default:
		query += ` ORDER BY created_at DESC `
	}

	offset := (brandFilter.Page - 1) * brandFilter.Limit
	query += ` LIMIT ? OFFSET ? `
	args = append(args, brandFilter.Limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var brands []Brands
	for rows.Next() {
		var brand Brands
		err := rows.Scan(&brand.BrandId, &brand.BrandName, &brand.CreatedAt, &brand.UpdatedAt)
		if err != nil {
			return nil, err
		}
		brands = append(brands, brand)
	}

	return brands, nil
}

func (r *repository) GetBrandById(ctx context.Context, brandId string) (Brands, error) {
	query := `SELECT brand_id, brand_name, created_at, updated_at FROM brands WHERE brand_id = ?`
	var brand Brands
	err := r.db.QueryRowContext(ctx, query, brandId).Scan(&brand.BrandId, &brand.BrandName, &brand.CreatedAt, &brand.UpdatedAt)
	if err != nil {
		return Brands{}, err
	}
	return brand, nil
}

func (r *repository) CreateBrand(ctx context.Context, brand Brands) error {
	query := `INSERT INTO brands (brand_id, brand_name, created_at, updated_at) VALUES (?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, brand.BrandId, brand.BrandName, brand.CreatedAt, brand.UpdatedAt)
	return err
}

func (r *repository) UpdateBrand(ctx context.Context, brandId string, brand Brands) error {
	query := `UPDATE brands SET brand_name = ? WHERE brand_id = ?`
	result, err := r.db.ExecContext(ctx, query, brand.BrandName, brandId)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *repository) DeleteBrand(ctx context.Context, brandId string) error {
	query := `DELETE FROM brands WHERE brand_id = ?`
	result, err := r.db.ExecContext(ctx, query, brandId)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
