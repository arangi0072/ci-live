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
				// Log the actual panic internally.
				// Never expose the panic details to the client.
				log.Printf(
					"PANIC recovered: %v\n%s",
					err,
					debug.Stack(),
				)

				// Stop processing the request.
				c.Abort()

				// Return a generic error.
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
