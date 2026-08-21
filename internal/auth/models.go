package auth

import "time"

// ============================================================
// User
// ============================================================

type User struct {
	ID            string     `json:"id"`
	Email         string     `json:"email"`
	PasswordHash  string     `json:"-"`
	EmailVerified bool       `json:"email_verified"`
	IsActive      bool       `json:"is_active"`
	IsBanned      bool       `json:"is_banned"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// ============================================================
// Signup
// ============================================================

type SignupRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=128"`
}

// ============================================================
// Login
// ============================================================

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// ============================================================
// Refresh Token
// ============================================================

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ============================================================
// Logout
// ============================================================

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ============================================================
// Email Verification
// ============================================================

type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

// ============================================================
// Forgot Password
// ============================================================

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ============================================================
// Reset Password
// ============================================================

type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=8,max=128"`
}

// ============================================================
// Refresh Token Database Model
// ============================================================

type RefreshToken struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	TokenHash string     `json:"-"`
	DeviceID  *string    `json:"device_id,omitempty"`
	ExpiresAt time.Time  `json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

// ============================================================
// Email Verification Token
// ============================================================

type EmailVerificationToken struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// ============================================================
// Password Reset Token
// ============================================================

type PasswordResetToken struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// ============================================================
// Authentication Response
// ============================================================

type AuthResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"`
	ExpiresIn    int64        `json:"expires_in"`
}

// ============================================================
// Public User Response
// ============================================================

// Never expose password_hash or other sensitive internal fields.
type UserResponse struct {
	ID            string     `json:"id"`
	Email         string     `json:"email"`
	EmailVerified bool       `json:"email_verified"`
	CreatedAt     time.Time  `json:"created_at"`
}

// ============================================================
// Generic Message Response
// ============================================================

type MessageResponse struct {
	Message string `json:"message"`
}

// ============================================================
// Device Information
// ============================================================

type DeviceInfo struct {
	DeviceID    string `json:"device_id,omitempty"`
	Platform    string `json:"platform,omitempty"`
	AppVersion  string `json:"app_version,omitempty"`
}