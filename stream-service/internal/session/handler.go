package session

import (
	"errors"
	"net/http"
	"strconv"

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
// Create Session
//
// POST /sessions
//

func (h *Handler) CreateSession(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err)
		return
	}

	var req CreateSessionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	session, err := h.service.CreateSession(
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
		"data":    session,
	})
}

//
// Get Session
//
// GET /sessions/:id
//

func (h *Handler) GetSession(c *gin.Context) {
	sessionID, err := parseUUIDParam(c, "id")
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	session, err := h.service.GetSession(
		c.Request.Context(),
		sessionID,
	)

	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    session,
	})
}

//
// List Sessions
//
// GET /sessions
//

func (h *Handler) ListSessions(c *gin.Context) {
	filter, err := parseSessionFilter(c)
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	result, err := h.service.ListSessions(
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
// Get Stream Sessions
//
// GET /streams/:stream_id/sessions
//

func (h *Handler) GetStreamSessions(c *gin.Context) {
	streamID, err := parseUUIDParam(c, "stream_id")
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	filter := SessionFilter{
		StreamID: &streamID,
		Page:     1,
		Limit:    20,
	}

	if value := c.Query("page"); value != "" {
		page, err := strconv.Atoi(value)

		if err != nil || page < 1 {
			respondError(
				c,
				http.StatusBadRequest,
				errors.New("invalid page"),
			)
			return
		}

		filter.Page = page
	}

	if value := c.Query("limit"); value != "" {
		limit, err := strconv.Atoi(value)

		if err != nil || limit < 1 || limit > 100 {
			respondError(
				c,
				http.StatusBadRequest,
				errors.New("limit must be between 1 and 100"),
			)
			return
		}

		filter.Limit = limit
	}

	result, err := h.service.ListSessions(
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
// Update Session
//
// PATCH /sessions/:id
//

func (h *Handler) UpdateSession(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err)
		return
	}

	sessionID, err := parseUUIDParam(c, "id")
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	var req UpdateSessionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	session, err := h.service.UpdateSession(
		c.Request.Context(),
		userID,
		sessionID,
		req,
	)

	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    session,
	})
}

//
// End Session
//
// POST /sessions/:id/end
//

func (h *Handler) EndSession(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err)
		return
	}

	sessionID, err := parseUUIDParam(c, "id")
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	var req EndSessionRequest

	// Body is optional.
	if c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			respondError(c, http.StatusBadRequest, err)
			return
		}
	}

	session, err := h.service.EndSession(
		c.Request.Context(),
		userID,
		sessionID,
		req,
	)

	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    session,
	})
}

//
// Get Session Statistics
//
// GET /sessions/:id/stats
//

func (h *Handler) GetSessionStats(c *gin.Context) {
	sessionID, err := parseUUIDParam(c, "id")
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	stats, err := h.service.GetSessionStats(
		c.Request.Context(),
		sessionID,
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
// Delete Session
//
// DELETE /sessions/:id
//

func (h *Handler) DeleteSession(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err)
		return
	}

	sessionID, err := parseUUIDParam(c, "id")
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	if err := h.service.DeleteSession(
		c.Request.Context(),
		userID,
		sessionID,
	); err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "session deleted successfully",
	})
}

//
// Parse UUID
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

	if err != nil {
		return uuid.Nil, ErrInvalidUUID
	}

	return id, nil
}

//
// Parse Session Filter
//

func parseSessionFilter(
	c *gin.Context,
) (SessionFilter, error) {

	filter := SessionFilter{
		Page:  1,
		Limit: 20,
	}

	//
	// stream_id
	//

	if value := c.Query("stream_id"); value != "" {
		streamID, err := uuid.Parse(value)

		if err != nil {
			return filter, ErrInvalidUUID
		}

		filter.StreamID = &streamID
	}

	//
	// page
	//

	if value := c.Query("page"); value != "" {
		page, err := strconv.Atoi(value)

		if err != nil || page < 1 {
			return filter, errors.New("invalid page")
		}

		filter.Page = page
	}

	//
	// limit
	//

	if value := c.Query("limit"); value != "" {
		limit, err := strconv.Atoi(value)

		if err != nil || limit < 1 || limit > 100 {
			return filter, errors.New(
				"limit must be between 1 and 100",
			)
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

	case errors.Is(err, ErrSessionNotFound):
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
