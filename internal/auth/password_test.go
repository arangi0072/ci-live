package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		password := "my-secure-password"
		hash, err := HashPassword(password)
		
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		
		if hash == "" {
			t.Fatalf("expected a valid hash, got empty string")
		}
		
		if hash == password {
			t.Fatalf("expected hash to be different from password")
		}
	})
	
	t.Run("empty password", func(t *testing.T) {
		_, err := HashPassword("")
		if err != ErrInvalidPassword {
			t.Fatalf("expected ErrInvalidPassword, got %v", err)
		}
	})
}

func TestComparePassword(t *testing.T) {
	password := "another-secure-password"
	hash, err := HashPassword(password)
	
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	t.Run("correct password", func(t *testing.T) {
		err := ComparePassword(hash, password)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("incorrect password", func(t *testing.T) {
		err := ComparePassword(hash, "wrong-password")
		if err != ErrInvalidPassword {
			t.Fatalf("expected ErrInvalidPassword, got %v", err)
		}
	})

	t.Run("empty hash", func(t *testing.T) {
		err := ComparePassword("", password)
		if err != ErrInvalidPassword {
			t.Fatalf("expected ErrInvalidPassword, got %v", err)
		}
	})

	t.Run("empty password", func(t *testing.T) {
		err := ComparePassword(hash, "")
		if err != ErrInvalidPassword {
			t.Fatalf("expected ErrInvalidPassword, got %v", err)
		}
	})
}
