package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port        string
	Environment string

	DatabaseURL    string
	MigrationsPath string

	JWTSecret string
	JWTIssuer string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:        getEnv("PORT", "8082"),
		Environment: getEnv("ENVIRONMENT", "development"),

		DatabaseURL: getEnv(
			"DATABASE_URL",
			"",
		),

		MigrationsPath: getEnv(
			"MIGRATIONS_PATH",
			"./migrations",
		),

		JWTSecret: getEnv(
			"JWT_SECRET",
			"",
		),

		JWTIssuer: getEnv(
			"JWT_ISSUER",
			"live-stream-auth-service",
		),
	}

	// --------------------------------------------------
	// Required configuration
	// --------------------------------------------------

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf(
			"DATABASE_URL is required",
		)
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf(
			"JWT_SECRET is required",
		)
	}

	if len(cfg.JWTSecret) < 32 {
		return nil, fmt.Errorf(
			"JWT_SECRET must be at least 32 characters",
		)
	}

	if cfg.MigrationsPath == "" {
		return nil, fmt.Errorf(
			"MIGRATIONS_PATH is required",
		)
	}

	return cfg, nil
}

// --------------------------------------------------
// Environment helper
// --------------------------------------------------

func getEnv(
	key string,
	fallback string,
) string {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	return value
}
