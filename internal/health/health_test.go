package health

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

// mockDB is a simple mock for testing
type mockDB struct {
	pingErr error
	pingFn  func() error
}

func (m *mockDB) PingContext(ctx context.Context) error {
	if m.pingFn != nil {
		return m.pingFn()
	}
	return m.pingErr
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Timeout != 30*time.Second {
		t.Errorf("expected Timeout 30s, got %v", cfg.Timeout)
	}
	if cfg.Interval != 500*time.Millisecond {
		t.Errorf("expected Interval 500ms, got %v", cfg.Interval)
	}
	if cfg.Retries != 5 {
		t.Errorf("expected Retries 5, got %d", cfg.Retries)
	}
}

func TestWithTimeout(t *testing.T) {
	cfg := NewConfig(WithTimeout(10 * time.Second))
	if cfg.Timeout != 10*time.Second {
		t.Errorf("expected Timeout 10s, got %v", cfg.Timeout)
	}
}

func TestWithInterval(t *testing.T) {
	cfg := NewConfig(WithInterval(100 * time.Millisecond))
	if cfg.Interval != 100*time.Millisecond {
		t.Errorf("expected Interval 100ms, got %v", cfg.Interval)
	}
}

func TestWithRetries(t *testing.T) {
	cfg := NewConfig(WithRetries(10))
	if cfg.Retries != 10 {
		t.Errorf("expected Retries 10, got %d", cfg.Retries)
	}
}

func TestCheck_Success(t *testing.T) {
	// This test uses a real sql.DB but with minimal configuration
	// In practice we'd use a mock, but for unit testing the interface:
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Skipf("skipping: could not open sqlite: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	result, err := Check(ctx, db, WithRetries(1), WithTimeout(time.Second))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !result.Healthy {
		t.Errorf("expected healthy=true, got false: %s", result.Message)
	}
}

func TestCheck_Failure(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:notexist")
	if err != nil {
		t.Skipf("skipping: could not open sqlite: %v", err)
	}
	// Force an error by closing immediately
	db.Close()

	ctx := context.Background()
	result, err := Check(ctx, db, WithRetries(2), WithInterval(10*time.Millisecond))
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	if result.Healthy {
		t.Errorf("expected healthy=false")
	}
}

func TestCheck_Timeout(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Skipf("skipping: could not open sqlite: %v", err)
	}
	db.Close()

	ctx := context.Background()
	result, err := Check(ctx, db,
		WithTimeout(10*time.Millisecond),
		WithInterval(5*time.Millisecond),
		WithRetries(2))
	if err == nil {
		t.Errorf("expected error due to timeout")
	}
	if result.Healthy {
		t.Errorf("expected healthy=false on timeout")
	}
}

func TestCheckResult(t *testing.T) {
	result := &CheckResult{
		Healthy: true,
		Message: "ok",
		Latency: 5 * time.Millisecond,
	}
	if !result.Healthy {
		t.Error("expected Healthy=true")
	}
	if result.Message != "ok" {
		t.Errorf("expected Message='ok', got '%s'", result.Message)
	}
	if result.Latency != 5*time.Millisecond {
		t.Errorf("expected Latency=5ms, got %v", result.Latency)
	}
}

func TestWaitForReady_Success(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Skipf("skipping: could not open sqlite: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	err = WaitForReady(ctx, db, WithRetries(3), WithTimeout(time.Second))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestWaitForReady_Failure(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Skipf("skipping: could not open sqlite: %v", err)
	}
	db.Close()

	ctx := context.Background()
	err = WaitForReady(ctx, db,
		WithRetries(1),
		WithInterval(10*time.Millisecond),
		WithTimeout(50*time.Millisecond))
	if err == nil {
		t.Errorf("expected error")
	}
}

var _ interface {
	PingContext(ctx context.Context) error
} = (*sql.DB)(nil)
