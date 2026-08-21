package channel

import "time"

// ============================================================
// Channel
// ============================================================

type Channel struct {
	ID string `json:"id"`

	UserID string `json:"user_id"`

	Username string `json:"username"`
	Name     string `json:"name"`

	Description string `json:"description,omitempty"`

	AvatarURL string `json:"avatar_url,omitempty"`
	BannerURL string `json:"banner_url,omitempty"`

	IsVerified bool `json:"is_verified"`

	FollowerCount int64 `json:"follower_count"`
	TotalViews    int64 `json:"total_views"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ============================================================
// Create Channel
// ============================================================

type CreateChannelRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`

	Name string `json:"name" binding:"required,min=1,max=100"`

	Description string `json:"description,omitempty" binding:"omitempty,max=1000"`

	AvatarURL string `json:"avatar_url,omitempty" binding:"omitempty,max=2048"`

	BannerURL string `json:"banner_url,omitempty" binding:"omitempty,max=2048"`
}

// ============================================================
// Update Channel
// ============================================================

type UpdateChannelRequest struct {
	Name *string `json:"name,omitempty" binding:"omitempty,min=1,max=100"`

	Description *string `json:"description,omitempty" binding:"omitempty,max=1000"`

	AvatarURL *string `json:"avatar_url,omitempty" binding:"omitempty,max=2048"`

	BannerURL *string `json:"banner_url,omitempty" binding:"omitempty,max=2048"`
}
