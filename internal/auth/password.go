package auth

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

var (
	ErrInvalidPassword = errors.New("invalid password")
)

func HashPassword(password string) (string, error) {
	if password == "" {
		return "", ErrInvalidPassword
	}

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcryptCost,
	)
	if err != nil {
		return "", fmt.Errorf(
			"hash password: %w",
			err,
		)
	}

	return string(hash), nil
}

func ComparePassword(
	hashedPassword string,
	password string,
) error {
	if hashedPassword == "" || password == "" {
		return ErrInvalidPassword
	}

	err := bcrypt.CompareHashAndPassword(
		[]byte(hashedPassword),
		[]byte(password),
	)

	if err != nil {
		return ErrInvalidPassword
	}

	return nil
}