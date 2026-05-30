// Package postgres provides PostgreSQL container management for testing.
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"github.com/atop0914/containerdb/internal/config"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

// Option configures PostgreSQL container options.
type Option func(*config.PostgresConfig)

// WithImage sets the Docker image for PostgreSQL.
func WithImage(image string) Option {
	return func(c *config.PostgresConfig) {
		c.Image = image
	}
}

// WithUsername sets the PostgreSQL username.
func WithUsername(username string) Option {
	return func(c *config.PostgresConfig) {
		c.Username = username
	}
}

// WithPassword sets the PostgreSQL password.
func WithPassword(password string) Option {
	return func(c *config.PostgresConfig) {
		c.Password = password
	}
}

// WithDatabase sets the PostgreSQL database name.
func WithDatabase(database string) Option {
	return func(c *config.PostgresConfig) {
		c.Database = database
	}
}

// WithHealthCheckTimeout sets the health check timeout duration.
func WithHealthCheckTimeout(timeout time.Duration) Option {
	return func(c *config.PostgresConfig) {
		c.HealthCheckTimeout = timeout
	}
}

// WithPoolSettings configures the connection pool.
func WithPoolSettings(maxOpen, maxIdle int, maxLifetime time.Duration) Option {
	return func(c *config.PostgresConfig) {
		c.MaxOpenConns = maxOpen
		c.MaxIdleConns = maxIdle
		c.ConnMaxLifetime = maxLifetime
	}
}

// WithHealthCheckRetry configures health check retry attempts.
func WithHealthCheckRetry(retries int) Option {
	return func(c *config.PostgresConfig) {
		if retries > 0 {
			c.HealthCheckRetries = retries
		}
	}
}

// WithHealthCheckInterval configures the interval between health check retries.
func WithHealthCheckInterval(interval time.Duration) Option {
	return func(c *config.PostgresConfig) {
		if interval > 0 {
			c.HealthCheckInterval = interval
		}
	}
}

// New starts a new PostgreSQL container and returns a connected *sql.DB.
// Call the returned cleanup function to stop and remove the container.
//
// Example:
//
//	ctx := context.Background()
//	db, cleanup, err := postgres.New(ctx)
//	if err != nil {
//	    panic(err)
//	}
//	defer cleanup()
func New(ctx context.Context) (*sql.DB, func(), error) {
	cfg := config.DefaultPostgresConfig()
	return NewWithConfig(ctx, cfg)
}

// NewWithOptions starts a PostgreSQL container with functional options.
func NewWithOptions(ctx context.Context, opts ...Option) (*sql.DB, func(), error) {
	cfg := config.DefaultPostgresConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	return NewWithConfig(ctx, cfg)
}

// NewWithConfig starts a PostgreSQL container with custom configuration.
func NewWithConfig(ctx context.Context, cfg *config.PostgresConfig) (*sql.DB, func(), error) {
	dbName := cfg.Database
	dbUser := cfg.Username
	dbPassword := cfg.Password

	pgContainer, err := postgres.Run(ctx,
		cfg.Image,
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	// Get connection details using ConnectionString (v0.42.0+ API)
	connStr, err := pgContainer.ConnectionString(ctx)
	if err != nil {
		_ = pgContainer.Terminate(ctx)
		return nil, nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	pool, err := sql.Open("postgres", connStr)
	if err != nil {
		_ = pgContainer.Terminate(ctx)
		return nil, nil, fmt.Errorf("failed to open db: %w", err)
	}

	// Apply pool settings from config
	pool.SetMaxOpenConns(cfg.MaxOpenConns)
	pool.SetMaxIdleConns(cfg.MaxIdleConns)
	pool.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	ctx, cancel := context.WithTimeout(ctx, cfg.HealthCheckTimeout)
	defer cancel()

	if err := pool.PingContext(ctx); err != nil {
		pool.Close()
		_ = pgContainer.Terminate(ctx)
		return nil, nil, fmt.Errorf("postgres not ready: %w", err)
	}

	cleanup := func() {
		pool.Close()
		_ = pgContainer.Terminate(context.Background())
	}

	return pool, cleanup, nil
}

// MustNew starts a PostgreSQL container and panics on error.
// Use for tests where you want quick setup.
func MustNew(ctx context.Context) (*sql.DB, func()) {
	db, cleanup, err := New(ctx)
	if err != nil {
		panic(fmt.Errorf("MustNew failed: %w", err))
	}
	return db, cleanup
}

// MustNewWithOptions starts a PostgreSQL container with options and panics on error.
func MustNewWithOptions(ctx context.Context, opts ...Option) (*sql.DB, func()) {
	db, cleanup, err := NewWithOptions(ctx, opts...)
	if err != nil {
		panic(fmt.Errorf("MustNewWithOptions failed: %w", err))
	}
	return db, cleanup
}

// Container exposes the underlying testcontainers PostgreSQL module for advanced use.
func Container(ctx context.Context, opts ...Option) (*postgres.PostgresContainer, error) {
	cfg := config.DefaultPostgresConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	pgContainer, err := postgres.Run(ctx,
		cfg.Image,
		postgres.WithDatabase(cfg.Database),
		postgres.WithUsername(cfg.Username),
		postgres.WithPassword(cfg.Password),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres container: %w", err)
	}

	return pgContainer, nil
}
