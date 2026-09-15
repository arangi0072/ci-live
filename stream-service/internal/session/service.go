package session

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrUnauthorized    = errors.New("unauthorized")
	ErrForbidden       = errors.New("forbidden")
	ErrSessionNotFound = errors.New("session not found")
	ErrStreamNotFound  = errors.New("stream not found")
)

type Service interface {
	CreateSession(
		ctx context.Context,
		userID uuid.UUID,
		req CreateSessionRequest,
	) (*SessionResponse, error)

	GetSession(
		ctx context.Context,
		sessionID uuid.UUID,
	) (*SessionResponse, error)

	ListSessions(
		ctx context.Context,
		filter SessionFilter,
	) (*SessionListResponse, error)

	UpdateSession(
		ctx context.Context,
		userID uuid.UUID,
		sessionID uuid.UUID,
		req UpdateSessionRequest,
	) (*SessionResponse, error)

	EndSession(
		ctx context.Context,
		userID uuid.UUID,
		sessionID uuid.UUID,
		req EndSessionRequest,
	) (*SessionResponse, error)

	GetSessionStats(
		ctx context.Context,
		sessionID uuid.UUID,
	) (*SessionStats, error)

	DeleteSession(
		ctx context.Context,
		userID uuid.UUID,
		sessionID uuid.UUID,
	) error
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{
		repository: repository,
	}
}

//
// Create Session
//

func (s *service) CreateSession(
	ctx context.Context,
	userID uuid.UUID,
	req CreateSessionRequest,
) (*SessionResponse, error) {

	if userID == uuid.Nil {
		return nil, ErrUnauthorized
	}

	if req.StreamID == uuid.Nil {
		return nil, ErrStreamNotFound
	}

	// Verify that the authenticated user owns
	// the stream through the channel.
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

	now := time.Now().UTC()

	startedAt := req.StartedAt

	if startedAt == nil {
		startedAt = &now
	}

	session := &StreamSession{
		ID:              uuid.New(),
		StreamID:        req.StreamID,
		MediaServerID:   req.MediaServerID,
		MediaPath:       req.MediaPath,
		StartedAt:       startedAt,
		EndedAt:         nil,
		DurationSeconds: 0,
		PeakViewers:     0,
		CreatedAt:       now,
	}

	if err := s.repository.Create(
		ctx,
		session,
	); err != nil {
		return nil, fmt.Errorf(
			"create session: %w",
			err,
		)
	}

	return ToSessionResponse(session), nil
}

//
// Get Session
//

func (s *service) GetSession(
	ctx context.Context,
	sessionID uuid.UUID,
) (*SessionResponse, error) {

	if sessionID == uuid.Nil {
		return nil, ErrSessionNotFound
	}

	session, err := s.repository.GetByID(
		ctx,
		sessionID,
	)

	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			return nil, ErrSessionNotFound
		}

		return nil, fmt.Errorf(
			"get session: %w",
			err,
		)
	}

	return ToSessionResponse(session), nil
}

//
// List Sessions
//

func (s *service) ListSessions(
	ctx context.Context,
	filter SessionFilter,
) (*SessionListResponse, error) {

	if filter.Page < 1 {
		filter.Page = 1
	}

	if filter.Limit < 1 {
		filter.Limit = 20
	}

	if filter.Limit > 100 {
		filter.Limit = 100
	}

	sessions, total, err := s.repository.List(
		ctx,
		filter,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"list sessions: %w",
			err,
		)
	}

	items := make(
		[]SessionResponse,
		0,
		len(sessions),
	)

	for i := range sessions {
		items = append(
			items,
			*ToSessionResponse(&sessions[i]),
		)
	}

	return &SessionListResponse{
		Sessions: items,
		Total:    total,
		Page:     filter.Page,
		Limit:    filter.Limit,
	}, nil
}

//
// Update Session
//

func (s *service) UpdateSession(
	ctx context.Context,
	userID uuid.UUID,
	sessionID uuid.UUID,
	req UpdateSessionRequest,
) (*SessionResponse, error) {

	if userID == uuid.Nil {
		return nil, ErrUnauthorized
	}

	session, err := s.getOwnedSession(
		ctx,
		userID,
		sessionID,
	)

	if err != nil {
		return nil, err
	}

	if req.MediaServerID != nil {
		session.MediaServerID = *req.MediaServerID
	}

	if req.MediaPath != nil {
		session.MediaPath = *req.MediaPath
	}

	if req.StartedAt != nil {
		session.StartedAt = req.StartedAt
	}

	if req.EndedAt != nil {
		session.EndedAt = req.EndedAt
	}

	if req.DurationSeconds != nil {
		session.DurationSeconds = *req.DurationSeconds
	}

	if req.PeakViewers != nil {
		session.PeakViewers = *req.PeakViewers
	}

	// If an end time exists but duration wasn't explicitly
	// supplied, calculate the duration automatically.
	if session.EndedAt != nil &&
		session.StartedAt != nil &&
		req.DurationSeconds == nil {

		duration := session.EndedAt.Sub(
			*session.StartedAt,
		)

		if duration < 0 {
			return nil, errors.New(
				"ended_at cannot be before started_at",
			)
		}

		session.DurationSeconds = int64(
			duration.Seconds(),
		)
	}

	if err := s.repository.Update(
		ctx,
		session,
	); err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			return nil, ErrSessionNotFound
		}

		return nil, fmt.Errorf(
			"update session: %w",
			err,
		)
	}

	return ToSessionResponse(session), nil
}

