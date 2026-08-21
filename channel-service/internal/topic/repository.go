package topic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type topicRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &topicRepository{
		db: db,
	}
}

// Compile-time interface check.
var _ Repository = (*topicRepository)(nil)

// ============================================================
// Create Topic
// ============================================================

func (r *topicRepository) Create(
	ctx context.Context,
	topic *Topic,
) error {
	const query = `
		INSERT INTO topics (
			id,
			category_id,
			name,
			slug,
			description,
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
		topic.ID,
		topic.CategoryID,
		topic.Name,
		topic.Slug,
		topic.Description,
		topic.StreamCount,
	)

	if err != nil {
		return fmt.Errorf(
			"create topic: %w",
			err,
		)
	}

	return nil
}

// ============================================================
// Get Topic By ID
// ============================================================

func (r *topicRepository) GetByID(
	ctx context.Context,
	id string,
) (*Topic, error) {
	const query = `
		SELECT
			id,
			category_id,
			name,
			slug,
			description,
			stream_count,
			created_at,
			updated_at
		FROM topics
		WHERE id = $1
	`

	var topic Topic

	err := r.db.QueryRowContext(
		ctx,
		query,
		id,
	).Scan(
		&topic.ID,
		&topic.CategoryID,
		&topic.Name,
		&topic.Slug,
		&topic.Description,
		&topic.StreamCount,
		&topic.CreatedAt,
		&topic.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTopicNotFound
		}

		return nil, fmt.Errorf(
			"get topic by id: %w",
			err,
		)
	}

	return &topic, nil
}

// ============================================================
// Get Topic By Slug
// ============================================================

func (r *topicRepository) GetBySlug(
	ctx context.Context,
	slug string,
) (*Topic, error) {
	const query = `
		SELECT
			id,
			category_id,
			name,
			slug,
			description,
			stream_count,
			created_at,
			updated_at
		FROM topics
		WHERE slug = $1
	`

	var topic Topic

	err := r.db.QueryRowContext(
		ctx,
		query,
		slug,
	).Scan(
		&topic.ID,
		&topic.CategoryID,
		&topic.Name,
		&topic.Slug,
		&topic.Description,
		&topic.StreamCount,
		&topic.CreatedAt,
		&topic.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTopicNotFound
		}

		return nil, fmt.Errorf(
			"get topic by slug: %w",
			err,
		)
	}

	return &topic, nil
}

// ============================================================
// List Topics
// ============================================================

func (r *topicRepository) List(
	ctx context.Context,
	categoryID string,
) ([]Topic, error) {

	var (
		rows *sql.Rows
		err  error
	)

	// --------------------------------------------------------
	// List all topics
	// --------------------------------------------------------

	if categoryID == "" {
		const query = `
			SELECT
				id,
				category_id,
				name,
				slug,
				description,
				stream_count,
				created_at,
				updated_at
			FROM topics
			ORDER BY name ASC
		`

		rows, err = r.db.QueryContext(
			ctx,
			query,
		)
	} else {

		// ----------------------------------------------------
		// List topics belonging to category
		// ----------------------------------------------------

		const query = `
			SELECT
				id,
				category_id,
				name,
				slug,
				description,
				stream_count,
				created_at,
				updated_at
			FROM topics
			WHERE category_id = $1
			ORDER BY name ASC
		`

		rows, err = r.db.QueryContext(
			ctx,
			query,
			categoryID,
		)
	}

	if err != nil {
		return nil, fmt.Errorf(
			"list topics: %w",
			err,
		)
	}

	defer rows.Close()

	topics := make(
		[]Topic,
		0,
	)

	for rows.Next() {
		var topic Topic

		if err := rows.Scan(
			&topic.ID,
			&topic.CategoryID,
			&topic.Name,
			&topic.Slug,
			&topic.Description,
			&topic.StreamCount,
			&topic.CreatedAt,
			&topic.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf(
				"scan topic: %w",
				err,
			)
		}

		topics = append(
			topics,
			topic,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate topics: %w",
			err,
		)
	}

	return topics, nil
}

// ============================================================
// Update Topic
// ============================================================

func (r *topicRepository) Update(
	ctx context.Context,
	id string,
	req UpdateTopicRequest,
) (*Topic, error) {

	const query = `
		UPDATE topics
		SET
			name = COALESCE($1, name),
			slug = COALESCE($2, slug),
			description = COALESCE($3, description),
			updated_at = NOW()
		WHERE id = $4
		RETURNING
			id,
			category_id,
			name,
			slug,
			description,
			stream_count,
			created_at,
			updated_at
	`

	var (
		name        interface{}
		slug        interface{}
		description interface{}
	)

	if req.Name != nil {
		name = *req.Name
		slug = generateSlug(*req.Name)
	}

	if req.Description != nil {
		description = *req.Description
	}

	var topic Topic

	err := r.db.QueryRowContext(
		ctx,
		query,
		name,
		slug,
		description,
		id,
	).Scan(
		&topic.ID,
		&topic.CategoryID,
		&topic.Name,
		&topic.Slug,
		&topic.Description,
		&topic.StreamCount,
		&topic.CreatedAt,
		&topic.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTopicNotFound
		}

		return nil, fmt.Errorf(
			"update topic: %w",
			err,
		)
	}

	return &topic, nil
}

// ============================================================
// Delete Topic
// ============================================================

func (r *topicRepository) Delete(
	ctx context.Context,
	id string,
) error {
	const query = `
		DELETE FROM topics
		WHERE id = $1
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		id,
	)

	if err != nil {
		return fmt.Errorf(
			"delete topic: %w",
			err,
		)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"get deleted topic rows: %w",
			err,
		)
	}

	if rows == 0 {
		return ErrTopicNotFound
	}

	return nil
}
