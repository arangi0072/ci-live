package database

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//
// Run Migration From SQL File
//

func RunMigration(
	ctx context.Context,
	db *pgxpool.Pool,
	migrationPath string,
) error {

	if db == nil {
		return fmt.Errorf(
			"postgres pool is nil",
		)
	}

	if strings.TrimSpace(migrationPath) == "" {
		return fmt.Errorf(
			"migration path cannot be empty",
		)
	}

	//
	// Read migration file.
	//

	data, err := os.ReadFile(migrationPath)
	if err != nil {
		return fmt.Errorf(
			"failed to read migration file: %w",
			err,
		)
	}

	sql := strings.TrimSpace(
		string(data),
	)

	if sql == "" {
		return fmt.Errorf(
			"migration file is empty",
		)
	}

	//
	// Execute migration.
	//

	if _, err := db.Exec(
		ctx,
		sql,
	); err != nil {
		return fmt.Errorf(
			"failed to execute database migration: %w",
			err,
		)
	}

	return nil
}

//
// Run Migration With Transaction
//
// Useful when you want the entire migration to
// succeed or fail as one operation.
//

func RunMigrationTx(
	ctx context.Context,
	db *pgxpool.Pool,
	migrationPath string,
) error {

	if db == nil {
		return fmt.Errorf(
			"postgres pool is nil",
		)
	}

	if strings.TrimSpace(migrationPath) == "" {
		return fmt.Errorf(
			"migration path cannot be empty",
		)
	}

	//
	// Read migration file.
	//

	data, err := os.ReadFile(migrationPath)
	if err != nil {
		return fmt.Errorf(
			"failed to read migration file: %w",
			err,
		)
	}

	sql := strings.TrimSpace(
		string(data),
	)

	if sql == "" {
		return fmt.Errorf(
			"migration file is empty",
		)
	}

	//
	// Begin transaction.
	//

	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf(
			"failed to begin migration transaction: %w",
			err,
		)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	//
	// Execute migration.
	//

	if _, err := tx.Exec(
		ctx,
		sql,
	); err != nil {
		return fmt.Errorf(
			"failed to execute migration: %w",
			err,
		)
	}

	//
	// Commit.
	//

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf(
			"failed to commit migration: %w",
			err,
		)
	}

	return nil
}
