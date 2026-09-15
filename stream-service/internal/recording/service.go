package recording

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

//
// Errors
//

var (
	ErrUnauthorized           = errors.New("unauthorized")
	ErrForbidden              = errors.New("forbidden")
	ErrRecordingNotFound      = errors.New("recording not found")
	ErrStreamNotFound         = errors.New("stream not found")
	ErrSessionNotFound        = errors.New("session not found")
	ErrInvalidRecordingStatus = errors.New("invalid recording status")
	ErrInvalidUUID            = errors.New("invalid UUID")
	ErrInvalidStorageProvider = errors.New("invalid storage provider")
)

//
// Service Interface
//

type Service interface {
	CreateRecording(
		ctx context.Context,
		userID uuid.UUID,
		req CreateRecordingRequest,
	) (*RecordingResponse, error)

	GetRecording(
		ctx context.Context,
		recordingID uuid.UUID,
	) (*RecordingResponse, error)

	ListRecordings(
		ctx context.Context,
		filter RecordingFilter,
	) (*RecordingListResponse, error)

	UpdateRecording(
		ctx context.Context,
		userID uuid.UUID,
		recordingID uuid.UUID,
		req UpdateRecordingRequest,
	) (*RecordingResponse, error)

	UpdateRecordingStatus(
		ctx context.Context,
		userID uuid.UUID,
		recordingID uuid.UUID,
		req UpdateRecordingStatusRequest,
	) (*RecordingResponse, error)

	DeleteRecording(
		ctx context.Context,
		userID uuid.UUID,
		recordingID uuid.UUID,
	) error
}

//
// Service Implementation
//

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{
		repository: repository,
	}
}

//
// Create Recording
//

func (s *service) CreateRecording(
	ctx context.Context,
	userID uuid.UUID,
	req CreateRecordingRequest,
) (*RecordingResponse, error) {

	if userID == uuid.Nil {
		return nil, ErrUnauthorized
	}

	if req.StreamID == uuid.Nil {
		return nil, ErrInvalidUUID
	}

	if req.StorageKey == "" {
		return nil, errors.New("storage key is required")
	}

	//
	// Storage provider
	//

	provider := req.StorageProvider

	if provider == "" {
		provider = StorageProviderMinIO
	}

	if !provider.IsValid() {
		return nil, ErrInvalidStorageProvider
	}

	//
	// Validate dimensions
	//

	if req.Width != nil || req.Height != nil {

		if req.Width == nil || req.Height == nil {
			return nil, errors.New(
				"width and height must be provided together",
			)
		}

		if *req.Width <= 0 || *req.Height <= 0 {
			return nil, errors.New(
				"width and height must be greater than zero",
			)
		}
	}

	//
	// Validate file size
	//

	if req.FileSize != nil && *req.FileSize < 0 {
		return nil, errors.New(
			"file size cannot be negative",
		)
	}

	//
	// Validate duration
	//

	if req.DurationSeconds != nil &&
		*req.DurationSeconds < 0 {

		return nil, errors.New(
			"duration cannot be negative",
		)
	}

	//
	// Verify stream ownership.
	//

	owner, err := s.repository.GetStreamOwner(
		ctx,
		req.StreamID,
	)

	if err != nil {

		if errors.Is(err, ErrStreamNotFound) {
			return nil, ErrStreamNotFound
		}

		return nil, fmt.Errorf(
			"get stream owner: %w",
			err,
		)
	}

	if owner.UserID != userID {
		return nil, ErrForbidden
	}

	//
	// Validate session.
	//

	if req.SessionID != nil {

		if *req.SessionID == uuid.Nil {
			return nil, ErrInvalidUUID
		}

		sessionOwner, err := s.repository.GetSessionOwner(
			ctx,
			*req.SessionID,
		)

		if err != nil {

			if errors.Is(err, ErrSessionNotFound) {
				return nil, ErrSessionNotFound
			}

			return nil, fmt.Errorf(
				"get session owner: %w",
				err,
			)
		}

		if sessionOwner.UserID != userID {
			return nil, ErrForbidden
		}

		if sessionOwner.StreamID != req.StreamID {
			return nil, errors.New(
				"session does not belong to stream",
			)
		}
	}

	//
	// Create recording.
	//

	now := time.Now().UTC()

	recording := &StreamRecording{
		ID:              uuid.New(),
		StreamID:        req.StreamID,
		SessionID:       req.SessionID,
		StorageProvider: provider,
		StorageKey:      req.StorageKey,
		PlaybackURL:     req.PlaybackURL,
		ThumbnailURL:    req.ThumbnailURL,
		Format:          req.Format,
		Width:           req.Width,
		Height:          req.Height,
		FileSize:        req.FileSize,
		DurationSeconds: req.DurationSeconds,
		ViewCount:       0,
		Status:          RecordingStatusProcessing,
		CreatedAt:       now,
		CompletedAt:     nil,
	}

	if err := s.repository.Create(
		ctx,
		recording,
	); err != nil {

		return nil, fmt.Errorf(
			"create recording: %w",
			err,
		)
	}

	return ToRecordingResponse(recording), nil
}

