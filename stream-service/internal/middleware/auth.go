package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	UserIDContextKey = "user_id"
)

//
// Auth Middleware
//
// Expects:
//
// Authorization: Bearer <token>
//
// The JWT validation itself should be handled by your
// JWT/auth service. This middleware expects a validated
// user ID to be available from the token claims.
//

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {

		authorization := c.GetHeader("Authorization")

		if authorization == "" {
			unauthorized(c, "authorization header is required")
			return
		}

		parts := strings.Fields(authorization)

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

		//
		// TODO:
		// Validate JWT using your Auth/JWT service.
		//
		// The validated token must provide the user ID.
		//

		userID, err := extractUserIDFromToken(token)
		if err != nil {
			unauthorized(c, "invalid or expired access token")
			return
		}

		if userID == uuid.Nil {
			unauthorized(c, "invalid user ID")
			return
		}

		//
		// Make user ID available to handlers/services.
		//

		c.Set(UserIDContextKey, userID)

		c.Next()
	}
}

//
// Optional Auth Middleware
//
// Allows both authenticated and anonymous requests.
//

func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {

		authorization := c.GetHeader("Authorization")

		if authorization == "" {
			c.Next()
			return
		}

		parts := strings.Fields(authorization)

		if len(parts) != 2 ||
			!strings.EqualFold(parts[0], "Bearer") {

			c.Next()
			return
		}

		token := parts[1]

		userID, err := extractUserIDFromToken(token)

		if err == nil && userID != uuid.Nil {
			c.Set(UserIDContextKey, userID)
		}

		c.Next()
	}
}

//
// Extract User ID
//
// Replace the implementation with your actual JWT
// verification service when auth package is connected.
//

func extractUserIDFromToken(token string) (uuid.UUID, error) {

	//
	// IMPORTANT:
	// Do not trust a raw UUID/token in production.
	//
	// This function is intentionally isolated so your
	// RS256 JWT validation can be plugged in here.
	//

	return uuid.Nil, ErrTokenValidationNotConfigured
}

//
// Errors
//

var (
	ErrTokenValidationNotConfigured = &AuthError{
		Message: "token validation is not configured",
	}
)

type AuthError struct {
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}

//
// Unauthorized Response
//

func unauthorized(
	c *gin.Context,
	message string,
) {
	c.AbortWithStatusJSON(
		http.StatusUnauthorized,
		gin.H{
			"success": false,
			"error": gin.H{
				"message": message,
			},
		},
	)
}
