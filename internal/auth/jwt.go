package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

type Claims struct {
	UserID string `json:"sub"`
	Email  string `json:"email"`

	TokenType string `json:"type"`

	jwt.RegisteredClaims
}

type JWTManager struct {
	secret               []byte
	accessTokenExpiry    time.Duration
	refreshTokenExpiry   time.Duration
	issuer               string
}

func NewJWTManager(
	secret string,
	accessTokenExpiry time.Duration,
	refreshTokenExpiry time.Duration,
	issuer string,
) *JWTManager {
	return &JWTManager{
		secret:             []byte(secret),
		accessTokenExpiry:  accessTokenExpiry,
		refreshTokenExpiry: refreshTokenExpiry,
		issuer:             issuer,
	}
}

func (j *JWTManager) GenerateAccessToken(
	userID string,
	email string,
) (string, error) {
	now := time.Now()

	claims := Claims{
		UserID: userID,
		Email:  email,
		TokenType: "access",

		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    j.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.accessTokenExpiry)),
		},
	}

	return j.sign(claims)
}

func (j *JWTManager) GenerateRefreshToken(
	userID string,
	email string,
) (string, error) {
	now := time.Now()

	claims := Claims{
		UserID: userID,
		Email:  email,
		TokenType: "refresh",

		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    j.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.refreshTokenExpiry)),
		},
	}

	return j.sign(claims)
}

func (j *JWTManager) ValidateAccessToken(
	tokenString string,
) (*Claims, error) {
	claims, err := j.validate(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "access" {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (j *JWTManager) ValidateRefreshToken(
	tokenString string,
) (*Claims, error) {
	claims, err := j.validate(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "refresh" {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (j *JWTManager) validate(
	tokenString string,
) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			// Make sure the token uses HMAC.
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf(
					"%w: unexpected signing method",
					ErrInvalidToken,
				)
			}

			return j.secret, nil
		},
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}

		return nil, fmt.Errorf(
			"%w: %v",
			ErrInvalidToken,
			err,
		)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.Issuer != j.issuer {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (j *JWTManager) sign(claims Claims) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	tokenString, err := token.SignedString(j.secret)
	if err != nil {
		return "", fmt.Errorf(
			"sign jwt: %w",
			err,
		)
	}

	return tokenString, nil
}