package endpoint

import (
	"time"

	"github.com/google/uuid"
)

//
// Endpoint Type
//

type EndpointType string

const (
	EndpointTypePublish  EndpointType = "PUBLISH"
	EndpointTypePlayback EndpointType = "PLAYBACK"
)

func (e EndpointType) IsValid() bool {
	switch e {
	case EndpointTypePublish, EndpointTypePlayback:
		return true
	default:
		return false
	}
}

//
// Endpoint Protocol
//

type EndpointProtocol string

const (
	EndpointProtocolRTMP EndpointProtocol = "RTMP"
	EndpointProtocolHLS  EndpointProtocol = "HLS"
)

func (p EndpointProtocol) IsValid() bool {
	switch p {
	case EndpointProtocolRTMP, EndpointProtocolHLS:
		return true
	default:
		return false
	}
}

//
// Stream Endpoint
//

type StreamEndpoint struct {
	ID uuid.UUID `json:"id" db:"id"`

	StreamID uuid.UUID `json:"stream_id" db:"stream_id"`

	EndpointType EndpointType `json:"endpoint_type" db:"endpoint_type"`

	Protocol EndpointProtocol `json:"protocol" db:"protocol"`

	URL string `json:"url" db:"url"`

	IsActive bool `json:"is_active" db:"is_active"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

//
// Create Endpoint Request
//

type CreateEndpointRequest struct {
	StreamID uuid.UUID `json:"stream_id" binding:"required"`

	EndpointType EndpointType `json:"endpoint_type" binding:"required"`

	Protocol EndpointProtocol `json:"protocol" binding:"required"`

	URL string `json:"url" binding:"required"`

	IsActive *bool `json:"is_active,omitempty"`
}

//
// Update Endpoint Request
//

type UpdateEndpointRequest struct {
	URL *string `json:"url,omitempty"`

	IsActive *bool `json:"is_active,omitempty"`
}

//
// Endpoint Response
//

type EndpointResponse struct {
	ID uuid.UUID `json:"id"`

	StreamID uuid.UUID `json:"stream_id"`

	EndpointType EndpointType `json:"endpoint_type"`

	Protocol EndpointProtocol `json:"protocol"`

	URL string `json:"url"`

	IsActive bool `json:"is_active"`

	CreatedAt time.Time `json:"created_at"`
}

//
// Endpoint List Response
//

type EndpointListResponse struct {
	Endpoints []EndpointResponse `json:"endpoints"`

	Total int64 `json:"total"`
}

//
// Endpoint Filter
//

type EndpointFilter struct {
	StreamID *uuid.UUID

	EndpointType *EndpointType

	Protocol *EndpointProtocol

	IsActive *bool
}

//
// Endpoint Owner
//
// stream_endpoints
//        ↓ stream_id
// streams
//        ↓ channel_id
// channels
//        ↓ user_id
//

type EndpointOwner struct {
	EndpointID uuid.UUID `db:"endpoint_id"`

	StreamID uuid.UUID `db:"stream_id"`

	ChannelID uuid.UUID `db:"channel_id"`

	UserID uuid.UUID `db:"user_id"`
}

//
// Response Conversion
//

func ToEndpointResponse(
	e *StreamEndpoint,
) *EndpointResponse {

	if e == nil {
		return nil
	}

	return &EndpointResponse{
		ID:           e.ID,
		StreamID:     e.StreamID,
		EndpointType: e.EndpointType,
		Protocol:     e.Protocol,
		URL:          e.URL,
		IsActive:     e.IsActive,
		CreatedAt:    e.CreatedAt,
	}
}

//
// List Response Conversion
//

func ToEndpointResponses(
	endpoints []StreamEndpoint,
) []EndpointResponse {

	responses := make(
		[]EndpointResponse,
		0,
		len(endpoints),
	)

	for i := range endpoints {
		response := ToEndpointResponse(
			&endpoints[i],
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
