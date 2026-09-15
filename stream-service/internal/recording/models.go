package recording

import (
	"time"

	"github.com/google/uuid"
)

//
// Recording Status
//

type RecordingStatus string

const (
	RecordingStatusProcessing RecordingStatus = "PROCESSING"
	RecordingStatusReady      RecordingStatus = "READY"
	RecordingStatusFailed     RecordingStatus = "FAILED"
	RecordingStatusDeleted    RecordingStatus = "DELETED"
)

func (s RecordingStatus) IsValid() bool {
	switch s {
	case RecordingStatusProcessing,
		RecordingStatusReady,
		RecordingStatusFailed,
		RecordingStatusDeleted:
		return true
	default:
		return false
	}
}

//
// Storage Provider
//

type StorageProvider string

const (
	StorageProviderMinIO StorageProvider = "MINIO"
)

func (s StorageProvider) IsValid() bool {
	return s == StorageProviderMinIO
}

//
// Stream Recording
//

type StreamRecording struct {
	ID uuid.UUID `json:"id" db:"id"`

	StreamID uuid.UUID `json:"stream_id" db:"stream_id"`

	SessionID *uuid.UUID `json:"session_id,omitempty" db:"session_id"`

	StorageProvider StorageProvider `json:"storage_provider" db:"storage_provider"`

	StorageKey string `json:"storage_key" db:"storage_key"`

	PlaybackURL *string `json:"playback_url,omitempty" db:"playback_url"`

	ThumbnailURL *string `json:"thumbnail_url,omitempty" db:"thumbnail_url"`

	Format *string `json:"format,omitempty" db:"format"`

	Width *int `json:"width,omitempty" db:"width"`

	Height *int `json:"height,omitempty" db:"height"`

	FileSize *int64 `json:"file_size,omitempty" db:"file_size"`

	DurationSeconds *int64 `json:"duration_seconds,omitempty" db:"duration_seconds"`

	ViewCount int64 `json:"view_count" db:"view_count"`

	Status RecordingStatus `json:"status" db:"status"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`

	CompletedAt *time.Time `json:"completed_at,omitempty" db:"completed_at"`
}

//
// Create Recording Request
//

type CreateRecordingRequest struct {
	StreamID uuid.UUID `json:"stream_id" binding:"required"`

	SessionID *uuid.UUID `json:"session_id,omitempty"`

	StorageProvider StorageProvider `json:"storage_provider,omitempty"`

	StorageKey string `json:"storage_key" binding:"required"`

	PlaybackURL *string `json:"playback_url,omitempty"`

	ThumbnailURL *string `json:"thumbnail_url,omitempty"`

	Format *string `json:"format,omitempty"`

	Width *int `json:"width,omitempty"`

	Height *int `json:"height,omitempty"`

	FileSize *int64 `json:"file_size,omitempty"`

	DurationSeconds *int64 `json:"duration_seconds,omitempty"`
}

//
// Update Recording Request
//

type UpdateRecordingRequest struct {
	PlaybackURL *string `json:"playback_url,omitempty"`

	ThumbnailURL *string `json:"thumbnail_url,omitempty"`

	Format *string `json:"format,omitempty"`

	Width *int `json:"width,omitempty"`

	Height *int `json:"height,omitempty"`

	FileSize *int64 `json:"file_size,omitempty"`

	DurationSeconds *int64 `json:"duration_seconds,omitempty"`

	Status *RecordingStatus `json:"status,omitempty"`

	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

//
// Update Recording Status Request
//

type UpdateRecordingStatusRequest struct {
	Status RecordingStatus `json:"status" binding:"required"`
}

//
// Recording Response
//

type RecordingResponse struct {
	ID uuid.UUID `json:"id"`

	StreamID uuid.UUID `json:"stream_id"`

	SessionID *uuid.UUID `json:"session_id,omitempty"`

	StorageProvider StorageProvider `json:"storage_provider"`

	StorageKey string `json:"storage_key"`

	PlaybackURL *string `json:"playback_url,omitempty"`

	ThumbnailURL *string `json:"thumbnail_url,omitempty"`

	Format *string `json:"format,omitempty"`

	Width *int `json:"width,omitempty"`

	Height *int `json:"height,omitempty"`

	FileSize *int64 `json:"file_size,omitempty"`

	DurationSeconds *int64 `json:"duration_seconds,omitempty"`

	ViewCount int64 `json:"view_count"`

	Status RecordingStatus `json:"status"`

	CreatedAt time.Time `json:"created_at"`

	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

//
// Recording List Response
//

type RecordingListResponse struct {
	Recordings []RecordingResponse `json:"recordings"`

	Total int64 `json:"total"`

	Page int `json:"page"`

	Limit int `json:"limit"`
}

//
// Recording Filter
//

type RecordingFilter struct {
	StreamID  *uuid.UUID
	SessionID *uuid.UUID
	Status    *RecordingStatus

	Page  int
	Limit int
}

//
// Recording Owner
//
// stream_recordings
//        ↓ stream_id
// streams
//        ↓ channel_id
// channels
//        ↓ user_id
//

type RecordingOwner struct {
	RecordingID uuid.UUID `db:"recording_id"`

	StreamID uuid.UUID `db:"stream_id"`

	ChannelID uuid.UUID `db:"channel_id"`

	UserID uuid.UUID `db:"user_id"`
}

//
// Recording Statistics
//

type RecordingStats struct {
	RecordingID uuid.UUID `json:"recording_id"`

	StreamID uuid.UUID `json:"stream_id"`

	SessionID *uuid.UUID `json:"session_id,omitempty"`

	Status RecordingStatus `json:"status"`

	FileSize *int64 `json:"file_size,omitempty"`

	DurationSeconds *int64 `json:"duration_seconds,omitempty"`

	ViewCount int64 `json:"view_count"`

	CreatedAt time.Time `json:"created_at"`

	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

//
// Duration
//

func (r *StreamRecording) Duration() time.Duration {
	if r == nil {
		return 0
	}

	if r.DurationSeconds != nil {
		return time.Duration(*r.DurationSeconds) * time.Second
	}

	return 0
}

//
// Response Conversion
//

func ToRecordingResponse(
	recording *StreamRecording,
) *RecordingResponse {

	if recording == nil {
		return nil
	}

	return &RecordingResponse{
		ID:              recording.ID,
		StreamID:        recording.StreamID,
		SessionID:       recording.SessionID,
		StorageProvider: recording.StorageProvider,
		StorageKey:      recording.StorageKey,
		PlaybackURL:     recording.PlaybackURL,
		ThumbnailURL:    recording.ThumbnailURL,
		Format:          recording.Format,
		Width:           recording.Width,
		Height:          recording.Height,
		FileSize:        recording.FileSize,
		DurationSeconds: recording.DurationSeconds,
		ViewCount:       recording.ViewCount,
		Status:          recording.Status,
		CreatedAt:       recording.CreatedAt,
		CompletedAt:     recording.CompletedAt,
	}
}

//
// List Conversion
//

func ToRecordingResponses(
	recordings []StreamRecording,
) []RecordingResponse {

	responses := make(
		[]RecordingResponse,
		0,
		len(recordings),
	)

	for i := range recordings {
		response := ToRecordingResponse(&recordings[i])

		if response != nil {
			responses = append(
				responses,
				*response,
			)
		}
	}

	return responses
}
