package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {

		defer func() {
			if err := recover(); err != nil {

				// Log panic internally with stack trace.
				log.Printf(
					"PANIC recovered: %v\n%s",
					err,
					debug.Stack(),
				)

				// Stop the request.
				c.Abort()

				// Don't expose internal panic details.
				if !c.Writer.Written() {
					c.JSON(
						http.StatusInternalServerError,
						gin.H{
							"success": false,
							"error": gin.H{
								"code":    "INTERNAL_SERVER_ERROR",
								"message": "internal server error",
							},
						},
					)
				}
			}
		}()

		c.Next()
	}
}
