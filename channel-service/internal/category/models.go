package category

import "time"

// ============================================================
// Category
// ============================================================

type Category struct {
	ID string `json:"id"`

	Name string `json:"name"`

	Slug string `json:"slug"`

	Description string `json:"description,omitempty"`

	IconURL string `json:"icon_url,omitempty"`

	StreamCount int64 `json:"stream_count"`

	CreatedAt time.Time `json:"created_at"`

	UpdatedAt time.Time `json:"updated_at"`
}

// ============================================================
// Create Category
// ============================================================

type CreateCategoryRequest struct {
	Name string `json:"name" binding:"required,min=2,max=100"`

	Description string `json:"description,omitempty" binding:"omitempty,max=500"`

	IconURL string `json:"icon_url,omitempty" binding:"omitempty,max=2048"`
}

// ============================================================
// Update Category
// ============================================================

type UpdateCategoryRequest struct {
	Name *string `json:"name,omitempty" binding:"omitempty,min=2,max=100"`

	Description *string `json:"description,omitempty" binding:"omitempty,max=500"`

	IconURL *string `json:"icon_url,omitempty" binding:"omitempty,max=2048"`
}
