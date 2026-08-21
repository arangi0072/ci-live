package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

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