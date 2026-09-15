package recording

import (
	"errors"
	"fmt"
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
// Create Recording
//
// POST /recordings
//

func (h *Handler) CreateRecording(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err)
		return
	}

	var req CreateRecordingRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	recording, err := h.service.CreateRecording(
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
		"data":    recording,
	})
}

//
// Get Recording
//
// GET /recordings/:id
//

func (h *Handler) GetRecording(c *gin.Context) {
	recordingID, err := parseUUIDParam(c, "id")
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	recording, err := h.service.GetRecording(
		c.Request.Context(),
		recordingID,
	)

	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    recording,
	})
}

//
// List Recordings
//
// GET /recordings
//

func (h *Handler) ListRecordings(c *gin.Context) {
	filter, err := parseRecordingFilter(c)
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	result, err := h.service.ListRecordings(
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
// Get Stream Recordings
//
// GET /streams/:stream_id/recordings
//

func (h *Handler) GetStreamRecordings(c *gin.Context) {
	streamID, err := parseUUIDParam(c, "stream_id")
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	filter := RecordingFilter{
		StreamID: &streamID,
		Page:     1,
		Limit:    20,
	}

	//
	// Allow pagination.
	//

	if page := c.Query("page"); page != "" {
		value, err := strconv.Atoi(page)
		if err != nil || value < 1 {
			respondError(
				c,
				http.StatusBadRequest,
				errors.New("invalid page"),
			)
			return
		}

		filter.Page = value
	}

	if limit := c.Query("limit"); limit != "" {
		value, err := strconv.Atoi(limit)
		if err != nil || value < 1 || value > 100 {
			respondError(
				c,
				http.StatusBadRequest,
				errors.New("invalid limit"),
			)
			return
		}

		filter.Limit = value
	}

	result, err := h.service.ListRecordings(
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
// Get Session Recordings
//
// GET /sessions/:session_id/recordings
//

func (h *Handler) GetSessionRecordings(c *gin.Context) {
	sessionID, err := parseUUIDParam(c, "session_id")
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	filter := RecordingFilter{
		SessionID: &sessionID,
		Page:      1,
		Limit:     20,
	}

	if page := c.Query("page"); page != "" {
		value, err := strconv.Atoi(page)
		if err != nil || value < 1 {
			respondError(
				c,
				http.StatusBadRequest,
				errors.New("invalid page"),
			)
			return
		}

		filter.Page = value
	}

	if limit := c.Query("limit"); limit != "" {
		value, err := strconv.Atoi(limit)
		if err != nil || value < 1 || value > 100 {
			respondError(
				c,
				http.StatusBadRequest,
				errors.New("invalid limit"),
			)
			return
		}

		filter.Limit = value
	}

	result, err := h.service.ListRecordings(
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
// Update Recording
//
// PATCH /recordings/:id
//

func (h *Handler) UpdateRecording(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err)
		return
	}

	recordingID, err := parseUUIDParam(c, "id")
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	var req UpdateRecordingRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	recording, err := h.service.UpdateRecording(
		c.Request.Context(),
		userID,
		recordingID,
		req,
	)

	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    recording,
	})
}

//
// Update Recording Status
//
// PATCH /recordings/:id/status
//

func (h *Handler) UpdateRecordingStatus(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err)
		return
	}

	recordingID, err := parseUUIDParam(c, "id")
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	var req UpdateRecordingStatusRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	recording, err := h.service.UpdateRecordingStatus(
		c.Request.Context(),
		userID,
		recordingID,
		req,
	)

	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    recording,
	})
}

//
// Delete Recording
//
// DELETE /recordings/:id
//

func (h *Handler) DeleteRecording(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err)
		return
	}

	recordingID, err := parseUUIDParam(c, "id")
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	err = h.service.DeleteRecording(
		c.Request.Context(),
		userID,
		recordingID,
	)

	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "recording deleted successfully",
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
// Parse Recording Filter
//

func parseRecordingFilter(
	c *gin.Context,
) (RecordingFilter, error) {

	filter := RecordingFilter{
		Page:  1,
		Limit: 20,
	}

	//
	// Stream ID
	//

	if value := c.Query("stream_id"); value != "" {
		streamID, err := uuid.Parse(value)

		if err != nil || streamID == uuid.Nil {
			return filter, ErrInvalidUUID
		}

		filter.StreamID = &streamID
	}

	//
	// Session ID
	//

	if value := c.Query("session_id"); value != "" {
		sessionID, err := uuid.Parse(value)

		if err != nil || sessionID == uuid.Nil {
			return filter, ErrInvalidUUID
		}

		filter.SessionID = &sessionID
	}

	//
	// Status
	//

	if value := c.Query("status"); value != "" {
		status := RecordingStatus(value)

		if !status.IsValid() {
			return filter, ErrInvalidRecordingStatus
		}

		filter.Status = &status
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
// Get Authenticated User ID
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

	case errors.Is(err, ErrRecordingNotFound):
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

	case errors.Is(err, ErrSessionNotFound):
		respondError(
			c,
			http.StatusNotFound,
			err,
		)

	case errors.Is(err, ErrInvalidRecordingStatus):
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

//
// Keep fmt available for packages that extend
// handler validation without changing imports.
//

var _ = fmt.Sprintf
