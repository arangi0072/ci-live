package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

// ============================================================
// Users
// ============================================================

func (r *Repository) CreateUser(
	ctx context.Context,
	user *User,
) error {
	const query = `
		INSERT INTO users (
			id,
			email,
			password_hash,
			email_verified,
			is_active,
			is_banned,
			last_login_at,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.EmailVerified,
		user.IsActive,
		user.IsBanned,
		user.LastLoginAt,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (r *Repository) GetUserByID(
	ctx context.Context,
	userID string,
) (*User, error) {
	const query = `
		SELECT
			id,
			email,
			password_hash,
			email_verified,
			is_active,
			is_banned,
			last_login_at,
			created_at,
			updated_at
		FROM users
		WHERE id = $1
	`

	var user User

	err := r.db.QueryRowContext(
		ctx,
		query,
		userID,
	).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.EmailVerified,
		&user.IsActive,
		&user.IsBanned,
		&user.LastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf(
			"get user by id: %w",
			err,
		)
	}

	return &user, nil
}

func (r *Repository) GetUserByEmail(
	ctx context.Context,
	email string,
) (*User, error) {
	const query = `
		SELECT
			id,
			email,
			password_hash,
			email_verified,
			is_active,
			is_banned,
			last_login_at,
			created_at,
			updated_at
		FROM users
		WHERE email = $1
	`

	var user User

	err := r.db.QueryRowContext(
		ctx,
		query,
		email,
	).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.EmailVerified,
		&user.IsActive,
		&user.IsBanned,
		&user.LastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf(
			"get user by email: %w",
			err,
		)
	}

	return &user, nil
}

func (r *Repository) UpdateLastLogin(
	ctx context.Context,
	userID string,
) error {
	const query = `
		UPDATE users
		SET
			last_login_at = NOW(),
			updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		userID,
	)

	if err != nil {
		return fmt.Errorf(
			"update last login: %w",
			err,
		)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"get affected rows: %w",
			err,
		)
	}

	if rows == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *Repository) MarkEmailVerified(
	ctx context.Context,
	userID string,
) error {
	const query = `
		UPDATE users
		SET
			email_verified = TRUE,
			updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		userID,
	)

	if err != nil {
		return fmt.Errorf(
			"mark email verified: %w",
			err,
		)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"get affected rows: %w",
			err,
		)
	}

	if rows == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *Repository) UpdatePassword(
	ctx context.Context,
	userID string,
	passwordHash string,
) error {
	const query = `
		UPDATE users
		SET
			password_hash = $1,
			updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		passwordHash,
		userID,
	)

	if err != nil {
		return fmt.Errorf(
			"update password: %w",
			err,
		)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"get affected rows: %w",
			err,
		)
	}

	if rows == 0 {
		return ErrUserNotFound
	}

	return nil
}

// ============================================================
// Refresh Tokens
// ============================================================

func (r *Repository) CreateRefreshToken(
	ctx context.Context,
	token *RefreshToken,
) error {
	const query = `
		INSERT INTO refresh_tokens (
			id,
			user_id,
			token_hash,
			device_id,
			expires_at,
			revoked_at,
			created_at,
			last_used_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.DeviceID,
		token.ExpiresAt,
		token.RevokedAt,
		token.CreatedAt,
		token.LastUsedAt,
	)

	if err != nil {
		return fmt.Errorf(
			"create refresh token: %w",
			err,
		)
	}

	return nil
}

func (r *Repository) GetRefreshTokenByHash(
	ctx context.Context,
	tokenHash string,
) (*RefreshToken, error) {
	const query = `
		SELECT
			id,
			user_id,
			token_hash,
			device_id,
			expires_at,
			revoked_at,
			created_at,
			last_used_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`

	var token RefreshToken

	err := r.db.QueryRowContext(
		ctx,
		query,
		tokenHash,
	).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.DeviceID,
		&token.ExpiresAt,
		&token.RevokedAt,
		&token.CreatedAt,
		&token.LastUsedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidToken
		}

		return nil, fmt.Errorf(
			"get refresh token: %w",
			err,
		)
	}

	return &token, nil
}

func (r *Repository) RevokeRefreshToken(
	ctx context.Context,
	tokenHash string,
) error {
	const query = `
		UPDATE refresh_tokens
		SET revoked_at = COALESCE(revoked_at, NOW())
		WHERE token_hash = $1
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		tokenHash,
	)

	if err != nil {
		return fmt.Errorf(
			"revoke refresh token: %w",
			err,
		)
	}

	return nil
}

