// Package config provides configuration types for containerized databases.
package config

import (
	"time"
)

// MySQLConfig holds configuration for MySQL containers.
type MySQLConfig struct {
	// Image is the Docker image to use. Defaults to mysql:8.0
	Image string
	// Username for MySQL. Defaults to root
	Username string
	// Password for MySQL. Defaults to rootpassword
	Password string
	// Database name. Defaults to testdb
	Database string
	// HealthCheckTimeout for container readiness. Defaults to 30s
	HealthCheckTimeout time.Duration
	// HealthCheckRetries for connection attempts. Defaults to 3
	HealthCheckRetries int
	// HealthCheckInterval between retries. Defaults to 500ms
	HealthCheckInterval time.Duration
	// MaxOpenConns is the max number of open connections. Defaults to 10
	MaxOpenConns int
	// MaxIdleConns is the max number of idle connections. Defaults to 5
	MaxIdleConns int
	// ConnMaxLifetime is the max lifetime of a connection. Defaults to 1 hour
	ConnMaxLifetime time.Duration
}

// DefaultMySQLConfig returns a config with sensible defaults.
func DefaultMySQLConfig() *MySQLConfig {
	return &MySQLConfig{
		Image:                "mysql:8.0",
		Username:             "root",
		Password:             "rootpassword",
		Database:             "testdb",
		HealthCheckTimeout:   30 * time.Second,
		HealthCheckRetries:   3,
		HealthCheckInterval:  500 * time.Millisecond,
		MaxOpenConns:         10,
		MaxIdleConns:         5,
		ConnMaxLifetime:      time.Hour,
	}
}

// PostgresConfig holds configuration for PostgreSQL containers.
type PostgresConfig struct {
	// Image is the Docker image to use. Defaults to postgres:16-alpine
	Image string
	// Username. Defaults to postgres
	Username string
	// Password. Defaults to postgres
	Password string
	// Database name. Defaults to testdb
	Database string
	// HealthCheckTimeout. Defaults to 30s
	HealthCheckTimeout time.Duration
	// HealthCheckRetries for connection attempts. Defaults to 3
	HealthCheckRetries int
	// HealthCheckInterval between retries. Defaults to 500ms
	HealthCheckInterval time.Duration
	// MaxOpenConns is the max number of open connections. Defaults to 10
	MaxOpenConns int
	// MaxIdleConns is the max number of idle connections. Defaults to 5
	MaxIdleConns int
	// ConnMaxLifetime is the max lifetime of a connection. Defaults to 1 hour
	ConnMaxLifetime time.Duration
}

// DefaultPostgresConfig returns a config with sensible defaults.
func DefaultPostgresConfig() *PostgresConfig {
	return &PostgresConfig{
		Image:                "postgres:16-alpine",
		Username:             "postgres",
		Password:             "postgres",
		Database:             "testdb",
		HealthCheckTimeout:   30 * time.Second,
		HealthCheckRetries:   3,
		HealthCheckInterval:  500 * time.Millisecond,
		MaxOpenConns:         10,
		MaxIdleConns:         5,
		ConnMaxLifetime:      time.Hour,
	}
}

// SQLiteConfig holds configuration for SQLite databases.
type SQLiteConfig struct {
	// Mode determines the database mode: "memory", "temp", or "file"
	Mode string
	// Path is the file path (used when Mode is "file")
	Path string
	// Cache controls the cache mode: "shared", "private", "write"
	Cache string
}

// DefaultSQLiteConfig returns a config with sensible defaults (in-memory).
func DefaultSQLiteConfig() *SQLiteConfig {
	return &SQLiteConfig{
		Mode:  "memory",
		Path:  "",
		Cache: "shared",
	}
}
