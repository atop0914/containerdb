package examples

import (
	"context"
	"fmt"
	"testing"

	"github.com/atop0914/containerdb/pkg/mysql"
)

func TestMySQL_Example(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping mysql example in short mode")
	}
	ctx := context.Background()
	
	db, cleanup, err := mysql.New(ctx)
	if err != nil {
		t.Fatalf("failed to start mysql: %v", err)
	}
	defer cleanup()
	
	var version string
	err = db.QueryRowContext(ctx, "SELECT VERSION()").Scan(&version)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	
	fmt.Println("MySQL version:", version)
	t.Log("MySQL test passed")
}
