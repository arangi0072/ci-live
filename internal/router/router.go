package router

import (
	"net/http"

	"ci-live/internal/auth"
	"ci-live/internal/config"
	"ci-live/internal/middleware"
	"ci-live/internal/response"

	"github.com/gin-gonic/gin"
)

func New(
	cfg *config.Config,
	authHandler *auth.AuthHandler,
	jwtManager *auth.JWTManager,
) *gin.Engine {

	// --------------------------------------------------
	// Gin
	// --------------------------------------------------

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

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
			"service": "auth-service",
			"status":  "healthy",
		})
	})

	// --------------------------------------------------
	// API v1
	// --------------------------------------------------

	api := r.Group("/api/v1")

	// --------------------------------------------------
	// Authentication routes
	// --------------------------------------------------

	authRoutes := api.Group("/auth")

	// Public routes
	authRoutes.POST(
		"/signup",
		authHandler.Signup,
	)

	authRoutes.POST(
		"/login",
		authHandler.Login,
	)

	authRoutes.POST(
		"/refresh",
		authHandler.Refresh,
	)

	authRoutes.POST(
		"/verify-email",
		authHandler.VerifyEmail,
	)

	authRoutes.POST(
		"/forgot-password",
		authHandler.ForgotPassword,
	)

	authRoutes.POST(
		"/reset-password",
		authHandler.ResetPassword,
	)

	// --------------------------------------------------
	// Protected routes
	// --------------------------------------------------

	protected := authRoutes.Group("")

	protected.Use(
		middleware.Auth(jwtManager),
	)

	protected.GET(
		"/me",
		authHandler.Me,
	)

	protected.POST(
		"/logout",
		authHandler.Logout,
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