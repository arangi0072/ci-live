package like

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//
// PostgreSQL Repository
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
// Create Like
//

func (r *repository) CreateLike(
	ctx context.Context,
	like *StreamLike,
) (*StreamLike, error) {

	const query = `
		INSERT INTO stream_likes (
			stream_id,
			user_id
		)
		VALUES ($1, $2)
		RETURNING
			stream_id,
			user_id,
			created_at
	`

	var result StreamLike

	err := r.db.QueryRow(
		ctx,
		query,
		like.StreamID,
		like.UserID,
	).Scan(
		&result.StreamID,
		&result.UserID,
		&result.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

//
// Get Like
//

func (r *repository) GetLike(
	ctx context.Context,
	streamID uuid.UUID,
	userID uuid.UUID,
) (*StreamLike, error) {

	const query = `
		SELECT
			stream_id,
			user_id,
			created_at
		FROM stream_likes
		WHERE
			stream_id = $1
			AND user_id = $2
	`

	var like StreamLike

	err := r.db.QueryRow(
		ctx,
		query,
		streamID,
		userID,
	).Scan(
		&like.StreamID,
		&like.UserID,
		&like.CreatedAt,
	)

	if err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &like, nil
}

//
// Delete Like
//

func (r *repository) DeleteLike(
	ctx context.Context,
	streamID uuid.UUID,
	userID uuid.UUID,
) error {

	const query = `
		DELETE FROM stream_likes
		WHERE
			stream_id = $1
			AND user_id = $2
	`

	commandTag, err := r.db.Exec(
		ctx,
		query,
		streamID,
		userID,
	)

	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return ErrLikeNotFound
	}

	return nil
}

//
// Is Liked
//

func (r *repository) IsLiked(
	ctx context.Context,
	streamID uuid.UUID,
	userID uuid.UUID,
) (bool, error) {

	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM stream_likes
			WHERE
				stream_id = $1
				AND user_id = $2
		)
	`

	var liked bool

	err := r.db.QueryRow(
		ctx,
		query,
		streamID,
		userID,
	).Scan(&liked)

	if err != nil {
		return false, err
	}

	return liked, nil
}

//
// List Likes
//

func (r *repository) ListLikes(
	ctx context.Context,
	filter LikeFilter,
) ([]StreamLike, int64, error) {

	conditions := make([]string, 0)
	args := make([]any, 0)

	parameterIndex := 1

	//
	// Stream filter
	//

	if filter.StreamID != nil {
		conditions = append(
			conditions,
			fmt.Sprintf(
				"stream_id = $%d",
				parameterIndex,
			),
		)

		args = append(
			args,
			*filter.StreamID,
		)

		parameterIndex++
	}

	//
	// User filter
	//

	if filter.UserID != nil {
		conditions = append(
			conditions,
			fmt.Sprintf(
				"user_id = $%d",
				parameterIndex,
			),
		)

		args = append(
			args,
			*filter.UserID,
		)

		parameterIndex++
	}

	//
	// Build WHERE clause.
	//

	whereClause := ""

	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(
			conditions,
			" AND ",
		)
	}

	//
	// Count
	//

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM stream_likes
		%s
	`, whereClause)

	var total int64

	err := r.db.QueryRow(
		ctx,
		countQuery,
		args...,
	).Scan(&total)

	if err != nil {
		return nil, 0, err
	}

	//
	// Pagination
	//

	page := filter.Page

	if page < 1 {
		page = 1
	}

	limit := filter.Limit

	if limit < 1 {
		limit = 50
	}

	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	//
	// Fetch likes.
	//

	query := fmt.Sprintf(`
		SELECT
			stream_id,
			user_id,
			created_at
		FROM stream_likes
		%s
		ORDER BY created_at DESC
		LIMIT $%d
		OFFSET $%d
	`,
		whereClause,
		parameterIndex,
		parameterIndex+1,
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
		return nil, 0, err
	}

	defer rows.Close()

	likes := make(
		[]StreamLike,
		0,
		limit,
	)

	for rows.Next() {

		var like StreamLike

		err := rows.Scan(
			&like.StreamID,
			&like.UserID,
			&like.CreatedAt,
		)

		if err != nil {
			return nil, 0, err
		}

		likes = append(
			likes,
			like,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return likes, total, nil
}

//
// Get Stream Owner
//
// streams -> channels -> channels.user_id
//

func (r *repository) GetStreamOwner(
	ctx context.Context,
	streamID uuid.UUID,
) (*StreamOwner, error) {

	const query = `
		SELECT
			s.id,
			c.id,
			c.user_id
		FROM streams s
		INNER JOIN channels c
			ON c.id = s.channel_id
		WHERE s.id = $1
	`

	var owner StreamOwner

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

		return nil, err
	}

	return &owner, nil
}

//
// Get Like Statistics
//

func (r *repository) GetLikeStats(
	ctx context.Context,
	streamID uuid.UUID,
) (*LikeStats, error) {

	const query = `
		SELECT COUNT(*)
		FROM stream_likes
		WHERE stream_id = $1
	`

	var likeCount int64

	err := r.db.QueryRow(
		ctx,
		query,
		streamID,
	).Scan(&likeCount)

	if err != nil {
		return nil, err
	}

	return &LikeStats{
		StreamID:  streamID,
		LikeCount: likeCount,
	}, nil
}
