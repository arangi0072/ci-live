package stream

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

//
// Errors
//

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrInvalidUUID  = errors.New("invalid UUID")

	ErrStreamNotFound  = errors.New("stream not found")
	ErrChannelNotFound = errors.New("channel not found")

	ErrInvalidStreamStatus     = errors.New("invalid stream status")
	ErrInvalidStreamVisibility = errors.New("invalid stream visibility")

	ErrInvalidTitle       = errors.New("invalid stream title")
	ErrInvalidViewerCount = errors.New("invalid viewer count")

	ErrStreamAlreadyLive = errors.New("stream is already live")
	ErrStreamNotLive     = errors.New("stream is not live")

	ErrInvalidStatusTransition = errors.New("invalid stream status transition")
)

//
// Service Interface
//

type Service interface {
	CreateStream(
		ctx context.Context,
		userID uuid.UUID,
		req CreateStreamRequest,
	) (*StreamResponse, error)

	GetStream(
		ctx context.Context,
		streamID uuid.UUID,
	) (*StreamResponse, error)

	GetPublicStream(
		ctx context.Context,
		streamID uuid.UUID,
	) (*PublicStreamResponse, error)

	ListStreams(
		ctx context.Context,
		filter StreamFilter,
	) (*StreamListResponse, error)

	UpdateStream(
		ctx context.Context,
		userID uuid.UUID,
		streamID uuid.UUID,
		req UpdateStreamRequest,
	) (*StreamResponse, error)

	DeleteStream(
		ctx context.Context,
		userID uuid.UUID,
		streamID uuid.UUID,
	) error

	StartStream(
		ctx context.Context,
		userID uuid.UUID,
		streamID uuid.UUID,
	) (*StreamResponse, error)

	EndStream(
		ctx context.Context,
		userID uuid.UUID,
		streamID uuid.UUID,
	) (*StreamResponse, error)

	UpdateStreamStatus(
		ctx context.Context,
		userID uuid.UUID,
		streamID uuid.UUID,
		req UpdateStreamStatusRequest,
	) (*StreamResponse, error)

	GetStreamStats(
		ctx context.Context,
		streamID uuid.UUID,
	) (*StreamStats, error)

	IncrementView(
		ctx context.Context,
		streamID uuid.UUID,
	) (*StreamStats, error)

	UpdateViewerCount(
		ctx context.Context,
		userID uuid.UUID,
		streamID uuid.UUID,
		viewerCount int,
	) (*StreamStats, error)
}

//
// Repository
//
// This interface is implemented by stream/repository.go.
//

type Repository interface {
	Create(
		ctx context.Context,
		stream *Stream,
	) error

	GetByID(
		ctx context.Context,
		streamID uuid.UUID,
	) (*Stream, error)

	GetPublicByID(
		ctx context.Context,
		streamID uuid.UUID,
	) (*Stream, error)

	List(
		ctx context.Context,
		filter StreamFilter,
	) ([]Stream, int64, error)

	Update(
		ctx context.Context,
		stream *Stream,
	) error

	Delete(
		ctx context.Context,
		streamID uuid.UUID,
	) error

	GetStreamOwner(
		ctx context.Context,
		streamID uuid.UUID,
	) (*StreamOwner, error)

	GetChannelByUserID(
		ctx context.Context,
		userID uuid.UUID,
	) (*ChannelOwner, error)

	GetStats(
		ctx context.Context,
		streamID uuid.UUID,
	) (*StreamStats, error)

	IncrementViews(
		ctx context.Context,
		streamID uuid.UUID,
	) error

	UpdateViewerCount(
		ctx context.Context,
		streamID uuid.UUID,
		viewerCount int,
		peakViewers int,
	) error

	UpdateStatus(
		ctx context.Context,
		streamID uuid.UUID,
		status StreamStatus,
		startedAt *time.Time,
		endedAt *time.Time,
	) error
}

//
// Channel owner
//
// streams.channel_id -> channels.id -> channels.user_id
//

type ChannelOwner struct {
	ID     uuid.UUID `db:"id"`
	UserID uuid.UUID `db:"user_id"`
}

//
// Service implementation
//

type service struct {
	repository Repository
}

//
// Constructor
//

func NewService(
	repository Repository,
) Service {
	return &service{
		repository: repository,
	}
}

//
// Create Stream
//

