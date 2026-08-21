package follow

import (
	"errors"
	"net/http"

	"channel-service/internal/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// ============================================================
// POST /api/v1/channels/:channel_id/follow
// ============================================================

func (h *Handler) Follow(c *gin.Context) {
	userID, err := getAuthenticatedUserID(c)
	if err != nil {
		response.Unauthorized(
			c,
			"UNAUTHORIZED",
			"authentication required",
		)
		return
	}

	channelID := c.Param("channel_id")

	if channelID == "" {
		response.BadRequest(
			c,
			"INVALID_CHANNEL_ID",
			"channel_id is required",
		)
		return
	}

	err = h.service.Follow(
		c.Request.Context(),
		userID,
		channelID,
	)

	if err != nil {
		handleError(c, err)
		return
	}

	response.OK(c, gin.H{
		"following": true,
	})
}

// ============================================================
// DELETE /api/v1/channels/:channel_id/follow
// ============================================================

func (h *Handler) Unfollow(c *gin.Context) {
	userID, err := getAuthenticatedUserID(c)
	if err != nil {
		response.Unauthorized(
			c,
			"UNAUTHORIZED",
			"authentication required",
		)
		return
	}

	channelID := c.Param("channel_id")

	if channelID == "" {
		response.BadRequest(
			c,
			"INVALID_CHANNEL_ID",
			"channel_id is required",
		)
		return
	}

	err = h.service.Unfollow(
		c.Request.Context(),
		userID,
		channelID,
	)

	if err != nil {
		handleError(c, err)
		return
	}

	response.OK(c, gin.H{
		"following": false,
	})
}

// ============================================================
// GET /api/v1/channels/:channel_id/follow
// Check whether current user follows channel
// ============================================================

func (h *Handler) IsFollowing(c *gin.Context) {
	userID, err := getAuthenticatedUserID(c)
	if err != nil {
		response.Unauthorized(
			c,
			"UNAUTHORIZED",
			"authentication required",
		)
		return
	}

	channelID := c.Param("channel_id")

	if channelID == "" {
		response.BadRequest(
			c,
			"INVALID_CHANNEL_ID",
			"channel_id is required",
		)
		return
	}

	following, err := h.service.IsFollowing(
		c.Request.Context(),
		userID,
		channelID,
	)

	if err != nil {
		handleError(c, err)
		return
	}

	response.OK(c, gin.H{
		"following": following,
	})
}

// ============================================================
// GET /api/v1/channels/:channel_id/followers/count
// ============================================================

func (h *Handler) GetFollowerCount(c *gin.Context) {
	channelID := c.Param("channel_id")

	if channelID == "" {
		response.BadRequest(
			c,
			"INVALID_CHANNEL_ID",
			"channel_id is required",
		)
		return
	}

	count, err := h.service.GetFollowerCount(
		c.Request.Context(),
		channelID,
	)

	if err != nil {
		handleError(c, err)
		return
	}

	response.OK(c, gin.H{
		"follower_count": count,
	})
}

// ============================================================
// GET /api/v1/users/me/following
// ============================================================

func (h *Handler) GetFollowing(c *gin.Context) {
	userID, err := getAuthenticatedUserID(c)
	if err != nil {
		response.Unauthorized(
			c,
			"UNAUTHORIZED",
			"authentication required",
		)
		return
	}

	following, err := h.service.GetFollowing(
		c.Request.Context(),
		userID,
	)

	if err != nil {
		handleError(c, err)
		return
	}

	response.OK(c, following)
}

// ============================================================
// Authentication Helper
// ============================================================

func getAuthenticatedUserID(
	c *gin.Context,
) (string, error) {
	value, exists := c.Get("user_id")

	if !exists {
		return "", errors.New(
			"user_id not found in context",
		)
	}

	userID, ok := value.(string)

	if !ok || userID == "" {
		return "", errors.New(
			"invalid user_id",
		)
	}

	return userID, nil
}

// ============================================================
// Error Handler
// ============================================================

func handleError(
	c *gin.Context,
	err error,
) {
	switch {
	case errors.Is(err, ErrAlreadyFollowing):
		response.Conflict(
			c,
			"ALREADY_FOLLOWING",
			"you are already following this channel",
		)

	case errors.Is(err, ErrFollowNotFound):
		response.NotFound(
			c,
			"FOLLOW_NOT_FOUND",
			"follow relationship not found",
		)

	default:
		_ = c.Error(err)

		response.Error(
			c,
			http.StatusInternalServerError,
			"INTERNAL_SERVER_ERROR",
			"internal server error",
		)
	}
}
