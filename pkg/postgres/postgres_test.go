// Package postgres provides PostgreSQL container management for testing.
package postgres

import (
	"context"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// TestNewPostgresBasic tests basic PostgreSQL container creation.
func TestNewPostgresBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping PostgreSQL container test in short mode")
	}
	ctx := context.Background()
	db, cleanup, err := New(ctx)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer cleanup()

	if err := db.Ping(); err != nil {
		t.Fatalf("db.Ping() error = %v", err)
	}

	// Test basic query
	var result int
	err = db.QueryRow("SELECT 42").Scan(&result)
	if err != nil {
		t.Fatalf("QueryRow() error = %v", err)
	}
	if result != 42 {
		t.Fatalf("expected 42, got %d", result)
	}
}

// TestNewPostgresWithOptions tests PostgreSQL container with functional options.
func TestNewPostgresWithOptions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping PostgreSQL container test in short mode")
	}
	ctx := context.Background()
	db, cleanup, err := NewWithOptions(ctx,
		WithImage("postgres:16-alpine"),
		WithUsername("testuser"),
		WithPassword("testpass"),
		WithDatabase("testdb"),
		WithHealthCheckTimeout(30*time.Second),
	)
	if err != nil {
		t.Fatalf("NewWithOptions() error = %v", err)
	}
	defer cleanup()

	if err := db.Ping(); err != nil {
		t.Fatalf("db.Ping() error = %v", err)
	}
}

// TestNewPostgresWithPoolSettings tests connection pool configuration.
func TestNewPostgresWithPoolSettings(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping PostgreSQL container test in short mode")
	}
	ctx := context.Background()
	db, cleanup, err := NewWithOptions(ctx,
		WithPoolSettings(5, 2, 30*time.Minute),
	)
	if err != nil {
		t.Fatalf("NewWithOptions() error = %v", err)
	}
	defer cleanup()

	if err := db.Ping(); err != nil {
		t.Fatalf("db.Ping() error = %v", err)
	}

	// Verify pool settings
	stats := db.Stats()
	if stats.MaxOpenConnections != 5 {
		t.Errorf("expected MaxOpenConnections=5, got %d", stats.MaxOpenConnections)
	}
	if stats.Idle != 2 {
		t.Errorf("expected Idle=2, got %d", stats.Idle)
	}
}

// TestMustNewPostgres tests MustNew panic behavior with invalid image.
func TestMustNewPostgres(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping PostgreSQL container test in short mode")
	}
	ctx := context.Background()

	// This should not panic (valid image)
	db, cleanup := MustNew(ctx)
	if db == nil {
		t.Fatal("expected non-nil db")
	}
	defer cleanup()

	// MustNewWithOptions should also not panic with valid options
	db2, cleanup2 := MustNewWithOptions(ctx, WithImage("postgres:16-alpine"))
	if db2 == nil {
		t.Fatal("expected non-nil db2")
	}
	defer cleanup2()
}

// TestPostgresConnection tests that we can execute SQL statements.
func TestPostgresConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping PostgreSQL container test in short mode")
	}
	ctx := context.Background()
	db, cleanup, err := New(ctx)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer cleanup()

	// Create a test table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS test_table (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("CREATE TABLE error = %v", err)
	}

	// Insert a row
	result, err := db.Exec("INSERT INTO test_table (name) VALUES ($1)", "test")
	if err != nil {
		t.Fatalf("INSERT error = %v", err)
	}

	id, _ := result.LastInsertId()
	if id == 0 {
		t.Error("expected non-zero LastInsertId")
	}

	// Query the row
	var name string
	err = db.QueryRow("SELECT name FROM test_table WHERE id = $1", id).Scan(&name)
	if err != nil {
		t.Fatalf("QueryRow error = %v", err)
	}
	if name != "test" {
		t.Errorf("expected 'test', got '%s'", name)
	}
}

// TestPostgresMultipleConns tests multiple concurrent connections.
func TestPostgresMultipleConns(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping PostgreSQL container test in short mode")
	}
	ctx := context.Background()
	db, cleanup, err := New(ctx)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer cleanup()

	// Execute multiple queries concurrently
	results := make(chan int, 5)
	for i := 0; i < 5; i++ {
		go func() {
			var val int
			err := db.QueryRow("SELECT 1").Scan(&val)
			if err != nil {
				t.Errorf("query error: %v", err)
				results <- 0
				return
			}
			results <- val
		}()
	}

	for i := 0; i < 5; i++ {
		val := <-results
		if val != 1 {
			t.Errorf("expected 1, got %d", val)
		}
	}
}
