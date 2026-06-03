// package assetassignments

// import (
// 	"context"
// 	"database/sql"
// )

// type Repository interface {
// 	CreateAssetAssignmentTx(context.Context, *sql.Tx, AssetAssignment) error
// 	GetAllAssetAssignments(context.Context, AssetAssignmentFilter) ([]AssetAssignment, error)
// 	GetAssetAssignmentById(context.Context, string) (AssetAssignment, error)
// 	UpdateAssetAssignment(context.Context, string, AssetAssignment) error
// 	DeleteAssetAssignment(context.Context, string) error
// }

// type repository struct {
// 	db *sql.DB
// }

// func NewAssetAssignmentRepository(db *sql.DB) Repository {
// 	return &repository{
// 		db: db,
// 	}
// }

// func (r *repository) CreateAssetAssignmentTx(ctx context.Context, tx *sql.Tx, assetAssignment AssetAssignment) error {
// 	query := `
// 		INSERT INTO asset_assignments (
// 			assignment_id,
// 			asset_id,
// 			user_id,
// 			assigned_by,
// 			status,
// 			notes,
// 			assigned_date,
// 			return_date
// 		)
// 		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
// 	`

// 	_, err := tx.ExecContext(
// 		ctx,
// 		query,
// 		assetAssignment.AssignmentId,
// 		assetAssignment.AssetId,
// 		assetAssignment.UserId,
// 		assetAssignment.AssignedBy,
// 		assetAssignment.Status,
// 		assetAssignment.Notes,
// 		assetAssignment.AssignedDate,
// 		assetAssignment.ReturnDate,
// 	)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (r *repository) GetAssetAssignmentById(ctx context.Context, assignmentId string) (AssetAssignment, error) {
// 	query := `
// 	SELECT
// 		aa.assignment_id,
// 		aa.asset_id,
// 		aa.user_id,
// 		aa.assigned_by,
// 		aa.status,
// 		aa.notes,
// 		aa.assigned_date,
// 		aa.return_date,
// 		aa.created_at,
// 		aa.updated_at,

// 		a.asset_name,
// 		a.serial_number,

// 		b.brand_id,
// 		b.brand_name,

// 		c.category_id,
// 		c.category_name,

// 		u.username,
// 		u.email,

// 	FROM asset_assignments aa

// 	INNER JOIN assets a
// 		ON aa.asset_id = a.asset_id

// 	INNER JOIN brands b
// 		ON a.brand_id = b.brand_id

// 	INNER JOIN categories c
// 		ON a.category_id = c.category_id

// 	INNER JOIN users u
// 		ON aa.user_id = u.user_id

// 	INNER JOIN users admin
// 		ON aa.assigned_by = admin.user_id

// 	WHERE aa.assignment_id = ?
// `

// 	var assetAssignment AssetAssignment
// 	err := r.db.QueryRowContext(ctx, query, assignmentId).Scan(
// 		&assetAssignment.AssignmentId,
// 		&assetAssignment.AssetId,
// 		&assetAssignment.UserId,
// 		&assetAssignment.AssignedBy,
// 		&assetAssignment.Status,
// 		&assetAssignment.Notes,
// 		&assetAssignment.AssignedDate,
// 		&assetAssignment.ReturnDate,
// 		&assetAssignment.CreatedAt,
// 		&assetAssignment.UpdatedAt,
// 		&assetAssignment.Asset.AssetName,
// 		&assetAssignment.Asset.SerialNumber,
// 		&assetAssignment.Asset.BrandId,
// 		&assetAssignment.Asset.BrandName,
// 		&assetAssignment.Asset.CategoryId,
// 		&assetAssignment.Asset.CategoryName,
// 		&assetAssignment.User.Username,
// 		&assetAssignment.User.Email,
// 	)
// 	if err != nil {
// 		return AssetAssignment{}, err
// 	}
// 	return assetAssignment, nil
// }
// func (r *repository) GetAllAssetAssignments(
// 	ctx context.Context,
// 	filter AssetAssignmentFilter,
// ) ([]AssetAssignment, error) {

