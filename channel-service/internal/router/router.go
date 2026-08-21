package router

import (
	"net/http"

	"channel-service/internal/category"
	"channel-service/internal/channel"
	"channel-service/internal/config"
	"channel-service/internal/follow"
	"channel-service/internal/middleware"
	"channel-service/internal/response"
	"channel-service/internal/topic"

	"github.com/gin-gonic/gin"
)

func New(
	cfg *config.Config,
	channelHandler *channel.Handler,
	categoryHandler *category.Handler,
	topicHandler *topic.Handler,
	followHandler *follow.Handler,
) *gin.Engine {

	// --------------------------------------------------
	// Gin mode
	// --------------------------------------------------

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Use gin.New() because we're using our own
	// Logger and Recovery middleware.
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
			"service": "channel-service",
			"status":  "healthy",
		})
	})

	// --------------------------------------------------
	// API v1
	// --------------------------------------------------

	api := r.Group("/api/v1")

	// ==================================================
	// Public Channel Routes
	// ==================================================

	channels := api.Group("/channels")

	channels.GET(
		"/:channel_id",
		channelHandler.GetByID,
	)

	channels.GET(
		"/username/:username",
		channelHandler.GetByUsername,
	)

	// ==================================================
	// Public Category Routes
	// ==================================================

	categories := api.Group("/categories")

	categories.GET(
		"/",
		categoryHandler.List,
	)

	categories.GET(
		"/:category_id",
		categoryHandler.GetByID,
	)

	categories.GET(
		"/slug/:slug",
		categoryHandler.GetBySlug,
	)

	// ==================================================
	// Public Topic Routes
	// ==================================================

	topics := api.Group("/topics")

	topics.GET(
		"/",
		topicHandler.List,
	)

	topics.GET(
		"/:topic_id",
		topicHandler.GetByID,
	)

	topics.GET(
		"/slug/:slug",
		topicHandler.GetBySlug,
	)

	// ==================================================
	// Protected Routes
	// ==================================================

	protected := api.Group("")

	protected.Use(
		middleware.Auth(cfg),
	)

	// ==================================================
	// My Channel
	// ==================================================

	protected.POST(
		"/channels",
		channelHandler.Create,
	)

	protected.GET(
		"/channels/me",
		channelHandler.GetMe,
	)

	protected.PATCH(
		"/channels/me",
		channelHandler.Update,
	)

	// ==================================================
	// Follow / Unfollow
	// ==================================================

	protected.POST(
		"/channels/:channel_id/follow",
		followHandler.Follow,
	)

	protected.DELETE(
		"/channels/:channel_id/follow",
		followHandler.Unfollow,
	)

	protected.GET(
		"/channels/:channel_id/follow",
		followHandler.IsFollowing,
	)

	// Follower count can technically be public,
	// but keeping it protected for now is fine.
	protected.GET(
		"/channels/:channel_id/followers/count",
		followHandler.GetFollowerCount,
	)

	protected.GET(
		"/users/me/following",
		followHandler.GetFollowing,
	)

	// ==================================================
	// Category Admin Routes
	// ==================================================

	// TODO:
	// Add admin middleware before exposing these.
	//
	// protected.POST("/categories", categoryHandler.Create)
	// protected.PATCH("/categories/:category_id", categoryHandler.Update)
	// protected.DELETE("/categories/:category_id", categoryHandler.Delete)

	// ==================================================
	// Topic Admin Routes
	// ==================================================

	// TODO:
	// Add admin middleware before exposing these.
	//
	// protected.POST("/topics", topicHandler.Create)
	// protected.PATCH("/topics/:topic_id", topicHandler.Update)
	// protected.DELETE("/topics/:topic_id", topicHandler.Delete)

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
