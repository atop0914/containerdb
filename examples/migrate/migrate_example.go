package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/atop0914/containerdb/pkg/migrate"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Create a temporary directory for our example
	tmpDir, err := os.MkdirTemp("", "migrate-example-*")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create migrations directory
	migrationsDir := filepath.Join(tmpDir, "migrations")
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		log.Fatal(err)
	}

	// Create sample migration files
	upSQL := `CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	downSQL := `DROP TABLE users;`

	// Write migration files
	version := "20240101000000"
	upFile := filepath.Join(migrationsDir, version+"_create_users.up.sql")
	downFile := filepath.Join(migrationsDir, version+"_create_users.down.sql")

	if err := os.WriteFile(upFile, []byte(upSQL), 0644); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(downFile, []byte(downSQL), 0644); err != nil {
		log.Fatal(err)
	}

	// Create a SQLite database
	dbPath := filepath.Join(tmpDir, "example.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()

	// Create a migration runner
	runner := migrate.NewRunner(db,
		migrate.WithDir(migrationsDir),
		migrate.WithTableName("schema_migrations"),
	)

	// Run migrations
	fmt.Println("Running migrations...")
	if err := runner.Up(ctx); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Migrations completed!")

	// Check migration status
	statuses, err := runner.Status(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nMigration Status:")
	for _, s := range statuses {
		status := "Pending"
		if s.Applied {
			status = "Applied"
		}
		fmt.Printf("  %s: %s (%s)\n", s.Version, s.Name, status)
	}

	// Demonstrate creating a new migration
	fmt.Println("\nCreating new migration...")
	newVersion, err := migrate.CreateMigration(migrationsDir, "Add Age Column")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created migration with version: %s\n", newVersion)

	// List all migrations
	fmt.Println("\nAll migrations:")
	migrations, err := migrate.ListMigrations(migrationsDir)
	if err != nil {
		log.Fatal(err)
	}
	for _, m := range migrations {
		fmt.Printf("  %s: %s\n", m.Version, filepath.Base(m.Name))
	}

	fmt.Println("\nExample completed successfully!")
}
