package assets

import (
	"context"
	"database/sql"
	"inventory-it/internal/pkg"
	"math"
)

type Repository interface {
	CreateAsset(context.Context, Asset) error
	GetAllAssets(context.Context, AssetFilter) ([]Asset, error)
	GetAssetById(context.Context, string) (Asset, error)
	DeleteAssetById(context.Context, string) error
	UpdateAssetById(context.Context, string, Asset) error
	UpdateAssetStatusById(context.Context, *sql.Tx, string, AssetStatus) error
	GetTotalPageAndTotalDataAssets(context.Context, AssetFilter) (pkg.PaginationMeta, error)
}

type repository struct {
	db *sql.DB
}

func NewAssetRepository(db *sql.DB) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) CreateAsset(ctx context.Context, asset Asset) error {
	query := `
		INSERT INTO assets (
			asset_id,
			asset_name,
			serial_number,
			purchased_date,
			status,
			brand_id,
			category_id

		)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		asset.AssetId,
		asset.AssetName,
		asset.SerialNumber,
		asset.PurchasedDate,
		asset.Status,
		asset.BrandId,
		asset.CategoryId,
	)

	if err != nil {
		return err
	}

	return nil
}
func (r *repository) GetAllAssets(ctx context.Context, assetFilter AssetFilter) ([]Asset, error) {
	query := `
		SELECT
			a.asset_id,
			a.asset_name,
			a.serial_number,
			a.purchased_date,
			a.status,

			b.brand_id,
			b.brand_name,

			c.category_id,
			c.category_name,

			a.created_at,
			a.updated_at
		FROM assets a
		INNER JOIN brands b
			ON a.brand_id = b.brand_id
		INNER JOIN categories c
			ON a.category_id = c.category_id
		WHERE 1=1
	`

	args := []any{}

	if assetFilter.Search != "" {
		query += `
			AND (
				a.asset_name LIKE ?
				OR a.serial_number LIKE ?
				OR b.brand_name LIKE ?
				OR c.category_name LIKE ?
			)
		`

		search := "%" + assetFilter.Search + "%"
		args = append(args, search, search, search, search)
	}

	switch assetFilter.OrderBy {
	case "asc":
		query += ` ORDER BY c.category_name ASC `
	case "desc":
		query += ` ORDER BY c.category_name DESC `
	default:
		query += ` ORDER BY a.created_at DESC `
	}

	if assetFilter.Page <= 0 {
		assetFilter.Page = 1
	}

	if assetFilter.Limit <= 0 {
		assetFilter.Limit = 10
	}

	offset := (assetFilter.Page - 1) * assetFilter.Limit

	query += ` LIMIT ? OFFSET ? `
	args = append(args, assetFilter.Limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var listAssets []Asset

	for rows.Next() {
		var asset Asset

		err := rows.Scan(
			&asset.AssetId,
			&asset.AssetName,
			&asset.SerialNumber,
			&asset.PurchasedDate,
			&asset.Status,
			&asset.Brand.BrandId,
			&asset.Brand.BrandName,
			&asset.Category.CategoryId,
			&asset.Category.CategoryName,
			&asset.CategoryName,
			&asset.CreatedAt,
			&asset.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		listAssets = append(listAssets, asset)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return listAssets, nil
}
func (r *repository) GetAssetById(ctx context.Context, assetId string) (Asset, error) {
	query := `
		SELECT
			a.asset_id,
			a.asset_name,
			a.serial_number,
			a.purchased_date,
			a.status,

			b.brand_id,
			b.brand_name,

			c.category_id,
			c.category_name,

			a.created_at,
			a.updated_at
		FROM assets a
		INNER JOIN brands b
			ON a.brand_id = b.brand_id
		INNER JOIN categories c
			ON a.category_id = c.category_id
		WHERE a.asset_id = ?
	`

	var asset Asset

	err := r.db.QueryRowContext(ctx, query, assetId).Scan(
		&asset.AssetId,
		&asset.AssetName,
		&asset.SerialNumber,
		&asset.PurchasedDate,
		&asset.Status,
		&asset.Brand.BrandId,
		&asset.Brand.BrandName,
		&asset.Category.CategoryId,
		&asset.Category.CategoryName,
		&asset.CreatedAt,
		&asset.UpdatedAt,
	)

	if err != nil {
		return Asset{}, err
	}

	return asset, nil
}

func (r *repository) DeleteAssetById(ctx context.Context, assetId string) error {

	query := `
		DELETE FROM assets
		WHERE asset_id = ?
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		assetId,
	)

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

func (r *repository) UpdateAssetById(ctx context.Context, assetId string, asset Asset) error {
	query := `
		UPDATE assets
		SET
			asset_name = ?,
			serial_number = ?,
			purchased_date = ?,
			status = ?,
			brand_id = ?,
			category_id = ?
		WHERE asset_id = ?
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		asset.AssetName,
		asset.SerialNumber,
		asset.PurchasedDate,
		asset.Status,
		asset.BrandId,
		asset.CategoryId,
		assetId,
	)

	if err != nil {
		return err
	}

	return nil
}
func (r *repository) UpdateAssetStatusById(
	ctx context.Context, tx *sql.Tx, assetId string, status AssetStatus) error {
	query := `
		UPDATE assets
		SET status =  ?
		WHERE asset_id = ?
	`

	_, err := tx.ExecContext(
		ctx,
		query,
		status,
		assetId,
	)

	if err != nil {
		return err
	}

	return nil
}

func (r *repository) GetTotalPageAndTotalDataAssets(ctx context.Context, filter AssetFilter) (pkg.PaginationMeta, error) {
	query := `
		SELECT
			a.asset_id,
			a.asset_name,
			a.serial_number,
			a.purchased_date,
			a.status,

			b.brand_id,
			b.brand_name,

			c.category_id,
			c.category_name,

			a.created_at,
			a.updated_at
		FROM assets a
		INNER JOIN brands b
			ON a.brand_id = b.brand_id
		INNER JOIN categories c
			ON a.category_id = c.category_id
		WHERE 1=1
	`

	args := []any{}

	if filter.Search != "" {
		query += `
			AND (
				a.asset_name LIKE ?
				OR a.serial_number LIKE ?
				OR b.brand_name LIKE ?
				OR c.category_name LIKE ?
			)
		`

		search := "%" + filter.Search + "%"
		args = append(args, search, search, search, search)
	}

	var paginationData pkg.PaginationMeta

	var totalData int

	err := r.db.QueryRowContext(
		ctx,
		query,
		args...,
	).Scan(&totalData)

	if err != nil {
		return paginationData, err
	}

	var totalPage int

	if filter.Limit > 0 {
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