//
// End Session
//

func (s *service) EndSession(
	ctx context.Context,
	userID uuid.UUID,
	sessionID uuid.UUID,
	req EndSessionRequest,
) (*SessionResponse, error) {

	if userID == uuid.Nil {
		return nil, ErrUnauthorized
	}

	session, err := s.getOwnedSession(
		ctx,
		userID,
		sessionID,
	)

	if err != nil {
		return nil, err
	}

	// Already ended.
	if session.EndedAt != nil {
		return ToSessionResponse(session), nil
	}

	endedAt := req.EndedAt

	if endedAt == nil {
		now := time.Now().UTC()
		endedAt = &now
	}

	if session.StartedAt != nil &&
		endedAt.Before(*session.StartedAt) {

		return nil, errors.New(
			"ended_at cannot be before started_at",
		)
	}

	durationSeconds := req.DurationSeconds

	if durationSeconds == nil &&
		session.StartedAt != nil {

		duration := endedAt.Sub(
			*session.StartedAt,
		)

		value := int64(duration.Seconds())

		if value < 0 {
			value = 0
		}

		durationSeconds = &value
	}

	if durationSeconds == nil {
		value := int64(0)
		durationSeconds = &value
	}

	if err := s.repository.EndSession(
		ctx,
		sessionID,
		*endedAt,
		*durationSeconds,
	); err != nil {

		if errors.Is(err, ErrSessionNotFound) {
			return nil, ErrSessionNotFound
		}

		return nil, fmt.Errorf(
			"end session: %w",
			err,
		)
	}

	session.EndedAt = endedAt
	session.DurationSeconds = *durationSeconds

	return ToSessionResponse(session), nil
}

//
// Get Session Statistics
//

func (s *service) GetSessionStats(
	ctx context.Context,
	sessionID uuid.UUID,
) (*SessionStats, error) {

	if sessionID == uuid.Nil {
		return nil, ErrSessionNotFound
	}

	session, err := s.repository.GetByID(
		ctx,
		sessionID,
	)

	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			return nil, ErrSessionNotFound
		}

		return nil, fmt.Errorf(
			"get session: %w",
			err,
		)
	}

	duration := session.DurationSeconds

	// For an active session calculate current duration.
	if session.StartedAt != nil &&
		session.EndedAt == nil {

		current := time.Since(
			*session.StartedAt,
		)

		if current > 0 {
			duration = int64(
				current.Seconds(),
			)
		}
	}

	return &SessionStats{
		SessionID:       session.ID,
		StreamID:        session.StreamID,
		MediaServerID:   session.MediaServerID,
		MediaPath:       session.MediaPath,
		StartedAt:       session.StartedAt,
		EndedAt:         session.EndedAt,
		DurationSeconds: duration,
		PeakViewers:     session.PeakViewers,
	}, nil
}

//
// Delete Session
//

func (s *service) DeleteSession(
	ctx context.Context,
	userID uuid.UUID,
	sessionID uuid.UUID,
) error {

	if userID == uuid.Nil {
		return ErrUnauthorized
	}

	_, err := s.getOwnedSession(
		ctx,
		userID,
		sessionID,
	)

	if err != nil {
		return err
	}

	if err := s.repository.Delete(
		ctx,
		sessionID,
	); err != nil {

		if errors.Is(err, ErrSessionNotFound) {
			return ErrSessionNotFound
		}

		return fmt.Errorf(
			"delete session: %w",
			err,
		)
	}

	return nil
}

//
// Get Owned Session
//

func (s *service) getOwnedSession(
	ctx context.Context,
	userID uuid.UUID,
	sessionID uuid.UUID,
) (*StreamSession, error) {

	if sessionID == uuid.Nil {
		return nil, ErrSessionNotFound
	}

	session, err := s.repository.GetByID(
		ctx,
		sessionID,
	)

	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			return nil, ErrSessionNotFound
		}

		return nil, fmt.Errorf(
			"get session: %w",
			err,
		)
	}

	owner, err := s.repository.GetSessionOwner(
		ctx,
		sessionID,
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

	if owner.UserID != userID {
		return nil, ErrForbidden
	}

	return session, nil
}
