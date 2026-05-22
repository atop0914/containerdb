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

	"github.com/atop0914/containerdb/internal/config"
	internalsvc "github.com/atop0914/containerdb/internal/compose"
)

// Re-export types from internal/compose for public API use.
type (
	// Service represents a single service in a Docker Compose file.
	Service = internalsvc.Service
	// HealthCheck represents a Docker health check configuration.
	HealthCheck = internalsvc.HealthCheck
	// ComposeFile represents a Docker Compose file structure.
	ComposeFile = internalsvc.ComposeFile
)

// Re-export functions from internal/compose for public API use.
var (
	// BuildComposeFile creates a complete Docker Compose file from services.
	BuildComposeFile = internalsvc.BuildComposeFile
	// Parse reads and parses a Docker Compose file from the given path.
	Parse = internalsvc.Parse
	// ParseFromString parses a Docker Compose file from a string.
	ParseFromString = internalsvc.ParseFromString
	// DetectComposeVersion detects whether docker compose v2 or v1 is available.
	DetectComposeVersion = internalsvc.DetectComposeVersion
)

// ComposeVersion represents the Docker Compose CLI version.
type ComposeVersion string

const (
	// VersionV1 is the legacy docker-compose (Python) CLI.
	VersionV1 ComposeVersion = "v1"
	// VersionV2 is the new docker compose (Go plugin) CLI.
	VersionV2 ComposeVersion = "v2"
)

// Runner handles docker-compose operations for containerized databases.
type Runner struct {
	composeFile string
	projectName string
	version     ComposeVersion
}

// NewRunner creates a new compose runner with the specified project name.
func NewRunner(projectName string) *Runner {
	return &Runner{
		composeFile: "docker-compose.yml",
		projectName: projectName,
		version:     VersionV2, // default to v2
	}
}

// NewRunnerWithFile creates a new compose runner with a custom compose file path.
func NewRunnerWithFile(projectName, composeFile string) *Runner {
	return &Runner{
		composeFile: composeFile,
		projectName: projectName,
		version:     VersionV2,
	}
}

// SetVersion sets the compose CLI version to use.
func (r *Runner) SetVersion(v ComposeVersion) {
	r.version = v
}

// GetVersion returns the current compose version setting.
func (r *Runner) GetVersion() ComposeVersion {
	return r.version
}

// DetectVersion auto-detects the available compose version.
func (r *Runner) DetectVersion() error {
	v, err := internalsvc.DetectComposeVersion()
	if err != nil {
		return err
	}
	r.version = ComposeVersion(v)
	return nil
}

// buildCommand constructs the appropriate docker compose command.
func (r *Runner) buildCommand(args ...string) *exec.Cmd {
	if r.version == VersionV1 {
		cmdArgs := append([]string{"-p", r.projectName, "-f", r.composeFile}, args...)
		return exec.Command("docker-compose", cmdArgs...)
	}
	// V2: docker compose -p project -f file args...
	cmdArgs := append([]string{"compose", "-p", r.projectName, "-f", r.composeFile}, args...)
	return exec.Command("docker", cmdArgs...)
}

