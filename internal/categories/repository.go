package categories

import (
	"context"
	"database/sql"
	"inventory-it/internal/pkg"
	"math"
)

type Repository interface {
	GetAllCategories(context.Context, CategoryFilter) ([]Categories, error)
	GetCategoryById(context.Context, string) (Categories, error)
	CreateCategory(context.Context, Categories) error
	UpdateCategory(context.Context, string, Categories) error
	DeleteCategory(context.Context, string) error
	GetTotalPageAndTotalDataCategories(context.Context, CategoryFilter) (pkg.PaginationMeta, error)
}

type repository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) GetAllCategories(ctx context.Context, filter CategoryFilter) ([]Categories, error) {
	query := `SELECT category_id, category_name, created_at, updated_at FROM categories WHERE 1=1`
	args := []any{}

	if filter.Search != "" {
		query += ` AND category_name LIKE ? `
		search := "%" + filter.Search + "%"
		args = append(args, search)
	}

	switch filter.OrderBy {
	case "asc":
		query += ` ORDER BY category_name ASC `
	case "desc":
		query += ` ORDER BY category_name DESC `
	default:
		query += ` ORDER BY created_at DESC `
	}

	offset := (filter.Page - 1) * filter.Limit
	query += ` LIMIT ? OFFSET ? `
	args = append(args, filter.Limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var listCategories []Categories
	for rows.Next() {
		var category Categories
		err := rows.Scan(&category.CategoryId, &category.CategoryName, &category.CreatedAt, &category.UpdatedAt)
		if err != nil {
			return nil, err
		}
		listCategories = append(listCategories, category)
	}

	return listCategories, nil
}

func (r *repository) GetCategoryById(ctx context.Context, categoryId string) (Categories, error) {
	query := `SELECT category_id, category_name, created_at, updated_at FROM categories WHERE category_id = ?`
	row := r.db.QueryRowContext(ctx, query, categoryId)
	var category Categories
	err := row.Scan(&category.CategoryId, &category.CategoryName, &category.CreatedAt, &category.UpdatedAt)
	if err != nil {
		return Categories{}, err
	}

	return category, nil
}

func (r *repository) CreateCategory(ctx context.Context, category Categories) error {
	query := `INSERT INTO categories (category_id, category_name) VALUES (?, ?)`
	_, err := r.db.ExecContext(ctx, query, category.CategoryId, category.CategoryName)
	return err
}

func (r *repository) UpdateCategory(ctx context.Context, categoryId string, category Categories) error {
	query := `UPDATE categories SET category_name = ? WHERE category_id = ?`
	result, err := r.db.ExecContext(ctx, query, category.CategoryName, categoryId)
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

func (r *repository) DeleteCategory(ctx context.Context, categoryId string) error {
	query := `DELETE FROM categories WHERE category_id = ?`
	result, err := r.db.ExecContext(ctx, query, categoryId)
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

func (r *repository) GetTotalPageAndTotalDataCategories(ctx context.Context, filter CategoryFilter) (pkg.PaginationMeta, error) {
	query := `
		SELECT COUNT(*)
		FROM categories
	`

	args := []any{}
	var paginationData pkg.PaginationMeta

	if filter.Search != "" {
		query += `
			AND (
				category_name LIKE ?
	
			)
		`

		search := "%" + filter.Search + "%"
		args = append(args, search)
	}
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
