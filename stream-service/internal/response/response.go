package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

//
// API Response
//

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorBody  `json:"error,omitempty"`
}

//
// Error Body
//

type ErrorBody struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

//
// Success Response
//

func Success(
	c *gin.Context,
	status int,
	data interface{},
) {
	c.JSON(
		status,
		Response{
			Success: true,
			Data:    data,
		},
	)
}

//
// Created Response
//

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

//
// OK Response
//

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

//
// No Content Response
//

func NoContent(
	c *gin.Context,
) {
	c.Status(http.StatusNoContent)
}

//
// Error Response
//

func Error(
	c *gin.Context,
	status int,
	message string,
) {
	c.JSON(
		status,
		Response{
			Success: false,
			Error: &ErrorBody{
				Message: message,
			},
		},
	)
}

//
// Error Response With Code
//

func ErrorWithCode(
	c *gin.Context,
	status int,
	code string,
	message string,
) {
	c.JSON(
		status,
		Response{
			Success: false,
			Error: &ErrorBody{
				Code:    code,
				Message: message,
			},
		},
	)
}

//
// Bad Request
//

func BadRequest(
	c *gin.Context,
	message string,
) {
	Error(
		c,
		http.StatusBadRequest,
		message,
	)
}

//
// Unauthorized
//

func Unauthorized(
	c *gin.Context,
	message string,
) {
	Error(
		c,
		http.StatusUnauthorized,
		message,
	)
}

//
// Forbidden
//

func Forbidden(
	c *gin.Context,
	message string,
) {
	Error(
		c,
		http.StatusForbidden,
		message,
	)
}

//
// Not Found
//

func NotFound(
	c *gin.Context,
	message string,
) {
	Error(
		c,
		http.StatusNotFound,
		message,
	)
}

//
// Conflict
//

func Conflict(
	c *gin.Context,
	message string,
) {
	Error(
		c,
		http.StatusConflict,
		message,
	)
}

//
// Too Many Requests
//

func TooManyRequests(
	c *gin.Context,
	message string,
) {
	Error(
		c,
		http.StatusTooManyRequests,
		message,
	)
}

//
// Internal Server Error
//

func InternalServerError(
	c *gin.Context,
	message string,
) {
	Error(
		c,
		http.StatusInternalServerError,
		message,
	)
}
