// Example: Using ContainerDB with Docker Compose
//
// This example demonstrates how to:
// 1. Generate a docker-compose.yml with healthchecks
// 2. Parse an existing compose file
// 3. Manage services via the Runner API
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/atop0914/containerdb-bootcamp/internal/config"
	"github.com/atop0914/containerdb-bootcamp/pkg/compose"
)

func main() {
	// Example 1: Generate a compose file with healthchecks
	fmt.Println("=== Generating compose file with healthchecks ===")
	generateComposeWithHealthcheck()

	// Example 2: Parse an existing compose file
	fmt.Println("\n=== Parsing existing compose file ===")
	parseExistingCompose()

	// Example 3: Use the Runner API
	fmt.Println("\n=== Using Runner API ===")
	useRunnerAPI()
}

func generateComposeWithHealthcheck() {
	mysqlCfg := config.DefaultMySQLConfig()
	mysqlCfg.Username = "myuser"
	mysqlCfg.Password = "mypassword"
	mysqlCfg.Database = "myapp"

	pgCfg := config.DefaultPostgresConfig()
	pgCfg.Username = "pguser"
	pgCfg.Password = "pgpassword"
	pgCfg.Database = "myapp"

	// Generate services with healthchecks
	mysqlSvc := compose.GenerateMySQLServiceWithHealthCheck("mysql", mysqlCfg)
	mysqlSvc.Volumes = []string{"mysql_data:/var/lib/mysql"}

	pgSvc := compose.GeneratePostgresServiceWithHealthCheck("postgres", pgCfg)
	pgSvc.Volumes = []string{"pg_data:/var/lib/postgresql/data"}

	services := map[string]compose.Service{
		"mysql":    mysqlSvc,
		"postgres": pgSvc,
	}

	cf := compose.BuildComposeFile(services)
	yaml, err := cf.String()
	if err != nil {
		log.Fatalf("Failed to generate compose: %v", err)
	}

	fmt.Println(yaml)
}

func parseExistingCompose() {
	tmpDir, err := os.MkdirTemp("", "containerdb-example")
	if err != nil {
		log.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	composePath := filepath.Join(tmpDir, "docker-compose.yml")

	sampleCompose := `version: "3.8"
services:
  mysql:
    image: mysql:8.0
    container_name: my_mysql
    environment:
      MYSQL_ROOT_PASSWORD: secret123
      MYSQL_DATABASE: myapp
    ports:
      - "3306:3306"
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: "10s"
      timeout: "5s"
      retries: 5
  postgres:
    image: postgres:16-alpine
    container_name: my_postgres
    environment:
      POSTGRES_USER: pguser
      POSTGRES_PASSWORD: pgpass
      POSTGRES_DB: mydb
    ports:
      - "5432:5432"
`

	if err := os.WriteFile(composePath, []byte(sampleCompose), 0644); err != nil {
		log.Fatalf("Failed to write compose file: %v", err)
	}

	services, err := compose.ParseExistingServices(composePath)
	if err != nil {
		log.Fatalf("Failed to parse compose: %v", err)
	}

	for name, cfg := range services {
		fmt.Printf("Service: %s\n", name)
		fmt.Printf("  Image: %s\n", cfg.Image)
		fmt.Printf("  Port: %s\n", cfg.Port)
		fmt.Printf("  IsMySQL: %v\n", cfg.IsMySQL())
		fmt.Printf("  IsPostgres: %v\n", cfg.IsPostgres())
		fmt.Println()
	}
}

func useRunnerAPI() {
	r := compose.NewRunner("myproject")

	if err := r.DetectVersion(); err != nil {
		fmt.Printf("Docker Compose not available: %v\n", err)
		return
	}

	fmt.Printf("Compose version: %s\n", r.GetVersion())

	mysqlSvc := compose.GenerateMySQLService("mysql", config.DefaultMySQLConfig())
	services := map[string]compose.Service{"mysql": mysqlSvc}

	tmpDir, err := os.MkdirTemp("", "containerdb-runner")
	if err != nil {
		log.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := r.GenerateFile(services, tmpDir); err != nil {
		log.Fatalf("Failed to generate file: %v", err)
	}

	fmt.Println("Generated compose file in", tmpDir)

	ctx := context.Background()
	_ = ctx
}
