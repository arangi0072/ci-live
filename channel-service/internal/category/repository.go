package category

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type categoryRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &categoryRepository{
		db: db,
	}
}

// Compile-time interface check.
var _ Repository = (*categoryRepository)(nil)

// ============================================================
// Create
// ============================================================

func (r *categoryRepository) Create(
	ctx context.Context,
	category *Category,
) error {
	const query = `
		INSERT INTO categories (
			id,
			name,
			slug,
			description,
			icon_url,
			stream_count,
			created_at,
			updated_at
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			NOW(),
			NOW()
		)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		category.ID,
		category.Name,
		category.Slug,
		category.Description,
		category.IconURL,
		category.StreamCount,
	)

	if err != nil {
		return fmt.Errorf(
			"create category: %w",
			err,
		)
	}

	return nil
}

// ============================================================
// Get By ID
// ============================================================

func (r *categoryRepository) GetByID(
	ctx context.Context,
	id string,
) (*Category, error) {
	const query = `
		SELECT
			id,
			name,
			slug,
			description,
			icon_url,
			stream_count,
			created_at,
			updated_at
		FROM categories
		WHERE id = $1
	`

	var category Category

	err := r.db.QueryRowContext(
		ctx,
		query,
		id,
	).Scan(
		&category.ID,
		&category.Name,
		&category.Slug,
		&category.Description,
		&category.IconURL,
		&category.StreamCount,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}

		return nil, fmt.Errorf(
			"get category by id: %w",
			err,
		)
	}

	return &category, nil
}

// ============================================================
// Get By Slug
// ============================================================

func (r *categoryRepository) GetBySlug(
	ctx context.Context,
	slug string,
) (*Category, error) {
	const query = `
		SELECT
			id,
			name,
			slug,
			description,
			icon_url,
			stream_count,
			created_at,
			updated_at
		FROM categories
		WHERE slug = $1
	`

	var category Category

	err := r.db.QueryRowContext(
		ctx,
		query,
		slug,
	).Scan(
		&category.ID,
		&category.Name,
		&category.Slug,
		&category.Description,
		&category.IconURL,
		&category.StreamCount,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}

		return nil, fmt.Errorf(
			"get category by slug: %w",
			err,
		)
	}

	return &category, nil
}

// ============================================================
// List
// ============================================================

func (r *categoryRepository) List(
	ctx context.Context,
) ([]Category, error) {
	const query = `
		SELECT
			id,
			name,
			slug,
			description,
			icon_url,
			stream_count,
			created_at,
			updated_at
		FROM categories
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"list categories: %w",
			err,
		)
	}

	defer rows.Close()

	categories := make(
		[]Category,
		0,
	)

	for rows.Next() {
		var category Category

		if err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Slug,
			&category.Description,
			&category.IconURL,
			&category.StreamCount,
			&category.CreatedAt,
			&category.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf(
				"scan category: %w",
				err,
			)
		}

		categories = append(
			categories,
			category,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate categories: %w",
			err,
		)
	}

	return categories, nil
}

// ============================================================
// Update
// ============================================================

func (r *categoryRepository) Update(
	ctx context.Context,
	id string,
	req UpdateCategoryRequest,
) (*Category, error) {

	const query = `
		UPDATE categories
		SET
			name = COALESCE($1, name),
			slug = COALESCE($2, slug),
			description = COALESCE($3, description),
			icon_url = COALESCE($4, icon_url),
			updated_at = NOW()
		WHERE id = $5
		RETURNING
			id,
			name,
			slug,
			description,
			icon_url,
			stream_count,
			created_at,
			updated_at
	`

	var name interface{}
	var slug interface{}
	var description interface{}
	var iconURL interface{}

	if req.Name != nil {
		name = *req.Name

		// Keep slug synchronized with category name.
		slug = generateSlug(*req.Name)
	}

	if req.Description != nil {
		description = *req.Description
	}

	if req.IconURL != nil {
		iconURL = *req.IconURL
	}

	var category Category

	err := r.db.QueryRowContext(
		ctx,
		query,
		name,
		slug,
		description,
		iconURL,
		id,
	).Scan(
		&category.ID,
		&category.Name,
		&category.Slug,
		&category.Description,
		&category.IconURL,
		&category.StreamCount,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}

		return nil, fmt.Errorf(
			"update category: %w",
			err,
		)
	}

	return &category, nil
}

// ============================================================
// Delete
// ============================================================

func (r *categoryRepository) Delete(
	ctx context.Context,
	id string,
) error {
	const query = `
		DELETE FROM categories
		WHERE id = $1
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		id,
	)

	if err != nil {
		return fmt.Errorf(
			"delete category: %w",
			err,
		)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"get deleted category rows: %w",
			err,
		)
	}

	if rows == 0 {
		return ErrCategoryNotFound
	}

	return nil
}
