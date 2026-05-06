// Package migrate provides database migration helpers for containerized databases.
package migrate

import (
	"context"
	"database/sql"

	"github.com/atop0914/containerdb-bootcamp/internal/migrate"
)

// Runner handles database migrations for containerized databases.
type Runner struct {
	db       *sql.DB
	opts     []migrate.Option
	migrationsDir string
}

// NewRunner creates a new migration runner.
func NewRunner(db *sql.DB, opts ...migrate.Option) *Runner {
	return &Runner{
		db:             db,
		opts:           opts,
		migrationsDir:  "migrations",
	}
}

// Up runs all pending up migrations.
func (r *Runner) Up(ctx context.Context) error {
	return migrate.Run(ctx, r.db, r.migrationsDir, r.opts...)
}

// Down rolls back the last migration.
func (r *Runner) Down(ctx context.Context) error {
	migrations, err := migrate.ReadDirectionalMigrations(r.migrationsDir, migrate.Down)
	if err != nil {
		return err
	}

	if len(migrations) == 0 {
		return nil
	}

	// Get current version (last applied)
	last := migrations[len(migrations)-1]
	
	_, err = r.db.ExecContext(ctx, last.Name)
	return err
}

// ListMigrations returns all available migrations.
func ListMigrations(dir string) ([]migrate.Migration, error) {
	return migrate.ReadMigrations(dir)
}

// ForceVersion sets the schema version without running migrations.
func ForceVersion(ctx context.Context, db *sql.DB, version string, opts ...migrate.Option) error {
	return migrate.Force(ctx, db, version, opts...)
}