// 	query := `
// 		SELECT
// 			aa.assignment_id,
// 			aa.asset_id,
// 			aa.user_id,
// 			aa.assigned_by,
// 			aa.status,
// 			aa.notes,
// 			aa.assigned_date,
// 			aa.return_date,
// 			aa.created_at,
// 			aa.updated_at,

// 			a.asset_name,
// 			a.serial_number,

// 			b.brand_id,
// 			b.brand_name,

// 			c.category_id,
// 			c.category_name,

// 			u.username,
// 			u.email,

// 			admin.username AS assigned_by_username
// 		FROM asset_assignments aa
// 		INNER JOIN assets a
// 			ON aa.asset_id = a.asset_id
// 		INNER JOIN brands b
// 			ON a.brand_id = b.brand_id
// 		INNER JOIN categories c
// 			ON a.category_id = c.category_id
// 		INNER JOIN users u
// 			ON aa.user_id = u.user_id
// 		INNER JOIN users admin
// 			ON aa.assigned_by = admin.user_id
// 		WHERE 1=1
// 	`

// 	args := []any{}

// 	if filter.Search != "" {
// 		query += `
// 			AND (
// 				a.asset_name LIKE ?
// 				OR a.serial_number LIKE ?
// 				OR b.brand_name LIKE ?
// 				OR c.category_name LIKE ?
// 				OR u.username LIKE ?
// 				OR admin.username LIKE ?
// 			)
// 		`

// 		search := "%" + filter.Search + "%"
// 		args = append(args, search, search, search, search, search, search)
// 	}

// 	if filter.Status != "" {
// 		query += ` AND aa.status = ? `
// 		args = append(args, filter.Status)
// 	}

// 	switch filter.OrderBy {
// 	case "assigned_date_asc":
// 		query += ` ORDER BY aa.assigned_date ASC `
// 	case "assigned_date_desc":
// 		query += ` ORDER BY aa.assigned_date DESC `
// 	case "created_at_asc":
// 		query += ` ORDER BY aa.created_at ASC `
// 	default:
// 		query += ` ORDER BY aa.created_at DESC `
// 	}

// 	if filter.Page <= 0 {
// 		filter.Page = 1
// 	}

// 	if filter.Limit <= 0 {
// 		filter.Limit = 10
// 	}

// 	offset := (filter.Page - 1) * filter.Limit

// 	query += ` LIMIT ? OFFSET ? `
// 	args = append(args, filter.Limit, offset)

// 	rows, err := r.db.QueryContext(ctx, query, args...)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	assetAssignments := []AssetAssignment{}

// 	for rows.Next() {
// 		var assetAssignment AssetAssignment
// 		var returnDate sql.NullTime

// 		err := rows.Scan(
// 			&assetAssignment.AssignmentId,
// 			&assetAssignment.AssetId,
// 			&assetAssignment.UserId,
// 			&assetAssignment.AssignedBy,
// 			&assetAssignment.Status,
// 			&assetAssignment.Notes,
// 			&assetAssignment.AssignedDate,
// 			&returnDate,
// 			&assetAssignment.CreatedAt,
// 			&assetAssignment.UpdatedAt,

// 			&assetAssignment.Asset.AssetName,
// 			&assetAssignment.Asset.SerialNumber,

// 			&assetAssignment.Brand.BrandId,
// 			&assetAssignment.Brand.BrandName,

// 			&assetAssignment.Category.CategoryId,
// 			&assetAssignment.Category.CategoryName,

// 			&assetAssignment.User.Username,
// 			&assetAssignment.User.Email,

// 			&assetAssignment.AssignedByUser.Username,
// 		)
// 		if err != nil {
// 			return nil, err
// 		}

// 		if returnDate.Valid {
// 			assetAssignment.ReturnDate = &returnDate.Time
// 		} else {
// 			assetAssignment.ReturnDate = nil
// 		}

// 		assetAssignments = append(assetAssignments, assetAssignment)
// 	}

// 	if err := rows.Err(); err != nil {
// 		return nil, err
// 	}

// 	return assetAssignments, nil
// }
