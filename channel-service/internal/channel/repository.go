package channel

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type channelRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &channelRepository{
		db: db,
	}
}

// Compile-time interface check.
var _ Repository = (*channelRepository)(nil)

// ============================================================
// Create Channel
// ============================================================

func (r *channelRepository) Create(
	ctx context.Context,
	channel *Channel,
) error {
	const query = `
		INSERT INTO channels (
			id,
			user_id,
			username,
			name,
			description,
			avatar_url,
			banner_url,
			is_verified,
			follower_count,
			total_views,
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
			NOW(),
			NOW()
		)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		channel.ID,
		channel.UserID,
		channel.Username,
		channel.Name,
		channel.Description,
		channel.AvatarURL,
		channel.BannerURL,
		channel.IsVerified,
		channel.FollowerCount,
		channel.TotalViews,
	)

	if err != nil {
		return fmt.Errorf(
			"create channel: %w",
			err,
		)
	}

	return nil
}

// ============================================================
// Get Channel By ID
// ============================================================

func (r *channelRepository) GetByID(
	ctx context.Context,
	id string,
) (*Channel, error) {
	const query = `
		SELECT
			id,
			user_id,
			username,
			name,
			description,
			avatar_url,
			banner_url,
			is_verified,
			follower_count,
			total_views,
			created_at,
			updated_at
		FROM channels
		WHERE id = $1
	`

	var channel Channel

	err := r.db.QueryRowContext(
		ctx,
		query,
		id,
	).Scan(
		&channel.ID,
		&channel.UserID,
		&channel.Username,
		&channel.Name,
		&channel.Description,
		&channel.AvatarURL,
		&channel.BannerURL,
		&channel.IsVerified,
		&channel.FollowerCount,
		&channel.TotalViews,
		&channel.CreatedAt,
		&channel.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrChannelNotFound
		}

		return nil, fmt.Errorf(
			"get channel by id: %w",
			err,
		)
	}

	return &channel, nil
}

// ============================================================
// Get Channel By User ID
// ============================================================

func (r *channelRepository) GetByUserID(
	ctx context.Context,
	userID string,
) (*Channel, error) {
	const query = `
		SELECT
			id,
			user_id,
			username,
			name,
			description,
			avatar_url,
			banner_url,
			is_verified,
			follower_count,
			total_views,
			created_at,
			updated_at
		FROM channels
		WHERE user_id = $1
	`

	var channel Channel

	err := r.db.QueryRowContext(
		ctx,
		query,
		userID,
	).Scan(
		&channel.ID,
		&channel.UserID,
		&channel.Username,
		&channel.Name,
		&channel.Description,
		&channel.AvatarURL,
		&channel.BannerURL,
		&channel.IsVerified,
		&channel.FollowerCount,
		&channel.TotalViews,
		&channel.CreatedAt,
		&channel.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrChannelNotFound
		}

		return nil, fmt.Errorf(
			"get channel by user id: %w",
			err,
		)
	}

	return &channel, nil
}

// ============================================================
// Get Channel By Username
// ============================================================

func (r *channelRepository) GetByUsername(
	ctx context.Context,
	username string,
) (*Channel, error) {
	const query = `
		SELECT
			id,
			user_id,
			username,
			name,
			description,
			avatar_url,
			banner_url,
			is_verified,
			follower_count,
			total_views,
			created_at,
			updated_at
		FROM channels
		WHERE username = $1
	`

	var channel Channel

	err := r.db.QueryRowContext(
		ctx,
		query,
		username,
	).Scan(
		&channel.ID,
		&channel.UserID,
		&channel.Username,
		&channel.Name,
		&channel.Description,
		&channel.AvatarURL,
		&channel.BannerURL,
		&channel.IsVerified,
		&channel.FollowerCount,
		&channel.TotalViews,
		&channel.CreatedAt,
		&channel.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrChannelNotFound
		}

		return nil, fmt.Errorf(
			"get channel by username: %w",
			err,
		)
	}

	return &channel, nil
}

// ============================================================
// Update Channel
// ============================================================

func (r *channelRepository) Update(
	ctx context.Context,
	id string,
	req UpdateChannelRequest,
) (*Channel, error) {
	const query = `
		UPDATE channels
		SET
			name = COALESCE($1, name),
			description = COALESCE($2, description),
			avatar_url = COALESCE($3, avatar_url),
			banner_url = COALESCE($4, banner_url),
			updated_at = NOW()
		WHERE id = $5
		RETURNING
			id,
			user_id,
			username,
			name,
			description,
			avatar_url,
			banner_url,
			is_verified,
			follower_count,
			total_views,
			created_at,
			updated_at
	`

	var name interface{}
	var description interface{}
	var avatarURL interface{}
	var bannerURL interface{}

	if req.Name != nil {
		name = *req.Name
	}

	if req.Description != nil {
		description = *req.Description
	}

	if req.AvatarURL != nil {
		avatarURL = *req.AvatarURL
	}

	if req.BannerURL != nil {
		bannerURL = *req.BannerURL
	}

	var channel Channel

	err := r.db.QueryRowContext(
		ctx,
		query,
		name,
		description,
		avatarURL,
		bannerURL,
		id,
	).Scan(
		&channel.ID,
		&channel.UserID,
		&channel.Username,
		&channel.Name,
		&channel.Description,
		&channel.AvatarURL,
		&channel.BannerURL,
		&channel.IsVerified,
		&channel.FollowerCount,
		&channel.TotalViews,
		&channel.CreatedAt,
		&channel.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrChannelNotFound
		}

		return nil, fmt.Errorf(
			"update channel: %w",
			err,
		)
	}

	return &channel, nil
}

// ============================================================
// Delete Channel
// ============================================================

func (r *channelRepository) Delete(
	ctx context.Context,
	id string,
) error {
	const query = `
		DELETE FROM channels
		WHERE id = $1
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		id,
	)

	if err != nil {
		return fmt.Errorf(
			"delete channel: %w",
			err,
		)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"get deleted channel rows: %w",
			err,
		)
	}

	if rows == 0 {
		return ErrChannelNotFound
	}

	return nil
}
