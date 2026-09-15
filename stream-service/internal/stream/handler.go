package stream

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
// Create Stream
//
// POST /streams
//

func (h *Handler) CreateStream(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err)
		return
	}

	var req CreateStreamRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(
			c,
			http.StatusBadRequest,
			err,
		)
		return
	}

	stream, err := h.service.CreateStream(
		c.Request.Context(),
		userID,
		req,
	)

	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(
		http.StatusCreated,
		gin.H{
			"success": true,
			"data":    stream,
		},
	)
}

//
// Get Stream
//
// GET /streams/:id
//

func (h *Handler) GetStream(c *gin.Context) {
	streamID, err := parseUUIDParam(c, "id")
	if err != nil {
		respondError(
			c,
			http.StatusBadRequest,
			err,
		)
		return
	}

	stream, err := h.service.GetStream(
		c.Request.Context(),
		streamID,
	)

	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"success": true,
			"data":    stream,
		},
	)
}

//
// Get Public Stream
//
// GET /streams/:id/public
//

func (h *Handler) GetPublicStream(c *gin.Context) {
	streamID, err := parseUUIDParam(c, "id")
	if err != nil {
		respondError(
			c,
			http.StatusBadRequest,
			err,
		)
		return
	}

	stream, err := h.service.GetPublicStream(
		c.Request.Context(),
		streamID,
	)

	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"success": true,
			"data":    stream,
		},
	)
}

//
// List Streams
//
// GET /streams
//

func (h *Handler) ListStreams(c *gin.Context) {
	filter, err := parseStreamFilter(c)
	if err != nil {
		respondError(
			c,
			http.StatusBadRequest,
			err,
		)
		return
	}

	result, err := h.service.ListStreams(
		c.Request.Context(),
		filter,
	)

	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"success": true,
			"data":    result,
		},
	)
}

//
// Update Stream
//
// PATCH /streams/:id
//

func (h *Handler) UpdateStream(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		respondError(
			c,
			http.StatusUnauthorized,
			err,
		)
		return
	}

	streamID, err := parseUUIDParam(c, "id")
	if err != nil {
		respondError(
			c,
			http.StatusBadRequest,
			err,
		)
		return
	}

	var req UpdateStreamRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(
			c,
			http.StatusBadRequest,
			err,
		)
		return
	}

	stream, err := h.service.UpdateStream(
		c.Request.Context(),
		userID,
		streamID,
		req,
	)

	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"success": true,
			"data":    stream,
		},
	)
}

//
// Delete Stream
//
// DELETE /streams/:id
//

func (h *Handler) DeleteStream(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		respondError(
			c,
			http.StatusUnauthorized,
			err,
		)
		return
	}

	streamID, err := parseUUIDParam(c, "id")
	if err != nil {
		respondError(
			c,
			http.StatusBadRequest,
			err,
		)
		return
	}

	if err := h.service.DeleteStream(
		c.Request.Context(),
		userID,
		streamID,
	); err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"success": true,
			"message": "stream deleted successfully",
		},
	)
}

//
// Start Stream
//
// POST /streams/:id/start
//

func (h *Handler) StartStream(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		respondError(
			c,
			http.StatusUnauthorized,
			err,
		)
		return
	}

	streamID, err := parseUUIDParam(c, "id")
	if err != nil {
		respondError(
			c,
			http.StatusBadRequest,
			err,
		)
		return
	}

	stream, err := h.service.StartStream(
		c.Request.Context(),
		userID,
		streamID,
	)

	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"success": true,
			"data":    stream,
		},
	)
}

//
// End Stream
//
// POST /streams/:id/end
//

func (h *Handler) EndStream(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		respondError(
			c,
			http.StatusUnauthorized,
			err,
		)
		return
	}

	streamID, err := parseUUIDParam(c, "id")
	if err != nil {
		respondError(
			c,
			http.StatusBadRequest,
			err,
		)
		return
	}

	stream, err := h.service.EndStream(
		c.Request.Context(),
		userID,
		streamID,
	)

	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"success": true,
			"data":    stream,
		},
	)
}

//
// Update Stream Status
//
// PATCH /streams/:id/status
//

func (h *Handler) UpdateStreamStatus(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		respondError(
			c,
			http.StatusUnauthorized,
			err,
		)
		return
	}

	streamID, err := parseUUIDParam(c, "id")
	if err != nil {
		respondError(
			c,
			http.StatusBadRequest,
			err,
		)
		return
	}

	var req UpdateStreamStatusRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(
			c,
			http.StatusBadRequest,
			err,
		)
		return
	}

	stream, err := h.service.UpdateStreamStatus(
		c.Request.Context(),
		userID,
		streamID,
		req,
	)

	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"success": true,
			"data":    stream,
		},
	)
}

