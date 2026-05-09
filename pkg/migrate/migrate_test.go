package migrate

import (
	"testing"
)

func TestNewRunner(t *testing.T) {
	// Test that NewRunner creates a Runner with correct defaults
	// without a real DB connection
	r := NewRunner(nil)
	if r == nil {
		t.Fatal("expected non-nil Runner")
	}
	if r.dir != "migrations" {
		t.Errorf("expected default dir 'migrations', got %s", r.dir)
	}
}

func TestNewRunnerWithDir(t *testing.T) {
	r := NewRunner(nil, WithDir("/custom/path"))
	if r.dir != "/custom/path" {
		t.Errorf("expected dir '/custom/path', got %s", r.dir)
	}
}

func TestNewRunnerWithTableName(t *testing.T) {
	r := NewRunner(nil, WithTableName("custom_migrations"))
	if r == nil {
		t.Fatal("expected non-nil Runner")
	}
	// We can't directly check opts, but we can verify it doesn't panic
}

func TestNewRunnerWithTimeout(t *testing.T) {
	r := NewRunner(nil, WithTimeout(60))
	if r == nil {
		t.Fatal("expected non-nil Runner")
	}
}

func TestNewRunnerWithMultipleOptions(t *testing.T) {
	r := NewRunner(nil,
		WithDir("/test/migrations"),
		WithTableName("test_migrations"),
		WithTimeout(120),
	)
	if r.dir != "/test/migrations" {
		t.Errorf("expected dir '/test/migrations', got %s", r.dir)
	}
}
