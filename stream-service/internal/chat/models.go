package chat

import (
	"time"

	"github.com/google/uuid"
)

//
// Chat Message
//

type ChatMessage struct {
	ID int64 `json:"id" db:"id"`

	StreamID uuid.UUID `json:"stream_id" db:"stream_id"`

	UserID uuid.UUID `json:"user_id" db:"user_id"`

	Message string `json:"message" db:"message"`

	IsDeleted bool `json:"is_deleted" db:"is_deleted"`

	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

//
// Create Message Request
//

type CreateMessageRequest struct {
	StreamID uuid.UUID `json:"stream_id" binding:"required"`

	Message string `json:"message" binding:"required"`
}

//
// Update Message Request
//
// Currently chat_messages does not have
// an editable message column workflow.
// This request is intentionally omitted.
//

//
// Delete Message Request
//

type DeleteMessageRequest struct {
	MessageID int64 `json:"message_id" binding:"required"`
}

//
// Chat Message Response
//

type MessageResponse struct {
	ID int64 `json:"id"`

	StreamID uuid.UUID `json:"stream_id"`

	UserID uuid.UUID `json:"user_id"`

	Message string `json:"message"`

	IsDeleted bool `json:"is_deleted"`

	DeletedAt *time.Time `json:"deleted_at,omitempty"`

	CreatedAt time.Time `json:"created_at"`
}

//
// Message List Response
//

type MessageListResponse struct {
	Messages []MessageResponse `json:"messages"`

	Total int64 `json:"total"`

	Page int `json:"page"`

	Limit int `json:"limit"`
}

//
// Message Filter
//

type MessageFilter struct {
	StreamID uuid.UUID

	UserID *uuid.UUID

	IncludeDeleted bool

	Page  int
	Limit int
}

//
// Message Owner
//
// chat_messages
//       ↓ user_id
//
// user_id belongs to Auth Service.
// There is intentionally no cross-database FK.
//

type MessageOwner struct {
	MessageID int64 `db:"message_id"`

	StreamID uuid.UUID `db:"stream_id"`

	UserID uuid.UUID `db:"user_id"`
}

//
// Message Statistics
//

type MessageStats struct {
	StreamID uuid.UUID `json:"stream_id"`

	TotalMessages int64 `json:"total_messages"`

	DeletedMessages int64 `json:"deleted_messages"`

	ActiveMessages int64 `json:"active_messages"`
}

//
// Response Conversion
//

func ToMessageResponse(
	message *ChatMessage,
) *MessageResponse {

	if message == nil {
		return nil
	}

	return &MessageResponse{
		ID:        message.ID,
		StreamID:  message.StreamID,
		UserID:    message.UserID,
		Message:   message.Message,
		IsDeleted: message.IsDeleted,
		DeletedAt: message.DeletedAt,
		CreatedAt: message.CreatedAt,
	}
}

//
// List Response Conversion
//

func ToMessageResponses(
	messages []ChatMessage,
) []MessageResponse {

	responses := make(
		[]MessageResponse,
		0,
		len(messages),
	)

	for i := range messages {
		response := ToMessageResponse(
			&messages[i],
		)

		if response != nil {
			responses = append(
				responses,
				*response,
			)
		}
	}

	return responses
}
