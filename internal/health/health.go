// Package health provides health checking utilities for database containers.
package health

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Config holds health check configuration.
type Config struct {
	// Timeout is the max time to wait for health check. Defaults to 30s
	Timeout time.Duration
	// Interval between retry attempts. Defaults to 500ms
	Interval time.Duration
	// Retries is the number of retry attempts. Defaults to 5
	Retries int
}

// DefaultConfig returns a config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Timeout:  30 * time.Second,
		Interval: 500 * time.Millisecond,
		Retries:  5,
	}
}

// WithTimeout sets the health check timeout.
func WithTimeout(timeout time.Duration) func(*Config) {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithInterval sets the retry interval.
func WithInterval(interval time.Duration) func(*Config) {
	return func(c *Config) {
		c.Interval = interval
	}
}

// WithRetries sets the number of retry attempts.
func WithRetries(retries int) func(*Config) {
	return func(c *Config) {
		c.Retries = retries
	}
}

// Option is a functional option for health check configuration.
type Option func(*Config)

// NewConfig creates a new health check config with options.
func NewConfig(opts ...Option) *Config {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

// CheckResult holds the result of a health check.
type CheckResult struct {
	Healthy bool
	Message string
	Latency time.Duration
}

// Check verifies the database is reachable with retry logic.
func Check(ctx context.Context, db *sql.DB, opts ...Option) (*CheckResult, error) {
	cfg := NewConfig(opts...)

	ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	var lastErr error
	for i := 0; i < cfg.Retries; i++ {
		start := time.Now()
		if err := db.PingContext(ctx); err != nil {
			lastErr = err
			select {
			case <-ctx.Done():
				return &CheckResult{
					Healthy: false,
					Message: fmt.Sprintf("health check timed out after %v", cfg.Timeout),
				}, ctx.Err()
			case <-time.After(cfg.Interval):
				continue
			}
		}
		return &CheckResult{
			Healthy: true,
			Message: "database is healthy",
			Latency: time.Since(start),
		}, nil
	}

	return &CheckResult{
		Healthy: false,
		Message: fmt.Sprintf("health check failed after %d attempts: %v", cfg.Retries, lastErr),
	}, lastErr
}

// WaitForReady blocks until the database is ready or timeout is reached.
func WaitForReady(ctx context.Context, db *sql.DB, opts ...Option) error {
	result, err := Check(ctx, db, opts...)
	if err != nil || !result.Healthy {
		return fmt.Errorf("database not ready: %s", result.Message)
	}
	return nil
}
