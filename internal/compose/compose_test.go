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
		t.Errorf("expected MYSQL_ROOT_PASSWORD=*** got %s, ok=%v", val, ok)
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
		t.Errorf("expected MYSQL_ROOT_PASSWORD=*** got %s, ok=%v", val, ok)
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
		t.Errorf("expected POSTGRES_USER=pguser got %s, ok=%v", val, ok)
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

func TestAddHealthCheck(t *testing.T) {
	svc := GenerateMySQLCompose("mysql", "mysql:8.0", "root", "pass", "db", "3306")

	if svc.HealthCheck != nil {
		t.Error("expected no healthcheck before adding")
	}

	svc.AddHealthCheck("10s", "5s", 3, "CMD", "mysqladmin", "ping", "-h", "localhost")

	if svc.HealthCheck == nil {
		t.Fatal("expected healthcheck after adding")
	}

	if len(svc.HealthCheck.Test) != 5 {
		t.Errorf("expected 5 test cmd parts, got %d", len(svc.HealthCheck.Test))
	}

	if svc.HealthCheck.Interval != "10s" {
		t.Errorf("expected interval 10s, got %s", svc.HealthCheck.Interval)
	}

	if svc.HealthCheck.Timeout != "5s" {
		t.Errorf("expected timeout 5s, got %s", svc.HealthCheck.Timeout)
	}

	if svc.HealthCheck.Retries != 3 {
		t.Errorf("expected retries 3, got %d", svc.HealthCheck.Retries)
	}

	if svc.HealthCheck.StartPeriod != "10s" {
		t.Errorf("expected start_period 10s, got %s", svc.HealthCheck.StartPeriod)
	}
}

func TestAddDependsOn(t *testing.T) {
	svc := Service{Image: "app:latest"}

	svc.AddDependsOn("mysql", "service_healthy")

	if svc.DependsOn == nil {
		t.Fatal("expected depends_on after adding")
	}

	dep, ok := svc.DependsOn["mysql"]
	if !ok {
		t.Fatal("expected mysql dependency")
	}

	depMap, ok := dep.(map[string]string)
	if !ok {
		t.Fatalf("expected map[string]string, got %T", dep)
	}

	if depMap["condition"] != "service_healthy" {
		t.Errorf("expected condition service_healthy, got %s", depMap["condition"])
	}

	// Test simple depends_on without condition
	svc.AddDependsOn("redis", "")
	_, ok = svc.DependsOn["redis"]
	if !ok {
		t.Error("expected redis dependency")
	}
}

func TestSetRestart(t *testing.T) {
	svc := Service{Image: "mysql:8.0"}

	svc.SetRestart("unless-stopped")

	if svc.Restart != "unless-stopped" {
		t.Errorf("expected restart unless-stopped, got %s", svc.Restart)
	}
}

func TestParseWithHealthCheck(t *testing.T) {
	content := `
version: "3.8"
services:
  mysql:
    image: mysql:8.0
    container_name: test_mysql
    environment:
      MYSQL_ROOT_PASSWORD: secret
    ports:
      - "3306:3306"
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: "10s"
      timeout: "5s"
      retries: 3
      start_period: "10s"
    restart: "unless-stopped"
`
	cf, err := ParseFromString(content)
	if err != nil {
		t.Fatalf("ParseFromString failed: %v", err)
	}

	svc, ok := cf.GetService("mysql")
	if !ok {
		t.Fatal("mysql service not found")
	}

	if svc.HealthCheck == nil {
		t.Fatal("expected healthcheck to be parsed")
	}

	if len(svc.HealthCheck.Test) != 5 {
		t.Errorf("expected 5 test cmd parts, got %d", len(svc.HealthCheck.Test))
	}

	if svc.HealthCheck.Interval != "10s" {
		t.Errorf("expected interval 10s, got %s", svc.HealthCheck.Interval)
	}

	if svc.Restart != "unless-stopped" {
		t.Errorf("expected restart unless-stopped, got %s", svc.Restart)
	}
}

func TestParseWithDependsOn(t *testing.T) {
	content := `
version: "3.8"
services:
  app:
    image: myapp:latest
    depends_on:
      mysql:
        condition: service_healthy
  mysql:
    image: mysql:8.0
`
	cf, err := ParseFromString(content)
	if err != nil {
		t.Fatalf("ParseFromString failed: %v", err)
	}

	app, ok := cf.GetService("app")
	if !ok {
		t.Fatal("app service not found")
	}

	if app.DependsOn == nil {
		t.Fatal("expected depends_on to be parsed")
	}

	dep, ok := app.DependsOn["mysql"]
	if !ok {
		t.Fatal("expected mysql dependency")
	}

	depMap, ok := dep.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", dep)
	}

	if depMap["condition"] != "service_healthy" {
		t.Errorf("expected condition service_healthy, got %v", depMap["condition"])
	}
}

func TestDetectComposeVersion(t *testing.T) {
	// This test just verifies the function doesn't panic
	// In CI environments, docker may not be available
	_, err := DetectComposeVersion()
	if err != nil {
		t.Logf("DetectComposeVersion: %v (docker not available in test env)", err)
	}
}
