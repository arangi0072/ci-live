package user

import "time"

// ============================================================
// User Profile
// ============================================================

type UserProfile struct {
	UserID string `json:"user_id"`

	Username    string `json:"username"`
	DisplayName string `json:"display_name"`

	Bio       string `json:"bio,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ============================================================
// Create Profile
// ============================================================

type CreateProfileRequest struct {
	Username    string `json:"username" binding:"required,min=3,max=50"`
	DisplayName string `json:"display_name" binding:"required,min=1,max=100"`
	Bio         string `json:"bio,omitempty" binding:"max=500"`
	AvatarURL   string `json:"avatar_url,omitempty" binding:"omitempty,max=2048"`
}

// ============================================================
// Update Profile
// ============================================================

type UpdateProfileRequest struct {
	Username    *string `json:"username,omitempty" binding:"omitempty,min=3,max=50"`
	DisplayName *string `json:"display_name,omitempty" binding:"omitempty,min=1,max=100"`
	Bio         *string `json:"bio,omitempty" binding:"omitempty,max=500"`
	AvatarURL   *string `json:"avatar_url,omitempty" binding:"omitempty,max=2048"`
}

// ============================================================
// Public Profile Response
// ============================================================

type ProfileResponse struct {
	UserID string `json:"user_id"`

	Username    string `json:"username"`
	DisplayName string `json:"display_name"`

	Bio       string `json:"bio,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`

	CreatedAt time.Time `json:"created_at"`
}

// ============================================================
// Username Availability
// ============================================================

type UsernameAvailabilityResponse struct {
	Username  string `json:"username"`
	Available bool   `json:"available"`
}
