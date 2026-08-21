package follow

import "time"

// ============================================================
// Channel Follow
// ============================================================

type ChannelFollow struct {
	ID string `json:"id"`

	UserID    string `json:"user_id"`
	ChannelID string `json:"channel_id"`

	CreatedAt time.Time `json:"created_at"`
}

// ============================================================
// Follow Status
// ============================================================

type FollowStatus struct {
	Following bool `json:"following"`
}

// ============================================================
// Follower Count
// ============================================================

type FollowerCount struct {
	Count int64 `json:"follower_count"`
}
