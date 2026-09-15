package recording

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//
// Repository
//

type Repository interface {
	Create(
		ctx context.Context,
		recording *StreamRecording,
	) error

	GetByID(
		ctx context.Context,
		recordingID uuid.UUID,
	) (*StreamRecording, error)

	List(
		ctx context.Context,
		filter RecordingFilter,
	) ([]StreamRecording, int64, error)

	Update(
		ctx context.Context,
		recording *StreamRecording,
	) error

	Delete(
		ctx context.Context,
		recordingID uuid.UUID,
	) error

	GetRecordingOwner(
		ctx context.Context,
		recordingID uuid.UUID,
	) (*RecordingOwner, error)

	GetStreamOwner(
		ctx context.Context,
		streamID uuid.UUID,
	) (*RecordingOwner, error)

	GetSessionOwner(
		ctx context.Context,
		sessionID uuid.UUID,
	) (*RecordingOwner, error)
}

//
// Repository Implementation
//

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
	recording *StreamRecording,
) error {

	const query = `
		INSERT INTO stream_recordings (
			id,
			stream_id,
			session_id,
			storage_provider,
			storage_key,
			playback_url,
			thumbnail_url,
			format,
			width,
			height,
			file_size,
			duration_seconds,
			view_count,
			status,
			created_at,
			completed_at
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
		recording.ID,
		recording.StreamID,
		recording.SessionID,
		recording.StorageProvider,
		recording.StorageKey,
		recording.PlaybackURL,
		recording.ThumbnailURL,
		recording.Format,
		recording.Width,
		recording.Height,
		recording.FileSize,
		recording.DurationSeconds,
		recording.ViewCount,
		recording.Status,
		recording.CreatedAt,
		recording.CompletedAt,
	)

	if err != nil {
		return fmt.Errorf(
			"create recording: %w",
			err,
		)
	}

	return nil
}

//
// Get By ID
//

func (r *repository) GetByID(
	ctx context.Context,
	recordingID uuid.UUID,
) (*StreamRecording, error) {

	const query = `
		SELECT
			id,
			stream_id,
			session_id,
			storage_provider,
			storage_key,
			playback_url,
			thumbnail_url,
			format,
			width,
			height,
			file_size,
			duration_seconds,
			view_count,
			status,
			created_at,
			completed_at
		FROM stream_recordings
		WHERE id = $1
	`

	recording := &StreamRecording{}

	err := r.db.QueryRow(
		ctx,
		query,
		recordingID,
	).Scan(
		&recording.ID,
		&recording.StreamID,
		&recording.SessionID,
		&recording.StorageProvider,
		&recording.StorageKey,
		&recording.PlaybackURL,
		&recording.ThumbnailURL,
		&recording.Format,
		&recording.Width,
		&recording.Height,
		&recording.FileSize,
		&recording.DurationSeconds,
		&recording.ViewCount,
		&recording.Status,
		&recording.CreatedAt,
		&recording.CompletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRecordingNotFound
		}

		return nil, fmt.Errorf(
			"get recording: %w",
			err,
		)
	}

	return recording, nil
}

//
// List
//

func (r *repository) List(
	ctx context.Context,
	filter RecordingFilter,
) ([]StreamRecording, int64, error) {

	page := filter.Page
	if page < 1 {
		page = 1
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}

	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	args := make([]interface{}, 0, 5)
	conditions := make([]string, 0, 3)

	arg := 1

	//
	// Stream filter
	//

	if filter.StreamID != nil {
		conditions = append(
			conditions,
			fmt.Sprintf("stream_id = $%d", arg),
		)

		args = append(
			args,
			*filter.StreamID,
		)

		arg++
	}

	//
	// Session filter
	//

	if filter.SessionID != nil {
		conditions = append(
			conditions,
			fmt.Sprintf("session_id = $%d", arg),
		)

		args = append(
			args,
			*filter.SessionID,
		)

		arg++
	}

	//
	// Status filter
	//

	if filter.Status != nil {
		conditions = append(
			conditions,
			fmt.Sprintf("status = $%d", arg),
		)

		args = append(
			args,
			*filter.Status,
		)

		arg++
	}

	where := ""

	if len(conditions) > 0 {
		where = "WHERE " + joinConditions(conditions)
	}

	//
	// Count
	//

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM stream_recordings
		%s
	`, where)

	var total int64

	if err := r.db.QueryRow(
		ctx,
		countQuery,
		args...,
	).Scan(&total); err != nil {

		return nil, 0, fmt.Errorf(
			"count recordings: %w",
			err,
		)
	}

	//
	// List
	//

	query := fmt.Sprintf(`
		SELECT
			id,
			stream_id,
			session_id,
			storage_provider,
			storage_key,
			playback_url,
			thumbnail_url,
			format,
			width,
			height,
			file_size,
			duration_seconds,
			view_count,
			status,
			created_at,
			completed_at
		FROM stream_recordings
		%s
		ORDER BY created_at DESC
		LIMIT $%d
		OFFSET $%d
	`,
		where,
		arg,
		arg+1,
	)

	args = append(
		args,
		limit,
		offset,
	)

	rows, err := r.db.Query(
		ctx,
		query,
		args...,
	)

	if err != nil {
		return nil, 0, fmt.Errorf(
			"list recordings: %w",
			err,
		)
	}

	defer rows.Close()

	recordings := make(
		[]StreamRecording,
		0,
	)

	for rows.Next() {

		var recording StreamRecording

		err := rows.Scan(
			&recording.ID,
			&recording.StreamID,
			&recording.SessionID,
			&recording.StorageProvider,
			&recording.StorageKey,
			&recording.PlaybackURL,
			&recording.ThumbnailURL,
			&recording.Format,
			&recording.Width,
			&recording.Height,
			&recording.FileSize,
			&recording.DurationSeconds,
			&recording.ViewCount,
			&recording.Status,
			&recording.CreatedAt,
			&recording.CompletedAt,
		)

		if err != nil {
			return nil, 0, fmt.Errorf(
				"scan recording: %w",
				err,
			)
		}

		recordings = append(
			recordings,
			recording,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf(
			"iterate recordings: %w",
			err,
		)
	}

	return recordings, total, nil
}

