// Package migrate provides database migration utilities for containerized databases.
package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Direction represents migration direction.
type Direction bool

const (
	// Up migration direction
	Up Direction = true
	// Down migration direction
	Down Direction = false
)

// Config holds migration configuration.
type Config struct {
	// TableName is the migration table name (default: "schema_migrations")
	TableName string
	// Timeout for each migration step
	Timeout time.Duration
}

// Option applies configuration to Config.
type Option func(*Config)

// WithTableName sets the migration table name.
func WithTableName(name string) Option {
	return func(c *Config) {
		c.TableName = name
	}
}

// WithTimeout sets the timeout for each migration step.
func WithTimeout(d time.Duration) Option {
	return func(c *Config) {
		c.Timeout = d
	}
}

// defaultConfig returns Config with default values.
func defaultConfig() Config {
	return Config{
		TableName: "schema_migrations",
		Timeout:   30 * time.Second,
	}
}

// Migration represents a single migration file.
type Migration struct {
	Version   string
	Name      string
	Direction Direction
}

// Run executes migrations from a directory.
func Run(ctx context.Context, db *sql.DB, dir string, opts ...Option) error {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	migrations, err := ReadMigrations(dir)
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	if len(migrations) == 0 {
		return nil
	}

	// Ensure migration table exists
	if err := createMigrationsTable(ctx, db, cfg.TableName); err != nil {
		return err
	}

	for _, m := range migrations {
		applied, err := isApplied(ctx, db, cfg.TableName, m.Version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		sqlContent, err := os.ReadFile(m.Name)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", m.Name, err)
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		_, err = tx.ExecContext(ctx, string(sqlContent))
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", m.Version, err)
		}

		_, err = tx.ExecContext(ctx,
			fmt.Sprintf("INSERT INTO %s (version, applied_at) VALUES (?, ?)", cfg.TableName),
			m.Version, time.Now())
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", m.Version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", m.Version, err)
		}
	}

	return nil
}

// ReadMigrations reads migration files from a directory.
// Expects files named: {version}_{name}.up.sql or {version}_{name}.down.sql
// Only .up.sql files are included by default (use ReadDirectionalMigrations for both).
func ReadMigrations(dir string) ([]Migration, error) {
	return ReadDirectionalMigrations(dir, Up)
}

// ReadDirectionalMigrations reads only up or down migrations from a directory.
func ReadDirectionalMigrations(dir string, direction Direction) ([]Migration, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var migrations []Migration
	suffix := ".up.sql"
	if direction == Down {
		suffix = ".down.sql"
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, suffix) {
			continue
		}

		// Parse version from filename
		base := strings.TrimSuffix(name, suffix)
		parts := strings.SplitN(base, "_", 2)
		if len(parts) < 1 {
			continue
		}
		version := parts[0]

		migrations = append(migrations, Migration{
			Version:   version,
			Name:      filepath.Join(dir, name),
			Direction: direction,
		})
	}

	// Sort by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// createMigrationsTable creates the migrations tracking table.
func createMigrationsTable(ctx context.Context, db *sql.DB, tableName string) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP NOT NULL
		)
	`, tableName)
	_, err := db.ExecContext(ctx, query)
	return err
}

// isApplied checks if a migration version has been applied.
func isApplied(ctx context.Context, db *sql.DB, tableName, version string) (bool, error) {
	query := fmt.Sprintf("SELECT 1 FROM %s WHERE version = ?", tableName)
	var exists int
	err := db.QueryRowContext(ctx, query, version).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// Force sets the migration version without running migrations.
// Useful for initial schema setup.
func Force(ctx context.Context, db *sql.DB, version string, opts ...Option) error {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	if err := createMigrationsTable(ctx, db, cfg.TableName); err != nil {
		return err
	}

	_, err := db.ExecContext(ctx,
		fmt.Sprintf("INSERT OR REPLACE INTO %s (version, applied_at) VALUES (?, ?)", cfg.TableName),
		version, time.Now())
	return err
}
