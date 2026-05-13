package config

import (
	"testing"
	"time"
)

func TestDefaultMySQLConfig(t *testing.T) {
	cfg := DefaultMySQLConfig()

	if cfg.Image != "mysql:8.0" {
		t.Errorf("expected Image=mysql:8.0, got %s", cfg.Image)
	}
	if cfg.Username != "root" {
		t.Errorf("expected Username=root, got %s", cfg.Username)
	}
	if cfg.Password != "rootpassword" {
		t.Errorf("expected Password=rootpassword, got %s", cfg.Password)
	}
	if cfg.Database != "testdb" {
		t.Errorf("expected Database=testdb, got %s", cfg.Database)
	}
	if cfg.HealthCheckTimeout != 30*time.Second {
		t.Errorf("expected HealthCheckTimeout=30s, got %v", cfg.HealthCheckTimeout)
	}
	if cfg.MaxOpenConns != 10 {
		t.Errorf("expected MaxOpenConns=10, got %d", cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns != 5 {
		t.Errorf("expected MaxIdleConns=5, got %d", cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime != time.Hour {
		t.Errorf("expected ConnMaxLifetime=1h, got %v", cfg.ConnMaxLifetime)
	}
}

func TestDefaultPostgresConfig(t *testing.T) {
	cfg := DefaultPostgresConfig()

	if cfg.Image != "postgres:16-alpine" {
		t.Errorf("expected Image=postgres:16-alpine, got %s", cfg.Image)
	}
	if cfg.Username != "postgres" {
		t.Errorf("expected Username=postgres, got %s", cfg.Username)
	}
	if cfg.Password != "postgres" {
		t.Errorf("expected Password=postgres, got %s", cfg.Password)
	}
	if cfg.Database != "testdb" {
		t.Errorf("expected Database=testdb, got %s", cfg.Database)
	}
	if cfg.HealthCheckTimeout != 30*time.Second {
		t.Errorf("expected HealthCheckTimeout=30s, got %v", cfg.HealthCheckTimeout)
	}
	if cfg.MaxOpenConns != 10 {
		t.Errorf("expected MaxOpenConns=10, got %d", cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns != 5 {
		t.Errorf("expected MaxIdleConns=5, got %d", cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime != time.Hour {
		t.Errorf("expected ConnMaxLifetime=1h, got %v", cfg.ConnMaxLifetime)
	}
}

func TestDefaultSQLiteConfig(t *testing.T) {
	cfg := DefaultSQLiteConfig()

	if cfg.Mode != "memory" {
		t.Errorf("expected Mode=memory, got %s", cfg.Mode)
	}
	if cfg.Path != "" {
		t.Errorf("expected Path='', got %s", cfg.Path)
	}
	if cfg.Cache != "shared" {
		t.Errorf("expected Cache=shared, got %s", cfg.Cache)
	}
}

func TestMySQLConfig_Modify(t *testing.T) {
	cfg := DefaultMySQLConfig()
	cfg.Image = "mysql:5.7"
	cfg.Username = "custom"
	cfg.MaxOpenConns = 20

	if cfg.Image != "mysql:5.7" {
		t.Errorf("expected Image=mysql:5.7, got %s", cfg.Image)
	}
	if cfg.Username != "custom" {
		t.Errorf("expected Username=custom, got %s", cfg.Username)
	}
	if cfg.MaxOpenConns != 20 {
		t.Errorf("expected MaxOpenConns=20, got %d", cfg.MaxOpenConns)
	}
}

func TestPostgresConfig_Modify(t *testing.T) {
	cfg := DefaultPostgresConfig()
	cfg.Image = "postgres:15-alpine"
	cfg.Database = "mydb"
	cfg.MaxIdleConns = 10

	if cfg.Image != "postgres:15-alpine" {
		t.Errorf("expected Image=postgres:15-alpine, got %s", cfg.Image)
	}
	if cfg.Database != "mydb" {
		t.Errorf("expected Database=mydb, got %s", cfg.Database)
	}
	if cfg.MaxIdleConns != 10 {
		t.Errorf("expected MaxIdleConns=10, got %d", cfg.MaxIdleConns)
	}
}

func TestSQLiteConfig_Modify(t *testing.T) {
	cfg := DefaultSQLiteConfig()
	cfg.Mode = "file"
	cfg.Path = "/tmp/test.db"
	cfg.Cache = "private"

	if cfg.Mode != "file" {
		t.Errorf("expected Mode=file, got %s", cfg.Mode)
	}
	if cfg.Path != "/tmp/test.db" {
		t.Errorf("expected Path=/tmp/test.db, got %s", cfg.Path)
	}
	if cfg.Cache != "private" {
		t.Errorf("expected Cache=private, got %s", cfg.Cache)
	}
}