func (s *service) CreateStream(
	ctx context.Context,
	userID uuid.UUID,
	req CreateStreamRequest,
) (*StreamResponse, error) {

	if userID == uuid.Nil {
		return nil, ErrUnauthorized
	}

	//
	// Validate title.
	//

	title := strings.TrimSpace(req.Title)

	if title == "" {
		return nil, ErrInvalidTitle
	}

	if len(title) > 200 {
		return nil, ErrInvalidTitle
	}

	//
	// Default visibility.
	//

	visibility := req.Visibility

	if visibility == "" {
		visibility = StreamVisibilityPublic
	}

	if !visibility.IsValid() {
		return nil, ErrInvalidStreamVisibility
	}

	//
	// Find the user's channel.
	//
	// A stream belongs to a channel, not directly to a user.
	//

	channel, err := s.repository.GetChannelByUserID(
		ctx,
		userID,
	)

	if err != nil {
		if errors.Is(err, ErrChannelNotFound) {
			return nil, ErrChannelNotFound
		}

		return nil, fmt.Errorf(
			"get user channel: %w",
			err,
		)
	}

	now := time.Now().UTC()

	stream := &Stream{
		ID:           uuid.New(),
		ChannelID:    channel.ID,
		Title:        title,
		Description:  cleanOptionalString(req.Description),
		ThumbnailURL: cleanOptionalString(req.ThumbnailURL),
		CategoryID:   req.CategoryID,
		Status:       StreamStatusCreated,
		Visibility:   visibility,
		ViewerCount:  0,
		PeakViewers:  0,
		LikeCount:    0,
		TotalViews:   0,
		StartedAt:    nil,
		EndedAt:      nil,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repository.Create(
		ctx,
		stream,
	); err != nil {
		return nil, fmt.Errorf(
			"create stream: %w",
			err,
		)
	}

	return toStreamResponse(stream), nil
}

//
// Get Stream
//

func (s *service) GetStream(
	ctx context.Context,
	streamID uuid.UUID,
) (*StreamResponse, error) {

	if streamID == uuid.Nil {
		return nil, ErrStreamNotFound
	}

	stream, err := s.repository.GetByID(
		ctx,
		streamID,
	)

	if err != nil {
		if errors.Is(err, ErrStreamNotFound) {
			return nil, ErrStreamNotFound
		}

		return nil, fmt.Errorf(
			"get stream: %w",
			err,
		)
	}

	return toStreamResponse(stream), nil
}

//
// Get Public Stream
//

func (s *service) GetPublicStream(
	ctx context.Context,
	streamID uuid.UUID,
) (*PublicStreamResponse, error) {

	if streamID == uuid.Nil {
		return nil, ErrStreamNotFound
	}

	stream, err := s.repository.GetPublicByID(
		ctx,
		streamID,
	)

	if err != nil {
		if errors.Is(err, ErrStreamNotFound) {
			return nil, ErrStreamNotFound
		}

		return nil, fmt.Errorf(
			"get public stream: %w",
			err,
		)
	}

	//
	// Only public streams are returned here.
	//

	if stream.Visibility != StreamVisibilityPublic {
		return nil, ErrStreamNotFound
	}

	return toPublicStreamResponse(stream), nil
}

//
// List Streams
//

func (s *service) ListStreams(
	ctx context.Context,
	filter StreamFilter,
) (*StreamListResponse, error) {

	//
	// Pagination defaults.
	//

	if filter.Page < 1 {
		filter.Page = 1
	}

	if filter.Limit < 1 {
		filter.Limit = 20
	}

	if filter.Limit > 100 {
		filter.Limit = 100
	}

	//
	// Validate filters.
	//

	if filter.Status != nil &&
		!filter.Status.IsValid() {
		return nil, ErrInvalidStreamStatus
	}

	if filter.Visibility != nil &&
		!filter.Visibility.IsValid() {
		return nil, ErrInvalidStreamVisibility
	}

	streams, total, err := s.repository.List(
		ctx,
		filter,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"list streams: %w",
			err,
		)
	}

	result := make(
		[]StreamResponse,
		0,
		len(streams),
	)

	for i := range streams {
		result = append(
			result,
			*toStreamResponse(&streams[i]),
		)
	}

	return &StreamListResponse{
		Streams: result,
		Total:   total,
		Page:    filter.Page,
		Limit:   filter.Limit,
	}, nil
}

//
// Update Stream
//

