package topic

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
// POST /api/v1/topics
// Create Topic
// ============================================================

func (h *Handler) Create(c *gin.Context) {
	var req CreateTopicRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(
			c,
			"INVALID_REQUEST",
			"invalid topic information",
		)
		return
	}

	topic, err := h.service.Create(
		c.Request.Context(),
		req,
	)

	if err != nil {
		handleError(c, err)
		return
	}

	response.Created(c, topic)
}

// ============================================================
// GET /api/v1/topics/:topic_id
// Get Topic By ID
// ============================================================

func (h *Handler) GetByID(c *gin.Context) {
	topicID := c.Param("topic_id")

	if topicID == "" {
		response.BadRequest(
			c,
			"INVALID_TOPIC_ID",
			"topic_id is required",
		)
		return
	}

	topic, err := h.service.GetByID(
		c.Request.Context(),
		topicID,
	)

	if err != nil {
		handleError(c, err)
		return
	}

	response.OK(c, topic)
}

// ============================================================
// GET /api/v1/topics/slug/:slug
// Get Topic By Slug
// ============================================================

func (h *Handler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")

	if slug == "" {
		response.BadRequest(
			c,
			"INVALID_TOPIC_SLUG",
			"topic slug is required",
		)
		return
	}

	topic, err := h.service.GetBySlug(
		c.Request.Context(),
		slug,
	)

	if err != nil {
		handleError(c, err)
		return
	}

	response.OK(c, topic)
}

// ============================================================
// GET /api/v1/topics/category/:category_id
// List Topics By Category
// ============================================================

func (h *Handler) List(c *gin.Context) {
	categoryID := c.Query("category_id")

	// category_id is optional.
	//
	// Example:
	// GET /api/v1/topics
	//
	// or:
	// GET /api/v1/topics?category_id=<uuid>

	topics, err := h.service.List(
		c.Request.Context(),
		categoryID,
	)

	if err != nil {
		handleError(c, err)
		return
	}

	response.OK(c, topics)
}

// ============================================================
// PATCH /api/v1/topics/:topic_id
// Update Topic
// ============================================================

func (h *Handler) Update(c *gin.Context) {
	topicID := c.Param("topic_id")

	if topicID == "" {
		response.BadRequest(
			c,
			"INVALID_TOPIC_ID",
			"topic_id is required",
		)
		return
	}

	var req UpdateTopicRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(
			c,
			"INVALID_REQUEST",
			"invalid topic information",
		)
		return
	}

	topic, err := h.service.Update(
		c.Request.Context(),
		topicID,
		req,
	)

	if err != nil {
		handleError(c, err)
		return
	}

	response.OK(c, topic)
}

// ============================================================
// DELETE /api/v1/topics/:topic_id
// Delete Topic
// ============================================================

func (h *Handler) Delete(c *gin.Context) {
	topicID := c.Param("topic_id")

	if topicID == "" {
		response.BadRequest(
			c,
			"INVALID_TOPIC_ID",
			"topic_id is required",
		)
		return
	}

	if err := h.service.Delete(
		c.Request.Context(),
		topicID,
	); err != nil {
		handleError(c, err)
		return
	}

	response.OK(c, gin.H{
		"message": "topic deleted successfully",
	})
}

// ============================================================
// Error Handler
// ============================================================

func handleError(
	c *gin.Context,
	err error,
) {
	switch {
	case errors.Is(err, ErrTopicNotFound):
		response.NotFound(
			c,
			"TOPIC_NOT_FOUND",
			"topic not found",
		)

	case errors.Is(err, ErrTopicExists):
		response.Conflict(
			c,
			"TOPIC_EXISTS",
			"topic already exists",
		)

	case errors.Is(err, ErrInvalidTopic):
		response.BadRequest(
			c,
			"INVALID_TOPIC",
			"invalid topic information",
		)

	default:
		_ = c.Error(err)

		response.InternalServerError(c)
	}
}
