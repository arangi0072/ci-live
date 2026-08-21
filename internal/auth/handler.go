package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthService interface {
	Signup(ctx context.Context, req SignupRequest) (*AuthResponse, error)
	Login(ctx context.Context, req LoginRequest) (*AuthResponse, error)
	Refresh(ctx context.Context, req RefreshTokenRequest) (*AuthResponse, error)
	Logout(ctx context.Context, req LogoutRequest) error
	GetUser(ctx context.Context, userID string) (*UserResponse, error)
	VerifyEmail(ctx context.Context, req VerifyEmailRequest) error
	ForgotPassword(ctx context.Context, req ForgotPasswordRequest) error
	ResetPassword(ctx context.Context, req ResetPasswordRequest) error
}

type AuthHandler struct {
	service AuthService
}

func NewHandler(service AuthService) *AuthHandler {
	return &AuthHandler{
		service: service,
	}
}

// ============================================================
// POST /api/v1/auth/signup
// ============================================================

func (h *AuthHandler) Signup(c *gin.Context) {
	var req SignupRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
		})
		return
	}

	response, err := h.service.Signup(c.Request.Context(), req)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response)
}

// ============================================================
// POST /api/v1/auth/login
// ============================================================

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
		})
		return
	}

	response, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// ============================================================
// POST /api/v1/auth/refresh
// ============================================================

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshTokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
		})
		return
	}

	response, err := h.service.Refresh(
		c.Request.Context(),
		req,
	)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// ============================================================
// POST /api/v1/auth/logout
// ============================================================

func (h *AuthHandler) Logout(c *gin.Context) {
	var req LogoutRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
		})
		return
	}

	if err := h.service.Logout(
		c.Request.Context(),
		req,
	); err != nil {
		handleAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, MessageResponse{
		Message: "logged out successfully",
	})
}

// ============================================================
// GET /api/v1/auth/me
// ============================================================

func (h *AuthHandler) Me(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	userID, ok := userIDValue.(string)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	user, err := h.service.GetUser(
		c.Request.Context(),
		userID,
	)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

// ============================================================
// POST /api/v1/auth/verify-email
// ============================================================

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req VerifyEmailRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
		})
		return
	}

	if err := h.service.VerifyEmail(
		c.Request.Context(),
		req,
	); err != nil {
		handleAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, MessageResponse{
		Message: "email verified successfully",
	})
}

// ============================================================
// POST /api/v1/auth/forgot-password
// ============================================================

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
		})
		return
	}

	// The service should deliberately return the same successful
	// response whether or not the email exists.
	if err := h.service.ForgotPassword(
		c.Request.Context(),
		req,
	); err != nil {
		handleAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, MessageResponse{
		Message: "if the email exists, a password reset link has been sent",
	})
}

// ============================================================
// POST /api/v1/auth/reset-password
// ============================================================

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
		})
		return
	}

	if err := h.service.ResetPassword(
		c.Request.Context(),
		req,
	); err != nil {
		handleAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, MessageResponse{
		Message: "password reset successfully",
	})
}

// ============================================================
// Error handling
// ============================================================

func handleAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid email or password",
		})

	case errors.Is(err, ErrInvalidToken):
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid token",
		})

	case errors.Is(err, ErrExpiredToken):
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "token expired",
		})

	case errors.Is(err, ErrUserNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"error": "user not found",
		})

	case errors.Is(err, ErrEmailAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{
			"error": "email already registered",
		})

	case errors.Is(err, ErrEmailNotVerified):
		c.JSON(http.StatusForbidden, gin.H{
			"error": "email is not verified",
		})

	case errors.Is(err, ErrUserInactive):
		c.JSON(http.StatusForbidden, gin.H{
			"error": "account is inactive",
		})

	case errors.Is(err, ErrUserBanned):
		c.JSON(http.StatusForbidden, gin.H{
			"error": "account is banned",
		})

	default:
		c.Error(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
	}
}