// GenerateMySQLService generates a compose service for MySQL from config.
func GenerateMySQLService(name string, cfg *config.MySQLConfig) Service {
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
func GeneratePostgresService(name string, cfg *config.PostgresConfig) Service {
	return internalsvc.GeneratePostgresCompose(
		name,
		cfg.Image,
		cfg.Username,
		cfg.Password,
		cfg.Database,
		"", // port is handled by docker-compose
	)
}

// GenerateMySQLServiceWithHealthCheck generates a compose service with healthcheck.
func GenerateMySQLServiceWithHealthCheck(name string, cfg *config.MySQLConfig) Service {
	svc := GenerateMySQLService(name, cfg)
	svc.AddHealthCheck("10s", "5s", 5,
		"CMD", "mysqladmin", "ping", "-h", "localhost",
	)
	svc.SetRestart("unless-stopped")
	return svc
}

// GeneratePostgresServiceWithHealthCheck generates a compose service with healthcheck.
func GeneratePostgresServiceWithHealthCheck(name string, cfg *config.PostgresConfig) Service {
	svc := GeneratePostgresService(name, cfg)
	svc.AddHealthCheck("10s", "5s", 5,
		"CMD", "pg_isready", "-U", cfg.Username,
	)
	svc.SetRestart("unless-stopped")
	return svc
}

// GenerateFile creates a docker-compose.yml file with the given services.
func (r *Runner) GenerateFile(services map[string]Service, dir string) error {
	cf := internalsvc.BuildComposeFile(services)
	path := filepath.Join(dir, r.composeFile)
	return cf.WriteToFile(path)
}

// GenerateFileTo generates a docker-compose.yml file at a specific path.
func (r *Runner) GenerateFileTo(services map[string]Service, path string) error {
	cf := internalsvc.BuildComposeFile(services)
	return cf.WriteToFile(path)
}

// Up starts the docker-compose services.
func (r *Runner) Up(ctx context.Context, dir string) error {
	cmd := r.buildCommand("up", "-d")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("compose up failed: %w\n%s", err, output)
	}
	return nil
}

// UpWithWait starts services and waits for healthchecks to pass.
func (r *Runner) UpWithWait(ctx context.Context, dir string) error {
	cmd := r.buildCommand("up", "-d", "--wait")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("compose up --wait failed: %w\n%s", err, output)
	}
	return nil
}

// Down stops and removes the docker-compose services.
func (r *Runner) Down(ctx context.Context, dir string) error {
	cmd := r.buildCommand("down")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("compose down failed: %w\n%s", err, output)
	}
	return nil
}

// DownWithVolumes stops and removes services and their volumes.
func (r *Runner) DownWithVolumes(ctx context.Context, dir string) error {
	cmd := r.buildCommand("down", "-v")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("compose down -v failed: %w\n%s", err, output)
	}
	return nil
}

// Ps shows the status of docker-compose services.
func (r *Runner) Ps(ctx context.Context, dir string) (string, error) {
	cmd := r.buildCommand("ps")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("compose ps failed: %w", err)
	}
	return string(output), nil
}

// Logs shows the logs of docker-compose services.
func (r *Runner) Logs(ctx context.Context, dir string, service string) (string, error) {
	args := []string{"logs"}
	if service != "" {
		args = append(args, service)
	}
	cmd := r.buildCommand(args...)
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("compose logs failed: %w", err)
	}
	return string(output), nil
}

// Status shows the health status of services.
func (r *Runner) Status(ctx context.Context, dir string) (string, error) {
	cmd := r.buildCommand("ps", "--format", "json")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("compose status failed: %w", err)
	}
	return string(output), nil
}

// ParseExisting reads and parses an existing docker-compose.yml file.
func ParseExisting(path string) (*ComposeFile, error) {
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
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: "10s"
      timeout: "5s"
      retries: 5
      start_period: "10s"
    restart: "unless-stopped"
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
    healthcheck:
      test: ["CMD", "pg_isready", "-U", "postgres"]
      interval: "10s"
      timeout: "5s"
      retries: 5
      start_period: "10s"
    restart: "unless-stopped"
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
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: "10s"
      timeout: "5s"
      retries: 5
      start_period: "10s"
    restart: "unless-stopped"

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
    healthcheck:
      test: ["CMD", "pg_isready", "-U", "postgres"]
      interval: "10s"
      timeout: "5s"
      retries: 5
      start_period: "10s"
    restart: "unless-stopped"

volumes:
  mysql_data:
  pg_data:
`
}

// EnsureDockerCompose checks if docker compose is available.
func EnsureDockerCompose() error {
	v, err := internalsvc.DetectComposeVersion()
	if err != nil {
		return err
	}
	if v == "v2" {
		return nil
	}
	// Also check v1
	if _, err := exec.LookPath("docker-compose"); err == nil {
		return nil
	}
	return fmt.Errorf("docker compose not available")
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
