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

		duration := time.Since(start)

		status := c.Writer.Status()

		clientIP := c.ClientIP()
		method := c.Request.Method
		path := c.Request.URL.Path

		log.Printf(
			"HTTP %s %s status=%d duration=%s ip=%s",
			method,
			path,
			status,
			duration,
			clientIP,
		)
	}
}