func (s *service) UpdateStream(
	ctx context.Context,
	userID uuid.UUID,
	streamID uuid.UUID,
	req UpdateStreamRequest,
) (*StreamResponse, error) {

	if userID == uuid.Nil {
		return nil, ErrUnauthorized
	}

	if streamID == uuid.Nil {
		return nil, ErrStreamNotFound
	}

	stream, err := s.repository.GetByID(
		ctx,
		streamID,
	)

	if err != nil {
		if errors.Is(err, ErrStreamNotFound) {
			return nil, ErrStreamNotFound
		}

		return nil, fmt.Errorf(
			"get stream: %w",
			err,
		)
	}

	//
	// Verify channel ownership.
	//

	if err := s.verifyOwnership(
		ctx,
		userID,
		streamID,
	); err != nil {
		return nil, err
	}

	//
	// Do not allow metadata changes after ending.
	//

	if stream.Status == StreamStatusEnded ||
		stream.Status == StreamStatusCancelled {

		return nil, ErrInvalidStatusTransition
	}

	//
	// Title.
	//

	if req.Title != nil {
		title := strings.TrimSpace(
			*req.Title,
		)

		if title == "" || len(title) > 200 {
			return nil, ErrInvalidTitle
		}

		stream.Title = title
	}

	//
	// Description.
	//

	if req.Description != nil {
		stream.Description = cleanOptionalString(
			req.Description,
		)
	}

	//
	// Thumbnail.
	//

	if req.ThumbnailURL != nil {
		stream.ThumbnailURL = cleanOptionalString(
			req.ThumbnailURL,
		)
	}

	//
	// Category.
	//

	if req.CategoryID != nil {
		stream.CategoryID = req.CategoryID
	}

	//
	// Visibility.
	//

	if req.Visibility != nil {

		if !req.Visibility.IsValid() {
			return nil, ErrInvalidStreamVisibility
		}

		stream.Visibility = *req.Visibility
	}

	stream.UpdatedAt = time.Now().UTC()

	if err := s.repository.Update(
		ctx,
		stream,
	); err != nil {
		return nil, fmt.Errorf(
			"update stream: %w",
			err,
		)
	}

	return toStreamResponse(stream), nil
}

//
// Delete Stream
//

func (s *service) DeleteStream(
	ctx context.Context,
	userID uuid.UUID,
	streamID uuid.UUID,
) error {

	if userID == uuid.Nil {
		return ErrUnauthorized
	}

	if streamID == uuid.Nil {
		return ErrStreamNotFound
	}

	stream, err := s.repository.GetByID(
		ctx,
		streamID,
	)

	if err != nil {
		if errors.Is(err, ErrStreamNotFound) {
			return ErrStreamNotFound
		}

		return fmt.Errorf(
			"get stream: %w",
			err,
		)
	}

	if err := s.verifyOwnership(
		ctx,
		userID,
		stream.ID,
	); err != nil {
		return err
	}

	//
	// A live stream should be ended first.
	//

	if stream.Status == StreamStatusLive ||
		stream.Status == StreamStatusStarting ||
		stream.Status == StreamStatusReconnecting {

		return ErrInvalidStatusTransition
	}

	if err := s.repository.Delete(
		ctx,
		streamID,
	); err != nil {
		if errors.Is(err, ErrStreamNotFound) {
			return ErrStreamNotFound
		}

		return fmt.Errorf(
			"delete stream: %w",
			err,
		)
	}

	return nil
}

//
// Start Stream
//

func (s *service) StartStream(
	ctx context.Context,
	userID uuid.UUID,
	streamID uuid.UUID,
) (*StreamResponse, error) {

	if userID == uuid.Nil {
		return nil, ErrUnauthorized
	}

	if streamID == uuid.Nil {
		return nil, ErrStreamNotFound
	}

	stream, err := s.repository.GetByID(
		ctx,
		streamID,
	)

	if err != nil {
		return nil, err
	}

	if err := s.verifyOwnership(
		ctx,
		userID,
		streamID,
	); err != nil {
		return nil, err
	}

	//
	// Cannot start an already-live stream.
	//

	if stream.Status == StreamStatusLive {
		return nil, ErrStreamAlreadyLive
	}

	//
	// A finished/cancelled stream cannot be restarted.
	//

	if stream.Status == StreamStatusEnded ||
		stream.Status == StreamStatusCancelled {

		return nil, ErrInvalidStatusTransition
	}

	//
	// Only CREATED/FAILED streams can begin again.
	//

	if stream.Status != StreamStatusCreated &&
		stream.Status != StreamStatusFailed {

		return nil, ErrInvalidStatusTransition
	}

	now := time.Now().UTC()

	if err := s.repository.UpdateStatus(
		ctx,
		streamID,
		StreamStatusStarting,
		nil,
		nil,
	); err != nil {
		return nil, fmt.Errorf(
			"start stream: %w",
			err,
		)
	}

	stream.Status = StreamStatusStarting
	stream.UpdatedAt = now

	return toStreamResponse(stream), nil
}

