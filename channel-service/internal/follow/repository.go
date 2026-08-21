package follow

import (
	"context"
	"database/sql"
	"fmt"
)

type followRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &followRepository{
		db: db,
	}
}

// Compile-time interface check.
var _ Repository = (*followRepository)(nil)

// ============================================================
// Create Follow
// ============================================================

func (r *followRepository) Create(
	ctx context.Context,
	follow *ChannelFollow,
) error {
	const query = `
		INSERT INTO channel_follows (
			id,
			user_id,
			channel_id,
			created_at
		)
		VALUES (
			$1,
			$2,
			$3,
			NOW()
		)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		follow.ID,
		follow.UserID,
		follow.ChannelID,
	)

	if err != nil {
		return fmt.Errorf(
			"create channel follow: %w",
			err,
		)
	}

	return nil
}

// ============================================================
// Delete Follow
// ============================================================

func (r *followRepository) Delete(
	ctx context.Context,
	userID string,
	channelID string,
) error {
	const query = `
		DELETE FROM channel_follows
		WHERE user_id = $1
		  AND channel_id = $2
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		userID,
		channelID,
	)

	if err != nil {
		return fmt.Errorf(
			"delete channel follow: %w",
			err,
		)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"get deleted follow rows: %w",
			err,
		)
	}

	if rows == 0 {
		return ErrFollowNotFound
	}

	return nil
}

// ============================================================
// Check Follow Exists
// ============================================================

func (r *followRepository) Exists(
	ctx context.Context,
	userID string,
	channelID string,
) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM channel_follows
			WHERE user_id = $1
			  AND channel_id = $2
		)
	`

	var exists bool

	err := r.db.QueryRowContext(
		ctx,
		query,
		userID,
		channelID,
	).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf(
			"check channel follow: %w",
			err,
		)
	}

	return exists, nil
}

// ============================================================
// Count Followers
// ============================================================

func (r *followRepository) CountFollowers(
	ctx context.Context,
	channelID string,
) (int64, error) {
	const query = `
		SELECT COUNT(*)
		FROM channel_follows
		WHERE channel_id = $1
	`

	var count int64

	err := r.db.QueryRowContext(
		ctx,
		query,
		channelID,
	).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf(
			"count channel followers: %w",
			err,
		)
	}

	return count, nil
}

// ============================================================
// Get Following
// ============================================================

func (r *followRepository) GetFollowing(
	ctx context.Context,
	userID string,
) ([]ChannelFollow, error) {
	const query = `
		SELECT
			id,
			user_id,
			channel_id,
			created_at
		FROM channel_follows
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		userID,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"get following: %w",
			err,
		)
	}

	defer rows.Close()

	follows := make(
		[]ChannelFollow,
		0,
	)

	for rows.Next() {
		var follow ChannelFollow

		if err := rows.Scan(
			&follow.ID,
			&follow.UserID,
			&follow.ChannelID,
			&follow.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf(
				"scan channel follow: %w",
				err,
			)
		}

		follows = append(
			follows,
			follow,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate channel follows: %w",
			err,
		)
	}

	return follows, nil
}
