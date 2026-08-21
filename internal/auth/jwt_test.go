package auth

import (
	"testing"
	"time"
)

func setupJWTManager() *JWTManager {
	return NewJWTManager(
		"test-secret-key-that-is-long-enough",
		15*time.Minute,
		24*time.Hour,
		"test-issuer",
	)
}

func TestGenerateAndValidateAccessToken(t *testing.T) {
	manager := setupJWTManager()
	userID := "user-123"
	email := "test@example.com"

	token, err := manager.GenerateAccessToken(userID, email)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" {
		t.Fatalf("expected valid token, got empty string")
	}

	claims, err := manager.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("expected no error validating token, got %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("expected userID %s, got %s", userID, claims.UserID)
	}
	if claims.Email != email {
		t.Errorf("expected email %s, got %s", email, claims.Email)
	}
	if claims.TokenType != "access" {
		t.Errorf("expected token type access, got %s", claims.TokenType)
	}
}

func TestGenerateAndValidateRefreshToken(t *testing.T) {
	manager := setupJWTManager()
	userID := "user-123"
	email := "test@example.com"

	token, err := manager.GenerateRefreshToken(userID, email)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	claims, err := manager.ValidateRefreshToken(token)
	if err != nil {
		t.Fatalf("expected no error validating token, got %v", err)
	}
	if claims.TokenType != "refresh" {
		t.Errorf("expected token type refresh, got %s", claims.TokenType)
	}
}

func TestValidateToken_WrongType(t *testing.T) {
	manager := setupJWTManager()
	
	accessToken, _ := manager.GenerateAccessToken("user-1", "user@test.com")
	refreshToken, _ := manager.GenerateRefreshToken("user-1", "user@test.com")

	// Validate access token using refresh token method should fail
	_, err := manager.ValidateRefreshToken(accessToken)
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken when validating access token as refresh token, got %v", err)
	}

	// Validate refresh token using access token method should fail
	_, err = manager.ValidateAccessToken(refreshToken)
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken when validating refresh token as access token, got %v", err)
	}
}

func TestValidateToken_Expired(t *testing.T) {
	// Create manager with very short expiry
	manager := NewJWTManager("secret", -1*time.Minute, -1*time.Minute, "issuer")

	token, _ := manager.GenerateAccessToken("user-1", "test@test.com")
	
	_, err := manager.ValidateAccessToken(token)
	if err != ErrExpiredToken {
		t.Errorf("expected ErrExpiredToken, got %v", err)
	}
}

func TestValidateToken_InvalidSignature(t *testing.T) {
	manager1 := setupJWTManager()
	manager2 := NewJWTManager("different-secret-key", 15*time.Minute, 24*time.Hour, "test-issuer")

	token, _ := manager1.GenerateAccessToken("user-1", "test@test.com")

	// Validate token from manager1 using manager2
	_, err := manager2.ValidateAccessToken(token)
	if err == nil {
		t.Fatal("expected error validating token with wrong secret, got nil")
	}
}
