package stream

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{
		db: db,
	}
}

//
// Create
//

func (r *repository) Create(
	ctx context.Context,
	stream *Stream,
) error {
	const query = `
		INSERT INTO streams (
			id,
			channel_id,
			title,
			description,
			thumbnail_url,
			category_id,
			status,
			visibility,
			viewer_count,
			peak_viewers,
			like_count,
			total_views,
			started_at,
			ended_at,
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
			$7,
			$8,
			$9,
			$10,
			$11,
			$12,
			$13,
			$14,
			$15,
			$16
		)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		stream.ID,
		stream.ChannelID,
		stream.Title,
		stream.Description,
		stream.ThumbnailURL,
		stream.CategoryID,
		stream.Status,
		stream.Visibility,
		stream.ViewerCount,
		stream.PeakViewers,
		stream.LikeCount,
		stream.TotalViews,
		stream.StartedAt,
		stream.EndedAt,
		stream.CreatedAt,
		stream.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("create stream: %w", err)
	}

	return nil
}

//
// Get By ID
//

func (r *repository) GetByID(
	ctx context.Context,
	streamID uuid.UUID,
) (*Stream, error) {
	const query = `
		SELECT
			id,
			channel_id,
			title,
			description,
			thumbnail_url,
			category_id,
			status,
			visibility,
			viewer_count,
			peak_viewers,
			like_count,
			total_views,
			started_at,
			ended_at,
			created_at,
			updated_at
		FROM streams
		WHERE id = $1
	`

	stream := &Stream{}

	err := r.db.QueryRow(
		ctx,
		query,
		streamID,
	).Scan(
		&stream.ID,
		&stream.ChannelID,
		&stream.Title,
		&stream.Description,
		&stream.ThumbnailURL,
		&stream.CategoryID,
		&stream.Status,
		&stream.Visibility,
		&stream.ViewerCount,
		&stream.PeakViewers,
		&stream.LikeCount,
		&stream.TotalViews,
		&stream.StartedAt,
		&stream.EndedAt,
		&stream.CreatedAt,
		&stream.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrStreamNotFound
		}

		return nil, fmt.Errorf("get stream: %w", err)
	}

	return stream, nil
}

//
// Get Public Stream
//

func (r *repository) GetPublicByID(
	ctx context.Context,
	streamID uuid.UUID,
) (*Stream, error) {
	const query = `
		SELECT
			id,
			channel_id,
			title,
			description,
			thumbnail_url,
			category_id,
			status,
			visibility,
			viewer_count,
			peak_viewers,
			like_count,
			total_views,
			started_at,
			ended_at,
			created_at,
			updated_at
		FROM streams
		WHERE id = $1
		  AND visibility = 'PUBLIC'
	`

	stream := &Stream{}

	err := r.db.QueryRow(
		ctx,
		query,
		streamID,
	).Scan(
		&stream.ID,
		&stream.ChannelID,
		&stream.Title,
		&stream.Description,
		&stream.ThumbnailURL,
		&stream.CategoryID,
		&stream.Status,
		&stream.Visibility,
		&stream.ViewerCount,
		&stream.PeakViewers,
		&stream.LikeCount,
		&stream.TotalViews,
		&stream.StartedAt,
		&stream.EndedAt,
		&stream.CreatedAt,
		&stream.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrStreamNotFound
		}

		return nil, fmt.Errorf(
			"get public stream: %w",
			err,
		)
	}

	return stream, nil
}

//
// List
//

func (r *repository) List(
	ctx context.Context,
	filter StreamFilter,
) ([]Stream, int64, error) {

	if filter.Page < 1 {
		filter.Page = 1
	}

	if filter.Limit < 1 {
		filter.Limit = 20
	}

	if filter.Limit > 100 {
		filter.Limit = 100
	}

	offset := (filter.Page - 1) * filter.Limit

	conditions := []string{
		"1 = 1",
	}

	args := make([]any, 0, 6)
	argIndex := 1

	//
	// Channel filter
	//

	if filter.ChannelID != nil {
		conditions = append(
			conditions,
			fmt.Sprintf(
				"channel_id = $%d",
				argIndex,
			),
		)

		args = append(
			args,
			*filter.ChannelID,
		)

		argIndex++
	}

	//
	// Category filter
	//

	if filter.CategoryID != nil {
		conditions = append(
			conditions,
			fmt.Sprintf(
				"category_id = $%d",
				argIndex,
			),
		)

		args = append(
			args,
			*filter.CategoryID,
		)

		argIndex++
	}

	//
	// Status filter
	//

	if filter.Status != nil {
		conditions = append(
			conditions,
			fmt.Sprintf(
				"status = $%d",
				argIndex,
			),
		)

		args = append(
			args,
			*filter.Status,
		)

		argIndex++
	}

	//
	// Visibility filter
	//

	if filter.Visibility != nil {
		conditions = append(
			conditions,
			fmt.Sprintf(
				"visibility = $%d",
				argIndex,
			),
		)

		args = append(
			args,
			*filter.Visibility,
		)

		argIndex++
	}

	whereClause := ""

	for i, condition := range conditions {
		if i == 0 {
			whereClause = " WHERE " + condition
			continue
		}

		whereClause += " AND " + condition
	}

	//
	// Count
	//

	countQuery := `
		SELECT COUNT(*)
		FROM streams
	` + whereClause

	var total int64

	if err := r.db.QueryRow(
		ctx,
		countQuery,
		args...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf(
			"count streams: %w",
			err,
		)
	}

	//
	// Data
	//

	dataQuery := `
		SELECT
			id,
			channel_id,
			title,
			description,
			thumbnail_url,
			category_id,
			status,
			visibility,
			viewer_count,
			peak_viewers,
			like_count,
			total_views,
			started_at,
			ended_at,
			created_at,
			updated_at
		FROM streams
	` + whereClause + fmt.Sprintf(`
		ORDER BY created_at DESC
		LIMIT $%d
		OFFSET $%d
	`, argIndex, argIndex+1)

	args = append(
		args,
		filter.Limit,
		offset,
	)

	rows, err := r.db.Query(
		ctx,
		dataQuery,
		args...,
	)

	if err != nil {
		return nil, 0, fmt.Errorf(
			"list streams: %w",
			err,
		)
	}

	defer rows.Close()

	streams := make(
		[]Stream,
		0,
	)

	for rows.Next() {
		var stream Stream

		if err := rows.Scan(
			&stream.ID,
			&stream.ChannelID,
			&stream.Title,
			&stream.Description,
			&stream.ThumbnailURL,
			&stream.CategoryID,
			&stream.Status,
			&stream.Visibility,
			&stream.ViewerCount,
			&stream.PeakViewers,
			&stream.LikeCount,
			&stream.TotalViews,
			&stream.StartedAt,
			&stream.EndedAt,
			&stream.CreatedAt,
			&stream.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf(
				"scan stream: %w",
				err,
			)
		}

		streams = append(
			streams,
			stream,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf(
			"iterate streams: %w",
			err,
		)
	}

	return streams, total, nil
}

//
// Update
//

func (r *repository) Update(
	ctx context.Context,
	stream *Stream,
) error {
	const query = `
		UPDATE streams
		SET
			title = $2,
			description = $3,
			thumbnail_url = $4,
			category_id = $5,
			visibility = $6,
			updated_at = $7
		WHERE id = $1
	`

	result, err := r.db.Exec(
		ctx,
		query,
		stream.ID,
		stream.Title,
		stream.Description,
		stream.ThumbnailURL,
		stream.CategoryID,
		stream.Visibility,
		stream.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf(
			"update stream: %w",
			err,
		)
	}

	if result.RowsAffected() == 0 {
		return ErrStreamNotFound
	}

	return nil
}

//
// Delete
//

func (r *repository) Delete(
	ctx context.Context,
	streamID uuid.UUID,
) error {
	const query = `
		DELETE FROM streams
		WHERE id = $1
	`

	result, err := r.db.Exec(
		ctx,
		query,
		streamID,
	)

	if err != nil {
		return fmt.Errorf(
			"delete stream: %w",
			err,
		)
	}

	if result.RowsAffected() == 0 {
		return ErrStreamNotFound
	}

	return nil
}

//
// Get Stream Owner
//
// streams.channel_id -> channels.id -> channels.user_id
//

func (r *repository) GetStreamOwner(
	ctx context.Context,
	streamID uuid.UUID,
) (*StreamOwner, error) {
	const query = `
		SELECT
			s.id,
			s.channel_id,
			c.user_id
		FROM streams s
		INNER JOIN channels c
			ON c.id = s.channel_id
		WHERE s.id = $1
	`

	owner := &StreamOwner{}

	err := r.db.QueryRow(
		ctx,
		query,
		streamID,
	).Scan(
		&owner.StreamID,
		&owner.ChannelID,
		&owner.UserID,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrStreamNotFound
		}

		return nil, fmt.Errorf(
			"get stream owner: %w",
			err,
		)
	}

	return owner, nil
}

//
// Get Channel By User ID
//

func (r *repository) GetChannelByUserID(
	ctx context.Context,
	userID uuid.UUID,
) (*ChannelOwner, error) {
	const query = `
		SELECT
			id,
			user_id
		FROM channels
		WHERE user_id = $1
		  AND is_active = TRUE
		LIMIT 1
	`

	channel := &ChannelOwner{}

	err := r.db.QueryRow(
		ctx,
		query,
		userID,
	).Scan(
		&channel.ID,
		&channel.UserID,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrChannelNotFound
		}

		return nil, fmt.Errorf(
			"get channel by user: %w",
			err,
		)
	}

	return channel, nil
}

//
// Get Statistics
//

func (r *repository) GetStats(
	ctx context.Context,
	streamID uuid.UUID,
) (*StreamStats, error) {
	const query = `
		SELECT
			id,
			viewer_count,
			peak_viewers,
			like_count,
			total_views
		FROM streams
		WHERE id = $1
	`

	stats := &StreamStats{}

	err := r.db.QueryRow(
		ctx,
		query,
		streamID,
	).Scan(
		&stats.StreamID,
		&stats.ViewerCount,
		&stats.PeakViewers,
		&stats.LikeCount,
		&stats.TotalViews,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrStreamNotFound
		}

		return nil, fmt.Errorf(
			"get stream stats: %w",
			err,
		)
	}

	return stats, nil
}

//
// Increment Views
//

func (r *repository) IncrementViews(
	ctx context.Context,
	streamID uuid.UUID,
) error {
	const query = `
		UPDATE streams
		SET
			total_views = total_views + 1,
			updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Exec(
		ctx,
		query,
		streamID,
	)

	if err != nil {
		return fmt.Errorf(
			"increment stream views: %w",
			err,
		)
	}

	if result.RowsAffected() == 0 {
		return ErrStreamNotFound
	}

	return nil
}

