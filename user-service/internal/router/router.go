package router

import (
	"net/http"

	"user-service/internal/config"
	"user-service/internal/device"
	"user-service/internal/middleware"
	"user-service/internal/response"
	"user-service/internal/user"

	"github.com/gin-gonic/gin"
)

func New(
	cfg *config.Config,
	userHandler *user.Handler,
	deviceHandler *device.Handler,
) *gin.Engine {

	// --------------------------------------------------
	// Gin mode
	// --------------------------------------------------

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Use gin.New() because we're providing our own
	// Recovery and Logger middleware.
	r := gin.New()

	// --------------------------------------------------
	// Global middleware
	// --------------------------------------------------

	r.Use(
		middleware.Recovery(),
		middleware.Logger(),
	)

	// --------------------------------------------------
	// Health check
	// --------------------------------------------------

	r.GET("/health", func(c *gin.Context) {
		response.OK(c, gin.H{
			"service": "user-service",
			"status":  "healthy",
		})
	})

	// --------------------------------------------------
	// API v1
	// --------------------------------------------------

	api := r.Group("/api/v1")

	// ==================================================
	// Public User Routes
	// ==================================================

	users := api.Group("/users")

	// Public profile.
	users.GET(
		"/:user_id",
		userHandler.GetByID,
	)

	// Username availability.
	users.GET(
		"/username/:username/availability",
		userHandler.CheckUsername,
	)

	// ==================================================
	// Protected Routes
	// ==================================================

	protected := api.Group("")

	protected.Use(
		middleware.Auth(cfg),
	)

	// --------------------------------------------------
	// User Profile
	// --------------------------------------------------

	protected.GET(
		"/users/me",
		userHandler.GetMe,
	)

	protected.POST(
		"/users/me",
		userHandler.CreateMe,
	)

	protected.PATCH(
		"/users/me",
		userHandler.UpdateMe,
	)

	// --------------------------------------------------
	// Devices
	// --------------------------------------------------

	protected.GET(
		"/devices",
		deviceHandler.List,
	)

	protected.POST(
		"/devices",
		deviceHandler.Register,
	)

	protected.PATCH(
		"/devices/:device_id",
		deviceHandler.Update,
	)

	protected.DELETE(
		"/devices/:device_id",
		deviceHandler.Delete,
	)

	// --------------------------------------------------
	// 404
	// --------------------------------------------------

	r.NoRoute(func(c *gin.Context) {
		response.Error(
			c,
			http.StatusNotFound,
			"ROUTE_NOT_FOUND",
			"route not found",
		)
	})

	return r
}
