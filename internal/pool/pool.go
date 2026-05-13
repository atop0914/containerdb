// Package pool provides enhanced connection pool management utilities.
package pool

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"
)

// Stats holds connection pool statistics.
type Stats struct {
	MaxOpenConnections int
	OpenConnections     int
	InUse               int
	Idle                int
	WaitCount           int64
	WaitDuration        time.Duration
	MaxIdleClosed       int64
	MaxLifetimeClosed   int64
}

// GetStats retrieves current pool statistics.
func GetStats(db *sql.DB) *Stats {
	stats := db.Stats()
	return &Stats{
		MaxOpenConnections: stats.MaxOpenConnections,
		OpenConnections:     stats.OpenConnections,
		InUse:               stats.InUse,
		Idle:                stats.Idle,
		WaitCount:           stats.WaitCount,
		WaitDuration:        stats.WaitDuration,
		MaxIdleClosed:       stats.MaxIdleClosed,
		MaxLifetimeClosed:   stats.MaxLifetimeClosed,
	}
}

// PoolConfig holds advanced pool configuration.
type PoolConfig struct {
	// MaxOpenConns is the maximum number of open connections to the database.
	MaxOpenConns int
	// MaxIdleConns is the maximum number of connections in the idle connection pool.
	MaxIdleConns int
	// ConnMaxLifetime is the maximum amount of time a connection may be reused.
	ConnMaxLifetime time.Duration
	// ConnMaxIdleTime is the maximum amount of time a connection may be idle.
	ConnMaxIdleTime time.Duration
	// ConnReqDurThreshold is the duration threshold for slow query logging.
	ConnReqDurThreshold time.Duration
}

// DefaultPoolConfig returns sensible defaults.
func DefaultPoolConfig() *PoolConfig {
	return &PoolConfig{
		MaxOpenConns:        10,
		MaxIdleConns:        5,
		ConnMaxLifetime:     time.Hour,
		ConnMaxIdleTime:     30 * time.Minute,
		ConnReqDurThreshold: 100 * time.Millisecond,
	}
}

// Validate checks if the pool configuration is valid.
func (p *PoolConfig) Validate() error {
	if p.MaxOpenConns <= 0 {
		return fmt.Errorf("MaxOpenConns must be positive, got %d", p.MaxOpenConns)
	}
	if p.MaxIdleConns < 0 {
		return fmt.Errorf("MaxIdleConns cannot be negative, got %d", p.MaxIdleConns)
	}
	if p.MaxIdleConns > p.MaxOpenConns {
		return fmt.Errorf("MaxIdleConns (%d) cannot exceed MaxOpenConns (%d)", p.MaxIdleConns, p.MaxOpenConns)
	}
	if p.ConnMaxLifetime <= 0 {
		return fmt.Errorf("ConnMaxLifetime must be positive, got %v", p.ConnMaxLifetime)
	}
	return nil
}

// Configure applies pool configuration to a sql.DB.
func Configure(db *sql.DB, cfg *PoolConfig) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	return nil
}

// Monitor provides continuous pool monitoring.
type Monitor struct {
	db       *sql.DB
	interval time.Duration
	stopCh   chan struct{}
	wg       sync.WaitGroup
	onSlow   func(d time.Duration, query string)
	onWait   func(d time.Duration)
}

// NewMonitor creates a new pool monitor.
func NewMonitor(db *sql.DB, interval time.Duration) *Monitor {
	return &Monitor{
		db:       db,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// OnSlowQuery registers a callback for slow queries.
func (m *Monitor) OnSlowQuery(fn func(d time.Duration, query string)) {
	m.onSlow = fn
}

// OnWait registers a callback for connection wait events.
func (m *Monitor) OnWait(fn func(d time.Duration)) {
	m.onWait = fn
}

// Start begins monitoring the pool.
func (m *Monitor) Start(ctx context.Context) {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		ticker := time.NewTicker(m.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-m.stopCh:
				return
			case <-ticker.C:
				stats := GetStats(m.db)
				if stats.WaitCount > 0 && m.onWait != nil {
					m.onWait(stats.WaitDuration / time.Duration(stats.WaitCount))
				}
			}
		}
	}()
}

// Stop halts the monitor.
func (m *Monitor) Stop() {
	close(m.stopCh)
	m.wg.Wait()
}

// TracedDB wraps a sql.DB with query timing capabilities.
type TracedDB struct {
	DB      *sql.DB
	SlowLog func(query string, duration time.Duration)
	mu      sync.Mutex
}

// NewTracedDB creates a traced database wrapper.
func NewTracedDB(db *sql.DB) *TracedDB {
	return &TracedDB{DB: db}
}

// SetSlowQueryThreshold sets the threshold for slow query logging.
func (t *TracedDB) SetSlowQueryThreshold(threshold time.Duration) {
	t.SlowLog = func(query string, d time.Duration) {
		fmt.Printf("slow query (%s): %s\n", d, query)
	}
}

// QueryContext wraps QueryContext with timing.
func (t *TracedDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := t.DB.QueryContext(ctx, query, args...)
	d := time.Since(start)
	if d > 0 && t.SlowLog != nil {
		t.SlowLog(query, d)
	}
	return rows, err
}

// ExecContext wraps ExecContext with timing.
func (t *TracedDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := t.DB.ExecContext(ctx, query, args...)
	d := time.Since(start)
	if d > 0 && t.SlowLog != nil {
		t.SlowLog(query, d)
	}
	return result, err
}

// QueryRowContext wraps QueryRowContext with timing.
func (t *TracedDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := t.DB.QueryRowContext(ctx, query, args...)
	d := time.Since(start)
	if d > 0 && t.SlowLog != nil {
		t.SlowLog(query, d)
	}
	return row
}
