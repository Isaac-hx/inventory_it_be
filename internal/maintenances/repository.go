package maintenances

import (
	"context"
	"database/sql"
	"inventory-it/internal/pkg"
	"math"
)

type Repository interface {
	GetAllMaintenances(context.Context, MaintenanceFilter) ([]Maintenance, error)
	GetMaintenanceById(context.Context, string) (Maintenance, error)
	CreateMaintenanceTx(context.Context, *sql.Tx, Maintenance) error
	UpdateMaintenanceStatusTx(context.Context, *sql.Tx, string, Maintenance) error
	UpdateMaintenanceCostTx(context.Context, *sql.Tx, string, int64) error
	UpdateMaintenanceDescriptionTx(context.Context, *sql.Tx, string, string) error
	GetTotalPageAndTotalDataMaintenances(context.Context, MaintenanceFilter) (pkg.PaginationMeta, error)
	UpdateMaintenanceTx(context.Context, *sql.Tx, string, Maintenance) error
	GetAllMaintenancesData(context.Context) ([]Maintenance, error)
}

type MaintenanceRepository struct {
	db *sql.DB
}

func NewMaintenanceRepository(db *sql.DB) Repository {
	return &MaintenanceRepository{db: db}
}

func (r *MaintenanceRepository) GetAllMaintenances(ctx context.Context, maintenanceFilter MaintenanceFilter) ([]Maintenance, error) {
	query := `
		SELECT 
			m.maintenance_id,
			m.description,
			m.cost,
			m.status,
			m.asset_id,
			m.maintenance_at,
			m.completed_at,

			a.asset_id,
			a.asset_name,


			b.brand_id,
			b.brand_name,
			
			c.category_id,
			c.category_name
		FROM maintenances m
		INNER JOIN assets a
			ON m.asset_id = a.asset_id
		INNER JOIN brands b
			ON a.brand_id = b.brand_id
		INNER JOIN categories c
			ON a.category_id = c.category_id
		WHERE 1=1
	`
	args := []any{}

	if maintenanceFilter.Search != "" {
		query += `
			AND (
				a.asset_name LIKE ?

			)
		`

		search := "%" + maintenanceFilter.Search + "%"
		args = append(args, search)
	}

	if maintenanceFilter.Status != "" {
		query += ` AND m.status = ? `
		args = append(args, maintenanceFilter.Status)
	}

	switch maintenanceFilter.OrderBy {
	case "created_at_asc":
		query += ` ORDER BY m.created_at ASC `
	case "created_at_desc":
		query += ` ORDER BY m.created_at DESC `
	case "cost_asc":
		query += ` ORDER BY m.cost ASC `
	case "cost_desc":
		query += ` ORDER BY m.cost DESC `
	default:
		query += ` ORDER BY m.created_at DESC `
	}
	offset := (maintenanceFilter.Page - 1) * maintenanceFilter.Limit

	query += ` LIMIT ? OFFSET ? `
	args = append(args, maintenanceFilter.Limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	maintenances := []Maintenance{}
	var completedAt sql.NullTime
	for rows.Next() {
		var maintenance Maintenance

		err := rows.Scan(
			&maintenance.MaintenanceId,
			&maintenance.Description,
			&maintenance.Cost,
			&maintenance.Status,
			&maintenance.AssetId,
			&maintenance.MaintenanceAt,
			&completedAt,
			&maintenance.Asset.AssetId,
			&maintenance.Asset.AssetName,
			&maintenance.Brand.BrandId,
			&maintenance.Brand.BrandName,
			&maintenance.Category.CategoryId,
			&maintenance.Category.CategoryName,
		)
		if completedAt.Valid {
			maintenance.CompletedAt = &completedAt.Time
		} else {
			maintenance.CompletedAt = nil
		}
		if err != nil {
			return nil, err
		}

		maintenances = append(maintenances, maintenance)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return maintenances, nil
}

func (r *MaintenanceRepository) GetMaintenanceById(ctx context.Context, maintenance_id string) (Maintenance, error) {
	query := `
		SELECT 
			m.maintenance_id,
			m.description,
			m.cost,
			m.status,
			m.asset_id,
			m.maintenance_at,
			m.completed_at,
			m.created_at,
			m.updated_at,

			a.asset_name,
			a.serial_number,

			b.brand_id,
			b.brand_name,
			
			c.category_id,
			c.category_name
		FROM maintenances m
		INNER JOIN assets a
			ON m.asset_id = a.asset_id
		INNER JOIN brands b
			ON a.brand_id = b.brand_id
		INNER JOIN categories c
			ON a.category_id = c.category_id
		WHERE m.maintenance_id = ?
	`

	row := r.db.QueryRowContext(ctx, query, maintenance_id)

	var maintenance Maintenance
	var completedAt sql.NullTime

	err := row.Scan(
		&maintenance.MaintenanceId,
		&maintenance.Description,
		&maintenance.Cost,
		&maintenance.Status,
		&maintenance.AssetId,
		&maintenance.MaintenanceAt,
		&completedAt,
		&maintenance.CreatedAt,
		&maintenance.UpdatedAt,
		&maintenance.Asset.AssetName,
		&maintenance.Asset.SerialNumber,
		&maintenance.Brand.BrandId,
		&maintenance.Brand.BrandName,
		&maintenance.Category.CategoryId,
		&maintenance.Category.CategoryName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Maintenance{}, nil
		}
		return Maintenance{}, err
	}

	if completedAt.Valid {
		maintenance.CompletedAt = &completedAt.Time
	} else {
		maintenance.CompletedAt = nil
	}

	return maintenance, nil
}

func (r *MaintenanceRepository) CreateMaintenanceTx(
	ctx context.Context,
	tx *sql.Tx,
	maintenance Maintenance,
) error {
	query := `
		INSERT INTO maintenances (
			maintenance_id,
			asset_id,
			description,
			cost,
			status,
			maintenance_at
		)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := tx.ExecContext(
		ctx,
		query,
		maintenance.MaintenanceId,
		maintenance.AssetId,
		maintenance.Description,
		maintenance.Cost,
		maintenance.Status,
		maintenance.MaintenanceAt,
	)

	return err
}

func (r *MaintenanceRepository) UpdateMaintenanceStatusTx(
	ctx context.Context,
	tx *sql.Tx,
	maintenance_id string,
	maintenanceUpdated Maintenance,
) error {
	query := `
		UPDATE maintenances
		SET status = ?,
		completed_at = ?
		WHERE maintenance_id = ?
	`

	result, err := tx.ExecContext(ctx, query, maintenanceUpdated.Status, maintenanceUpdated.CompletedAt, maintenance_id)
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
func (r *MaintenanceRepository) UpdateMaintenanceCostTx(
	ctx context.Context,
	tx *sql.Tx,
	maintenanceId string,
	cost int64,
) error {
	query := `
		UPDATE maintenances
		SET cost = ?
			WHERE maintenance_id = ?
	`

	result, err := tx.ExecContext(ctx, query, cost, maintenanceId)
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

func (r *MaintenanceRepository) UpdateMaintenanceDescriptionTx(
	ctx context.Context,
	tx *sql.Tx,
	maintenanceId string,
	description string,
) error {
	query := `
		UPDATE maintenances
		SET description = ?
			WHERE maintenance_id = ?
	`

	result, err := tx.ExecContext(ctx, query, description, maintenanceId)
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

func (r *MaintenanceRepository) GetTotalPageAndTotalDataMaintenances(ctx context.Context, filter MaintenanceFilter) (pkg.PaginationMeta, error) {
	// 1. Ubah SELECT menjadi COUNT untuk menghitung total baris saja
	// Hapus JOIN ke brand dan category jika tidak digunakan untuk filter SEARCH agar query jauh lebih cepat (Ringan)
	query := `
        SELECT COUNT(m.maintenance_id)
        FROM maintenances m
        INNER JOIN assets a ON m.asset_id = a.asset_id
        INNER JOIN brands b ON a.brand_id = b.brand_id
        INNER JOIN categories c ON a.category_id = c.category_id
        WHERE 1=1
    `
	args := []any{}

	// 2. Pertahankan logika filter search yang sama persis agar sinkron dengan data tabel
	if filter.Search != "" {
		query += `
            AND (
                a.asset_name LIKE ?
                OR a.serial_number LIKE ?
                OR b.brand_name LIKE ?
                OR c.category_name LIKE ?
                OR m.description LIKE ?
            )
        `
		search := "%" + filter.Search + "%"
		args = append(args, search, search, search, search, search)
	}

	var paginationData pkg.PaginationMeta
	var totalData int

	// 3. Scan sekarang akan berjalan sukses karena kolom yang di-SELECT hanya satu (Hasil COUNT)
	err := r.db.QueryRowContext(
		ctx,
		query,
		args...,
	).Scan(&totalData)

	if err != nil {
		return paginationData, err
	}

	// 4. Hitung estimasi total halaman berdasarkan limit baris per halaman
	var totalPage int
	if filter.Limit > 0 {
		totalPage = int(math.Ceil(
			float64(totalData) / float64(filter.Limit),
		))
	} else {
		totalPage = 1 // Fallback default jika limit kosong
	}

	// 5. Mapping metadata pagination kembali ke domain struct
	paginationData.Page = filter.Page
	paginationData.Limit = filter.Limit
	paginationData.TotalData = totalData
	paginationData.TotalPage = totalPage

	return paginationData, nil
}

// UpdateMaintenanceTx memperbarui data log maintenance menggunakan context transaksi aktif
func (r *MaintenanceRepository) UpdateMaintenanceTx(ctx context.Context, tx *sql.Tx, maintenanceID string, maintenance Maintenance) error {
	query := `
		UPDATE maintenances 
		SET 
			cost = ?, 
			description = ?, 
			status = ?, 
			completed_at = ?
		WHERE maintenance_id = ?;
	`

	// Eksekusi query menggunakan objek transaksi 'tx', bukan 'r.db'
	result, err := tx.ExecContext(ctx, query,
		maintenance.Cost,
		maintenance.Description,
		maintenance.Status,
		maintenance.CompletedAt, // Menggunakan pointer *time.Time agar bisa bernilai NULL di DB
		maintenanceID,
	)
	if err != nil {
		return err
	}

	// Opsional: Memastikan ada baris data yang benar-benar terupdate
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows // Mengembalikan error jika ID tidak ditemukan di DB
	}

	return nil
}
func (r *MaintenanceRepository) GetAllMaintenancesData(ctx context.Context) ([]Maintenance, error) {
	query := `
        SELECT 
            m.maintenance_id,
            m.description,
            m.cost,
            m.status,
            m.asset_id,
            m.maintenance_at,
            m.completed_at,
            m.created_at,
            m.updated_at,

            a.asset_name,
            a.serial_number,
			a.description,
			a.purchased_date,

            b.brand_id,
            b.brand_name,
            
            c.category_id,
            c.category_name
        FROM maintenances m
        INNER JOIN assets a
            ON m.asset_id = a.asset_id
        INNER JOIN brands b
            ON a.brand_id = b.brand_id
        INNER JOIN categories c
            ON a.category_id = c.category_id
    `
	// 1. Jalankan QueryContext untuk mengambil banyak baris data (rows)
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // Pastikan selalu di-close agar tidak kebocoran memori (memory leak)

	// 2. Siapkan slice kosong untuk menampung list maintenance
	var maintenances []Maintenance

	// 3. Looping baris demi baris dari database
	for rows.Next() {
		var m Maintenance

		// 4. Scan data secara berurutan sesuai dengan urutan kolom di SELECT query
		err := rows.Scan(
			&m.MaintenanceId,
			&m.Description,
			&m.Cost,
			&m.Status,
			&m.AssetId,
			&m.MaintenanceAt,
			&m.CompletedAt,
			&m.CreatedAt,
			&m.UpdatedAt,

			// Arahkan data Asset ke dalam nested struct Asset di dalam Maintenance
			&m.Asset.AssetName,
			&m.Asset.SerialNumber,
			&m.Asset.Description,
			&m.Asset.PurchasedDate,

			// Arahkan data Brand & Category ke dalam nested struct milik Asset
			&m.Asset.Brand.BrandId,
			&m.Asset.Brand.BrandName,

			&m.Asset.Category.CategoryId,
			&m.Asset.Category.CategoryName,
		)
		if err != nil {
			return nil, err
		}

		// 5. Masukkan data yang berhasil di-scan ke dalam slice utama
		maintenances = append(maintenances, m)
	}

	// 6. Cek apakah ada error yang terjadi selama proses looping berjalan
	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Kembalikan semua data
	return maintenances, nil
}