//
// Update Viewer Count
//
// Peak viewers is updated atomically:
//
// peak_viewers = GREATEST(peak_viewers, viewer_count)
//

func (r *repository) UpdateViewerCount(
	ctx context.Context,
	streamID uuid.UUID,
	viewerCount int,
	peakViewers int,
) error {
	const query = `
		UPDATE streams
		SET
			viewer_count = $2,
			peak_viewers = GREATEST(peak_viewers, $3),
			updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Exec(
		ctx,
		query,
		streamID,
		viewerCount,
		peakViewers,
	)

	if err != nil {
		return fmt.Errorf(
			"update viewer count: %w",
			err,
		)
	}

	if result.RowsAffected() == 0 {
		return ErrStreamNotFound
	}

	return nil
}

//
// Update Status
//

func (r *repository) UpdateStatus(
	ctx context.Context,
	streamID uuid.UUID,
	status StreamStatus,
	startedAt *time.Time,
	endedAt *time.Time,
) error {
	const query = `
		UPDATE streams
		SET
			status = $2,
			started_at = COALESCE($3, started_at),
			ended_at = COALESCE($4, ended_at),
			updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Exec(
		ctx,
		query,
		streamID,
		status,
		startedAt,
		endedAt,
	)

	if err != nil {
		return fmt.Errorf(
			"update stream status: %w",
			err,
		)
	}

	if result.RowsAffected() == 0 {
		return ErrStreamNotFound
	}

	return nil
}
