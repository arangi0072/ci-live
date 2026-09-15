package like

import (
	"time"

	"github.com/google/uuid"
)

//
// Like
//

type StreamLike struct {
	StreamID  uuid.UUID `json:"stream_id" db:"stream_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

//
// Create Like Request
//

type CreateLikeRequest struct {
	StreamID uuid.UUID `json:"stream_id" binding:"required"`
}

//
// Like Response
//

type LikeResponse struct {
	StreamID  uuid.UUID `json:"stream_id"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

//
// Like List Response
//

type LikeListResponse struct {
	Likes []LikeResponse `json:"likes"`
	Total int64          `json:"total"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
}

//
// Like Filter
//

type LikeFilter struct {
	StreamID *uuid.UUID
	UserID   *uuid.UUID
	Page     int
	Limit    int
}

//
// Like Owner
//

type LikeOwner struct {
	StreamID uuid.UUID `json:"stream_id"`
	UserID   uuid.UUID `json:"user_id"`
}

//
// Like Statistics
//

type LikeStats struct {
	StreamID  uuid.UUID `json:"stream_id"`
	LikeCount int64     `json:"like_count"`
}

//
// Convert StreamLike -> LikeResponse
//

func ToLikeResponse(
	like *StreamLike,
) *LikeResponse {

	if like == nil {
		return nil
	}

	return &LikeResponse{
		StreamID:  like.StreamID,
		UserID:    like.UserID,
		CreatedAt: like.CreatedAt,
	}
}

//
// Convert []StreamLike -> []LikeResponse
//

func ToLikeResponses(
	likes []StreamLike,
) []LikeResponse {

	if len(likes) == 0 {
		return []LikeResponse{}
	}

	responses := make(
		[]LikeResponse,
		0,
		len(likes),
	)

	for i := range likes {
		response := ToLikeResponse(&likes[i])

		if response != nil {
			responses = append(
				responses,
				*response,
			)
		}
	}

	return responses
}

//
// Stream Owner
//

type StreamOwner struct {
	StreamID  uuid.UUID `json:"stream_id"`
	ChannelID uuid.UUID `json:"channel_id"`
	UserID    uuid.UUID `json:"user_id"`
}
