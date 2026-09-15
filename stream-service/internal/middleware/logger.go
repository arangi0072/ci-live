package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

//
// Logger Middleware
//

func Logger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {

		start := time.Now()

		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery

		c.Next()

		duration := time.Since(start)

		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		//
		// Build request path.
		//

		if rawQuery != "" {
			path = path + "?" + rawQuery
		}

		//
		// Log request.
		//

		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.String("client_ip", clientIP),
			zap.Duration("latency", duration),
			zap.String("user_agent", c.Request.UserAgent()),
		}

		//
		// Add authenticated user if available.
		//

		if userID, exists := c.Get(UserIDContextKey); exists {
			switch id := userID.(type) {
			case string:
				fields = append(
					fields,
					zap.String("user_id", id),
				)

			case interface {
				String() string
			}:
				fields = append(
					fields,
					zap.String("user_id", id.String()),
				)
			}
		}

		//
		// Error information.
		//

		if len(c.Errors) > 0 {
			fields = append(
				fields,
				zap.String("errors", c.Errors.String()),
			)
		}

		//
		// Choose log level based on status.
		//

		switch {
		case statusCode >= 500:
			logger.Error(
				"http request",
				fields...,
			)

		case statusCode >= 400:
			logger.Warn(
				"http request",
				fields...,
			)

		default:
			logger.Info(
				"http request",
				fields...,
			)
		}
	}
}
