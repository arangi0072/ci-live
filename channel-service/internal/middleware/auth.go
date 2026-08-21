package middleware

import (
	"net/http"
	"strings"

	"channel-service/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"user_id"`

	jwt.RegisteredClaims
}

func Auth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {

		// --------------------------------------------------
		// Get Authorization header
		// --------------------------------------------------

		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			unauthorized(
				c,
				"AUTHENTICATION_REQUIRED",
				"authorization token is required",
			)
			return
		}

		// --------------------------------------------------
		// Expected format:
		//
		// Authorization: Bearer <token>
		// --------------------------------------------------

		parts := strings.SplitN(
			authHeader,
			" ",
			2,
		)

		if len(parts) != 2 ||
			!strings.EqualFold(parts[0], "Bearer") {

			unauthorized(
				c,
				"INVALID_AUTHORIZATION",
				"invalid authorization header",
			)
			return
		}

		tokenString := strings.TrimSpace(parts[1])

		if tokenString == "" {
			unauthorized(
				c,
				"INVALID_TOKEN",
				"token is empty",
			)
			return
		}

		// --------------------------------------------------
		// Parse JWT
		// --------------------------------------------------

		claims := &Claims{}

		token, err := jwt.ParseWithClaims(
			tokenString,
			claims,
			func(token *jwt.Token) (interface{}, error) {

				// Only accept HMAC signing methods.
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrTokenSignatureInvalid
				}

				return []byte(cfg.JWTSecret), nil
			},

			jwt.WithIssuer(cfg.JWTIssuer),

			jwt.WithValidMethods(
				[]string{
					jwt.SigningMethodHS256.Alg(),
				},
			),
		)

		if err != nil {
			unauthorized(
				c,
				"INVALID_TOKEN",
				"invalid or expired token",
			)
			return
		}

		if !token.Valid {
			unauthorized(
				c,
				"INVALID_TOKEN",
				"invalid or expired token",
			)
			return
		}

		// --------------------------------------------------
		// Validate User ID
		// --------------------------------------------------

		if claims.UserID == "" {
			unauthorized(
				c,
				"INVALID_TOKEN",
				"user identity missing from token",
			)
			return
		}

		// --------------------------------------------------
		// Store authenticated user
		// --------------------------------------------------

		c.Set(
			"user_id",
			claims.UserID,
		)

		c.Set(
			"claims",
			claims,
		)

		// Continue request.
		c.Next()
	}
}

// ============================================================
// Unauthorized Response
// ============================================================

func unauthorized(
	c *gin.Context,
	code string,
	message string,
) {
	c.AbortWithStatusJSON(
		http.StatusUnauthorized,
		gin.H{
			"success": false,
			"error": gin.H{
				"code":    code,
				"message": message,
			},
		},
	)
}