//
// Get Recording
//

func (s *service) GetRecording(
	ctx context.Context,
	recordingID uuid.UUID,
) (*RecordingResponse, error) {

	if recordingID == uuid.Nil {
		return nil, ErrInvalidUUID
	}

	recording, err := s.repository.GetByID(
		ctx,
		recordingID,
	)

	if err != nil {

		if errors.Is(err, ErrRecordingNotFound) {
			return nil, ErrRecordingNotFound
		}

		return nil, fmt.Errorf(
			"get recording: %w",
			err,
		)
	}

	return ToRecordingResponse(recording), nil
}

//
// List Recordings
//

func (s *service) ListRecordings(
	ctx context.Context,
	filter RecordingFilter,
) (*RecordingListResponse, error) {

	if filter.Page < 1 {
		filter.Page = 1
	}

	if filter.Limit <= 0 {
		filter.Limit = 20
	}

	if filter.Limit > 100 {
		filter.Limit = 100
	}

	if filter.StreamID != nil &&
		*filter.StreamID == uuid.Nil {

		return nil, ErrInvalidUUID
	}

	if filter.SessionID != nil &&
		*filter.SessionID == uuid.Nil {

		return nil, ErrInvalidUUID
	}

	if filter.Status != nil &&
		!filter.Status.IsValid() {

		return nil, ErrInvalidRecordingStatus
	}

	recordings, total, err := s.repository.List(
		ctx,
		filter,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"list recordings: %w",
			err,
		)
	}

	return &RecordingListResponse{
		Recordings: ToRecordingResponses(recordings),
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
	}, nil
}

//
// Update Recording
//

func (s *service) UpdateRecording(
	ctx context.Context,
	userID uuid.UUID,
	recordingID uuid.UUID,
	req UpdateRecordingRequest,
) (*RecordingResponse, error) {

	if userID == uuid.Nil {
		return nil, ErrUnauthorized
	}

	if recordingID == uuid.Nil {
		return nil, ErrInvalidUUID
	}

	recording, err := s.getOwnedRecording(
		ctx,
		userID,
		recordingID,
	)

	if err != nil {
		return nil, err
	}

	//
	// Playback URL
	//

	if req.PlaybackURL != nil {
		recording.PlaybackURL = req.PlaybackURL
	}

	//
	// Thumbnail URL
	//

	if req.ThumbnailURL != nil {
		recording.ThumbnailURL = req.ThumbnailURL
	}

	//
	// Format
	//

	if req.Format != nil {
		recording.Format = req.Format
	}

	//
	// Dimensions
	//

	if req.Width != nil {
		if *req.Width <= 0 {
			return nil, errors.New(
				"width must be greater than zero",
			)
		}

		recording.Width = req.Width
	}

	if req.Height != nil {
		if *req.Height <= 0 {
			return nil, errors.New(
				"height must be greater than zero",
			)
		}

		recording.Height = req.Height
	}

	//
	// Ensure dimensions remain valid together.
	//

	if recording.Width != nil ||
		recording.Height != nil {

		if recording.Width == nil ||
			recording.Height == nil {

			return nil, errors.New(
				"width and height must be provided together",
			)
		}
	}

	//
	// File size
	//

	if req.FileSize != nil {

		if *req.FileSize < 0 {
			return nil, errors.New(
				"file size cannot be negative",
			)
		}

		recording.FileSize = req.FileSize
	}

	//
	// Duration
	//

	if req.DurationSeconds != nil {

		if *req.DurationSeconds < 0 {
			return nil, errors.New(
				"duration cannot be negative",
			)
		}

		recording.DurationSeconds =
			req.DurationSeconds
	}

	//
	// Status
	//

	if req.Status != nil {

		if !req.Status.IsValid() {
			return nil, ErrInvalidRecordingStatus
		}

		recording.Status = *req.Status
	}

	//
	// Completed At
	//

	if req.CompletedAt != nil {
		recording.CompletedAt = req.CompletedAt
	}

	//
	// Automatically complete recording.
	//

	if recording.Status == RecordingStatusReady ||
		recording.Status == RecordingStatusFailed ||
		recording.Status == RecordingStatusDeleted {

		if recording.CompletedAt == nil {
			now := time.Now().UTC()
			recording.CompletedAt = &now
		}
	}

	if err := s.repository.Update(
		ctx,
		recording,
	); err != nil {

		if errors.Is(err, ErrRecordingNotFound) {
			return nil, ErrRecordingNotFound
		}

		return nil, fmt.Errorf(
			"update recording: %w",
			err,
		)
	}

	return ToRecordingResponse(recording), nil
}

