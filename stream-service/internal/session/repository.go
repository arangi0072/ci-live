package session

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

type Repository interface {
	Create(
		ctx context.Context,
		session *StreamSession,
	) error

	GetByID(
		ctx context.Context,
		sessionID uuid.UUID,
	) (*StreamSession, error)

	List(
		ctx context.Context,
		filter SessionFilter,
	) ([]StreamSession, int64, error)

	Update(
		ctx context.Context,
		session *StreamSession,
	) error

	EndSession(
		ctx context.Context,
		sessionID uuid.UUID,
		endedAt time.Time,
		durationSeconds int64,
	) error

	GetByStreamID(
		ctx context.Context,
		streamID uuid.UUID,
	) ([]StreamSession, error)

	GetActiveByStreamID(
		ctx context.Context,
		streamID uuid.UUID,
	) (*StreamSession, error)

	GetSessionOwner(
		ctx context.Context,
		sessionID uuid.UUID,
	) (*SessionOwner, error)

	GetStreamOwner(
		ctx context.Context,
		streamID uuid.UUID,
	) (*SessionOwner, error)

	Delete(
		ctx context.Context,
		sessionID uuid.UUID,
	) error
}

//
// Create Session
//

func (r *repository) Create(
	ctx context.Context,
	session *StreamSession,
) error {

	const query = `
		INSERT INTO stream_sessions (
			id,
			stream_id,
			media_server_id,
			media_path,
			started_at,
			ended_at,
			duration_seconds,
			peak_viewers,
			created_at
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
			$9
		)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		session.ID,
		session.StreamID,
		session.MediaServerID,
		session.MediaPath,
		session.StartedAt,
		session.EndedAt,
		session.DurationSeconds,
		session.PeakViewers,
		session.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf(
			"create stream session: %w",
			err,
		)
	}

	return nil
}

//
// Get Session By ID
//

func (r *repository) GetByID(
	ctx context.Context,
	sessionID uuid.UUID,
) (*StreamSession, error) {

	const query = `
		SELECT
			id,
			stream_id,
			media_server_id,
			media_path,
			started_at,
			ended_at,
			duration_seconds,
			peak_viewers,
			created_at
		FROM stream_sessions
		WHERE id = $1
	`

	session := &StreamSession{}

	err := r.db.QueryRow(
		ctx,
		query,
		sessionID,
	).Scan(
		&session.ID,
		&session.StreamID,
		&session.MediaServerID,
		&session.MediaPath,
		&session.StartedAt,
		&session.EndedAt,
		&session.DurationSeconds,
		&session.PeakViewers,
		&session.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSessionNotFound
		}

		return nil, fmt.Errorf(
			"get stream session: %w",
			err,
		)
	}

	return session, nil
}

//
// List Sessions
//

func (r *repository) List(
	ctx context.Context,
	filter SessionFilter,
) ([]StreamSession, int64, error) {

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

	//
	// Count
	//

	var total int64

	if filter.StreamID != nil {

		const countQuery = `
			SELECT COUNT(*)
			FROM stream_sessions
			WHERE stream_id = $1
		`

		err := r.db.QueryRow(
			ctx,
			countQuery,
			*filter.StreamID,
		).Scan(&total)

		if err != nil {
			return nil, 0, fmt.Errorf(
				"count stream sessions: %w",
				err,
			)
		}

	} else {

		const countQuery = `
			SELECT COUNT(*)
			FROM stream_sessions
		`

		err := r.db.QueryRow(
			ctx,
			countQuery,
		).Scan(&total)

		if err != nil {
			return nil, 0, fmt.Errorf(
				"count stream sessions: %w",
				err,
			)
		}
	}

	//
	// Query
	//

	var (
		rows pgx.Rows
		err  error
	)

	if filter.StreamID != nil {

		const query = `
			SELECT
				id,
				stream_id,
				media_server_id,
				media_path,
				started_at,
				ended_at,
				duration_seconds,
				peak_viewers,
				created_at
			FROM stream_sessions
			WHERE stream_id = $1
			ORDER BY created_at DESC
			LIMIT $2
			OFFSET $3
		`

		rows, err = r.db.Query(
			ctx,
			query,
			*filter.StreamID,
			filter.Limit,
			offset,
		)

	} else {

		const query = `
			SELECT
				id,
				stream_id,
				media_server_id,
				media_path,
				started_at,
				ended_at,
				duration_seconds,
				peak_viewers,
				created_at
			FROM stream_sessions
			ORDER BY created_at DESC
			LIMIT $1
			OFFSET $2
		`

		rows, err = r.db.Query(
			ctx,
			query,
			filter.Limit,
			offset,
		)
	}

	if err != nil {
		return nil, 0, fmt.Errorf(
			"list stream sessions: %w",
			err,
		)
	}

	defer rows.Close()

	sessions := make(
		[]StreamSession,
		0,
		filter.Limit,
	)

	for rows.Next() {

		var session StreamSession

		err := rows.Scan(
			&session.ID,
			&session.StreamID,
			&session.MediaServerID,
			&session.MediaPath,
			&session.StartedAt,
			&session.EndedAt,
			&session.DurationSeconds,
			&session.PeakViewers,
			&session.CreatedAt,
		)

		if err != nil {
			return nil, 0, fmt.Errorf(
				"scan stream session: %w",
				err,
			)
		}

		sessions = append(
			sessions,
			session,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf(
			"iterate stream sessions: %w",
			err,
		)
	}

	return sessions, total, nil
}

//
// Update Session
//

func (r *repository) Update(
	ctx context.Context,
	session *StreamSession,
) error {

	const query = `
		UPDATE stream_sessions
		SET
			media_server_id = $2,
			media_path = $3,
			started_at = $4,
			ended_at = $5,
			duration_seconds = $6,
			peak_viewers = $7
		WHERE id = $1
	`

	result, err := r.db.Exec(
		ctx,
		query,
		session.ID,
		session.MediaServerID,
		session.MediaPath,
		session.StartedAt,
		session.EndedAt,
		session.DurationSeconds,
		session.PeakViewers,
	)

	if err != nil {
		return fmt.Errorf(
			"update stream session: %w",
			err,
		)
	}

	if result.RowsAffected() == 0 {
		return ErrSessionNotFound
	}

	return nil
}

//
// End Session
//

func (r *repository) EndSession(
	ctx context.Context,
	sessionID uuid.UUID,
	endedAt time.Time,
	durationSeconds int64,
) error {

	const query = `
		UPDATE stream_sessions
		SET
			ended_at = $2,
			duration_seconds = $3
		WHERE id = $1
		  AND ended_at IS NULL
	`

	result, err := r.db.Exec(
		ctx,
		query,
		sessionID,
		endedAt,
		durationSeconds,
	)

	if err != nil {
		return fmt.Errorf(
			"end stream session: %w",
			err,
		)
	}

	if result.RowsAffected() == 0 {
		return ErrSessionNotFound
	}

	return nil
}

//
// Get Sessions By Stream
//

func (r *repository) GetByStreamID(
	ctx context.Context,
	streamID uuid.UUID,
) ([]StreamSession, error) {

	const query = `
		SELECT
			id,
			stream_id,
			media_server_id,
			media_path,
			started_at,
			ended_at,
			duration_seconds,
			peak_viewers,
			created_at
		FROM stream_sessions
		WHERE stream_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(
		ctx,
		query,
		streamID,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"get stream sessions: %w",
			err,
		)
	}

	defer rows.Close()

	sessions := make(
		[]StreamSession,
		0,
	)

	for rows.Next() {

		var session StreamSession

		err := rows.Scan(
			&session.ID,
			&session.StreamID,
			&session.MediaServerID,
			&session.MediaPath,
			&session.StartedAt,
			&session.EndedAt,
			&session.DurationSeconds,
			&session.PeakViewers,
			&session.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf(
				"scan stream session: %w",
				err,
			)
		}

		sessions = append(
			sessions,
			session,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate stream sessions: %w",
			err,
		)
	}

	return sessions, nil
}

