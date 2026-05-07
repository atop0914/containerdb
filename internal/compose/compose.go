// Package compose provides Docker Compose file parsing and generation utilities.
package compose

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Service represents a single service in a Docker Compose file.
type Service struct {
	Image       string            `yaml:"image,omitempty"`
	ContainerName string          `yaml:"container_name,omitempty"`
	Environment map[string]string `yaml:"environment,omitempty"`
	Ports       []string          `yaml:"ports,omitempty"`
	Volumes     []string          `yaml:"volumes,omitempty"`
	Command     string            `yaml:"command,omitempty"`
	Networks    []string          `yaml:"networks,omitempty"`
}

// ComposeFile represents a Docker Compose file structure.
type ComposeFile struct {
	Version  string             `yaml:"version,omitempty"`
	Services map[string]Service `yaml:"services"`
	Networks map[string]any     `yaml:"networks,omitempty"`
	Volumes  map[string]any     `yaml:"volumes,omitempty"`
}

// Parse reads and parses a Docker Compose file from the given path.
func Parse(path string) (*ComposeFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read compose file: %w", err)
	}

	var cf ComposeFile
	if err := yaml.Unmarshal(data, &cf); err != nil {
		return nil, fmt.Errorf("failed to parse compose file: %w", err)
	}

	return &cf, nil
}

// ParseFromString parses a Docker Compose file from a string.
func ParseFromString(content string) (*ComposeFile, error) {
	var cf ComposeFile
	if err := yaml.Unmarshal([]byte(content), &cf); err != nil {
		return nil, fmt.Errorf("failed to parse compose content: %w", err)
	}
	return &cf, nil
}

// GetService returns a service by name.
func (c *ComposeFile) GetService(name string) (Service, bool) {
	svc, ok := c.Services[name]
	return svc, ok
}

// GetEnv returns an environment variable value from a service.
func (s *Service) GetEnv(key string) (string, bool) {
	val, ok := s.Environment[key]
	return val, ok
}

// GetPort returns the host port for a container port mapping.
// Format expected: "host:container" or just "container" if not mapped.
func (s *Service) GetPort(containerPort string) (string, error) {
	for _, portMapping := range s.Ports {
		parts := strings.Split(portMapping, ":")
		if len(parts) == 2 && parts[1] == containerPort {
			return parts[0], nil
		}
	}
	return "", fmt.Errorf("port %s not found in service ports", containerPort)
}

// GenerateMySQLCompose generates a Docker Compose snippet for MySQL.
func GenerateMySQLCompose(serviceName, image, username, password, database string, port string) Service {
	env := map[string]string{
		"MYSQL_ROOT_PASSWORD": password,
		"MYSQL_DATABASE":     database,
	}
	if username != "root" {
		env["MYSQL_USER"] = username
		env["MYSQL_PASSWORD"] = password
	}

	svc := Service{
		Image:         image,
		ContainerName: serviceName,
		Environment:   env,
	}

	if port != "" {
		svc.Ports = []string{port + ":3306"}
	}

	return svc
}

// GeneratePostgresCompose generates a Docker Compose snippet for PostgreSQL.
func GeneratePostgresCompose(serviceName, image, username, password, database string, port string) Service {
	env := map[string]string{
		"POSTGRES_USER":     username,
		"POSTGRES_PASSWORD": password,
		"POSTGRES_DB":       database,
	}

	svc := Service{
		Image:         image,
		ContainerName: serviceName,
		Environment:   env,
	}

	if port != "" {
		svc.Ports = []string{port + ":5432"}
	}

	return svc
}

// BuildComposeFile creates a complete Docker Compose file from services.
func BuildComposeFile(services map[string]Service) *ComposeFile {
	return &ComposeFile{
		Version:  "3.8",
		Services: services,
	}
}

// WriteToFile writes the compose file to disk.
func (c *ComposeFile) WriteToFile(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal compose file: %w", err)
	}

	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write compose file: %w", err)
	}

	return nil
}

// String returns the YAML representation of the compose file.
func (c *ComposeFile) String() (string, error) {
	data, err := yaml.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("failed to marshal compose file: %w", err)
	}
	return string(data), nil
}
