package device

import "time"

// ============================================================
// Device
// ============================================================

type Device struct {
	ID string `json:"id"`

	// User ID comes from the Auth Service JWT.
	UserID string `json:"user_id"`

	DeviceID string `json:"device_id"`

	Platform string `json:"platform"`

	FCMToken string `json:"-"`

	AppVersion string `json:"app_version,omitempty"`

	IsActive bool `json:"is_active"`

	LastSeenAt *time.Time `json:"last_seen_at,omitempty"`

	CreatedAt time.Time `json:"created_at"`

	UpdatedAt time.Time `json:"updated_at"`
}

// ============================================================
// Register Device
// ============================================================

type RegisterDeviceRequest struct {
	DeviceID string `json:"device_id" binding:"required,max=255"`

	Platform string `json:"platform" binding:"required,oneof=android ios web"`

	FCMToken string `json:"fcm_token" binding:"required,max=4096"`

	AppVersion string `json:"app_version" binding:"omitempty,max=30"`
}

// ============================================================
// Update Device
// ============================================================

type UpdateDeviceRequest struct {
	FCMToken *string `json:"fcm_token,omitempty"`

	AppVersion *string `json:"app_version,omitempty"`

	IsActive *bool `json:"is_active,omitempty"`
}

// ============================================================
// Device Response
// ============================================================

// Never expose the FCM token to the Android client through
// normal device-list responses.
type DeviceResponse struct {
	ID string `json:"id"`

	DeviceID string `json:"device_id"`

	Platform string `json:"platform"`

	AppVersion string `json:"app_version,omitempty"`

	IsActive bool `json:"is_active"`

	LastSeenAt *time.Time `json:"last_seen_at,omitempty"`

	CreatedAt time.Time `json:"created_at"`

	UpdatedAt time.Time `json:"updated_at"`
}

// ============================================================
// Device List Response
// ============================================================

type DeviceListResponse struct {
	Devices []DeviceResponse `json:"devices"`
	Total   int              `json:"total"`
}
