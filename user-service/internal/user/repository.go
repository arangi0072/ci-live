package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type userRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &userRepository{
		db: db,
	}
}

// Compile-time interface check.
var _ Repository = (*userRepository)(nil)

// ============================================================
// Create Profile
// ============================================================

func (r *userRepository) Create(
	ctx context.Context,
	profile *UserProfile,
) error {
	const query = `
		INSERT INTO user_profiles (
			user_id,
			username,
			display_name,
			bio,
			avatar_url,
			created_at,
			updated_at
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			NOW(),
			NOW()
		)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		profile.UserID,
		profile.Username,
		profile.DisplayName,
		profile.Bio,
		profile.AvatarURL,
	)

	if err != nil {
		return fmt.Errorf(
			"create user profile: %w",
			err,
		)
	}

	return nil
}

// ============================================================
// Get Profile by User ID
// ============================================================

func (r *userRepository) GetByUserID(
	ctx context.Context,
	userID string,
) (*UserProfile, error) {
	const query = `
		SELECT
			user_id,
			username,
			display_name,
			bio,
			avatar_url,
			created_at,
			updated_at
		FROM user_profiles
		WHERE user_id = $1
	`

	var profile UserProfile

	err := r.db.QueryRowContext(
		ctx,
		query,
		userID,
	).Scan(
		&profile.UserID,
		&profile.Username,
		&profile.DisplayName,
		&profile.Bio,
		&profile.AvatarURL,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrProfileNotFound
		}

		return nil, fmt.Errorf(
			"get user profile: %w",
			err,
		)
	}

	return &profile, nil
}

// ============================================================
// Update Profile
// ============================================================

func (r *userRepository) Update(
	ctx context.Context,
	userID string,
	req UpdateProfileRequest,
) (*UserProfile, error) {
	const query = `
		UPDATE user_profiles
		SET
			username = COALESCE($1, username),
			display_name = COALESCE($2, display_name),
			bio = COALESCE($3, bio),
			avatar_url = COALESCE($4, avatar_url),
			updated_at = NOW()
		WHERE user_id = $5
		RETURNING
			user_id,
			username,
			display_name,
			bio,
			avatar_url,
			created_at,
			updated_at
	`

	var username interface{}
	var displayName interface{}
	var bio interface{}
	var avatarURL interface{}

	if req.Username != nil {
		username = *req.Username
	}

	if req.DisplayName != nil {
		displayName = *req.DisplayName
	}

	if req.Bio != nil {
		bio = *req.Bio
	}

	if req.AvatarURL != nil {
		avatarURL = *req.AvatarURL
	}

	var profile UserProfile

	err := r.db.QueryRowContext(
		ctx,
		query,
		username,
		displayName,
		bio,
		avatarURL,
		userID,
	).Scan(
		&profile.UserID,
		&profile.Username,
		&profile.DisplayName,
		&profile.Bio,
		&profile.AvatarURL,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrProfileNotFound
		}

		return nil, fmt.Errorf(
			"update user profile: %w",
			err,
		)
	}

	return &profile, nil
}

// ============================================================
// Username Exists
// ============================================================

func (r *userRepository) UsernameExists(
	ctx context.Context,
	username string,
) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM user_profiles
			WHERE username = $1
		)
	`

	var exists bool

	err := r.db.QueryRowContext(
		ctx,
		query,
		username,
	).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf(
			"check username existence: %w",
			err,
		)
	}

	return exists, nil
}
