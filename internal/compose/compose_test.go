package compose

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	content := `
version: "3.8"
services:
  mysql:
    image: mysql:8.0
    container_name: test_mysql
    environment:
      MYSQL_ROOT_PASSWORD: secret
      MYSQL_DATABASE: testdb
    ports:
      - "3306:3306"
`
	cf, err := ParseFromString(content)
	if err != nil {
		t.Fatalf("ParseFromString failed: %v", err)
	}

	if len(cf.Services) != 1 {
		t.Errorf("expected 1 service, got %d", len(cf.Services))
	}

	svc, ok := cf.GetService("mysql")
	if !ok {
		t.Fatal("mysql service not found")
	}

	if svc.Image != "mysql:8.0" {
		t.Errorf("expected image mysql:8.0, got %s", svc.Image)
	}

	if svc.ContainerName != "test_mysql" {
		t.Errorf("expected container_name test_mysql, got %s", svc.ContainerName)
	}

	val, ok := svc.GetEnv("MYSQL_ROOT_PASSWORD")
	if !ok || val != "secret" {
		t.Errorf("expected MYSQL_ROOT_PASSWORD=secret, got %s, ok=%v", val, ok)
	}
}

func TestGetService(t *testing.T) {
	cf := &ComposeFile{
		Services: map[string]Service{
			"pg": {Image: "postgres:16"},
		},
	}

	svc, ok := cf.GetService("pg")
	if !ok {
		t.Fatal("pg service not found")
	}
	if svc.Image != "postgres:16" {
		t.Errorf("expected postgres:16, got %s", svc.Image)
	}

	_, ok = cf.GetService("nonexistent")
	if ok {
		t.Error("nonexistent service should not be found")
	}
}

func TestServiceGetPort(t *testing.T) {
	svc := Service{
		Ports: []string{"3306:3306", "33060:33060"},
	}

	port, err := svc.GetPort("3306")
	if err != nil {
		t.Errorf("GetPort failed: %v", err)
	}
	if port != "3306" {
		t.Errorf("expected 3306, got %s", port)
	}

	_, err = svc.GetPort("9999")
	if err == nil {
		t.Error("expected error for unmapped port")
	}
}

func TestGenerateMySQLCompose(t *testing.T) {
	svc := GenerateMySQLCompose("mydb", "mysql:8.0", "root", "secret", "appdb", "3306")

	if svc.Image != "mysql:8.0" {
		t.Errorf("expected mysql:8.0, got %s", svc.Image)
	}

	if svc.ContainerName != "mydb" {
		t.Errorf("expected mydb, got %s", svc.ContainerName)
	}

	if len(svc.Ports) != 1 || svc.Ports[0] != "3306:3306" {
		t.Errorf("expected ports [3306:3306], got %v", svc.Ports)
	}

	val, ok := svc.GetEnv("MYSQL_ROOT_PASSWORD")
	if !ok || val != "secret" {
		t.Errorf("expected MYSQL_ROOT_PASSWORD=secret, got %s, ok=%v", val, ok)
	}
}

func TestGeneratePostgresCompose(t *testing.T) {
	svc := GeneratePostgresCompose("pgdb", "postgres:16-alpine", "pguser", "pgpass", "mydb", "5432")

	if svc.Image != "postgres:16-alpine" {
		t.Errorf("expected postgres:16-alpine, got %s", svc.Image)
	}

	if len(svc.Ports) != 1 || svc.Ports[0] != "5432:5432" {
		t.Errorf("expected ports [5432:5432], got %v", svc.Ports)
	}

	val, ok := svc.GetEnv("POSTGRES_USER")
	if !ok || val != "pguser" {
		t.Errorf("expected POSTGRES_USER=pguser, got %s, ok=%v", val, ok)
	}
}

func TestBuildComposeFile(t *testing.T) {
	services := map[string]Service{
		"mysql": GenerateMySQLCompose("mysql", "mysql:8.0", "root", "pass", "db", "3306"),
		"pg":    GeneratePostgresCompose("pg", "postgres:16", "pg", "pg", "db", "5432"),
	}

	cf := BuildComposeFile(services)

	if cf.Version != "3.8" {
		t.Errorf("expected version 3.8, got %s", cf.Version)
	}

	if len(cf.Services) != 2 {
		t.Errorf("expected 2 services, got %d", len(cf.Services))
	}
}

func TestWriteToFile(t *testing.T) {
	cf := BuildComposeFile(map[string]Service{
		"mysql": GenerateMySQLCompose("test", "mysql:8.0", "root", "pass", "db", "3306"),
	})

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "docker-compose.yml")

	err := cf.WriteToFile(path)
	if err != nil {
		t.Fatalf("WriteToFile failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if !strings.Contains(string(data), "mysql:8.0") {
		t.Error("generated compose file does not contain expected image")
	}

	if !strings.Contains(string(data), "MYSQL_ROOT_PASSWORD") {
		t.Error("generated compose file does not contain expected env var")
	}
}

func TestComposeString(t *testing.T) {
	cf := &ComposeFile{
		Version: "3.8",
		Services: map[string]Service{
			"test": {Image: "alpine:latest"},
		},
	}

	s, err := cf.String()
	if err != nil {
		t.Fatalf("String() failed: %v", err)
	}

	if !strings.Contains(s, "version:") {
		t.Error("String() should contain version")
	}

	if !strings.Contains(s, "services:") {
		t.Error("String() should contain services")
	}
}

func TestGenerateMySQLCompose_NonRootUser(t *testing.T) {
	// Test that non-root users get MYSQL_USER and MYSQL_PASSWORD set
	svc := GenerateMySQLCompose("mydb", "mysql:8.0", "appuser", "apppass", "appdb", "")

	_, ok := svc.GetEnv("MYSQL_USER")
	if !ok {
		t.Error("expected MYSQL_USER to be set for non-root user")
	}

	_, ok = svc.GetEnv("MYSQL_PASSWORD")
	if !ok {
		t.Error("expected MYSQL_PASSWORD to be set for non-root user")
	}

	// root user should not get MYSQL_USER set
	svcRoot := GenerateMySQLCompose("rootdb", "mysql:8.0", "root", "secret", "rootdb", "")
	_, ok = svcRoot.GetEnv("MYSQL_USER")
	if ok {
		t.Error("MYSQL_USER should not be set for root user")
	}
}
