package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrEmailNotVerified   = errors.New("email not verified")
	ErrUserInactive       = errors.New("user inactive")
	ErrUserBanned         = errors.New("user banned")
	ErrInvalidResetToken  = errors.New("invalid reset token")
)

// ============================================================
// Repository interface
// ============================================================

type AuthRepository interface {
	CreateUser(
		ctx context.Context,
		user *User,
	) error

	GetUserByID(
		ctx context.Context,
		userID string,
	) (*User, error)

	GetUserByEmail(
		ctx context.Context,
		email string,
	) (*User, error)

	UpdateLastLogin(
		ctx context.Context,
		userID string,
	) error

	MarkEmailVerified(
		ctx context.Context,
		userID string,
	) error

	UpdatePassword(
		ctx context.Context,
		userID string,
		passwordHash string,
	) error

	// Refresh tokens
	CreateRefreshToken(
		ctx context.Context,
		token *RefreshToken,
	) error

	GetRefreshTokenByHash(
		ctx context.Context,
		tokenHash string,
	) (*RefreshToken, error)

	RevokeRefreshToken(
		ctx context.Context,
		tokenHash string,
	) error

	RevokeAllRefreshTokens(
		ctx context.Context,
		userID string,
	) error

	UpdateRefreshTokenLastUsed(
		ctx context.Context,
		tokenHash string,
	) error

	// Email verification
	CreateEmailVerificationToken(
		ctx context.Context,
		token *EmailVerificationToken,
	) error

	GetEmailVerificationToken(
		ctx context.Context,
		tokenHash string,
	) (*EmailVerificationToken, error)

	MarkEmailVerificationTokenUsed(
		ctx context.Context,
		tokenHash string,
	) error

	// Password reset
	CreatePasswordResetToken(
		ctx context.Context,
		token *PasswordResetToken,
	) error

	GetPasswordResetToken(
		ctx context.Context,
		tokenHash string,
	) (*PasswordResetToken, error)

	MarkPasswordResetTokenUsed(
		ctx context.Context,
		tokenHash string,
	) error
}

// ============================================================
// Service
// ============================================================

type Service struct {
	repository AuthRepository
	jwtManager *JWTManager
}

// Constructor
func NewService(
	repository AuthRepository,
	jwtManager *JWTManager,
) *Service {
	return &Service{
		repository: repository,
		jwtManager: jwtManager,
	}
}

// Make Service satisfy the AuthService interface used by handler.go.
var _ AuthService = (*Service)(nil)

// ============================================================
// Signup
// ============================================================

