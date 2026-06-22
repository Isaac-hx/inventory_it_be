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
	GetAllAssetsData(context.Context) ([]Asset, error)
	GetAssetById(context.Context, string) (Asset, error)
	DeleteAssetById(context.Context, string) error
	UpdateAssetById(context.Context, string, Asset) error
	UpdateAssetStatusById(context.Context, *sql.Tx, string, AssetStatus) error
	GetTotalPageAndTotalDataAssets(context.Context, AssetFilter) (pkg.PaginationMeta, error)
	GetOverview(context.Context) (pkg.OverviewData, error)
	GetCountGroupCategoryAssets(context.Context) ([]ChartDataAssetByCategories, error)
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
			description,
			purchased_date,
			quantity_stock,
			status,
			brand_id,
			category_id

		)
		VALUES (?, ?, ?, ?,?,?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		asset.AssetId,
		asset.AssetName,
		asset.SerialNumber,
		asset.Description,
		asset.PurchasedDate,
		asset.QuantityStock,
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
            a.purchased_date,
            a.status,
            a.serial_number,
            a.quantity_stock,

            b.brand_id,

            c.category_id,

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

	// 1. PINDAHKAN FILTER STATUS KE SINI (Sebelum ORDER BY)
	// Berikan spasi di awal string agar aman
	if assetFilter.Status != "" {
		query += ` AND a.status = ? `
		args = append(args, assetFilter.Status)
	}

	// 2. KLAUSA ORDER BY SEKARANG BERADA DI TEMPAT YANG BENAR
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

	// 3. KLAUSA LIMIT & OFFSET DI AKHIR QUERY
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
			&asset.PurchasedDate,
			&asset.Status,
			&asset.SerialNumber,
			&asset.QuantityStock,
			&asset.BrandId,
			&asset.CategoryId,
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
			a.description,
			a.quantity_stock,

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
		&asset.Description,
		&asset.QuantityStock,
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
			quantity_stock = ?,
			status = ?,
			brand_id = ?,
			category_id = ?,
			description = ?
		WHERE asset_id = ?
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		asset.AssetName,
		asset.SerialNumber,
		asset.PurchasedDate,
		asset.QuantityStock,
		asset.Status,
		asset.BrandId,
		asset.CategoryId,
		asset.Description,
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

func (r *repository) GetTotalPageAndTotalDataAssets(
	ctx context.Context,
	filter AssetFilter,
) (pkg.PaginationMeta, error) {
	query := `
		SELECT COUNT(*)
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

	var totalData int

	err := r.db.QueryRowContext(ctx, query, args...).Scan(&totalData)
	if err != nil {
		return pkg.PaginationMeta{}, err
	}

	totalPage := 0
	if filter.Limit > 0 {
		totalPage = int(math.Ceil(float64(totalData) / float64(filter.Limit)))
	}

	paginationData := pkg.PaginationMeta{
		Page:      filter.Page,
		Limit:     filter.Limit,
		TotalData: totalData,
		TotalPage: totalPage,
	}

	return paginationData, nil
}

func (r *repository) GetAllAssetsData(ctx context.Context) ([]Asset, error) {
	// 1. Hapus klausa 'WHERE a.asset_id = ?' agar mengambil semua baris data
	query := `
        SELECT
            a.asset_id,
            a.asset_name,
            a.serial_number,
            a.purchased_date,
            a.status,
            a.description,
            a.quantity_stock,

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
            ON a.category_id = c.category_id`

	// 2. Gunakan QueryContext untuk mengambil data dalam jumlah banyak (multiple rows)
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 3. Inisialisasi slice kosong untuk menampung seluruh daftar asset
	var assets []Asset

	// 4. Looping untuk membaca data baris demi baris
	for rows.Next() {
		var asset Asset

		// Scan data ke struct utama dan juga ke nested struct (Brand & Category)
		err := rows.Scan(
			&asset.AssetId,
			&asset.AssetName,
			&asset.SerialNumber,
			&asset.PurchasedDate,
			&asset.Status,
			&asset.Description,
			&asset.QuantityStock,
			&asset.Brand.BrandId,
			&asset.Brand.BrandName,
			&asset.Category.CategoryId,
			&asset.Category.CategoryName,
			&asset.CreatedAt,
			&asset.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Masukkan objek asset yang berhasil di-scan ke dalam slice/array
		assets = append(assets, asset)
	}

	// 5. Pastikan tidak ada error tersembunyi setelah looping selesai
	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Kembalikan seluruh list asset
	return assets, nil
}

func (r *repository) GetOverview(
	ctx context.Context,
) (pkg.OverviewData, error) {

	query := `
	SELECT
		COUNT(*) AS total_asset,
		SUM(CASE WHEN status = 'assigned' THEN 1 ELSE 0 END) AS assigned_asset,
		SUM(CASE WHEN status = 'available' THEN 1 ELSE 0 END) AS available_asset,
		SUM(CASE WHEN status = 'retired' THEN 1 ELSE 0 END) AS damaged_asset
	FROM assets
	`

	var overview pkg.OverviewData

	err := r.db.QueryRowContext(ctx, query).Scan(
		&overview.TotalAsset,
		&overview.TotalAssetAssigned,
		&overview.TotalAssetAvailable,
		&overview.TotalAssetRetired,
	)

	if err != nil {
		return pkg.OverviewData{}, err
	}

	return overview, nil
}
func (r *repository) GetCountGroupCategoryAssets(ctx context.Context) ([]ChartDataAssetByCategories, error) {
	query := `
        SELECT 
            c.category_name, 
            COUNT(a.asset_id) AS total_asset
        FROM assets a
        INNER JOIN categories c 
            ON a.category_id = c.category_id
        GROUP BY c.category_name;
    `

	// 1. Gunakan QueryContext untuk mengambil multiple rows
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 2. Inisialisasi slice kosong agar tidak mengembalikan nilai nil ke json
	chartDataList := []ChartDataAssetByCategories{}

	// 3. Lakukan looping untuk membaca setiap baris kategori
	for rows.Next() {
		var item ChartDataAssetByCategories

		err := rows.Scan(&item.CategoryName, &item.TotalAsset)
		if err != nil {
			return nil, err
		}

		chartDataList = append(chartDataList, item)
	}

	// 4. Cek apakah ada error saat proses looping berjalan
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return chartDataList, nil
}
