package migrate

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReadMigrations(t *testing.T) {
	// Create temp migration directory
	tmpDir, err := os.MkdirTemp("", "migrations-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test migration files
	files := []struct {
		name    string
		content string
	}{
		{"001_create_users.up.sql", "CREATE TABLE users (id INT PRIMARY KEY);"},
		{"002_create_posts.up.sql", "CREATE TABLE posts (id INT PRIMARY KEY);"},
		{"001_create_users.down.sql", "DROP TABLE users;"},
		{"002_create_posts.down.sql", "DROP TABLE posts;"},
	}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, f.name), []byte(f.content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	migrations, err := ReadMigrations(tmpDir)
	if err != nil {
		t.Fatalf("ReadMigrations failed: %v", err)
	}

	if len(migrations) != 2 {
		t.Errorf("expected 2 migrations, got %d", len(migrations))
	}

	if migrations[0].Version != "001" {
		t.Errorf("expected version 001, got %s", migrations[0].Version)
	}
	if migrations[1].Version != "002" {
		t.Errorf("expected version 002, got %s", migrations[1].Version)
	}
}

func TestReadMigrationsEmptyDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "migrations-empty-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	migrations, err := ReadMigrations(tmpDir)
	if err != nil {
		t.Fatalf("ReadMigrations failed: %v", err)
	}

	if len(migrations) != 0 {
		t.Errorf("expected 0 migrations, got %d", len(migrations))
	}
}

func TestReadMigrationsNonSQLFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "migrations-non-sql-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Write non-SQL file
	if err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# migrations"), 0644); err != nil {
		t.Fatal(err)
	}

	migrations, err := ReadMigrations(tmpDir)
	if err != nil {
		t.Fatalf("ReadMigrations failed: %v", err)
	}

	if len(migrations) != 0 {
		t.Errorf("expected 0 migrations, got %d", len(migrations))
	}
}

func TestMigrationSorting(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "migrations-sort-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create migrations in random order
	files := []string{
		"003_third.up.sql",
		"001_first.up.sql",
		"002_second.up.sql",
	}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, f), []byte("SELECT 1;"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	migrations, err := ReadMigrations(tmpDir)
	if err != nil {
		t.Fatalf("ReadMigrations failed: %v", err)
	}

	expected := []string{"001", "002", "003"}
	for i, exp := range expected {
		if migrations[i].Version != exp {
			t.Errorf("position %d: expected %s, got %s", i, exp, migrations[i].Version)
		}
	}
}

func TestConfigOptions(t *testing.T) {
	cfg := defaultConfig()
	if cfg.TableName != "schema_migrations" {
		t.Errorf("expected default table name 'schema_migrations', got %s", cfg.TableName)
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("expected default timeout 30s, got %v", cfg.Timeout)
	}

	opt := WithTableName("custom_table")
	opt(&cfg)
	if cfg.TableName != "custom_table" {
		t.Errorf("expected 'custom_table', got %s", cfg.TableName)
	}

	opt2 := WithTimeout(60 * time.Second)
	opt2(&cfg)
	if cfg.Timeout != 60*time.Second {
		t.Errorf("expected timeout 60s, got %v", cfg.Timeout)
	}
}

func TestReadDirectionalMigrations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "migrations-directional-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create both up and down migrations
	files := []struct {
		name    string
		content string
	}{
		{"001_create_users.up.sql", "CREATE TABLE users (id INT PRIMARY KEY);"},
		{"001_create_users.down.sql", "DROP TABLE users;"},
		{"002_create_posts.up.sql", "CREATE TABLE posts (id INT PRIMARY KEY);"},
		{"002_create_posts.down.sql", "DROP TABLE posts;"},
	}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, f.name), []byte(f.content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Read up migrations
	upMigrations, err := ReadDirectionalMigrations(tmpDir, Up)
	if err != nil {
		t.Fatalf("ReadDirectionalMigrations(Up) failed: %v", err)
	}
	if len(upMigrations) != 2 {
		t.Errorf("expected 2 up migrations, got %d", len(upMigrations))
	}
	for _, m := range upMigrations {
		if m.Direction != Up {
			t.Errorf("expected Up direction, got %v", m.Direction)
		}
	}

	// Read down migrations
	downMigrations, err := ReadDirectionalMigrations(tmpDir, Down)
	if err != nil {
		t.Fatalf("ReadDirectionalMigrations(Down) failed: %v", err)
	}
	if len(downMigrations) != 2 {
		t.Errorf("expected 2 down migrations, got %d", len(downMigrations))
	}
	for _, m := range downMigrations {
		if m.Direction != Down {
			t.Errorf("expected Down direction, got %v", m.Direction)
		}
	}
}

func TestCreateMigration(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "migrations-create-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	version, err := CreateMigration(tmpDir, "Create Users Table")
	if err != nil {
		t.Fatalf("CreateMigration failed: %v", err)
	}

	if version == "" {
		t.Error("expected non-empty version")
	}

	// Check that files were created
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 2 {
		t.Errorf("expected 2 files, got %d", len(entries))
	}

	// Check file names
	foundUp := false
	foundDown := false
	for _, entry := range entries {
		name := entry.Name()
		if name == version+"_create_users_table.up.sql" {
			foundUp = true
		}
		if name == version+"_create_users_table.down.sql" {
			foundDown = true
		}
	}

	if !foundUp {
		t.Error("up migration file not found")
	}
	if !foundDown {
		t.Error("down migration file not found")
	}
}

func TestCreateMigrationWithSubdirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "migrations-subdir-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	subDir := filepath.Join(tmpDir, "migrations")
	version, err := CreateMigration(subDir, "test")
	if err != nil {
		t.Fatalf("CreateMigration failed: %v", err)
	}

	if version == "" {
		t.Error("expected non-empty version")
	}

	// Check that directory was created
	if _, err := os.Stat(subDir); os.IsNotExist(err) {
		t.Error("migrations directory was not created")
	}
}
