package examples

import (
	"context"
	"fmt"
	"testing"

	"github.com/atop0914/containerdb-bootcamp/pkg/sqlite"
)

func TestSQLite_InMemory_Example(t *testing.T) {
	db, cleanup, err := sqlite.InMemory()
	if err != nil {
		t.Fatalf("failed to create in-memory sqlite: %v", err)
	}
	defer cleanup()

	ctx := context.Background()

	// Create a test table
	_, err = db.ExecContext(ctx, "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	// Insert data
	result, err := db.ExecContext(ctx, "INSERT INTO users (name) VALUES (?)", "alice")
	if err != nil {
		t.Fatalf("failed to insert: %v", err)
	}

	id, _ := result.LastInsertId()
	fmt.Printf("Inserted user with id: %d\n", id)

	// Query data
	var name string
	err = db.QueryRowContext(ctx, "SELECT name FROM users WHERE id = ?", id).Scan(&name)
	if err != nil {
		t.Fatalf("failed to query: %v", err)
	}

	fmt.Printf("Retrieved user: %s\n", name)
	t.Log("SQLite in-memory test passed")
}

func TestSQLite_TempDB_Example(t *testing.T) {
	db, cleanup, err := sqlite.TempDB()
	if err != nil {
		t.Fatalf("failed to create temp sqlite: %v", err)
	}
	defer cleanup()

	ctx := context.Background()

	// Create a test table
	_, err = db.ExecContext(ctx, "CREATE TABLE products (id INTEGER PRIMARY KEY, name TEXT, price REAL)")
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	// Insert data
	_, err = db.ExecContext(ctx, "INSERT INTO products (name, price) VALUES (?, ?)", "widget", 29.99)
	if err != nil {
		t.Fatalf("failed to insert: %v", err)
	}

	// Query data
	var name string
	var price float64
	err = db.QueryRowContext(ctx, "SELECT name, price FROM products").Scan(&name, &price)
	if err != nil {
		t.Fatalf("failed to query: %v", err)
	}

	fmt.Printf("Product: %s, Price: %.2f\n", name, price)
	t.Log("SQLite tempdb test passed")
}

func TestSQLite_WithOptions_Example(t *testing.T) {
	db, cleanup, err := sqlite.NewWithOptions(
		sqlite.WithMode("memory"),
		sqlite.WithCache("shared"),
	)
	if err != nil {
		t.Fatalf("failed to create sqlite with options: %v", err)
	}
	defer cleanup()

	ctx := context.Background()

	// Verify we can use the database
	_, err = db.ExecContext(ctx, "SELECT 1")
	if err != nil {
		t.Fatalf("failed to execute query: %v", err)
	}

	t.Log("SQLite with options test passed")
}