//
// Update Recording Status
//

func (s *service) UpdateRecordingStatus(
	ctx context.Context,
	userID uuid.UUID,
	recordingID uuid.UUID,
	req UpdateRecordingStatusRequest,
) (*RecordingResponse, error) {

	if userID == uuid.Nil {
		return nil, ErrUnauthorized
	}

	if recordingID == uuid.Nil {
		return nil, ErrInvalidUUID
	}

	if !req.Status.IsValid() {
		return nil, ErrInvalidRecordingStatus
	}

	recording, err := s.getOwnedRecording(
		ctx,
		userID,
		recordingID,
	)

	if err != nil {
		return nil, err
	}

	recording.Status = req.Status

	//
	// Automatically set completed_at.
	//

	switch req.Status {

	case RecordingStatusReady,
		RecordingStatusFailed,
		RecordingStatusDeleted:

		if recording.CompletedAt == nil {
			now := time.Now().UTC()
			recording.CompletedAt = &now
		}

	case RecordingStatusProcessing:

		//
		// Processing recordings should not have
		// a completion timestamp.
		//

		recording.CompletedAt = nil
	}

	if err := s.repository.Update(
		ctx,
		recording,
	); err != nil {

		if errors.Is(err, ErrRecordingNotFound) {
			return nil, ErrRecordingNotFound
		}

		return nil, fmt.Errorf(
			"update recording status: %w",
			err,
		)
	}

	return ToRecordingResponse(recording), nil
}

//
// Delete Recording
//

func (s *service) DeleteRecording(
	ctx context.Context,
	userID uuid.UUID,
	recordingID uuid.UUID,
) error {

	if userID == uuid.Nil {
		return ErrUnauthorized
	}

	if recordingID == uuid.Nil {
		return ErrInvalidUUID
	}

	_, err := s.getOwnedRecording(
		ctx,
		userID,
		recordingID,
	)

	if err != nil {
		return err
	}

	//
	// Soft delete.
	//
	// Database supports DELETED status,
	// so we preserve the recording row.
	//

	if err := s.repository.Delete(
		ctx,
		recordingID,
	); err != nil {

		if errors.Is(err, ErrRecordingNotFound) {
			return ErrRecordingNotFound
		}

		return fmt.Errorf(
			"delete recording: %w",
			err,
		)
	}

	return nil
}

//
// Get Owned Recording
//

func (s *service) getOwnedRecording(
	ctx context.Context,
	userID uuid.UUID,
	recordingID uuid.UUID,
) (*StreamRecording, error) {

	recording, err := s.repository.GetByID(
		ctx,
		recordingID,
	)

	if err != nil {

		if errors.Is(err, ErrRecordingNotFound) {
			return nil, ErrRecordingNotFound
		}

		return nil, fmt.Errorf(
			"get recording: %w",
			err,
		)
	}

	owner, err := s.repository.GetRecordingOwner(
		ctx,
		recordingID,
	)

	if err != nil {

		if errors.Is(err, ErrRecordingNotFound) {
			return nil, ErrRecordingNotFound
		}

		return nil, fmt.Errorf(
			"get recording owner: %w",
			err,
		)
	}

	if owner.UserID != userID {
		return nil, ErrForbidden
	}

	return recording, nil
}
