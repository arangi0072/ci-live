package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request.
		c.Next()

		// Calculate request duration.
		duration := time.Since(start)

		// Default for unauthenticated requests.
		userID := "-"

		if value, exists := c.Get("user_id"); exists {
			if id, ok := value.(string); ok && id != "" {
				userID = id
			}
		}

		log.Printf(
			"HTTP method=%s path=%s status=%d duration=%s ip=%s user_id=%s",
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration,
			c.ClientIP(),
			userID,
		)
	}
}
