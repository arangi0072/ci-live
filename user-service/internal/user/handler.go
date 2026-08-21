package user

import (
	"errors"
	"net/http"

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
// GET /api/v1/users/me
// ============================================================

func (h *Handler) GetMe(c *gin.Context) {
	userID, err := getAuthenticatedUserID(c)
	if err != nil {
		unauthorized(c)
		return
	}

	profile, err := h.service.GetProfile(
		c.Request.Context(),
		userID,
	)

	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    profile,
	})
}

// ============================================================
// GET /api/v1/users/:user_id
// Public profile
// ============================================================

func (h *Handler) GetByID(c *gin.Context) {
	userID := c.Param("user_id")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_USER_ID",
				"message": "user_id is required",
			},
		})
		return
	}

	profile, err := h.service.GetProfile(
		c.Request.Context(),
		userID,
	)

	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    profile,
	})
}

// ============================================================
// POST /api/v1/users/me
// Create initial profile
// ============================================================

func (h *Handler) CreateMe(c *gin.Context) {
	userID, err := getAuthenticatedUserID(c)
	if err != nil {
		unauthorized(c)
		return
	}

	var req CreateProfileRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "invalid profile information",
			},
		})
		return
	}

	profile, err := h.service.CreateProfile(
		c.Request.Context(),
		userID,
		req,
	)

	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    profile,
	})
}

// ============================================================
// PATCH /api/v1/users/me
// ============================================================

func (h *Handler) UpdateMe(c *gin.Context) {
	userID, err := getAuthenticatedUserID(c)
	if err != nil {
		unauthorized(c)
		return
	}

	var req UpdateProfileRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "invalid profile information",
			},
		})
		return
	}

	profile, err := h.service.UpdateProfile(
		c.Request.Context(),
		userID,
		req,
	)

	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    profile,
	})
}

// ============================================================
// GET /api/v1/users/username/:username/availability
// ============================================================

func (h *Handler) CheckUsername(c *gin.Context) {
	username := c.Param("username")

	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_USERNAME",
				"message": "username is required",
			},
		})
		return
	}

	result, err := h.service.CheckUsernameAvailability(
		c.Request.Context(),
		username,
	)

	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// ============================================================
// Error handling
// ============================================================

func handleError(
	c *gin.Context,
	err error,
) {
	switch {
	case errors.Is(err, ErrProfileNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "PROFILE_NOT_FOUND",
				"message": "user profile not found",
			},
		})

	case errors.Is(err, ErrProfileAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "PROFILE_ALREADY_EXISTS",
				"message": "user profile already exists",
			},
		})

	case errors.Is(err, ErrUsernameTaken):
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "USERNAME_TAKEN",
				"message": "username is already taken",
			},
		})

	case errors.Is(err, ErrInvalidUsername):
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_USERNAME",
				"message": "invalid username",
			},
		})

	case errors.Is(err, ErrInvalidProfile):
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_PROFILE",
				"message": "invalid profile information",
			},
		})

	default:
		c.Error(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "internal server error",
			},
		})
	}
}

// ============================================================
// Helpers
// ============================================================

func getAuthenticatedUserID(
	c *gin.Context,
) (string, error) {
	value, exists := c.Get("user_id")

	if !exists {
		return "", errors.New("user id not found")
	}

	userID, ok := value.(string)

	if !ok || userID == "" {
		return "", errors.New("invalid user id")
	}

	return userID, nil
}

func unauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(
		http.StatusUnauthorized,
		gin.H{
			"success": false,
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		},
	)
}
