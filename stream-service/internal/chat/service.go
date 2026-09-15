package chat

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
)

//
// Service Errors
//

var (
	ErrUnauthorized = errors.New(
		"unauthorized",
	)

	ErrForbidden = errors.New(
		"forbidden",
	)

	ErrMessageNotFound = errors.New(
		"chat message not found",
	)

	ErrStreamNotFound = errors.New(
		"stream not found",
	)

	ErrInvalidUUID = errors.New(
		"invalid UUID",
	)

	ErrInvalidMessageID = errors.New(
		"invalid message ID",
	)

	ErrMessageEmpty = errors.New(
		"message cannot be empty",
	)

	ErrMessageTooLong = errors.New(
		"message cannot exceed 500 characters",
	)
)

//
// Repository Interface
//
// Implemented by repository.go.
//

type Repository interface {
	CreateMessage(
		ctx context.Context,
		message *ChatMessage,
	) (*ChatMessage, error)

	GetMessageByID(
		ctx context.Context,
		messageID int64,
	) (*ChatMessage, error)

	ListMessages(
		ctx context.Context,
		filter MessageFilter,
	) ([]ChatMessage, int64, error)

	DeleteMessage(
		ctx context.Context,
		messageID int64,
	) error

	GetMessageOwner(
		ctx context.Context,
		messageID int64,
	) (*MessageOwner, error)

	GetStreamOwner(
		ctx context.Context,
		streamID uuid.UUID,
	) (*StreamOwner, error)

	GetMessageStats(
		ctx context.Context,
		streamID uuid.UUID,
	) (*MessageStats, error)
}

//
// Service Interface
//

type Service interface {
	CreateMessage(
		ctx context.Context,
		userID uuid.UUID,
		req CreateMessageRequest,
	) (*MessageResponse, error)

	GetMessage(
		ctx context.Context,
		messageID int64,
	) (*MessageResponse, error)

	ListMessages(
		ctx context.Context,
		filter MessageFilter,
	) (*MessageListResponse, error)

	DeleteMessage(
		ctx context.Context,
		userID uuid.UUID,
		messageID int64,
	) error

	GetMessageStats(
		ctx context.Context,
		streamID uuid.UUID,
	) (*MessageStats, error)
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
// Create Message
//

func (s *service) CreateMessage(
	ctx context.Context,
	userID uuid.UUID,
	req CreateMessageRequest,
) (*MessageResponse, error) {

	if userID == uuid.Nil {
		return nil, ErrUnauthorized
	}

	if req.StreamID == uuid.Nil {
		return nil, ErrInvalidUUID
	}

	//
	// Trim whitespace before validation.
	//

	messageText := strings.TrimSpace(req.Message)

	if messageText == "" {
		return nil, ErrMessageEmpty
	}

	//
	// PostgreSQL schema allows 1–500 characters.
	//

	if len([]rune(messageText)) > 500 {
		return nil, ErrMessageTooLong
	}

	//
	// Verify that the stream exists.
	//

	_, err := s.repository.GetStreamOwner(
		ctx,
		req.StreamID,
	)

	if err != nil {
		if errors.Is(err, ErrStreamNotFound) {
			return nil, ErrStreamNotFound
		}

		return nil, err
	}

	message := &ChatMessage{
		StreamID: req.StreamID,
		UserID:   userID,
		Message:  messageText,
	}

	createdMessage, err := s.repository.CreateMessage(
		ctx,
		message,
	)

	if err != nil {
		return nil, err
	}

	return ToMessageResponse(createdMessage), nil
}

//
// Get Message
//

func (s *service) GetMessage(
	ctx context.Context,
	messageID int64,
) (*MessageResponse, error) {

	if messageID <= 0 {
		return nil, ErrInvalidMessageID
	}

	message, err := s.repository.GetMessageByID(
		ctx,
		messageID,
	)

	if err != nil {
		return nil, err
	}

	if message == nil {
		return nil, ErrMessageNotFound
	}

	return ToMessageResponse(message), nil
}

//
// List Messages
//

func (s *service) ListMessages(
	ctx context.Context,
	filter MessageFilter,
) (*MessageListResponse, error) {

	if filter.StreamID == uuid.Nil {
		return nil, ErrInvalidUUID
	}

	//
	// Safe pagination defaults.
	//

	if filter.Page < 1 {
		filter.Page = 1
	}

	if filter.Limit < 1 {
		filter.Limit = 50
	}

	if filter.Limit > 100 {
		filter.Limit = 100
	}

	//
	// Verify stream exists.
	//

	_, err := s.repository.GetStreamOwner(
		ctx,
		filter.StreamID,
	)

	if err != nil {
		if errors.Is(err, ErrStreamNotFound) {
			return nil, ErrStreamNotFound
		}

		return nil, err
	}

	messages, total, err := s.repository.ListMessages(
		ctx,
		filter,
	)

	if err != nil {
		return nil, err
	}

	return &MessageListResponse{
		Messages: ToMessageResponses(messages),
		Total:    total,
		Page:     filter.Page,
		Limit:    filter.Limit,
	}, nil
}

//
// Delete Message
//
// This is a soft delete.
// The repository updates:
//
// is_deleted = TRUE
// deleted_at = NOW()
//

func (s *service) DeleteMessage(
	ctx context.Context,
	userID uuid.UUID,
	messageID int64,
) error {

	if userID == uuid.Nil {
		return ErrUnauthorized
	}

	if messageID <= 0 {
		return ErrInvalidMessageID
	}

	owner, err := s.repository.GetMessageOwner(
		ctx,
		messageID,
	)

	if err != nil {
		return err
	}

	if owner == nil {
		return ErrMessageNotFound
	}

	//
	// Currently only the message author can delete
	// their own message.
	//
	// There is no moderator/role table in the current
	// database schema, so we do not invent one here.
	//

	if owner.UserID != userID {
		return ErrForbidden
	}

	return s.repository.DeleteMessage(
		ctx,
		messageID,
	)
}

//
// Get Message Statistics
//

func (s *service) GetMessageStats(
	ctx context.Context,
	streamID uuid.UUID,
) (*MessageStats, error) {

	if streamID == uuid.Nil {
		return nil, ErrInvalidUUID
	}

	//
	// Verify stream exists.
	//

	_, err := s.repository.GetStreamOwner(
		ctx,
		streamID,
	)

	if err != nil {
		if errors.Is(err, ErrStreamNotFound) {
			return nil, ErrStreamNotFound
		}

		return nil, err
	}

	return s.repository.GetMessageStats(
		ctx,
		streamID,
	)
}
