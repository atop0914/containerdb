package benchmarks

import (
	"testing"

	"github.com/atop0914/containerdb-bootcamp/internal/config"
	internalsvc "github.com/atop0914/containerdb-bootcamp/internal/compose"
	pkgcompose "github.com/atop0914/containerdb-bootcamp/pkg/compose"
)

// BenchmarkCompose_GenerateMySQLCompose measures MySQL service generation (internal).
func BenchmarkCompose_GenerateMySQLCompose(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = internalsvc.GenerateMySQLCompose("test-db", "mysql:8.0", "root", "password", "testdb", "3306")
	}
}

// BenchmarkCompose_GeneratePostgresCompose measures PostgreSQL service generation (internal).
func BenchmarkCompose_GeneratePostgresCompose(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = internalsvc.GeneratePostgresCompose("test-db", "postgres:16-alpine", "postgres", "password", "testdb", "5432")
	}
}

// BenchmarkCompose_GenerateMySQLService measures MySQL service generation (pkg).
func BenchmarkCompose_GenerateMySQLService(b *testing.B) {
	cfg := config.DefaultMySQLConfig()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pkgcompose.GenerateMySQLService("test-db", cfg)
	}
}

// BenchmarkCompose_GeneratePostgresService measures PostgreSQL service generation (pkg).
func BenchmarkCompose_GeneratePostgresService(b *testing.B) {
	cfg := config.DefaultPostgresConfig()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pkgcompose.GeneratePostgresService("test-db", cfg)
	}
}

// BenchmarkCompose_GenerateMySQLServiceWithHealthCheck measures MySQL service with healthcheck.
func BenchmarkCompose_GenerateMySQLServiceWithHealthCheck(b *testing.B) {
	cfg := config.DefaultMySQLConfig()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pkgcompose.GenerateMySQLServiceWithHealthCheck("test-db", cfg)
	}
}

// BenchmarkCompose_GeneratePostgresServiceWithHealthCheck measures PostgreSQL service with healthcheck.
func BenchmarkCompose_GeneratePostgresServiceWithHealthCheck(b *testing.B) {
	cfg := config.DefaultPostgresConfig()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pkgcompose.GeneratePostgresServiceWithHealthCheck("test-db", cfg)
	}
}

// BenchmarkCompose_BuildComposeFile measures compose file building.
func BenchmarkCompose_BuildComposeFile(b *testing.B) {
	svc := internalsvc.GenerateMySQLCompose("test-db", "mysql:8.0", "root", "password", "testdb", "3306")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		services := map[string]internalsvc.Service{"mysql": svc}
		_ = internalsvc.BuildComposeFile(services)
	}
}

// BenchmarkCompose_NewRunner measures runner creation.
func BenchmarkCompose_NewRunner(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = pkgcompose.NewRunner("test-project")
	}
}

// BenchmarkCompose_GenerateFile measures full compose file generation.
func BenchmarkCompose_GenerateFile(b *testing.B) {
	cfg := config.DefaultMySQLConfig()
	svc := pkgcompose.GenerateMySQLService("test-db", cfg)
	services := map[string]pkgcompose.Service{"mysql": svc}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runner := pkgcompose.NewRunner("test-project")
		runner.GenerateFileTo(services, "/dev/null")
	}
}

// BenchmarkCompose_TemplateMySQL measures MySQL template generation.
func BenchmarkCompose_TemplateMySQL(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = pkgcompose.TemplateMySQL()
	}
}

// BenchmarkCompose_TemplatePostgres measures PostgreSQL template generation.
func BenchmarkCompose_TemplatePostgres(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = pkgcompose.TemplatePostgres()
	}
}

// BenchmarkCompose_TemplateMySQLPostgres measures combined template generation.
func BenchmarkCompose_TemplateMySQLPostgres(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = pkgcompose.TemplateMySQLPostgres()
	}
}
