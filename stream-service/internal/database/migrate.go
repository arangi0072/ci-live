package database

import (
	"context"
	"fmt"
	"time"

	"stream-service/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

//
// PostgreSQL Configuration
//

type PostgresConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

//
// New PostgreSQL Connection Pool
//

func NewPostgresPool(
	ctx context.Context,
	cfg config.PostgresConfig,
) (*pgxpool.Pool, error) {

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to parse postgres config: %w",
			err,
		)
	}

	//
	// Connection pool settings
	//

	if cfg.MaxConns > 0 {
		poolConfig.MaxConns = cfg.MaxConns
	}

	if cfg.MinConns > 0 {
		poolConfig.MinConns = cfg.MinConns
	}

	if cfg.MaxConnLifetime > 0 {
		poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	}

	if cfg.MaxConnIdleTime > 0 {
		poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	}

	//
	// Create pool
	//

	pool, err := pgxpool.NewWithConfig(
		ctx,
		poolConfig,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to create postgres pool: %w",
			err,
		)
	}

	//
	// Verify database connection.
	//

	pingCtx, cancel := context.WithTimeout(
		ctx,
		5*time.Second,
	)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()

		return nil, fmt.Errorf(
			"failed to ping postgres: %w",
			err,
		)
	}

	return pool, nil
}

//
// Close PostgreSQL Pool
//

func ClosePostgresPool(
	pool *pgxpool.Pool,
) {
	if pool == nil {
		return
	}

	pool.Close()
}

//
// Ping PostgreSQL
//

func PingPostgres(
	ctx context.Context,
	pool *pgxpool.Pool,
) error {

	if pool == nil {
		return fmt.Errorf(
			"postgres pool is nil",
		)
	}

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf(
			"postgres ping failed: %w",
			err,
		)
	}

	return nil
}
