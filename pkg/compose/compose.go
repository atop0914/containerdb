// Package compose provides high-level Docker Compose integration for ContainerDB.
// It allows generating compose files from existing containerdb configurations
// and running databases via docker-compose.
package compose

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/atop0914/containerdb-bootcamp/internal/config"
	internalsvc "github.com/atop0914/containerdb-bootcamp/internal/compose"
)

// Runner handles docker-compose operations for containerized databases.
type Runner struct {
	composeFile string
	projectName string
}

// NewRunner creates a new compose runner with the specified project name.
func NewRunner(projectName string) *Runner {
	return &Runner{
		composeFile: "docker-compose.yml",
		projectName: projectName,
	}
}

// NewRunnerWithFile creates a new compose runner with a custom compose file path.
func NewRunnerWithFile(projectName, composeFile string) *Runner {
	return &Runner{
		composeFile: composeFile,
		projectName: projectName,
	}
}

// GenerateMySQLService generates a compose service for MySQL from config.
func GenerateMySQLService(name string, cfg *config.MySQLConfig) internalsvc.Service {
	return internalsvc.GenerateMySQLCompose(
		name,
		cfg.Image,
		cfg.Username,
		cfg.Password,
		cfg.Database,
		"", // port is handled by docker-compose
	)
}

// GeneratePostgresService generates a compose service for PostgreSQL from config.
func GeneratePostgresService(name string, cfg *config.PostgresConfig) internalsvc.Service {
	return internalsvc.GeneratePostgresCompose(
		name,
		cfg.Image,
		cfg.Username,
		cfg.Password,
		cfg.Database,
		"", // port is handled by docker-compose
	)
}

// GenerateFile creates a docker-compose.yml file with the given services.
func (r *Runner) GenerateFile(services map[string]internalsvc.Service, dir string) error {
	cf := internalsvc.BuildComposeFile(services)
	path := filepath.Join(dir, r.composeFile)
	return cf.WriteToFile(path)
}

// GenerateFileTo generates a docker-compose.yml file at a specific path.
func (r *Runner) GenerateFileTo(services map[string]internalsvc.Service, path string) error {
	cf := internalsvc.BuildComposeFile(services)
	return cf.WriteToFile(path)
}

// Up starts the docker-compose services.
func (r *Runner) Up(ctx context.Context, dir string) error {
	cmd := exec.Command("docker-compose", "-p", r.projectName, "-f", r.composeFile, "up", "-d")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker-compose up failed: %w\n%s", err, output)
	}
	return nil
}

// Down stops and removes the docker-compose services.
func (r *Runner) Down(ctx context.Context, dir string) error {
	cmd := exec.Command("docker-compose", "-p", r.projectName, "-f", r.composeFile, "down")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker-compose down failed: %w\n%s", err, output)
	}
	return nil
}

// Ps shows the status of docker-compose services.
func (r *Runner) Ps(ctx context.Context, dir string) (string, error) {
	cmd := exec.Command("docker-compose", "-p", r.projectName, "-f", r.composeFile, "ps")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("docker-compose ps failed: %w", err)
	}
	return string(output), nil
}

// Logs shows the logs of docker-compose services.
func (r *Runner) Logs(ctx context.Context, dir string, service string) (string, error) {
	args := []string{"-p", r.projectName, "-f", r.composeFile, "logs"}
	if service != "" {
		args = append(args, service)
	}
	cmd := exec.Command("docker-compose", args...)
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("docker-compose logs failed: %w", err)
	}
	return string(output), nil
}

// ParseExisting reads and parses an existing docker-compose.yml file.
func ParseExisting(path string) (*internalsvc.ComposeFile, error) {
	return internalsvc.Parse(path)
}

// ParseExistingServices reads an existing compose file and extracts service configs.
func ParseExistingServices(path string) (map[string]ServiceConfig, error) {
	cf, err := internalsvc.Parse(path)
	if err != nil {
		return nil, err
	}

	result := make(map[string]ServiceConfig)
	for name, svc := range cf.Services {
		cfg := ServiceConfig{
			Name:     name,
			Image:    svc.Image,
			Port:     extractFirstPort(svc.Ports),
			Username: extractEnv(svc.Environment, []string{"MYSQL_USER", "POSTGRES_USER"}),
			Password: extractEnv(svc.Environment, []string{"MYSQL_ROOT_PASSWORD", "POSTGRES_PASSWORD"}),
			Database: extractEnv(svc.Environment, []string{"MYSQL_DATABASE", "POSTGRES_DB"}),
		}
		result[name] = cfg
	}
	return result, nil
}

// ServiceConfig holds parsed service configuration.
type ServiceConfig struct {
	Name     string
	Image    string
	Port     string
	Username string
	Password string
	Database string
}

// IsMySQL returns true if the service appears to be MySQL.
func (s *ServiceConfig) IsMySQL() bool {
	return strings.Contains(s.Image, "mysql") || s.Port == "3306"
}

// IsPostgres returns true if the service appears to be PostgreSQL.
func (s *ServiceConfig) IsPostgres() bool {
	return strings.Contains(s.Image, "postgres") || s.Port == "5432"
}

func extractFirstPort(ports []string) string {
	for _, p := range ports {
		parts := strings.Split(p, ":")
		if len(parts) >= 2 {
			return parts[0]
		}
	}
	return ""
}

func extractEnv(env map[string]string, keys []string) string {
	for _, key := range keys {
		if val, ok := env[key]; ok {
			return val
		}
	}
	return ""
}

// TemplateMySQL generates a docker-compose.yml template for MySQL.
func TemplateMySQL() string {
	return `version: "3.8"
services:
  mysql:
    image: mysql:8.0
    container_name: containerdb_mysql
    environment:
      MYSQL_ROOT_PASSWORD: rootpassword
      MYSQL_DATABASE: testdb
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
volumes:
  mysql_data:
`
}

// TemplatePostgres generates a docker-compose.yml template for PostgreSQL.
func TemplatePostgres() string {
	return `version: "3.8"
services:
  postgres:
    image: postgres:16-alpine
    container_name: containerdb_postgres
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: testdb
    ports:
      - "5432:5432"
    volumes:
      - pg_data:/var/lib/postgresql/data
volumes:
  pg_data:
`
}

// TemplateMySQLPostgres generates a docker-compose.yml template with both MySQL and PostgreSQL.
func TemplateMySQLPostgres() string {
	return `version: "3.8"
services:
  mysql:
    image: mysql:8.0
    container_name: containerdb_mysql
    environment:
      MYSQL_ROOT_PASSWORD: rootpassword
      MYSQL_DATABASE: testdb
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql

  postgres:
    image: postgres:16-alpine
    container_name: containerdb_postgres
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: testdb
    ports:
      - "5432:5432"
    volumes:
      - pg_data:/var/lib/postgresql/data

volumes:
  mysql_data:
  pg_data:
`
}

// EnsureDockerCompose checks if docker-compose is available.
func EnsureDockerCompose() error {
	_, err := exec.LookPath("docker-compose")
	if err != nil {
		return fmt.Errorf("docker-compose not found in PATH: %w", err)
	}
	return nil
}

// EnsureDocker checks if docker is available.
func EnsureDocker() error {
	cmd := exec.Command("docker", "info")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker is not available: %w\n%s", err, output)
	}
	return nil
}

// WriteTemplate writes a compose template to a file.
func WriteTemplate(path, template string) error {
	return os.WriteFile(path, []byte(template), 0644)
}
