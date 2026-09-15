package like

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
// Create Like
//
// POST /streams/:stream_id/like
//

func (h *Handler) CreateLike(c *gin.Context) {
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

	req := CreateLikeRequest{
		StreamID: streamID,
	}

	like, err := h.service.CreateLike(
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
		"data":    like,
	})
}

//
// Remove Like
//
// DELETE /streams/:stream_id/like
//

func (h *Handler) RemoveLike(c *gin.Context) {
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

	err = h.service.RemoveLike(
		c.Request.Context(),
		userID,
		streamID,
	)

	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "like removed successfully",
	})
}

//
// Get Like
//
// GET /streams/:stream_id/like
//
// Returns the authenticated user's like.
//

func (h *Handler) GetLike(c *gin.Context) {
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

	like, err := h.service.GetLike(
		c.Request.Context(),
		userID,
		streamID,
	)

	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    like,
	})
}

//
// Check Like
//
// GET /streams/:stream_id/like/check
//

func (h *Handler) CheckLike(c *gin.Context) {
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

	liked, err := h.service.IsLiked(
		c.Request.Context(),
		userID,
		streamID,
	)

	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"stream_id": streamID,
			"liked":     liked,
		},
	})
}

//
// List Stream Likes
//
// GET /streams/:stream_id/likes
//

func (h *Handler) ListLikes(c *gin.Context) {
	streamID, err := parseUUIDParam(c, "stream_id")
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	filter, err := parseLikeFilter(c)
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	filter.StreamID = &streamID

	result, err := h.service.ListLikes(
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
// Get Like Statistics
//
// GET /streams/:stream_id/likes/stats
//

func (h *Handler) GetLikeStats(c *gin.Context) {
	streamID, err := parseUUIDParam(c, "stream_id")
	if err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	stats, err := h.service.GetLikeStats(
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
// Parse Like Filter
//

func parseLikeFilter(
	c *gin.Context,
) (LikeFilter, error) {

	filter := LikeFilter{
		Page:  1,
		Limit: 50,
	}

	//
	// Optional user filter
	//

	if value := c.Query("user_id"); value != "" {
		userID, err := uuid.Parse(value)

		if err != nil || userID == uuid.Nil {
			return filter, ErrInvalidUUID
		}

		filter.UserID = &userID
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

	case errors.Is(err, ErrLikeNotFound):
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

	case errors.Is(err, ErrAlreadyLiked):
		respondError(
			c,
			http.StatusConflict,
			err,
		)

	case errors.Is(err, ErrNotLiked):
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
