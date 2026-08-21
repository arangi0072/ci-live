package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"user-service/internal/config"
)

type Claims struct {
	UserID string `json:"sub"`
	Email  string `json:"email"`
	Type   string `json:"type"`

	jwt.RegisteredClaims
}

func Auth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// --------------------------------------------------
		// 1. Read Authorization header
		// --------------------------------------------------

		header := c.GetHeader("Authorization")

		if header == "" {
			unauthorized(
				c,
				"authorization header is required",
			)
			return
		}

		// --------------------------------------------------
		// 2. Parse Bearer token
		// --------------------------------------------------

		parts := strings.Fields(header)

		if len(parts) != 2 ||
			!strings.EqualFold(parts[0], "Bearer") {

			unauthorized(
				c,
				"invalid authorization header",
			)
			return
		}

		tokenString := parts[1]

		if tokenString == "" {
			unauthorized(
				c,
				"access token is required",
			)
			return
		}

		// --------------------------------------------------
		// 3. Parse and validate JWT
		// --------------------------------------------------

		claims := &Claims{}

		token, err := jwt.ParseWithClaims(
			tokenString,
			claims,
			func(token *jwt.Token) (interface{}, error) {

				// Only allow HMAC signing methods.
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New(
						"unexpected signing method",
					)
				}

				return []byte(cfg.JWTSecret), nil
			},
			jwt.WithIssuer(cfg.JWTIssuer),
		)

		if err != nil {
			unauthorized(
				c,
				"invalid access token",
			)
			return
		}

		if !token.Valid {
			unauthorized(
				c,
				"invalid access token",
			)
			return
		}

		// --------------------------------------------------
		// 4. Validate token type
		// --------------------------------------------------

		if claims.Type != "access" {
			unauthorized(
				c,
				"invalid token type",
			)
			return
		}

		// --------------------------------------------------
		// 5. Validate user ID
		// --------------------------------------------------

		if claims.UserID == "" {
			unauthorized(
				c,
				"invalid token claims",
			)
			return
		}

		// --------------------------------------------------
		// 6. Store authenticated user information
		// --------------------------------------------------

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)

		// --------------------------------------------------
		// 7. Continue request
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
			"success": false,
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": message,
			},
		},
	)
}