//
// Get Active Session
//

func (r *repository) GetActiveByStreamID(
	ctx context.Context,
	streamID uuid.UUID,
) (*StreamSession, error) {

	const query = `
		SELECT
			id,
			stream_id,
			media_server_id,
			media_path,
			started_at,
			ended_at,
			duration_seconds,
			peak_viewers,
			created_at
		FROM stream_sessions
		WHERE stream_id = $1
		  AND ended_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`

	session := &StreamSession{}

	err := r.db.QueryRow(
		ctx,
		query,
		streamID,
	).Scan(
		&session.ID,
		&session.StreamID,
		&session.MediaServerID,
		&session.MediaPath,
		&session.StartedAt,
		&session.EndedAt,
		&session.DurationSeconds,
		&session.PeakViewers,
		&session.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSessionNotFound
		}

		return nil, fmt.Errorf(
			"get active stream session: %w",
			err,
		)
	}

	return session, nil
}

//
// Get Session Owner
//
// stream_sessions
//      ↓ stream_id
// streams
//      ↓ channel_id
// channels
//      ↓ user_id
//

func (r *repository) GetSessionOwner(
	ctx context.Context,
	sessionID uuid.UUID,
) (*SessionOwner, error) {

	const query = `
		SELECT
			ss.id AS session_id,
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

	owner := &SessionOwner{}

	err := r.db.QueryRow(
		ctx,
		query,
		sessionID,
	).Scan(
		&owner.SessionID,
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
// Get Stream Owner
//

func (r *repository) GetStreamOwner(
	ctx context.Context,
	streamID uuid.UUID,
) (*SessionOwner, error) {

	const query = `
		SELECT
			s.id AS stream_id,
			s.channel_id,
			c.user_id
		FROM streams s
		INNER JOIN channels c
			ON c.id = s.channel_id
		WHERE s.id = $1
	`

	owner := &SessionOwner{
		StreamID: streamID,
	}

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
// Delete Session
//

func (r *repository) Delete(
	ctx context.Context,
	sessionID uuid.UUID,
) error {

	const query = `
		DELETE FROM stream_sessions
		WHERE id = $1
	`

	result, err := r.db.Exec(
		ctx,
		query,
		sessionID,
	)

	if err != nil {
		return fmt.Errorf(
			"delete stream session: %w",
			err,
		)
	}

	if result.RowsAffected() == 0 {
		return ErrSessionNotFound
	}

	return nil
}
