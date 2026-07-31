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
	CreateRequest(context.Context, Maintenance) error
	GetAllMaintenancesByUserId(context.Context, string) ([]Maintenance, error)
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
			m.assignment_id,
			m.maintenance_at,
			m.completed_at,

			a.asset_id,
			a.asset_name,

			b.brand_id,
			b.brand_name,
			
			c.category_id,
			c.category_name
		FROM maintenances m
		INNER JOIN asset_assignments aa ON m.assignment_id = aa.assignment_id
		INNER JOIN assets a ON aa.asset_id = a.asset_id
		INNER JOIN brands b ON a.brand_id = b.brand_id
		INNER JOIN categories c ON a.category_id = c.category_id
		WHERE 1=1
	`
	args := []any{}

	if maintenanceFilter.Search != "" {
		query += ` AND (a.asset_name LIKE ?) `
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

	// Inisialisasi slice kosong biar gak nge-return nil ke JSON parser frontend
	maintenances := []Maintenance{}

	for rows.Next() {
		var m Maintenance
		var completedAt sql.NullTime

		err := rows.Scan(
			&m.MaintenanceId,
			&m.Description,
			&m.Cost,
			&m.Status,
			&m.Assignment.AssignmentId, // Pas di urutan m.assignment_id
			&m.MaintenanceAt,
			&completedAt,
			&m.Asset.AssetId,
			&m.Asset.AssetName,
			&m.Brand.BrandId,
			&m.Brand.BrandName,
			&m.Category.CategoryId,
			&m.Category.CategoryName,
		)
		if err != nil {
			return nil, err
		}

		if completedAt.Valid {
			m.CompletedAt = &completedAt.Time
		} else {
			m.CompletedAt = nil
		}

		maintenances = append(maintenances, m)
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
			m.assignment_id, -- Menggunakan assignment_id sesuai relasi table terbaru
			m.maintenance_at,
			m.completed_at,
			m.created_at,
			m.updated_at,

			a.asset_id,      -- Ditambahkan agar struct m.Asset.AssetId terisi dengan benar
			a.asset_name,
			a.serial_number,

			b.brand_id,
			b.brand_name,
			
			c.category_id,
			c.category_name
		FROM maintenances m
		INNER JOIN asset_assignments aa ON m.assignment_id = aa.assignment_id
		INNER JOIN assets a ON aa.asset_id = a.asset_id
		INNER JOIN brands b ON a.brand_id = b.brand_id
		INNER JOIN categories c ON a.category_id = c.category_id
		WHERE m.maintenance_id = ?
	`
	row := r.db.QueryRowContext(ctx, query, maintenance_id)

	var m Maintenance
	var completedAt sql.NullTime

	err := row.Scan(
		&m.MaintenanceId,
		&m.Description,
		&m.Cost,
		&m.Status,
		&m.Assignment.AssignmentId, // Menampung m.assignment_id
		&m.MaintenanceAt,
		&completedAt,
		&m.CreatedAt,
		&m.UpdatedAt,
		&m.Asset.AssetId, // Menampung a.asset_id asli
		&m.Asset.AssetName,
		&m.Asset.SerialNumber,
		&m.Brand.BrandId,
		&m.Brand.BrandName,
		&m.Category.CategoryId,
		&m.Category.CategoryName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Maintenance{}, nil // Atau return custom error ErrNotFound jika dibutuhkan
		}
		return Maintenance{}, err
	}

	if completedAt.Valid {
		m.CompletedAt = &completedAt.Time
	} else {
		m.CompletedAt = nil
	}

	return m, nil
}

func (r *MaintenanceRepository) CreateMaintenanceTx(ctx context.Context, tx *sql.Tx, m Maintenance) error {
	query := `
        INSERT INTO maintenances (
            maintenance_id,
            assignment_id,
            description,
            cost,
            status,
            maintenance_at
        ) VALUES (?, ?, ?, ?, ?, ?)
    `
	_, err := tx.ExecContext(ctx, query, m.MaintenanceId, m.Assignment.AssignmentId, m.Description, m.Cost, m.Status, m.MaintenanceAt)
	return err
}