func (s *Service) Signup(
	ctx context.Context,
	req SignupRequest,
) (*AuthResponse, error) {

	email := normalizeEmail(req.Email)

	if email == "" {
		return nil, ErrInvalidCredentials
	}

	if len(req.Password) < 8 {
		return nil, ErrInvalidPassword
	}

	// Check whether email already exists.
	existingUser, err := s.repository.GetUserByEmail(
		ctx,
		email,
	)

	if err == nil && existingUser != nil {
		return nil, ErrEmailAlreadyExists
	}

	if !errors.Is(err, ErrUserNotFound) {
		if err != nil {
			return nil, fmt.Errorf(
				"check existing user: %w",
				err,
			)
		}
	}

	// Hash password.
	passwordHash, err := HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf(
			"hash password: %w",
			err,
		)
	}

	user := &User{
		ID:            uuid.NewString(),
		Email:         email,
		PasswordHash:  passwordHash,
		EmailVerified: false,
		IsActive:      true,
		IsBanned:      false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.repository.CreateUser(
		ctx,
		user,
	); err != nil {
		return nil, fmt.Errorf(
			"create user: %w",
			err,
		)
	}

	// Generate email verification token.
	rawToken, tokenHash, err := generateSecureToken()
	if err != nil {
		return nil, fmt.Errorf(
			"generate verification token: %w",
			err,
		)
	}

	verificationToken := &EmailVerificationToken{
		ID:        uuid.NewString(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	if err := s.repository.CreateEmailVerificationToken(
		ctx,
		verificationToken,
	); err != nil {
		return nil, fmt.Errorf(
			"create verification token: %w",
			err,
		)
	}

	// TODO:
	// Send rawToken through email service.
	//
	// Never store rawToken in PostgreSQL.
	_ = rawToken

	return s.createAuthResponse(
		ctx,
		user,
	)
}

// ============================================================
// Login
// ============================================================

func (s *Service) Login(
	ctx context.Context,
	req LoginRequest,
) (*AuthResponse, error) {

	email := normalizeEmail(req.Email)

	user, err := s.repository.GetUserByEmail(
		ctx,
		email,
	)

	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}

		return nil, fmt.Errorf(
			"get user by email: %w",
			err,
		)
	}

	// Do not reveal whether the account exists.
	if err := ComparePassword(
		user.PasswordHash,
		req.Password,
	); err != nil {
		return nil, ErrInvalidCredentials
	}

	if user.IsBanned {
		return nil, ErrUserBanned
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// For the prototype, email verification is required.
	if !user.EmailVerified {
		return nil, ErrEmailNotVerified
	}

	if err := s.repository.UpdateLastLogin(
		ctx,
		user.ID,
	); err != nil {
		return nil, fmt.Errorf(
			"update last login: %w",
			err,
		)
	}

	return s.createAuthResponse(
		ctx,
		user,
	)
}

// ============================================================
// Refresh
// ============================================================

func (s *Service) Refresh(
	ctx context.Context,
	req RefreshTokenRequest,
) (*AuthResponse, error) {

	claims, err := s.jwtManager.ValidateRefreshToken(
		req.RefreshToken,
	)

	if err != nil {
		return nil, err
	}

	tokenHash := hashToken(req.RefreshToken)

	storedToken, err := s.repository.GetRefreshTokenByHash(
		ctx,
		tokenHash,
	)

	if err != nil {
		return nil, ErrInvalidToken
	}

	if storedToken.RevokedAt != nil {
		return nil, ErrInvalidToken
	}

	if time.Now().After(storedToken.ExpiresAt) {
		return nil, ErrExpiredToken
	}

	// Ensure JWT subject matches database record.
	if storedToken.UserID != claims.UserID {
		return nil, ErrInvalidToken
	}

	user, err := s.repository.GetUserByID(
		ctx,
		claims.UserID,
	)

	if err != nil {
		return nil, ErrUserNotFound
	}

	if user.IsBanned {
		return nil, ErrUserBanned
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	if err := s.repository.UpdateRefreshTokenLastUsed(
		ctx,
		tokenHash,
	); err != nil {
		return nil, fmt.Errorf(
			"update refresh token usage: %w",
			err,
		)
	}

	// Token rotation:
	// Revoke the old refresh token and issue a new pair.
	if err := s.repository.RevokeRefreshToken(
		ctx,
		tokenHash,
	); err != nil {
		return nil, fmt.Errorf(
			"revoke old refresh token: %w",
			err,
		)
	}

	return s.createAuthResponse(
		ctx,
		user,
	)
}

// ============================================================
// Logout
// ============================================================

func (s *Service) Logout(
	ctx context.Context,
	req LogoutRequest,
) error {

	tokenHash := hashToken(req.RefreshToken)

	if err := s.repository.RevokeRefreshToken(
		ctx,
		tokenHash,
	); err != nil {
		return fmt.Errorf(
			"revoke refresh token: %w",
			err,
		)
	}

	return nil
}

// ============================================================
// Get User
// ============================================================

func (s *Service) GetUser(
	ctx context.Context,
	userID string,
) (*UserResponse, error) {

	user, err := s.repository.GetUserByID(
		ctx,
		userID,
	)

	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf(
			"get user: %w",
			err,
		)
	}

	return toUserResponse(user), nil
}

// ============================================================
// Verify Email
// ============================================================

func (s *Service) VerifyEmail(
	ctx context.Context,
	req VerifyEmailRequest,
) error {

	tokenHash := hashToken(req.Token)

	token, err := s.repository.GetEmailVerificationToken(
		ctx,
		tokenHash,
	)

	if err != nil {
		return ErrInvalidToken
	}

	if token.UsedAt != nil {
		return ErrInvalidToken
	}

	if time.Now().After(token.ExpiresAt) {
		return ErrExpiredToken
	}

	if err := s.repository.MarkEmailVerified(
		ctx,
		token.UserID,
	); err != nil {
		return fmt.Errorf(
			"mark email verified: %w",
			err,
		)
	}

	if err := s.repository.MarkEmailVerificationTokenUsed(
		ctx,
		tokenHash,
	); err != nil {
		return fmt.Errorf(
			"mark verification token used: %w",
			err,
		)
	}

	return nil
}

// ============================================================
// Forgot Password
// ============================================================

func (s *Service) ForgotPassword(
	ctx context.Context,
	req ForgotPasswordRequest,
) error {

	email := normalizeEmail(req.Email)

	user, err := s.repository.GetUserByEmail(
		ctx,
		email,
	)

	// IMPORTANT:
	// Do not reveal whether the email exists.
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil
		}

		return fmt.Errorf(
			"get user for password reset: %w",
			err,
		)
	}

	if !user.IsActive || user.IsBanned {
		return nil
	}

	rawToken, tokenHash, err := generateSecureToken()
	if err != nil {
		return fmt.Errorf(
			"generate password reset token: %w",
			err,
		)
	}

	resetToken := &PasswordResetToken{
		ID:        uuid.NewString(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(30 * time.Minute),
		CreatedAt: time.Now(),
	}

	if err := s.repository.CreatePasswordResetToken(
		ctx,
		resetToken,
	); err != nil {
		return fmt.Errorf(
			"create password reset token: %w",
			err,
		)
	}

	// TODO:
	// Send rawToken through email service.
	_ = rawToken

	return nil
}

