package topic

import "time"

// ============================================================
// Topic
// ============================================================

type Topic struct {
	ID string `json:"id"`

	CategoryID string `json:"category_id"`

	Name string `json:"name"`

	Slug string `json:"slug"`

	Description string `json:"description,omitempty"`

	StreamCount int64 `json:"stream_count"`

	CreatedAt time.Time `json:"created_at"`

	UpdatedAt time.Time `json:"updated_at"`
}

// ============================================================
// Create Topic
// ============================================================

type CreateTopicRequest struct {
	CategoryID string `json:"category_id" binding:"required"`

	Name string `json:"name" binding:"required,min=2,max=100"`

	Description string `json:"description,omitempty" binding:"omitempty,max=500"`
}

// ============================================================
// Update Topic
// ============================================================

type UpdateTopicRequest struct {
	Name *string `json:"name,omitempty" binding:"omitempty,min=2,max=100"`

	Description *string `json:"description,omitempty" binding:"omitempty,max=500"`
}
