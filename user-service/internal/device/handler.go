package device

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
// POST /api/v1/devices
// Register or update a user's device
// ============================================================

func (h *Handler) Register(c *gin.Context) {
	var req RegisterDeviceRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "invalid device information",
			},
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "authentication required",
			},
		})
		return
	}

	userIDString, ok := userID.(string)
	if !ok || userIDString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "invalid authenticated user",
			},
		})
		return
	}

	device, err := h.service.Register(
		c.Request.Context(),
		userIDString,
		req,
	)

	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    device,
	})
}

// ============================================================
// GET /api/v1/devices
// Get all devices belonging to authenticated user
// ============================================================

func (h *Handler) List(c *gin.Context) {
	userID, exists := c.Get("user_id")

	if !exists {
		unauthorized(c)
		return
	}

	userIDString, ok := userID.(string)
	if !ok || userIDString == "" {
		unauthorized(c)
		return
	}

	devices, err := h.service.List(
		c.Request.Context(),
		userIDString,
	)

	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    devices,
	})
}

// ============================================================
// PATCH /api/v1/devices/:device_id
// Update device information
// ============================================================

func (h *Handler) Update(c *gin.Context) {
	deviceID := c.Param("device_id")

	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_DEVICE_ID",
				"message": "device_id is required",
			},
		})
		return
	}

	userID, exists := c.Get("user_id")

	if !exists {
		unauthorized(c)
		return
	}

	userIDString, ok := userID.(string)
	if !ok || userIDString == "" {
		unauthorized(c)
		return
	}

	var req UpdateDeviceRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "invalid device information",
			},
		})
		return
	}

	device, err := h.service.Update(
		c.Request.Context(),
		userIDString,
		deviceID,
		req,
	)

	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    device,
	})
}

// ============================================================
// DELETE /api/v1/devices/:device_id
// Remove/deactivate device
// ============================================================

func (h *Handler) Delete(c *gin.Context) {
	deviceID := c.Param("device_id")

	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_DEVICE_ID",
				"message": "device_id is required",
			},
		})
		return
	}

	userID, exists := c.Get("user_id")

	if !exists {
		unauthorized(c)
		return
	}

	userIDString, ok := userID.(string)
	if !ok || userIDString == "" {
		unauthorized(c)
		return
	}

	if err := h.service.Delete(
		c.Request.Context(),
		userIDString,
		deviceID,
	); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "device removed successfully",
		},
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
	case errors.Is(err, ErrDeviceNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "DEVICE_NOT_FOUND",
				"message": "device not found",
			},
		})

	case errors.Is(err, ErrInvalidDevice):
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_DEVICE",
				"message": "invalid device",
			},
		})

	case errors.Is(err, ErrUnauthorizedDevice):
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "UNAUTHORIZED_DEVICE",
				"message": "device does not belong to this user",
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