//
// Get Stream Statistics
//
// GET /streams/:id/stats
//

func (h *Handler) GetStreamStats(c *gin.Context) {
	streamID, err := parseUUIDParam(c, "id")
	if err != nil {
		respondError(
			c,
			http.StatusBadRequest,
			err,
		)
		return
	}

	stats, err := h.service.GetStreamStats(
		c.Request.Context(),
		streamID,
	)

	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"success": true,
			"data":    stats,
		},
	)
}

//
// Increment View
//
// POST /streams/:id/view
//

func (h *Handler) IncrementView(c *gin.Context) {
	streamID, err := parseUUIDParam(c, "id")
	if err != nil {
		respondError(
			c,
			http.StatusBadRequest,
			err,
		)
		return
	}

	stats, err := h.service.IncrementView(
		c.Request.Context(),
		streamID,
	)

	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"success": true,
			"data":    stats,
		},
	)
}

//
// Update Viewer Count
//
// PATCH /streams/:id/viewers
//

func (h *Handler) UpdateViewerCount(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		respondError(
			c,
			http.StatusUnauthorized,
			err,
		)
		return
	}

	streamID, err := parseUUIDParam(c, "id")
	if err != nil {
		respondError(
			c,
			http.StatusBadRequest,
			err,
		)
		return
	}

	var body struct {
		ViewerCount int `json:"viewer_count" binding:"required,min=0"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		respondError(
			c,
			http.StatusBadRequest,
			err,
		)
		return
	}

	stats, err := h.service.UpdateViewerCount(
		c.Request.Context(),
		userID,
		streamID,
		body.ViewerCount,
	)

	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"success": true,
			"data":    stats,
		},
	)
}

//
// Helper: User ID
//

func getUserID(c *gin.Context) (uuid.UUID, error) {
	value, exists := c.Get("user_id")

	if !exists {
		return uuid.Nil, errors.New(
			"user authentication required",
		)
	}

	switch v := value.(type) {

	case uuid.UUID:
		if v == uuid.Nil {
			return uuid.Nil, errors.New(
				"invalid authenticated user",
			)
		}

		return v, nil

	case string:
		id, err := uuid.Parse(v)
		if err != nil {
			return uuid.Nil, errors.New(
				"invalid authenticated user",
			)
		}

		return id, nil

	default:
		return uuid.Nil, errors.New(
			"invalid authenticated user",
		)
	}
}

//
// Helper: UUID path parameter
//

func parseUUIDParam(
	c *gin.Context,
	name string,
) (uuid.UUID, error) {

	value := c.Param(name)

	if value == "" {
		return uuid.Nil, errors.New(
			"missing " + name,
		)
	}

	id, err := uuid.Parse(value)

	if err != nil {
		return uuid.Nil, errors.New(
			"invalid " + name,
		)
	}

	return id, nil
}

//
// Helper: Stream Filters
//

func parseStreamFilter(
	c *gin.Context,
) (StreamFilter, error) {

	filter := StreamFilter{
		Page:  1,
		Limit: 20,
	}

	//
	// channel_id
	//

	if value := c.Query("channel_id"); value != "" {
		id, err := uuid.Parse(value)

		if err != nil {
			return filter, errors.New(
				"invalid channel_id",
			)
		}

		filter.ChannelID = &id
	}

	//
	// category_id
	//

	if value := c.Query("category_id"); value != "" {
		id, err := uuid.Parse(value)

		if err != nil {
			return filter, errors.New(
				"invalid category_id",
			)
		}

		filter.CategoryID = &id
	}

	//
	// status
	//

	if value := c.Query("status"); value != "" {
		status := StreamStatus(value)

		if !status.IsValid() {
			return filter, errors.New(
				"invalid stream status",
			)
		}

		filter.Status = &status
	}

	//
	// visibility
	//

	if value := c.Query("visibility"); value != "" {
		visibility := StreamVisibility(value)

		if !visibility.IsValid() {
			return filter, errors.New(
				"invalid stream visibility",
			)
		}

		filter.Visibility = &visibility
	}

	//
	// page
	//

	if value := c.Query("page"); value != "" {
		page, err := strconv.Atoi(value)

		if err != nil || page < 1 {
			return filter, errors.New(
				"invalid page",
			)
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
// Error Handling
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

	case errors.Is(err, ErrStreamNotFound):
		respondError(
			c,
			http.StatusNotFound,
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
// Generic Error Response
//

func respondError(
	c *gin.Context,
	status int,
	err error,
) {

	c.JSON(
		status,
		gin.H{
			"success": false,
			"error": gin.H{
				"message": err.Error(),
			},
		},
	)
}
