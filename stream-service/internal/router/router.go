package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"stream-service/internal/chat"
	"stream-service/internal/endpoint"
	"stream-service/internal/like"
	"stream-service/internal/middleware"
	"stream-service/internal/recording"
	"stream-service/internal/session"
	"stream-service/internal/stream"
)

//
// Router Dependencies
//

type Dependencies struct {
	Logger *zap.Logger

	ChatHandler      *chat.Handler
	EndpointHandler  *endpoint.Handler
	LikeHandler      *like.Handler
	RecordingHandler *recording.Handler
	SessionHandler   *session.Handler
	StreamHandler    *stream.Handler
}

//
// New Router
//

func NewRouter(
	deps Dependencies,
) *gin.Engine {

	//
	// Gin engine
	//

	router := gin.New()

	//
	// Global middleware
	//

	router.Use(
		middleware.Logger(deps.Logger),
		middleware.Recovery(deps.Logger),
	)

	//
	// Health check
	//

	router.GET(
		"/health",
		healthCheck,
	)

	router.GET(
		"/ready",
		healthCheck,
	)

	//
	// API v1
	//

	api := router.Group("/api/v1")

	registerRoutes(
		api,
		deps,
	)

	return router
}

//
// Register Routes
//

func registerRoutes(
	api *gin.RouterGroup,
	deps Dependencies,
) {

	//
	// =========================
	// PUBLIC ROUTES
	// =========================
	//

	public := api.Group("")

	//
	// Streams
	//

	if deps.StreamHandler != nil {

		public.GET(
			"/streams",
			deps.StreamHandler.ListStreams,
		)

		public.GET(
			"/streams/:id",
			deps.StreamHandler.GetStream,
		)
	}

	//
	// Recordings
	//

	if deps.RecordingHandler != nil {

		public.GET(
			"/recordings/:id",
			deps.RecordingHandler.GetRecording,
		)

		public.GET(
			"/streams/:stream_id/recordings",
			deps.RecordingHandler.GetStreamRecordings,
		)

		public.GET(
			"/sessions/:session_id/recordings",
			deps.RecordingHandler.GetSessionRecordings,
		)
	}

	//
	// Chat history
	//

	if deps.ChatHandler != nil {

		public.GET(
			"/streams/:stream_id/chat/messages",
			deps.ChatHandler.ListMessages,
		)
	}

	//
	// =========================
	// AUTHENTICATED ROUTES
	// =========================
	//

	protected := api.Group("")

	protected.Use(
		middleware.Auth(),
	)

	//
	// Streams
	//

	if deps.StreamHandler != nil {

		protected.POST(
			"/streams",
			deps.StreamHandler.CreateStream,
		)

		protected.PATCH(
			"/streams/:id",
			deps.StreamHandler.UpdateStream,
		)

		protected.DELETE(
			"/streams/:id",
			deps.StreamHandler.DeleteStream,
		)

		protected.PATCH(
			"/streams/:id/status",
			deps.StreamHandler.UpdateStreamStatus,
		)
	}

	//
	// Stream Sessions
	//

	if deps.SessionHandler != nil {

		protected.POST(
			"/streams/:stream_id/sessions",
			deps.SessionHandler.CreateSession,
		)

		protected.GET(
			"/sessions/:id",
			deps.SessionHandler.GetSession,
		)

		protected.GET(
			"/streams/:stream_id/sessions",
			deps.SessionHandler.ListSessions,
		)

		protected.PATCH(
			"/sessions/:id",
			deps.SessionHandler.UpdateSession,
		)

		protected.DELETE(
			"/sessions/:id",
			deps.SessionHandler.DeleteSession,
		)
	}

	//
	// Stream Endpoints
	//

	if deps.EndpointHandler != nil {

		protected.POST(
			"/streams/:stream_id/endpoints",
			deps.EndpointHandler.CreateEndpoint,
		)

		protected.GET(
			"/endpoints/:id",
			deps.EndpointHandler.GetEndpoint,
		)

		protected.GET(
			"/streams/:stream_id/endpoints",
			deps.EndpointHandler.ListEndpoints,
		)

		protected.PATCH(
			"/endpoints/:id",
			deps.EndpointHandler.UpdateEndpoint,
		)

		protected.DELETE(
			"/endpoints/:id",
			deps.EndpointHandler.DeleteEndpoint,
		)
	}

	//
	// Recordings
	//

	if deps.RecordingHandler != nil {

		protected.POST(
			"/recordings",
			deps.RecordingHandler.CreateRecording,
		)

		protected.PATCH(
			"/recordings/:id",
			deps.RecordingHandler.UpdateRecording,
		)

		protected.PATCH(
			"/recordings/:id/status",
			deps.RecordingHandler.UpdateRecordingStatus,
		)

		protected.DELETE(
			"/recordings/:id",
			deps.RecordingHandler.DeleteRecording,
		)
	}

	//
	// Chat
	//

	if deps.ChatHandler != nil {

		protected.POST(
			"/streams/:stream_id/chat/messages",
			deps.ChatHandler.CreateMessage,
		)

		protected.GET(
			"/chat/messages/:id",
			deps.ChatHandler.GetMessage,
		)

		protected.DELETE(
			"/chat/messages/:id",
			deps.ChatHandler.DeleteMessage,
		)

		protected.GET(
			"/streams/:stream_id/chat/stats",
			deps.ChatHandler.GetMessageStats,
		)
	}

	//
	// Likes
	//

	if deps.LikeHandler != nil {

		protected.POST(
			"/streams/:stream_id/like",
			deps.LikeHandler.CreateLike,
		)

		protected.DELETE(
			"/streams/:stream_id/like",
			deps.LikeHandler.RemoveLike,
		)

		protected.GET(
			"/streams/:stream_id/like",
			deps.LikeHandler.GetLike,
		)

		protected.GET(
			"/streams/:stream_id/like/check",
			deps.LikeHandler.CheckLike,
		)

		protected.GET(
			"/streams/:stream_id/likes",
			deps.LikeHandler.ListLikes,
		)

		protected.GET(
			"/streams/:stream_id/likes/stats",
			deps.LikeHandler.GetLikeStats,
		)
	}

}

//
// Health Check
//

func healthCheck(
	c *gin.Context,
) {
	c.JSON(
		http.StatusOK,
		gin.H{
			"success": true,
			"data": gin.H{
				"status": "ok",
			},
		},
	)
}
