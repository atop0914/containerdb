package examples

import (
	"context"
	"fmt"
	"testing"

	"github.com/atop0914/containerdb/pkg/postgres"
)

func TestPostgres_Example(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgres example in short mode")
	}
	ctx := context.Background()

	db, cleanup, err := postgres.New(ctx)
	if err != nil {
		t.Fatalf("failed to start postgres: %v", err)
	}
	defer cleanup()

	var version string
	err = db.QueryRowContext(ctx, "SELECT version()").Scan(&version)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	fmt.Println("PostgreSQL version:", version)
	t.Log("Postgres test passed")
}
