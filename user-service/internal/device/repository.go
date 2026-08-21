package device

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type deviceRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &deviceRepository{
		db: db,
	}
}

// Make sure implementation satisfies Repository.
var _ Repository = (*deviceRepository)(nil)

// ============================================================
// Create
// ============================================================

func (r *deviceRepository) Create(
	ctx context.Context,
	device *Device,
) error {
	const query = `
		INSERT INTO user_devices (
			id,
			user_id,
			device_id,
			platform,
			fcm_token,
			app_version,
			is_active,
			last_seen_at,
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
			NOW(),
			NOW(),
			NOW()
		)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		device.ID,
		device.UserID,
		device.DeviceID,
		device.Platform,
		device.FCMToken,
		device.AppVersion,
		device.IsActive,
	)

	if err != nil {
		return fmt.Errorf(
			"create device: %w",
			err,
		)
	}

	return nil
}

// ============================================================
// Get by user + device ID
// ============================================================

func (r *deviceRepository) GetByUserAndDeviceID(
	ctx context.Context,
	userID string,
	deviceID string,
) (*Device, error) {
	const query = `
		SELECT
			id,
			user_id,
			device_id,
			platform,
			fcm_token,
			app_version,
			is_active,
			last_seen_at,
			created_at,
			updated_at
		FROM user_devices
		WHERE user_id = $1
		  AND device_id = $2
		LIMIT 1
	`

	var device Device

	err := r.db.QueryRowContext(
		ctx,
		query,
		userID,
		deviceID,
	).Scan(
		&device.ID,
		&device.UserID,
		&device.DeviceID,
		&device.Platform,
		&device.FCMToken,
		&device.AppVersion,
		&device.IsActive,
		&device.LastSeenAt,
		&device.CreatedAt,
		&device.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDeviceNotFound
		}

		return nil, fmt.Errorf(
			"get device: %w",
			err,
		)
	}

	return &device, nil
}

// ============================================================
// List user's devices
// ============================================================

func (r *deviceRepository) ListByUser(
	ctx context.Context,
	userID string,
) ([]Device, error) {
	const query = `
		SELECT
			id,
			user_id,
			device_id,
			platform,
			fcm_token,
			app_version,
			is_active,
			last_seen_at,
			created_at,
			updated_at
		FROM user_devices
		WHERE user_id = $1
		ORDER BY last_seen_at DESC NULLS LAST,
		         created_at DESC
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		userID,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"list devices: %w",
		)
	}

	defer rows.Close()

	devices := make(
		[]Device,
		0,
	)

	for rows.Next() {
		var device Device

		if err := rows.Scan(
			&device.ID,
			&device.UserID,
			&device.DeviceID,
			&device.Platform,
			&device.FCMToken,
			&device.AppVersion,
			&device.IsActive,
			&device.LastSeenAt,
			&device.CreatedAt,
			&device.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf(
				"scan device: %w",
				err,
			)
		}

		devices = append(
			devices,
			device,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate devices: %w",
			err,
		)
	}

	return devices, nil
}

// ============================================================
// Update
// ============================================================

func (r *deviceRepository) Update(
	ctx context.Context,
	userID string,
	deviceID string,
	req UpdateDeviceRequest,
) (*Device, error) {
	const query = `
		UPDATE user_devices
		SET
			fcm_token = COALESCE($1, fcm_token),
			app_version = COALESCE($2, app_version),
			is_active = COALESCE($3, is_active),
			last_seen_at = NOW(),
			updated_at = NOW()
		WHERE user_id = $4
		  AND device_id = $5
		RETURNING
			id,
			user_id,
			device_id,
			platform,
			fcm_token,
			app_version,
			is_active,
			last_seen_at,
			created_at,
			updated_at
	`

	var device Device

	var fcmToken interface{}
	var appVersion interface{}
	var isActive interface{}

	if req.FCMToken != nil {
		fcmToken = *req.FCMToken
	}

	if req.AppVersion != nil {
		appVersion = *req.AppVersion
	}

	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	err := r.db.QueryRowContext(
		ctx,
		query,
		fcmToken,
		appVersion,
		isActive,
		userID,
		deviceID,
	).Scan(
		&device.ID,
		&device.UserID,
		&device.DeviceID,
		&device.Platform,
		&device.FCMToken,
		&device.AppVersion,
		&device.IsActive,
		&device.LastSeenAt,
		&device.CreatedAt,
		&device.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDeviceNotFound
		}

		return nil, fmt.Errorf(
			"update device: %w",
			err,
		)
	}

	return &device, nil
}

// ============================================================
// Delete
// ============================================================

func (r *deviceRepository) Delete(
	ctx context.Context,
	userID string,
	deviceID string,
) error {
	const query = `
		DELETE FROM user_devices
		WHERE user_id = $1
		  AND device_id = $2
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		userID,
		deviceID,
	)

	if err != nil {
		return fmt.Errorf(
			"delete device: %w",
			err,
		)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"get deleted rows: %w",
			err,
		)
	}

	if rows == 0 {
		return ErrDeviceNotFound
	}

	return nil
}

// ============================================================
// Optional: Update last seen
// ============================================================

func (r *deviceRepository) UpdateLastSeen(
	ctx context.Context,
	userID string,
	deviceID string,
) error {
	const query = `
		UPDATE user_devices
		SET
			last_seen_at = NOW(),
			updated_at = NOW()
		WHERE user_id = $1
		  AND device_id = $2
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		userID,
		deviceID,
	)

	if err != nil {
		return fmt.Errorf(
			"update device last seen: %w",
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
		return ErrDeviceNotFound
	}

	return nil
}
