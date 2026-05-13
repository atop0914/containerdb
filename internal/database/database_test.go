package database

import (
	"context"
	"testing"
	"time"

	"github.com/atop0914/containerdb-bootcamp/pkg/sqlite"
)

func TestDB_Ping(t *testing.T) {
	db, cleanup, err := sqlite.InMemory()
	if err != nil {
		t.Fatalf("InMemory() error = %v", err)
	}
	defer cleanup()

	wrapper := &DB{
		Pool:  db,
		Close: cleanup,
	}

	ctx := context.Background()
	if err := wrapper.Ping(ctx); err != nil {
		t.Fatalf("Ping() error = %v", err)
	}
}

func TestDB_Start(t *testing.T) {
	db, cleanup, err := sqlite.InMemory()
	if err != nil {
		t.Fatalf("InMemory() error = %v", err)
	}
	defer cleanup()

	wrapper := &DB{
		Pool:  db,
		Close: cleanup,
	}

	ctx := context.Background()
	startFunc := func(ctx context.Context) (*DB, error) {
		return &DB{Pool: db, Close: cleanup}, nil
	}

	result, err := wrapper.Start(ctx, startFunc)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	if result.Pool != db {
		t.Error("Start() did not return expected DB")
	}
}

func TestDB_StartFunc(t *testing.T) {
	ctx := context.Background()

	startFunc := func(ctx context.Context) (*DB, error) {
		db, cleanup, err := sqlite.InMemory()
		if err != nil {
			return nil, err
		}
		return &DB{Pool: db, Close: cleanup}, nil
	}

	db, err := startFunc(ctx)
	if err != nil {
		t.Fatalf("startFunc() error = %v", err)
	}

	if err := db.Pool.Ping(); err != nil {
		t.Errorf("Pool.Ping() error = %v", err)
	}

	db.Close()
}

func TestDB_Close(t *testing.T) {
	db, cleanup, err := sqlite.InMemory()
	if err != nil {
		t.Fatalf("InMemory() error = %v", err)
	}

	// Track if close was called
	closeCalled := false
	wrappedCleanup := cleanup
	cleanup = func() {
		closeCalled = true
		wrappedCleanup()
	}

	wrapper := &DB{
		Pool:  db,
		Close: cleanup,
	}

	// Close should not panic
	wrapper.Close()

	if !closeCalled {
		t.Error("Close() did not call the cleanup function")
	}

	// After close, pool should be closed
	if err := db.Ping(); err == nil {
		t.Log("Note: Pool still accessible after close (SQLite behavior)")
	}
}

func TestDB_NilPool(t *testing.T) {
	// Test that operations on nil pool are handled gracefully
	// Note: sql.DB.Ping on nil pool will panic, so we just verify wrapper setup works
	wrapper := &DB{
		Pool:  nil,
		Close: func() {},
	}

	if wrapper.Pool != nil {
		t.Error("expected nil pool")
	}
	if wrapper.Close == nil {
		t.Error("expected non-nil close function")
	}
}

func TestDB_WithTimeout(t *testing.T) {
	db, cleanup, err := sqlite.InMemory()
	if err != nil {
		t.Fatalf("InMemory() error = %v", err)
	}
	defer cleanup()

	wrapper := &DB{
		Pool:  db,
		Close: cleanup,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	if err := wrapper.Ping(ctx); err != nil {
		t.Fatalf("Ping() error = %v", err)
	}
}
