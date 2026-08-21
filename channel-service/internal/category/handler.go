package category

import (
	"errors"
	"net/http"

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
// GET /api/v1/categories
// ============================================================

func (h *Handler) List(c *gin.Context) {
	categories, err := h.service.List(
		c.Request.Context(),
	)

	if err != nil {
		handleError(c, err)
		return
	}

	response.OK(c, categories)
}

// ============================================================
// GET /api/v1/categories/:category_id
// ============================================================

func (h *Handler) GetByID(c *gin.Context) {
	categoryID := c.Param("category_id")

	if categoryID == "" {
		response.BadRequest(
			c,
			"INVALID_CATEGORY_ID",
			"category_id is required",
		)
		return
	}

	category, err := h.service.GetByID(
		c.Request.Context(),
		categoryID,
	)

	if err != nil {
		handleError(c, err)
		return
	}

	response.OK(c, category)
}

// ============================================================
// GET /api/v1/categories/slug/:slug
// ============================================================

func (h *Handler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")

	if slug == "" {
		response.BadRequest(
			c,
			"INVALID_CATEGORY_SLUG",
			"category slug is required",
		)
		return
	}

	category, err := h.service.GetBySlug(
		c.Request.Context(),
		slug,
	)

	if err != nil {
		handleError(c, err)
		return
	}

	response.OK(c, category)
}

// ============================================================
// POST /api/v1/categories
// ============================================================
//
// This should normally be an admin-only endpoint.
// JWT authentication alone is not enough if you later
// introduce normal users with access to this route.
// ============================================================

func (h *Handler) Create(c *gin.Context) {
	var req CreateCategoryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(
			c,
			"INVALID_REQUEST",
			"invalid category information",
		)
		return
	}

	category, err := h.service.Create(
		c.Request.Context(),
		req,
	)

	if err != nil {
		handleError(c, err)
		return
	}

	response.Created(c, category)
}

// ============================================================
// PATCH /api/v1/categories/:category_id
// ============================================================

func (h *Handler) Update(c *gin.Context) {
	categoryID := c.Param("category_id")

	if categoryID == "" {
		response.BadRequest(
			c,
			"INVALID_CATEGORY_ID",
			"category_id is required",
		)
		return
	}

	var req UpdateCategoryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(
			c,
			"INVALID_REQUEST",
			"invalid category information",
		)
		return
	}

	category, err := h.service.Update(
		c.Request.Context(),
		categoryID,
		req,
	)

	if err != nil {
		handleError(c, err)
		return
	}

	response.OK(c, category)
}

// ============================================================
// DELETE /api/v1/categories/:category_id
// ============================================================

func (h *Handler) Delete(c *gin.Context) {
	categoryID := c.Param("category_id")

	if categoryID == "" {
		response.BadRequest(
			c,
			"INVALID_CATEGORY_ID",
			"category_id is required",
		)
		return
	}

	if err := h.service.Delete(
		c.Request.Context(),
		categoryID,
	); err != nil {
		handleError(c, err)
		return
	}

	response.OK(c, gin.H{
		"message": "category deleted successfully",
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
	case errors.Is(err, ErrCategoryNotFound):
		response.NotFound(
			c,
			"CATEGORY_NOT_FOUND",
			"category not found",
		)

	case errors.Is(err, ErrCategoryExists):
		response.Conflict(
			c,
			"CATEGORY_EXISTS",
			"category already exists",
		)

	case errors.Is(err, ErrInvalidCategory):
		response.BadRequest(
			c,
			"INVALID_CATEGORY",
			"invalid category information",
		)

	default:
		// Log the actual error through Gin's error chain.
		_ = c.Error(err)

		response.InternalServerError(c)
	}
}

// Prevent unused import issues if the compiler checks
// http usage after future handler extensions.
var _ = http.StatusOK
