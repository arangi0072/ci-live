package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewPostgres(databaseURL string) (*sql.DB, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("database URL is empty")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf(
			"open postgres connection: %w",
			err,
		)
	}

	// --------------------------------------------------
	// Connection pool configuration
	// --------------------------------------------------

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)

	db.SetConnMaxLifetime(
		30 * time.Minute,
	)

	db.SetConnMaxIdleTime(
		5 * time.Minute,
	)

	// --------------------------------------------------
	// Verify database connectivity
	// --------------------------------------------------

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)

	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()

		return nil, fmt.Errorf(
			"ping postgres: %w",
			err,
		)
	}

	return db, nil
}
