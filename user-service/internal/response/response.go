package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ============================================================
// API Response
// ============================================================

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

// ============================================================
// API Error
// ============================================================

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ============================================================
// Success
// ============================================================

func Success(
	c *gin.Context,
	status int,
	data interface{},
) {
	c.JSON(status, APIResponse{
		Success: true,
		Data:    data,
	})
}

// ============================================================
// 200 OK
// ============================================================

func OK(
	c *gin.Context,
	data interface{},
) {
	Success(
		c,
		http.StatusOK,
		data,
	)
}

// ============================================================
// 201 Created
// ============================================================

func Created(
	c *gin.Context,
	data interface{},
) {
	Success(
		c,
		http.StatusCreated,
		data,
	)
}

// ============================================================
// Error
// ============================================================

func Error(
	c *gin.Context,
	status int,
	code string,
	message string,
) {
	c.AbortWithStatusJSON(
		status,
		APIResponse{
			Success: false,
			Error: &APIError{
				Code:    code,
				Message: message,
			},
		},
	)
}

// ============================================================
// 400 Bad Request
// ============================================================

func BadRequest(
	c *gin.Context,
	code string,
	message string,
) {
	Error(
		c,
		http.StatusBadRequest,
		code,
		message,
	)
}

// ============================================================
// 401 Unauthorized
// ============================================================

func Unauthorized(
	c *gin.Context,
	code string,
	message string,
) {
	Error(
		c,
		http.StatusUnauthorized,
		code,
		message,
	)
}

// ============================================================
// 403 Forbidden
// ============================================================

func Forbidden(
	c *gin.Context,
	code string,
	message string,
) {
	Error(
		c,
		http.StatusForbidden,
		code,
		message,
	)
}

// ============================================================
// 404 Not Found
// ============================================================

func NotFound(
	c *gin.Context,
	code string,
	message string,
) {
	Error(
		c,
		http.StatusNotFound,
		code,
		message,
	)
}

// ============================================================
// 409 Conflict
// ============================================================

func Conflict(
	c *gin.Context,
	code string,
	message string,
) {
	Error(
		c,
		http.StatusConflict,
		code,
		message,
	)
}

// ============================================================
// 500 Internal Server Error
// ============================================================

func InternalServerError(
	c *gin.Context,
) {
	Error(
		c,
		http.StatusInternalServerError,
		"INTERNAL_SERVER_ERROR",
		"internal server error",
	)
}
