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

// Status represents the status of a migration.
type Status struct {
	Version    string
	Name       string
	Applied    bool
	AppliedAt  *time.Time
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

// Rollback rolls back the last N applied migrations.
func Rollback(ctx context.Context, db *sql.DB, dir string, steps int, opts ...Option) error {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	// Get applied versions in reverse order
	appliedVersions, err := getAppliedVersions(ctx, db, cfg.TableName)
	if err != nil {
		return fmt.Errorf("failed to get applied versions: %w", err)
	}

	if len(appliedVersions) == 0 {
		return nil
	}

	// Read down migrations
	downMigrations, err := ReadDirectionalMigrations(dir, Down)
	if err != nil {
		return fmt.Errorf("failed to read down migrations: %w", err)
	}

	// Create a map of down migrations by version
	downMap := make(map[string]Migration)
	for _, m := range downMigrations {
		downMap[m.Version] = m
	}

	// Rollback specified number of steps
	if steps <= 0 || steps > len(appliedVersions) {
		steps = len(appliedVersions)
	}

	for i := 0; i < steps; i++ {
		version := appliedVersions[i]
		down, ok := downMap[version]
		if !ok {
			return fmt.Errorf("no down migration found for version %s", version)
		}

		sqlContent, err := os.ReadFile(down.Name)
		if err != nil {
			return fmt.Errorf("failed to read down migration %s: %w", down.Name, err)
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		_, err = tx.ExecContext(ctx, string(sqlContent))
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute down migration %s: %w", version, err)
		}

		_, err = tx.ExecContext(ctx,
			fmt.Sprintf("DELETE FROM %s WHERE version = ?", cfg.TableName),
			version)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete migration record %s: %w", version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit rollback %s: %w", version, err)
		}
	}

	return nil
}

// GetStatus returns the status of all migrations in a directory.
func GetStatus(ctx context.Context, db *sql.DB, dir string, opts ...Option) ([]Status, error) {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	migrations, err := ReadMigrations(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations: %w", err)
	}

	// Ensure migration table exists
	if err := createMigrationsTable(ctx, db, cfg.TableName); err != nil {
		return nil, err
	}

	// Get applied versions with timestamps
	appliedMap, err := getAppliedVersionsWithTime(ctx, db, cfg.TableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get applied versions: %w", err)
	}

	statuses := make([]Status, 0, len(migrations))
	for _, m := range migrations {
		s := Status{
			Version: m.Version,
			Name:    filepath.Base(m.Name),
			Applied: false,
		}
		if appliedAt, ok := appliedMap[m.Version]; ok {
			s.Applied = true
			s.AppliedAt = &appliedAt
		}
		statuses = append(statuses, s)
	}

	return statuses, nil
}

// CreateMigration creates a new migration file pair (up and down).
func CreateMigration(dir, name string) (string, error) {
	// Ensure directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create migrations directory: %w", err)
	}

	// Generate version from timestamp
	version := time.Now().Format("20060102150405")
	
	// Clean the name
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ToLower(name)

	// Create up migration
	upFile := filepath.Join(dir, fmt.Sprintf("%s_%s.up.sql", version, name))
	if err := os.WriteFile(upFile, []byte("-- Migration: "+name+"\n\n"), 0644); err != nil {
		return "", fmt.Errorf("failed to create up migration: %w", err)
	}

	// Create down migration
	downFile := filepath.Join(dir, fmt.Sprintf("%s_%s.down.sql", version, name))
	if err := os.WriteFile(downFile, []byte("-- Rollback: "+name+"\n\n"), 0644); err != nil {
		return "", fmt.Errorf("failed to create down migration: %w", err)
	}

	return version, nil
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

// getAppliedVersions returns all applied versions in reverse order.
func getAppliedVersions(ctx context.Context, db *sql.DB, tableName string) ([]string, error) {
	query := fmt.Sprintf("SELECT version FROM %s ORDER BY applied_at DESC", tableName)
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []string
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}
	return versions, nil
}

// getAppliedVersionsWithTime returns a map of version to applied time.
func getAppliedVersionsWithTime(ctx context.Context, db *sql.DB, tableName string) (map[string]time.Time, error) {
	query := fmt.Sprintf("SELECT version, applied_at FROM %s", tableName)
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]time.Time)
	for rows.Next() {
		var version string
		var appliedAt time.Time
		if err := rows.Scan(&version, &appliedAt); err != nil {
			return nil, err
		}
		result[version] = appliedAt
	}
	return result, nil
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
