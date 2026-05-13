// Package mysql provides MySQL container management for testing.
package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/atop0914/containerdb-bootcamp/internal/config"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
)

// Option is a functional option for MySQL configuration.
type Option func(*config.MySQLConfig)

// WithImage sets the Docker image for MySQL.
func WithImage(image string) Option {
	return func(cfg *config.MySQLConfig) {
		cfg.Image = image
	}
}

// WithUsername sets the MySQL username.
func WithUsername(username string) Option {
	return func(cfg *config.MySQLConfig) {
		cfg.Username = username
	}
}

// WithPassword sets the MySQL password.
func WithPassword(password string) Option {
	return func(cfg *config.MySQLConfig) {
		cfg.Password = password
	}
}

// WithDatabase sets the MySQL database name.
func WithDatabase(database string) Option {
	return func(cfg *config.MySQLConfig) {
		cfg.Database = database
	}
}

// WithHealthCheckTimeout sets the timeout for container health checks.
func WithHealthCheckTimeout(timeout time.Duration) Option {
	return func(cfg *config.MySQLConfig) {
		cfg.HealthCheckTimeout = timeout
	}
}

// WithPoolSettings configures connection pool settings.
func WithPoolSettings(maxOpen, maxIdle int, maxLifetime time.Duration) Option {
	return func(cfg *config.MySQLConfig) {
		cfg.MaxOpenConns = maxOpen
		cfg.MaxIdleConns = maxIdle
		cfg.ConnMaxLifetime = maxLifetime
	}
}

// WithHealthCheckRetry configures health check retry attempts.
func WithHealthCheckRetry(retries int) Option {
	return func(cfg *config.MySQLConfig) {
		if retries > 0 {
			cfg.HealthCheckRetries = retries
		}
	}
}

// WithHealthCheckInterval configures the interval between health check retries.
func WithHealthCheckInterval(interval time.Duration) Option {
	return func(cfg *config.MySQLConfig) {
		if interval > 0 {
			cfg.HealthCheckInterval = interval
		}
	}
}

// New starts a new MySQL container and returns a connected *sql.DB.
// Call the returned cleanup function to stop and remove the container.
//
// Example:
//
//	ctx := context.Background()
//	db, cleanup, err := mysql.New(ctx)
//	if err != nil {
//	    panic(err)
//	}
//	defer cleanup()
func New(ctx context.Context) (*sql.DB, func(), error) {
	cfg := config.DefaultMySQLConfig()
	return NewWithConfig(ctx, cfg)
}

// NewWithConfig starts a MySQL container with custom configuration.
func NewWithConfig(ctx context.Context, cfg *config.MySQLConfig) (*sql.DB, func(), error) {
	mysqlContainer, err := mysql.Run(ctx,
		cfg.Image,
		mysql.WithUsername(cfg.Username),
		mysql.WithPassword(cfg.Password),
		mysql.WithDatabase(cfg.Database),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start mysql container: %w", err)
	}

	// Get connection details using ConnectionString (v0.42.0+ API)
	connStr, err := mysqlContainer.ConnectionString(ctx)
	if err != nil {
		mysqlContainer.Terminate(ctx)
		return nil, nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	pool, err := sql.Open("mysql", connStr)
	if err != nil {
		mysqlContainer.Terminate(ctx)
		return nil, nil, fmt.Errorf("failed to open db: %w", err)
	}

	// Configure pool
	pool.SetMaxOpenConns(cfg.MaxOpenConns)
	pool.SetMaxIdleConns(cfg.MaxIdleConns)
	pool.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Verify connection with timeout
	ctx, cancel := context.WithTimeout(ctx, cfg.HealthCheckTimeout)
	defer cancel()

	if err := pool.PingContext(ctx); err != nil {
		pool.Close()
		mysqlContainer.Terminate(ctx)
		return nil, nil, fmt.Errorf("mysql not ready: %w", err)
	}

	cleanup := func() {
		pool.Close()
		mysqlContainer.Terminate(context.Background())
	}

	return pool, cleanup, nil
}

// NewWithOptions starts a MySQL container with functional options.
// This is the preferred way to create a customized MySQL container.
//
// Example:
//
//	ctx := context.Background()
//	db, cleanup, err := mysql.NewWithOptions(ctx,
//	    mysql.WithImage("mysql:8.0"),
//	    mysql.WithUsername("myuser"),
//	    mysql.WithPassword("mypass"),
//	    mysql.WithDatabase("mydb"),
//	)
//	if err != nil {
//	    panic(err)
//	}
//	defer cleanup()
func NewWithOptions(ctx context.Context, opts ...Option) (*sql.DB, func(), error) {
	cfg := config.DefaultMySQLConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	return NewWithConfig(ctx, cfg)
}

// MustNew starts a new MySQL container and panics on error.
// Use for quick setup in tests.
func MustNew(ctx context.Context) (*sql.DB, func()) {
	db, cleanup, err := New(ctx)
	if err != nil {
		panic(fmt.Errorf("MustNew failed: %w", err))
	}
	return db, cleanup
}

// MustNewWithOptions starts a MySQL container with options and panics on error.
func MustNewWithOptions(ctx context.Context, opts ...Option) (*sql.DB, func()) {
	db, cleanup, err := NewWithOptions(ctx, opts...)
	if err != nil {
		panic(fmt.Errorf("MustNewWithOptions failed: %w", err))
	}
	return db, cleanup
}

// Container exposes the underlying testcontainers MySQL module for advanced use.
func Container(ctx context.Context) (*mysql.MySQLContainer, error) {
	cfg := config.DefaultMySQLConfig()
	return mysql.Run(ctx, cfg.Image)
}

// NewWithOptionsContainer starts a MySQL container with functional options and returns the raw container.
func NewWithOptionsContainer(ctx context.Context, opts ...Option) (*mysql.MySQLContainer, error) {
	cfg := config.DefaultMySQLConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	return mysql.Run(ctx,
		cfg.Image,
		mysql.WithUsername(cfg.Username),
		mysql.WithPassword(cfg.Password),
		mysql.WithDatabase(cfg.Database),
	)
}
