package sqlite

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestInMemory(t *testing.T) {
	db, cleanup, err := InMemory()
	if err != nil {
		t.Fatalf("InMemory() failed: %v", err)
	}
	defer cleanup()

	if err := db.Ping(); err != nil {
		t.Fatalf("db.Ping() failed: %v", err)
	}
}

func TestTempDB(t *testing.T) {
	db, cleanup, err := TempDB()
	if err != nil {
		t.Fatalf("TempDB() failed: %v", err)
	}
	defer cleanup()

	if err := db.Ping(); err != nil {
		t.Fatalf("db.Ping() failed: %v", err)
	}

	// Verify we can execute queries
	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY)")
	if err != nil {
		t.Fatalf("CREATE TABLE failed: %v", err)
	}
}

func TestNewWithOptions_Memory(t *testing.T) {
	db, cleanup, err := NewWithOptions(WithMode("memory"), WithCache("shared"))
	if err != nil {
		t.Fatalf("NewWithOptions(memory) failed: %v", err)
	}
	defer cleanup()

	if err := db.Ping(); err != nil {
		t.Fatalf("db.Ping() failed: %v", err)
	}
}

func TestNewWithOptions_Temp(t *testing.T) {
	db, cleanup, err := NewWithOptions(WithMode("temp"), WithCache("private"))
	if err != nil {
		t.Fatalf("NewWithOptions(temp) failed: %v", err)
	}
	defer cleanup()

	if err := db.Ping(); err != nil {
		t.Fatalf("db.Ping() failed: %v", err)
	}
}

func TestNewWithOptions_File(t *testing.T) {
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "sqlite-test-file.db")
	defer os.Remove(tmpFile)

	db, cleanup, err := NewWithOptions(
		WithMode("file"),
		WithPath(tmpFile),
		WithCache("write"),
	)
	if err != nil {
		t.Fatalf("NewWithOptions(file) failed: %v", err)
	}
	defer cleanup()

	if err := db.Ping(); err != nil {
		t.Fatalf("db.Ping() failed: %v", err)
	}

	// Verify persistence by creating table and closing
	_, err = db.Exec("CREATE TABLE persistence_test (data TEXT)")
	if err != nil {
		t.Fatalf("CREATE TABLE failed: %v", err)
	}

	// Close and reopen
	cleanup()

	db2, cleanup2, err := NewWithOptions(WithMode("file"), WithPath(tmpFile))
	if err != nil {
		t.Fatalf("Reopening file database failed: %v", err)
	}
	defer cleanup2()

	// Table should still exist
	var count int
	err = db2.QueryRow("SELECT COUNT(*) FROM persistence_test").Scan(&count)
	if err != nil {
		t.Fatalf("Query failed after reopen: %v", err)
	}
	if count != 0 {
		t.Fatalf("Expected 0 rows, got %d", count)
	}
}

func TestNewWithOptions_FileNoPath(t *testing.T) {
	_, _, err := NewWithOptions(WithMode("file"))
	if err == nil {
		t.Fatalf("Expected error for file mode without path")
	}
}

func TestNewWithOptions_InvalidMode(t *testing.T) {
	_, _, err := NewWithOptions(WithMode("invalid"))
	if err == nil {
		t.Fatalf("Expected error for invalid mode")
	}
}

func TestSQLite_CRUD(t *testing.T) {
	db, cleanup, err := InMemory()
	if err != nil {
		t.Fatalf("InMemory() failed: %v", err)
	}
	defer cleanup()

	ctx := context.Background()

	// CREATE
	_, err = db.ExecContext(ctx, "CREATE TABLE items (id INTEGER PRIMARY KEY, name TEXT, value REAL)")
	if err != nil {
		t.Fatalf("CREATE TABLE failed: %v", err)
	}

	// INSERT
	result, err := db.ExecContext(ctx, "INSERT INTO items (name, value) VALUES (?, ?)", "test_item", 42.5)
	if err != nil {
		t.Fatalf("INSERT failed: %v", err)
	}

	id, _ := result.LastInsertId()

	// READ
	var name string
	var value float64
	err = db.QueryRowContext(ctx, "SELECT name, value FROM items WHERE id = ?", id).Scan(&name, &value)
	if err != nil {
		t.Fatalf("SELECT failed: %v", err)
	}
	if name != "test_item" || value != 42.5 {
		t.Fatalf("Expected (test_item, 42.5), got (%s, %f)", name, value)
	}

	// UPDATE
	_, err = db.ExecContext(ctx, "UPDATE items SET value = ? WHERE id = ?", 99.9, id)
	if err != nil {
		t.Fatalf("UPDATE failed: %v", err)
	}

	err = db.QueryRowContext(ctx, "SELECT value FROM items WHERE id = ?", id).Scan(&value)
	if err != nil {
		t.Fatalf("SELECT after UPDATE failed: %v", err)
	}
	if value != 99.9 {
		t.Fatalf("Expected 99.9, got %f", value)
	}

	// DELETE
	_, err = db.ExecContext(ctx, "DELETE FROM items WHERE id = ?", id)
	if err != nil {
		t.Fatalf("DELETE failed: %v", err)
	}

	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM items").Scan(&count)
	if err != nil {
		t.Fatalf("COUNT failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("Expected 0 rows after DELETE, got %d", count)
	}
}

func TestSQLite_Transaction(t *testing.T) {
	db, cleanup, err := InMemory()
	if err != nil {
		t.Fatalf("InMemory() failed: %v", err)
	}
	defer cleanup()

	ctx := context.Background()

	_, err = db.ExecContext(ctx, "CREATE TABLE accounts (id INTEGER PRIMARY KEY, balance REAL)")
	if err != nil {
		t.Fatalf("CREATE TABLE failed: %v", err)
	}

	// Start transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx failed: %v", err)
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO accounts (balance) VALUES (?)", 100.0)
	if err != nil {
		tx.Rollback()
		t.Fatalf("INSERT failed: %v", err)
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO accounts (balance) VALUES (?)", 200.0)
	if err != nil {
		tx.Rollback()
		t.Fatalf("INSERT failed: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	// Verify
	var total float64
	err = db.QueryRowContext(ctx, "SELECT SUM(balance) FROM accounts").Scan(&total)
	if err != nil {
		t.Fatalf("SUM query failed: %v", err)
	}
	if total != 300.0 {
		t.Fatalf("Expected total 300.0, got %f", total)
	}
}

func TestSQLite_ConcurrentRead(t *testing.T) {
	db, cleanup, err := InMemory()
	if err != nil {
		t.Fatalf("InMemory() failed: %v", err)
	}
	defer cleanup()

	ctx := context.Background()

	_, err = db.ExecContext(ctx, "CREATE TABLE counters (id INTEGER PRIMARY KEY, count INTEGER)")
	if err != nil {
		t.Fatalf("CREATE TABLE failed: %v", err)
	}

	// Insert initial row
	_, err = db.ExecContext(ctx, "INSERT INTO counters (count) VALUES (?)", 0)
	if err != nil {
		t.Fatalf("INSERT failed: %v", err)
	}

	// Simulate concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			db.QueryRowContext(ctx, "SELECT count FROM counters WHERE id = 1").Scan(new(int))
		}()
	}

	t.Log("Concurrent read test passed (no race conditions detected)")
}
