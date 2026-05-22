// Package sqlite provides SQLite helpers for testing without containers.
package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	"github.com/atop0914/containerdb/internal/config"
)

// Option is a functional option for SQLite configuration.
type Option func(*config.SQLiteConfig)

// WithMode sets the database mode: "memory", "temp", or "file".
func WithMode(mode string) Option {
	return func(cfg *config.SQLiteConfig) {
		cfg.Mode = mode
	}
}

// WithPath sets the file path for persistent SQLite databases.
func WithPath(path string) Option {
	return func(cfg *config.SQLiteConfig) {
		cfg.Path = path
	}
}

// WithCache sets the cache mode: "shared", "private", or "write".
func WithCache(cache string) Option {
	return func(cfg *config.SQLiteConfig) {
		cfg.Cache = cache
	}
}

// InMemory creates an in-memory SQLite database.
// This is the fastest option, useful for unit tests.
//
// Example:
//
//	db, cleanup, err := sqlite.InMemory()
//	if err != nil {
//	    panic(err)
//	}
//	defer cleanup()
func InMemory() (*sql.DB, func(), error) {
	cfg := config.DefaultSQLiteConfig()
	cfg.Mode = "memory"
	return newSQLite(cfg)
}

// TempDB creates a temporary SQLite database file.
// The file is automatically removed when the database is closed.
// Useful when you need persistence but automatic cleanup.
//
// Example:
//
//	db, cleanup, err := sqlite.TempDB()
//	if err != nil {
//	    panic(err)
//	}
//	defer cleanup()
func TempDB() (*sql.DB, func(), error) {
	cfg := config.DefaultSQLiteConfig()
	cfg.Mode = "temp"
	return newSQLite(cfg)
}

// NewWithOptions starts a SQLite database with custom configuration.
//
// Example:
//
//	db, cleanup, err := sqlite.NewWithOptions(
//	    sqlite.WithMode("memory"),
//	    sqlite.WithCache("shared"),
//	)
//	if err != nil {
//	    panic(err)
//	}
//	defer cleanup()
func NewWithOptions(opts ...Option) (*sql.DB, func(), error) {
	cfg := config.DefaultSQLiteConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	return newSQLite(cfg)
}

// newSQLite creates a new SQLite database based on config.
func newSQLite(cfg *config.SQLiteConfig) (*sql.DB, func(), error) {
	var dsn string
	var cleanupFunc func()

	switch cfg.Mode {
	case "memory":
		dsn = ":memory:"
		if cfg.Cache != "" {
			dsn += fmt.Sprintf("?cache=%s", cfg.Cache)
		}
		cleanupFunc = func() {}
	case "temp":
		tmpDir := os.TempDir()
		tmpFile := filepath.Join(tmpDir, fmt.Sprintf("testdb-%d.sqlite", os.Getpid()))
		dsn = tmpFile
		if cfg.Cache != "" {
			dsn += fmt.Sprintf("?cache=%s", cfg.Cache)
		}
		cleanupFunc = func() {
			os.Remove(tmpFile)
		}
	case "file":
		if cfg.Path == "" {
			return nil, nil, fmt.Errorf("file mode requires a path")
		}
		dsn = cfg.Path
		if cfg.Cache != "" {
			dsn += fmt.Sprintf("?cache=%s", cfg.Cache)
		}
		cleanupFunc = func() {}
	default:
		return nil, nil, fmt.Errorf("unknown SQLite mode: %s (use 'memory', 'temp', or 'file')", cfg.Mode)
	}

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open sqlite: %w", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to ping sqlite: %w", err)
	}

	cleanup := func() {
		db.Close()
		cleanupFunc()
	}

	return db, cleanup, nil
}

