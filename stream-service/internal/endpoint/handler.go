package endpoint

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var (
	ErrInvalidUUID = errors.New("invalid UUID")
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

//
// Create Endpoint
//
// POST /endpoints
//

func (h *Handler) CreateEndpoint(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err)
		return
	}

	var req CreateEndpointRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	endpoint, err := h.service.CreateEndpoint(
		c.Request.Context(),
		userID,
		req,
	)

	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    endpoint,
	})
}

//
// Get Endpoint
//
// GET /endpoints/:id
//

func (h *Handler) GetEndpoint(c *gin.Context) {
	endpointID, err := parseUUIDParam(c, "id")
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	endpoint, err := h.service.GetEndpoint(
		c.Request.Context(),
		endpointID,
	)

	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    endpoint,
	})
}

//
// List Endpoints
//
// GET /endpoints
//

func (h *Handler) ListEndpoints(c *gin.Context) {
	filter, err := parseEndpointFilter(c)
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	result, err := h.service.ListEndpoints(
		c.Request.Context(),
		filter,
	)

	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

//
// Get Stream Endpoints
//
// GET /streams/:stream_id/endpoints
//

func (h *Handler) GetStreamEndpoints(c *gin.Context) {
	streamID, err := parseUUIDParam(c, "stream_id")
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	filter := EndpointFilter{
		StreamID: &streamID,
	}

	result, err := h.service.ListEndpoints(
		c.Request.Context(),
		filter,
	)

	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

//
// Update Endpoint
//
// PATCH /endpoints/:id
//

func (h *Handler) UpdateEndpoint(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err)
		return
	}

	endpointID, err := parseUUIDParam(c, "id")
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	var req UpdateEndpointRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	endpoint, err := h.service.UpdateEndpoint(
		c.Request.Context(),
		userID,
		endpointID,
		req,
	)

	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    endpoint,
	})
}

//
// Activate Endpoint
//
// POST /endpoints/:id/activate
//

func (h *Handler) ActivateEndpoint(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err)
		return
	}

	endpointID, err := parseUUIDParam(c, "id")
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	endpoint, err := h.service.SetEndpointActive(
		c.Request.Context(),
		userID,
		endpointID,
		true,
	)

	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    endpoint,
	})
}

//
// Deactivate Endpoint
//
// POST /endpoints/:id/deactivate
//

func (h *Handler) DeactivateEndpoint(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err)
		return
	}

	endpointID, err := parseUUIDParam(c, "id")
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	endpoint, err := h.service.SetEndpointActive(
		c.Request.Context(),
		userID,
		endpointID,
		false,
	)

	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    endpoint,
	})
}

//
// Delete Endpoint
//
// DELETE /endpoints/:id
//

func (h *Handler) DeleteEndpoint(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err)
		return
	}

	endpointID, err := parseUUIDParam(c, "id")
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	if err := h.service.DeleteEndpoint(
		c.Request.Context(),
		userID,
		endpointID,
	); err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "endpoint deleted successfully",
	})
}

//
// Parse UUID Parameter
//

func parseUUIDParam(
	c *gin.Context,
	name string,
) (uuid.UUID, error) {

	value := c.Param(name)

	if value == "" {
		return uuid.Nil, ErrInvalidUUID
	}

	id, err := uuid.Parse(value)

	if err != nil || id == uuid.Nil {
		return uuid.Nil, ErrInvalidUUID
	}

	return id, nil
}

//
// Parse Endpoint Filter
//

func parseEndpointFilter(
	c *gin.Context,
) (EndpointFilter, error) {

	filter := EndpointFilter{}

	//
	// stream_id
	//

	if value := c.Query("stream_id"); value != "" {
		streamID, err := uuid.Parse(value)

		if err != nil || streamID == uuid.Nil {
			return filter, ErrInvalidUUID
		}

		filter.StreamID = &streamID
	}

	//
	// endpoint_type
	//

	if value := c.Query("endpoint_type"); value != "" {
		endpointType := EndpointType(value)

		if !endpointType.IsValid() {
			return filter, errors.New(
				"invalid endpoint_type",
			)
		}

		filter.EndpointType = &endpointType
	}

	//
	// protocol
	//

	if value := c.Query("protocol"); value != "" {
		protocol := EndpointProtocol(value)

		if !protocol.IsValid() {
			return filter, errors.New(
				"invalid protocol",
			)
		}

		filter.Protocol = &protocol
	}

	//
	// is_active
	//

	if value := c.Query("is_active"); value != "" {
		switch value {
		case "true":
			active := true
			filter.IsActive = &active

		case "false":
			active := false
			filter.IsActive = &active

		default:
			return filter, errors.New(
				"is_active must be true or false",
			)
		}
	}

	return filter, nil
}

//
// Get Authenticated User
//

func getUserID(c *gin.Context) (uuid.UUID, error) {
	value, exists := c.Get("user_id")

	if !exists {
		return uuid.Nil, ErrUnauthorized
	}

	switch v := value.(type) {

	case uuid.UUID:
		if v == uuid.Nil {
			return uuid.Nil, ErrUnauthorized
		}

		return v, nil

	case string:
		id, err := uuid.Parse(v)

		if err != nil || id == uuid.Nil {
			return uuid.Nil, ErrUnauthorized
		}

		return id, nil

	default:
		return uuid.Nil, ErrUnauthorized
	}
}

//
// Service Error Handler
//

func handleServiceError(
	c *gin.Context,
	err error,
) {
	switch {

	case errors.Is(err, ErrUnauthorized):
		respondError(
			c,
			http.StatusUnauthorized,
			err,
		)

	case errors.Is(err, ErrForbidden):
		respondError(
			c,
			http.StatusForbidden,
			err,
		)

	case errors.Is(err, ErrEndpointNotFound):
		respondError(
			c,
			http.StatusNotFound,
			err,
		)

	case errors.Is(err, ErrStreamNotFound):
		respondError(
			c,
			http.StatusNotFound,
			err,
		)

	case errors.Is(err, ErrInvalidEndpointType):
		respondError(
			c,
			http.StatusBadRequest,
			err,
		)

	case errors.Is(err, ErrInvalidProtocol):
		respondError(
			c,
			http.StatusBadRequest,
			err,
		)

	case errors.Is(err, ErrEndpointAlreadyExists):
		respondError(
			c,
			http.StatusConflict,
			err,
		)

	default:
		respondError(
			c,
			http.StatusInternalServerError,
			err,
		)
	}
}

//
// Error Response
//

func respondError(
	c *gin.Context,
	status int,
	err error,
) {
	c.JSON(status, gin.H{
		"success": false,
		"error": gin.H{
			"message": err.Error(),
		},
	})
}
