package brands

import (
	"context"
	"database/sql"
	"inventory-it/internal/pkg"
	"math"
)

type Repository interface {
	GetAllBrands(context.Context, BrandFilter) ([]Brand, error)
	GetBrandById(context.Context, string) (Brand, error)
	CreateBrand(context.Context, Brand) error
	UpdateBrand(context.Context, string, Brand) error
	DeleteBrand(context.Context, string) error
	GetTotalPageAndTotalDataBrands(context.Context, BrandFilter) (pkg.PaginationMeta, error)
}

type repository struct {
	db *sql.DB
}

func NewBrandRepository(db *sql.DB) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) GetAllBrands(ctx context.Context, brandFilter BrandFilter) ([]Brand, error) {
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

	var BrandList []Brand
	for rows.Next() {
		var brand Brand
		err := rows.Scan(&brand.BrandId, &brand.BrandName, &brand.CreatedAt, &brand.UpdatedAt)
		if err != nil {
			return nil, err
		}
		BrandList = append(BrandList, brand)
	}
	return BrandList, nil
}

func (r *repository) GetBrandById(ctx context.Context, brandId string) (Brand, error) {
	query := `SELECT brand_id, brand_name, created_at, updated_at FROM brands WHERE brand_id = ?`
	var brand Brand
	err := r.db.QueryRowContext(ctx, query, brandId).Scan(&brand.BrandId, &brand.BrandName, &brand.CreatedAt, &brand.UpdatedAt)
	if err != nil {
		return Brand{}, err
	}
	return brand, nil
}

func (r *repository) CreateBrand(ctx context.Context, brand Brand) error {
	query := `INSERT INTO brands (brand_id, brand_name) VALUES (?, ?)`
	_, err := r.db.ExecContext(ctx, query, brand.BrandId, brand.BrandName)
	return err
}

func (r *repository) UpdateBrand(ctx context.Context, brandId string, brand Brand) error {
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

func (r *repository) GetTotalPageAndTotalDataBrands(ctx context.Context, filter BrandFilter) (pkg.PaginationMeta, error) {
	// 👇 1. Tambahkan WHERE 1=1 agar struktur SQL valid saat digabung dengan AND
	query := `
        SELECT COUNT(*)
        FROM brands
        WHERE 1=1
    `

	args := []any{}
	var paginationData pkg.PaginationMeta

	if filter.Search != "" {
		query += `
            AND (
                brand_name LIKE ?
            )
        `

		search := "%" + filter.Search + "%"
		args = append(args, search)
	}

	var totalData int

	// 2. Jalankan query row
	err := r.db.QueryRowContext(
		ctx,
		query,
		args...,
	).Scan(&totalData)

	if err != nil {
		// Jika ada error di sini, sekarang usecase yang baru akan meneruskannya ke handler
		return paginationData, err
	}

	var totalPage int

	if filter.Limit > 0 {
		// Gunakan math.Ceil untuk menghitung total halaman
		totalPage = int(math.Ceil(
			float64(totalData) / float64(filter.Limit),
		))
	}

	paginationData.Page = filter.Page
	paginationData.Limit = filter.Limit
	paginationData.TotalData = totalData
	paginationData.TotalPage = totalPage

	return paginationData, nil
}
