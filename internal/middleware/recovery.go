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
				// Log the panic and stack trace internally.
				log.Printf(
					"PANIC recovered: %v\n%s",
					err,
					debug.Stack(),
				)

				// If the response hasn't already been written,
				// return a generic error to the client.
				if !c.Writer.Written() {
					c.AbortWithStatusJSON(
						http.StatusInternalServerError,
						gin.H{
							"error": "internal server error",
						},
					)

					return
				}

				// Response was already started.
				c.Abort()
			}
		}()

		c.Next()
	}
}