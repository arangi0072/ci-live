package middleware

import (
	"net/http"
	"strings"

	"ci-live/internal/auth"

	"github.com/gin-gonic/gin"
)

func Auth(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// --------------------------------------------------
		// 1. Get Authorization header
		// --------------------------------------------------

		header := c.GetHeader("Authorization")

		if header == "" {
			unauthorized(c, "authorization header is required")
			return
		}

		// --------------------------------------------------
		// 2. Validate Bearer scheme
		// --------------------------------------------------

		parts := strings.Fields(header)

		if len(parts) != 2 ||
			!strings.EqualFold(parts[0], "Bearer") {
			unauthorized(c, "invalid authorization header")
			return
		}

		token := parts[1]

		if token == "" {
			unauthorized(c, "access token is required")
			return
		}

		// --------------------------------------------------
		// 3. Validate JWT
		// --------------------------------------------------

		claims, err := jwtManager.ValidateAccessToken(token)
		if err != nil {
			switch err {
			case auth.ErrExpiredToken:
				unauthorized(c, "access token expired")

			default:
				unauthorized(c, "invalid access token")
			}

			return
		}

		// --------------------------------------------------
		// 4. Validate user ID
		// --------------------------------------------------

		if claims.UserID == "" {
			unauthorized(c, "invalid token claims")
			return
		}

		// --------------------------------------------------
		// 5. Store authenticated user information
		// --------------------------------------------------

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)

		// --------------------------------------------------
		// 6. Continue request
		// --------------------------------------------------

		c.Next()
	}
}

func unauthorized(
	c *gin.Context,
	message string,
) {
	c.AbortWithStatusJSON(
		http.StatusUnauthorized,
		gin.H{
			"error": message,
		},
	)
}