//
// End Stream
//

func (s *service) EndStream(
	ctx context.Context,
	userID uuid.UUID,
	streamID uuid.UUID,
) (*StreamResponse, error) {

	if userID == uuid.Nil {
		return nil, ErrUnauthorized
	}

	if streamID == uuid.Nil {
		return nil, ErrStreamNotFound
	}

	stream, err := s.repository.GetByID(
		ctx,
		streamID,
	)

	if err != nil {
		return nil, err
	}

	if err := s.verifyOwnership(
		ctx,
		userID,
		streamID,
	); err != nil {
		return nil, err
	}

	//
	// The stream must actually be running.
	//

	if stream.Status != StreamStatusLive &&
		stream.Status != StreamStatusStarting &&
		stream.Status != StreamStatusReconnecting {

		return nil, ErrStreamNotLive
	}

	now := time.Now().UTC()

	if err := s.repository.UpdateStatus(
		ctx,
		streamID,
		StreamStatusEnded,
		nil,
		&now,
	); err != nil {
		return nil, fmt.Errorf(
			"end stream: %w",
			err,
		)
	}

	stream.Status = StreamStatusEnded
	stream.EndedAt = &now
	stream.UpdatedAt = now

	return toStreamResponse(stream), nil
}

//
// Update Stream Status
//

func (s *service) UpdateStreamStatus(
	ctx context.Context,
	userID uuid.UUID,
	streamID uuid.UUID,
	req UpdateStreamStatusRequest,
) (*StreamResponse, error) {

	if userID == uuid.Nil {
		return nil, ErrUnauthorized
	}

	if streamID == uuid.Nil {
		return nil, ErrStreamNotFound
	}

	if !req.Status.IsValid() {
		return nil, ErrInvalidStreamStatus
	}

	stream, err := s.repository.GetByID(
		ctx,
		streamID,
	)

	if err != nil {
		return nil, err
	}

	if err := s.verifyOwnership(
		ctx,
		userID,
		streamID,
	); err != nil {
		return nil, err
	}

	if !isValidStatusTransition(
		stream.Status,
		req.Status,
	) {
		return nil, ErrInvalidStatusTransition
	}

	var startedAt *time.Time
	var endedAt *time.Time

	now := time.Now().UTC()

	switch req.Status {

	case StreamStatusStarting:
		startedAt = stream.StartedAt

	case StreamStatusLive:
		if stream.StartedAt == nil {
			startedAt = &now
		} else {
			startedAt = stream.StartedAt
		}

	case StreamStatusEnded,
		StreamStatusCancelled,
		StreamStatusFailed:

		endedAt = &now
	}

	if err := s.repository.UpdateStatus(
		ctx,
		streamID,
		req.Status,
		startedAt,
		endedAt,
	); err != nil {
		return nil, fmt.Errorf(
			"update stream status: %w",
			err,
		)
	}

	stream.Status = req.Status

	if startedAt != nil {
		stream.StartedAt = startedAt
	}

	if endedAt != nil {
		stream.EndedAt = endedAt
	}

	stream.UpdatedAt = now

	return toStreamResponse(stream), nil
}

//
// Get Stream Statistics
//

func (s *service) GetStreamStats(
	ctx context.Context,
	streamID uuid.UUID,
) (*StreamStats, error) {

	if streamID == uuid.Nil {
		return nil, ErrStreamNotFound
	}

	stats, err := s.repository.GetStats(
		ctx,
		streamID,
	)

	if err != nil {
		if errors.Is(err, ErrStreamNotFound) {
			return nil, ErrStreamNotFound
		}

		return nil, fmt.Errorf(
			"get stream stats: %w",
			err,
		)
	}

	return stats, nil
}

//
// Increment View
//

func (s *service) IncrementView(
	ctx context.Context,
	streamID uuid.UUID,
) (*StreamStats, error) {

	if streamID == uuid.Nil {
		return nil, ErrStreamNotFound
	}

	//
	// Verify stream exists.
	//

	_, err := s.repository.GetByID(
		ctx,
		streamID,
	)

	if err != nil {
		return nil, err
	}

	if err := s.repository.IncrementViews(
		ctx,
		streamID,
	); err != nil {
		return nil, fmt.Errorf(
			"increment stream views: %w",
			err,
		)
	}

	return s.repository.GetStats(
		ctx,
		streamID,
	)
}

//
// Update Viewer Count
//

