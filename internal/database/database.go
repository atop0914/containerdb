// Package database provides common interfaces for containerized databases.
package database

import (
	"context"
	"database/sql"
)

// DB wraps a SQL database connection pool with its cleanup function.
type DB struct {
	Pool  *sql.DB
	Close func() // Releases container resources
}

// StartFunc is a function that starts a database container.
type StartFunc func(ctx context.Context) (*DB, error)

// Start starts the database and returns a cleanup function.
func (d *DB) Start(ctx context.Context, fn StartFunc) (*DB, error) {
	return fn(ctx)
}

// Ping checks if the database is reachable.
func (d *DB) Ping(ctx context.Context) error {
	return d.Pool.PingContext(ctx)
}
