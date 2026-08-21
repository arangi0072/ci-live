package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func RunMigrations(
	ctx context.Context,
	db *sql.DB,
	migrationsPath string,
) error {
	if err := createMigrationsTable(ctx, db); err != nil {
		return err
	}

	files, err := os.ReadDir(migrationsPath)
	if err != nil {
		return fmt.Errorf("read migrations directory: %w", err)
	}

	var migrationFiles []string

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		migrationFiles = append(
			migrationFiles,
			file.Name(),
		)
	}

	sort.Strings(migrationFiles)

	for _, filename := range migrationFiles {
		if err := runMigration(
			ctx,
			db,
			migrationsPath,
			filename,
		); err != nil {
			return fmt.Errorf(
				"migration %s failed: %w",
				filename,
				err,
			)
		}
	}

	return nil
}

func createMigrationsTable(
	ctx context.Context,
	db *sql.DB,
) error {
	const query = `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`

	if _, err := db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf(
			"create schema_migrations table: %w",
			err,
		)
	}

	return nil
}

func runMigration(
	ctx context.Context,
	db *sql.DB,
	migrationsPath string,
	filename string,
) error {
	applied, err := migrationApplied(
		ctx,
		db,
		filename,
	)
	if err != nil {
		return err
	}

	if applied {
		return nil
	}

	path := filepath.Join(
		migrationsPath,
		filename,
	)

	sqlBytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf(
			"read migration file: %w",
			err,
		)
	}

	migrationSQL := strings.TrimSpace(
		string(sqlBytes),
	)

	if migrationSQL == "" {
		return fmt.Errorf(
			"migration file %s is empty",
			filename,
		)
	}

	tx, err := db.BeginTx(
		ctx,
		nil,
	)
	if err != nil {
		return fmt.Errorf(
			"begin migration transaction: %w",
			err,
		)
	}

	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.ExecContext(
		ctx,
		migrationSQL,
	); err != nil {
		return fmt.Errorf(
			"execute migration: %w",
			err,
		)
	}

	const insertQuery = `
		INSERT INTO schema_migrations (version)
		VALUES ($1)
	`

	if _, err := tx.ExecContext(
		ctx,
		insertQuery,
		filename,
	); err != nil {
		return fmt.Errorf(
			"record migration: %w",
			err,
		)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf(
			"commit migration: %w",
			err,
		)
	}

	return nil
}

func migrationApplied(
	ctx context.Context,
	db *sql.DB,
	version string,
) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM schema_migrations
			WHERE version = $1
		)
	`

	var exists bool

	if err := db.QueryRowContext(
		ctx,
		query,
		version,
	).Scan(&exists); err != nil {
		return false, fmt.Errorf(
			"check migration %s: %w",
			version,
			err,
		)
	}

	return exists, nil
}