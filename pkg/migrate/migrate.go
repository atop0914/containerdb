// Package migrate provides database migration helpers for containerized databases.
package migrate

import (
	"context"
	"database/sql"
	"time"

	"github.com/atop0914/containerdb/internal/migrate"
)

// Runner handles database migrations for containerized databases.
type Runner struct {
	db            *sql.DB
	dir           string
	opts          []migrate.Option
}

// Option configures the Runner.
type Option func(*Runner)

// WithDir sets the migrations directory.
func WithDir(dir string) Option {
	return func(r *Runner) {
		r.dir = dir
	}
}

// WithTableName sets the migration tracking table name.
func WithTableName(name string) Option {
	return func(r *Runner) {
		r.opts = append(r.opts, migrate.WithTableName(name))
	}
}

// WithTimeout sets the timeout for each migration step.
func WithTimeout(d time.Duration) Option {
	return func(r *Runner) {
		r.opts = append(r.opts, migrate.WithTimeout(d))
	}
}

// NewRunner creates a new migration runner.
func NewRunner(db *sql.DB, opts ...Option) *Runner {
	r := &Runner{
		db:  db,
		dir: "migrations",
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Up runs all pending up migrations.
func (r *Runner) Up(ctx context.Context) error {
	return migrate.Run(ctx, r.db, r.dir, r.opts...)
}

// Down rolls back the last N migrations. If steps <= 0, rolls back all.
func (r *Runner) Down(ctx context.Context, steps int) error {
	return migrate.Rollback(ctx, r.db, r.dir, steps, r.opts...)
}

// Status returns the status of all migrations.
func (r *Runner) Status(ctx context.Context) ([]migrate.Status, error) {
	return migrate.GetStatus(ctx, r.db, r.dir, r.opts...)
}

// CreateMigration creates a new migration file pair (up and down).
func CreateMigration(dir, name string) (string, error) {
	return migrate.CreateMigration(dir, name)
}

// ListMigrations returns all available migrations.
func ListMigrations(dir string) ([]migrate.Migration, error) {
	return migrate.ReadMigrations(dir)
}

// ForceVersion sets the schema version without running migrations.
func ForceVersion(ctx context.Context, db *sql.DB, version string, opts ...migrate.Option) error {
	return migrate.Force(ctx, db, version, opts...)
}
