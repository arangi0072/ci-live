package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

//
// Recovery Middleware
//

func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {

		defer func() {

			if recovered := recover(); recovered != nil {

				//
				// Capture stack trace.
				//

				stack := debug.Stack()

				//
				// Log panic.
				//

				logger.Error(
					"panic recovered",
					zap.Any(
						"panic",
						recovered,
					),
					zap.ByteString(
						"stack",
						stack,
					),
					zap.String(
						"method",
						c.Request.Method,
					),
					zap.String(
						"path",
						c.Request.URL.Path,
					),
					zap.String(
						"client_ip",
						c.ClientIP(),
					),
				)

				//
				// Abort request with JSON.
				//

				c.AbortWithStatusJSON(
					http.StatusInternalServerError,
					gin.H{
						"success": false,
						"error": gin.H{
							"message": "internal server error",
						},
					},
				)

				return
			}
		}()

		c.Next()
	}
}