func (s *service) UpdateViewerCount(
	ctx context.Context,
	userID uuid.UUID,
	streamID uuid.UUID,
	viewerCount int,
) (*StreamStats, error) {

	if userID == uuid.Nil {
		return nil, ErrUnauthorized
	}

	if streamID == uuid.Nil {
		return nil, ErrStreamNotFound
	}

	if viewerCount < 0 {
		return nil, ErrInvalidViewerCount
	}

	stream, err := s.repository.GetByID(
		ctx,
		streamID,
	)

	if err != nil {
		return nil, err
	}

	if err := s.verifyOwnership(
		ctx,
		userID,
		streamID,
	); err != nil {
		return nil, err
	}

	//
	// Peak viewers can only increase.
	//

	peakViewers := stream.PeakViewers

	if viewerCount > peakViewers {
		peakViewers = viewerCount
	}

	if err := s.repository.UpdateViewerCount(
		ctx,
		streamID,
		viewerCount,
		peakViewers,
	); err != nil {
		return nil, fmt.Errorf(
			"update viewer count: %w",
			err,
		)
	}

	return s.repository.GetStats(
		ctx,
		streamID,
	)
}

//
// Ownership
//

func (s *service) verifyOwnership(
	ctx context.Context,
	userID uuid.UUID,
	streamID uuid.UUID,
) error {

	owner, err := s.repository.GetStreamOwner(
		ctx,
		streamID,
	)

	if err != nil {
		if errors.Is(err, ErrStreamNotFound) {
			return ErrStreamNotFound
		}

		return fmt.Errorf(
			"get stream ownership: %w",
			err,
		)
	}

	if owner.UserID != userID {
		return ErrForbidden
	}

	return nil
}

//
// Status transitions
//

func isValidStatusTransition(
	current StreamStatus,
	next StreamStatus,
) bool {

	if current == next {
		return true
	}

	switch current {

	case StreamStatusCreated:
		return next == StreamStatusStarting ||
			next == StreamStatusCancelled

	case StreamStatusStarting:
		return next == StreamStatusLive ||
			next == StreamStatusReconnecting ||
			next == StreamStatusEnding ||
			next == StreamStatusFailed

	case StreamStatusLive:
		return next == StreamStatusReconnecting ||
			next == StreamStatusEnding ||
			next == StreamStatusEnded ||
			next == StreamStatusFailed

	case StreamStatusReconnecting:
		return next == StreamStatusLive ||
			next == StreamStatusEnding ||
			next == StreamStatusFailed

	case StreamStatusEnding:
		return next == StreamStatusEnded ||
			next == StreamStatusFailed

	case StreamStatusFailed:
		return next == StreamStatusStarting ||
			next == StreamStatusCancelled

	case StreamStatusEnded,
		StreamStatusCancelled:

		return false

	default:
		return false
	}
}

//
// Response helpers
//

func toStreamResponse(
	stream *Stream,
) *StreamResponse {

	if stream == nil {
		return nil
	}

	return &StreamResponse{
		ID:           stream.ID,
		ChannelID:    stream.ChannelID,
		Title:        stream.Title,
		Description:  stream.Description,
		ThumbnailURL: stream.ThumbnailURL,
		CategoryID:   stream.CategoryID,
		Status:       stream.Status,
		Visibility:   stream.Visibility,
		ViewerCount:  stream.ViewerCount,
		PeakViewers:  stream.PeakViewers,
		LikeCount:    stream.LikeCount,
		TotalViews:   stream.TotalViews,
		StartedAt:    stream.StartedAt,
		EndedAt:      stream.EndedAt,
		CreatedAt:    stream.CreatedAt,
		UpdatedAt:    stream.UpdatedAt,
	}
}

func toPublicStreamResponse(
	stream *Stream,
) *PublicStreamResponse {

	if stream == nil {
		return nil
	}

	return &PublicStreamResponse{
		ID:           stream.ID,
		ChannelID:    stream.ChannelID,
		Title:        stream.Title,
		Description:  stream.Description,
		ThumbnailURL: stream.ThumbnailURL,
		CategoryID:   stream.CategoryID,
		Status:       stream.Status,
		Visibility:   stream.Visibility,
		ViewerCount:  stream.ViewerCount,
		PeakViewers:  stream.PeakViewers,
		LikeCount:    stream.LikeCount,
		TotalViews:   stream.TotalViews,
		StartedAt:    stream.StartedAt,
		CreatedAt:    stream.CreatedAt,
	}
}

//
// Optional string helper
//

func cleanOptionalString(
	value *string,
) *string {

	if value == nil {
		return nil
	}

	cleaned := strings.TrimSpace(*value)

	if cleaned == "" {
		return nil
	}

	return &cleaned
}
