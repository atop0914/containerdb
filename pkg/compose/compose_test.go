package compose

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/atop0914/containerdb-bootcamp/internal/config"
	"github.com/atop0914/containerdb-bootcamp/internal/compose"
)

func TestNewRunner(t *testing.T) {
	r := NewRunner("testproject")
	if r.projectName != "testproject" {
		t.Errorf("expected projectName testproject, got %s", r.projectName)
	}
	if r.composeFile != "docker-compose.yml" {
		t.Errorf("expected composeFile docker-compose.yml, got %s", r.composeFile)
	}
}

func TestNewRunnerWithFile(t *testing.T) {
	r := NewRunnerWithFile("custom", "custom-compose.yml")
	if r.projectName != "custom" {
		t.Errorf("expected projectName custom, got %s", r.projectName)
	}
	if r.composeFile != "custom-compose.yml" {
		t.Errorf("expected composeFile custom-compose.yml, got %s", r.composeFile)
	}
}

func TestGenerateMySQLService(t *testing.T) {
	cfg := config.DefaultMySQLConfig()
	cfg.Username = "testuser"
	cfg.Password = "testpass"
	cfg.Database = "testdb"

	svc := GenerateMySQLService("mysql_test", cfg)

	if svc.Image != "mysql:8.0" {
		t.Errorf("expected mysql:8.0, got %s", svc.Image)
	}
	if svc.ContainerName != "mysql_test" {
		t.Errorf("expected mysql_test, got %s", svc.ContainerName)
	}
}

func TestGeneratePostgresService(t *testing.T) {
	cfg := config.DefaultPostgresConfig()
	cfg.Username = "pguser"
	cfg.Password = "pgpass"
	cfg.Database = "pgdb"

	svc := GeneratePostgresService("postgres_test", cfg)

	if svc.Image != "postgres:16-alpine" {
		t.Errorf("expected postgres:16-alpine, got %s", svc.Image)
	}
	if svc.ContainerName != "postgres_test" {
		t.Errorf("expected postgres_test, got %s", svc.ContainerName)
	}
}

func TestRunnerGenerateFile(t *testing.T) {
	r := NewRunner("test")
	services := map[string]compose.Service{
		"mysql": GenerateMySQLService("mysql", config.DefaultMySQLConfig()),
	}

	tmpDir := t.TempDir()
	err := r.GenerateFile(services, tmpDir)
	if err != nil {
		t.Fatalf("GenerateFile failed: %v", err)
	}

	path := filepath.Join(tmpDir, "docker-compose.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if !strings.Contains(string(data), "mysql:8.0") {
		t.Error("generated file should contain mysql:8.0")
	}
}

func TestParseExistingServices(t *testing.T) {
	content := `
version: "3.8"
services:
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: secret
      MYSQL_DATABASE: appdb
    ports:
      - "3306:3306"
  postgres:
    image: postgres:16
    environment:
      POSTGRES_USER: pguser
      POSTGRES_PASSWORD: pgpass
      POSTGRES_DB: pgdb
    ports:
      - "5432:5432"
`

	tmpDir := t.TempDir()
	composePath := filepath.Join(tmpDir, "docker-compose.yml")
	err := os.WriteFile(composePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	services, err := ParseExistingServices(composePath)
	if err != nil {
		t.Fatalf("ParseExistingServices failed: %v", err)
	}

	if len(services) != 2 {
		t.Errorf("expected 2 services, got %d", len(services))
	}

	mysql, ok := services["mysql"]
	if !ok {
		t.Fatal("mysql service not found")
	}
	if mysql.Password != "secret" {
		t.Errorf("expected password secret, got %s", mysql.Password)
	}
	if mysql.Database != "appdb" {
		t.Errorf("expected database appdb, got %s", mysql.Database)
	}

	pg, ok := services["postgres"]
	if !ok {
		t.Fatal("postgres service not found")
	}
	if pg.Username != "pguser" {
		t.Errorf("expected username pguser, got %s", pg.Username)
	}
}

func TestServiceConfigIsMySQL(t *testing.T) {
	cfg := &ServiceConfig{Image: "mysql:8.0", Port: "33060"}
	if !cfg.IsMySQL() {
		t.Error("expected IsMySQL to return true for mysql:8.0")
	}

	cfg = &ServiceConfig{Image: "alpine", Port: "3306"}
	if !cfg.IsMySQL() {
		t.Error("expected IsMySQL to return true for port 3306")
	}

	cfg = &ServiceConfig{Image: "alpine", Port: "5432"}
	if cfg.IsMySQL() {
		t.Error("expected IsMySQL to return false for port 5432")
	}
}

func TestServiceConfigIsPostgres(t *testing.T) {
	cfg := &ServiceConfig{Image: "postgres:16", Port: "5432"}
	if !cfg.IsPostgres() {
		t.Error("expected IsPostgres to return true for postgres:16")
	}

	cfg = &ServiceConfig{Image: "alpine", Port: "5432"}
	if !cfg.IsPostgres() {
		t.Error("expected IsPostgres to return true for port 5432")
	}

	cfg = &ServiceConfig{Image: "alpine", Port: "3306"}
	if cfg.IsPostgres() {
		t.Error("expected IsPostgres to return false for port 3306")
	}
}

func TestTemplateMySQL(t *testing.T) {
	template := TemplateMySQL()
	if !strings.Contains(template, "mysql:8.0") {
		t.Error("template should contain mysql:8.0")
	}
	if !strings.Contains(template, "MYSQL_ROOT_PASSWORD") {
		t.Error("template should contain MYSQL_ROOT_PASSWORD")
	}
}

func TestTemplatePostgres(t *testing.T) {
	template := TemplatePostgres()
	if !strings.Contains(template, "postgres:16-alpine") {
		t.Error("template should contain postgres:16-alpine")
	}
	if !strings.Contains(template, "POSTGRES_USER") {
		t.Error("template should contain POSTGRES_USER")
	}
}

func TestTemplateMySQLPostgres(t *testing.T) {
	template := TemplateMySQLPostgres()
	if !strings.Contains(template, "mysql:8.0") {
		t.Error("template should contain mysql:8.0")
	}
	if !strings.Contains(template, "postgres:16-alpine") {
		t.Error("template should contain postgres:16-alpine")
	}
	if !strings.Contains(template, "volumes:") {
		t.Error("template should contain volumes")
	}
}

func TestWriteTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test-compose.yml")

	err := WriteTemplate(path, TemplateMySQL())
	if err != nil {
		t.Fatalf("WriteTemplate failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if !strings.Contains(string(data), "mysql:8.0") {
		t.Error("written template should contain mysql:8.0")
	}
}
