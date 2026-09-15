package chat

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

//
// Handler
//

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

//
// Create Message
//
// POST /streams/:stream_id/chat/messages
//

func (h *Handler) CreateMessage(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err)
		return
	}

	streamID, err := parseUUIDParam(c, "stream_id")
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	var req CreateMessageRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	//
	// The stream ID comes from the URL.
	//

	req.StreamID = streamID

	message, err := h.service.CreateMessage(
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
		"data":    message,
	})
}

//
// Get Message
//
// GET /chat/messages/:id
//

func (h *Handler) GetMessage(c *gin.Context) {
	messageID, err := parseMessageIDParam(c, "id")
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	message, err := h.service.GetMessage(
		c.Request.Context(),
		messageID,
	)

	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    message,
	})
}

//
// List Messages
//
// GET /streams/:stream_id/chat/messages
//

func (h *Handler) ListMessages(c *gin.Context) {
	streamID, err := parseUUIDParam(c, "stream_id")
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	filter, err := parseMessageFilter(c)
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	filter.StreamID = streamID

	result, err := h.service.ListMessages(
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
// Delete Message
//
// DELETE /chat/messages/:id
//

func (h *Handler) DeleteMessage(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err)
		return
	}

	messageID, err := parseMessageIDParam(c, "id")
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	err = h.service.DeleteMessage(
		c.Request.Context(),
		userID,
		messageID,
	)

	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "message deleted successfully",
	})
}

//
// Get Stream Message Statistics
//
// GET /streams/:stream_id/chat/stats
//

func (h *Handler) GetMessageStats(c *gin.Context) {
	streamID, err := parseUUIDParam(c, "stream_id")
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	stats, err := h.service.GetMessageStats(
		c.Request.Context(),
		streamID,
	)

	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
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
// Parse Message ID
//

func parseMessageIDParam(
	c *gin.Context,
	name string,
) (int64, error) {

	value := c.Param(name)

	if value == "" {
		return 0, ErrInvalidMessageID
	}

	id, err := strconv.ParseInt(value, 10, 64)

	if err != nil || id <= 0 {
		return 0, ErrInvalidMessageID
	}

	return id, nil
}

//
// Parse Message Filter
//

func parseMessageFilter(
	c *gin.Context,
) (MessageFilter, error) {

	filter := MessageFilter{
		Page:  1,
		Limit: 50,
	}

	//
	// User ID
	//

	if value := c.Query("user_id"); value != "" {
		userID, err := uuid.Parse(value)

		if err != nil || userID == uuid.Nil {
			return filter, ErrInvalidUUID
		}

		filter.UserID = &userID
	}

	//
	// Include deleted messages
	//

	if value := c.Query("include_deleted"); value != "" {

		switch value {
		case "true":
			filter.IncludeDeleted = true

		case "false":
			filter.IncludeDeleted = false

		default:
			return filter, errors.New(
				"invalid include_deleted value",
			)
		}
	}

	//
	// Page
	//

	if value := c.Query("page"); value != "" {
		page, err := strconv.Atoi(value)

		if err != nil || page < 1 {
			return filter, errors.New("invalid page")
		}

		filter.Page = page
	}

	//
	// Limit
	//

	if value := c.Query("limit"); value != "" {
		limit, err := strconv.Atoi(value)

		if err != nil || limit < 1 || limit > 100 {
			return filter, errors.New("invalid limit")
		}

		filter.Limit = limit
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
// Handle Service Errors
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

	case errors.Is(err, ErrMessageNotFound):
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

	case errors.Is(err, ErrInvalidMessageID):
		respondError(
			c,
			http.StatusBadRequest,
			err,
		)

	case errors.Is(err, ErrInvalidUUID):
		respondError(
			c,
			http.StatusBadRequest,
			err,
		)

	case errors.Is(err, ErrMessageTooLong):
		respondError(
			c,
			http.StatusBadRequest,
			err,
		)

	case errors.Is(err, ErrMessageEmpty):
		respondError(
			c,
			http.StatusBadRequest,
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