// ============================================================
// Reset Password
// ============================================================

func (s *Service) ResetPassword(
	ctx context.Context,
	req ResetPasswordRequest,
) error {

	if len(req.Password) < 8 {
		return ErrInvalidPassword
	}

	tokenHash := hashToken(req.Token)

	resetToken, err := s.repository.GetPasswordResetToken(
		ctx,
		tokenHash,
	)

	if err != nil {
		return ErrInvalidResetToken
	}

	if resetToken.UsedAt != nil {
		return ErrInvalidResetToken
	}

	if time.Now().After(resetToken.ExpiresAt) {
		return ErrExpiredToken
	}

	passwordHash, err := HashPassword(
		req.Password,
	)
	if err != nil {
		return fmt.Errorf(
			"hash new password: %w",
			err,
		)
	}

	if err := s.repository.UpdatePassword(
		ctx,
		resetToken.UserID,
		passwordHash,
	); err != nil {
		return fmt.Errorf(
			"update password: %w",
			err,
		)
	}

	if err := s.repository.MarkPasswordResetTokenUsed(
		ctx,
		tokenHash,
	); err != nil {
		return fmt.Errorf(
			"mark reset token used: %w",
			err,
		)
	}

	// Invalidate all existing sessions after a password reset.
	if err := s.repository.RevokeAllRefreshTokens(
		ctx,
		resetToken.UserID,
	); err != nil {
		return fmt.Errorf(
			"revoke refresh tokens: %w",
			err,
		)
	}

	return nil
}

// ============================================================
// Auth Response
// ============================================================

func (s *Service) createAuthResponse(
	ctx context.Context,
	user *User,
) (*AuthResponse, error) {

	accessToken, err := s.jwtManager.GenerateAccessToken(
		user.ID,
		user.Email,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"generate access token: %w",
			err,
		)
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(
		user.ID,
		user.Email,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"generate refresh token: %w",
			err,
		)
	}

	refreshTokenHash := hashToken(refreshToken)

	refreshTokenModel := &RefreshToken{
		ID:        uuid.NewString(),
		UserID:    user.ID,
		TokenHash: refreshTokenHash,
		ExpiresAt: time.Now().Add(
			s.jwtManager.refreshTokenExpiry,
		),
		CreatedAt: time.Now(),
	}

	if err := s.repository.CreateRefreshToken(
		ctx,
		refreshTokenModel,
	); err != nil {
		return nil, fmt.Errorf(
			"store refresh token: %w",
			err,
		)
	}

	return &AuthResponse{
		User:         *toUserResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(
			s.jwtManager.accessTokenExpiry.Seconds(),
		),
	}, nil
}

// ============================================================
// Helpers
// ============================================================

func normalizeEmail(email string) string {
	return strings.ToLower(
		strings.TrimSpace(email),
	)
}

func generateSecureToken() (
	string,
	string,
	error,
) {
	bytes := make([]byte, 32)

	if _, err := rand.Read(bytes); err != nil {
		return "", "", fmt.Errorf(
			"generate secure random token: %w",
			err,
		)
	}

	rawToken := hex.EncodeToString(bytes)

	return rawToken, hashToken(rawToken), nil
}

func hashToken(token string) string {
	hash := sha256.Sum256(
		[]byte(token),
	)

	return hex.EncodeToString(hash[:])
}

func toUserResponse(user *User) *UserResponse {
	return &UserResponse{
		ID:            user.ID,
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt,
	}
}