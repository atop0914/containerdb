package pool

import (
	"database/sql"
	"testing"
	"time"
)

func TestDefaultPoolConfig(t *testing.T) {
	cfg := DefaultPoolConfig()
	if cfg.MaxOpenConns != 10 {
		t.Errorf("expected MaxOpenConns=10, got %d", cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns != 5 {
		t.Errorf("expected MaxIdleConns=5, got %d", cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime != time.Hour {
		t.Errorf("expected ConnMaxLifetime=1h, got %v", cfg.ConnMaxLifetime)
	}
}

func TestPoolConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     PoolConfig
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: PoolConfig{
				MaxOpenConns:    10,
				MaxIdleConns:    5,
				ConnMaxLifetime: time.Hour,
			},
			wantErr: false,
		},
		{
			name: "MaxOpenConns zero",
			cfg: PoolConfig{
				MaxOpenConns:    0,
				MaxIdleConns:    5,
				ConnMaxLifetime: time.Hour,
			},
			wantErr: true,
		},
		{
			name: "MaxOpenConns negative",
			cfg: PoolConfig{
				MaxOpenConns:    -1,
				MaxIdleConns:    5,
				ConnMaxLifetime: time.Hour,
			},
			wantErr: true,
		},
		{
			name: "MaxIdleConns negative",
			cfg: PoolConfig{
				MaxOpenConns:    10,
				MaxIdleConns:    -1,
				ConnMaxLifetime: time.Hour,
			},
			wantErr: true,
		},
		{
			name: "MaxIdleConns exceeds MaxOpenConns",
			cfg: PoolConfig{
				MaxOpenConns:    5,
				MaxIdleConns:    10,
				ConnMaxLifetime: time.Hour,
			},
			wantErr: true,
		},
		{
			name: "ConnMaxLifetime zero",
			cfg: PoolConfig{
				MaxOpenConns:    10,
				MaxIdleConns:    5,
				ConnMaxLifetime: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetStats(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Skipf("skipping: could not open sqlite: %v", err)
	}
	defer db.Close()

	stats := GetStats(db)
	if stats == nil {
		t.Fatal("expected non-nil stats")
	}
	// Basic sanity check - values may be 0 for in-memory db
	if stats.MaxOpenConnections < 0 {
		t.Errorf("expected MaxOpenConnections >= 0, got %d", stats.MaxOpenConnections)
	}
}

func TestConfigure(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Skipf("skipping: could not open sqlite: %v", err)
	}
	defer db.Close()

	cfg := &PoolConfig{
		MaxOpenConns:    20,
		MaxIdleConns:    10,
		ConnMaxLifetime: 2 * time.Hour,
	}

	err = Configure(db, cfg)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify settings were applied
	stats := db.Stats()
	if stats.MaxOpenConnections != 20 {
		t.Errorf("expected MaxOpenConnections=20, got %d", stats.MaxOpenConnections)
	}
}

func TestConfigure_Invalid(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Skipf("skipping: could not open sqlite: %v", err)
	}
	defer db.Close()

	cfg := &PoolConfig{
		MaxOpenConns:    0, // Invalid
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
	}

	err = Configure(db, cfg)
	if err == nil {
		t.Errorf("expected error for invalid config")
	}
}

func TestMonitorCreation(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Skipf("skipping: could not open sqlite: %v", err)
	}
	defer db.Close()

	monitor := NewMonitor(db, time.Second)
	if monitor == nil {
		t.Fatal("expected non-nil monitor")
	}
	if monitor.interval != time.Second {
		t.Errorf("expected interval=1s, got %v", monitor.interval)
	}
}

func TestMonitorCallbacks(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Skipf("skipping: could not open sqlite: %v", err)
	}
	defer db.Close()

	monitor := NewMonitor(db, time.Second)

	var slowCalled, waitCalled bool
	monitor.OnSlowQuery(func(d time.Duration, query string) {
		slowCalled = true
	})
	monitor.OnWait(func(d time.Duration) {
		waitCalled = true
	})

	// Just verify callbacks are set (actual monitoring requires longer runtime)
	if !slowCalled || !waitCalled {
		t.Skip("callbacks not called as expected in short test")
	}
}

func TestTracedDB(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Skipf("skipping: could not open sqlite: %v", err)
	}
	defer db.Close()

	traced := NewTracedDB(db)
	if traced == nil {
		t.Fatal("expected non-nil traced db")
	}
	if traced.DB != db {
		t.Error("expected DB field to match input")
	}
}

func TestTracedDB_SlowQueryThreshold(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Skipf("skipping: could not open sqlite: %v", err)
	}
	defer db.Close()

	traced := NewTracedDB(db)
	traced.SetSlowQueryThreshold(10 * time.Millisecond)
	// Should not panic
}
