package stream

import (
	"time"

	"github.com/google/uuid"
)

//
// Stream Status
//

type StreamStatus string

const (
	StreamStatusCreated      StreamStatus = "CREATED"
	StreamStatusStarting     StreamStatus = "STARTING"
	StreamStatusLive         StreamStatus = "LIVE"
	StreamStatusReconnecting StreamStatus = "RECONNECTING"
	StreamStatusEnding       StreamStatus = "ENDING"
	StreamStatusEnded        StreamStatus = "ENDED"
	StreamStatusFailed       StreamStatus = "FAILED"
	StreamStatusCancelled    StreamStatus = "CANCELLED"
)

//
// Stream Visibility
//

type StreamVisibility string

const (
	StreamVisibilityPublic   StreamVisibility = "PUBLIC"
	StreamVisibilityUnlisted StreamVisibility = "UNLISTED"
	StreamVisibilityPrivate  StreamVisibility = "PRIVATE"
)

//
// Stream
//

type Stream struct {
	ID uuid.UUID `json:"id" db:"id"`

	ChannelID uuid.UUID `json:"channel_id" db:"channel_id"`

	Title string `json:"title" db:"title"`

	Description *string `json:"description,omitempty" db:"description"`

	ThumbnailURL *string `json:"thumbnail_url,omitempty" db:"thumbnail_url"`

	CategoryID *uuid.UUID `json:"category_id,omitempty" db:"category_id"`

	Status StreamStatus `json:"status" db:"status"`

	Visibility StreamVisibility `json:"visibility" db:"visibility"`

	ViewerCount int `json:"viewer_count" db:"viewer_count"`

	PeakViewers int `json:"peak_viewers" db:"peak_viewers"`

	LikeCount int64 `json:"like_count" db:"like_count"`

	TotalViews int64 `json:"total_views" db:"total_views"`

	StartedAt *time.Time `json:"started_at,omitempty" db:"started_at"`

	EndedAt *time.Time `json:"ended_at,omitempty" db:"ended_at"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`

	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

//
// Create Stream Request
//

type CreateStreamRequest struct {
	Title string `json:"title" binding:"required,max=200"`

	Description *string `json:"description,omitempty"`

	ThumbnailURL *string `json:"thumbnail_url,omitempty"`

	CategoryID *uuid.UUID `json:"category_id,omitempty"`

	Visibility StreamVisibility `json:"visibility,omitempty"`
}

//
// Update Stream Request
//

type UpdateStreamRequest struct {
	Title *string `json:"title,omitempty"`

	Description *string `json:"description,omitempty"`

	ThumbnailURL *string `json:"thumbnail_url,omitempty"`

	CategoryID *uuid.UUID `json:"category_id,omitempty"`

	Visibility *StreamVisibility `json:"visibility,omitempty"`
}

//
// Stream Response
//

type StreamResponse struct {
	ID uuid.UUID `json:"id"`

	ChannelID uuid.UUID `json:"channel_id"`

	Title string `json:"title"`

	Description *string `json:"description,omitempty"`

	ThumbnailURL *string `json:"thumbnail_url,omitempty"`

	CategoryID *uuid.UUID `json:"category_id,omitempty"`

	Status StreamStatus `json:"status"`

	Visibility StreamVisibility `json:"visibility"`

	ViewerCount int `json:"viewer_count"`

	PeakViewers int `json:"peak_viewers"`

	LikeCount int64 `json:"like_count"`

	TotalViews int64 `json:"total_views"`

	StartedAt *time.Time `json:"started_at,omitempty"`

	EndedAt *time.Time `json:"ended_at,omitempty"`

	CreatedAt time.Time `json:"created_at"`

	UpdatedAt time.Time `json:"updated_at"`
}

//
// Stream List Response
//

type StreamListResponse struct {
	Streams []StreamResponse `json:"streams"`

	Total int64 `json:"total"`

	Page int `json:"page"`

	Limit int `json:"limit"`
}

//
// Stream Filter
//

type StreamFilter struct {
	ChannelID *uuid.UUID

	CategoryID *uuid.UUID

	Status *StreamStatus

	Visibility *StreamVisibility

	Page int

	Limit int
}

//
// Stream Ownership
//
// The streams table does NOT contain user_id/owner_id.
// Ownership is resolved through:
//
// streams.channel_id -> channels.id -> channels.user_id
//

type StreamOwner struct {
	StreamID uuid.UUID `db:"stream_id"`

	ChannelID uuid.UUID `db:"channel_id"`

	UserID uuid.UUID `db:"user_id"`
}

//
// Stream Statistics
//

type StreamStats struct {
	StreamID uuid.UUID `json:"stream_id"`

	ViewerCount int `json:"viewer_count"`

	PeakViewers int `json:"peak_viewers"`

	LikeCount int64 `json:"like_count"`

	TotalViews int64 `json:"total_views"`
}

//
// Stream Status Update
//

type UpdateStreamStatusRequest struct {
	Status StreamStatus `json:"status" binding:"required"`
}

//
// Start Stream Request
//

type StartStreamRequest struct {
	StreamID uuid.UUID `json:"stream_id" binding:"required"`
}

//
// End Stream Request
//

type EndStreamRequest struct {
	StreamID uuid.UUID `json:"stream_id" binding:"required"`
}

//
// Stream Ownership Check
//

type StreamOwnership struct {
	StreamID uuid.UUID `json:"stream_id"`

	ChannelID uuid.UUID `json:"channel_id"`

	UserID uuid.UUID `json:"user_id"`
}

//
// Public Stream
//
// Used when returning stream information to viewers.
//

type PublicStreamResponse struct {
	ID uuid.UUID `json:"id"`

	ChannelID uuid.UUID `json:"channel_id"`

	Title string `json:"title"`

	Description *string `json:"description,omitempty"`

	ThumbnailURL *string `json:"thumbnail_url,omitempty"`

	CategoryID *uuid.UUID `json:"category_id,omitempty"`

	Status StreamStatus `json:"status"`

	Visibility StreamVisibility `json:"visibility"`

	ViewerCount int `json:"viewer_count"`

	PeakViewers int `json:"peak_viewers"`

	LikeCount int64 `json:"like_count"`

	TotalViews int64 `json:"total_views"`

	StartedAt *time.Time `json:"started_at,omitempty"`

	CreatedAt time.Time `json:"created_at"`
}

//
// Stream Summary
//
// Useful for feeds/discovery pages.
//

type StreamSummary struct {
	ID uuid.UUID `json:"id"`

	ChannelID uuid.UUID `json:"channel_id"`

	Title string `json:"title"`

	ThumbnailURL *string `json:"thumbnail_url,omitempty"`

	CategoryID *uuid.UUID `json:"category_id,omitempty"`

	Status StreamStatus `json:"status"`

	Visibility StreamVisibility `json:"visibility"`

	ViewerCount int `json:"viewer_count"`

	LikeCount int64 `json:"like_count"`

	StartedAt *time.Time `json:"started_at,omitempty"`
}

//
// Validation
//

func (s StreamStatus) IsValid() bool {
	switch s {
	case StreamStatusCreated,
		StreamStatusStarting,
		StreamStatusLive,
		StreamStatusReconnecting,
		StreamStatusEnding,
		StreamStatusEnded,
		StreamStatusFailed,
		StreamStatusCancelled:
		return true

	default:
		return false
	}
}

func (v StreamVisibility) IsValid() bool {
	switch v {
	case StreamVisibilityPublic,
		StreamVisibilityUnlisted,
		StreamVisibilityPrivate:
		return true

	default:
		return false
	}
}