func (r *MaintenanceRepository) UpdateMaintenanceStatusTx(ctx context.Context, tx *sql.Tx, maintenance_id string, maintenanceUpdated Maintenance) error {
	query := `
        UPDATE maintenances
        SET status = ?, completed_at = ?
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

func (r *MaintenanceRepository) UpdateMaintenanceCostTx(ctx context.Context, tx *sql.Tx, maintenanceId string, cost int64) error {
	query := `UPDATE maintenances SET cost = ? WHERE maintenance_id = ?`
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

func (r *MaintenanceRepository) UpdateMaintenanceDescriptionTx(ctx context.Context, tx *sql.Tx, maintenanceId string, description string) error {
	query := `UPDATE maintenances SET description = ? WHERE maintenance_id = ?`
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
	query := `
		SELECT COUNT(m.maintenance_id)
		FROM maintenances m
		INNER JOIN asset_assignments aa ON m.assignment_id = aa.assignment_id
		INNER JOIN assets a ON aa.asset_id = a.asset_id
		INNER JOIN brands b ON a.brand_id = b.brand_id
		INNER JOIN categories c ON a.category_id = c.category_id
		WHERE 1=1
	`
	args := []any{}

	// Menambahkan filter status jika ada di filter payload (opsional, tapi biasanya sinkron dengan GetAllMaintenances)
	if filter.Status != "" {
		query += ` AND m.status = ? `
		args = append(args, filter.Status)
	}

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

	err := r.db.QueryRowContext(ctx, query, args...).Scan(&totalData)
	if err != nil {
		return paginationData, err
	}

	var totalPage int
	if filter.Limit > 0 {
		totalPage = int(math.Ceil(float64(totalData) / float64(filter.Limit)))
	} else {
		totalPage = 1
	}

	paginationData.Page = filter.Page
	paginationData.Limit = filter.Limit
	paginationData.TotalData = totalData
	paginationData.TotalPage = totalPage

	return paginationData, nil
}

func (r *MaintenanceRepository) UpdateMaintenanceTx(ctx context.Context, tx *sql.Tx, maintenanceID string, m Maintenance) error {
	query := `
        UPDATE maintenances
        SET cost = ?, description = ?, status = ?, completed_at = ?
        WHERE maintenance_id = ?;
    `
	result, err := tx.ExecContext(ctx, query, m.Cost, m.Description, m.Status, m.CompletedAt, maintenanceID)
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
func (r *MaintenanceRepository) GetAllMaintenancesData(ctx context.Context) ([]Maintenance, error) {
	query := `
		SELECT
			m.maintenance_id,
			m.description,
			m.cost,
			m.status,
			m.assignment_id, -- Diubah dari m.asset_id ke m.assignment_id
			m.maintenance_at,
			m.completed_at,
			m.created_at,
			m.updated_at,

			a.asset_id,      -- Ditambahkan supaya m.Asset.AssetId dapet ID aslinya
			a.asset_name,
			a.serial_number,
			a.description,
			a.purchased_date,

			b.brand_id,
			b.brand_name,
			
			c.category_id,
			c.category_name
		FROM maintenances m
		INNER JOIN asset_assignments aa ON m.assignment_id = aa.assignment_id
		INNER JOIN assets a ON aa.asset_id = a.asset_id
		INNER JOIN brands b ON a.brand_id = b.brand_id
		INNER JOIN categories c ON a.category_id = c.category_id
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Inisialisasi slice kosong biar ga nil pas dilempar ke front-end
	maintenances := []Maintenance{}

	for rows.Next() {
		var m Maintenance
		var completedAt sql.NullTime

		err := rows.Scan(
			&m.MaintenanceId,
			&m.Description,
			&m.Cost,
			&m.Status,
			&m.Assignment.AssignmentId, // Menampung m.assignment_id dengan benar
			&m.MaintenanceAt,
			&completedAt,
			&m.CreatedAt,
			&m.UpdatedAt,
			&m.Asset.AssetId, // Menampung a.asset_id asli
			&m.Asset.AssetName,
			&m.Asset.SerialNumber,
			&m.Asset.Description,
			&m.Asset.PurchasedDate,
			&m.Brand.BrandId,
			&m.Brand.BrandName,
			&m.Category.CategoryId,
			&m.Category.CategoryName,
		)
		if err != nil {
			return nil, err
		}

		if completedAt.Valid {
			m.CompletedAt = &completedAt.Time
		} else {
			m.CompletedAt = nil
		}

		maintenances = append(maintenances, m)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return maintenances, nil
}

func (r *MaintenanceRepository) CreateRequest(ctx context.Context, m Maintenance) error {
	query := `
        INSERT INTO maintenances(maintenance_id, user_id, asset_id, description) 
        VALUES(?, ?, ?, ?)
    `
	_, err := r.db.ExecContext(ctx, query, m.MaintenanceId, m.User.UserId, m.Asset.AssetId, m.Description)
	return err
}
func (r *MaintenanceRepository) GetAllMaintenancesByUserId(ctx context.Context, userId string) ([]Maintenance, error) {
	query := `
	SELECT
		m.maintenance_id,
		m.status,
		m.description,
		m.cost,
		m.maintenance_at,
		m.completed_at,
		m.created_at,
		m.updated_at,
		m.assignment_id, -- Diubah dari m.asset_id ke m.assignment_id
		aa.user_id,
		
		a.asset_id,      -- Ditambahkan agar m.Asset.AssetId dapet ID aslinya
		a.asset_name,
		a.serial_number,
		a.purchased_date,
		a.processor,
		a.ram,
		a.storage,

		c.category_name,
		b.brand_name
	FROM maintenances m
	LEFT JOIN asset_assignments aa ON m.assignment_id = aa.assignment_id
	LEFT JOIN assets a ON aa.asset_id = a.asset_id
	LEFT JOIN categories c ON a.category_id = c.category_id
	LEFT JOIN brands b ON a.brand_id = b.brand_id
	WHERE aa.user_id = ?
	ORDER BY m.maintenance_at DESC;
	`
	rows, err := r.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Inisialisasi slice kosong agar tidak me-return null ke front-end
	maintenances := []Maintenance{}

	for rows.Next() {
		var m Maintenance
		var completedAt sql.NullTime

		err := rows.Scan(
			&m.MaintenanceId,
			&m.Status,
			&m.Description,
			&m.Cost,
			&m.MaintenanceAt,
			&completedAt,
			&m.CreatedAt,
			&m.UpdatedAt,
			&m.Assignment.AssignmentId, // Menampung m.assignment_id
			&m.User.UserId,             // Menampung m.user_id langsung ke field UserId utama
			&m.Asset.AssetId,           // Menampung a.asset_id asli
			&m.Asset.AssetName,
			&m.Asset.SerialNumber,
			&m.Asset.PurchasedDate,
			&m.Asset.Processor,
			&m.Asset.Ram,
			&m.Asset.Storage,
			&m.Category.CategoryName,
			&m.Brand.BrandName,
		)
		if err != nil {
			return nil, err
		}

		if completedAt.Valid {
			m.CompletedAt = &completedAt.Time
		} else {
			m.CompletedAt = nil
		}

		maintenances = append(maintenances, m)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return maintenances, nil
}
