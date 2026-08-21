package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port        string
	Environment string

	DatabaseURL string

	MigrationsPath string

	JWTSecret string
	JWTIssuer string
}

func Load() (*Config, error) {
	// Load environment variables.
	// In production, these should be injected by the
	// deployment environment rather than relying on .env.
	loadDotEnv()

	cfg := &Config{
		Port:        getEnv("PORT", "8081"),
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
// Environment helpers
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

// --------------------------------------------------
// .env loader
// --------------------------------------------------

func loadDotEnv() {
	// We intentionally don't fail if .env doesn't exist.
	//
	// In production, environment variables should normally
	// come from Docker/Kubernetes/systemd/secrets management.
	//
	// If you're using godotenv:
	//
	//     _ = godotenv.Load()
	//
	// We'll add it once the dependency is installed.
}
