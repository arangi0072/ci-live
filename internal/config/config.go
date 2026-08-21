package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string

	Environment string

	DatabaseURL string

	JWTSecret string

	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
}

func Load() (*Config, error) {
	// Load .env if it exists.
	// In production, environment variables should be provided
	// directly by the deployment environment.
	_ = godotenv.Load()

	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),

		DatabaseURL: getEnv("DATABASE_URL", ""),

		JWTSecret: getEnv("JWT_SECRET", ""),

		AccessTokenExpiry: getDuration(
			"ACCESS_TOKEN_EXPIRY",
			15*time.Minute,
		),

		RefreshTokenExpiry: getDuration(
			"REFRESH_TOKEN_EXPIRY",
			30*24*time.Hour,
		),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	if len(cfg.JWTSecret) < 32 {
		return nil, fmt.Errorf(
			"JWT_SECRET must be at least 32 characters",
		)
	}

	return cfg, nil
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	return value
}

func getDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return duration
}

func getBool(key string, fallback bool) bool {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}