func (r *Repository) RevokeAllRefreshTokens(
	ctx context.Context,
	userID string,
) error {
	const query = `
		UPDATE refresh_tokens
		SET revoked_at = COALESCE(revoked_at, NOW())
		WHERE user_id = $1
		  AND revoked_at IS NULL
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		userID,
	)

	if err != nil {
		return fmt.Errorf(
			"revoke all refresh tokens: %w",
			err,
		)
	}

	return nil
}

func (r *Repository) UpdateRefreshTokenLastUsed(
	ctx context.Context,
	tokenHash string,
) error {
	const query = `
		UPDATE refresh_tokens
		SET last_used_at = NOW()
		WHERE token_hash = $1
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		tokenHash,
	)

	if err != nil {
		return fmt.Errorf(
			"update refresh token last used: %w",
			err,
		)
	}

	return nil
}

// ============================================================
// Email Verification Tokens
// ============================================================

func (r *Repository) CreateEmailVerificationToken(
	ctx context.Context,
	token *EmailVerificationToken,
) error {
	const query = `
		INSERT INTO email_verification_tokens (
			id,
			user_id,
			token_hash,
			expires_at,
			used_at,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
		token.UsedAt,
		token.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf(
			"create email verification token: %w",
			err,
		)
	}

	return nil
}

func (r *Repository) GetEmailVerificationToken(
	ctx context.Context,
	tokenHash string,
) (*EmailVerificationToken, error) {
	const query = `
		SELECT
			id,
			user_id,
			token_hash,
			expires_at,
			used_at,
			created_at
		FROM email_verification_tokens
		WHERE token_hash = $1
	`

	var token EmailVerificationToken

	err := r.db.QueryRowContext(
		ctx,
		query,
		tokenHash,
	).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.UsedAt,
		&token.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidToken
		}

		return nil, fmt.Errorf(
			"get email verification token: %w",
			err,
		)
	}

	return &token, nil
}

func (r *Repository) MarkEmailVerificationTokenUsed(
	ctx context.Context,
	tokenHash string,
) error {
	const query = `
		UPDATE email_verification_tokens
		SET used_at = COALESCE(used_at, NOW())
		WHERE token_hash = $1
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		tokenHash,
	)

	if err != nil {
		return fmt.Errorf(
			"mark email verification token used: %w",
			err,
		)
	}

	return nil
}

// ============================================================
// Password Reset Tokens
// ============================================================

func (r *Repository) CreatePasswordResetToken(
	ctx context.Context,
	token *PasswordResetToken,
) error {
	const query = `
		INSERT INTO password_reset_tokens (
			id,
			user_id,
			token_hash,
			expires_at,
			used_at,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
		token.UsedAt,
		token.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf(
			"create password reset token: %w",
			err,
		)
	}

	return nil
}

func (r *Repository) GetPasswordResetToken(
	ctx context.Context,
	tokenHash string,
) (*PasswordResetToken, error) {
	const query = `
		SELECT
			id,
			user_id,
			token_hash,
			expires_at,
			used_at,
			created_at
		FROM password_reset_tokens
		WHERE token_hash = $1
	`

	var token PasswordResetToken

	err := r.db.QueryRowContext(
		ctx,
		query,
		tokenHash,
	).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.UsedAt,
		&token.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidResetToken
		}

		return nil, fmt.Errorf(
			"get password reset token: %w",
			err,
		)
	}

	return &token, nil
}

func (r *Repository) MarkPasswordResetTokenUsed(
	ctx context.Context,
	tokenHash string,
) error {
	const query = `
		UPDATE password_reset_tokens
		SET used_at = COALESCE(used_at, NOW())
		WHERE token_hash = $1
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		tokenHash,
	)

	if err != nil {
		return fmt.Errorf(
			"mark password reset token used: %w",
			err,
		)
	}

	return nil
}

// ============================================================
// Compile-time interface check
// ============================================================

var _ AuthRepository = (*Repository)(nil)