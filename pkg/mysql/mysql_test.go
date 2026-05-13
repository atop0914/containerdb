// Package mysql provides MySQL container management for testing.
package mysql

import (
	"context"
	"testing"
	"time"
)

// TestNewMySQLBasic tests basic MySQL container creation.
func TestNewMySQLBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping MySQL container test in short mode")
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
	err = db.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		t.Fatalf("QueryRow() error = %v", err)
	}
	if result != 1 {
		t.Fatalf("expected 1, got %d", result)
	}
}

// TestNewMySQLWithOptions tests MySQL container with functional options.
func TestNewMySQLWithOptions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping MySQL container test in short mode")
	}
	ctx := context.Background()
	db, cleanup, err := NewWithOptions(ctx,
		WithImage("mysql:8.0"),
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

	// Verify we can query the database
	var version string
	err = db.QueryRow("SELECT VERSION()").Scan(&version)
	if err != nil {
		t.Fatalf("SELECT VERSION() error = %v", err)
	}
	t.Logf("MySQL version: %s", version)
}

// TestNewMySQLWithPoolSettings tests connection pool configuration.
func TestNewMySQLWithPoolSettings(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping MySQL container test in short mode")
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
	if stats.Idle > 5 {
		t.Errorf("expected Idle<=5, got %d", stats.Idle)
	}
}

// TestMustNewMySQL tests MustNew panic behavior.
func TestMustNewMySQL(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping MySQL container test in short mode")
	}
	ctx := context.Background()

	// This should not panic (valid image)
	db, cleanup := MustNew(ctx)
	if db == nil {
		t.Fatal("expected non-nil db")
	}
	defer cleanup()

	if err := db.Ping(); err != nil {
		t.Fatalf("db.Ping() error = %v", err)
	}
}

// TestMustNewMySQLWithOptions tests MustNewWithOptions.
func TestMustNewMySQLWithOptions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping MySQL container test in short mode")
	}
	ctx := context.Background()
	db, cleanup := MustNewWithOptions(ctx, WithImage("mysql:8.0"))
	if db == nil {
		t.Fatal("expected non-nil db")
	}
	defer cleanup()

	if err := db.Ping(); err != nil {
		t.Fatalf("db.Ping() error = %v", err)
	}
}

// TestMySQLConnection tests that we can execute SQL statements.
func TestMySQLConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping MySQL container test in short mode")
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
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(100) NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("CREATE TABLE error = %v", err)
	}

	// Insert a row
	result, err := db.Exec("INSERT INTO test_table (name) VALUES (?)", "test")
	if err != nil {
		t.Fatalf("INSERT error = %v", err)
	}

	id, _ := result.LastInsertId()
	if id == 0 {
		t.Error("expected non-zero LastInsertId")
	}

	// Query the row
	var name string
	err = db.QueryRow("SELECT name FROM test_table WHERE id = ?", id).Scan(&name)
	if err != nil {
		t.Fatalf("QueryRow error = %v", err)
	}
	if name != "test" {
		t.Errorf("expected 'test', got '%s'", name)
	}
}

// TestMySQLMultipleConns tests multiple concurrent connections.
func TestMySQLMultipleConns(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping MySQL container test in short mode")
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

// TestMySQLTransaction tests transaction support.
func TestMySQLTransaction(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping MySQL container test in short mode")
	}
	ctx := context.Background()
	db, cleanup, err := New(ctx)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer cleanup()

	// Create test table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS accounts (
			id INT AUTO_INCREMENT PRIMARY KEY,
			balance DECIMAL(10,2)
		)
	`)
	if err != nil {
		t.Fatalf("CREATE TABLE error = %v", err)
	}

	// Start transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx failed: %v", err)
	}

	_, err = tx.Exec("INSERT INTO accounts (balance) VALUES (?)", 100.0)
	if err != nil {
		tx.Rollback()
		t.Fatalf("INSERT failed: %v", err)
	}

	_, err = tx.Exec("INSERT INTO accounts (balance) VALUES (?)", 200.0)
	if err != nil {
		tx.Rollback()
		t.Fatalf("INSERT failed: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	// Verify
	var total float64
	err = db.QueryRow("SELECT SUM(balance) FROM accounts").Scan(&total)
	if err != nil {
		t.Fatalf("SUM query failed: %v", err)
	}
	if total != 300.0 {
		t.Errorf("expected total 300.0, got %f", total)
	}
}

// TestMySQLPreparedStatements tests prepared statement usage.
func TestMySQLPreparedStatements(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping MySQL container test in short mode")
	}
	ctx := context.Background()
	db, cleanup, err := New(ctx)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer cleanup()

	// Create test table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(100)
		)
	`)
	if err != nil {
		t.Fatalf("CREATE TABLE error = %v", err)
	}

	// Prepare statement
	stmt, err := db.PrepareContext(ctx, "INSERT INTO users (name) VALUES (?)")
	if err != nil {
		t.Fatalf("PrepareContext error = %v", err)
	}
	defer stmt.Close()

	// Execute multiple inserts
	for _, name := range []string{"alice", "bob", "charlie"} {
		_, err = stmt.ExecContext(ctx, name)
		if err != nil {
			t.Fatalf("ExecContext error = %v", err)
		}
	}

	// Verify count
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		t.Fatalf("COUNT query failed: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 rows, got %d", count)
	}
}
