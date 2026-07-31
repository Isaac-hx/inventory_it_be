package assetassignments

import (
	"context"
	"database/sql"
	"inventory-it/internal/pkg"
	"math"
)

type Repository interface {
	CreateAssignmentTx(context.Context, *sql.Tx, AssetAssignment) error
	GetAllAssetAssignments(context.Context, AssetAssignmentFilter) ([]AssetAssignment, error)
	GetAssetAssignmentById(context.Context, string) (AssetAssignment, error)
	UpdateAssetAssignmentByIdTx(context.Context, *sql.Tx, string, AssetAssignment) error
	GetTotalPageAndTotalDataAssetAssignments(context.Context, AssetAssignmentFilter) (pkg.PaginationMeta, error)
	GetAllAssignmentsData(context.Context) ([]AssetAssignment, error)
	UpdateStatusAssignmentByIdTx(context.Context, *sql.Tx, string, AssetAssignment) error
	UpdateStatusAssignmentByAssetIdTx(context.Context, *sql.Tx, string, AssignmentStatus) error
	GetAssignmentsByUserId(context.Context, string, AssignmentStatus) ([]AssetAssignment, error)
}

type repository struct {
	db *sql.DB
}

func NewAssetAssignmentRepository(db *sql.DB) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) CreateAssignmentTx(ctx context.Context, tx *sql.Tx, assetAssignment AssetAssignment) error {
	query := `
		INSERT INTO asset_assignments (
			assignment_id,
			asset_id,
			user_id,
			corporation,
			assigned_by,
			status,
			notes,
			assigned_date
		)
		VALUES (?, ?, ?, ?,?, ?, ?, ?)
	`
	_, err := tx.ExecContext(
		ctx,
		query,
		assetAssignment.AssignmentId,
		assetAssignment.AssetId,
		assetAssignment.UserId,
		assetAssignment.Corporation,
		assetAssignment.AssignedById,
		assetAssignment.Status,
		assetAssignment.Notes,
		assetAssignment.AssignedDate,
	)
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) GetAssetAssignmentById(ctx context.Context, assignmentId string) (AssetAssignment, error) {
	query := `
	SELECT
    aa.assignment_id,
    aa.asset_id,
    aa.user_id,
    aa.assigned_by,
    aa.status,
    aa.notes,
    aa.assigned_date,
    aa.return_date,
    aa.created_at,
    aa.updated_at,
	aa.corporation,

    a.asset_name,
	a.processor,
	a.ram,
	a.storage,
    a.serial_number,
	a.quantity_stock,

    b.brand_id,
    b.brand_name,

    c.category_id,
    c.category_name,

    u.username AS assigned_to_username,
    u.email AS assigned_to_email,

	d.department_id,
	d.department_name,

    admin.username AS assigned_by_username

	FROM asset_assignments aa

	INNER JOIN assets a
		ON aa.asset_id = a.asset_id

	INNER JOIN brands b
		ON a.brand_id = b.brand_id

	INNER JOIN categories c
		ON a.category_id = c.category_id

	INNER JOIN users u
		ON aa.user_id = u.user_id

	INNER JOIN users admin
		ON aa.assigned_by = admin.user_id
	
	INNER JOIN departments d
		ON d.department_id = u.department_id

	WHERE aa.assignment_id = ?`

	var assetAssignment AssetAssignment
	err := r.db.QueryRowContext(ctx, query, assignmentId).Scan(
		&assetAssignment.AssignmentId,
		&assetAssignment.Asset.AssetId,
		&assetAssignment.UserId,
		&assetAssignment.AssignedById,
		&assetAssignment.Status,
		&assetAssignment.Notes,
		&assetAssignment.AssignedDate,
		&assetAssignment.ReturnDate,
		&assetAssignment.CreatedAt,
		&assetAssignment.UpdatedAt,
		&assetAssignment.Corporation,

		&assetAssignment.Asset.AssetName,
		&assetAssignment.Asset.Processor,
		&assetAssignment.Asset.Ram,
		&assetAssignment.Asset.Storage,
		&assetAssignment.Asset.SerialNumber,
		&assetAssignment.Asset.QuantityStock,
		&assetAssignment.Asset.BrandId,
		&assetAssignment.Asset.Brand.BrandName,

		&assetAssignment.Asset.CategoryId,
		&assetAssignment.Asset.CategoryName,

		&assetAssignment.User.Username,
		&assetAssignment.User.Email,

		&assetAssignment.User.DepartmentId,
		&assetAssignment.User.DepartmentName,
		&assetAssignment.AssignedByUsername,
	)
	if err != nil {
		return AssetAssignment{}, err
	}
	return assetAssignment, nil
}
func (r *repository) GetAllAssetAssignments(
	ctx context.Context,
	filter AssetAssignmentFilter,
) ([]AssetAssignment, error) {

	query := `
		SELECT
			aa.assignment_id,
			aa.asset_id,
			aa.user_id,
			aa.assigned_by,
			aa.status,
			aa.assigned_date,
			aa.corporation,

			a.asset_name,
			a.ram,
			a.processor,
			a.storage,

			u.username,

			admin.username AS assigned_by_username
		FROM asset_assignments aa
		INNER JOIN assets a
			ON aa.asset_id = a.asset_id
		INNER JOIN users u
			ON aa.user_id = u.user_id
		INNER JOIN users admin
			ON aa.assigned_by = admin.user_id
		WHERE 1=1
	`

	args := []any{}

	if filter.Search != "" {
		query += `
			AND (
				a.asset_name LIKE ?

			)
		`

		search := "%" + filter.Search + "%"
		args = append(args, search)
	}

	if filter.Status != "" {
		query += ` AND aa.status = ? `
		args = append(args, filter.Status)
	}

	switch filter.OrderBy {
	case "assigned_date_asc":
		query += ` ORDER BY aa.assigned_date ASC `
	case "assigned_date_desc":
		query += ` ORDER BY aa.assigned_date DESC `
	case "created_at_asc":
		query += ` ORDER BY aa.created_at ASC `
	default:
		query += ` ORDER BY aa.created_at DESC `
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}

	if filter.Limit <= 0 {
		filter.Limit = 10
	}

	offset := (filter.Page - 1) * filter.Limit

	query += ` LIMIT ? OFFSET ? `
	args = append(args, filter.Limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	assetAssignments := []AssetAssignment{}

	for rows.Next() {
		var assetAssignment AssetAssignment

		err := rows.Scan(
			&assetAssignment.AssignmentId,
			&assetAssignment.AssetId,
			&assetAssignment.UserId,
			&assetAssignment.AssignedById,
			&assetAssignment.Status,
			&assetAssignment.AssignedDate,
			&assetAssignment.Corporation,
			&assetAssignment.Asset.AssetName,
			&assetAssignment.Asset.Ram,
			&assetAssignment.Asset.Processor,
			&assetAssignment.Asset.Storage,
			&assetAssignment.User.Username,
			&assetAssignment.AssignedByUsername,
		)
		if err != nil {
			return nil, err
		}

		assetAssignments = append(assetAssignments, assetAssignment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return assetAssignments, nil
}

func (r *repository) GetTotalPageAndTotalDataAssetAssignments(ctx context.Context, filter AssetAssignmentFilter) (pkg.PaginationMeta, error) {
	// 👇 1. Tambahkan WHERE 1=1 agar struktur SQL valid saat digabung dengan AND
	query := `
        SELECT COUNT(*)
        FROM asset_assignments
        WHERE 1=1
    `

	args := []any{}
	var paginationData pkg.PaginationMeta

	if filter.Search != "" {
		query += `
            AND (
                asset_assignments.assignment_id LIKE ?
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

func (r *repository) UpdateAssetAssignmentByIdTx(ctx context.Context, tx *sql.Tx, assignmentId string, updatedAssignment AssetAssignment) error {
	query := `
	UPDATE asset_assignments SET
	asset_id = ?,
	status = ?,
	user_id = ?,
	Corporation = ?,
	notes = ?,
	return_date = ?
	WHERE assignment_id = ?

	`

	_, err := tx.ExecContext(ctx, query,
		updatedAssignment.AssetId,
		updatedAssignment.Status,
		updatedAssignment.UserId,
		updatedAssignment.Corporation,
		updatedAssignment.Notes,
		updatedAssignment.ReturnDate,
		assignmentId,
	)
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) GetAllAssignmentsData(ctx context.Context) ([]AssetAssignment, error) {
	query := `
    SELECT
        aa.assignment_id,
        aa.asset_id,
        aa.user_id,
        aa.assigned_by,
        aa.status,
        aa.notes,
        aa.assigned_date,
        aa.return_date,
        aa.created_at,
        aa.updated_at,
		aa.corporation,

        a.asset_id,        -- Ditambahkan koma yang kurang dari kode sebelumnya
        a.asset_name,
        a.serial_number,
        a.purchased_date,

        b.brand_id,
        b.brand_name,

        c.category_id,
        c.category_name,

        u.username AS assigned_to_username,
        u.email,
        u.password,
		u.Role,
        -- Ambil data departemen milik user penerima
        COALESCE(dept.department_id, '') AS department_id,
        COALESCE(dept.department_name, '') AS department_name,

        admin.username AS assigned_by_username
    FROM asset_assignments aa
    INNER JOIN assets a ON aa.asset_id = a.asset_id
    INNER JOIN brands b ON a.brand_id = b.brand_id
    INNER JOIN categories c ON a.category_id = c.category_id
    INNER JOIN users u ON aa.user_id = u.user_id
    INNER JOIN users admin ON aa.assigned_by = admin.user_id
    -- JOIN ke tabel departments berdasarkan department_id milik user (u)
    LEFT JOIN departments dept ON u.department_id = dept.department_id`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assignments []AssetAssignment

	for rows.Next() {
		var aa AssetAssignment

		// Scan harus 100% berurutan dengan kolom SELECT di atas!
		// Scan sekarang pas menjadi 25 argumen, presisi dengan SELECT
		err := rows.Scan(
			&aa.AssignmentId,
			&aa.AssetId,
			&aa.UserId,
			&aa.AssignedById,
			&aa.Status,
			&aa.Notes,
			&aa.AssignedDate,
			&aa.ReturnDate,
			&aa.CreatedAt,
			&aa.UpdatedAt,
			aa.Corporation,

			// Relasi Asset (SEKARANG SUDAH SINKRON)
			&aa.Asset.AssetId, // <-- TAMBAHKAN BARIS INI DI SINI
			&aa.Asset.AssetName,
			&aa.Asset.SerialNumber,
			&aa.Asset.PurchasedDate,

			// Relasi Brand
			&aa.Asset.Brand.BrandId,
			&aa.Asset.Brand.BrandName,

			// Relasi Category
			&aa.Asset.Category.CategoryId,
			&aa.Asset.Category.CategoryName,

			// Relasi User (Penerima)
			&aa.User.Username,
			&aa.User.Email,
			&aa.User.Password,
			&aa.User.Role,

			// Relasi Department milik User Penerima
			&aa.User.Department.DepartmentId,
			&aa.User.Department.DepartmentName,

			// Relasi Admin / User (Pemberi)
			&aa.AssignedByUsername)
		if err != nil {
			return nil, err
		}

		assignments = append(assignments, aa)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return assignments, nil
}

func (r *repository) UpdateStatusAssignmentByIdTx(ctx context.Context, tx *sql.Tx, assignmentId string, assignment AssetAssignment) error {
	query :=
		`
		UPDATE asset_assignments 
		SET status = ?,
		return_date = ?
		WHERE assignment_id = ?
	`

	_, err := tx.ExecContext(ctx, query,
		assignment.Status,
		assignment.ReturnDate,
		assignmentId,
	)
	if err != nil {
		return err
	}
	return nil
}
func (r *repository) GetAssignmentsByUserId(ctx context.Context, userId string, status AssignmentStatus) ([]AssetAssignment, error) {
	// Menambahkan filter status ke dalam WHERE clause menggunakan AND
	query := `
    SELECT
        aa.assignment_id,
        aa.asset_id,
        aa.user_id,
        aa.assigned_by,
        aa.status,
        aa.notes,
        aa.assigned_date,
        aa.return_date,
        COALESCE(aa.corporation, '') AS corporation,
        aa.created_at,
        aa.updated_at,

        a.asset_id,
        a.asset_name,
        a.serial_number,
        a.purchased_date,
        a.brand_id,
        a.category_id,

        b.brand_id,
        b.brand_name,

        c.category_id,
        c.category_name,

        u.username AS assigned_to_username,
        u.email,
        u.password,
        u.role, 
        COALESCE(dept.department_id, '') AS department_id,
        COALESCE(dept.department_name, '') AS department_name,

        admin.username AS assigned_by_username
    FROM asset_assignments aa
    INNER JOIN assets a ON aa.asset_id = a.asset_id
    INNER JOIN brands b ON a.brand_id = b.brand_id
    INNER JOIN categories c ON a.category_id = c.category_id
    INNER JOIN users u ON aa.user_id = u.user_id
    INNER JOIN users admin ON aa.assigned_by = admin.user_id
    LEFT JOIN departments dept ON u.department_id = dept.department_id
    WHERE aa.user_id = ? AND aa.status = ?` // <- Ditambahkan di sini

	// Passing userId dan status secara berurutan sesuai tanda tanya (?) di query
	rows, err := r.db.QueryContext(ctx, query, userId, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var assignments []AssetAssignment

	for rows.Next() {
		var aa AssetAssignment

		// Scan tetap presisi tanpa ada perubahan urutan kolom
		err := rows.Scan(
			&aa.AssignmentId,
			&aa.AssetId,
			&aa.UserId,
			&aa.AssignedById,
			&aa.Status,
			&aa.Notes,
			&aa.AssignedDate,
			&aa.ReturnDate,
			&aa.Corporation,
			&aa.CreatedAt,
			&aa.UpdatedAt,

			&aa.Asset.AssetId,
			&aa.Asset.AssetName,
			&aa.Asset.SerialNumber,
			&aa.Asset.PurchasedDate,
			&aa.Asset.BrandId,
			&aa.Asset.CategoryId,

			&aa.Asset.Brand.BrandId,
			&aa.Asset.Brand.BrandName,

			&aa.Asset.Category.CategoryId,
			&aa.Asset.Category.CategoryName,

			&aa.User.Username,
			&aa.User.Email,
			&aa.User.Password,
			&aa.User.Role,
			&aa.User.Department.DepartmentId,
			&aa.User.Department.DepartmentName,
			&aa.AssignedByUsername,
		)
		if err != nil {
			return nil, err
		}

		assignments = append(assignments, aa)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return assignments, nil
}
func (r *repository) UpdateStatusAssignmentByAssetIdTx(ctx context.Context, tx *sql.Tx, assetId string, status AssignmentStatus) error {
	query :=
		`
		UPDATE asset_assignments 
		SET status = ?
		WHERE asset_id = ?
	`

	_, err := tx.ExecContext(ctx, query,
		status,
		assetId,
	)
	if err != nil {
		return err
	}
	return nil
}
