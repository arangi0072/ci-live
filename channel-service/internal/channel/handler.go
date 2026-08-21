package channel

import (
	"errors"

	"github.com/gin-gonic/gin"

	"channel-service/internal/response"
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
// POST /api/v1/channels
// Create channel
// ============================================================

func (h *Handler) Create(c *gin.Context) {
	userID, err := getAuthenticatedUserID(c)
	if err != nil {
		response.Unauthorized(
			c,
			"UNAUTHORIZED",
			"authentication required",
		)
		return
	}

	var req CreateChannelRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(
			c,
			"INVALID_REQUEST",
			"invalid channel information",
		)
		return
	}

	channel, err := h.service.Create(
		c.Request.Context(),
		userID,
		req,
	)

	if err != nil {
		handleError(c, err)
		return
	}

	response.Created(c, channel)
}

// ============================================================
// GET /api/v1/channels/me
// Get authenticated user's channel
// ============================================================

func (h *Handler) GetMe(c *gin.Context) {
	userID, err := getAuthenticatedUserID(c)
	if err != nil {
		response.Unauthorized(
			c,
			"UNAUTHORIZED",
			"authentication required",
		)
		return
	}

	channel, err := h.service.GetMyChannel(
		c.Request.Context(),
		userID,
	)

	if err != nil {
		handleError(c, err)
		return
	}

	response.OK(c, channel)
}

// ============================================================
// GET /api/v1/channels/:channel_id
// Get public channel
// ============================================================

func (h *Handler) GetByID(c *gin.Context) {
	channelID := c.Param("channel_id")

	if channelID == "" {
		response.BadRequest(
			c,
			"INVALID_CHANNEL_ID",
			"channel_id is required",
		)
		return
	}

	channel, err := h.service.GetByID(
		c.Request.Context(),
		channelID,
	)

	if err != nil {
		handleError(c, err)
		return
	}

	response.OK(c, channel)
}

// ============================================================
// GET /api/v1/channels/username/:username
// Get channel by username
// ============================================================

func (h *Handler) GetByUsername(c *gin.Context) {
	username := c.Param("username")

	if username == "" {
		response.BadRequest(
			c,
			"INVALID_USERNAME",
			"username is required",
		)
		return
	}

	channel, err := h.service.GetByUsername(
		c.Request.Context(),
		username,
	)

	if err != nil {
		handleError(c, err)
		return
	}

	response.OK(c, channel)
}

// ============================================================
// PATCH /api/v1/channels/me
// Update authenticated user's channel
// ============================================================

func (h *Handler) Update(c *gin.Context) {
	userID, err := getAuthenticatedUserID(c)
	if err != nil {
		response.Unauthorized(
			c,
			"UNAUTHORIZED",
			"authentication required",
		)
		return
	}

	var req UpdateChannelRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(
			c,
			"INVALID_REQUEST",
			"invalid channel information",
		)
		return
	}

	channel, err := h.service.Update(
		c.Request.Context(),
		userID,
		req,
	)

	if err != nil {
		handleError(c, err)
		return
	}

	response.OK(c, channel)
}

// ============================================================
// Error handling
// ============================================================

func handleError(
	c *gin.Context,
	err error,
) {
	switch {
	case errors.Is(err, ErrChannelNotFound):
		response.NotFound(
			c,
			"CHANNEL_NOT_FOUND",
			"channel not found",
		)

	case errors.Is(err, ErrChannelExists):
		response.Conflict(
			c,
			"CHANNEL_EXISTS",
			"user already has a channel",
		)

	case errors.Is(err, ErrUsernameTaken):
		response.Conflict(
			c,
			"USERNAME_TAKEN",
			"channel username is already taken",
		)

	case errors.Is(err, ErrInvalidChannel):
		response.BadRequest(
			c,
			"INVALID_CHANNEL",
			"invalid channel information",
		)

	case errors.Is(err, ErrUnauthorizedChannel):
		response.Unauthorized(
			c,
			"UNAUTHORIZED_CHANNEL",
			"you are not authorized to modify this channel",
		)

	default:
		_ = c.Error(err)

		response.InternalServerError(c)
	}
}

// ============================================================
// Authentication helper
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
