package session

import (
	"time"

	"github.com/google/uuid"
)

//
// Stream Session
//

type StreamSession struct {
	ID uuid.UUID `json:"id" db:"id"`

	StreamID uuid.UUID `json:"stream_id" db:"stream_id"`

	MediaServerID string `json:"media_server_id" db:"media_server_id"`

	MediaPath string `json:"media_path" db:"media_path"`

	StartedAt *time.Time `json:"started_at,omitempty" db:"started_at"`

	EndedAt *time.Time `json:"ended_at,omitempty" db:"ended_at"`

	DurationSeconds int64 `json:"duration_seconds" db:"duration_seconds"`

	PeakViewers int `json:"peak_viewers" db:"peak_viewers"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

//
// Create Session Request
//

type CreateSessionRequest struct {
	StreamID uuid.UUID `json:"stream_id" binding:"required"`

	MediaServerID string `json:"media_server_id" binding:"required"`

	MediaPath string `json:"media_path" binding:"required"`

	StartedAt *time.Time `json:"started_at,omitempty"`
}

//
// Update Session Request
//

type UpdateSessionRequest struct {
	MediaServerID *string `json:"media_server_id,omitempty"`

	MediaPath *string `json:"media_path,omitempty"`

	StartedAt *time.Time `json:"started_at,omitempty"`

	EndedAt *time.Time `json:"ended_at,omitempty"`

	DurationSeconds *int64 `json:"duration_seconds,omitempty"`

	PeakViewers *int `json:"peak_viewers,omitempty"`
}

//
// End Session Request
//

type EndSessionRequest struct {
	EndedAt *time.Time `json:"ended_at,omitempty"`

	DurationSeconds *int64 `json:"duration_seconds,omitempty"`
}

//
// Session Response
//

type SessionResponse struct {
	ID uuid.UUID `json:"id"`

	StreamID uuid.UUID `json:"stream_id"`

	MediaServerID string `json:"media_server_id"`

	MediaPath string `json:"media_path"`

	StartedAt *time.Time `json:"started_at,omitempty"`

	EndedAt *time.Time `json:"ended_at,omitempty"`

	DurationSeconds int64 `json:"duration_seconds"`

	PeakViewers int `json:"peak_viewers"`

	CreatedAt time.Time `json:"created_at"`
}

//
// Session List Response
//

type SessionListResponse struct {
	Sessions []SessionResponse `json:"sessions"`

	Total int64 `json:"total"`

	Page int `json:"page"`

	Limit int `json:"limit"`
}

//
// Session Filter
//

type SessionFilter struct {
	StreamID *uuid.UUID

	Page int

	Limit int
}

//
// Session Statistics
//

type SessionStats struct {
	SessionID uuid.UUID `json:"session_id"`

	StreamID uuid.UUID `json:"stream_id"`

	MediaServerID string `json:"media_server_id"`

	MediaPath string `json:"media_path"`

	StartedAt *time.Time `json:"started_at,omitempty"`

	EndedAt *time.Time `json:"ended_at,omitempty"`

	DurationSeconds int64 `json:"duration_seconds"`

	PeakViewers int `json:"peak_viewers"`
}

//
// Session Ownership
//
// stream_sessions
//      ↓ stream_id
// streams
//      ↓ channel_id
// channels
//      ↓ user_id
//

type SessionOwner struct {
	SessionID uuid.UUID `db:"session_id"`

	StreamID uuid.UUID `db:"stream_id"`

	ChannelID uuid.UUID `db:"channel_id"`

	UserID uuid.UUID `db:"user_id"`
}

//
// Media Server Session
//

type MediaServerSession struct {
	SessionID uuid.UUID `json:"session_id"`

	StreamID uuid.UUID `json:"stream_id"`

	MediaServerID string `json:"media_server_id"`

	MediaPath string `json:"media_path"`

	StartedAt *time.Time `json:"started_at,omitempty"`

	EndedAt *time.Time `json:"ended_at,omitempty"`
}

//
// Duration
//

func (s *StreamSession) Duration() time.Duration {
	if s == nil || s.StartedAt == nil {
		return 0
	}

	end := time.Now().UTC()

	if s.EndedAt != nil {
		end = *s.EndedAt
	}

	if end.Before(*s.StartedAt) {
		return 0
	}

	return end.Sub(*s.StartedAt)
}

//
// Response Conversion
//

func ToSessionResponse(
	s *StreamSession,
) *SessionResponse {

	if s == nil {
		return nil
	}

	return &SessionResponse{
		ID:              s.ID,
		StreamID:        s.StreamID,
		MediaServerID:   s.MediaServerID,
		MediaPath:       s.MediaPath,
		StartedAt:       s.StartedAt,
		EndedAt:         s.EndedAt,
		DurationSeconds: s.DurationSeconds,
		PeakViewers:     s.PeakViewers,
		CreatedAt:       s.CreatedAt,
	}
}