//
// Update
//

func (r *repository) Update(
	ctx context.Context,
	recording *StreamRecording,
) error {

	const query = `
		UPDATE stream_recordings
		SET
			storage_provider = $2,
			storage_key = $3,
			playback_url = $4,
			thumbnail_url = $5,
			format = $6,
			width = $7,
			height = $8,
			file_size = $9,
			duration_seconds = $10,
			view_count = $11,
			status = $12,
			completed_at = $13
		WHERE id = $1
	`

	result, err := r.db.Exec(
		ctx,
		query,
		recording.ID,
		recording.StorageProvider,
		recording.StorageKey,
		recording.PlaybackURL,
		recording.ThumbnailURL,
		recording.Format,
		recording.Width,
		recording.Height,
		recording.FileSize,
		recording.DurationSeconds,
		recording.ViewCount,
		recording.Status,
		recording.CompletedAt,
	)

	if err != nil {
		return fmt.Errorf(
			"update recording: %w",
			err,
		)
	}

	if result.RowsAffected() == 0 {
		return ErrRecordingNotFound
	}

	return nil
}

//
// Delete
//
// Soft delete using the DELETED status.
//

func (r *repository) Delete(
	ctx context.Context,
	recordingID uuid.UUID,
) error {

	const query = `
		UPDATE stream_recordings
		SET
			status = 'DELETED',
			completed_at = COALESCE(completed_at, NOW())
		WHERE id = $1
	`

	result, err := r.db.Exec(
		ctx,
		query,
		recordingID,
	)

	if err != nil {
		return fmt.Errorf(
			"delete recording: %w",
			err,
		)
	}

	if result.RowsAffected() == 0 {
		return ErrRecordingNotFound
	}

	return nil
}

//
// Get Recording Owner
//

func (r *repository) GetRecordingOwner(
	ctx context.Context,
	recordingID uuid.UUID,
) (*RecordingOwner, error) {

	const query = `
		SELECT
			sr.id,
			sr.stream_id,
			s.channel_id,
			c.user_id
		FROM stream_recordings sr
		INNER JOIN streams s
			ON s.id = sr.stream_id
		INNER JOIN channels c
			ON c.id = s.channel_id
		WHERE sr.id = $1
	`

	owner := &RecordingOwner{}

	err := r.db.QueryRow(
		ctx,
		query,
		recordingID,
	).Scan(
		&owner.RecordingID,
		&owner.StreamID,
		&owner.ChannelID,
		&owner.UserID,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRecordingNotFound
		}

		return nil, fmt.Errorf(
			"get recording owner: %w",
			err,
		)
	}

	return owner, nil
}

//
// Get Stream Owner
//

func (r *repository) GetStreamOwner(
	ctx context.Context,
	streamID uuid.UUID,
) (*RecordingOwner, error) {

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

	owner := &RecordingOwner{}

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
// Get Session Owner
//

func (r *repository) GetSessionOwner(
	ctx context.Context,
	sessionID uuid.UUID,
) (*RecordingOwner, error) {

	const query = `
		SELECT
			ss.stream_id,
			s.channel_id,
			c.user_id
		FROM stream_sessions ss
		INNER JOIN streams s
			ON s.id = ss.stream_id
		INNER JOIN channels c
			ON c.id = s.channel_id
		WHERE ss.id = $1
	`

	owner := &RecordingOwner{}

	err := r.db.QueryRow(
		ctx,
		query,
		sessionID,
	).Scan(
		&owner.StreamID,
		&owner.ChannelID,
		&owner.UserID,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSessionNotFound
		}

		return nil, fmt.Errorf(
			"get session owner: %w",
			err,
		)
	}

	return owner, nil
}

//
// SQL Condition Helper
//

func joinConditions(
	conditions []string,
) string {

	if len(conditions) == 0 {
		return ""
	}

	result := conditions[0]

	for i := 1; i < len(conditions); i++ {
		result += " AND " + conditions[i]
	}

	return result
}
